package bling

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/security"
)

const (
	blingOAuthSessionTTL        = 5 * time.Minute
	blingRefreshTokenLifetime   = 30 * 24 * time.Hour
	blingOAuthStatusAuthorizing = "AUTHORIZING"
	blingOAuthStatusConnected   = "CONNECTED"
	blingOAuthStatusError       = "AUTH_ERROR"
	blingOAuthStatusCancelled   = "CANCELLED"
)

var (
	ErrBlingOAuthConfiguration = errors.New("bling oauth configuration is invalid")
	ErrBlingOAuthSession       = errors.New("bling oauth session is invalid or expired")
	ErrBlingOAuthCallback      = errors.New("bling oauth callback is invalid")
	ErrBlingOAuthNotConnected  = errors.New("bling oauth connection is not authorized")
	ErrBlingOAuthStorage       = errors.New("bling oauth token storage failed")
)

type BlingOAuthStartResult struct {
	SessionID        string `json:"session_id"`
	AuthorizationURL string `json:"authorization_url"`
	Status           string `json:"status"`
}

type BlingOAuthStatusResult struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

type BlingOAuthTestResult struct {
	Status          string `json:"status"`
	PageRecordCount int    `json:"page_record_count"`
}

type blingOAuthRepository interface {
	Connection(context.Context) (database.IntegrationConnection, error)
	MarkAuthorizing(context.Context) error
	CompleteOAuth(context.Context, []string, *time.Time, *time.Time) error
	MarkOAuthError(context.Context, string) error
	RecordTestSuccess(context.Context) error
	RecordTestFailure(context.Context, string) error
}

type postgresBlingOAuthRepository struct {
	queries *database.Queries
}

func newPostgresBlingOAuthRepository(pool *pgxpool.Pool) *postgresBlingOAuthRepository {
	if pool == nil {
		return nil
	}
	return &postgresBlingOAuthRepository{queries: database.New(pool)}
}

func (r *postgresBlingOAuthRepository) Connection(ctx context.Context) (database.IntegrationConnection, error) {
	if r == nil || r.queries == nil {
		return database.IntegrationConnection{}, ErrBlingOAuthConfiguration
	}
	return r.queries.GetIntegrationConnectionByProvider(ctx, "BLING")
}

func (r *postgresBlingOAuthRepository) MarkAuthorizing(ctx context.Context) error {
	return r.queries.MarkBlingAuthorizing(ctx)
}

func (r *postgresBlingOAuthRepository) CompleteOAuth(ctx context.Context, scopes []string, accessExpiresAt, refreshExpiresAt *time.Time) error {
	return r.queries.CompleteBlingOAuth(ctx, database.CompleteBlingOAuthParams{
		AuthorizedScopes:      scopes,
		TokenExpiresAt:        nullableTimestamp(accessExpiresAt),
		RefreshTokenExpiresAt: nullableTimestamp(refreshExpiresAt),
	})
}

func (r *postgresBlingOAuthRepository) MarkOAuthError(ctx context.Context, code string) error {
	return r.queries.MarkBlingOAuthError(ctx, &code)
}

func (r *postgresBlingOAuthRepository) RecordTestSuccess(ctx context.Context) error {
	return r.queries.RecordBlingTestSuccess(ctx)
}

func (r *postgresBlingOAuthRepository) RecordTestFailure(ctx context.Context, code string) error {
	return r.queries.RecordBlingTestFailure(ctx, &code)
}

