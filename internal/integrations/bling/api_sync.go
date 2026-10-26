package bling

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/dbtx"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/integrations/httpretry"
)

const ReceivablesAPIResource = "contas.receber"
const CashMovementsAPIResource = "caixas"

// Bound a full read even if the provider never returns a verifiable end marker.
const maxBlingAPIPageReads = 1001

const staleRawPageBatchSize = 256

var (
	ErrBlingAPISyncDatabaseUnavailable = errors.New("bling api sync database is unavailable")
	ErrBlingAPISyncClientRequired      = errors.New("bling api sync client is required")
	ErrBlingAPISyncConnectionMissing   = errors.New("bling api sync connection is missing")
	ErrBlingAPISyncPageLimit           = errors.New("bling api sync reached the page safety limit")
	ErrBlingAPIPaginationUnknown       = errors.New("bling API response does not identify whether another page exists")
	ErrBlingAPICheckpointInvalid       = errors.New("bling API page checkpoint is invalid")
	ErrBlingAPICheckpointConflict      = errors.New("bling API page checkpoint conflicts with a previously stored page")
)

type APISyncResult struct {
	BatchID      string
	PagesRead    int
	RecordsRead  int
	PagesCreated int
	PagesUpdated int
	RetryCount   int32
	Status       string
}

// apiResourceSyncService stores API pages as append-only RAW evidence. It
// intentionally does not normalize financial fields: source IDs and meanings
// are still subject to BK-040 confirmation from the real Bling account.
type apiResourceSyncService struct {
	pool         *pgxpool.Pool
	client       *BlingAPIClient
	maxPages     int
	resource     string
	sourceEntity string
	list         func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error)
}

type ReceivablesAPISyncService struct{ *apiResourceSyncService }

func NewReceivablesAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *ReceivablesAPISyncService {
	return &ReceivablesAPISyncService{apiResourceSyncService: &apiResourceSyncService{pool: pool, client: client, maxPages: maxBlingAPIPageReads, resource: ReceivablesAPIResource, sourceEntity: "bling.contas_receber.page", list: client.ListReceivables}}
}

// SyncWithCorrelationID runs receivables as part of a stable daily operation.
// The ID lets a restarted worker determine whether that scheduled run finished.
func (s *ReceivablesAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}

const PayablesAPIResource = "contas.pagar"

type PayablesAPISyncService struct{ *apiResourceSyncService }

func NewPayablesAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *PayablesAPISyncService {
	return &PayablesAPISyncService{apiResourceSyncService: &apiResourceSyncService{pool: pool, client: client, maxPages: maxBlingAPIPageReads, resource: PayablesAPIResource, sourceEntity: "bling.contas_pagar.page", list: client.ListPayables}}
}

// SyncWithCorrelationID runs payables as part of a stable daily operation.
// The ID lets a restarted worker determine whether that scheduled run finished.
func (s *PayablesAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}

// CashMovementsAPISyncService stores the Bling cash/bank movement pages as
// append-only RAW evidence. It does not classify a movement as CAPEX or a
// financial fact; those meanings still require account homologation.
type CashMovementsAPISyncService struct{ *apiResourceSyncService }

func NewCashMovementsAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *CashMovementsAPISyncService {
	var list func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error)
	if client != nil {
		list = func(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			return client.ListCashMovements(ctx, CashMovementsFilter{Page: filter.Page, MaxRecords: filter.Limit})
		}
	}
	return &CashMovementsAPISyncService{apiResourceSyncService: &apiResourceSyncService{pool: pool, client: client, maxPages: maxBlingAPIPageReads, resource: CashMovementsAPIResource, sourceEntity: "bling.caixas.page", list: list}}
}

// SyncWithCorrelationID runs cash movements as part of the stable daily load.
func (s *CashMovementsAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}

func (s *apiResourceSyncService) Sync(ctx context.Context, filter ReceivablesFilter) (out APISyncResult, returnErr error) {
	return s.sync(ctx, filter, nil)
}

