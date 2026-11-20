package drememory

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/financial"
)

// This integration test uses only synthetic amounts and a disposable database.
// It verifies storage mechanics, not the company's financial rules or data.
func TestRepositoryPersistsAndReadsArithmeticMemory(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MAJUCAU_TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to validate DRE memory persistence against PostgreSQL")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to disposable PostgreSQL: %v", err)
	}
	defer pool.Close()

	fixtureID := uuid.NewString()
	var userID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO system_users (windows_sid, display_name, role_id)
		SELECT $1, 'DRE memory integration fixture', id FROM roles WHERE code = 'ADMIN'
		RETURNING id`, "DRE-MEMORY-TEST-"+fixtureID).Scan(&userID); err != nil {
		t.Fatalf("create synthetic approved-rule actor: %v", err)
	}
	var ruleVersionID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO financial_rule_versions (rule_code, version, description, status, approved_by, approved_at)
		VALUES ($1, 1, 'Synthetic DRE memory persistence test', 'APPROVED', $2, clock_timestamp())
		RETURNING id`, "DRE_MEMORY_TEST_"+strings.ReplaceAll(fixtureID, "-", ""), userID).Scan(&ruleVersionID); err != nil {
		t.Fatalf("create synthetic approved rule: %v", err)
	}

	periodStart := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC)
	input := financial.DREInput{
		PeriodStart: periodStart, PeriodEnd: periodEnd, RuleVersion: "ignored-in-favor-of-database-rule-version",
		GrossRevenue:                DREAmount("100.00"),
		RevenueDeductions:           DREAmount("-10.00"),
		CostOfGoodsSold:             DREAmount("-20.00"),
		SellingExpenses:             DREAmount("-5.00"),
		Administrative:              DREAmount("-5.00"),
		OtherOperating:              DREAmount("0.00"),
		DepreciationAndAmortization: DREAmount("-2.00"),
		FinancialResult:             DREAmount("0.00"),
		OtherIncomeExpenses:         DREAmount("0.00"),
		IncomeTaxes:                 financial.DREAmount{Available: false},
	}

	repository := NewRepository(pool)
	snapshotID, err := repository.Execute(ctx, ruleVersionID, strings.Repeat("a", 64), input)
	if err != nil {
		t.Fatalf("persist synthetic DRE calculation: %v", err)
	}
	snapshot, err := repository.Load(ctx, snapshotID)
	if err != nil {
		t.Fatalf("load persisted DRE calculation memory: %v", err)
	}
	if snapshot.Status != "PARTIALLY_CONFIRMED" {
		t.Fatalf("snapshot status = %q, want PARTIALLY_CONFIRMED", snapshot.Status)
	}
	if snapshot.RuleCode != "DRE_MEMORY_TEST_"+strings.ReplaceAll(fixtureID, "-", "") || snapshot.RuleVersion != 1 || snapshot.InputHash != strings.Repeat("a", 64) {
		t.Fatalf("snapshot audit identity is incomplete: code=%q version=%d hash=%q", snapshot.RuleCode, snapshot.RuleVersion, snapshot.InputHash)
	}
	if got := findLine(t, snapshot, "NET_REVENUE"); got.Amount == nil || got.Amount.String() != "90.0000" || len(got.Steps) != 1 || got.Steps[0].Formula != "GROSS_REVENUE + REVENUE_DEDUCTIONS" {
		t.Fatalf("net-revenue calculation memory is incomplete: %+v", got)
	}
	if got := findLine(t, snapshot, "NET_INCOME"); got.Available || got.Amount != nil || got.UnavailableReason == "" || len(got.Steps) != 1 || got.Steps[0].Available {
		t.Fatalf("unavailable taxes were incorrectly turned into a result: %+v", got)
	}
	if got := findLine(t, snapshot, "EBITDA"); got.Amount == nil || got.Amount.String() != "62.0000" {
		t.Fatalf("EBITDA input trace result = %+v, want 62.0000", got)
	}
	var runStatus string
	if err := pool.QueryRow(ctx, `SELECT status FROM calculation_runs WHERE id = $1`, snapshot.RunID).Scan(&runStatus); err != nil {
		t.Fatalf("read finalized DRE run: %v", err)
	}
	if runStatus != "PARTIAL" {
		t.Fatalf("calculation run status = %q, want PARTIAL", runStatus)
	}
}

func DREAmount(value string) financial.DREAmount {
	amount, err := domain.ParseMoney(value)
	if err != nil {
		panic(err)
	}
	return financial.DREAmount{Amount: amount, Available: true}
}

func findLine(t *testing.T, snapshot Snapshot, code string) Line {
	t.Helper()
	for _, line := range snapshot.Lines {
		if line.Code == code {
			return line
		}
	}
	t.Fatalf("DRE line %q was not persisted", code)
	return Line{}
}
