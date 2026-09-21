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

func TestParseFutureCSVNormalizesProjectedNuvemPagoRows(t *testing.T) {
	report, err := ParseFutureCSV(strings.NewReader(sampleFutureCSV))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Receivables) != 2 || len(report.Errors) != 2 {
		t.Fatalf("receivables=%d errors=%d report=%#v", len(report.Receivables), len(report.Errors), report)
	}
	first := report.Receivables[0]
	if first.SourceSystem != "NUVEM_PAGO" || first.SourceEntity != FutureSourceEntity || first.SourceID != "9483" || first.Status != "PROJECTED" {
		t.Fatalf("identity/status mapping: %#v", first)
	}
	if first.GrossAmount.String() != "236.3600" || first.FeeAmount.String() != "8.1300" || first.InterestAmount.String() != "13.5300" || first.TotalCostAmount.String() != "21.6600" || first.NetAmount.String() != "214.7000" {
		t.Fatalf("amount mapping: %#v", first)
	}
	if first.InstallmentCount == nil || *first.InstallmentCount != 2 || first.PaymentDate.Format("2006-01-02") != "2026-08-20" || first.ExpectedReceiptDate.Format("2006-01-02") != "2026-09-21" {
		t.Fatalf("date/installment mapping: %#v", first)
	}
}

func TestParseFutureCSVRejectsInvalidContractBeforeRows(t *testing.T) {
	_, err := ParseFutureCSV(strings.NewReader("Cliente;Recebido\nA;10,00\n"))
	if err == nil {
		t.Fatal("missing future columns must fail before import")
	}
}

func TestParseFutureCSVRejectsReceiptDateBeforePayment(t *testing.T) {
	input := strings.Replace(sampleFutureCSV, "20/08/2026;21/09/2026;Entrada;Venda;9483", "20/08/2026;19/08/2026;Entrada;Venda;9999", 1)
	report, err := ParseFutureCSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Receivables) != 2 || len(report.Errors) != 2 || report.Errors[0].Code != "RECEIPT_DATE_BEFORE_PAYMENT" {
		t.Fatalf("unexpected validation result: %#v", report)
	}
}