func (s *apiResourceSyncService) sync(ctx context.Context, filter ReceivablesFilter, correlationID *pgtype.UUID) (out APISyncResult, returnErr error) {
	if s == nil || s.pool == nil {
		return APISyncResult{}, ErrBlingAPISyncDatabaseUnavailable
	}
	if s.client == nil || s.list == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	if s.maxPages <= 0 {
		return APISyncResult{}, ErrBlingAPISyncPageLimit
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return APISyncResult{}, fmt.Errorf("begin Bling API sync: %w", err)
	}
	defer func() {
		if err := dbtx.Rollback(ctx, tx); err != nil {
			out = APISyncResult{}
			returnErr = errors.Join(returnErr, err)
		}
	}()
	queries := database.New(tx)
	connection, err := queries.GetIntegrationConnectionByProvider(ctx, string(domain.OriginBling))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return APISyncResult{}, ErrBlingAPISyncConnectionMissing
		}
		return APISyncResult{}, fmt.Errorf("load Bling API connection: %w", err)
	}
	if !connection.ID.Valid {
		return APISyncResult{}, ErrBlingAPISyncConnectionMissing
	}
	// Keep manual and scheduled syncs of the same Bling resource from racing
	// while their page snapshots and current RAW records are being replaced.
	// The transaction lock is released automatically on commit or rollback.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, blingAPIResourceSyncLockKey(connection.ID, s.resource)); err != nil {
		return APISyncResult{}, fmt.Errorf("lock Bling API resource sync: %w", err)
	}
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: s.resource})
	if err != nil {
		return APISyncResult{}, fmt.Errorf("create Bling API sync batch: %w", err)
	}
	if correlationID != nil {
		if !correlationID.Valid || correlationID.Bytes == ([16]byte{}) {
			return APISyncResult{}, errors.New("invalid Bling API sync correlation ID")
		}
		if _, err := tx.Exec(ctx, `UPDATE integration_sync_batches SET correlation_id = $1 WHERE id = $2`, *correlationID, batch.ID); err != nil {
			return APISyncResult{}, fmt.Errorf("set Bling API sync correlation ID: %w", err)
		}
	}
	checkpointCorrelationID := batch.CorrelationID
	if correlationID != nil {
		checkpointCorrelationID = *correlationID
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id <> $3`,
		connection.ID, s.resource, checkpointCorrelationID); err != nil {
		return APISyncResult{}, fmt.Errorf("remove superseded Bling page checkpoints: %w", err)
	}
	result := APISyncResult{BatchID: batch.ID.String(), Status: "RUNNING"}
	retryCounter := httpretry.NewCounter()
	ctx = httpretry.WithCounter(ctx, retryCounter)
	if _, err := tx.Exec(ctx, "SAVEPOINT bling_api_sync_data"); err != nil {
		return result, fmt.Errorf("create Bling API sync savepoint: %w", err)
	}
	fail := func(code string, cause error) (APISyncResult, error) {
		result.Status = "FAILED"
		result.RetryCount = retryCounter.Count()
		return result, s.finishFailed(ctx, tx, queries, batch.ID, result, code, cause)
	}
	current := filter
	if current.Page == 0 {
		current.Page = 1
	}
	startingPage := current.Page
	checkpoint, err := loadAPIPageCheckpointSummary(ctx, tx, connection.ID, s.resource, checkpointCorrelationID, startingPage, s.maxPages)
	if err != nil {
		return fail("BLING_API_CHECKPOINT_INVALID", err)
	}
	result.PagesRead = checkpoint.PageCount
	if checkpoint.RecordCount > int64(^uint(0)>>1) {
		return fail("BLING_API_CHECKPOINT_INVALID", ErrBlingAPICheckpointInvalid)
	}
	result.RecordsRead = int(checkpoint.RecordCount)
	completed := checkpoint.PageCount > 0 && !checkpoint.LastHasNext
	finalPage := checkpoint.LastPage
	if checkpoint.PageCount > 0 {
		current.Page = checkpoint.LastPage + 1
		current.Limit = checkpoint.LastLimit
	}
	for result.PagesRead < s.maxPages {
		if completed {
			break
		}
		if err := ctx.Err(); err != nil {
			return fail("SYNC_CANCELLED", err)
		}
		page, err := s.list(ctx, current)
		if err != nil {
			return fail("BLING_API_READ_FAILED", err)
		}
		if page.Page != current.Page {
			return fail("BLING_API_SCHEMA_MISMATCH", fmt.Errorf("%w: response page does not match the requested page", ErrBlingAPISchemaMismatch))
		}
		if !page.HasNextKnown {
			return fail("BLING_API_PAGINATION_UNKNOWN", ErrBlingAPIPaginationUnknown)
		}
		payload, err := marshalReceivablesPage(page)
		if err != nil {
			return fail("BLING_API_PAYLOAD_INVALID", err)
		}
		if err := persistAPIPageCheckpoint(ctx, s.pool, connection.ID, s.resource, checkpointCorrelationID, page, payload); err != nil {
			return fail("BLING_API_CHECKPOINT_FAILED", err)
		}
		result.PagesRead++
		result.RecordsRead += len(page.Records)
		if !page.HasNext {
			completed = true
			finalPage = page.Page
			break
		}
		current.Page = page.Page + 1
		current.Limit = page.Limit
	}
	if !completed {
		return fail("BLING_API_PAGE_LIMIT", ErrBlingAPISyncPageLimit)
	}
	for pageNumber := startingPage; pageNumber <= finalPage; pageNumber++ {
		var payload []byte
		var payloadHash string
		var hasNext bool
		if err := tx.QueryRow(ctx, `
			SELECT payload, payload_hash, has_next
			FROM bling_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`,
			connection.ID, s.resource, checkpointCorrelationID, pageNumber,
		).Scan(&payload, &payloadHash, &hasNext); err != nil {
			return fail("BLING_API_CHECKPOINT_INVALID", fmt.Errorf("%w: missing staged page", ErrBlingAPICheckpointInvalid))
		}
		if (pageNumber < finalPage && !hasNext) || (pageNumber == finalPage && hasNext) {
			return fail("BLING_API_CHECKPOINT_INVALID", ErrBlingAPICheckpointInvalid)
		}
		created, updated, err := persistResourcePageWithHash(ctx, queries, connection.ID, batch.ID, pageNumber, s.sourceEntity, payload, payloadHash)
		if err != nil {
			return fail("BLING_API_RAW_FAILED", err)
		}
		if created {
			result.PagesCreated++
		}
		if updated {
			result.PagesUpdated++
		}
	}
	if err := retireCurrentResourcePagesAfter(ctx, tx, connection.ID, s.sourceEntity, finalPage); err != nil {
		return fail("BLING_API_RAW_RETIRE_FAILED", err)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`,
		connection.ID, s.resource, checkpointCorrelationID); err != nil {
		return fail("BLING_API_CHECKPOINT_CLEANUP_FAILED", err)
	}
	cursorValue := fmt.Sprintf("page:%d", finalPage)
	watermark := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if err := queries.UpsertSyncCursor(ctx, database.UpsertSyncCursorParams{ConnectionID: connection.ID, Resource: s.resource, CursorValue: &cursorValue, WatermarkAt: watermark}); err != nil {
		return fail("BLING_SYNC_CURSOR_FAILED", err)
	}
	result.RetryCount = retryCounter.Count()
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: "SUCCESS", RecordsRead: int64(result.RecordsRead), RecordsCreated: int64(result.PagesCreated), RecordsUpdated: int64(result.PagesUpdated), RetryCount: result.RetryCount,
	}); err != nil {
		return fail("BLING_SYNC_BATCH_FINISH_FAILED", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit Bling API sync: %w", err)
	}
	result.Status = "SUCCESS"
	return result, nil
}

