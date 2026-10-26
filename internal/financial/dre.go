package financial

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidDREInput = errors.New("invalid DRE input")

// DREAmount distinguishes a verified zero from an amount that is not available.
// Amounts are signed effects on the result: revenue is positive, while deductions
// and expenses are negative. Inputs must already be classified by an approved
// upstream rule; this calculator does not classify source records.
type DREAmount struct {
	Amount    domain.Money
	Available bool
}

type DRELineCode string

const (
	DREGrossRevenueCode       DRELineCode = "GROSS_REVENUE"
	DRERevenueDeductionsCode  DRELineCode = "REVENUE_DEDUCTIONS"
	DRENetRevenueCode         DRELineCode = "NET_REVENUE"
	DRECostOfGoodsSoldCode    DRELineCode = "COST_OF_GOODS_SOLD"
	DREGrossProfitCode        DRELineCode = "GROSS_PROFIT"
	DRESellingExpensesCode    DRELineCode = "SELLING_EXPENSES"
	DREAdministrativeCode     DRELineCode = "ADMINISTRATIVE_EXPENSES"
	DREOtherOperatingCode     DRELineCode = "OTHER_OPERATING_EXPENSES"
	DREOperatingResultCode    DRELineCode = "OPERATING_RESULT"
	DREFinancialResultCode    DRELineCode = "FINANCIAL_RESULT"
	DREOtherIncomeExpenseCode DRELineCode = "OTHER_INCOME_EXPENSES"
	DREBeforeTaxCode          DRELineCode = "RESULT_BEFORE_INCOME_TAX"
	DREIncomeTaxesCode        DRELineCode = "INCOME_TAXES"
	DRENetIncomeCode          DRELineCode = "NET_INCOME"
)

type DREInput struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	RuleVersion string

	GrossRevenue                DREAmount
	RevenueDeductions           DREAmount
	CostOfGoodsSold             DREAmount
	SellingExpenses             DREAmount
	Administrative              DREAmount
	OtherOperating              DREAmount
	DepreciationAndAmortization DREAmount
	FinancialResult             DREAmount
	OtherIncomeExpenses         DREAmount
	IncomeTaxes                 DREAmount
}

type DRELine struct {
	Code      DRELineCode
	Amount    domain.Money
	Available bool
}

type DREResult struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	RuleVersion string
	Status      domain.DataStatus
	Lines       []DRELine
	Indicators  DREIndicators
}

