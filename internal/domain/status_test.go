package domain

import (
	"math"
	"testing"
)

func TestExternalIdentityNormalizedAndStable(t *testing.T) {
	a, err := NewExternalIdentity(" c-1 ", "bling", "contas_receber", " 42 ")
	if err != nil || a.SourceSystem != "BLING" || a.SourceEntity != "CONTAS_RECEBER" || a.SourceID != "42" {
		t.Fatalf("normalize: %#v %v", a, err)
	}
	b, _ := NewExternalIdentity("c-1", "BLING", "CONTAS_RECEBER", "42")
	if a.Key() != b.Key() || a.Hash() != b.Hash() {
		t.Fatal("identity must be stable")
	}
	if _, err := NewExternalIdentity("", "BLING", "ORDER", "1"); err == nil {
		t.Fatal("empty identity component must fail")
	}
}
func TestAbsExactMinInt64(t *testing.T) {
	m := Money{units: math.MinInt64}
	if _, err := m.AbsExact(); err != ErrMoneyOverflow {
		t.Fatalf("expected exact abs overflow, got %v", err)
	}
}
