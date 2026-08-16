package installer

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func passingProbes() Probes {
	return Probes{
		Platform: func() (PlatformInfo, error) {
			return PlatformInfo{OS: "windows", Architecture: "amd64", Version: "10.0.26100", Supported: true}, nil
		},
		Admin:    func() (bool, error) { return true, nil },
		Reboot:   func() (bool, error) { return false, nil },
		DiskFree: func(string) (uint64, error) { return 2 << 30, nil },
		Path: func(string) (PathObservation, error) {
			return PathObservation{Exists: false, ParentExists: true, Accessible: true}, nil
		},
		Port: func(uint16) (bool, error) { return true, nil },
		WebView2: func() (WebViewObservation, error) {
			return WebViewObservation{Installed: true, Version: "128.0.1"}, nil
		},
	}
}

func testConfig() Config {
	return Config{
		MinimumFreeBytes: 1 << 30,
		PreferredPort:    54329,
		FreeSpacePath:    `C:\ProgramData\Majucau`,
		Paths: []PathSpec{
			{Label: "data_dir", Path: `C:\ProgramData\Majucau`},
			{Label: "install_dir", Path: `C:\Program Files\Majucau`},
		},
	}
}

func TestEvaluateReady(t *testing.T) {
	result := Evaluate(testConfig(), passingProbes())
	if result.Status != StatusReady || result.ExitCode != ExitReady {
		t.Fatalf("expected ready, got %#v", result)
	}
	if len(result.Issues) != 0 || len(result.Checks.Paths) != 2 {
		t.Fatalf("unexpected ready details: %#v", result)
	}
	if result.Checks.Paths[0].Label != "data_dir" {
		t.Fatalf("paths must be deterministic and sorted: %#v", result.Checks.Paths)
	}
}

func TestEvaluateBlockedBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Probes)
		code   string
	}{
		{"not-admin", func(p *Probes) { p.Admin = func() (bool, error) { return false, nil } }, "INSTALL_NOT_ELEVATED"},
		{"reboot", func(p *Probes) { p.Reboot = func() (bool, error) { return true, nil } }, "REBOOT_PENDING"},
		{"disk-boundary", func(p *Probes) { p.DiskFree = func(string) (uint64, error) { return (1 << 30) - 1, nil } }, "DISK_SPACE_LOW"},
		{"path-file", func(p *Probes) {
			p.Path = func(string) (PathObservation, error) {
				return PathObservation{Exists: true, IsDirectory: false, ParentExists: true, Accessible: true}, nil
			}
		}, "PATH_UNAVAILABLE"},
		{"port", func(p *Probes) { p.Port = func(uint16) (bool, error) { return false, nil } }, "DB_PORT_CONFLICT"},
		{"webview", func(p *Probes) { p.WebView2 = func() (WebViewObservation, error) { return WebViewObservation{}, nil } }, "WEBVIEW2_MISSING"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			probes := passingProbes()
			test.mutate(&probes)
			result := Evaluate(testConfig(), probes)
			if result.Status != StatusBlocked || result.ExitCode != ExitBlocked {
				t.Fatalf("expected blocked, got %#v", result)
			}
			if test.code != "" && !hasIssue(result, test.code) {
				t.Fatalf("missing issue %q: %#v", test.code, result.Issues)
			}
		})
	}
}

func TestEvaluateDiskMinimumBoundaryIsReady(t *testing.T) {
	probes := passingProbes()
	probes.DiskFree = func(string) (uint64, error) { return testConfig().MinimumFreeBytes, nil }
	if result := Evaluate(testConfig(), probes); result.Status != StatusReady {
		t.Fatalf("minimum free bytes should be inclusive: %#v", result)
	}
}

func TestEvaluateProbeFailureIsInternal(t *testing.T) {
	probes := passingProbes()
	probes.Platform = func() (PlatformInfo, error) { return PlatformInfo{}, errors.New("machine detail must not escape") }
	result := Evaluate(testConfig(), probes)
	if result.Status != StatusInternal || result.ExitCode != ExitInternal {
		t.Fatalf("expected internal error, got %#v", result)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "machine detail") || strings.Contains(string(encoded), "ProgramData") {
		t.Fatalf("result leaked probe details or path: %s", encoded)
	}
}

func TestEvaluateUnsupportedPlatformIsBlocked(t *testing.T) {
	probes := passingProbes()
	probes.Platform = func() (PlatformInfo, error) {
		return PlatformInfo{OS: "linux", Architecture: "amd64", Version: "6", Supported: false}, nil
	}
	result := Evaluate(testConfig(), probes)
	if result.Status != StatusBlocked || !hasIssue(result, "OS_UNSUPPORTED") {
		t.Fatalf("expected unsupported platform block: %#v", result)
	}
}

func hasIssue(result Result, code string) bool {
	for _, issue := range result.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
