package application

import (
	"context"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

type HealthState string

const (
	HealthOK          HealthState = "OK"
	HealthDegraded    HealthState = "DEGRADED"
	HealthUnavailable HealthState = "UNAVAILABLE"
)

type HealthResponse struct {
	Service      string             `json:"service"`
	Version      string             `json:"version"`
	State        HealthState        `json:"state"`
	CheckedAt    time.Time          `json:"checked_at"`
	Dependencies []DependencyHealth `json:"dependencies,omitempty"`
}
type DependencyHealth struct {
	Name   string      `json:"name"`
	State  HealthState `json:"state"`
	Detail string      `json:"detail,omitempty"`
}
type IntegrationStatus struct {
	Provider      domain.Origin            `json:"provider"`
	Status        domain.IntegrationStatus `json:"status"`
	LastSuccessAt *time.Time               `json:"last_success_at,omitempty"`
	LastAttemptAt *time.Time               `json:"last_attempt_at,omitempty"`
	ErrorCode     string                   `json:"error_code,omitempty"`
}

type BlingReceiptImportFile struct {
	Name         string `json:"name"`
	SHA256       string `json:"sha256"`
	ReceiptCount int    `json:"receipt_count"`
	ErrorCount   int    `json:"error_count"`
}

type BlingReceiptImportIssue struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BlingReceiptImportPreview is deliberately summary-only: customer names,
// histories and financial rows never cross the worker/UI boundary in preview.
type BlingReceiptImportPreview struct {
	Files        []BlingReceiptImportFile  `json:"files"`
	ReceiptCount int                       `json:"receipt_count"`
	ErrorCount   int                       `json:"error_count"`
	IgnoredCount int                       `json:"ignored_count"`
	Issues       []BlingReceiptImportIssue `json:"issues,omitempty"`
	ErrorCode    string                    `json:"error_code,omitempty"`
	Message      string                    `json:"message,omitempty"`
}

type BlingReceiptImportResult struct {
	BatchID        string `json:"batch_id,omitempty"`
	Status         string `json:"status,omitempty"`
	RecordsRead    int    `json:"records_read"`
	RecordsCreated int    `json:"records_created"`
	RecordsUpdated int    `json:"records_updated"`
	RecordsFailed  int    `json:"records_failed"`
	IgnoredCount   int    `json:"ignored_count"`
	ErrorCode      string `json:"error_code,omitempty"`
	Message        string `json:"message,omitempty"`
}

type NuvemPagoFutureImportFile struct {
	Name             string `json:"name"`
	SHA256           string `json:"sha256"`
	ReceivableCount  int    `json:"receivable_count"`
	RejectedRowCount int    `json:"rejected_row_count"`
}

type NuvemPagoFutureImportIssue struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NuvemPagoFutureImportPreview is summary-only. Customer names and financial
// rows remain in the worker and are never returned to the frontend preview.
type NuvemPagoFutureImportPreview struct {
	Files           []NuvemPagoFutureImportFile  `json:"files"`
	ReceivableCount int                          `json:"receivable_count"`
	ErrorCount      int                          `json:"error_count"`
	IgnoredCount    int                          `json:"ignored_count"`
	Issues          []NuvemPagoFutureImportIssue `json:"issues,omitempty"`
	ErrorCode       string                       `json:"error_code,omitempty"`
	Message         string                       `json:"message,omitempty"`
}

type NuvemPagoFutureImportResult struct {
	BatchID        string `json:"batch_id,omitempty"`
	Status         string `json:"status,omitempty"`
	RecordsRead    int    `json:"records_read"`
	RecordsCreated int    `json:"records_created"`
	RecordsUpdated int    `json:"records_updated"`
	RecordsFailed  int    `json:"records_failed"`
	IgnoredCount   int    `json:"ignored_count"`
	ErrorCode      string `json:"error_code,omitempty"`
	Message        string `json:"message,omitempty"`
}

// BlingConfigRequest crosses the authenticated local pipe only. The worker
// consumes ClientSecret immediately and never echoes it back.
type BlingConfigRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	ClientSecret string `json:"client_secret"`
}

type BlingConfigResponse struct {
	ClientID         string `json:"client_id,omitempty"`
	RedirectURI      string `json:"redirect_uri,omitempty"`
	SecretConfigured bool   `json:"secret_configured"`
	Status           string `json:"status,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
	Message          string `json:"message,omitempty"`
}

// NuvemshopConfigRequest carries the editable app metadata and transient
// client secret to the worker. The secret is never echoed or persisted here.
type NuvemshopConfigRequest struct {
	AppID        string `json:"app_id"`
	RedirectURI  string `json:"redirect_uri"`
	ClientSecret string `json:"client_secret"`
}

type NuvemshopConfigResponse struct {
	AppID            string `json:"app_id,omitempty"`
	RedirectURI      string `json:"redirect_uri,omitempty"`
	SecretConfigured bool   `json:"secret_configured"`
	Status           string `json:"status,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
	Message          string `json:"message,omitempty"`
}

