package nuvempago

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportFutureFolderIsDeterministicAndDeduplicatesAcrossFiles(t *testing.T) {
	folder := t.TempDir()
	valid := strings.Replace(sampleFutureCSV, ";Saída;", ";Entrada;", 1)
	if err := os.WriteFile(filepath.Join(folder, "B-futuros.csv"), []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "A-futuros.csv"), []byte(strings.Replace(valid, "9483", "9486", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "leia-me.txt"), []byte("fora do contrato"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := ImportFutureFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 2 || result.Files[0].Name != "A-futuros.csv" || result.Files[1].Name != "B-futuros.csv" {
		t.Fatalf("files are not lexical/deterministic: %#v", result.Files)
	}
	if len(result.Receivables) != 4 || len(result.Errors) != 4 || len(result.Ignored) != 1 {
		t.Fatalf("unexpected folder result: %#v", result)
	}
	if result.Ignored[0] != "leia-me.txt" || result.Errors[0].Code != "DUPLICATE_SOURCE_ID" {
		t.Fatalf("unexpected ignored/error result: %#v", result)
	}
	if result.Files[0].SHA256 == "" || result.Files[1].SHA256 == "" || result.Files[0].SHA256 == result.Files[1].SHA256 {
		t.Fatal("each file must have a distinct content hash")
	}
}

func TestPreviewFutureFolderBoundsMetadataAndIssuesAndCleansTemporaryData(t *testing.T) {
	folder := t.TempDir()
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	t.Setenv("TEMP", tempRoot)
	t.Setenv("TMP", tempRoot)

	for index := 0; index < 51; index++ {
		content := strings.ReplaceAll(sampleFutureCSV, "9483", fmt.Sprintf("TX%04d", index))
		content = strings.ReplaceAll(content, "9484", fmt.Sprintf("OT%04d", index))
		name := fmt.Sprintf("%02d.csv", index)
		if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(folder, "readme.txt"), []byte("ignorado"), 0o600); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewFutureFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if preview.FileCount != 51 || len(preview.Files) != 50 || preview.ReceivableCount != 102 || preview.ErrorCount != 102 || preview.IgnoredCount != 1 || len(preview.Issues) != 50 {
		t.Fatalf("preview did not preserve total counts while bounding details: files=%d/%d receivables=%d errors=%d ignored=%d issues=%d", preview.FileCount, len(preview.Files), preview.ReceivableCount, preview.ErrorCount, preview.IgnoredCount, len(preview.Issues))
	}
	if preview.Files[0].Name != "00.csv" || preview.Files[49].Name != "49.csv" || preview.Files[0].ReceivableCount != 2 || preview.Files[0].RejectedRowCount != 2 || len(preview.Files[0].SHA256) != 64 {
		t.Fatalf("unexpected bounded file metadata: first=%#v last=%#v", preview.Files[0], preview.Files[49])
	}
	if preview.Issues[0].File != "00.csv" || preview.Issues[49].File != "24.csv" {
		t.Fatalf("issues should be collected in stable filename order: first=%#v last=%#v", preview.Issues[0], preview.Issues[49])
	}
	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("preview left temporary files behind: %v", entries)
	}
}
