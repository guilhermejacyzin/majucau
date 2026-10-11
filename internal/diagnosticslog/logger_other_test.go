//go:build !windows

package diagnosticslog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewWritesStructuredJSONToPrivateDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "logs")
	logger, err := New(dir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	logger.Slog().Info("worker event", "operation", "sync", "result", "success")
	if err := logger.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "majucau-worker.jsonl"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	if event["msg"] != "worker event" || event["operation"] != "sync" || event["result"] != "success" {
		t.Fatalf("unexpected structured event fields: %#v", event)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat log directory: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Errorf("log directory permissions = %04o, want 0700", got)
	}
	fileInfo, err := os.Stat(filepath.Join(dir, "majucau-worker.jsonl"))
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0600 {
		t.Errorf("log file permissions = %04o, want 0600", got)
	}
}
