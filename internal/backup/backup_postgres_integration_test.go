package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestCreateVerifyAndRestoreAgainstDisposablePostgres(t *testing.T) {
	sourceURL := strings.TrimSpace(os.Getenv("MAJUCAU_TEST_DATABASE_URL"))
	targetURL := strings.TrimSpace(os.Getenv("MAJUCAU_TEST_RESTORE_DATABASE_URL"))
	if sourceURL == "" || targetURL == "" {
		t.Skip("source and restore disposable PostgreSQL URLs are not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	const marker = "streaming-backup-restore-ci"
	source, err := pgx.Connect(ctx, sourceURL)
	if err != nil {
		t.Fatalf("connect to disposable backup source PostgreSQL: %v", err)
	}
	defer func() { _ = source.Close(context.Background()) }()
	if _, err := source.Exec(ctx, `CREATE TABLE backup_restore_evidence (marker text PRIMARY KEY)`); err != nil {
		t.Fatalf("create restore evidence table in source: %v", err)
	}
	if _, err := source.Exec(ctx, `INSERT INTO backup_restore_evidence (marker) VALUES ($1)`, marker); err != nil {
		t.Fatalf("insert restore evidence row in source: %v", err)
	}

	passphrase := []byte("CI-only backup validation passphrase")
	outputDir := filepath.Join(t.TempDir(), "source-backups")
	created, err := Create(ctx, Options{
		DatabaseURL: sourceURL, OutputDir: outputDir,
		AppVersion: "ci", SchemaVersion: "ci-schema",
		InstallID: "ci-backup-integration", Passphrase: passphrase,
	})
	if err != nil {
		t.Fatalf("create encrypted backup from disposable PostgreSQL: %v", err)
	}
	if created.Manifest.FormatVersion != FormatVersion || created.Manifest.Cipher != streamCipher {
		t.Fatalf("backup did not use the authenticated streaming format: %#v", created.Manifest)
	}

	entries := make(map[string]int64, len(created.Manifest.Files))
	for _, file := range created.Manifest.Files {
		entries[file.Path] = file.SizeBytes
	}
	for _, path := range []string{databaseDumpEntry, globalsDumpEntry} {
		if entries[path] <= 0 {
			t.Fatalf("backup omitted non-empty PostgreSQL payload %q: %#v", path, created.Manifest.Files)
		}
	}

	verification := VerifyContext(ctx, created.Path, passphrase, "ci-schema")
	if !verification.Valid || len(verification.Issues) != 0 {
		t.Fatalf("real PostgreSQL backup did not verify: %#v", verification)
	}

	targetBackupDir := filepath.Join(t.TempDir(), "restore-backups")
	events := make([]string, 0, 3)
	restored, err := Restore(ctx, RestoreOptions{
		PackagePath: created.Path, DatabaseURL: targetURL, BackupDir: targetBackupDir,
		CurrentSchema: "ci-schema", AppVersion: "ci", InstallID: "ci-restore-integration",
		Passphrase: passphrase,
		StopWorker: func(context.Context) error {
			events = append(events, "stop")
			return nil
		},
		Validate: func(ctx context.Context) error {
			events = append(events, "validate")
			connection, err := pgx.Connect(ctx, targetURL)
			if err != nil {
				return err
			}
			defer func() { _ = connection.Close(context.Background()) }()

			var restoredMarker string
			if err := connection.QueryRow(ctx, `SELECT marker FROM backup_restore_evidence`).Scan(&restoredMarker); err != nil {
				return err
			}
			if restoredMarker != marker {
				return fmt.Errorf("restored marker mismatch: got %q", restoredMarker)
			}
			var sourceRoleExists bool
			if err := connection.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'majucau')`).Scan(&sourceRoleExists); err != nil {
				return err
			}
			if !sourceRoleExists {
				return fmt.Errorf("source PostgreSQL role was not restored")
			}
			return nil
		},
		StartWorker: func(context.Context) error {
			events = append(events, "start")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("restore encrypted backup into fresh PostgreSQL: result=%#v err=%v", restored, err)
	}
	if restored.Status != "RESTORED_NEEDS_RECONNECT" || restored.Manifest.InstallID != created.Manifest.InstallID {
		t.Fatalf("unexpected real PostgreSQL restore result: %#v", restored)
	}
	if restored.PreRestoreBackup.Path == "" {
		t.Fatal("restore did not create a pre-restore backup of the target database")
	}
	preRestore := VerifyContext(ctx, restored.PreRestoreBackup.Path, passphrase, "ci-schema")
	if !preRestore.Valid || len(preRestore.Issues) != 0 {
		t.Fatalf("pre-restore backup did not verify: %#v", preRestore)
	}
	if !reflect.DeepEqual(events, []string{"stop", "validate", "start"}) {
		t.Fatalf("unexpected worker lifecycle around PostgreSQL restore: %#v", events)
	}
}
