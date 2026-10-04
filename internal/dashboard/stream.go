package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
)

const (
	dashboardFieldMaxBytes  = 1 << 20
	dashboardRecordMaxBytes = 4 << 20
)

var ErrSnapshotRecordTooLarge = errors.New("dashboard snapshot record exceeds the configured limit")

type dashboardStreamDB interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

// WriteDashboardSnapshot emits the same dashboard data as a JSON document but
// reads and writes detail rows one at a time. The read-only repeatable-read
// transaction keeps summary metrics and both tables on the same database view.
func (r *Reader) WriteDashboardSnapshot(ctx context.Context, dst io.Writer) error {
	if r == nil || r.db == nil {
		return ErrDatabaseUnavailable
	}
	db, ok := r.db.(dashboardStreamDB)
	if !ok {
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
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

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
	if err := streamReceivableRows(ctx, tx, dst); err != nil {
		return err
	}
	if err := writeAll(dst, []byte(`],"payable_rows":[`)); err != nil {
		return err
	}
	if err := streamPayableRows(ctx, tx, dst); err != nil {
		return err
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
		receivablesValue, futureValue, receiptsMonthValue, receivablesOverdueValue       string
		payablesValue, payablesTodayValue, payablesOverdueValue, paymentsMonthValue      string
		receivablesCount, futureCount, receiptsTotalCount, receiptsMonthCount            int64
		receivablesOverdueCount, payablesCount, payablesTodayCount, payablesOverdueCount int64
		paymentsTotalCount, paymentsMonthCount                                           int64
	)
	err := tx.QueryRow(ctx, dashboardStreamSummarySQL,
		businessDate.Format("2006-01-02"),
		monthStart.Format("2006-01-02"),
		monthEnd.Format("2006-01-02"),
	).Scan(
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
  FROM eligible_receivables
), rc AS (
  SELECT
    COALESCE(SUM(amount) FILTER (
      WHERE source_system = 'BLING' AND status = 'CONFIRMED'
        AND receipt_date >= $2::date AND receipt_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE source_system = 'BLING' AND status = 'CONFIRMED'
        AND receipt_date >= $2::date AND receipt_date < $3::date
    ) AS month_count,
    COUNT(id) FILTER (WHERE source_system = 'BLING' AND status = 'CONFIRMED') AS total_count
  FROM receipts
), p AS (
  SELECT
    COALESCE(SUM(open_balance), 0)::text AS total_value,
    COUNT(id) AS total_count,
    COALESCE(SUM(open_balance) FILTER (WHERE due_date = $1::date), 0)::text AS today_value,
    COUNT(id) FILTER (WHERE due_date = $1::date) AS today_count,
    COALESCE(SUM(open_balance) FILTER (
      WHERE open_balance > 0 AND due_date < $1::date
    ), 0)::text AS overdue_value,
    COUNT(id) FILTER (WHERE open_balance > 0 AND due_date < $1::date) AS overdue_count
  FROM payables WHERE open_balance > 0
), pm AS (
  SELECT
    COALESCE(SUM(amount) FILTER (
      WHERE source_system = 'BLING' AND payment_date >= $2::date AND payment_date < $3::date
    ), 0)::text AS month_value,
    COUNT(id) FILTER (
      WHERE source_system = 'BLING' AND payment_date >= $2::date AND payment_date < $3::date
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
SELECT CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.supplier END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.document END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.due_date END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.value END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.category END,
       CASE WHEN limits.exceeds_limit THEN '' ELSE row_data.status END,
       limits.exceeds_limit
  FROM (SELECT * FROM payables WHERE open_balance > 0
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
