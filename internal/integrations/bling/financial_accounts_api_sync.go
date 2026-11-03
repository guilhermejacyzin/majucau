package bling

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	database "majucau.local/financial-intelligence/database/gen"
)

const FinancialAccountsAPIResource = "contas-contabeis"

// FinancialAccountsAPISyncService stores a bounded, read-only RAW snapshot of
// the financial-account identities configured in Bling. It does not treat
// account metadata as a balance or as a projected cash-flow movement.
type FinancialAccountsAPISyncService struct{ *apiResourceSyncService }

func NewFinancialAccountsAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *FinancialAccountsAPISyncService {
	_, normalizePage := normalizeFinancialAccountsAPISnapshotPage()
	var list func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error)
	if client != nil {
		list = func(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			return client.ListFinancialAccounts(ctx, filter.Page, filter.Limit)
		}
	}
	return &FinancialAccountsAPISyncService{apiResourceSyncService: &apiResourceSyncService{
		pool:          pool,
		client:        client,
		maxPages:      maxBlingAPIPageReads,
		resource:      FinancialAccountsAPIResource,
		sourceEntity:  "bling.contas_financeiras.page",
		list:          list,
		normalizePage: normalizePage,
	}}
}

type financialAccountAPIIdentity struct {
	ID          int64  `json:"id"`
	Description string `json:"descricao"`
}

func parseFinancialAccountAPIIdentity(payload json.RawMessage) (int64, string, error) {
	var account financialAccountAPIIdentity
	if len(payload) == 0 || json.Unmarshal(payload, &account) != nil || account.ID <= 0 {
		return 0, "", fmt.Errorf("financial-account identity is invalid")
	}
	description := strings.TrimSpace(account.Description)
	if description == "" || len(description) > 512 {
		return 0, "", fmt.Errorf("financial-account description is invalid")
	}
	return account.ID, description, nil
}

func normalizeFinancialAccountsAPISnapshotPage() (
	func(context.Context, pgtype.UUID, []json.RawMessage) ([]json.RawMessage, error),
	func(context.Context, *database.Queries, pgtype.UUID, pgtype.UUID, []json.RawMessage) error,
) {
	normalize := func(ctx context.Context, queries *database.Queries, connectionID, batchID pgtype.UUID, records []json.RawMessage) error {
		if queries == nil || !connectionID.Valid || !batchID.Valid {
			return fmt.Errorf("financial-account normalization dependencies are missing")
		}
		for _, payload := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			accountID, description, err := parseFinancialAccountAPIIdentity(payload)
			if err != nil {
				return err
			}
			sourceID := strconv.FormatInt(accountID, 10)
			rawID, err := persistCurrentBlingEntity(ctx, queries, connectionID, batchID, "bling.conta_financeira", sourceID, payload)
			if err != nil {
				return fmt.Errorf("persist Bling financial-account source: %w", err)
			}
			if err := queries.UpsertBlingFinancialAccountIdentity(ctx, database.UpsertBlingFinancialAccountIdentityParams{
				ConnectionID: connectionID,
				SyncBatchID:  batchID,
				SourceID:     sourceID,
				Description:  description,
				RawRecordID:  rawID,
			}); err != nil {
				return fmt.Errorf("persist normalized Bling financial-account identity: %w", err)
			}
		}
		return nil
	}
	return nil, normalize
}

func (s *FinancialAccountsAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}
