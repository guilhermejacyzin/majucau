package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"majucau.local/financial-intelligence/internal/installer"
)

func TestRunInvalidCommandUsesStableInternalContract(t *testing.T) {
	var output bytes.Buffer
	code := run([]string{"unknown"}, &output)
	if code != installer.ExitInternal {
		t.Fatalf("expected %d, got %d", installer.ExitInternal, code)
	}
	var result installer.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v; output=%q", err, output.String())
	}
	if result.Status != installer.StatusInternal || len(result.Issues) != 1 || result.Issues[0].Code != "INVALID_ARGUMENT" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestRunPreflightProducesContractWithoutPaths(t *testing.T) {
	var output bytes.Buffer
	// The maximum uint64 makes the disk gate fail on every realistic machine,
	// regardless of whether this test is running elevated or not.
	code := run([]string{"preflight", "--min-free-bytes=18446744073709551615"}, &output)
	if code != installer.ExitBlocked {
		t.Fatalf("expected blocked, got %d: %s", code, output.String())
	}
	var result installer.Result
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != installer.StatusBlocked {
		t.Fatalf("unexpected status: %#v", result)
	}
}

func TestRunDiagnosticsWritesSanitizedBundleWhenBlocked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "majucau-diagnostic.zip")
	var output bytes.Buffer
	code := run([]string{"diagnostics", "--output", path, "--min-free-bytes=18446744073709551615"}, &output)
	if code != installer.ExitBlocked {
		t.Fatalf("expected blocked preflight with bundle, got %d: %s", code, output.String())
	}
	var response diagnosticResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("invalid diagnostics JSON: %v; output=%q", err, output.String())
	}
	if response.Status != installer.StatusBlocked || response.Bundle.Bytes <= 0 || len(response.Bundle.SHA256) != 64 {
		t.Fatalf("unexpected diagnostics response: %#v", response)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("bundle was not created: %v", err)
	}
}

func TestRunVerifyPackageReturnsVerifiedContract(t *testing.T) {
	root := t.TempDir()
	payload := []byte("desktop-binary")
	if err := os.WriteFile(filepath.Join(root, "majucau.exe"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := installer.ReleaseManifest{
		ManifestVersion: installer.ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "1.0",
		MinSchema:       "1.0",
		MaxSchema:       "1.0",
		Artifacts: []installer.ManifestFile{{
			Path:   "majucau.exe",
			SHA256: sha256Hex(payload),
		}},
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "release-manifest.json"), encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	code := run([]string{"verify-package", "--package-dir", root, "--current-schema", "1.0"}, &output)
	if code != installer.ExitReady {
		t.Fatalf("expected package verified, got %d: %s", code, output.String())
	}
	var response manifestResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "PACKAGE_VERIFIED" || !response.Valid || len(response.Issues) != 0 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestRunVerifyPackageBlocksSchemaMismatchWithoutAbsolutePath(t *testing.T) {
	root := t.TempDir()
	payload := []byte("desktop-binary")
	if err := os.WriteFile(filepath.Join(root, "majucau.exe"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	manifest := installer.ReleaseManifest{
		ManifestVersion: installer.ReleaseManifestVersion,
		AppVersion:      "0.1.0",
		SchemaVersion:   "2.0",
		MinSchema:       "2.0",
		MaxSchema:       "3.0",
		Artifacts:       []installer.ManifestFile{{Path: "../secret.json", SHA256: sha256Hex(payload)}},
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "release-manifest.json"), encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	code := run([]string{"verify-package", "--package-dir", root, "--current-schema", "1.0"}, &output)
	if code != installer.ExitBlocked {
		t.Fatalf("expected package blocked, got %d: %s", code, output.String())
	}
	var response manifestResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "PACKAGE_INVALID" || response.Valid || len(response.Issues) < 1 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if bytes.Contains(output.Bytes(), []byte(root)) {
		t.Fatalf("response leaked absolute package path: %s", output.String())
	}
}

func sha256Hex(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
