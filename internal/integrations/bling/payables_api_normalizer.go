package bling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	database "majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/domain"
)

const (
	payableAPISourceEntity = "bling.conta_pagar"
	maxPayableTextBytes    = 512
	maxPayableLookupCache  = 2048
)

var (
	errBlingPayableListSchemaMismatch    = errors.New("Bling payable list schema mismatch")
	errBlingPayableDetailSchemaMismatch  = errors.New("Bling payable detail schema mismatch")
	errBlingContactDetailSchemaMismatch  = errors.New("Bling contact detail schema mismatch")
	errBlingCategoryDetailSchemaMismatch = errors.New("Bling category detail schema mismatch")
)

type payableAPIListRecord struct {
	ID int64 `json:"id"`
}

type payableAPIContact struct {
	ID int64 `json:"id"`
}

type payableAPIContactDetail struct {
	ID   int64  `json:"id"`
	Name string `json:"nome"`
}

type payableAPICategory struct {
	ID *int64 `json:"id"`
}

type payableAPICategoryDetail struct {
	ID          int64  `json:"id"`
	Description string `json:"descricao"`
	Type        int64  `json:"tipo"`
}

type payableAPIDetail struct {
	ID             int64               `json:"id"`
	Status         int64               `json:"situacao"`
	DueDate        string              `json:"vencimento"`
	Amount         json.RawMessage     `json:"valor"`
	OpenBalance    json.RawMessage     `json:"saldo"`
	CompetenceDate *string             `json:"competencia"`
	Document       *string             `json:"numeroDocumento"`
	Category       *payableAPICategory `json:"categoria"`
	Contact        *payableAPIContact  `json:"contato"`
}

type normalizedPayable struct {
	SourceID           string
	SupplierSource     string
	SupplierName       string
	SupplierRawID      pgtype.UUID
	CategoryID         int64
	CategoryUnassigned bool
	CategoryRawID      pgtype.UUID
	Document           *string
	Category           *string
	CompetenceDate     pgtype.Date
	DueDate            pgtype.Date
	PaymentDate        pgtype.Date
	OriginalAmount     pgtype.Numeric
	OpenBalance        pgtype.Numeric
	Status             string
}

// payableSchemaFieldError carries a reviewed field identifier for the batch
// audit. Its text intentionally contains no provider payload or field value.
type payableSchemaFieldError struct {
	code string
}

// payableAPIMoneyError reports only a reviewed, non-sensitive reason. It never
// retains or formats the provider's financial value.
type payableAPIMoneyError struct {
	code string
}

func (e *payableAPIMoneyError) Error() string { return "invalid provider money field" }

func (e *payableSchemaFieldError) Error() string { return ErrBlingAPISchemaMismatch.Error() }

func (e *payableSchemaFieldError) Unwrap() error { return ErrBlingAPISchemaMismatch }

func payableSchemaField(code string) error { return &payableSchemaFieldError{code: code} }

