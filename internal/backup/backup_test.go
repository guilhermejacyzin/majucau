package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	if !strings.Contains(calls[1], "--database majucau") {
		t.Fatalf("pg_dumpall must connect through the configured database for scoped password authentication: %#v", calls[1])
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

func TestV2StreamingBackupAuthenticatesAndRestoresLargeSegmentedPayload(t *testing.T) {
	root := t.TempDir()
	passphrase := []byte("uma-frase-local-forte")
	largeDump := bytes.Repeat([]byte("D"), 3<<20)
	runner := func(_ context.Context, name string, args []string, _ []string) error {
		for index, arg := range args {
			if arg != "--file" || index+1 >= len(args) {
				continue
			}
			payload := largeDump
			if strings.Contains(name, "pg_dumpall") {
				payload = []byte("CREATE ROLE majucau_runtime;\n")
			}
			return os.WriteFile(args[index+1], payload, 0600)
		}
		t.Fatalf("PostgreSQL command did not specify an output file: %q", name)
		return nil
	}

	created, err := Create(context.Background(), Options{
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		OutputDir: root, AppVersion: "0.1.0", SchemaVersion: "1.0", InstallID: "install-1",
		Passphrase: passphrase, RunCommand: runner,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Manifest.FormatVersion != FormatVersion || created.Manifest.Cipher != streamCipher {
		t.Fatalf("Create() did not use the streaming format: %#v", created.Manifest)
	}
	var databaseRecord File
	for _, record := range created.Manifest.Files {
		if record.Path == databaseDumpEntry {
			databaseRecord = record
		}
	}
	if databaseRecord.Path == "" || databaseRecord.SizeBytes <= int64(len(largeDump)) {
		t.Fatalf("database payload did not cross streaming segment boundaries: %#v", databaseRecord)
	}
	if verification := Verify(created.Path, passphrase, "1.0"); !verification.Valid {
		t.Fatalf("Verify() rejected valid multi-segment V2 package: %#v", verification)
	}

	var restoreEvents []string
	databaseDumpRestored := false
	globalsDumpRestored := false
	restoreRunner := func(_ context.Context, name string, args []string, _ []string) error {
		switch {
		case strings.Contains(name, "pg_restore"):
			if len(args) == 0 {
				t.Error("pg_restore did not receive the staged database dump")
				return nil
			}
			payload, err := os.ReadFile(args[len(args)-1])
			if err != nil {
				t.Errorf("read staged database dump: %v", err)
				return nil
			}
			databaseDumpRestored = bytes.Equal(payload, largeDump)
		case strings.Contains(name, "psql"):
			for index, arg := range args {
				if arg == "--file" && index+1 < len(args) {
					payload, err := os.ReadFile(args[index+1])
					if err != nil {
						t.Errorf("read staged globals dump: %v", err)
						return nil
					}
					globalsDumpRestored = bytes.Equal(payload, []byte("CREATE ROLE majucau_runtime;\n"))
					break
				}
			}
		case strings.Contains(name, "pg_dump"):
			return writeFakeDump(name, args)
		}
		return nil
	}
	restored, err := Restore(context.Background(), RestoreOptions{
		PackagePath: created.Path, DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		BackupDir: filepath.Join(root, "pre-restore"), CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1",
		Passphrase: passphrase, RunCommand: restoreRunner,
		StopWorker: func(context.Context) error { restoreEvents = append(restoreEvents, "stop"); return nil },
		Validate:   func(context.Context) error { restoreEvents = append(restoreEvents, "validate"); return nil },
		StartWorker: func(context.Context) error {
			restoreEvents = append(restoreEvents, "start")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Restore() rejected valid multi-segment V2 package: %v", err)
	}
	if restored.Manifest.FormatVersion != FormatVersion || restored.Status != "RESTORED_NEEDS_RECONNECT" || !databaseDumpRestored || !globalsDumpRestored {
		t.Fatalf("V2 restore did not preserve both payloads: result=%#v database=%t globals=%t", restored, databaseDumpRestored, globalsDumpRestored)
	}
	if strings.Join(restoreEvents, ",") != "stop,validate,start" {
		t.Fatalf("unexpected V2 restore lifecycle: %#v", restoreEvents)
	}

	tamperedPath := filepath.Join(root, "tampered.mjbk")
	rewriteTamperedV2Backup(t, created.Path, tamperedPath, databaseDumpEntry)
	tampered := Verify(tamperedPath, passphrase, "1.0")
	if tampered.Valid || !containsIssue(tampered.Issues, "backup payload authentication or checksum validation failed") {
		t.Fatalf("Verify() accepted a payload tampered after updating the manifest hash: %#v", tampered)
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
		CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1", Passphrase: passphrase,
		PGDumpPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_dump.exe",
		PGDumpAllPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_dumpall.exe",
		PGRestorePath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_restore.exe",
		PSQLPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\psql.exe", RunCommand: runner,
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
	if len(calls) != 6 || !strings.Contains(calls[2], "pg_dump.exe") || !strings.Contains(calls[3], "pg_dumpall.exe") || !strings.Contains(calls[4], "pg_restore.exe") || !strings.Contains(calls[4], "--clean") || !strings.Contains(calls[4], "--exit-on-error") || !strings.Contains(calls[5], "psql.exe") || !strings.Contains(calls[5], "--set=ON_ERROR_STOP=1") {
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

// rewriteTamperedV2Backup changes ciphertext and updates the manifest digest,
// proving verification relies on authenticated encryption and not only SHA-256.
func rewriteTamperedV2Backup(t *testing.T, sourcePath, destinationPath, payloadName string) {
	t.Helper()
	archive, err := zip.OpenReader(sourcePath)
	if err != nil {
		t.Fatalf("open backup package: %v", err)
	}
	defer archive.Close()

	entries := make(map[string][]byte, len(archive.File))
	var manifest Manifest
	for _, entry := range archive.File {
		reader, err := entry.Open()
		if err != nil {
			t.Fatalf("open backup entry %q: %v", entry.Name, err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			t.Fatalf("read backup entry %q: %v", entry.Name, readErr)
		}
		if closeErr != nil {
			t.Fatalf("close backup entry %q: %v", entry.Name, closeErr)
		}
		entries[entry.Name] = data
		if entry.Name == manifestEntry {
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatalf("decode backup manifest: %v", err)
			}
		}
	}

	ciphertext, ok := entries[payloadName]
	if !ok || len(ciphertext) < 2 {
		t.Fatalf("backup payload %q is missing or too small", payloadName)
	}
	ciphertext[len(ciphertext)*3/4] ^= 0x01
	entries[payloadName] = ciphertext
	digest := sha256.Sum256(ciphertext)
	updated := false
	for index := range manifest.Files {
		if manifest.Files[index].Path == payloadName {
			manifest.Files[index].SHA256 = hex.EncodeToString(digest[:])
			manifest.Files[index].SizeBytes = int64(len(ciphertext))
			updated = true
		}
	}
	if !updated {
		t.Fatalf("manifest has no record for payload %q", payloadName)
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode tampered manifest: %v", err)
	}
	entries[manifestEntry] = append(manifestBytes, '\n')

	output, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatalf("create tampered package: %v", err)
	}
	writer := zip.NewWriter(output)
	for _, entry := range archive.File {
		header := &zip.FileHeader{Name: entry.Name, Method: zip.Store}
		header.SetMode(0600)
		entryWriter, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatalf("create tampered package entry %q: %v", entry.Name, err)
		}
		if _, err := entryWriter.Write(entries[entry.Name]); err != nil {
			t.Fatalf("write tampered package entry %q: %v", entry.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("finish tampered package: %v", err)
	}
	if err := output.Close(); err != nil {
		t.Fatalf("close tampered package: %v", err)
	}
}

func containsIssue(issues []string, expected string) bool {
	for _, issue := range issues {
		if issue == expected {
			return true
		}
	}
	return false
}

