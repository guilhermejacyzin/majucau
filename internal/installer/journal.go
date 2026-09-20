package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const JournalSchemaVersion = "1.0"

type InstallPhase string

const (
	PhaseNew                  InstallPhase = "NEW"
	PhasePreflightOK          InstallPhase = "PREFLIGHT_OK"
	PhasePackageVerified      InstallPhase = "PACKAGE_VERIFIED"
	PhaseDependenciesReady    InstallPhase = "DEPENDENCIES_READY"
	PhaseDatabaseServiceReady InstallPhase = "DATABASE_SERVICE_READY"
	PhaseDatabaseInitialized  InstallPhase = "DATABASE_INITIALIZED"
	PhaseMigrationsApplied    InstallPhase = "MIGRATIONS_APPLIED"
	PhaseWorkerRegistered     InstallPhase = "WORKER_REGISTERED"
	PhaseDesktopInstalled     InstallPhase = "DESKTOP_INSTALLED"
	PhaseSmokeTestOK          InstallPhase = "SMOKE_TEST_OK"
	PhaseCommitted            InstallPhase = "COMMITTED"
	PhasePreflightBlocked     InstallPhase = "PREFLIGHT_BLOCKED"
	PhasePackageInvalid       InstallPhase = "PACKAGE_INVALID"
	PhaseDependencyBlocked    InstallPhase = "DEPENDENCY_BLOCKED"
	PhaseDatabaseBlocked      InstallPhase = "DATABASE_BLOCKED"
	PhaseMigrationBlocked     InstallPhase = "MIGRATION_BLOCKED"
	PhaseServiceBlocked       InstallPhase = "SERVICE_BLOCKED"
	PhaseSmokeTestFailed      InstallPhase = "SMOKE_TEST_FAILED"
	PhaseRecoveryRequired     InstallPhase = "RECOVERY_REQUIRED"
)

var (
	ErrJournalInvalid = errors.New("installer journal is invalid")
	ErrJournalCorrupt = errors.New("installer journal is corrupt")
)

type JournalState struct {
	SchemaVersion  string       `json:"schema_version"`
	InstallID      string       `json:"install_id"`
	ProductVersion string       `json:"product_version"`
	Phase          InstallPhase `json:"phase"`
	UpdatedAt      time.Time    `json:"updated_at"`
	PackageSHA256  string       `json:"package_sha256,omitempty"`
	InstallPath    string       `json:"install_path,omitempty"`
	DataPath       string       `json:"data_path,omitempty"`
	ServiceName    string       `json:"service_name,omitempty"`
	Port           uint16       `json:"port,omitempty"`
	LastErrorCode  string       `json:"last_error_code,omitempty"`
	NextStep       string       `json:"next_step,omitempty"`
}

type JournalStore struct {
	path string
}

func NewJournalStore(path string) (*JournalStore, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || clean == "" || !filepath.IsAbs(clean) || strings.ToLower(filepath.Base(clean)) != "state.json" {
		return nil, ErrJournalInvalid
	}
	return &JournalStore{path: clean}, nil
}

func (s *JournalStore) Load() (JournalState, error) {
	if s == nil || s.path == "" {
		return JournalState{}, ErrJournalInvalid
	}
	payload, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return JournalState{SchemaVersion: JournalSchemaVersion, Phase: PhaseNew}, nil
		}
		return JournalState{}, fmt.Errorf("%w: read", ErrJournalCorrupt)
	}
	var state JournalState
	if err := json.Unmarshal(payload, &state); err != nil {
		return JournalState{}, fmt.Errorf("%w: decode", ErrJournalCorrupt)
	}
	if err := validateJournalState(state); err != nil {
		return JournalState{}, fmt.Errorf("%w: %v", ErrJournalCorrupt, err)
	}
	return state, nil
}

func (s *JournalStore) Save(state JournalState) error {
	if s == nil || s.path == "" {
		return ErrJournalInvalid
	}
	if state.SchemaVersion == "" {
		state.SchemaVersion = JournalSchemaVersion
	}
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	} else {
		state.UpdatedAt = state.UpdatedAt.UTC()
	}
	if err := validateJournalState(state); err != nil {
		return fmt.Errorf("%w: %v", ErrJournalInvalid, err)
	}
	parent := filepath.Dir(s.path)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return fmt.Errorf("%w: create parent", ErrJournalInvalid)
	}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: encode", ErrJournalInvalid)
	}
	temporary, err := os.CreateTemp(parent, ".majucau-install-*.tmp")
	if err != nil {
		return fmt.Errorf("%w: create temporary", ErrJournalInvalid)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	_ = temporary.Chmod(0600)
	if _, err := temporary.Write(payload); err != nil {
		return fmt.Errorf("%w: write", ErrJournalInvalid)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("%w: flush", ErrJournalInvalid)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("%w: close temporary", ErrJournalInvalid)
	}
	if err := os.Rename(temporaryPath, s.path); err != nil {
		return fmt.Errorf("%w: commit", ErrJournalInvalid)
	}
	removeTemporary = false
	return nil
}

func validateJournalState(state JournalState) error {
	if state.SchemaVersion != JournalSchemaVersion || strings.TrimSpace(state.InstallID) == "" || strings.TrimSpace(state.ProductVersion) == "" || !validInstallPhase(state.Phase) || state.UpdatedAt.IsZero() {
		return errors.New("required fields are missing")
	}
	if len(state.InstallID) > 128 || len(state.ProductVersion) > 128 || len(state.ServiceName) > 128 || len(state.NextStep) > 256 || len(state.LastErrorCode) > 128 {
		return errors.New("field is too long")
	}
	for _, value := range []string{state.InstallID, state.ProductVersion, state.ServiceName, state.NextStep, state.LastErrorCode} {
		if strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("field contains a control character")
		}
	}
	if state.PackageSHA256 != "" {
		digest, err := hex.DecodeString(state.PackageSHA256)
		if err != nil || len(digest) != sha256.Size {
			return errors.New("package hash is not SHA-256")
		}
	}
	return nil
}

func validInstallPhase(phase InstallPhase) bool {
	switch phase {
	case PhaseNew, PhasePreflightOK, PhasePackageVerified, PhaseDependenciesReady, PhaseDatabaseServiceReady,
		PhaseDatabaseInitialized, PhaseMigrationsApplied, PhaseWorkerRegistered, PhaseDesktopInstalled,
		PhaseSmokeTestOK, PhaseCommitted, PhasePreflightBlocked, PhasePackageInvalid, PhaseDependencyBlocked,
		PhaseDatabaseBlocked, PhaseMigrationBlocked, PhaseServiceBlocked, PhaseSmokeTestFailed, PhaseRecoveryRequired:
		return true
	default:
		return false
	}
}
