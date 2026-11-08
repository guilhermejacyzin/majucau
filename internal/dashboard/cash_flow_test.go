package dashboard

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/application"
)

func TestGroupCashFlowReportWeeksUsesBlingWeekBoundariesAndExactValues(t *testing.T) {
	days := []application.DashboardCashFlowDay{
		{Date: "2026-11-03", Inflows: dashboardFlowMetric("10.1000", true), Outflows: dashboardFlowMetric("-2.0000", true)},
		{Date: "2026-11-04", Inflows: dashboardFlowMetric("0.0000", true), Outflows: dashboardFlowMetric("-3.2500", true)},
		{Date: "2026-11-05", Inflows: dashboardFlowMetric("0.0000", true), Outflows: dashboardFlowMetric("0.0000", true)},
		{Date: "2026-11-06", Inflows: dashboardFlowMetric("0.0000", true), Outflows: dashboardFlowMetric("0.0000", true)},
		{Date: "2026-11-07", Inflows: dashboardFlowMetric("0.0000", true), Outflows: dashboardFlowMetric("0.0000", true)},
		{Date: "2026-11-08", Inflows: dashboardFlowMetric("0.2000", true), Outflows: dashboardFlowMetric("-1.0000", true)},
		{Date: "2026-11-09", Inflows: dashboardFlowMetric("0.0000", true), Outflows: dashboardFlowMetric("0.0000", true)},
	}
	periods, err := groupCashFlowReportWeeks(days, true)
	if err != nil {
		t.Fatalf("groupCashFlowReportWeeks() error = %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("got %d periods, want 2: %+v", len(periods), periods)
	}
	if periods[0].From != "2026-11-03" || periods[0].Through != "2026-11-07" || periods[0].Inflows.Value != "10.1000" || periods[0].Outflows.Value != "-5.2500" {
		t.Fatalf("first report bucket mismatch: %+v", periods[0])
	}
	if periods[1].From != "2026-11-08" || periods[1].Through != "2026-11-09" || periods[1].Inflows.Value != "0.2000" || periods[1].Outflows.Value != "-1.0000" {
		t.Fatalf("second report bucket mismatch: %+v", periods[1])
	}
	for _, period := range periods {
		if period.Inflows.State != "PARTIAL" || period.Outflows.State != "PARTIAL" {
			t.Fatalf("partial source state was lost: %+v", period)
		}
	}
}

func TestApplyBlingCashFlowFloatMovesSourceToNextCalendarDay(t *testing.T) {
	source := time.Date(2026, time.November, 21, 0, 0, 0, 0, time.UTC)
	got := applyBlingCashFlowFloat(source)
	want := time.Date(2026, time.November, 22, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("applyBlingCashFlowFloat(%s) = %s, want %s", source.Format(time.DateOnly), got.Format(time.DateOnly), want.Format(time.DateOnly))
	}
}

func TestRecordedCashMovementKeepsItsActualDateWhileForecastUsesFloat(t *testing.T) {
	source := time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)
	recorded := blingCashFlowReportDate("MOVEMENT", source)
	if !recorded.Equal(source) {
		t.Fatalf("recorded movement date = %s, want actual date %s", recorded.Format(time.DateOnly), source.Format(time.DateOnly))
	}
	wantForecast := time.Date(2026, time.November, 7, 0, 0, 0, 0, time.UTC)
	forecast := blingCashFlowReportDate("PAYABLE", source)
	if !forecast.Equal(wantForecast) {
		t.Fatalf("forecast date = %s, want effective float date %s", forecast.Format(time.DateOnly), wantForecast.Format(time.DateOnly))
	}
}

func TestCashFlowPeriodOmitsUnconfirmedClosingBalance(t *testing.T) {
	period := application.DashboardCashFlowPeriod{
		From:     "2026-11-01",
		Through:  "2026-11-07",
		Inflows:  dashboardFlowMetric("1250.0000", true),
		Outflows: dashboardFlowMetric("-800.0000", true),
	}
	encoded, err := json.Marshal(period)
	if err != nil {
		t.Fatalf("marshal cash-flow period: %v", err)
	}
	if strings.Contains(string(encoded), `"closing_balance"`) {
		t.Fatalf("unconfirmed closing balance must be omitted, got %s", encoded)
	}
}

func TestGroupCashFlowReportWeeksRejectsInvalidMoney(t *testing.T) {
	days := []application.DashboardCashFlowDay{{
		Date: "2026-11-03", Inflows: dashboardFlowMetric("not-money", true), Outflows: dashboardFlowMetric("0.0000", true),
	}}
	if _, err := groupCashFlowReportWeeks(days, true); err == nil {
		t.Fatal("invalid API amount must not be published in a report bucket")
	}
}

func TestSelectCriticalCashFlowWeekUsesLargestOutflowAndPreservesPartialState(t *testing.T) {
	periods := []application.DashboardCashFlowPeriod{
		{From: "2026-11-03", Through: "2026-11-07", Inflows: dashboardFlowMetric("1250.00", true), Outflows: dashboardFlowMetric("-800.00", true)},
		{From: "2026-11-08", Through: "2026-11-14", Inflows: dashboardFlowMetric("900.00", true), Outflows: dashboardFlowMetric("-152735.18", true)},
		{From: "2026-11-15", Through: "2026-11-21", Inflows: application.DashboardMetric{State: "UNAVAILABLE"}, Outflows: application.DashboardMetric{State: "UNAVAILABLE"}},
	}
	got, err := selectCriticalCashFlowWeek(periods)
	if err != nil {
		t.Fatalf("selectCriticalCashFlowWeek() error = %v", err)
	}
	if got == nil || got.From != "2026-11-08" || got.Through != "2026-11-14" || got.Outflows.Value != "-152735.18" || got.Outflows.State != "PARTIAL" || got.Inflows.Value != "900.00" {
		t.Fatalf("critical week = %+v, want largest partial outflow week", got)
	}
}

func TestSelectCriticalCashFlowWeekKeepsFirstWeekOnTieAndRejectsInvalidValue(t *testing.T) {
	periods := []application.DashboardCashFlowPeriod{
		{From: "2026-11-03", Through: "2026-11-07", Outflows: dashboardFlowMetric("-20.00", true)},
		{From: "2026-11-08", Through: "2026-11-14", Outflows: dashboardFlowMetric("20.00", true)},
	}
	got, err := selectCriticalCashFlowWeek(periods)
	if err != nil || got == nil || got.From != "2026-11-03" {
		t.Fatalf("critical week tie = %+v, %v; want first period", got, err)
	}
	if _, err := selectCriticalCashFlowWeek([]application.DashboardCashFlowPeriod{{Outflows: dashboardFlowMetric("bad", true)}}); err == nil {
		t.Fatal("invalid weekly outflow must not be published as the critical week")
	}
}
