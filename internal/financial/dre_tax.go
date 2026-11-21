package financial

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidDRETaxInput = errors.New("invalid DRE tax input")

const (
	DRETaxValueUnavailable           DRETaxValueKind = "UNAVAILABLE"
	DRETaxValueEstimated             DRETaxValueKind = "ESTIMATED"
	DRETaxValueActual                DRETaxValueKind = "ACTUAL"
	DRETaxRateRuleVersion                            = "DRE_TAX_REFERENCE_V1"
	DRETaxRevenueActualNF                            = "ACTUAL_NF"
	DRETaxRevenueActualSalesFallback                 = "ACTUAL_SALES_FALLBACK"
	DRETaxRevenueProjectedNF                         = "PROJECTED_NF"
)

type DRETaxValueKind string

type DRETaxPayment struct {
	SourceRecordID string
	Competence     string
	PaymentDate    time.Time
	Amount         domain.Money // Positive source amount; the DRE effect is negative.
	Paid           bool
	Installment    bool
}

type DRETaxRateSample struct {
	Competence        string
	Payments          []DRETaxPayment
	NFRevenue         domain.Money
	NFSourceRecordIDs []string
}

type DRETaxWeeklyRevenue struct {
	StartDate       time.Time
	EndDate         time.Time
	Revenue         domain.Money
	SourceRecordIDs []string
	SourceKind      string // ACTUAL_NF, ACTUAL_SALES_FALLBACK, or PROJECTED_NF
	SourceComplete  bool
}

type DRETaxEstimateInput struct {
	Competence              string
	CurrentPayments         []DRETaxPayment
	CurrentPaymentsComplete bool
	HistoricalSamples       []DRETaxRateSample
	HistoryComplete         bool
	Weeks                   []DRETaxWeeklyRevenue
	WeeksComplete           bool
}

type DRETaxRate struct {
	Numerator          domain.Money `json:"numerator"`
	Denominator        domain.Money `json:"denominator"`
	SourceCompetence   string       `json:"source_competence,omitempty"`
	TaxSourceRecordIDs []string     `json:"tax_source_record_ids,omitempty"`
	NFSourceRecordIDs  []string     `json:"nf_source_record_ids,omitempty"`
	UsedFallback       bool         `json:"used_fallback"`
}

type DRETaxWeeklyContribution struct {
	StartDate        time.Time    `json:"start_date"`
	EndDate          time.Time    `json:"end_date"`
	Revenue          domain.Money `json:"revenue"`
	RevenueKind      string       `json:"revenue_kind"`
	RevenueSourceIDs []string     `json:"revenue_source_ids,omitempty"`
	TaxAmount        domain.Money `json:"tax_amount"` // Negative effect in the DRE.
}

type DRETaxResolution struct {
	Competence         string                     `json:"competence"`
	Kind               DRETaxValueKind            `json:"kind"`
	Amount             DREAmount                  `json:"amount"`
	Reason             string                     `json:"reason,omitempty"`
	RuleVersion        string                     `json:"rule_version"`
	TaxSourceRecordIDs []string                   `json:"tax_source_record_ids,omitempty"`
	Rate               *DRETaxRate                `json:"rate,omitempty"`
	Weekly             []DRETaxWeeklyContribution `json:"weekly,omitempty"`
	WeeklyReason       string                     `json:"weekly_reason,omitempty"`
}

