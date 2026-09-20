package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BlingAPIBaseURL   = "https://api.bling.com.br/Api/v3"
	BlingAuthorizeURL = "https://www.bling.com.br/Api/v3/oauth/authorize"
	BlingTokenURL     = "https://api.bling.com.br/Api/v3/oauth/token"

	// Bling documents a maximum of 100 records per page for the resources used
	// by the financial integration. This is deliberately a hard contract here;
	// callers must not silently ask the API for an unsupported page size.
	BlingMaxPageSize = 100

	defaultBlingResponseLimit int64 = 4 << 20 // 4 MiB; protects the worker from an unexpected payload.
	defaultBlingMaxAttempts         = 3
	defaultBlingRetryBase           = 250 * time.Millisecond
	defaultBlingRetryJitter         = 100 * time.Millisecond
	defaultBlingMaxRetryAfter       = 30 * time.Second
)

var (
	ErrBlingAPIUnauthorized    = errors.New("bling api is not authorized")
	ErrBlingAPIRateLimited     = errors.New("bling api rate limit reached")
	ErrBlingAPIUnavailable     = errors.New("bling api is unavailable")
	ErrBlingAPIInvalidResponse = errors.New("bling api returned an invalid response")
	ErrBlingAPISchemaMismatch  = errors.New("bling api response schema is not homologated")
	ErrBlingAPIConfiguration   = errors.New("bling api configuration is invalid")
)

// HTTPDoer is intentionally small so the adapter can be tested with an
// httptest server without introducing a second HTTP stack.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// BlingAPIError contains only safe transport metadata. Response bodies are
// never copied into this error because they can contain account data.
type BlingAPIError struct {
	StatusCode int
	Code       string
	RetryAfter time.Duration
}

func (e *BlingAPIError) Error() string {
	if e == nil {
		return "bling api error"
	}
	if e.Code == "" {
		return fmt.Sprintf("bling api request failed with status %d", e.StatusCode)
	}
	return fmt.Sprintf("bling api request failed with status %d (%s)", e.StatusCode, e.Code)
}

func (e *BlingAPIError) Unwrap() error {
	if e == nil {
		return nil
	}
	switch {
	case e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden:
		return ErrBlingAPIUnauthorized
	case e.StatusCode == http.StatusTooManyRequests:
		return ErrBlingAPIRateLimited
	case e.StatusCode >= http.StatusInternalServerError || e.StatusCode == http.StatusRequestTimeout:
		return ErrBlingAPIUnavailable
	default:
		return nil
	}
}

type BlingAPIClient struct {
	baseURL          *url.URL
	httpClient       HTTPDoer
	accessToken      string
	maxResponseBytes int64
	maxAttempts      int
	wait             func(context.Context, time.Duration) error
	random           func() float64
}

// NewBlingAPIClient builds a read-only Bling v3 client. The production URL is
// HTTPS. HTTP is accepted only for loopback hosts so tests can use httptest
// without weakening the production transport contract.
func NewBlingAPIClient(httpClient HTTPDoer, rawBaseURL, accessToken string) (*BlingAPIClient, error) {
	if strings.TrimSpace(rawBaseURL) == "" {
		rawBaseURL = BlingAPIBaseURL
	}
	baseURL, err := parseBlingURL(rawBaseURL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("%w: access token is required", ErrBlingAPIConfiguration)
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &BlingAPIClient{
		baseURL:          baseURL,
		httpClient:       httpClient,
		accessToken:      accessToken,
		maxResponseBytes: defaultBlingResponseLimit,
		maxAttempts:      defaultBlingMaxAttempts,
		wait:             waitBlingRetry,
		random:           rand.Float64,
	}, nil
}

type ReceivablesFilter struct {
	Page             int
	Limit            int
	DueDateFrom      string
	DueDateTo        string
	ReceivedDateFrom string
	ReceivedDateTo   string
	PaymentDateFrom  string
	PaymentDateTo    string
	Status           string
}

func (f ReceivablesFilter) normalized() (ReceivablesFilter, error) {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = BlingMaxPageSize
	}
	if f.Page < 1 {
		return ReceivablesFilter{}, fmt.Errorf("%w: page must be positive", ErrBlingAPIConfiguration)
	}
	if f.Limit < 1 || f.Limit > BlingMaxPageSize {
		return ReceivablesFilter{}, fmt.Errorf("%w: limit must be between 1 and %d", ErrBlingAPIConfiguration, BlingMaxPageSize)
	}
	return f, nil
}

