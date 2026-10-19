package main

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type emptyGoogleDriveCredentialStore struct{}

func (emptyGoogleDriveCredentialStore) Put(context.Context, string, []byte) error { return nil }
func (emptyGoogleDriveCredentialStore) Get(context.Context, string) ([]byte, error) {
	return nil, fs.ErrNotExist
}
func (emptyGoogleDriveCredentialStore) Delete(context.Context, string) error { return nil }

func TestGoogleDriveFolderIDsRequireProtectedBundle(t *testing.T) {
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID", "folder-a")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID", "folder-b")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID", "folder-c")
	folders, configured, err := googleDriveFolderIDsFromProtectedStore(context.Background(), emptyGoogleDriveCredentialStore{})
	if err != nil || configured || folders != nil {
		t.Fatalf("folder configuration = %#v configured:%v error:%v; environment variables must not configure Drive access", folders, configured, err)
	}
}

func TestGoogleDriveDailyScheduleCanLoadConfigurationAddedAfterWorkerStarts(t *testing.T) {
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID", "")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID", "")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID", "")
	syncer, setupCode := newDailyGoogleDriveSyncer(context.Background(), &pgxpool.Pool{}, emptyGoogleDriveCredentialStore{})
	if syncer == nil || setupCode != "" {
		t.Fatalf("daily syncer=%v setupCode=%q; schedule should remain available before Drive setup", syncer, setupCode)
	}
	scheduledAt := time.Date(2026, 10, 19, 9, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	_, err := syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, errGoogleDriveFoldersNotConfigured) {
		t.Fatalf("first sync error=%v; want missing folder configuration", err)
	}
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID", "folder-a")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID", "folder-b")
	t.Setenv("MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID", "folder-c")
	_, err = syncer.SyncDaily(context.Background(), scheduledAt)
	if !errors.Is(err, errGoogleDriveFoldersNotConfigured) {
		t.Fatalf("sync after adding environment folder IDs error=%v; want protected folder bundle missing", err)
	}
}
