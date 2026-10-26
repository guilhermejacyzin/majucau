package bling

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

func TestBlingDailySyncSkipsACompletedScheduledLoad(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL daily sync integration test")
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
		VALUES ('BLING', 'CONNECTED', 'daily-sync-idempotency-ci')
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
	scheduledAt := time.Date(uniqueDate.Year(), uniqueDate.Month(), uniqueDate.Day(), 9, 30, 0, 0, location)
	correlationID, err := dailyBlingCorrelationID(scheduledAt)
	if err != nil {
		t.Fatal(err)
	}
	correlation := pgUUID(correlationID)
	if _, err := pool.Exec(ctx, `DELETE FROM integration_sync_batches WHERE correlation_id = $1`, correlation); err != nil {
		t.Fatalf("clear previous fixture batches: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE correlation_id = $1`, correlation)
	}()
	for _, resource := range []string{ReceivablesAPIResource, PayablesAPIResource} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id)
			VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3)`, connection.ID, resource, correlation); err != nil {
			t.Fatalf("seed completed %s batch: %v", resource, err)
		}
	}

	service := &BlingOAuthService{pool: pool}
	result, skipped, err := service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("incomplete two-resource run was accepted: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3)`, connection.ID, CashMovementsAPIResource, correlation); err != nil {
		t.Fatalf("seed completed %s batch: %v", CashMovementsAPIResource, err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		result, skipped, err = service.SyncDaily(ctx, scheduledAt)
		if err != nil || !skipped || result.Status != "SUCCESS" {
			t.Fatalf("attempt %d: result=%+v skipped=%v error=%v", attempt+1, result, skipped, err)
		}
	}
	var batchCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM integration_sync_batches WHERE correlation_id = $1`, correlation).Scan(&batchCount); err != nil {
		t.Fatalf("count daily sync batches: %v", err)
	}
	if batchCount != 3 {
		t.Fatalf("restarted daily load created duplicate batches: got %d, want 3", batchCount)
	}
}
