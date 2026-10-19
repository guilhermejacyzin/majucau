package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/integrations/nuvempago"
	"majucau.local/financial-intelligence/internal/security"
)

const (
	googleDriveLossesFolderEnv   = "MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID"
	googleDriveReceiptsFolderEnv = "MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID"
	googleDriveShippingFolderEnv = "MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID"
)

var (
	errGoogleDriveFoldersNotConfigured  = errors.New("Google Drive folder configuration is missing")
	errGoogleDriveConfigurationInvalid  = errors.New("Google Drive configuration is invalid")
	errGoogleDriveCredentialUnavailable = errors.New("Google Drive credential is unavailable")
)

type dailyGoogleDriveSyncer interface {
	SyncDaily(context.Context, time.Time) (googledrive.DailyFileInventorySyncResult, error)
}

type lazyDailyGoogleDriveSyncer struct {
	pool  *pgxpool.Pool
	store security.SecretStore
}

func newDailyGoogleDriveSyncer(ctx context.Context, pool *pgxpool.Pool, store security.SecretStore) (dailyGoogleDriveSyncer, string) {
	if ctx == nil {
		return nil, "GOOGLE_DRIVE_CONFIGURATION_INVALID"
	}
	if pool == nil {
		return nil, "GOOGLE_DRIVE_DATABASE_UNAVAILABLE"
	}
	if store == nil {
		return nil, "GOOGLE_DRIVE_SECRET_STORE_UNAVAILABLE"
	}
	// Keep the approved 09:30 schedule alive even when the account has not been
	// provisioned yet. Each run reloads the protected configuration so setup
	// completed while the worker is running takes effect without a restart.
	return &lazyDailyGoogleDriveSyncer{pool: pool, store: store}, ""
}

func (s *lazyDailyGoogleDriveSyncer) SyncDaily(ctx context.Context, scheduledAt time.Time) (googledrive.DailyFileInventorySyncResult, error) {
	if s == nil || ctx == nil || s.pool == nil || s.store == nil {
		return googledrive.DailyFileInventorySyncResult{Status: "FAILED"}, errGoogleDriveConfigurationInvalid
	}
	syncer, err := buildDailyGoogleDriveSyncer(ctx, s.pool, s.store)
	if err != nil {
		return googledrive.DailyFileInventorySyncResult{Status: "FAILED"}, err
	}
	return syncer.SyncDaily(ctx, scheduledAt)
}

func buildDailyGoogleDriveSyncer(ctx context.Context, pool *pgxpool.Pool, store security.SecretStore) (dailyGoogleDriveSyncer, error) {
	if ctx == nil || pool == nil || store == nil {
		return nil, errGoogleDriveConfigurationInvalid
	}
	folderIDs, configured, err := googleDriveFolderIDsFromStoreOrEnvironment(ctx, store)
	if err != nil {
		if errors.Is(err, googledrive.ErrServiceAccountCredentialStore) {
			return nil, errGoogleDriveCredentialUnavailable
		}
		return nil, errGoogleDriveConfigurationInvalid
	}
	if !configured {
		return nil, errGoogleDriveFoldersNotConfigured
	}
	tokens, err := googledrive.NewServiceAccountTokenProvider(ctx, store)
	if err != nil {
		return nil, errGoogleDriveCredentialUnavailable
	}
	client, err := googledrive.NewClient(tokens, []string{
		folderIDs[googledrive.DriveFolderOperationalLosses],
		folderIDs[googledrive.DriveFolderNuvemPagoReceipts],
		folderIDs[googledrive.DriveFolderNuvemShipping],
	})
	if err != nil {
		return nil, errGoogleDriveConfigurationInvalid
	}
	folderSync, err := googledrive.NewFileInventorySyncService(pool, client)
	if err != nil {
		return nil, errGoogleDriveConfigurationInvalid
	}
	dailySync, err := googledrive.NewDailyFileInventorySyncService(folderSync, folderIDs)
	if err != nil {
		return nil, errGoogleDriveConfigurationInvalid
	}
	futureImporter := nuvempago.NewFutureImportService(pool)
	pipeline, err := newDailyGoogleDriveImportSyncer(dailySync, folderSync, client, futureImporter, os.TempDir(), folderIDs)
	if err != nil {
		return nil, errGoogleDriveConfigurationInvalid
	}
	return pipeline, nil
}

func googleDriveFolderIDsFromStoreOrEnvironment(ctx context.Context, store security.SecretStore) (googledrive.DriveFolderIDs, bool, error) {
	protected, configured, err := googledrive.ReadDriveFolderIDs(ctx, store)
	if err != nil || configured {
		return protected, configured, err
	}
	return googleDriveFolderIDsFromEnvironment()
}

func googleDriveFolderIDsFromEnvironment() (googledrive.DriveFolderIDs, bool, error) {
	values := googledrive.DriveFolderIDs{
		googledrive.DriveFolderOperationalLosses: strings.TrimSpace(os.Getenv(googleDriveLossesFolderEnv)),
		googledrive.DriveFolderNuvemPagoReceipts: strings.TrimSpace(os.Getenv(googleDriveReceiptsFolderEnv)),
		googledrive.DriveFolderNuvemShipping:     strings.TrimSpace(os.Getenv(googleDriveShippingFolderEnv)),
	}
	configured := 0
	for _, value := range values {
		if value != "" {
			configured++
		}
	}
	if configured == 0 {
		return nil, false, nil
	}
	if googledrive.ValidateDriveFolderIDs(values) != nil {
		return nil, false, googledrive.ErrSyncConfiguration
	}
	return values, true, nil
}