// ResolveDRETax resolves one competence to a single current tax amount.
// A complete set of actual paid records always takes precedence over an
// estimate. Estimates use the latest complete paid-DAS/NF sample; the
// reference's 8.5% fallback is only available after history coverage is proven.
func ResolveDRETax(in DRETaxEstimateInput) (DRETaxResolution, error) {
	if _, err := parseDRECompetence(in.Competence); err != nil {
		return DRETaxResolution{}, ErrInvalidDRETaxInput
	}
	result := DRETaxResolution{
		Competence:  in.Competence,
		Kind:        DRETaxValueUnavailable,
		RuleVersion: DRETaxRateRuleVersion,
	}
	if !in.CurrentPaymentsComplete {
		result.Reason = "A cobertura dos pagamentos da competência não foi confirmada."
		return result, nil
	}
	for _, payment := range in.CurrentPayments {
		if payment.Paid && effectiveDRETaxPaymentCompetence(payment) == "" {
			result.Reason = "Há pagamento marcado como pago sem competência identificável; a estimativa foi bloqueada."
			return result, nil
		}
	}
	actualTotal, actualIDs, hasActual, err := sumPaidDRETax(in.Competence, in.CurrentPayments)
	if err != nil {
		return DRETaxResolution{}, err
	}
	if hasActual {
		signed, err := actualTotal.Negate()
		if err != nil {
			return DRETaxResolution{}, err
		}
		result.Kind = DRETaxValueActual
		result.Amount = DREAmount{Amount: signed, Available: true}
		result.TaxSourceRecordIDs = actualIDs
		result.Weekly, result.WeeklyReason = allocateActualDRETax(actualTotal, in.Weeks, in.WeeksComplete, in.Competence)
		return result, nil
	}
	if !in.HistoryComplete {
		result.Reason = "A cobertura histórica dos pagamentos e das notas fiscais não foi confirmada."
		return result, nil
	}
	if !in.WeeksComplete || !completeDRETaxWeeks(in.Weeks, in.Competence) {
		result.Reason = "A cobertura semanal das notas fiscais/projeções está incompleta."
		return result, nil
	}
	weeks := append([]DRETaxWeeklyRevenue(nil), in.Weeks...)
	sort.Slice(weeks, func(i, j int) bool { return dateOnly(weeks[i].StartDate).Before(dateOnly(weeks[j].StartDate)) })
	rate, err := latestDRETaxRate(in.HistoricalSamples, in.Competence)
	if err != nil {
		return DRETaxResolution{}, err
	}
	if rate == nil {
		numerator, _ := domain.ParseMoney("0.085")
		denominator, _ := domain.ParseMoney("1")
		rate = &DRETaxRate{Numerator: numerator, Denominator: denominator, UsedFallback: true}
	}
	weekly := make([]DRETaxWeeklyContribution, 0, len(weeks))
	total := domain.ZeroMoney()
	for _, week := range weeks {
		amount, err := week.Revenue.MulRatioRoundHalfUp(rate.Numerator.Units(), rate.Denominator.Units())
		if err != nil {
			return DRETaxResolution{}, fmt.Errorf("estimate weekly DRE tax: %w", err)
		}
		amount, err = amount.RoundHalfUp(2)
		if err != nil {
			return DRETaxResolution{}, fmt.Errorf("round weekly DRE tax: %w", err)
		}
		total, err = total.Add(amount)
		if err != nil {
			return DRETaxResolution{}, fmt.Errorf("sum weekly DRE tax: %w", err)
		}
		signed, err := amount.Negate()
		if err != nil {
			return DRETaxResolution{}, err
		}
		weekly = append(weekly, DRETaxWeeklyContribution{
			StartDate: dateOnly(week.StartDate), EndDate: dateOnly(week.EndDate),
			Revenue: week.Revenue, RevenueKind: week.SourceKind,
			RevenueSourceIDs: append([]string(nil), week.SourceRecordIDs...), TaxAmount: signed,
		})
	}
	signed, err := total.Negate()
	if err != nil {
		return DRETaxResolution{}, err
	}
	result.Kind = DRETaxValueEstimated
	result.Amount = DREAmount{Amount: signed, Available: true}
	result.Rate = rate
	result.Weekly = weekly
	return result, nil
}

func sumPaidDRETax(competence string, payments []DRETaxPayment) (domain.Money, []string, bool, error) {
	total := domain.ZeroMoney()
	ids := make([]string, 0)
	for _, payment := range payments {
		if !payment.Paid || effectiveDRETaxPaymentCompetence(payment) != competence {
			continue
		}
		sourceID := strings.TrimSpace(payment.SourceRecordID)
		if sourceID == "" || payment.Amount.IsNegative() {
			return domain.ZeroMoney(), nil, false, ErrInvalidDRETaxInput
		}
		var err error
		total, err = total.Add(payment.Amount)
		if err != nil {
			return domain.ZeroMoney(), nil, false, fmt.Errorf("sum paid DRE tax records: %w", err)
		}
		ids = append(ids, sourceID)
	}
	sort.Strings(ids)
	return total, ids, len(ids) > 0, nil
}

