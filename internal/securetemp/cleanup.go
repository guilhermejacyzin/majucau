package securetemp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCleanupPathRequired = errors.New("temporary cleanup path is required")
	ErrUnsafeCleanupPath   = errors.New("refusing unsafe temporary cleanup path")
)

// Workspace represents a private temporary directory created by this package.
// Its path cannot be supplied by callers when requesting recursive cleanup.
type Workspace struct {
	path string
}

// NewWorkspace creates a private temporary directory and returns a cleanup
// capability for that directory.
func NewWorkspace(parent, prefix string) (*Workspace, error) {
	path, err := createPrivateDir(parent, prefix)
	if err != nil {
		return nil, err
	}
	return &Workspace{path: path}, nil
}

// Path returns the private directory path for creating files inside it.
func (w *Workspace) Path() string {
	if w == nil {
		return ""
	}
	return w.path
}

// RemoveAll removes this workspace and treats an already-absent path as
// successfully cleaned up. A failed cleanup retains the path for retry.
func (w *Workspace) RemoveAll() error {
	if w == nil {
		return ErrCleanupPathRequired
	}
	return removeAllPath(w.path)
}

// CleanupError reports a failed cleanup. Path is available to authorized
// callers for manual cleanup but is omitted from Error to keep account names
// and other local path details out of ordinary logs.
type CleanupError struct {
	Path  string
	Cause error
}

func (e *CleanupError) Error() string {
	return "temporary data cleanup failed; manual cleanup may be required"
}

func (e *CleanupError) Unwrap() error { return e.Cause }

// RemoveFile removes one temporary file and treats an already-absent file as
// successfully cleaned up.
func RemoveFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrCleanupPathRequired
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return &CleanupError{Path: path, Cause: err}
	}
	return nil
}

func removeAllPath(path string) error {
	if strings.TrimSpace(path) == "" {
		return ErrCleanupPathRequired
	}
	clean := filepath.Clean(path)
	root := filepath.VolumeName(clean) + string(os.PathSeparator)
	if !filepath.IsAbs(clean) || clean == root {
		return ErrUnsafeCleanupPath
	}
	if err := os.RemoveAll(path); err != nil {
		return &CleanupError{Path: path, Cause: err}
	}
	return nil
}
