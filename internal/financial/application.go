package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidApplicationInput = errors.New("invalid application input")

type ApplicationInput struct {
	EligibleAccumulatedProfit, AlreadyInvested, MinimumReserve domain.Money
	Balances                                                   []DailyBalance
	RuleVersion                                                string
}
type ApplicationResult struct {
	EligibleAccumulatedProfit, AlreadyInvested, AvailableProfit, WorstWeekCash, MinimumReserve, FinancialLimit, MaximumInvestment domain.Money
	RuleVersion                                                                                                                   string
	Status                                                                                                                        domain.DataStatus
}

// MaximumApplication implements D-003-A. Only balances supplied by an
// approved source should be passed by the caller; this package does not infer
// or manufacture expected receipts.
func MaximumApplication(in ApplicationInput) (ApplicationResult, error) {
	if in.RuleVersion == "" || len(in.Balances) != 61 ||
		in.AlreadyInvested.IsNegative() || in.MinimumReserve.IsNegative() {
		return ApplicationResult{}, ErrInvalidApplicationInput
	}
	for index := 1; index < len(in.Balances); index++ {
		expected := dateOnly(in.Balances[index-1].Date).AddDate(0, 0, 1)
		if !dateOnly(in.Balances[index].Date).Equal(expected) {
			return ApplicationResult{}, ErrInvalidApplicationInput
		}
	}
	available := in.EligibleAccumulatedProfit
	var err error
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
	worst := in.Balances[0].ClosingBalance
	for _, b := range in.Balances[1:] {
		if b.ClosingBalance.Compare(worst) < 0 {
			worst = b.ClosingBalance
		}
	}
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

func IsWithinProjection(b DailyBalance, now time.Time) bool {
	return !b.Date.Before(dateOnly(now)) && !b.Date.After(dateOnly(now).AddDate(0, 0, 60))
}
