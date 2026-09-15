// Command installer-helper exposes read-only checks used by the Windows
// installer. It is intentionally independent from the installer engine.
package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"

	"majucau.local/financial-intelligence/internal/installer"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

func run(args []string, output io.Writer) int {
	if len(args) == 0 || args[0] != "preflight" {
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
