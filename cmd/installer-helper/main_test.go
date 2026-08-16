package main

import (
	"bytes"
	"encoding/json"
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