func normalizePayablesAPISnapshotPage(client *BlingAPIClient) (func(context.Context, *database.Queries, pgtype.UUID, pgtype.UUID, []json.RawMessage) error, func() string) {
	contactCache := make(map[int64]string)
	categoryCache := make(map[int64]string)
	contactRawCache := make(map[int64]pgtype.UUID)
	categoryRawCache := make(map[int64]pgtype.UUID)
	categoryScopeUnavailable := false
	categoryUnassigned := false
	var cachedBatch pgtype.UUID
	normalize := func(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, records []json.RawMessage) error {
		if client == nil || queries == nil || !connectionID.Valid || !batchID.Valid {
			return fmt.Errorf("%w: payable normalizer dependencies are missing", ErrBlingAPIConfiguration)
		}
		if !cachedBatch.Valid || cachedBatch.Bytes != batchID.Bytes {
			clear(contactCache)
			clear(categoryCache)
			clear(contactRawCache)
			clear(categoryRawCache)
			categoryScopeUnavailable = false
			categoryUnassigned = false
			cachedBatch = batchID
		}
		for _, record := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			var listItem payableAPIListRecord
			if err := json.Unmarshal(record, &listItem); err != nil || listItem.ID <= 0 {
				return errors.Join(errBlingPayableListSchemaMismatch, payableSchemaField("BLING_PAYABLE_LIST_ID_INVALID"))
			}
			detailPayload, err := client.GetPayable(ctx, listItem.ID)
			if err != nil {
				return err
			}
			payable, err := parsePayableAPIDetail(detailPayload, listItem.ID)
			if err != nil {
				return errors.Join(errBlingPayableDetailSchemaMismatch, err)
			}
			if payable.CategoryUnassigned {
				categoryUnassigned = true
			}
			if payable.SupplierSource != "" {
				contactID, _ := strconv.ParseInt(payable.SupplierSource, 10, 64)
				name, ok := contactCache[contactID]
				if !ok {
					contactPayload, err := client.GetContact(ctx, contactID)
					if err != nil {
						return err
					}
					name, err = parsePayableContactDetail(contactPayload, contactID)
					if err != nil {
						return errors.Join(errBlingContactDetailSchemaMismatch, err)
					}
					contactRawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, "bling.contato", payable.SupplierSource, contactPayload)
					if err != nil {
						return fmt.Errorf("persist Bling contact source: %w", err)
					}
					if len(contactCache) >= maxPayableLookupCache {
						clear(contactCache)
						clear(contactRawCache)
					}
					contactCache[contactID] = name
					contactRawCache[contactID] = contactRawID
				}
				payable.SupplierName = name
				payable.SupplierRawID = contactRawCache[contactID]
			}
			if payable.CategoryID > 0 && !categoryScopeUnavailable {
				name, ok := categoryCache[payable.CategoryID]
				if !ok {
					categoryPayload, err := client.GetFinancialCategory(ctx, payable.CategoryID)
					if err != nil {
						if errors.Is(err, ErrBlingFinancialCategoryScopeInsufficient) {
							categoryScopeUnavailable = true
						} else {
							return err
						}
					}
					if !categoryScopeUnavailable {
						name, err = parsePayableCategoryDetail(categoryPayload, payable.CategoryID)
						if err != nil {
							return errors.Join(errBlingCategoryDetailSchemaMismatch, err)
						}
						categoryRawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, "bling.categoria_receita_despesa", strconv.FormatInt(payable.CategoryID, 10), categoryPayload)
						if err != nil {
							return fmt.Errorf("persist Bling category source: %w", err)
						}
						if len(categoryCache) >= maxPayableLookupCache {
							clear(categoryCache)
							clear(categoryRawCache)
						}
						categoryCache[payable.CategoryID] = name
						categoryRawCache[payable.CategoryID] = categoryRawID
					}
				}
				if !categoryScopeUnavailable {
					payable.Category = &name
					payable.CategoryRawID = categoryRawCache[payable.CategoryID]
				}
			}
			rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, payableAPISourceEntity, payable.SourceID, detailPayload)
			if err != nil {
				return fmt.Errorf("persist Bling payable source: %w", err)
			}
			var supplierID pgtype.UUID
			if payable.SupplierSource != "" {
				supplierID, err = queries.UpsertBlingSupplierContact(ctx, database.UpsertBlingSupplierContactParams{
					PConnectionID: connectionID,
					PSourceID:     payable.SupplierSource,
					PName:         payable.SupplierName,
					PRawRecordID:  payable.SupplierRawID,
				})
				if err != nil {
					return fmt.Errorf("persist Bling payable supplier: %w", err)
				}
			}
			payableID, err := queries.UpsertBlingPayable(ctx, database.UpsertBlingPayableParams{
				ConnectionID:   connectionID,
				SourceID:       payable.SourceID,
				SupplierID:     supplierID,
				Document:       payable.Document,
				Category:       payable.Category,
				CompetenceDate: payable.CompetenceDate,
				DueDate:        payable.DueDate,
				PaymentDate:    payable.PaymentDate,
				OriginalAmount: payable.OriginalAmount,
				OpenBalance:    payable.OpenBalance,
				Status:         payable.Status,
				RawRecordID:    rawID,
			})
			if err != nil {
				return fmt.Errorf("persist normalized Bling payable: %w", err)
			}
			if err := queries.InsertBlingPayableSource(ctx, database.InsertBlingPayableSourceParams{PayableID: payableID, RawRecordID: rawID}); err != nil {
				return fmt.Errorf("link normalized payable to RAW source: %w", err)
			}
			for _, relatedRawID := range []pgtype.UUID{payable.SupplierRawID, payable.CategoryRawID} {
				if relatedRawID.Valid && relatedRawID.Bytes != rawID.Bytes {
					if err := queries.InsertBlingPayableSource(ctx, database.InsertBlingPayableSourceParams{PayableID: payableID, RawRecordID: relatedRawID}); err != nil {
						return fmt.Errorf("link payable reference to RAW source: %w", err)
					}
				}
			}
			inserted, err := queries.InsertPayableSyncSnapshotRow(ctx, database.InsertPayableSyncSnapshotRowParams{SyncBatchID: batchID, PayableID: payableID})
			if err != nil {
				return fmt.Errorf("link payable to daily snapshot: %w", err)
			}
			if inserted != 1 {
				return fmt.Errorf("%w: duplicate payable identity in daily snapshot", ErrBlingAPISchemaMismatch)
			}
		}
		return nil
	}
	return normalize, func() string {
		switch {
		case categoryScopeUnavailable && categoryUnassigned:
			return "BLING_CATEGORY_INCOMPLETE"
		case categoryScopeUnavailable:
			return "BLING_CATEGORY_SCOPE_INSUFFICIENT"
		case categoryUnassigned:
			return "BLING_CATEGORY_UNASSIGNED"
		default:
			return ""
		}
	}
}

