package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/security"
)

const BlingCredentialSecretRef = "bling_credentials_v1"

var (
	ErrBlingCredentialValidation = errors.New("bling credential configuration is invalid")
	ErrBlingCredentialMissing    = errors.New("bling client secret is not configured")
	ErrBlingCredentialVault      = errors.New("bling credential vault is unavailable")
)

type BlingConfigInput struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ClientSecret string `json:"client_secret"`
}

type BlingConfigResult struct {
	ClientID         string `json:"client_id"`
	RedirectURI      string `json:"redirect_uri"`
	SecretConfigured bool   `json:"secret_configured"`
	Status           string `json:"status"`
}

type blingSecretBundle struct {
	ClientSecret string `json:"client_secret"`
}

type BlingCredentialService struct {
	pool  *pgxpool.Pool
	store security.SecretStore
}

func NewBlingCredentialService(pool *pgxpool.Pool, store security.SecretStore) *BlingCredentialService {
	return &BlingCredentialService{pool: pool, store: store}
}

// Save stores the secret only in the worker vault and public metadata in the
// integration table. An empty secret means “keep the existing secret”; it is
// never interpreted as a request to erase credentials.
func (s *BlingCredentialService) Save(ctx context.Context, input BlingConfigInput) (BlingConfigResult, error) {
	if s == nil || s.pool == nil {
		return BlingConfigResult{}, ErrBlingCredentialVault
	}
	if s.store == nil {
		return BlingConfigResult{}, ErrBlingCredentialVault
	}
	input, err := normalizeBlingConfig(input)
	if err != nil {
		return BlingConfigResult{}, err
	}
	previous, previousErr := s.store.Get(ctx, BlingCredentialSecretRef)
	if previousErr != nil && !errors.Is(previousErr, fs.ErrNotExist) {
		return BlingConfigResult{}, fmt.Errorf("%w: read existing secret", ErrBlingCredentialVault)
	}
	secret := input.ClientSecret
	if secret == "" {
		secret, err = clientSecretFromBundle(previous)
		if err != nil {
			return BlingConfigResult{}, ErrBlingCredentialMissing
		}
	}
	bundle, err := json.Marshal(blingSecretBundle{ClientSecret: secret})
	if err != nil {
		return BlingConfigResult{}, fmt.Errorf("%w: encode secret", ErrBlingCredentialVault)
	}
	if err := s.store.Put(ctx, BlingCredentialSecretRef, bundle); err != nil {
		return BlingConfigResult{}, fmt.Errorf("%w: write secret", ErrBlingCredentialVault)
	}
	queries := database.New(s.pool)
	clientID, redirectURI, secretRef := input.ClientID, input.RedirectURI, BlingCredentialSecretRef
	if err := queries.UpsertBlingConnectionConfig(ctx, database.UpsertBlingConnectionConfigParams{ClientID: &clientID, RedirectUri: &redirectURI, SecretRef: &secretRef}); err != nil {
		restoreSecret(ctx, s.store, previous, previousErr)
		return BlingConfigResult{}, fmt.Errorf("save Bling public configuration: %w", err)
	}
	return BlingConfigResult{ClientID: input.ClientID, RedirectURI: input.RedirectURI, SecretConfigured: true, Status: "NOT_CONFIGURED"}, nil
}

func normalizeBlingConfig(input BlingConfigInput) (BlingConfigInput, error) {
	input.ClientID = strings.TrimSpace(input.ClientID)
	input.RedirectURI = strings.TrimSpace(input.RedirectURI)
	if input.ClientID == "" || len(input.ClientID) > 256 {
		return BlingConfigInput{}, fmt.Errorf("%w: client id", ErrBlingCredentialValidation)
	}
	parsed, err := url.Parse(input.RedirectURI)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "https" && !isLoopbackHost(parsed.Hostname())) {
		return BlingConfigInput{}, fmt.Errorf("%w: redirect uri must use HTTPS or loopback HTTP", ErrBlingCredentialValidation)
	}
	if strings.TrimSpace(input.ClientSecret) == "" {
		input.ClientSecret = ""
	}
	if len(input.ClientSecret) > 2048 {
		return BlingConfigInput{}, fmt.Errorf("%w: client secret", ErrBlingCredentialValidation)
	}
	return input, nil
}

func clientSecretFromBundle(value []byte) (string, error) {
	if len(value) == 0 {
		return "", ErrBlingCredentialMissing
	}
	var bundle blingSecretBundle
	if err := json.Unmarshal(value, &bundle); err != nil || strings.TrimSpace(bundle.ClientSecret) == "" {
		return "", ErrBlingCredentialMissing
	}
	return bundle.ClientSecret, nil
}

func restoreSecret(ctx context.Context, store security.SecretStore, previous []byte, previousErr error) {
	if len(previous) > 0 && previousErr == nil {
		_ = store.Put(ctx, BlingCredentialSecretRef, previous)
		return
	}
	_ = store.Delete(ctx, BlingCredentialSecretRef)
}
