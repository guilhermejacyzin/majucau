package diagnosticslog

import (
	"errors"
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
