//go:build !windows

package diagnosticslog

import (
	"errors"
	"fmt"
	"os"
)

func prepareLogDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("prepare private diagnostic log directory: %w", err)
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return fmt.Errorf("inspect diagnostic log directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("diagnostic log directory is not a regular directory")
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("restrict diagnostic log directory permissions: %w", err)
	}
	return nil
}

func openSafeLogFile(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err == nil && (!info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0) {
		return nil, fmt.Errorf("diagnostic log target is not a regular file")
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect diagnostic log file: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open diagnostic log file: %w", err)
	}
	if err := file.Chmod(0600); err != nil {
		return nil, fmt.Errorf("restrict diagnostic log file permissions: %w", errors.Join(err, file.Close()))
	}
	return file, nil
}

func inspectExistingLogFile(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("inspect diagnostic log file")
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("diagnostic log target is not a regular file")
	}
	return info, nil
}
