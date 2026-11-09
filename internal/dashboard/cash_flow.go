package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
)

const latestApprovedItauBalanceCheckpointSQL = `
SELECT checkpoint.connection_id::text, checkpoint.account_source_id,
       checkpoint.reconciled_amount::text, checkpoint.reference_date
  FROM bling_cash_balance_checkpoints checkpoint
  JOIN integration_connections connection ON connection.id = checkpoint.connection_id
 WHERE connection.provider = 'BLING'
   AND checkpoint.source = 'USER_APPROVED_BLING_SCREENSHOT'
   AND checkpoint.reference_date <= $1::date
 ORDER BY checkpoint.reference_date DESC, checkpoint.created_at DESC
 LIMIT 1
`

const reconciledCashMovementDeltaSQL = `
WITH current_reconciliation_batch AS (
  SELECT id, connection_id
    FROM integration_sync_batches
   WHERE correlation_id = $1 AND connection_id = $5::uuid AND resource = 'caixas.conciliadas'
     AND status = 'SUCCESS' AND data_stage = 'NORMALIZED'
   ORDER BY finished_at DESC NULLS LAST, id DESC LIMIT 1
), current_movement_batch AS (
  SELECT id, connection_id
    FROM integration_sync_batches
   WHERE correlation_id = $1 AND connection_id = $5::uuid AND resource = 'caixas'
     AND status IN ('SUCCESS', 'PARTIAL') AND data_stage = 'NORMALIZED'
   ORDER BY finished_at DESC NULLS LAST, id DESC LIMIT 1
), reconciled_movements AS (
  SELECT m.movement_date, m.direction, m.amount
    FROM current_reconciliation_batch reconciliation_batch
    JOIN bling_cash_movement_reconciled_snapshot_rows membership
      ON membership.sync_batch_id = reconciliation_batch.id
     AND membership.connection_id = reconciliation_batch.connection_id
    JOIN current_movement_batch movement_batch
      ON movement_batch.connection_id = reconciliation_batch.connection_id
    JOIN bling_cash_movements m
      ON m.sync_batch_id = movement_batch.id
     AND m.source_id = membership.source_id
     AND m.account_source_id = $2::text
   WHERE m.movement_date > $3::date AND m.movement_date <= $4::date
)
SELECT EXISTS (SELECT 1 FROM current_reconciliation_batch),
       COALESCE((SELECT SUM(CASE WHEN direction = 'INFLOW' THEN amount ELSE -amount END) FROM reconciled_movements), 0)::text,
       (SELECT MAX(movement_date) FROM reconciled_movements)
`

const reconciledCashMovementSnapshotExistsSQL = `
SELECT to_regclass('public.bling_cash_movement_reconciled_snapshot_rows') IS NOT NULL
`

