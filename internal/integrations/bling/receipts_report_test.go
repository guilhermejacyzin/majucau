package bling

import (
	"strings"
	"testing"
)

const sampleReceiptsCSV = "\ufeffCliente;Histórico;Forma de pagamento;Nº documento;Vencimento;Liquidação;Situação;Valor taxa;Recebido\n" +
	"Cliente A;Ref. ao pedido de venda nº 16223;NUVEMPAGO 3X;016223/01;01/08/2026;03/08/2026;pago;53,43;461,45\n" +
	"Cliente B;Ref. a NF nº 007882;BOLETO;007882/01;16/07/2026;21/08/2026;pago;0,00;7.251,06\n" +
	"Cliente C;Ref. ao pedido nº 17785;Nuvemshop PIX;017785/01;01/08/2026;01/08/2026;pago;1,18;118,11\n" +
	"Cliente D;Ref. ao pedido nº 17786;NUVEMPAGO 1X;017786/01;02/08/2026;;aberto;5,00;100,00\n" +
	"Cliente E;Ref. duplicado;NUVEMPAGO 2X;016223/01;02/08/2026;03/08/2026;pago;10,00;100,00\n"

func TestParseReceiptsCSVUsesBlingAsSourceForNuvemPaymentMethods(t *testing.T) {
	report, err := ParseReceiptsCSV(strings.NewReader(sampleReceiptsCSV))
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Receipts) != 3 || len(report.Errors) != 2 {
		t.Fatalf("receipts=%d errors=%d report=%#v", len(report.Receipts), len(report.Errors), report)
	}
	first := report.Receipts[0]
	if string(first.SourceSystem) != "BLING" || first.Status != "CONFIRMED" || first.Amount.String() != "461.4500" || first.Fee.String() != "53.4300" {
		t.Fatalf("source/amount mapping: %#v", first)
	}
	if first.ReceiptDate.Format("2006-01-02") != "2026-08-03" || first.PaymentMethod != "NUVEMPAGO 3X" {
		t.Fatalf("date/payment mapping: %#v", first)
	}
}

func TestParseReceiptsCSVRejectsMissingColumn(t *testing.T) {
	_, err := ParseReceiptsCSV(strings.NewReader("Cliente;Recebido\nA;10,00\n"))
	if err == nil {
		t.Fatal("missing report columns must fail before import")
	}
}

func TestParseBrazilianMoney(t *testing.T) {
	for input, want := range map[string]string{"0,00": "0.0000", "7.251,06": "7251.0600", "(10,25)": "-10.2500"} {
		got, err := parseBrazilianMoney(input)
		if err != nil || got.String() != want {
			t.Fatalf("%q -> %s (%v), want %s", input, got.String(), err, want)
		}
	}
}
