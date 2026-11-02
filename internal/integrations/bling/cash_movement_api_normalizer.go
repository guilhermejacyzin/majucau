package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	database "majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/domain"
)

const (
	blingCashMovementSourceEntity = "bling.caixa"
	maxCashMovementSourceIDBytes  = 64
	maxCashMovementAccountName    = 512
	maxCashMovementOriginType     = 128
	maxCashMovementStatusBytes    = 128
)

type cashMovementSchemaFieldError struct{ code string }

func (e *cashMovementSchemaFieldError) Error() string {
	return "Bling cash movement field is invalid"
}

func cashMovementFieldError(code string) error {
	return &cashMovementSchemaFieldError{code: code}
}

type cashMovementAPIAccount struct {
	ID          json.RawMessage `json:"id"`
	Description *string         `json:"descricao"`
}

type cashMovementAPIOrigin struct {
	ID   json.RawMessage `json:"id"`
	Type *string         `json:"tipo"`
}

type cashMovementAPIRecord struct {
	ID           json.RawMessage         `json:"id"`
	Direction    string                  `json:"debCred"`
	Status       string                  `json:"situacao"`
	Amount       json.RawMessage         `json:"valor"`
	MovementDate string                  `json:"data"`
	Account      *cashMovementAPIAccount `json:"contaFinanceira"`
	Origin       *cashMovementAPIOrigin  `json:"origem"`
}

type normalizedCashMovement struct {
	SourceID        string
	AccountSourceID *string
	AccountName     *string
	Direction       string
	Amount          pgtype.Numeric
	MovementDate    pgtype.Date
	SourceStatus    string
	OriginType      *string
}

func normalizeCashMovementAPISnapshotPage() (func(context.Context, *database.Queries, pgtype.UUID, pgtype.UUID, []json.RawMessage) error, func() string) {
	accountMissing := false
	statusIncomplete := false
	normalize := func(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, records []json.RawMessage) error {
		if queries == nil || !connectionID.Valid || !batchID.Valid {
			return fmt.Errorf("%w: cash movement normalizer dependencies are missing", ErrBlingAPIConfiguration)
		}
		for _, payload := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			movement, err := parseCashMovementAPIRecord(payload)
			if err != nil {
				return err
			}
			rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, blingCashMovementSourceEntity, movement.SourceID, payload)
			if err != nil {
				return fmt.Errorf("persist Bling cash movement source: %w", err)
			}
			if movement.AccountSourceID == nil {
				accountMissing = true
			}
			if !cashMovementStatusIsKnown(movement.SourceStatus) {
				statusIncomplete = true
			}
			if err := queries.UpsertBlingCashMovement(ctx, database.UpsertBlingCashMovementParams{
				ConnectionID:    connectionID,
				SyncBatchID:     batchID,
				SourceID:        movement.SourceID,
				AccountSourceID: movement.AccountSourceID,
				AccountName:     movement.AccountName,
				Direction:       movement.Direction,
				Amount:          movement.Amount,
				MovementDate:    movement.MovementDate,
				SourceStatus:    movement.SourceStatus,
				OriginType:      movement.OriginType,
				RawRecordID:     rawID,
			}); err != nil {
				return fmt.Errorf("persist normalized Bling cash movement: %w", err)
			}
		}
		return nil
	}
	partialErrorCode := func() string {
		if accountMissing {
			return "BLING_CASH_ACCOUNT_ID_MISSING"
		}
		if statusIncomplete {
			return "BLING_CASH_MOVEMENT_STATUS_MISSING"
		}
		return ""
	}
	return normalize, partialErrorCode
}

func cashMovementStatusIsKnown(value string) bool {
	return strings.TrimSpace(value) == "R"
}