// T02 mirrors the Bling Contas a Receber/Contas a Pagar reports, grouped by due
// date. Receivables use the displayed liquid value calculated from the
// same-run official payment-method terms. Do not add the external Nuvem file to
// T02: the Bling receivables report already includes Nuvem Pago payment methods.
// Do not source values from Bling's prepared Flow of Cash report.
const cashFlowCandidatesSQL = `
WITH ` + dashboardEligibleReceivablesCTE + `, daily AS (
  SELECT DISTINCT ON (b.resource) b.resource, b.id AS batch_id, b.status
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING' AND b.correlation_id = $1
     AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED'
   ORDER BY b.resource, b.finished_at DESC NULLS LAST, b.id DESC
), current_payables AS (
  SELECT p.*
    FROM daily d
    JOIN payable_sync_snapshot_rows membership ON membership.sync_batch_id = d.batch_id
    JOIN payables p ON p.id = membership.payable_id
   WHERE d.resource = 'contas.pagar'
), candidates AS (
  SELECT 'RECEIVABLE'::text AS kind, r.effective_due_date AS flow_date, 'INFLOW'::text AS direction,
         COALESCE(r.liquid_amount, 0)::text AS amount,
         ''::text AS source_status, r.source_financial_account_id::text AS account_source_id,
         r.status AS financial_status, COALESCE(r.source_payment_method_id::text, '') AS payment_method_id,
         NULL::text AS fee_rate, NULL::text AS fixed_fee, NULL::integer AS settlement_days,
         NULL::boolean AS use_business_days, NULL::smallint AS destination,
         false AS fee_rate_known, false AS fixed_fee_known, false AS settlement_known,
         false AS business_days_known, false AS destination_known,
         COALESCE((SELECT status FROM daily WHERE resource = 'contas.receber'), 'SUCCESS') AS batch_status
    FROM eligible_receivables r
   WHERE COALESCE(r.open_balance, r.gross_amount, 0) > 0
     AND r.liquid_amount > 0
     AND r.source_system = 'BLING'
     AND r.source_entity = 'bling.conta_receber'
     AND r.status IN ('OPEN', 'PARTIALLY_PAID')
     AND r.effective_due_date >= $2::date AND r.effective_due_date <= $3::date
  UNION ALL
  SELECT 'PAYABLE', p.due_date, 'OUTFLOW', abs(p.open_balance)::text,
         ''::text, p.source_financial_account_id::text,
         p.status, COALESCE(p.source_payment_method_id::text, ''),
         NULL::text, NULL::text, NULL::integer, NULL::boolean, NULL::smallint,
         false, false, false, false, false,
         d.status
    FROM daily d
    JOIN current_payables p ON true
   WHERE d.resource = 'contas.pagar'
     AND p.source_system = 'BLING' AND p.source_entity = 'bling.conta_pagar'
     AND p.open_balance <> 0
     AND UPPER(p.status) NOT IN ('5', 'CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
     AND p.due_date >= $4::date AND p.due_date <= $3::date
)
SELECT kind, flow_date, direction, amount, source_status,
       COALESCE(account_source_id, ''), financial_status,
       COALESCE(payment_method_id, ''), COALESCE(fee_rate, ''), COALESCE(fixed_fee, ''),
       settlement_days, use_business_days, destination,
       fee_rate_known, fixed_fee_known, settlement_known, business_days_known,
       destination_known, batch_status
  FROM candidates
 ORDER BY flow_date, kind
`

type approvedItauBalance struct {
	Metric        application.DashboardMetric
	Projection    application.DashboardMetric
	ReferenceDate time.Time
	BaseMetric    application.DashboardMetric
	BaseDate      time.Time
}

type cashFlowCandidate struct {
	kind, direction, amount, sourceStatus, accountID, financialStatus string
	methodID, feeRate, fixedFee                                       string
	flowDate                                                          time.Time
	settlementDays                                                    pgtype.Int4
	useBusinessDays                                                   pgtype.Bool
	destination                                                       pgtype.Int2
	feeRateKnown, fixedFeeKnown, settlementKnown, businessDaysKnown   bool
	destinationKnown, batchPartial                                    bool
}

type cashFlowDayAggregate struct {
	inflows, outflows domain.Money
}

func readLatestItauBalance(ctx context.Context, tx pgx.Tx, asOf time.Time, correlationID string) (approvedItauBalance, error) {
	var connectionID, accountID, value string
	var referenceDate time.Time
	if err := tx.QueryRow(ctx, latestApprovedItauBalanceCheckpointSQL, dashboardDateOnly(asOf).Format(time.DateOnly)).Scan(&connectionID, &accountID, &value, &referenceDate); err != nil {
		if err == pgx.ErrNoRows {
			return approvedItauBalance{Metric: application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}}, nil
		}
		return approvedItauBalance{}, err
	}
	base := metric(value, 1, "CONFIRMED", "BLING")
	baseDate := dashboardDateOnly(referenceDate)
	ready, delta, latestDate, err := reconciledCashMovementDelta(ctx, tx, correlationID, connectionID, accountID, baseDate, dashboardDateOnly(asOf))
	if err != nil {
		return approvedItauBalance{}, err
	}
	result := approvedItauBalance{
		Metric:        base,
		Projection:    application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
		ReferenceDate: baseDate,
		BaseMetric:    base,
		BaseDate:      baseDate,
	}
	if !ready {
		return result, nil
	}
	updated, parseErr := domain.ParseMoney(base.Value)
	if parseErr != nil {
		return approvedItauBalance{}, fmt.Errorf("approved Itaú balance is invalid")
	}
	movementDelta, parseErr := domain.ParseMoney(delta)
	if parseErr != nil {
		return approvedItauBalance{}, fmt.Errorf("reconciled Itaú movement total is invalid")
	}
	updated, err = updated.Add(movementDelta)
	if err != nil {
		return approvedItauBalance{}, err
	}
	result.Metric = metric(updated.String(), 1, "CONFIRMED", "BLING")
	result.Projection = result.Metric
	if latestDate.Valid {
		result.ReferenceDate = dashboardDateOnly(latestDate.Time)
	}
	return result, nil
}

