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

func ImportFutureFolder(ctx context.Context, folder string) (FutureFolderImport, error) {
	if strings.TrimSpace(folder) == "" {
		return FutureFolderImport{}, ErrFutureFolderRequired
	}
	if err := ctx.Err(); err != nil {
		return FutureFolderImport{}, err
	}
	info, err := os.Stat(folder)
	if err != nil {
		return FutureFolderImport{}, fmt.Errorf("inspect future import folder: %w", err)
	}
	if !info.IsDir() {
		return FutureFolderImport{}, fmt.Errorf("future import path is not a folder: %s", filepath.Base(folder))
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
