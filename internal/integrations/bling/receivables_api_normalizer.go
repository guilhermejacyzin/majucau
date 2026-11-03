package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	database "majucau.local/financial-intelligence/database/gen"
)

const (
	receivableAPISourceEntity = "bling.conta_receber"
	maxReceivableTextBytes    = 512
)

var (
	errBlingReceivableListSchemaMismatch   = errors.New("Bling receivable list schema mismatch")
	errBlingReceivableDetailSchemaMismatch = errors.New("Bling receivable detail schema mismatch")
)

type receivableAPIContact struct {
	ID   int64  `json:"id"`
	Name string `json:"nome"`
}

type receivableAPISourceReference struct {
	ID int64 `json:"id"`
}

type receivableAPIListRecord struct {
	ID                int64                         `json:"id"`
	Status            int64                         `json:"situacao"`
	DueDate           string                        `json:"vencimento"`
	IssueDate         *string                       `json:"dataEmissao"`
	Amount            json.RawMessage               `json:"valor"`
	Document          *string                       `json:"numeroDocumento"`
	Competence        *string                       `json:"competencia"`
	Contact           *receivableAPIContact         `json:"contato"`
	PaymentMethod     *receivableAPISourceReference `json:"formaPagamento"`
	FinancialAccount  *receivableAPISourceReference `json:"portador"`
	AccountingAccount *receivableAPISourceReference `json:"contaContabil"`
}

type receivableAPIDetail struct {
	ID         int64           `json:"id"`
	Status     int64           `json:"situacao"`
	DueDate    string          `json:"vencimento"`
	IssueDate  *string         `json:"dataEmissao"`
	Amount     json.RawMessage `json:"valor"`
	OpenAmount json.RawMessage `json:"saldo"`
	Document   *string         `json:"numeroDocumento"`
	Competence *string         `json:"competencia"`
	Contact    *struct {
		ID int64 `json:"id"`
	} `json:"contato"`
	PaymentMethod     *receivableAPISourceReference `json:"formaPagamento"`
	FinancialAccount  *receivableAPISourceReference `json:"portador"`
	AccountingAccount *receivableAPISourceReference `json:"contaContabil"`
}

type preparedReceivableAPISnapshotRecord struct {
	List    json.RawMessage `json:"list"`
	Detail  json.RawMessage `json:"detail,omitempty"`
	Contact json.RawMessage `json:"contact"`
}

type normalizedReceivable struct {
	SourceID                  string
	CustomerSource            string
	CustomerName              string
	CustomerRawID             pgtype.UUID
	Document                  *string
	IssueDate                 pgtype.Date
	CompetenceDate            pgtype.Date
	DueDate                   pgtype.Date
	GrossAmount               pgtype.Numeric
	OpenBalance               pgtype.Numeric
	Status                    string
	SourcePaymentMethodID     *int64
	SourceFinancialAccountID  *int64
	SourceAccountingAccountID *int64
}

