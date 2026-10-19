package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
)

type fakeDailyGoogleDriveSyncer struct {
	scheduledAt time.Time
	result      googledrive.DailyFileInventorySyncResult
	err         error
	calls       int
}

func (f *fakeDailyGoogleDriveSyncer) SyncDaily(_ context.Context, scheduledAt time.Time) (googledrive.DailyFileInventorySyncResult, error) {
	f.calls++
	f.scheduledAt = scheduledAt
	return f.result, f.err
}

func TestRunDailyGoogleDriveSyncPassesScheduleAndAcceptsCompleteResult(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	syncer := &fakeDailyGoogleDriveSyncer{result: googledrive.DailyFileInventorySyncResult{Status: "SUCCESS", FoldersRead: 3}}
	result, err := runDailyGoogleDriveSync(context.Background(), syncer, scheduledAt)
	if err != nil || result.Status != "SUCCESS" || syncer.calls != 1 || !syncer.scheduledAt.Equal(scheduledAt) {
		t.Fatalf("daily Drive result=%+v error=%v calls=%d scheduled=%s", result, err, syncer.calls, syncer.scheduledAt)
	}
}

func TestRunDailyGoogleDriveSyncRejectsMissingAndPartialSources(t *testing.T) {
	if _, err := runDailyGoogleDriveSync(context.Background(), nil, time.Now()); !errors.Is(err, googledrive.ErrSyncConfiguration) {
		t.Fatalf("missing Drive source error = %v", err)
	}
	syncer := &fakeDailyGoogleDriveSyncer{result: googledrive.DailyFileInventorySyncResult{Status: "PARTIAL", FoldersRead: 2, FoldersFailed: 1}}
	if _, err := runDailyGoogleDriveSync(context.Background(), syncer, time.Now()); !errors.Is(err, googledrive.ErrDailySyncIncomplete) {
		t.Fatalf("partial Drive source error = %v", err)
	}
}

func TestDailyGoogleDriveCatchUpUsesBrasiliaSchedule(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	before := time.Date(2026, 10, 18, 9, 29, 59, 0, location)
	if _, due := dailyGoogleDriveCatchUpAt(before); due {
		t.Fatal("run before 09:30 should wait for regular schedule")
	}
	after := time.Date(2026, 10, 18, 14, 15, 0, 0, time.UTC)
	scheduledAt, due := dailyGoogleDriveCatchUpAt(after)
	if !due || scheduledAt.Hour() != 9 || scheduledAt.Minute() != 30 || scheduledAt.Day() != 18 || scheduledAt.Month() != time.October {
		t.Fatalf("catch-up schedule = %s due=%v", scheduledAt, due)
	}
}

func TestDailyGoogleDriveSyncErrorCodeDoesNotExposeDetails(t *testing.T) {
	if got := dailyGoogleDriveSyncErrorCode(errors.New("sensitive file name and token")); got != "GOOGLE_DRIVE_SYNC_FAILED" {
		t.Fatalf("error code = %q", got)
	}
	if got := dailyGoogleDriveSyncErrorCode(googledrive.ErrDailySyncIncomplete); got != "GOOGLE_DRIVE_SYNC_INCOMPLETE" {
		t.Fatalf("incomplete error code = %q", got)
	}
}
