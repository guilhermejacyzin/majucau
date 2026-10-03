// Package backup creates and verifies encrypted, portable PostgreSQL backups.
// It deliberately does not restore or mutate a live cluster; restoration is a
// separate, explicitly authorized operation with compatibility checks.
package backup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"

	"majucau.local/financial-intelligence/internal/securetemp"
)

const (
	FormatVersion               = "2"
	ManifestSchemaVersion       = "2.0"
	legacyFormatVersion         = "1"
	legacyManifestSchemaVersion = "1.0"
	BackupSource                = "MAJUCAU_LOCAL_POSTGRESQL"
	TokenExportPolicy           = "EXCLUDED"
	databaseDumpEntry           = "payload/database.dump.saead"
	globalsDumpEntry            = "payload/globals.sql.saead"
	legacyDatabaseDumpEntry     = "payload/database.dump.gcm"
	legacyGlobalsDumpEntry      = "payload/globals.sql.gcm"
	streamKeysetEntry           = "metadata/streaming-keyset.gcm"
	streamCipher                = "TINK_AES256_GCM_HKDF_SEGMENT_1_MIB"
	manifestEntry               = "backup-manifest.json"
	encryptedHeader             = "MAJUCAU-BACKUP-1"
	argonMemoryKB               = 64 * 1024
	argonIterations             = 1
	argonParallelism            = 4
	argonKeyLength              = 32
	argonSaltLength             = 16
)

var (
	ErrInvalidOptions       = errors.New("invalid backup options")
	ErrInvalidPackage       = errors.New("invalid backup package")
	ErrWrongPassphrase      = errors.New("backup passphrase is invalid")
	ErrUnsupportedSchema    = errors.New("backup schema is incompatible")
	ErrRestoreHooksRequired = errors.New("restore worker lifecycle hooks are required")
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
	Cipher            string `json:"cipher,omitempty"`
	KeysetFile        File `json:"keyset_file,omitempty"`
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

// TemporaryWorkspaceCleanupError reports that a private workspace could not
// be removed. Path is provided to authorized callers for manual cleanup; it is
// intentionally omitted from Error so paths containing account names are not
// copied into ordinary logs.
type TemporaryWorkspaceCleanupError struct {
	Path  string
	Cause error
}

func (e *TemporaryWorkspaceCleanupError) Error() string {
	return "private backup workspace cleanup failed; manual cleanup is required"
}

func (e *TemporaryWorkspaceCleanupError) Unwrap() error { return e.Cause }

type Verification struct {
	Valid    bool
	Manifest Manifest
	Issues   []string
}

// RestoreOptions makes the destructive boundary explicit. Callers must own
// the worker lifecycle and provide a validation hook; this package never
// guesses which Windows service is safe to stop or starts a worker before the
// restored database has been validated.
type RestoreOptions struct {
	PackagePath    string
	DatabaseURL    string
	BackupDir      string
	CurrentSchema  string
	AppVersion     string
	InstallID      string
	Passphrase     []byte
	PGDumpPath     string
	PGDumpAllPath  string
	PGRestorePath  string
	PSQLPath       string
	RunCommand     CommandRunner
	StopWorker     func(context.Context) error
	Validate       func(context.Context) error
	StartWorker    func(context.Context) error
}

type RestoreResult struct {
	Manifest          Manifest
	PreRestoreBackup  Result
	Status             string
}

func Create(ctx context.Context, options Options) (result Result, retErr error) {
	if ctx == nil {
		return Result{}, ErrInvalidOptions
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
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
	temporary, err := securetemp.NewDir(options.OutputDir, ".backup-")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary backup directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(temporary); cleanupErr != nil {
			retErr = errors.Join(retErr, &TemporaryWorkspaceCleanupError{Path: temporary, Cause: cleanupErr})
		}
	}()
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
		Cipher: streamCipher,
	}
	key := deriveKey(options.Passphrase, salt, manifest.KDF)
	defer clearBytes(key)
	stamp := now().UTC().Format("20060102T150405.000000000Z")
	finalPath := filepath.Join(options.OutputDir, "majucau-backup-"+stamp+".mjbk")
	temporaryPackage := filepath.Join(temporary, "backup-package.tmp")
	manifest, err = writeV2Package(ctx, temporaryPackage, manifest, key, dumpPath, globalsPath)
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := os.Rename(temporaryPackage, finalPath); err != nil {
		return Result{}, fmt.Errorf("commit backup package: %w", err)
	}
	return Result{Path: finalPath, Manifest: manifest}, nil
}

func Verify(path string, passphrase []byte, currentSchema string) Verification {
	return VerifyContext(context.Background(), path, passphrase, currentSchema)
}

// VerifyContext authenticates the complete archive while allowing callers to
// cancel long-running verification of large packages.
func VerifyContext(ctx context.Context, path string, passphrase []byte, currentSchema string) Verification {
	return verifyPackage(ctx, path, passphrase, currentSchema)
}

