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
	"sync/atomic"
	"testing"
	"time"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/security"
)

const oauthFailureProbePayload = `{"data":[],"pagination":{"page":1,"limit":1,"total":0}}`

type failingPutSecretStore struct {
	base *memorySecretStore
	err  error
}

func (s *failingPutSecretStore) Put(ctx context.Context, ref string, value []byte) error {
	if s.err != nil {
		return s.err
	}
	return s.base.Put(ctx, ref, value)
}
func (s *failingPutSecretStore) Get(ctx context.Context, ref string) ([]byte, error) {
	return s.base.Get(ctx, ref)
}
func (s *failingPutSecretStore) Delete(ctx context.Context, ref string) error {
	return s.base.Delete(ctx, ref)
}

func newOAuthFailureHarness(t *testing.T, store security.SecretStore, httpDoer HTTPDoer) (*BlingOAuthService, *fakeBlingOAuthRepository, BlingOAuthStartResult, string) {
	t.Helper()

	portListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := portListener.Addr().(*net.TCPAddr).Port
	_ = portListener.Close()

	clientID := "client-id"
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	repo := &fakeBlingOAuthRepository{
		connection: database.IntegrationConnection{ClientID: &clientID, RedirectUri: &redirectURI},
	}
	service := newBlingOAuthService(repo, store, httpDoer)
	result, err := service.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = service.Disconnect(context.Background()) })
	return service, repo, result, redirectURI
}

func sendOAuthFailureHarnessCallback(t *testing.T, redirectURI, state, code string) (int, error) {
	t.Helper()
	query := url.Values{"code": {code}, "state": {state}}
	response, err := http.Get(redirectURI + "?" + query.Encode())
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	return response.StatusCode, nil
}

