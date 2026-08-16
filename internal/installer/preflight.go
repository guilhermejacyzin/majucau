package installer

import "errors"

var errMissingProbe = errors.New("installer preflight probe is not configured")

// Evaluate applies installation policy to observations. It never invokes a
// mutating operation; all filesystem and network checks are delegated to the
// read-only probes.
func Evaluate(config Config, probes Probes) Result {
	config = config.normalize()
	result := Result{
		SchemaVersion: SchemaVersion,
		Status:        StatusReady,
		ExitCode:      ExitReady,
		Issues:        []Issue{},
		Checks: Checks{
			DiskSpace: DiskCheck{MinimumFreeBytes: config.MinimumFreeBytes},
			Paths:     []PathCheck{},
		},
	}

	platform, err := callPlatform(probes)
	if err != nil {
		return internalResult(result, "platform")
	}
	result.Checks.Platform = PlatformCheck{
		OS: platform.OS, Architecture: platform.Architecture,
		Version: platform.Version, Supported: platform.Supported,
	}
	if !platform.Supported {
		result.Issues = append(result.Issues, Issue{Code: "OS_UNSUPPORTED", Check: "platform"})
	}

	admin, err := callAdmin(probes)
	if err != nil {
		return internalResult(result, "administrator")
	}
	result.Checks.Administrator = BoolCheck{Passed: admin}
	if !admin {
		result.Issues = append(result.Issues, Issue{Code: "INSTALL_NOT_ELEVATED", Check: "administrator"})
	}

	pending, err := callReboot(probes)
	if err != nil {
		return internalResult(result, "pending_reboot")
	}
	result.Checks.PendingReboot = BoolCheck{Passed: !pending}
	if pending {
		result.Issues = append(result.Issues, Issue{Code: "REBOOT_PENDING", Check: "pending_reboot"})
	}

	free, err := callDisk(probes, config.FreeSpacePath)
	if err != nil {
		return internalResult(result, "disk_space")
	}
	result.Checks.DiskSpace = DiskCheck{
		Available: true, FreeBytes: free, MinimumFreeBytes: config.MinimumFreeBytes,
		Passed: free >= config.MinimumFreeBytes,
	}
	if free < config.MinimumFreeBytes {
		result.Issues = append(result.Issues, Issue{Code: "DISK_SPACE_LOW", Check: "disk_space"})
	}

	for _, spec := range config.Paths {
		observation, probeErr := callPath(probes, spec.Path)
		if probeErr != nil {
			return internalResult(result, "paths")
		}
		check := PathCheck{
			Label: spec.Label, Exists: observation.Exists,
			IsDirectory: observation.IsDirectory, ParentExists: observation.ParentExists,
			Accessible: observation.Accessible,
		}
		// A destination may not exist yet, but it must be creatable in an
		// existing accessible parent. Existing destinations must be directories.
		check.Passed = observation.ParentExists && observation.Accessible &&
			(!observation.Exists || observation.IsDirectory)
		result.Checks.Paths = append(result.Checks.Paths, check)
		if !check.Passed {
			result.Issues = append(result.Issues, Issue{Code: "PATH_UNAVAILABLE", Check: spec.Label})
		}
	}

	available, err := callPort(probes, config.PreferredPort)
	if err != nil {
		return internalResult(result, "preferred_port")
	}
	result.Checks.PreferredPort = PortCheck{Port: config.PreferredPort, Available: available, Passed: available}
	if !available {
		result.Issues = append(result.Issues, Issue{Code: "DB_PORT_CONFLICT", Check: "preferred_port"})
	}

	webview, err := callWebView(probes)
	if err != nil {
		return internalResult(result, "webview2")
	}
	result.Checks.WebView2 = WebViewCheck{
		Installed: webview.Installed, Version: webview.Version, Passed: webview.Installed,
	}
	if !webview.Installed {
		result.Issues = append(result.Issues, Issue{Code: "WEBVIEW2_MISSING", Check: "webview2"})
	}

	if len(result.Issues) > 0 {
		result.Status, result.ExitCode = StatusBlocked, ExitBlocked
	}
	return result
}

func internalResult(result Result, check string) Result {
	result.Status, result.ExitCode = StatusInternal, ExitInternal
	result.Issues = append(result.Issues, Issue{Code: "PROBE_FAILED", Check: check})
	return result
}

func callPlatform(p Probes) (PlatformInfo, error) {
	if p.Platform == nil {
		return PlatformInfo{}, errMissingProbe
	}
	return p.Platform()
}
func callAdmin(p Probes) (bool, error) {
	if p.Admin == nil {
		return false, errMissingProbe
	}
	return p.Admin()
}
func callReboot(p Probes) (bool, error) {
	if p.Reboot == nil {
		return false, errMissingProbe
	}
	return p.Reboot()
}
func callDisk(p Probes, path string) (uint64, error) {
	if p.DiskFree == nil {
		return 0, errMissingProbe
	}
	return p.DiskFree(path)
}
func callPath(p Probes, path string) (PathObservation, error) {
	if p.Path == nil {
		return PathObservation{}, errMissingProbe
	}
	return p.Path(path)
}
func callPort(p Probes, port uint16) (bool, error) {
	if p.Port == nil {
		return false, errMissingProbe
	}
	return p.Port(port)
}
func callWebView(p Probes) (WebViewObservation, error) {
	if p.WebView2 == nil {
		return WebViewObservation{}, errMissingProbe
	}
	return p.WebView2()
}
