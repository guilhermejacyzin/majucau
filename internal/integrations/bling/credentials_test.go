package bling

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"testing"

	"majucau.local/financial-intelligence/internal/security"
)

type memorySecretStore struct {
	mu     sync.Mutex
	values map[string][]byte
}

func (s *memorySecretStore) Put(_ context.Context, ref string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := security.ValidateSecretRef(ref); err != nil {
		return nil, err
	}
	value, ok := s.values[ref]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return append([]byte(nil), value...), nil
}

type fakeBlingConfigRepository struct {
	clientID    string
	redirectURI string
	secretRef   *string
	calls       int
	err         error
}

func (r *fakeBlingConfigRepository) UpsertBlingConnectionConfig(_ context.Context, clientID, redirectURI string, secretRef *string) error {
	r.calls++
	r.clientID = clientID
	r.redirectURI = redirectURI
	r.secretRef = nil
	if secretRef != nil {
		copy := *secretRef
		r.secretRef = &copy
	}
	return r.err
}
func (s *memorySecretStore) Delete(_ context.Context, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func TestSaveBlingConfigPersistsPublicMetadataWithoutMissingSecret(t *testing.T) {
	store := &memorySecretStore{}
	repository := &fakeBlingConfigRepository{}
	service := &BlingCredentialService{repository: repository, credentials: NewBlingCredentialState(store)}

	result, err := service.Save(context.Background(), BlingConfigInput{
		ClientID: "public-client-id", RedirectURI: "https://app.example.test/callback",
	})
	if err != nil {
		t.Fatalf("save public configuration: %v", err)
	}
	if result.ClientID != "public-client-id" || result.RedirectURI != "https://app.example.test/callback" || result.SecretConfigured {
		t.Fatalf("unexpected public result: %#v", result)
	}
	if repository.calls != 1 || repository.clientID != result.ClientID || repository.redirectURI != result.RedirectURI || repository.secretRef != nil {
		t.Fatalf("public configuration was not saved without a vault reference: %#v", repository)
	}
	if _, err := store.Get(context.Background(), BlingCredentialSecretRef); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing secret must not be fabricated or stored, got err %v", err)
	}
}

func TestSaveBlingConfigKeepsExistingSecretWhenDatabaseFails(t *testing.T) {
	store := &memorySecretStore{}
	if err := store.Put(context.Background(), BlingCredentialSecretRef, []byte(`{"client_secret":"previous-secret"}`)); err != nil {
		t.Fatal(err)
	}
	repository := &fakeBlingConfigRepository{err: errors.New("database unavailable")}
	service := &BlingCredentialService{repository: repository, credentials: NewBlingCredentialState(store)}
	_, err := service.Save(context.Background(), BlingConfigInput{
		ClientID: "public-client-id", RedirectURI: "https://app.example.test/callback", ClientSecret: "new-secret",
	})
	if err == nil {
		t.Fatal("expected database error")
	}
	secret, secretErr := clientSecretFromBundle(mustGetSecret(t, store))
	if secretErr != nil || secret != "previous-secret" {
		t.Fatalf("previous secret was not restored: secret=%q err=%v", secret, secretErr)
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
