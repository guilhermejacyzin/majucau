package nuvempago

import (
	"context"
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
