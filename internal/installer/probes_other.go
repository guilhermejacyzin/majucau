//go:build !windows

package installer

import "runtime"

// DefaultProbes is deliberately harmless off Windows. This lets CI compile
// and exercise the CLI without pretending that a non-Windows host is eligible.
func DefaultProbes() Probes {
	return Probes{
		Platform: func() (PlatformInfo, error) {
			return PlatformInfo{OS: runtime.GOOS, Architecture: runtime.GOARCH, Version: "unsupported", Supported: false}, nil
		},
		Admin:    func() (bool, error) { return false, nil },
		Reboot:   func() (bool, error) { return false, nil },
		DiskFree: func(string) (uint64, error) { return 0, nil },
		Path:     func(string) (PathObservation, error) { return PathObservation{LocationSafe: false}, nil },
		Port:     func(uint16) (bool, error) { return false, nil },
		WebView2: func() (WebViewObservation, error) { return WebViewObservation{}, nil },
	}
}
