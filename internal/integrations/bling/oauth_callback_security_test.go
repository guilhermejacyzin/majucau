package bling

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type oauthCallbackOutcome struct {
	code string
	err  error
}

func startOAuthCallbackTest(t *testing.T, state string) (func(url.Values) (int, error), <-chan oauthCallbackOutcome) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	session := &blingOAuthSession{state: state, path: "/callback", listener: listener, expiresAt: time.Now().Add(time.Minute)}
	service := &BlingOAuthService{now: time.Now}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	outcomes := make(chan oauthCallbackOutcome, 1)
	go func() {
		code, err := service.awaitCallback(ctx, session)
		outcomes <- oauthCallbackOutcome{code: code, err: err}
	}()

	request := func(query url.Values) (int, error) {
		endpoint := fmt.Sprintf("http://%s/callback?%s", listener.Addr(), query.Encode())
		response, err := http.Get(endpoint)
		if err != nil {
			return 0, err
		}
		defer response.Body.Close()
		_, _ = io.Copy(io.Discard, response.Body)
		return response.StatusCode, nil
	}
	return request, outcomes
}

func TestBlingOAuthCallbackRejectsInvalidStateAndAcceptsValidState(t *testing.T) {
	request, outcomes := startOAuthCallbackTest(t, "expected-state")

	status, err := request(url.Values{"code": {"untrusted-code"}, "state": {"wrong-state"}})
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusBadRequest {
		t.Fatalf("invalid state status = %d, want %d", status, http.StatusBadRequest)
	}

	status, err = request(url.Values{"code": {"authorization-code"}, "state": {"expected-state"}})
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusOK {
		t.Fatalf("valid callback status = %d, want %d", status, http.StatusOK)
	}

	select {
	case outcome := <-outcomes:
		if outcome.err != nil || outcome.code != "authorization-code" {
			t.Fatalf("callback outcome = %+v", outcome)
		}
	case <-time.After(time.Second):
		t.Fatal("valid callback was not received")
	}
}

func TestBlingOAuthCallbackRejectsProviderErrorAndMissingCode(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
	}{
		{name: "provider error", query: url.Values{"error": {"access_denied"}, "state": {"expected-state"}}},
		{name: "missing code", query: url.Values{"state": {"expected-state"}}},
		{name: "blank code", query: url.Values{"code": {"  "}, "state": {"expected-state"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, outcomes := startOAuthCallbackTest(t, "expected-state")
			status, err := request(test.query)
			if err != nil {
				t.Fatal(err)
			}
			if status != http.StatusOK {
				t.Fatalf("callback status = %d, want %d", status, http.StatusOK)
			}
			select {
			case outcome := <-outcomes:
				if !errors.Is(outcome.err, ErrBlingOAuthRejected) || outcome.code != "" {
					t.Fatalf("callback outcome = %+v, want rejected without code", outcome)
				}
			case <-time.After(time.Second):
				t.Fatal("rejected callback was not received")
			}
		})
	}
}

func TestBlingOAuthStatusExpiresSession(t *testing.T) {
	now := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	session := &blingOAuthSession{
		id:        "session-id",
		expiresAt: now.Add(-time.Second),
		status:    blingOAuthStatusAuthorizing,
	}
	service := &BlingOAuthService{
		now:      func() time.Time { return now },
		sessions: map[string]*blingOAuthSession{session.id: session},
	}

	status, err := service.Status(context.Background(), session.id)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != blingOAuthStatusCancelled || status.ErrorCode != "BLING_OAUTH_EXPIRED" {
		t.Fatalf("expired session status = %+v", status)
	}
}
