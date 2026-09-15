// Package installer contains the read-only checks used before installing
// Majucau on a Windows workstation. It deliberately has no installation
// side-effects and keeps machine-specific values out of its public result.
package installer

import "sort"

const (
	SchemaVersion  = "1.0"
	StatusReady    = "READY"
	StatusBlocked  = "BLOCKED"
	StatusInternal = "INTERNAL_ERROR"

	ExitReady    = 0
	ExitBlocked  = 2
	ExitInternal = 3

	// DefaultMinimumFreeBytes is a fresh-install safety floor. Upgrade and
	// restore flows must raise it using current data, backup and rollback size.
	DefaultMinimumFreeBytes = 4 << 30
)

// Config controls a preflight run. Paths are consumed by probes but are never
// emitted in the result, preventing accidental disclosure of user names.
type Config struct {
	MinimumFreeBytes uint64
	PreferredPort    uint16
	FreeSpacePath    string
	Paths            []PathSpec
}

type PathSpec struct {
	Label      string
	Path       string
	ParentPath string
}

type PlatformInfo struct {
	OS           string
	Architecture string
	Version      string
	Supported    bool
}

type PathObservation struct {
	Exists       bool
	IsDirectory  bool
	ParentExists bool
	Accessible   bool
}

type WebViewObservation struct {
	Installed bool
	Version   string
}

// Probes is the seam between policy and operating-system discovery. Tests can
// inject deterministic implementations; production receives DefaultProbes.
type Probes struct {
	Platform func() (PlatformInfo, error)
	Admin    func() (bool, error)
	Reboot   func() (bool, error)
	DiskFree func(path string) (uint64, error)
	Path     func(path string) (PathObservation, error)
	Port     func(port uint16) (bool, error)
	WebView2 func() (WebViewObservation, error)
}

type Result struct {
	SchemaVersion string  `json:"schema_version"`
	Status        string  `json:"status"`
	ExitCode      int     `json:"exit_code"`
	Checks        Checks  `json:"checks"`
	Issues        []Issue `json:"issues"`
}

type Checks struct {
	Platform      PlatformCheck `json:"platform"`
	Administrator BoolCheck     `json:"administrator"`
	PendingReboot BoolCheck     `json:"pending_reboot"`
	DiskSpace     DiskCheck     `json:"disk_space"`
	Paths         []PathCheck   `json:"paths"`
	PreferredPort PortCheck     `json:"preferred_port"`
	WebView2      WebViewCheck  `json:"webview2"`
}

type PlatformCheck struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Version      string `json:"version"`
	Supported    bool   `json:"supported"`
}

type BoolCheck struct {
	Passed bool `json:"passed"`
}

type DiskCheck struct {
	Available        bool   `json:"available"`
	FreeBytes        uint64 `json:"free_bytes"`
	MinimumFreeBytes uint64 `json:"minimum_free_bytes"`
	Passed           bool   `json:"passed"`
}

type PathCheck struct {
	Label        string `json:"label"`
	Exists       bool   `json:"exists"`
	IsDirectory  bool   `json:"is_directory"`
	ParentExists bool   `json:"parent_exists"`
	Accessible   bool   `json:"accessible"`
	Passed       bool   `json:"passed"`
}

type PortCheck struct {
	Port      uint16 `json:"port"`
	Available bool   `json:"available"`
	Passed    bool   `json:"passed"`
}

type WebViewCheck struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Passed    bool   `json:"passed"`
}

type Issue struct {
	Code  string `json:"code"`
	Check string `json:"check"`
}

// DefaultConfig is intentionally conservative for a bundled PostgreSQL and
// application install. Both values can be overridden by the CLI.
func DefaultConfig(installPath, dataPath string) Config {
	return Config{
		MinimumFreeBytes: DefaultMinimumFreeBytes,
		PreferredPort:    54329,
		FreeSpacePath:    dataPath,
		Paths: []PathSpec{
			{Label: "install_dir", Path: installPath},
			{Label: "data_dir", Path: dataPath},
		},
	}
}

func (c Config) normalize() Config {
	if c.MinimumFreeBytes == 0 {
		c.MinimumFreeBytes = DefaultMinimumFreeBytes
	}
	if c.PreferredPort == 0 {
		c.PreferredPort = 54329
	}
	if c.FreeSpacePath == "" && len(c.Paths) > 0 {
		c.FreeSpacePath = c.Paths[0].Path
	}
	c.Paths = append([]PathSpec(nil), c.Paths...)
	sort.Slice(c.Paths, func(i, j int) bool { return c.Paths[i].Label < c.Paths[j].Label })
	return c
}
