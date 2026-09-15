package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type DataStatus string

const (
	StatusConfirmed             DataStatus = "CONFIRMED"
	StatusPartiallyConfirmed    DataStatus = "PARTIALLY_CONFIRMED"
	StatusProvisional           DataStatus = "PROVISIONAL"
	StatusProjected             DataStatus = "PROJECTED"
	StatusPendingReconciliation DataStatus = "PENDING_RECONCILIATION"
	StatusDivergent             DataStatus = "DIVERGENT"
)

func (s DataStatus) Valid() bool {
	switch s {
	case StatusConfirmed, StatusPartiallyConfirmed, StatusProvisional, StatusProjected, StatusPendingReconciliation, StatusDivergent:
		return true
	}
	return false
}

type Origin string

const (
	OriginBling      Origin = "BLING"
	OriginNuvemshop  Origin = "NUVEMSHOP"
	OriginNuvemPago  Origin = "NUVEM_PAGO"
	OriginPayroll    Origin = "PAYROLL"
	OriginManual     Origin = "MANUAL"
	OriginCalculated Origin = "CALCULATED"
)

func (o Origin) Valid() bool {
	switch o {
	case OriginBling, OriginNuvemshop, OriginNuvemPago, OriginPayroll, OriginManual, OriginCalculated:
		return true
	}
	return false
}

type IntegrationStatus string

const (
	IntegrationNotConfigured      IntegrationStatus = "NOT_CONFIGURED"
	IntegrationAuthorizing        IntegrationStatus = "AUTHORIZING"
	IntegrationConnected          IntegrationStatus = "CONNECTED"
	IntegrationTokenExpiring      IntegrationStatus = "TOKEN_EXPIRING"
	IntegrationAuthError          IntegrationStatus = "AUTH_ERROR"
	IntegrationSchemaMismatch     IntegrationStatus = "SCHEMA_MISMATCH"
	IntegrationSyncing            IntegrationStatus = "SYNCING"
	IntegrationStale              IntegrationStatus = "STALE"
	IntegrationPartiallyAvailable IntegrationStatus = "PARTIALLY_AVAILABLE"
	IntegrationUnavailable        IntegrationStatus = "UNAVAILABLE"
)

func (s IntegrationStatus) Valid() bool {
	switch s {
	case IntegrationNotConfigured, IntegrationAuthorizing, IntegrationConnected, IntegrationTokenExpiring, IntegrationAuthError, IntegrationSchemaMismatch, IntegrationSyncing, IntegrationStale, IntegrationPartiallyAvailable, IntegrationUnavailable:
		return true
	}
	return false
}

type ExternalIdentity struct{ ConnectionID, SourceSystem, SourceEntity, SourceID string }

var ErrInvalidExternalIdentity = errors.New("invalid external identity")

func NewExternalIdentity(connectionID, sourceSystem, sourceEntity, sourceID string) (ExternalIdentity, error) {
	i := ExternalIdentity{strings.TrimSpace(connectionID), strings.ToUpper(strings.TrimSpace(sourceSystem)), strings.ToUpper(strings.TrimSpace(sourceEntity)), strings.TrimSpace(sourceID)}
	if i.ConnectionID == "" || i.SourceSystem == "" || i.SourceEntity == "" || i.SourceID == "" {
		return ExternalIdentity{}, ErrInvalidExternalIdentity
	}
	return i, nil
}
func (i ExternalIdentity) Key() string {
	return strings.Join([]string{i.ConnectionID, i.SourceSystem, i.SourceEntity, i.SourceID}, "\x1f")
}
func (i ExternalIdentity) Hash() string {
	h := sha256.Sum256([]byte(i.Key()))
	return hex.EncodeToString(h[:])
}

type RecordVersion struct {
	Identity    ExternalIdentity
	PayloadHash string
	IsCurrent   bool
}
