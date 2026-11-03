package financial

import (
	"errors"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

func TestValidateBlingCashFlowReportMatchesReferenceScreenshot(t *testing.T) {
	periods := referenceBlingCashFlowPeriods(t)
	summary := BlingCashFlowReportSummary{
		Inflows:        mustFlowMoney(t, "76600.30"),
		Outflows:       mustFlowMoney(t, "-281052.64"),
		Adjustments:    domain.ZeroMoney(),
		ClosingBalance: mustFlowMoney(t, "-129988.98"),
	}
	if err := ValidateBlingCashFlowReport(periods, summary); err != nil {
		t.Fatalf("valid report should pass arithmetic and summary checks: %v", err)
	}
}

func TestCompareBlingCashFlowReportRequiresExactPeriodValues(t *testing.T) {
	reference := referenceBlingCashFlowPeriods(t)
	projection := append([]BlingCashFlowReportPeriod(nil), reference...)
	if err := CompareBlingCashFlowReport(reference, projection); err != nil {
		t.Fatalf("matching API projection should pass: %v", err)
	}
	projection[2].Inflows = mustFlowMoney(t, "37383.23")
	projection[2].Outflows = mustFlowMoney(t, "-152735.19")
	if err := CompareBlingCashFlowReport(reference, projection); !errors.Is(err, ErrBlingCashFlowReportMismatch) {
		t.Fatalf("mismatched period must remain unconfirmed, got %v", err)
	}
}

func TestValidateBlingCashFlowReportRejectsBrokenContinuity(t *testing.T) {
	periods := referenceBlingCashFlowPeriods(t)
	periods[1].OpeningBalance = mustFlowMoney(t, "57381.55")
	err := ValidateBlingCashFlowReport(periods, BlingCashFlowReportSummary{
		Inflows:        mustFlowMoney(t, "76600.30"),
		Outflows:       mustFlowMoney(t, "-281052.64"),
		Adjustments:    domain.ZeroMoney(),
		ClosingBalance: mustFlowMoney(t, "-129988.98"),
	})
	if !errors.Is(err, ErrInvalidBlingCashFlowReport) {
		t.Fatalf("broken opening/closing continuity must be rejected, got %v", err)
	}
}

func TestBlingCashFlowReportFromDailyBalancesUsesBlingWeeklyPeriods(t *testing.T) {
	start := time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC)
	daily, err := DailyProjection(start, mustFlowMoney(t, "100.00"), []Movement{
		{Date: start, Inflows: mustFlowMoney(t, "10.00"), Outflows: mustFlowMoney(t, "2.00")},
		{Date: start.AddDate(0, 0, 7), Inflows: mustFlowMoney(t, "5.00"), Outflows: mustFlowMoney(t, "1.00")},
	})
	if err != nil {
		t.Fatalf("DailyProjection() error = %v", err)
	}
	periods, summary, err := BlingCashFlowReportFromDailyBalances(daily[:9])
	if err != nil {
		t.Fatalf("BlingCashFlowReportFromDailyBalances() error = %v", err)
	}
	if len(periods) != 2 || periods[0].From.Format("2006-01-02") != "2026-11-02" || periods[0].Through.Format("2006-01-02") != "2026-11-07" || periods[1].From.Format("2006-01-02") != "2026-11-08" || periods[1].Through.Format("2006-01-02") != "2026-11-10" {
		t.Fatalf("period boundaries do not match the Bling weekly report: %+v", periods)
	}
	if periods[0].Outflows.String() != "-2.0000" || periods[1].Outflows.String() != "-1.0000" || summary.Inflows.String() != "15.0000" || summary.Outflows.String() != "-3.0000" || summary.ClosingBalance.String() != "112.0000" {
		t.Fatalf("signed values or summary are incorrect: periods=%+v summary=%+v", periods, summary)
	}
}

func TestBlingCashFlowReportFromDailyBalancesRejectsMissingOpeningBalance(t *testing.T) {
	start := time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC)
	daily, err := DailyProjection(start, mustFlowMoney(t, "100.00"), []Movement{
		{Date: start, Outflows: mustFlowMoney(t, "5.00")},
	})
	if err != nil {
		t.Fatalf("DailyProjection() error = %v", err)
	}
	daily[0].ClosingBalance = mustFlowMoney(t, "94.00")
	if _, _, err := BlingCashFlowReportFromDailyBalances(daily[:1]); !errors.Is(err, ErrInvalidBlingCashFlowReport) {
		t.Fatalf("invalid daily balance should fail closed, got %v", err)
	}
}

func referenceBlingCashFlowPeriods(t *testing.T) []BlingCashFlowReportPeriod {
	t.Helper()
	local, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	values := [][7]string{
		{"2026-11-02", "2026-11-07", "74463.36", "8401.31", "-25483.13", "0.00", "57381.54"},
		{"2026-11-08", "2026-11-14", "57381.54", "12362.61", "-18020.38", "0.00", "51723.77"},
		{"2026-11-15", "2026-11-21", "51723.77", "37383.22", "-152735.18", "0.00", "-63628.19"},
		{"2026-11-22", "2026-11-28", "-63628.19", "9722.56", "-22264.52", "0.00", "-76170.15"},
		{"2026-11-29", "2026-12-05", "-76170.15", "7124.43", "-57445.14", "0.00", "-126490.86"},
		{"2026-12-06", "2026-12-09", "-126490.86", "1606.17", "-5104.29", "0.00", "-129988.98"},
	}
	periods := make([]BlingCashFlowReportPeriod, 0, len(values))
	for _, row := range values {
		from, err := time.ParseInLocation("2006-01-02", row[0], local)
		if err != nil {
			t.Fatal(err)
		}
		through, err := time.ParseInLocation("2006-01-02", row[1], local)
		if err != nil {
			t.Fatal(err)
		}
		periods = append(periods, BlingCashFlowReportPeriod{
			From: from, Through: through,
			OpeningBalance: mustFlowMoney(t, row[2]),
			Inflows:        mustFlowMoney(t, row[3]),
			Outflows:       mustFlowMoney(t, row[4]),
			Adjustments:    mustFlowMoney(t, row[5]),
			ClosingBalance: mustFlowMoney(t, row[6]),
		})
	}
	return periods
}

func mustFlowMoney(t *testing.T, value string) domain.Money {
	t.Helper()
	amount, err := domain.ParseMoney(value)
	if err != nil {
		t.Fatal(err)
	}
	return amount
}
