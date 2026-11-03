package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/financial"
)

const PaymentMethodsAPIResource = "formas-pagamentos"

var errBlingPaymentMethodSchemaMismatch = errors.New("Bling payment method schema mismatch")

type paymentMethodAPIIdentity struct {
	ID int64 `json:"id"`
}

type preparedPaymentMethodAPIRecord struct {
	List   json.RawMessage `json:"list"`
	Detail json.RawMessage `json:"detail"`
}

type paymentMethodFeeConfig struct {
	Rate          json.RawMessage `json:"aliquota"`
	Fixed         json.RawMessage `json:"valor"`
	SettlementDay json.RawMessage `json:"prazo"`
}

type paymentMethodDetailConfig struct {
	ID              int64                   `json:"id"`
	Destination     *int                    `json:"destino"`
	Fees            *paymentMethodFeeConfig `json:"taxas"`
	UseBusinessDays *bool                   `json:"utilizaDiasUteis"`
}

// parseBlingPaymentMethodTerms turns the read-only detail captured in the RAW
// sync into explicit calculation inputs. It does not write back to Bling or
// infer missing values from a similarly named payment method.
func parseBlingPaymentMethodTerms(record json.RawMessage, expectedID int64) (financial.BlingPaymentMethodTerms, error) {
	var prepared preparedPaymentMethodAPIRecord
	if expectedID <= 0 || len(record) == 0 || json.Unmarshal(record, &prepared) != nil || len(prepared.Detail) == 0 {
		return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
	}
	var detail paymentMethodDetailConfig
	if json.Unmarshal(prepared.Detail, &detail) != nil || detail.ID != expectedID || detail.ID <= 0 {
		return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
	}
	terms := financial.BlingPaymentMethodTerms{}
	if detail.Destination != nil {
		if *detail.Destination < int(financial.BlingDestinationReceivablePayable) || *detail.Destination > int(financial.BlingDestinationCashAndBanks) {
			return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
		}
		terms.Destination = financial.BlingPaymentDestination(*detail.Destination)
		terms.DestinationKnown = true
	}
	if detail.Fees != nil {
		var err error
		terms.FeeRate, err = optionalPaymentMethodMoney(detail.Fees.Rate)
		if err != nil {
			return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
		}
		terms.FeeRateKnown = len(detail.Fees.Rate) > 0 && string(detail.Fees.Rate) != "null"
		terms.FixedFee, err = optionalPaymentMethodMoney(detail.Fees.Fixed)
		if err != nil {
			return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
		}
		terms.FixedFeeKnown = len(detail.Fees.Fixed) > 0 && string(detail.Fees.Fixed) != "null"
		terms.SettlementDays, err = optionalPaymentMethodDays(detail.Fees.SettlementDay)
		if err != nil {
			return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
		}
		terms.SettlementKnown = len(detail.Fees.SettlementDay) > 0 && string(detail.Fees.SettlementDay) != "null"
	}
	if detail.UseBusinessDays != nil {
		terms.UseBusinessDays = *detail.UseBusinessDays
		terms.BusinessDaysKnown = true
	}
	if terms.FeeRate.IsNegative() || terms.FixedFee.IsNegative() {
		return financial.BlingPaymentMethodTerms{}, errBlingPaymentMethodSchemaMismatch
	}
	return terms, nil
}

func optionalPaymentMethodMoney(raw json.RawMessage) (domain.Money, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return domain.ZeroMoney(), nil
	}
	textValue, err := decodeAPIStringOrNumber(raw)
	if err != nil {
		return domain.Money{}, errBlingPaymentMethodSchemaMismatch
	}
	value, err := domain.ParseMoney(textValue)
	if err != nil || value.IsNegative() {
		return domain.Money{}, errBlingPaymentMethodSchemaMismatch
	}
	return value, nil
}

func optionalPaymentMethodDays(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var days int
	if json.Unmarshal(raw, &days) != nil || days < 0 || days > 3650 {
		return 0, errBlingPaymentMethodSchemaMismatch
	}
	return days, nil
}

// PaymentMethodsAPISyncService records both the immutable RAW response and a
// typed read-only snapshot of the payment-method settings used by T02.
type PaymentMethodsAPISyncService struct{ *apiResourceSyncService }

func NewPaymentMethodsAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *PaymentMethodsAPISyncService {
	preparePage, normalizePage := normalizePaymentMethodsAPISnapshotPage(client)
	var list func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error)
	if client != nil {
		list = func(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			return client.ListPaymentMethods(ctx, filter.Page, filter.Limit)
		}
	}
	return &PaymentMethodsAPISyncService{apiResourceSyncService: &apiResourceSyncService{
		pool:          pool,
		client:        client,
		maxPages:      maxBlingAPIPageReads,
		resource:      PaymentMethodsAPIResource,
		sourceEntity:  "bling.formas_pagamentos.page",
		list:          list,
		preparePage:   preparePage,
		normalizePage: normalizePage,
	}}
}

func (s *PaymentMethodsAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}

