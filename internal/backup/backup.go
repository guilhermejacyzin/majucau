// Package backup creates and verifies encrypted, portable PostgreSQL backups.
// It deliberately does not restore or mutate a live cluster; restoration is a
// separate, explicitly authorized operation with compatibility checks.
package backup

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"
)

const (
	FormatVersion           = "1"
	ManifestSchemaVersion   = "1.0"
	BackupSource             = "MAJUCAU_LOCAL_POSTGRESQL"
	TokenExportPolicy        = "EXCLUDED"
	databaseDumpEntry        = "payload/database.dump.gcm"
	globalsDumpEntry         = "payload/globals.sql.gcm"
	manifestEntry            = "backup-manifest.json"
	encryptedHeader           = "MAJUCAU-BACKUP-1"
	argonMemoryKB             = 64 * 1024
	argonIterations           = 1
	argonParallelism          = 4
	argonKeyLength            = 32
	argonSaltLength            = 16
)

var (
	ErrInvalidOptions       = errors.New("invalid backup options")
	ErrInvalidPackage       = errors.New("invalid backup package")
	ErrWrongPassphrase      = errors.New("backup passphrase is invalid")
	ErrUnsupportedSchema    = errors.New("backup schema is incompatible")
)

// CommandRunner is injectable so package tests never require PostgreSQL.
// env contains a temporary PGPASSFILE path, never a password or DSN.
type CommandRunner func(ctx context.Context, name string, args []string, env []string) error

type Options struct {
	DatabaseURL string
	OutputDir   string
	AppVersion  string
	SchemaVersion string
	InstallID   string
	Passphrase  []byte
	PGDumpPath  string
	PGDumpAllPath string
	Now         func() time.Time
	RunCommand  CommandRunner
}

type File struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

type Manifest struct {
	FormatVersion     string `json:"format_version"`
	ManifestSchema    string `json:"manifest_schema"`
	AppVersion        string `json:"app_version"`
	SchemaVersion     string `json:"schema_version"`
	InstallID         string `json:"install_id"`
	Source            string `json:"source"`
	CreatedAt         time.Time `json:"created_at"`
	TokensExported    bool `json:"tokens_exported"`
	TokenPolicy       string `json:"token_policy"`
	KDF               KDFParameters `json:"kdf"`
	Files             []File `json:"files"`
}

type KDFParameters struct {
	Name         string `json:"name"`
	Salt         string `json:"salt"`
	MemoryKB     uint32 `json:"memory_kib"`
	Iterations   uint32 `json:"iterations"`
	Parallelism  uint8  `json:"parallelism"`
	KeyLength    uint32 `json:"key_length"`
}

type Result struct {
	Path     string
	Manifest Manifest
}

type Verification struct {
	Valid    bool
	Manifest Manifest
	Issues   []string
}