func latestDRETaxRate(samples []DRETaxRateSample, targetCompetence string) (*DRETaxRate, error) {
	targetDate, _ := parseDRECompetence(targetCompetence)
	valid := make([]DRETaxRateSample, 0, len(samples))
	for _, sample := range samples {
		sampleDate, err := parseDRECompetence(sample.Competence)
		if err != nil {
			return nil, ErrInvalidDRETaxInput
		}
		if !sampleDate.Before(targetDate) || sample.NFRevenue.IsNegative() || sample.NFRevenue.IsZero() || !hasDRETaxSourceIDs(sample.NFSourceRecordIDs) {
			continue
		}
		paid := domain.ZeroMoney()
		ids := make([]string, 0)
		eligiblePayments := make([]DRETaxPayment, 0)
		for _, payment := range sample.Payments {
			if !payment.Paid || payment.Installment || effectiveDRETaxPaymentCompetence(payment) != sample.Competence {
				continue
			}
			sourceID := strings.TrimSpace(payment.SourceRecordID)
			if sourceID == "" || payment.Amount.IsNegative() {
				return nil, ErrInvalidDRETaxInput
			}
			paid, err = paid.Add(payment.Amount)
			if err != nil {
				return nil, fmt.Errorf("sum historical DRE tax records: %w", err)
			}
			ids = append(ids, sourceID)
			payment.SourceRecordID = sourceID
			eligiblePayments = append(eligiblePayments, payment)
		}
		if paid.IsZero() || len(ids) == 0 {
			continue
		}
		sample.Payments = eligiblePayments
		valid = append(valid, sample)
	}
	if len(valid) == 0 {
		return nil, nil
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Competence > valid[j].Competence })
	latest := valid[0]
	tax := domain.ZeroMoney()
	ids := make([]string, 0)
	for _, payment := range latest.Payments {
		if payment.Paid && !payment.Installment && effectiveDRETaxPaymentCompetence(payment) == latest.Competence {
			var err error
			tax, err = tax.Add(payment.Amount)
			if err != nil {
				return nil, err
			}
			ids = append(ids, payment.SourceRecordID)
		}
	}
	return &DRETaxRate{
		Numerator: tax, Denominator: latest.NFRevenue, SourceCompetence: latest.Competence,
		TaxSourceRecordIDs: ids, NFSourceRecordIDs: append([]string(nil), latest.NFSourceRecordIDs...),
	}, nil
}

func allocateActualDRETax(total domain.Money, weeks []DRETaxWeeklyRevenue, complete bool, competence string) ([]DRETaxWeeklyContribution, string) {
	if !complete || !completeDRETaxWeeks(weeks, competence) {
		return nil, "O valor mensal pago está disponível; a distribuição semanal não foi fornecida por uma fonte completa."
	}
	weeks = append([]DRETaxWeeklyRevenue(nil), weeks...)
	sort.Slice(weeks, func(i, j int) bool { return dateOnly(weeks[i].StartDate).Before(dateOnly(weeks[j].StartDate)) })
	for _, week := range weeks {
		if !isDRETaxActualRevenueKind(week.SourceKind) {
			return nil, "O valor mensal pago está disponível; o rateio semanal exige notas fiscais reais, não projeções."
		}
	}
	weeklyRevenueTotal := domain.ZeroMoney()
	for _, week := range weeks {
		var err error
		weeklyRevenueTotal, err = weeklyRevenueTotal.Add(week.Revenue)
		if err != nil {
			return nil, "A distribuição semanal não pôde ser calculada por overflow."
		}
	}
	var allocations []domain.Money
	var err error
	if weeklyRevenueTotal.IsZero() {
		// The reference splits a paid monthly guide equally when there is no NF
		// revenue base for the month. Keep the exact monthly total in cents.
		allocations, err = allocateDRETaxUniformly(total, len(weeks))
	} else {
		allocations, err = allocateDRETaxProportionally(total, weeks, weeklyRevenueTotal)
	}
	if err != nil {
		return nil, "A distribuição semanal não pôde ser calculada por overflow."
	}
	result := make([]DRETaxWeeklyContribution, 0, len(weeks))
	for i, week := range weeks {
		amount := allocations[i]
		signed, err := amount.Negate()
		if err != nil {
			return nil, "A distribuição semanal não pôde ser calculada por overflow."
		}
		result = append(result, DRETaxWeeklyContribution{
			StartDate: dateOnly(week.StartDate), EndDate: dateOnly(week.EndDate),
			Revenue: week.Revenue, RevenueKind: week.SourceKind,
			RevenueSourceIDs: append([]string(nil), week.SourceRecordIDs...), TaxAmount: signed,
		})
	}
	return result, ""
}

func allocateDRETaxUniformly(total domain.Money, count int) ([]domain.Money, error) {
	if total.IsNegative() || count <= 0 {
		return nil, ErrInvalidDRETaxInput
	}
	base := total.Units() / int64(count)
	remainder := total.Units() % int64(count)
	result := make([]domain.Money, count)
	for i := range result {
		units := base
		if int64(i) < remainder {
			units++
		}
		amount, err := moneyFromUnits(units)
		if err != nil {
			return nil, err
		}
		result[i] = amount
	}
	return result, nil
}

