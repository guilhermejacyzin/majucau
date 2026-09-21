package nuvempago

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
	ErrInvalidFuturePersistenceIdentity = errors.New("invalid future receivable persistence identity")
	ErrFutureRawHistoricalVersion       = errors.New("future raw payload hash already exists as a historical version")
)

type FuturePersistenceSummary struct {
	RecordsRead    int `json:"records_read"`
	RecordsCreated int `json:"records_created"`
	RecordsUpdated int `json:"records_updated"`
	RecordsFailed  int `json:"records_failed"`
}

// PersistFutureReceivables writes Nuvem Pago projected rights to receive. It
// never touches the Bling-only receipts table. The caller owns the transaction
// so RAW and the normalized upsert commit atomically.
func PersistFutureReceivables(ctx context.Context, tx pgx.Tx, connectionID, syncBatchID pgtype.UUID, report FutureReport) (FuturePersistenceSummary, error) {
	if !connectionID.Valid || !syncBatchID.Valid {
		return FuturePersistenceSummary{}, ErrInvalidFuturePersistenceIdentity
	}
	if tx == nil {
		return FuturePersistenceSummary{}, errors.New("future receivable persistence transaction is required")
	}
	queries := database.New(tx)
	summary := FuturePersistenceSummary{RecordsRead: len(report.Receivables) + len(report.Errors), RecordsFailed: len(report.Errors)}
	for _, candidate := range report.Receivables {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		if candidate.SourceSystem != domain.OriginNuvemPago || candidate.SourceEntity != FutureSourceEntity || candidate.Status != domain.StatusProjected || candidate.SourceID == "" {
			return summary, fmt.Errorf("%w: source=%s entity=%s status=%s id=%s", ErrInvalidFuturePersistenceIdentity, candidate.SourceSystem, candidate.SourceEntity, candidate.Status, candidate.SourceID)
		}
		if candidate.PaymentDate.IsZero() || candidate.ExpectedReceiptDate.IsZero() || candidate.GrossAmount.IsNegative() || candidate.FeeAmount.IsNegative() || candidate.NetAmount.IsNegative() {
			return summary, fmt.Errorf("%w: dates and non-negative amounts are required", ErrInvalidFuturePersistenceIdentity)
		}
		rawID, err := ensureCurrentFutureRaw(ctx, queries, connectionID, syncBatchID, candidate)
		if err != nil {
			return summary, fmt.Errorf("persist future raw %s: %w", candidate.SourceID, err)
		}
		exists, err := futureReceivableExists(ctx, tx, connectionID, candidate.SourceID)
		if err != nil {
			return summary, fmt.Errorf("check future receivable %s: %w", candidate.SourceID, err)
		}
		gross, err := futureNumeric(candidate.GrossAmount)
		if err != nil {
			return summary, fmt.Errorf("gross amount %s: %w", candidate.SourceID, err)
		}
		fee, err := futureNumeric(candidate.FeeAmount)
		if err != nil {
			return summary, fmt.Errorf("fee amount %s: %w", candidate.SourceID, err)
		}
		net, err := futureNumeric(candidate.NetAmount)
		if err != nil {
			return summary, fmt.Errorf("net amount %s: %w", candidate.SourceID, err)
		}
		var installmentCount any
		if candidate.InstallmentCount != nil {
			installmentCount = int32(*candidate.InstallmentCount)
		}
		if _, err := tx.Exec(ctx, upsertFutureReceivableSQL,
			connectionID, candidate.SourceEntity, candidate.SourceID,
			date(candidate.PaymentDate), date(candidate.ExpectedReceiptDate), gross, fee, net,
			string(candidate.Status), candidate.PaymentMethod, installmentCount, rawID,
		); err != nil {
			return summary, fmt.Errorf("upsert future receivable %s: %w", candidate.SourceID, err)
		}
		if exists {
			summary.RecordsUpdated++
		} else {
			summary.RecordsCreated++
		}
	}
	return summary, nil
}

