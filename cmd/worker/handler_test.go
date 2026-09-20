package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/ipc"
)

func TestWorkerPreviewReturnsSanitizedBlingSummary(t *testing.T) {
	folder := t.TempDir()
	csv := "Cliente;Histórico;Forma de pagamento;Nº documento;Vencimento;Liquidação;Situação;Valor taxa;Recebido\n" +
		"Cliente Interno;Pedido confidencial;NUVEMPAGO 1X;DOC-1;01/08/2026;03/08/2026;pago;1,00;100,00\n"
	if err := os.WriteFile(filepath.Join(folder, "recebidos.csv"), []byte(csv), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]string{"folder": folder})
	if err != nil {
		t.Fatal(err)
	}
	response, err := (newWorkerHandler()).Handle(context.Background(), ipc.Request{RequestID: "test", Method: ipc.MethodBlingReceiptsPreview, Payload: payload})
	if err != nil || !response.OK {
		t.Fatalf("preview response: %#v %v", response, err)
	}
	var preview application.BlingReceiptImportPreview
	if err := json.Unmarshal(response.Payload, &preview); err != nil {
		t.Fatal(err)
	}
	if preview.ReceiptCount != 1 || preview.ErrorCount != 0 || len(preview.Files) != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	if strings.Contains(string(response.Payload), "Cliente Interno") || strings.Contains(string(response.Payload), "confidencial") {
		t.Fatalf("preview leaked financial row data: %s", response.Payload)
	}
}

func TestWorkerPreviewRejectsMissingFolder(t *testing.T) {
	response, err := (newWorkerHandler()).Handle(context.Background(), ipc.Request{RequestID: "test", Method: ipc.MethodBlingReceiptsPreview, Payload: []byte(`{"folder":""}`)})
	if err != nil || response.OK || response.Error == nil || response.Error.Code != "BLING_IMPORT_FOLDER_REQUIRED" {
		t.Fatalf("missing folder response: %#v %v", response, err)
	}
}

func TestWorkerImportFailsClosedWhenDatabaseIsNotConfigured(t *testing.T) {
	response, err := (workerHandler{health: application.StaticHealth{Service: "test"}}).Handle(context.Background(), ipc.Request{RequestID: "test", Method: ipc.MethodBlingReceiptsImport, Payload: []byte(`{"folder":"C:\\imports"}`)})
	if err != nil || response.OK || response.Error == nil || response.Error.Code != "BLING_DATABASE_NOT_CONFIGURED" {
		t.Fatalf("database-disabled import response: %#v %v", response, err)
	}
}
