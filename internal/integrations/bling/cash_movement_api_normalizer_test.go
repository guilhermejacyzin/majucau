package bling

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseCashMovementAPIRecordKeepsDirectionSeparateFromMagnitude(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		direction string
		status    string
		wantError string
	}{
		{
			name:      "debit negative value becomes outflow magnitude",
			payload:   `{"id":1,"debCred":"D","situacao":"R","valor":-125.50,"data":"2026-11-02","contaFinanceira":{"id":8,"descricao":"Conta teste"}}`,
			direction: "OUTFLOW",
			status:    "R",
		},
		{
			name:      "credit positive value becomes inflow magnitude",
			payload:   `{"id":2,"debCred":"C","situacao":"R","valor":75.25,"data":"2026-11-02","contaFinanceira":{"id":8,"descricao":"Conta teste"}}`,
			direction: "INFLOW",
			status:    "R",
		},
		{
			name:      "optional empty status is preserved without classification",
			payload:   `{"id":6,"debCred":"D","situacao":"","valor":-10,"data":"2026-11-02"}`,
			direction: "OUTFLOW",
		},
		{
			name:      "new provider status remains source metadata",
			payload:   `{"id":7,"debCred":"D","situacao":"Pendente","valor":-10,"data":"2026-11-02"}`,
			direction: "OUTFLOW",
			status:    "Pendente",
		},
		{
			name:      "debit positive value is inconsistent",
			payload:   `{"id":3,"debCred":"D","situacao":"R","valor":10,"data":"2026-11-02"}`,
			wantError: "BLING_CASH_MOVEMENT_DIRECTION_AMOUNT_MISMATCH",
		},
		{
			name:      "credit negative value is inconsistent",
			payload:   `{"id":8,"debCred":"C","situacao":"R","valor":-10,"data":"2026-11-02"}`,
			wantError: "BLING_CASH_MOVEMENT_DIRECTION_AMOUNT_MISMATCH",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCashMovementAPIRecord(json.RawMessage(test.payload))
			if test.wantError != "" {
				if err == nil {
					t.Fatal("expected movement field error")
				}
				var fieldErr *cashMovementSchemaFieldError
				if !errors.As(err, &fieldErr) || fieldErr.code != test.wantError {
					t.Fatalf("error = %v, want reviewed code %s", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Direction != test.direction || !got.Amount.Valid || got.Amount.NaN || got.Amount.Int == nil || got.Amount.Int.Sign() <= 0 {
				t.Fatalf("normalized movement = %+v, want direction %s and exact positive magnitude", got, test.direction)
			}
			if got.SourceStatus != test.status {
				t.Fatalf("source status = %q, want source value %q", got.SourceStatus, test.status)
			}
		})
	}
}

func TestParseCashMovementAPIRecordWithoutAccountCanBeIdentifiedAsPartial(t *testing.T) {
	got, err := parseCashMovementAPIRecord(json.RawMessage(`{"id":4,"debCred":"D","situacao":"R","valor":-10,"data":"2026-11-02"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Direction != "OUTFLOW" || got.AccountSourceID != nil || got.AccountName != nil {
		t.Fatalf("normalized movement = %+v, want valid outflow with account awaiting identification", got)
	}
}

func TestCashMovementStatusRequiresExplicitRegularStatusForConfirmedTotals(t *testing.T) {
	if !cashMovementStatusIsKnown("R") {
		t.Fatal("regular Bling status should be considered known")
	}
	for _, status := range []string{"", " ", "Pendente", "E"} {
		if cashMovementStatusIsKnown(status) {
			t.Errorf("status %q should remain unconfirmed until its meaning is mapped", status)
		}
	}
}

func TestParseCashMovementAPIRecordPreservesAccountIDWithoutDescription(t *testing.T) {
	got, err := parseCashMovementAPIRecord(json.RawMessage(`{"id":5,"debCred":"D","situacao":"R","valor":-10,"data":"2026-11-02","contaFinanceira":{"id":9}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.AccountSourceID == nil || *got.AccountSourceID != "9" || got.AccountName != nil {
		t.Fatalf("account mapping = %v / %v, want source ID 9 retained without a display name", got.AccountSourceID, got.AccountName)
	}
}
