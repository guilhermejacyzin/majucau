//go:build windows

package securetemp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

func TestNewWorkspaceAppliesProtectedDACLAndRestrictsInheritedAccess(t *testing.T) {
	workspace, err := NewWorkspace(t.TempDir(), "majucau-securetemp-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := workspace.RemoveAll(); err != nil {
			t.Errorf("remove private directory: %v", err)
		}
	})
	dir := workspace.Path()

	dirSecurity := readDACL(t, dir)
	control, _, err := dirSecurity.Control()
	if err != nil {
		t.Fatalf("read private directory security descriptor control: %v", err)
	}
	if control&windows.SE_DACL_PROTECTED == 0 {
		t.Fatalf("private directory DACL is not protected: %s", dirSecurity)
	}
	assertPrivateDACL(t, dirSecurity.String(), "OICI")

	child := filepath.Join(dir, "private.tmp")
	if err := os.WriteFile(child, []byte("temporary"), 0600); err != nil {
		t.Fatal(err)
	}
	childSecurity := readDACL(t, child)
	assertPrivateDACL(t, childSecurity.String(), "ID")
}

func TestNewDirPropagatesDescriptorReleaseFailure(t *testing.T) {
	var created string
	api := failingPrivateDirectoryAPI(t, &created, false)
	dir, err := newDirWithAPI(t.TempDir(), "majucau-release-failure-", api)
	if err == nil {
		t.Fatal("descriptor release failure must be returned")
	}
	if dir != "" {
		t.Fatalf("directory path = %q, want empty after cleanup", dir)
	}
	if created == "" {
		t.Fatal("fake directory API was not called")
	}
	if _, statErr := os.Stat(created); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("directory should be removed after descriptor release failure, stat error = %v", statErr)
	}
}

func TestNewDirReportsDescriptorReleaseAndCleanupFailures(t *testing.T) {
	var created string
	api := failingPrivateDirectoryAPI(t, &created, true)
	dir, err := newDirWithAPI(t.TempDir(), "majucau-cleanup-failure-", api)
	if err == nil {
		t.Fatal("descriptor release and directory cleanup failures must be returned")
	}
	if dir != "" {
		t.Fatalf("directory path = %q, want empty after failed cleanup", dir)
	}
	var cleanupErr *CleanupError
	if !errors.As(err, &cleanupErr) {
		t.Fatalf("error type %T does not expose a manual cleanup path", err)
	}
	if cleanupErr.Path != created {
		t.Fatalf("manual cleanup path = %q, want %q", cleanupErr.Path, created)
	}
	if strings.Contains(err.Error(), created) {
		t.Fatalf("ordinary error exposed the local path: %q", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(created); err != nil {
			t.Errorf("remove injected cleanup-failure fixture: %v", err)
		}
	})
}

func failingPrivateDirectoryAPI(t *testing.T, created *string, leaveChild bool) privateDirectoryAPI {
	t.Helper()
	return privateDirectoryAPI{
		convert: func(_ *uint16, descriptor **windows.SECURITY_DESCRIPTOR, size *uint32) (uintptr, error) {
			*descriptor = &windows.SECURITY_DESCRIPTOR{}
			*size = 1
			return 1, nil
		},
		free: func(address uintptr) (uintptr, error) {
			if address == 0 {
				t.Error("LocalFree received a nil security descriptor")
			}
			return address, syscall.Errno(5)
		},
		create: func(path *uint16, attributes *windows.SecurityAttributes) (uintptr, error) {
			if attributes == nil || attributes.SecurityDescriptor == nil {
				t.Error("CreateDirectoryW did not receive the protected security descriptor")
			}
			*created = windows.UTF16PtrToString(path)
			if err := os.Mkdir(*created, 0o700); err != nil {
				return 0, err
			}
			if leaveChild {
				if err := os.WriteFile(filepath.Join(*created, "keep.tmp"), []byte("temporary"), 0o600); err != nil {
					return 0, err
				}
			}
			return 1, nil
		},
		remove: os.Remove,
	}
}

func readDACL(t *testing.T, path string) *windows.SECURITY_DESCRIPTOR {
	t.Helper()
	security, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatalf("read DACL for temporary path: %v", err)
	}
	if security == nil {
		t.Fatal("temporary path has no security descriptor")
	}
	dacl, _, err := security.DACL()
	if err != nil {
		t.Fatalf("read DACL for temporary path: %v", err)
	}
	if dacl == nil {
		t.Fatal("temporary path has a null DACL")
	}
	return security
}

func assertPrivateDACL(t *testing.T, sddl, flags string) {
	t.Helper()
	for _, sid := range []string{"OW", "SY", "BA"} {
		ace := "(A;" + flags + ";FA;;;" + sid + ")"
		if !strings.Contains(sddl, ace) {
			t.Errorf("DACL %q is missing required ACE %q", sddl, ace)
		}
	}
	if got := strings.Count(sddl, "("); got != 3 {
		t.Errorf("DACL %q contains %d ACEs; want exactly the owner, SYSTEM, and administrators", sddl, got)
	}
}
