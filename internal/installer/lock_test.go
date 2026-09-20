package installer

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestInstallLockPreventsConcurrentInstance(t *testing.T) {
	name := fmt.Sprintf("Local\\MajucauInstallerTest-%d", os.Getpid())
	first, err := AcquireInstallLock(name)
	if errors.Is(err, ErrInstallerLockFailed) {
		t.Skip("platform does not provide an installer mutex")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	second, err := AcquireInstallLock(name)
	if !errors.Is(err, ErrInstallerLocked) {
		if second != nil {
			_ = second.Release()
		}
		t.Fatalf("second lock error = %v, want %v", err, ErrInstallerLocked)
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	third, err := AcquireInstallLock(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := third.Release(); err != nil {
		t.Fatal(err)
	}
}
