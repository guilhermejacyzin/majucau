package main

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
)

type emptyGoogleDriveCredentialStore struct{}

func (emptyGoogleDriveCredentialStore) Put(context.Context, string, []byte) error { return nil }
func (emptyGoogleDriveCredentialStore) Get(context.Context, string) ([]byte, error) {
	return nil, fs.ErrNotExist
}
func (emptyGoogleDriveCredentialStore) Delete(context.Context, string) error { return nil }

func TestGoogleDriveFolderIDsFromEnvironmentRequiresAllThreeIDs(t *testing.T) {
	t.Setenv(googleDriveLossesFolderEnv, "folder-a")
	t.Setenv(googleDriveReceiptsFolderEnv, "")
	t.Setenv(googleDriveShippingFolderEnv, "folder-c")
	if _, configured, err := googleDriveFolderIDsFromEnvironment(); configured || !errors.Is(err, googledrive.ErrSyncConfiguration) {
		t.Fatalf("partial folder configuration = configured:%v error:%v", configured, err)
	}
}

func TestGoogleDriveFolderIDsFromEnvironmentAcceptsThreeConfiguredIDs(t *testing.T) {
	t.Setenv(googleDriveLossesFolderEnv, " folder-a ")
	t.Setenv(googleDriveReceiptsFolderEnv, "folder-b")
	t.Setenv(googleDriveShippingFolderEnv, "folder-c")
	folders, configured, err := googleDriveFolderIDsFromEnvironment()
	if err != nil || !configured || len(folders) != 3 || folders[googledrive.DriveFolderOperationalLosses] != "folder-a" {
		t.Fatalf("folder configuration = %#v configured:%v error:%v", folders, configured, err)
	}
}

func TestGoogleDriveFolderIDsFromEnvironmentAllowsAllIDsToBeAbsent(t *testing.T) {
	t.Setenv(googleDriveLossesFolderEnv, "")
	t.Setenv(googleDriveReceiptsFolderEnv, "")
	t.Setenv(googleDriveShippingFolderEnv, "")
	folders, configured, err := googleDriveFolderIDsFromEnvironment()
	if err != nil || configured || folders != nil {
		t.Fatalf("empty folder configuration = %#v configured:%v error:%v", folders, configured, err)
	}
}

func TestGoogleDriveDailyScheduleCanLoadConfigurationAddedAfterWorkerStarts(t *testing.T) {
	t.Setenv(googleDriveLossesFolderEnv, "")
	t.Setenv(googleDriveReceiptsFolderEnv, "")
	t.Setenv(googleDriveShippingFolderEnv, "")
	syncer, setupCode := newDailyGoogleDriveSyncer(context.Background(), &pgxpool.Pool{}, emptyGoogleDriveCredentialStore{})
	if syncer == nil || setupCode != "" {
		t.Fatalf("daily syncer=%v setupCode=%q; schedule should remain available before Drive setup", syncer, setupCode)
	}
	scheduledAt := time.Date(2026, 10, 19, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	_, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, errGoogleDriveFoldersNotConfigured) {
		t.Fatalf("first sync error=%v; want missing folder configuration", err)
	}
	t.Setenv(googleDriveLossesFolderEnv, "folder-a")
	t.Setenv(googleDriveReceiptsFolderEnv, "folder-b")
	t.Setenv(googleDriveShippingFolderEnv, "folder-c")
	_, err = syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, errGoogleDriveCredentialUnavailable) {
		t.Fatalf("sync after adding folder IDs error=%v; want missing service credential", err)
	}
}
