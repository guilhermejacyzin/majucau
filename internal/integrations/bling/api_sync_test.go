package bling

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestBlingAPINormalizationFailureCodeDoesNotExposeErrorText(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "category permission", err: ErrBlingFinancialCategoryScopeInsufficient, want: "BLING_CATEGORY_SCOPE_INSUFFICIENT"},
		{name: "resource permission", err: ErrBlingAPIScopeInsufficient, want: "BLING_API_SCOPE_INSUFFICIENT"},
		{name: "authorization", err: ErrBlingAPIUnauthorized, want: "BLING_API_UNAUTHORIZED"},
		{name: "invalid response", err: ErrBlingAPIInvalidResponse, want: "BLING_API_INVALID_RESPONSE"},
		{name: "configuration", err: ErrBlingAPIConfiguration, want: "BLING_API_CONFIGURATION_INVALID"},
		{name: "payable due date field", err: errors.Join(errBlingPayableDetailSchemaMismatch, payableSchemaField("BLING_PAYABLE_DUE_DATE_INVALID")), want: "BLING_PAYABLE_DUE_DATE_INVALID"},
		{name: "payable amount field", err: errors.Join(errBlingPayableDetailSchemaMismatch, payableSchemaField("BLING_PAYABLE_AMOUNT_INVALID")), want: "BLING_PAYABLE_AMOUNT_INVALID"},
		{name: "contact name field", err: errors.Join(errBlingContactDetailSchemaMismatch, payableSchemaField("BLING_CONTACT_DETAIL_NAME_INVALID")), want: "BLING_CONTACT_DETAIL_NAME_INVALID"},
		{name: "category type field", err: errors.Join(errBlingCategoryDetailSchemaMismatch, payableSchemaField("BLING_CATEGORY_DETAIL_TYPE_INVALID")), want: "BLING_CATEGORY_DETAIL_TYPE_INVALID"},
		{name: "payable list schema", err: errors.Join(errBlingPayableListSchemaMismatch, ErrBlingAPISchemaMismatch), want: "BLING_PAYABLE_LIST_SCHEMA_MISMATCH"},
		{name: "payable detail schema", err: errors.Join(errBlingPayableDetailSchemaMismatch, ErrBlingAPISchemaMismatch), want: "BLING_PAYABLE_DETAIL_SCHEMA_MISMATCH"},
		{name: "contact detail schema", err: errors.Join(errBlingContactDetailSchemaMismatch, ErrBlingAPISchemaMismatch), want: "BLING_CONTACT_DETAIL_SCHEMA_MISMATCH"},
		{name: "category detail schema", err: errors.Join(errBlingCategoryDetailSchemaMismatch, ErrBlingAPISchemaMismatch), want: "BLING_CATEGORY_DETAIL_SCHEMA_MISMATCH"},
		{name: "payment method detail schema", err: errors.Join(errBlingPaymentMethodSchemaMismatch, ErrBlingAPISchemaMismatch), want: "BLING_PAYMENT_METHOD_DETAIL_SCHEMA_MISMATCH"},
		{name: "schema", err: ErrBlingAPISchemaMismatch, want: "BLING_API_SCHEMA_MISMATCH"},
		{name: "missing detail", err: &BlingAPIError{StatusCode: http.StatusNotFound, Code: "provider detail"}, want: "BLING_API_RECORD_NOT_FOUND"},
		{name: "rejected request", err: &BlingAPIError{StatusCode: http.StatusBadRequest, Code: "provider detail"}, want: "BLING_API_REQUEST_REJECTED"},
		{name: "other provider response", err: &BlingAPIError{StatusCode: http.StatusConflict, Code: "sensitive provider payload"}, want: "BLING_API_RESPONSE_REJECTED"},
		{name: "duplicate database row", err: &pgconn.PgError{Code: "23505", Message: "sensitive record details"}, want: "BLING_NORMALIZATION_DUPLICATE"},
		{name: "database constraint", err: &pgconn.PgError{Code: "23514", Message: "sensitive constraint detail"}, want: "BLING_NORMALIZATION_CONSTRAINT_FAILED"},
		{name: "other database state", err: &pgconn.PgError{Code: "22007", Message: "sensitive date detail"}, want: "BLING_NORMALIZATION_STORAGE_22007"},
		{name: "unsafe database state", err: &pgconn.PgError{Code: "bad-state", Message: "sensitive detail"}, want: "BLING_NORMALIZATION_STORAGE_FAILED"},
		{name: "unknown", err: errors.New("sensitive provider payload"), want: "BLING_API_NORMALIZATION_FAILED"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := blingAPINormalizationFailureCode(test.err); got != test.want {
				t.Fatalf("normalization failure code = %q, want %q", got, test.want)
			}
		})
	}
}

