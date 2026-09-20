package bling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/domain"
)

var (
	ErrInvalidPersistenceIdentity = errors.New("invalid receipt persistence identity")
	ErrHistoricalRawVersion       = errors.New("raw payload hash already exists as a historical version")
)

type PersistenceSummary struct {
	RecordsRead    int `json:"records_read"`
	RecordsCreated int `json:"records_created"`
	RecordsUpdated int `json:"records_updated"`
	RecordsFailed  int `json:"records_failed"`
}

type receiptRawPayload struct {
	CustomerName     string `json:"customer_name"`
	History          string `json:"history"`
	PaymentMethod    string `json:"payment_method"`
	ExternalDocument string `json:"external_document"`
	DueDate          string `json:"due_date"`
	ReceiptDate      string `json:"receipt_date"`
	Fee              string `json:"fee"`
	Amount           string `json:"amount"`
	Status           string `json:"status"`
}

// PersistReceipts appends the source payload and upserts the normalized Bling
// receipt in the caller-owned transaction. A caller can commit the returned
// summary as SUCCESS/PARTIAL or roll the transaction back on any DB error.
func PersistReceipts(ctx context.Context, tx pgx.Tx, connectionID, syncBatchID pgtype.UUID, report ReceiptsReport) (PersistenceSummary, error) {
	if !connectionID.Valid || !syncBatchID.Valid {
		return PersistenceSummary{}, ErrInvalidPersistenceIdentity
	}
	if tx == nil {
		return PersistenceSummary{}, errors.New("receipt persistence transaction is required")
	}
	q := database.New(tx)
	summary := PersistenceSummary{RecordsRead: len(report.Receipts) + len(report.Errors), RecordsFailed: len(report.Errors)}
	for _, candidate := range report.Receipts {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		if candidate.SourceSystem != domain.OriginBling || candidate.SourceEntity != ReceiptsReportSourceEntity || candidate.Status != domain.StatusConfirmed {
			return summary, fmt.Errorf("%w: source=%s entity=%s status=%s", ErrInvalidPersistenceIdentity, candidate.SourceSystem, candidate.SourceEntity, candidate.Status)
		}
		if candidate.SourceID == "" || candidate.ReceiptDate.IsZero() || candidate.Amount.IsNegative() || candidate.Fee.IsNegative() {
			return summary, fmt.Errorf("%w: source_id, date and non-negative amounts are required", ErrInvalidPersistenceIdentity)
		}
		rawID, err := ensureCurrentRaw(ctx, q, connectionID, syncBatchID, candidate)
		if err != nil {
			return summary, fmt.Errorf("persist raw receipt %s: %w", candidate.SourceID, err)
		}
		exists, err := q.ReceiptExists(ctx, database.ReceiptExistsParams{ConnectionID: connectionID, SourceEntity: candidate.SourceEntity, SourceID: candidate.SourceID})
		if err != nil {
			return summary, fmt.Errorf("check receipt %s: %w", candidate.SourceID, err)
		}
		amount, err := numeric(candidate.Amount)
		if err != nil {
			return summary, fmt.Errorf("amount %s: %w", candidate.SourceID, err)
		}
		if _, err := q.UpsertReceipt(ctx, database.UpsertReceiptParams{
			ConnectionID: connectionID,
			SourceEntity: candidate.SourceEntity,
			SourceID:     candidate.SourceID,
			ReceiptDate:  date(candidate.ReceiptDate),
			Amount:       amount,
			Status:       string(candidate.Status),
			RawRecordID:  rawID,
		}); err != nil {
			return summary, fmt.Errorf("upsert receipt %s: %w", candidate.SourceID, err)
		}
		if exists {
			summary.RecordsUpdated++
		} else {
			summary.RecordsCreated++
		}
	}
	return summary, nil
}

func ensureCurrentRaw(ctx context.Context, q *database.Queries, connectionID, syncBatchID pgtype.UUID, candidate ReceiptCandidate) (pgtype.UUID, error) {
	payload, err := json.Marshal(receiptRawPayload{
		CustomerName: candidate.CustomerName, History: candidate.History, PaymentMethod: candidate.PaymentMethod,
		ExternalDocument: candidate.ExternalDocument, DueDate: candidate.DueDate.Format("2006-01-02"),
		ReceiptDate: candidate.ReceiptDate.Format("2006-01-02"), Fee: candidate.Fee.String(), Amount: candidate.Amount.String(),
		Status: string(candidate.Status),
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	current, err := q.GetCurrentRawRecord(ctx, database.GetCurrentRawRecordParams{
		ConnectionID: connectionID, SourceSystem: string(domain.OriginBling), SourceEntity: candidate.SourceEntity, SourceID: candidate.SourceID,
	})
	if err == nil {
		if current.PayloadHash == hash {
			return current.ID, nil
		}
		if err := q.RetireCurrentRawRecord(ctx, current.ID); err != nil {
			return pgtype.UUID{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}
	inserted, err := q.InsertRawRecord(ctx, database.InsertRawRecordParams{
		SyncBatchID: syncBatchID, ConnectionID: connectionID, SourceSystem: string(domain.OriginBling),
		SourceEntity: candidate.SourceEntity, SourceID: candidate.SourceID, PayloadHash: hash, Payload: payload,
		SourceUpdatedAt: pgtype.Timestamptz{}, IsCurrent: true,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	if !inserted.IsCurrent {
		return pgtype.UUID{}, ErrHistoricalRawVersion
	}
	return inserted.ID, nil
}

func date(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: !value.IsZero()}
}

func numeric(value domain.Money) (pgtype.Numeric, error) {
	var result pgtype.Numeric
	if err := result.Scan(value.String()); err != nil {
		return pgtype.Numeric{}, err
	}
	return result, nil
}
