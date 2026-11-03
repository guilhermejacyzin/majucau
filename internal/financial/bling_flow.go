package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidBlingFlowRule = errors.New("invalid Bling cash-flow rule")

// BusinessCalendar is supplied by the caller because the Bling payment-method
// setting can count business days, including holidays. The calculation must
// not silently treat every weekday as a business day.
type BusinessCalendar interface {
	IsBusinessDay(time.Time) bool
}

// BlingPaymentMethodTerms contains the read-only settings that affect an open
// receivable in the Bling cash-flow report. FeeRate is expressed as percent
// (for example, 2.5 means 2.5%), and FixedFee is charged per receivable parcel.
type BlingPaymentMethodTerms struct {
	FeeRate           domain.Money
	FixedFee          domain.Money
	SettlementDays    int
	UseBusinessDays   bool
	FeeRateKnown      bool
	FixedFeeKnown     bool
	SettlementKnown   bool
	BusinessDaysKnown bool
}

// BlingNetReceivableValue applies the configured rate and fixed fee to one
// open receivable parcel, matching the Bling help example. The fee is rounded
// to cents before it is subtracted from the gross amount.
func BlingNetReceivableValue(gross domain.Money, terms BlingPaymentMethodTerms) (domain.Money, error) {
	if gross.IsNegative() || terms.FeeRate.IsNegative() || terms.FixedFee.IsNegative() || terms.SettlementDays < 0 || !terms.FeeRateKnown || !terms.FixedFeeKnown {
		return domain.Money{}, ErrInvalidBlingFlowRule
	}
	percentageFee, err := gross.MulRatioRoundHalfUp(terms.FeeRate.Units(), 100*10_000)
	if err != nil {
		return domain.Money{}, err
	}
	fee, err := percentageFee.Add(terms.FixedFee)
	if err != nil {
		return domain.Money{}, err
	}
	fee, err = fee.RoundHalfUp(2)
	if err != nil {
		return domain.Money{}, err
	}
	net, err := gross.Sub(fee)
	if err != nil {
		return domain.Money{}, err
	}
	if net.IsNegative() {
		return domain.Money{}, ErrInvalidBlingFlowRule
	}
	return net, nil
}

// BlingExpectedReceiptDate applies the payment-method settlement term to the
// receivable's due date. Business-day terms require an explicit calendar; after
// adding the term, a non-business result moves to the next business day.
func BlingExpectedReceiptDate(dueDate time.Time, terms BlingPaymentMethodTerms, calendar BusinessCalendar) (time.Time, error) {
	if dueDate.IsZero() || terms.SettlementDays < 0 || !terms.SettlementKnown || !terms.BusinessDaysKnown {
		return time.Time{}, ErrInvalidBlingFlowRule
	}
	result := dateOnly(dueDate)
	if !terms.UseBusinessDays {
		return result.AddDate(0, 0, terms.SettlementDays), nil
	}
	if calendar == nil {
		return time.Time{}, ErrInvalidBlingFlowRule
	}
	for days := 0; days < terms.SettlementDays; {
		result = result.AddDate(0, 0, 1)
		if calendar.IsBusinessDay(result) {
			days++
		}
	}
	for !calendar.IsBusinessDay(result) {
		result = result.AddDate(0, 0, 1)
	}
	return result, nil
}
