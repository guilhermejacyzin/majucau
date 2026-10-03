package bling

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreviewReceiptsFolderIsDeterministicAndRejectsNuvemExport(t *testing.T) {
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, "01-bling.csv"), []byte(sampleReceiptsCSV), 0o600); err != nil {
		t.Fatal(err)
	}
	nuvemCSV := "Data do pagamento;Data de recebimento;Tipo de movimentação;Tipo de transação;Nº transação;Nome;Forma de pagamento;Bandeira;Nº Parcelas;Valor Bruto (R$);Taxas (R$);Juros (R$);Custos Totais (R$);Valor Liquido (R$)\n" +
		"20/09/2026;20/09/2026;Entrada;Venda;11008;Cliente;Pix;-;1;156,71;-1,55;0,00;-1,55;155,16\n"
	if err := os.WriteFile(filepath.Join(folder, "02-nuvem.csv"), []byte(nuvemCSV), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "leia-me.txt"), []byte("não importar"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := PreviewReceiptsFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if result.FileCount != 2 || len(result.Files) != 2 || result.ReceiptCount != 3 || result.ErrorCount != 3 || len(result.Issues) != 3 {
		t.Fatalf("files=%d/%d receipts=%d errors=%d issues=%#v", result.FileCount, len(result.Files), result.ReceiptCount, result.ErrorCount, result.Issues)
	}
	if result.IgnoredCount != 1 {
		t.Fatalf("ignored file count: %d", result.IgnoredCount)
	}
	if result.Files[0].Name != "01-bling.csv" || result.Files[0].ReceiptCount != 3 || len(result.Files[0].SHA256) != 64 {
		t.Fatalf("first file metadata: %#v", result.Files[0])
	}
	if result.Files[1].Name != "02-nuvem.csv" || result.Files[1].ErrorCount != 1 {
		t.Fatalf("second file metadata: %#v", result.Files[1])
	}
	if !strings.Contains(result.Issues[len(result.Issues)-1].Message, "Bling") {
		t.Fatalf("header error must identify the accepted contract: %#v", result.Issues)
	}
}

func TestPreviewReceiptsFolderHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := PreviewReceiptsFolder(ctx, t.TempDir())
	if err == nil {
		t.Fatal("cancelled import must fail")
	}
}

func TestPreviewReceiptsFolderRejectsDuplicateAcrossFiles(t *testing.T) {
	folder := t.TempDir()
	row := "Cliente;Histórico;Forma de pagamento;Nº documento;Vencimento;Liquidação;Situação;Valor taxa;Recebido\n" +
		"Cliente;Pedido;NUVEMPAGO 1X;DUP-1;01/08/2026;03/08/2026;pago;1,00;100,00\n"
	for _, name := range []string{"a.csv", "b.csv"} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(row), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	result, err := PreviewReceiptsFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if result.ReceiptCount != 1 || result.ErrorCount != 1 || len(result.Issues) != 1 || result.Issues[0].Code != "DUPLICATE_SOURCE_ID" {
		t.Fatalf("cross-file duplicate was not rejected: %#v", result)
	}
}

func TestPreviewReceiptsFolderBoundsMetadataAndIssuesAndCleansTemporaryData(t *testing.T) {
	folder := t.TempDir()
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	t.Setenv("TEMP", tempRoot)
	t.Setenv("TMP", tempRoot)

	for index := 0; index < 51; index++ {
		content := strings.ReplaceAll(sampleReceiptsCSV, "016223/01", fmt.Sprintf("%06d/01", 100000+index))
		content = strings.ReplaceAll(content, "007882/01", fmt.Sprintf("%06d/01", 200000+index))
		content = strings.ReplaceAll(content, "017785/01", fmt.Sprintf("%06d/01", 300000+index))
		name := fmt.Sprintf("%02d.csv", index)
		if err := os.WriteFile(filepath.Join(folder, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(folder, "readme.txt"), []byte("ignorado"), 0o600); err != nil {
		t.Fatal(err)
	}

	preview, err := PreviewReceiptsFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if preview.FileCount != 51 || len(preview.Files) != 50 || preview.ReceiptCount != 153 || preview.ErrorCount != 102 || preview.IgnoredCount != 1 || len(preview.Issues) != 50 {
		t.Fatalf("preview did not preserve total counts while bounding details: files=%d/%d receipts=%d errors=%d ignored=%d issues=%d", preview.FileCount, len(preview.Files), preview.ReceiptCount, preview.ErrorCount, preview.IgnoredCount, len(preview.Issues))
	}
	if preview.Files[0].Name != "00.csv" || preview.Files[49].Name != "49.csv" || preview.Files[0].ReceiptCount != 3 || preview.Files[0].ErrorCount != 2 || len(preview.Files[0].SHA256) != 64 {
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
