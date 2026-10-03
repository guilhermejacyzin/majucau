package bling

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"majucau.local/financial-intelligence/internal/integrations/csvlimit"
	"majucau.local/financial-intelligence/internal/integrations/csvstream"
)

var ErrImportFolderRequired = errors.New("bling receipts import folder is required")

// ImportedFile describes one input file without exposing its absolute path.
// SHA-256 identifies the content in preview metadata; persistence idempotency
// is enforced by the normalized source identity and RAW payload hash.
type ImportedFile struct {
	Name         string `json:"name"`
	SHA256       string `json:"sha256"`
	ReceiptCount int    `json:"receipt_count"`
	ErrorCount   int    `json:"error_count"`
}

type FileRowError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FolderStreamSummary struct {
	ReceiptCount int
	ErrorCount   int
	IgnoredCount int
}

type FolderPreview struct {
	FolderStreamSummary
	Files     []ImportedFile
	FileCount int
	Issues    []FileRowError
}

type StreamReceiptHandler func(file string, line int, candidate ReceiptCandidate) error
type StreamReceiptIssueHandler func(issue FileRowError) error
type StreamReceiptFileHandler func(file ImportedFile)

func validateReceiptsFolder(folder string) error {
	if strings.TrimSpace(folder) == "" {
		return ErrImportFolderRequired
	}
	info, err := os.Stat(folder)
	if err != nil {
		return fmt.Errorf("inspect import folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("import path is not a folder: %s", filepath.Base(folder))
	}
	return nil
}

// StreamReceiptsFolder processes valid CSV rows one at a time. Deduplication
// is kept in the caller's transaction, and callbacks decide whether each
// candidate can be persisted. Only per-file summary metadata is retained.
func StreamReceiptsFolder(ctx context.Context, folder string, seen csvstream.SourceIDIndex, accept StreamReceiptHandler, reject StreamReceiptIssueHandler, onFile StreamReceiptFileHandler) (FolderStreamSummary, error) {
	return streamReceiptsFolder(ctx, folder, seen, accept, reject, onFile, false)
}

// StreamReceiptsFolderPreview keeps whole-file validation compatible with the
// current preview while scanning each CSV incrementally. A short preflight
// prevents a malformed file from contributing rows before it is rejected.
func StreamReceiptsFolderPreview(ctx context.Context, folder string, seen csvstream.SourceIDIndex, reject StreamReceiptIssueHandler, onFile StreamReceiptFileHandler) (FolderStreamSummary, error) {
	return streamReceiptsFolder(ctx, folder, seen, nil, reject, onFile, true)
}

func streamReceiptsFolder(ctx context.Context, folder string, seen csvstream.SourceIDIndex, accept StreamReceiptHandler, reject StreamReceiptIssueHandler, onFile StreamReceiptFileHandler, preview bool) (FolderStreamSummary, error) {
	if seen == nil {
		return FolderStreamSummary{}, errors.New("receipt import deduplication is required")
	}
	if err := ctx.Err(); err != nil {
		return FolderStreamSummary{}, err
	}
	if err := validateReceiptsFolder(folder); err != nil {
		return FolderStreamSummary{}, err
	}
	queue, err := seen.QueueCSVFiles(ctx, folder)
	if err != nil {
		return FolderStreamSummary{}, err
	}
	defer queue.Close()
	root, err := os.OpenRoot(folder)
	if err != nil {
		return FolderStreamSummary{}, fmt.Errorf("open receipt import root: %w", err)
	}
	defer root.Close()

	summary := FolderStreamSummary{IgnoredCount: queue.IgnoredCount()}
	for {
		if err := ctx.Err(); err != nil {
			return FolderStreamSummary{}, err
		}
		name, ok, err := queue.Next()
		if err != nil {
			return FolderStreamSummary{}, fmt.Errorf("read ordered receipt filenames: %w", err)
		}
		if !ok {
			break
		}
		file, err := root.Open(name)
		if err != nil {
			if preview {
				summary.ErrorCount++
				if reject != nil {
					if err := reject(FileRowError{File: name, Code: "FILE_READ", Message: "não foi possível abrir o arquivo"}); err != nil {
						return FolderStreamSummary{}, err
					}
				}
				continue
			}
			return FolderStreamSummary{}, fmt.Errorf("open receipt CSV %q: %w", name, err)
		}
		var validationHash []byte
		if preview {
			validationHasher := sha256.New()
			validationErr := StreamReceiptsCSV(io.TeeReader(file, validationHasher), func(_ int, _ ReceiptCandidate) error {
				return ctx.Err()
			}, func(RowError) error {
				return ctx.Err()
			})
			if err := ctx.Err(); err != nil {
				_ = file.Close()
				return FolderStreamSummary{}, err
			}
			if errors.Is(validationErr, csvlimit.ErrLimitExceeded) {
				_ = file.Close()
				return FolderStreamSummary{}, validationErr
			}
			if validationErr != nil {
				closeErr := file.Close()
				if closeErr != nil {
					return FolderStreamSummary{}, fmt.Errorf("close invalid receipt CSV %q: %w", name, closeErr)
				}
				metadata := ImportedFile{Name: name, SHA256: hex.EncodeToString(validationHasher.Sum(nil)), ErrorCount: 1}
				summary.ErrorCount++
				if reject != nil {
					if err := reject(FileRowError{File: name, Line: 1, Code: "INVALID_REPORT", Message: "arquivo não corresponde ao relatório de recebimentos do Bling"}); err != nil {
						return FolderStreamSummary{}, err
					}
				}
				if onFile != nil {
					onFile(metadata)
				}
				continue
			}
			validationHash = validationHasher.Sum(nil)
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				_ = file.Close()
				return FolderStreamSummary{}, fmt.Errorf("rewind receipt CSV for preview: %w", err)
			}
		}
		hasher := sha256.New()
		metadata := ImportedFile{Name: name}
		parseErr := StreamReceiptsCSV(io.TeeReader(file, hasher), func(line int, candidate ReceiptCandidate) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if previousFile, previousLine, duplicate, err := seen.CheckAndAdd(ctx, candidate.SourceID, name, line); err != nil {
				return err
			} else if duplicate {
				summary.ErrorCount++
				metadata.ErrorCount++
				issue := FileRowError{File: name, Line: line, Code: "DUPLICATE_SOURCE_ID"}
				if previousFile == name {
					issue.Message = fmt.Sprintf("documento repetido; primeira ocorrência na linha %d", previousLine)
				} else {
					issue.Message = fmt.Sprintf("documento repetido; primeira ocorrência no arquivo %s", previousFile)
				}
				if reject != nil {
					return reject(issue)
				}
				return nil
			}
			if accept != nil {
				if err := accept(name, line, candidate); err != nil {
					return err
				}
			}
			summary.ReceiptCount++
			metadata.ReceiptCount++
			return nil
		}, func(rowErr RowError) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			summary.ErrorCount++
			metadata.ErrorCount++
			if reject != nil {
				return reject(FileRowError{File: name, Line: rowErr.Line, Code: rowErr.Code, Message: rowErr.Message})
			}
			return nil
		})
		closeErr := file.Close()
		if closeErr != nil {
			return FolderStreamSummary{}, fmt.Errorf("close receipt CSV %q: %w", name, closeErr)
		}
		if err := ctx.Err(); err != nil {
			return FolderStreamSummary{}, err
		}
		if preview && !bytes.Equal(validationHash, hasher.Sum(nil)) {
			return FolderStreamSummary{}, fmt.Errorf("receipt CSV %q changed during preview", name)
		}
		if errors.Is(parseErr, ErrMissingReportColumn) || errors.Is(parseErr, ErrInvalidReportHeader) {
			summary.ErrorCount++
			metadata.ErrorCount++
			if reject != nil {
				if err := reject(FileRowError{File: name, Line: 1, Code: "INVALID_REPORT", Message: "arquivo não corresponde ao relatório de recebimentos do Bling"}); err != nil {
					return FolderStreamSummary{}, err
				}
			}
		} else if parseErr != nil {
			return FolderStreamSummary{}, fmt.Errorf("read receipt CSV %q: %w", name, parseErr)
		}
		metadata.SHA256 = hex.EncodeToString(hasher.Sum(nil))
		if onFile != nil {
			onFile(metadata)
		}
	}
	return summary, nil
}

// PreviewReceiptsFolder returns a bounded summary without requiring
// PostgreSQL. Temporary fingerprints and encrypted filenames are removed when
// the scan completes.
func PreviewReceiptsFolder(ctx context.Context, folder string) (preview FolderPreview, err error) {
	if err := validateReceiptsFolder(folder); err != nil {
		return FolderPreview{}, err
	}
	seen, err := csvstream.NewDiskSourceIDSet(ctx)
	if err != nil {
		return FolderPreview{}, err
	}
	defer func() { err = errors.Join(err, seen.Close()) }()
	preview.Files = make([]ImportedFile, 0, 50)
	preview.Issues = make([]FileRowError, 0, 50)
	preview.FolderStreamSummary, err = StreamReceiptsFolderPreview(ctx, folder, seen, func(issue FileRowError) error {
		if len(preview.Issues) < 50 {
			preview.Issues = append(preview.Issues, issue)
		}
		return nil
	}, func(file ImportedFile) {
		preview.FileCount++
		if len(preview.Files) < 50 {
			preview.Files = append(preview.Files, file)
		}
	})
	if err != nil {
		return FolderPreview{}, err
	}
	return preview, nil
}
