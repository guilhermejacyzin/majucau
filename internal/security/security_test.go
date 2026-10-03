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

func TestRedactJSONBoundsInputAndFailsClosed(t *testing.T) {
	input := `{"access_token":"private","padding":"` + strings.Repeat("x", maxRedactJSONBytes) + `"}`
	if len(input) <= maxRedactJSONBytes {
		t.Fatal("test input must exceed the redaction limit")
	}
	if got := string(RedactJSON([]byte(input))); got != `"[REDACTED]"` {
		t.Fatalf("oversized JSON was not fully redacted: got %q", got)
	}
}