func Create(ctx context.Context, options Options) (Result, error) {
	if err := validateOptions(options); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(options.OutputDir, 0700); err != nil {
		return Result{}, fmt.Errorf("create backup directory: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	run := options.RunCommand
	if run == nil {
		run = runExternalCommand
	}
	pgDump := options.PGDumpPath
	if pgDump == "" {
		pgDump = "pg_dump"
	}
	pgDumpAll := options.PGDumpAllPath
	if pgDumpAll == "" {
		pgDumpAll = "pg_dumpall"
	}

	config, err := pgx.ParseConfig(options.DatabaseURL)
	if err != nil || config.Host == "" || config.Database == "" || config.User == "" {
		return Result{}, ErrInvalidOptions
	}
	temporary, err := os.MkdirTemp(options.OutputDir, ".backup-")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary backup directory: %w", err)
	}
	defer os.RemoveAll(temporary)
	passwordFile, env, err := connectionEnvironment(config, temporary)
	if err != nil {
		return Result{}, err
	}
	if passwordFile != "" {
		defer os.Remove(passwordFile)
	}
	dumpPath := filepath.Join(temporary, "database.dump")
	globalsPath := filepath.Join(temporary, "globals.sql")
	common := []string{"--host", config.Host, "--port", fmt.Sprintf("%d", config.Port), "--username", config.User, "--dbname", config.Database}
	if err := run(ctx, pgDump, append([]string{"--format=custom", "--no-owner", "--no-privileges", "--file", dumpPath}, common...), env); err != nil {
		return Result{}, fmt.Errorf("pg_dump failed: %w", err)
	}
	if err := run(ctx, pgDumpAll, append([]string{"--globals-only", "--no-role-passwords", "--file", globalsPath}, common[:len(common)-2]...), env); err != nil {
		return Result{}, fmt.Errorf("pg_dumpall failed: %w", err)
	}
	dump, err := os.ReadFile(dumpPath)
	if err != nil {
		return Result{}, fmt.Errorf("read database dump: %w", err)
	}
	globals, err := os.ReadFile(globalsPath)
	if err != nil {
		return Result{}, fmt.Errorf("read globals dump: %w", err)
	}
	salt := make([]byte, argonSaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return Result{}, fmt.Errorf("generate backup salt: %w", err)
	}
	manifest := Manifest{
		FormatVersion: FormatVersion, ManifestSchema: ManifestSchemaVersion,
		AppVersion: options.AppVersion, SchemaVersion: options.SchemaVersion,
		InstallID: options.InstallID, Source: BackupSource, CreatedAt: now().UTC(),
		TokensExported: false, TokenPolicy: TokenExportPolicy,
		KDF: KDFParameters{Name: "argon2id", Salt: hex.EncodeToString(salt), MemoryKB: argonMemoryKB, Iterations: argonIterations, Parallelism: argonParallelism, KeyLength: argonKeyLength},
	}
	key := deriveKey(options.Passphrase, salt, manifest.KDF)
	databaseEncrypted, err := encrypt(key, []byte(databaseDumpEntry), dump)
	if err != nil {
		return Result{}, err
	}
	globalsEncrypted, err := encrypt(key, []byte(globalsDumpEntry), globals)
	if err != nil {
		return Result{}, err
	}
	manifest.Files = []File{fileRecord(databaseDumpEntry, databaseEncrypted), fileRecord(globalsDumpEntry, globalsEncrypted)}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("encode backup manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	stamp := now().UTC().Format("20060102T150405.000000000Z")
	finalPath := filepath.Join(options.OutputDir, "majucau-backup-"+stamp+".mjbk")
	temporaryPackage := finalPath + ".tmp"
	if err := writePackage(temporaryPackage, manifestBytes, databaseEncrypted, globalsEncrypted); err != nil {
		_ = os.Remove(temporaryPackage)
		return Result{}, err
	}
	if err := os.Rename(temporaryPackage, finalPath); err != nil {
		_ = os.Remove(temporaryPackage)
		return Result{}, fmt.Errorf("commit backup package: %w", err)
	}
	return Result{Path: finalPath, Manifest: manifest}, nil
}

func Verify(path string, passphrase []byte, currentSchema string) Verification {
	result := Verification{}
	if !filepath.IsAbs(path) || strings.TrimSpace(currentSchema) == "" || len(passphrase) == 0 {
		result.Issues = append(result.Issues, "invalid verification arguments")
		return result
	}
	reader, err := zip.OpenReader(path)
	if err != nil {
		result.Issues = append(result.Issues, "backup package is unreadable")
		return result
	}
	defer reader.Close()
	entries := make(map[string][]byte, len(reader.File))
	seenEntries := make(map[string]bool, len(reader.File))
	expectedEntries := map[string]bool{manifestEntry: true, databaseDumpEntry: true, globalsDumpEntry: true}
	for _, entry := range reader.File {
		if !expectedEntries[entry.Name] {
			result.Issues = append(result.Issues, "backup package contains an unexpected entry")
			continue
		}
		if seenEntries[entry.Name] {
			result.Issues = append(result.Issues, "backup package contains a duplicate entry")
			continue
		}
		seenEntries[entry.Name] = true
		file, openErr := entry.Open()
		if openErr != nil {
			result.Issues = append(result.Issues, "backup package entry is unreadable")
			continue
		}
		payload, readErr := io.ReadAll(file)
		_ = file.Close()
		if readErr != nil {
			result.Issues = append(result.Issues, "backup package entry is unreadable")
			continue
		}
		entries[entry.Name] = payload
	}
	for expected := range expectedEntries {
		if !seenEntries[expected] {
			result.Issues = append(result.Issues, "backup package entry is missing")
		}
	}
	manifestBytes, ok := entries[manifestEntry]
	if !ok {
		result.Issues = append(result.Issues, "backup manifest is missing")
		return result
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result.Manifest); err != nil {
		result.Issues = append(result.Issues, "backup manifest is invalid")
		return result
	}
	if result.Manifest.FormatVersion != FormatVersion || result.Manifest.ManifestSchema != ManifestSchemaVersion || result.Manifest.Source != BackupSource || result.Manifest.TokensExported || result.Manifest.TokenPolicy != TokenExportPolicy {
		result.Issues = append(result.Issues, "backup manifest identity or token policy is invalid")
	}
	if result.Manifest.SchemaVersion != currentSchema {
		result.Issues = append(result.Issues, ErrUnsupportedSchema.Error())
	}
	if len(result.Manifest.Files) != 2 {
		result.Issues = append(result.Issues, "backup manifest file list is invalid")
	}
	seenManifestFiles := make(map[string]bool, len(result.Manifest.Files))
	for _, expected := range result.Manifest.Files {
		if expected.Path != databaseDumpEntry && expected.Path != globalsDumpEntry {
			result.Issues = append(result.Issues, "backup manifest contains an unexpected file")
			continue
		}
		if seenManifestFiles[expected.Path] {
			result.Issues = append(result.Issues, "backup manifest contains a duplicate file")
			continue
		}
		seenManifestFiles[expected.Path] = true
	}
	if !seenManifestFiles[databaseDumpEntry] || !seenManifestFiles[globalsDumpEntry] {
		result.Issues = append(result.Issues, "backup manifest file list is incomplete")
	}
	if !validKDF(result.Manifest.KDF) {
		result.Issues = append(result.Issues, "backup key derivation parameters are invalid")
		return result
	}
	key := deriveKey(passphrase, mustDecodeSalt(result.Manifest.KDF.Salt), result.Manifest.KDF)
	for _, expected := range result.Manifest.Files {
		payload, found := entries[expected.Path]
		if !found || !validFileDigest(expected, payload) {
			result.Issues = append(result.Issues, "backup checksum mismatch")
			continue
		}
		if _, err := decrypt(key, []byte(expected.Path), payload); err != nil {
			result.Issues = append(result.Issues, ErrWrongPassphrase.Error())
		}
	}
	result.Valid = len(result.Issues) == 0
	return result
}

func validateOptions(options Options) error {
	if strings.TrimSpace(options.DatabaseURL) == "" || strings.TrimSpace(options.OutputDir) == "" || !filepath.IsAbs(options.OutputDir) || strings.TrimSpace(options.AppVersion) == "" || strings.TrimSpace(options.SchemaVersion) == "" || len(options.Passphrase) < 12 {
		return ErrInvalidOptions
	}
	return nil
}

func connectionEnvironment(config *pgx.ConnConfig, directory string) (string, []string, error) {
	env := make([]string, 0, len(os.Environ())+5)
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "MAJUCAU_DATABASE_URL=") {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, "PGHOST="+config.Host, fmt.Sprintf("PGPORT=%d", config.Port), "PGUSER="+config.User, "PGDATABASE="+config.Database)
	if config.Password == "" {
		return "", env, nil
	}
	path := filepath.Join(directory, "pgpass.conf")
	line := fmt.Sprintf("%s:%d:%s:%s:%s\n", config.Host, config.Port, config.Database, config.User, config.Password)
	if err := os.WriteFile(path, []byte(line), 0600); err != nil {
		return "", nil, fmt.Errorf("create temporary PostgreSQL password file: %w", err)
	}
	env = append(env, "PGPASSFILE="+path)
	return path, env, nil
}

func runExternalCommand(ctx context.Context, name string, args []string, env []string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = env
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	return command.Run()
}

func deriveKey(passphrase, salt []byte, parameters KDFParameters) []byte {
	return argon2.IDKey(passphrase, salt, parameters.Iterations, parameters.MemoryKB, parameters.Parallelism, uint32(parameters.KeyLength))
}

func encrypt(key, associatedData, plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plain, associatedData)
	result := make([]byte, 0, len(encryptedHeader)+1+len(nonce)+len(sealed))
	result = append(result, encryptedHeader...)
	result = append(result, byte(len(nonce)))
	result = append(result, nonce...)
	result = append(result, sealed...)
	return result, nil
}