func normalizePaymentMethodsAPISnapshotPage(client *BlingAPIClient) (
	func(context.Context, pgtype.UUID, []json.RawMessage) ([]json.RawMessage, error),
	func(context.Context, *database.Queries, pgtype.UUID, pgtype.UUID, []json.RawMessage) error,
) {
	detailCache := make(map[int64]json.RawMessage)
	var cacheBatch pgtype.UUID

	prepare := func(ctx context.Context, batchID pgtype.UUID, records []json.RawMessage) ([]json.RawMessage, error) {
		if client == nil || !batchID.Valid {
			return nil, fmt.Errorf("%w: payment method preparation dependencies are missing", ErrBlingAPIConfiguration)
		}
		if !cacheBatch.Valid || cacheBatch.Bytes != batchID.Bytes {
			clear(detailCache)
			cacheBatch = batchID
		}
		prepared := make([]json.RawMessage, 0, len(records))
		for _, rawRecord := range records {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			var identity paymentMethodAPIIdentity
			if len(rawRecord) == 0 || json.Unmarshal(rawRecord, &identity) != nil || identity.ID <= 0 {
				return nil, errors.Join(errBlingPaymentMethodSchemaMismatch, ErrBlingAPISchemaMismatch)
			}
			detail, ok := detailCache[identity.ID]
			if !ok {
				var err error
				detail, err = client.GetPaymentMethod(ctx, identity.ID)
				if err != nil {
					return nil, err
				}
				var detailIdentity paymentMethodAPIIdentity
				if len(detail) == 0 || json.Unmarshal(detail, &detailIdentity) != nil || detailIdentity.ID != identity.ID {
					return nil, errors.Join(errBlingPaymentMethodSchemaMismatch, ErrBlingAPISchemaMismatch)
				}
				detailCache[identity.ID] = detail
			}
			encoded, err := json.Marshal(preparedPaymentMethodAPIRecord{List: rawRecord, Detail: detail})
			if err != nil {
				return nil, errors.Join(errBlingPaymentMethodSchemaMismatch, err)
			}
			prepared = append(prepared, encoded)
		}
		return prepared, nil
	}

	normalize := func(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, records []json.RawMessage) error {
		if queries == nil || !connectionID.Valid || !batchID.Valid {
			return fmt.Errorf("%w: payment method normalizer dependencies are missing", ErrBlingAPIConfiguration)
		}
		for _, record := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			var prepared preparedPaymentMethodAPIRecord
			if len(record) == 0 || json.Unmarshal(record, &prepared) != nil || len(prepared.Detail) == 0 {
				return errors.Join(errBlingPaymentMethodSchemaMismatch, ErrBlingAPISchemaMismatch)
			}
			var identity paymentMethodAPIIdentity
			if json.Unmarshal(prepared.List, &identity) != nil || identity.ID <= 0 {
				return errors.Join(errBlingPaymentMethodSchemaMismatch, ErrBlingAPISchemaMismatch)
			}
			terms, err := parseBlingPaymentMethodTerms(record, identity.ID)
			if err != nil {
				return errors.Join(errBlingPaymentMethodSchemaMismatch, err)
			}
			rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, "bling.forma_pagamento", strconv.FormatInt(identity.ID, 10), record)
			if err != nil {
				return fmt.Errorf("persist Bling payment method source: %w", err)
			}
			feeRate, err := knownPaymentMethodMoney(terms.FeeRate)
			if err != nil {
				return err
			}
			fixedFee, err := knownPaymentMethodMoney(terms.FixedFee)
			if err != nil {
				return err
			}
			destination := (*int16)(nil)
			if terms.DestinationKnown {
				value := int16(terms.Destination)
				destination = &value
			}
			settlementDays := (*int32)(nil)
			if terms.SettlementKnown {
				value := int32(terms.SettlementDays)
				settlementDays = &value
			}
			businessDays := (*bool)(nil)
			if terms.BusinessDaysKnown {
				value := terms.UseBusinessDays
				businessDays = &value
			}
			if err := queries.UpsertBlingPaymentMethodTerms(ctx, database.UpsertBlingPaymentMethodTermsParams{
				ConnectionID: connectionID, SyncBatchID: batchID, SourceID: strconv.FormatInt(identity.ID, 10),
				Destination: destination, FeeRate: feeRate, FeeRateKnown: terms.FeeRateKnown,
				FixedFee: fixedFee, FixedFeeKnown: terms.FixedFeeKnown,
				SettlementDays: settlementDays, SettlementKnown: terms.SettlementKnown,
				UseBusinessDays: businessDays, BusinessDaysKnown: terms.BusinessDaysKnown,
				RawRecordID: rawID,
			}); err != nil {
				return fmt.Errorf("persist normalized Bling payment method terms: %w", err)
			}
		}
		return nil
	}
	return prepare, normalize
}

func knownPaymentMethodMoney(value domain.Money) (pgtype.Numeric, error) {
	parsed, err := numeric(value)
	if err != nil {
		return pgtype.Numeric{}, errBlingPaymentMethodSchemaMismatch
	}
	return parsed, nil
}
