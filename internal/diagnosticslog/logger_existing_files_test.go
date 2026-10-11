package diagnosticslog

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRotatingWriterRejectsOversizedExistingLogWithoutChangingIt(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"majucau-worker.jsonl", "majucau-worker.jsonl.2"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, name)
			original := []byte("older log larger than the configured limit")
			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatalf("write existing log: %v", err)
			}

			writer, err := newRotatingWriter(dir, 8, 3)
			if writer != nil {
				_ = writer.Close()
			}
			if !errors.Is(err, ErrExistingLogTooLarge) {
				t.Fatalf("newRotatingWriter() error = %v, want %v", err, ErrExistingLogTooLarge)
			}

			got, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatalf("read existing log after rejection: %v", readErr)
			}
			if string(got) != string(original) {
				t.Fatalf("existing log changed: got %d bytes, want %d bytes", len(got), len(original))
			}
		})
	}
}
