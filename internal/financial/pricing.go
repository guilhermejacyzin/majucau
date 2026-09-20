package financial

import (
	"errors"
	"time"

	"majucau.local/financial-intelligence/internal/domain"
)

const NuvemPagoPricingVersion = "nuvem-pago-2026-09-20-v1"

var (
	ErrInvalidPricingMethod = errors.New("invalid pricing method")
	ErrInvalidInstallments  = errors.New("invalid installments")
	ErrNegativeGross        = errors.New("gross amount cannot be negative")
)

type PaymentMethod string

const (
	PaymentCard   PaymentMethod = "CARD"
	PaymentBoleto PaymentMethod = "BOLETO"
	PaymentPIX    PaymentMethod = "PIX"
)

type PricingRule struct {
	Version          string        `json:"pricing_version"`
	Method           PaymentMethod `json:"method"`
	Installments     int           `json:"installments,omitempty"`
	ReceiptDays      *int          `json:"receipt_days,omitempty"`
	ReceiptImmediate bool          `json:"receipt_immediate,omitempty"`
	PercentNumerator int64         `json:"percent_numerator"`
	FixedFee         domain.Money  `json:"fixed_fee"`
	TPVFree          bool          `json:"tpv_free"`
	Source           string        `json:"source"`
	ValidFrom        time.Time     `json:"valid_from"`
}

type PricingResult struct {
	Rule         PricingRule  `json:"rule"`
	Gross        domain.Money `json:"gross"`
	Fee          domain.Money `json:"fee"`
	EstimatedNet domain.Money `json:"estimated_net"`
}

func NuvemPagoRule(method PaymentMethod, installments int) (PricingRule, error) {
	fixed, err := domain.ParseMoney("0.35")
	if err != nil {
		return PricingRule{}, err
	}
	validFrom := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	days := 30
	switch method {
	case PaymentCard:
		var percentage int64
		switch installments {
		case 1:
			percentage = 259
		case 2:
			percentage = 449
		case 3:
			percentage = 544
		default:
			return PricingRule{}, ErrInvalidInstallments
		}
		return PricingRule{Version: NuvemPagoPricingVersion, Method: method, Installments: installments, ReceiptDays: &days, PercentNumerator: percentage, FixedFee: fixed, TPVFree: true, Source: "user-approved-panel-captures-2026-09-20", ValidFrom: validFrom}, nil
	case PaymentBoleto:
		fee, parseErr := domain.ParseMoney("2.39")
		if parseErr != nil {
			return PricingRule{}, parseErr
		}
		days = 2
		return PricingRule{Version: NuvemPagoPricingVersion, Method: method, ReceiptDays: &days, FixedFee: fee, TPVFree: true, Source: "user-approved-panel-captures-2026-09-20", ValidFrom: validFrom}, nil
	case PaymentPIX:
		return PricingRule{Version: NuvemPagoPricingVersion, Method: method, ReceiptImmediate: true, PercentNumerator: 99, TPVFree: true, Source: "user-approved-panel-captures-2026-09-20", ValidFrom: validFrom}, nil
	default:
		return PricingRule{}, ErrInvalidPricingMethod
	}
}

func CalculateNuvemPagoPricing(method PaymentMethod, installments int, gross domain.Money) (PricingResult, error) {
	if gross.IsNegative() {
		return PricingResult{}, ErrNegativeGross
	}
	rule, err := NuvemPagoRule(method, installments)
	if err != nil {
		return PricingResult{}, err
	}
	fee := rule.FixedFee
	if rule.PercentNumerator > 0 {
		percentageFee, ratioErr := gross.MulRatioRoundHalfUp(rule.PercentNumerator, 10000)
		if ratioErr != nil {
			return PricingResult{}, ratioErr
		}
		fee, err = fee.Add(percentageFee)
		if err != nil {
			return PricingResult{}, err
		}
	}
	net, err := gross.Sub(fee)
	if err != nil {
		return PricingResult{}, err
	}
	return PricingResult{Rule: rule, Gross: gross, Fee: fee, EstimatedNet: net}, nil
}
