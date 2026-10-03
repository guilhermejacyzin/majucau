package nuvempago

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

// FutureFolderImport is a deterministic, non-persistent preview. The caller
// must point it at the controlled 02_nuvem_pago/recebimentos_futuros folder;
// no received/realized ledger is ever written by this function.
type FutureFolderImport struct {
	Files       []ImportedFutureFile        `json:"files"`
	Receivables []FutureReceivableCandidate `json:"receivables"`
	Errors      []FutureFileRowError        `json:"errors"`
	Ignored     []string                    `json:"ignored"`
}

type FutureFolderStreamSummary struct {
	ReceivableCount int
	ErrorCount      int
	IgnoredCount    int
}

type StreamFutureHandler func(file string, line int, candidate FutureReceivableCandidate) error
type StreamFutureIssueHandler func(issue FutureFileRowError) error
type StreamFutureFileHandler func(file ImportedFutureFile)

func ImportFutureFolder(ctx context.Context, folder string) (FutureFolderImport, error) {
	if err := ctx.Err(); err != nil {
		return FutureFolderImport{}, err
	}
	if err := validateFutureFolder(folder); err != nil {
		return FutureFolderImport{}, err
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return FutureFolderImport{}, fmt.Errorf("read future import folder: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	result := FutureFolderImport{}
	seenSourceIDs := make(map[string]string)
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
			result.Ignored = append(result.Ignored, entry.Name())
			continue
		}
		file, err := os.Open(filepath.Join(folder, entry.Name()))
		if err != nil {
			result.Errors = append(result.Errors, FutureFileRowError{File: entry.Name(), Line: 0, Code: "FILE_READ", Message: "não foi possível abrir o arquivo"})
			continue
		}
		hasher := sha256.New()
		report, parseErr := ParseFutureCSV(io.TeeReader(file, hasher))
		closeErr := file.Close()
		metadata := ImportedFutureFile{Name: entry.Name(), SHA256: hex.EncodeToString(hasher.Sum(nil))}
		if closeErr != nil && parseErr == nil {
			parseErr = closeErr
		}
		if errors.Is(parseErr, csvlimit.ErrLimitExceeded) {
			return FutureFolderImport{}, parseErr
		}
		if parseErr != nil {
			metadata.RejectedRowCount = 1
			result.Files = append(result.Files, metadata)
			result.Errors = append(result.Errors, FutureFileRowError{File: entry.Name(), Line: 1, Code: "INVALID_FUTURE_EXPORT", Message: "arquivo não corresponde ao extrato de recebimentos futuros do Nuvem Pago"})
			continue
		}
		metadata.RejectedRowCount = len(report.Errors)
		for _, candidate := range report.Receivables {
			if previousFile, exists := seenSourceIDs[candidate.SourceID]; exists {
				metadata.RejectedRowCount++
				result.Errors = append(result.Errors, FutureFileRowError{File: entry.Name(), Line: 0, Code: "DUPLICATE_SOURCE_ID", Message: fmt.Sprintf("transação repetida; primeira ocorrência no arquivo %s", previousFile)})
				continue
			}
			seenSourceIDs[candidate.SourceID] = entry.Name()
			metadata.ReceivableCount++
			result.Receivables = append(result.Receivables, candidate)
		}
		result.Files = append(result.Files, metadata)
		for _, rowErr := range report.Errors {
			result.Errors = append(result.Errors, FutureFileRowError{File: entry.Name(), Line: rowErr.Line, Code: rowErr.Code, Message: rowErr.Message})
		}
	}
	return result, nil
}

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
func StreamFutureFolder(ctx context.Context, folder string, seen *csvstream.SourceIDSet, accept StreamFutureHandler, reject StreamFutureIssueHandler, onFile StreamFutureFileHandler) (FutureFolderStreamSummary, error) {
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
			return FutureFolderStreamSummary{}, fmt.Errorf("open future CSV %q: %w", name, err)
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
