package nuvempago

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/dbtx"
	"majucau.local/financial-intelligence/internal/integrations/csvstream"
)

var (
	ErrFutureDatabaseUnavailable  = errors.New("future receivable database is unavailable")
	ErrNuvemPagoConnectionMissing = errors.New("nuvem pago integration connection is missing")
	ErrFuturePendingInvalid       = errors.New("future import pending source summary is invalid")
)

type FutureImportResult struct {
	BatchID        string
	Status         string
	RecordsRead    int
	RecordsCreated int
	RecordsUpdated int
	RecordsFailed  int
	IgnoredCount   int
	PendingFiles   int64
	StaleFiles     int64
	SourceMissing  bool
}

// FutureImportPending describes files deliberately left for a later run.
// It contains counts and fixed categories only; file names and source data
// never enter the batch audit message.
type FutureImportPending struct {
	NoCompatibleCSV      bool
	DeferredFiles        int64
	StaleFiles           int64
	UnsupportedFiles     int64
	DownloadFailedFiles  int64
	InvalidMetadataFiles int64
}

func (p FutureImportPending) hasPending() bool {
	return p.NoCompatibleCSV || p.DeferredFiles > 0 || p.StaleFiles > 0 || p.UnsupportedFiles > 0 || p.DownloadFailedFiles > 0 || p.InvalidMetadataFiles > 0
}

func (p FutureImportPending) pendingFileCount() (int64, bool) {
	if p.DeferredFiles < 0 || p.UnsupportedFiles < 0 || p.DownloadFailedFiles < 0 || p.InvalidMetadataFiles < 0 {
		return 0, false
	}
	total := int64(0)
	for _, count := range []int64{p.DeferredFiles, p.StaleFiles, p.UnsupportedFiles, p.DownloadFailedFiles, p.InvalidMetadataFiles} {
		if count > int64(^uint64(0)>>1)-total {
			return 0, false
		}
		total += count
	}
	return total, true
}

func (p FutureImportPending) auditDetails() (string, string, int64, error) {
	files, valid := p.pendingFileCount()
	if !valid {
		return "", "", 0, ErrFuturePendingInvalid
	}
	if !p.hasPending() {
		return "", "", 0, nil
	}
	parts := make([]string, 0, 5)
	if p.NoCompatibleCSV {
		parts = append(parts, "nenhum CSV compatível disponível")
	}
	if p.DeferredFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d arquivo(s) aguardam a próxima atualização", p.DeferredFiles))
	}
	if p.StaleFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d arquivo(s) estão antigos para a atualização diária", p.StaleFiles))
	}
	if p.UnsupportedFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d arquivo(s) usam formato ainda não suportado", p.UnsupportedFiles))
	}
	if p.DownloadFailedFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d arquivo(s) não puderam ser baixados nesta execução", p.DownloadFailedFiles))
	}
	if p.InvalidMetadataFiles > 0 {
		parts = append(parts, fmt.Sprintf("%d arquivo(s) têm informações incompletas", p.InvalidMetadataFiles))
	}
	code := "NUVEM_PAGO_FUTURE_SOURCE_PENDING"
	message := "Importação parcial; " + strings.Join(parts, "; ") + "."
	return code, message, files, nil
}

type FutureImportService struct{ pool *pgxpool.Pool }

func NewFutureImportService(pool *pgxpool.Pool) *FutureImportService {
	return &FutureImportService{pool: pool}
}

func (s *FutureImportService) Import(ctx context.Context, folder string) (FutureImportResult, error) {
	return s.ImportWithPending(ctx, folder, FutureImportPending{})
}

// ImportWithPending imports available CSV files in one transaction and audits
// unsupported or deferred source files as PARTIAL without counting files as
// invalid data rows.
func (s *FutureImportService) ImportWithPending(ctx context.Context, folder string, pending FutureImportPending) (FutureImportResult, error) {
	return s.importWithPending(ctx, folder, pending, nil)
}

// ImportStaged reads a complete secure staging folder in one transaction and
// invokes cleanup after the last CSV byte has been read but before committing
// financial rows. A cleanup failure therefore rolls the database transaction
// back instead of reporting a successful import with sensitive files left on disk.
func (s *FutureImportService) ImportStaged(ctx context.Context, folder string, pending FutureImportPending, cleanupBeforeCommit func() error) (FutureImportResult, error) {
	if cleanupBeforeCommit == nil {
		return FutureImportResult{}, ErrFuturePendingInvalid
	}
	return s.importWithPending(ctx, folder, pending, cleanupBeforeCommit)
}

