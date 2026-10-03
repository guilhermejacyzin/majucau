//go:build windows

package securetemp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestNewDirAppliesProtectedDACLAndRestrictsInheritedAccess(t *testing.T) {
	dir, err := NewDir(t.TempDir(), "majucau-securetemp-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := RemoveAll(dir); err != nil {
			t.Errorf("remove private directory: %v", err)
		}
	})

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
