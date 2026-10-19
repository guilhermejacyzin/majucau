package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"syscall"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/httpretry"
)

const (
	BlingAPIBaseURL   = "https://api.bling.com.br/Api/v3"
	BlingAuthorizeURL = "https://www.bling.com.br/Api/v3/oauth/authorize"
	BlingTokenURL     = "https://api.bling.com.br/Api/v3/oauth/token"

	// The adapter deliberately caps each page at 100 records as a memory and
	// batch safety bound. Bling's general guide says 100 is the default and that
	// callers may configure the limit; it does not establish 100 as the maximum
	// for these specific endpoints. Raising this application cap requires
	// endpoint and account homologation.
	BlingMaxPageSize = 100

	defaultBlingResponseLimit int64 = 4 << 20 // 4 MiB; protects the worker from an unexpected payload.
	defaultBlingMaxAttempts         = httpretry.DefaultMaxAttempts
	defaultBlingRetryBase           = httpretry.DefaultBaseDelay
	defaultBlingRetryJitter         = httpretry.DefaultJitter
	defaultBlingMaxRetryAfter       = httpretry.MaxRetryAfter
)

var (
	ErrBlingAPIUnauthorized    = errors.New("bling api is not authorized")
	ErrBlingAPIRateLimited     = errors.New("bling api rate limit reached")
	ErrBlingAPIUnavailable     = errors.New("bling api is unavailable")
	ErrBlingAPIInvalidResponse = errors.New("bling api returned an invalid response")
	ErrBlingAPISchemaMismatch  = errors.New("bling api response schema is not homologated")
	ErrBlingAPIConfiguration   = errors.New("bling api configuration is invalid")
	errBlingDataMissing        = errors.New("bling response data array is missing")
	errBlingDataNotArray       = errors.New("bling response data is not an array")
	errBlingPageTooLarge       = errors.New("bling response page exceeds the supported record count")
	errBlingResponseTooLarge   = errors.New("bling response exceeds the configured byte limit")
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

// blingTransportFailure preserves a safe public error while retaining whether
// the underlying transport failure is temporary enough for a replay-safe GET.
// The raw transport error is intentionally not exposed because it may contain
// account URLs or other request details.
type blingTransportFailure struct {
	retryable bool
}

func (e *blingTransportFailure) Error() string {
	return fmt.Sprintf("%v: request failed", ErrBlingAPIUnavailable)
}

func (e *blingTransportFailure) Unwrap() error {
	return ErrBlingAPIUnavailable
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
	case isBlingServerError(e.StatusCode) || e.StatusCode == http.StatusRequestTimeout:
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
	httpClient = withBlingRequestPacing(httpClient)
	return &BlingAPIClient{
		baseURL:          baseURL,
		httpClient:       httpClient,
		accessToken:      accessToken,
		maxResponseBytes: defaultBlingResponseLimit,
		maxAttempts:      defaultBlingMaxAttempts,
		wait:             httpretry.Wait,
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

const (
	blingDateFilterNone     = ""
	blingDateFilterDue      = "due"
	blingDateFilterReceived = "received"
	blingDateFilterPayment  = "payment"
)

func financialDateFilterMode(filter ReceivablesFilter) (string, error) {
	modes := 0
	mode := blingDateFilterNone
	for _, candidate := range []struct {
		mode     string
		from, to string
	}{
		{blingDateFilterDue, filter.DueDateFrom, filter.DueDateTo},
		{blingDateFilterReceived, filter.ReceivedDateFrom, filter.ReceivedDateTo},
		{blingDateFilterPayment, filter.PaymentDateFrom, filter.PaymentDateTo},
	} {
		if strings.TrimSpace(candidate.from) != "" || strings.TrimSpace(candidate.to) != "" {
			modes++
			mode = candidate.mode
		}
	}
	if modes > 1 {
		return blingDateFilterNone, fmt.Errorf("%w: select only one financial date filter", ErrBlingAPIConfiguration)
	}
	return mode, nil
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
	// Limit is the server/request page size when known. It is zero when no
	// limit was sent and the response does not report one.
	Limit   int
	Total   *int
	HasNext bool
	// HasNextKnown is true when API metadata establishes continuation or
	// completion, when a documented page size allows completion to be inferred
	// from the number of returned records, or when a non-empty page without a
	// known size is conservatively followed. Empty data with no known size is
	// still ambiguous.
	HasNextKnown bool
}

// BlingAPIResourcePage is the generic name for a bounded raw page returned by
// any list endpoint. BlingReceivablesPage remains as a compatibility alias for
// existing callers.
type BlingAPIResourcePage = BlingReceivablesPage

type blingPagination struct {
	Page    *int  `json:"page"`
	Limit   *int  `json:"limit"`
	Total   *int  `json:"total"`
	HasNext *bool `json:"hasNext"`
}

type blingReceivablesEnvelope struct {
	Data       []json.RawMessage `json:"data"`
	Pagination *blingPagination  `json:"pagination"`
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
	dateMode, err := financialDateFilterMode(filter)
	if err != nil {
		return BlingReceivablesPage{}, err
	}
	switch resource {
	case "contas/receber":
		if dateMode == blingDateFilterPayment {
			return BlingReceivablesPage{}, fmt.Errorf("%w: payment date is not available for accounts receivable", ErrBlingAPIConfiguration)
		}
		if dateMode != blingDateFilterNone {
			query.Set("tipoFiltroData", "V")
			from, to := filter.DueDateFrom, filter.DueDateTo
			if dateMode == blingDateFilterReceived {
				query.Set("tipoFiltroData", "R")
				from, to = filter.ReceivedDateFrom, filter.ReceivedDateTo
			}
			setOptionalQuery(query, "dataInicial", from)
			setOptionalQuery(query, "dataFinal", to)
		}
		setOptionalQuery(query, "situacoes[]", filter.Status)
	case "contas/pagar":
		if dateMode == blingDateFilterReceived {
			return BlingReceivablesPage{}, fmt.Errorf("%w: received date is not available for accounts payable", ErrBlingAPIConfiguration)
		}
		if dateMode == blingDateFilterDue {
			setOptionalQuery(query, "dataVencimentoInicial", filter.DueDateFrom)
			setOptionalQuery(query, "dataVencimentoFinal", filter.DueDateTo)
		}
		if dateMode == blingDateFilterPayment {
			setOptionalQuery(query, "dataPagamentoInicial", filter.PaymentDateFrom)
			setOptionalQuery(query, "dataPagamentoFinal", filter.PaymentDateTo)
		}
		setOptionalQuery(query, "situacao", filter.Status)
	default:
		if dateMode != blingDateFilterNone || strings.TrimSpace(filter.Status) != "" {
			return BlingReceivablesPage{}, fmt.Errorf("%w: financial filters are not available for this resource", ErrBlingAPIConfiguration)
		}
	}
	pathURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pathURL.String(), nil)
	if err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: create request", ErrBlingAPIConfiguration)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("enable-jwt", "1")
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
	return c.doFinancialResourceRequestWithLimit(req, filter, true)
}

// doFinancialResourceRequestWithoutPageLimit is used for endpoints whose
// public contract does not accept a page-size parameter. The local record cap
// still limits decoding but cannot be used to infer pagination completion.
func (c *BlingAPIClient) doFinancialResourceRequestWithoutPageLimit(req *http.Request, filter ReceivablesFilter) (BlingReceivablesPage, error) {
	return c.doFinancialResourceRequestWithLimit(req, filter, false)
}

func (c *BlingAPIClient) doFinancialResourceRequestWithLimit(req *http.Request, filter ReceivablesFilter, requestedLimitKnown bool) (BlingReceivablesPage, error) {
	return withBlingReadRetry(req.Context(), req.Method, c.maxAttempts, c.wait, c.random, func() (BlingReceivablesPage, error) {
		return c.doFinancialResourceAttempt(req, filter, requestedLimitKnown)
	})
}

func (c *BlingAPIClient) doFinancialResourceAttempt(req *http.Request, filter ReceivablesFilter, requestedLimitKnown bool) (BlingReceivablesPage, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return BlingReceivablesPage{}, req.Context().Err()
		}
		return BlingReceivablesPage{}, newBlingTransportFailure(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		apiErr := &BlingAPIError{StatusCode: resp.StatusCode, Code: "BLING_API_REJECTED", RetryAfter: retryAfter(resp.Header.Get("Retry-After"))}
		if resp.StatusCode == http.StatusTooManyRequests {
			apiErr.Code = "BLING_RATE_LIMITED"
		}
		return BlingReceivablesPage{}, apiErr
	}
	envelope, err := decodeBlingReceivablesEnvelope(resp.Body, c.maxResponseBytes)
	if errors.Is(err, errBlingDataMissing) {
		return BlingReceivablesPage{}, fmt.Errorf("%w: data array is missing", ErrBlingAPISchemaMismatch)
	}
	if errors.Is(err, errBlingDataNotArray) || errors.Is(err, errBlingPageTooLarge) {
		return BlingReceivablesPage{}, fmt.Errorf("%w: %v", ErrBlingAPISchemaMismatch, err)
	}
	if err != nil {
		if errors.Is(err, errBlingResponseTooLarge) {
			return BlingReceivablesPage{}, fmt.Errorf("%w: response body exceeds limit", ErrBlingAPIInvalidResponse)
		}
		return BlingReceivablesPage{}, fmt.Errorf("%w: response is not valid bounded json", ErrBlingAPIInvalidResponse)
	}
	page := BlingReceivablesPage{Records: envelope.Data, Page: filter.Page}
	pageLimitKnown := requestedLimitKnown
	if requestedLimitKnown {
		page.Limit = filter.Limit
	}
	if envelope.Pagination != nil {
		if envelope.Pagination.Page != nil {
			if *envelope.Pagination.Page < 1 || *envelope.Pagination.Page != filter.Page {
				return BlingReceivablesPage{}, fmt.Errorf("%w: pagination page does not match the request", ErrBlingAPISchemaMismatch)
			}
			page.Page = *envelope.Pagination.Page
		}
		if envelope.Pagination.Limit != nil {
			if *envelope.Pagination.Limit < 1 || *envelope.Pagination.Limit > BlingMaxPageSize {
				return BlingReceivablesPage{}, fmt.Errorf("%w: pagination limit is outside the supported range", ErrBlingAPISchemaMismatch)
			}
			if requestedLimitKnown && *envelope.Pagination.Limit > filter.Limit {
				return BlingReceivablesPage{}, fmt.Errorf("%w: response page limit exceeds the requested limit", ErrBlingAPISchemaMismatch)
			}
			page.Limit = *envelope.Pagination.Limit
			pageLimitKnown = true
		}
		if envelope.Pagination.Total != nil && *envelope.Pagination.Total < 0 {
			return BlingReceivablesPage{}, fmt.Errorf("%w: pagination total cannot be negative", ErrBlingAPISchemaMismatch)
		}
		if envelope.Pagination.Total != nil && *envelope.Pagination.Total == 0 && len(envelope.Data) > 0 {
			return BlingReceivablesPage{}, fmt.Errorf("%w: pagination total is empty but the response contains records", ErrBlingAPISchemaMismatch)
		}
		page.Total = envelope.Pagination.Total
		if envelope.Pagination.HasNext != nil {
			page.HasNext = *envelope.Pagination.HasNext
			page.HasNextKnown = true
		} else if page.Total != nil && pageLimitKnown {
			page.HasNext = hasNextFromTotal(page.Page, page.Limit, *page.Total)
			page.HasNextKnown = true
		}
	}
	if pageLimitKnown && page.Limit > 0 && len(page.Records) > page.Limit {
		return BlingReceivablesPage{}, fmt.Errorf("%w: response data exceeds the page limit", ErrBlingAPISchemaMismatch)
	}
	if !page.HasNextKnown {
		// The documented list responses expose a data array, while pagina/limite
		// (default 100) identify the page window. A full page may have another
		// page; a short page is terminal. This also handles an empty page after
		// an exact multiple of the page size. If no page size is documented or
		// returned, keep the conservative behavior and reject an empty result.
		if pageLimitKnown && page.Limit > 0 {
			page.HasNext = len(page.Records) == page.Limit
			page.HasNextKnown = true
		} else if len(page.Records) > 0 {
			page.HasNext = true
			page.HasNextKnown = true
		}
	}
	return page, nil
}

// hasNextFromTotal compares page position without multiplying untrusted
// pagination values, which could overflow and falsely mark a list complete.
func hasNextFromTotal(page, limit, total int) bool {
	if page < 1 || limit < 1 || total < 1 {
		return false
	}
	fullPages := total / limit
	return page < fullPages || (page == fullPages && total%limit != 0)
}

func decodeBlingReceivablesEnvelope(reader io.Reader, limit int64) (blingReceivablesEnvelope, error) {
	if limit <= 0 {
		limit = defaultBlingResponseLimit
	}
	const maxInt64 = int64(^uint64(0) >> 1)
	if limit >= maxInt64 {
		return blingReceivablesEnvelope{}, errors.New("invalid Bling response byte limit")
	}
	limited := &io.LimitedReader{R: reader, N: limit + 1}
	decoder := json.NewDecoder(limited)
	var envelope blingReceivablesEnvelope
	opening, err := decoder.Token()
	if err != nil {
		if limited.N == 0 {
			return blingReceivablesEnvelope{}, errBlingResponseTooLarge
		}
		return blingReceivablesEnvelope{}, err
	}
	if opening != json.Delim('{') {
		return blingReceivablesEnvelope{}, errors.New("Bling response envelope is not an object")
	}
	seenData := false
	for decoder.More() {
		propertyToken, err := decoder.Token()
		if err != nil {
			if limited.N == 0 {
				return blingReceivablesEnvelope{}, errBlingResponseTooLarge
			}
			return blingReceivablesEnvelope{}, err
		}
		property, ok := propertyToken.(string)
		if !ok {
			return blingReceivablesEnvelope{}, errors.New("invalid Bling response property")
		}
		switch property {
		case "data":
			if seenData {
				return blingReceivablesEnvelope{}, errBlingDataNotArray
			}
			seenData = true
			opening, err := decoder.Token()
			if err != nil {
				if limited.N == 0 {
					return blingReceivablesEnvelope{}, errBlingResponseTooLarge
				}
				return blingReceivablesEnvelope{}, err
			}
			if opening == nil {
				return blingReceivablesEnvelope{}, errBlingDataMissing
			}
			if opening != json.Delim('[') {
				return blingReceivablesEnvelope{}, errBlingDataNotArray
			}
			envelope.Data = make([]json.RawMessage, 0, BlingMaxPageSize)
			for decoder.More() {
				if len(envelope.Data) >= BlingMaxPageSize {
					return blingReceivablesEnvelope{}, errBlingPageTooLarge
				}
				var record json.RawMessage
				if err := decoder.Decode(&record); err != nil {
					if limited.N == 0 {
						return blingReceivablesEnvelope{}, errBlingResponseTooLarge
					}
					return blingReceivablesEnvelope{}, err
				}
				envelope.Data = append(envelope.Data, record)
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				if limited.N == 0 {
					return blingReceivablesEnvelope{}, errBlingResponseTooLarge
				}
				return blingReceivablesEnvelope{}, errors.New("invalid Bling response data array")
			}
		case "pagination":
			if err := decoder.Decode(&envelope.Pagination); err != nil {
				if limited.N == 0 {
					return blingReceivablesEnvelope{}, errBlingResponseTooLarge
				}
				return blingReceivablesEnvelope{}, err
			}
		default:
			var ignored json.RawMessage
			if err := decoder.Decode(&ignored); err != nil {
				if limited.N == 0 {
					return blingReceivablesEnvelope{}, errBlingResponseTooLarge
				}
				return blingReceivablesEnvelope{}, err
			}
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		if limited.N == 0 {
			return blingReceivablesEnvelope{}, errBlingResponseTooLarge
		}
		return blingReceivablesEnvelope{}, errors.New("invalid Bling response envelope")
	}
	if !seenData {
		return blingReceivablesEnvelope{}, errBlingDataMissing
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if limited.N == 0 {
			return blingReceivablesEnvelope{}, errBlingResponseTooLarge
		}
		if err == nil {
			return blingReceivablesEnvelope{}, errors.New("Bling response contains multiple JSON values")
		}
		return blingReceivablesEnvelope{}, err
	}
	if limited.N == 0 {
		return blingReceivablesEnvelope{}, errBlingResponseTooLarge
	}
	return envelope, nil
}

func retryableBlingError(err error) bool {
	var transportFailure *blingTransportFailure
	if errors.As(err, &transportFailure) {
		return transportFailure.retryable
	}
	var apiErr *BlingAPIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusRequestTimeout || apiErr.StatusCode == http.StatusTooManyRequests || isBlingServerError(apiErr.StatusCode)
	}
	return false
}

func newBlingTransportFailure(err error) error {
	return &blingTransportFailure{retryable: isRetryableBlingTransportError(err)}
}

func isRetryableBlingTransportError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return true
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		return dnsError.IsTimeout || dnsError.IsTemporary
	}
	return errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.EHOSTUNREACH) ||
		errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EPIPE)
}

func isBlingServerError(statusCode int) bool {
	return statusCode >= http.StatusInternalServerError && statusCode < 600
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
	return httpretry.ParseRetryAfter(value, now)
}
