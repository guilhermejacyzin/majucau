package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/application"
)

const databaseReadinessTimeout = 1500 * time.Millisecond

type workerHealthChecker struct {
	pool              *pgxpool.Pool
	databaseErrorCode string
}

func newWorkerHealthChecker(pool *pgxpool.Pool, databaseErrorCode string) application.HealthChecker {
	return workerHealthChecker{pool: pool, databaseErrorCode: databaseErrorCode}
}

func (h workerHealthChecker) CheckHealth(ctx context.Context) application.HealthResponse {
	dependency := application.DependencyHealth{Name: "postgresql", State: application.HealthOK}
	if h.pool == nil {
		dependency.State = application.HealthDegraded
		dependency.Detail = h.databaseErrorCode
		if dependency.Detail == "" {
			dependency.Detail = "DATABASE_UNAVAILABLE"
		}
	} else {
		pingContext, cancel := context.WithTimeout(ctx, databaseReadinessTimeout)
		defer cancel()
		if err := h.pool.Ping(pingContext); err != nil {
			dependency.State = application.HealthDegraded
			dependency.Detail = "DATABASE_UNAVAILABLE"
		}
	}

	return (application.StaticHealth{
		Service:      "majucau-worker",
		Version:      "dev",
		Dependencies: []application.DependencyHealth{dependency},
	}).CheckHealth(ctx)
}
