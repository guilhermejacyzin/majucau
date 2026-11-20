package database_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestDRECalculationMemoryPostgresConstraintsAndImmutability(t *testing.T) {
	dsn := os.Getenv("MAJUCAU_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAJUCAU_TEST_DATABASE_URL to validate the DRE memory schema against PostgreSQL")
	}

	ctx := context.Background()
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to disposable PostgreSQL: %v", err)
	}
	defer connection.Close(ctx)

	tx, err := connection.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ruleVersionID, calculationRunID, snapshotID, inputLineID, resultLineID, unavailableLineID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO financial_rule_versions (rule_code, version, description, status)
		VALUES ('DRE_MEMORY_SCHEMA_TEST', 1, 'CI-only DRE memory schema fixture', 'DRAFT')
		RETURNING id
	`).Scan(&ruleVersionID); err != nil {
		t.Fatalf("insert rule version fixture: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO calculation_runs (calculation_type, rule_version_id, reference_date, input_hash)
		VALUES ('DRE_MEMORY_SCHEMA_TEST', $1, DATE '2026-09-01', repeat('a', 64))
		RETURNING id
	`, ruleVersionID).Scan(&calculationRunID); err != nil {
		t.Fatalf("insert calculation run fixture: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_snapshots (calculation_run_id, statement_type, period_start, period_end, status)
		VALUES ($1, 'DRE', DATE '2026-09-01', DATE '2026-09-30', 'PROVISIONAL')
		RETURNING id
	`, calculationRunID).Scan(&snapshotID); err != nil {
		t.Fatalf("insert statement snapshot fixture: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_lines (snapshot_id, line_code, label, amount, display_order, rule_version_id)
		VALUES ($1, 'GROSS_REVENUE', 'Receita bruta', 120.00, 1, $2)
		RETURNING id
	`, snapshotID, ruleVersionID).Scan(&inputLineID); err != nil {
		t.Fatalf("insert input statement line: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_lines (snapshot_id, line_code, label, amount, display_order, rule_version_id)
		VALUES ($1, 'NET_REVENUE', 'Receita líquida', 100.00, 2, $2)
		RETURNING id
	`, snapshotID, ruleVersionID).Scan(&resultLineID); err != nil {
		t.Fatalf("insert result statement line: %v", err)
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_lines (
		  snapshot_id, line_code, label, amount, display_order, rule_version_id,
		  availability_status, unavailable_reason
		) VALUES ($1, 'GROSS_PROFIT', 'Lucro bruto', NULL, 3, $2, 'UNAVAILABLE', 'CMV ainda não disponível')
		RETURNING id
	`, snapshotID, ruleVersionID).Scan(&unavailableLineID); err != nil {
		t.Fatalf("insert unavailable statement line without a fake zero: %v", err)
	}
	var workingCapitalSnapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_snapshots (calculation_run_id, statement_type, period_start, period_end, status)
		VALUES ($1, 'WORKING_CAPITAL', DATE '2026-09-01', DATE '2026-09-30', 'PARTIALLY_CONFIRMED')
		RETURNING id
	`, calculationRunID).Scan(&workingCapitalSnapshotID); err != nil {
		t.Fatalf("insert working-capital snapshot fixture: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO statement_lines (
		  snapshot_id, line_code, label, amount, display_order, rule_version_id,
		  availability_status, unavailable_reason
		) VALUES ($1, 'CURRENT_ASSETS', 'Ativos circulantes', NULL, 1, $2, 'UNAVAILABLE', 'Fonte indisponível')
	`, workingCapitalSnapshotID, ruleVersionID); err != nil {
		t.Fatalf("insert unavailable working-capital line: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO statement_lines (snapshot_id, line_code, label, amount, display_order, rule_version_id)
		VALUES ($1, 'CURRENT_LIABILITIES', 'Passivos circulantes', 50.00, 2, $2)
	`, workingCapitalSnapshotID, ruleVersionID); err != nil {
		t.Fatalf("insert available working-capital line: %v", err)
	}
	var unavailableWorkingCapital bool
	if err := tx.QueryRow(ctx, `
		SELECT current_assets IS NULL AND working_capital IS NULL
		  FROM working_capital_snapshots WHERE id = $1
	`, workingCapitalSnapshotID).Scan(&unavailableWorkingCapital); err != nil {
		t.Fatalf("read working-capital snapshot: %v", err)
	}
	if !unavailableWorkingCapital {
		t.Fatal("working-capital view converted an unavailable source to zero")
	}

	var stepID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO dre_calculation_steps (
		  calculation_run_id, snapshot_id, statement_line_id, step_order,
		  operation_code, formula_expression, output_amount,
		  availability_status, rule_version_id
		) VALUES ($1, $2, $3, 0, 'SUM', 'GROSS_REVENUE + REVENUE_DEDUCTIONS', 100.00, 'AVAILABLE', $4)
		RETURNING id
	`, calculationRunID, snapshotID, resultLineID, ruleVersionID).Scan(&stepID); err != nil {
		t.Fatalf("insert available calculation step: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO dre_calculation_operands (
		  calculation_run_id, snapshot_id, calculation_step_id, operand_order,
		  operand_code, operand_type, statement_line_id, operand_amount, availability_status
		) VALUES ($1, $2, $3, 0, 'GROSS_REVENUE', 'STATEMENT_LINE', $4, 120.00, 'AVAILABLE')
	`, calculationRunID, snapshotID, stepID, inputLineID); err != nil {
		t.Fatalf("insert calculation operand: %v", err)
	}
	var unavailableStepID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO dre_calculation_steps (
		  calculation_run_id, snapshot_id, statement_line_id, step_order,
		  operation_code, formula_expression, output_amount,
		  availability_status, unavailable_reason, rule_version_id
		) VALUES ($1, $2, $3, 0, 'SUM', 'NET_REVENUE + COST_OF_GOODS_SOLD', NULL, 'UNAVAILABLE', 'CMV ainda não disponível', $4)
		RETURNING id
	`, calculationRunID, snapshotID, unavailableLineID, ruleVersionID).Scan(&unavailableStepID); err != nil {
		t.Fatalf("insert unavailable calculation step without a result amount: %v", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO dre_calculation_operands (
		  calculation_run_id, snapshot_id, calculation_step_id, operand_order,
		  operand_code, operand_type, statement_line_id, operand_amount,
		  availability_status, unavailable_reason
		) VALUES ($1, $2, $3, 0, 'COST_OF_GOODS_SOLD', 'STATEMENT_LINE', $4, NULL, 'UNAVAILABLE', 'CMV ainda não disponível')
	`, calculationRunID, snapshotID, unavailableStepID, unavailableLineID); err != nil {
		t.Fatalf("insert unavailable operand without a fake zero: %v", err)
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT invalid_unavailable_value"); err != nil {
		t.Fatalf("create constraint savepoint: %v", err)
	}
	_, invalidErr := tx.Exec(ctx, `
		INSERT INTO dre_calculation_steps (
		  calculation_run_id, snapshot_id, statement_line_id, step_order,
		  operation_code, formula_expression, output_amount,
		  availability_status, unavailable_reason, rule_version_id
		) VALUES ($1, $2, $3, 1, 'SUM', 'missing input', 0, 'UNAVAILABLE', 'input missing', $4)
	`, calculationRunID, snapshotID, resultLineID, ruleVersionID)
	var postgresErr *pgconn.PgError
	if !errors.As(invalidErr, &postgresErr) || postgresErr.Code != "23514" {
		t.Fatalf("unavailable formula with numeric result should violate a check constraint, got %v", invalidErr)
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT invalid_unavailable_value"); err != nil {
		t.Fatalf("rollback invalid insert: %v", err)
	}

	if _, err := tx.Exec(ctx, "SAVEPOINT immutable_step"); err != nil {
		t.Fatalf("create immutability savepoint: %v", err)
	}
	_, mutationErr := tx.Exec(ctx, `UPDATE dre_calculation_steps SET operation_code = 'CHANGED' WHERE id = $1`, stepID)
	if !errors.As(mutationErr, &postgresErr) || postgresErr.Code != "P0001" {
		t.Fatalf("mutating calculation memory should be rejected, got %v", mutationErr)
	}
	if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT immutable_step"); err != nil {
		t.Fatalf("rollback immutable update: %v", err)
	}
}
