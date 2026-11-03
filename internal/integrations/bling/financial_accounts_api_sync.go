package bling

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const FinancialAccountsAPIResource = "contas-contabeis"

// FinancialAccountsAPISyncService stores a bounded, read-only RAW snapshot of
// the financial-account identities configured in Bling. It does not treat
// account metadata as a balance or as a projected cash-flow movement.
type FinancialAccountsAPISyncService struct{ *apiResourceSyncService }

func NewFinancialAccountsAPISyncService(pool *pgxpool.Pool, client *BlingAPIClient) *FinancialAccountsAPISyncService {
	var list func(context.Context, ReceivablesFilter) (BlingReceivablesPage, error)
	if client != nil {
		list = func(ctx context.Context, filter ReceivablesFilter) (BlingReceivablesPage, error) {
			return client.ListFinancialAccounts(ctx, filter.Page, filter.Limit)
		}
	}
	return &FinancialAccountsAPISyncService{apiResourceSyncService: &apiResourceSyncService{
		pool:         pool,
		client:       client,
		maxPages:     maxBlingAPIPageReads,
		resource:     FinancialAccountsAPIResource,
		sourceEntity: "bling.contas_financeiras.page",
		list:         list,
	}}
}

func (s *FinancialAccountsAPISyncService) SyncWithCorrelationID(ctx context.Context, filter ReceivablesFilter, correlationID pgtype.UUID) (APISyncResult, error) {
	if s == nil || s.apiResourceSyncService == nil {
		return APISyncResult{}, ErrBlingAPISyncClientRequired
	}
	return s.apiResourceSyncService.sync(ctx, filter, &correlationID)
}
