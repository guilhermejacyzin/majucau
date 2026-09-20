//go:build !windows

package installer

import "errors"

const DefaultInstallLockName = "MajucauInstaller"

var (
	ErrInstallerLocked     = errors.New("another installer instance is active")
	ErrInstallerLockFailed = errors.New("installer lock is unavailable")
)

type InstallLock struct{}

func AcquireInstallLock(string) (*InstallLock, error) { return nil, ErrInstallerLockFailed }
func (*InstallLock) Release() error                   { return nil }