func parsePayableAPIDetail(payload []byte, expectedID int64) (normalizedPayable, error) {
	var detail payableAPIDetail
	if len(payload) == 0 || json.Unmarshal(payload, &detail) != nil || detail.ID != expectedID || detail.ID <= 0 {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_DETAIL_ID_INVALID")
	}
	if detail.Status < 1 || detail.Status > 7 {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_STATUS_INVALID")
	}
	dueDate, err := parseRequiredAPIDate(detail.DueDate)
	if err != nil {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_DUE_DATE_INVALID")
	}
	amount, err := parseRequiredAPIMoney(detail.Amount)
	if err != nil {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_AMOUNT_INVALID")
	}
	openBalance, err := parseRequiredAPIMoney(detail.OpenBalance)
	if err != nil {
		var moneyErr *payableAPIMoneyError
		if errors.As(err, &moneyErr) {
			return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_OPEN_BALANCE_" + moneyErr.code)
		}
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_OPEN_BALANCE_INVALID")
	}
	var competenceDate pgtype.Date
	if detail.CompetenceDate != nil && strings.TrimSpace(*detail.CompetenceDate) != "" {
		competenceDate, err = parseOptionalAPIDate(*detail.CompetenceDate)
		if err != nil {
			return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_COMPETENCE_DATE_INVALID")
		}
	}
	document, err := optionalPayableText(detail.Document)
	if err != nil {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_DOCUMENT_INVALID")
	}
	result := normalizedPayable{
		SourceID: strconv.FormatInt(detail.ID, 10), Document: document,
		CompetenceDate: competenceDate, DueDate: dueDate,
		OriginalAmount: amount, OpenBalance: openBalance, Status: strconv.FormatInt(detail.Status, 10),
	}
	if detail.Category != nil {
		if detail.Category.ID == nil || *detail.Category.ID < 0 {
			return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_CATEGORY_ID_INVALID")
		}
		if *detail.Category.ID == 0 {
			result.CategoryUnassigned = true
		} else {
			result.CategoryID = *detail.Category.ID
		}
	}
	if detail.Contact == nil || detail.Contact.ID <= 0 {
		return normalizedPayable{}, payableSchemaField("BLING_PAYABLE_SUPPLIER_ID_INVALID")
	}
	result.SupplierSource = strconv.FormatInt(detail.Contact.ID, 10)
	return result, nil
}

func parseRequiredAPIMoney(raw json.RawMessage) (pgtype.Numeric, error) {
	textValue, err := decodeAPIStringOrNumber(raw)
	if err != nil {
		return pgtype.Numeric{}, &payableAPIMoneyError{code: classifyPayableMoneyShape(raw)}
	}
	if !isExactPayableDecimal(textValue) {
		return pgtype.Numeric{}, &payableAPIMoneyError{code: classifyPayableMoneyShape(raw)}
	}
	var value pgtype.Numeric
	if err := value.Scan(textValue); err != nil || !value.Valid || value.Int == nil || value.Int.Sign() < 0 {
		return pgtype.Numeric{}, &payableAPIMoneyError{code: "VALUE_INVALID"}
	}
	return value, nil
}

func classifyPayableMoneyShape(raw json.RawMessage) string {
	value := strings.TrimSpace(string(raw))
	switch {
	case value == "" || value == "null":
		return "MISSING"
	case strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") || value == "true" || value == "false":
		return "TYPE_INVALID"
	case strings.HasPrefix(value, `"`):
		var decoded string
		if json.Unmarshal(raw, &decoded) != nil {
			return "FORMAT_INVALID"
		}
		return classifyPayableMoneyText(strings.TrimSpace(decoded))
	default:
		return classifyPayableMoneyText(value)
	}
}

func classifyPayableMoneyText(value string) string {
	if value == "" {
		return "MISSING"
	}
	if strings.HasPrefix(value, "-") && isExactPayableDecimal(strings.TrimPrefix(value, "-")) {
		return "NEGATIVE"
	}
	if strings.ContainsAny(value, "eE") {
		return "SCIENTIFIC_NOTATION"
	}
	if strings.Contains(value, ",") {
		return "DECIMAL_SEPARATOR_INVALID"
	}
	if len(value) > 20 {
		return "OUT_OF_RANGE"
	}
	whole, fraction := value, ""
	if decimalAt := strings.IndexByte(value, '.'); decimalAt >= 0 {
		if strings.IndexByte(value[decimalAt+1:], '.') >= 0 {
			return "FORMAT_INVALID"
		}
		whole, fraction = value[:decimalAt], value[decimalAt+1:]
	}
	if allASCIIDigits(whole) && len(whole) > 15 {
		return "OUT_OF_RANGE"
	}
	if allASCIIDigits(whole) && len(fraction) > 4 && allASCIIDigits(fraction) {
		return "PRECISION_EXCEEDED"
	}
	return "FORMAT_INVALID"
}

func allASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func isExactPayableDecimal(value string) bool {
	if value == "" || len(value) > 20 {
		return false
	}
	decimalAt := strings.IndexByte(value, '.')
	whole := value
	fraction := ""
	if decimalAt >= 0 {
		if strings.IndexByte(value[decimalAt+1:], '.') >= 0 {
			return false
		}
		whole, fraction = value[:decimalAt], value[decimalAt+1:]
		if len(fraction) < 1 || len(fraction) > 4 {
			return false
		}
	}
	if whole == "" || len(whole) > 15 || (len(whole) > 1 && whole[0] == '0') {
		return false
	}
	for _, digit := range whole + fraction {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func decodeAPIStringOrNumber(raw json.RawMessage) (string, error) {
	textValue := strings.TrimSpace(string(raw))
	if textValue == "" || textValue == "null" {
		return "", errors.New("missing decimal")
	}
	if strings.HasPrefix(textValue, "\"") {
		var decoded string
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return "", err
		}
		textValue = strings.TrimSpace(decoded)
	}
	return textValue, nil
}

func parseRequiredAPIDate(value string) (pgtype.Date, error) {
	parsed, err := parseOptionalAPIDate(value)
	if err != nil || !parsed.Valid {
		return pgtype.Date{}, errors.New("missing date")
	}
	return parsed, nil
}

func parseOptionalAPIDate(value string) (pgtype.Date, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil || parsed.Format("2006-01-02") != strings.TrimSpace(value) {
		return pgtype.Date{}, errors.New("invalid date")
	}
	return pgtype.Date{Time: parsed, Valid: true}, nil
}

func optionalPayableText(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	clean := strings.TrimSpace(*value)
	if len(clean) > maxPayableTextBytes {
		return nil, errors.New("text exceeds field limit")
	}
	if clean == "" {
		return nil, nil
	}
	return &clean, nil
}

func parsePayableContactDetail(payload []byte, expectedID int64) (string, error) {
	var detail payableAPIContactDetail
	if len(payload) == 0 || json.Unmarshal(payload, &detail) != nil || detail.ID != expectedID {
		return "", payableSchemaField("BLING_CONTACT_DETAIL_ID_INVALID")
	}
	name, err := optionalPayableText(&detail.Name)
	if err != nil || name == nil {
		return "", payableSchemaField("BLING_CONTACT_DETAIL_NAME_INVALID")
	}
	return *name, nil
}

func parsePayableCategoryDetail(payload []byte, expectedID int64) (string, error) {
	var detail payableAPICategoryDetail
	if len(payload) == 0 || json.Unmarshal(payload, &detail) != nil || detail.ID != expectedID {
		return "", payableSchemaField("BLING_CATEGORY_DETAIL_ID_INVALID")
	}
	if detail.Type < 1 || detail.Type > 3 {
		return "", payableSchemaField("BLING_CATEGORY_DETAIL_TYPE_INVALID")
	}
	description, err := optionalPayableText(&detail.Description)
	if err != nil || description == nil {
		return "", payableSchemaField("BLING_CATEGORY_DETAIL_DESCRIPTION_INVALID")
	}
	return *description, nil
}

func persistCurrentBlingEntity(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, sourceEntity, sourceID string, payload []byte) (pgtype.UUID, error) {
	if queries == nil || !connectionID.Valid || !batchID.Valid || sourceEntity == "" || sourceID == "" || len(payload) == 0 {
		return pgtype.UUID{}, errors.New("invalid Bling entity identity")
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	current, err := queries.GetCurrentRawRecord(ctx, database.GetCurrentRawRecordParams{
		ConnectionID: connectionID, SourceSystem: string(domain.OriginBling), SourceEntity: sourceEntity, SourceID: sourceID,
	})
	if err == nil && current.PayloadHash == hash {
		return current.ID, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}
	if err == nil {
		if err := queries.RetireCurrentRawRecord(ctx, current.ID); err != nil {
			return pgtype.UUID{}, err
		}
	}
	inserted, err := queries.InsertRawRecord(ctx, database.InsertRawRecordParams{
		SyncBatchID: batchID, ConnectionID: connectionID, SourceSystem: string(domain.OriginBling),
		SourceEntity: sourceEntity, SourceID: sourceID, PayloadHash: hash, Payload: payload, IsCurrent: true,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	if !inserted.IsCurrent {
		return pgtype.UUID{}, ErrHistoricalRawVersion
	}
	return inserted.ID, nil
}
