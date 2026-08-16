package domain

import (
	"errors"
	"math"
	"testing"
)

func TestMoneyExactArithmeticAndParsing(t *testing.T) {
	a, err := ParseMoney("0.10")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseMoney("0.20")
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.Add(b)
	if err != nil || c.String() != "0.3000" {
		t.Fatalf("got %v, %v", c, err)
	}
	if _, err := ParseMoney("1.00001"); !errors.Is(err, ErrInvalidMoney) {
		t.Fatalf("expected precision rejection, got %v", err)
	}
	if _, err := ParseMoney("1,20"); err == nil {
		t.Fatal("locale separators must be rejected")
	}
	neg, err := ParseMoney("-12.3456")
	if err != nil || neg.String() != "-12.3456" {
		t.Fatalf("negative parse: %v %v", neg, err)
	}
}
func TestMoneyRoundHalfUp(t *testing.T) {
	for input, expected := range map[string]string{"1.2345": "1.2300", "1.2350": "1.2400", "-1.2350": "-1.2400"} {
		m, _ := ParseMoney(input)
		got, err := m.RoundHalfUp(2)
		if err != nil || got.String() != expected {
			t.Errorf("%s => %v (%v), want %s", input, got, err, expected)
		}
	}
}
func TestMoneyOverflowAndDivision(t *testing.T) {
	m, _ := ParseMoney("922337203685477.5807")
	if _, err := m.Add(Money{units: 1}); !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("overflow not detected: %v", err)
	}
	if _, err := m.MulRatio(1, 0); !errors.Is(err, ErrDivisionByZero) {
		t.Fatalf("division error not detected: %v", err)
	}
	min, err := ParseMoney("-922337203685477.5808")
	if err != nil || min.Units() != math.MinInt64 || min.String() != "-922337203685477.5808" {
		t.Fatalf("MinInt64 money: %v %v", min, err)
	}
	if _, err := min.MulRatio(-1, 1); !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("MinInt64 multiplication overflow not detected: %v", err)
	}
	max, err := ParseMoney("922337203685477.5807")
	if err != nil || max.Units() != math.MaxInt64 || max.String() != "922337203685477.5807" {
		t.Fatalf("MaxInt64 money: %v %v", max, err)
	}
}
