package bling

import (
	"encoding/json"
	"fmt"
	"testing"

	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/financial"
)

func TestParseBlingPaymentMethodTerms(t *testing.T) {
	record := json.RawMessage(`{"list":{"id":11},"detail":{"id":11,"destino":1,"taxas":{"aliquota":0.25,"valor":1.99,"prazo":30},"utilizaDiasUteis":true}}`)
	terms, err := parseBlingPaymentMethodTerms(record, 11)
	if err != nil {
		t.Fatal(err)
	}
	if terms.FeeRate.String() != "0.2500" || terms.FixedFee.String() != "1.9900" || terms.SettlementDays != 30 || !terms.UseBusinessDays || !terms.FeeRateKnown || !terms.FixedFeeKnown || !terms.SettlementKnown || !terms.BusinessDaysKnown || !terms.DestinationKnown || !terms.ContributesToBlingCashFlowForecast() {
		t.Fatalf("terms = %+v", terms)
	}
	got, err := financial.BlingNetReceivableValue(moneyValue(t, "166.66"), terms)
	if err != nil {
		t.Fatal(err)
	}
	if want := moneyValue(t, "164.25"); got.Compare(want) != 0 {
		t.Fatalf("net value = %s, want %s", got, want)
	}
}

func TestParseBlingPaymentMethodTermsDoesNotInventOnInvalidOrMismatchedDetails(t *testing.T) {
	for _, record := range []json.RawMessage{
		json.RawMessage(`{"list":{"id":11},"detail":{"id":12,"taxas":{"prazo":30}}}`),
		json.RawMessage(`{"list":{"id":11},"detail":{"id":11,"taxas":{"aliquota":"invalida"}}}`),
		json.RawMessage(`{"list":{"id":11},"detail":{"id":11,"taxas":{"prazo":-1}}}`),
		json.RawMessage(`{"list":{"id":11},"detail":{"id":11,"destino":4}}`),
	} {
		if _, err := parseBlingPaymentMethodTerms(record, 11); err == nil {
			t.Fatalf("invalid payment method detail accepted: %s", record)
		}
	}
}

func TestParseBlingPaymentMethodTermsKeepsMissingFieldsUnknown(t *testing.T) {
	terms, err := parseBlingPaymentMethodTerms(json.RawMessage(`{"list":{"id":9},"detail":{"id":9,"taxas":{}}}`), 9)
	if err != nil {
		t.Fatal(err)
	}
	if terms.FeeRate.Compare(domain.ZeroMoney()) != 0 || terms.FixedFee.Compare(domain.ZeroMoney()) != 0 || terms.SettlementDays != 0 || terms.UseBusinessDays || terms.FeeRateKnown || terms.FixedFeeKnown || terms.SettlementKnown || terms.BusinessDaysKnown || terms.DestinationKnown || terms.ContributesToBlingCashFlowForecast() {
		t.Fatalf("empty optional settings were not normalized to zero: %+v", terms)
	}
	if _, err := financial.BlingNetReceivableValue(moneyValue(t, "100.00"), terms); err == nil {
		t.Fatal("missing fee fields were treated as a confirmed zero-fee method")
	}
}

func TestBlingPaymentMethodDestinationControlsCashFlowForecastEligibility(t *testing.T) {
	for _, test := range []struct {
		destination int
		eligible    bool
	}{
		{destination: 1, eligible: true},
		{destination: 2, eligible: false},
		{destination: 3, eligible: false},
	} {
		record := json.RawMessage(fmt.Sprintf(`{"list":{"id":9},"detail":{"id":9,"destino":%d}}`, test.destination))
		terms, err := parseBlingPaymentMethodTerms(record, 9)
		if err != nil {
			t.Fatalf("destination %d parse: %v", test.destination, err)
		}
		if got := terms.ContributesToBlingCashFlowForecast(); got != test.eligible {
			t.Fatalf("destination %d eligibility = %v, want %v", test.destination, got, test.eligible)
		}
	}
}

func moneyValue(t *testing.T, text string) domain.Money {
	t.Helper()
	value, err := domain.ParseMoney(text)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
