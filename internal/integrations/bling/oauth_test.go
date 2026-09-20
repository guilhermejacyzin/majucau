package bling

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlingOAuthExchangeUsesBasicAuthAndForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		id, secret, ok := r.BasicAuth()
		if !ok || id != "client-id" || secret != "client-secret" {
			t.Fatalf("basic auth = %q, %q, %v", id, secret, ok)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("content type = %q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "one-time-code" || r.Form.Get("redirect_uri") != "http://127.0.0.1/callback" {
			t.Fatalf("form = %#v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"access","refresh_token":"refresh","token_type":"Bearer","scope":"read","expires_in":21600}`))
	}))
	defer server.Close()

	client, err := NewBlingOAuthClient(server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	token, err := client.ExchangeAuthorizationCode(context.Background(), "client-id", "client-secret", "http://127.0.0.1/callback", "one-time-code")
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "access" || token.RefreshToken != "refresh" || token.ExpiresIn != 21600 {
		t.Fatalf("token = %+v", token)
	}
}

func TestBlingOAuthErrorsDoNotExposeSecretOrBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","client_secret":"do-not-leak"}`))
	}))
	defer server.Close()
	client, err := NewBlingOAuthClient(server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.RefreshAccessToken(context.Background(), "client-id", "client-secret", "refresh")
	if !errors.Is(err, ErrBlingOAuthRejected) {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "do-not-leak") || strings.Contains(err.Error(), "client-secret") {
		t.Fatalf("oauth error leaked sensitive data: %v", err)
	}
}

func TestBlingOAuthRejectsMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer server.Close()
	client, err := NewBlingOAuthClient(server.Client(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.RefreshAccessToken(context.Background(), "client-id", "client-secret", "refresh")
	if !errors.Is(err, ErrBlingOAuthSchema) {
		t.Fatalf("error = %v", err)
	}
}
