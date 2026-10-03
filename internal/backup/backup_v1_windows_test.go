//go:build windows

package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This fixture writer intentionally reproduces the V1 envelope written by
// earlier releases. It exists only in the Windows test build because restore
// of that format uses the Windows CNG AES-GCM implementation.
func TestRestoreLegacyV1PackageThroughWindowsCNG(t *testing.T) {
	root := t.TempDir()
	packagePath := filepath.Join(root, "legacy-v1.mjbk")
	backupDir := filepath.Join(root, "pre-restore")
	passphrase := []byte("uma-frase-local-forte")
	databaseDump := bytes.Repeat([]byte("historical-database-row\n"), 16_000)
	globalsDump := bytes.Repeat([]byte("CREATE ROLE historical_role;\n"), 5_000)
	writeV1TestPackage(t, packagePath, passphrase, databaseDump, globalsDump, false)

	var events []string
	var commands []string
	runner := func(_ context.Context, name string, args []string, _ []string) error {
		commands = append(commands, name)
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "pg_restore") {
			if len(args) == 0 {
				t.Fatal("pg_restore did not receive the staged database dump")
			}
			actual, err := os.ReadFile(args[len(args)-1])
			if err != nil {
				return err
			}
			if !bytes.Equal(actual, databaseDump) {
				t.Fatal("V1 database dump did not survive multi-block restore")
			}
			return nil
		}
		if strings.Contains(lowerName, "psql") {
			for index, arg := range args {
				if arg == "--file" && index+1 < len(args) {
					actual, err := os.ReadFile(args[index+1])
					if err != nil {
						return err
					}
					if !bytes.Equal(actual, globalsDump) {
						t.Fatal("V1 globals dump did not survive multi-block restore")
					}
					return nil
				}
			}
			t.Fatal("psql did not receive the staged globals dump")
			return nil
		}
		return writeFakeDump(name, args)
	}

	result, err := Restore(context.Background(), RestoreOptions{
		PackagePath: packagePath,
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		BackupDir: backupDir, CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1",
		Passphrase: passphrase,
		PGDumpPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_dump.exe",
		PGDumpAllPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_dumpall.exe",
		PGRestorePath: "C:\\Program Files\\PostgreSQL\\16\\bin\\pg_restore.exe",
		PSQLPath: "C:\\Program Files\\PostgreSQL\\16\\bin\\psql.exe",
		RunCommand: runner,
		StopWorker: func(context.Context) error { events = append(events, "stop"); return nil },
		Validate: func(context.Context) error { events = append(events, "validate"); return nil },
		StartWorker: func(context.Context) error { events = append(events, "start"); return nil },
	})
	if err != nil {
		t.Fatalf("Restore() of historical V1 package failed: %v", err)
	}
	if result.Manifest.FormatVersion != legacyFormatVersion || result.Status != "RESTORED_NEEDS_RECONNECT" {
		t.Fatalf("unexpected V1 restore result: %#v", result)
	}
	if strings.Join(events, ",") != "stop,validate,start" {
		t.Fatalf("unexpected worker lifecycle: %#v", events)
	}
	if len(commands) != 4 {
		t.Fatalf("expected pre-restore backup and restore commands, got %#v", commands)
	}
}

func TestRestoreRejectsTamperedLegacyV1BeforeStoppingWorker(t *testing.T) {
	root := t.TempDir()
	packagePath := filepath.Join(root, "legacy-v1-tampered.mjbk")
	passphrase := []byte("uma-frase-local-forte")
	databaseDump := bytes.Repeat([]byte("historical-database-row\n"), 4_000)
	globalsDump := bytes.Repeat([]byte("CREATE ROLE historical_role;\n"), 1_000)
	writeV1TestPackage(t, packagePath, passphrase, databaseDump, globalsDump, true)

	stopped := false
	result, err := Restore(context.Background(), RestoreOptions{
		PackagePath: packagePath,
		DatabaseURL: "postgres://user:password@127.0.0.1:54329/majucau?sslmode=disable",
		BackupDir: filepath.Join(root, "pre-restore"), CurrentSchema: "1.0", AppVersion: "0.1.0", InstallID: "install-1",
		Passphrase: passphrase, RunCommand: fakeDumpRunner,
		StopWorker: func(context.Context) error { stopped = true; return nil },
		Validate: func(context.Context) error { return nil },
		StartWorker: func(context.Context) error { return nil },
	})
	if err == nil || result.Status != "RECOVERY_REQUIRED" {
		t.Fatalf("tampered V1 package was not rejected: result=%#v err=%v", result, err)
	}
	if stopped {
		t.Fatal("restore stopped the worker before authenticating the V1 package")
	}
}

