package financial

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type BusinessType string

const (
	B2B                        BusinessType = "B2B"
	B2CExcludedFromBlingFuture BusinessType = "B2C_EXCLUDED_FROM_BLING_FUTURE"
	Unclassified               BusinessType = "UNCLASSIFIED"
)

type ClassificationInput struct {
	ManualOverride       BusinessType
	OverrideReason       string
	ReconciledToNuvemB2C bool
	Document             string
	ApprovedB2BListMatch bool
}
type ClassificationResult struct {
	BusinessType BusinessType
	RuleVersion  string
	Reason       string
}

// ClassifyB2B applies D-001-A in order and deliberately returns UNCLASSIFIED
// for ambiguous data. Manual overrides require a reason and are not silently
// accepted as B2B without an audit explanation.
func ClassifyB2B(in ClassificationInput, ruleVersion string) ClassificationResult {
	if ruleVersion == "" {
		ruleVersion = "D-001-A"
	}
	if validBusinessType(in.ManualOverride) && in.OverrideReason != "" {
		return ClassificationResult{in.ManualOverride, ruleVersion, "audited_manual_override"}
	}
	if in.ReconciledToNuvemB2C {
		return ClassificationResult{B2CExcludedFromBlingFuture, ruleVersion, "reconciled_nuvem_b2c"}
	}
	if ValidCNPJ(in.Document) {
		return ClassificationResult{B2B, ruleVersion, "valid_cnpj"}
	}
	if in.ApprovedB2BListMatch {
		return ClassificationResult{B2B, ruleVersion, "approved_b2b_list"}
	}
	return ClassificationResult{Unclassified, ruleVersion, "insufficient_deterministic_evidence"}
}

func validBusinessType(value BusinessType) bool {
	return value == B2B || value == B2CExcludedFromBlingFuture || value == Unclassified
}

func ValidCNPJ(value string) bool {
	digits := make([]int, 0, 14)
	for _, r := range strings.TrimSpace(value) {
		if r >= '0' && r <= '9' {
			digits = append(digits, int(r-'0'))
		} else if r != '.' && r != '/' && r != '-' {
			return false
		}
	}
	if len(digits) != 14 {
		return false
	}
	allSame := true
	for _, d := range digits[1:] {
		if d != digits[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}
	for n := 12; n <= 13; n++ {
		sum := 0
		weight := n - 7
		for i := 0; i < n; i++ {
			sum += digits[i] * weight
			weight--
			if weight == 1 {
				weight = 9
			}
		}
		check := (sum * 10) % 11
		if check == 10 {
			check = 0
		}
		if check != digits[n] {
			return false
		}
	}
	return true
}

func PayloadHash(payload []byte) string { h := sha256.Sum256(payload); return hex.EncodeToString(h[:]) }