func allocateDRETaxProportionally(total domain.Money, weeks []DRETaxWeeklyRevenue, revenueTotal domain.Money) ([]domain.Money, error) {
	if total.IsNegative() || revenueTotal.IsNegative() || revenueTotal.IsZero() || len(weeks) == 0 {
		return nil, ErrInvalidDRETaxInput
	}
	unitsTotal := big.NewInt(total.Units())
	denominator := big.NewInt(revenueTotal.Units())
	quotients := make([]*big.Int, len(weeks))
	remainders := make([]*big.Int, len(weeks))
	allocated := big.NewInt(0)
	for i, week := range weeks {
		if week.Revenue.IsNegative() {
			return nil, ErrInvalidDRETaxInput
		}
		product := new(big.Int).Mul(unitsTotal, big.NewInt(week.Revenue.Units()))
		quotients[i], remainders[i] = new(big.Int), new(big.Int)
		quotients[i].QuoRem(product, denominator, remainders[i])
		allocated.Add(allocated, quotients[i])
	}
	remaining := new(big.Int).Sub(unitsTotal, allocated)
	if !remaining.IsInt64() || remaining.Int64() < 0 || remaining.Int64() >= int64(len(weeks)) {
		return nil, ErrInvalidDRETaxInput
	}
	order := make([]int, len(weeks))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return remainders[order[i]].Cmp(remainders[order[j]]) > 0
	})
	for i := int64(0); i < remaining.Int64(); i++ {
		quotients[order[i]].Add(quotients[order[i]], big.NewInt(1))
	}
	result := make([]domain.Money, len(weeks))
	for i, quotient := range quotients {
		if !quotient.IsInt64() {
			return nil, domain.ErrMoneyOverflow
		}
		amount, err := moneyFromUnits(quotient.Int64())
		if err != nil {
			return nil, err
		}
		result[i] = amount
	}
	return result, nil
}

func moneyFromUnits(units int64) (domain.Money, error) {
	if units < 0 {
		return domain.Money{}, ErrInvalidDRETaxInput
	}
	return domain.ParseMoney(fmt.Sprintf("%d.%04d", units/10000, units%10000))
}

func hasDRETaxSourceIDs(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			return false
		}
	}
	return true
}

func completeDRETaxWeeks(weeks []DRETaxWeeklyRevenue, competence string) bool {
	if len(weeks) == 0 {
		return false
	}
	monthStart, monthEnd := time.Time{}, time.Time{}
	if competence != "" {
		var err error
		monthStart, err = parseDRECompetence(competence)
		if err != nil {
			return false
		}
		monthEnd = monthStart.AddDate(0, 1, -1)
	}
	ordered := append([]DRETaxWeeklyRevenue(nil), weeks...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].StartDate.Before(ordered[j].StartDate) })
	var previousEnd time.Time
	for _, week := range ordered {
		start, end := dateOnly(week.StartDate), dateOnly(week.EndDate)
		if start.IsZero() || end.IsZero() || start.After(end) || week.Revenue.IsNegative() ||
			(week.Revenue.IsPositive() && !hasDRETaxSourceIDs(week.SourceRecordIDs)) || !week.SourceComplete ||
			!isDRETaxRevenueKind(week.SourceKind) ||
			(!previousEnd.IsZero() && !start.After(previousEnd)) {
			return false
		}
		if competence != "" && (start.Before(monthStart) || end.After(monthEnd)) {
			return false
		}
		previousEnd = end
	}
	if competence != "" {
		if !ordered[0].StartDate.Equal(monthStart) && !dateOnly(ordered[0].StartDate).Equal(monthStart) {
			return false
		}
		if !dateOnly(ordered[len(ordered)-1].EndDate).Equal(monthEnd) {
			return false
		}
		for i := 1; i < len(ordered); i++ {
			if !dateOnly(ordered[i].StartDate).Equal(dateOnly(ordered[i-1].EndDate).AddDate(0, 0, 1)) {
				return false
			}
		}
	}
	return true
}

func isDRETaxActualRevenueKind(kind string) bool {
	return kind == DRETaxRevenueActualNF || kind == DRETaxRevenueActualSalesFallback
}

func isDRETaxRevenueKind(kind string) bool {
	return isDRETaxActualRevenueKind(kind) || kind == DRETaxRevenueProjectedNF
}

func effectiveDRETaxPaymentCompetence(payment DRETaxPayment) string {
	// The reference derives DAS competence from the actual caixa/banco posting
	// month (M+1 -> M). Its FCB competence field commonly represents the payment
	// month, so payment date takes precedence when this source is available.
	if !payment.PaymentDate.IsZero() {
		return dateOnly(payment.PaymentDate.AddDate(0, -1, 0)).Format("2006-01")
	}
	if _, err := parseDRECompetence(payment.Competence); err == nil {
		return payment.Competence
	}
	return ""
}

func parseDRECompetence(value string) (time.Time, error) {
	if len(value) != 7 || value[4] != '-' {
		return time.Time{}, ErrInvalidDRETaxInput
	}
	date, err := time.Parse("2006-01", value)
	if err != nil || date.Format("2006-01") != value {
		return time.Time{}, ErrInvalidDRETaxInput
	}
	return date, nil
}
