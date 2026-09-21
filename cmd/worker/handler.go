package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/integrations/bling"
	"majucau.local/financial-intelligence/internal/integrations/nuvempago"
	"majucau.local/financial-intelligence/internal/integrations/nuvemshop"
	"majucau.local/financial-intelligence/internal/ipc"
)

type receiptImporter interface {
	Import(context.Context, string) (bling.ReceiptImportResult, error)
	Close()
}

type futureImporter interface {
	Import(context.Context, string) (nuvempago.FutureImportResult, error)
	Close()
}

type blingConfigSaver interface {
	Save(context.Context, bling.BlingConfigInput) (bling.BlingConfigResult, error)
}

type nuvemshopConfigSaver interface {
	Save(context.Context, nuvemshop.ConfigInput) (nuvemshop.ConfigResult, error)
}

type blingOAuthService interface {
	Start(context.Context) (bling.BlingOAuthStartResult, error)
	Status(context.Context, string) (bling.BlingOAuthStatusResult, error)
	Test(context.Context) (bling.BlingOAuthTestResult, error)
}

type blingRawSyncer interface {
	Sync(context.Context, bling.ReceivablesFilter) (bling.BlingRawSyncResult, error)
}

type workerHandler struct {
	health          application.StaticHealth
	receiptImporter receiptImporter
	futureImporter  futureImporter
	blingConfig     blingConfigSaver
	nuvemshopConfig nuvemshopConfigSaver
	blingOAuth      blingOAuthService
	blingSync       blingRawSyncer
	statusReader    application.IntegrationStatusReader
}

func newWorkerHandler() workerHandler {
	pool := newDatabasePoolFromEnvironment()
	var importer receiptImporter
	var future futureImporter
	var configSaver blingConfigSaver
	var nuvemshopSaver nuvemshopConfigSaver
	var oauthService blingOAuthService
	var rawSync blingRawSyncer
	var statusReader application.IntegrationStatusReader
	if pool != nil {
		importer = bling.NewReceiptImportService(pool)
		future = nuvempago.NewFutureImportService(pool)
		statusReader = newIntegrationStatusReader(pool)
		if store := newWorkerSecretStore(); store != nil {
			configSaver = bling.NewBlingCredentialService(pool, store)
			nuvemshopSaver = nuvemshop.NewCredentialService(pool, store)
			oauthService = bling.NewBlingOAuthService(pool, store)
			rawSync = oauthService.(blingRawSyncer)
		}
	}
	return workerHandler{health: application.StaticHealth{Service: "majucau-worker", Version: "dev"}, receiptImporter: importer, futureImporter: future, blingConfig: configSaver, nuvemshopConfig: nuvemshopSaver, blingOAuth: oauthService, blingSync: rawSync, statusReader: statusReader}
}
func (h workerHandler) Handle(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	switch req.Method {
	case ipc.MethodHealth:
		return ipc.NewResponse(req.RequestID, h.health.CheckHealth(ctx))
	case ipc.MethodIntegrationStatus:
		return h.integrationStatus(ctx, req)
	case ipc.MethodBlingConfigSave:
		return h.saveBlingConfig(ctx, req)
	case ipc.MethodNuvemshopConfigSave:
		return h.saveNuvemshopConfig(ctx, req)
	case ipc.MethodBlingOAuthStart:
		return h.startBlingOAuth(ctx, req)
	case ipc.MethodBlingOAuthStatus:
		return h.statusBlingOAuth(ctx, req)
	case ipc.MethodBlingOAuthTest:
		return h.testBlingOAuth(ctx, req)
	case ipc.MethodBlingSync:
		return h.syncBling(ctx, req)
	case ipc.MethodBlingReceiptsPreview:
		return h.previewBlingReceipts(ctx, req)
	case ipc.MethodBlingReceiptsImport:
		return h.importBlingReceipts(ctx, req)
	case ipc.MethodNuvemPagoFuturePreview:
		return h.previewNuvemPagoFuture(ctx, req)
	case ipc.MethodNuvemPagoFutureImport:
		return h.importNuvemPagoFuture(ctx, req)
	default:
		return ipc.Response{}, ipc.ErrUnsupportedMethod
	}
}

