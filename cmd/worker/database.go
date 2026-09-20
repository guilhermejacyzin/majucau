package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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