func normalizeReceivablesAPISnapshotPage(client *BlingAPIClient) (
	func(context.Context, pgtype.UUID, []json.RawMessage) ([]json.RawMessage, error),
	func(context.Context, *database.Queries, pgtype.UUID, pgtype.UUID, []json.RawMessage) error,
	func() string,
) {
	contactCache := make(map[int64]json.RawMessage)
	contactNameCache := make(map[int64]string)
	contactRawCache := make(map[int64]pgtype.UUID)
	var preparedBatch pgtype.UUID
	var normalizedBatch pgtype.UUID

	prepare := func(ctx context.Context, batchID pgtype.UUID, records []json.RawMessage) ([]json.RawMessage, error) {
		if client == nil || !batchID.Valid {
			return nil, fmt.Errorf("%w: receivable API preparation dependencies are missing", ErrBlingAPIConfiguration)
		}
		if !preparedBatch.Valid || preparedBatch.Bytes != batchID.Bytes {
			clear(contactCache)
			preparedBatch = batchID
		}
		preparedRecords := make([]json.RawMessage, 0, len(records))
		for _, record := range records {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			list, err := parseReceivableAPIListRecord(record)
			if err != nil {
				return nil, errors.Join(errBlingReceivableListSchemaMismatch, err)
			}
			prepared := preparedReceivableAPISnapshotRecord{List: record}
			if receivableNeedsDetail(list.Status) {
				detailPayload, err := client.GetReceivable(ctx, list.ID)
				if err != nil {
					return nil, err
				}
				detail, err := parseReceivableAPIDetail(detailPayload, list.ID)
				if err != nil || detail.Status != list.Status || detail.Contact.ID != list.Contact.ID {
					return nil, errors.Join(errBlingReceivableDetailSchemaMismatch, receivableSchemaField("BLING_RECEIVABLE_DETAIL_STATUS_MISMATCH"), err)
				}
				prepared.Detail = detailPayload
			}
			contactID := list.Contact.ID
			contactPayload, ok := contactCache[contactID]
			if !ok {
				contactPayload, err = json.Marshal(list.Contact)
				if err != nil {
					return nil, errors.Join(errBlingReceivableListSchemaMismatch, receivableSchemaField("BLING_RECEIVABLE_CONTACT_INVALID"))
				}
				contactName, parseErr := parseReceivableContactDetail(contactPayload, contactID)
				if parseErr != nil && contactID > 0 {
					contactPayload, err = client.GetContact(ctx, contactID)
					if err != nil {
						return nil, errors.Join(errBlingReceivableDetailSchemaMismatch, parseErr, err)
					}
					contactName, parseErr = parseReceivableContactDetail(contactPayload, contactID)
				}
				if parseErr != nil {
					return nil, errors.Join(errBlingReceivableDetailSchemaMismatch, parseErr)
				}
				_ = contactName
				if len(contactCache) >= maxPayableLookupCache {
					clear(contactCache)
				}
				contactCache[contactID] = contactPayload
			}
			prepared.Contact = contactPayload
			encoded, err := json.Marshal(prepared)
			if err != nil {
				return nil, fmt.Errorf("encode prepared Bling receivable: %w", err)
			}
			preparedRecords = append(preparedRecords, encoded)
		}
		return preparedRecords, nil
	}

	normalize := func(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, records []json.RawMessage) error {
		if queries == nil || !connectionID.Valid || !batchID.Valid {
			return fmt.Errorf("%w: receivable normalizer dependencies are missing", ErrBlingAPIConfiguration)
		}
		if !normalizedBatch.Valid || normalizedBatch.Bytes != batchID.Bytes {
			clear(contactNameCache)
			clear(contactRawCache)
			normalizedBatch = batchID
		}
		for _, record := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			var prepared preparedReceivableAPISnapshotRecord
			if len(record) == 0 || json.Unmarshal(record, &prepared) != nil || len(prepared.List) == 0 || len(prepared.Contact) == 0 {
				return errors.Join(errBlingReceivableDetailSchemaMismatch, receivableSchemaField("BLING_RECEIVABLE_PREPARED_RECORD_INVALID"))
			}
			list, err := parseReceivableAPIListRecord(prepared.List)
			if err != nil {
				return errors.Join(errBlingReceivableListSchemaMismatch, err)
			}
			receivable := normalizedReceivable{SourceID: strconv.FormatInt(list.ID, 10), CustomerSource: strconv.FormatInt(list.Contact.ID, 10)}
			if _, err := strconv.ParseInt(receivable.CustomerSource, 10, 64); err != nil {
				return errors.Join(errBlingReceivableListSchemaMismatch, receivableSchemaField("BLING_RECEIVABLE_CONTACT_INVALID"))
			}
			contactID, _ := strconv.ParseInt(receivable.CustomerSource, 10, 64)
			receivable.CustomerName, _ = contactNameCache[contactID]
			if receivable.CustomerName == "" {
				receivable.CustomerName, err = parseReceivableContactDetail(prepared.Contact, contactID)
				if err != nil {
					return errors.Join(errBlingReceivableDetailSchemaMismatch, err)
				}
				rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, "bling.contato", receivable.CustomerSource, prepared.Contact)
				if err != nil {
					return fmt.Errorf("persist Bling customer source: %w", err)
				}
				if len(contactNameCache) >= maxPayableLookupCache {
					clear(contactNameCache)
					clear(contactRawCache)
				}
				contactNameCache[contactID] = receivable.CustomerName
				contactRawCache[contactID] = rawID
			}
			receivable.CustomerRawID = contactRawCache[contactID]
			if !receivable.CustomerRawID.Valid {
				return errors.Join(errBlingReceivableDetailSchemaMismatch, receivableSchemaField("BLING_RECEIVABLE_CONTACT_SOURCE_MISSING"))
			}
			if len(prepared.Detail) > 0 {
				detail, err := parseReceivableAPIDetail(prepared.Detail, list.ID)
				if err != nil || detail.Status != list.Status || detail.Contact.ID != list.Contact.ID {
					return errors.Join(errBlingReceivableDetailSchemaMismatch, err)
				}
				receivable, err = normalizedReceivableFromDetail(detail, receivable)
				if err != nil {
					return errors.Join(errBlingReceivableDetailSchemaMismatch, err)
				}
			} else {
				receivable, err = normalizedReceivableFromList(list, receivable)
				if err != nil {
					return errors.Join(errBlingReceivableListSchemaMismatch, err)
				}
			}
			accountPayload := prepared.List
			if len(prepared.Detail) > 0 {
				accountPayload = prepared.Detail
			}
			rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, receivableAPISourceEntity, receivable.SourceID, accountPayload)
			if err != nil {
				return fmt.Errorf("persist Bling receivable source: %w", err)
			}
			customerID, err := queries.UpsertBlingCustomerContact(ctx, database.UpsertBlingCustomerContactParams{
				PConnectionID: connectionID, PSourceID: receivable.CustomerSource,
				PName: receivable.CustomerName, PRawRecordID: receivable.CustomerRawID,
			})
			if err != nil {
				return fmt.Errorf("persist normalized Bling customer: %w", err)
			}
			receivableID, err := queries.UpsertBlingReceivable(ctx, database.UpsertBlingReceivableParams{
				ConnectionID: connectionID, SourceID: receivable.SourceID, CustomerID: customerID,
				ExternalDocument: receivable.Document, IssueDate: receivable.IssueDate,
				CompetenceDate: receivable.CompetenceDate, DueDate: receivable.DueDate,
				GrossAmount: receivable.GrossAmount, OpenBalance: receivable.OpenBalance,
				Status: receivable.Status, RawRecordID: rawID,
				SourcePaymentMethodID:     receivable.SourcePaymentMethodID,
				SourceFinancialAccountID:  receivable.SourceFinancialAccountID,
				SourceAccountingAccountID: receivable.SourceAccountingAccountID,
			})
			if err != nil {
				return fmt.Errorf("persist normalized Bling receivable: %w", err)
			}
			for _, sourceRawID := range []pgtype.UUID{rawID, receivable.CustomerRawID} {
				if err := queries.InsertBlingReceivableSource(ctx, database.InsertBlingReceivableSourceParams{ReceivableID: receivableID, RawRecordID: sourceRawID}); err != nil {
					return fmt.Errorf("link Bling receivable to RAW source: %w", err)
				}
			}
			inserted, err := queries.InsertReceivableSyncSnapshotRow(ctx, database.InsertReceivableSyncSnapshotRowParams{SyncBatchID: batchID, ReceivableID: receivableID})
			if err != nil {
				return fmt.Errorf("link receivable to daily snapshot: %w", err)
			}
			if inserted != 1 {
				return fmt.Errorf("%w: duplicate receivable identity in daily snapshot", ErrBlingAPISchemaMismatch)
			}
		}
		return nil
	}
	return prepare, normalize, func() string { return "" }
}