func reconciledCashMovementDelta(ctx context.Context, tx pgx.Tx, correlationID, connectionID, accountID string, after, through time.Time) (bool, string, pgtype.Date, error) {
	if connectionID == "" || accountID == "" || after.IsZero() || through.Before(after) {
		return false, "0", pgtype.Date{}, nil
	}
	var snapshotTableExists bool
	if err := tx.QueryRow(ctx, reconciledCashMovementSnapshotExistsSQL).Scan(&snapshotTableExists); err != nil {
		return false, "0", pgtype.Date{}, err
	}
	if !snapshotTableExists {
		// Do not issue a query that references a missing optional table: PostgreSQL
		// marks the whole transaction failed even if the caller ignores SQLSTATE 42P01.
		return false, "0", pgtype.Date{}, nil
	}
	var ready bool
	var delta string
	var latestDate pgtype.Date
	if err := tx.QueryRow(ctx, reconciledCashMovementDeltaSQL, correlationID, accountID, after.Format(time.DateOnly), through.Format(time.DateOnly), connectionID).Scan(&ready, &delta, &latestDate); err != nil {
		return false, "0", pgtype.Date{}, err
	}
	return ready, delta, latestDate, nil
}

func readBlingCashFlow(ctx context.Context, tx pgx.Tx, asOf time.Time, correlationID string, itauBalance approvedItauBalance) (*application.DashboardCashFlow, error) {
	// readStreamSummary passes a date-only business date represented at UTC
	// midnight. Convert only to date here; converting that normalized value to
	// Sao Paulo again shifts it to the previous calendar day (21:00 BRT).
	start := dashboardDateOnly(asOf)
	end := start.AddDate(0, 0, 60)
	rows, err := tx.Query(ctx, cashFlowCandidatesSQL, correlationID, start.Format(time.DateOnly), end.Format(time.DateOnly), start.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var daily [61]cashFlowDayAggregate
	for rows.Next() {
		candidate, err := scanCashFlowCandidate(rows)
		if err != nil {
			return nil, err
		}
		amount, err := domain.ParseMoney(candidate.amount)
		if err != nil {
			return nil, fmt.Errorf("cash-flow candidate contains invalid amount")
		}
		flowDate := candidate.flowDate
		switch candidate.kind {
		case "PAYABLE":
			amount, err = amount.AbsExact()
			if err != nil {
				return nil, err
			}
		case "RECEIVABLE":
			// Receivables already carry their positive open balance.
		default:
			continue
		}
		day := dashboardDateOnly(flowDate)
		if day.Before(start) {
			continue
		}
		if day.After(end) {
			continue
		}
		index := int(day.Sub(start) / (24 * time.Hour))
		if candidate.direction == "INFLOW" {
			daily[index].inflows, err = daily[index].inflows.Add(amount)
		} else if candidate.direction == "OUTFLOW" {
			daily[index].outflows, err = daily[index].outflows.Add(amount)
		} else {
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	flow := &application.DashboardCashFlow{Days: make([]application.DashboardCashFlowDay, 0, len(daily))}
	inflows30, outflows30 := domain.ZeroMoney(), domain.ZeroMoney()
	inflows60, outflows60 := domain.ZeroMoney(), domain.ZeroMoney()
	for index := range daily {
		date := start.AddDate(0, 0, index)
		if index < 60 {
			inflows60, err = inflows60.Add(daily[index].inflows)
			if err == nil {
				outflows60, err = outflows60.Sub(daily[index].outflows)
			}
			if err != nil {
				return nil, err
			}
		}
		if index < 30 {
			inflows30, err = inflows30.Add(daily[index].inflows)
			if err == nil {
				outflows30, err = outflows30.Sub(daily[index].outflows)
			}
			if err != nil {
				return nil, err
			}
		}
		outflow, err := domain.ZeroMoney().Sub(daily[index].outflows)
		if err != nil {
			return nil, err
		}
		flow.Days = append(flow.Days, application.DashboardCashFlowDay{
			Date:     date.Format(time.DateOnly),
			Inflows:  dashboardFlowMetric(daily[index].inflows.String(), false),
			Outflows: dashboardFlowMetric(outflow.String(), false),
		})
	}
	flow.InflowsNext30Days = dashboardFlowMetric(inflows30.String(), false)
	flow.OutflowsNext30Days = dashboardFlowMetric(outflows30.String(), false)
	flow.InflowsNext60Days = dashboardFlowMetric(inflows60.String(), false)
	flow.OutflowsNext60Days = dashboardFlowMetric(outflows60.String(), false)
	flow.Periods, err = groupCashFlowReportWeeks(flow.Days, false)
	if err != nil {
		return nil, err
	}
	projectedOpening := itauBalance.Projection
	if projectedOpening.Value == "" || itauBalance.ReferenceDate.IsZero() {
		projectedOpening = application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}
	}
	if err := applyWeeklyCashBalances(flow.Periods, projectedOpening, false); err != nil {
		return nil, err
	}
	flow.CriticalWeek, err = selectCriticalCashFlowWeek(flow.Periods)
	if err != nil {
		return nil, err
	}
	flow.MinimumBalance30Days, flow.MinimumBalance60Days, err = minimumCashFlowBalances(daily[:], start, projectedOpening, false)
	if err != nil {
		return nil, err
	}
	return flow, nil
}

func applyWeeklyCashBalances(periods []application.DashboardCashFlowPeriod, opening application.DashboardMetric, partial bool) error {
	if opening.State == "UNAVAILABLE" || opening.Value == "" {
		return nil
	}
	balance, err := domain.ParseMoney(opening.Value)
	if err != nil {
		return fmt.Errorf("approved opening balance is invalid")
	}
	state := "PROJECTED"
	if partial || opening.State == "PARTIAL" {
		state = "PARTIAL"
	}
	for index := range periods {
		period := &periods[index]
		inflow, err := domain.ParseMoney(period.Inflows.Value)
		if err != nil {
			return fmt.Errorf("cash-flow period contains invalid inflows")
		}
		outflow, err := domain.ParseMoney(period.Outflows.Value)
		if err != nil {
			return fmt.Errorf("cash-flow period contains invalid outflows")
		}
		period.OpeningBalance = ptrMetric(metric(balance.String(), 1, state, "BLING"))
		balance, err = balance.Add(inflow)
		if err == nil {
			balance, err = balance.Add(outflow)
		}
		if err != nil {
			return err
		}
		period.ClosingBalance = ptrMetric(metric(balance.String(), 1, state, "BLING"))
	}
	return nil
}

func ptrMetric(value application.DashboardMetric) *application.DashboardMetric { return &value }

const approvedCashReserve = "50000.00"

func minimumCashFlowBalances(days []cashFlowDayAggregate, start time.Time, projectionOpening application.DashboardMetric, partial bool) (*application.DashboardMinimumBalance, *application.DashboardMinimumBalance, error) {
	if projectionOpening.State == "UNAVAILABLE" || projectionOpening.Value == "" || len(days) < 60 {
		return unavailableMinimumBalance(), unavailableMinimumBalance(), nil
	}
	opening, err := domain.ParseMoney(projectionOpening.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("current account balance is invalid")
	}
	reserve, err := domain.ParseMoney(approvedCashReserve)
	if err != nil {
		return nil, nil, err
	}
	makeHorizon := func(horizon int) (*application.DashboardMinimumBalance, error) {
		balance := opening
		var minimum domain.Money
		var minimumDate time.Time
		for index := 0; index < horizon; index++ {
			inflow, outflow := days[index].inflows, days[index].outflows
			balance, err = balance.Add(inflow)
			if err == nil {
				balance, err = balance.Sub(outflow)
			}
			if err != nil {
				return nil, err
			}
			if index == 0 || balance.Compare(minimum) < 0 {
				minimum = balance
				minimumDate = start.AddDate(0, 0, index)
			}
		}
		headroom, err := minimum.Sub(reserve)
		if err != nil {
			return nil, err
		}
		state := "PROJECTED"
		if partial || projectionOpening.State == "PARTIAL" {
			state = "PARTIAL"
		}
		return &application.DashboardMinimumBalance{
			Balance:  metric(minimum.String(), 1, state, "BLING"),
			Date:     minimumDate.Format(time.DateOnly),
			Headroom: metric(headroom.String(), 1, state, "MAJUCAU_RULE"),
		}, nil
	}
	minimum30, err := makeHorizon(30)
	if err != nil {
		return nil, nil, err
	}
	minimum60, err := makeHorizon(60)
	if err != nil {
		return nil, nil, err
	}
	return minimum30, minimum60, nil
}

func unavailableMinimumBalance() *application.DashboardMinimumBalance {
	return &application.DashboardMinimumBalance{
		Balance:  application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
		Headroom: application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "MAJUCAU_RULE"},
	}
}

// selectCriticalCashFlowWeek returns the report period with the greatest
// absolute outflow, matching the approved T02 description. Partial values
// remain partial; unavailable weeks are not considered.
func selectCriticalCashFlowWeek(periods []application.DashboardCashFlowPeriod) (*application.DashboardCashFlowPeriod, error) {
	var selected *application.DashboardCashFlowPeriod
	var greatestOutflow domain.Money
	for index := range periods {
		metric := periods[index].Outflows
		if metric.State == "UNAVAILABLE" || metric.Value == "" {
			continue
		}
		outflow, err := domain.ParseMoney(metric.Value)
		if err != nil {
			return nil, fmt.Errorf("cash-flow report contains invalid weekly outflow")
		}
		outflow, err = outflow.AbsExact()
		if err != nil {
			return nil, err
		}
		if selected == nil || outflow.Compare(greatestOutflow) > 0 {
			selected = &periods[index]
			greatestOutflow = outflow
		}
	}
	return selected, nil
}

// groupCashFlowReportWeeks follows the Bling report's week boundaries: the
// first bucket starts on the requested start date and ends on Saturday; the
// following buckets run Sunday through Saturday. Values remain exact decimals.
func groupCashFlowReportWeeks(days []application.DashboardCashFlowDay, partial bool) ([]application.DashboardCashFlowPeriod, error) {
	if len(days) == 0 {
		return nil, nil
	}
	periods := make([]application.DashboardCashFlowPeriod, 0, (len(days)+6)/7)
	var current application.DashboardCashFlowPeriod
	inflows, outflows := domain.ZeroMoney(), domain.ZeroMoney()
	for index, day := range days {
		date, err := time.Parse(time.DateOnly, day.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid cash-flow report date")
		}
		if current.From == "" {
			current.From = day.Date
		}
		current.Through = day.Date
		inflow, err := domain.ParseMoney(day.Inflows.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid cash-flow report inflow")
		}
		outflow, err := domain.ParseMoney(day.Outflows.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid cash-flow report outflow")
		}
		inflows, err = inflows.Add(inflow)
		if err != nil {
			return nil, err
		}
		outflows, err = outflows.Add(outflow)
		if err != nil {
			return nil, err
		}
		if date.Weekday() == time.Saturday || index == len(days)-1 {
			current.Inflows = dashboardFlowMetric(inflows.String(), partial)
			current.Outflows = dashboardFlowMetric(outflows.String(), partial)
			periods = append(periods, current)
			current = application.DashboardCashFlowPeriod{}
			inflows, outflows = domain.ZeroMoney(), domain.ZeroMoney()
		}
	}
	return periods, nil
}

func dashboardDateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func scanCashFlowCandidate(rows pgx.Rows) (cashFlowCandidate, error) {
	var candidate cashFlowCandidate
	var status pgtype.Text
	if err := rows.Scan(
		&candidate.kind, &candidate.flowDate, &candidate.direction, &candidate.amount,
		&candidate.sourceStatus, &candidate.accountID, &candidate.financialStatus,
		&candidate.methodID, &candidate.feeRate, &candidate.fixedFee,
		&candidate.settlementDays, &candidate.useBusinessDays, &candidate.destination,
		&candidate.feeRateKnown, &candidate.fixedFeeKnown, &candidate.settlementKnown,
		&candidate.businessDaysKnown, &candidate.destinationKnown, &status,
	); err != nil {
		return cashFlowCandidate{}, err
	}
	candidate.batchPartial = status.Valid && status.String == "PARTIAL"
	return candidate, nil
}

func dashboardFlowMetric(value string, partial bool) application.DashboardMetric {
	state := "CONFIRMED"
	if partial {
		state = "PARTIAL"
	}
	return application.DashboardMetric{Value: value, State: state, SourceSystem: "BLING"}
}
