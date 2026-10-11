package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// GetPayable reads one account-payable detail from Bling and preserves its
// payload as RAW JSON. The ID is numeric so it cannot inject a path or query.
func (c *BlingAPIClient) GetPayable(ctx context.Context, id int64) (json.RawMessage, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: payable ID must be positive", ErrBlingAPIConfiguration)
	}
	return c.getDataRecord(ctx, "contas/pagar/"+strconv.FormatInt(id, 10))
}

// GetBordero reads one payment bordero from Bling and preserves the response
// as RAW JSON. It does not infer category membership or allocate payments.
func (c *BlingAPIClient) GetBordero(ctx context.Context, id int64) (json.RawMessage, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: bordero ID must be positive", ErrBlingAPIConfiguration)
	}
	return c.getDataRecord(ctx, "borderos/"+strconv.FormatInt(id, 10))
}

// GetFinancialCategory reads one Bling income/expense category and preserves
// its payload as RAW JSON. The category is not mapped to a Majucau rule here.
func (c *BlingAPIClient) GetFinancialCategory(ctx context.Context, id int64) (json.RawMessage, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: financial category ID must be positive", ErrBlingAPIConfiguration)
	}
	return c.getDataRecord(ctx, "categorias/receitas-despesas/"+strconv.FormatInt(id, 10))
}

// ListFinancialCategories reads one bounded page of Bling income/expense
// categories. Callers must use HasNextKnown; an omitted pagination marker is
// never treated as the end of the list.
func (c *BlingAPIClient) ListFinancialCategories(ctx context.Context, page, limit int) (BlingReceivablesPage, error) {
	return c.listFinancialResource(ctx, "categorias/receitas-despesas", ReceivablesFilter{Page: page, Limit: limit})
}

func (c *BlingAPIClient) getDataRecord(ctx context.Context, resource string) (json.RawMessage, error) {
	if c == nil || c.baseURL == nil || c.httpClient == nil {
		return nil, ErrBlingAPIConfiguration
	}
	pathURL := *c.baseURL
	pathURL.Path = strings.TrimRight(pathURL.Path, "/") + "/" + strings.TrimLeft(resource, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pathURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: create request", ErrBlingAPIConfiguration)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("enable-jwt", "1")

	return withBlingReadRetry(req.Context(), req.Method, c.maxAttempts, c.wait, c.random, func() (json.RawMessage, error) {
		return c.getDataRecordAttempt(req)
	})
}

func (c *BlingAPIClient) getDataRecordAttempt(req *http.Request) (json.RawMessage, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if req.Context().Err() != nil {
			return nil, req.Context().Err()
		}
		return nil, fmt.Errorf("%w: request failed", ErrBlingAPIUnavailable)
	}
	defer resp.Body.Close()
	if err := req.Context().Err(); err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		code := "BLING_API_REJECTED"
		if resp.StatusCode == http.StatusTooManyRequests {
			code = "BLING_RATE_LIMITED"
		}
		return nil, &BlingAPIError{StatusCode: resp.StatusCode, Code: code, RetryAfter: retryAfter(resp.Header.Get("Retry-After"))}
	}

	record, err := decodeBlingDataRecord(resp.Body, c.maxResponseBytes)
	if err == nil {
		return record, nil
	}
	if errors.Is(err, errBlingResponseTooLarge) {
		return nil, fmt.Errorf("%w: response body exceeds limit", ErrBlingAPIInvalidResponse)
	}
	return nil, fmt.Errorf("%w: detail response is not a bounded data object", ErrBlingAPISchemaMismatch)
}

func decodeBlingDataRecord(reader io.Reader, limit int64) (json.RawMessage, error) {
	if limit <= 0 {
		limit = defaultBlingResponseLimit
	}
	const maxInt64 = int64(^uint64(0) >> 1)
	if limit >= maxInt64 {
		return nil, errors.New("invalid Bling response byte limit")
	}
	limited := &io.LimitedReader{R: reader, N: limit + 1}
	decoder := json.NewDecoder(limited)
	opening, err := decoder.Token()
	if err != nil {
		if limited.N == 0 {
			return nil, errBlingResponseTooLarge
		}
		return nil, err
	}
	if opening != json.Delim('{') {
		return nil, errors.New("Bling detail envelope is not an object")
	}
	var record json.RawMessage
	seenData := false
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			if limited.N == 0 {
				return nil, errBlingResponseTooLarge
			}
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("invalid Bling detail property")
		}
		if key == "data" {
			if seenData {
				return nil, errors.New("duplicate Bling detail data property")
			}
			seenData = true
			if err := decoder.Decode(&record); err != nil {
				if limited.N == 0 {
					return nil, errBlingResponseTooLarge
				}
				return nil, err
			}
		} else {
			var ignored json.RawMessage
			if err := decoder.Decode(&ignored); err != nil {
				if limited.N == 0 {
					return nil, errBlingResponseTooLarge
				}
				return nil, err
			}
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		if limited.N == 0 {
			return nil, errBlingResponseTooLarge
		}
		return nil, errors.New("invalid Bling detail envelope")
	}
	if !seenData || len(record) == 0 {
		return nil, errors.New("Bling detail data is missing")
	}
	if len(strings.TrimSpace(string(record))) == 0 || strings.TrimSpace(string(record))[0] != '{' {
		return nil, errors.New("Bling detail data is not an object")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if limited.N == 0 {
			return nil, errBlingResponseTooLarge
		}
		return nil, errors.New("Bling detail response has trailing data")
	}
	if limited.N == 0 {
		return nil, errBlingResponseTooLarge
	}
	return record, nil
}
