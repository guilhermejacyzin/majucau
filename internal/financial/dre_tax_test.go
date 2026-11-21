package financial

import (
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

func TestResolveDRETaxActualPaymentReplacesEstimateAndMapsPaymentMonth(t *testing.T) {
	weeks := dreTaxWeeks("2026-10", []string{"8000", "2000", "0", "0"}, "ACTUAL_NF")
	input := DRETaxEstimateInput{
		Competence: "2026-10", CurrentPaymentsComplete: true, HistoryComplete: true,
		CurrentPayments: []DRETaxPayment{{
			SourceRecordID: "das-paid-october", PaymentDate: time.Date(2026, time.November, 20, 0, 0, 0, 0, time.UTC),
			Amount: dreTaxMoney(t, "550"), Paid: true,
		}, {
			SourceRecordID: "das-installment-october", PaymentDate: time.Date(2026, time.November, 25, 0, 0, 0, 0, time.UTC),
			Amount: dreTaxMoney(t, "20"), Paid: true, Installment: true,
		}},
		HistoricalSamples: []DRETaxRateSample{{
			Competence: "2026-09", NFRevenue: dreTaxMoney(t, "10000"), NFSourceRecordIDs: []string{"nf-september"},
			Payments: []DRETaxPayment{{SourceRecordID: "das-september", Competence: "2026-09", Amount: dreTaxMoney(t, "1000"), Paid: true}},
		}},
		Weeks: weeks, WeeksComplete: true,
	}

	result, err := ResolveDRETax(input)
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueActual || !result.Amount.Available || result.Amount.Amount.String() != "-570.0000" {
		t.Fatalf("tax resolution = %+v, want paid actual -570.0000 including paid installment", result)
	}
	if len(result.TaxSourceRecordIDs) != 2 || result.TaxSourceRecordIDs[0] != "das-installment-october" || result.TaxSourceRecordIDs[1] != "das-paid-october" {
		t.Fatalf("actual tax sources = %v", result.TaxSourceRecordIDs)
	}
	if len(result.Weekly) != 4 || result.Weekly[0].TaxAmount.String() != "-456.0000" || result.Weekly[1].TaxAmount.String() != "-114.0000" || result.Weekly[2].TaxAmount.String() != "0.0000" || result.Weekly[3].TaxAmount.String() != "0.0000" {
		t.Fatalf("weekly actual allocation = %+v", result.Weekly)
	}
}

func TestResolveDRETaxUsesLatestEligibleRateAndActualOrProjectedWeeklyBase(t *testing.T) {
	input := DRETaxEstimateInput{
		Competence: "2026-12", CurrentPaymentsComplete: true, HistoryComplete: true,
		HistoricalSamples: []DRETaxRateSample{
			{Competence: "2026-08", NFRevenue: dreTaxMoney(t, "1000"), NFSourceRecordIDs: []string{"nf-august"}, Payments: []DRETaxPayment{{SourceRecordID: "das-august", Competence: "2026-08", Amount: dreTaxMoney(t, "200"), Paid: true}}},
			{Competence: "2026-10", NFRevenue: dreTaxMoney(t, "1000"), NFSourceRecordIDs: []string{"nf-october"}, Payments: []DRETaxPayment{
				{SourceRecordID: "das-october", Competence: "2026-10", Amount: dreTaxMoney(t, "100"), Paid: true},
				{SourceRecordID: "das-installment", Competence: "2026-10", Amount: dreTaxMoney(t, "900"), Paid: true, Installment: true},
			}},
		},
		Weeks: []DRETaxWeeklyRevenue{
			{StartDate: time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 7, 0, 0, 0, 0, time.UTC), Revenue: dreTaxMoney(t, "500"), SourceRecordIDs: []string{"nf-week-1"}, SourceKind: "ACTUAL_NF", SourceComplete: true},
			{StartDate: time.Date(2026, 12, 8, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 14, 0, 0, 0, 0, time.UTC), Revenue: dreTaxMoney(t, "500"), SourceRecordIDs: []string{"nf-week-2"}, SourceKind: "ACTUAL_NF", SourceComplete: true},
			{StartDate: time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 21, 0, 0, 0, 0, time.UTC), Revenue: dreTaxMoney(t, "1000"), SourceRecordIDs: []string{"projection-week-3"}, SourceKind: "PROJECTED_NF", SourceComplete: true},
			{StartDate: time.Date(2026, 12, 22, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), Revenue: dreTaxMoney(t, "1000"), SourceRecordIDs: []string{"projection-week-4"}, SourceKind: "PROJECTED_NF", SourceComplete: true},
		},
		WeeksComplete: true,
	}

	result, err := ResolveDRETax(input)
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueEstimated || !result.Amount.Available || result.Amount.Amount.String() != "-300.0000" {
		t.Fatalf("tax resolution = %+v, want estimate -300.0000", result)
	}
	if result.Rate == nil || result.Rate.UsedFallback || result.Rate.SourceCompetence != "2026-10" || result.Rate.Numerator.String() != "100.0000" || result.Rate.Denominator.String() != "1000.0000" {
		t.Fatalf("rate = %+v, want 10%% based on October paid tax excluding installment", result.Rate)
	}
	if len(result.Weekly) != 4 || result.Weekly[0].TaxAmount.String() != "-50.0000" || result.Weekly[2].TaxAmount.String() != "-100.0000" || result.Weekly[2].RevenueKind != "PROJECTED_NF" {
		t.Fatalf("weekly estimates = %+v", result.Weekly)
	}
}

