package bling

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportReceiptsFolderIsDeterministicAndRejectsNuvemExport(t *testing.T) {
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

	result, err := ImportReceiptsFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 2 || len(result.Receipts) != 3 || len(result.Errors) != 3 {
		t.Fatalf("files=%d receipts=%d errors=%d ignored=%#v", len(result.Files), len(result.Receipts), len(result.Errors), result.Ignored)
	}
	if len(result.Ignored) != 1 || result.Ignored[0] != "leia-me.txt" {
		t.Fatalf("ignored files: %#v", result.Ignored)
	}
	if result.Files[0].Name != "01-bling.csv" || result.Files[0].ReceiptCount != 3 || len(result.Files[0].SHA256) != 64 {
		t.Fatalf("first file metadata: %#v", result.Files[0])
	}
	if result.Files[1].Name != "02-nuvem.csv" || result.Files[1].ErrorCount != 1 {
		t.Fatalf("second file metadata: %#v", result.Files[1])
	}
	if !strings.Contains(result.Errors[len(result.Errors)-1].Message, "Bling") {
		t.Fatalf("header error must identify the accepted contract: %#v", result.Errors)
	}
}

func TestImportReceiptsFolderHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ImportReceiptsFolder(ctx, t.TempDir())
	if err == nil {
		t.Fatal("cancelled import must fail")
	}
}

func TestImportReceiptsFolderRejectsDuplicateAcrossFiles(t *testing.T) {
	folder := t.TempDir()
	row := "Cliente;Histórico;Forma de pagamento;Nº documento;Vencimento;Liquidação;Situação;Valor taxa;Recebido\n" +
		"Cliente;Pedido;NUVEMPAGO 1X;DUP-1;01/08/2026;03/08/2026;pago;1,00;100,00\n"
	for _, name := range []string{"a.csv", "b.csv"} {
		if err := os.WriteFile(filepath.Join(folder, name), []byte(row), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	result, err := ImportReceiptsFolder(context.Background(), folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Receipts) != 1 || len(result.Errors) != 1 || result.Errors[0].Code != "DUPLICATE_SOURCE_ID" {
		t.Fatalf("cross-file duplicate was not rejected: %#v", result)
	}
}
