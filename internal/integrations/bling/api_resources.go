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
	"time"
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

// GetCashMovement reads one Bling cash/bank movement and preserves its payload
// as RAW JSON. It does not classify the movement or infer financial meaning.
func (c *BlingAPIClient) GetCashMovement(ctx context.Context, id int64) (json.RawMessage, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: cash movement ID must be positive", ErrBlingAPIConfiguration)
	}
	return c.getDataRecord(ctx, "caixas/"+strconv.FormatInt(id, 10))
}

// GetFinancialCategory reads one Bling income/expense category and preserves
// its payload as RAW JSON. The category is not mapped to a Majucau rule here.
func (c *BlingAPIClient) GetFinancialCategory(ctx context.Context, id int64) (json.RawMessage, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: financial category ID must be positive", ErrBlingAPIConfiguration)
	}
	return c.getDataRecord(ctx, "categorias/receitas-despesas/"+strconv.FormatInt(id, 10))
}

const maxFinancialCategoryHierarchyDepth = 32

// FinancialCategoryNode is one bounded, read-only category record. Payload
// remains the original JSON so callers can preserve its lineage as RAW.
type FinancialCategoryNode struct {
	ID       int64
	ParentID int64
	Type     int64
	Payload  json.RawMessage
}

// WalkFinancialCategoryHierarchy visits the requested category and each parent
// up to the root, in leaf-to-root order. It never loads the complete category
// list or hierarchy into memory. A missing or invalid parent ID, cycle, or
// hierarchy deeper than the safety limit fails closed. Callers that persist
// visitor results must use a transaction and roll it back if this method fails.
func (c *BlingAPIClient) WalkFinancialCategoryHierarchy(ctx context.Context, id int64, visit func(FinancialCategoryNode) error) error {
	if c == nil || c.baseURL == nil || c.httpClient == nil {
		return ErrBlingAPIConfiguration
	}
	if id <= 0 || visit == nil {
		return fmt.Errorf("%w: category ID and visitor are required", ErrBlingAPIConfiguration)
	}
	seen := make(map[int64]struct{}, maxFinancialCategoryHierarchyDepth)
	currentID := id
	for depth := 0; depth < maxFinancialCategoryHierarchyDepth; depth++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, exists := seen[currentID]; exists {
			return fmt.Errorf("%w: category hierarchy contains a cycle", ErrBlingAPISchemaMismatch)
		}
		seen[currentID] = struct{}{}
		payload, err := c.GetFinancialCategory(ctx, currentID)
		if err != nil {
			return err
		}
		var category struct {
			ID          int64   `json:"id"`
			ParentID    *int64  `json:"idCategoriaPai"`
			Type        int64   `json:"tipo"`
			Description *string `json:"descricao"`
		}
		if err := json.Unmarshal(payload, &category); err != nil {
			return fmt.Errorf("%w: category response is invalid", ErrBlingAPISchemaMismatch)
		}
		if category.ID != currentID || category.ParentID == nil || *category.ParentID < 0 || category.Type < 1 || category.Type > 3 || category.Description == nil {
			return fmt.Errorf("%w: category identity or hierarchy fields are missing or invalid", ErrBlingAPISchemaMismatch)
		}
		if *category.ParentID == currentID {
			return fmt.Errorf("%w: category is its own parent", ErrBlingAPISchemaMismatch)
		}
		if *category.ParentID > 0 {
			if _, exists := seen[*category.ParentID]; exists {
				return fmt.Errorf("%w: category hierarchy contains a cycle", ErrBlingAPISchemaMismatch)
			}
		}
		if err := visit(FinancialCategoryNode{ID: category.ID, ParentID: *category.ParentID, Type: category.Type, Payload: payload}); err != nil {
			return err
		}
		if *category.ParentID == 0 {
			return nil
		}
		currentID = *category.ParentID
	}
	return fmt.Errorf("%w: category hierarchy exceeds the safety limit", ErrBlingAPISchemaMismatch)
}

// ListFinancialCategories reads one bounded page of Bling income/expense
// categories. When explicit next-page metadata is absent, the documented
// requested page size determines completion: full pages continue and short
// pages are terminal.
func (c *BlingAPIClient) ListFinancialCategories(ctx context.Context, page, limit int) (BlingReceivablesPage, error) {
	return c.listFinancialResource(ctx, "categorias/receitas-despesas", ReceivablesFilter{Page: page, Limit: limit})
}

// CashMovementsFilter contains documented date/page/category parameters and a
// local response record cap. MaxRecords is not sent to Bling because the
// current GET /caixas contract does not document a limit query parameter.
type CashMovementsFilter struct {
	Page        int
	MaxRecords  int
	DateFrom    string
	DateTo      string
	CategoryIDs []int64
}

