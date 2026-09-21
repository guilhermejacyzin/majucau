package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/application"
)

var ErrDatabaseUnavailable = errors.New("dashboard database is unavailable")

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Reader struct {
	db  rowQuerier
	now func() time.Time
}

func NewReader(pool *pgxpool.Pool) *Reader {
	return &Reader{db: pool, now: func() time.Time { return time.Now().UTC() }}
}

// ReadDashboardSnapshot reads only normalized tables. In particular, it never
// treats Bling RAW API pages as financial facts and never writes anything.
func (r *Reader) ReadDashboardSnapshot(ctx context.Context) (application.DashboardSnapshot, error) {
	if r == nil || r.db == nil {
		return application.DashboardSnapshot{}, ErrDatabaseUnavailable
	}
	asOf := r.now().UTC()
	monthStart := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var (
		receivablesValue, futureValue, receiptsMonthValue, receivablesOverdueValue       string
		payablesValue, payablesTodayValue, payablesOverdueValue, paymentsMonthValue      string
		receivablesCount, futureCount, receiptsTotalCount, receiptsMonthCount            int64
		receivablesOverdueCount, payablesCount, payablesTodayCount, payablesOverdueCount int64
		paymentsTotalCount, paymentsMonthCount                                           int64
	)

	err := r.db.QueryRow(ctx, dashboardSnapshotSQL, asOf, monthStart, monthEnd).Scan(
		&receivablesValue, &receivablesCount,
		&futureValue, &futureCount,
		&receiptsMonthValue, &receiptsMonthCount, &receiptsTotalCount,
		&receivablesOverdueValue, &receivablesOverdueCount,
		&payablesValue, &payablesCount,
		&payablesTodayValue, &payablesTodayCount,
		&payablesOverdueValue, &payablesOverdueCount,
		&paymentsMonthValue, &paymentsMonthCount, &paymentsTotalCount,
	)
	if err != nil {
		return application.DashboardSnapshot{}, err
	}

	state := "UNAVAILABLE"
	if receivablesCount > 0 || futureCount > 0 || receiptsTotalCount > 0 || payablesCount > 0 || paymentsTotalCount > 0 {
		state = "PARTIAL"
	}
	return application.DashboardSnapshot{
		AsOf:               asOf,
		DataState:          state,
		Receivables:        metric(receivablesValue, receivablesCount, "CONFIRMED", "BLING + NUVEM_PAGO"),
		FutureB2C:          metric(futureValue, futureCount, projectedState(futureCount), "NUVEM_PAGO"),
		ReceiptsMonth:      metric(receiptsMonthValue, receiptsMonthCount, stateForCount(receiptsTotalCount), "BLING"),
		ReceivablesOverdue: metric(receivablesOverdueValue, receivablesOverdueCount, stateForCount(receivablesCount), "BLING + NUVEM_PAGO"),
		Payables:           metric(payablesValue, payablesCount, stateForCount(payablesCount), "BLING"),
		PayablesDueToday:   metric(payablesTodayValue, payablesTodayCount, stateForCount(payablesCount), "BLING"),
		PayablesOverdue:    metric(payablesOverdueValue, payablesOverdueCount, stateForCount(payablesCount), "BLING"),
		PaymentsMonth:      metric(paymentsMonthValue, paymentsMonthCount, stateForCount(paymentsTotalCount), "BLING"),
	}, nil
}

func metric(value string, count int64, state, source string) application.DashboardMetric {
	if count == 0 && state == "UNAVAILABLE" {
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

const dashboardSnapshotSQL = `
WITH r AS (
  SELECT
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (
      WHERE source_system = 'NUVEM_PAGO'
        AND business_type = 'B2C'
        AND status = 'PROJECTED'
    ), 0)::text AS future_value,
    COUNT(id) FILTER (
      WHERE source_system = 'NUVEM_PAGO'
        AND business_type = 'B2C'
        AND status = 'PROJECTED'
    ) AS future_count,
    COALESCE(SUM(COALESCE(open_balance, net_amount, gross_amount)) FILTER (WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND due_date < $1::date), 0)::text AS overdue_value,
    COUNT(id) FILTER (WHERE COALESCE(open_balance, net_amount, gross_amount) > 0 AND due_date < $1::date) AS overdue_count
  FROM receivables
), rc AS (
  SELECT
    COALESCE(SUM(amount) FILTER (
      WHERE source_system = 'BLING'
        AND status = 'CONFIRMED'
        AND receipt_date >= $2::date
        AND receipt_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE source_system = 'BLING'
        AND status = 'CONFIRMED'
        AND receipt_date >= $2::date
        AND receipt_date < $3::date
    ) AS month_count,
    COUNT(id) FILTER (WHERE source_system = 'BLING' AND status = 'CONFIRMED') AS total_count
  FROM receipts
), p AS (
  SELECT
    COALESCE(SUM(open_balance), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(open_balance) FILTER (WHERE due_date = $1::date), 0)::text AS today_value,
    COUNT(id) FILTER (WHERE due_date = $1::date) AS today_count,
    COALESCE(SUM(open_balance) FILTER (WHERE open_balance > 0 AND due_date < $1::date), 0)::text AS overdue_value,
    COUNT(id) FILTER (WHERE open_balance > 0 AND due_date < $1::date) AS overdue_count
  FROM payables
), pm AS (
  SELECT
    COALESCE(SUM(amount) FILTER (
      WHERE source_system = 'BLING'
        AND payment_date >= $2::date
        AND payment_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE source_system = 'BLING'
        AND payment_date >= $2::date
        AND payment_date < $3::date
    ) AS month_count,
    COUNT(id) FILTER (WHERE source_system = 'BLING') AS total_count
  FROM payments
)
SELECT r.total_value, r.total_count, r.future_value, r.future_count,
       rc.month_value, rc.month_count, rc.total_count,
       r.overdue_value, r.overdue_count,
       p.total_value, p.total_count, p.today_value, p.today_count,
       p.overdue_value, p.overdue_count,
       pm.month_value, pm.month_count, pm.total_count
FROM r CROSS JOIN rc CROSS JOIN p CROSS JOIN pm
`

