package bling

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

func TestDailyPayablesAPIIsNormalizedAtomicallyAndKeepsLastSnapshotOnFailure(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL payable API integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	var databaseName string
	if err := pool.QueryRow(ctx, `SELECT current_database()`).Scan(&databaseName); err != nil {
		t.Fatalf("identify PostgreSQL database: %v", err)
	}
	if databaseName != "majucau_test" {
		t.Skip("payable API integration test only runs against the dedicated majucau_test database")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'api-payables-normalizer-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}
	firstRunAt := time.Date(2026, 10, 19, 12, 30, 0, 0, time.UTC)
	firstCorrelation, err := dailyBlingCorrelationID(firstRunAt)
	if err != nil {
		t.Fatal(err)
	}
	secondCorrelation, err := dailyBlingCorrelationID(firstRunAt.Add(24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	const sourceID = "900000000001"
	const contactSourceID = "900000000002"
	const categorySourceID = "900000000003"
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4)`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation))
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payable_sync_snapshot_rows WHERE sync_batch_id IN (SELECT id FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4))`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation))
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payable_sources WHERE payable_id IN (SELECT id FROM payables WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = $2 AND source_id = $3)`, connection.ID, payableAPISourceEntity, sourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payables WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = $2 AND source_id = $3`, connection.ID, payableAPISourceEntity, sourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM contacts WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = 'contatos' AND source_id = $2`, connection.ID, contactSourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_records WHERE connection_id = $1 AND source_system = 'BLING' AND ((source_entity = 'bling.contas_pagar.page' AND source_id = 'pagina:1') OR (source_entity = 'bling.conta_pagar' AND source_id = $2) OR (source_entity = 'bling.contato' AND source_id = $3) OR (source_entity = 'bling.categoria_receita_despesa' AND source_id = $4))`, connection.ID, sourceID, contactSourceID, categorySourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, PayablesAPIResource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4)`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation))
	}
	cleanup()
	defer cleanup()

	failCategory := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("enable-jwt") != "1" {
			t.Errorf("unexpected method/header for %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/Api/v3/contas/pagar":
			if r.URL.Query().Get("pagina") != "1" || r.URL.Query().Get("limite") != "100" {
				t.Errorf("payables list query = %q", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"data":[{"id":900000000001}]}`))
		case "/Api/v3/contas/pagar/900000000001":
			_, _ = w.Write([]byte(`{"data":{"id":900000000001,"situacao":3,"vencimento":"2026-10-31","valor":250.0000,"saldo":150.2500,"competencia":"2026-10-01","numeroDocumento":"NF-21","categoria":{"id":900000000003},"contato":{"id":900000000002}}}`))
		case "/Api/v3/contatos/900000000002":
			_, _ = w.Write([]byte(`{"data":{"id":900000000002,"nome":"Fornecedor de teste","situacao":"A","numeroDocumento":"00.000.000/0000-00"}}`))
		case "/Api/v3/categorias/receitas-despesas/900000000003":
			if failCategory {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":{"type":"FORBIDDEN"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":900000000003,"descricao":"CAPEX teste","tipo":1,"situacao":1}}`))
		default:
			t.Errorf("unexpected Bling API path %q", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewBlingAPIClient(server.Client(), server.URL+"/Api/v3", "fake-access-token")
	if err != nil {
		t.Fatal(err)
	}
	service := NewPayablesAPISyncService(pool, client)
	first, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(firstCorrelation))
	if err != nil {
		t.Fatalf("first daily payable sync failed: %v", err)
	}
	if first.Status != "SUCCESS" {
		t.Fatalf("first sync status = %q", first.Status)
	}
	var stage, status string
	if err := pool.QueryRow(ctx, `SELECT data_stage, status FROM integration_sync_batches WHERE id = $1::uuid`, first.BatchID).Scan(&stage, &status); err != nil {
		t.Fatalf("read first batch: %v", err)
	}
	if stage != "NORMALIZED" || status != "SUCCESS" {
		t.Fatalf("first batch stage/status = %s/%s", stage, status)
	}
	var payableCount int
	var dueDate, statusCode, document, category, supplier string
	var originalAmount, openBalance string
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*), MIN(due_date)::text, MIN(status), MIN(document), MIN(category),
		       MIN(original_amount)::text, MIN(open_balance)::text,
		       MIN(c.name)
		FROM payables p LEFT JOIN contacts c ON c.id = p.supplier_id
		WHERE p.connection_id = $1 AND p.source_system = 'BLING' AND p.source_entity = $2 AND p.source_id = $3`, connection.ID, payableAPISourceEntity, sourceID).Scan(&payableCount, &dueDate, &statusCode, &document, &category, &originalAmount, &openBalance, &supplier); err != nil {
		t.Fatalf("read normalized payable: %v", err)
	}
	if payableCount != 1 || dueDate != "2026-10-31" || statusCode != "3" || document != "NF-21" || category != "CAPEX teste" || originalAmount != "250.0000" || openBalance != "150.2500" || supplier != "Fornecedor de teste" {
		t.Fatalf("normalized payable = count:%d due:%s status:%s document:%s category:%s original:%s balance:%s supplier:%s", payableCount, dueDate, statusCode, document, category, originalAmount, openBalance, supplier)
	}
	var lineageCount, snapshotCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sources WHERE payable_id IN (SELECT id FROM payables WHERE connection_id = $1 AND source_entity = $2 AND source_id = $3)`, connection.ID, payableAPISourceEntity, sourceID).Scan(&lineageCount); err != nil {
		t.Fatalf("count payable lineage: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, first.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("count payable snapshot membership: %v", err)
	}
	if lineageCount != 3 || snapshotCount != 1 {
		t.Fatalf("lineage/snapshot counts = %d/%d, want 3/1", lineageCount, snapshotCount)
	}
	var rawContactDocument string
	if err := pool.QueryRow(ctx, `SELECT payload->>'numeroDocumento' FROM raw_records WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = 'bling.contato' AND source_id = $2 AND is_current`, connection.ID, contactSourceID).Scan(&rawContactDocument); err != nil {
		t.Fatalf("read current contact RAW data: %v", err)
	}
	if rawContactDocument != "00.000.000/0000-00" {
		t.Fatalf("contact RAW payload was not preserved exactly: %q", rawContactDocument)
	}

	failCategory = true
	failed, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(secondCorrelation))
	if !errors.Is(err, ErrBlingAPIUnauthorized) || failed.Status != "FAILED" {
		t.Fatalf("failed sync result/error = %+v/%v, want failed authorization", failed, err)
	}
	if err := pool.QueryRow(ctx, `SELECT data_stage, status FROM integration_sync_batches WHERE id = $1::uuid`, failed.BatchID).Scan(&stage, &status); err != nil {
		t.Fatalf("read failed batch: %v", err)
	}
	if stage != "RAW_ONLY" || status != "FAILED" {
		t.Fatalf("failed batch stage/status = %s/%s", stage, status)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, failed.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("count failed snapshot members: %v", err)
	}
	if snapshotCount != 0 {
		t.Fatalf("failed batch published %d snapshot rows", snapshotCount)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, first.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("recheck previous snapshot membership: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("previous valid snapshot membership = %d, want 1 after failure", snapshotCount)
	}
	var cursor string
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, PayablesAPIResource).Scan(&cursor); err != nil {
		t.Fatalf("read payable cursor: %v", err)
	}
	if cursor != "page:1" {
		t.Fatalf("cursor after failed refresh = %q, want previous page:1", cursor)
	}
}