// CalculateDRE applies only the arithmetic of the approved DRE structure to
// amounts that have already been classified. Missing inputs make dependent
// totals unavailable; they are never silently replaced by zero. Even complete
// arithmetic remains provisional until sources and classifications are
// reconciled by the application workflow.
func CalculateDRE(in DREInput) (DREResult, error) {
	if strings.TrimSpace(in.RuleVersion) == "" || in.PeriodStart.IsZero() || in.PeriodEnd.IsZero() ||
		dateOnly(in.PeriodStart).After(dateOnly(in.PeriodEnd)) {
		return DREResult{}, ErrInvalidDREInput
	}

	netRevenue, err := sumDREAmounts(in.GrossRevenue, in.RevenueDeductions)
	if err != nil {
		return DREResult{}, fmt.Errorf("calculate net revenue: %w", err)
	}
	grossProfit, err := sumDREAmounts(netRevenue, in.CostOfGoodsSold)
	if err != nil {
		return DREResult{}, fmt.Errorf("calculate gross profit: %w", err)
	}
	operatingResult, err := sumDREAmounts(grossProfit, in.SellingExpenses, in.Administrative, in.OtherOperating)
	if err != nil {
		return DREResult{}, fmt.Errorf("calculate operating result: %w", err)
	}
	beforeTax, err := sumDREAmounts(operatingResult, in.FinancialResult, in.OtherIncomeExpenses)
	if err != nil {
		return DREResult{}, fmt.Errorf("calculate result before income tax: %w", err)
	}
	netIncome, err := sumDREAmounts(beforeTax, in.IncomeTaxes)
	if err != nil {
		return DREResult{}, fmt.Errorf("calculate net income: %w", err)
	}
	indicators, err := CalculateDREIndicators(operatingResult, in.DepreciationAndAmortization)
	if err != nil {
		return DREResult{}, err
	}

	inputs := []DREAmount{
		in.GrossRevenue, in.RevenueDeductions, in.CostOfGoodsSold,
		in.SellingExpenses, in.Administrative, in.OtherOperating,
		in.FinancialResult, in.OtherIncomeExpenses, in.IncomeTaxes,
	}
	status := domain.StatusProvisional
	for _, amount := range inputs {
		if !amount.Available {
			status = domain.StatusPartiallyConfirmed
			break
		}
	}

	return DREResult{
		PeriodStart: dateOnly(in.PeriodStart),
		PeriodEnd:   dateOnly(in.PeriodEnd),
		RuleVersion: strings.TrimSpace(in.RuleVersion),
		Status:      status,
		Lines: []DRELine{
			dreLine(DREGrossRevenueCode, in.GrossRevenue),
			dreLine(DRERevenueDeductionsCode, in.RevenueDeductions),
			dreLine(DRENetRevenueCode, netRevenue),
			dreLine(DRECostOfGoodsSoldCode, in.CostOfGoodsSold),
			dreLine(DREGrossProfitCode, grossProfit),
			dreLine(DRESellingExpensesCode, in.SellingExpenses),
			dreLine(DREAdministrativeCode, in.Administrative),
			dreLine(DREOtherOperatingCode, in.OtherOperating),
			dreLine(DREOperatingResultCode, operatingResult),
			dreLine(DREFinancialResultCode, in.FinancialResult),
			dreLine(DREOtherIncomeExpenseCode, in.OtherIncomeExpenses),
			dreLine(DREBeforeTaxCode, beforeTax),
			dreLine(DREIncomeTaxesCode, in.IncomeTaxes),
			dreLine(DRENetIncomeCode, netIncome),
		},
		Indicators: indicators,
	}, nil
}

// DREIndicators contains the two approved result cards. These are derived
// indicators, not additional DRE table lines.
type DREIndicators struct {
	EBIT   DREAmount
	EBITDA DREAmount
}

// CalculateDREIndicators treats operatingResult as EBIT after the period's
// recognized depreciation and amortization. DepreciationAndAmortization is a
// signed operating expense (zero or negative); EBITDA adds that expense back.
// Source classification is outside this function, so missing inputs remain
// unavailable rather than being treated as zero.
func CalculateDREIndicators(operatingResult, depreciationAndAmortization DREAmount) (DREIndicators, error) {
	if depreciationAndAmortization.Available && depreciationAndAmortization.Amount.IsPositive() {
		return DREIndicators{}, ErrInvalidDREInput
	}
	if !operatingResult.Available || !depreciationAndAmortization.Available {
		return DREIndicators{}, nil
	}
	ebitda, err := operatingResult.Amount.Sub(depreciationAndAmortization.Amount)
	if err != nil {
		return DREIndicators{}, fmt.Errorf("calculate EBITDA: %w", err)
	}
	return DREIndicators{
		EBIT:   DREAmount{Amount: operatingResult.Amount, Available: true},
		EBITDA: DREAmount{Amount: ebitda, Available: true},
	}, nil
}

func dreLine(code DRELineCode, value DREAmount) DRELine {
	return DRELine{Code: code, Amount: value.Amount, Available: value.Available}
}

func sumDREAmounts(values ...DREAmount) (DREAmount, error) {
	total := domain.ZeroMoney()
	for _, value := range values {
		if !value.Available {
			return DREAmount{}, nil
		}
		var err error
		total, err = total.Add(value.Amount)
		if err != nil {
			return DREAmount{}, err
		}
	}
	return DREAmount{Amount: total, Available: true}, nil
}