const maxCashMovementDateFilterBytes = 64
const maxCashMovementCategoryIDs = 100

func (f CashMovementsFilter) normalized() (CashMovementsFilter, error) {
	f.DateFrom = strings.TrimSpace(f.DateFrom)
	f.DateTo = strings.TrimSpace(f.DateTo)
	if len(f.DateFrom) > maxCashMovementDateFilterBytes || len(f.DateTo) > maxCashMovementDateFilterBytes {
		return CashMovementsFilter{}, fmt.Errorf("%w: date filters exceed %d bytes", ErrBlingAPIConfiguration, maxCashMovementDateFilterBytes)
	}
	dateFrom, hasDateFrom := parseCashMovementDate(f.DateFrom)
	if f.DateFrom != "" && !hasDateFrom {
		return CashMovementsFilter{}, fmt.Errorf("%w: invalid cash movement start date", ErrBlingAPIConfiguration)
	}
	dateTo, hasDateTo := parseCashMovementDate(f.DateTo)
	if f.DateTo != "" && !hasDateTo {
		return CashMovementsFilter{}, fmt.Errorf("%w: invalid cash movement end date", ErrBlingAPIConfiguration)
	}
	if hasDateFrom && hasDateTo && dateFrom.After(dateTo) {
		return CashMovementsFilter{}, fmt.Errorf("%w: cash movement start date is after end date", ErrBlingAPIConfiguration)
	}
	if len(f.CategoryIDs) > maxCashMovementCategoryIDs {
		return CashMovementsFilter{}, fmt.Errorf("%w: category filter exceeds %d IDs", ErrBlingAPIConfiguration, maxCashMovementCategoryIDs)
	}
	seenCategoryIDs := make(map[int64]struct{}, len(f.CategoryIDs))
	for _, id := range f.CategoryIDs {
		if id <= 0 {
			return CashMovementsFilter{}, fmt.Errorf("%w: category IDs must be positive", ErrBlingAPIConfiguration)
		}
		if _, exists := seenCategoryIDs[id]; exists {
			return CashMovementsFilter{}, fmt.Errorf("%w: category IDs must be unique", ErrBlingAPIConfiguration)
		}
		seenCategoryIDs[id] = struct{}{}
	}
	if f.Page == 0 {
		f.Page = 1
	}
	if f.MaxRecords == 0 {
		f.MaxRecords = BlingMaxPageSize
	}
	if f.Page < 1 {
		return CashMovementsFilter{}, fmt.Errorf("%w: page must be positive", ErrBlingAPIConfiguration)
	}
	if f.MaxRecords < 1 || f.MaxRecords > BlingMaxPageSize {
		return CashMovementsFilter{}, fmt.Errorf("%w: max records must be between 1 and %d", ErrBlingAPIConfiguration, BlingMaxPageSize)
	}
	return f, nil
}

func parseCashMovementDate(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return time.Time{}, false
	}
	return parsed, true
}

// ListCashMovements reads one bounded page from GET /caixas and preserves
// each item as RAW JSON. MaxRecords is an application-side cap, not a query
// parameter. If the response also omits a page size or explicit page metadata,
// HasNextKnown stays false for an empty page; callers must fail closed.
func (c *BlingAPIClient) ListCashMovements(ctx context.Context, filter CashMovementsFilter) (BlingAPIResourcePage, error) {
	if c == nil || c.baseURL == nil || c.httpClient == nil {
		return BlingReceivablesPage{}, ErrBlingAPIConfiguration
	}
	filter, err := filter.normalized()
	if err != nil {
		return BlingReceivablesPage{}, err
	}
	pathURL := *c.baseURL
	pathURL.Path = strings.TrimRight(pathURL.Path, "/") + "/caixas"
	query := pathURL.Query()
	query.Set("pagina", strconv.Itoa(filter.Page))
	setOptionalQuery(query, "dataInicial", filter.DateFrom)
	setOptionalQuery(query, "dataFinal", filter.DateTo)
	for _, categoryID := range filter.CategoryIDs {
		query.Add("idsCategorias", strconv.FormatInt(categoryID, 10))
	}
	pathURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pathURL.String(), nil)
	if err != nil {
		return BlingReceivablesPage{}, fmt.Errorf("%w: create request", ErrBlingAPIConfiguration)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("enable-jwt", "1")
	return c.doFinancialResourceRequestWithoutPageLimit(req, ReceivablesFilter{Page: filter.Page, Limit: filter.MaxRecords})
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
		return nil, newBlingTransportFailure(err)
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
