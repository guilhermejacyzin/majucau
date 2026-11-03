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
	for _, statement := range []string{
		`CREATE TEMP TABLE integration_connections (id uuid, provider text) ON COMMIT DROP`,
		`CREATE TEMP TABLE integration_sync_batches (id uuid DEFAULT gen_random_uuid(), connection_id uuid, resource text, started_at timestamptz, finished_at timestamptz, status text, correlation_id uuid, data_stage text) ON COMMIT DROP`,
		`CREATE TEMP TABLE contacts (id uuid, name text) ON COMMIT DROP`,
		`CREATE TEMP TABLE raw_records (id uuid, payload jsonb) ON COMMIT DROP`,
		`CREATE TEMP TABLE receivables (id uuid, source_system text, source_entity text, source_id text, business_type text, customer_id uuid, raw_record_id uuid, status text, open_balance numeric, net_amount numeric, gross_amount numeric, due_date date, expected_receipt_date date) ON COMMIT DROP`,
		`CREATE TEMP TABLE receivable_sync_snapshot_rows (sync_batch_id uuid, receivable_id uuid) ON COMMIT DROP`,
		`CREATE TEMP TABLE receipts (id uuid, source_system text, status text, receipt_date date, amount numeric) ON COMMIT DROP`,
		`CREATE TEMP TABLE payables (id uuid, source_entity text, open_balance numeric, due_date date, status text) ON COMMIT DROP`,
		`CREATE TEMP TABLE payable_sync_snapshot_rows (sync_batch_id uuid, payable_id uuid) ON COMMIT DROP`,
		`CREATE TEMP TABLE payments (id uuid, source_system text, payment_date date, amount numeric) ON COMMIT DROP`,
		`CREATE TEMP TABLE bling_cash_movements (id uuid, sync_batch_id uuid, account_source_id text, account_name text, direction text, source_status text, movement_date date, amount numeric) ON COMMIT DROP`,
	} {
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatalf("create isolated dashboard fixture table: %v", err)
		}
	}
	connectionID := "00000000-0000-4000-8000-000000000001"
	if _, err := tx.Exec(ctx, `INSERT INTO integration_connections (id, provider) VALUES ($1::uuid, 'BLING')`, connectionID); err != nil {
		t.Fatalf("seed isolated Bling connection: %v", err)
	}
	var receivableBatchID, payableBatchID, cashBatchID string
	for _, resource := range []string{"contas.receber", "contas.pagar", "caixas"} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO integration_sync_batches (connection_id, resource, started_at, finished_at, status, correlation_id, data_stage)
			VALUES ($1::uuid, $2, $3, $4, 'SUCCESS', $5, 'RAW_ONLY')`, connectionID, resource, asOf.Add(-time.Hour), asOf, correlationID); err != nil {
			t.Fatalf("seed successful RAW-only %s batch: %v", resource, err)
		}
		if resource == "contas.pagar" {
			if err := tx.QueryRow(ctx, `SELECT id::text FROM integration_sync_batches WHERE connection_id = $1::uuid AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID).Scan(&payableBatchID); err != nil {
				t.Fatalf("read seeded payable batch ID: %v", err)
			}
		}
		if resource == "contas.receber" {
			if err := tx.QueryRow(ctx, `SELECT id::text FROM integration_sync_batches WHERE connection_id = $1::uuid AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID).Scan(&receivableBatchID); err != nil {
				t.Fatalf("read seeded receivable batch ID: %v", err)
			}
		}
		if resource == "caixas" {
			if err := tx.QueryRow(ctx, `SELECT id::text FROM integration_sync_batches WHERE connection_id = $1::uuid AND resource = $2 AND correlation_id = $3`, connectionID, resource, correlationID).Scan(&cashBatchID); err != nil {
				t.Fatalf("read seeded cash movement batch ID: %v", err)
			}
		}
	}
	snapshot, err := readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with RAW-only batches: %v", err)
	}
	if snapshot.LastBlingSyncAt == nil {
		t.Fatal("last RAW sync time is missing")
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
	blingCustomerID := "00000000-0000-4000-8000-000000000041"
	nuvemRawID := "00000000-0000-4000-8000-000000000042"
	blingReceivableID := "00000000-0000-4000-8000-000000000043"
	nuvemReceivableID := "00000000-0000-4000-8000-000000000044"
	independentReceivableID := "00000000-0000-4000-8000-000000000045"
	if _, err := tx.Exec(ctx, `INSERT INTO contacts (id, name) VALUES ($1::uuid, 'José Cliente')`, blingCustomerID); err != nil {
		t.Fatalf("seed normalized Bling customer name: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO raw_records (id, payload) VALUES ($1::uuid, '{"customer_name":" jose   cliente "}'::jsonb)`, nuvemRawID); err != nil {
		t.Fatalf("seed Nuvem customer name in RAW fixture: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO receivables (id, source_system, source_entity, source_id, business_type, customer_id, raw_record_id, status, open_balance, net_amount, gross_amount, due_date, expected_receipt_date) VALUES
		  ($1::uuid, 'BLING', 'bling.conta_receber', 'bling-1', 'B2C', $2::uuid, NULL, '1', 100, NULL, 100, $4::date + 25, NULL),
		  ($3::uuid, 'NUVEM_PAGO', 'nuvem_pago_recebimentos_futuros_csv_v1', 'nuvem-1', 'B2C', NULL, $5::uuid, 'PROJECTED', NULL, 97, 100, NULL, $4::date + 20),
		  ($6::uuid, 'BLING', 'bling.conta_receber', 'bling-2', 'B2C', $2::uuid, NULL, '1', 50, NULL, 50, $4::date + 10, NULL)`,
		blingReceivableID, blingCustomerID, nuvemReceivableID, asOf.Format("2006-01-02"), nuvemRawID, independentReceivableID); err != nil {
		t.Fatalf("seed cross-source matching and separate receivables: %v", err)
	}
	for _, receivableID := range []string{blingReceivableID, independentReceivableID} {
		if _, err := tx.Exec(ctx, `INSERT INTO receivable_sync_snapshot_rows (sync_batch_id, receivable_id) VALUES ($1::uuid, $2::uuid)`, receivableBatchID, receivableID); err != nil {
			t.Fatalf("seed current receivable snapshot membership: %v", err)
		}
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with normalized account batches: %v", err)
	}
	if snapshot.Receivables.State != "PARTIAL" || snapshot.Payables.State != "CONFIRMED" {
		t.Fatalf("receivables should remain partial until a same-day Nuvem Pago load succeeds: receivables=%+v payables=%+v", snapshot.Receivables, snapshot.Payables)
	}
	if snapshot.ReceiptsMonth.State != "UNAVAILABLE" || snapshot.PaymentsMonth.State != "UNAVAILABLE" {
		t.Fatalf("cash movement metrics should remain unavailable until caixas is normalized: receipts=%+v payments=%+v", snapshot.ReceiptsMonth, snapshot.PaymentsMonth)
	}
	if snapshot.ReceivablesDueNext90Days.Value != "147" || snapshot.ReceivablesDueNext90Days.Count != 2 {
		t.Fatalf("next 90 day receivables should pair one Bling/Nuvem match by name and gross value without using date, while retaining a different value: %+v", snapshot.ReceivablesDueNext90Days)
	}
	nuvemConnectionID := "00000000-0000-4000-8000-000000000046"
	if _, err := tx.Exec(ctx, `INSERT INTO integration_connections (id, provider) VALUES ($1::uuid, 'NUVEM_PAGO')`, nuvemConnectionID); err != nil {
		t.Fatalf("seed isolated Nuvem Pago connection: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO integration_sync_batches (connection_id, resource, started_at, finished_at, status, data_stage) VALUES ($1::uuid, 'nuvem_pago_recebimentos_futuros_csv_v1', $2, $2, 'SUCCESS', 'NORMALIZED')`, nuvemConnectionID, asOf); err != nil {
		t.Fatalf("seed same-day successful Nuvem Pago load: %v", err)
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil || snapshot.Receivables.State != "CONFIRMED" || snapshot.ReceivablesDueNext90Days.State != "CONFIRMED" {
		t.Fatalf("same-day successful Bling and Nuvem sources should confirm receivables, snapshot=%+v err=%v", snapshot, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET data_stage = 'NORMALIZED' WHERE connection_id = $1::uuid AND resource = 'caixas' AND correlation_id = $2`, connectionID, correlationID); err != nil {
		t.Fatalf("mark cash movement batch normalized: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET status = 'PARTIAL' WHERE id = $1::uuid`, cashBatchID); err != nil {
		t.Fatalf("mark cash source with missing statuses partial: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO bling_cash_movements (id, sync_batch_id, account_source_id, account_name, direction, source_status, movement_date, amount) VALUES
		('00000000-0000-4000-8000-000000000031', $1::uuid, '1', 'Itaú', 'INFLOW', 'R', '2026-10-10', 100.00),
		('00000000-0000-4000-8000-000000000032', $1::uuid, '1', 'Itaú', 'OUTFLOW', 'R', '2026-10-11', 25.00),
		('00000000-0000-4000-8000-000000000033', $1::uuid, '2', 'Mercado Pago', 'INFLOW', '', '2026-10-12', 20.00),
		('00000000-0000-4000-8000-000000000034', $1::uuid, '2', 'Mercado Pago', 'OUTFLOW', '', '2026-10-13', 5.00)`, cashBatchID); err != nil {
		t.Fatalf("seed cash movements including entries with missing source status: %v", err)
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("readStreamSummary() failed against PostgreSQL with normalized cash batch: %v", err)
	}
	if snapshot.ReceiptsMonth.State != "PARTIAL" || snapshot.PaymentsMonth.State != "PARTIAL" {
		t.Fatalf("cash movement metrics should keep values visible and mark missing source status partial: receipts=%+v payments=%+v", snapshot.ReceiptsMonth, snapshot.PaymentsMonth)
	}
	if snapshot.ReceiptsMonth.Value != "120.00" || snapshot.PaymentsMonth.Value != "-30.00" {
		t.Fatalf("cash flow should include listed movements, with inflows positive and outflows negative: receipts=%+v payments=%+v", snapshot.ReceiptsMonth, snapshot.PaymentsMonth)
	}
	if snapshot.CashBalance.State != "UNAVAILABLE" || snapshot.CashBalance.Value != "" {
		t.Fatalf("current bank balance must remain unavailable until Bling provides an official balance, not be inferred from movements: %+v", snapshot.CashBalance)
	}
	currentPayableID := "00000000-0000-4000-8000-000000000011"
	stalePayableID := "00000000-0000-4000-8000-000000000012"
	legacyPayableID := "00000000-0000-4000-8000-000000000013"
	dayOnePayableID := "00000000-0000-4000-8000-000000000014"
	daySevenPayableID := "00000000-0000-4000-8000-000000000015"
	dayEightPayableID := "00000000-0000-4000-8000-000000000016"
	cancelledPayableID := "00000000-0000-4000-8000-000000000017"
	canceledPayableID := "00000000-0000-4000-8000-000000000018"
	refundedPayableID := "00000000-0000-4000-8000-000000000019"
	reversedPayableID := "00000000-0000-4000-8000-000000000020"
	estornadoPayableID := "00000000-0000-4000-8000-000000000021"
	businessDate := asOf.Format("2006-01-02")
	if _, err := tx.Exec(ctx, `
		INSERT INTO payables (id, source_entity, open_balance, due_date, status) VALUES
		($1::uuid, 'bling.conta_pagar', -200, $12::date, 'OPEN'),
		($2::uuid, 'bling.conta_pagar', -900, $12::date, 'OPEN'),
		($3::uuid, 'legacy.report', -20, $12::date, 'OPEN'),
		($4::uuid, 'bling.conta_pagar', -10, $12::date + 1, 'OPEN'),
		($5::uuid, 'bling.conta_pagar', -30, $12::date + 7, 'OPEN'),
		($6::uuid, 'bling.conta_pagar', -50, $12::date + 8, 'OPEN'),
		($7::uuid, 'bling.conta_pagar', -60, $12::date, 'CANCELLED'),
		($8::uuid, 'bling.conta_pagar', -70, $12::date - 1, 'CANCELED'),
		($9::uuid, 'bling.conta_pagar', -80, $12::date + 2, 'REFUNDED'),
		($10::uuid, 'bling.conta_pagar', -90, $12::date + 2, 'REVERSED'),
		($11::uuid, 'bling.conta_pagar', -100, $12::date + 2, 'ESTORNADO')`,
		currentPayableID, stalePayableID, legacyPayableID, dayOnePayableID, daySevenPayableID, dayEightPayableID,
		cancelledPayableID, canceledPayableID, refundedPayableID, reversedPayableID, estornadoPayableID, businessDate); err != nil {
		t.Fatalf("seed current, stale, legacy, boundary, and canceled payable rows: %v", err)
	}
	for _, payableID := range []string{currentPayableID, dayOnePayableID, daySevenPayableID, dayEightPayableID, cancelledPayableID, canceledPayableID, refundedPayableID, reversedPayableID, estornadoPayableID} {
		if _, err := tx.Exec(ctx, `INSERT INTO payable_sync_snapshot_rows (sync_batch_id, payable_id) VALUES ($1::uuid, $2::uuid)`, payableBatchID, payableID); err != nil {
			t.Fatalf("seed current payable snapshot membership: %v", err)
		}
	}
	snapshot, err = readStreamSummary(ctx, tx, asOf)
	if err != nil {
		t.Fatalf("read summary with daily payable membership: %v", err)
	}
	if snapshot.Payables.Value != "-710" || snapshot.Payables.Count != 10 {
		t.Fatalf("payable total includes a stale account or excludes a legacy row: %+v, want -R$ 710 from 10 rows", snapshot.Payables)
	}
	if snapshot.PayablesDueNext7Days.State != "CONFIRMED" || snapshot.PayablesDueNext7Days.Value != "-40" || snapshot.PayablesDueNext7Days.Count != 2 {
		t.Fatalf("upcoming payables should include only open outflows due on days 1–7 from the current snapshot: %+v, want -R$ 40 from 2 rows", snapshot.PayablesDueNext7Days)
	}
	if snapshot.PayablesDueNext90Days.State != "CONFIRMED" || snapshot.PayablesDueNext90Days.Value != "-310" || snapshot.PayablesDueNext90Days.Count != 5 {
		t.Fatalf("next 90 day payables = %+v, want -R$ 310 across five open titles including today", snapshot.PayablesDueNext90Days)
	}
	if snapshot.PayablesDueToday.State != "CONFIRMED" || snapshot.PayablesDueToday.Value != "-220" || snapshot.PayablesDueToday.Count != 2 {
		t.Fatalf("today's aging should ignore canceled titles and retain current and legacy outflows: %+v, want -R$ 220 from 2 rows", snapshot.PayablesDueToday)
	}
	if snapshot.PayablesOverdue.State != "CONFIRMED" || snapshot.PayablesOverdue.Value != "0" || snapshot.PayablesOverdue.Count != 0 {
		t.Fatalf("overdue aging should exclude a canceled title: %+v, want confirmed zero", snapshot.PayablesOverdue)
	}
}
