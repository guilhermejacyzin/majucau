package bling

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestParsePayableAPIDetailUsesExactSourceAmountsAndDate(t *testing.T) {
	const payload = `{"id":21,"situacao":3,"vencimento":"2026-10-31","valor":120.5000,"saldo":45.2500,"competencia":"2026-10-01","numeroDocumento":"NF-21","categoria":{"id":9},"contato":{"id":88}}`
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
	if err := wantBalance.Scan("45.2500"); err != nil {
		t.Fatal(err)
	}
	if got.OriginalAmount.Int.Cmp(wantOriginal.Int) != 0 || got.OriginalAmount.Exp != wantOriginal.Exp || got.OpenBalance.Int.Cmp(wantBalance.Int) != 0 || got.OpenBalance.Exp != wantBalance.Exp {
		t.Fatalf("exact amounts = %s e%d / %s e%d", got.OriginalAmount.Int, got.OriginalAmount.Exp, got.OpenBalance.Int, got.OpenBalance.Exp)
	}
	if got.SupplierSource != "88" || got.CategoryID != 9 || got.Document == nil || *got.Document != "NF-21" {
		t.Fatalf("optional source mapping = %+v", got)
	}
}

func TestParsePayableAPIDetailRejectsMissingOrConflictingFinancialFields(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		id      int64
	}{
		{name: "wrong account ID", payload: `{"id":22,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10}`, id: 21},
		{name: "missing open balance", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10}`, id: 21},
		{name: "invalid due date", payload: `{"id":21,"situacao":1,"vencimento":"31/10/2026","valor":10,"saldo":10}`, id: 21},
		{name: "contact without ID", payload: `{"id":21,"situacao":1,"vencimento":"2026-10-31","valor":10,"saldo":10,"contato":{"id":0}}`, id: 21},
		{name: "unknown status", payload: `{"id":21,"situacao":8,"vencimento":"2026-10-31","valor":10,"saldo":10,"contato":{"id":9}}`, id: 21},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parsePayableAPIDetail([]byte(test.payload), test.id); !errors.Is(err, ErrBlingAPISchemaMismatch) {
				t.Fatalf("parsePayableAPIDetail() error = %v, want schema mismatch", err)
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
