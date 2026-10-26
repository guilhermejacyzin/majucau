package financial

import (
	"errors"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

func availableDREAmount(t *testing.T, amount string) DREAmount {
	t.Helper()
	return DREAmount{Amount: money(t, amount), Available: true}
}

func findDRELine(t *testing.T, result DREResult, code DRELineCode) DRELine {
	t.Helper()
	for _, line := range result.Lines {
		if line.Code == code {
			return line
		}
	}
	t.Fatalf("DRE line %q was not returned", code)
	return DRELine{}
}

func TestCalculateDREUsesExactSignedAmountsForApprovedLines(t *testing.T) {
	start := time.Date(2026, 10, 1, 12, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	end := time.Date(2026, 10, 31, 22, 0, 0, 0, start.Location())
	result, err := CalculateDRE(DREInput{
		PeriodStart: start, PeriodEnd: end, RuleVersion: "dre-v1",
		GrossRevenue:                availableDREAmount(t, "1000000"),
		RevenueDeductions:           availableDREAmount(t, "0"),
		CostOfGoodsSold:             availableDREAmount(t, "-600000"),
		SellingExpenses:             availableDREAmount(t, "-100000"),
		Administrative:              availableDREAmount(t, "-100000"),
		OtherOperating:              availableDREAmount(t, "-80000"),
		DepreciationAndAmortization: availableDREAmount(t, "-30000"),
		FinancialResult:             availableDREAmount(t, "-20000"),
		OtherIncomeExpenses:         availableDREAmount(t, "0"),
		IncomeTaxes:                 availableDREAmount(t, "-25000"),
	})
	if err != nil {
		t.Fatalf("CalculateDRE() error = %v", err)
	}
	if result.Status != domain.StatusProvisional {
		t.Fatalf("status = %q, want provisional", result.Status)
	}
	if !result.PeriodStart.Equal(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)) ||
		!result.PeriodEnd.Equal(time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("period dates were not normalized: %s through %s", result.PeriodStart, result.PeriodEnd)
	}
	for code, want := range map[DRELineCode]string{
		DRENetRevenueCode:      "1000000.0000",
		DREGrossProfitCode:     "400000.0000",
		DREOperatingResultCode: "120000.0000",
		DREBeforeTaxCode:       "100000.0000",
		DRENetIncomeCode:       "75000.0000",
	} {
		line := findDRELine(t, result, code)
		if !line.Available || line.Amount.String() != want {
			t.Errorf("line %s = (%s, available=%t), want (%s, available=true)", code, line.Amount, line.Available, want)
		}
	}
	if !result.Indicators.EBIT.Available || result.Indicators.EBIT.Amount.String() != "120000.0000" {
		t.Errorf("EBIT card = %#v, want available 120000.0000", result.Indicators.EBIT)
	}
	if !result.Indicators.EBITDA.Available || result.Indicators.EBITDA.Amount.String() != "150000.0000" {
		t.Errorf("EBITDA card = %#v, want available 150000.0000", result.Indicators.EBITDA)
	}
}

func TestCalculateDREKeepsDependentTotalsUnavailable(t *testing.T) {
	zero := availableDREAmount(t, "0")
	result, err := CalculateDRE(DREInput{
		PeriodStart:  time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:    time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
		RuleVersion:  "dre-v1",
		GrossRevenue: availableDREAmount(t, "100"), RevenueDeductions: zero,
		// Missing CMV must not be interpreted as a zero cost.
		SellingExpenses: zero, Administrative: zero, OtherOperating: zero,
		FinancialResult: zero, OtherIncomeExpenses: zero, IncomeTaxes: zero,
	})
	if err != nil {
		t.Fatalf("CalculateDRE() error = %v", err)
	}
	if result.Status != domain.StatusPartiallyConfirmed {
		t.Fatalf("status = %q, want partially confirmed", result.Status)
	}
	if line := findDRELine(t, result, DRECostOfGoodsSoldCode); line.Available || !line.Amount.IsZero() {
		t.Fatalf("missing CMV line must be unavailable, got %#v", line)
	}
	for _, code := range []DRELineCode{DREGrossProfitCode, DREOperatingResultCode, DREBeforeTaxCode, DRENetIncomeCode} {
		if line := findDRELine(t, result, code); line.Available {
			t.Errorf("line %s should be unavailable while CMV is missing", code)
		}
	}
	if line := findDRELine(t, result, DRENetRevenueCode); !line.Available || line.Amount.String() != "100.0000" {
		t.Fatalf("independent net revenue should remain available: %#v", line)
	}
}

func TestCalculateDRERejectsInvalidPeriodAndRuleVersion(t *testing.T) {
	periodStart := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if _, err := CalculateDRE(DREInput{PeriodStart: periodStart, PeriodEnd: periodEnd, RuleVersion: "dre-v1"}); !errors.Is(err, ErrInvalidDREInput) {
		t.Fatalf("reversed period error = %v, want ErrInvalidDREInput", err)
	}
	if _, err := CalculateDRE(DREInput{PeriodStart: periodStart, PeriodEnd: periodStart, RuleVersion: " "}); !errors.Is(err, ErrInvalidDREInput) {
		t.Fatalf("empty rule version error = %v, want ErrInvalidDREInput", err)
	}
}

func TestCalculateDREPropagatesOverflow(t *testing.T) {
	max, err := domain.ParseMoney("922337203685477.5807")
	if err != nil {
		t.Fatal(err)
	}
	result, err := CalculateDRE(DREInput{
		PeriodStart:       time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:         time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
		RuleVersion:       "dre-v1",
		GrossRevenue:      DREAmount{Amount: max, Available: true},
		RevenueDeductions: availableDREAmount(t, "0.0001"),
	})
	if !errors.Is(err, domain.ErrMoneyOverflow) || result.Lines != nil {
		t.Fatalf("overflow should fail without a partial result: result=%#v err=%v", result, err)
	}
}

func TestCalculateDREIndicatorsMatchApprovedExample(t *testing.T) {
	indicators, err := CalculateDREIndicators(
		availableDREAmount(t, "120000"),
		availableDREAmount(t, "-30000"),
	)
	if err != nil {
		t.Fatalf("CalculateDREIndicators() error = %v", err)
	}
	if !indicators.EBIT.Available || indicators.EBIT.Amount.String() != "120000.0000" {
		t.Fatalf("EBIT = %#v, want available 120000", indicators.EBIT)
	}
	if !indicators.EBITDA.Available || indicators.EBITDA.Amount.String() != "150000.0000" {
		t.Fatalf("EBITDA = %#v, want available 150000", indicators.EBITDA)
	}
}

func TestCalculateDREIndicatorsStayUnavailableWhenInputsAreMissing(t *testing.T) {
	zero := availableDREAmount(t, "0")
	for _, input := range [][2]DREAmount{
		{{}, zero},
		{zero, {}},
	} {
		indicators, err := CalculateDREIndicators(input[0], input[1])
		if err != nil {
			t.Fatalf("CalculateDREIndicators() error = %v", err)
		}
		if indicators.EBIT.Available || indicators.EBITDA.Available {
			t.Fatalf("missing required source must keep both cards unavailable: %#v", indicators)
		}
	}
}

func TestCalculateDREIndicatorsRejectPositiveDepreciationAndOverflow(t *testing.T) {
	if _, err := CalculateDREIndicators(availableDREAmount(t, "120"), availableDREAmount(t, "1")); !errors.Is(err, ErrInvalidDREInput) {
		t.Fatalf("positive depreciation error = %v, want ErrInvalidDREInput", err)
	}
	max, err := domain.ParseMoney("922337203685477.5807")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CalculateDREIndicators(
		DREAmount{Amount: max, Available: true},
		availableDREAmount(t, "-0.0001"),
	); !errors.Is(err, domain.ErrMoneyOverflow) {
		t.Fatalf("overflow error = %v, want ErrMoneyOverflow", err)
	}
}
