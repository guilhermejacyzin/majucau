package backup

import (
	"errors"
	"testing"
)

type closeFailureStub struct {
	err error
}

func (stub closeFailureStub) Close() error {
	return stub.err
}

func TestCloseWithFailureSanitizesMessageAndPreservesCause(t *testing.T) {
	cause := errors.New(`access denied for C:\private\financial-backup.mjbk`)
	err := closeWithFailure(nil, "backup package", closeFailureStub{err: cause})
	if err == nil {
		t.Fatal("closeWithFailure() returned nil")
	}
	if !errors.Is(err, cause) {
		t.Fatal("close cause is not available to internal error inspection")
	}
	if got, want := err.Error(), "backup package close failed"; got != want {
		t.Fatalf("public error = %q, want sanitized %q", got, want)
	}
}

func TestCloseWithFailurePreservesOriginalAndCloseErrors(t *testing.T) {
	original := errors.New("payload read failed")
	closeCause := errors.New("close failed")
	err := closeWithFailure(original, "backup archive entry", closeFailureStub{err: closeCause})
	if !errors.Is(err, original) || !errors.Is(err, closeCause) {
		t.Fatal("combined error does not preserve both causes")
	}
	if got, want := err.Error(), "payload read failed\nbackup archive entry close failed"; got != want {
		t.Fatalf("combined error = %q, want %q", got, want)
	}
}

func TestCloseWithFailureLeavesOriginalWhenCloseSucceeds(t *testing.T) {
	original := errors.New("payload read failed")
	if got := closeWithFailure(original, "backup archive entry", closeFailureStub{}); got != original {
		t.Fatalf("closeWithFailure() = %v, want original error", got)
	}
}
