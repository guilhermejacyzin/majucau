// Command installer-helper exposes read-only checks used by the Windows
// installer. It is intentionally independent from the installer engine.
package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"majucau.local/financial-intelligence/internal/installer"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, output io.Writer) int {
	if len(args) == 0 {
		return writeResult(output, invalidArgumentResult())
	}
	if args[0] == "diagnostics" {
		return runDiagnostics(args[1:], output)
	}
	if args[0] == "verify-package" {
		return runVerifyPackage(args[1:], output)
	}
	if args[0] != "preflight" {
		return writeResult(output, invalidArgumentResult())
	}

	executable, err := os.Executable()
	if err != nil {
		return writeResult(output, invalidArgumentResult())
	}
	installDefault := filepath.Dir(executable)
	dataRoot := os.Getenv("ProgramData")
	if dataRoot == "" {
		dataRoot = installDefault
	}
	dataDefault := filepath.Join(dataRoot, "Majucau")

	flags := flag.NewFlagSet("preflight", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	installPath := flags.String("install-dir", installDefault, "installation directory")
	dataPath := flags.String("data-dir", dataDefault, "application data directory")
	freePath := flags.String("free-space-path", dataRoot, "volume used for free-space check")
	minimumFree := flags.Uint64("min-free-bytes", installer.DefaultMinimumFreeBytes, "minimum free bytes")
	port := flags.Uint("port", 54329, "preferred local port")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *port == 0 || *port > 65535 {
		return writeResult(output, invalidArgumentResult())
	}

	config := installer.DefaultConfig(*installPath, *dataPath)
	config.MinimumFreeBytes = *minimumFree
	config.PreferredPort = uint16(*port)
	config.FreeSpacePath = *freePath
	return writeResult(output, installer.Evaluate(config, installer.DefaultProbes()))
}

type manifestResponse struct {
	SchemaVersion string                     `json:"schema_version"`
	Status        string                     `json:"status"`
	ExitCode      int                        `json:"exit_code"`
	Valid         bool                       `json:"valid"`
	Issues        []installer.ManifestIssue  `json:"issues,omitempty"`
	Manifest      installer.ReleaseManifest  `json:"manifest"`
}

func runVerifyPackage(args []string, output io.Writer) int {
	flags := flag.NewFlagSet("verify-package", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	packageDir := flags.String("package-dir", "", "absolute package directory")
	manifestPath := flags.String("manifest", "release-manifest.json", "manifest path relative to the package")
	currentSchema := flags.String("current-schema", "1.0", "currently installed database schema version")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || strings.TrimSpace(*packageDir) == "" || !filepath.IsAbs(*packageDir) || strings.TrimSpace(*currentSchema) == "" {
		return writeManifestResponse(output, manifestResponse{
			SchemaVersion: installer.ManifestValidationSchemaVersion,
			Status:        "PACKAGE_INVALID",
			ExitCode:      installer.ExitInternal,
			Issues:        []installer.ManifestIssue{{Code: installer.ManifestIssuePackageInvalid, Detail: "invalid verify-package arguments"}},
		})
	}
	validation := installer.ValidateReleaseManifest(*packageDir, *manifestPath, *currentSchema)
	response := manifestResponse{
		SchemaVersion: installer.ManifestValidationSchemaVersion,
		Status:        "PACKAGE_VERIFIED",
		ExitCode:      installer.ExitReady,
		Valid:         validation.Valid,
		Issues:        validation.Issues,
		Manifest:      validation.Manifest,
	}
	if !validation.Valid {
		response.Status = "PACKAGE_INVALID"
		response.ExitCode = installer.ExitBlocked
	}
	return writeManifestResponse(output, response)
}

func writeManifestResponse(output io.Writer, response manifestResponse) int {
	if err := json.NewEncoder(output).Encode(response); err != nil {
		return installer.ExitInternal
	}
	return response.ExitCode
}

type diagnosticResponse struct {
	SchemaVersion string                         `json:"schema_version"`
	Status        string                         `json:"status"`
	ExitCode      int                            `json:"exit_code"`
	Issues        []installer.Issue              `json:"issues"`
	Bundle        installer.DiagnosticBundleInfo `json:"bundle"`
}

func runDiagnostics(args []string, output io.Writer) int {
	executable, err := os.Executable()
	if err != nil {
		return writeResult(output, invalidArgumentResult())
	}
	installDefault := filepath.Dir(executable)
	dataRoot := os.Getenv("ProgramData")
	if dataRoot == "" {
		dataRoot = installDefault
	}
	dataDefault := filepath.Join(dataRoot, "Majucau")

	flags := flag.NewFlagSet("diagnostics", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	outputPath := flags.String("output", "", "absolute path for the sanitized ZIP bundle")
	installPath := flags.String("install-dir", installDefault, "installation directory")
	dataPath := flags.String("data-dir", dataDefault, "application data directory")
	freePath := flags.String("free-space-path", dataRoot, "volume used for free-space check")
	minimumFree := flags.Uint64("min-free-bytes", installer.DefaultMinimumFreeBytes, "minimum free bytes")
	port := flags.Uint("port", 54329, "preferred local port")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || strings.TrimSpace(*outputPath) == "" || *port == 0 || *port > 65535 {
		return writeResult(output, invalidArgumentResult())
	}
	config := installer.DefaultConfig(*installPath, *dataPath)
	config.MinimumFreeBytes = *minimumFree
	config.PreferredPort = uint16(*port)
	config.FreeSpacePath = *freePath
	result := installer.Evaluate(config, installer.DefaultProbes())
	bundle, err := installer.WriteDiagnosticBundle(*outputPath, result, installer.DiagnosticMetadata{AppVersion: "installer-helper", GeneratedAt: time.Now().UTC()})
	if err != nil {
		return writeResult(output, installer.Result{
			SchemaVersion: installer.SchemaVersion,
			Status:        installer.StatusInternal,
			ExitCode:      installer.ExitInternal,
			Checks:        installer.Checks{Paths: []installer.PathCheck{}},
			Issues:        []installer.Issue{{Code: "DIAGNOSTIC_WRITE_FAILED", Check: "diagnostics"}},
		})
	}
	response := diagnosticResponse{SchemaVersion: installer.SchemaVersion, Status: result.Status, ExitCode: result.ExitCode, Issues: result.Issues, Bundle: bundle}
	if err := json.NewEncoder(output).Encode(response); err != nil {
		return installer.ExitInternal
	}
	return result.ExitCode
}

func writeResult(output io.Writer, result installer.Result) int {
	if err := json.NewEncoder(output).Encode(result); err != nil {
		return installer.ExitInternal
	}
	return result.ExitCode
}

func invalidArgumentResult() installer.Result {
	return installer.Result{
		SchemaVersion: installer.SchemaVersion,
		Status:        installer.StatusInternal,
		ExitCode:      installer.ExitInternal,
		Checks:        installer.Checks{Paths: []installer.PathCheck{}},
		Issues:        []installer.Issue{{Code: "INVALID_ARGUMENT", Check: "cli"}},
	}
}
