package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/dbtx"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/scheduler"
)

const (
	// Each detail table returns at most 50 rows. These limits cap the raw
	// detail strings retained by the desktop snapshot at about 1.6 MiB.
	dashboardFieldMaxBytes  = 4 << 10
	dashboardRecordMaxBytes = 16 << 10
)

var (
	ErrDatabaseUnavailable    = errors.New("dashboard database is unavailable")
	ErrSnapshotRecordTooLarge = errors.New("dashboard snapshot record exceeds the configured limit")
)

type dashboardStreamDB interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type Reader struct {
	db  dashboardStreamDB
	now func() time.Time
}

func NewReader(pool *pgxpool.Pool) *Reader {
	if pool == nil {
		return &Reader{}
	}
	return &Reader{db: pool, now: func() time.Time { return time.Now().UTC() }}
}

// WriteDashboardSnapshot emits the same dashboard data as a JSON document but
// reads and writes detail rows one at a time. The read-only repeatable-read
// transaction keeps summary metrics and both tables on the same database view.
func (r *Reader) WriteDashboardSnapshot(ctx context.Context, dst io.Writer) (returnErr error) {
	if r == nil || r.db == nil {
		return ErrDatabaseUnavailable
	}
	if dst == nil {
		return errors.New("dashboard snapshot output is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if r.now != nil {
		now = r.now().UTC()
	}
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() {
		if err := dbtx.Rollback(ctx, tx); err != nil {
			returnErr = errors.Join(returnErr, err)
		}
	}()

	snapshot, err := readStreamSummary(ctx, tx, now)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	if len(encoded) == 0 || encoded[len(encoded)-1] != '}' {
		return errors.New("dashboard snapshot could not be encoded")
	}
	if err := writeAll(dst, encoded[:len(encoded)-1]); err != nil {
		return err
	}
	if err := writeAll(dst, []byte(`,"receivable_rows":[`)); err != nil {
		return err
	}
	if err := flushDashboardOutput(dst); err != nil {
		return err
	}
	if snapshot.Receivables.State != "UNAVAILABLE" {
		if err := streamReceivableRows(ctx, tx, dst); err != nil {
			return err
		}
	}
	if err := writeAll(dst, []byte(`],"payable_rows":[`)); err != nil {
		return err
	}
	if snapshot.Payables.State != "UNAVAILABLE" {
		if err := streamPayableRows(ctx, tx, dst); err != nil {
			return err
		}
	}
	if err := writeAll(dst, []byte(`]}`)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func readStreamSummary(ctx context.Context, tx pgx.Tx, asOf time.Time) (application.DashboardSnapshot, error) {
	businessDate := domain.BusinessDate(asOf)
	monthStart := time.Date(businessDate.Year(), businessDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	var (
		receivablesValue, futureValue, receiptsMonthValue, receivablesOverdueValue, receivablesDueNext90DaysValue                                                              string
		payablesValue, payablesTodayValue, payablesOverdueValue, payablesDueNext7DaysValue, payablesDueNext90DaysValue, paymentsMonthValue                                     string
		receivablesCount, futureCount, receiptsTotalCount, receiptsMonthCount                                                                                                  int64
		receivablesOverdueCount, receivablesDueNext90DaysCount, payablesCount, payablesTodayCount, payablesOverdueCount, payablesDueNext7DaysCount, payablesDueNext90DaysCount int64
		paymentsTotalCount, paymentsMonthCount                                                                                                                                 int64
		lastBlingSyncAt, dailyReceivablesSyncAt, dailyPayablesSyncAt, dailyCashMovementsSyncAt                                                                                 pgtype.Timestamptz
		dailyPayablesPartial, dailyCashMovementsPartial                                                                                                                        bool
	)
	dailyCorrelationID, err := dailyBlingCorrelationID(asOf)
	if err != nil {
		return application.DashboardSnapshot{}, err
	}
	err = tx.QueryRow(ctx, dashboardStreamSummarySQL,
		businessDate.Format("2006-01-02"),
		monthStart.Format("2006-01-02"),
		monthEnd.Format("2006-01-02"),
		dailyCorrelationID,
	).Scan(
		&receivablesValue, &receivablesCount,
		&futureValue, &futureCount,
		&receiptsMonthValue, &receiptsMonthCount, &receiptsTotalCount,
		&receivablesOverdueValue, &receivablesOverdueCount,
		&payablesValue, &payablesCount,
		&payablesTodayValue, &payablesTodayCount,
		&payablesOverdueValue, &payablesOverdueCount,
		&paymentsMonthValue, &paymentsMonthCount, &paymentsTotalCount,
		&lastBlingSyncAt,
		&dailyReceivablesSyncAt,
		&dailyPayablesSyncAt,
		&dailyPayablesPartial,
		&dailyCashMovementsSyncAt,
		&dailyCashMovementsPartial,
		&payablesDueNext7DaysValue,
		&payablesDueNext7DaysCount,
		&receivablesDueNext90DaysValue,
		&receivablesDueNext90DaysCount,
		&payablesDueNext90DaysValue,
		&payablesDueNext90DaysCount,
	)
	if err != nil {
		return application.DashboardSnapshot{}, err
	}
	receivablesFresh := completedBy(dailyReceivablesSyncAt, asOf)
	payablesFresh := completedBy(dailyPayablesSyncAt, asOf)
	cashMovementsFresh := completedBy(dailyCashMovementsSyncAt, asOf)
	state := "UNAVAILABLE"
	if futureCount > 0 || receivablesFresh || payablesFresh || cashMovementsFresh {
		state = "PARTIAL"
	}
	var lastBlingSyncTime *time.Time
	if lastBlingSyncAt.Valid {
		value := lastBlingSyncAt.Time.UTC()
		lastBlingSyncTime = &value
	}
	return application.DashboardSnapshot{
		AsOf:                     asOf,
		LastBlingSyncAt:          lastBlingSyncTime,
		DataState:                state,
		Receivables:              metricForSourceRun(receivablesValue, receivablesCount, receivablesCount, receivablesFresh, "BLING + NUVEM_PAGO"),
		FutureB2C:                metric(futureValue, futureCount, projectedState(futureCount), "NUVEM_PAGO"),
		ReceiptsMonth:            metricForSourceRun(receiptsMonthValue, receiptsMonthCount, receiptsTotalCount, cashMovementsFresh, "BLING", dailyCashMovementsPartial),
		ReceivablesOverdue:       metricForSourceRun(receivablesOverdueValue, receivablesOverdueCount, receivablesCount, receivablesFresh, "BLING + NUVEM_PAGO"),
		ReceivablesDueNext90Days: metricForSourceRun(receivablesDueNext90DaysValue, receivablesDueNext90DaysCount, receivablesCount, receivablesFresh, "BLING + NUVEM_PAGO"),
		Payables:                 metricForSourceRun(payablesValue, payablesCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueToday:         metricForSourceRun(payablesTodayValue, payablesTodayCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesOverdue:          metricForSourceRun(payablesOverdueValue, payablesOverdueCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueNext7Days:     metricForSourceRun(payablesDueNext7DaysValue, payablesDueNext7DaysCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueNext90Days:    metricForSourceRun(payablesDueNext90DaysValue, payablesDueNext90DaysCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		CashBalance:              application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
		PaymentsMonth:            metricForSourceRun(paymentsMonthValue, paymentsMonthCount, paymentsTotalCount, cashMovementsFresh, "BLING", dailyCashMovementsPartial),
	}, nil
}

func dailyBlingCorrelationID(asOf time.Time) (string, error) {
	schedule, err := scheduler.NewSaoPaulo0930()
	if err != nil {
		return "", err
	}
	runAt, err := schedule.LatestRunAtOrBefore(asOf)
	if err != nil {
		return "", err
	}
	return scheduler.DailyRunCorrelationID(scheduler.BlingDailyRunScope, runAt).String(), nil
}

func completedBy(syncAt pgtype.Timestamptz, asOf time.Time) bool {
	return syncAt.Valid && !syncAt.Time.After(asOf)
}

func metricForSourceRun(value string, count, sourceCount int64, fresh bool, source string, partial ...bool) application.DashboardMetric {
	if !fresh {
		return application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: source}
	}
	state := stateForCount(sourceCount)
	if len(partial) > 0 && partial[0] {
		state = "PARTIAL"
	}
	if state == "UNAVAILABLE" {
		// A successful daily load with no matching records confirms a zero; it
		// is different from a source that has not completed its current load.
		state = "CONFIRMED"
	}
	return metric(value, count, state, source)
}

func metric(value string, count int64, state, source string) application.DashboardMetric {
	if state == "UNAVAILABLE" {
		return application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: source}
	}
	return application.DashboardMetric{Value: value, Count: count, State: state, SourceSystem: source}
}

func stateForCount(count int64) string {
	if count == 0 {
		return "UNAVAILABLE"
	}
	return "CONFIRMED"
}

func projectedState(count int64) string {
	if count == 0 {
		return "UNAVAILABLE"
	}
	return "PROJECTED"
}

func streamReceivableRows(ctx context.Context, tx pgx.Tx, dst io.Writer) error {
	rows, err := tx.Query(ctx, dashboardReceivableRowsSQL, dashboardFieldMaxBytes, dashboardRecordMaxBytes)
	if err != nil {
		return err
	}
	defer rows.Close()
	first := true
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		var row application.DashboardReceivableRow
		var oversized bool
		if err := rows.Scan(&row.Customer, &row.Origin, &row.DueDate, &row.GrossValue, &row.NetValue, &row.Status, &oversized); err != nil {
			return err
		}
		if oversized {
			return ErrSnapshotRecordTooLarge
		}
		if !first {
			if err := writeAll(dst, []byte{','}); err != nil {
				return err
			}
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if err := writeAll(dst, encoded); err != nil {
			return err
		}
		if err := flushDashboardOutput(dst); err != nil {
			return err
		}
		first = false
	}
	return rows.Err()
}

func streamPayableRows(ctx context.Context, tx pgx.Tx, dst io.Writer) error {
	rows, err := tx.Query(ctx, dashboardPayableRowsSQL, dashboardFieldMaxBytes, dashboardRecordMaxBytes)
	if err != nil {
		return err
	}
	defer rows.Close()
	first := true
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		var row application.DashboardPayableRow
		var oversized bool
		if err := rows.Scan(&row.Supplier, &row.Document, &row.DueDate, &row.Value, &row.Category, &row.Status, &oversized); err != nil {
			return err
		}
		if oversized {
			return ErrSnapshotRecordTooLarge
		}
		if !first {
			if err := writeAll(dst, []byte{','}); err != nil {
				return err
			}
		}
		encoded, err := json.Marshal(row)
		if err != nil {
			return err
		}
		if err := writeAll(dst, encoded); err != nil {
			return err
		}
		if err := flushDashboardOutput(dst); err != nil {
			return err
		}
		first = false
	}
	return rows.Err()
}

func writeAll(dst io.Writer, data []byte) error {
	for len(data) > 0 {
		count, err := dst.Write(data)
		if err != nil {
			return err
		}
		if count <= 0 || count > len(data) {
			return io.ErrShortWrite
		}
		data = data[count:]
	}
	return nil
}

func flushDashboardOutput(dst io.Writer) error {
	if flusher, ok := dst.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}
	return nil
}

const dashboardStreamSummarySQL = `
WITH eligible_receivables AS (
  SELECT *
    FROM receivables
   WHERE (source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED')
      OR (source_system = 'BLING' AND business_type = 'B2B')
), r AS (
  SELECT
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED'
    ), 0)::text AS future_value,
    COUNT(id) FILTER (
      WHERE source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED'
    ) AS future_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND due_date < $1::date
    ), 0)::text AS overdue_value,
    COUNT(id) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND due_date < $1::date
    ) AS overdue_count
    ,COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0
        AND due_date >= $1::date AND due_date < ($1::date + 90)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS due_next_90_days_value,
    COUNT(id) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0
        AND due_date >= $1::date AND due_date < ($1::date + 90)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS due_next_90_days_count
  FROM eligible_receivables
), cash_snapshot AS (
  SELECT b.id AS sync_batch_id, b.status
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.resource = 'caixas'
     AND b.correlation_id = $4
     AND b.status IN ('SUCCESS', 'PARTIAL')
     AND b.data_stage = 'NORMALIZED'
   ORDER BY b.finished_at DESC, b.id DESC
   LIMIT 1
), rc AS (
  SELECT
    COALESCE(SUM(amount) FILTER (
      WHERE direction = 'INFLOW' AND source_status IN ('R', '')
        AND movement_date >= $2::date AND movement_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE direction = 'INFLOW' AND source_status IN ('R', '')
        AND movement_date >= $2::date AND movement_date < $3::date
    ) AS month_count,
    COUNT(id) FILTER (WHERE direction = 'INFLOW' AND source_status IN ('R', '')) AS total_count
  FROM bling_cash_movements movement
  JOIN cash_snapshot snapshot ON snapshot.sync_batch_id = movement.sync_batch_id
), latest_daily_payable_snapshot AS (
  SELECT b.id AS sync_batch_id
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.resource = 'contas.pagar'
     AND b.correlation_id IS NOT NULL
     AND b.status IN ('SUCCESS', 'PARTIAL')
     AND b.data_stage = 'NORMALIZED'
   ORDER BY b.finished_at DESC, b.id DESC
   LIMIT 1
), p AS (
  SELECT
    COALESCE(SUM(open_balance), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(open_balance) FILTER (
      WHERE due_date = $1::date
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS today_value,
    COUNT(id) FILTER (
      WHERE due_date = $1::date
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS today_count,
    COALESCE(SUM(open_balance) FILTER (
      WHERE open_balance <> 0 AND due_date < $1::date
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS overdue_value,
    COUNT(id) FILTER (
      WHERE open_balance <> 0 AND due_date < $1::date
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS overdue_count,
    COALESCE(SUM(open_balance) FILTER (
      WHERE due_date > $1::date AND due_date < ($1::date + 8)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS due_next_7_days_value,
    COUNT(id) FILTER (
      WHERE due_date > $1::date AND due_date < ($1::date + 8)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS due_next_7_days_count
    ,COALESCE(SUM(open_balance) FILTER (
      WHERE due_date >= $1::date AND due_date < ($1::date + 90)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS due_next_90_days_value,
    COUNT(id) FILTER (
      WHERE due_date >= $1::date AND due_date < ($1::date + 90)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS due_next_90_days_count
  FROM payables pw
  WHERE pw.open_balance <> 0
    AND (pw.source_entity <> 'bling.conta_pagar' OR EXISTS (
      SELECT 1
        FROM latest_daily_payable_snapshot latest
        JOIN payable_sync_snapshot_rows membership ON membership.sync_batch_id = latest.sync_batch_id
       WHERE membership.payable_id = pw.id
    ))
), pm AS (
  SELECT
    COALESCE(SUM(-amount) FILTER (
      WHERE direction = 'OUTFLOW' AND source_status IN ('R', '')
        AND movement_date >= $2::date AND movement_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE direction = 'OUTFLOW' AND source_status IN ('R', '')
        AND movement_date >= $2::date AND movement_date < $3::date
    ) AS month_count,
    COUNT(id) FILTER (WHERE direction = 'OUTFLOW' AND source_status IN ('R', '')) AS total_count
  FROM bling_cash_movements movement
  JOIN cash_snapshot snapshot ON snapshot.sync_batch_id = movement.sync_batch_id
), last_bling_sync AS (
  SELECT MAX(b.finished_at) AS finished_at
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.status IN ('SUCCESS', 'PARTIAL')
), current_daily_bling_sync AS (
  SELECT
    MAX(b.finished_at) FILTER (WHERE b.resource = 'contas.receber' AND b.status = 'SUCCESS' AND b.data_stage = 'NORMALIZED') AS receivables_finished_at,
    MAX(b.finished_at) FILTER (WHERE b.resource = 'contas.pagar' AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED') AS payables_finished_at,
    COALESCE(BOOL_OR(b.status = 'PARTIAL') FILTER (WHERE b.resource = 'contas.pagar' AND b.data_stage = 'NORMALIZED'), false) AS payables_partial,
    MAX(b.finished_at) FILTER (WHERE b.resource = 'caixas' AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED') AS cash_movements_finished_at,
    COALESCE(BOOL_OR(b.status = 'PARTIAL') FILTER (WHERE b.resource = 'caixas' AND b.data_stage = 'NORMALIZED'), false) AS cash_movements_partial
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.correlation_id = $4
)
SELECT r.total_value, r.total_count, r.future_value, r.future_count,
       rc.month_value, rc.month_count, rc.total_count,
       r.overdue_value, r.overdue_count,
       p.total_value, p.total_count, p.today_value, p.today_count,
       p.overdue_value, p.overdue_count,
       pm.month_value, pm.month_count, pm.total_count,
       last_bling_sync.finished_at,
       current_daily_bling_sync.receivables_finished_at,
       current_daily_bling_sync.payables_finished_at,
       current_daily_bling_sync.payables_partial,
       current_daily_bling_sync.cash_movements_finished_at,
       current_daily_bling_sync.cash_movements_partial,
       p.due_next_7_days_value, p.due_next_7_days_count,
       r.due_next_90_days_value, r.due_next_90_days_count,
       p.due_next_90_days_value, p.due_next_90_days_count
FROM r CROSS JOIN rc CROSS JOIN p CROSS JOIN pm CROSS JOIN last_bling_sync CROSS JOIN current_daily_bling_sync
`

const dashboardReceivableRowsSQL = `
WITH eligible_receivables AS (
  SELECT * FROM receivables
   WHERE (source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED')
      OR (source_system = 'BLING' AND business_type = 'B2B')
)
SELECT CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.customer END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.origin END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.due_date END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.gross_value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.net_value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.status END,
       limits.exceeds_limit
  FROM (SELECT * FROM eligible_receivables
         WHERE COALESCE(open_balance, net_amount, gross_amount) > 0
         ORDER BY due_date NULLS LAST, source_id LIMIT 50) er
  LEFT JOIN contacts c ON c.id = er.customer_id
 CROSS JOIN LATERAL (
   SELECT COALESCE(c.name, '') AS customer,
          CASE er.source_system WHEN 'NUVEM_PAGO' THEN 'Nuvem Pago' WHEN 'BLING' THEN 'Bling' ELSE er.source_system END AS origin,
          COALESCE(to_char(er.due_date, 'YYYY-MM-DD'), '') AS due_date,
          COALESCE(er.gross_amount::text, '') AS gross_value,
          COALESCE(COALESCE(er.net_amount, er.gross_amount)::text, '') AS net_value,
          COALESCE(er.status, '') AS status
 ) AS row_data
 CROSS JOIN LATERAL (
   SELECT
     octet_length(row_data.customer)::bigint > $1 OR
     octet_length(row_data.origin)::bigint > $1 OR
     octet_length(row_data.due_date)::bigint > $1 OR
     octet_length(row_data.gross_value)::bigint > $1 OR
     octet_length(row_data.net_value)::bigint > $1 OR
     octet_length(row_data.status)::bigint > $1 OR
     octet_length(row_data.customer)::bigint + octet_length(row_data.origin)::bigint +
     octet_length(row_data.due_date)::bigint + octet_length(row_data.gross_value)::bigint +
     octet_length(row_data.net_value)::bigint + octet_length(row_data.status)::bigint > $2 AS exceeds_limit
 ) AS limits
`

const dashboardPayableRowsSQL = `
WITH latest_daily_payable_snapshot AS (
  SELECT b.id AS sync_batch_id
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.resource = 'contas.pagar'
     AND b.correlation_id IS NOT NULL
     AND b.status IN ('SUCCESS', 'PARTIAL')
     AND b.data_stage = 'NORMALIZED'
   ORDER BY b.finished_at DESC, b.id DESC
   LIMIT 1
)
SELECT CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.supplier END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.document END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.due_date END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.category END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.status END,
       limits.exceeds_limit
  FROM (SELECT * FROM payables pw
         WHERE pw.open_balance <> 0
           AND (pw.source_entity <> 'bling.conta_pagar' OR EXISTS (
             SELECT 1
               FROM latest_daily_payable_snapshot latest
               JOIN payable_sync_snapshot_rows membership ON membership.sync_batch_id = latest.sync_batch_id
              WHERE membership.payable_id = pw.id
           ))
         ORDER BY due_date, source_id LIMIT 50) pw
  LEFT JOIN contacts c ON c.id = pw.supplier_id
 CROSS JOIN LATERAL (
   SELECT COALESCE(c.name, '') AS supplier,
          COALESCE(pw.document, '') AS document,
          COALESCE(to_char(pw.due_date, 'YYYY-MM-DD'), '') AS due_date,
          COALESCE(pw.open_balance::text, '') AS value,
          COALESCE(pw.category, '') AS category,
          COALESCE(pw.status, '') AS status
 ) AS row_data
 CROSS JOIN LATERAL (
   SELECT
     octet_length(row_data.supplier)::bigint > $1 OR
     octet_length(row_data.document)::bigint > $1 OR
     octet_length(row_data.due_date)::bigint > $1 OR
     octet_length(row_data.value)::bigint > $1 OR
     octet_length(row_data.category)::bigint > $1 OR
     octet_length(row_data.status)::bigint > $1 OR
     octet_length(row_data.supplier)::bigint + octet_length(row_data.document)::bigint +
     octet_length(row_data.due_date)::bigint + octet_length(row_data.value)::bigint +
     octet_length(row_data.category)::bigint + octet_length(row_data.status)::bigint > $2 AS exceeds_limit
 ) AS limits
`