func receivableNeedsDetail(status int64) bool { return status != 2 && status != 5 }

func parseReceivableAPIListRecord(payload []byte) (receivableAPIListRecord, error) {
	var list receivableAPIListRecord
	if len(payload) == 0 || json.Unmarshal(payload, &list) != nil || list.ID <= 0 {
		return receivableAPIListRecord{}, receivableSchemaField("BLING_RECEIVABLE_LIST_ID_INVALID")
	}
	if list.Status < 1 || list.Status > 7 {
		return receivableAPIListRecord{}, receivableSchemaField("BLING_RECEIVABLE_STATUS_INVALID")
	}
	if list.Contact == nil || list.Contact.ID <= 0 {
		return receivableAPIListRecord{}, receivableSchemaField("BLING_RECEIVABLE_CONTACT_INVALID")
	}
	if err := validateReceivableSourceReferences(list.PaymentMethod, list.FinancialAccount, list.AccountingAccount); err != nil {
		return receivableAPIListRecord{}, err
	}
	return list, nil
}

func parseReceivableAPIDetail(payload []byte, expectedID int64) (receivableAPIDetail, error) {
	var detail receivableAPIDetail
	if len(payload) == 0 || json.Unmarshal(payload, &detail) != nil || detail.ID != expectedID || detail.ID <= 0 {
		return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_DETAIL_ID_INVALID")
	}
	if detail.Status < 1 || detail.Status > 7 {
		return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_STATUS_INVALID")
	}
	if detail.Contact == nil || detail.Contact.ID <= 0 {
		return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_CONTACT_INVALID")
	}
	if err := validateReceivableSourceReferences(detail.PaymentMethod, detail.FinancialAccount, detail.AccountingAccount); err != nil {
		return receivableAPIDetail{}, err
	}
	if _, err := parseRequiredAPIDate(detail.DueDate); err != nil {
		return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_DUE_DATE_INVALID")
	}
	if _, err := parseRequiredAPIMoney(detail.Amount); err != nil {
		return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_AMOUNT_INVALID")
	}
	if receivableNeedsDetail(detail.Status) {
		if _, err := parseRequiredAPIMoney(detail.OpenAmount); err != nil {
			return receivableAPIDetail{}, receivableSchemaField("BLING_RECEIVABLE_OPEN_BALANCE_INVALID")
		}
	}
	return detail, nil
}

