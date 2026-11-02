package bling

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

func TestDailyPayablesAPIIsNormalizedAtomicallyWithCategoryScopeGap(t *testing.T) {
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
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	dateOffset := int(time.Now().UnixNano() % int64(36500))
	uniqueDate := time.Date(2100, time.January, 1, 0, 0, 0, 0, location).AddDate(0, 0, dateOffset)
	firstRunAt := time.Date(uniqueDate.Year(), uniqueDate.Month(), uniqueDate.Day(), 9, 30, 0, 0, location)
	firstCorrelation, err := dailyBlingCorrelationID(firstRunAt)
	if err != nil {
		t.Fatal(err)
	}
	secondCorrelation, err := dailyBlingCorrelationID(firstRunAt.Add(24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	thirdCorrelation, err := dailyBlingCorrelationID(firstRunAt.Add(48 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	fourthCorrelation, err := dailyBlingCorrelationID(firstRunAt.Add(72 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	baseSourceID := time.Now().UnixNano()
	sourceID := strconv.FormatInt(baseSourceID, 10)
	contactSourceID := strconv.FormatInt(baseSourceID+1, 10)
	categorySourceID := strconv.FormatInt(baseSourceID+2, 10)
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4, $5, $6)`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation), pgUUID(thirdCorrelation), pgUUID(fourthCorrelation))
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payable_sync_snapshot_rows WHERE sync_batch_id IN (SELECT id FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4, $5, $6))`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation), pgUUID(thirdCorrelation), pgUUID(fourthCorrelation))
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payable_sources WHERE payable_id IN (SELECT id FROM payables WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = $2 AND source_id = $3)`, connection.ID, payableAPISourceEntity, sourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM payables WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = $2 AND source_id = $3`, connection.ID, payableAPISourceEntity, sourceID)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM contacts WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = 'contatos' AND source_id = $2`, connection.ID, contactSourceID)
		// RAW rows are append-only. Unique synthetic IDs keep each test run isolated.
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, PayablesAPIResource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2 AND correlation_id IN ($3, $4, $5, $6)`, connection.ID, PayablesAPIResource, pgUUID(firstCorrelation), pgUUID(secondCorrelation), pgUUID(thirdCorrelation), pgUUID(fourthCorrelation))
	}
	cleanup()
	defer cleanup()

	failCategory := false
	failPayableDetail := false
	unassignedCategory := false
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
			_, _ = w.Write([]byte(`{"data":[{"id":` + sourceID + `}]}`))
		case "/Api/v3/contas/pagar/" + sourceID:
			if failPayableDetail {
				_, _ = w.Write([]byte(`{"data":{"id":` + sourceID + `}}`))
				return
			}
			categoryID := categorySourceID
			if unassignedCategory {
				categoryID = "0"
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%s,"situacao":3,"vencimento":"2026-10-31","valor":250.0000,"saldo":150.2500,"competencia":"2026-10-01","numeroDocumento":"NF-21","categoria":{"id":%s},"contato":{"id":%s}}}`, sourceID, categoryID, contactSourceID)))
		case "/Api/v3/contatos/" + contactSourceID:
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%s,"nome":"Fornecedor de teste","situacao":"A","numeroDocumento":"00.000.000/0000-00"}}`, contactSourceID)))
		case "/Api/v3/categorias/receitas-despesas/" + categorySourceID:
			if failCategory {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":{"type":"FORBIDDEN"}}`))
				return
			}
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":{"id":%s,"descricao":"CAPEX teste","tipo":1,"situacao":1}}`, categorySourceID)))
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
		t.Fatalf("first daily payable sync failed (%s): %v", blingAPINormalizationFailureCode(err), err)
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
		SELECT COUNT(*), MIN(p.due_date)::text, MIN(p.status), MIN(p.document), COALESCE(MIN(p.category), ''),
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
	partial, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(secondCorrelation))
	if err != nil || partial.Status != "PARTIAL" {
		t.Fatalf("category-scope gap should publish financial fields as partial: result=%+v error=%v", partial, err)
	}
	var batchErrorCode string
	if err := pool.QueryRow(ctx, `SELECT data_stage, status, error_code FROM integration_sync_batches WHERE id = $1::uuid`, partial.BatchID).Scan(&stage, &status, &batchErrorCode); err != nil {
		t.Fatalf("read partial batch: %v", err)
	}
	if stage != "NORMALIZED" || status != "PARTIAL" || batchErrorCode != "BLING_CATEGORY_SCOPE_INSUFFICIENT" {
		t.Fatalf("partial batch stage/status/code = %s/%s/%s", stage, status, batchErrorCode)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, partial.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("count partial snapshot members: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("partial batch published %d snapshot rows, want 1 with its validated amount and due date", snapshotCount)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, first.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("check previous snapshot membership: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("previous valid snapshot membership = %d, want 1 after a later partial snapshot", snapshotCount)
	}
	unassignedCategory = true
	unassigned, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(thirdCorrelation))
	if err != nil || unassigned.Status != "PARTIAL" {
		t.Fatalf("unassigned category should preserve validated payable fields as partial: result=%+v error=%v", unassigned, err)
	}
	if err := pool.QueryRow(ctx, `SELECT data_stage, status, error_code FROM integration_sync_batches WHERE id = $1::uuid`, unassigned.BatchID).Scan(&stage, &status, &batchErrorCode); err != nil {
		t.Fatalf("read unassigned-category batch: %v", err)
	}
	if stage != "NORMALIZED" || status != "PARTIAL" || batchErrorCode != "BLING_CATEGORY_UNASSIGNED" {
		t.Fatalf("unassigned-category batch stage/status/code = %s/%s/%s", stage, status, batchErrorCode)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, unassigned.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("count unassigned-category snapshot members: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("unassigned-category batch published %d rows, want 1 with validated financial fields", snapshotCount)
	}
	var unassignedCategoryValue string
	if err := pool.QueryRow(ctx, `SELECT COALESCE(category, '') FROM payables WHERE connection_id = $1 AND source_system = 'BLING' AND source_entity = $2 AND source_id = $3`, connection.ID, payableAPISourceEntity, sourceID).Scan(&unassignedCategoryValue); err != nil {
		t.Fatalf("read unassigned payable category: %v", err)
	}
	if unassignedCategoryValue != "" {
		t.Fatalf("unassigned category was inferred as %q", unassignedCategoryValue)
	}
	failPayableDetail = true
	unassignedCategory = false
	failed, err := service.SyncWithCorrelationID(ctx, ReceivablesFilter{}, pgUUID(fourthCorrelation))
	if !errors.Is(err, ErrBlingAPISchemaMismatch) || failed.Status != "FAILED" {
		t.Fatalf("invalid financial fields should fail closed: result=%+v error=%v", failed, err)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, failed.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("count failed snapshot members: %v", err)
	}
	if snapshotCount != 0 {
		t.Fatalf("failed batch published %d rows", snapshotCount)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM payable_sync_snapshot_rows WHERE sync_batch_id = $1::uuid`, partial.BatchID).Scan(&snapshotCount); err != nil {
		t.Fatalf("recheck last valid partial snapshot: %v", err)
	}
	if snapshotCount != 1 {
		t.Fatalf("last valid partial snapshot membership = %d after malformed input, want 1", snapshotCount)
	}
	var cursor string
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, PayablesAPIResource).Scan(&cursor); err != nil {
		t.Fatalf("read payable cursor: %v", err)
	}
	if cursor != "page:1" {
		t.Fatalf("cursor after failed refresh = %q, want previous page:1", cursor)
	}
}
