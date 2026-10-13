package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	gen "majucau.local/financial-intelligence/database/gen"

	"majucau.local/financial-intelligence/internal/application"
)

const databaseReadinessTimeout = 1500 * time.Millisecond

type workerHealthProbe interface {
	Ping(context.Context) error
	SchemaReady(context.Context) (bool, error)
}

type postgresWorkerHealthProbe struct {
	pool *pgxpool.Pool
}

func (p postgresWorkerHealthProbe) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p postgresWorkerHealthProbe) SchemaReady(ctx context.Context) (bool, error) {
	result, err := gen.New(p.pool).SchemaHealth(ctx)
	if err != nil {
		return false, err
	}
	return workerSchemaHealthReady(result), nil
}

func workerSchemaHealthReady(result gen.SchemaHealthRow) bool {
	return schemaHealthFlag(result.HasIntegrations) &&
		schemaHealthFlag(result.HasRaw) &&
		schemaHealthFlag(result.HasAudit) &&
		schemaHealthFlag(result.HasDailyBalances)
}

func schemaHealthFlag(value interface{}) bool {
	ready, ok := value.(bool)
	return ok && ready
}

type workerHealthChecker struct {
	probe             workerHealthProbe
	databaseErrorCode string
}

func newWorkerHealthChecker(pool *pgxpool.Pool, databaseErrorCode string) application.HealthChecker {
	var probe workerHealthProbe
	if pool != nil {
		probe = postgresWorkerHealthProbe{pool: pool}
	}
	return workerHealthChecker{probe: probe, databaseErrorCode: databaseErrorCode}
}

func (h workerHealthChecker) CheckHealth(ctx context.Context) application.HealthResponse {
	dependency := application.DependencyHealth{Name: "postgresql", State: application.HealthOK}
	if h.probe == nil {
		dependency.State = application.HealthDegraded
		dependency.Detail = h.databaseErrorCode
		if dependency.Detail == "" {
			dependency.Detail = "DATABASE_UNAVAILABLE"
		}
	} else {
		pingContext, cancel := context.WithTimeout(ctx, databaseReadinessTimeout)
		defer cancel()
		if err := h.probe.Ping(pingContext); err != nil {
			dependency.State = application.HealthDegraded
			dependency.Detail = "DATABASE_UNAVAILABLE"
		} else if schemaReady, err := h.probe.SchemaReady(pingContext); err != nil {
			dependency.State = application.HealthDegraded
			dependency.Detail = "DATABASE_UNAVAILABLE"
		} else if !schemaReady {
			dependency.State = application.HealthDegraded
			dependency.Detail = "DATABASE_SCHEMA_NOT_READY"
		}
	}

	return (application.StaticHealth{
		Service:      "majucau-worker",
		Version:      "dev",
		Dependencies: []application.DependencyHealth{dependency},
	}).CheckHealth(ctx)
}
