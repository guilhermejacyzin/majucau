package bling

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

var ErrImportFolderRequired = errors.New("bling receipts import folder is required")

// ImportedFile describes one input file without exposing its absolute path.
// The SHA-256 is the stable identity used by the later persistence worker for
// idempotent batch tracking.
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

// FolderImport is the deterministic, non-persistent result of one folder
// scan. Persistence is intentionally a separate worker transaction so a
// malformed file can never partially alter the ledger.
type FolderImport struct {
	Files    []ImportedFile     `json:"files"`
	Receipts []ReceiptCandidate `json:"receipts"`
	Errors   []FileRowError     `json:"errors"`
	Ignored  []string           `json:"ignored"`
}

// ImportReceiptsFolder reads only CSV files in a single directory. Files are
// processed in lexical order, making previews and retries reproducible. A
// Nuvem/Nuvem Pago export does not match the Bling report contract and is
// returned as an explicit header error; it is never reclassified as a Bling
// receipt.
func ImportReceiptsFolder(ctx context.Context, folder string) (FolderImport, error) {
	if strings.TrimSpace(folder) == "" {
		return FolderImport{}, ErrImportFolderRequired
	}
	if err := ctx.Err(); err != nil {
		return FolderImport{}, err
	}
	info, err := os.Stat(folder)
	if err != nil {
		return FolderImport{}, fmt.Errorf("inspect import folder: %w", err)
	}
	if !info.IsDir() {
		return FolderImport{}, fmt.Errorf("import path is not a folder: %s", filepath.Base(folder))
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return FolderImport{}, fmt.Errorf("read import folder: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	result := FolderImport{}
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
			result.Errors = append(result.Errors, FileRowError{File: entry.Name(), Line: 0, Code: "FILE_READ", Message: "não foi possível abrir o arquivo"})
			continue
		}
		hasher := sha256.New()
		report, parseErr := ParseReceiptsCSV(io.TeeReader(file, hasher))
		closeErr := file.Close()
		metadata := ImportedFile{Name: entry.Name(), SHA256: hex.EncodeToString(hasher.Sum(nil))}
		if closeErr != nil && parseErr == nil {
			parseErr = closeErr
		}
		if parseErr != nil {
			metadata.ErrorCount = 1
			result.Files = append(result.Files, metadata)
			result.Errors = append(result.Errors, FileRowError{File: entry.Name(), Line: 1, Code: "INVALID_REPORT", Message: "arquivo não corresponde ao relatório de recebimentos do Bling"})
			continue
		}
		metadata.ErrorCount = len(report.Errors)
		for _, receipt := range report.Receipts {
			if previousFile, exists := seenSourceIDs[receipt.SourceID]; exists {
				metadata.ErrorCount++
				result.Errors = append(result.Errors, FileRowError{File: entry.Name(), Line: 0, Code: "DUPLICATE_SOURCE_ID", Message: fmt.Sprintf("documento repetido; primeira ocorrência no arquivo %s", previousFile)})
				continue
			}
			seenSourceIDs[receipt.SourceID] = entry.Name()
			metadata.ReceiptCount++
			result.Receipts = append(result.Receipts, receipt)
		}
		result.Files = append(result.Files, metadata)
		for _, rowErr := range report.Errors {
			result.Errors = append(result.Errors, FileRowError{File: entry.Name(), Line: rowErr.Line, Code: rowErr.Code, Message: rowErr.Message})
		}
	}
	return result, nil
}
