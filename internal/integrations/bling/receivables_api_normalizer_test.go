package bling

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestParseReceivableAPIListAndDetailWithoutClassifyingCustomerType(t *testing.T) {
	listPayload := json.RawMessage(`{"id":321,"situacao":3,"vencimento":"2026-11-12","dataEmissao":"2026-10-01","valor":100.25,"contato":{"id":9,"nome":"Cliente Teste","tipo":"F"},"formaPagamento":{"id":11},"contaContabil":{"id":12},"portador":{"id":13}}`)
	list, err := parseReceivableAPIListRecord(listPayload)
	if err != nil {
		t.Fatalf("parse receivable list = %v", err)
	}
	if list.ID != 321 || list.Status != 3 || list.Contact.ID != 9 {
		t.Fatalf("parsed list identity/status/contact = %+v", list)
	}
	listNormalized, err := normalizedReceivableFromList(list, normalizedReceivable{})
	if err != nil || listNormalized.SourcePaymentMethodID == nil || *listNormalized.SourcePaymentMethodID != 11 || listNormalized.SourceAccountingAccountID == nil || *listNormalized.SourceAccountingAccountID != 12 || listNormalized.SourceFinancialAccountID == nil || *listNormalized.SourceFinancialAccountID != 13 {
		t.Fatalf("parsed list source references = %+v, %v", listNormalized, err)
	}
	amount, err := parseRequiredAPIMoney(list.Amount)
	wantAmount := pgtype.Numeric{}
	if scanErr := wantAmount.Scan("100.25"); scanErr != nil {
		t.Fatal(scanErr)
	}
	if err != nil || amount.Int.Cmp(wantAmount.Int) != 0 || amount.Exp != wantAmount.Exp {
		t.Fatalf("parsed receivable amount = %+v, %v", amount, err)
	}

	detailPayload := json.RawMessage(`{"id":321,"situacao":3,"vencimento":"2026-11-12","dataEmissao":"2026-10-01","competencia":"2026-10-01","numeroDocumento":"R-321","valor":100.25,"saldo":40.10,"contato":{"id":9},"formaPagamento":{"id":11},"contaContabil":{"id":12},"portador":{"id":13}}`)
	detail, err := parseReceivableAPIDetail(detailPayload, 321)
	if err != nil {
		t.Fatalf("parse receivable detail = %v", err)
	}
	if detail.Status != list.Status || detail.Contact.ID != list.Contact.ID {
		t.Fatalf("detail changed list status/contact: %+v", detail)
	}
	detailNormalized, err := normalizedReceivableFromDetail(detail, normalizedReceivable{})
	if err != nil || detailNormalized.SourcePaymentMethodID == nil || *detailNormalized.SourcePaymentMethodID != 11 || detailNormalized.SourceAccountingAccountID == nil || *detailNormalized.SourceAccountingAccountID != 12 || detailNormalized.SourceFinancialAccountID == nil || *detailNormalized.SourceFinancialAccountID != 13 {
		t.Fatalf("parsed detail source references = %+v, %v", detailNormalized, err)
	}
	open, err := parseRequiredAPIMoney(detail.OpenAmount)
	wantOpen := pgtype.Numeric{}
	if scanErr := wantOpen.Scan("40.10"); scanErr != nil {
		t.Fatal(scanErr)
	}
	if err != nil || open.Int.Cmp(wantOpen.Int) != 0 || open.Exp != wantOpen.Exp {
		t.Fatalf("parsed outstanding balance = %+v, %v", open, err)
	}
}

func TestReceivableStatusesAndDetailsAreConservative(t *testing.T) {
	statuses := map[int64]string{
		1: "OPEN", 2: "PAID", 3: "PARTIALLY_PAID", 4: "RETURNED",
		5: "CANCELLED", 6: "PARTIALLY_RETURNED", 7: "CONFIRMED",
	}
	for code, want := range statuses {
		if got := receivableStatus(code); got != want {
			t.Errorf("status %d = %q, want %q", code, got, want)
		}
	}
	if receivableNeedsDetail(2) || receivableNeedsDetail(5) {
		t.Fatal("paid/cancelled accounts should not require an outstanding-balance detail read")
	}
	for _, code := range []int64{1, 3, 4, 6, 7} {
		if !receivableNeedsDetail(code) {
			t.Errorf("status %d may have a balance and must require a detail read", code)
		}
	}
}

func TestReceivableContactMatchingNameIsRequiredButPersonTypeIsIgnored(t *testing.T) {
	payload := json.RawMessage(`{"id":9,"nome":"Cliente Teste","tipo":"F"}`)
	name, err := parseReceivableContactDetail(payload, 9)
	if err != nil || name != "Cliente Teste" {
		t.Fatalf("contact name = %q, err = %v", name, err)
	}
	if _, err := parseReceivableContactDetail(payload, 10); err == nil {
		t.Fatal("mismatched contact ID must fail closed")
	}
}
