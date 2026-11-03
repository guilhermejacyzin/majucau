package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidBlingCashFlowReport = errors.New("invalid Bling cash-flow report")
var ErrBlingCashFlowReportMismatch = errors.New("Bling cash-flow report does not match the API projection")

// BlingCashFlowReportPeriod is one contiguous period as displayed by the
// Bling cash-flow report. Outflows are signed negative, matching the report.
type BlingCashFlowReportPeriod struct {
	From           time.Time
	Through        time.Time
	OpeningBalance domain.Money
	Inflows        domain.Money
	Outflows       domain.Money
	Adjustments    domain.Money
	ClosingBalance domain.Money
}

// BlingCashFlowReportSummary contains the total row displayed by Bling. The
// opening balance is the first period's opening balance; the closing balance
// is the last period's closing balance.
type BlingCashFlowReportSummary struct {
	Inflows        domain.Money
	Outflows       domain.Money
	Adjustments    domain.Money
	ClosingBalance domain.Money
}

// ValidateBlingCashFlowReport checks the accounting arithmetic and continuity
// of a report used as the oracle for the API-built T02 projection. It does not
// store report values or make them production data.
func ValidateBlingCashFlowReport(periods []BlingCashFlowReportPeriod, summary BlingCashFlowReportSummary) error {
	if len(periods) == 0 {
		return ErrInvalidBlingCashFlowReport
	}
	var inflows, outflows, adjustments domain.Money
	for i, period := range periods {
		from := reportDateOnly(period.From)
		through := reportDateOnly(period.Through)
		if period.From.IsZero() || period.Through.IsZero() || through.Before(from) || period.Inflows.IsNegative() || period.Outflows.IsPositive() {
			return ErrInvalidBlingCashFlowReport
		}
		if i > 0 {
			previous := periods[i-1]
			if !from.Equal(reportDateOnly(previous.Through).AddDate(0, 0, 1)) || period.OpeningBalance.Compare(previous.ClosingBalance) != 0 {
				return ErrInvalidBlingCashFlowReport
			}
		}
		closing, err := period.OpeningBalance.Add(period.Inflows)
		if err != nil {
			return err
		}
		closing, err = closing.Add(period.Outflows)
		if err != nil {
			return err
		}
		closing, err = closing.Add(period.Adjustments)
		if err != nil || closing.Compare(period.ClosingBalance) != 0 {
			if err != nil {
				return err
			}
			return ErrInvalidBlingCashFlowReport
		}
		inflows, err = inflows.Add(period.Inflows)
		if err != nil {
			return err
		}
		outflows, err = outflows.Add(period.Outflows)
		if err != nil {
			return err
		}
		adjustments, err = adjustments.Add(period.Adjustments)
		if err != nil {
			return err
		}
	}
	last := periods[len(periods)-1]
	if inflows.Compare(summary.Inflows) != 0 || outflows.Compare(summary.Outflows) != 0 || adjustments.Compare(summary.Adjustments) != 0 || last.ClosingBalance.Compare(summary.ClosingBalance) != 0 {
		return ErrInvalidBlingCashFlowReport
	}
	return nil
}

// CompareBlingCashFlowReport compares an API-composed series with the Bling
// report periods before those values can be marked confirmed in T02.
func CompareBlingCashFlowReport(reference, projection []BlingCashFlowReportPeriod) error {
	if err := validateReportPeriods(reference); err != nil {
		return err
	}
	if err := validateReportPeriods(projection); err != nil {
		return err
	}
	if len(reference) != len(projection) {
		return ErrBlingCashFlowReportMismatch
	}
	for i, expected := range reference {
		actual := projection[i]
		if !reportDateOnly(expected.From).Equal(reportDateOnly(actual.From)) ||
			!reportDateOnly(expected.Through).Equal(reportDateOnly(actual.Through)) ||
			expected.OpeningBalance.Compare(actual.OpeningBalance) != 0 ||
			expected.Inflows.Compare(actual.Inflows) != 0 ||
			expected.Outflows.Compare(actual.Outflows) != 0 ||
			expected.Adjustments.Compare(actual.Adjustments) != 0 ||
			expected.ClosingBalance.Compare(actual.ClosingBalance) != 0 {
			return ErrBlingCashFlowReportMismatch
		}
	}
	return nil
}

func validateReportPeriods(periods []BlingCashFlowReportPeriod) error {
	if len(periods) == 0 {
		return ErrInvalidBlingCashFlowReport
	}
	var summary BlingCashFlowReportSummary
	for _, period := range periods {
		var err error
		summary.Inflows, err = summary.Inflows.Add(period.Inflows)
		if err != nil {
			return err
		}
		summary.Outflows, err = summary.Outflows.Add(period.Outflows)
		if err != nil {
			return err
		}
		summary.Adjustments, err = summary.Adjustments.Add(period.Adjustments)
		if err != nil {
			return err
		}
	}
	summary.ClosingBalance = periods[len(periods)-1].ClosingBalance
	return ValidateBlingCashFlowReport(periods, summary)
}

func reportDateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
