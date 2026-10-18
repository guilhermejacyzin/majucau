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
)

const ReceivablesAPIResource = "contas.receber"

// A full 1,000-page read may need one additional request to observe the
// empty-page response used as the provisional end-of-list sentinel.
const maxBlingAPIPageReads = 1001

const staleRawPageBatchSize = 256

var (
	ErrBlingAPISyncDatabaseUnavailable = errors.New("bling api sync database is unavailable")
	ErrBlingAPISyncClientRequired      = errors.New("bling api sync client is required")
	ErrBlingAPISyncConnectionMissing   = errors.New("bling api sync connection is missing")
	ErrBlingAPISyncPageLimit           = errors.New("bling api sync reached the page safety limit")
	ErrBlingAPIPaginationUnknown       = errors.New("bling API response does not identify whether another page exists")
)

type APISyncResult struct {
	BatchID      string
	PagesRead    int
	RecordsRead  int
	PagesCreated int
	PagesUpdated int
	Status       string
}

// ReceivablesAPISyncService stores API pages as append-only RAW evidence. It
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
	result := APISyncResult{BatchID: batch.ID.String(), Status: "RUNNING"}
	if _, err := tx.Exec(ctx, "SAVEPOINT bling_api_sync_data"); err != nil {
		return result, fmt.Errorf("create Bling API sync savepoint: %w", err)
	}
	fail := func(code string, cause error) (APISyncResult, error) {
		result.Status = "FAILED"
		return result, s.finishFailed(ctx, tx, queries, batch.ID, result, code, cause)
	}
	current := filter
	if current.Page == 0 {
		current.Page = 1
	}
	completed := false
	finalPage := 0
	for result.PagesRead < s.maxPages {
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
		created, updated, err := persistResourcePage(ctx, queries, connection.ID, batch.ID, page.Page, s.sourceEntity, payload)
		if err != nil {
			return fail("BLING_API_RAW_FAILED", err)
		}
		result.PagesRead++
		result.RecordsRead += len(page.Records)
		if created {
			result.PagesCreated++
		}
		if updated {
			result.PagesUpdated++
		}
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
	if err := retireCurrentResourcePagesAfter(ctx, tx, connection.ID, s.sourceEntity, finalPage); err != nil {
		return fail("BLING_API_RAW_RETIRE_FAILED", err)
	}
	cursorValue := fmt.Sprintf("page:%d", finalPage)
	watermark := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if err := queries.UpsertSyncCursor(ctx, database.UpsertSyncCursorParams{ConnectionID: connection.ID, Resource: s.resource, CursorValue: &cursorValue, WatermarkAt: watermark}); err != nil {
		return fail("BLING_SYNC_CURSOR_FAILED", err)
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: "SUCCESS", RecordsRead: int64(result.RecordsRead), RecordsCreated: int64(result.PagesCreated), RecordsUpdated: int64(result.PagesUpdated),
	}); err != nil {
		return fail("BLING_SYNC_BATCH_FINISH_FAILED", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit Bling API sync: %w", err)
	}
	result.Status = "SUCCESS"
	return result, nil
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
		ErrorCode: &code, ErrorMessageSanitized: &message,
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
	if !connectionID.Valid || !batchID.Valid || page < 1 || len(payload) == 0 {
		return false, false, errors.New("invalid Bling API RAW page identity")
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
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