func TestResolveDRETaxUsesVisibleFallbackOnlyWhenHistoryIsComplete(t *testing.T) {
	input := DRETaxEstimateInput{
		Competence: "2026-11", CurrentPaymentsComplete: true, HistoryComplete: true,
		Weeks: dreTaxWeeks("2026-11", []string{"1000", "0", "0", "0"}, "ACTUAL_NF"), WeeksComplete: true,
	}
	result, err := ResolveDRETax(input)
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueEstimated || result.Amount.Amount.String() != "-85.0000" || result.Rate == nil || !result.Rate.UsedFallback {
		t.Fatalf("fallback resolution = %+v, want visibly marked 8.5%% estimate", result)
	}
	input.HistoryComplete = false
	result, err = ResolveDRETax(input)
	if err != nil {
		t.Fatalf("ResolveDRETax() with missing history error = %v", err)
	}
	if result.Kind != DRETaxValueUnavailable || result.Amount.Available || result.Rate != nil {
		t.Fatalf("incomplete history must not trigger fallback: %+v", result)
	}
}

func TestResolveDRETaxDoesNotEstimateWithIncompleteWeeklySources(t *testing.T) {
	input := DRETaxEstimateInput{
		Competence: "2026-11", CurrentPaymentsComplete: true, HistoryComplete: true,
		Weeks: dreTaxWeeks("2026-11", []string{"1000", "0", "0", "0"}, "ACTUAL_NF"), WeeksComplete: false,
	}
	result, err := ResolveDRETax(input)
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueUnavailable || result.Amount.Available {
		t.Fatalf("incomplete weekly source must remain unavailable: %+v", result)
	}
}

func TestResolveDRETaxBlocksPaidRecordWithoutCompetence(t *testing.T) {
	result, err := ResolveDRETax(DRETaxEstimateInput{
		Competence: "2026-11", CurrentPaymentsComplete: true, HistoryComplete: true,
		CurrentPayments: []DRETaxPayment{{SourceRecordID: "paid-unmapped", Amount: dreTaxMoney(t, "100"), Paid: true}},
	})
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueUnavailable || result.Amount.Available || !strings.Contains(result.Reason, "sem competência") {
		t.Fatalf("unmapped paid tax must block the estimate: %+v", result)
	}
}

func TestResolveDRETaxUsesPostingMonthAndUniformAllocationWithoutNFBase(t *testing.T) {
	weeks := dreTaxWeeks("2026-10", []string{"0", "0", "0", "0"}, "ACTUAL_NF")
	result, err := ResolveDRETax(DRETaxEstimateInput{
		Competence: "2026-10", CurrentPaymentsComplete: true,
		// The source's FCB competence may reflect November (payment month); the
		// reference uses the posting date and maps November back to October.
		CurrentPayments: []DRETaxPayment{{
			SourceRecordID: "das-october-paid-in-november", Competence: "2026-11",
			PaymentDate: time.Date(2026, time.November, 20, 0, 0, 0, 0, time.UTC),
			Amount:      dreTaxMoney(t, "100"), Paid: true,
		}},
		Weeks: weeks, WeeksComplete: true,
	})
	if err != nil {
		t.Fatalf("ResolveDRETax() error = %v", err)
	}
	if result.Kind != DRETaxValueActual || result.Amount.Amount.String() != "-100.0000" || len(result.Weekly) != 4 {
		t.Fatalf("tax resolution = %+v, want actual October payment", result)
	}
	for i, week := range result.Weekly {
		if week.TaxAmount.String() != "-25.0000" {
			t.Fatalf("week %d allocation = %s, want -25.0000", i, week.TaxAmount)
		}
	}
}

