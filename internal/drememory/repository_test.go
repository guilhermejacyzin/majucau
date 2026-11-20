package drememory

import (
	"testing"

	"majucau.local/financial-intelligence/internal/financial"
)

func TestFormulaExpressionUsesRecordedOperationsAndOperands(t *testing.T) {
	tests := []struct {
		name string
		step financial.DRECalculationStep
		want string
	}{
		{name: "sum", step: financial.DRECalculationStep{Operation: "SUM", OperandCodes: []financial.DRELineCode{financial.DREGrossRevenueCode, financial.DRERevenueDeductionsCode}, OperandValues: []financial.DREAmount{{Available: true}, {Available: true}}}, want: "GROSS_REVENUE + REVENUE_DEDUCTIONS"},
		{name: "identity", step: financial.DRECalculationStep{Operation: "IDENTITY", OperandCodes: []financial.DRELineCode{financial.DREOperatingResultCode}, OperandValues: []financial.DREAmount{{Available: true}}}, want: "OPERATING_RESULT"},
		{name: "signed expense subtraction", step: financial.DRECalculationStep{Operation: "SUBTRACT_SIGNED_EXPENSE", OperandCodes: []financial.DRELineCode{financial.DREOperatingResultCode, financial.DREDepreciationAmortizationCode}, OperandValues: []financial.DREAmount{{Available: true}, {Available: true}}}, want: "OPERATING_RESULT - (DEPRECIATION_AND_AMORTIZATION)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := formulaExpression(test.step)
			if err != nil || got != test.want {
				t.Fatalf("formulaExpression() = %q, want %q", got, test.want)
			}
		})
	}
	if _, err := formulaExpression(financial.DRECalculationStep{Operation: "UNMAPPED", OperandCodes: []financial.DRELineCode{financial.DREGrossRevenueCode}, OperandValues: []financial.DREAmount{{Available: true}}}); err == nil {
		t.Fatal("unknown operation was accepted as a formula")
	}
}

func TestDRELabelsAndInputHashValidation(t *testing.T) {
	if got := lineLabel(financial.DREIncomeTaxesCode); got != "Tributos sobre o lucro" {
		t.Fatalf("lineLabel(taxes) = %q", got)
	}
	if !isHex("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef") {
		t.Fatal("valid SHA-256 hash was rejected")
	}
	if isHex("0123456789ABCDEF") || isHex("not-a-hash") {
		t.Fatal("invalid hash was accepted")
	}
}
