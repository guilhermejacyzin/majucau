package bling

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/scheduler"
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
			INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id, data_stage)
			VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3, 'RAW_ONLY')`, connection.ID, resource, correlation); err != nil {
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
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("RAW-only payable load was accepted as complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE correlation_id = $1 AND resource = $2`, correlation, PayablesAPIResource); err != nil {
		t.Fatalf("mark payable fixture as normalized: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE correlation_id = $1 AND resource = $2`, correlation, CashMovementsAPIResource); err != nil {
		t.Fatalf("mark cash movements fixture as normalized: %v", err)
	}
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("RAW-only receivables load was accepted as complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE correlation_id = $1 AND resource = $2`, correlation, ReceivablesAPIResource); err != nil {
		t.Fatalf("mark receivables fixture as normalized: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id, data_stage)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3, 'RAW_ONLY')`, connection.ID, PaymentMethodsAPIResource, correlation); err != nil {
		t.Fatalf("seed completed %s batch: %v", PaymentMethodsAPIResource, err)
	}
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("RAW-only payment-method settings were accepted as complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE correlation_id = $1 AND resource = $2`, correlation, PaymentMethodsAPIResource); err != nil {
		t.Fatalf("mark payment-method fixture as normalized: %v", err)
	}
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("missing financial-account list was accepted as complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id, data_stage)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3, 'RAW_ONLY')`, connection.ID, FinancialAccountsAPIResource, correlation); err != nil {
		t.Fatalf("seed completed %s batch: %v", FinancialAccountsAPIResource, err)
	}
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err == nil || skipped || result.Status == "SUCCESS" {
		t.Fatalf("RAW-only financial-account identities were accepted as complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE correlation_id = $1 AND resource = $2`, correlation, FinancialAccountsAPIResource); err != nil {
		t.Fatalf("mark financial-account fixture as normalized: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at, correlation_id, data_stage)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp(), $3, 'NORMALIZED')`, connection.ID, ReconciledCashMovementsAPIResource, correlation); err != nil {
		t.Fatalf("seed completed %s batch: %v", ReconciledCashMovementsAPIResource, err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		result, skipped, err = service.SyncDaily(ctx, scheduledAt)
		if err != nil || !skipped || result.Status != "SUCCESS" {
			t.Fatalf("attempt %d: result=%+v skipped=%v error=%v", attempt+1, result, skipped, err)
		}
	}
	if _, err := pool.Exec(ctx, `UPDATE integration_sync_batches SET status = 'PARTIAL' WHERE correlation_id = $1 AND resource = $2`, correlation, PayablesAPIResource); err != nil {
		t.Fatalf("mark normalized partial payable fixture: %v", err)
	}
	result, skipped, err = service.SyncDaily(ctx, scheduledAt)
	if err != nil || !skipped || result.Status != "SUCCESS" {
		t.Fatalf("normalized partial payable snapshot should remain idempotently complete: result=%+v skipped=%v error=%v", result, skipped, err)
	}
	var batchCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM integration_sync_batches WHERE correlation_id = $1`, correlation).Scan(&batchCount); err != nil {
		t.Fatalf("count daily sync batches: %v", err)
	}
	if batchCount != 6 {
		t.Fatalf("restarted daily load created duplicate batches: got %d, want 6", batchCount)
	}
}

func TestManualSyncResumesNewestUnpublishedPayablePages(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL resume integration test")
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
		t.Fatal(err)
	}
	if databaseName != "majucau_test" {
		t.Skip("resume integration test only runs against the dedicated majucau_test database")
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_connections (provider, status, external_account_id)
		VALUES ('BLING', 'CONNECTED', 'manual-sync-resume-ci')
		ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("prepare Bling connection: %v", err)
	}
	connection, err := database.New(pool).GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil || !connection.ID.Valid {
		t.Fatalf("load Bling connection: %v", err)
	}
	location, err := time.LoadLocation(scheduler.SaoPauloTimeZone)
	if err != nil {
		t.Fatal(err)
	}
	older, err := dailyBlingCorrelationID(time.Date(2100, 1, 2, 9, 30, 0, 0, location))
	if err != nil {
		t.Fatal(err)
	}
	newer, err := dailyBlingCorrelationID(time.Date(2100, 1, 3, 9, 30, 0, 0, location))
	if err != nil {
		t.Fatal(err)
	}
	olderID, newerID := pgUUID(older), pgUUID(newer)
	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource IN ($2, $2 || '.prepared') AND correlation_id = ANY($3::uuid[])`, connection.ID, PayablesAPIResource, []pgtype.UUID{olderID, newerID})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM integration_sync_batches WHERE connection_id = $1 AND correlation_id = ANY($2::uuid[])`, connection.ID, []pgtype.UUID{olderID, newerID})
	}
	cleanup()
	defer cleanup()
	for _, candidate := range []struct {
		id        pgtype.UUID
		startedAt time.Time
	}{{olderID, time.Now().Add(-time.Minute)}, {newerID, time.Now()}} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO integration_sync_batches (connection_id, resource, status, started_at, finished_at, correlation_id, data_stage)
			VALUES ($1, $2, 'FAILED', $3, $3, $4, 'RAW_ONLY')`, connection.ID, PayablesAPIResource, candidate.startedAt, candidate.id); err != nil {
			t.Fatalf("seed incomplete payable batch: %v", err)
		}
		for _, resource := range []string{PayablesAPIResource, PayablesAPIResource + ".prepared"} {
			if _, err := pool.Exec(ctx, `
				INSERT INTO bling_sync_page_checkpoints (connection_id, resource, correlation_id, page_number, page_limit, has_next, record_count, payload_hash, payload)
				VALUES ($1, $2, $3, 1, 100, false, 0, repeat('0', 64), '{}'::jsonb)`, connection.ID, resource, candidate.id); err != nil {
				t.Fatalf("seed %s checkpoint: %v", resource, err)
			}
		}
		for _, resource := range []string{CashMovementsAPIResource, ReconciledCashMovementsAPIResource, ReceivablesAPIResource, PaymentMethodsAPIResource, FinancialAccountsAPIResource} {
			if _, err := pool.Exec(ctx, `
				INSERT INTO integration_sync_batches (connection_id, resource, status, started_at, finished_at, correlation_id, data_stage)
				VALUES ($1, $2, 'SUCCESS', $3, $3, $4, 'NORMALIZED')`, connection.ID, resource, candidate.startedAt, candidate.id); err != nil {
				t.Fatalf("seed full snapshot resource %s: %v", resource, err)
			}
		}
	}

	resumed, found, err := resumableBlingSyncCorrelation(ctx, pool, connection.ID)
	if err != nil || !found || resumed != newer {
		t.Fatalf("resumable correlation = %s, found=%t, err=%v; want newest %s", resumed, found, err, newer)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO integration_sync_batches (connection_id, resource, status, started_at, finished_at, correlation_id, data_stage)
		VALUES ($1, $2, 'SUCCESS', clock_timestamp(), clock_timestamp(), $3, 'NORMALIZED')`, connection.ID, PayablesAPIResource, newerID); err != nil {
		t.Fatalf("mark newest payable snapshot as published: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM bling_sync_page_checkpoints WHERE connection_id = $1 AND resource IN ($2, $2 || '.prepared') AND correlation_id = $3`, connection.ID, PayablesAPIResource, newerID); err != nil {
		t.Fatalf("clear published payable checkpoints: %v", err)
	}
	resumed, found, err = resumableBlingSyncCorrelation(ctx, pool, connection.ID)
	if err != nil || !found || resumed != older {
		t.Fatalf("after publish, resumable correlation = %s, found=%t, err=%v; want older unfinished %s", resumed, found, err, older)
	}

	var attempts int
	var result BlingRawSyncResult
	steps := []blingSyncStep{{resource: PayablesAPIResource, target: &result.Payables, run: func() (APISyncResult, error) {
		attempts++
		return APISyncResult{Status: "FAILED"}, nil
	}}}
	skipPublishedBlingSyncSteps(ctx, pool, connection.ID, &newer, steps)
	if failures := runIndependentBlingSyncSteps(steps...); len(failures) != 0 || attempts != 0 || result.Payables.Status != "SUCCESS" {
		t.Fatalf("published source was not skipped: failures=%v attempts=%d result=%+v", failures, attempts, result.Payables)
	}
}
