package nuvempago

import (
	"context"
	"os"
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
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('NUVEM_PAGO', 'CONNECTED', 'e2e-nuvem-pago')
		ON CONFLICT (provider) DO UPDATE
		SET status = EXCLUDED.status, external_account_id = EXCLUDED.external_account_id,
		    updated_at = clock_timestamp()`); err != nil {
		t.Fatalf("prepare Nuvem Pago connection: %v", err)
	}

	service := NewFutureImportService(pool)
	first, err := service.Import(ctx, folder)
	if err != nil {
		t.Fatalf("first future import: %v", err)
	}
	if first.Status != "PARTIAL" || first.RecordsRead != 3 || first.RecordsFailed != 1 || first.RecordsCreated != 2 || first.RecordsUpdated != 0 {
		t.Fatalf("unexpected first import result: %+v", first)
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

