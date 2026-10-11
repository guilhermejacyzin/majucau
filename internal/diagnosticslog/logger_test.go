//go:build !windows

package diagnosticslog

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRotatingWriterRejectsInvalidCaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxBytes int64
		maxFiles int
	}{
		{name: "zero bytes", maxBytes: 0, maxFiles: 2},
		{name: "negative bytes", maxBytes: -1, maxFiles: 2},
		{name: "zero files", maxBytes: 8, maxFiles: 0},
		{name: "one file cannot rotate", maxBytes: 8, maxFiles: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newRotatingWriter(t.TempDir(), test.maxBytes, test.maxFiles)
			if !errors.Is(err, ErrInvalidRotationCaps) {
				t.Fatalf("newRotatingWriter() error = %v, want %v", err, ErrInvalidRotationCaps)
			}
		})
	}
}

func TestRotatingWriterRejectsRecordLargerThanFileCap(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writer, err := newRotatingWriter(dir, 8, 3)
	if err != nil {
		t.Fatalf("newRotatingWriter() error = %v", err)
	}
	t.Cleanup(func() { _ = writer.Close() })

	if n, err := writer.Write([]byte("123456789")); n != 0 || !errors.Is(err, ErrRecordTooLarge) {
		t.Fatalf("Write() = (%d, %v), want (0, %v)", n, err, ErrRecordTooLarge)
	}

	info, err := os.Stat(filepath.Join(dir, "majucau-worker.jsonl"))
	if err != nil {
		t.Fatalf("stat current log: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("oversized record wrote %d bytes, want 0", info.Size())
	}
}

func TestRotatingWriterRejectsRecordLargerThanEncodedRecordLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writer, err := newRotatingWriter(dir, maxEncodedRecord*2, 3)
	if err != nil {
		t.Fatalf("newRotatingWriter() error = %v", err)
	}
	t.Cleanup(func() { _ = writer.Close() })

	if n, err := writer.Write(make([]byte, maxEncodedRecord+1)); n != 0 || !errors.Is(err, ErrRecordTooLarge) {
		t.Fatalf("Write() = (%d, %v), want (0, %v)", n, err, ErrRecordTooLarge)
	}

	info, err := os.Stat(filepath.Join(dir, "majucau-worker.jsonl"))
	if err != nil {
		t.Fatalf("stat current log: %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("oversized record wrote %d bytes, want 0", info.Size())
	}
}

func TestRotatingWriterKeepsOnlyConfiguredNewestFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writer, err := newRotatingWriter(dir, 8, 3)
	if err != nil {
		t.Fatalf("newRotatingWriter() error = %v", err)
	}
	defer writer.Close()

	for _, record := range []string{"record01", "record02", "record03", "record04"} {
		if n, err := writer.Write([]byte(record)); err != nil || n != len(record) {
			t.Fatalf("Write(%q) = (%d, %v), want (%d, nil)", record, n, err, len(record))
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	wantFiles := map[string]string{
		"majucau-worker.jsonl":   "record04",
		"majucau-worker.jsonl.1": "record03",
		"majucau-worker.jsonl.2": "record02",
	}
	for name, want := range wantFiles {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
		if int64(len(got)) > writer.maxBytes {
			t.Errorf("%s has %d bytes, exceeds cap %d", name, len(got), writer.maxBytes)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "majucau-worker.jsonl.3")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected fourth log file: stat error = %v", err)
	}
}
