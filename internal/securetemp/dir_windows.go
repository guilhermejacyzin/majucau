//go:build windows

package securetemp

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type privateDirectoryAPI struct {
	convert func(*uint16, **windows.SECURITY_DESCRIPTOR, *uint32) (uintptr, error)
	free    func(uintptr) (uintptr, error)
	create  func(*uint16, *windows.SecurityAttributes) (uintptr, error)
	remove  func(string) error
}

// NewDir creates a temporary directory with a protected DACL. The owner,
// LocalSystem, and administrators can access the directory and its contents.
func NewDir(parent, prefix string) (string, error) {
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
	return newDirWithAPI(parent, prefix, privateDirectoryAPI{
		convert: func(sddl *uint16, descriptor **windows.SECURITY_DESCRIPTOR, size *uint32) (uintptr, error) {
			result, _, callErr := convert.Call(
				uintptr(unsafe.Pointer(sddl)), 1,
				uintptr(unsafe.Pointer(descriptor)), uintptr(unsafe.Pointer(size)),
			)
			return result, callErr
		},
		free: func(address uintptr) (uintptr, error) {
			result, _, callErr := localFree.Call(address)
			return result, callErr
		},
		create: func(path *uint16, attributes *windows.SecurityAttributes) (uintptr, error) {
			result, _, callErr := createDir.Call(uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(attributes)))
			return result, callErr
		},
		remove: os.Remove,
	})
}

func newDirWithAPI(parent, prefix string, api privateDirectoryAPI) (createdPath string, returnErr error) {
	if !filepath.IsAbs(parent) {
		return "", fmt.Errorf("private temporary directory parent must be absolute")
	}
	if api.convert == nil || api.free == nil || api.create == nil || api.remove == nil {
		return "", errors.New("private temporary directory API is incomplete")
	}
	sddl, err := syscall.UTF16PtrFromString("D:P(A;OICI;FA;;;OW)(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)")
	if err != nil {
		return "", err
	}
	var descriptor *windows.SECURITY_DESCRIPTOR
	var descriptorSize uint32
	result, callErr := api.convert(sddl, &descriptor, &descriptorSize)
	if result == 0 {
		if callErr == nil || errors.Is(callErr, syscall.Errno(0)) {
			callErr = errors.New("Windows returned failure without an error code")
		}
		return "", fmt.Errorf("create private temporary directory security descriptor: %w", callErr)
	}
	if descriptor == nil {
		return "", errors.New("Windows returned no temporary-directory security descriptor")
	}
	defer func() {
		freed, freeErr := api.free(uintptr(unsafe.Pointer(descriptor)))
		if freed == 0 {
			return
		}
		if freeErr == nil || errors.Is(freeErr, syscall.Errno(0)) {
			freeErr = errors.New("LocalFree returned the allocated security descriptor")
		}
		cleanupErr := fmt.Errorf("release temporary-directory security descriptor: %w", freeErr)
		if createdPath != "" {
			if err := api.remove(createdPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				cleanupErr = errors.Join(cleanupErr, &CleanupError{Path: createdPath, Cause: err})
			}
			createdPath = ""
		}
		returnErr = errors.Join(returnErr, cleanupErr)
	}()

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
		created, createErr := api.create(pathPtr, &attributes)
		if created != 0 {
			return path, nil
		}
		if createErr != syscall.ERROR_ALREADY_EXISTS {
			if createErr == nil || errors.Is(createErr, syscall.Errno(0)) {
				createErr = errors.New("Windows returned failure without an error code")
			}
			return "", fmt.Errorf("create private temporary directory: %w", createErr)
		}
	}
	return "", fmt.Errorf("could not allocate a unique private temporary directory")
}
