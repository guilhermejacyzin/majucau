//go:build !windows

package backup

import (
	"fmt"
	"os"
	"path/filepath"
)

// The desktop product is Windows-only. This fallback keeps non-Windows CI and
// package tooling buildable while restricting temporary files to their owner.
func newPrivateTempDir(parent, prefix string) (string, error) {
	if !filepath.IsAbs(parent) {
		return "", fmt.Errorf("private temporary directory parent must be absolute")
	}
	path, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.RemoveAll(path)
		return "", err
	}
	return path, nil
}