// Restore validates and restores a package into the explicitly supplied
// PostgreSQL database. It creates a fresh backup before stopping the worker,
// never runs pg_restore for an invalid package, and leaves the caller in
// RECOVERY_REQUIRED when a destructive step or post-restore validation fails.
func Restore(ctx context.Context, options RestoreOptions) (result RestoreResult, retErr error) {
	if ctx == nil {
		return RestoreResult{}, ErrInvalidOptions
	}
	if err := ctx.Err(); err != nil {
		return RestoreResult{}, err
	}
	if err := validateRestoreOptions(options); err != nil {
		return RestoreResult{}, err
	}
	if err := os.MkdirAll(options.BackupDir, 0700); err != nil {
		return RestoreResult{}, fmt.Errorf("create restore parent directory: %w", err)
	}
	temporary, err := securetemp.NewDir(options.BackupDir, ".restore-")
	if err != nil {
		return RestoreResult{Status: "RECOVERY_REQUIRED"}, fmt.Errorf("create private restore workspace: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(temporary); cleanupErr != nil {
			retErr = errors.Join(retErr, &TemporaryWorkspaceCleanupError{Path: temporary, Cause: cleanupErr})
		}
	}()
	stagedPackage := filepath.Join(temporary, "input.mjbk")
	if err := copyEncryptedPackage(ctx, options.PackagePath, stagedPackage); err != nil {
		return RestoreResult{Status: "RECOVERY_REQUIRED"}, fmt.Errorf("stage encrypted backup: %w", err)
	}
	verification := VerifyContext(ctx, stagedPackage, options.Passphrase, options.CurrentSchema)
	if !verification.Valid {
		return RestoreResult{Manifest: verification.Manifest, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("%w: %s", ErrInvalidPackage, strings.Join(verification.Issues, "; "))
	}
	databasePath, globalsPath, err := decryptPackageToFiles(ctx, stagedPackage, options.Passphrase, verification.Manifest, temporary)
	if err != nil {
		return RestoreResult{Manifest: verification.Manifest, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("authenticate and stage restore payloads: %w", err)
	}
	preRestore, err := Create(ctx, Options{
		DatabaseURL: options.DatabaseURL, OutputDir: options.BackupDir,
		AppVersion: options.AppVersion, SchemaVersion: options.CurrentSchema,
		InstallID: options.InstallID, Passphrase: options.Passphrase,
		PGDumpPath: options.PGDumpPath, PGDumpAllPath: options.PGDumpAllPath, RunCommand: options.RunCommand,
	})
	if err != nil {
		return RestoreResult{Manifest: verification.Manifest, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("create pre-restore backup: %w", err)
	}
	if err := options.StopWorker(ctx); err != nil {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("stop worker before restore: %w", err)
	}
	config, err := pgx.ParseConfig(options.DatabaseURL)
	if err != nil || config.Host == "" || config.Database == "" || config.User == "" {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, ErrInvalidOptions
	}
	_, env, err := connectionEnvironment(config, temporary)
	if err != nil {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, err
	}
	run := options.RunCommand
	if run == nil {
		run = runExternalCommand
	}
	pgRestore := options.PGRestorePath
	if pgRestore == "" {
		pgRestore = "pg_restore"
	}
	psql := options.PSQLPath
	if psql == "" {
		psql = "psql"
	}
	common := []string{"--host", config.Host, "--port", fmt.Sprintf("%d", config.Port), "--username", config.User, "--dbname", config.Database}
	restoreArgs := append([]string{"--clean", "--if-exists", "--no-owner", "--no-privileges", "--exit-on-error"}, append(common, databasePath)...)
	if err := run(ctx, pgRestore, restoreArgs, env); err != nil {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("pg_restore failed: %w", err)
	}
	globalsArgs := append([]string{"--set=ON_ERROR_STOP=1", "--file", globalsPath}, common...)
	if err := run(ctx, psql, globalsArgs, env); err != nil {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("restore globals failed: %w", err)
	}
	if options.Validate != nil {
		if err := options.Validate(ctx); err != nil {
			return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("validate restored database: %w", err)
		}
	}
	if err := options.StartWorker(ctx); err != nil {
		return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RECOVERY_REQUIRED"}, fmt.Errorf("start worker after restore: %w", err)
	}
	return RestoreResult{Manifest: verification.Manifest, PreRestoreBackup: preRestore, Status: "RESTORED_NEEDS_RECONNECT"}, nil
}

func validateRestoreOptions(options RestoreOptions) error {
	if !filepath.IsAbs(options.PackagePath) || !filepath.IsAbs(options.BackupDir) || strings.TrimSpace(options.DatabaseURL) == "" || strings.TrimSpace(options.CurrentSchema) == "" || strings.TrimSpace(options.AppVersion) == "" || strings.TrimSpace(options.InstallID) == "" || len(options.Passphrase) < 12 {
		return ErrInvalidOptions
	}
	if options.StopWorker == nil || options.Validate == nil || options.StartWorker == nil {
		return ErrRestoreHooksRequired
	}
	return nil
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

func fileRecord(path string, payload []byte) File {
	digest := sha256.Sum256(payload)
	return File{Path: path, SHA256: hex.EncodeToString(digest[:]), SizeBytes: int64(len(payload))}
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

