package nuvempago

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

var ErrFutureFolderRequired = errors.New("nuvem pago future import folder is required")

type ImportedFutureFile struct {
	Name             string `json:"name"`
	SHA256           string `json:"sha256"`
	ReceivableCount  int    `json:"receivable_count"`
	RejectedRowCount int    `json:"rejected_row_count"`
}

type FutureFileRowError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FutureFolderStreamSummary struct {
	ReceivableCount int
	ErrorCount      int
	IgnoredCount    int
}

type FutureFolderPreview struct {
	FutureFolderStreamSummary
	Files     []ImportedFutureFile
	FileCount int
	Issues    []FutureFileRowError
}

type StreamFutureHandler func(file string, line int, candidate FutureReceivableCandidate) error
type StreamFutureIssueHandler func(issue FutureFileRowError) error
type StreamFutureFileHandler func(file ImportedFutureFile)

func validateFutureFolder(folder string) error {
	if strings.TrimSpace(folder) == "" {
		return ErrFutureFolderRequired
	}
	info, err := os.Stat(folder)
	if err != nil {
		return fmt.Errorf("inspect future import folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("future import path is not a folder: %s", filepath.Base(folder))
	}
	return nil
}

// StreamFutureFolder processes valid rows one at a time. Deduplication lives
// in the caller's transaction; the folder walker retains only file metadata.
func StreamFutureFolder(ctx context.Context, folder string, seen csvstream.SourceIDIndex, accept StreamFutureHandler, reject StreamFutureIssueHandler, onFile StreamFutureFileHandler) (FutureFolderStreamSummary, error) {
	return streamFutureFolder(ctx, folder, seen, accept, reject, onFile, false)
}

// StreamFutureFolderPreview preserves whole-file validation in the current
// preview while processing rows incrementally.
func StreamFutureFolderPreview(ctx context.Context, folder string, seen csvstream.SourceIDIndex, reject StreamFutureIssueHandler, onFile StreamFutureFileHandler) (FutureFolderStreamSummary, error) {
	return streamFutureFolder(ctx, folder, seen, nil, reject, onFile, true)
}

func streamFutureFolder(ctx context.Context, folder string, seen csvstream.SourceIDIndex, accept StreamFutureHandler, reject StreamFutureIssueHandler, onFile StreamFutureFileHandler, preview bool) (FutureFolderStreamSummary, error) {
	if seen == nil {
		return FutureFolderStreamSummary{}, errors.New("future import deduplication is required")
	}
	if err := ctx.Err(); err != nil {
		return FutureFolderStreamSummary{}, err
	}
	if err := validateFutureFolder(folder); err != nil {
		return FutureFolderStreamSummary{}, err
	}
	queue, err := seen.QueueCSVFiles(ctx, folder)
	if err != nil {
		return FutureFolderStreamSummary{}, err
	}
	defer queue.Close()
	root, err := os.OpenRoot(folder)
	if err != nil {
		return FutureFolderStreamSummary{}, fmt.Errorf("open future import root: %w", err)
	}
	defer root.Close()

	summary := FutureFolderStreamSummary{IgnoredCount: queue.IgnoredCount()}
	for {
		if err := ctx.Err(); err != nil {
			return FutureFolderStreamSummary{}, err
		}
		name, ok, err := queue.Next()
		if err != nil {
			return FutureFolderStreamSummary{}, fmt.Errorf("read ordered future filenames: %w", err)
		}
		if !ok {
			break
		}
		file, err := root.Open(name)
		if err != nil {
			if preview {
				summary.ErrorCount++
				if reject != nil {
					if err := reject(FutureFileRowError{File: name, Code: "FILE_READ", Message: "não foi possível abrir o arquivo"}); err != nil {
						return FutureFolderStreamSummary{}, err
					}
				}
				continue
			}
			return FutureFolderStreamSummary{}, fmt.Errorf("open future CSV %q: %w", name, err)
		}
		var validationHash []byte
		if preview {
			validationHasher := sha256.New()
			validationErr := StreamFutureCSV(io.TeeReader(file, validationHasher), func(_ int, _ FutureReceivableCandidate) error {
				return ctx.Err()
			}, func(RowError) error {
				return ctx.Err()
			})
			if err := ctx.Err(); err != nil {
				_ = file.Close()
				return FutureFolderStreamSummary{}, err
			}
			if errors.Is(validationErr, csvlimit.ErrLimitExceeded) {
				_ = file.Close()
				return FutureFolderStreamSummary{}, validationErr
			}
			if validationErr != nil {
				closeErr := file.Close()
				if closeErr != nil {
					return FutureFolderStreamSummary{}, fmt.Errorf("close invalid future CSV %q: %w", name, closeErr)
				}
				metadata := ImportedFutureFile{Name: name, SHA256: hex.EncodeToString(validationHasher.Sum(nil)), RejectedRowCount: 1}
				summary.ErrorCount++
				if reject != nil {
					if err := reject(FutureFileRowError{File: name, Line: 1, Code: "INVALID_FUTURE_EXPORT", Message: "arquivo não corresponde ao extrato de recebimentos futuros do Nuvem Pago"}); err != nil {
						return FutureFolderStreamSummary{}, err
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
				return FutureFolderStreamSummary{}, fmt.Errorf("rewind future CSV for preview: %w", err)
			}
		}
		hasher := sha256.New()
		metadata := ImportedFutureFile{Name: name}
		parseErr := StreamFutureCSV(io.TeeReader(file, hasher), func(line int, candidate FutureReceivableCandidate) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			previousFile, previousLine, duplicate, err := seen.CheckAndAdd(ctx, candidate.SourceID, name, line)
			if err != nil {
				return err
			}
			if duplicate {
				summary.ErrorCount++
				metadata.RejectedRowCount++
				issue := FutureFileRowError{File: name, Line: line, Code: "DUPLICATE_SOURCE_ID"}
				if previousFile == name {
					issue.Message = fmt.Sprintf("transação repetida; primeira ocorrência na linha %d", previousLine)
				} else {
					issue.Message = fmt.Sprintf("transação repetida; primeira ocorrência no arquivo %s", previousFile)
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
			summary.ReceivableCount++
			metadata.ReceivableCount++
			return nil
		}, func(rowErr RowError) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			summary.ErrorCount++
			metadata.RejectedRowCount++
			if reject != nil {
				return reject(FutureFileRowError{File: name, Line: rowErr.Line, Code: rowErr.Code, Message: rowErr.Message})
			}
			return nil
		})
		closeErr := file.Close()
		if closeErr != nil {
			return FutureFolderStreamSummary{}, fmt.Errorf("close future CSV %q: %w", name, closeErr)
		}
		if err := ctx.Err(); err != nil {
			return FutureFolderStreamSummary{}, err
		}
		if preview && !bytes.Equal(validationHash, hasher.Sum(nil)) {
			return FutureFolderStreamSummary{}, fmt.Errorf("future CSV %q changed during preview", name)
		}
		if errors.Is(parseErr, ErrMissingFutureColumn) || errors.Is(parseErr, ErrInvalidFutureHeader) {
			summary.ErrorCount++
			metadata.RejectedRowCount++
			if reject != nil {
				if err := reject(FutureFileRowError{File: name, Line: 1, Code: "INVALID_FUTURE_EXPORT", Message: "arquivo não corresponde ao extrato de recebimentos futuros do Nuvem Pago"}); err != nil {
					return FutureFolderStreamSummary{}, err
				}
			}
		} else if parseErr != nil {
			return FutureFolderStreamSummary{}, fmt.Errorf("read future CSV %q: %w", name, parseErr)
		}
		metadata.SHA256 = hex.EncodeToString(hasher.Sum(nil))
		if onFile != nil {
			onFile(metadata)
		}
	}
	return summary, nil
}

// PreviewFutureFolder returns a bounded summary without requiring PostgreSQL.
func PreviewFutureFolder(ctx context.Context, folder string) (preview FutureFolderPreview, err error) {
	if err := validateFutureFolder(folder); err != nil {
		return FutureFolderPreview{}, err
	}
	seen, err := csvstream.NewDiskSourceIDSet(ctx)
	if err != nil {
		return FutureFolderPreview{}, err
	}
	defer func() { err = errors.Join(err, seen.Close()) }()
	preview.Files = make([]ImportedFutureFile, 0, 50)
	preview.Issues = make([]FutureFileRowError, 0, 50)
	preview.FutureFolderStreamSummary, err = StreamFutureFolderPreview(ctx, folder, seen, func(issue FutureFileRowError) error {
		if len(preview.Issues) < 50 {
			preview.Issues = append(preview.Issues, issue)
		}
		return nil
	}, func(file ImportedFutureFile) {
		preview.FileCount++
		if len(preview.Files) < 50 {
			preview.Files = append(preview.Files, file)
		}
	})
	if err != nil {
		return FutureFolderPreview{}, err
	}
	return preview, nil
}
