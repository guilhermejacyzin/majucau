package bling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/database/gen"
)

type fakeBlingOAuthRepository struct {
	connection       database.IntegrationConnection
	authorizing      bool
	completed        bool
	disconnected     bool
	testSucceeded    bool
	lastErrorCode    string
	scopes           []string
	accessExpiresAt  *time.Time
	refreshExpiresAt *time.Time
}

func (r *fakeBlingOAuthRepository) Connection(context.Context) (database.IntegrationConnection, error) {
	return r.connection, nil
}
func (r *fakeBlingOAuthRepository) MarkAuthorizing(context.Context) error {
	r.authorizing = true
	return nil
}
func (r *fakeBlingOAuthRepository) CompleteOAuth(_ context.Context, scopes []string, accessExpiresAt, refreshExpiresAt *time.Time) error {
	r.completed, r.scopes, r.accessExpiresAt, r.refreshExpiresAt = true, scopes, accessExpiresAt, refreshExpiresAt
	return nil
}
func (r *fakeBlingOAuthRepository) MarkDisconnected(context.Context) error {
	r.disconnected = true
	return nil
}
func (r *fakeBlingOAuthRepository) MarkOAuthError(_ context.Context, code string) error {
	r.lastErrorCode = code
	return nil
}
func (r *fakeBlingOAuthRepository) RecordTestSuccess(context.Context) error {
	r.testSucceeded = true
	return nil
}
func (r *fakeBlingOAuthRepository) RecordTestFailure(_ context.Context, code string) error {
	r.lastErrorCode = code
	return nil
}

type fakeOAuthHTTPDoer func(*http.Request) (*http.Response, error)

func (f fakeOAuthHTTPDoer) Do(req *http.Request) (*http.Response, error) { return f(req) }

func TestBuildBlingAuthorizeURLDoesNotContainSecret(t *testing.T) {
	got, err := buildBlingAuthorizeURL("client-id", "http://127.0.0.1:43821/callback", "state-value")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("client_id") != "client-id" || parsed.Query().Get("state") != "state-value" {
		t.Fatalf("unexpected authorization URL: %s", got)
	}
	if strings.Contains(got, "client-secret") {
		t.Fatalf("authorization URL leaked secret: %s", got)
	}
}

func TestValidateBlingLoopbackRedirect(t *testing.T) {
	for _, raw := range []string{"https://app.example.test/callback", "http://localhost:43821/callback", "http://127.0.0.1/callback"} {
		if _, err := validateBlingLoopbackRedirect(raw); err == nil {
			t.Fatalf("redirect should be rejected: %s", raw)
		}
	}
	if parsed, err := validateBlingLoopbackRedirect("http://127.0.0.1:43821/callback"); err != nil || parsed.Port() != "43821" {
		t.Fatalf("valid loopback redirect rejected: %v", err)
	}
}

func TestBlingOAuthLoopbackSessionExchangesAndTestsWithoutLeakingToken(t *testing.T) {
	probePayload := `{"data":[],"pagination":{"page":1,"limit":1,"total":0}}`
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		body := probePayload
		if strings.Contains(req.URL.Path, "/oauth/token") {
			body = `{"access_token":"access-token-private","refresh_token":"refresh-token-private","token_type":"Bearer","scope":"read_orders read_finance","expires_in":21600}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	secretStore := &memorySecretStore{values: map[string][]byte{}}
	initial, _ := json.Marshal(blingSecretBundle{ClientSecret: "client-secret-private"}) // test-only fixture
	if err := secretStore.Put(context.Background(), BlingCredentialSecretRef, initial); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	clientID, redirectURI := "client-id", fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	repo := &fakeBlingOAuthRepository{connection: database.IntegrationConnection{ClientID: &clientID, RedirectUri: &redirectURI}}
	service := newBlingOAuthService(repo, secretStore, httpDoer)
	result, err := service.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !repo.authorizing || result.Status != blingOAuthStatusAuthorizing || result.AuthorizationURL == "" {
		t.Fatalf("unexpected start result: %+v", result)
	}
	authorizeURL, _ := url.Parse(result.AuthorizationURL)
	callbackURL := fmt.Sprintf("%s?code=code-value&state=%s", redirectURI, url.QueryEscape(authorizeURL.Query().Get("state")))
	callbackResponse, err := http.Get(callbackURL)
	if err != nil {
		t.Fatal(err)
	}
	_ = callbackResponse.Body.Close()
	deadline := time.Now().Add(2 * time.Second)
	var status BlingOAuthStatusResult
	for time.Now().Before(deadline) {
		status, _ = service.Status(context.Background(), result.SessionID)
		if status.Status != blingOAuthStatusAuthorizing {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status.Status != blingOAuthStatusConnected || !repo.completed || !strings.Contains(status.Message, "sucesso") {
		t.Fatalf("unexpected OAuth status: %+v, repo=%+v", status, repo)
	}
	stored, err := secretStore.Get(context.Background(), BlingCredentialSecretRef)
	if err != nil {
		t.Fatal(err)
	}
	var bundle blingSecretBundle
	if err := json.Unmarshal(stored, &bundle); err != nil || bundle.AccessToken != "access-token-private" {
		t.Fatalf("token was not stored in worker vault: %v %+v", err, bundle)
	}
	if strings.Contains(status.Message, "access-token-private") || strings.Contains(result.AuthorizationURL, "client-secret-private") {
		t.Fatal("OAuth result leaked a secret or token")
	}
}

func TestOAuthRepositoryTimestampHelper(t *testing.T) {
	if got := nullableTimestamp(nil); got.Valid {
		t.Fatal("nil timestamp must remain SQL NULL")
	}
	value := time.Now().UTC()
	got := nullableTimestamp(&value)
	if !got.Valid || !got.Time.Equal(value) {
		t.Fatalf("unexpected timestamp: %+v", got)
	}
}

func TestBlingOAuthDisconnectRemovesTokensAndPreservesClientSecret(t *testing.T) {
	store := &memorySecretStore{values: map[string][]byte{}}
	initial, err := json.Marshal(blingSecretBundle{ClientSecret: "client-secret", AccessToken: "access-token", RefreshToken: "refresh-token", TokenType: "Bearer", Scope: "read_finance", AccessTokenExpiresUnix: 123, RefreshTokenExpiresUnix: 456})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), BlingCredentialSecretRef, initial); err != nil {
		t.Fatal(err)
	}
	repo := &fakeBlingOAuthRepository{}
	service := newBlingOAuthService(repo, store, fakeOAuthHTTPDoer(func(*http.Request) (*http.Response, error) { return nil, fmt.Errorf("unused") }))
	if err := service.Disconnect(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !repo.disconnected {
		t.Fatal("repository was not marked disconnected")
	}
	stored, err := store.Get(context.Background(), BlingCredentialSecretRef)
	if err != nil {
		t.Fatal(err)
	}
	var got blingSecretBundle
	if err := json.Unmarshal(stored, &got); err != nil {
		t.Fatal(err)
	}
	if got.ClientSecret != "client-secret" || got.AccessToken != "" || got.RefreshToken != "" || got.Scope != "" || got.AccessTokenExpiresUnix != 0 || got.RefreshTokenExpiresUnix != 0 {
		t.Fatalf("disconnect did not clear only authorization tokens: %+v", got)
	}
}

