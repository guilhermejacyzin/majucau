package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/ipc"
)

func TestGetBootstrapStateFromWorker(t *testing.T) {
	handler := ipc.HandlerFunc(func(ctx context.Context, req ipc.Request) (ipc.Response, error) {
		switch req.Method {
		case ipc.MethodHealth:
			return ipc.NewResponse(req.RequestID, application.StaticHealth{
				Service: "majucau-worker",
				Version: "test",
			}.CheckHealth(ctx))
		case ipc.MethodIntegrationStatus:
			return ipc.NewResponse(req.RequestID, []application.IntegrationStatus{{
				Provider: domain.OriginBling,
				Status:   domain.IntegrationConnected,
			}})
		default:
			return ipc.Response{}, ipc.ErrUnsupportedMethod
		}
	})
	app := NewApp()
	app.clientFactory = func() (ipc.Client, error) { return ipc.NewFakeClient(handler), nil }

	got := app.GetBootstrapState()
	if got.ErrorCode != "" {
		t.Fatalf("unexpected public error: %s", got.ErrorCode)
	}
	if got.Worker.State != application.HealthOK || got.Worker.Version != "test" {
		t.Fatalf("unexpected worker state: %+v", got.Worker)
	}
	if len(got.Integrations) != 1 || got.Integrations[0].Status != domain.IntegrationConnected {
		t.Fatalf("unexpected integration state: %+v", got.Integrations)
	}
}

func TestGetBootstrapStateFailsClosed(t *testing.T) {
	app := NewApp()
	app.clientFactory = func() (ipc.Client, error) {
		return nil, errors.New("transport details must not leak")
	}

	got := app.GetBootstrapState()
	if got.ErrorCode != "WORKER_UNAVAILABLE" {
		t.Fatalf("expected stable unavailable code, got %q", got.ErrorCode)
	}
	if got.Worker.State != application.HealthUnavailable {
		t.Fatalf("expected unavailable worker, got %q", got.Worker.State)
	}
	if len(got.Integrations) != 3 {
		t.Fatalf("expected safe integration defaults, got %d", len(got.Integrations))
	}
}

func TestSaveBlingConfigForwardsSecretOnlyToWorker(t *testing.T) {
	handler := ipc.HandlerFunc(func(ctx context.Context, req ipc.Request) (ipc.Response, error) {
		if req.Method != ipc.MethodBlingConfigSave {
			return ipc.Response{}, ipc.ErrUnsupportedMethod
		}
		var input application.BlingConfigRequest
		if err := json.Unmarshal(req.Payload, &input); err != nil || input.ClientSecret != "secret-only-in-worker-call" {
			t.Fatalf("unexpected config payload: %#v %v", input, err)
		}
		return ipc.NewResponse(req.RequestID, application.BlingConfigResponse{ClientID: input.ClientID, RedirectURI: input.RedirectURI, SecretConfigured: true, Status: "NOT_CONFIGURED"})
	})
	app := NewApp()
	app.clientFactory = func() (ipc.Client, error) { return ipc.NewFakeClient(handler), nil }
	result := app.SaveBlingConfig(application.BlingConfigRequest{ClientID: "client", RedirectURI: "https://app.example.test/callback", ClientSecret: "secret-only-in-worker-call"})
	if result.ErrorCode != "" || !result.SecretConfigured || result.ClientID != "client" {
		t.Fatalf("unexpected config result: %+v", result)
	}
}

func TestBlingOAuthMethodsUseSanitizedWorkerContracts(t *testing.T) {
	handler := ipc.HandlerFunc(func(ctx context.Context, req ipc.Request) (ipc.Response, error) {
		switch req.Method {
		case ipc.MethodBlingOAuthStart:
			return ipc.NewResponse(req.RequestID, application.BlingOAuthStartResponse{SessionID: "session-1", AuthorizationURL: "https://www.bling.com.br/authorize?state=opaque", Status: "AUTHORIZING"})
		case ipc.MethodBlingOAuthStatus:
			return ipc.NewResponse(req.RequestID, application.BlingOAuthStatusResponse{SessionID: "session-1", Status: "CONNECTED", Message: "Autorização concluída."})
		case ipc.MethodBlingOAuthTest:
			return ipc.NewResponse(req.RequestID, application.BlingOAuthTestResponse{Status: "SUCCESS", PageRecordCount: 1})
		default:
			return ipc.Response{}, ipc.ErrUnsupportedMethod
		}
	})
	app := NewApp()
	app.clientFactory = func() (ipc.Client, error) { return ipc.NewFakeClient(handler), nil }
	start := app.StartBlingOAuth()
	if start.ErrorCode != "" || start.SessionID != "session-1" || start.AuthorizationURL == "" {
		t.Fatalf("unexpected OAuth start result: %+v", start)
	}
	status := app.GetBlingOAuthStatus("session-1")
	if status.ErrorCode != "" || status.Status != "CONNECTED" {
		t.Fatalf("unexpected OAuth status result: %+v", status)
	}
	testResult := app.TestBlingConnection()
	if testResult.ErrorCode != "" || testResult.PageRecordCount != 1 {
		t.Fatalf("unexpected OAuth test result: %+v", testResult)
	}
}

func TestSyncBlingForwardsExplicitFilterAndOnlyCounts(t *testing.T) {
	handler := ipc.HandlerFunc(func(ctx context.Context, req ipc.Request) (ipc.Response, error) {
		if req.Method != ipc.MethodBlingSync {
			return ipc.Response{}, ipc.ErrUnsupportedMethod
		}
		var input application.BlingSyncRequest
		if err := json.Unmarshal(req.Payload, &input); err != nil || input.ReceivedDateFrom != "2026-09-01" {
			t.Fatalf("unexpected sync request: %#v %v", input, err)
		}
		return ipc.NewResponse(req.RequestID, application.BlingSyncResponse{Status: "SUCCESS", Receivables: application.BlingSyncResourceResult{Status: "SUCCESS", RecordsRead: 2}, Payables: application.BlingSyncResourceResult{Status: "SUCCESS", RecordsRead: 1}})
	})
	app := NewApp()
	app.clientFactory = func() (ipc.Client, error) { return ipc.NewFakeClient(handler), nil }
	result := app.SyncBling(application.BlingSyncRequest{ReceivedDateFrom: "2026-09-01", ReceivedDateTo: "2026-09-30", PaymentDateFrom: "2026-09-01", PaymentDateTo: "2026-09-30"})
	if result.ErrorCode != "" || result.Status != "SUCCESS" || result.Receivables.RecordsRead != 2 || result.Payables.RecordsRead != 1 {
		t.Fatalf("unexpected sync result: %+v", result)
	}
}