func TestPreparePaymentMethodPageKeepsRawListAndDetailAndCachesWithinPage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/formas-pagamentos/11" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"id":11,"destino":1,"taxas":{"aliquota":2.5,"valor":1.25,"prazo":30}}}`))
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	batchID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	prepare := preparePaymentMethodAPISnapshotPage(client)
	prepared, err := prepare(context.Background(), batchID, []json.RawMessage{
		json.RawMessage(`{"id":11,"descricao":"Cartao"}`),
		json.RawMessage(`{"id":11,"descricao":"Cartao"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(prepared) != 2 {
		t.Fatalf("calls=%d prepared=%d, want one detail read and two preserved list rows", calls, len(prepared))
	}
	for _, row := range prepared {
		var decoded preparedPaymentMethodAPIRecord
		if err := json.Unmarshal(row, &decoded); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(decoded.List), `"id":11`) || !strings.Contains(string(decoded.Detail), `"destino":1`) {
			t.Fatalf("prepared row lost raw list/detail: %s", row)
		}
	}
}

func TestMarshalReceivablesPagePreservesUnhomologatedPayload(t *testing.T) {
	total := 3
	payload, err := marshalReceivablesPage(BlingReceivablesPage{
		Records: []json.RawMessage{json.RawMessage(`{"id":42,"situacao":"pago"}`)},
		Page:    1, Limit: 1, Total: &total, HasNext: true, HasNextKnown: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded receivablesPagePayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Page != 1 || decoded.Limit != 1 || decoded.Total == nil || *decoded.Total != 3 || !decoded.HasNext || len(decoded.Records) != 1 {
		t.Fatalf("decoded payload = %+v", decoded)
	}
	if !strings.Contains(string(decoded.Records[0]), `"id":42`) {
		t.Fatalf("raw record changed: %s", decoded.Records[0])
	}
}

func TestReceivablesPageSourceIdentityIsStable(t *testing.T) {
	const expected = "pagina:7"
	if got := receivablesPageSourceID(7); got != expected {
		t.Fatalf("source id = %q, want %q", got, expected)
	}
}

func TestBlingAPIResourceSyncLockKeyIsScoped(t *testing.T) {
	connection := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	otherConnection := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	key := blingAPIResourceSyncLockKey(connection, ReceivablesAPIResource)
	if got := blingAPIResourceSyncLockKey(connection, ReceivablesAPIResource); got != key {
		t.Fatalf("same resource lock key = %d, want stable key %d", got, key)
	}
	if got := blingAPIResourceSyncLockKey(connection, PayablesAPIResource); got == key {
		t.Fatalf("different resource reused lock key %d", key)
	}
	if got := blingAPIResourceSyncLockKey(otherConnection, ReceivablesAPIResource); got == key {
		t.Fatalf("different connection reused lock key %d", key)
	}
}

func TestCashMovementsAPISyncServicePreservesPagedRAWContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/caixas" {
			t.Fatalf("request = %s %s, want GET /caixas", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("pagina") != "3" {
			t.Fatalf("page = %q, want 3", query.Get("pagina"))
		}
		if query.Get("dataInicial") != "2026-01-01" || query.Get("dataFinal") != "2026-10-18" {
			t.Fatalf("date range = %q..%q, want 2026-01-01..2026-10-18", query.Get("dataInicial"), query.Get("dataFinal"))
		}
		if r.URL.Query().Get("limite") != "100" {
			t.Fatalf("cash movement page size = %q, want 100", r.URL.Query().Get("limite"))
		}
		if r.Header.Get("enable-jwt") != "1" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("cash movement request did not preserve JWT authorization headers")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":91}],"pagination":{"page":3,"hasNext":false}}`))
	}))
	defer server.Close()

	client, err := NewBlingAPIClient(server.Client(), server.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	service := newCashMovementsAPISyncService(nil, client, CashMovementsFilter{DateFrom: "2026-01-01", DateTo: "2026-10-18"})
	if service.resource != CashMovementsAPIResource || service.sourceEntity != "bling.caixas.page" {
		t.Fatalf("resource/source = %q/%q", service.resource, service.sourceEntity)
	}
	page, err := service.list(context.Background(), ReceivablesFilter{Page: 3, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if page.Page != 3 || !page.HasNextKnown || page.HasNext || len(page.Records) != 1 || string(page.Records[0]) != `{"id":91}` {
		t.Fatalf("page = %+v", page)
	}
}

type fakeStaleRawPageStore struct {
	ids         []pgtype.UUID
	retired     map[pgtype.UUID]bool
	requested   []int
	retireErrAt int
	retireCalls int
	afterRetire func()
}

func (s *fakeStaleRawPageStore) ListCurrentStalePageIDs(_ context.Context, _ pgtype.UUID, _ string, _ int, after pgtype.UUID, limit int) ([]pgtype.UUID, error) {
	s.requested = append(s.requested, limit)
	batch := make([]pgtype.UUID, 0, limit)
	for _, id := range s.ids {
		if s.retired[id] || (after.Valid && bytes.Compare(id.Bytes[:], after.Bytes[:]) <= 0) {
			continue
		}
		batch = append(batch, id)
		if len(batch) == limit {
			break
		}
	}
	return batch, nil
}

func (s *fakeStaleRawPageStore) RetireCurrentRawPage(_ context.Context, id pgtype.UUID) error {
	s.retireCalls++
	if s.retireErrAt > 0 && s.retireCalls == s.retireErrAt {
		return errors.New("injected retirement failure")
	}
	s.retired[id] = true
	if s.afterRetire != nil {
		s.afterRetire()
	}
	return nil
}

func TestRetireCurrentResourcePagesUsesBoundedBatches(t *testing.T) {
	store := &fakeStaleRawPageStore{retired: make(map[pgtype.UUID]bool)}
	for i := 1; i <= staleRawPageBatchSize*2+17; i++ {
		var id pgtype.UUID
		id.Bytes[14] = byte(i >> 8)
		id.Bytes[15] = byte(i)
		id.Valid = true
		store.ids = append(store.ids, id)
	}

	if err := retireCurrentResourcePagesInBatches(context.Background(), store, pgtype.UUID{Valid: true}, "bling.contas_receber.page", 2); err != nil {
		t.Fatalf("retire stale pages: %v", err)
	}
	if len(store.retired) != len(store.ids) {
		t.Fatalf("retired %d IDs, want %d", len(store.retired), len(store.ids))
	}
	if len(store.requested) != 3 {
		t.Fatalf("queries = %d, want 3 bounded queries", len(store.requested))
	}
	for _, requested := range store.requested {
		if requested != staleRawPageBatchSize {
			t.Fatalf("requested batch size = %d, want %d", requested, staleRawPageBatchSize)
		}
	}
}

func TestRetireCurrentResourcePagesStopsOnFailure(t *testing.T) {
	store := &fakeStaleRawPageStore{retired: make(map[pgtype.UUID]bool), retireErrAt: 2}
	for i := 1; i <= 3; i++ {
		var id pgtype.UUID
		id.Bytes[15] = byte(i)
		id.Valid = true
		store.ids = append(store.ids, id)
	}

	err := retireCurrentResourcePagesInBatches(context.Background(), store, pgtype.UUID{Valid: true}, "bling.contas_receber.page", 1)
	if err == nil || !strings.Contains(err.Error(), "injected retirement failure") {
		t.Fatalf("retirement error = %v, want injected failure", err)
	}
	if store.retireCalls != 2 || len(store.retired) != 1 {
		t.Fatalf("retire calls/successes = %d/%d, want 2/1", store.retireCalls, len(store.retired))
	}
}

func TestRetireCurrentResourcePagesStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := &fakeStaleRawPageStore{retired: make(map[pgtype.UUID]bool)}
	store.afterRetire = cancel
	for i := 1; i <= 3; i++ {
		var id pgtype.UUID
		id.Bytes[15] = byte(i)
		id.Valid = true
		store.ids = append(store.ids, id)
	}

	err := retireCurrentResourcePagesInBatches(ctx, store, pgtype.UUID{Valid: true}, "bling.contas_receber.page", 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("retirement error = %v, want context cancellation", err)
	}
	if store.retireCalls != 1 || len(store.retired) != 1 {
		t.Fatalf("retire calls/successes = %d/%d, want 1/1", store.retireCalls, len(store.retired))
	}
}
