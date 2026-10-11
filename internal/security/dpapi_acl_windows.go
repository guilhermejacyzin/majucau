//go:build windows

package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const privateSecretAccessMask = windows.ACCESS_MASK(0x001F01FF)

func currentProcessSID() (string, error) {
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		return "", err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil || user == nil || user.User.Sid == nil {
		if err == nil {
			err = errors.New("current process token has no user SID")
		}
		return "", err
	}
	return user.User.Sid.String(), nil
}

func preparePrivateSecretDirectory(path, userSID string) error {
	if userSID == "" || !filepath.IsAbs(path) {
		return errors.New("secret directory identity or path is invalid")
	}
	path = filepath.Clean(path)
	parent := filepath.Dir(path)
	if err := requirePlainWindowsDirectory(parent); err != nil {
		return errors.New("secret directory parent is unavailable or unsafe")
	}
	if err := validatePrivateSecretACL(parent, true, userSID); err != nil {
		return errors.New("secret directory parent permissions are unsafe")
	}
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		if err := createPrivateSecretDirectory(path, userSID); err != nil {
			return errors.New("could not create a private secret directory")
		}
	} else if err != nil {
		return errors.New("secret directory is unavailable")
	}
	if err := requirePlainWindowsDirectory(path); err != nil {
		return errors.New("secret directory is unavailable or unsafe")
	}
	if err := setPrivateSecretACL(path, true, userSID); err != nil {
		return errors.New("could not restrict secret directory access")
	}
	if err := validatePrivateSecretACL(path, true, userSID); err != nil {
		return errors.New("secret directory permissions are unsafe")
	}
	return nil
}

func createPrivateSecretDirectory(path, userSID string) error {
	security, err := privateSecretSecurityDescriptor(true, userSID)
	if err != nil {
		return err
	}
	attributes := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: security,
	}
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	err = windows.CreateDirectory(pathPtr, attributes)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return nil
	}
	return err
}

func preparePrivateSecretFile(path, userSID string) error {
	if err := requirePlainWindowsFile(path); err != nil {
		return err
	}
	if err := setPrivateSecretACL(path, false, userSID); err != nil {
		return errors.New("could not restrict secret file access")
	}
	if err := validatePrivateSecretACL(path, false, userSID); err != nil {
		return errors.New("secret file permissions are unsafe")
	}
	return nil
}

func setPrivateSecretACL(path string, directory bool, userSID string) error {
	security, err := privateSecretSecurityDescriptor(directory, userSID)
	if err != nil {
		return err
	}
	dacl, _, err := security.DACL()
	if err != nil || dacl == nil {
		return errors.New("secret security descriptor has no DACL")
	}
	information := windows.DACL_SECURITY_INFORMATION | windows.PROTECTED_DACL_SECURITY_INFORMATION
	return windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.SECURITY_INFORMATION(information), nil, nil, dacl, nil)
}

func privateSecretSecurityDescriptor(directory bool, userSID string) (*windows.SECURITY_DESCRIPTOR, error) {
	user, err := windows.StringToSid(userSID)
	if err != nil || user == nil {
		return nil, errors.New("current process SID is invalid")
	}
	inherit := ""
	if directory {
		inherit = "OICI"
	}
	sddl := fmt.Sprintf("D:P(A;%s;FA;;;%s)(A;%s;FA;;;SY)(A;%s;FA;;;BA)", inherit, userSID, inherit, inherit)
	security, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return nil, errors.New("could not prepare private secret permissions")
	}
	return security, nil
}

func validatePrivateSecretACL(path string, directory bool, userSID string) error {
	security, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil || security == nil {
		return errors.New("secret security descriptor is unavailable")
	}
	control, _, err := security.Control()
	if err != nil || control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("secret path must use a protected DACL")
	}
	dacl, _, err := security.DACL()
	if err != nil || dacl == nil || dacl.AceCount != 3 {
		return errors.New("secret path must have exactly three access entries")
	}
	allowed := map[string]bool{userSID: false, "S-1-5-18": false, "S-1-5-32-544": false}
	wantFlags := uint8(0)
	if directory {
		wantFlags = uint8(windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE)
	}
	for index := uint32(0); index < uint32(dacl.AceCount); index++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &ace); err != nil || ace == nil || ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return errors.New("secret path has an unexpected access entry")
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart)).String()
		seen, ok := allowed[sid]
		if !ok || seen || ace.Mask != privateSecretAccessMask || ace.Header.AceFlags != wantFlags {
			return errors.New("secret path has an unexpected access entry")
		}
		allowed[sid] = true
	}
	for _, found := range allowed {
		if !found {
			return errors.New("secret path is missing a required access entry")
		}
	}
	return nil
}

func requirePlainWindowsDirectory(path string) error {
	if err := rejectReparsePath(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("secret path is not a regular directory")
	}
	return nil
}

func requirePlainWindowsFile(path string) error {
	if err := rejectReparsePath(path); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("secret path is not a regular file")
	}
	return nil
}

func rejectReparsePath(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	volume := filepath.VolumeName(absolute)
	if volume == "" {
		return errors.New("secret path has no volume")
	}
	remainder := strings.TrimLeft(absolute[len(volume):], `/\`)
	current := volume + string(filepath.Separator)
	for _, component := range strings.FieldsFunc(remainder, func(r rune) bool { return r == '/' || r == '\\' }) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("secret path contains a symbolic link")
		}
		pathPtr, err := windows.UTF16PtrFromString(current)
		if err != nil {
			return err
		}
		attributes, err := windows.GetFileAttributes(pathPtr)
		if err != nil || attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return errors.New("secret path contains an unsafe reparse point")
		}
	}
	return nil
}
