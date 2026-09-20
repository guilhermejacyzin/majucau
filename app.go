package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/ipc"
)

const appVersion = "0.1.0-g1"

// App is the deliberately small Wails boundary. It never receives or stores
// provider secrets; privileged actions belong to the worker.
type App struct {
	ctx           context.Context
	clientFactory func() (ipc.Client, error)
}

type BootstrapState struct {
	AppVersion   string                          `json:"app_version"`
	Worker       application.HealthResponse      `json:"worker"`
	Integrations []application.IntegrationStatus `json:"integrations"`
	ErrorCode    string                          `json:"error_code,omitempty"`
	Message      string                          `json:"message,omitempty"`
	CheckedAt    time.Time                       `json:"checked_at"`
}

func NewApp() *App {
	return &App{clientFactory: func() (ipc.Client, error) {
		return ipc.NewNamedPipeClient(ipc.DefaultNamedPipeConfig())
	}}
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }

// GetBootstrapState asks the local worker for live health and integration
// status. Failure is explicit and does not synthesize financial values.
func (a *App) GetBootstrapState() BootstrapState {
	checkedAt := time.Now().UTC()
	unavailable := BootstrapState{
		AppVersion: appVersion,
		Worker: application.HealthResponse{
			Service:   "majucau-worker",
			Version:   "unknown",
			State:     application.HealthUnavailable,
			CheckedAt: checkedAt,
		},
		Integrations: unavailableIntegrations(),
		ErrorCode:    "WORKER_UNAVAILABLE",
		Message:      "O serviço local ainda não está disponível.",
		CheckedAt:    checkedAt,
	}

	client, err := a.clientFactory()
	if err != nil {
		return unavailable
	}
	defer client.Close()

	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()

	healthResponse, err := client.Call(ctx, ipc.Request{
		Version:   ipc.ProtocolVersion,
		RequestID: requestID("health"),
		Method:    ipc.MethodHealth,
	})
	if err != nil || !healthResponse.OK {
		return unavailable
	}
	var health application.HealthResponse
	if err := json.Unmarshal(healthResponse.Payload, &health); err != nil {
		unavailable.ErrorCode = "WORKER_INVALID_RESPONSE"
		unavailable.Message = "O serviço local respondeu em formato inválido."
		return unavailable
	}

	integrationResponse, err := client.Call(ctx, ipc.Request{
		Version:   ipc.ProtocolVersion,
		RequestID: requestID("integrations"),
		Method:    ipc.MethodIntegrationStatus,
	})
	if err != nil || !integrationResponse.OK {
		return BootstrapState{
			AppVersion:   appVersion,
			Worker:       health,
			Integrations: unavailableIntegrations(),
			ErrorCode:    "INTEGRATION_STATUS_UNAVAILABLE",
			Message:      "O status das integrações não pôde ser consultado.",
			CheckedAt:    checkedAt,
		}
	}
	var integrations []application.IntegrationStatus
	if err := json.Unmarshal(integrationResponse.Payload, &integrations); err != nil {
		return BootstrapState{
			AppVersion:   appVersion,
			Worker:       health,
			Integrations: unavailableIntegrations(),
			ErrorCode:    "WORKER_INVALID_RESPONSE",
			Message:      "O serviço local respondeu em formato inválido.",
			CheckedAt:    checkedAt,
		}
	}
	return BootstrapState{
		AppVersion:   appVersion,
		Worker:       health,
		Integrations: integrations,
		CheckedAt:    checkedAt,
	}
}

