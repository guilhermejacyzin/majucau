package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"majucau.local/financial-intelligence/internal/ipc"
)

func TestReadDriveCredentialFileReturnsOnlyBoundedRegularFileContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service-account.json")
	want := []byte(`{"type":"service_account","private_key":"test-only"}`)
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readDriveCredentialFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("credential contents match=%v error=%v", bytes.Equal(got, want), err)
	}
	wipeDriveCredentialBytes(got)
}

func TestReadDriveCredentialFileRejectsDirectoriesAndOversizedInput(t *testing.T) {
	root := t.TempDir()
	if _, err := readDriveCredentialFile(root); err == nil {
		t.Fatal("directory was accepted as a credential file")
	}
	path := filepath.Join(root, "large.json")
	if err := os.WriteFile(path, make([]byte, 1<<20+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readDriveCredentialFile(path); err == nil {
		t.Fatal("oversized credential file was accepted")
	}
}

func TestReadDriveFolderIDsFileStreamsAndValidatesApprovedRoots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drive-folders.json")
	contents := []byte(`{"01_perdas_operacionais":"losses-folder","02_nuvem_pago/recebimentos_futuros":"receipts-folder","03_nuvem_envio":"shipping-folder"}`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readDriveFolderIDsFile(path)
	if err != nil || len(got) != 3 || got["01_perdas_operacionais"] != "losses-folder" {
		t.Fatalf("folder IDs=%v error=%v", got, err)
	}
	for _, invalid := range []string{
		`{"01_perdas_operacionais":"losses-folder","02_nuvem_pago/recebimentos_futuros":"receipts-folder"}`,
		`{"01_perdas_operacionais":"same","02_nuvem_pago/recebimentos_futuros":"same","03_nuvem_envio":"shipping-folder"}`,
		`{"unexpected":"folder"}`,
		`{"01_perdas_operacionais":"losses-folder","02_nuvem_pago/recebimentos_futuros":"receipts-folder","03_nuvem_envio":"shipping-folder"} {}`,
	} {
		if err := os.WriteFile(path, []byte(invalid), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := readDriveFolderIDsFile(path); err == nil {
			t.Fatalf("invalid folder configuration accepted: %s", invalid)
		}
	}
}

func TestNormalizeResolvedDrivePathAcceptsExtendedDrivePrefixOnly(t *testing.T) {
	got, ok := normalizeResolvedDrivePath(`\\?\C:\Users\runner\Temp\service-account.json`)
	if !ok || got != `C:\Users\runner\Temp\service-account.json` {
		t.Fatalf("extended local path = %q, accepted=%v", got, ok)
	}
	for _, path := range []string{
		`\\?\UNC\server\share\service-account.json`,
		`\\server\share\service-account.json`,
		`\\.\C:\device\service-account.json`,
	} {
		if got, ok := normalizeResolvedDrivePath(path); ok {
			t.Fatalf("non-local path %q was accepted as %q", path, got)
		}
	}
}

func TestDriveCredentialProvisionErrorCodeUsesOnlyKnownCodes(t *testing.T) {
	if got := driveCredentialProvisionErrorCode(ipc.NewErrorResponse("id", "GOOGLE_DRIVE_CREDENTIAL_INVALID", "safe")); got != "GOOGLE_DRIVE_CREDENTIAL_INVALID" {
		t.Fatalf("known error code = %q", got)
	}
	for _, untrusted := range []string{"GOOGLE_DRIVE_PRIVATE_KEY_LEAK\nsecret", "BLING_ERROR", ""} {
		response := ipc.NewErrorResponse("id", untrusted, "untrusted message")
		if got := driveCredentialProvisionErrorCode(response); got != "GOOGLE_DRIVE_CREDENTIAL_SAVE_FAILED" {
			t.Fatalf("untrusted error code %q was returned as %q", untrusted, got)
		}
	}
}
