package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidProjection = errors.New("invalid projection input")

type Movement struct {
	Date                           time.Time
	Inflows, Outflows, Adjustments domain.Money
}
type BalanceType string

const (
	Realized    BalanceType = "REALIZED"
	Provisional BalanceType = "PROVISIONAL"
	Projected   BalanceType = "PROJECTED"
)

type DailyBalance struct {
	Date           time.Time
	OpeningBalance domain.Money
	Inflows        domain.Money
	Outflows       domain.Money
	Adjustments    domain.Money
	ClosingBalance domain.Money
	Type           BalanceType
}

// DailyProjection returns D0 through D+60 (61 consecutive calendar dates).
// Dates are compared in the caller's business timezone; the date portion is
// normalized to midnight UTC to make results independent of local process TZ.
func DailyProjection(start time.Time, opening domain.Money, movements []Movement) ([]DailyBalance, error) {
	start = dateOnly(start)
	byDate := make(map[string]Movement, len(movements))
	for _, m := range movements {
		m.Date = dateOnly(m.Date)
		key := m.Date.Format("2006-01-02")
		old, ok := byDate[key]
		if ok {
			var err error
			old.Inflows, err = old.Inflows.Add(m.Inflows)
			if err != nil {
				return nil, err
			}
			old.Outflows, err = old.Outflows.Add(m.Outflows)
			if err != nil {
				return nil, err
			}
			old.Adjustments, err = old.Adjustments.Add(m.Adjustments)
			if err != nil {
				return nil, err
			}
			byDate[key] = old
		} else {
			byDate[key] = m
		}
	}
	result := make([]DailyBalance, 0, 61)
	closing := opening
	for day := 0; day <= 60; day++ {
		date := start.AddDate(0, 0, day)
		m := byDate[date.Format("2006-01-02")]
		var err error
		closing, err = closing.Add(m.Inflows)
		if err != nil {
			return nil, err
		}
		closing, err = closing.Sub(m.Outflows)
		if err != nil {
			return nil, err
		}
		closing, err = closing.Add(m.Adjustments)
		if err != nil {
			return nil, err
		}
		openingForDay := opening
		if day > 0 {
			openingForDay = result[day-1].ClosingBalance
		}
		kind := Projected
		if day == 0 {
			kind = Provisional
		}
		result = append(result, DailyBalance{Date: date, OpeningBalance: openingForDay, Inflows: m.Inflows, Outflows: m.Outflows, Adjustments: m.Adjustments, ClosingBalance: closing, Type: kind})
	}
	return result, nil
}

// MinimumBalance is the dated low point of a complete D0-D+60 projection.
// The movement fields are retained so the dashboard and drill-down can explain
// the low point without recomputing or losing its daily composition.
type MinimumBalance struct {
	Date           time.Time
	ClosingBalance domain.Money
	Inflows        domain.Money
	Outflows       domain.Money
	Adjustments    domain.Money
}

// FindMinimumProjectedBalance validates and finds the first lowest closing
// balance in an approved 61-day projection. The first lowest point is chosen
// deterministically when two days have the same value. Requiring the temporal
// and status contract prevents D-1 data or a truncated series from appearing
// as the projected minimum.
func FindMinimumProjectedBalance(rows []DailyBalance) (MinimumBalance, error) {
	if len(rows) != 61 {
		return MinimumBalance{}, ErrInvalidProjection
	}
	for index, row := range rows {
		if row.Date.IsZero() {
			return MinimumBalance{}, ErrInvalidProjection
		}
		if index == 0 {
			if row.Type != Provisional {
				return MinimumBalance{}, ErrInvalidProjection
			}
		} else {
			expectedDate := dateOnly(rows[index-1].Date).AddDate(0, 0, 1)
			if !dateOnly(row.Date).Equal(expectedDate) || row.Type != Projected {
				return MinimumBalance{}, ErrInvalidProjection
			}
		}
		expectedClosing, err := row.OpeningBalance.Add(row.Inflows)
		if err != nil {
			return MinimumBalance{}, err
		}
		expectedClosing, err = expectedClosing.Sub(row.Outflows)
		if err != nil {
			return MinimumBalance{}, err
		}
		expectedClosing, err = expectedClosing.Add(row.Adjustments)
		if err != nil || expectedClosing.Compare(row.ClosingBalance) != 0 {
			if err != nil {
				return MinimumBalance{}, err
			}
			return MinimumBalance{}, ErrInvalidProjection
		}
	}

	minimum := rows[0]
	for _, row := range rows[1:] {
		if row.ClosingBalance.Compare(minimum.ClosingBalance) < 0 {
			minimum = row
		}
	}
	return MinimumBalance{
		Date:           dateOnly(minimum.Date),
		ClosingBalance: minimum.ClosingBalance,
		Inflows:        minimum.Inflows,
		Outflows:       minimum.Outflows,
		Adjustments:    minimum.Adjustments,
	}, nil
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

