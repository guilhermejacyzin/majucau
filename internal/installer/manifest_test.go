package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReleaseManifestAcceptsPackageAndMigration(t *testing.T) {
	root := t.TempDir()
	writeManifestFile(t, root, "migrations/000001_init.up.sql", []byte("create table smoke(id integer);\n"))
	writeManifestFile(t, root, "majucau.exe", []byte("binary"))
	manifest := ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "2.0",
		Migrations: []ManifestFile{{
			Path:   "migrations/000001_init.up.sql",
			SHA256: digestFor([]byte("create table smoke(id integer);\n")),
		}},
		Artifacts: []ManifestFile{{
			Path:      "majucau.exe",
			SHA256:    digestFor([]byte("binary")),
			SizeBytes: int64(len("binary")),
		}},
	}
	writeManifest(t, root, manifest)

	result := ValidateReleaseManifest(root, "", "1.5")
	if !result.Valid {
		t.Fatalf("expected valid package, issues: %+v", result.Issues)
	}
}

func TestValidateReleaseManifestRejectsChecksumMismatch(t *testing.T) {
	root := t.TempDir()
	writeManifestFile(t, root, "majucau.exe", []byte("actual"))
	writeManifest(t, root, ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "1.0",
		Artifacts:       []ManifestFile{{Path: "majucau.exe", SHA256: digestFor([]byte("expected"))}},
	})

	result := ValidateReleaseManifest(root, "release-manifest.json", "1.0")
	if result.Valid || !hasManifestIssue(result, ManifestIssueChecksumMismatch) {
		t.Fatalf("expected checksum mismatch, got valid=%v issues=%+v", result.Valid, result.Issues)
	}
}

func TestValidateReleaseManifestRejectsIncompatibleSchema(t *testing.T) {
	root := t.TempDir()
	writeManifestFile(t, root, "majucau.exe", []byte("binary"))
	writeManifest(t, root, ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "2.0",
		MinSchema:       "2.0",
		MaxSchema:       "3.0",
		Artifacts:       []ManifestFile{{Path: "majucau.exe", SHA256: digestFor([]byte("binary"))}},
	})

	result := ValidateReleaseManifest(root, "release-manifest.json", "1.0")
	if result.Valid || !hasManifestIssue(result, ManifestIssueSchemaIncompatible) {
		t.Fatalf("expected schema incompatibility, got valid=%v issues=%+v", result.Valid, result.Issues)
	}
}

func TestValidateReleaseManifestRejectsUnsafeDuplicateAndProtectedEntries(t *testing.T) {
	root := t.TempDir()
	writeManifestFile(t, root, "majucau.exe", []byte("binary"))
	writeManifest(t, root, ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "1.0",
		Artifacts: []ManifestFile{
			{Path: "majucau.exe", SHA256: digestFor([]byte("binary"))},
			{Path: "MAJUCAU.EXE", SHA256: digestFor([]byte("binary"))},
			{Path: "../outside.bin", SHA256: digestFor([]byte("binary"))},
			{Path: "secret/token.json", SHA256: digestFor([]byte("binary"))},
		},
	})

	result := ValidateReleaseManifest(root, "release-manifest.json", "1.0")
	if result.Valid || !hasManifestIssue(result, ManifestIssuePackageInvalid) {
		t.Fatalf("expected invalid package, got valid=%v issues=%+v", result.Valid, result.Issues)
	}
	if len(result.Issues) < 3 {
		t.Fatalf("expected duplicate, unsafe and protected entry issues, got %+v", result.Issues)
	}
}

func TestValidateReleaseManifestRejectsTooManyEntries(t *testing.T) {
	root := t.TempDir()
	entries := make([]ManifestFile, maxReleaseManifestEntries+1)
	for index := range entries {
		entries[index] = ManifestFile{
			Path:   fmt.Sprintf("artifact-%04d.exe", index),
			SHA256: digestFor([]byte("binary")),
		}
	}
	writeManifest(t, root, ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "1.0",
		Artifacts:       entries,
	})

	result := ValidateReleaseManifest(root, "release-manifest.json", "1.0")
	if result.Valid || len(result.Issues) != 1 || result.Issues[0].Detail != "manifest contains too many files" {
		t.Fatalf("expected one bounded entry-count issue, got valid=%v issues=%+v", result.Valid, result.Issues)
	}
}

func TestParseSchemaVersionRejectsIntegerOverflow(t *testing.T) {
	if _, ok := parseSchemaVersion(strings.Repeat("9", maxManifestVersionFieldBytes)); ok {
		t.Fatal("schema version with an overflowing integer component was accepted")
	}
}

func TestValidateReleaseManifestRejectsSymlinkedDirectory(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	payload := []byte("outside package")
	writeManifestFile(t, outside, "artifact.bin", payload)
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("cannot create directory symlink on this runner: %v", err)
	}
	writeManifest(t, root, ReleaseManifest{
		ManifestVersion: ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "1.0",
		Artifacts: []ManifestFile{{
			Path:      "linked/artifact.bin",
			SHA256:    digestFor(payload),
			SizeBytes: int64(len(payload)),
		}},
	})

	result := ValidateReleaseManifest(root, "release-manifest.json", "1.0")
	if result.Valid || !hasManifestIssue(result, ManifestIssueChecksumMismatch) {
		t.Fatalf("expected linked directory rejection, got valid=%v issues=%+v", result.Valid, result.Issues)
	}
}

func writeManifestFile(t *testing.T, root, relative string, payload []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeManifest(t *testing.T, root string, manifest ReleaseManifest) {
	t.Helper()
	payload, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	writeManifestFile(t, root, "release-manifest.json", payload)
}

func digestFor(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func hasManifestIssue(result ManifestValidation, code string) bool {
	for _, issue := range result.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
