package security

import (
	"context"
	"strings"
	"testing"
)

func TestRedactionAndFakeStore(t *testing.T) {
	v := RedactJSON([]byte(`{"access_token":"abcdef1234","name":"ok","nested":{"password":"x"}}`))
	if string(v) == "" || string(v) == `{"access_token":"abcdef1234","name":"ok","nested":{"password":"x"}}` {
		t.Fatalf("secret was not redacted: %s", v)
	}
	s := NewFakeSecretStore()
	if err := s.Put(context.Background(), "bling_token", []byte("secret")); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(context.Background(), "bling_token")
	if err != nil || string(got) != "secret" {
		t.Fatalf("fake store: %q %v", got, err)
	}
	got[0] = 'X'
	again, _ := s.Get(context.Background(), "bling_token")
	if string(again) != "secret" {
		t.Fatal("fake store leaked mutable buffer")
	}
}

func TestRedactJSONRecognizesSensitiveKeysAcrossCommonFormats(t *testing.T) {
	input := []byte(`{"accessToken":"access-token-value","clientSecret":"client-secret-value","refresh-token":"refresh-token-value","API-Key":"api-key-value","Authorization":"Bearer authorization-value","nested":{"setCookie":"session-cookie-value","credentials":"stored-credentials"},"display_name":"Majucau"}`)
	redacted := string(RedactJSON(input))
	for _, secret := range []string{
		"access-token-value", "client-secret-value", "refresh-token-value", "api-key-value",
		"Bearer authorization-value", "session-cookie-value", "stored-credentials",
	} {
		if strings.Contains(redacted, secret) {
			t.Errorf("redacted JSON still contains sensitive value %q: %s", secret, redacted)
		}
	}
	if !strings.Contains(redacted, "display_name") || !strings.Contains(redacted, "Majucau") {
		t.Fatalf("non-sensitive fields were not preserved: %s", redacted)
	}
}

func TestRedactJSONBoundsInputAndFailsClosed(t *testing.T) {
	input := `{"access_token":"private","padding":"` + strings.Repeat("x", maxRedactJSONBytes) + `"}`
	if len(input) <= maxRedactJSONBytes {
		t.Fatal("test input must exceed the redaction limit")
	}
	if got := string(RedactJSON([]byte(input))); got != `"[REDACTED]"` {
		t.Fatalf("oversized JSON was not fully redacted: got %q", got)
	}
}
