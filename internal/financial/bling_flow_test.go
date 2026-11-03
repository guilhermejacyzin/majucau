package financial

import (
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

type weekdayCalendar struct{ holidays map[string]bool }

func (c weekdayCalendar) IsBusinessDay(date time.Time) bool {
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday || c.holidays[date.Format("2006-01-02")] {
		return false
	}
	return true
}

func mustBlingFlowMoney(t *testing.T, value string) domain.Money {
	t.Helper()
	parsed, err := domain.ParseMoney(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestBlingNetReceivableValueMatchesOfficialExample(t *testing.T) {
	got, err := BlingNetReceivableValue(mustBlingFlowMoney(t, "166.66"), BlingPaymentMethodTerms{
		FeeRate:       mustBlingFlowMoney(t, "0.25"),
		FixedFee:      mustBlingFlowMoney(t, "1.99"),
		FeeRateKnown:  true,
		FixedFeeKnown: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := mustBlingFlowMoney(t, "164.25"); got.Compare(want) != 0 {
		t.Fatalf("net receivable = %s, want %s", got, want)
	}
}

func TestBlingExpectedReceiptDateUsesCalendarDaysWhenConfigured(t *testing.T) {
	got, err := BlingExpectedReceiptDate(time.Date(2018, time.July, 23, 0, 0, 0, 0, time.UTC), BlingPaymentMethodTerms{SettlementDays: 30, UseBusinessDays: false, SettlementKnown: true, BusinessDaysKnown: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2018, time.August, 22, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("expected receipt date = %s, want %s", got, want)
	}
}

func TestBlingExpectedReceiptDateRequiresHolidayAwareBusinessCalendar(t *testing.T) {
	due := time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC)
	terms := BlingPaymentMethodTerms{SettlementDays: 1, UseBusinessDays: true, SettlementKnown: true, BusinessDaysKnown: true}
	if _, err := BlingExpectedReceiptDate(due, terms, nil); err != ErrInvalidBlingFlowRule {
		t.Fatalf("missing calendar error = %v, want ErrInvalidBlingFlowRule", err)
	}

	calendar := weekdayCalendar{holidays: map[string]bool{"2026-11-03": true}}
	got, err := BlingExpectedReceiptDate(due, terms, calendar)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, time.November, 4, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("expected receipt date = %s, want %s", got, want)
	}
}

func TestBlingNetReceivableValueRejectsNegativeNetAmount(t *testing.T) {
	_, err := BlingNetReceivableValue(mustBlingFlowMoney(t, "1.00"), BlingPaymentMethodTerms{FixedFee: mustBlingFlowMoney(t, "1.01"), FeeRateKnown: true, FixedFeeKnown: true})
	if err != ErrInvalidBlingFlowRule {
		t.Fatalf("negative-net error = %v, want ErrInvalidBlingFlowRule", err)
	}
}
