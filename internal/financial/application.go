package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidApplicationInput = errors.New("invalid application input")

type ApplicationInput struct {
	EligibleAccumulatedProfit, AlreadyInvested, MinimumReserve domain.Money
	// AsOfDate is the business date represented by D0 in the 61-day balances.
	AsOfDate    time.Time
	Balances    []DailyBalance
	RuleVersion string
}
type ApplicationResult struct {
	EligibleAccumulatedProfit, AlreadyInvested, AvailableProfit, WorstWeekCash, MinimumReserve, FinancialLimit, MaximumInvestment domain.Money
	RuleVersion                                                                                                                   string
	Status                                                                                                                        domain.DataStatus
}

// MaximumApplication implements D-003-A. AsOfDate must match the provisional
// D0 balance, and all 61 daily balances must be complete and internally
// consistent. Only balances supplied by an approved source should be passed by
// the caller; this package does not infer or manufacture expected receipts.
func MaximumApplication(in ApplicationInput) (ApplicationResult, error) {
	if in.RuleVersion == "" || in.AsOfDate.IsZero() || len(in.Balances) != 61 ||
		in.AlreadyInvested.IsNegative() || in.MinimumReserve.IsNegative() {
		return ApplicationResult{}, ErrInvalidApplicationInput
	}
	minimum, err := FindMinimumProjectedBalance(in.Balances)
	if err != nil {
		if errors.Is(err, ErrInvalidProjection) {
			return ApplicationResult{}, ErrInvalidApplicationInput
		}
		return ApplicationResult{}, err
	}
	if !dateOnly(in.Balances[0].Date).Equal(dateOnly(in.AsOfDate)) {
		return ApplicationResult{}, ErrInvalidApplicationInput
	}
	available := in.EligibleAccumulatedProfit
	if available.IsNegative() {
		available = domain.ZeroMoney()
	}
	available, err = available.Sub(in.AlreadyInvested)
	if err != nil {
		return ApplicationResult{}, err
	}
	if available.IsNegative() {
		available = domain.ZeroMoney()
	}
	worst := minimum.ClosingBalance
	limit, err := worst.Sub(in.MinimumReserve)
	if err != nil {
		return ApplicationResult{}, err
	}
	if limit.IsNegative() {
		limit = domain.ZeroMoney()
	}
	max := available
	if limit.Compare(max) < 0 {
		max = limit
	}
	return ApplicationResult{in.EligibleAccumulatedProfit, in.AlreadyInvested, available, worst, in.MinimumReserve, limit, max, in.RuleVersion, domain.StatusProvisional}, nil
}

// SuggestedReserve calculates the approved, informational D-003-C suggestion
// from a validated D0-D+60 projection. It does not change MaximumApplication.
func SuggestedReserve(asOfDate time.Time, balances []DailyBalance, minimumReserve domain.Money) (domain.Money, error) {
	if asOfDate.IsZero() || minimumReserve.IsNegative() {
		return domain.ZeroMoney(), ErrInvalidApplicationInput
	}
	minimum, err := FindMinimumProjectedBalance(balances)
	if err != nil {
		return domain.ZeroMoney(), err
	}
	if !dateOnly(balances[0].Date).Equal(dateOnly(asOfDate)) {
		return domain.ZeroMoney(), ErrInvalidApplicationInput
	}
	drawdown, err := balances[0].ClosingBalance.Sub(minimum.ClosingBalance)
	if err != nil {
		return domain.ZeroMoney(), err
	}
	if drawdown.IsNegative() {
		drawdown = domain.ZeroMoney()
	}
	return minimumReserve.Add(drawdown)
}

func IsWithinProjection(b DailyBalance, now time.Time) bool {
	return !b.Date.Before(dateOnly(now)) && !b.Date.After(dateOnly(now).AddDate(0, 0, 60))
}