func validateReceivableSourceReferences(references ...*receivableAPISourceReference) error {
	for _, reference := range references {
		if reference != nil && reference.ID < 0 {
			return receivableSchemaField("BLING_RECEIVABLE_SOURCE_REFERENCE_INVALID")
		}
	}
	return nil
}

func receivableSourceReferenceID(reference *receivableAPISourceReference) *int64 {
	if reference == nil || reference.ID == 0 {
		return nil
	}
	value := reference.ID
	return &value
}

func normalizedReceivableFromList(list receivableAPIListRecord, base normalizedReceivable) (normalizedReceivable, error) {
	var err error
	base.SourcePaymentMethodID = receivableSourceReferenceID(list.PaymentMethod)
	base.SourceFinancialAccountID = receivableSourceReferenceID(list.FinancialAccount)
	base.SourceAccountingAccountID = receivableSourceReferenceID(list.AccountingAccount)
	base.GrossAmount, err = parseRequiredAPIMoney(list.Amount)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_AMOUNT_INVALID")
	}
	base.DueDate, err = parseRequiredAPIDate(list.DueDate)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_DUE_DATE_INVALID")
	}
	base.IssueDate, err = optionalReceivableDate(list.IssueDate)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_ISSUE_DATE_INVALID")
	}
	base.CompetenceDate, err = optionalReceivableDate(list.Competence)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_COMPETENCE_DATE_INVALID")
	}
	base.Document, err = optionalReceivableText(list.Document)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_DOCUMENT_INVALID")
	}
	base.OpenBalance = zeroReceivableNumeric()
	base.Status = receivableStatus(list.Status)
	return base, nil
}

func normalizedReceivableFromDetail(detail receivableAPIDetail, base normalizedReceivable) (normalizedReceivable, error) {
	var err error
	base.SourcePaymentMethodID = receivableSourceReferenceID(detail.PaymentMethod)
	base.SourceFinancialAccountID = receivableSourceReferenceID(detail.FinancialAccount)
	base.SourceAccountingAccountID = receivableSourceReferenceID(detail.AccountingAccount)
	base.GrossAmount, err = parseRequiredAPIMoney(detail.Amount)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_AMOUNT_INVALID")
	}
	base.OpenBalance = zeroReceivableNumeric()
	if receivableNeedsDetail(detail.Status) {
		base.OpenBalance, err = parseRequiredAPIMoney(detail.OpenAmount)
		if err != nil {
			return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_OPEN_BALANCE_INVALID")
		}
	}
	base.DueDate, err = parseRequiredAPIDate(detail.DueDate)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_DUE_DATE_INVALID")
	}
	base.IssueDate, err = optionalReceivableDate(detail.IssueDate)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_ISSUE_DATE_INVALID")
	}
	base.CompetenceDate, err = optionalReceivableDate(detail.Competence)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_COMPETENCE_DATE_INVALID")
	}
	base.Document, err = optionalReceivableText(detail.Document)
	if err != nil {
		return normalizedReceivable{}, receivableSchemaField("BLING_RECEIVABLE_DOCUMENT_INVALID")
	}
	base.Status = receivableStatus(detail.Status)
	return base, nil
}

func receivableStatus(status int64) string {
	switch status {
	case 1:
		return "OPEN"
	case 2:
		return "PAID"
	case 3:
		return "PARTIALLY_PAID"
	case 4:
		return "RETURNED"
	case 5:
		return "CANCELLED"
	case 6:
		return "PARTIALLY_RETURNED"
	case 7:
		return "CONFIRMED"
	default:
		return "UNCLASSIFIED"
	}
}

func parseReceivableContactDetail(payload []byte, expectedID int64) (string, error) {
	var contact receivableAPIContact
	if len(payload) == 0 || json.Unmarshal(payload, &contact) != nil || contact.ID != expectedID || contact.ID <= 0 {
		return "", receivableSchemaField("BLING_RECEIVABLE_CONTACT_INVALID")
	}
	name, err := optionalReceivableText(&contact.Name)
	if err != nil || name == nil {
		return "", receivableSchemaField("BLING_RECEIVABLE_CUSTOMER_NAME_INVALID")
	}
	return *name, nil
}

func optionalReceivableText(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	clean := strings.TrimSpace(*value)
	if len(clean) > maxReceivableTextBytes {
		return nil, errors.New("receivable text exceeds field limit")
	}
	if clean == "" {
		return nil, nil
	}
	return &clean, nil
}

func optionalReceivableDate(value *string) (pgtype.Date, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Date{}, nil
	}
	return parseOptionalAPIDate(*value)
}

func zeroReceivableNumeric() pgtype.Numeric {
	var value pgtype.Numeric
	_ = value.Scan("0")
	return value
}

func receivableSchemaField(code string) error { return &payableSchemaFieldError{code: code} }
