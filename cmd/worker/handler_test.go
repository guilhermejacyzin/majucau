package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/integrations/bling"
	"majucau.local/financial-intelligence/internal/ipc"
)

type fakeBlingConfigSaver struct{}

func (fakeBlingConfigSaver) Save(_ context.Context, input bling.BlingConfigInput) (bling.BlingConfigResult, error) {
	return bling.BlingConfigResult{ClientID: input.ClientID, RedirectURI: input.RedirectURI, SecretConfigured: input.ClientSecret != "", Status: "NOT_CONFIGURED"}, nil
}

type fakeBlingOAuthService struct{}

func (fakeBlingOAuthService) Start(context.Context) (bling.BlingOAuthStartResult, error) {
	return bling.BlingOAuthStartResult{SessionID: "session-1", AuthorizationURL: "https://www.bling.com.br/authorize?state=opaque", Status: "AUTHORIZING"}, nil
}
func (fakeBlingOAuthService) Status(context.Context, string) (bling.BlingOAuthStatusResult, error) {
	return bling.BlingOAuthStatusResult{SessionID: "session-1", Status: "CONNECTED", Message: "Autorização concluída."}, nil
}
func (fakeBlingOAuthService) Test(context.Context) (bling.BlingOAuthTestResult, error) {
	return bling.BlingOAuthTestResult{Status: "SUCCESS", PageRecordCount: 1}, nil
}

type fakeBlingRawSyncer struct {
	called bool
	filter bling.ReceivablesFilter
}

func (f *fakeBlingRawSyncer) Sync(_ context.Context, filter bling.ReceivablesFilter) (bling.BlingRawSyncResult, error) {
	f.called = true
	f.filter = filter
	return bling.BlingRawSyncResult{Status: "SUCCESS", Receivables: bling.APISyncResult{Status: "SUCCESS", PagesRead: 1, RecordsRead: 2}, Payables: bling.APISyncResult{Status: "SUCCESS", PagesRead: 1, RecordsRead: 1}}, nil
}

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

func TestWorkerBlingConfigNeverEchoesSecret(t *testing.T) {
	payload, err := json.Marshal(application.BlingConfigRequest{ClientID: "client", RedirectURI: "https://app.example.test/callback", ClientSecret: "secret-do-not-echo"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := (workerHandler{health: application.StaticHealth{Service: "test"}, blingConfig: fakeBlingConfigSaver{}}).Handle(context.Background(), ipc.Request{RequestID: "test", Method: ipc.MethodBlingConfigSave, Payload: payload})
	if err != nil || !response.OK {
		t.Fatalf("config response: %#v %v", response, err)
	}
	if strings.Contains(string(response.Payload), "secret-do-not-echo") {
		t.Fatalf("secret leaked in response: %s", response.Payload)
	}
	var result application.BlingConfigResponse
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		t.Fatal(err)
	}
	if !result.SecretConfigured || result.ClientID != "client" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestWorkerOAuthMethodsReturnOnlySanitizedState(t *testing.T) {
	handler := workerHandler{health: application.StaticHealth{Service: "test"}, blingOAuth: fakeBlingOAuthService{}}
	start, err := handler.Handle(context.Background(), ipc.Request{RequestID: "start", Method: ipc.MethodBlingOAuthStart})
	if err != nil || !start.OK || strings.Contains(string(start.Payload), "secret") {
		t.Fatalf("unexpected OAuth start response: %#v %v", start, err)
	}
	status, err := handler.Handle(context.Background(), ipc.Request{RequestID: "status", Method: ipc.MethodBlingOAuthStatus, Payload: []byte(`{"session_id":"session-1"}`)})
	if err != nil || !status.OK || !strings.Contains(string(status.Payload), "CONNECTED") {
		t.Fatalf("unexpected OAuth status response: %#v %v", status, err)
	}
	testResponse, err := handler.Handle(context.Background(), ipc.Request{RequestID: "test", Method: ipc.MethodBlingOAuthTest, Payload: []byte(`{}`)})
	if err != nil || !testResponse.OK || !strings.Contains(string(testResponse.Payload), "page_record_count") {
		t.Fatalf("unexpected OAuth test response: %#v %v", testResponse, err)
	}
}

func TestWorkerSyncRequiresExplicitFilterAndReturnsCounts(t *testing.T) {
	fake := &fakeBlingRawSyncer{}
	handler := workerHandler{health: application.StaticHealth{Service: "test"}, blingSync: fake}
	missing, err := handler.Handle(context.Background(), ipc.Request{RequestID: "missing", Method: ipc.MethodBlingSync, Payload: []byte(`{}`)})
	if err != nil || missing.OK || missing.Error == nil || missing.Error.Code != "BLING_SYNC_FILTER_REQUIRED" {
		t.Fatalf("missing filter response: %#v %v", missing, err)
	}
	payload := []byte(`{"received_date_from":"2026-09-01","received_date_to":"2026-09-30","payment_date_from":"2026-09-01","payment_date_to":"2026-09-30"}`)
	response, err := handler.Handle(context.Background(), ipc.Request{RequestID: "sync", Method: ipc.MethodBlingSync, Payload: payload})
	if err != nil || !response.OK || !fake.called {
		t.Fatalf("sync response: %#v %v", response, err)
	}
	if fake.filter.ReceivedDateFrom != "2026-09-01" || fake.filter.PaymentDateTo != "2026-09-30" || !strings.Contains(string(response.Payload), `"records_read":2`) {
		t.Fatalf("filter/counts not preserved: filter=%+v payload=%s", fake.filter, response.Payload)
	}
}

func TestBlingSyncErrorClassificationIsSanitized(t *testing.T) {
	tests := []struct {
		err  error
		code string
	}{
		{err: bling.ErrBlingAPIUnauthorized, code: "BLING_API_UNAUTHORIZED"},
		{err: bling.ErrBlingAPIRateLimited, code: "BLING_API_RATE_LIMITED"},
		{err: bling.ErrBlingAPIUnavailable, code: "BLING_API_UNAVAILABLE"},
		{err: bling.ErrBlingAPISchemaMismatch, code: "BLING_SCHEMA_MISMATCH"},
		{err: bling.ErrBlingAPIInvalidResponse, code: "BLING_INVALID_RESPONSE"},
	}
	for _, test := range tests {
		t.Run(test.code, func(t *testing.T) {
			if got := blingSyncErrorCode(test.err); got != test.code {
				t.Fatalf("code = %q, want %q", got, test.code)
			}
			if strings.Contains(blingSyncErrorMessage(test.err), "token") || errors.Is(test.err, context.Canceled) {
				t.Fatal("sanitized message leaked implementation detail")
			}
		})
	}
}
