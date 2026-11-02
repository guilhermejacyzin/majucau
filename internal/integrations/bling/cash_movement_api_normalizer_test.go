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
		wantError string
	}{
		{
			name:      "debit is an outflow",
			payload:   `{"id":1,"debCred":"D","situacao":"R","valor":125.50,"data":"2026-11-02","contaFinanceira":{"id":8,"descricao":"Conta teste"}}`,
			direction: "OUTFLOW",
		},
		{
			name:      "credit is an inflow",
			payload:   `{"id":2,"debCred":"C","situacao":"R","valor":75.25,"data":"2026-11-02","contaFinanceira":{"id":8,"descricao":"Conta teste"}}`,
			direction: "INFLOW",
		},
		{
			name:      "negative magnitude rejected without changing sign semantics",
			payload:   `{"id":3,"debCred":"D","situacao":"R","valor":-10,"data":"2026-11-02","contaFinanceira":{"id":8,"descricao":"Conta teste"}}`,
			wantError: "BLING_CASH_MOVEMENT_AMOUNT_INVALID",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCashMovementAPIRecord(json.RawMessage(test.payload))
			if test.wantError != "" {
				if err == nil {
					t.Fatal("expected invalid movement amount")
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
			if got.Direction != test.direction || !got.Amount.Valid || got.Amount.NaN || got.Amount.Int == nil {
				t.Fatalf("normalized movement = %+v, want direction %s and exact positive magnitude", got, test.direction)
			}
		})
	}
}

func TestParseCashMovementAPIRecordWithoutAccountCanBeIdentifiedAsPartial(t *testing.T) {
	got, err := parseCashMovementAPIRecord(json.RawMessage(`{"id":4,"debCred":"D","situacao":"R","valor":10,"data":"2026-11-02"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Direction != "OUTFLOW" || got.AccountSourceID != nil || got.AccountName != nil {
		t.Fatalf("normalized movement = %+v, want valid outflow with account awaiting identification", got)
	}
}

func TestParseCashMovementAPIRecordPreservesAccountIDWithoutDescription(t *testing.T) {
	got, err := parseCashMovementAPIRecord(json.RawMessage(`{"id":5,"debCred":"D","situacao":"R","valor":10,"data":"2026-11-02","contaFinanceira":{"id":9}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.AccountSourceID == nil || *got.AccountSourceID != "9" || got.AccountName != nil {
		t.Fatalf("account mapping = %v / %v, want source ID 9 retained without a display name", got.AccountSourceID, got.AccountName)
	}
}
