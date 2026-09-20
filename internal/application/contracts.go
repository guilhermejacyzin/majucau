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
