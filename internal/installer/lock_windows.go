//go:build windows

package installer

import (
	"errors"
	"strings"

	"golang.org/x/sys/windows"
)

const DefaultInstallLockName = `Global\MajucauInstaller`

var (
	ErrInstallerLocked     = errors.New("another installer instance is active")
	ErrInstallerLockFailed = errors.New("installer lock is unavailable")
)

type InstallLock struct {
	handle windows.Handle
}

func AcquireInstallLock(name string) (*InstallLock, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 260 || strings.ContainsAny(name, "\r\n\x00") {
		return nil, ErrInstallerLockFailed
	}
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, ErrInstallerLockFailed
	}
	handle, err := windows.CreateMutex(nil, true, ptr)
	if err != nil {
		if handle != 0 {
			_ = windows.CloseHandle(handle)
		}
		if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
			return nil, ErrInstallerLocked
		}
		return nil, ErrInstallerLockFailed
	}
	return &InstallLock{handle: handle}, nil
}

func (lock *InstallLock) Release() error {
	if lock == nil || lock.handle == 0 {
		return nil
	}
	releaseErr := windows.ReleaseMutex(lock.handle)
	closeErr := windows.CloseHandle(lock.handle)
	lock.handle = 0
	if releaseErr != nil {
		return releaseErr
	}
	return closeErr
}