type apiPageCheckpointSummary struct {
	PageCount   int
	FirstPage   int
	LastPage    int
	RecordCount int64
	LastLimit   int
	LastHasNext bool
}

func loadAPIPageCheckpointSummary(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, startingPage, maxPages int) (apiPageCheckpointSummary, error) {
	if !connectionID.Valid || !correlationID.Valid || resource == "" || startingPage < 1 || maxPages < 1 {
		return apiPageCheckpointSummary{}, ErrBlingAPICheckpointInvalid
	}
	var summary apiPageCheckpointSummary
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::integer, COALESCE(MIN(page_number), 0), COALESCE(MAX(page_number), 0),
		       COALESCE(SUM(record_count), 0)::bigint
		FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3`,
		connectionID, resource, correlationID,
	).Scan(&summary.PageCount, &summary.FirstPage, &summary.LastPage, &summary.RecordCount); err != nil {
		return apiPageCheckpointSummary{}, fmt.Errorf("read Bling page checkpoint summary: %w", err)
	}
	if summary.PageCount == 0 {
		return summary, nil
	}
	if summary.FirstPage != startingPage || summary.LastPage-summary.FirstPage+1 != summary.PageCount || summary.PageCount > maxPages || summary.RecordCount < 0 {
		return apiPageCheckpointSummary{}, ErrBlingAPICheckpointInvalid
	}
	if err := tx.QueryRow(ctx, `
		SELECT page_limit, has_next
		FROM bling_sync_page_checkpoints
		WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`,
		connectionID, resource, correlationID, summary.LastPage,
	).Scan(&summary.LastLimit, &summary.LastHasNext); err != nil {
		return apiPageCheckpointSummary{}, fmt.Errorf("read last Bling page checkpoint: %w", err)
	}
	return summary, nil
}

func persistAPIPageCheckpoint(ctx context.Context, pool *pgxpool.Pool, connectionID pgtype.UUID, resource string, correlationID pgtype.UUID, page BlingReceivablesPage, payload []byte) error {
	if pool == nil || !connectionID.Valid || !correlationID.Valid || resource == "" || page.Page < 1 || page.Limit < 0 || !page.HasNextKnown || len(payload) == 0 {
		return ErrBlingAPICheckpointInvalid
	}
	digest := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(digest[:])
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin Bling page checkpoint: %w", err)
	}
	defer func() { _ = dbtx.Rollback(ctx, tx) }()
	tag, err := tx.Exec(ctx, `
		INSERT INTO bling_sync_page_checkpoints (
			connection_id, resource, correlation_id, page_number, page_limit,
			has_next, record_count, payload_hash, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
		ON CONFLICT (connection_id, resource, correlation_id, page_number) DO NOTHING`,
		connectionID, resource, correlationID, page.Page, page.Limit, page.HasNext, len(page.Records), payloadHash, payload,
	)
	if err != nil {
		return fmt.Errorf("store Bling page checkpoint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		var existingHash string
		var existingLimit int
		var existingHasNext bool
		var existingRecordCount int
		if err := tx.QueryRow(ctx, `
			SELECT payload_hash, page_limit, has_next, record_count
			FROM bling_sync_page_checkpoints
			WHERE connection_id = $1 AND resource = $2 AND correlation_id = $3 AND page_number = $4`,
			connectionID, resource, correlationID, page.Page,
		).Scan(&existingHash, &existingLimit, &existingHasNext, &existingRecordCount); err != nil {
			return fmt.Errorf("verify existing Bling page checkpoint: %w", err)
		}
		if existingHash != payloadHash || existingLimit != page.Limit || existingHasNext != page.HasNext || existingRecordCount != len(page.Records) {
			return ErrBlingAPICheckpointConflict
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit Bling page checkpoint: %w", err)
	}
	return nil
}

func blingAPIResourceSyncLockKey(connectionID pgtype.UUID, resource string) int64 {
	hash := sha256.New()
	_, _ = hash.Write([]byte("majucau:bling-api-resource-sync\x00"))
	_, _ = hash.Write(connectionID.Bytes[:])
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(resource))
	return int64(binary.BigEndian.Uint64(hash.Sum(nil)[:8]))
}

// retireCurrentResourcePagesAfter removes stale page snapshots only after a
// complete read. The database procedure keeps RAW history append-only while
// allowing the current view to reflect a list that became shorter.
func retireCurrentResourcePagesAfter(ctx context.Context, tx pgx.Tx, connectionID pgtype.UUID, sourceEntity string, finalPage int) error {
	if !connectionID.Valid || sourceEntity == "" || finalPage < 1 {
		return errors.New("invalid Bling page snapshot identity")
	}
	store := staleRawPageStore{tx: tx, queries: database.New(tx)}
	return retireCurrentResourcePagesInBatches(ctx, store, connectionID, sourceEntity, finalPage)
}

type staleRawPageBatchStore interface {
	ListCurrentStalePageIDs(context.Context, pgtype.UUID, string, int, pgtype.UUID, int) ([]pgtype.UUID, error)
	RetireCurrentRawPage(context.Context, pgtype.UUID) error
}

type staleRawPageStore struct {
	tx      pgx.Tx
	queries *database.Queries
}

func (s staleRawPageStore) ListCurrentStalePageIDs(ctx context.Context, connectionID pgtype.UUID, sourceEntity string, finalPage int, afterID pgtype.UUID, limit int) ([]pgtype.UUID, error) {
	rows, err := s.tx.Query(ctx, `
		SELECT id
		FROM raw_records
		WHERE connection_id = $1
		  AND source_system = 'BLING'
		  AND source_entity = $2
		  AND is_current
		  AND source_id ~ '^pagina:[0-9]+$'
		  AND substring(source_id FROM 8)::numeric > $3
		  AND ($4::uuid IS NULL OR id > $4)
		ORDER BY id
		LIMIT $5`, connectionID, sourceEntity, finalPage, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("find stale Bling RAW pages: %w", err)
	}
	defer rows.Close()

	ids := make([]pgtype.UUID, 0, limit)
	for rows.Next() {
		var id pgtype.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("read stale Bling RAW page ID: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stale Bling RAW pages: %w", err)
	}
	return ids, nil
}

func (s staleRawPageStore) RetireCurrentRawPage(ctx context.Context, id pgtype.UUID) error {
	return s.queries.RetireCurrentRawRecord(ctx, id)
}

func retireCurrentResourcePagesInBatches(ctx context.Context, store staleRawPageBatchStore, connectionID pgtype.UUID, sourceEntity string, finalPage int) error {
	if store == nil || !connectionID.Valid || sourceEntity == "" || finalPage < 1 {
		return errors.New("invalid Bling page retirement request")
	}
	cursor := pgtype.UUID{}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		ids, err := store.ListCurrentStalePageIDs(ctx, connectionID, sourceEntity, finalPage, cursor, staleRawPageBatchSize)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if len(ids) > staleRawPageBatchSize {
			return errors.New("stale Bling page query exceeded its batch limit")
		}
		for _, id := range ids {
			if err := ctx.Err(); err != nil {
				return err
			}
			if !id.Valid || (cursor.Valid && bytes.Compare(id.Bytes[:], cursor.Bytes[:]) <= 0) {
				return errors.New("stale Bling page query returned an invalid cursor order")
			}
			if err := store.RetireCurrentRawPage(ctx, id); err != nil {
				return fmt.Errorf("retire stale Bling RAW page: %w", err)
			}
			cursor = id
		}
		if len(ids) < staleRawPageBatchSize {
			return nil
		}
	}
}

func (s *apiResourceSyncService) finishFailed(ctx context.Context, tx pgx.Tx, queries *database.Queries, batchID pgtype.UUID, result APISyncResult, code string, cause error) error {
	// Keep the failure audit while rolling back all RAW and cursor mutations from
	// the attempted sync. A detached, bounded context also lets cancellation be
	// recorded after the caller's context has ended.
	auditCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if _, err := tx.Exec(auditCtx, "ROLLBACK TO SAVEPOINT bling_api_sync_data"); err != nil {
		return fmt.Errorf("%w: rollback failed Bling sync data: %v", cause, err)
	}
	if _, err := tx.Exec(auditCtx, "RELEASE SAVEPOINT bling_api_sync_data"); err != nil {
		return fmt.Errorf("%w: release failed Bling sync savepoint: %v", cause, err)
	}
	message := code
	if err := queries.FinishIntegrationSyncBatch(auditCtx, database.FinishIntegrationSyncBatchParams{
		ID: batchID, Status: "FAILED", RecordsRead: int64(result.RecordsRead), RecordsCreated: int64(result.PagesCreated), RecordsUpdated: int64(result.PagesUpdated),
		RetryCount: result.RetryCount, ErrorCode: &code, ErrorMessageSanitized: &message,
	}); err != nil {
		return fmt.Errorf("%w: finish failed Bling sync batch: %v", cause, err)
	}
	if err := tx.Commit(auditCtx); err != nil {
		return fmt.Errorf("%w: commit failed Bling sync audit: %v", cause, err)
	}
	return cause
}

type receivablesPagePayload struct {
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	Total   *int              `json:"total,omitempty"`
	HasNext bool              `json:"has_next"`
	Records []json.RawMessage `json:"records"`
}

func marshalReceivablesPage(page BlingReceivablesPage) ([]byte, error) {
	return json.Marshal(receivablesPagePayload{Page: page.Page, Limit: page.Limit, Total: page.Total, HasNext: page.HasNext, Records: page.Records})
}

func persistReceivablesPage(ctx context.Context, q *database.Queries, connectionID, batchID pgtype.UUID, page int, payload []byte) (created, updated bool, err error) {
	return persistResourcePage(ctx, q, connectionID, batchID, page, "bling.contas_receber.page", payload)
}

func persistResourcePage(ctx context.Context, q *database.Queries, connectionID, batchID pgtype.UUID, page int, sourceEntity string, payload []byte) (created, updated bool, err error) {
	digest := sha256.Sum256(payload)
	return persistResourcePageWithHash(ctx, q, connectionID, batchID, page, sourceEntity, payload, hex.EncodeToString(digest[:]))
}

func persistResourcePageWithHash(ctx context.Context, q *database.Queries, connectionID, batchID pgtype.UUID, page int, sourceEntity string, payload []byte, hash string) (created, updated bool, err error) {
	if !connectionID.Valid || !batchID.Valid || page < 1 || len(payload) == 0 {
		return false, false, errors.New("invalid Bling API RAW page identity")
	}
	if len(hash) != sha256.Size*2 {
		return false, false, errors.New("invalid Bling API RAW page hash")
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return false, false, errors.New("invalid Bling API RAW page hash")
	}
	sourceID := receivablesPageSourceID(page)
	current, currentErr := q.GetCurrentRawRecord(ctx, database.GetCurrentRawRecordParams{ConnectionID: connectionID, SourceSystem: string(domain.OriginBling), SourceEntity: sourceEntity, SourceID: sourceID})
	if currentErr == nil && current.PayloadHash == hash {
		return false, false, nil
	}
	if currentErr != nil && !errors.Is(currentErr, pgx.ErrNoRows) {
		return false, false, currentErr
	}
	if currentErr == nil {
		if err := q.RetireCurrentRawRecord(ctx, current.ID); err != nil {
			return false, false, err
		}
		updated = true
	} else {
		created = true
	}
	inserted, err := q.InsertRawRecord(ctx, database.InsertRawRecordParams{
		SyncBatchID: batchID, ConnectionID: connectionID, SourceSystem: string(domain.OriginBling), SourceEntity: sourceEntity, SourceID: sourceID, PayloadHash: hash, Payload: payload, IsCurrent: true,
	})
	if err != nil {
		return false, false, err
	}
	if !inserted.IsCurrent {
		return false, false, ErrHistoricalRawVersion
	}
	return created, updated, nil
}

func receivablesPageSourceID(page int) string {
	return fmt.Sprintf("pagina:%d", page)
}
