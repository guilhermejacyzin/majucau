package googledrive

import (
	"bytes"
	"strings"
	"testing"
)

func TestDriveCheckpointMetadataIsCanonicalAndBounded(t *testing.T) {
	file := File{
		ID: "file-a", Name: "ledger.csv", MIMEType: "text/csv", Size: 12, SizeKnown: true,
		ModifiedTime: "2026-10-18T09:00:00Z", MD5Checksum: "0123456789abcdef0123456789abcdef",
		Version: "1", Parents: []string{"folder-a"},
	}
	payload, err := marshalDriveCheckpointFiles([]File{file})
	if err != nil {
		t.Fatalf("marshal checkpoint metadata: %v", err)
	}
	canonical, files, err := canonicalDriveCheckpointPayload(payload)
	if err != nil {
		t.Fatalf("canonicalize checkpoint metadata: %v", err)
	}
	if !bytes.Equal(payload, canonical) || len(files) != 1 || files[0].ID != file.ID || files[0].Parents[0] != "folder-a" {
		t.Fatalf("checkpoint metadata round trip = %s, files=%#v", canonical, files)
	}
	empty, err := marshalDriveCheckpointFiles(nil)
	if err != nil || string(empty) != "[]" {
		t.Fatalf("empty page payload = %q, error=%v; want an empty JSON array", empty, err)
	}
	if _, _, err := canonicalDriveCheckpointPayload([]byte("null")); err == nil {
		t.Fatal("null checkpoint payload was accepted")
	}
}

func TestDriveCheckpointMetadataRejectsOversizedFilename(t *testing.T) {
	file := File{ID: "file-a", Name: strings.Repeat("x", 4097), MIMEType: "text/csv", Parents: []string{"folder-a"}}
	if _, err := marshalDriveCheckpointFiles([]File{file}); err == nil {
		t.Fatal("oversized Drive filename was accepted")
	}
}

func TestDriveCheckpointResourcesAreRestrictedToApprovedFolders(t *testing.T) {
	for _, scope := range []DriveFolderScope{DriveFolderOperationalLosses, DriveFolderNuvemPagoReceipts, DriveFolderNuvemShipping} {
		if !driveResourceValid(googleDriveCheckpointResource + string(scope)) {
			t.Errorf("approved folder scope %q was rejected", scope)
		}
	}
	if driveResourceValid("google_drive.files.arbitrary-folder") {
		t.Fatal("unapproved Drive folder resource was accepted")
	}
}
