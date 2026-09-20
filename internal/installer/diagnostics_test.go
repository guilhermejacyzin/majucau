package installer

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteDiagnosticBundleIsSanitizedAndDeterministicContract(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "support.zip")
	result := Result{
		SchemaVersion: SchemaVersion,
		Status:        StatusBlocked,
		ExitCode:      ExitBlocked,
		Issues:        []Issue{{Code: "PATH_UNSAFE_LOCATION", Check: "data_dir"}},
	}
	info, err := WriteDiagnosticBundle(path, result, DiagnosticMetadata{
		AppVersion:    "0.1.0-test",
		CorrelationID: "corr-test",
		GeneratedAt:   time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Bytes <= 0 || len(info.SHA256) != 64 || len(info.Entries) != 2 {
		t.Fatalf("unexpected bundle info: %#v", info)
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	seen := map[string]bool{}
	for _, entry := range archive.File {
		seen[entry.Name] = true
		reader, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(reader)
		_ = reader.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(content), dir) || strings.Contains(string(content), "secret-token") {
			t.Fatalf("bundle leaked local or secret data: %s", entry.Name)
		}
		if entry.Name == "diagnostic.json" {
			var document diagnosticDocument
			if err := json.Unmarshal(content, &document); err != nil {
				t.Fatal(err)
			}
			if document.CorrelationID != "corr-test" || document.Preflight.Status != StatusBlocked {
				t.Fatalf("unexpected diagnostic document: %#v", document)
			}
		}
	}
	if !seen["diagnostic.json"] || !seen["README.txt"] {
		t.Fatalf("missing entries: %#v", seen)
	}
}

func TestWriteDiagnosticBundleRejectsRelativeAndNonZipPaths(t *testing.T) {
	for _, path := range []string{"support.zip", filepath.Join(t.TempDir(), "support.txt")} {
		if _, err := WriteDiagnosticBundle(path, Result{}, DiagnosticMetadata{}); err != ErrDiagnosticOutput {
			t.Fatalf("path %q error = %v, want %v", path, err, ErrDiagnosticOutput)
		}
	}
}

func TestDiagnosticBundleDoesNotLeaveTemporaryFileOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "support.zip")
	if _, err := WriteDiagnosticBundle(path, Result{}, DiagnosticMetadata{}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".majucau-diagnostic-") {
			t.Fatalf("temporary file was not removed: %s", entry.Name())
		}
	}
}
