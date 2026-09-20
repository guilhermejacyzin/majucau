package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/integrations/bling"
	"majucau.local/financial-intelligence/internal/ipc"
)

type workerHandler struct{ health application.StaticHealth }

func newWorkerHandler() workerHandler {
	return workerHandler{health: application.StaticHealth{Service: "majucau-worker", Version: "dev"}}
}
func (h workerHandler) Handle(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	switch req.Method {
	case ipc.MethodHealth:
		return ipc.NewResponse(req.RequestID, h.health.CheckHealth(ctx))
	case ipc.MethodIntegrationStatus:
		return ipc.NewResponse(req.RequestID, []application.IntegrationStatus{
			{Provider: domain.OriginBling, Status: domain.IntegrationNotConfigured},
			{Provider: domain.OriginNuvemshop, Status: domain.IntegrationNotConfigured},
			{Provider: domain.OriginNuvemPago, Status: domain.IntegrationUnavailable},
		})
	case ipc.MethodBlingReceiptsPreview:
		return h.previewBlingReceipts(ctx, req)
	default:
		return ipc.Response{}, ipc.ErrUnsupportedMethod
	}
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
