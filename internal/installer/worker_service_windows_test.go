//go:build windows

package installer

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestWorkerServiceFailureCodeContainsOnlyKnownStage(t *testing.T) {
	err := workerServiceFailure{stage: "PREPARE_PARAMETERS_ACL", cause: errors.New(`failure at C:\private\path`)}
	if got := WorkerServiceFailureCode(err); got != "WORKER_SERVICE_PREPARE_PARAMETERS_ACL" {
		t.Fatalf("unexpected stage code: %q", got)
	}
}

func TestServiceImageWithinRequiresQuotedImageBelowInstallRoot(t *testing.T) {
	root := filepath.Join(`C:\Program Files`, "Majucau Suite")
	image := filepath.Join(root, "worker", "v1", "majucau-worker.exe")
	valid := quoteWorkerServiceImage(image)
	if valid != `"`+image+`"` {
		t.Fatalf("worker service image was not quoted: %q", valid)
	}
	if !serviceImageWithin(valid, root) {
		t.Fatal("expected worker image below install root to be accepted")
	}
	for _, invalid := range []string{
		`"C:\Program Files\Majucau SuiteEvil\majucau-worker.exe"`,
		`"C:\Program Files\Other Product\majucau-worker.exe"`,
		`C:\Program Files\Majucau Suite\..\Other Product\worker.exe`,
		`"C:\Program Files\Majucau Suite\worker.exe" --unsafe-argument`,
	} {
		if serviceImageWithin(invalid, root) {
			t.Fatalf("expected unsafe image to be rejected: %q", invalid)
		}
	}
}

func TestWorkerServiceImageMustBeVersionedWorkerExecutable(t *testing.T) {
	root := filepath.Join(`C:\Program Files`, "Majucau Suite")
	valid := quoteWorkerServiceImage(filepath.Join(root, "worker", "v1", "majucau-worker.exe"))
	if !workerServiceImageWithin(valid, root) {
		t.Fatalf("expected versioned worker image to be accepted: %q", valid)
	}
	for _, invalid := range []string{
		quoteWorkerServiceImage(filepath.Join(root, "Majucau Financial Intelligence.exe")),
		quoteWorkerServiceImage(filepath.Join(root, "worker", "v1", "installer-helper.exe")),
		quoteWorkerServiceImage(filepath.Join(root, "worker-malicious", "v1", "majucau-worker.exe")),
	} {
		if workerServiceImageWithin(invalid, root) {
			t.Fatalf("expected non-worker or non-versioned path to be rejected: %q", invalid)
		}
	}
}

func TestWorkerExecutablePathAllowsProgramFilesSpaces(t *testing.T) {
	root := filepath.Join(`C:\Program Files`, "Majucau Suite")
	worker := filepath.Join(root, "worker", "test", "majucau-worker.exe")
	if !workerExecutableWithin(worker, root) {
		t.Fatalf("expected an unquoted worker filesystem path with spaces in its parent to be accepted: %q", worker)
	}
	for _, invalid := range []string{
		filepath.Join(root, "majucau-worker.exe"),
		filepath.Join(root+"Evil", "worker", "majucau-worker.exe"),
		filepath.Join(`C:\Program Files\Other Product`, "worker", "majucau-worker.exe"),
		filepath.Join(root, "worker", "installer-helper.exe"),
	} {
		if workerExecutableWithin(invalid, root) {
			t.Fatalf("expected invalid worker filesystem path to be rejected: %q", invalid)
		}
	}
}

func TestServicePrivateDirectoryUsesProtectedMinimalACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "WorkerSecrets")
	serviceSID := "S-1-5-80-0"
	if err := createServicePrivateDirectory(path, serviceSID); err != nil {
		t.Fatalf("create protected worker directory: %v", err)
	}
	if err := validateServicePrivateDirectory(path, serviceSID); err != nil {
		t.Fatalf("validate protected worker directory: %v", err)
	}
	if err := ensureServicePrivateDirectory(path, serviceSID); err != nil {
		t.Fatalf("accept existing directory with approved ACL: %v", err)
	}
	if err := ensureServicePrivateDirectory(path, "S-1-5-80-1"); err == nil {
		t.Fatal("accepted existing secret directory bound to a different service SID")
	}
}
