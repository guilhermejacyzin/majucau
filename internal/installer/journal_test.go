package installer

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestJournalStoreRoundTripAndFreshState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Install", "state.json")
	store, err := NewJournalStore(path)
	if err != nil {
		t.Fatal(err)
	}
	state, err := store.Load()
	if err != nil || state.Phase != PhaseNew || state.SchemaVersion != JournalSchemaVersion {
		t.Fatalf("fresh journal: %#v %v", state, err)
	}
	expected := JournalState{InstallID: "install-test", ProductVersion: "0.1.0", Phase: PhasePreflightOK, InstallPath: `C:\Program Files\Majucau`, DataPath: `C:\ProgramData\Majucau`, ServiceName: "MajucauWorker", Port: 54329, NextStep: "verify package"}
	if err := store.Save(expected); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.InstallID != expected.InstallID || got.Phase != expected.Phase || got.Port != expected.Port || got.UpdatedAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatalf("round trip mismatch: %#v", got)
	}
}

func TestJournalRejectsUnknownPhaseAndInvalidHash(t *testing.T) {
	store, err := NewJournalStore(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	base := JournalState{InstallID: "install-test", ProductVersion: "0.1.0", Phase: PhaseNew}
	invalidPhase := base
	invalidPhase.Phase = InstallPhase("UNKNOWN")
	if err := store.Save(invalidPhase); !errors.Is(err, ErrJournalInvalid) {
		t.Fatalf("unknown phase error = %v", err)
	}
	invalidHash := base
	invalidHash.PackageSHA256 = "abc"
	if err := store.Save(invalidHash); !errors.Is(err, ErrJournalInvalid) {
		t.Fatalf("invalid hash error = %v", err)
	}
}
