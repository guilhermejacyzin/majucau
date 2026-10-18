package bling

import (
	"errors"
	"testing"
	"time"
)

func TestDailyBlingCorrelationIDIsStableBySaoPauloDate(t *testing.T) {
	local := time.FixedZone("Brasilia", -3*60*60)
	first, err := dailyBlingCorrelationID(time.Date(2026, 10, 18, 9, 30, 0, 0, local))
	if err != nil {
		t.Fatal(err)
	}
	second, err := dailyBlingCorrelationID(time.Date(2026, 10, 18, 12, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	nextDay, err := dailyBlingCorrelationID(time.Date(2026, 10, 19, 9, 30, 0, 0, local))
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("same São Paulo date produced different IDs: %s and %s", first, second)
	}
	if first == nextDay {
		t.Fatal("different scheduled dates produced the same ID")
	}
}

func TestDailyBlingCorrelationIDRequiresTheApprovedSchedule(t *testing.T) {
	if _, err := dailyBlingCorrelationID(time.Time{}); !errors.Is(err, ErrBlingDailyScheduleInvalid) {
		t.Fatalf("zero time error = %v", err)
	}
	wrongTime := time.Date(2026, 10, 18, 9, 31, 0, 0, time.FixedZone("Brasilia", -3*60*60))
	if _, err := dailyBlingCorrelationID(wrongTime); !errors.Is(err, ErrBlingDailyScheduleInvalid) {
		t.Fatalf("wrong schedule time error = %v", err)
	}
}
