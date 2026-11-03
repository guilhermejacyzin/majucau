package dashboard

import (
	"testing"

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

func TestGroupCashFlowReportWeeksRejectsInvalidMoney(t *testing.T) {
	days := []application.DashboardCashFlowDay{{
		Date: "2026-11-03", Inflows: dashboardFlowMetric("not-money", true), Outflows: dashboardFlowMetric("0.0000", true),
	}}
	if _, err := groupCashFlowReportWeeks(days, true); err == nil {
		t.Fatal("invalid API amount must not be published in a report bucket")
	}
}
