//go:build windows

package backup

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// newPrivateTempDir creates the directory with a protected DACL at creation
// time. It never relies on the ACL inherited from a shared ProgramData or TEMP
// parent for plaintext dumps, passfiles, or restore workspaces.
func newPrivateTempDir(parent, prefix string) (string, error) {
	if !filepath.IsAbs(parent) {
		return "", fmt.Errorf("private temporary directory parent must be absolute")
	}
	convert := windows.NewLazySystemDLL("advapi32.dll").NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	localFree := windows.NewLazySystemDLL("kernel32.dll").NewProc("LocalFree")
	if err := convert.Find(); err != nil {
		return "", err
	}
	if err := localFree.Find(); err != nil {
		return "", err
	}
	createDir := windows.NewLazySystemDLL("kernel32.dll").NewProc("CreateDirectoryW")
	if err := createDir.Find(); err != nil {
		return "", err
	}

	// Protected DACL: only the directory owner, LocalSystem, and administrators
	// can access the tree. OICI passes the same restriction to every child.
	sddl, err := syscall.UTF16PtrFromString("D:P(A;OICI;FA;;;OW)(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)")
	if err != nil {
		return "", err
	}
	var descriptor *windows.SECURITY_DESCRIPTOR
	var descriptorSize uint32
	result, _, callErr := convert.Call(
		uintptr(unsafe.Pointer(sddl)), 1,
		uintptr(unsafe.Pointer(&descriptor)), uintptr(unsafe.Pointer(&descriptorSize)),
	)
	if result == 0 {
		return "", fmt.Errorf("create private temporary directory security descriptor (last error %v)", callErr)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(descriptor)))

	attributes := windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: descriptor,
	}
	for attempt := 0; attempt < 8; attempt++ {
		suffix := make([]byte, 16)
		if _, err := rand.Read(suffix); err != nil {
			return "", fmt.Errorf("generate private temporary directory name: %w", err)
		}
		path := filepath.Join(parent, prefix+hex.EncodeToString(suffix))
		pathPtr, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return "", err
		}
		created, _, createErr := createDir.Call(uintptr(unsafe.Pointer(pathPtr)), uintptr(unsafe.Pointer(&attributes)))
		if created != 0 {
			return path, nil
		}
		if createErr != syscall.ERROR_ALREADY_EXISTS {
			return "", fmt.Errorf("create private temporary directory: %w", createErr)
		}
	}
	return "", fmt.Errorf("could not allocate a unique private temporary directory")
}
