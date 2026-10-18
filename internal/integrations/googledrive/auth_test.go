package googledrive

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"majucau.local/financial-intelligence/internal/security"
)

func TestServiceAccountTokenProviderUsesReadOnlyScopeAndCachesToken(t *testing.T) {
	privateKey, raw := serviceAccountFixture(t)
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, raw); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		if request.Method != http.MethodPost || request.URL.Path != "/" {
			t.Errorf("token request = %s %s", request.Method, request.URL.Path)
		}
		if err := request.ParseForm(); err != nil {
			t.Errorf("parse token form: %v", err)
		}
		if request.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("grant_type = %q", request.Form.Get("grant_type"))
		}
		verifyJWTAssertion(t, request.Form.Get("assertion"), privateKey, server.URL)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"access_token":"short-lived-access-token","token_type":"Bearer","expires_in":3600}`)
	}))
	defer server.Close()
	provider, err := newServiceAccountTokenProvider(context.Background(), store, server.URL)
	if err != nil {
		t.Fatalf("newServiceAccountTokenProvider() error = %v", err)
	}
	for range 2 {
		token, tokenErr := provider.AccessToken(context.Background())
		if tokenErr != nil || token != "short-lived-access-token" {
			t.Fatalf("AccessToken() = %q, %v", token, tokenErr)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("token exchange calls = %d, want cached single exchange", calls.Load())
	}
	if strings.Contains(fmt.Sprint(provider), "short-lived-access-token") {
		t.Fatal("provider string representation exposed the access token")
	}
}

func TestServiceAccountTokenProviderRejectsInvalidIdentityAndTokenEndpoint(t *testing.T) {
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, []byte(`{"type":"authorized_user"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := newServiceAccountTokenProvider(context.Background(), store, "https://attacker.example/token"); !errors.Is(err, ErrConfiguration) {
		t.Fatalf("unsafe token endpoint error = %v", err)
	}
	if _, err := NewServiceAccountTokenProvider(context.Background(), store); !errors.Is(err, ErrAuthorization) {
		t.Fatalf("invalid credential error = %v", err)
	}
}

func TestServiceAccountTokenProviderDoesNotFollowTokenEndpointRedirect(t *testing.T) {
	privateKey, raw := serviceAccountFixture(t)
	_ = privateKey
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, raw); err != nil {
		t.Fatal(err)
	}
	var redirected atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected.Add(1) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	provider, err := newServiceAccountTokenProvider(context.Background(), store, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.AccessToken(context.Background()); !errors.Is(err, ErrAuthorization) {
		t.Fatalf("AccessToken() error = %v, want sanitized authorization failure", err)
	}
	if redirected.Load() != 0 {
		t.Fatalf("redirect target received %d requests", redirected.Load())
	}
}

func TestServiceAccountTokenProviderHonorsCallerCancellation(t *testing.T) {
	privateKey, raw := serviceAccountFixture(t)
	_ = privateKey
	store := newMemorySecretStore()
	if err := store.Put(context.Background(), ServiceAccountSecretRef, raw); err != nil {
		t.Fatal(err)
	}
	provider, err := newServiceAccountTokenProvider(context.Background(), store, "http://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.AccessToken(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("AccessToken() error = %v, want caller cancellation", err)
	}
}

func serviceAccountFixture(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("marshal test key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	credential := struct {
		Type         string `json:"type"`
		ProjectID    string `json:"project_id"`
		PrivateKeyID string `json:"private_key_id"`
		PrivateKey   string `json:"private_key"`
		ClientEmail  string `json:"client_email"`
		ClientID     string `json:"client_id"`
		TokenURI     string `json:"token_uri"`
	}{
		Type: "service_account", ProjectID: "majucau-test", PrivateKeyID: "test-key-id",
		PrivateKey: string(keyPEM), ClientEmail: "majucau-test@majucau-test.iam.gserviceaccount.com",
		ClientID: "1234567890", TokenURI: googleTokenURL,
	}
	raw, err := json.Marshal(credential)
	if err != nil {
		t.Fatalf("marshal test credential: %v", err)
	}
	return privateKey, raw
}

type memorySecretStore struct {
	mu     sync.Mutex
	values map[string][]byte
}

func newMemorySecretStore() *memorySecretStore {
	return &memorySecretStore{values: make(map[string][]byte)}
}

func (s *memorySecretStore) Put(_ context.Context, ref string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[ref] = append([]byte(nil), value...)
	return nil
}

func (s *memorySecretStore) Get(_ context.Context, ref string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.values[ref]
	if !ok {
		return nil, errors.New("missing secret")
	}
	return append([]byte(nil), value...), nil
}

func (s *memorySecretStore) Delete(_ context.Context, ref string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, ref)
	return nil
}

var _ security.SecretStore = (*memorySecretStore)(nil)

func verifyJWTAssertion(t *testing.T, assertion string, privateKey *rsa.PrivateKey, expectedAudience string) {
	t.Helper()
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		t.Fatalf("JWT has %d parts, want 3", len(parts))
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode JWT header: %v", err)
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "RS256" {
		t.Fatalf("JWT header algorithm = %q, err=%v", header.Algorithm, err)
	}
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode JWT claims: %v", err)
	}
	var claims struct {
		Issuer   string `json:"iss"`
		Scope    string `json:"scope"`
		Audience string `json:"aud"`
	}
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		t.Fatalf("unmarshal JWT claims: %v", err)
	}
	if claims.Issuer != "majucau-test@majucau-test.iam.gserviceaccount.com" || claims.Scope != driveReadOnlyScope || claims.Audience != expectedAudience {
		t.Fatalf("unexpected JWT claims: %#v", claims)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode JWT signature: %v", err)
	}
	message := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(&privateKey.PublicKey, crypto.SHA256, digest[:], signature); err != nil {
		t.Fatalf("verify JWT signature: %v", err)
	}
}
