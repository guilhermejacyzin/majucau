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

func TestNewWorkspaceRejectsPathSeparatorsInPrefix(t *testing.T) {
	for _, prefix := range []string{"../outside-", `..\outside-`, `/outside-`, `\outside-`} {
		workspace, err := NewWorkspace(t.TempDir(), prefix)
		if workspace != nil {
			t.Fatalf("prefix %q created workspace %q", prefix, workspace.Path())
		}
		if !errors.Is(err, ErrUnsafeWorkspacePrefix) {
			t.Fatalf("prefix %q error = %v, want ErrUnsafeWorkspacePrefix", prefix, err)
		}
	}
}

func TestNewWorkspaceKeepsSafePrefixInsideParent(t *testing.T) {
	parent := t.TempDir()
	workspace, err := NewWorkspace(parent, "majucau-safe-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := workspace.RemoveAll(); err != nil {
			t.Errorf("remove private workspace: %v", err)
		}
	})
	if got := filepath.Dir(workspace.Path()); got != parent {
		t.Fatalf("workspace parent = %q, want %q", got, parent)
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
