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
	"sync"
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
	completedCh      chan struct{}
	disconnectedCh   chan struct{}
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
	if r.completedCh != nil {
		close(r.completedCh)
	}
	return nil
}
func (r *fakeBlingOAuthRepository) MarkDisconnected(context.Context) error {
	r.disconnected = true
	if r.disconnectedCh != nil {
		close(r.disconnectedCh)
	}
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

func TestResolveBlingOAuthCallbackSupportsRegisteredBridge(t *testing.T) {
	callback, err := resolveBlingOAuthCallback(blingBridgeRedirectURI)
	if err != nil {
		t.Fatal(err)
	}
	if got := callback.String(); got != blingBridgeLocalCallback {
		t.Fatalf("unexpected local callback: %s", got)
	}
	if _, err := resolveBlingOAuthCallback("https://untrusted.example.test/oauth/callback"); err == nil {
		t.Fatal("untrusted HTTPS callback should be rejected")
	}
}

func TestBlingOAuthBridgeReturnsCodeToLocalSessionAndExchangesRegisteredRedirect(t *testing.T) {
	var tokenRedirect string
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		body := `{"data":[],"pagination":{"page":1,"limit":1,"total":0}}`
		if strings.Contains(req.URL.Path, "/oauth/token") {
			if err := req.ParseForm(); err != nil {
				return nil, err
			}
			tokenRedirect = req.Form.Get("redirect_uri")
			body = `{"access_token":"access-token-private","refresh_token":"refresh-token-private","token_type":"Bearer","scope":"read_finance","expires_in":21600}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	secretStore := &memorySecretStore{values: map[string][]byte{}}
	initial, _ := json.Marshal(blingSecretBundle{ClientSecret: "client-secret-private"}) // test-only fixture
	if err := secretStore.Put(context.Background(), BlingCredentialSecretRef, initial); err != nil {
		t.Fatal(err)
	}
	clientID, redirectURI := "client-id", blingBridgeRedirectURI
	repo := &fakeBlingOAuthRepository{connection: database.IntegrationConnection{ClientID: &clientID, RedirectUri: &redirectURI}}
	service := newBlingOAuthService(repo, secretStore, httpDoer)
	var localListener net.Listener
	service.listen = func(network, address string) (net.Listener, error) {
		if network != "tcp" || address != "127.0.0.1:43821" {
			t.Fatalf("unexpected listener address: %s %s", network, address)
		}
		var err error
		localListener, err = net.Listen("tcp", "127.0.0.1:0")
		return localListener, err
	}
	result, err := service.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	authorizeURL, err := url.Parse(result.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	state := authorizeURL.Query().Get("state")
	if authorizeURL.Query().Get("redirect_uri") != blingBridgeRedirectURI || len(state) != 64 {
		t.Fatalf("authorization URL did not preserve the registered bridge: %s", result.AuthorizationURL)
	}
	oversizedURL := fmt.Sprintf("http://%s/callback?code=%s&state=%s", localListener.Addr().String(), strings.Repeat("x", blingOAuthCodeMaxBytes+1), url.QueryEscape(state))
	oversizedResponse, err := http.Get(oversizedURL)
	if err != nil {
		t.Fatal(err)
	}
	_ = oversizedResponse.Body.Close()
	if oversizedResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("oversized local OAuth code was not rejected: status=%d", oversizedResponse.StatusCode)
	}
	callbackURL := fmt.Sprintf("http://%s/callback?code=short-lived-code&state=%s", localListener.Addr().String(), url.QueryEscape(state))
	response, err := http.Get(callbackURL)
	if err != nil {
		t.Fatal(err)
	}
	callbackPage, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("Referrer-Policy") != "no-referrer" || !strings.Contains(response.Header.Get("Content-Security-Policy"), "script-src 'nonce-") || !strings.Contains(string(callbackPage), "history.replaceState") || strings.Contains(string(callbackPage), "short-lived-code") {
		t.Fatalf("local callback did not clear the short-lived code from the browser URL safely: status=%d headers=%v", response.StatusCode, response.Header)
	}
	_ = response.Body.Close()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status, _ := service.Status(context.Background(), result.SessionID)
		if status.Status == blingOAuthStatusConnected {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	status, _ := service.Status(context.Background(), result.SessionID)
	if status.Status != blingOAuthStatusConnected || !repo.completed {
		t.Fatalf("bridge OAuth session did not complete: %+v", status)
	}
	if tokenRedirect != blingBridgeRedirectURI {
		t.Fatalf("token exchange used redirect %q, want the registered bridge", tokenRedirect)
	}
}

func TestBlingOAuthLoopbackSessionExchangesAndTestsWithoutLeakingToken(t *testing.T) {
	probePayload := `{"data":[],"pagination":{"page":1,"limit":1,"total":0}}`
	tokenRequestStarted := make(chan struct{})
	releaseTokenExchange := make(chan struct{})
	var releaseOnce sync.Once
	releaseExchange := func() { releaseOnce.Do(func() { close(releaseTokenExchange) }) }
	defer releaseExchange()
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		body := probePayload
		if strings.Contains(req.URL.Path, "/oauth/token") {
			close(tokenRequestStarted)
			<-releaseTokenExchange
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
	callbackPage, err := io.ReadAll(callbackResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	if callbackResponse.StatusCode != http.StatusOK || callbackResponse.Header.Get("Cache-Control") != "no-store" || callbackResponse.Header.Get("Referrer-Policy") != "no-referrer" || !strings.Contains(callbackResponse.Header.Get("Content-Security-Policy"), "script-src 'nonce-") || !strings.Contains(string(callbackPage), "history.replaceState") || strings.Contains(string(callbackPage), "code-value") {
		t.Fatalf("callback response did not clear the short-lived code from the browser URL safely: status=%d headers=%v", callbackResponse.StatusCode, callbackResponse.Header)
	}
	_ = callbackResponse.Body.Close()
	select {
	case <-tokenRequestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("authorization exchange did not start")
	}
	type oauthTestOutcome struct {
		result BlingOAuthTestResult
		err    error
	}
	testDone := make(chan oauthTestOutcome, 1)
	go func() {
		result, err := service.Test(context.Background())
		testDone <- oauthTestOutcome{result: result, err: err}
	}()
	earlyTest := false
	var testOutcome oauthTestOutcome
	select {
	case testOutcome = <-testDone:
		earlyTest = true
	case <-time.After(50 * time.Millisecond):
	}
	releaseExchange()
	deadline := time.Now().Add(2 * time.Second)
	var status BlingOAuthStatusResult
	for time.Now().Before(deadline) {
		status, _ = service.Status(context.Background(), result.SessionID)
		if status.Status != blingOAuthStatusAuthorizing {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !earlyTest {
		select {
		case testOutcome = <-testDone:
		case <-time.After(2 * time.Second):
			t.Fatal("OAuth test did not resume after authorization completed")
		}
	}
	if earlyTest {
		t.Fatal("OAuth test read credentials before the authorization exchange finished")
	}
	if testOutcome.err != nil || testOutcome.result.Status != "SUCCESS" {
		t.Fatalf("concurrent OAuth test failed after authorization: %+v, %v", testOutcome.result, testOutcome.err)
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
	initial, err := json.Marshal(blingSecretBundle{ClientSecret: "cs-fixture", AccessToken: "at-fixture", RefreshToken: "rt-fixture", TokenType: "Bearer", Scope: "read_finance", AccessTokenExpiresUnix: 123, RefreshTokenExpiresUnix: 456})
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
	if got.ClientSecret != "cs-fixture" || got.AccessToken != "" || got.RefreshToken != "" || got.Scope != "" || got.AccessTokenExpiresUnix != 0 || got.RefreshTokenExpiresUnix != 0 {
		t.Fatalf("disconnect did not clear only authorization tokens: %+v", got)
	}
}

func TestBlingOAuthDisconnectWaitsForInFlightAuthorization(t *testing.T) {
	tokenRequestStarted := make(chan struct{})
	releaseTokenExchange := make(chan struct{})
	var releaseOnce sync.Once
	releaseExchange := func() { releaseOnce.Do(func() { close(releaseTokenExchange) }) }
	defer releaseExchange()

	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		if strings.HasSuffix(req.URL.Path, "/oauth/token") {
			close(tokenRequestStarted)
			<-releaseTokenExchange
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"access_token":"test-only-access","refresh_token":"test-only-refresh","token_type":"Bearer","scope":"read_finance","expires_in":21600}`)), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"data":[],"pagination":{"page":1,"limit":1,"total":0}}`)), Header: make(http.Header)}, nil
	})
	store := &memorySecretStore{values: map[string][]byte{}}
	initial, err := json.Marshal(blingSecretBundle{ClientSecret: "client-secret-fixture"}) // test-only fixture
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), BlingCredentialSecretRef, initial); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	clientID, redirectURI := "client-id", fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	repo := &fakeBlingOAuthRepository{
		connection:     database.IntegrationConnection{ClientID: &clientID, RedirectUri: &redirectURI},
		completedCh:    make(chan struct{}),
		disconnectedCh: make(chan struct{}),
	}
	service := newBlingOAuthService(repo, store, httpDoer)
	started, err := service.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	authorizeURL, err := url.Parse(started.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	callbackURL := fmt.Sprintf("%s?code=code-value&state=%s", redirectURI, url.QueryEscape(authorizeURL.Query().Get("state")))
	var callbackResponse *http.Response
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		callbackResponse, err = http.Get(callbackURL)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("OAuth callback failed: %v", err)
	}
	_ = callbackResponse.Body.Close()
	select {
	case <-tokenRequestStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("authorization exchange did not start")
	}

	disconnectStarted := make(chan struct{})
	disconnectDone := make(chan error, 1)
	go func() {
		close(disconnectStarted)
		disconnectDone <- service.Disconnect(context.Background())
	}()
	<-disconnectStarted
	earlyDisconnect := false
	var disconnectErr error
	select {
	case disconnectErr = <-disconnectDone:
		earlyDisconnect = true
	case <-time.After(50 * time.Millisecond):
	}
	releaseExchange()
	select {
	case <-repo.completedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("authorization did not finish after token response")
	}
	if !earlyDisconnect {
		select {
		case disconnectErr = <-disconnectDone:
		case <-time.After(2 * time.Second):
			t.Fatal("disconnect did not finish after authorization")
		}
	}
	if disconnectErr != nil {
		t.Fatalf("disconnect failed: %v", disconnectErr)
	}
	if earlyDisconnect {
		t.Fatal("disconnect returned while an authorization token exchange was in progress")
	}
	select {
	case <-repo.disconnectedCh:
	default:
		t.Fatal("repository did not record disconnection")
	}

	stored, err := store.Get(context.Background(), BlingCredentialSecretRef)
	if err != nil {
		t.Fatal(err)
	}
	var bundle blingSecretBundle
	if err := json.Unmarshal(stored, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.ClientSecret != "client-secret-fixture" || bundle.AccessToken != "" || bundle.RefreshToken != "" {
		t.Fatalf("authorization overwrote disconnected credentials: %+v", bundle)
	}
}
