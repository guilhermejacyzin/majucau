//go:build windows

package diagnosticslog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

const privateLogDirectorySDDL = "D:P(A;OICI;FA;;;OW)(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"

func prepareLogDirectory(dir string) error {
	parent := filepath.Dir(dir)
	if err := requirePlainDirectory(parent); err != nil {
		return fmt.Errorf("Majucau data directory is unavailable or unsafe")
	}
	if err := validatePrivateACL(parent, true); err != nil {
		return fmt.Errorf("Majucau data directory permissions are unsafe")
	}
	security, err := windows.SecurityDescriptorFromString(privateLogDirectorySDDL)
	if err != nil {
		return fmt.Errorf("prepare private diagnostic log permissions: %w", err)
	}
	defer windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(security))))
	attributes := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: security,
	}
	dirPtr, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return fmt.Errorf("prepare diagnostic log directory path: %w", err)
	}
	if err := windows.CreateDirectory(dirPtr, attributes); err != nil && !errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return fmt.Errorf("create private diagnostic log directory: %w", err)
	}
	if err := requirePlainDirectory(dir); err != nil {
		return fmt.Errorf("diagnostic log directory is unavailable or unsafe")
	}
	if err := validatePrivateACL(dir, true); err != nil {
		return fmt.Errorf("diagnostic log directory permissions are unsafe")
	}
	return nil
}

func openSafeLogFile(path string) (*os.File, error) {
	if err := requirePlainFileOrMissing(path); err != nil {
		return nil, fmt.Errorf("diagnostic log target is unsafe")
	}
	if _, err := os.Lstat(path); err == nil {
		if err := validatePrivateACL(path, false); err != nil {
			return nil, fmt.Errorf("diagnostic log file permissions are unsafe")
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open diagnostic log file: %w", err)
	}
	if err := requirePlainFileOrMissing(path); err != nil {
		return nil, errors.Join(fmt.Errorf("diagnostic log target changed while opening"), file.Close())
	}
	if err := validatePrivateACL(path, false); err != nil {
		return nil, errors.Join(fmt.Errorf("diagnostic log file permissions are unsafe"), file.Close())
	}
	return file, nil
}

func requirePlainDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("path is not a regular directory")
	}
	attributes, err := windows.GetFileAttributes(path)
	if err != nil {
		return err
	}
	if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("reparse points are not allowed")
	}
	return nil
}

func requirePlainFileOrMissing(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("path is not a regular file")
	}
	attributes, err := windows.GetFileAttributes(path)
	if err != nil {
		return err
	}
	if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return fmt.Errorf("reparse points are not allowed")
	}
	return nil
}

func validatePrivateACL(path string, directory bool) error {
	security, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		return err
	}
	control, _, err := security.Control()
	if err != nil || (directory && control&windows.SE_DACL_PROTECTED == 0) {
		return fmt.Errorf("diagnostic path must use a protected DACL")
	}
	dacl, _, err := security.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 3 {
		return fmt.Errorf("diagnostic path must have exactly three private access entries")
	}
	allowed := map[string]bool{
		"S-1-3-4":      false, // Owner Rights
		"S-1-5-18":     false, // LocalSystem
		"S-1-5-32-544": false, // Built-in Administrators
	}
	wantFlags := uint8(windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE)
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return fmt.Errorf("unexpected diagnostic access entry")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		sidText, err := sid.String()
		if err != nil {
			return err
		}
		if seen, ok := allowed[sidText]; !ok || seen || ace.Mask != windows.ACCESS_MASK(0x001F01FF) {
			return fmt.Errorf("diagnostic path has an unexpected access entry")
		}
		if directory && ace.Header.AceFlags&wantFlags != wantFlags {
			return fmt.Errorf("diagnostic directory access must inherit to its files")
		}
		if !directory && ace.Header.AceFlags&windows.INHERITED_ACE == 0 {
			return fmt.Errorf("diagnostic file permissions must inherit from its directory")
		}
		allowed[sidText] = true
	}
	for _, found := range allowed {
		if !found {
			return fmt.Errorf("diagnostic path is missing a required access entry")
		}
	}
	return nil
}