func writeV1TestPackage(t *testing.T, path string, passphrase, databaseDump, globalsDump []byte, tamperDatabaseTag bool) {
	t.Helper()
	salt := make([]byte, argonSaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		t.Fatalf("generate V1 fixture salt: %v", err)
	}
	parameters := KDFParameters{
		Name: "argon2id", Salt: hex.EncodeToString(salt), MemoryKB: argonMemoryKB,
		Iterations: argonIterations, Parallelism: argonParallelism, KeyLength: argonKeyLength,
	}
	key := deriveKey(passphrase, salt, parameters)
	defer clearBytes(key)
	databaseCiphertext := sealV1TestPayload(t, key, []byte(legacyDatabaseDumpEntry), databaseDump)
	if tamperDatabaseTag {
		// Keep the manifest digest correct so rejection proves AES-GCM
		// authentication, rather than only the outer ZIP-entry checksum.
		databaseCiphertext[len(databaseCiphertext)-1] ^= 0x80
	}
	globalsCiphertext := sealV1TestPayload(t, key, []byte(legacyGlobalsDumpEntry), globalsDump)
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	manifest := Manifest{
		FormatVersion: legacyFormatVersion, ManifestSchema: legacyManifestSchemaVersion,
		AppVersion: "0.1.0", SchemaVersion: "1.0", InstallID: "install-1",
		Source: BackupSource, CreatedAt: now, TokensExported: false, TokenPolicy: TokenExportPolicy,
		KDF: parameters,
		Files: []File{
			fileRecord(legacyDatabaseDumpEntry, databaseCiphertext),
			fileRecord(legacyGlobalsDumpEntry, globalsCiphertext),
		},
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode V1 fixture manifest: %v", err)
	}
	manifestBytes = append(manifestBytes, '\n')

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatalf("create V1 fixture: %v", err)
	}
	archive := zip.NewWriter(file)
	entries := map[string][]byte{
		manifestEntry: manifestBytes,
		legacyDatabaseDumpEntry: databaseCiphertext,
		legacyGlobalsDumpEntry: globalsCiphertext,
	}
	for _, name := range []string{legacyDatabaseDumpEntry, legacyGlobalsDumpEntry, manifestEntry} {
		entry, createErr := archive.Create(name) // V1 used ZIP's default Deflate method.
		if createErr != nil {
			_ = archive.Close()
			_ = file.Close()
			t.Fatalf("create V1 fixture entry %q: %v", name, createErr)
		}
		if _, writeErr := entry.Write(entries[name]); writeErr != nil {
			_ = archive.Close()
			_ = file.Close()
			t.Fatalf("write V1 fixture entry %q: %v", name, writeErr)
		}
	}
	if err := archive.Close(); err != nil {
		_ = file.Close()
		t.Fatalf("finish V1 fixture ZIP: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close V1 fixture: %v", err)
	}
}

func sealV1TestPayload(t *testing.T, key, associatedData, plaintext []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("create V1 fixture AES key: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("create V1 fixture GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("generate V1 fixture nonce: %v", err)
	}
	envelope := make([]byte, 0, len(encryptedHeader)+1+len(nonce)+len(plaintext)+gcm.Overhead())
	envelope = append(envelope, encryptedHeader...)
	envelope = append(envelope, byte(len(nonce)))
	envelope = append(envelope, nonce...)
	envelope = gcm.Seal(envelope, nonce, plaintext, associatedData)
	return envelope
}
