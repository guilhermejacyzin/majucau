//go:build !windows

package securetemp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// NewDir creates a temporary directory restricted to its owner.
func NewDir(parent, prefix string) (string, error) {
	if !filepath.IsAbs(parent) {
		return "", fmt.Errorf("private temporary directory parent must be absolute")
	}
	path, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0700); err != nil {
		return "", errors.Join(err, RemoveAll(path))
	}
	return path, nil
}
