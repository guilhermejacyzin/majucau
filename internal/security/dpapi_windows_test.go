//go:build windows

package security

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestDPAPIStoreRoundTripAndPersistence(t *testing.T) {
	store, err := NewDPAPIStore(t.TempDir())
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
