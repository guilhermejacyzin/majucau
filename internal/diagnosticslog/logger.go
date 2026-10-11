package diagnosticslog

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	// MaxLogFileBytes and MaxLogFiles keep technical diagnostics bounded. The
	// active file plus its rotated copies use at most 50 MiB per installation.
	MaxLogFileBytes  int64 = 10 << 20
	MaxLogFiles            = 5
	maxEncodedRecord       = 32 << 10
)

var ErrRecordTooLarge = errors.New("diagnostic log record exceeds its size limit")

// Logger writes JSON diagnostics to one bounded, rotating local file set.
// Callers must only pass allowlisted fields and must never pass payloads,
// credentials, customer data, file paths, or raw error strings.
type Logger struct {
	logger *slog.Logger
	writer *rotatingWriter
}

func New(dir string) (*Logger, error) {
	if !filepath.IsAbs(dir) {
		return nil, fmt.Errorf("diagnostic log directory must be absolute")
	}
	if err := prepareLogDirectory(filepath.Clean(dir)); err != nil {
		return nil, err
	}
	writer, err := newRotatingWriter(filepath.Clean(dir), MaxLogFileBytes, MaxLogFiles)
	if err != nil {
		return nil, err
	}
	return &Logger{logger: slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: slog.LevelInfo})), writer: writer}, nil
}

func (l *Logger) Slog() *slog.Logger {
	if l == nil {
		return slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}
	return l.logger
}

func (l *Logger) Close() error {
	if l == nil || l.writer == nil {
		return nil
	}
	return l.writer.Close()
}

type rotatingWriter struct {
	mu       sync.Mutex
	dir      string
	file     *os.File
	size     int64
	maxBytes int64
	maxFiles int
}

func newRotatingWriter(dir string, maxBytes int64, maxFiles int) (*rotatingWriter, error) {
	w := &rotatingWriter{dir: dir, maxBytes: maxBytes, maxFiles: maxFiles}
	if err := w.openCurrent(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	if len(p) > maxEncodedRecord {
		return 0, ErrRecordTooLarge
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, os.ErrClosed
	}
	if w.size > 0 && w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	syncErr := w.file.Sync()
	closeErr := w.file.Close()
	w.file = nil
	return errors.Join(syncErr, closeErr)
}

func (w *rotatingWriter) rotate() error {
	if err := w.file.Sync(); err != nil {
		return err
	}
	if err := w.file.Close(); err != nil {
		w.file = nil
		return err
	}
	w.file = nil
	for index := w.maxFiles - 1; index >= 1; index-- {
		source := w.path(index - 1)
		destination := w.path(index)
		if _, err := os.Lstat(source); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		if err := os.Remove(destination); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Rename(source, destination); err != nil {
			return err
		}
	}
	return w.openCurrent()
}

func (w *rotatingWriter) openCurrent() error {
	path := w.path(0)
	file, err := openSafeLogFile(path)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		return errors.Join(err, file.Close())
	}
	w.file = file
	w.size = info.Size()
	return nil
}

func (w *rotatingWriter) path(index int) string {
	name := "majucau-worker.jsonl"
	if index > 0 {
		name = fmt.Sprintf("%s.%d", name, index)
	}
	return filepath.Join(w.dir, name)
}