func nullableTimestamp(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

type blingOAuthSession struct {
	id        string
	state     string
	path      string
	listener  net.Listener
	expiresAt time.Time
	status    string
	errorCode string
	message   string
}

type BlingOAuthService struct {
	repo       blingOAuthRepository
	store      security.SecretStore
	httpClient HTTPDoer
	listen     func(string, string) (net.Listener, error)
	now        func() time.Time

	mu       sync.Mutex
	sessions map[string]*blingOAuthSession
}

func NewBlingOAuthService(pool *pgxpool.Pool, store security.SecretStore) *BlingOAuthService {
	return &BlingOAuthService{
		repo:       newPostgresBlingOAuthRepository(pool),
		store:      store,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		listen:     net.Listen,
		now:        time.Now,
		sessions:   make(map[string]*blingOAuthSession),
	}
}

func newBlingOAuthService(repo blingOAuthRepository, store security.SecretStore, httpClient HTTPDoer) *BlingOAuthService {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &BlingOAuthService{repo: repo, store: store, httpClient: httpClient, listen: net.Listen, now: time.Now, sessions: make(map[string]*blingOAuthSession)}
}

func (s *BlingOAuthService) Start(ctx context.Context) (BlingOAuthStartResult, error) {
	if s == nil || s.repo == nil || s.store == nil {
		return BlingOAuthStartResult{}, ErrBlingOAuthConfiguration
	}
	connection, err := s.repo.Connection(ctx)
	if err != nil {
		return BlingOAuthStartResult{}, ErrBlingOAuthConfiguration
	}
	clientID, redirectURI, err := connectionOAuthConfig(connection)
	if err != nil {
		return BlingOAuthStartResult{}, err
	}
	if _, err := loadBlingSecret(ctx, s.store); err != nil {
		return BlingOAuthStartResult{}, err
	}
	redirect, err := validateBlingLoopbackRedirect(redirectURI)
	if err != nil {
		return BlingOAuthStartResult{}, err
	}
	state, err := randomHex(32)
	if err != nil {
		return BlingOAuthStartResult{}, fmt.Errorf("%w: state", ErrBlingOAuthConfiguration)
	}
	sessionID, err := randomHex(16)
	if err != nil {
		return BlingOAuthStartResult{}, fmt.Errorf("%w: session", ErrBlingOAuthConfiguration)
	}
	listener, err := s.listen("tcp", net.JoinHostPort("127.0.0.1", redirect.Port()))
	if err != nil {
		return BlingOAuthStartResult{}, fmt.Errorf("%w: callback port unavailable", ErrBlingOAuthConfiguration)
	}
	if err := s.repo.MarkAuthorizing(ctx); err != nil {
		_ = listener.Close()
		return BlingOAuthStartResult{}, fmt.Errorf("%w: state could not be persisted", ErrBlingOAuthConfiguration)
	}
	authorizationURL, err := buildBlingAuthorizeURL(clientID, redirectURI, state)
	if err != nil {
		_ = listener.Close()
		return BlingOAuthStartResult{}, err
	}
	session := &blingOAuthSession{
		id: sessionID, state: state, path: redirect.Path, listener: listener,
		expiresAt: s.now().Add(blingOAuthSessionTTL), status: blingOAuthStatusAuthorizing,
	}
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()
	go s.runSession(session, clientID, redirectURI)
	return BlingOAuthStartResult{SessionID: sessionID, AuthorizationURL: authorizationURL, Status: blingOAuthStatusAuthorizing}, nil
}

func (s *BlingOAuthService) Status(_ context.Context, sessionID string) (BlingOAuthStatusResult, error) {
	if s == nil || strings.TrimSpace(sessionID) == "" {
		return BlingOAuthStatusResult{}, ErrBlingOAuthSession
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[sessionID]
	if !ok {
		return BlingOAuthStatusResult{}, ErrBlingOAuthSession
	}
	if session.status == blingOAuthStatusAuthorizing && s.now().After(session.expiresAt) {
		session.status = blingOAuthStatusCancelled
		session.errorCode = "BLING_OAUTH_EXPIRED"
		session.message = "A autorização expirou. Inicie a conexão novamente."
	}
	return BlingOAuthStatusResult{SessionID: session.id, Status: session.status, ErrorCode: session.errorCode, Message: session.message}, nil
}

func (s *BlingOAuthService) Test(ctx context.Context) (BlingOAuthTestResult, error) {
	if s == nil || s.repo == nil || s.store == nil {
		return BlingOAuthTestResult{}, ErrBlingOAuthConfiguration
	}
	connection, err := s.repo.Connection(ctx)
	if err != nil {
		return BlingOAuthTestResult{}, ErrBlingOAuthConfiguration
	}
	clientID, _, err := connectionOAuthConfig(connection)
	if err != nil {
		return BlingOAuthTestResult{}, err
	}
	bundle, err := loadBlingSecret(ctx, s.store)
	if err != nil {
		return BlingOAuthTestResult{}, err
	}
	if bundle.AccessToken == "" {
		return BlingOAuthTestResult{}, ErrBlingOAuthNotConnected
	}
	if bundle.AccessTokenExpiresUnix > 0 && bundle.AccessTokenExpiresUnix <= s.now().Add(time.Minute).Unix() {
		if bundle.RefreshToken == "" || (bundle.RefreshTokenExpiresUnix > 0 && bundle.RefreshTokenExpiresUnix <= s.now().Unix()) {
			return BlingOAuthTestResult{}, ErrBlingOAuthNotConnected
		}
		oauthClient, clientErr := NewBlingOAuthClient(s.httpClient, "")
		if clientErr != nil {
			_ = s.repo.RecordTestFailure(ctx, "BLING_OAUTH_CLIENT_INVALID")
			return BlingOAuthTestResult{}, clientErr
		}
		refreshed, refreshErr := oauthClient.RefreshAccessToken(ctx, clientID, bundle.ClientSecret, bundle.RefreshToken)
		if refreshErr != nil {
			_ = s.repo.RecordTestFailure(ctx, "BLING_REFRESH_FAILED")
			return BlingOAuthTestResult{}, refreshErr
		}
		bundle = updateBundleWithToken(bundle, refreshed, s.now())
		encoded, encodeErr := json.Marshal(bundle)
		if encodeErr != nil || s.store.Put(ctx, BlingCredentialSecretRef, encoded) != nil {
			_ = s.repo.RecordTestFailure(ctx, "BLING_TOKEN_STORE_FAILED")
			return BlingOAuthTestResult{}, ErrBlingOAuthStorage
		}
		if err := s.repo.CompleteOAuth(ctx, splitScopes(bundle.Scope), tokenExpiry(s.now(), refreshed.ExpiresIn), refreshExpiry(s.now(), bundle.RefreshToken)); err != nil {
			_ = s.repo.RecordTestFailure(ctx, "BLING_TOKEN_METADATA_FAILED")
			return BlingOAuthTestResult{}, ErrBlingOAuthStorage
		}
	}
	client, err := NewBlingAPIClient(s.httpClient, "", bundle.AccessToken)
	if err != nil {
		return BlingOAuthTestResult{}, err
	}
	probe, err := client.TestConnection(ctx)
	if err != nil {
		_ = s.repo.RecordTestFailure(ctx, oauthErrorCode(err))
		return BlingOAuthTestResult{}, err
	}
	if err := s.repo.RecordTestSuccess(ctx); err != nil {
		return BlingOAuthTestResult{}, ErrBlingOAuthStorage
	}
	return BlingOAuthTestResult{Status: "SUCCESS", PageRecordCount: probe.PageRecordCount}, nil
}

func (s *BlingOAuthService) runSession(session *blingOAuthSession, clientID, redirectURI string) {
	ctx, cancel := context.WithTimeout(context.Background(), blingOAuthSessionTTL)
	defer cancel()
	code, err := s.awaitCallback(ctx, session)
	if err != nil {
		s.failSession(session, oauthErrorCode(err), oauthErrorMessage(err))
		_ = s.repo.MarkOAuthError(ctx, oauthErrorCode(err))
		return
	}
	bundle, err := loadBlingSecret(ctx, s.store)
	if err != nil {
		s.failSession(session, "BLING_SECRET_REQUIRED", "Configure o Client Secret antes de conectar.")
		_ = s.repo.MarkOAuthError(ctx, "BLING_SECRET_REQUIRED")
		return
	}
	oauthClient, err := NewBlingOAuthClient(s.httpClient, "")
	if err != nil {
		s.failSession(session, "BLING_OAUTH_CLIENT_INVALID", "A configuração do cliente OAuth do Bling é inválida.")
		_ = s.repo.MarkOAuthError(ctx, "BLING_OAUTH_CLIENT_INVALID")
		return
	}
	token, err := oauthClient.ExchangeAuthorizationCode(ctx, clientID, bundle.ClientSecret, redirectURI, code)
	if err != nil {
		s.failSession(session, oauthErrorCode(err), oauthErrorMessage(err))
		_ = s.repo.MarkOAuthError(ctx, oauthErrorCode(err))
		return
	}
	apiClient, err := NewBlingAPIClient(s.httpClient, "", token.AccessToken)
	if err != nil {
		s.failSession(session, "BLING_TEST_FAILED", "A conexão foi autorizada, mas não pôde ser validada.")
		_ = s.repo.MarkOAuthError(ctx, "BLING_TEST_FAILED")
		return
	}
	if _, err := apiClient.TestConnection(ctx); err != nil {
		s.failSession(session, oauthErrorCode(err), "A autorização foi recebida, mas o teste do Bling falhou.")
		_ = s.repo.MarkOAuthError(ctx, oauthErrorCode(err))
		return
	}
	previous, previousErr := s.store.Get(ctx, BlingCredentialSecretRef)
	updated := updateBundleWithToken(bundle, token, s.now())
	encoded, err := json.Marshal(updated)
	if err != nil || s.store.Put(ctx, BlingCredentialSecretRef, encoded) != nil {
		s.failSession(session, "BLING_TOKEN_STORE_FAILED", "A autorização foi concluída, mas o token não pôde ser protegido.")
		_ = s.repo.MarkOAuthError(ctx, "BLING_TOKEN_STORE_FAILED")
		return
	}
	if err := s.repo.CompleteOAuth(ctx, splitScopes(updated.Scope), tokenExpiry(s.now(), token.ExpiresIn), refreshExpiry(s.now(), updated.RefreshToken)); err != nil {
		restoreSecret(ctx, s.store, previous, previousErr)
		s.failSession(session, "BLING_OAUTH_PERSIST_FAILED", "A autorização não pôde ser registrada no banco local.")
		_ = s.repo.MarkOAuthError(ctx, "BLING_OAUTH_PERSIST_FAILED")
		return
	}
	s.completeSession(session)
}

func (s *BlingOAuthService) awaitCallback(ctx context.Context, session *blingOAuthSession) (string, error) {
	callback := make(chan struct {
		code string
		err  error
	}, 1)
	sendCallback := func(result struct {
		code string
		err  error
	}) {
		select {
		case callback <- result:
		default:
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc(session.path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != session.path {
			http.Error(w, "callback inválido", http.StatusBadRequest)
			return
		}
		query := r.URL.Query()
		if query.Get("state") != session.state {
			http.Error(w, "estado inválido", http.StatusBadRequest)
			return
		}
		if query.Get("error") != "" || strings.TrimSpace(query.Get("code")) == "" {
			sendCallback(struct {
				code string
				err  error
			}{err: ErrBlingOAuthRejected})
			_, _ = w.Write([]byte("A autorização foi recusada. Você pode voltar ao Majucau."))
			return
		}
		sendCallback(struct {
			code string
			err  error
		}{code: query.Get("code")})
		_, _ = w.Write([]byte("Autorização recebida. Você pode voltar ao Majucau."))
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = server.Serve(session.listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		_ = session.listener.Close()
	}()
	select {
	case result := <-callback:
		return result.code, result.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *BlingOAuthService) failSession(session *blingOAuthSession, code, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session.status = blingOAuthStatusError
	session.errorCode = code
	session.message = message
}

func (s *BlingOAuthService) completeSession(session *blingOAuthSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session.status = blingOAuthStatusConnected
	session.errorCode = ""
	session.message = "Bling autorizado e testado com sucesso."
}

func connectionOAuthConfig(connection database.IntegrationConnection) (string, string, error) {
	if connection.ClientID == nil || strings.TrimSpace(*connection.ClientID) == "" || connection.RedirectUri == nil || strings.TrimSpace(*connection.RedirectUri) == "" {
		return "", "", ErrBlingOAuthConfiguration
	}
	return strings.TrimSpace(*connection.ClientID), strings.TrimSpace(*connection.RedirectUri), nil
}

func validateBlingLoopbackRedirect(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.Port() == "" || parsed.Path == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("%w: redirect must be registered as http://127.0.0.1:port/path", ErrBlingOAuthConfiguration)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("%w: callback port", ErrBlingOAuthConfiguration)
	}
	return parsed, nil
}

func buildBlingAuthorizeURL(clientID, redirectURI, state string) (string, error) {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(state) == "" {
		return "", ErrBlingOAuthConfiguration
	}
	parsed, err := url.Parse(BlingAuthorizeURL)
	if err != nil {
		return "", ErrBlingOAuthConfiguration
	}
	query := parsed.Query()
	query.Set("response_type", "code")
	query.Set("client_id", clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("state", state)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func randomHex(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func loadBlingSecret(ctx context.Context, store security.SecretStore) (blingSecretBundle, error) {
	if store == nil {
		return blingSecretBundle{}, ErrBlingCredentialVault
	}
	encoded, err := store.Get(ctx, BlingCredentialSecretRef)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return blingSecretBundle{}, ErrBlingCredentialMissing
		}
		return blingSecretBundle{}, ErrBlingCredentialVault
	}
	var bundle blingSecretBundle
	if err := json.Unmarshal(encoded, &bundle); err != nil || strings.TrimSpace(bundle.ClientSecret) == "" {
		return blingSecretBundle{}, ErrBlingCredentialMissing
	}
	return bundle, nil
}

func updateBundleWithToken(bundle blingSecretBundle, token BlingOAuthToken, now time.Time) blingSecretBundle {
	bundle.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		bundle.RefreshToken = token.RefreshToken
	}
	if token.TokenType != "" {
		bundle.TokenType = token.TokenType
	}
	if token.Scope != "" {
		bundle.Scope = token.Scope
	}
	if token.ExpiresIn > 0 {
		bundle.AccessTokenExpiresUnix = now.Add(time.Duration(token.ExpiresIn) * time.Second).Unix()
	} else {
		bundle.AccessTokenExpiresUnix = 0
	}
	if token.RefreshToken != "" || bundle.RefreshToken != "" {
		bundle.RefreshTokenExpiresUnix = now.Add(blingRefreshTokenLifetime).Unix()
	}
	return bundle
}

func tokenExpiry(now time.Time, seconds int64) *time.Time {
	if seconds <= 0 {
		return nil
	}
	value := now.Add(time.Duration(seconds) * time.Second)
	return &value
}

func refreshExpiry(now time.Time, refreshToken string) *time.Time {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}
	value := now.Add(blingRefreshTokenLifetime)
	return &value
}

func splitScopes(scope string) []string {
	if strings.TrimSpace(scope) == "" {
		return []string{}
	}
	return strings.Fields(scope)
}

func oauthErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrBlingOAuthCallback):
		return "BLING_OAUTH_CALLBACK_INVALID"
	case errors.Is(err, ErrBlingOAuthRejected):
		return "BLING_OAUTH_REJECTED"
	case errors.Is(err, ErrBlingAPIUnauthorized):
		return "BLING_UNAUTHORIZED"
	case errors.Is(err, ErrBlingAPIRateLimited):
		return "BLING_RATE_LIMITED"
	case errors.Is(err, ErrBlingAPIUnavailable):
		return "BLING_PROVIDER_UNAVAILABLE"
	case errors.Is(err, context.DeadlineExceeded):
		return "BLING_OAUTH_EXPIRED"
	case errors.Is(err, ErrBlingOAuthNotConnected):
		return "BLING_NOT_CONNECTED"
	default:
		return "BLING_OAUTH_FAILED"
	}
}

func oauthErrorMessage(err error) string {
	switch oauthErrorCode(err) {
	case "BLING_OAUTH_CALLBACK_INVALID":
		return "O retorno do Bling não passou na validação de segurança."
	case "BLING_OAUTH_REJECTED":
		return "A autorização foi recusada ou cancelada no Bling."
	case "BLING_UNAUTHORIZED":
		return "O Bling recusou a autorização. Confira o aplicativo e os escopos."
	case "BLING_RATE_LIMITED":
		return "O Bling limitou temporariamente as tentativas. Aguarde e tente novamente."
	case "BLING_PROVIDER_UNAVAILABLE":
		return "O Bling está indisponível no momento."
	case "BLING_OAUTH_EXPIRED":
		return "A autorização expirou. Inicie a conexão novamente."
	default:
		return "Não foi possível concluir a autorização do Bling."
	}
}
