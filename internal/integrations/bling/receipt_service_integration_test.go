package bling

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestReceiptImportServicePostgresE2E is opt-in so the normal unit suite never
// requires a local database. It is used by release validation with a disposable
// PostgreSQL instance and a sanitized Bling report folder.
func TestReceiptImportServicePostgresE2E(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	folder := os.Getenv("MAJUCAU_TEST_RECEIPTS_FOLDER")
	if dsn == "" || folder == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL and MAJUCAU_TEST_RECEIPTS_FOLDER to run the PostgreSQL e2e")
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

	service := NewReceiptImportService(pool)
	result, err := service.Import(ctx, folder)
	if err != nil {
		t.Fatalf("import sanitized Bling folder: %v", err)
	}
	if result.Status != "PARTIAL" || result.RecordsRead != 3 || result.RecordsFailed != 1 || result.IgnoredCount != 0 {
		t.Fatalf("unexpected import result: %+v", result)
	}
	if result.RecordsCreated+result.RecordsUpdated != 2 {
		t.Fatalf("expected two paid receipts to be created or updated: %+v", result)
	}

	var persisted int
	if err := pool.QueryRow(ctx, `
		SELECT count(*)
		FROM receipts
		WHERE source_system = 'BLING'
		  AND source_entity = $1
		  AND source_id IN ('E2E-20260920-001', 'E2E-20260920-002')`, ReceiptsReportSourceEntity).Scan(&persisted); err != nil {
		t.Fatalf("count persisted receipts: %v", err)
	}
	if persisted != 2 {
		t.Fatalf("expected two persisted receipts, got %d", persisted)
	}
}
