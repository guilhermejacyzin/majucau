package securetemp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCleanupPathRequired = errors.New("temporary cleanup path is required")
	ErrUnsafeCleanupPath  = errors.New("refusing unsafe temporary cleanup path")
)

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

// RemoveAll removes a temporary workspace and treats an already-absent path
// as successfully cleaned up.
func RemoveAll(path string) error {
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
