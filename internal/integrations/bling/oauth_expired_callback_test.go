package bling

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

func TestBlingOAuthExpiredCallbackDoesNotExchangeOrPersistTokens(t *testing.T) {
	var requests atomic.Int32
	httpDoer := fakeOAuthHTTPDoer(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return nil, errors.New("unexpected HTTP request after expired callback")
	})
	store := newOAuthFailureHarnessStore(t)
	service, repo, start, redirectURI := newOAuthFailureHarness(t, store, httpDoer)

	parsed, err := url.Parse(start.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	service.mu.Lock()
	service.sessions[start.SessionID].expiresAt = time.Now().Add(-time.Second)
	service.mu.Unlock()

	statusCode, err := sendOAuthFailureHarnessCallback(t, redirectURI, parsed.Query().Get("state"), "expired-code")
	if err != nil {
		t.Fatal(err)
	}
	if statusCode != http.StatusGone {
		t.Fatalf("expired callback status = %d, want %d", statusCode, http.StatusGone)
	}

	deadline := time.Now().Add(3 * time.Second)
	var status BlingOAuthStatusResult
	for time.Now().Before(deadline) {
		status, err = service.Status(context.Background(), start.SessionID)
		if err != nil {
			t.Fatal(err)
		}
		if status.Status == blingOAuthStatusError {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status.Status != blingOAuthStatusError || status.ErrorCode != "BLING_OAUTH_EXPIRED" {
		t.Fatalf("expired OAuth status = %+v, want AUTH_ERROR/BLING_OAUTH_EXPIRED", status)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("HTTP requests after expired callback = %d, want 0", got)
	}
	if repo.completed {
		t.Fatal("expired OAuth callback marked the connection complete")
	}

	var stored blingSecretBundle
	if err := json.Unmarshal(mustGetSecret(t, store), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.AccessToken != "" || stored.RefreshToken != "" {
		t.Fatal("expired OAuth callback persisted token material")
	}
}
