package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/application"
	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/integrations/bling"
)

// newReceiptImporterFromEnvironment keeps the DSN out of arguments and logs.
// The installer will provide MAJUCAU_DATABASE_URL after creating the dedicated
// loopback PostgreSQL instance; local development without a database remains
// usable for health and folder preview.
func newReceiptImporterFromEnvironment() *bling.ReceiptImportService {
	return bling.NewReceiptImportService(newDatabasePoolFromEnvironment())
}

func newDatabasePoolFromEnvironment() *pgxpool.Pool {
	dsn := strings.TrimSpace(os.Getenv("MAJUCAU_DATABASE_URL"))
	if dsn == "" {
		return nil
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil
	}
	config.MaxConns = 4
	config.MinConns = 1
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil
	}
	return pool
}

type postgresIntegrationStatusReader struct {
	queries *database.Queries
}

func newIntegrationStatusReader(pool *pgxpool.Pool) application.IntegrationStatusReader {
	if pool == nil {
		return nil
	}
	return &postgresIntegrationStatusReader{queries: database.New(pool)}
}

func (r *postgresIntegrationStatusReader) IntegrationStatuses(ctx context.Context) ([]application.IntegrationStatus, error) {
	rows, err := r.queries.ListIntegrationStatus(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]application.IntegrationStatus, 0, len(rows)+2)
	seen := map[domain.Origin]bool{}
	for _, row := range rows {
		provider := domain.Origin(row.Provider)
		if provider != domain.OriginBling && provider != domain.OriginNuvemshop && provider != domain.OriginNuvemPago {
			continue
		}
		seen[provider] = true
		status := domain.IntegrationStatus(row.Status)
		if !status.Valid() {
			status = domain.IntegrationSchemaMismatch
		}
		result = append(result, application.IntegrationStatus{Provider: provider, Status: status, LastSuccessAt: nullableTime(row.LastSuccessAt), LastAttemptAt: nullableTime(row.LastAttemptAt), ErrorCode: optionalString(row.LastErrorCode)})
	}
	if !seen[domain.OriginBling] {
		result = append(result, application.IntegrationStatus{Provider: domain.OriginBling, Status: domain.IntegrationNotConfigured})
	}
	if !seen[domain.OriginNuvemshop] {
		result = append(result, application.IntegrationStatus{Provider: domain.OriginNuvemshop, Status: domain.IntegrationNotConfigured})
	}
	if !seen[domain.OriginNuvemPago] {
		result = append(result, application.IntegrationStatus{Provider: domain.OriginNuvemPago, Status: domain.IntegrationUnavailable})
	}
	return result, nil
}

func nullableTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	copy := value.Time
	return &copy
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
