package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	gen "majucau.local/financial-intelligence/database/gen"
	"majucau.local/financial-intelligence/internal/application"
)

type fakeWorkerHealthProbe struct {
	pingErr     error
	schemaReady bool
	schemaErr   error
}

func (p fakeWorkerHealthProbe) Ping(context.Context) error {
	return p.pingErr
}

func (p fakeWorkerHealthProbe) SchemaReady(context.Context) (bool, error) {
	return p.schemaReady, p.schemaErr
}

func TestWorkerHealthCheckerReportsDatabaseReadiness(t *testing.T) {
	tests := []struct {
		name       string
		probe      workerHealthProbe
		dbCode     string
		wantState  application.HealthState
		wantDetail string
	}{
		{
			name:      "ready database and schema",
			probe:     fakeWorkerHealthProbe{schemaReady: true},
			wantState: application.HealthOK,
		},
		{
			name:       "missing core schema",
			probe:      fakeWorkerHealthProbe{},
			wantState:  application.HealthDegraded,
			wantDetail: "DATABASE_SCHEMA_NOT_READY",
		},
		{
			name:       "database unavailable",
			probe:      fakeWorkerHealthProbe{pingErr: errors.New("private driver error")},
			wantState:  application.HealthDegraded,
			wantDetail: "DATABASE_UNAVAILABLE",
		},
		{
			name:       "schema query unavailable",
			probe:      fakeWorkerHealthProbe{schemaErr: errors.New("private driver error")},
			wantState:  application.HealthDegraded,
			wantDetail: "DATABASE_UNAVAILABLE",
		},
		{
			name:       "database not configured",
			dbCode:     "DATABASE_NOT_CONFIGURED",
			wantState:  application.HealthDegraded,
			wantDetail: "DATABASE_NOT_CONFIGURED",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := (workerHealthChecker{probe: test.probe, databaseErrorCode: test.dbCode}).CheckHealth(context.Background())
			if response.State != test.wantState {
				t.Fatalf("state = %q, want %q", response.State, test.wantState)
			}
			if len(response.Dependencies) != 1 {
				t.Fatalf("dependencies = %d, want 1", len(response.Dependencies))
			}
			if response.Dependencies[0].Detail != test.wantDetail {
				t.Fatalf("detail = %q, want %q", response.Dependencies[0].Detail, test.wantDetail)
			}
		})
	}
}

func TestWorkerSchemaHealthRequiresEveryCoreTable(t *testing.T) {
	ready := gen.SchemaHealthRow{
		HasIntegrations:  true,
		HasRaw:           true,
		HasAudit:         true,
		HasDailyBalances: true,
	}
	if !workerSchemaHealthReady(ready) {
		t.Fatal("all required tables present should be ready")
	}

	tests := []struct {
		name string
		row  gen.SchemaHealthRow
	}{
		{
			name: "integration table missing",
			row:  gen.SchemaHealthRow{HasIntegrations: false, HasRaw: true, HasAudit: true, HasDailyBalances: true},
		},
		{
			name: "raw table missing",
			row:  gen.SchemaHealthRow{HasIntegrations: true, HasRaw: false, HasAudit: true, HasDailyBalances: true},
		},
		{
			name: "audit table missing",
			row:  gen.SchemaHealthRow{HasIntegrations: true, HasRaw: true, HasAudit: false, HasDailyBalances: true},
		},
		{
			name: "balance table missing",
			row:  gen.SchemaHealthRow{HasIntegrations: true, HasRaw: true, HasAudit: true, HasDailyBalances: false},
		},
		{
			name: "unexpected query value",
			row:  gen.SchemaHealthRow{HasIntegrations: "true", HasRaw: true, HasAudit: true, HasDailyBalances: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if workerSchemaHealthReady(test.row) {
				t.Fatal("incomplete or invalid schema result must not be ready")
			}
		})
	}
}

func TestPostgresWorkerHealthProbeReportsMigratedSchema(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MAJUCAU_TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to run the PostgreSQL worker health integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("create PostgreSQL pool for worker health integration test")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal("PostgreSQL health integration test database is unavailable")
	}

	response := newWorkerHealthChecker(pool, "").CheckHealth(ctx)
	if response.State != application.HealthOK {
		t.Fatalf("worker health state = %q, want %q; dependencies=%+v", response.State, application.HealthOK, response.Dependencies)
	}
	if len(response.Dependencies) != 1 || response.Dependencies[0].State != application.HealthOK {
		t.Fatalf("worker PostgreSQL dependency is not healthy: %+v", response.Dependencies)
	}
}
