package dashboard

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/scheduler"
)

func TestReadStreamSummaryPostgresRequiresNormalizedDailyStages(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL dashboard query test")
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
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		t.Fatalf("begin PostgreSQL transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	asOf := time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	schedule, err := scheduler.NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}
	scheduledAt, err := schedule.LatestRunAtOrBefore(asOf)
	if err != nil {
		t.Fatal(err)
	}
	correlationID := scheduler.DailyRunCorrelationID(scheduler.BlingDailyRunScope, scheduledAt)
	if _, err := tx.Exec(ctx, `INSERT INTO integration_connections (provider) VALUES ('BLING') ON CONFLICT (provider) DO NOTHING`); err != nil {
		t.Fatalf("seed Bling connection: %v", err)
	}
	var connectionID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM integration_connections WHERE provider = 'BLING'`).Scan(&connectionID); err != nil {
		t.Fatalf("read Bling connection id: %v", err)
	}
	for _, resource := range []string{"contas.receber", "contas.pagar", "caixas"} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO integration_sync_batches (connection_id, resource, started_at, finished_at, status, correlation_id, data_stage)
			VALUES ($1::uuid, $2, $3, $4, 'SUCCESS', $5, 'RAW_ONLY')`, connectionID, resource, asOf.Add(-time.Hour), asOf, correlationID); err != nil {
			t.Fatalf("seed successful RAW-only %s batch: %v", resource, err)
		}
	}
	snapshot, err := readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with RAW-only batches: %v", err)
	}
	if snapshot.LastBlingSyncAt == nil || !snapshot.LastBlingSyncAt.Equal(asOf) {
		t.Fatalf("last RAW sync = %v, want %s", snapshot.LastBlingSyncAt, asOf)
	}
	for _, item := range []struct {
		name   string
		metric application.DashboardMetric
	}{
		{name: "receivables", metric: snapshot.Receivables},
		{name: "receipts this month", metric: snapshot.ReceiptsMonth},
		{name: "payables", metric: snapshot.Payables},
		{name: "payments this month", metric: snapshot.PaymentsMonth},
	} {
		if item.metric.State != "UNAVAILABLE" || item.metric.Value != "" {
			t.Errorf("%s after successful RAW-only sync = %+v, want unavailable", item.name, item.metric)
		}
	}
	for _, resource := range []string{"contas.receber", "contas.pagar"} {
		if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE connection_id = $1::uuid AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID); err != nil {
			t.Fatalf("mark %s normalized: %v", resource, err)
		}
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with normalized account batches: %v", err)
	}
	if snapshot.Receivables.State != "CONFIRMED" || snapshot.Payables.State != "CONFIRMED" {
		t.Fatalf("account metrics should follow normalized account batches: receivables=%+v payables=%+v", snapshot.Receivables, snapshot.Payables)
	}
	if snapshot.ReceiptsMonth.State != "UNAVAILABLE" || snapshot.PaymentsMonth.State != "UNAVAILABLE" {
		t.Fatalf("cash movement metrics should remain unavailable until caixas is normalized: receipts=%+v payments=%+v", snapshot.ReceiptsMonth, snapshot.PaymentsMonth)
	}
	if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE connection_id = $1::uuid AND resource = 'caixas' AND correlation_id = $2`, connectionID, correlationID); err != nil {
		t.Fatalf("mark cash movement batch normalized: %v", err)
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with normalized cash batch: %v", err)
	}
	if snapshot.ReceiptsMonth.State != "CONFIRMED" || snapshot.PaymentsMonth.State != "CONFIRMED" {
		t.Fatalf("cash movement metrics should follow normalized caixas: receipts=%+v payments=%+v", snapshot.ReceiptsMonth, snapshot.PaymentsMonth)
	}
}
