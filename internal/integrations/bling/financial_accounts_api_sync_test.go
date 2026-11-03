package bling

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseFinancialAccountAPIIdentity(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		wantID      int64
		wantName    string
		wantFailure bool
	}{
		{name: "valid account", payload: `{"id":123,"descricao":" Itaú "}`, wantID: 123, wantName: "Itaú"},
		{name: "missing id", payload: `{"descricao":"Itaú"}`, wantFailure: true},
		{name: "nonpositive id", payload: `{"id":0,"descricao":"Itaú"}`, wantFailure: true},
		{name: "malformed json", payload: `{"id":`, wantFailure: true},
		{name: "empty description", payload: `{"id":123,"descricao":"  "}`, wantFailure: true},
		{name: "description too long", payload: `{"id":123,"descricao":"` + strings.Repeat("a", 513) + `"}`, wantFailure: true},
		{name: "invalid utf8 json", payload: string([]byte{0xff}), wantFailure: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, description, err := parseFinancialAccountAPIIdentity(json.RawMessage(test.payload))
			if test.wantFailure {
				if err == nil {
					t.Fatalf("parseFinancialAccountAPIIdentity() unexpectedly succeeded: id=%d description=%q", id, description)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseFinancialAccountAPIIdentity() error = %v", err)
			}
			if id != test.wantID || description != test.wantName {
				t.Fatalf("identity = (%d, %q), want (%d, %q)", id, description, test.wantID, test.wantName)
			}
		})
	}
}
