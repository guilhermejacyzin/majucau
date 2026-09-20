package bling

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalReceivablesPagePreservesUnhomologatedPayload(t *testing.T) {
	total := 3
	payload, err := marshalReceivablesPage(BlingReceivablesPage{
		Records: []json.RawMessage{json.RawMessage(`{"id":42,"situacao":"pago"}`)},
		Page:    1, Limit: 1, Total: &total, HasNext: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var decoded receivablesPagePayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Page != 1 || decoded.Limit != 1 || decoded.Total == nil || *decoded.Total != 3 || !decoded.HasNext || len(decoded.Records) != 1 {
		t.Fatalf("decoded payload = %+v", decoded)
	}
	if !strings.Contains(string(decoded.Records[0]), `"id":42`) {
		t.Fatalf("raw record changed: %s", decoded.Records[0])
	}
}

func TestReceivablesPageSourceIdentityIsStable(t *testing.T) {
	const expected = "pagina:7"
	if got := receivablesPageSourceID(7); got != expected {
		t.Fatalf("source id = %q, want %q", got, expected)
	}
}
