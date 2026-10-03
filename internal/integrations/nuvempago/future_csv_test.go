package nuvempago

import (
	"strings"
	"testing"
)

const sampleFutureCSV = "\ufeffData do pagamento;Data de recebimento;Tipo de movimentação;Tipo de transação;Nº transação;Nome;Forma de pagamento;Bandeira;Nº Parcelas;Valor Bruto (R$);Taxas (R$);Juros (R$);Custos Totais (R$);Valor Liquido (R$)\n" +
	"20/08/2026;21/09/2026;Entrada;Venda;9483;Cliente sintético;Cartão de crédito;visa;2;236,36;-8,13;-13,53;-21,66;214,70\n" +
	"20/08/2026;20/09/2026;Entrada;Venda;9484;Outro cliente;Pix;; ;100,00;-0,99;0,00;-0,99;99,01\n" +
	"20/08/2026;21/09/2026;Saída;Venda;9485;Não importar;Cartão de crédito;visa;1;10,00;0,00;0,00;0,00;10,00\n" +
	"20/08/2026;21/09/2026;Entrada;Venda;9483;Duplicado;Cartão de crédito;visa;2;236,36;-8,13;-13,53;-21,66;214,70\n"

func TestStreamFutureCSVNormalizesProjectedNuvemPagoRows(t *testing.T) {
	var first FutureReceivableCandidate
	accepted, rejected := 0, 0
	err := StreamFutureCSV(strings.NewReader(sampleFutureCSV), func(_ int, candidate FutureReceivableCandidate) error {
		accepted++
		if accepted == 1 {
			first = candidate
		}
		return nil
	}, func(RowError) error {
		rejected++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if accepted != 3 || rejected != 1 {
		t.Fatalf("stream accepted=%d rejected=%d", accepted, rejected)
	}
	if first.SourceSystem != "NUVEM_PAGO" || first.SourceEntity != FutureSourceEntity || first.SourceID != "9483" || first.Status != "PROJECTED" {
		t.Fatalf("identity/status mapping: %#v", first)
	}
	if first.GrossAmount.String() != "236.3600" || first.FeeAmount.String() != "8.1300" || first.FeeAmountSigned.String() != "-8.1300" || first.InterestAmount.String() != "13.5300" || first.InterestAmountSigned.String() != "-13.5300" || first.TotalCostAmount.String() != "21.6600" || first.TotalCostAmountSigned.String() != "-21.6600" || first.NetAmount.String() != "214.7000" {
		t.Fatalf("amount mapping: %#v", first)
	}
	if first.InstallmentCount == nil || *first.InstallmentCount != 2 || first.PaymentDate.Format("2006-01-02") != "2026-08-20" || first.ExpectedReceiptDate.Format("2006-01-02") != "2026-09-21" {
		t.Fatalf("date/installment mapping: %#v", first)
	}
}

func TestStreamFutureCSVRejectsInvalidContractBeforeRows(t *testing.T) {
	err := StreamFutureCSV(strings.NewReader("Cliente;Recebido\nA;10,00\n"), nil, nil)
	if err == nil {
		t.Fatal("missing future columns must fail before import")
	}
}

func TestStreamFutureCSVRejectsReceiptDateBeforePayment(t *testing.T) {
	input := strings.Replace(sampleFutureCSV, "20/08/2026;21/09/2026;Entrada;Venda;9483", "20/08/2026;19/08/2026;Entrada;Venda;9999", 1)
	firstRejected := ""
	accepted, rejected := 0, 0
	err := StreamFutureCSV(strings.NewReader(input), func(_ int, _ FutureReceivableCandidate) error {
		accepted++
		return nil
	}, func(issue RowError) error {
		rejected++
		if firstRejected == "" {
			firstRejected = issue.Code
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if accepted != 2 || rejected != 2 || firstRejected != "RECEIPT_DATE_BEFORE_PAYMENT" {
		t.Fatalf("unexpected stream result: accepted=%d rejected=%d first rejection=%q", accepted, rejected, firstRejected)
	}
}
