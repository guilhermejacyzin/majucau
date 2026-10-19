package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"majucau.local/financial-intelligence/internal/diagnosticslog"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/ipc"
)

var (
	diagnosticRequestIDPattern = regexp.MustCompile(`^[a-z0-9-]{1,48}-[0-9]{1,20}$`)
	diagnosticCodePattern      = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)
	diagnosticStatusPattern    = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,31}$`)
)

type diagnosticRequestHandler struct {
	next   ipc.Handler
	logger *slog.Logger
}

func newWorkerDiagnosticLogger() (*slog.Logger, func() error, error) {
	programData := strings.TrimSpace(os.Getenv("ProgramData"))
	if programData == "" {
		return slog.New(slog.NewJSONHandler(os.Stderr, nil)), func() error { return nil }, os.ErrNotExist
	}
	fileLogger, err := diagnosticslog.New(filepath.Join(programData, "Majucau", "Logs"))
	if err != nil {
		return slog.New(slog.NewJSONHandler(os.Stderr, nil)), func() error { return nil }, err
	}
	return fileLogger.Slog(), fileLogger.Close, nil
}

func (h diagnosticRequestHandler) Handle(ctx context.Context, req ipc.Request) (ipc.Response, error) {
	started := time.Now()
	response, err := h.next.Handle(ctx, req)
	h.logCompletion(req, response, err, started)
	return response, err
}

func (h diagnosticRequestHandler) HandleStream(ctx context.Context, req ipc.Request, dst io.Writer) (ipc.Response, error) {
	started := time.Now()
	streamHandler, ok := h.next.(ipc.StreamingHandler)
	if !ok {
		response := ipc.NewErrorResponse(req.RequestID, "INTERNAL_ERROR", "worker request failed")
		h.logCompletion(req, response, ipc.ErrUnsupportedMethod, started)
		return response, nil
	}
	response, err := streamHandler.HandleStream(ctx, req, dst)
	h.logCompletion(req, response, err, started)
	return response, err
}

func (h diagnosticRequestHandler) logCompletion(req ipc.Request, response ipc.Response, err error, started time.Time) {
	if h.logger == nil {
		return
	}
	requestID := req.RequestID
	if !diagnosticRequestIDPattern.MatchString(requestID) {
		requestID = "invalid"
	}
	errorCode := ""
	if err != nil {
		errorCode = "INTERNAL_ERROR"
	}
	if response.Error != nil {
		errorCode = safeDiagnosticCode(response.Error.Code)
	}
	result := "succeeded"
	if !response.OK || err != nil {
		result = "failed"
	}
	fields := responseSummary{}
	if len(response.Payload) > 0 {
		// The response is bounded by the IPC protocol. Decode only approved
		// summary fields; never serialize or log the payload itself.
		_ = json.Unmarshal(response.Payload, &fields)
	}
	if fields.ErrorCode != "" {
		errorCode = safeDiagnosticCode(fields.ErrorCode)
	}
	if errorCode != "" {
		result = "failed"
	}
	level := slog.LevelInfo
	if result == "failed" {
		level = slog.LevelError
	}
	attrs := []slog.Attr{
		slog.String("event", "worker_request_completed"),
		slog.String("request_id", requestID),
		slog.String("operation", string(req.Method)),
		slog.String("source_system", diagnosticSource(req.Method)),
		slog.String("result", result),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	}
	if errorCode != "" {
		attrs = append(attrs, slog.String("error_code", errorCode))
	}
	if status := safeBlingIntegrationStatus(req.Method, response.Payload); status != "" {
		attrs = append(attrs, slog.String("bling_integration_status", status))
	}
	if status := safeDiagnosticStatus(fields.Status); status != "" {
		attrs = append(attrs, slog.String("batch_status", status))
	}
	if id := safeBatchID(fields.BatchID); id != "" {
		attrs = append(attrs, slog.String("batch_id", id))
	}
	if id := safeBatchID(fields.Receivables.BatchID); id != "" {
		attrs = append(attrs, slog.String("receivables_batch_id", id))
	}
	if id := safeBatchID(fields.Payables.BatchID); id != "" {
		attrs = append(attrs, slog.String("payables_batch_id", id))
	}
	appendCount := func(key string, value *int64) {
		if value != nil && *value >= 0 && *value <= 1_000_000_000 {
			attrs = append(attrs, slog.Int64(key, *value))
		}
	}
	appendCount("pages_read", fields.PagesRead)
	appendCount("records_read", fields.RecordsRead)
	appendCount("records_created", fields.RecordsCreated)
	appendCount("records_updated", fields.RecordsUpdated)
	appendCount("records_failed", fields.RecordsFailed)
	appendCount("records_ignored", fields.IgnoredCount)
	appendCount("receivables_records_read", fields.Receivables.RecordsRead)
	appendCount("payables_records_read", fields.Payables.RecordsRead)
	h.logger.LogAttrs(context.Background(), level, "worker request completed", attrs...)
}

type responseSummary struct {
	ErrorCode      string `json:"error_code"`
	Status         string `json:"status"`
	BatchID        string `json:"batch_id"`
	PagesRead      *int64 `json:"pages_read"`
	RecordsRead    *int64 `json:"records_read"`
	RecordsCreated *int64 `json:"records_created"`
	RecordsUpdated *int64 `json:"records_updated"`
	RecordsFailed  *int64 `json:"records_failed"`
	IgnoredCount   *int64 `json:"ignored_count"`
	Receivables    struct {
		BatchID     string `json:"batch_id"`
		RecordsRead *int64 `json:"records_read"`
	} `json:"receivables"`
	Payables struct {
		BatchID     string `json:"batch_id"`
		RecordsRead *int64 `json:"records_read"`
	} `json:"payables"`
}

type integrationStatusSummary struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

func safeBlingIntegrationStatus(method ipc.Method, payload []byte) string {
	const maxStatusPayload = 4 << 10
	if method != ipc.MethodIntegrationStatus || len(payload) == 0 || len(payload) > maxStatusPayload {
		return ""
	}
	var integrations []integrationStatusSummary
	if err := json.Unmarshal(payload, &integrations); err != nil {
		return ""
	}
	for _, integration := range integrations {
		if domain.Origin(integration.Provider) != domain.OriginBling {
			continue
		}
		status := domain.IntegrationStatus(integration.Status)
		if status.Valid() {
			return string(status)
		}
		return ""
	}
	return ""
}

func diagnosticSource(method ipc.Method) string {
	switch method {
	case ipc.MethodBlingConfigRead, ipc.MethodBlingConfigSave, ipc.MethodBlingOAuthStart, ipc.MethodBlingOAuthStatus, ipc.MethodBlingOAuthTest, ipc.MethodBlingOAuthDisconnect, ipc.MethodBlingSync, ipc.MethodBlingReceiptsPreview, ipc.MethodBlingReceiptsImport:
		return "BLING"
	case ipc.MethodNuvemshopConfigSave:
		return "NUVEMSHOP"
	case ipc.MethodNuvemPagoFuturePreview, ipc.MethodNuvemPagoFutureImport:
		return "NUVEM_PAGO"
	case ipc.MethodGoogleDriveCredentialSave:
		return "GOOGLE_DRIVE"
	default:
		return "LOCAL"
	}
}

func safeDiagnosticCode(value string) string {
	if diagnosticCodePattern.MatchString(value) {
		return value
	}
	return "INTERNAL_ERROR"
}

func safeDiagnosticStatus(value string) string {
	if diagnosticStatusPattern.MatchString(value) {
		return value
	}
	return ""
}

func safeBatchID(value string) string {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return ""
	}
	return parsed.String()
}

var _ ipc.Handler = diagnosticRequestHandler{}
var _ ipc.StreamingHandler = diagnosticRequestHandler{}
