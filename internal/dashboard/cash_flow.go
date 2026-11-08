package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/financial"
)

const selectedCashFlowAccountsSQL = `
SELECT source_id, description
  FROM bling_financial_account_identities
 WHERE sync_batch_id = (
   SELECT b.id
     FROM integration_sync_batches b
     JOIN integration_connections c ON c.id = b.connection_id
    WHERE c.provider = 'BLING' AND b.resource = 'contas-contabeis'
      AND b.correlation_id = $1 AND b.status IN ('SUCCESS', 'PARTIAL')
      AND b.data_stage = 'NORMALIZED'
    ORDER BY b.finished_at DESC, b.id DESC LIMIT 1
 )
 ORDER BY source_id
`

// GET /caixas includes the financial-account ID and description with each
// movement. Use those official fields when the optional account catalog is
// unavailable or incomplete. The approved report screenshot selects one
// financial account; project only the unique Itaú account identified in the
// approved context. Keep the separate bank-balance card and report balance
// formulas independent; do not infer Nuvem transfers.
const selectedCashFlowMovementAccountsSQL = `
SELECT DISTINCT m.account_source_id, m.account_name
  FROM bling_cash_movements m
  JOIN integration_sync_batches b ON b.id = m.sync_batch_id
  JOIN integration_connections c ON c.id = b.connection_id
 WHERE c.provider = 'BLING' AND b.resource = 'caixas'
   AND b.correlation_id = $1 AND b.status IN ('SUCCESS', 'PARTIAL')
   AND b.data_stage = 'NORMALIZED'
   AND NULLIF(BTRIM(m.account_source_id), '') IS NOT NULL
   AND NULLIF(BTRIM(m.account_name), '') IS NOT NULL
 ORDER BY m.account_source_id, m.account_name
`

const latestApprovedItauBalanceCheckpointSQL = `
SELECT checkpoint.reconciled_amount::text, checkpoint.reference_date
  FROM bling_cash_balance_checkpoints checkpoint
  JOIN integration_connections connection ON connection.id = checkpoint.connection_id
 WHERE connection.provider = 'BLING'
   AND checkpoint.account_source_id = ANY($1::text[])
   AND checkpoint.reference_date <= $2::date
 ORDER BY checkpoint.reference_date DESC, checkpoint.created_at DESC
 LIMIT 1
`

const reconciledCashMovementDeltaSQL = `
WITH current_reconciliation_batch AS (
  SELECT id, connection_id
    FROM integration_sync_batches
   WHERE correlation_id = $1 AND resource = 'caixas.conciliadas'
     AND status = 'SUCCESS' AND data_stage = 'NORMALIZED'
   ORDER BY finished_at DESC NULLS LAST, id DESC LIMIT 1
), current_movement_batch AS (
  SELECT id, connection_id
    FROM integration_sync_batches
   WHERE correlation_id = $1 AND resource = 'caixas'
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
     AND m.account_source_id = ANY($2::text[])
   WHERE m.movement_date > $3::date AND m.movement_date <= $4::date
)
SELECT EXISTS (SELECT 1 FROM current_reconciliation_batch),
       COALESCE((SELECT SUM(CASE WHEN direction = 'INFLOW' THEN amount ELSE -amount END) FROM reconciled_movements), 0)::text,
       (SELECT MAX(movement_date) FROM reconciled_movements)
`

// The report comparison indicates one calendar day for forecast dates while
// the bank-float option is enabled. Registered cash movements already carry
// their actual movement date and must not be shifted into the report window.
// Keep forecasts PARTIAL until the account's saved float setting is verified.
const blingCashFlowEffectiveFloatDays = 1