func TestCalculateDREUsesExactlyOneResolvedTaxValueForSelectedCompetence(t *testing.T) {
	for _, tc := range []struct {
		name           string
		kind           DRETaxValueKind
		amount         string
		wantDeductions string
		wantNetIncome  string
	}{
		{name: "estimated until paid", kind: DRETaxValueEstimated, amount: "-85", wantDeductions: "-85.0000", wantNetIncome: "915.0000"},
		{name: "paid replaces estimate", kind: DRETaxValueActual, amount: "-120", wantDeductions: "-120.0000", wantNetIncome: "880.0000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
			resolution := &DRETaxResolution{
				Competence: "2026-11", Kind: tc.kind,
				Amount: DREAmount{Amount: dreTaxMoney(t, tc.amount), Available: true},
			}
			result, err := CalculateDRE(DREInput{
				PeriodStart: start, PeriodEnd: start.AddDate(0, 1, -1), RuleVersion: "dre-v1",
				GrossRevenue:      availableDREAmount(t, "1000"),
				RevenueDeductions: availableDREAmount(t, "0"), CostOfGoodsSold: availableDREAmount(t, "0"),
				SellingExpenses: availableDREAmount(t, "0"), Administrative: availableDREAmount(t, "0"),
				OtherOperating: availableDREAmount(t, "0"), FinancialResult: availableDREAmount(t, "0"),
				OtherIncomeExpenses: availableDREAmount(t, "0"), IncomeTaxes: availableDREAmount(t, "0"),
				TaxResolution: resolution,
			})
			if err != nil {
				t.Fatalf("CalculateDRE() error = %v", err)
			}
			if result.TaxValueKind != tc.kind || findDRELine(t, result, DRENetIncomeCode).Amount.String() != tc.wantNetIncome {
				t.Fatalf("tax kind/result = %q / %s, want %q / %s", result.TaxValueKind, findDRELine(t, result, DRENetIncomeCode).Amount, tc.kind, tc.wantNetIncome)
			}
			if got := findDRELine(t, result, DRERevenueDeductionsCode).Amount.String(); got != tc.wantDeductions {
				t.Errorf("revenue deductions = %s, want DAS included once at %s", got, tc.wantDeductions)
			}
			if got := findDRELine(t, result, DREIncomeTaxesCode).Amount.String(); got != "0.0000" {
				t.Errorf("tax on profit = %s, want 0.0000; DAS must not be placed on this line", got)
			}
		})
	}
}

func dreTaxMoney(t *testing.T, value string) domain.Money {
	t.Helper()
	amount, err := domain.ParseMoney(value)
	if err != nil {
		t.Fatalf("parse test money %q: %v", value, err)
	}
	return amount
}

func dreTaxWeeks(competence string, revenues []string, kind string) []DRETaxWeeklyRevenue {
	start, _ := parseDRECompetence(competence)
	result := make([]DRETaxWeeklyRevenue, 0, len(revenues))
	weekStart := start
	for i, value := range revenues {
		weekEnd := weekStart.AddDate(0, 0, 6)
		monthEnd := start.AddDate(0, 1, -1)
		if i == len(revenues)-1 || weekEnd.After(monthEnd) {
			weekEnd = monthEnd
		}
		amount, _ := domain.ParseMoney(value)
		ids := []string(nil)
		if amount.IsPositive() {
			ids = []string{"nf-week-" + string(rune('1'+i))}
		}
		result = append(result, DRETaxWeeklyRevenue{
			StartDate: weekStart, EndDate: weekEnd, Revenue: amount,
			SourceRecordIDs: ids, SourceKind: kind, SourceComplete: true,
		})
		weekStart = weekEnd.AddDate(0, 0, 1)
	}
	return result
}
