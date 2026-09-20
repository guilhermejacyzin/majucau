package bling

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
)

var (
	ErrReceiptDatabaseUnavailable = errors.New("receipt database is unavailable")
	ErrBlingConnectionMissing     = errors.New("bling integration connection is missing")
)

type ReceiptImportResult struct {
	BatchID        string
	Status         string
	RecordsRead    int
	RecordsCreated int
	RecordsUpdated int
	RecordsFailed  int
	IgnoredCount   int
}

// ReceiptImportService owns one PostgreSQL pool supplied by the worker. It
// deliberately has no UI or credential concerns; secrets stay outside this
// path and the transaction owns the complete RAW/ledger write.
type ReceiptImportService struct {
	pool *pgxpool.Pool
}

func NewReceiptImportService(pool *pgxpool.Pool) *ReceiptImportService {
	return &ReceiptImportService{pool: pool}
}

func (s *ReceiptImportService) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *ReceiptImportService) Import(ctx context.Context, folder string) (ReceiptImportResult, error) {
	if s == nil || s.pool == nil {
		return ReceiptImportResult{}, ErrReceiptDatabaseUnavailable
	}
	report, err := ImportReceiptsFolder(ctx, folder)
	if err != nil {
		return ReceiptImportResult{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ReceiptImportResult{}, fmt.Errorf("begin receipt import: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := database.New(tx)
	connection, err := queries.GetIntegrationConnectionByProvider(ctx, "BLING")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ReceiptImportResult{}, ErrBlingConnectionMissing
		}
		return ReceiptImportResult{}, fmt.Errorf("load Bling connection: %w", err)
	}
	if !connection.ID.Valid {
		return ReceiptImportResult{}, ErrBlingConnectionMissing
	}
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: ReceiptsReportSourceEntity})
	if err != nil {
		return ReceiptImportResult{}, fmt.Errorf("create receipt import batch: %w", err)
	}
	receiptsReport := ReceiptsReport{Receipts: report.Receipts, Errors: make([]RowError, 0, len(report.Errors))}
	for _, issue := range report.Errors {
		receiptsReport.Errors = append(receiptsReport.Errors, RowError{Line: issue.Line, Code: issue.Code, Message: issue.Message})
	}
	summary, err := PersistReceipts(ctx, tx, connection.ID, batch.ID, receiptsReport)
	if err != nil {
		return ReceiptImportResult{}, err
	}
	status := "SUCCESS"
	if summary.RecordsFailed > 0 {
		status = "PARTIAL"
	}
	var errorCode, errorMessage *string
	if summary.RecordsFailed > 0 {
		code := "BLING_IMPORT_ROW_ERRORS"
		message := "Algumas linhas não foram importadas; consulte os erros do lote."
		errorCode, errorMessage = &code, &message
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: status, RecordsRead: int32(summary.RecordsRead), RecordsCreated: int32(summary.RecordsCreated),
		RecordsUpdated: int32(summary.RecordsUpdated), RecordsFailed: int32(summary.RecordsFailed), ErrorCode: errorCode, ErrorMessageSanitized: errorMessage,
	}); err != nil {
		return ReceiptImportResult{}, fmt.Errorf("finish receipt import batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return ReceiptImportResult{}, fmt.Errorf("commit receipt import: %w", err)
	}
	return ReceiptImportResult{
		BatchID: batchIDString(batch.ID), Status: status, RecordsRead: summary.RecordsRead, RecordsCreated: summary.RecordsCreated,
		RecordsUpdated: summary.RecordsUpdated, RecordsFailed: summary.RecordsFailed, IgnoredCount: len(report.Ignored),
	}, nil
}

func batchIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}
