package bling

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestParsePayableAPIDetailUsesExactSourceAmountsAndDate(t *testing.T) {
	const payload = `{"id":21,"situacao":3,"vencimento":"2026-10-31","valor":120.5000,"saldo":45.2500,"competencia":"2026-10-01","numeroDocumento":"NF-21","categoria":{"id":9},"contato":{"id":88},"formaPagamento":{"id":11},"portador":{"id":13}}`
	got, err := parsePayableAPIDetail([]byte(payload), 21)
	if err != nil {
		t.Fatalf("parsePayableAPIDetail() error = %v", err)
	}
	if got.SourceID != "21" || got.Status != "3" || !got.DueDate.Valid || got.DueDate.Time.Format("2006-01-02") != "2026-10-31" {
		t.Fatalf("identity/date mapping = %+v", got)
	}
	wantOriginal := pgtype.Numeric{}
	wantBalance := pgtype.Numeric{}
	if err := wantOriginal.Scan("120.5000"); err != nil {
		t.Fatal(err)
	}
	if err := wantBalance.Scan("-45.2500"); err != nil {
		t.Fatal(err)
	}
	if got.OriginalAmount.Int.Cmp(wantOriginal.Int) != 0 || got.OriginalAmount.Exp != wantOriginal.Exp || got.OpenBalance.Int.Cmp(wantBalance.Int) != 0 || got.OpenBalance.Exp != wantBalance.Exp {
		t.Fatalf("exact amounts = %s e%d / %s e%d", got.OriginalAmount.Int, got.OriginalAmount.Exp, got.OpenBalance.Int, got.OpenBalance.Exp)
	}
	signedOutflow, err := parsePayableAPIDetail([]byte(`{"id":22,"situacao":3,"vencimento":"2026-10-31","valor":120.5,"saldo":-45.2500,"categoria":{"id":9},"contato":{"id":88}}`), 22)
	if err != nil || signedOutflow.OpenBalance.Int.Cmp(wantBalance.Int) != 0 || signedOutflow.OpenBalance.Exp != wantBalance.Exp {
		t.Fatalf("already-negative source balance = %+v, %v; want exact negative outflow", signedOutflow.OpenBalance, err)
	}
	if got.SupplierSource != "88" || got.CategoryID != 9 || got.Document == nil || *got.Document != "NF-21" {
		t.Fatalf("optional source mapping = %+v", got)
	}
	if got.SourcePaymentMethodID == nil || *got.SourcePaymentMethodID != 11 || got.SourceFinancialAccountID == nil || *got.SourceFinancialAccountID != 13 {
		t.Fatalf("cash-flow source references = %+v", got)
	}
	uncategorized, err := parsePayableAPIDetail([]byte(`{"id":21,"situacao":3,"vencimento":"2026-10-31","valor":120.5,"saldo":45.25,"categoria":{"id":0},"contato":{"id":88}}`), 21)
	if err != nil || !uncategorized.CategoryUnassigned || uncategorized.CategoryID != 0 {
		t.Fatalf("zero category ID should remain unclassified: result=%+v error=%v", uncategorized, err)
	}
}

func TestParsePayableAPIDetailRejectsMissingOrConflictingFinancialFields(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		id       int64
		wantCode string
	}{
		{name: "wrong account ID", payload: `{"id":22,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10}`, id: 21, wantCode: "BLING_PAYABLE_DETAIL_ID_INVALID"},
		{name: "missing open balance", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10}`, id: 21, wantCode: "BLING_PAYABLE_OPEN_BALANCE_MISSING"},
		{name: "invalid due date", payload: `{"id":21,"situacao":1,"vencimento":"31/10/2026","valor":10,"saldo":10}`, id: 21, wantCode: "BLING_PAYABLE_DUE_DATE_INVALID"},
		{name: "contact without ID", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10,"contato":{"id":0}}`, id: 21, wantCode: "BLING_PAYABLE_SUPPLIER_ID_INVALID"},
		{name: "unknown status", payload: `{"id":21,"situacao":8,"vencimento":"2026-10-31","valor":10,"saldo":10,"contato":{"id":9}}`, id: 21, wantCode: "BLING_PAYABLE_STATUS_INVALID"},
		{name: "missing category ID", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10,"categoria":{},"contato":{"id":9}}`, id: 21, wantCode: "BLING_PAYABLE_CATEGORY_ID_INVALID"},
		{name: "negative category ID", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10,"categoria":{"id":-1},"contato":{"id":9}}`, id: 21, wantCode: "BLING_PAYABLE_CATEGORY_ID_INVALID"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parsePayableAPIDetail([]byte(test.payload), test.id); !errors.Is(err, ErrBlingAPISchemaMismatch) {
				t.Fatalf("parsePayableAPIDetail() error = %v, want schema mismatch", err)
			} else if gotCode := blingAPINormalizationFailureCode(err); gotCode != test.wantCode {
				t.Fatalf("safe schema code = %q, want %q", gotCode, test.wantCode)
			}
		})
	}
}

