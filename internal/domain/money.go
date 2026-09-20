package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Money is an exact BRL-compatible decimal with four fractional places.
// Values are stored as scaled integers, never as floating point numbers.
type Money struct{ units int64 }

const moneyScale int64 = 10000

var (
	ErrInvalidMoney   = errors.New("invalid money")
	ErrMoneyOverflow  = errors.New("money overflow")
	ErrDivisionByZero = errors.New("division by zero")
)

func ZeroMoney() Money { return Money{} }

// NewMoney creates a value from a whole-unit integer and a fractional scale.
// scale must be between 0 and 4; fraction follows the sign of whole.
func NewMoney(whole, fraction int64, scale int) (Money, error) {
	if scale < 0 || scale > 4 {
		return Money{}, ErrInvalidMoney
	}
	if fraction < 0 || (scale == 0 && fraction != 0) {
		return Money{}, ErrInvalidMoney
	}
	pow := int64(1)
	for i := 0; i < scale; i++ {
		pow *= 10
	}
	if fraction >= pow {
		return Money{}, ErrInvalidMoney
	}
	if whole > math.MaxInt64/moneyScale || whole < math.MinInt64/moneyScale {
		return Money{}, ErrMoneyOverflow
	}
	units := whole * moneyScale
	part := fraction
	for i := scale; i < 4; i++ {
		part *= 10
	}
	if whole < 0 {
		if units < math.MinInt64+part {
			return Money{}, ErrMoneyOverflow
		}
		units -= part
	} else {
		if units > math.MaxInt64-part {
			return Money{}, ErrMoneyOverflow
		}
		units += part
	}
	return Money{units: units}, nil
}

// ParseMoney accepts a decimal point and up to four fractional digits. It
// rejects locale separators and exponents so parsing is deterministic.
func ParseMoney(value string) (Money, error) {
	s := strings.TrimSpace(value)
	if s == "" {
		return Money{}, ErrInvalidMoney
	}
	neg := false
	if s[0] == '+' || s[0] == '-' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" {
		return Money{}, ErrInvalidMoney
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 || parts[0] == "" {
		return Money{}, ErrInvalidMoney
	}
	for _, r := range parts[0] {
		if r < '0' || r > '9' {
			return Money{}, ErrInvalidMoney
		}
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 4 {
		return Money{}, ErrInvalidMoney
	}
	for _, r := range frac {
		if r < '0' || r > '9' {
			return Money{}, ErrInvalidMoney
		}
	}
	whole, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return Money{}, ErrInvalidMoney
	}
	for len(frac) < 4 {
		frac += "0"
	}
	var part int64
	if frac != "" {
		part, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return Money{}, ErrInvalidMoney
		}
	}
	// Build the scaled magnitude with arbitrary precision so the exact
	// MinInt64 representation (-922337203685477.5808) remains valid.
	mag := new(big.Int).SetUint64(whole)
	mag.Mul(mag, big.NewInt(moneyScale))
	mag.Add(mag, big.NewInt(part))
	if neg {
		mag.Neg(mag)
	}
	if !mag.IsInt64() {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: mag.Int64()}, nil
}

func MoneyFromString(value string) (Money, error) { return ParseMoney(value) }
func (m Money) Units() int64                      { return m.units }
func (m Money) IsZero() bool                      { return m.units == 0 }
func (m Money) IsNegative() bool                  { return m.units < 0 }
func (m Money) IsPositive() bool                  { return m.units > 0 }

// AbsExact reports overflow when the magnitude cannot be represented.
func (m Money) AbsExact() (Money, error) {
	if m.units == math.MinInt64 {
		return Money{}, ErrMoneyOverflow
	}
	if m.units < 0 {
		return Money{units: -m.units}, nil
	}
	return m, nil
}
func (m Money) Negate() (Money, error) {
	if m.units == math.MinInt64 {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: -m.units}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if other.units > 0 && m.units > math.MaxInt64-other.units || other.units < 0 && m.units < math.MinInt64-other.units {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: m.units + other.units}, nil
}
func (m Money) Sub(other Money) (Money, error) {
	n, err := other.Negate()
	if err != nil {
		return Money{}, err
	}
	return m.Add(n)
}
func (m Money) Compare(other Money) int {
	if m.units < other.units {
		return -1
	}
	if m.units > other.units {
		return 1
	}
	return 0
}

// MulRatio multiplies by an exact rational number and truncates toward zero.
// Callers should use RoundHalfUp at the presentation/rule boundary.
func (m Money) MulRatio(numerator, denominator int64) (Money, error) {
	if denominator == 0 {
		return Money{}, ErrDivisionByZero
	}
	if numerator == 0 || m.units == 0 {
		return ZeroMoney(), nil
	}
	product := new(big.Int).Mul(big.NewInt(m.units), big.NewInt(numerator))
	quotient := new(big.Int).Quo(product, big.NewInt(denominator))
	if !quotient.IsInt64() {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: quotient.Int64()}, nil
}

// MulRatioRoundHalfUp multiplies by an exact rational number and rounds the
// result to the four decimal places used by Money. It is the safe boundary
// for tariff calculations where truncation would understate a fee.
func (m Money) MulRatioRoundHalfUp(numerator, denominator int64) (Money, error) {
	if denominator == 0 {
		return Money{}, ErrDivisionByZero
	}
	if numerator == 0 || m.units == 0 {
		return ZeroMoney(), nil
	}
	product := new(big.Int).Mul(big.NewInt(m.units), big.NewInt(numerator))
	den := big.NewInt(denominator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(product, den, remainder)
	if remainder.Sign() != 0 {
		absRemainder := new(big.Int).Abs(remainder)
		absDenominator := new(big.Int).Abs(den)
		absRemainder.Lsh(absRemainder, 1)
		if absRemainder.Cmp(absDenominator) >= 0 {
			direction := int64(1)
			if product.Sign() != den.Sign() {
				direction = -1
			}
			quotient.Add(quotient, big.NewInt(direction))
		}
	}
	if !quotient.IsInt64() {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: quotient.Int64()}, nil
}

func (m Money) RoundHalfUp(places int) (Money, error) {
	if places < 0 || places > 4 {
		return Money{}, ErrInvalidMoney
	}
	factor := int64(1)
	for i := places; i < 4; i++ {
		factor *= 10
	}
	if factor == 1 {
		return m, nil
	}
	q, r := m.units/factor, m.units%factor
	if r < 0 {
		r = -r
	}
	if r*2 >= factor {
		if m.units >= 0 {
			q++
		} else {
			q--
		}
	}
	if q > math.MaxInt64/factor || q < math.MinInt64/factor {
		return Money{}, ErrMoneyOverflow
	}
	return Money{units: q * factor}, nil
}

func (m Money) String() string {
	neg := m.units < 0
	var magnitude uint64
	if neg {
		magnitude = uint64(-(m.units + 1)) + 1
	} else {
		magnitude = uint64(m.units)
	}
	whole, frac := magnitude/uint64(moneyScale), magnitude%uint64(moneyScale)
	s := fmt.Sprintf("%d.%04d", whole, frac)
	if neg {
		return "-" + s
	}
	return s
}
func (m Money) MarshalJSON() ([]byte, error) { return json.Marshal(m.String()) }
func (m *Money) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return ErrInvalidMoney
	}
	v, err := ParseMoney(s)
	if err != nil {
		return err
	}
	*m = v
	return nil
}
