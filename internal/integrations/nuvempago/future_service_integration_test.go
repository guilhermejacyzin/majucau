package nuvempago

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestFutureImportServicePostgresE2E is opt-in. Release validation supplies a
// disposable PostgreSQL database with the migration and the sanitized fixture
// folder; the ordinary unit suite never requires a local database.
func TestFutureImportServicePostgresE2E(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	folder := os.Getenv("MAJUCAU_TEST_FUTURE_FOLDER")
	if dsn == "" || folder == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL and MAJUCAU_TEST_FUTURE_FOLDER to run the PostgreSQL e2e")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create PostgreSQL pool: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}
	var initialStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM integration_connections WHERE provider = 'NUVEM_PAGO'`).Scan(&initialStatus); err != nil {
		t.Fatalf("load migrated Nuvem Pago import connection: %v", err)
	}
	if initialStatus != "UNAVAILABLE" {
		t.Fatalf("initial Nuvem Pago status = %q, want UNAVAILABLE", initialStatus)
	}
	// Importing the approved Drive report must work while the live Nuvem Pago
	// integration remains unavailable for API confirmation.
	service := NewFutureImportService(pool)
	cleanupFailure := errors.New("staging cleanup failed")
	if _, err := service.ImportStaged(ctx, folder, FutureImportPending{}, func() error { return cleanupFailure }); !errors.Is(err, cleanupFailure) {
		t.Fatalf("staged import cleanup error = %v, want rollback", err)
	}
	assertFuturePersistence(t, ctx, pool, 0, 0, 0)

	first, err := service.ImportWithPending(ctx, folder, FutureImportPending{UnsupportedFiles: 1})
	if err != nil {
		t.Fatalf("first future import: %v", err)
	}
	if first.Status != "PARTIAL" || first.RecordsRead != 3 || first.RecordsFailed != 1 || first.RecordsCreated != 2 || first.RecordsUpdated != 0 || first.PendingFiles != 1 {
		t.Fatalf("unexpected first import result: %+v", first)
	}
	assertFuturePersistence(t, ctx, pool, 2, 2, 0)
	var batchStatus, errorCode, errorMessage string
	var recordsFailed int64
	if err := pool.QueryRow(ctx, `
		SELECT status, error_code, error_message_sanitized, records_failed
		FROM integration_sync_batches WHERE id::text = $1`, first.BatchID).Scan(&batchStatus, &errorCode, &errorMessage, &recordsFailed); err != nil {
		t.Fatalf("read partial future batch audit: %v", err)
	}
	if batchStatus != "PARTIAL" || errorCode != "NUVEM_PAGO_FUTURE_PARTIAL" || !strings.Contains(errorMessage, "linhas rejeitadas") || !strings.Contains(errorMessage, "arquivos pendentes") || recordsFailed != 1 {
		t.Fatalf("unexpected pending audit status=%q code=%q failed=%d message=%q", batchStatus, errorCode, recordsFailed, errorMessage)
	}

	pending, err := service.RecordPending(ctx, FutureImportPending{NoCompatibleCSV: true, UnsupportedFiles: 2})
	if err != nil {
		t.Fatalf("record missing compatible source: %v", err)
	}
	if pending.Status != "PARTIAL" || pending.RecordsRead != 0 || pending.RecordsCreated != 0 || pending.RecordsUpdated != 0 || pending.RecordsFailed != 0 || pending.PendingFiles != 2 || !pending.SourceMissing {
		t.Fatalf("unexpected pending-only result: %+v", pending)
	}
	assertFuturePersistence(t, ctx, pool, 2, 2, 0)

	second, err := service.Import(ctx, folder)
	if err != nil {
		t.Fatalf("idempotent future import: %v", err)
	}
	if second.Status != "PARTIAL" || second.RecordsRead != 3 || second.RecordsFailed != 1 || second.RecordsCreated != 0 || second.RecordsUpdated != 2 {
		t.Fatalf("unexpected idempotent import result: %+v", second)
	}
	assertFuturePersistence(t, ctx, pool, 2, 2, 0)
}

func assertFuturePersistence(t *testing.T, ctx context.Context, pool *pgxpool.Pool, receivables, currentRaw, receipts int) {
	t.Helper()
	var got int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM receivables
		WHERE source_system = 'NUVEM_PAGO' AND business_type = 'B2C'
		  AND status = 'PROJECTED'`).Scan(&got); err != nil {
		t.Fatalf("count projected receivables: %v", err)
	}
	if got != receivables {
		t.Fatalf("expected %d projected receivables, got %d", receivables, got)
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM raw_records
		WHERE source_system = 'NUVEM_PAGO' AND source_entity = $1 AND is_current`, FutureSourceEntity).Scan(&got); err != nil {
		t.Fatalf("count current future RAW: %v", err)
	}
	if got != currentRaw {
		t.Fatalf("expected %d current future RAW records, got %d", currentRaw, got)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM receipts WHERE source_system = 'BLING' AND source_entity = $1`, FutureSourceEntity).Scan(&got); err != nil {
		t.Fatalf("count future rows in receipts: %v", err)
	}
	if got != receipts {
		t.Fatalf("expected %d future rows in receipts, got %d", receipts, got)
	}
}