func decrypt(key, associatedData, payload []byte) ([]byte, error) {
	minimum := len(encryptedHeader) + 1
	if len(payload) < minimum || string(payload[:len(encryptedHeader)]) != encryptedHeader {
		return nil, ErrWrongPassphrase
	}
	nonceLength := int(payload[len(encryptedHeader)])
	if nonceLength <= 0 || len(payload) <= minimum+nonceLength {
		return nil, ErrWrongPassphrase
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrWrongPassphrase
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || nonceLength != gcm.NonceSize() {
		return nil, ErrWrongPassphrase
	}
	return gcm.Open(nil, payload[minimum:minimum+nonceLength], payload[minimum+nonceLength:], associatedData)
}

func fileRecord(path string, payload []byte) File {
	digest := sha256.Sum256(payload)
	return File{Path: path, SHA256: hex.EncodeToString(digest[:]), SizeBytes: int64(len(payload))}
}

func validFileDigest(expected File, payload []byte) bool {
	actual := fileRecord(expected.Path, payload)
	return strings.EqualFold(actual.SHA256, expected.SHA256) && actual.SizeBytes == expected.SizeBytes
}

func writePackage(path string, manifest, database, globals []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create backup package: %w", err)
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	entries := map[string][]byte{manifestEntry: manifest, databaseDumpEntry: database, globalsDumpEntry: globals}
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, entryPath := range paths {
		entry, createErr := archive.Create(entryPath)
		if createErr != nil {
			return createErr
		}
		if _, writeErr := entry.Write(entries[entryPath]); writeErr != nil {
			return writeErr
		}
	}
	if err := archive.Close(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return file.Close()
}

func validKDF(parameters KDFParameters) bool {
	return parameters.Name == "argon2id" && parameters.MemoryKB == argonMemoryKB && parameters.Iterations == argonIterations && parameters.Parallelism == argonParallelism && parameters.KeyLength == argonKeyLength && len(mustDecodeSalt(parameters.Salt)) == argonSaltLength
}

func mustDecodeSalt(value string) []byte {
	salt, err := hex.DecodeString(value)
	if err != nil {
		return nil
	}
	return salt
}
