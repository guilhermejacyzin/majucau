package googledrive

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type fakeFolderInventorySyncer struct {
	results map[DriveFolderScope]FileInventorySyncResult
	errors  map[DriveFolderScope]error
	calls   []DriveFolderScope
	ids     map[DriveFolderScope]pgtype.UUID
}

func (f *fakeFolderInventorySyncer) SyncFolder(_ context.Context, scope DriveFolderScope, _ string, correlation pgtype.UUID) (FileInventorySyncResult, error) {
	f.calls = append(f.calls, scope)
	if f.ids == nil {
		f.ids = make(map[DriveFolderScope]pgtype.UUID)
	}
	f.ids[scope] = correlation
	return f.results[scope], f.errors[scope]
}

func TestDailyDriveCorrelationIDIsStablePerDateAndScope(t *testing.T) {
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	first, err := dailyDriveCorrelationID(scheduledAt, DriveFolderNuvemPagoReceipts)
	if err != nil {
		t.Fatal(err)
	}
	again, err := dailyDriveCorrelationID(scheduledAt.In(time.UTC), DriveFolderNuvemPagoReceipts)
	if err != nil || first != again {
		t.Fatalf("correlation IDs differ for the same São Paulo date: %s %s error=%v", first, again, err)
	}
	otherScope, err := dailyDriveCorrelationID(scheduledAt, DriveFolderNuvemShipping)
	if err != nil || first == otherScope {
		t.Fatalf("different folder scopes must have different IDs: %s %s error=%v", first, otherScope, err)
	}
	wrongTime := scheduledAt.Add(time.Minute)
	if _, err := dailyDriveCorrelationID(wrongTime, DriveFolderNuvemPagoReceipts); !errors.Is(err, ErrDailySyncScheduleInvalid) {
		t.Fatalf("unexpected schedule time error = %v", err)
	}
}

func TestDailyDriveSyncContinuesOtherFoldersWhenOneFails(t *testing.T) {
	ctx := context.Background()
	syncer := &fakeFolderInventorySyncer{
		results: map[DriveFolderScope]FileInventorySyncResult{
			DriveFolderOperationalLosses: {Status: "SUCCESS", FilesRead: 3},
			DriveFolderNuvemShipping:     {Status: "SUCCESS", FilesRead: 2, Skipped: true},
		},
		errors: map[DriveFolderScope]error{DriveFolderNuvemPagoReceipts: errors.New("sanitized test error")},
	}
	service, err := NewDailyFileInventorySyncService(syncer, map[DriveFolderScope]string{
		DriveFolderOperationalLosses: "folder-a",
		DriveFolderNuvemPagoReceipts: "folder-b",
		DriveFolderNuvemShipping:     "folder-c",
	})
	if err != nil {
		t.Fatal(err)
	}
	scheduledAt := time.Date(2026, 10, 18, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	result, err := service.SyncDaily(ctx, scheduledAt)
	if !errors.Is(err, ErrDailySyncIncomplete) {
		t.Fatalf("incomplete error = %v", err)
	}
	if result.Status != "PARTIAL" || result.FoldersRead != 2 || result.FoldersFailed != 1 || result.FoldersSkipped != 1 || result.FilesRead != 5 {
		t.Fatalf("unexpected partial result: %+v", result)
	}
	if len(result.FolderResults) != len(approvedDriveFolderScopes) {
		t.Fatalf("folder results = %d, want %d", len(result.FolderResults), len(approvedDriveFolderScopes))
	}
	for index, scope := range approvedDriveFolderScopes {
		if result.FolderResults[index].Scope != scope {
			t.Fatalf("folder result %d scope = %q, want %q", index, result.FolderResults[index].Scope, scope)
		}
	}
	if result.FolderResults[1].Status != "FAILED" {
		t.Fatalf("failed folder status = %q, want FAILED", result.FolderResults[1].Status)
	}
	if len(syncer.calls) != len(approvedDriveFolderScopes) {
		t.Fatalf("processed %d folders, want all %d", len(syncer.calls), len(approvedDriveFolderScopes))
	}
	if len(syncer.ids) != len(approvedDriveFolderScopes) {
		t.Fatalf("not all folders received a stable correlation ID: %d", len(syncer.ids))
	}
}

func TestNewDailyDriveSyncServiceRequiresThreeDistinctApprovedFolders(t *testing.T) {
	syncer := &fakeFolderInventorySyncer{results: map[DriveFolderScope]FileInventorySyncResult{}, errors: map[DriveFolderScope]error{}}
	for _, folders := range []map[DriveFolderScope]string{
		{},
		{DriveFolderOperationalLosses: "folder-a"},
		{
			DriveFolderOperationalLosses: "folder-a",
			DriveFolderNuvemPagoReceipts: "folder-a",
			DriveFolderNuvemShipping:     "folder-c",
		},
	} {
		if _, err := NewDailyFileInventorySyncService(syncer, folders); !errors.Is(err, ErrSyncConfiguration) {
			t.Fatalf("invalid folder config error = %v", err)
		}
	}
}
