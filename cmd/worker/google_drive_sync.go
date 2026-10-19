package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"majucau.local/financial-intelligence/internal/integrations/googledrive"
	"majucau.local/financial-intelligence/internal/security"
)

const (
	googleDriveLossesFolderEnv   = "MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID"
	googleDriveReceiptsFolderEnv = "MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID"
	googleDriveShippingFolderEnv = "MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID"
)

type dailyGoogleDriveSyncer interface {
	SyncDaily(context.Context, time.Time) (googledrive.DailyFileInventorySyncResult, error)
}

func newDailyGoogleDriveSyncer(ctx context.Context, pool *pgxpool.Pool, store security.SecretStore) (dailyGoogleDriveSyncer, string) {
	folderIDs, configured, err := googleDriveFolderIDsFromEnvironment()
	if err != nil {
		return nil, "GOOGLE_DRIVE_CONFIGURATION_INVALID"
	}
	if !configured {
		return nil, "GOOGLE_DRIVE_NOT_CONFIGURED"
	}
	if pool == nil {
		return nil, "GOOGLE_DRIVE_DATABASE_UNAVAILABLE"
	}
	if store == nil {
		return nil, "GOOGLE_DRIVE_SECRET_STORE_UNAVAILABLE"
	}
	tokens, err := googledrive.NewServiceAccountTokenProvider(ctx, store)
	if err != nil {
		return nil, "GOOGLE_DRIVE_CREDENTIAL_UNAVAILABLE"
	}
	client, err := googledrive.NewClient(tokens, []string{
		folderIDs[googledrive.DriveFolderOperationalLosses],
		folderIDs[googledrive.DriveFolderNuvemPagoReceipts],
		folderIDs[googledrive.DriveFolderNuvemShipping],
	})
	if err != nil {
		return nil, "GOOGLE_DRIVE_CONFIGURATION_INVALID"
	}
	folderSync, err := googledrive.NewFileInventorySyncService(pool, client)
	if err != nil {
		return nil, "GOOGLE_DRIVE_CONFIGURATION_INVALID"
	}
	dailySync, err := googledrive.NewDailyFileInventorySyncService(folderSync, folderIDs)
	if err != nil {
		return nil, "GOOGLE_DRIVE_CONFIGURATION_INVALID"
	}
	return dailySync, ""
}

func googleDriveFolderIDsFromEnvironment() (map[googledrive.DriveFolderScope]string, bool, error) {
	values := map[googledrive.DriveFolderScope]string{
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
	if configured != len(values) {
		return nil, false, googledrive.ErrSyncConfiguration
	}
	return values, true, nil
}
