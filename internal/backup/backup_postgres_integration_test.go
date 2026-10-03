package backup

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestCreateAndVerifyAgainstDisposablePostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("MAJUCAU_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("MAJUCAU_TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	result, err := Create(ctx, Options{
		DatabaseURL:   databaseURL,
		OutputDir:     t.TempDir(),
		AppVersion:    "ci",
		SchemaVersion: "ci-schema",
		InstallID:     "ci-backup-integration",
		Passphrase:    []byte("CI-only backup validation passphrase"),
	})
	if err != nil {
		t.Fatalf("create encrypted backup from disposable PostgreSQL: %v", err)
	}
	if result.Manifest.FormatVersion != FormatVersion || result.Manifest.Cipher != streamCipher {
		t.Fatalf("backup did not use the authenticated streaming format: %#v", result.Manifest)
	}

	entries := make(map[string]int64, len(result.Manifest.Files))
	for _, file := range result.Manifest.Files {
		entries[file.Path] = file.SizeBytes
	}
	for _, path := range []string{databaseDumpEntry, globalsDumpEntry} {
		if entries[path] <= 0 {
			t.Fatalf("backup omitted non-empty PostgreSQL payload %q: %#v", path, result.Manifest.Files)
		}
	}

	verification := VerifyContext(ctx, result.Path, []byte("CI-only backup validation passphrase"), "ci-schema")
	if !verification.Valid || len(verification.Issues) != 0 {
		t.Fatalf("real PostgreSQL backup did not verify: %#v", verification)
	}
}
