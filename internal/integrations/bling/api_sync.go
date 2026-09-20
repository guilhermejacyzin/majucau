package bling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/domain"
)

const ReceivablesAPIResource = "contas.receber"

var (
	ErrBlingAPISyncDatabaseUnavailable = errors.New("bling api sync database is unavailable")
	ErrBlingAPISyncClientRequired      = errors.New("bling api sync client is required")
	ErrBlingAPISyncConnectionMissing   = errors.New("bling api sync connection is missing")
	ErrBlingAPISyncPageLimit           = errors.New("bling api sync reached the page safety limit")
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
	return &ReceivablesAPISyncService{apiResourceSyncService: &apiResourceSyncService{pool: pool, client: client, maxPages: 1000, resource: ReceivablesAPIResource, sourceEntity: "bling.contas_receber.page", list: client.ListReceivables}}
}

const PayablesAPIResource = "contas.pagar"

type PayablesAPISyncService struct{ *apiResourceSyncService }

func NewPayablesAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *PayablesAPISyncService {
	return &PayablesAPISyncService{apiResourceSyncService: &apiResourceSyncService{pool: pool, client: client, maxPages: 1000, resource: PayablesAPIResource, sourceEntity: "bling.contas_pagar.page", list: client.ListPayables}}
}

func (s *apiResourceSyncService) Sync(ctx context.Context, filter ReceivablesFilter) (APISyncResult, error) {
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
	defer tx.Rollback(ctx)
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
	batch, err := queries.InsertIntegrationSyncBatch(ctx, database.InsertIntegrationSyncBatchParams{ConnectionID: connection.ID, Resource: s.resource})
	if err != nil {
		return APISyncResult{}, fmt.Errorf("create Bling API sync batch: %w", err)
	}
	result := APISyncResult{BatchID: batch.ID.String(), Status: "RUNNING"}
	current := filter
	if current.Page == 0 {
		current.Page = 1
	}
	completed := false
	for result.PagesRead < s.maxPages {
		if err := ctx.Err(); err != nil {
			return result, s.finishFailed(ctx, queries, batch.ID, result, "SYNC_CANCELLED", err)
		}
		page, err := s.list(ctx, current)
		if err != nil {
			return result, s.finishFailed(ctx, queries, batch.ID, result, "BLING_API_READ_FAILED", err)
		}
		payload, err := marshalReceivablesPage(page)
		if err != nil {
			return result, s.finishFailed(ctx, queries, batch.ID, result, "BLING_API_PAYLOAD_INVALID", err)
		}
		created, updated, err := persistResourcePage(ctx, queries, connection.ID, batch.ID, page.Page, s.sourceEntity, payload)
		if err != nil {
			return result, s.finishFailed(ctx, queries, batch.ID, result, "BLING_API_RAW_FAILED", err)
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
			break
		}
		current.Page = page.Page + 1
		current.Limit = page.Limit
	}
	if !completed {
		return result, s.finishFailed(ctx, queries, batch.ID, result, "BLING_API_PAGE_LIMIT", ErrBlingAPISyncPageLimit)
	}
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batch.ID, Status: "SUCCESS", RecordsRead: int32(result.RecordsRead), RecordsCreated: int32(result.PagesCreated), RecordsUpdated: int32(result.PagesUpdated),
	}); err != nil {
		return result, fmt.Errorf("finish Bling API sync batch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit Bling API sync: %w", err)
	}
	result.Status = "SUCCESS"
	return result, nil
}

func (s *apiResourceSyncService) finishFailed(ctx context.Context, queries *database.Queries, batchID pgtype.UUID, result APISyncResult, code string, cause error) error {
	message := code
	if err := queries.FinishIntegrationSyncBatch(ctx, database.FinishIntegrationSyncBatchParams{
		ID: batchID, Status: "FAILED", RecordsRead: int32(result.RecordsRead), RecordsCreated: int32(result.PagesCreated), RecordsUpdated: int32(result.PagesUpdated),
		ErrorCode: &code, ErrorMessageSanitized: &message,
	}); err != nil {
		return fmt.Errorf("%w: %v; finish batch: %v", cause, cause, err)
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