const upsertFutureReceivableSQL = `
INSERT INTO receivables (
  connection_id, source_system, source_entity, source_id, business_type,
  issue_date, expected_receipt_date, gross_amount, fee_amount, net_amount,
  currency_code, status, payment_method, installment_number,
  installment_count, raw_record_id
) VALUES ($1, 'NUVEM_PAGO', $2, $3, 'B2C', $4, $5, $6, $7, $8,
          'BRL', $9, $10, NULL, $11, $12)
ON CONFLICT (connection_id, source_system, source_entity, source_id)
DO UPDATE SET
  business_type = EXCLUDED.business_type,
  issue_date = EXCLUDED.issue_date,
  expected_receipt_date = EXCLUDED.expected_receipt_date,
  gross_amount = EXCLUDED.gross_amount,
  fee_amount = EXCLUDED.fee_amount,
  net_amount = EXCLUDED.net_amount,
  status = EXCLUDED.status,
  payment_method = EXCLUDED.payment_method,
  installment_number = EXCLUDED.installment_number,
  installment_count = EXCLUDED.installment_count,
  raw_record_id = EXCLUDED.raw_record_id,
  updated_at = clock_timestamp()
`

func futureReceivableExists(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, sourceID string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (
  SELECT 1 FROM receivables
  WHERE connection_id = $1 AND source_system = 'NUVEM_PAGO'
    AND source_entity = $2 AND source_id = $3
)`, connectionID, FutureSourceEntity, sourceID).Scan(&exists)
	return exists, err
}

type futureRawPayload struct {
	CustomerName          string `json:"customer_name"`
	PaymentMethod         string `json:"payment_method"`
	Brand                 string `json:"brand"`
	PaymentDate           string `json:"payment_date"`
	ExpectedReceiptDate   string `json:"expected_receipt_date"`
	InstallmentCount      *int   `json:"installment_count,omitempty"`
	GrossAmount           string `json:"gross_amount"`
	FeeAmount             string `json:"fee_amount"`
	FeeAmountSigned       string `json:"fee_amount_signed"`
	InterestAmount        string `json:"interest_amount"`
	InterestAmountSigned  string `json:"interest_amount_signed"`
	TotalCostAmount       string `json:"total_cost_amount"`
	TotalCostAmountSigned string `json:"total_cost_amount_signed"`
	NetAmount             string `json:"net_amount"`
	Status                string `json:"status"`
}

func ensureCurrentFutureRaw(ctx context.Context, queries *database.Queries, connectionID, syncBatchID pgtype.UUID, candidate FutureReceivableCandidate) (pgtype.UUID, error) {
	payload, err := json.Marshal(futureRawPayload{
		CustomerName: candidate.CustomerName, PaymentMethod: candidate.PaymentMethod, Brand: candidate.Brand,
		PaymentDate: candidate.PaymentDate.Format("2006-01-02"), ExpectedReceiptDate: candidate.ExpectedReceiptDate.Format("2006-01-02"),
		InstallmentCount: candidate.InstallmentCount, GrossAmount: candidate.GrossAmount.String(),
		FeeAmount: candidate.FeeAmount.String(), FeeAmountSigned: candidate.FeeAmountSigned.String(),
		InterestAmount: candidate.InterestAmount.String(), InterestAmountSigned: candidate.InterestAmountSigned.String(),
		TotalCostAmount: candidate.TotalCostAmount.String(), TotalCostAmountSigned: candidate.TotalCostAmountSigned.String(),
		NetAmount: candidate.NetAmount.String(), Status: string(candidate.Status),
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	current, err := queries.GetCurrentRawRecord(ctx, database.GetCurrentRawRecordParams{
		ConnectionID: connectionID, SourceSystem: string(domain.OriginNuvemPago), SourceEntity: candidate.SourceEntity, SourceID: candidate.SourceID,
	})
	if err == nil {
		if current.PayloadHash == hash {
			return current.ID, nil
		}
		if err := queries.RetireCurrentRawRecord(ctx, current.ID); err != nil {
			return pgtype.UUID{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}
	inserted, err := queries.InsertRawRecord(ctx, database.InsertRawRecordParams{
		SyncBatchID: syncBatchID, ConnectionID: connectionID, SourceSystem: string(domain.OriginNuvemPago),
		SourceEntity: candidate.SourceEntity, SourceID: candidate.SourceID, PayloadHash: hash, Payload: payload,
		SourceUpdatedAt: pgtype.Timestamptz{}, IsCurrent: true,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	if !inserted.IsCurrent {
		return pgtype.UUID{}, ErrFutureRawHistoricalVersion
	}
	return inserted.ID, nil
}

func date(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: !value.IsZero()}
}

func futureNumeric(value domain.Money) (pgtype.Numeric, error) {
	var result pgtype.Numeric
	if err := result.Scan(value.String()); err != nil {
		return pgtype.Numeric{}, err
	}
	return result, nil
}
