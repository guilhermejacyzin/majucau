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
// movement. Use those official fields only when the optional account catalog
// could not be normalized (for example, when its read scope is unavailable).
// Distinct IDs and exact name prefixes keep the approved Itaú/Mercado Pago
// selection deterministic; ambiguous or missing matches remain unavailable.
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
     AND m.movement_date >= $2::date AND m.movement_date <= $3::date
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
     AND UPPER(p.status) NOT IN ('CANCELLED', 'CANCELED', 'REFUNDED', 'REVERSED', 'ESTORNADO')
     AND p.due_date >= $2::date AND p.due_date <= $3::date
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

func readBlingCashFlow(ctx context.Context, tx pgx.Tx, asOf time.Time, correlationID string) (*application.DashboardCashFlow, error) {
	start := domain.BusinessDate(asOf)
	end := start.AddDate(0, 0, 60)
	accountIDs, accountsReady, err := selectedCashFlowAccountIDs(ctx, tx, correlationID)
	if err != nil {
		return nil, err
	}
	if !accountsReady {
		return &application.DashboardCashFlow{
			InflowsNext30Days:  application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
			OutflowsNext30Days: application.DashboardMetric{State: "UNAVAILABLE", SourceSystem: "BLING"},
		}, nil
	}
	rows, err := tx.Query(ctx, cashFlowCandidatesSQL, correlationID, start.Format(time.DateOnly), end.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var daily [61]struct {
		inflows, outflows domain.Money
	}
	partial := true // Float and opening-balance equivalence are not in the documented API.
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
		day := dashboardDateOnly(flowDate)
		if day.Before(start) || day.After(end) {
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
	for index := range daily {
		date := start.AddDate(0, 0, index)
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
	return flow, nil
}

func dashboardDateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func selectedCashFlowAccountIDs(ctx context.Context, tx pgx.Tx, correlationID string) ([]string, bool, error) {
	rows, err := tx.Query(ctx, selectedCashFlowAccountsSQL, correlationID)
	if err != nil {
		return nil, false, err
	}
	ids, anyCatalogRows, err := selectedCashFlowAccountIDsFromRows(rows)
	if err != nil {
		return nil, false, err
	}
	if len(ids) == 2 {
		return ids, true, nil
	}
	if anyCatalogRows {
		return nil, false, nil
	}

	rows, err = tx.Query(ctx, selectedCashFlowMovementAccountsSQL, correlationID)
	if err != nil {
		return nil, false, err
	}
	ids, _, err = selectedCashFlowAccountIDsFromRows(rows)
	if err != nil {
		return nil, false, err
	}
	if len(ids) != 2 {
		return nil, false, nil
	}
	return ids, true, nil
}

func selectedCashFlowAccountIDsFromRows(rows pgx.Rows) ([]string, bool, error) {
	defer rows.Close()
	var itau, mercadoPago []string
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
		if strings.HasPrefix(name, "mercado pago") {
			mercadoPago = append(mercadoPago, id)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, anyRows, err
	}
	if len(itau) != 1 || len(mercadoPago) != 1 {
		return nil, anyRows, nil
	}
	return []string{itau[0], mercadoPago[0]}, anyRows, nil
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
