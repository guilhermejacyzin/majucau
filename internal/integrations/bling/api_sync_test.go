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

	"github.com/jackc/pgx/v5/pgtype"
)

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
		if _, ok := r.URL.Query()["limite"]; ok {
			t.Fatal("cash movement request sent undocumented limite parameter")
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
