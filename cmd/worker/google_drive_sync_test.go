package main

import (
	"errors"
	"testing"

	"majucau.local/financial-intelligence/internal/integrations/googledrive"
)

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
