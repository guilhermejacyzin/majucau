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

func TestDailyGoogleDriveCatchUpUsesConfiguredBrasiliaSchedule(t *testing.T) {
	schedule, err := scheduler.NewSaoPaulo0930()
	if err != nil {
		t.Fatal(err)
	}
	location := time.FixedZone("UTC-3", -3*60*60)
	beforeRun := time.Date(2026, 10, 18, 9, 29, 59, 0, location)
	syncer := &fakeDailyGoogleDriveSyncer{result: googledrive.DailyFileInventorySyncResult{Status: "SUCCESS", FoldersRead: 3}}
	if _, _, due, err := runDailyGoogleDriveCatchUp(context.Background(), syncer, beforeRun); due || err != nil || syncer.calls != 0 {
		t.Fatalf("Drive catch-up before schedule due=%v error=%v calls=%d; want no run", due, err, syncer.calls)
	}
	if _, _, due, err := runDailyGoogleDriveCatchUp(context.Background(), syncer, time.Time{}); due || err != nil || syncer.calls != 0 {
		t.Fatalf("Drive catch-up with zero time due=%v error=%v calls=%d; want no run", due, err, syncer.calls)
	}
	atSchedule := time.Date(2026, 10, 18, 9, 30, 0, 0, location)
	boundarySyncer := &fakeDailyGoogleDriveSyncer{result: googledrive.DailyFileInventorySyncResult{Status: "SUCCESS", FoldersRead: 3}}
	_, boundaryScheduledAt, due, err := runDailyGoogleDriveCatchUp(context.Background(), boundarySyncer, atSchedule)
	if err != nil || !due || boundarySyncer.calls != 1 || !boundaryScheduledAt.Equal(atSchedule) {
		t.Fatalf("Drive catch-up at schedule = %s due=%v error=%v calls=%d; want one run at 2026-10-18T09:30:00-03:00", boundaryScheduledAt.Format(time.RFC3339), due, err, boundarySyncer.calls)
	}
	startedAfterRun := time.Date(2026, 10, 18, 14, 15, 0, 0, time.UTC)
	_, scheduledAt, due, err := runDailyGoogleDriveCatchUp(context.Background(), syncer, startedAfterRun)
	if err != nil || !due || syncer.calls != 1 || !syncer.scheduledAt.Equal(scheduledAt) || scheduledAt.Format(time.RFC3339) != "2026-10-18T09:30:00-03:00" {
		t.Fatalf("Drive catch-up = %s due=%v error=%v calls=%d; want one run at 2026-10-18T09:30:00-03:00", scheduledAt.Format(time.RFC3339), due, err, syncer.calls)
	}
	previousSchedule, err := previousGoogleDriveSchedule(scheduledAt)
	if err != nil || previousSchedule.Format(time.RFC3339) != "2026-10-17T09:30:00-03:00" {
		t.Fatalf("catch-up file cutoff = %s error=%v; want previous day at 09:30", previousSchedule.Format(time.RFC3339), err)
	}
	nextRun, err := schedule.NextRun(startedAfterRun)
	if err != nil || nextRun.Format(time.RFC3339) != "2026-10-19T09:30:00-03:00" {
		t.Fatalf("next regular Drive run = %s error=%v; want 2026-10-19T09:30:00-03:00", nextRun.Format(time.RFC3339), err)
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
