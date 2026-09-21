package financial

import (
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

func money(t *testing.T, s string) domain.Money {
	t.Helper()
	m, err := domain.ParseMoney(s)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestProjectionHasD0ThroughD60AndContinuity(t *testing.T) {
	start := time.Date(2026, 1, 31, 23, 59, 0, 0, time.FixedZone("BRT", -3*60*60))
	rows, err := DailyProjection(start, money(t, "100.00"), []Movement{{Date: start, Inflows: money(t, "10"), Outflows: money(t, "3")}, {Date: start.AddDate(0, 0, 1), Outflows: money(t, "2")}})
	if err != nil || len(rows) != 61 {
		t.Fatalf("projection: %d %v", len(rows), err)
	}
	if rows[0].ClosingBalance.String() != "107.0000" || rows[1].OpeningBalance.Compare(rows[0].ClosingBalance) != 0 {
		t.Fatalf("continuity broken: %#v", rows[:2])
	}
	if rows[0].Type != Provisional || rows[1].Type != Projected || rows[60].Date.Sub(rows[0].Date) != 60*24*time.Hour {
		t.Fatal("date/type contract broken")
	}
}
func TestAgingBoundariesAndExclusions(t *testing.T) {
	r := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		d    int
		want AgingBucket
	}{{-1, Overdue}, {0, Today}, {1, Days1To7}, {7, Days1To7}, {8, Days8To15}, {15, Days8To15}, {16, Days16To30}, {30, Days16To30}, {31, Days31To45}, {45, Days31To45}, {46, Days46To60}, {60, Days46To60}, {61, OutsideHorizon}} {
		if got := AgingBucketFor(r, r.AddDate(0, 0, tc.d)); got != tc.want {
			t.Errorf("d=%d got %s want %s", tc.d, got, tc.want)
		}
	}
	totals, err := CalculateAging(r, []OpenReceivable{{DueDate: r, OpenBalance: money(t, "10")}, {DueDate: r, OpenBalance: money(t, "5"), Status: "CANCELLED"}})
	if err != nil || totals[Today].Amount.String() != "10.0000" {
		t.Fatalf("aging exclusions: %#v %v", totals, err)
	}
	if _, err := CalculateAging(r, []OpenReceivable{{DueDate: r, OpenBalance: money(t, "-1")}}); err == nil {
		t.Fatal("negative open balance must fail")
	}
}
func TestApplicationConservativeLimits(t *testing.T) {
	rows, _ := DailyProjection(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), money(t, "100"), []Movement{{Date: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Outflows: money(t, "80")}})
	got, err := MaximumApplication(ApplicationInput{EligibleAccumulatedProfit: money(t, "100"), AlreadyInvested: money(t, "20"), MinimumReserve: money(t, "30"), Balances: rows, RuleVersion: "D-003-A"})
	if err != nil || got.MaximumInvestment.String() != "0.0000" || got.WorstWeekCash.String() != "20.0000" {
		t.Fatalf("application: %#v %v", got, err)
	}
}

func TestApplicationGoldenFormulaD003A(t *testing.T) {
	start := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	rows, err := DailyProjection(start, money(t, "250.00"), []Movement{{
		Date: start.AddDate(0, 0, 10), Outflows: money(t, "70.00"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := MaximumApplication(ApplicationInput{
		EligibleAccumulatedProfit: money(t, "200.00"),
		AlreadyInvested:           money(t, "40.00"),
		MinimumReserve:            money(t, "50.00"),
		Balances:                  rows,
		RuleVersion:               "D-003-A",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AvailableProfit.String() != "160.0000" || got.WorstWeekCash.String() != "180.0000" ||
		got.FinancialLimit.String() != "130.0000" || got.MaximumInvestment.String() != "130.0000" {
		t.Fatalf("golden D-003-A mismatch: %#v", got)
	}
}

func TestApplicationFloorsInsufficientProfitAndReserveBreach(t *testing.T) {
	start := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	rows, err := DailyProjection(start, money(t, "20.00"), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := MaximumApplication(ApplicationInput{
		EligibleAccumulatedProfit: money(t, "30.00"),
		AlreadyInvested:           money(t, "40.00"),
		MinimumReserve:            money(t, "50.00"),
		Balances:                  rows,
		RuleVersion:               "D-003-A",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.AvailableProfit.String() != "0.0000" || got.FinancialLimit.String() != "0.0000" || got.MaximumInvestment.String() != "0.0000" {
		t.Fatalf("negative application inputs must floor at zero: %#v", got)
	}
}

func TestApplicationRejectsInvalidHorizonAndNegativeParameters(t *testing.T) {
	rows, err := DailyProjection(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), money(t, "100"), nil)
	if err != nil {
		t.Fatal(err)
	}
	base := ApplicationInput{
		EligibleAccumulatedProfit: money(t, "100"),
		AlreadyInvested:           money(t, "20"),
		MinimumReserve:            money(t, "30"),
		Balances:                  rows,
		RuleVersion:               "D-003-A",
	}
	invalidHorizon := base
	invalidHorizon.Balances = rows[:60]
	if _, err := MaximumApplication(invalidHorizon); err != ErrInvalidApplicationInput {
		t.Fatalf("expected invalid horizon, got %v", err)
	}
	negativeInvested := base
	negativeInvested.AlreadyInvested = money(t, "-0.0001")
	if _, err := MaximumApplication(negativeInvested); err != ErrInvalidApplicationInput {
		t.Fatalf("expected negative invested rejection, got %v", err)
	}
	negativeReserve := base
	negativeReserve.MinimumReserve = money(t, "-0.0001")
	if _, err := MaximumApplication(negativeReserve); err != ErrInvalidApplicationInput {
		t.Fatalf("expected negative reserve rejection, got %v", err)
	}
}

func TestAgingIgnoresClosedBalanceAndRejectsMissingDate(t *testing.T) {
	reference := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	got, err := CalculateAging(reference, []OpenReceivable{{
		DueDate: reference, OpenBalance: domain.ZeroMoney(), Status: "OPEN",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("zero open balance must not be counted: %#v", got)
	}
	if _, err := CalculateAging(reference, []OpenReceivable{{
		OpenBalance: money(t, "1"), Status: "OPEN",
	}}); err == nil {
		t.Fatal("expected missing due date rejection")
	}
}

func TestB2BRejectsUnknownManualOverride(t *testing.T) {
	got := ClassifyB2B(ClassificationInput{
		ManualOverride: BusinessType("FORGED"), OverrideReason: "not valid",
	}, "D-001-A")
	if got.BusinessType != Unclassified {
		t.Fatalf("unknown override must fail closed: %#v", got)
	}
}
func TestB2BRuleAndCNPJ(t *testing.T) {
	if !ValidCNPJ("11.222.333/0001-81") {
		t.Fatal("known valid CNPJ rejected")
	}
	if ValidCNPJ("11.222.333/0001-80") || ValidCNPJ("11.111.111/1111-11") {
		t.Fatal("invalid CNPJ accepted")
	}
	if got := ClassifyB2B(ClassificationInput{ReconciledToNuvemB2C: true, Document: "11.222.333/0001-81"}, "v1"); got.BusinessType != B2CExcludedFromBlingFuture {
		t.Fatal("reconciliation must take priority")
	}
	if got := ClassifyB2B(ClassificationInput{Document: "11.222.333/0001-81"}, "v1"); got.BusinessType != B2B {
		t.Fatal("CNPJ must classify B2B")
	}
	if got := ClassifyB2B(ClassificationInput{ManualOverride: B2B}, "v1"); got.BusinessType != Unclassified {
		t.Fatal("override without reason must remain conservative")
	}
}