// SaveBlingConfig forwards the editable Bling configuration to the worker.
// The Wails boundary does not persist or echo the secret; the worker owns the
// DPAPI write and returns only sanitized metadata.
func (a *App) SaveBlingConfig(input application.BlingConfigRequest) application.BlingConfigResponse {
	fallback := application.BlingConfigResponse{ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	payload, err := json.Marshal(input)
	if err != nil {
		return application.BlingConfigResponse{ErrorCode: "BLING_CONFIG_INVALID", Message: "A configuração do Bling não pôde ser preparada."}
	}
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-config"), Method: ipc.MethodBlingConfigSave, Payload: payload})
	if err != nil {
		return fallback
	}
	if !response.OK {
		if response.Error == nil {
			return application.BlingConfigResponse{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu sem explicar o erro."}
		}
		return application.BlingConfigResponse{ErrorCode: response.Error.Code, Message: response.Error.Message}
	}
	var result application.BlingConfigResponse
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		return application.BlingConfigResponse{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	return result
}

// StartBlingOAuth asks the worker to create a short-lived loopback callback,
// then opens the provider authorization page in the user's external browser.
// The URL contains no client secret or token.
func (a *App) StartBlingOAuth() application.BlingOAuthStartResponse {
	fallback := application.BlingOAuthStartResponse{ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-oauth-start"), Method: ipc.MethodBlingOAuthStart})
	if err != nil {
		return fallback
	}
	if !response.OK {
		return application.BlingOAuthStartResponse{ErrorCode: workerErrorCode(response), Message: workerErrorMessage(response)}
	}
	var result application.BlingOAuthStartResponse
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		return application.BlingOAuthStartResponse{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	if result.AuthorizationURL != "" && a.ctx != nil {
		runtime.BrowserOpenURL(a.ctx, result.AuthorizationURL)
	}
	return result
}

// GetBlingOAuthStatus returns only the sanitized state of a pending OAuth
// session. Codes, tokens and callback details never cross this boundary.
func (a *App) GetBlingOAuthStatus(sessionID string) application.BlingOAuthStatusResponse {
	fallback := application.BlingOAuthStatusResponse{SessionID: sessionID, ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	payload, err := json.Marshal(application.BlingOAuthStatusRequest{SessionID: sessionID})
	if err != nil {
		return application.BlingOAuthStatusResponse{SessionID: sessionID, ErrorCode: "BLING_OAUTH_SESSION_INVALID", Message: "A sessão de autorização não é válida."}
	}
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-oauth-status"), Method: ipc.MethodBlingOAuthStatus, Payload: payload})
	if err != nil {
		return fallback
	}
	if !response.OK {
		return application.BlingOAuthStatusResponse{SessionID: sessionID, ErrorCode: workerErrorCode(response), Message: workerErrorMessage(response)}
	}
	var result application.BlingOAuthStatusResponse
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		return application.BlingOAuthStatusResponse{SessionID: sessionID, ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	return result
}

// TestBlingConnection performs the documented one-record read without
// importing, changing or synchronizing any financial data.
func (a *App) TestBlingConnection() application.BlingOAuthTestResponse {
	fallback := application.BlingOAuthTestResponse{ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-oauth-test"), Method: ipc.MethodBlingOAuthTest, Payload: json.RawMessage(`{}`)})
	if err != nil {
		return fallback
	}
	if !response.OK {
		return application.BlingOAuthTestResponse{ErrorCode: workerErrorCode(response), Message: workerErrorMessage(response)}
	}
	var result application.BlingOAuthTestResponse
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		return application.BlingOAuthTestResponse{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	return result
}

func workerErrorCode(response ipc.Response) string {
	if response.Error == nil || response.Error.Code == "" {
		return "WORKER_INVALID_RESPONSE"
	}
	return response.Error.Code
}

func workerErrorMessage(response ipc.Response) string {
	if response.Error == nil || response.Error.Message == "" {
		return "O serviço local respondeu sem explicar o erro."
	}
	return response.Error.Message
}

// PreviewBlingReceipts asks the worker to validate a local Bling export folder.
// It returns counts and sanitized issues only; financial rows stay in the worker.
func (a *App) PreviewBlingReceipts(folder string) application.BlingReceiptImportPreview {
	fallback := application.BlingReceiptImportPreview{ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	payload, err := json.Marshal(map[string]string{"folder": folder})
	if err != nil {
		return application.BlingReceiptImportPreview{ErrorCode: "IMPORT_REQUEST_INVALID", Message: "A pasta informada não pôde ser preparada."}
	}
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-preview"), Method: ipc.MethodBlingReceiptsPreview, Payload: payload})
	if err != nil {
		return fallback
	}
	if !response.OK {
		if response.Error == nil {
			return application.BlingReceiptImportPreview{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu sem explicar o erro."}
		}
		return application.BlingReceiptImportPreview{ErrorCode: response.Error.Code, Message: response.Error.Message}
	}
	var preview application.BlingReceiptImportPreview
	if err := json.Unmarshal(response.Payload, &preview); err != nil {
		return application.BlingReceiptImportPreview{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	return preview
}

// ImportBlingReceipts asks the worker to persist the validated folder in one
// PostgreSQL transaction. The UI receives counts and stable error codes only.
func (a *App) ImportBlingReceipts(folder string) application.BlingReceiptImportResult {
	fallback := application.BlingReceiptImportResult{ErrorCode: "WORKER_UNAVAILABLE", Message: "O serviço local ainda não está disponível."}
	client, err := a.clientFactory()
	if err != nil {
		return fallback
	}
	defer client.Close()
	parent := a.ctx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()
	payload, err := json.Marshal(map[string]string{"folder": folder})
	if err != nil {
		return application.BlingReceiptImportResult{ErrorCode: "IMPORT_REQUEST_INVALID", Message: "A pasta informada não pôde ser preparada."}
	}
	response, err := client.Call(ctx, ipc.Request{Version: ipc.ProtocolVersion, RequestID: requestID("bling-import"), Method: ipc.MethodBlingReceiptsImport, Payload: payload})
	if err != nil {
		return fallback
	}
	if !response.OK {
		if response.Error == nil {
			return application.BlingReceiptImportResult{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu sem explicar o erro."}
		}
		return application.BlingReceiptImportResult{ErrorCode: response.Error.Code, Message: response.Error.Message}
	}
	var result application.BlingReceiptImportResult
	if err := json.Unmarshal(response.Payload, &result); err != nil {
		return application.BlingReceiptImportResult{ErrorCode: "WORKER_INVALID_RESPONSE", Message: "O serviço local respondeu em formato inválido."}
	}
	return result
}

func unavailableIntegrations() []application.IntegrationStatus {
	return []application.IntegrationStatus{
		{Provider: domain.OriginBling, Status: domain.IntegrationNotConfigured},
		{Provider: domain.OriginNuvemshop, Status: domain.IntegrationNotConfigured},
		{Provider: domain.OriginNuvemPago, Status: domain.IntegrationUnavailable},
	}
}

func requestID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UTC().UnixNano())
}
