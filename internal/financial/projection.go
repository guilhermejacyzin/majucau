package financial

import (
	"context"
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

var ErrInvalidProjection = errors.New("invalid projection input")

type Movement struct {
	Date                           time.Time
	Inflows, Outflows, Adjustments domain.Money
}

// MovementReader yields one movement at a time. The callback can read from a
// database cursor and return its context or rows error without buffering the
// full result set.
type MovementReader func(context.Context) (Movement, bool, error)

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

// DailyProjection is a slice-backed compatibility wrapper for callers that
// already hold a small set of movements. Database-backed production readers
// should use DailyProjectionFrom to avoid buffering the full source result.
// Dates are compared in the caller's business timezone; the date portion is
// normalized to midnight UTC to make results independent of local process TZ.
func DailyProjection(start time.Time, opening domain.Money, movements []Movement) ([]DailyBalance, error) {
	index := 0
	return DailyProjectionFrom(context.Background(), start, opening, func(context.Context) (Movement, bool, error) {
		if index >= len(movements) {
			return Movement{}, false, nil
		}
		movement := movements[index]
		index++
		return movement, true, nil
	})
}

// DailyProjectionFrom consumes movements once and retains only daily totals
// within D0 through D+60. Its memory use is bounded by the 61-day output,
// regardless of the number of source movements.
func DailyProjectionFrom(ctx context.Context, start time.Time, opening domain.Money, next MovementReader) ([]DailyBalance, error) {
	if ctx == nil || next == nil {
		return nil, ErrInvalidProjection
	}
	start = dateOnly(start)
	end := start.AddDate(0, 0, 60)
	var dailyMovements [61]Movement
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		m, ok, err := next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		m.Date = dateOnly(m.Date)
		if m.Date.Before(start) || m.Date.After(end) {
			continue
		}
		dayIndex := int(m.Date.Sub(start) / (24 * time.Hour))
		day := &dailyMovements[dayIndex]
		day.Inflows, err = day.Inflows.Add(m.Inflows)
		if err != nil {
			return nil, err
		}
		day.Outflows, err = day.Outflows.Add(m.Outflows)
		if err != nil {
			return nil, err
		}
		day.Adjustments, err = day.Adjustments.Add(m.Adjustments)
		if err != nil {
			return nil, err
		}
	}
	result := make([]DailyBalance, 0, 61)
	closing := opening
	for day := 0; day <= 60; day++ {
		date := start.AddDate(0, 0, day)
		m := dailyMovements[day]
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
			if !dateOnly(row.Date).Equal(expectedDate) || row.Type != Projected ||
				row.OpeningBalance.Compare(rows[index-1].ClosingBalance) != 0 {
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
