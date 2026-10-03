package bling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
)

func TestAPIResourceSyncKeepsFailureAuditAndCommitsCursorWithRAW(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL API sync integration test")
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
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'api-sync-atomicity-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}

	const resource = "ci.bling.api_sync.atomicity"
	const sourceEntity = "ci.bling.api_sync.atomicity.page"
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM raw_records WHERE source_entity = $1`, sourceEntity)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND resource = $2`, connection.ID, resource)
	}
	cleanup()
	defer cleanup()

	var priorBatchID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp())
		RETURNING id::text`, connection.ID, resource).Scan(&priorBatchID); err != nil {
		t.Fatalf("create prior successful batch: %v", err)
	}
	previousPayload := []byte(`{"page":"previous-valid"}`)
	previousDigest := sha256.Sum256(previousPayload)
	previousHash := hex.EncodeToString(previousDigest[:])
	if _, err := pool.Exec(ctx, `
		INSERT INTO raw_records (sync_batch_id, connection_id, source_system, source_entity, source_id, payload_hash, payload, is_current)
		VALUES ($1::uuid, $2, 'BLING', $3, 'pagina:1', $4, $5::jsonb, true)`, priorBatchID, connection.ID, sourceEntity, previousHash, previousPayload); err != nil {
		t.Fatalf("seed latest valid RAW page: %v", err)
	}
	const previousCursor = "checkpoint:previous-success"
	if _, err := pool.Exec(ctx, `
		INSERT INTO sync_cursors (connection_id, resource, cursor_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (connection_id, resource) DO UPDATE SET cursor_value = EXCLUDED.cursor_value`, connection.ID, resource, previousCursor); err != nil {
		t.Fatalf("seed previous cursor: %v", err)
	}

	pageOne := BlingReceivablesPage{Page: 1, Limit: 1, HasNext: true, Records: []json.RawMessage{json.RawMessage(`{"id":"page-one-new"}`)}}
	pageTwo := BlingReceivablesPage{Page: 2, Limit: 1, HasNext: false, Records: []json.RawMessage{json.RawMessage(`{"id":"page-two-new"}`)}}
	pageTwoFailure := errors.New("injected page read failure")
	failSecondPage := true
	service := &apiResourceSyncService{
		pool:         pool,
		client:       &BlingAPIClient{},
		maxPages:     3,
		resource:     resource,
		sourceEntity: sourceEntity,
		list: func(_ context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			switch filter.Page {
			case 1:
				return pageOne, nil
			case 2:
				if failSecondPage {
					return BlingReceivablesPage{}, pageTwoFailure
				}
				return pageTwo, nil
			default:
				return BlingReceivablesPage{}, errors.New("unexpected page request")
			}
		},
	}

	failed, err := service.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if !errors.Is(err, pageTwoFailure) {
		t.Fatalf("failed page sync error = %v, want injected read failure", err)
	}
	if failed.Status != "FAILED" || failed.BatchID == "" {
		t.Fatalf("failed sync result = %+v", failed)
	}
	var failedStatus, failedCode string
	var failedRecordsRead int
	if err := pool.QueryRow(ctx, `
		SELECT status, COALESCE(error_code, ''), records_read
		FROM integration_sync_batches WHERE id = $1::uuid`, failed.BatchID).Scan(&failedStatus, &failedCode, &failedRecordsRead); err != nil {
		t.Fatalf("read persisted failed batch: %v", err)
	}
	if failedStatus != "FAILED" || failedCode != "BLING_API_READ_FAILED" || failedRecordsRead != 1 {
		t.Fatalf("failed batch status/code/records = %q/%q/%d", failedStatus, failedCode, failedRecordsRead)
	}
	var currentHash, cursor string
	if err := pool.QueryRow(ctx, `
		SELECT payload_hash FROM raw_records
		WHERE connection_id = $1 AND source_entity = $2 AND source_id = 'pagina:1' AND is_current`, connection.ID, sourceEntity).Scan(&currentHash); err != nil {
		t.Fatalf("read preserved current RAW page: %v", err)
	}
	if currentHash != previousHash {
		t.Fatalf("failed sync replaced the latest valid RAW hash: got %s, want %s", currentHash, previousHash)
	}
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource).Scan(&cursor); err != nil {
		t.Fatalf("read cursor after failed sync: %v", err)
	}
	if cursor != previousCursor {
		t.Fatalf("failed sync advanced cursor to %q, want %q", cursor, previousCursor)
	}
	var rawCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&rawCount); err != nil {
		t.Fatalf("count RAW pages after failed sync: %v", err)
	}
	if rawCount != 1 {
		t.Fatalf("failed sync retained partial RAW pages: got %d rows, want 1", rawCount)
	}

	failSecondPage = false
	succeeded, err := service.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("successful paginated sync: %v", err)
	}
	if succeeded.Status != "SUCCESS" || succeeded.PagesRead != 2 {
		t.Fatalf("successful sync result = %+v", succeeded)
	}
	if err := pool.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE connection_id = $1 AND resource = $2`, connection.ID, resource).Scan(&cursor); err != nil {
		t.Fatalf("read cursor after successful sync: %v", err)
	}
	if cursor != "page:2" {
		t.Fatalf("successful sync cursor = %q, want page:2", cursor)
	}
	var currentPages, allPages int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE is_current), count(*)
		FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&currentPages, &allPages); err != nil {
		t.Fatalf("count RAW pages after successful sync: %v", err)
	}
	if currentPages != 2 || allPages != 3 {
		t.Fatalf("successful sync RAW current/total counts = %d/%d, want 2/3", currentPages, allPages)
	}

	replayed, err := service.Sync(ctx, ReceivablesFilter{Page: 1, Limit: 1})
	if err != nil {
		t.Fatalf("idempotent API sync replay: %v", err)
	}
	if replayed.Status != "SUCCESS" || replayed.PagesRead != 2 || replayed.PagesCreated != 0 || replayed.PagesUpdated != 0 {
		t.Fatalf("idempotent replay result = %+v", replayed)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE is_current), count(*)
		FROM raw_records WHERE connection_id = $1 AND source_entity = $2`, connection.ID, sourceEntity).Scan(&currentPages, &allPages); err != nil {
		t.Fatalf("count RAW pages after idempotent replay: %v", err)
	}
	if currentPages != 2 || allPages != 3 {
		t.Fatalf("idempotent replay changed RAW current/total counts to %d/%d, want 2/3", currentPages, allPages)
	}
}
