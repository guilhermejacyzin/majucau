//go:build windows

package installer

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func DefaultProbes() Probes {
	return Probes{
		Platform: probePlatform,
		Admin:    probeAdministrator,
		Reboot:   probePendingReboot,
		DiskFree: probeDiskFree,
		Path:     probePath,
		Port:     probePort,
		WebView2: probeWebView2,
	}
}

func probePlatform() (PlatformInfo, error) {
	info := windows.RtlGetVersion()
	return PlatformInfo{
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
		Version: fmt.Sprintf("%d.%d.%d", info.MajorVersion, info.MinorVersion, info.BuildNumber),
		// Windows 10 21H2 is the minimum baseline; Windows 11 uses
		// the same major version with newer builds. ProductType 1 is a
		// workstation; Windows Server is outside the supported target.
		Supported: runtime.GOARCH == "amd64" && info.MajorVersion == 10 && info.BuildNumber >= 19044 && info.ProductType == 1,
	}, nil
}

func probeAdministrator() (bool, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false, err
	}
	defer token.Close()
	var elevation uint32
	var returned uint32
	err := windows.GetTokenInformation(token, windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &returned)
	if err != nil {
		return false, err
	}
	return elevation != 0, nil
}

func probePendingReboot() (bool, error) {
	keys := []struct {
		path  string
		value string
	}{
		{`SYSTEM\CurrentControlSet\Control\Session Manager`, "PendingFileRenameOperations"},
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\Component Based Servicing\RebootPending`, ""},
		{`SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired`, ""},
	}
	for _, candidate := range keys {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, candidate.path, registry.QUERY_VALUE)
		if err != nil {
			if err != registry.ErrNotExist {
				return false, err
			}
			continue
		}
		if candidate.value == "" {
			key.Close()
			return true, nil
		}
		_, _, err = key.GetStringsValue(candidate.value)
		key.Close()
		if err == nil {
			return true, nil
		}
		if err != registry.ErrNotExist {
			return false, err
		}
	}
	return false, nil
}

func probeDiskFree(path string) (uint64, error) {
	if path == "" {
		path = `C:\`
	}
	path = existingAncestor(path)
	var available, total, free uint64
	directory, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	if err := windows.GetDiskFreeSpaceEx(directory, &available, &total, &free); err != nil {
		return 0, err
	}
	return available, nil
}

func existingAncestor(path string) string {
	path = filepath.Clean(path)
	for {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return path
		}
		path = parent
	}
}

func probePath(path string) (PathObservation, error) {
	if path == "" {
		return PathObservation{}, nil
	}
	clean := filepath.Clean(path)
	locationSafe := safeLocation(clean)
	parent := existingAncestor(filepath.Dir(clean))
	parentInfo, parentErr := os.Stat(parent)
	parentExists := parentErr == nil && parentInfo.IsDir()
	parentAccessible := false
	if parentExists {
		if handle, err := os.Open(parent); err == nil {
			parentAccessible = handle.Close() == nil
		}
	}
	info, err := os.Stat(clean)
	if err == nil {
		handle, openErr := os.Open(clean)
		accessible := openErr == nil
		if handle != nil {
			accessible = handle.Close() == nil
		}
		return PathObservation{Exists: true, IsDirectory: info.IsDir(), ParentExists: parentExists, Accessible: accessible, LocationSafe: locationSafe}, nil
	}
	if os.IsNotExist(err) {
		return PathObservation{ParentExists: parentExists, Accessible: parentAccessible, LocationSafe: locationSafe}, nil
	}
	return PathObservation{Exists: true, ParentExists: parentExists, Accessible: false, LocationSafe: locationSafe}, nil
}

func safeLocation(path string) bool {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) || strings.HasPrefix(clean, `\\`) || strings.HasPrefix(clean, `//`) {
		return false
	}
	volume := filepath.VolumeName(clean)
	if volume == "" {
		return false
	}
	root, err := windows.UTF16PtrFromString(volume + `\`)
	if err != nil {
		return false
	}
	switch windows.GetDriveType(root) {
	case windows.DRIVE_REMOVABLE, windows.DRIVE_REMOTE, windows.DRIVE_CDROM, windows.DRIVE_RAMDISK:
		return false
	case windows.DRIVE_UNKNOWN:
		return false
	}
	normalized := strings.ToLower(strings.ReplaceAll(clean, "/", `\`))
	for _, segment := range strings.Split(normalized, `\`) {
		segment = strings.TrimSpace(segment)
		if segment == "dropbox" || strings.HasPrefix(segment, "onedrive") || strings.HasPrefix(segment, "sharepoint") {
			return false
		}
	}
	for candidate := clean; ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return false
			}
			if data, ok := info.Sys().(*windows.Win32FileAttributeData); ok && data.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
				return false
			}
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			break
		}
	}
	return true
}

func probePort(port uint16) (bool, error) {
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false, nil
	}
	return listener.Close() == nil, nil
}

func probeWebView2() (WebViewObservation, error) {
	const client = `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`
	locations := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\` + client},
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\` + client},
		{registry.CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\` + client},
	}
	for _, location := range locations {
		key, err := registry.OpenKey(location.root, location.path, registry.QUERY_VALUE)
		if err != nil {
			if err != registry.ErrNotExist {
				return WebViewObservation{}, err
			}
			continue
		}
		version, _, valueErr := key.GetStringValue("pv")
		key.Close()
		if valueErr == nil && safeVersion(version) {
			return WebViewObservation{Installed: true, Version: version}, nil
		}
		return WebViewObservation{Installed: true}, nil
	}
	return WebViewObservation{}, nil
}

func safeVersion(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && char != '.' && char != '-' {
			return false
		}
	}
	return true
}