func TestParsePayableAPIDetailClassifiesOpenBalanceWithoutExposingValue(t *testing.T) {
	cases := []struct {
		name, raw, wantCode string
	}{
		{name: "object instead of decimal", raw: `{"amount":10}`, wantCode: "BLING_PAYABLE_OPEN_BALANCE_TYPE_INVALID"},
		{name: "negative decimal remains an outflow", raw: `-1.25`, wantCode: ""},
		{name: "unsupported precision", raw: `1.00001`, wantCode: "BLING_PAYABLE_OPEN_BALANCE_PRECISION_EXCEEDED"},
		{name: "scientific notation", raw: `1e2`, wantCode: "BLING_PAYABLE_OPEN_BALANCE_SCIENTIFIC_NOTATION"},
		{name: "locale decimal separator", raw: `"1,25"`, wantCode: "BLING_PAYABLE_OPEN_BALANCE_DECIMAL_SEPARATOR_INVALID"},
		{name: "valid exact amount", raw: `12.3400`, wantCode: ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload := `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":` + test.raw + `,"contato":{"id":9}}`
			got, err := parsePayableAPIDetail([]byte(payload), 21)
			if test.wantCode == "" {
				if err != nil {
					t.Fatalf("parsePayableAPIDetail() error = %v", err)
				}
				if got.OpenBalance.Int.Sign() > 0 {
					t.Fatalf("open balance = %s, want a non-positive outflow", got.OpenBalance.Int)
				}
				return
			}
			if err == nil {
				t.Fatal("parsePayableAPIDetail() unexpectedly succeeded")
			}
			if gotCode := blingAPINormalizationFailureCode(err); gotCode != test.wantCode {
				t.Fatalf("safe schema code = %q, want %q", gotCode, test.wantCode)
			}
			if strings.Contains(err.Error(), test.raw) {
				t.Fatal("provider balance value leaked into the error")
			}
		})
	}
}

func TestParseRequiredAPIMoneyKeepsPrecisionAndRejectsRounding(t *testing.T) {
	for _, raw := range []string{`0`, `100.00`, `999999999999999.9999`, `"12.3400"`} {
		var value json.RawMessage = []byte(raw)
		if _, err := parseRequiredAPIMoney(value); err != nil {
			t.Errorf("parseRequiredAPIMoney(%s) error = %v", raw, err)
		}
	}
	for _, raw := range []string{`-1`, `1.00001`, `1000000000000000`, `1e2`, `null`, `"NaN"`} {
		var value json.RawMessage = []byte(raw)
		if _, err := parseRequiredAPIMoney(value); err == nil {
			t.Errorf("parseRequiredAPIMoney(%s) unexpectedly accepted", raw)
		}
	}
}

func TestParsePayableReferenceDetailsRequireMatchingIDs(t *testing.T) {
	if _, err := parsePayableContactDetail([]byte(`{"id":89,"nome":"Fornecedor"}`), 88); !errors.Is(err, ErrBlingAPISchemaMismatch) {
		t.Fatalf("contact mismatch error = %v", err)
	}
	if _, err := parsePayableCategoryDetail([]byte(`{"id":10,"descricao":"Insumos","tipo":1}`), 9); !errors.Is(err, ErrBlingAPISchemaMismatch) {
		t.Fatalf("category mismatch error = %v", err)
	}
	if name, err := parsePayableCategoryDetail([]byte(`{"id":9,"descricao":"Insumos","tipo":1}`), 9); err != nil || name != "Insumos" {
		t.Fatalf("category detail = %q, %v", name, err)
	}
}
