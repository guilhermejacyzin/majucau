package googledrive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"majucau.local/financial-intelligence/internal/security"
)

const (
	ServiceAccountSecretRef    = "google-drive-service-account"
	driveReadOnlyScope         = "https://www.googleapis.com/auth/drive.readonly"
	googleTokenURL             = "https://oauth2.googleapis.com/token"
	MaxServiceAccountJSONBytes = 1 << 20
	maxServiceAccountJSON      = MaxServiceAccountJSONBytes
	refreshTokenEarlyBy        = time.Minute
)

// ServiceAccountTokenProvider reads its service-account key from the protected
// secret store and exchanges short-lived JWT assertions for read-only tokens.
// In the Windows prototype, the production store must be the worker's DPAPI
// store. No credential file or key text is accepted by this type.
type ServiceAccountTokenProvider struct {
	config *jwt.Config
	mu     sync.Mutex
	token  *oauth2.Token
	client *http.Client
}

// NewServiceAccountTokenProvider loads a service account key from the fixed
// protected-store reference. The key must describe a Google service account
// whose Drive access is separately limited to the three approved folders.
func NewServiceAccountTokenProvider(ctx context.Context, store security.SecretStore) (*ServiceAccountTokenProvider, error) {
	return newServiceAccountTokenProvider(ctx, store, googleTokenURL)
}

func newServiceAccountTokenProvider(ctx context.Context, store security.SecretStore, tokenURL string) (*ServiceAccountTokenProvider, error) {
	if ctx == nil || store == nil || ctx.Err() != nil || !validTokenURL(tokenURL) {
		return nil, ErrConfiguration
	}
	if err := security.ValidateSecretRef(ServiceAccountSecretRef); err != nil {
		return nil, ErrConfiguration
	}
	keyJSON, err := store.Get(ctx, ServiceAccountSecretRef)
	if err != nil {
		return nil, ErrAuthorization
	}
	defer wipeBytes(keyJSON)
	config, err := serviceAccountJWTConfig(keyJSON)
	if err != nil {
		return nil, ErrAuthorization
	}
	config.TokenURL = tokenURL
	parsedTokenURL, _ := url.Parse(tokenURL)
	client := &http.Client{Transport: tokenEndpointTransport{base: http.DefaultTransport, allowed: parsedTokenURL}}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &ServiceAccountTokenProvider{config: config, client: client}, nil
}

func (p *ServiceAccountTokenProvider) AccessToken(ctx context.Context) (string, error) {
	if p == nil || p.config == nil || p.client == nil || ctx == nil {
		return "", ErrAuthorization
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.token != nil && p.token.Valid() && p.token.Expiry.After(time.Now().Add(refreshTokenEarlyBy)) && strings.EqualFold(p.token.TokenType, "Bearer") {
		return p.token.AccessToken, nil
	}
	tokenContext, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	requestContext := context.WithValue(tokenContext, oauth2.HTTPClient, p.client)
	token, err := p.config.TokenSource(requestContext).Token()
	if err != nil || token == nil || !token.Valid() || !strings.EqualFold(token.TokenType, "Bearer") || !validToken(token.AccessToken) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", ErrAuthorization
	}
	p.token = token
	return token.AccessToken, nil
}

func (*ServiceAccountTokenProvider) String() string {
	return "Google Drive service-account token provider"
}

func (*ServiceAccountTokenProvider) GoString() string {
	return "Google Drive service-account token provider"
}

type serviceAccountIdentity struct {
	Type        string `json:"type"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

func serviceAccountJWTConfig(keyJSON []byte) (*jwt.Config, error) {
	var credentialBundle serviceAccountCredentialBundle
	bundleErr := json.Unmarshal(keyJSON, &credentialBundle)
	defer wipeBytes(credentialBundle.ServiceAccountKey)
	if bundleErr == nil && credentialBundle.Format != "" {
		if credentialBundle.Format != serviceAccountCredentialBundleFormat || ValidateDriveFolderIDs(credentialBundle.FolderIDs) != nil || len(credentialBundle.ServiceAccountKey) == 0 {
			return nil, ErrAuthorization
		}
		keyJSON = credentialBundle.ServiceAccountKey
	}
	if len(keyJSON) == 0 || len(keyJSON) > maxServiceAccountJSON {
		return nil, ErrAuthorization
	}
	var identity serviceAccountIdentity
	if err := json.Unmarshal(keyJSON, &identity); err != nil || identity.Type != "service_account" || len(identity.ClientEmail) == 0 || len(identity.ClientEmail) > 320 || !strings.HasSuffix(strings.ToLower(identity.ClientEmail), ".iam.gserviceaccount.com") || identity.TokenURI != googleTokenURL {
		return nil, ErrAuthorization
	}
	config, err := google.JWTConfigFromJSON(keyJSON, driveReadOnlyScope)
	if err != nil || config == nil || config.Email != identity.ClientEmail || config.TokenURL != googleTokenURL || len(config.PrivateKey) == 0 {
		if config != nil {
			wipeBytes(config.PrivateKey)
		}
		return nil, ErrAuthorization
	}
	return config, nil
}

func validTokenURL(value string) bool {
	if value == googleTokenURL {
		return true
	}
	// Tests may use a loopback HTTP token endpoint. Non-loopback endpoints must
	// always use Google's official HTTPS token service.
	parsed, err := url.Parse(value)
	return err == nil && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.User == nil && parsed.Scheme == "http" && safeBaseURLScheme(parsed)
}

type tokenEndpointTransport struct {
	base    http.RoundTripper
	allowed *url.URL
}

func (t tokenEndpointTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if t.base == nil || t.allowed == nil || request == nil || request.URL == nil || request.Method != http.MethodPost || request.URL.String() != t.allowed.String() {
		return nil, ErrAuthorization
	}
	return t.base.RoundTrip(request)
}

func wipeBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

var _ AccessTokenProvider = (*ServiceAccountTokenProvider)(nil)
