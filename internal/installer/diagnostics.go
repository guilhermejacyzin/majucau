package installer

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrDiagnosticOutput = errors.New("diagnostic output is invalid")

type DiagnosticMetadata struct {
	AppVersion    string
	CorrelationID string
	GeneratedAt   time.Time
}

type DiagnosticBundleInfo struct {
	SchemaVersion string   `json:"schema_version"`
	Entries       []string `json:"entries"`
	Bytes         int64    `json:"bytes"`
	SHA256        string   `json:"sha256"`
}

type diagnosticDocument struct {
	SchemaVersion string    `json:"schema_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	AppVersion    string    `json:"app_version,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Preflight     Result    `json:"preflight"`
}

// WriteDiagnosticBundle creates a small support bundle containing only the
// sanitized preflight contract and operator instructions. It intentionally
// does not read environment variables, logs, registry values or command-line
// arguments beyond the already-sanitized Result supplied by the caller.
func WriteDiagnosticBundle(outputPath string, result Result, metadata DiagnosticMetadata) (DiagnosticBundleInfo, error) {
	clean, err := validateDiagnosticPath(outputPath)
	if err != nil {
		return DiagnosticBundleInfo{}, err
	}
	if metadata.GeneratedAt.IsZero() {
		metadata.GeneratedAt = time.Now().UTC()
	}
	document, err := json.MarshalIndent(diagnosticDocument{
		SchemaVersion: SchemaVersion,
		GeneratedAt:   metadata.GeneratedAt.UTC(),
		AppVersion:    strings.TrimSpace(metadata.AppVersion),
		CorrelationID: strings.TrimSpace(metadata.CorrelationID),
		Preflight:     result,
	}, "", "  ")
	if err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: serialize", ErrDiagnosticOutput)
	}

	readme := []byte("Majucau Financial Intelligence — diagnóstico sanitizado\n\n" +
		"Este pacote contém somente o contrato de preflight e orientações técnicas.\n" +
		"Ele não contém senha, token, Client Secret, DSN, caminho completo, nome de usuário,\n" +
		"payload de API ou conteúdo de arquivos financeiros. Envie este arquivo somente\n" +
		"para o suporte autorizado e preserve o código exibido pelo instalador.\n")

	directory := filepath.Dir(clean)
	temporary, err := os.CreateTemp(directory, ".majucau-diagnostic-*.tmp")
	if err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: create temporary bundle", ErrDiagnosticOutput)
	}
	temporaryPath := temporary.Name()
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()
	_ = temporary.Chmod(0600)

	archive := zip.NewWriter(temporary)
	if err := writeZipEntry(archive, "diagnostic.json", document); err != nil {
		_ = archive.Close()
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: write contract", ErrDiagnosticOutput)
	}
	if err := writeZipEntry(archive, "README.txt", readme); err != nil {
		_ = archive.Close()
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: write instructions", ErrDiagnosticOutput)
	}
	if err := archive.Close(); err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: close bundle", ErrDiagnosticOutput)
	}
	if err := temporary.Sync(); err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: flush bundle", ErrDiagnosticOutput)
	}
	if err := temporary.Close(); err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: close temporary bundle", ErrDiagnosticOutput)
	}
	if err := os.Rename(temporaryPath, clean); err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: commit bundle", ErrDiagnosticOutput)
	}
	removeTemporary = false

	info, err := os.Stat(clean)
	if err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: inspect bundle", ErrDiagnosticOutput)
	}
	digest, err := fileSHA256(clean)
	if err != nil {
		return DiagnosticBundleInfo{}, fmt.Errorf("%w: hash bundle", ErrDiagnosticOutput)
	}
	return DiagnosticBundleInfo{SchemaVersion: SchemaVersion, Entries: []string{"diagnostic.json", "README.txt"}, Bytes: info.Size(), SHA256: digest}, nil
}

func validateDiagnosticPath(path string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || clean == "" || filepath.Ext(clean) != ".zip" || !filepath.IsAbs(clean) {
		return "", ErrDiagnosticOutput
	}
	parent := filepath.Dir(clean)
	if parent == "." || !pathExistsAsDirectory(parent) {
		return "", ErrDiagnosticOutput
	}
	if info, err := os.Lstat(clean); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", ErrDiagnosticOutput
	}
	return clean, nil
}

func pathExistsAsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func writeZipEntry(archive *zip.Writer, name string, content []byte) error {
	if strings.Contains(name, `\`) || strings.Contains(name, "..") {
		return ErrDiagnosticOutput
	}
	entry, err := archive.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write(content)
	return err
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}
