package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateAndVerifyEncryptedPackage(t *testing.T) {
	root := t.TempDir()
	passphrase := []byte("uma-frase-local-forte")
	var calls []string
	runner := func(_ context.Context, name string, args []string, env []string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		for _, entry := range env {
			if strings.HasPrefix(entry, "MAJUCAU_DATABASE_URL=") {
				t.Fatalf("DSN should not be passed to PostgreSQL child environment")
			}
		}
		return writeFakeDump(name, args)
	}
	result, err := Create(context.Background(), Options{
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		OutputDir: root, AppVersion: "0.1.0", SchemaVersion: "1.0", InstallID: "install-1",
		Passphrase: passphrase, RunCommand: runner,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if len(calls) != 2 || strings.Contains(calls[0], "user:password") || strings.Contains(calls[1], "user:password") || strings.Contains(strings.Join(calls, " "), "MAJUCAU_DATABASE_URL") {
		t.Fatalf("PostgreSQL command arguments leaked a password: %#v", calls)
	}
	if _, err := os.Stat(result.Path); err != nil {
		t.Fatalf("backup package missing: %v", err)
	}
	verification := Verify(result.Path, passphrase, "1.0")
	if !verification.Valid || len(verification.Issues) != 0 {
		t.Fatalf("Verify() invalid: %#v", verification)
	}
	if verification.Manifest.TokensExported || verification.Manifest.TokenPolicy != TokenExportPolicy {
		t.Fatalf("backup token policy is unsafe: %#v", verification.Manifest)
	}
	if filepath.Ext(result.Path) != ".mjbk" {
		t.Fatalf("unexpected package extension: %s", result.Path)
	}
}

func TestVerifyRejectsWrongPassphraseAndSchema(t *testing.T) {
	root := t.TempDir()
	result, err := Create(context.Background(), Options{
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		OutputDir: root, AppVersion: "0.1.0", SchemaVersion: "1.0", InstallID: "install-1",
		Passphrase: []byte("uma-frase-local-forte"), RunCommand: fakeDumpRunner,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	wrong := Verify(result.Path, []byte("outra-frase-local"), "1.0")
	if wrong.Valid || !containsIssue(wrong.Issues, ErrWrongPassphrase.Error()) {
		t.Fatalf("wrong passphrase was accepted: %#v", wrong)
	}
	schema := Verify(result.Path, []byte("uma-frase-local-forte"), "2.0")
	if schema.Valid || !containsIssue(schema.Issues, ErrUnsupportedSchema.Error()) {
		t.Fatalf("incompatible schema was accepted: %#v", schema)
	}
}

func TestCreateRequiresStrongPassphraseAndAbsoluteOutput(t *testing.T) {
	options := Options{DatabaseURL: "postgres://user@127.0.0.1/majucau", OutputDir: "backups", AppVersion: "0.1.0", SchemaVersion: "1.0", Passphrase: []byte("short")}
	if _, err := Create(context.Background(), options); err != ErrInvalidOptions {
		t.Fatalf("Create() error = %v, want ErrInvalidOptions", err)
	}
}

func TestRestoreCreatesPreRestoreBackupAndRequiresLifecycleHooks(t *testing.T) {
	root := t.TempDir()
	backupDir := filepath.Join(root, "pre-restore")
	passphrase := []byte("uma-frase-local-forte")
	var calls []string
	runner := func(_ context.Context, name string, args []string, _ []string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return writeFakeDump(name, args)
	}
	created, err := Create(context.Background(), Options{
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable", OutputDir: root,
		AppVersion: "0.1.0", SchemaVersion: "1.0", InstallID: "install-1", Passphrase: passphrase, RunCommand: runner,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	events := []string{}
	result, err := Restore(context.Background(), RestoreOptions{
		PackagePath: created.Path, DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable", BackupDir: backupDir,
		CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1", Passphrase: passphrase, RunCommand: runner,
		StopWorker: func(context.Context) error { events = append(events, "stop"); return nil },
		Validate: func(context.Context) error { events = append(events, "validate"); return nil },
		StartWorker: func(context.Context) error { events = append(events, "start"); return nil },
	})
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if result.Status != "RESTORED_NEEDS_RECONNECT" || result.Manifest.SchemaVersion != "1.0" {
		t.Fatalf("unexpected restore result: %#v", result)
	}
	if _, err := os.Stat(result.PreRestoreBackup.Path); err != nil {
		t.Fatalf("pre-restore backup missing: %v", err)
	}
	if strings.Join(events, ",") != "stop,validate,start" {
		t.Fatalf("unexpected lifecycle order: %#v", events)
	}
	if len(calls) != 6 || !strings.Contains(calls[4], "--clean") || !strings.Contains(calls[4], "--exit-on-error") || !strings.Contains(calls[5], "--set=ON_ERROR_STOP=1") {
		t.Fatalf("restore commands were not guarded: %#v", calls)
	}
	if strings.Contains(strings.Join(calls, " "), "user:password") {
		t.Fatalf("restore command arguments leaked the password: %#v", calls)
	}
}

func TestRestoreRequiresLifecycleHooks(t *testing.T) {
	if _, err := Restore(context.Background(), RestoreOptions{PackagePath: "C:\\backup.mjbk", BackupDir: "C:\\backups", DatabaseURL: "postgres://user@127.0.0.1/majucau", CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1", Passphrase: []byte("uma-frase-local-forte")}); err != ErrRestoreHooksRequired {
		t.Fatalf("Restore() error = %v, want ErrRestoreHooksRequired", err)
	}
}

func fakeDumpRunner(_ context.Context, name string, args []string, _ []string) error {
	return writeFakeDump(name, args)
}

func writeFakeDump(name string, args []string) error {
	for index, arg := range args {
		if arg == "--file" && index+1 < len(args) {
			payload := []byte("database dump")
			if strings.Contains(name, "pg_dumpall") {
				payload = []byte("CREATE ROLE majucau_runtime;\n")
			}
			if err := os.WriteFile(args[index+1], payload, 0600); err != nil {
				return err
			}
		}
	}
	return nil
}

func containsIssue(issues []string, expected string) bool {
	for _, issue := range issues {
		if issue == expected {
			return true
		}
	}
	return false
}