func (h workerHandler) saveNuvemshopConfig(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input application.NuvemshopConfigRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil {
		return ipc.NewErrorResponse(req.RequestID, "NUVEMSHOP_CONFIG_INVALID", "Os dados da configuração da Nuvemshop são inválidos."), nil
	}
	if h.nuvemshopConfig == nil {
		return ipc.NewErrorResponse(req.RequestID, "NUVEMSHOP_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.nuvemshopConfig.Save(ctx, nuvemshop.ConfigInput{AppID: input.AppID, RedirectURI: input.RedirectURI, ClientSecret: input.ClientSecret})
	if err != nil {
		code, message := "NUVEMSHOP_CONFIG_SAVE_FAILED", "Não foi possível salvar a configuração da Nuvemshop."
		switch {
		case errors.Is(err, nuvemshop.ErrNuvemshopCredentialValidation):
			code, message = "NUVEMSHOP_CONFIG_INVALID", "Confira o App ID e use o Redirect URI HTTPS do relay homologado."
		case errors.Is(err, nuvemshop.ErrNuvemshopCredentialMissing):
			code, message = "NUVEMSHOP_SECRET_REQUIRED", "Informe o Client Secret na primeira configuração."
		case errors.Is(err, nuvemshop.ErrNuvemshopCredentialVault):
			code, message = "NUVEMSHOP_VAULT_UNAVAILABLE", "O cofre seguro do Windows não está disponível."
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			code, message = "NUVEMSHOP_CONFIG_CANCELLED", "A configuração foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	return ipc.NewResponse(req.RequestID, application.NuvemshopConfigResponse{AppID: result.AppID, RedirectURI: result.RedirectURI, SecretConfigured: result.SecretConfigured, Status: result.Status})
}

func (h workerHandler) syncBling(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input application.BlingSyncRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || !hasBlingSyncFilter(input) {
		return ipc.NewErrorResponse(req.RequestID, "BLING_SYNC_FILTER_REQUIRED", "Informe uma janela de datas ou situação antes de sincronizar."), nil
	}
	if h.blingSync == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.blingSync.Sync(ctx, bling.ReceivablesFilter{Page: input.Page, Limit: input.Limit, DueDateFrom: input.DueDateFrom, DueDateTo: input.DueDateTo, ReceivedDateFrom: input.ReceivedDateFrom, ReceivedDateTo: input.ReceivedDateTo, PaymentDateFrom: input.PaymentDateFrom, PaymentDateTo: input.PaymentDateTo, Status: input.Status})
	response := application.BlingSyncResponse{Status: result.Status, Receivables: syncResourceResult(result.Receivables), Payables: syncResourceResult(result.Payables)}
	if err != nil {
		response.ErrorCode = blingSyncErrorCode(err)
		response.Message = blingSyncErrorMessage(err)
	}
	return ipc.NewResponse(req.RequestID, response)
}

func hasBlingSyncFilter(input application.BlingSyncRequest) bool {
	return strings.TrimSpace(input.DueDateFrom) != "" || strings.TrimSpace(input.DueDateTo) != "" || strings.TrimSpace(input.ReceivedDateFrom) != "" || strings.TrimSpace(input.ReceivedDateTo) != "" || strings.TrimSpace(input.PaymentDateFrom) != "" || strings.TrimSpace(input.PaymentDateTo) != "" || strings.TrimSpace(input.Status) != ""
}

func syncResourceResult(result bling.APISyncResult) application.BlingSyncResourceResult {
	return application.BlingSyncResourceResult{Status: result.Status, BatchID: result.BatchID, PagesRead: result.PagesRead, RecordsRead: result.RecordsRead}
}

func blingSyncErrorCode(err error) string {
	switch {
	case errors.Is(err, bling.ErrBlingAPISyncDatabaseUnavailable):
		return "BLING_DATABASE_NOT_CONFIGURED"
	case errors.Is(err, bling.ErrBlingAPISyncConnectionMissing):
		return "BLING_CONNECTION_NOT_CONFIGURED"
	case errors.Is(err, bling.ErrBlingAPISyncPageLimit):
		return "BLING_SYNC_PAGE_LIMIT"
	case errors.Is(err, bling.ErrBlingAPIUnauthorized):
		return "BLING_API_UNAUTHORIZED"
	case errors.Is(err, bling.ErrBlingAPIRateLimited):
		return "BLING_API_RATE_LIMITED"
	case errors.Is(err, bling.ErrBlingAPIUnavailable):
		return "BLING_API_UNAVAILABLE"
	case errors.Is(err, bling.ErrBlingAPISchemaMismatch):
		return "BLING_SCHEMA_MISMATCH"
	case errors.Is(err, bling.ErrBlingAPIInvalidResponse):
		return "BLING_INVALID_RESPONSE"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "BLING_SYNC_CANCELLED"
	default:
		return "BLING_SYNC_FAILED"
	}
}

func blingSyncErrorMessage(err error) string {
	switch blingSyncErrorCode(err) {
	case "BLING_DATABASE_NOT_CONFIGURED":
		return "O banco local ainda não está configurado para sincronizar o Bling."
	case "BLING_CONNECTION_NOT_CONFIGURED":
		return "Configure e autorize o Bling antes de sincronizar."
	case "BLING_SYNC_PAGE_LIMIT":
		return "A sincronização atingiu o limite de segurança de páginas; reduza a janela."
	case "BLING_API_UNAUTHORIZED":
		return "A autorização do Bling expirou ou foi revogada; conecte novamente."
	case "BLING_API_RATE_LIMITED":
		return "O Bling limitou temporariamente as consultas; aguarde e tente novamente."
	case "BLING_API_UNAVAILABLE":
		return "O Bling não respondeu após as tentativas seguras; o último dado válido foi preservado."
	case "BLING_SCHEMA_MISMATCH":
		return "O formato retornado pelo Bling não coincide com o contrato homologado; nenhum dado novo foi confirmado."
	case "BLING_INVALID_RESPONSE":
		return "A resposta do Bling não pôde ser validada; o último dado válido foi preservado."
	case "BLING_SYNC_CANCELLED":
		return "A sincronização foi cancelada antes do commit do lote."
	default:
		return "A sincronização do Bling falhou; o último dado válido foi preservado."
	}
}

func (h workerHandler) integrationStatus(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	if h.statusReader == nil {
		return ipc.NewResponse(req.RequestID, defaultIntegrationStatuses())
	}
	statuses, err := h.statusReader.IntegrationStatuses(ctx)
	if err != nil {
		return ipc.NewErrorResponse(req.RequestID, "INTEGRATION_STATUS_UNAVAILABLE", "O status das integrações não pôde ser consultado."), nil
	}
	return ipc.NewResponse(req.RequestID, statuses)
}

func defaultIntegrationStatuses() []application.IntegrationStatus {
	return []application.IntegrationStatus{
		{Provider: domain.OriginBling, Status: domain.IntegrationNotConfigured},
		{Provider: domain.OriginNuvemshop, Status: domain.IntegrationNotConfigured},
		{Provider: domain.OriginNuvemPago, Status: domain.IntegrationUnavailable},
	}
}

func (h workerHandler) startBlingOAuth(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	if h.blingOAuth == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.blingOAuth.Start(ctx)
	if err != nil {
		return ipc.NewErrorResponse(req.RequestID, blingOAuthErrorCode(err), blingOAuthErrorMessage(err)), nil
	}
	return ipc.NewResponse(req.RequestID, application.BlingOAuthStartResponse{SessionID: result.SessionID, AuthorizationURL: result.AuthorizationURL, Status: result.Status})
}

func (h workerHandler) statusBlingOAuth(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input application.BlingOAuthStatusRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || strings.TrimSpace(input.SessionID) == "" {
		return ipc.NewErrorResponse(req.RequestID, "BLING_OAUTH_SESSION_INVALID", "A sessão de autorização não é válida."), nil
	}
	if h.blingOAuth == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.blingOAuth.Status(ctx, input.SessionID)
	if err != nil {
		return ipc.NewErrorResponse(req.RequestID, blingOAuthErrorCode(err), blingOAuthErrorMessage(err)), nil
	}
	return ipc.NewResponse(req.RequestID, application.BlingOAuthStatusResponse{SessionID: result.SessionID, Status: result.Status, ErrorCode: result.ErrorCode, Message: result.Message})
}

func (h workerHandler) testBlingOAuth(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	if len(req.Payload) != 0 && string(req.Payload) != "null" && string(req.Payload) != "{}" {
		return ipc.NewErrorResponse(req.RequestID, "BLING_TEST_INVALID", "O teste do Bling não recebeu dados válidos."), nil
	}
	if h.blingOAuth == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.blingOAuth.Test(ctx)
	if err != nil {
		return ipc.NewErrorResponse(req.RequestID, blingOAuthErrorCode(err), blingOAuthErrorMessage(err)), nil
	}
	return ipc.NewResponse(req.RequestID, application.BlingOAuthTestResponse{Status: result.Status, PageRecordCount: result.PageRecordCount})
}

func blingOAuthErrorCode(err error) string {
	switch {
	case errors.Is(err, bling.ErrBlingCredentialMissing):
		return "BLING_SECRET_REQUIRED"
	case errors.Is(err, bling.ErrBlingCredentialVault):
		return "BLING_VAULT_UNAVAILABLE"
	case errors.Is(err, bling.ErrBlingOAuthConfiguration):
		return "BLING_OAUTH_CONFIG_INVALID"
	case errors.Is(err, bling.ErrBlingOAuthSession):
		return "BLING_OAUTH_SESSION_INVALID"
	case errors.Is(err, bling.ErrBlingOAuthNotConnected):
		return "BLING_NOT_CONNECTED"
	case errors.Is(err, bling.ErrBlingOAuthStorage):
		return "BLING_TOKEN_STORE_FAILED"
	default:
		return "BLING_OAUTH_FAILED"
	}
}

func blingOAuthErrorMessage(err error) string {
	switch blingOAuthErrorCode(err) {
	case "BLING_SECRET_REQUIRED":
		return "Informe e salve o Client Secret antes de conectar."
	case "BLING_VAULT_UNAVAILABLE":
		return "O cofre seguro do Windows não está disponível."
	case "BLING_OAUTH_CONFIG_INVALID":
		return "Use um Redirect URI loopback registrado no Bling: http://127.0.0.1:porta/caminho."
	case "BLING_OAUTH_SESSION_INVALID":
		return "A sessão de autorização expirou. Inicie a conexão novamente."
	case "BLING_NOT_CONNECTED":
		return "Autorize o Bling antes de testar a conexão."
	case "BLING_TOKEN_STORE_FAILED":
		return "O token não pôde ser protegido no cofre do Windows."
	default:
		return "Não foi possível concluir a operação do Bling."
	}
}

func (h workerHandler) saveBlingConfig(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input application.BlingConfigRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_CONFIG_INVALID", "Os dados da configuração do Bling são inválidos."), nil
	}
	if h.blingConfig == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_VAULT_UNAVAILABLE", "O cofre seguro e o banco local ainda não estão disponíveis."), nil
	}
	result, err := h.blingConfig.Save(ctx, bling.BlingConfigInput{ClientID: input.ClientID, RedirectURI: input.RedirectURI, ClientSecret: input.ClientSecret})
	if err != nil {
		code, message := "BLING_CONFIG_SAVE_FAILED", "Não foi possível salvar a configuração do Bling."
		switch {
		case errors.Is(err, bling.ErrBlingCredentialValidation):
			code, message = "BLING_CONFIG_INVALID", "Confira o Client ID e o Redirect URI do Bling."
		case errors.Is(err, bling.ErrBlingCredentialMissing):
			code, message = "BLING_SECRET_REQUIRED", "Informe o Client Secret na primeira configuração."
		case errors.Is(err, bling.ErrBlingCredentialVault):
			code, message = "BLING_VAULT_UNAVAILABLE", "O cofre seguro do Windows não está disponível."
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			code, message = "BLING_CONFIG_CANCELLED", "A configuração foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	return ipc.NewResponse(req.RequestID, application.BlingConfigResponse{ClientID: result.ClientID, RedirectURI: result.RedirectURI, SecretConfigured: result.SecretConfigured, Status: result.Status})
}

func (h workerHandler) importBlingReceipts(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input blingReceiptsPreviewRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || strings.TrimSpace(input.Folder) == "" {
		return ipc.NewErrorResponse(req.RequestID, "BLING_IMPORT_FOLDER_REQUIRED", "Informe a pasta dos relatórios CSV do Bling."), nil
	}
	if h.receiptImporter == nil {
		return ipc.NewErrorResponse(req.RequestID, "BLING_DATABASE_NOT_CONFIGURED", "O banco local ainda não está configurado para gravar os recebimentos."), nil
	}
	result, err := h.receiptImporter.Import(ctx, input.Folder)
	if err != nil {
		code := "BLING_IMPORT_FAILED"
		message := "Não foi possível gravar os recebimentos do Bling."
		switch {
		case errors.Is(err, bling.ErrReceiptDatabaseUnavailable):
			code, message = "BLING_DATABASE_NOT_CONFIGURED", "O banco local ainda não está configurado para gravar os recebimentos."
		case errors.Is(err, bling.ErrBlingConnectionMissing):
			code, message = "BLING_CONNECTION_NOT_CONFIGURED", "Configure a conexão do Bling antes de importar os recebimentos."
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			code, message = "BLING_IMPORT_CANCELLED", "A importação foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	return ipc.NewResponse(req.RequestID, application.BlingReceiptImportResult{BatchID: result.BatchID, Status: result.Status, RecordsRead: result.RecordsRead, RecordsCreated: result.RecordsCreated, RecordsUpdated: result.RecordsUpdated, RecordsFailed: result.RecordsFailed, IgnoredCount: result.IgnoredCount})
}

type blingReceiptsPreviewRequest struct {
	Folder string `json:"folder"`
}

func (h workerHandler) previewBlingReceipts(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input blingReceiptsPreviewRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || strings.TrimSpace(input.Folder) == "" {
		return ipc.NewErrorResponse(req.RequestID, "BLING_IMPORT_FOLDER_REQUIRED", "Informe a pasta dos relatórios CSV do Bling."), nil
	}
	report, err := bling.ImportReceiptsFolder(ctx, input.Folder)
	if err != nil {
		code := "BLING_IMPORT_FAILED"
		message := "Não foi possível ler a pasta dos relatórios do Bling."
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			code = "BLING_IMPORT_CANCELLED"
			message = "A leitura da pasta foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	preview := application.BlingReceiptImportPreview{
		ReceiptCount: len(report.Receipts), ErrorCount: len(report.Errors), IgnoredCount: len(report.Ignored),
		Files:  make([]application.BlingReceiptImportFile, 0, len(report.Files)),
		Issues: make([]application.BlingReceiptImportIssue, 0, minInt(len(report.Errors), 50)),
	}
	for _, file := range report.Files {
		preview.Files = append(preview.Files, application.BlingReceiptImportFile{Name: file.Name, SHA256: file.SHA256, ReceiptCount: file.ReceiptCount, ErrorCount: file.ErrorCount})
	}
	for index, issue := range report.Errors {
		if index >= 50 {
			break
		}
		preview.Issues = append(preview.Issues, application.BlingReceiptImportIssue{File: issue.File, Line: issue.Line, Code: issue.Code, Message: issue.Message})
	}
	return ipc.NewResponse(req.RequestID, preview)
}

func (h workerHandler) previewNuvemPagoFuture(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input blingReceiptsPreviewRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || strings.TrimSpace(input.Folder) == "" {
		return ipc.NewErrorResponse(req.RequestID, "NUVEM_PAGO_FUTURE_FOLDER_REQUIRED", "Informe a pasta dos recebimentos futuros do Nuvem Pago."), nil
	}
	report, err := nuvempago.ImportFutureFolder(ctx, input.Folder)
	if err != nil {
		code := "NUVEM_PAGO_FUTURE_PREVIEW_FAILED"
		message := "Não foi possível ler a pasta de recebimentos futuros do Nuvem Pago."
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			code = "NUVEM_PAGO_FUTURE_CANCELLED"
			message = "A leitura da pasta foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	preview := application.NuvemPagoFutureImportPreview{
		ReceivableCount: len(report.Receivables), ErrorCount: len(report.Errors), IgnoredCount: len(report.Ignored),
		Files:  make([]application.NuvemPagoFutureImportFile, 0, len(report.Files)),
		Issues: make([]application.NuvemPagoFutureImportIssue, 0, minInt(len(report.Errors), 50)),
	}
	for _, file := range report.Files {
		preview.Files = append(preview.Files, application.NuvemPagoFutureImportFile{Name: file.Name, SHA256: file.SHA256, ReceivableCount: file.ReceivableCount, RejectedRowCount: file.RejectedRowCount})
	}
	for index, issue := range report.Errors {
		if index >= 50 {
			break
		}
		preview.Issues = append(preview.Issues, application.NuvemPagoFutureImportIssue{File: issue.File, Line: issue.Line, Code: issue.Code, Message: issue.Message})
	}
	return ipc.NewResponse(req.RequestID, preview)
}

func (h workerHandler) importNuvemPagoFuture(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	var input blingReceiptsPreviewRequest
	if err := json.Unmarshal(req.Payload, &input); err != nil || strings.TrimSpace(input.Folder) == "" {
		return ipc.NewErrorResponse(req.RequestID, "NUVEM_PAGO_FUTURE_FOLDER_REQUIRED", "Informe a pasta dos recebimentos futuros do Nuvem Pago."), nil
	}
	if h.futureImporter == nil {
		return ipc.NewErrorResponse(req.RequestID, "NUVEM_PAGO_DATABASE_NOT_CONFIGURED", "O banco local ainda não está configurado para gravar os recebimentos futuros."), nil
	}
	result, err := h.futureImporter.Import(ctx, input.Folder)
	if err != nil {
		code := "NUVEM_PAGO_FUTURE_IMPORT_FAILED"
		message := "Não foi possível gravar os recebimentos futuros do Nuvem Pago."
		switch {
		case errors.Is(err, nuvempago.ErrFutureDatabaseUnavailable):
			code, message = "NUVEM_PAGO_DATABASE_NOT_CONFIGURED", "O banco local ainda não está configurado para gravar os recebimentos futuros."
		case errors.Is(err, nuvempago.ErrNuvemPagoConnectionMissing):
			code, message = "NUVEM_PAGO_CONNECTION_NOT_CONFIGURED", "Configure a conexão do Nuvem Pago antes de importar os recebimentos futuros."
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			code, message = "NUVEM_PAGO_FUTURE_CANCELLED", "A importação foi cancelada antes de terminar."
		}
		return ipc.NewErrorResponse(req.RequestID, code, message), nil
	}
	return ipc.NewResponse(req.RequestID, application.NuvemPagoFutureImportResult{BatchID: result.BatchID, Status: result.Status, RecordsRead: result.RecordsRead, RecordsCreated: result.RecordsCreated, RecordsUpdated: result.RecordsUpdated, RecordsFailed: result.RecordsFailed, IgnoredCount: result.IgnoredCount})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

