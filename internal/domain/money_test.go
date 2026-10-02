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
	cases := []struct {
		input    string
		places   int
		expected string
	}{
		{input: "1.2345", places: 2, expected: "1.2300"},
		{input: "1.2350", places: 2, expected: "1.2400"},
		{input: "-1.2350", places: 2, expected: "-1.2400"},
		{input: "-1.0050", places: 2, expected: "-1.0100"},
		{input: "1.5000", places: 0, expected: "2.0000"},
		{input: "-1.5000", places: 0, expected: "-2.0000"},
		{input: "1.2345", places: 3, expected: "1.2350"},
		{input: "1.2345", places: 4, expected: "1.2345"},
	}
	for _, tc := range cases {
		m, err := ParseMoney(tc.input)
		if err != nil {
			t.Fatalf("parse %s: %v", tc.input, err)
		}
		got, err := m.RoundHalfUp(tc.places)
		if err != nil || got.String() != tc.expected {
			t.Errorf("%s rounded to %d places => %v (%v), want %s", tc.input, tc.places, got, err, tc.expected)
		}
	}
}

func TestMoneyKeepsPrecisionUntilApprovedRoundingBoundary(t *testing.T) {
	first, err := ParseMoney("0.0049")
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseMoney("0.0049")
	if err != nil {
		t.Fatal(err)
	}
	total, err := first.Add(second)
	if err != nil || total.String() != "0.0098" {
		t.Fatalf("intermediate addition rounded or lost precision: %v %v", total, err)
	}
	rounded, err := total.RoundHalfUp(2)
	if err != nil || rounded.String() != "0.0100" {
		t.Fatalf("final approved boundary rounding: %v %v", rounded, err)
	}
}

func TestMoneyMulRatioRoundHalfUp(t *testing.T) {
	amount, _ := ParseMoney("100.01")
	fee, err := amount.MulRatioRoundHalfUp(259, 10000)
	if err != nil || fee.String() != "2.5903" {
		t.Fatalf("fee rounding: %v %v", fee, err)
	}
	half, _ := ParseMoney("1.0050")
	rounded, err := half.MulRatioRoundHalfUp(1, 1)
	if err != nil || rounded.String() != "1.0050" {
		t.Fatalf("identity must preserve scale: %v %v", rounded, err)
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