func parseCashMovementAPIRecord(payload json.RawMessage) (normalizedCashMovement, error) {
	var record cashMovementAPIRecord
	if len(payload) == 0 || json.Unmarshal(payload, &record) != nil {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_RECORD_INVALID")
	}
	sourceID, ok := parseBlingPositiveSourceID(record.ID)
	if !ok {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_ID_INVALID")
	}
	direction := ""
	switch strings.TrimSpace(record.Direction) {
	case "D":
		direction = "OUTFLOW"
	case "C":
		direction = "INFLOW"
	default:
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_DIRECTION_INVALID")
	}
	if !validCashMovementSourceStatus(record.Status) {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_STATUS_INVALID")
	}
	amountText, ok := decodeCashMovementDecimal(record.Amount)
	if !ok {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_AMOUNT_INVALID")
	}
	amount, err := domain.ParseMoney(amountText)
	if err != nil {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_AMOUNT_INVALID")
	}
	if (direction == "OUTFLOW" && amount.IsPositive()) || (direction == "INFLOW" && amount.IsNegative()) {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_DIRECTION_AMOUNT_MISMATCH")
	}
	amountMagnitude, err := amount.AbsExact()
	if err != nil {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_AMOUNT_INVALID")
	}
	date, err := parseCashMovementAPIDate(record.MovementDate)
	if err != nil {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_DATE_INVALID")
	}
	accountID, accountName := parseCashMovementAccount(record.Account)
	originType := optionalCashMovementOriginType(record.Origin)
	numericAmount, err := numeric(amountMagnitude)
	if err != nil {
		return normalizedCashMovement{}, cashMovementFieldError("BLING_CASH_MOVEMENT_AMOUNT_INVALID")
	}
	return normalizedCashMovement{
		SourceID: sourceID, AccountSourceID: accountID, AccountName: accountName,
		Direction: direction, Amount: numericAmount,
		MovementDate: pgtype.Date{Time: date, Valid: true}, SourceStatus: record.Status,
		OriginType: originType,
	}, nil
}

func parseBlingPositiveSourceID(raw json.RawMessage) (string, bool) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" {
		return "", false
	}
	if strings.HasPrefix(value, "\"") {
		if json.Unmarshal(raw, &value) != nil {
			return "", false
		}
		value = strings.TrimSpace(value)
	}
	if len(value) == 0 || len(value) > maxCashMovementSourceIDBytes || value[0] == '0' {
		return "", false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return "", false
		}
	}
	return value, true
}

func decodeCashMovementDecimal(raw json.RawMessage) (string, bool) {
	value := strings.TrimSpace(string(raw))
	if value == "" || value == "null" || len(value) > 32 || strings.ContainsAny(value, "eE,") {
		return "", false
	}
	if strings.HasPrefix(value, "\"") {
		if json.Unmarshal(raw, &value) != nil {
			return "", false
		}
		value = strings.TrimSpace(value)
	}
	if _, err := domain.ParseMoney(value); err != nil {
		return "", false
	}
	return value, true
}

func parseCashMovementAPIDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.DateOnly, time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, errors.New("invalid Bling cash movement date")
}

func validCashMovementSourceStatus(value string) bool {
	// Bling's public schema marks situacao as an optional string and gives R
	// only as an example. Preserve a missing/empty or new value without assigning
	// financial meaning to it; dashboard queries decide which statuses count.
	return len(value) <= maxCashMovementStatusBytes
}

func parseCashMovementAccount(account *cashMovementAPIAccount) (*string, *string) {
	if account == nil {
		return nil, nil
	}
	id, ok := parseBlingPositiveSourceID(account.ID)
	if !ok {
		return nil, nil
	}
	accountID := &id
	if account.Description == nil {
		return accountID, nil
	}
	name := strings.TrimSpace(*account.Description)
	if name == "" || len(name) > maxCashMovementAccountName {
		return accountID, nil
	}
	return accountID, &name
}

func optionalCashMovementOriginType(origin *cashMovementAPIOrigin) *string {
	if origin == nil || origin.Type == nil {
		return nil
	}
	value := strings.TrimSpace(*origin.Type)
	if value == "" || len(value) > maxCashMovementOriginType {
		return nil
	}
	return &value
}
