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