func waitForOAuthFailureHarnessStatus(t *testing.T, service *BlingOAuthService, sessionID string) BlingOAuthStatusResult {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		status, err := service.Status(context.Background(), sessionID)
		if err != nil {
			t.Fatal(err)
		}
		if status.Status != blingOAuthStatusAuthorizing {
			return status
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("OAuth session did not finish")
	return BlingOAuthStatusResult{}
}

func newOAuthFailureHarnessStore(t *testing.T) *memorySecretStore {
	t.Helper()
	store := &memorySecretStore{values: map[string][]byte{}}
	initial, err := json.Marshal(blingSecretBundle{ClientSecret: "client-secret-private"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), BlingCredentialSecretRef, initial); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestBlingOAuthCallbackDuplicateExchangesCodeOnlyOnce(t *testing.T) {
	var tokenRequests atomic.Int32
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		body := oauthFailureProbePayload
		if strings.Contains(req.URL.Path, "/oauth/token") {
			tokenRequests.Add(1)
			body = `{"access_token":"access-token-private","refresh_token":"refresh-token-private","token_type":"Bearer","scope":"read_orders","expires_in":21600}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})

	store := newOAuthFailureHarnessStore(t)
	service, repo, start, redirectURI := newOAuthFailureHarness(t, store, httpDoer)
	parsed, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	state := parsed.Query().Get("state")

	statusCode, err := sendOAuthFailureHarnessCallback(t, redirectURI, state, "one-time-code")
	if err != nil || statusCode != http.StatusOK {
		t.Fatalf("first callback status = %d, err = %v", statusCode, err)
	}
	// A duplicate can be rejected after shutdown or receive an acknowledgement
	// if it races with server shutdown. Either way, it must not trigger a second exchange.
	_, _ = sendOAuthFailureHarnessCallback(t, redirectURI, state, "one-time-code")

	status := waitForOAuthFailureHarnessStatus(t, service, start.SessionID)
	if status.Status != blingOAuthStatusConnected || !repo.completed {
		t.Fatalf("OAuth status = %+v, repo = %+v", status, repo)
	}
	if got := tokenRequests.Load(); got != 1 {
		t.Fatalf("token exchange count = %d, want 1", got)
	}
}

func TestBlingOAuthTokenExchangeFailureIsSanitizedAndNotPersisted(t *testing.T) {
	var tokenRequests, apiRequests atomic.Int32
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Path, "/oauth/token") {
			tokenRequests.Add(1)
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body: io.NopCloser(strings.NewReader(`{"error":"invalid_grant","client_secret":"client-secret-private","access_token":"access-token-private"}`)),
				Header: make(http.Header),
			}, nil
		}
		apiRequests.Add(1)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(oauthFailureProbePayload)), Header: make(http.Header)}, nil
	})

	store := newOAuthFailureHarnessStore(t)
	service, repo, start, redirectURI := newOAuthFailureHarness(t, store, httpDoer)
	parsed, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	statusCode, err := sendOAuthFailureHarnessCallback(t, redirectURI, parsed.Query().Get("state"), "one-time-code")
	if err != nil || statusCode != http.StatusOK {
		t.Fatalf("callback status = %d, err = %v", statusCode, err)
	}

	status := waitForOAuthFailureHarnessStatus(t, service, start.SessionID)
	if status.Status != blingOAuthStatusError || status.ErrorCode != "BLING_OAUTH_REJECTED" || repo.completed {
		t.Fatalf("OAuth status = %+v, repo = %+v", status, repo)
	}
	if strings.Contains(status.Message, "access-token-private") || strings.Contains(status.Message, "client-secret-private") {
		t.Fatalf("OAuth status leaked sensitive data: %+v", status)
	}
	if tokenRequests.Load() != 1 || apiRequests.Load() != 0 {
		t.Fatalf("token requests = %d, API requests = %d", tokenRequests.Load(), apiRequests.Load())
	}
	var stored blingSecretBundle
	if err := json.Unmarshal(mustGetSecret(t, store), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.AccessToken != "" || stored.RefreshToken != "" {
		t.Fatalf("failed exchange persisted token material: %+v", stored)
	}
}

func TestBlingOAuthDPAPIWriteFailureNeverMarksConnected(t *testing.T) {
	var tokenRequests atomic.Int32
	httpDoer := fakeOAuthHTTPDoer(func(req *http.Request) (*http.Response, error) {
		body := oauthFailureProbePayload
		if strings.Contains(req.URL.Path, "/oauth/token") {
			tokenRequests.Add(1)
			body = `{"access_token":"access-token-private","refresh_token":"refresh-token-private","token_type":"Bearer","scope":"read_orders","expires_in":21600}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})

	base := newOAuthFailureHarnessStore(t)
	store := &failingPutSecretStore{base: base, err: fmt.Errorf("DPAPI protection failed: access-token-private")}
	service, repo, start, redirectURI := newOAuthFailureHarness(t, store, httpDoer)
	parsed, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	statusCode, err := sendOAuthFailureHarnessCallback(t, redirectURI, parsed.Query().Get("state"), "one-time-code")
	if err != nil || statusCode != http.StatusOK {
		t.Fatalf("callback status = %d, err = %v", statusCode, err)
	}

	status := waitForOAuthFailureHarnessStatus(t, service, start.SessionID)
	if status.Status != blingOAuthStatusError || status.ErrorCode != "BLING_TOKEN_STORE_FAILED" || repo.completed {
		t.Fatalf("OAuth status = %+v, repo = %+v", status, repo)
	}
	if strings.Contains(status.Message, "access-token-private") || tokenRequests.Load() != 1 {
		t.Fatalf("OAuth failure was not sanitized: status=%+v token_requests=%d", status, tokenRequests.Load())
	}
	var stored blingSecretBundle
	if err := json.Unmarshal(mustGetSecret(t, base), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.AccessToken != "" || stored.RefreshToken != "" {
		t.Fatalf("failed protected persistence exposed tokens: %+v", stored)
	}
}