func (s *FutureImportService) importWithPending(ctx context.Context, folder string, pending FutureImportPending, cleanupBeforeCommit func() error) (out FutureImportResult, returnErr error) {
	if s == nil || s.pool == nil {
		return FutureImportResult{}, ErrFutureDatabaseUnavailable
	}
	if ctx == nil {
		return FutureImportResult{}, ErrFuturePendingInvalid
	}
	pendingCode, pendingMessage, pendingFiles, err := pending.auditDetails()
	if err != nil {
		return FutureImportResult{}, err
	}
	if err := validateFutureFolder(folder); err != nil {
		return FutureImportResult{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FutureImportResult{}, fmt.Errorf("begin future import: %w", err)
	}
	defer func() {
		if err := dbtx.Rollback(ctx, tx); err != nil {
			out = FutureImportResult{}
			returnErr = errors.Join(returnErr, err)
		}
	}()
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
	seen, err := csvstream.NewSourceIDSet(ctx, tx)
	if err != nil {
		return FutureImportResult{}, err
	}
	defer seen.Close()
	var summary FuturePersistenceSummary
	streamSummary, err := StreamFutureFolder(ctx, folder, seen, func(_ string, _ int, candidate FutureReceivableCandidate) error {
		exists, err := persistFutureCandidate(ctx, tx, queries, connection.ID, batch.ID, candidate)
		if err != nil {
			return err
		}
		if exists {
			summary.RecordsUpdated++
		} else {
			summary.RecordsCreated++
		}
		return nil
	}, nil, nil)
	if err != nil {
		return FutureImportResult{}, err
	}
	summary.RecordsRead = streamSummary.ReceivableCount + streamSummary.ErrorCount
	summary.RecordsFailed = streamSummary.ErrorCount
	status := "SUCCESS"
	if summary.RecordsFailed > 0 || pending.hasPending() {
		status = "PARTIAL"
	}
	if cleanupBeforeCommit != nil {
		if err := cleanupBeforeCommit(); err != nil {
			return FutureImportResult{}, err
		}
	}
	var errorCode, errorMessage *string
	if pending.hasPending() && summary.RecordsFailed > 0 {
		code := "NUVEM_PAGO_FUTURE_PARTIAL"
		message := "Importação parcial; existem linhas rejeitadas e arquivos pendentes. Consulte os contadores do lote."
		errorCode, errorMessage = &code, &message
	} else if pending.hasPending() {
		errorCode, errorMessage = &pendingCode, &pendingMessage
	} else if summary.RecordsFailed > 0 {
		code := "NUVEM_PAGO_FUTURE_ROW_ERRORS"
		message := "Algumas linhas do extrato futuro não foram importadas; consulte os erros do lote."
		errorCode, errorMessage = &code, &message
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: status, RecordsRead: int64(summary.RecordsRead), RecordsCreated: int64(summary.RecordsCreated),
		RecordsUpdated: int64(summary.RecordsUpdated), RecordsFailed: int64(summary.RecordsFailed), ErrorCode: errorCode, ErrorMessageSanitized: errorMessage,
	}); err != nil {
		return FutureImportResult{}, fmt.Errorf("finish future import batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return FutureImportResult{}, fmt.Errorf("commit future import: %w", err)
	}
	return FutureImportResult{BatchID: futureBatchIDString(batch.ID), Status: status, RecordsRead: summary.RecordsRead, RecordsCreated: summary.RecordsCreated, RecordsUpdated: summary.RecordsUpdated, RecordsFailed: summary.RecordsFailed, IgnoredCount: streamSummary.IgnoredCount, PendingFiles: pendingFiles, StaleFiles: pending.StaleFiles, SourceMissing: pending.NoCompatibleCSV}, nil
}

// RecordPending writes a zero-row PARTIAL batch when no compatible CSV could
// be safely staged. It does not mark files as failed financial records.
func (s *FutureImportService) RecordPending(ctx context.Context, pending FutureImportPending) (out FutureImportResult, returnErr error) {
	if s == nil || s.pool == nil {
		return FutureImportResult{}, ErrFutureDatabaseUnavailable
	}
	if ctx == nil {
		return FutureImportResult{}, ErrFuturePendingInvalid
	}
	code, message, files, err := pending.auditDetails()
	if err != nil || !pending.hasPending() {
		return FutureImportResult{}, ErrFuturePendingInvalid
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return FutureImportResult{}, fmt.Errorf("begin pending future source audit: %w", err)
	}
	defer func() {
		if rollbackErr := dbtx.Rollback(ctx, tx); rollbackErr != nil {
			out = FutureImportResult{}
			returnErr = errors.Join(returnErr, rollbackErr)
		}
	}()
	queries := database.New(tx)
	connection, err := queries.GetIntegrationConnectionByProvider(ctx, "NUVEM_PAGO")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FutureImportResult{}, ErrNuvemPagoConnectionMissing
		}
		return FutureImportResult{}, fmt.Errorf("load Nuvem Pago pending source connection: %w", err)
	}
	if !connection.ID.Valid {
		return FutureImportResult{}, ErrNuvemPagoConnectionMissing
	}
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: FutureSourceEntity})
	if err != nil {
		return FutureImportResult{}, fmt.Errorf("create pending future source batch: %w", err)
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: "PARTIAL", RecordsRead: 0, RecordsCreated: 0, RecordsUpdated: 0, RecordsFailed: 0,
		ErrorCode: &code, ErrorMessageSanitized: &message,
	}); err != nil {
		return FutureImportResult{}, fmt.Errorf("finish pending future source batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return FutureImportResult{}, fmt.Errorf("commit pending future source audit: %w", err)
	}
	return FutureImportResult{BatchID: futureBatchIDString(batch.ID), Status: "PARTIAL", PendingFiles: files, StaleFiles: pending.StaleFiles, SourceMissing: pending.NoCompatibleCSV}, nil
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
