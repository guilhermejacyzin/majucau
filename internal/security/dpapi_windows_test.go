//go:build windows

package security

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDPAPIStoreRoundTripAndPersistence(t *testing.T) {
	store, err := newTestDPAPIStore(t)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	secret := []byte("test-only-secret-that-must-not-appear-in-the-vault")
	if err := store.Put(ctx, "bling_test", secret); err != nil {
		t.Fatal(err)
	}
	encrypted, err := os.ReadFile(store.path("bling_test"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encrypted, secret) {
		t.Fatal("DPAPI vault persisted plaintext")
	}
	got, err := store.Get(ctx, "bling_test")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("round-trip mismatch")
	}
	reopened, err := NewDPAPIStore(store.dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err = reopened.Get(ctx, "bling_test")
	if err != nil || !bytes.Equal(got, secret) {
		t.Fatalf("restart persistence mismatch")
	}
	rotated := []byte("rotated-test-only-secret")
	if err := reopened.Put(ctx, "bling_test", rotated); err != nil {
		t.Fatal(err)
	}
	reopenedAgain, err := NewDPAPIStore(store.dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err = reopenedAgain.Get(ctx, "bling_test")
	if err != nil || !bytes.Equal(got, rotated) {
		t.Fatalf("atomic rotation persistence mismatch")
	}
	if err := reopenedAgain.Delete(ctx, "bling_test"); err != nil {
		t.Fatal(err)
	}
	if _, err := reopenedAgain.Get(ctx, "bling_test"); !os.IsNotExist(err) {
		t.Fatalf("expected deleted secret to be absent, got %v", err)
	}
}

func TestMachineDPAPIStoreRoundTripAndScopeIsolation(t *testing.T) {
	root := t.TempDir()
	prepareTestSecretDirectory(t, root)
	dir := filepath.Join(root, "machine-vault")
	store, err := NewMachineDPAPIStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	secret := []byte("machine-scope-test-secret")
	if err := store.Put(context.Background(), "drive_test", secret); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(context.Background(), "drive_test")
	if err != nil || !bytes.Equal(got, secret) {
		t.Fatalf("machine-scope round-trip mismatch: %v", err)
	}
	if store.scope != dpapiScopeMachine || store.protectionFlags != cryptProtectLocalMachine {
		t.Fatal("machine vault did not select the local-machine DPAPI protection flag")
	}
	if _, err := NewDPAPIStore(dir); err == nil {
		t.Fatal("current-user constructor must reject a machine-scope vault")
	}
	for _, target := range []struct {
		path      string
		directory bool
	}{
		{path: dir, directory: true},
		{path: filepath.Join(dir, ".scope")},
		{path: filepath.Join(dir, ".entropy")},
		{path: store.path("drive_test")},
	} {
		if err := validatePrivateSecretACL(target.path, target.directory, store.userSID); err != nil {
			t.Errorf("machine vault path is not restricted: %v", err)
		}
	}
}

func TestDPAPIStoreMigratesOnlyLegacyCurrentUserVault(t *testing.T) {
	root := t.TempDir()
	prepareTestSecretDirectory(t, root)
	dir := filepath.Join(root, "legacy-vault")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	userSID, err := currentProcessSID()
	if err != nil {
		t.Fatal(err)
	}
	entropy := bytes.Repeat([]byte{0xA5}, dpapiEntropyBytes)
	if err := os.WriteFile(filepath.Join(dir, ".entropy"), entropy, 0600); err != nil {
		t.Fatal(err)
	}
	if err := preparePrivateSecretFile(filepath.Join(dir, ".entropy"), userSID); err != nil {
		t.Fatal(err)
	}
	legacyBlob, err := protect([]byte("legacy-user-scope"), entropy, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := atomicWrite(filepath.Join(dir, "legacy.dpapi"), legacyBlob, 0600, userSID); err != nil {
		t.Fatal(err)
	}
	zeroBytes(legacyBlob)
	store, err := NewDPAPIStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if store.scope != dpapiScopeCurrentUser {
		t.Fatal("legacy vault should retain current-user DPAPI scope")
	}
	legacy, err := store.Get(context.Background(), "legacy")
	if err != nil || string(legacy) != "legacy-user-scope" {
		t.Fatalf("legacy current-user secret was not preserved: %v", err)
	}
	if _, err := NewMachineDPAPIStore(dir); err == nil {
		t.Fatal("machine-scope constructor must reject a legacy current-user vault")
	}
}

func TestMachineDPAPIStoreRejectsNonemptyUnmarkedVault(t *testing.T) {
	root := t.TempDir()
	prepareTestSecretDirectory(t, root)
	dir := filepath.Join(root, "unmarked-vault")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".entropy"), bytes.Repeat([]byte{0xA5}, dpapiEntropyBytes), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewMachineDPAPIStore(dir); err == nil {
		t.Fatal("machine-scope constructor must not adopt an existing unmarked vault")
	}
}

func TestDPAPIStoreRejectsOversizedEntropyAndSecretFiles(t *testing.T) {
	ctx := context.Background()
	badEntropyRoot := t.TempDir()
	prepareTestSecretDirectory(t, badEntropyRoot)
	badEntropyDir := filepath.Join(badEntropyRoot, "vault")
	if err := os.Mkdir(badEntropyDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(badEntropyDir, ".entropy"), bytes.Repeat([]byte{0xA5}, dpapiEntropyBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewDPAPIStore(badEntropyDir); err == nil {
		t.Fatal("entropy file larger than its fixed 32-byte format must be rejected")
	}

	store, err := newTestDPAPIStore(t)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(ctx, "bling_test", make([]byte, maxDPAPISecretValueBytes+1)); err == nil {
		t.Fatal("secret value larger than the IPC payload limit must be rejected")
	}
	if err := os.WriteFile(store.path("bling_test"), make([]byte, maxDPAPISecretBlobBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, "bling_test"); err == nil {
		t.Fatal("oversized encrypted secret file must be rejected before DPAPI processing")
	}
}

func TestDPAPIStoreRestrictsDirectoryAndSecretFiles(t *testing.T) {
	store, err := newTestDPAPIStore(t)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), "bling_test", []byte("test-only-secret")); err != nil {
		t.Fatal(err)
	}
	for _, target := range []struct {
		path      string
		directory bool
	}{
		{path: store.dir, directory: true},
		{path: filepath.Join(store.dir, ".scope")},
		{path: filepath.Join(store.dir, ".entropy")},
		{path: store.path("bling_test")},
	} {
		if err := validatePrivateSecretACL(target.path, target.directory, store.userSID); err != nil {
			t.Errorf("secret path is not restricted: %v", err)
		}
	}
}

func newTestDPAPIStore(t *testing.T) (*DPAPIStore, error) {
	t.Helper()
	root := t.TempDir()
	prepareTestSecretDirectory(t, root)
	return NewDPAPIStore(filepath.Join(root, "vault"))
}

func prepareTestSecretDirectory(t *testing.T, path string) {
	t.Helper()
	userSID, err := currentProcessSID()
	if err != nil {
		t.Fatal(err)
	}
	if err := setPrivateSecretACL(path, true, userSID); err != nil {
		t.Fatalf("prepare private test directory: %v", err)
	}
}
