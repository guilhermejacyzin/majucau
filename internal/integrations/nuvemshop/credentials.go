package nuvemshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/security"
)

const NuvemshopCredentialSecretRef = "nuvemshop_credentials_v1"

var (
	ErrNuvemshopCredentialValidation = errors.New("nuvemshop credential configuration is invalid")
	ErrNuvemshopCredentialMissing    = errors.New("nuvemshop client secret is not configured")
	ErrNuvemshopCredentialVault      = errors.New("nuvemshop credential vault is unavailable")
)

type ConfigInput struct {
	AppID        string `json:"app_id"`
	RedirectURI  string `json:"redirect_uri"`
	ClientSecret string `json:"client_secret"`
}

type ConfigResult struct {
	AppID            string `json:"app_id"`
	RedirectURI      string `json:"redirect_uri"`
	SecretConfigured bool   `json:"secret_configured"`
	Status           string `json:"status"`
}

type secretBundle struct {
	ClientSecret string `json:"client_secret"`
}

type CredentialService struct {
	pool  *pgxpool.Pool
	store security.SecretStore
}

func NewCredentialService(pool *pgxpool.Pool, store security.SecretStore) *CredentialService {
	return &CredentialService{pool: pool, store: store}
}

// Save stores only the Nuvemshop secret in the worker vault and public
// application metadata in PostgreSQL. OAuth is intentionally not started here.
func (s *CredentialService) Save(ctx context.Context, input ConfigInput) (ConfigResult, error) {
	if s == nil || s.pool == nil || s.store == nil {
		return ConfigResult{}, ErrNuvemshopCredentialVault
	}
	input, err := normalizeConfig(input)
	if err != nil {
		return ConfigResult{}, err
	}
	previous, previousErr := s.store.Get(ctx, NuvemshopCredentialSecretRef)
	if previousErr != nil && !errors.Is(previousErr, fs.ErrNotExist) {
		return ConfigResult{}, fmt.Errorf("%w: read existing secret", ErrNuvemshopCredentialVault)
	}
	secret := strings.TrimSpace(input.ClientSecret)
	if secret == "" {
		secret, err = secretFromBundle(previous)
		if err != nil {
			return ConfigResult{}, ErrNuvemshopCredentialMissing
		}
	}
	bundle, err := json.Marshal(secretBundle{ClientSecret: secret})
	if err != nil {
		return ConfigResult{}, fmt.Errorf("%w: encode secret", ErrNuvemshopCredentialVault)
	}
	if err := s.store.Put(ctx, NuvemshopCredentialSecretRef, bundle); err != nil {
		return ConfigResult{}, fmt.Errorf("%w: write secret", ErrNuvemshopCredentialVault)
	}
	queries := database.New(s.pool)
	appID, redirectURI, secretRef := input.AppID, input.RedirectURI, NuvemshopCredentialSecretRef
	if err := queries.UpsertNuvemshopConnectionConfig(ctx, database.UpsertNuvemshopConnectionConfigParams{ClientID: &appID, RedirectUri: &redirectURI, SecretRef: &secretRef}); err != nil {
		restoreSecret(ctx, s.store, previous, previousErr)
		return ConfigResult{}, fmt.Errorf("save Nuvemshop public configuration: %w", err)
	}
	return ConfigResult{AppID: input.AppID, RedirectURI: input.RedirectURI, SecretConfigured: true, Status: "NOT_CONFIGURED"}, nil
}

func normalizeConfig(input ConfigInput) (ConfigInput, error) {
	input.AppID = strings.TrimSpace(input.AppID)
	input.RedirectURI = strings.TrimSpace(input.RedirectURI)
	if input.AppID == "" || len(input.AppID) > 256 {
		return ConfigInput{}, fmt.Errorf("%w: app id", ErrNuvemshopCredentialValidation)
	}
	parsed, err := url.Parse(input.RedirectURI)
	host := strings.TrimSpace(strings.ToLower(parsed.Hostname()))
	if err != nil || parsed.Scheme != "https" || host == "" || host == "localhost" {
		return ConfigInput{}, fmt.Errorf("%w: redirect uri must use HTTPS relay", ErrNuvemshopCredentialValidation)
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return ConfigInput{}, fmt.Errorf("%w: redirect uri must not use loopback", ErrNuvemshopCredentialValidation)
	}
	if len(input.ClientSecret) > 2048 {
		return ConfigInput{}, fmt.Errorf("%w: client secret", ErrNuvemshopCredentialValidation)
	}
	return input, nil
}

func secretFromBundle(value []byte) (string, error) {
	if len(value) == 0 {
		return "", ErrNuvemshopCredentialMissing
	}
	var bundle secretBundle
	if err := json.Unmarshal(value, &bundle); err != nil || strings.TrimSpace(bundle.ClientSecret) == "" {
		return "", ErrNuvemshopCredentialMissing
	}
	return bundle.ClientSecret, nil
}

func restoreSecret(ctx context.Context, store security.SecretStore, previous []byte, previousErr error) {
	if len(previous) > 0 && previousErr == nil {
		_ = store.Put(ctx, NuvemshopCredentialSecretRef, previous)
		return
	}
	_ = store.Delete(ctx, NuvemshopCredentialSecretRef)
}
