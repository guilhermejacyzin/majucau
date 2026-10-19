package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/scheduler"
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

func TestDailyGoogleDriveMissedRunWaitsUntilNextBrasiliaSchedule(t *testing.T) {
	schedule, err := scheduler.NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}
	startedAfterRun := time.Date(2026, 10, 18, 14, 15, 0, 0, time.UTC)
	nextRun, err := schedule.NextRun(startedAfterRun)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := nextRun.Format(time.RFC3339), "2026-10-19T09:30:00-03:00"; got != want {
		t.Fatalf("next Drive run after missed schedule = %s, want %s", got, want)
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
