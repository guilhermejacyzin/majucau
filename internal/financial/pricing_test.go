package financial

import (
	"errors"
	"testing"

	"majucau.local/financial-intelligence/internal/domain"
)

func TestNuvemPagoCardPricingUsesD30AndApprovedRates(t *testing.T) {
	for _, tc := range []struct {
		installments int
		fee          string
	}{
		{1, "26.2500"},
		{2, "45.2500"},
		{3, "54.7500"},
	} {
		result, err := CalculateNuvemPagoPricing(PaymentCard, tc.installments, mustMoney(t, "1000"))
		if err != nil {
			t.Fatal(err)
		}
		if result.Fee.String() != tc.fee || result.EstimatedNet.String() == "" || result.Rule.ReceiptDays == nil || *result.Rule.ReceiptDays != 30 {
			t.Fatalf("installments=%d result=%#v", tc.installments, result)
		}
	}
}

func TestNuvemPagoBoletoAndPIXExceptions(t *testing.T) {
	boleto, err := CalculateNuvemPagoPricing(PaymentBoleto, 0, mustMoney(t, "1000"))
	if err != nil || boleto.Fee.String() != "2.3900" || boleto.Rule.ReceiptDays == nil || *boleto.Rule.ReceiptDays != 2 {
		t.Fatalf("boleto: %#v %v", boleto, err)
	}
	pix, err := CalculateNuvemPagoPricing(PaymentPIX, 0, mustMoney(t, "1000"))
	if err != nil || pix.Fee.String() != "9.9000" || !pix.Rule.ReceiptImmediate {
		t.Fatalf("pix: %#v %v", pix, err)
	}
}

func TestNuvemPagoPricingRejectsUnsupportedInputs(t *testing.T) {
	if _, err := NuvemPagoRule(PaymentCard, 4); !errors.Is(err, ErrInvalidInstallments) {
		t.Fatalf("expected installment rejection, got %v", err)
	}
	if _, err := CalculateNuvemPagoPricing(PaymentCard, 1, mustMoney(t, "-1")); !errors.Is(err, ErrNegativeGross) {
		t.Fatalf("expected negative gross rejection, got %v", err)
	}
	if _, err := NuvemPagoRule(PaymentMethod("CARDX"), 0); !errors.Is(err, ErrInvalidPricingMethod) {
		t.Fatalf("expected method rejection, got %v", err)
	}
}

func mustMoney(t *testing.T, value string) domain.Money {
	t.Helper()
	money, err := domain.ParseMoney(value)
	if err != nil {
		t.Fatal(err)
	}
	return money
}