// BlingReceivablesPage deliberately preserves each item as JSON until BK-040
// records a real, sanitized response from the customer's Bling account. No
// guessed field mapping can silently become a financial fact.
type BlingReceivablesPage struct {
	Records []json.RawMessage
	Page    int
	Limit   int
	Total   *int
	HasNext bool
}

type blingPagination struct {
	Page    *int  `json:"page"`
	Limit   *int  `json:"limit"`
	Total   *int  `json:"total"`
	HasNext *bool `json:"hasNext"`
}

type blingReceivablesEnvelope struct {
	Data       json.RawMessage  `json:"data"`
	Pagination *blingPagination `json:"pagination"`
}

func (c *BlingAPIClient) ListReceivables(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	return c.listFinancialResource(ctx, "contas/receber", filter)
}

// ListPayables reads the Bling accounts-payable resource using the same
// conservative raw-page contract. No payable field is normalized here.
func (c *BlingAPIClient) ListPayables(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	return c.listFinancialResource(ctx, "contas/pagar", filter)
}

func (c *BlingAPIClient) listFinancialResource(ctx context.Context, resource string, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	if c == nil || c.baseURL == nil || c.httpClient == nil {
		return BlingReceivablesPage{}, ErrBlingAPIConfiguration
	}
	filter, err := filter.normalized()
	if err != nil {
		return BlingReceivablesPage{}, err
	}
	pathURL := *c.baseURL
	pathURL.Path = strings.TrimRight(pathURL.Path, "/") + "/" + strings.TrimLeft(resource, "/")
	query := pathURL.Query()
	query.Set("pagina", strconv.Itoa(filter.Page))
	query.Set("limite", strconv.Itoa(filter.Limit))
	setOptionalQuery(query, "dataVencimentoInicial", filter.DueDateFrom)
	setOptionalQuery(query, "dataVencimentoFinal", filter.DueDateTo)
	setOptionalQuery(query, "dataRecebimentoInicial", filter.ReceivedDateFrom)
	setOptionalQuery(query, "dataRecebimentoFinal", filter.ReceivedDateTo)
	setOptionalQuery(query, "dataPagamentoInicial", filter.PaymentDateFrom)
	setOptionalQuery(query, "dataPagamentoFinal", filter.PaymentDateTo)
	setOptionalQuery(query, "situacao", filter.Status)
	pathURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pathURL.String(), nil)
	if err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: create request", ErrBlingAPIConfiguration)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	return c.doFinancialResourceRequest(req, filter)
}

type BlingConnectionProbe struct {
	PageRecordCount int
	Page            int
	Limit           int
}

func (c *BlingAPIClient) TestConnection(ctx context.Context) (BlingConnectionProbe, error) {
	page, err := c.ListReceivables(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if err != nil {
		return BlingConnectionProbe{}, err
	}
	return BlingConnectionProbe{PageRecordCount: len(page.Records), Page: page.Page, Limit: page.Limit}, nil
}

func (c *BlingAPIClient) doFinancialResourceRequest(req *http.Request, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	attempts := c.maxAttempts
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		page, err := c.doFinancialResourceAttempt(req, filter)
		if err == nil {
			return page, nil
		}
		if attempt == attempts || !retryableBlingError(err) {
			return BlingReceivablesPage{}, err
		}
		var apiErr *BlingAPIError
		var retryAfter time.Duration
		if errors.As(err, &apiErr) {
			retryAfter = apiErr.RetryAfter
		}
		wait := c.wait
		if wait == nil {
			wait = waitBlingRetry
		}
		if err := wait(req.Context(), c.retryDelay(attempt, retryAfter)); err != nil {
			return BlingReceivablesPage{}, err
		}
	}
	return BlingReceivablesPage{}, ErrBlingAPIUnavailable
}

