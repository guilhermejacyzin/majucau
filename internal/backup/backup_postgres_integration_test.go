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
	_, err = source.Exec(ctx, `
WITH connection AS (
  INSERT INTO integration_connections (provider, status)
  VALUES ('BLING', 'CONNECTED')
  RETURNING id
), batch AS (
  INSERT INTO integration_sync_batches (connection_id, resource, status, finished_at)
  SELECT id, 'ci_restore_evidence', 'SUCCESS', clock_timestamp() FROM connection
  RETURNING id, connection_id
), raw AS (
  INSERT INTO raw_records (sync_batch_id, connection_id, source_system, source_entity, source_id, payload_hash, payload)
  SELECT id, connection_id, 'BLING', 'ci_restore_evidence', 'streaming-backup-restore-ci-raw', repeat('a', 64),
         '{"marker":"streaming-backup-restore-ci-raw"}'::jsonb
  FROM batch
  RETURNING id
), cursor AS (
  INSERT INTO sync_cursors (connection_id, resource, cursor_value)
  SELECT connection_id, 'ci_restore_evidence', 'marker-page-1' FROM batch
), rule AS (
  INSERT INTO financial_rule_versions (rule_code, version, description, status)
  VALUES ('CI_RESTORE_EVIDENCE', 1, 'Test-only backup restoration fixture', 'DRAFT')
  RETURNING id
), calculation AS (
  INSERT INTO calculation_runs (calculation_type, rule_version_id, scenario_id, reference_date, status, input_hash, finished_at)
  SELECT 'CI_RESTORE_EVIDENCE', rule.id, scenario.id, DATE '2026-10-10', 'PARTIAL', repeat('b', 64), clock_timestamp()
  FROM rule CROSS JOIN scenarios AS scenario
  WHERE scenario.code = 'BASE'
  RETURNING id, rule_version_id, scenario_id
), snapshot AS (
  INSERT INTO statement_snapshots (calculation_run_id, statement_type, period_start, period_end, scenario_id, status)
  SELECT id, 'DRE', DATE '2026-10-01', DATE '2026-10-31', scenario_id, 'PROVISIONAL'
  FROM calculation
  RETURNING id, calculation_run_id
)
INSERT INTO statement_lines (snapshot_id, line_code, label, amount, display_order, rule_version_id)
SELECT snapshot.id, 'CI_RESTORE_EVIDENCE', 'CI-only backup restore marker', 123.45, 999, calculation.rule_version_id
FROM snapshot JOIN calculation ON calculation.id = snapshot.calculation_run_id`)
	if err != nil {
		t.Fatalf("create RAW, cursor, and statement snapshot evidence in source: %v", err)
	}
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
			var restoredRawMarker string
			if err := connection.QueryRow(ctx, `SELECT payload->>'marker' FROM raw_records WHERE source_entity = 'ci_restore_evidence' AND source_id = 'streaming-backup-restore-ci-raw'`).Scan(&restoredRawMarker); err != nil {
				return fmt.Errorf("query restored RAW record: %w", err)
			}
			if restoredRawMarker != "streaming-backup-restore-ci-raw" {
				return fmt.Errorf("restored RAW marker mismatch: got %q", restoredRawMarker)
			}
			var restoredCursor string
			if err := connection.QueryRow(ctx, `SELECT cursor_value FROM sync_cursors WHERE resource = 'ci_restore_evidence'`).Scan(&restoredCursor); err != nil {
				return fmt.Errorf("query restored sync cursor: %w", err)
			}
			if restoredCursor != "marker-page-1" {
				return fmt.Errorf("restored cursor mismatch: got %q", restoredCursor)
			}
			var restoredStatementAmount string
			if err := connection.QueryRow(ctx, `SELECT line.amount::text FROM statement_lines AS line JOIN statement_snapshots AS snapshot ON snapshot.id = line.snapshot_id WHERE line.line_code = 'CI_RESTORE_EVIDENCE' AND snapshot.statement_type = 'DRE'`).Scan(&restoredStatementAmount); err != nil {
				return fmt.Errorf("query restored statement snapshot line: %w", err)
			}
			if restoredStatementAmount != "123.4500" {
				return fmt.Errorf("restored statement amount mismatch: got %q", restoredStatementAmount)
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