type BlingOAuthStartResponse struct {
	SessionID        string `json:"session_id,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
	Status           string `json:"status,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`
	Message          string `json:"message,omitempty"`
}

type BlingOAuthStatusRequest struct {
	SessionID string `json:"session_id"`
}

type BlingOAuthStatusResponse struct {
	SessionID string `json:"session_id,omitempty"`
	Status    string `json:"status,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

type BlingOAuthTestResponse struct {
	Status          string `json:"status,omitempty"`
	PageRecordCount int    `json:"page_record_count"`
	ErrorCode       string `json:"error_code,omitempty"`
	Message         string `json:"message,omitempty"`
}

type BlingSyncRequest struct {
	Page             int    `json:"page,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	DueDateFrom      string `json:"due_date_from,omitempty"`
	DueDateTo        string `json:"due_date_to,omitempty"`
	ReceivedDateFrom string `json:"received_date_from,omitempty"`
	ReceivedDateTo   string `json:"received_date_to,omitempty"`
	PaymentDateFrom  string `json:"payment_date_from,omitempty"`
	PaymentDateTo    string `json:"payment_date_to,omitempty"`
	Status           string `json:"status,omitempty"`
}

type BlingSyncResourceResult struct {
	Status      string `json:"status,omitempty"`
	BatchID     string `json:"batch_id,omitempty"`
	PagesRead   int    `json:"pages_read"`
	RecordsRead int    `json:"records_read"`
}

type BlingSyncResponse struct {
	Status      string                  `json:"status,omitempty"`
	Receivables BlingSyncResourceResult `json:"receivables"`
	Payables    BlingSyncResourceResult `json:"payables"`
	ErrorCode   string                  `json:"error_code,omitempty"`
	Message     string                  `json:"message,omitempty"`
}

// DashboardMetric is a fail-closed, read-only projection of one normalized
// ledger metric. Value is serialized as text so the worker/UI boundary never
// rounds PostgreSQL numeric values through a binary floating-point type.
type DashboardMetric struct {
	Value        string `json:"value,omitempty"`
	Count        int64  `json:"count,omitempty"`
	State        string `json:"state"`
	SourceSystem string `json:"source_system,omitempty"`
}

// DashboardSnapshot contains only metrics that are already normalized in the
// local ledger. It deliberately does not manufacture cash balances, DRE,
// inventory or forecast values from RAW payloads.
type DashboardSnapshot struct {
	AsOf               time.Time       `json:"as_of"`
	DataState          string          `json:"data_state"`
	Receivables        DashboardMetric `json:"receivables"`
	FutureB2C          DashboardMetric `json:"future_b2c"`
	ReceiptsMonth      DashboardMetric `json:"receipts_month"`
	ReceivablesOverdue DashboardMetric `json:"receivables_overdue"`
	Payables           DashboardMetric `json:"payables"`
	PayablesDueToday   DashboardMetric `json:"payables_due_today"`
	PayablesOverdue    DashboardMetric `json:"payables_overdue"`
	PaymentsMonth      DashboardMetric `json:"payments_month"`
	ErrorCode          string          `json:"error_code,omitempty"`
	Message            string          `json:"message,omitempty"`
}

type DashboardSnapshotReader interface {
	ReadDashboardSnapshot(context.Context) (DashboardSnapshot, error)
}
type HealthChecker interface {
	CheckHealth(context.Context) HealthResponse
}
type IntegrationStatusReader interface {
	IntegrationStatuses(context.Context) ([]IntegrationStatus, error)
}

type StaticHealth struct {
	Service, Version string
	Dependencies     []DependencyHealth
}

func (h StaticHealth) CheckHealth(_ context.Context) HealthResponse {
	state := HealthOK
	for _, d := range h.Dependencies {
		if d.State == HealthUnavailable {
			state = HealthUnavailable
			break
		}
		if d.State == HealthDegraded {
			state = HealthDegraded
		}
	}
	return HealthResponse{Service: h.Service, Version: h.Version, State: state, CheckedAt: time.Now().UTC(), Dependencies: append([]DependencyHealth(nil), h.Dependencies...)}
}