func (c *BlingAPIClient) doFinancialResourceAttempt(req *http.Request, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return BlingReceivablesPage{}, req.Context().Err()
		}
		return BlingReceivablesPage{}, fmt.Errorf("%w: request failed", ErrBlingAPIUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		apiErr := &BlingAPIError{StatusCode: resp.StatusCode, Code: "BLING_API_REJECTED", RetryAfter: retryAfter(resp.Header.Get("Retry-After"))}
		if resp.StatusCode == http.StatusTooManyRequests {
			apiErr.Code = "BLING_RATE_LIMITED"
		}
		return BlingReceivablesPage{}, apiErr
	}
	body, err := readLimitedBody(resp.Body, c.maxResponseBytes)
	if err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: response body is too large or unreadable", ErrBlingAPIInvalidResponse)
	}
	var envelope blingReceivablesEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: response is not json", ErrBlingAPIInvalidResponse)
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return BlingReceivablesPage{}, fmt.Errorf("%w: data array is missing", ErrBlingAPISchemaMismatch)
	}
	var records []json.RawMessage
	if err := json.Unmarshal(envelope.Data, &records); err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: data is not an array", ErrBlingAPISchemaMismatch)
	}
	page := BlingReceivablesPage{Records: records, Page: filter.Page, Limit: filter.Limit}
	if envelope.Pagination != nil {
		if envelope.Pagination.Page != nil && *envelope.Pagination.Page > 0 {
			page.Page = *envelope.Pagination.Page
		}
		if envelope.Pagination.Limit != nil && *envelope.Pagination.Limit > 0 {
			page.Limit = *envelope.Pagination.Limit
		}
		page.Total = envelope.Pagination.Total
		if envelope.Pagination.HasNext != nil {
			page.HasNext = *envelope.Pagination.HasNext
		} else if page.Total != nil {
			page.HasNext = page.Page*page.Limit < *page.Total
		}
	}
	return page, nil
}

func retryableBlingError(err error) bool {
	var apiErr *BlingAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusRequestTimeout || apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= http.StatusInternalServerError
	}
	return errors.Is(err, ErrBlingAPIUnavailable)
}

func (c *BlingAPIClient) retryDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if retryAfter > defaultBlingMaxRetryAfter {
			return defaultBlingMaxRetryAfter
		}
		return retryAfter
	}
	delay := defaultBlingRetryBase
	for step := 1; step < attempt; step++ {
		delay *= 2
	}
	jitter := 0.5
	if c.random != nil {
		jitter = c.random()
	}
	if jitter < 0 {
		jitter = 0
	}
	if jitter > 1 {
		jitter = 1
	}
	return delay + time.Duration(jitter*float64(defaultBlingRetryJitter))
}

func waitBlingRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func parseBlingURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: base URL must be absolute", ErrBlingAPIConfiguration)
	}
	if parsed.Scheme != "https" && !isLoopbackHost(parsed.Hostname()) {
		return nil, fmt.Errorf("%w: Bling transport must use HTTPS", ErrBlingAPIConfiguration)
	}
	return parsed, nil
}

func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func setOptionalQuery(query url.Values, key, value string) {
	if strings.TrimSpace(value) != "" {
		query.Set(key, value)
	}
}

func retryAfter(value string) time.Duration {
	return retryAfterAt(value, time.Now())
}

func retryAfterAt(value string, now time.Time) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil {
		if seconds < 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(strings.TrimSpace(value))
	if err != nil || !when.After(now) {
		return 0
	}
	return when.Sub(now)
}

func readLimitedBody(reader io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		limit = defaultBlingResponseLimit
	}
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errors.New("response limit exceeded")
	}
	return body, nil
}
