package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrBlingOAuthRejected = errors.New("bling oauth authorization was rejected")
	ErrBlingOAuthSchema   = errors.New("bling oauth response schema is invalid")
)

type BlingOAuthClient struct {
	tokenURL         *url.URL
	httpClient       HTTPDoer
	maxResponseBytes int64
}

// NewBlingOAuthClient creates the server-side token exchange client. Client
// secrets are accepted only by this worker-side boundary and are never put in
// URLs, errors or logs. Loopback HTTP is allowed solely for local tests.
func NewBlingOAuthClient(httpClient HTTPDoer, rawTokenURL string) (*BlingOAuthClient, error) {
	if strings.TrimSpace(rawTokenURL) == "" {
		rawTokenURL = BlingTokenURL
	}
	tokenURL, err := parseBlingURL(rawTokenURL)
	if err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &BlingOAuthClient{tokenURL: tokenURL, httpClient: httpClient, maxResponseBytes: defaultBlingResponseLimit}, nil
}

type BlingOAuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scope        string
	ExpiresIn    int64
}

func (c *BlingOAuthClient) ExchangeAuthorizationCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (BlingOAuthToken, error) {
	return c.exchange(ctx, clientID, clientSecret, url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	})
}

func (c *BlingOAuthClient) RefreshAccessToken(ctx context.Context, clientID, clientSecret, refreshToken string) (BlingOAuthToken, error) {
	return c.exchange(ctx, clientID, clientSecret, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	})
}

func (c *BlingOAuthClient) exchange(ctx context.Context, clientID, clientSecret string, form url.Values) (BlingOAuthToken, error) {
	if c == nil || c.tokenURL == nil || c.httpClient == nil {
		return BlingOAuthToken{}, ErrBlingAPIConfiguration
	}
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return BlingOAuthToken{}, fmt.Errorf("%w: client credentials are required", ErrBlingAPIConfiguration)
	}
	for key, values := range form {
		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				return BlingOAuthToken{}, fmt.Errorf("%w: oauth field %s is required", ErrBlingAPIConfiguration, key)
			}
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return BlingOAuthToken{}, fmt.Errorf("%w: create oauth request", ErrBlingAPIConfiguration)
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return BlingOAuthToken{}, ctx.Err()
		}
		return BlingOAuthToken{}, fmt.Errorf("%w: token request failed", ErrBlingAPIUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.StatusCode == http.StatusTooManyRequests {
			return BlingOAuthToken{}, &BlingAPIError{StatusCode: resp.StatusCode, Code: "BLING_RATE_LIMITED", RetryAfter: retryAfter(resp.Header.Get("Retry-After"))}
		}
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return BlingOAuthToken{}, fmt.Errorf("%w: status %d", ErrBlingOAuthRejected, resp.StatusCode)
		}
		if resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusRequestTimeout {
			return BlingOAuthToken{}, &BlingAPIError{StatusCode: resp.StatusCode, Code: "BLING_OAUTH_UNAVAILABLE"}
		}
		return BlingOAuthToken{}, fmt.Errorf("%w: status %d", ErrBlingOAuthRejected, resp.StatusCode)
	}
	body, err := readLimitedBody(resp.Body, c.maxResponseBytes)
	if err != nil {
		return BlingOAuthToken{}, fmt.Errorf("%w: token response is too large or unreadable", ErrBlingAPIInvalidResponse)
	}
	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return BlingOAuthToken{}, fmt.Errorf("%w: token response is not json", ErrBlingOAuthSchema)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return BlingOAuthToken{}, fmt.Errorf("%w: access_token is missing", ErrBlingOAuthSchema)
	}
	return BlingOAuthToken{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		TokenType:    payload.TokenType,
		Scope:        payload.Scope,
		ExpiresIn:    payload.ExpiresIn,
	}, nil
}
