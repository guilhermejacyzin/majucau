package securetemp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveFileTreatsMissingPathAsAlreadyCleaned(t *testing.T) {
	path := filepath.Join(t.TempDir(), "already-absent.tmp")
	if err := RemoveFile(path); err != nil {
		t.Fatalf("first cleanup of a missing file: %v", err)
	}
	if err := RemoveFile(path); err != nil {
		t.Fatalf("repeated cleanup of a missing file: %v", err)
	}
}

func TestRemoveFileReturnsSanitizedErrorWhenCleanupFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "non-empty-directory")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(path, "keep.tmp")
	if err := os.WriteFile(child, []byte("temporary"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := RemoveFile(path)
	if err == nil {
		t.Fatal("removing a non-empty directory as a file should fail")
	}
	var cleanupErr *CleanupError
	if !errors.As(err, &cleanupErr) {
		t.Fatalf("cleanup error type = %T, want *CleanupError", err)
	}
	if cleanupErr.Path != path {
		t.Fatalf("manual cleanup path = %q, want %q", cleanupErr.Path, path)
	}
	if cleanupErr.Cause == nil {
		t.Fatal("cleanup error must preserve the underlying failure")
	}
	if strings.Contains(err.Error(), path) {
		t.Fatalf("ordinary error exposed the local path: %q", err)
	}
	if _, err := os.Stat(child); err != nil {
		t.Fatalf("failed cleanup unexpectedly removed its contents: %v", err)
	}
}
