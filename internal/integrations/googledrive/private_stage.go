package googledrive

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"

	"majucau.local/financial-intelligence/internal/securetemp"
)

// WithPrivateStage downloads a file into a private, short-lived workspace,
// lets the consumer read it as a stream, and removes the workspace before it
// returns. The callback must finish reading before it returns and must persist
// its own results transactionally and idempotently.
func (c *Client) WithPrivateStage(
	ctx context.Context,
	parent string,
	fileID string,
	maxBytes int64,
	consume func(context.Context, File, io.Reader, int64, string) error,
) (result StreamResult, returnErr error) {
	if ctx == nil || consume == nil || maxBytes <= 0 {
		return StreamResult{}, ErrConfiguration
	}
	workspace, err := securetemp.NewWorkspace(parent, "majucau-drive-stage-")
	if err != nil {
		return StreamResult{}, ErrSink
	}
	defer func() {
		returnErr = errors.Join(returnErr, workspace.RemoveAll())
	}()

	sink, err := newPrivateFileSink(workspace, maxBytes)
	if err != nil {
		return StreamResult{}, err
	}
	result, err = c.StreamFile(ctx, fileID, maxBytes, sink)
	if err != nil {
		return StreamResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return StreamResult{}, err
	}

	staged, err := os.Open(sink.committedPath())
	if err != nil {
		return StreamResult{}, ErrSink
	}
	consumeErr := consume(ctx, result.File, staged, result.Bytes, result.SHA256)
	if ctx.Err() != nil {
		consumeErr = errors.Join(consumeErr, ctx.Err())
	}
	closeErr := staged.Close()
	if consumeErr != nil {
		return StreamResult{}, &fileVisitFailure{cause: errors.Join(consumeErr, closeErr)}
	}
	if closeErr != nil {
		return StreamResult{}, ErrSink
	}
	return result, nil
}

type privateFileSink struct {
	partialPath string
	readyPath   string
	file        *os.File
	maxBytes    int64
	written     int64
	committed   bool
}

func newPrivateFileSink(workspace *securetemp.Workspace, maxBytes int64) (*privateFileSink, error) {
	if workspace == nil || workspace.Path() == "" || maxBytes <= 0 {
		return nil, ErrConfiguration
	}
	var randomName [16]byte
	if _, err := rand.Read(randomName[:]); err != nil {
		return nil, ErrSink
	}
	base := hex.EncodeToString(randomName[:])
	return &privateFileSink{
		partialPath: filepath.Join(workspace.Path(), base+".partial"),
		readyPath:   filepath.Join(workspace.Path(), base+".ready"),
		maxBytes:    maxBytes,
	}, nil
}

func (s *privateFileSink) Reset() error {
	if s == nil {
		return ErrSink
	}
	var errs []error
	if s.file != nil {
		if err := s.file.Close(); err != nil {
			errs = append(errs, ErrSink)
		}
		s.file = nil
	}
	for _, path := range []string{s.partialPath, s.readyPath} {
		if err := securetemp.RemoveFile(path); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	file, err := os.OpenFile(s.partialPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return ErrSink
	}
	s.file = file
	s.written = 0
	s.committed = false
	return nil
}

func (s *privateFileSink) Write(buffer []byte) (int, error) {
	if s == nil || s.file == nil || s.committed {
		return 0, ErrSink
	}
	if int64(len(buffer)) > s.maxBytes-s.written {
		return 0, ErrFileTooLarge
	}
	n, err := s.file.Write(buffer)
	s.written += int64(n)
	if err != nil {
		return n, ErrSink
	}
	return n, nil
}

func (s *privateFileSink) Commit() error {
	if s == nil || s.file == nil || s.committed {
		return ErrSink
	}
	if err := s.file.Sync(); err != nil {
		return ErrSink
	}
	if err := s.file.Close(); err != nil {
		s.file = nil
		return ErrSink
	}
	s.file = nil
	if err := os.Rename(s.partialPath, s.readyPath); err != nil {
		return ErrSink
	}
	s.committed = true
	return nil
}

func (s *privateFileSink) Rollback() error {
	if s == nil {
		return ErrSink
	}
	var errs []error
	if s.file != nil {
		if err := s.file.Close(); err != nil {
			errs = append(errs, ErrSink)
		}
		s.file = nil
	}
	for _, path := range []string{s.partialPath, s.readyPath} {
		if err := securetemp.RemoveFile(path); err != nil {
			errs = append(errs, err)
		}
	}
	s.committed = false
	return errors.Join(errs...)
}

func (s *privateFileSink) committedPath() string {
	if s == nil || !s.committed {
		return ""
	}
	return s.readyPath
}

var _ StagingSink = (*privateFileSink)(nil)
