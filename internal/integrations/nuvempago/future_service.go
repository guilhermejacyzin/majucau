package nuvempago

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
	ErrFutureDatabaseUnavailable  = errors.New("future receivable database is unavailable")
	ErrNuvemPagoConnectionMissing = errors.New("nuvem pago integration connection is missing")
)

type FutureImportResult struct {
	BatchID        string
	Status         string
	RecordsRead    int
	RecordsCreated int
	RecordsUpdated int
	RecordsFailed  int
	IgnoredCount   int
}

type FutureImportService struct{ pool *pgxpool.Pool }

func NewFutureImportService(pool *pgxpool.Pool) *FutureImportService {
	return &FutureImportService{pool: pool}
}

func (s *FutureImportService) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *FutureImportService) Import(ctx context.Context, folder string) (FutureImportResult, error) {
	if s == nil || s.pool == nil {
		return FutureImportResult{}, ErrFutureDatabaseUnavailable
	}
	report, err := ImportFutureFolder(ctx, folder)
	if err != nil {
		return FutureImportResult{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FutureImportResult{}, fmt.Errorf("begin future import: %w", err)
	}
	defer tx.Rollback(ctx)
	queries := database.New(tx)
	connection, err := queries.GetIntegrationConnectionByProvider(ctx, "NUVEM_PAGO")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FutureImportResult{}, ErrNuvemPagoConnectionMissing
		}
		return FutureImportResult{}, fmt.Errorf("load Nuvem Pago connection: %w", err)
	}
	if !connection.ID.Valid {
		return FutureImportResult{}, ErrNuvemPagoConnectionMissing
	}
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: FutureSourceEntity})
	if err != nil {
		return FutureImportResult{}, fmt.Errorf("create future import batch: %w", err)
	}
	summary, err := PersistFutureReceivables(ctx, tx, connection.ID, batch.ID, FutureReport{Receivables: report.Receivables, Errors: toRowErrors(report.Errors)})
	if err != nil {
		return FutureImportResult{}, err
	}
	status := "SUCCESS"
	if summary.RecordsFailed > 0 {
		status = "PARTIAL"
	}
	var errorCode, errorMessage *string
	if summary.RecordsFailed > 0 {
		code := "NUVEM_PAGO_FUTURE_ROW_ERRORS"
		message := "Algumas linhas do extrato futuro não foram importadas; consulte os erros do lote."
		errorCode, errorMessage = &code, &message
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: status, RecordsRead: int32(summary.RecordsRead), RecordsCreated: int32(summary.RecordsCreated),
		RecordsUpdated: int32(summary.RecordsUpdated), RecordsFailed: int32(summary.RecordsFailed), ErrorCode: errorCode, ErrorMessageSanitized: errorMessage,
	}); err != nil {
		return FutureImportResult{}, fmt.Errorf("finish future import batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return FutureImportResult{}, fmt.Errorf("commit future import: %w", err)
	}
	return FutureImportResult{BatchID: futureBatchIDString(batch.ID), Status: status, RecordsRead: summary.RecordsRead, RecordsCreated: summary.RecordsCreated, RecordsUpdated: summary.RecordsUpdated, RecordsFailed: summary.RecordsFailed, IgnoredCount: len(report.Ignored)}, nil
}

func toRowErrors(input []FutureFileRowError) []RowError {
	result := make([]RowError, 0, len(input))
	for _, issue := range input {
		result = append(result, RowError{Line: issue.Line, Code: issue.Code, Message: issue.Message})
	}
	return result
}

func futureBatchIDString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return ""
	}
	return parsed.String()
}
