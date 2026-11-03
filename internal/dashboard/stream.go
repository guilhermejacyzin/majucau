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
		receivablesValue, futureValue, receiptsMonthValue, receivablesOverdueValue, receivablesDueNext90DaysValue, receivablesDueNext30DaysValue                                                                                          string
		payablesValue, payablesTodayValue, payablesOverdueValue, payablesDueNext7DaysValue, payablesDueNext90DaysValue, payablesDueNext30DaysValue, paymentsMonthValue                                                                    string
		receivablesCount, futureCount, receiptsTotalCount, receiptsMonthCount                                                                                                                                                             int64
		receivablesOverdueCount, receivablesDueNext90DaysCount, receivablesDueNext30DaysCount, payablesCount, payablesTodayCount, payablesOverdueCount, payablesDueNext7DaysCount, payablesDueNext90DaysCount, payablesDueNext30DaysCount int64
		blingReceivablesCount, blingReceivablesOverdueCount, due90BlingCount, due90NuvemCount, due30BlingCount, due30NuvemCount                                                                                                           int64
		paymentsTotalCount, paymentsMonthCount                                                                                                                                                                                            int64
		blingReceivablesValue, blingReceivablesOverdueValue, due90BlingValue, due90NuvemValue, due30BlingValue, due30NuvemValue                                                                                                           string
		lastBlingSyncAt, dailyReceivablesSyncAt, dailyPayablesSyncAt, dailyCashMovementsSyncAt                                                                                                                                            pgtype.Timestamptz
		dailyReceivablesPartial, dailyPayablesPartial, dailyCashMovementsPartial, dailyNuvemFuturePartial                                                                                                                                 bool
	)
	dailyCorrelationID, err := latestPublishedBlingCorrelationID(ctx, tx, asOf)
	if err != nil {
		return application.DashboardSnapshot{}, err
	}
	cashFlow, err := readBlingCashFlow(ctx, tx, businessDate, dailyCorrelationID)
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
		&receivablesDueNext30DaysValue,
		&receivablesDueNext30DaysCount,
		&payablesDueNext90DaysValue,
		&payablesDueNext90DaysCount,
		&payablesDueNext30DaysValue,
		&payablesDueNext30DaysCount,
		&dailyNuvemFuturePartial,
		&blingReceivablesValue,
		&blingReceivablesCount,
		&blingReceivablesOverdueValue,
		&blingReceivablesOverdueCount,
		&due90BlingValue,
		&due90BlingCount,
		&due90NuvemValue,
		&due90NuvemCount,
		&due30BlingValue,
		&due30BlingCount,
		&due30NuvemValue,
		&due30NuvemCount,
		&dailyReceivablesPartial,
	)
	if err != nil {
		return application.DashboardSnapshot{}, err
	}
	receivablesFresh := completedBy(dailyReceivablesSyncAt, asOf)
	nuvemFutureFresh := !dailyNuvemFuturePartial
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
		Receivables:              combinedReceivableMetric(receivablesValue, receivablesCount, blingReceivablesValue, blingReceivablesCount, futureValue, futureCount, receivablesFresh, dailyReceivablesPartial, nuvemFutureFresh),
		FutureB2C:                metric(futureValue, futureCount, projectedState(futureCount), "NUVEM_PAGO"),
		ReceiptsMonth:            metricForSourceRun(receiptsMonthValue, receiptsMonthCount, receiptsTotalCount, cashMovementsFresh, "BLING", dailyCashMovementsPartial),
		ReceivablesOverdue:       combinedReceivableMetric(receivablesOverdueValue, receivablesOverdueCount, blingReceivablesOverdueValue, blingReceivablesOverdueCount, "0.0000", 0, receivablesFresh, dailyReceivablesPartial, nuvemFutureFresh),
		ReceivablesDueNext90Days: combinedReceivableMetric(receivablesDueNext90DaysValue, receivablesDueNext90DaysCount, due90BlingValue, due90BlingCount, due90NuvemValue, due90NuvemCount, receivablesFresh, dailyReceivablesPartial, nuvemFutureFresh),
		ReceivablesDueNext30Days: combinedReceivableMetric(receivablesDueNext30DaysValue, receivablesDueNext30DaysCount, due30BlingValue, due30BlingCount, due30NuvemValue, due30NuvemCount, receivablesFresh, dailyReceivablesPartial, nuvemFutureFresh),
		Payables:                 metricForSourceRun(payablesValue, payablesCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueToday:         metricForSourceRun(payablesTodayValue, payablesTodayCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesOverdue:          metricForSourceRun(payablesOverdueValue, payablesOverdueCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueNext7Days:     metricForSourceRun(payablesDueNext7DaysValue, payablesDueNext7DaysCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueNext90Days:    metricForSourceRun(payablesDueNext90DaysValue, payablesDueNext90DaysCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		PayablesDueNext30Days:    metricForSourceRun(payablesDueNext30DaysValue, payablesDueNext30DaysCount, payablesCount, payablesFresh, "BLING", dailyPayablesPartial),
		CashBalance:              application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
		CashFlow:                 cashFlow,
		PaymentsMonth:            metricForSourceRun(paymentsMonthValue, paymentsMonthCount, paymentsTotalCount, cashMovementsFresh, "BLING", dailyCashMovementsPartial),
	}, nil
}

const latestPublishedBlingCorrelationSQL = `
SELECT b.correlation_id::text
  FROM integration_sync_batches b
  JOIN integration_connections c ON c.id = b.connection_id
 WHERE c.provider = 'BLING'
   AND b.resource IN ('caixas', 'contas.pagar', 'contas.receber')
   AND b.correlation_id IS NOT NULL
   AND b.status IN ('SUCCESS', 'PARTIAL')
   AND b.data_stage = 'NORMALIZED'
   AND b.finished_at <= $1
 GROUP BY b.correlation_id
HAVING COUNT(DISTINCT b.resource) = 3
 ORDER BY MAX(b.finished_at) DESC, b.correlation_id DESC
 LIMIT 1
`

// The screen uses the newest fully published daily snapshot, including a
// manually rerun snapshot created before the next 09:30 slot. During an
// incomplete run, it continues using the previous complete set rather than
// combining sources from different runs. If no complete snapshot exists yet,
// use the scheduled slot so existing freshness handling remains fail-closed.
func latestPublishedBlingCorrelationID(ctx context.Context, tx pgx.Tx, asOf time.Time) (string, error) {
	var correlationID string
	err := tx.QueryRow(ctx, latestPublishedBlingCorrelationSQL, asOf.UTC()).Scan(&correlationID)
	if err == nil {
		return correlationID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	return dailyBlingCorrelationID(asOf)
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

// combinedReceivableMetric only includes sources that completed the current
// daily import. It never mixes a current source with a stale subtotal; the
// approved customer-name + gross-value de-duplication is used when both are current.
func combinedReceivableMetric(combinedValue string, combinedCount int64, blingValue string, blingCount int64, nuvemValue string, nuvemCount int64, blingFresh, blingPartial, nuvemFresh bool) application.DashboardMetric {
	const source = "BLING + NUVEM_PAGO"
	if !blingFresh && !nuvemFresh {
		return application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: source}
	}
	if blingFresh && nuvemFresh {
		state := "CONFIRMED"
		if blingPartial {
			state = "PARTIAL"
		}
		return metric(combinedValue, combinedCount, state, source)
	}
	if blingFresh {
		return metric(blingValue, blingCount, "PARTIAL", source)
	}
	return metric(nuvemValue, nuvemCount, "PARTIAL", source)
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

// Cross-source receivable matches use the approved customer-name + gross-value
// rule. Dates are intentionally excluded. Pair equal occurrences one-to-one so
// repeated sales within either source remain separate; when sources disagree
// on multiplicity, the larger source count is retained. Nuvem Pago is the
// representative row for matched pairs because it carries the projected
// settlement value/date; the Bling row is not added a second time.
const dashboardEligibleReceivablesCTE = `
latest_daily_receivable_snapshot AS (
  SELECT b.id AS sync_batch_id
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING'
     AND b.resource = 'contas.receber'
     AND b.correlation_id IS NOT NULL
     AND b.status IN ('SUCCESS', 'PARTIAL')
     AND b.data_stage = 'NORMALIZED'
   ORDER BY b.finished_at DESC, b.id DESC
   LIMIT 1
), eligible_receivables AS (
  SELECT r.*,
         COALESCE(NULLIF(regexp_replace(translate(lower(btrim(c.name)), 'áàâãäéèêëíìîïóòõôöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn'), '[[:space:]]+', ' ', 'g'), ''),
                  NULLIF(regexp_replace(translate(lower(btrim(rr.payload->>'customer_name')), 'áàâãäéèêëíìîïóòõôöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn'), '[[:space:]]+', ' ', 'g'), '')) AS customer_match_name,
         COALESCE(r.due_date, r.expected_receipt_date) AS effective_due_date
    FROM receivables r
    LEFT JOIN contacts c ON c.id = r.customer_id
    LEFT JOIN raw_records rr ON rr.id = r.raw_record_id
   WHERE (r.source_system = 'NUVEM_PAGO' AND r.business_type = 'B2C' AND r.status = 'PROJECTED')
      OR (r.source_system = 'BLING'
          AND (r.source_entity <> 'bling.conta_receber' OR EXISTS (
            SELECT 1
              FROM latest_daily_receivable_snapshot latest
              JOIN receivable_sync_snapshot_rows membership ON membership.sync_batch_id = latest.sync_batch_id
             WHERE membership.receivable_id = r.id
          )))
), ranked_receivables AS (
  SELECT er.*,
         row_number() OVER (
           PARTITION BY customer_match_name, gross_amount, source_system
           ORDER BY source_entity, source_id, id
         ) AS source_match_ordinal
    FROM eligible_receivables er
), deduplicated_receivables AS (
  SELECT bling.*
    FROM ranked_receivables bling
   WHERE bling.source_system = 'BLING'
     AND (bling.customer_match_name IS NULL OR NOT EXISTS (
       SELECT 1
         FROM ranked_receivables nuvem
        WHERE nuvem.source_system = 'NUVEM_PAGO'
          AND nuvem.customer_match_name = bling.customer_match_name
          AND nuvem.gross_amount = bling.gross_amount
          AND nuvem.source_match_ordinal = bling.source_match_ordinal
     ))
  UNION ALL
  SELECT nuvem.*
    FROM ranked_receivables nuvem
   WHERE nuvem.source_system = 'NUVEM_PAGO'
)
`

const dashboardStreamSummarySQL = `WITH ` + dashboardEligibleReceivablesCTE + `, r AS (
  SELECT
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE source_system = 'BLING'), 0)::text AS bling_value,
    COUNT(id) FILTER (WHERE source_system = 'BLING') AS bling_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED'
    ), 0)::text AS future_value,
    COUNT(id) FILTER (
      WHERE source_system = 'NUVEM_PAGO' AND business_type = 'B2C' AND status = 'PROJECTED'
    ) AS future_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND effective_due_date < $1::date
    ), 0)::text AS overdue_value,
    COUNT(id) FILTER (
      WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND effective_due_date < $1::date
    ) AS overdue_count
    ,COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE source_system = 'BLING' AND COALESCE(open_balance, net_amount, gross_amount) > 0 AND effective_due_date < $1::date
    ), 0)::text AS bling_overdue_value,
    COUNT(id) FILTER (
      WHERE source_system = 'BLING' AND COALESCE(open_balance, net_amount, gross_amount) > 0 AND effective_due_date < $1::date
    ) AS bling_overdue_count
  FROM deduplicated_receivables
), due90_candidates AS (
  SELECT *
    FROM eligible_receivables
   WHERE COALESCE(open_balance, net_amount, gross_amount) > 0
     AND effective_due_date >= $1::date AND effective_due_date < ($1::date + 90)
     AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
), due90_ranked AS (
  SELECT candidate.*,
         row_number() OVER (
           PARTITION BY customer_match_name, gross_amount, source_system
           ORDER BY source_entity, source_id, id
         ) AS source_match_ordinal
    FROM due90_candidates candidate
), due90_deduplicated AS (
  SELECT bling.*
    FROM due90_ranked bling
   WHERE bling.source_system = 'BLING'
     AND (bling.customer_match_name IS NULL OR NOT EXISTS (
       SELECT 1 FROM due90_ranked nuvem
        WHERE nuvem.source_system = 'NUVEM_PAGO'
          AND nuvem.customer_match_name = bling.customer_match_name
          AND nuvem.gross_amount = bling.gross_amount
          AND nuvem.source_match_ordinal = bling.source_match_ordinal
     ))
  UNION ALL
  SELECT nuvem.*
    FROM due90_ranked nuvem
   WHERE nuvem.source_system = 'NUVEM_PAGO'
), due90 AS (
  SELECT COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)), 0)::text AS value,
         COUNT(id) AS count,
         COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE source_system = 'BLING'), 0)::text AS bling_value,
         COUNT(id) FILTER (WHERE source_system = 'BLING') AS bling_count,
         COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE source_system = 'NUVEM_PAGO'), 0)::text AS nuvem_value,
         COUNT(id) FILTER (WHERE source_system = 'NUVEM_PAGO') AS nuvem_count,
         COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE effective_due_date < ($1::date + 30)), 0)::text AS due30_value,
         COUNT(id) FILTER (WHERE effective_due_date < ($1::date + 30)) AS due30_count,
         COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE source_system = 'BLING' AND effective_due_date < ($1::date + 30)), 0)::text AS due30_bling_value,
         COUNT(id) FILTER (WHERE source_system = 'BLING' AND effective_due_date < ($1::date + 30)) AS due30_bling_count,
         COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE source_system = 'NUVEM_PAGO' AND effective_due_date < ($1::date + 30)), 0)::text AS due30_nuvem_value,
         COUNT(id) FILTER (WHERE source_system = 'NUVEM_PAGO' AND effective_due_date < ($1::date + 30)) AS due30_nuvem_count
    FROM due90_deduplicated
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
    ) AS due_next_90_days_count,
    COALESCE(SUM(open_balance) FILTER (
      WHERE due_date >= $1::date AND due_date < ($1::date + 30)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ), 0)::text AS due_next_30_days_value,
    COUNT(id) FILTER (
      WHERE due_date >= $1::date AND due_date < ($1::date + 30)
        AND UPPER(status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
    ) AS due_next_30_days_count
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
    MAX(b.finished_at) FILTER (WHERE b.resource = 'contas.receber' AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED') AS receivables_finished_at,
    COALESCE(BOOL_OR(b.status = 'PARTIAL') FILTER (WHERE b.resource = 'contas.receber' AND b.data_stage = 'NORMALIZED'), false) AS receivables_partial,
    MAX(b.finished_at) FILTER (WHERE b.resource = 'contas.pagar' AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED') AS payables_finished_at,
    COALESCE(BOOL_OR(b.status = 'PARTIAL') FILTER (WHERE b.resource = 'contas.pagar' AND b.data_stage = 'NORMALIZED'), false) AS payables_partial,
    MAX(b.finished_at) FILTER (WHERE b.resource = 'caixas' AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED') AS cash_movements_finished_at,
    COALESCE(BOOL_OR(b.status = 'PARTIAL') FILTER (WHERE b.resource = 'caixas' AND b.data_stage = 'NORMALIZED'), false) AS cash_movements_partial
    ,NOT EXISTS (
      SELECT 1
        FROM integration_sync_batches nuvem
        JOIN integration_connections nuvem_connection ON nuvem_connection.id = nuvem.connection_id
       WHERE nuvem_connection.provider = 'NUVEM_PAGO'
         AND nuvem.resource = 'nuvem_pago_recebimentos_futuros_csv_v1'
         AND nuvem.started_at >= ($1::date::timestamp AT TIME ZONE 'America/Sao_Paulo')
         AND nuvem.status = 'SUCCESS'
         AND nuvem.data_stage = 'NORMALIZED'
    ) AS nuvem_future_partial
    FROM (SELECT 1) seed
    LEFT JOIN (
      SELECT b.*
        FROM integration_sync_batches b
        JOIN integration_connections c ON c.id = b.connection_id
       WHERE c.provider = 'BLING'
         AND b.correlation_id = $4
    ) b ON true
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
       due90.value, due90.count,
       due90.due30_value, due90.due30_count,
       p.due_next_90_days_value, p.due_next_90_days_count,
       p.due_next_30_days_value, p.due_next_30_days_count,
       current_daily_bling_sync.nuvem_future_partial,
       r.bling_value, r.bling_count, r.bling_overdue_value, r.bling_overdue_count,
       due90.bling_value, due90.bling_count, due90.nuvem_value, due90.nuvem_count,
       due90.due30_bling_value, due90.due30_bling_count, due90.due30_nuvem_value, due90.due30_nuvem_count,
       current_daily_bling_sync.receivables_partial
FROM r CROSS JOIN due90 CROSS JOIN rc CROSS JOIN p CROSS JOIN pm CROSS JOIN last_bling_sync CROSS JOIN current_daily_bling_sync
`

const dashboardReceivableRowsSQL = `
WITH ` + dashboardEligibleReceivablesCTE + `
SELECT CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.customer END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.origin END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.due_date END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.gross_value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.net_value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.status END,
       limits.exceeds_limit
  FROM (SELECT * FROM deduplicated_receivables
         WHERE COALESCE(open_balance, net_amount, gross_amount) > 0
         ORDER BY effective_due_date NULLS LAST, source_id LIMIT 50) er
  LEFT JOIN contacts c ON c.id = er.customer_id
  LEFT JOIN raw_records rr ON rr.id = er.raw_record_id
 CROSS JOIN LATERAL (
   SELECT COALESCE(c.name, rr.payload->>'customer_name', '') AS customer,
          CASE er.source_system WHEN 'NUVEM_PAGO' THEN 'Nuvem Pago' WHEN 'BLING' THEN 'Bling' ELSE er.source_system END AS origin,
          COALESCE(to_char(er.effective_due_date, 'YYYY-MM-DD'), '') AS due_date,
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