// Bling documents the prepared Flow of Cash report as a projection from
// receivables and payables, but does not expose a public API route for that
// report. This query reads the normalized official API snapshots to reproduce
// its inputs; it returns individual
// candidates so the domain can apply payment-method terms without buffering a
// transaction history in memory. Missing or unsupported fields are retained as
// candidates and make the resulting metric partial instead of silently zero.
const cashFlowCandidatesSQL = `
WITH daily AS (
  SELECT DISTINCT ON (b.resource) b.resource, b.id AS batch_id, b.status
    FROM integration_sync_batches b
    JOIN integration_connections c ON c.id = b.connection_id
   WHERE c.provider = 'BLING' AND b.correlation_id = $1
     AND b.status IN ('SUCCESS', 'PARTIAL') AND b.data_stage = 'NORMALIZED'
   ORDER BY b.resource, b.finished_at DESC NULLS LAST, b.id DESC
), current_receivables AS (
  SELECT r.*
    FROM daily d
    JOIN receivable_sync_snapshot_rows membership ON membership.sync_batch_id = d.batch_id
    JOIN receivables r ON r.id = membership.receivable_id
   WHERE d.resource = 'contas.receber'
), current_payables AS (
  SELECT p.*
    FROM daily d
    JOIN payable_sync_snapshot_rows membership ON membership.sync_batch_id = d.batch_id
    JOIN payables p ON p.id = membership.payable_id
   WHERE d.resource = 'contas.pagar'
), candidates AS (
  SELECT 'MOVEMENT'::text AS kind, m.movement_date AS flow_date,
         m.direction, m.amount::text AS amount, m.source_status AS source_status,
         m.account_source_id, ''::text AS financial_status,
         NULL::text AS payment_method_id, NULL::text AS fee_rate,
         NULL::text AS fixed_fee, NULL::integer AS settlement_days,
         NULL::boolean AS use_business_days, NULL::smallint AS destination,
         false AS fee_rate_known, false AS fixed_fee_known,
         false AS settlement_known, false AS business_days_known,
         false AS destination_known, d.status AS batch_status
    FROM daily d
    JOIN bling_cash_movements m ON m.sync_batch_id = d.batch_id
   WHERE d.resource = 'caixas'
     AND m.movement_date >= $5::date AND m.movement_date <= $3::date
  UNION ALL
  SELECT 'RECEIVABLE', r.due_date, 'INFLOW', COALESCE(r.open_balance, r.gross_amount)::text,
         ''::text, r.source_financial_account_id::text,
         r.status, r.source_payment_method_id::text,
         terms.fee_rate::text, terms.fixed_fee::text, terms.settlement_days,
         terms.use_business_days, terms.destination,
         COALESCE(terms.fee_rate_known, false), COALESCE(terms.fixed_fee_known, false),
         COALESCE(terms.settlement_known, false), COALESCE(terms.business_days_known, false),
         terms.destination IS NOT NULL,
         d.status
    FROM daily d
    JOIN current_receivables r ON true
    LEFT JOIN bling_payment_method_terms terms
      ON terms.connection_id = r.connection_id
     AND terms.source_id = r.source_payment_method_id::text
     AND terms.sync_batch_id = (SELECT forms.batch_id FROM daily forms WHERE forms.resource = 'formas-pagamentos' LIMIT 1)
   WHERE d.resource = 'contas.receber'
     AND r.source_system = 'BLING' AND r.source_entity = 'bling.conta_receber'
     AND r.status IN ('OPEN', 'PARTIALLY_PAID') AND COALESCE(r.open_balance, 0) > 0
     AND r.due_date >= $2::date - 3650 AND r.due_date <= $3::date
  UNION ALL
  SELECT 'PAYABLE', p.due_date, 'OUTFLOW', abs(p.open_balance)::text,
         ''::text, p.source_financial_account_id::text,
         p.status, p.source_payment_method_id::text,
         NULL::text, NULL::text, NULL::integer, NULL::boolean,
         terms.destination,
         false, false, false, false, terms.destination IS NOT NULL,
         d.status
    FROM daily d
    JOIN current_payables p ON true
    LEFT JOIN bling_payment_method_terms terms
      ON terms.connection_id = p.connection_id
     AND terms.source_id = p.source_payment_method_id::text
     AND terms.sync_batch_id = (SELECT forms.batch_id FROM daily forms WHERE forms.resource = 'formas-pagamentos' LIMIT 1)
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
	accountIDs, accountsReady, err := selectedCashFlowAccountIDs(ctx, tx, correlationID)
	if err != nil {
		return approvedItauBalance{}, err
	}
	if !accountsReady {
		return approvedItauBalance{Metric: application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}}, nil
	}
	var value string
	var referenceDate time.Time
	if err := tx.QueryRow(ctx, latestApprovedItauBalanceCheckpointSQL, accountIDs, dashboardDateOnly(asOf).Format(time.DateOnly)).Scan(&value, &referenceDate); err != nil {
		if err == pgx.ErrNoRows {
			return approvedItauBalance{Metric: application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}}, nil
		}
		return approvedItauBalance{}, err
	}
	base := metric(value, 1, "PARTIAL", "BLING")
	baseDate := dashboardDateOnly(referenceDate)
	ready, delta, latestDate, err := reconciledCashMovementDelta(ctx, tx, correlationID, accountIDs, baseDate, dashboardDateOnly(asOf))
	if err != nil {
		return approvedItauBalance{}, err
	}
	card := base
	cardDate := baseDate
	if ready {
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
		card = metric(updated.String(), 1, "PARTIAL", "BLING")
		if latestDate.Valid {
			cardDate = dashboardDateOnly(latestDate.Time)
		}
	}
	return approvedItauBalance{Metric: card, ReferenceDate: cardDate, BaseMetric: base, BaseDate: baseDate}, nil
}

func reconciledCashMovementDelta(ctx context.Context, tx pgx.Tx, correlationID string, accountIDs []string, after, through time.Time) (bool, string, pgtype.Date, error) {
	if len(accountIDs) == 0 || after.IsZero() || through.Before(after) {
		return false, "0", pgtype.Date{}, nil
	}
	var ready bool
	var delta string
	var latestDate pgtype.Date
	if err := tx.QueryRow(ctx, reconciledCashMovementDeltaSQL, correlationID, accountIDs, after.Format(time.DateOnly), through.Format(time.DateOnly)).Scan(&ready, &delta, &latestDate); err != nil {
		return false, "0", pgtype.Date{}, err
	}
	return ready, delta, latestDate, nil
}

func readBlingCashFlow(ctx context.Context, tx pgx.Tx, asOf time.Time, correlationID string, checkpoint approvedItauBalance) (*application.DashboardCashFlow, error) {
	start := domain.BusinessDate(asOf)
	end := start.AddDate(0, 0, 60)
	accountIDs, accountsReady, err := selectedCashFlowAccountIDs(ctx, tx, correlationID)
	if err != nil {
		return nil, err
	}
	if !accountsReady {
		return &application.DashboardCashFlow{
			InflowsNext30Days:    application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
			OutflowsNext30Days:   application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
			InflowsNext60Days:    application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
			OutflowsNext60Days:   application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
			MinimumBalance30Days: unavailableMinimumBalance(),
			MinimumBalance60Days: unavailableMinimumBalance(),
		}, nil
	}
	// Query one day before the visible window to include forecast rows that can
	// appear on D after the report's effective float is applied. Keep the query
	// through the actual end date because recorded movements are not shifted.
	sourceStart := start.AddDate(0, 0, -blingCashFlowEffectiveFloatDays)
	sourceEnd := end
	// Cash movements in the visible window reproduce the report's inflow/outflow
	// lines. Reconciled movements after the approved baseline but before this
	// window advance the opening balance; the API's `situacao=R` alone only means
	// registered, so bridge movements come from a separate reconciliation-filtered snapshot.
	movementStart := start
	rows, err := tx.Query(ctx, cashFlowCandidatesSQL, correlationID, sourceStart.Format(time.DateOnly), sourceEnd.Format(time.DateOnly), start.Format(time.DateOnly), movementStart.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var daily [61]cashFlowDayAggregate
	partial := true // Float equivalence still needs a live comparison with the selected report.
	for rows.Next() {
		candidate, err := scanCashFlowCandidate(rows)
		if err != nil {
			return nil, err
		}
		if candidate.batchPartial {
			partial = true
		}
		accountSelected := false
		for _, id := range accountIDs {
			if candidate.accountID == id {
				accountSelected = true
				break
			}
		}
		if !accountSelected {
			if candidate.accountID == "" {
				partial = true
			}
			continue
		}
		amount, err := domain.ParseMoney(candidate.amount)
		if err != nil {
			return nil, fmt.Errorf("cash-flow candidate contains invalid amount")
		}
		flowDate := candidate.flowDate
		if candidate.kind == "MOVEMENT" {
			if candidate.sourceStatus != "R" && candidate.sourceStatus != "" {
				partial = true
				continue
			}
			if candidate.sourceStatus == "" {
				partial = true
			}
		} else if candidate.kind == "RECEIVABLE" {
			if (candidate.financialStatus != "OPEN" && candidate.financialStatus != "PARTIALLY_PAID") || !candidate.destinationKnown || candidate.destination.Int16 != int16(financial.BlingDestinationReceivablePayable) || !candidate.feeRateKnown || !candidate.fixedFeeKnown || !candidate.settlementKnown || !candidate.businessDaysKnown || !candidate.useBusinessDays.Valid || candidate.useBusinessDays.Bool {
				partial = true
				continue
			}
			terms := financial.BlingPaymentMethodTerms{
				Destination: financial.BlingPaymentDestination(candidate.destination.Int16), DestinationKnown: true,
				FeeRateKnown: candidate.feeRateKnown, FixedFeeKnown: candidate.fixedFeeKnown,
				SettlementKnown: candidate.settlementKnown, BusinessDaysKnown: candidate.businessDaysKnown,
			}
			terms.FeeRate, err = parseCashFlowMoney(candidate.feeRate)
			if err != nil {
				partial = true
				continue
			}
			terms.FixedFee, err = parseCashFlowMoney(candidate.fixedFee)
			if err != nil {
				partial = true
				continue
			}
			terms.SettlementDays = int(candidate.settlementDays.Int32)
			amount, err = financial.BlingNetReceivableValue(amount, terms)
			if err != nil {
				partial = true
				continue
			}
			flowDate, err = financial.BlingExpectedReceiptDate(flowDate, terms, nil)
			if err != nil {
				partial = true
				continue
			}
		} else if candidate.kind == "PAYABLE" {
			if !candidate.destinationKnown || candidate.destination.Int16 != int16(financial.BlingDestinationReceivablePayable) {
				partial = true
				continue
			}
			amount, err = amount.AbsExact()
			if err != nil {
				return nil, err
			}
		} else {
			partial = true
			continue
		}
		// The cash-movement API already records the realized date. The observed
		// bank float applies only to forecasted receipts/payables, after receipt
		// settlement terms; never move a registered movement into another day.
		flowDate = blingCashFlowReportDate(candidate.kind, flowDate)
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
			partial = true
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
			Inflows:  dashboardFlowMetric(daily[index].inflows.String(), partial),
			Outflows: dashboardFlowMetric(outflow.String(), partial),
		})
	}
	flow.InflowsNext30Days = dashboardFlowMetric(inflows30.String(), partial)
	flow.OutflowsNext30Days = dashboardFlowMetric(outflows30.String(), partial)
	flow.InflowsNext60Days = dashboardFlowMetric(inflows60.String(), partial)
	flow.OutflowsNext60Days = dashboardFlowMetric(outflows60.String(), partial)
	flow.Periods, err = groupCashFlowReportWeeks(flow.Days, partial)
	if err != nil {
		return nil, err
	}
	projectedOpening := checkpoint.BaseMetric
	if projectedOpening.Value == "" || checkpoint.BaseDate.IsZero() {
		projectedOpening = application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}
	} else if start.After(checkpoint.BaseDate.AddDate(0, 0, 1)) {
		ready, delta, _, deltaErr := reconciledCashMovementDelta(ctx, tx, correlationID, accountIDs, checkpoint.BaseDate, start.AddDate(0, 0, -1))
		if deltaErr != nil {
			return nil, deltaErr
		}
		if !ready {
			projectedOpening = application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"}
		} else {
			base, parseErr := domain.ParseMoney(projectedOpening.Value)
			if parseErr != nil {
				return nil, fmt.Errorf("approved opening balance is invalid")
			}
			bridge, parseErr := domain.ParseMoney(delta)
			if parseErr != nil {
				return nil, fmt.Errorf("reconciled opening movement total is invalid")
			}
			base, err = base.Add(bridge)
			if err != nil {
				return nil, err
			}
			projectedOpening = metric(base.String(), 1, "PARTIAL", "BLING")
		}
	}
	if err := applyWeeklyCashBalances(flow.Periods, projectedOpening, partial); err != nil {
		return nil, err
	}
	flow.CriticalWeek, err = selectCriticalCashFlowWeek(flow.Periods)
	if err != nil {
		return nil, err
	}
	flow.MinimumBalance30Days, flow.MinimumBalance60Days, err = minimumCashFlowBalances(daily[:], start, projectedOpening, partial)
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

func applyBlingCashFlowFloat(sourceDate time.Time) time.Time {
	return sourceDate.AddDate(0, 0, blingCashFlowEffectiveFloatDays)
}

func blingCashFlowReportDate(kind string, sourceDate time.Time) time.Time {
	if kind == "MOVEMENT" {
		return sourceDate
	}
	return applyBlingCashFlowFloat(sourceDate)
}

func selectedCashFlowAccountIDs(ctx context.Context, tx pgx.Tx, correlationID string) ([]string, bool, error) {
	rows, err := tx.Query(ctx, selectedCashFlowAccountsSQL, correlationID)
	if err != nil {
		return nil, false, err
	}
	ids, _, err := selectedCashFlowAccountIDsFromRows(rows)
	if err != nil {
		return nil, false, err
	}
	if len(ids) == 1 {
		return ids, true, nil
	}

	// The catalog endpoint can return a partial identity set (for example, if
	// its read scope or normalized snapshot was incomplete). Do not let those
	// unrelated rows block recovery from the IDs and descriptions already
	// present in the official cash-movement snapshot.
	rows, err = tx.Query(ctx, selectedCashFlowMovementAccountsSQL, correlationID)
	if err != nil {
		return nil, false, err
	}
	ids, _, err = selectedCashFlowAccountIDsFromRows(rows)
	if err != nil {
		return nil, false, err
	}
	if len(ids) != 1 {
		return nil, false, nil
	}
	return ids, true, nil
}

func selectedCashFlowAccountIDsFromRows(rows pgx.Rows) ([]string, bool, error) {
	defer rows.Close()
	var itau []string
	anyRows := false
	for rows.Next() {
		anyRows = true
		var id, description string
		if err := rows.Scan(&id, &description); err != nil {
			return nil, anyRows, err
		}
		name := normalizeCashFlowAccountName(description)
		if strings.HasPrefix(name, "itau") {
			itau = append(itau, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, anyRows, err
	}
	if len(itau) != 1 {
		return nil, anyRows, nil
	}
	return []string{itau[0]}, anyRows, nil
}

func normalizeCashFlowAccountName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ü", "u", "ç", "c").Replace(value)
	return strings.Join(strings.Fields(value), " ")
}

// PostgreSQL returns the source's NUMERIC(20,8) percentage with all eight
// fractional places, including insignificant trailing zeros. The exact-money
// helper accepts up to four places, so strip only insignificant zeros before
// parsing; a meaningful fifth-through-eighth decimal remains unsupported and
// the affected forecast row is skipped as partial by the caller.
func parseCashFlowMoney(value string) (domain.Money, error) {
	value = strings.TrimSpace(value)
	if separator := strings.IndexByte(value, '.'); separator >= 0 {
		fraction := strings.TrimRight(value[separator+1:], "0")
		value = value[:separator]
		if fraction != "" {
			value += "." + fraction
		}
	}
	return domain.ParseMoney(value)
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
