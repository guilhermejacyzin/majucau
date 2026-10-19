package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"majucau.local/financial-intelligence/internal/ipc"
)

func TestDiagnosticLogIncludesOnlyAllowlistedBlingIntegrationStatus(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	payload := []byte(`[
		{"provider":"BLING","status":"CONNECTED","last_success_at":"2026-10-12T17:00:00Z","error_code":"BLING_SAFE_CODE","client_secret":"must-not-be-logged"},
		{"provider":"NUVEMSHOP","status":"AUTH_ERROR","error_code":"OTHER_PROVIDER_CODE"}
	]`)
	handler := diagnosticRequestHandler{logger: logger}
	handler.logCompletion(
		ipc.Request{RequestID: "integrations-123", Method: ipc.MethodIntegrationStatus},
		ipc.Response{OK: true, Payload: payload},
		nil,
		time.Now(),
	)

	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatalf("decode diagnostic event: %v", err)
	}
	if got := event["bling_integration_status"]; got != "CONNECTED" {
		t.Fatalf("bling integration status = %v, want CONNECTED", got)
	}
	for _, forbidden := range []string{"last_success_at", "BLING_SAFE_CODE", "client_secret", "must-not-be-logged", "OTHER_PROVIDER_CODE"} {
		if strings.Contains(output.String(), forbidden) {
			t.Fatalf("diagnostic log contains response detail %q: %s", forbidden, output.String())
		}
	}
}

func TestSafeBlingIntegrationStatusRejectsUnknownProviderStatusOrRequest(t *testing.T) {
	tests := []struct {
		name    string
		method  ipc.Method
		payload string
	}{
		{name: "other provider", method: ipc.MethodIntegrationStatus, payload: `[{"provider":"NUVEMSHOP","status":"CONNECTED"}]`},
		{name: "unknown status", method: ipc.MethodIntegrationStatus, payload: `[{"provider":"BLING","status":"CONNECTED_WITH_SECRET"}]`},
		{name: "wrong method", method: ipc.MethodDashboardSnapshot, payload: `[{"provider":"BLING","status":"CONNECTED"}]`},
		{name: "oversized payload", method: ipc.MethodIntegrationStatus, payload: `[{"provider":"BLING","status":"CONNECTED","padding":"` + strings.Repeat("x", (4<<10)) + `"}]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := safeBlingIntegrationStatus(test.method, []byte(test.payload)); got != "" {
				t.Fatalf("safeBlingIntegrationStatus() = %q, want empty", got)
			}
		})
	}
}

func TestDiagnosticLogDoesNotRecordGoogleDriveCredentialPayload(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	secret := []byte(`{"type":"service_account","private_key":"credential-must-not-appear"}`)
	handler := diagnosticRequestHandler{logger: logger}
	handler.logCompletion(
		ipc.Request{RequestID: "drive-credential-1", Method: ipc.MethodGoogleDriveCredentialSave, Payload: secret},
		ipc.Response{OK: true, Payload: []byte(`{"status":"CONFIGURED"}`)},
		nil,
		time.Now(),
	)
	if strings.Contains(output.String(), string(secret)) || strings.Contains(output.String(), "credential-must-not-appear") || strings.Contains(output.String(), "private_key") {
		t.Fatalf("diagnostic log contains credential material: %s", output.String())
	}
	if !strings.Contains(output.String(), `"source_system":"GOOGLE_DRIVE"`) {
		t.Fatalf("diagnostic log omitted the integration source: %s", output.String())
	}
}
