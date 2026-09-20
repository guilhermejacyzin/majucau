package bling

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"majucau.local/financial-intelligence/internal/security"
)

type memorySecretStore struct{ values map[string][]byte }

func (s *memorySecretStore) Put(_ context.Context, ref string, value []byte) error {
	if err := security.ValidateSecretRef(ref); err != nil {
		return err
	}
	if s.values == nil {
		s.values = map[string][]byte{}
	}
	s.values[ref] = append([]byte(nil), value...)
	return nil
}
func (s *memorySecretStore) Get(_ context.Context, ref string) ([]byte, error) {
	if err := security.ValidateSecretRef(ref); err != nil {
		return nil, err
	}
	return append([]byte(nil), s.values[ref]...), nil
}
func (s *memorySecretStore) Delete(_ context.Context, ref string) error {
	if err := security.ValidateSecretRef(ref); err != nil {
		return err
	}
	delete(s.values, ref)
	return nil
}

func TestNormalizeBlingConfigAcceptsHTTPSAndLoopbackRedirects(t *testing.T) {
	for _, redirect := range []string{"https://app.example.test/oauth/callback", "http://127.0.0.1:43821/callback"} {
		got, err := normalizeBlingConfig(BlingConfigInput{ClientID: "client", RedirectURI: redirect, ClientSecret: "secret"})
		if err != nil || got.RedirectURI != redirect {
			t.Fatalf("redirect %q: %#v, %v", redirect, got, err)
		}
	}
}

func TestNormalizeBlingConfigRejectsInsecureRedirect(t *testing.T) {
	_, err := normalizeBlingConfig(BlingConfigInput{ClientID: "client", RedirectURI: "http://public.example.test/callback", ClientSecret: "secret"})
	if !errors.Is(err, ErrBlingCredentialValidation) {
		t.Fatalf("error = %v", err)
	}
}

func TestBlingSecretBundleKeepsSecretOutOfPublicResult(t *testing.T) {
	store := &memorySecretStore{}
	if err := store.Put(context.Background(), BlingCredentialSecretRef, []byte(`{"client_secret":"secret-only-in-vault"}`)); err != nil {
		t.Fatal(err)
	}
	secret, err := clientSecretFromBundle(mustGetSecret(t, store))
	if err != nil || secret != "secret-only-in-vault" {
		t.Fatalf("secret = %q, err = %v", secret, err)
	}
	result := BlingConfigResult{ClientID: "client", RedirectURI: "https://app.example.test/callback", SecretConfigured: true, Status: "NOT_CONFIGURED"}
	if strings.Contains(fmt.Sprintf("%+v", result), secret) {
		t.Fatal("public result leaked client secret")
	}
}

func mustGetSecret(t *testing.T, store security.SecretStore) []byte {
	t.Helper()
	value, err := store.Get(context.Background(), BlingCredentialSecretRef)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
