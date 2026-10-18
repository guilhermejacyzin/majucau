package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/bling"
)

type fakeDailyBlingSyncer struct {
	scheduledAt time.Time
	result      bling.BlingRawSyncResult
	skipped     bool
	err         error
	calls       int
}

func (f *fakeDailyBlingSyncer) SyncDaily(_ context.Context, scheduledAt time.Time) (bling.BlingRawSyncResult, bool, error) {
	f.calls++
	f.scheduledAt = scheduledAt
	return f.result, f.skipped, f.err
}

func TestRunDailyBlingSyncPassesTheScheduledTimeAndReportsSkippedRuns(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	syncer := &fakeDailyBlingSyncer{result: bling.BlingRawSyncResult{Status: "SUCCESS"}, skipped: true}
	result, skipped, err := runDailyBlingSync(context.Background(), syncer, scheduledAt)
	if err != nil || result.Status != "SUCCESS" || !skipped || syncer.calls != 1 {
		t.Fatalf("daily sync result=%+v error=%v calls=%d", result, err, syncer.calls)
	}
	if !syncer.scheduledAt.Equal(scheduledAt) {
		t.Fatalf("scheduled time = %s, want %s", syncer.scheduledAt, scheduledAt)
	}
}

func TestRunDailyBlingSyncRejectsMissingOrIncompleteSource(t *testing.T) {
	if _, _, err := runDailyBlingSync(context.Background(), nil, time.Now()); !errors.Is(err, ErrDailyBlingUnavailable) {
		t.Fatalf("missing source error = %v", err)
	}
	syncer := &fakeDailyBlingSyncer{result: bling.BlingRawSyncResult{Status: "PARTIAL"}}
	if _, _, err := runDailyBlingSync(context.Background(), syncer, time.Now()); !errors.Is(err, ErrDailyBlingIncomplete) {
		t.Fatalf("incomplete result error = %v", err)
	}
}

func TestDailyBlingCatchUpUsesTheConfiguredLocalTime(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	before := time.Date(2026, 10, 18, 9, 29, 59, 0, location)
	if _, due := dailyBlingCatchUpAt(before); due {
		t.Fatal("run before 09:30 should wait for the regular schedule")
	}
	after := time.Date(2026, 10, 18, 14, 15, 0, 0, time.UTC)
	scheduledAt, due := dailyBlingCatchUpAt(after)
	if !due || scheduledAt.Hour() != 9 || scheduledAt.Minute() != 30 || scheduledAt.Day() != 18 || scheduledAt.Month() != time.October {
		t.Fatalf("catch-up schedule = %s due=%v", scheduledAt, due)
	}
}

func TestDailyBlingSyncErrorCodeDoesNotExposeUnderlyingError(t *testing.T) {
	err := errors.New("secret token and account data")
	if got := dailyBlingSyncErrorCode(err); got != "BLING_SYNC_FAILED" {
		t.Fatalf("error code = %q", got)
	}
}
