// Package drememory persists and reads the arithmetic trace of a DRE run.
// Source classification and document lineage are supplied by the importing
// application and remain separate from this package.
package drememory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"majucau.local/financial-intelligence/internal/domain"
	"majucau.local/financial-intelligence/internal/financial"
)

var (
	ErrUnavailable      = errors.New("DRE memory database is unavailable")
	ErrInvalidExecution = errors.New("DRE memory execution is invalid")
	ErrMemoryNotFound   = errors.New("DRE calculation memory was not found")
)

const unavailableInputReason = "Entrada não disponível nesta execução."

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Execute creates a run, snapshot, lines, and arithmetic trace atomically. The
// source hash must be computed from the exact immutable inputs by the caller.
// This method does not classify source records or make missing inputs zero.
func (r *Repository) Execute(ctx context.Context, ruleVersionID uuid.UUID, inputHash string, in financial.DREInput) (uuid.UUID, error) {
	if r == nil || r.pool == nil {
		return uuid.Nil, ErrUnavailable
	}
	inputHash = strings.ToLower(strings.TrimSpace(inputHash))
	if ruleVersionID == uuid.Nil || len(inputHash) != 64 || !isHex(inputHash) || in.PeriodStart.IsZero() || in.PeriodEnd.IsZero() {
		return uuid.Nil, ErrInvalidExecution
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var ruleCode string
	var ruleNumber int
	var approved bool
	if err := tx.QueryRow(ctx, `SELECT rule_code, version, status = 'APPROVED' FROM financial_rule_versions WHERE id = $1`, ruleVersionID).Scan(&ruleCode, &ruleNumber, &approved); err != nil {
		return uuid.Nil, fmt.Errorf("read DRE rule version: %w", err)
	}
	if !approved {
		return uuid.Nil, ErrInvalidExecution
	}
	if !strings.HasPrefix(strings.ToUpper(ruleCode), "DRE") {
		return uuid.Nil, ErrInvalidExecution
	}
	in.RuleVersion = fmt.Sprintf("%s-v%d", ruleCode, ruleNumber)
	result, err := financial.CalculateDRE(in)
	if err != nil {
		return uuid.Nil, fmt.Errorf("calculate DRE: %w", err)
	}
	var runID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO calculation_runs (calculation_type, rule_version_id, reference_date, status, input_hash)
		VALUES ('DRE', $1, $2, 'RUNNING', $3)
		RETURNING id`, ruleVersionID, result.PeriodStart, inputHash).Scan(&runID); err != nil {
		return uuid.Nil, fmt.Errorf("create DRE calculation run: %w", err)
	}
	status := statementStatus(result.Status)
	var snapshotID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO statement_snapshots (calculation_run_id, statement_type, period_start, period_end, status)
		VALUES ($1, 'DRE', $2, $3, $4)
		RETURNING id`, runID, result.PeriodStart, result.PeriodEnd, status).Scan(&snapshotID); err != nil {
		return uuid.Nil, fmt.Errorf("create DRE snapshot: %w", err)
	}
	lineIDs, err := writeLines(ctx, tx, snapshotID, ruleVersionID, in, result)
	if err != nil {
		return uuid.Nil, err
	}
	if err := writeSteps(ctx, tx, runID, snapshotID, ruleVersionID, lineIDs, result); err != nil {
		return uuid.Nil, err
	}
	runStatus := "SUCCESS"
	if status == "PARTIALLY_CONFIRMED" || status == "DIVERGENT" {
		runStatus = "PARTIAL"
	}
	if _, err := tx.Exec(ctx, `SELECT finish_dre_calculation($1, $2)`, runID, runStatus); err != nil {
		return uuid.Nil, fmt.Errorf("finish DRE calculation run: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit DRE calculation memory: %w", err)
	}
	return snapshotID, nil
}

type Snapshot struct {
	ID          uuid.UUID `json:"id"`
	RunID       uuid.UUID `json:"run_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Status      string    `json:"status"`
	RuleCode    string    `json:"rule_code"`
	RuleVersion int       `json:"rule_version"`
	InputHash   string    `json:"input_hash"`
	Lines       []Line    `json:"lines"`
}

type Line struct {
	ID                uuid.UUID     `json:"id"`
	Code              string        `json:"code"`
	Label             string        `json:"label"`
	Amount            *domain.Money `json:"amount,omitempty"`
	Available         bool          `json:"available"`
	UnavailableReason string        `json:"unavailable_reason,omitempty"`
	Steps             []Step        `json:"steps"`
}

type Step struct {
	Operation         string        `json:"operation"`
	Formula           string        `json:"formula"`
	Output            *domain.Money `json:"output,omitempty"`
	Available         bool          `json:"available"`
	UnavailableReason string        `json:"unavailable_reason,omitempty"`
	Operands          []Operand     `json:"operands"`
}

type Operand struct {
	Code              string        `json:"code"`
	Amount            *domain.Money `json:"amount,omitempty"`
	Available         bool          `json:"available"`
	UnavailableReason string        `json:"unavailable_reason,omitempty"`
}

// LoadPeriod returns the most recently completed DRE memory for an exact
// competence range. The caller supplies month boundaries; no other period is
// substituted when the requested one has no saved execution.
func (r *Repository) LoadPeriod(ctx context.Context, periodStart, periodEnd time.Time) (Snapshot, error) {
	if r == nil || r.pool == nil {
		return Snapshot{}, ErrUnavailable
	}
	start, end := dateOnly(periodStart), dateOnly(periodEnd)
	if periodStart.IsZero() || periodEnd.IsZero() || start.After(end) {
		return Snapshot{}, ErrInvalidExecution
	}
	var snapshotID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT snapshot.id
		  FROM statement_snapshots snapshot
		  JOIN calculation_runs run ON run.id = snapshot.calculation_run_id
		 WHERE snapshot.statement_type = 'DRE'
	   AND snapshot.period_start = $1
	   AND snapshot.period_end = $2
	   AND run.calculation_type = 'DRE'
	   AND run.status IN ('SUCCESS', 'PARTIAL')
	 ORDER BY run.finished_at DESC, snapshot.calculated_at DESC, snapshot.id DESC
	 LIMIT 1`, start, end).Scan(&snapshotID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrMemoryNotFound
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("find DRE memory for period: %w", err)
	}
	return r.Load(ctx, snapshotID)
}

// Load returns the persisted arithmetic memory in display order. It reads in
// one repeatable-read transaction so lines, steps and operands share a view.
func (r *Repository) Load(ctx context.Context, snapshotID uuid.UUID) (Snapshot, error) {
	if r == nil || r.pool == nil {
		return Snapshot{}, ErrUnavailable
	}
	if snapshotID == uuid.Nil {
		return Snapshot{}, ErrInvalidExecution
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return Snapshot{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var result Snapshot
	err = tx.QueryRow(ctx, `
		SELECT s.id, s.calculation_run_id, s.period_start, s.period_end, s.status,
		       rule.rule_code, rule.version, run.input_hash
		FROM statement_snapshots s
		JOIN calculation_runs run ON run.id = s.calculation_run_id
		JOIN financial_rule_versions rule ON rule.id = run.rule_version_id
		WHERE s.id = $1 AND s.statement_type = 'DRE'`, snapshotID).
		Scan(&result.ID, &result.RunID, &result.PeriodStart, &result.PeriodEnd, &result.Status,
			&result.RuleCode, &result.RuleVersion, &result.InputHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Snapshot{}, ErrMemoryNotFound
	}
	if err != nil {
		return Snapshot{}, fmt.Errorf("read DRE snapshot: %w", err)
	}
	rows, err := tx.Query(ctx, `
		SELECT id, line_code, label, amount::text, availability_status, unavailable_reason
		FROM statement_lines WHERE snapshot_id = $1
		ORDER BY display_order, line_code`, snapshotID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read DRE lines: %w", err)
	}
	for rows.Next() {
		var line Line
		var amount *string
		var availability string
		if err := rows.Scan(&line.ID, &line.Code, &line.Label, &amount, &availability, &line.UnavailableReason); err != nil {
			rows.Close()
			return Snapshot{}, fmt.Errorf("scan DRE line: %w", err)
		}
		line.Available = availability == "AVAILABLE"
		line.Amount, err = parseOptionalMoney(amount)
		if err != nil {
			rows.Close()
			return Snapshot{}, err
		}
		result.Lines = append(result.Lines, line)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Snapshot{}, fmt.Errorf("iterate DRE lines: %w", err)
	}
	rows.Close()
	for i := range result.Lines {
		steps, err := loadSteps(ctx, tx, snapshotID, result.Lines[i].ID)
		if err != nil {
			return Snapshot{}, err
		}
		result.Lines[i].Steps = steps
	}
	if err := tx.Commit(ctx); err != nil {
		return Snapshot{}, fmt.Errorf("finish DRE memory read: %w", err)
	}
	return result, nil
}

func writeLines(ctx context.Context, tx pgx.Tx, snapshotID, ruleVersionID uuid.UUID, in financial.DREInput, result financial.DREResult) (map[financial.DRELineCode]uuid.UUID, error) {
	lines := append([]financial.DRELine(nil), result.Lines...)
	lines = append(lines,
		financial.DRELine{Code: financial.DREDepreciationAmortizationCode, Amount: in.DepreciationAndAmortization.Amount, Available: in.DepreciationAndAmortization.Available},
		financial.DRELine{Code: financial.DREBITCode, Amount: result.Indicators.EBIT.Amount, Available: result.Indicators.EBIT.Available},
		financial.DRELine{Code: financial.DREBITDACode, Amount: result.Indicators.EBITDA.Amount, Available: result.Indicators.EBITDA.Available},
	)
	lineIDs := make(map[financial.DRELineCode]uuid.UUID, len(lines))
	for order, line := range lines {
		var amount any
		var unavailableReason any
		availability := "AVAILABLE"
		if line.Available {
			amount = line.Amount.String()
		} else {
			availability = "UNAVAILABLE"
			unavailableReason = unavailableInputReason
		}
		var id uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO statement_lines (snapshot_id, line_code, label, amount, display_order, rule_version_id, availability_status, unavailable_reason)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`, snapshotID, line.Code, lineLabel(line.Code), amount, order, ruleVersionID, availability, unavailableReason).Scan(&id); err != nil {
			return nil, fmt.Errorf("write DRE line %s: %w", line.Code, err)
		}
		lineIDs[line.Code] = id
	}
	return lineIDs, nil
}

func writeSteps(ctx context.Context, tx pgx.Tx, runID, snapshotID, ruleVersionID uuid.UUID, lineIDs map[financial.DRELineCode]uuid.UUID, result financial.DREResult) error {
	for _, step := range result.CalculationMemory {
		lineID, ok := lineIDs[step.LineCode]
		if !ok {
			return fmt.Errorf("DRE memory output line %s is missing", step.LineCode)
		}
		formula, err := formulaExpression(step)
		if err != nil {
			return fmt.Errorf("write DRE calculation formula %s: %w", step.LineCode, err)
		}
		availability := "AVAILABLE"
		var output any
		var unavailableReason any
		if step.Result.Available {
			output = step.Result.Amount.String()
		} else {
			availability = "UNAVAILABLE"
			unavailableReason = unavailableInputReason
		}
		var stepID uuid.UUID
		if err := tx.QueryRow(ctx, `
			INSERT INTO dre_calculation_steps (calculation_run_id, snapshot_id, statement_line_id, step_order, operation_code, formula_expression, output_amount, availability_status, unavailable_reason, rule_version_id)
			VALUES ($1, $2, $3, 0, $4, $5, $6, $7, $8, $9)
			RETURNING id`, runID, snapshotID, lineID, step.Operation, formula, output, availability, unavailableReason, ruleVersionID).Scan(&stepID); err != nil {
			return fmt.Errorf("write DRE calculation step %s: %w", step.LineCode, err)
		}
		for index, code := range step.OperandCodes {
			operand := step.OperandValues[index]
			operandType := "STATEMENT_LINE"
			var operandLineID any
			var constantName any
			if id, exists := lineIDs[code]; exists {
				operandLineID = id
			} else {
				operandType = "CONSTANT"
				constantName = string(code)
			}
			var amount any
			var operandAvailabilityReason any
			operandAvailability := "AVAILABLE"
			if operand.Available {
				amount = operand.Amount.String()
			} else {
				operandAvailability = "UNAVAILABLE"
				operandAvailabilityReason = unavailableInputReason
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO dre_calculation_operands (calculation_run_id, snapshot_id, calculation_step_id, operand_order, operand_code, operand_type, statement_line_id, constant_name, operand_amount, availability_status, unavailable_reason)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, runID, snapshotID, stepID, index, code, operandType, operandLineID, constantName, amount, operandAvailability, operandAvailabilityReason); err != nil {
				return fmt.Errorf("write DRE operand %s: %w", code, err)
			}
		}
	}
	return nil
}

func loadSteps(ctx context.Context, tx pgx.Tx, snapshotID, lineID uuid.UUID) ([]Step, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, operation_code, formula_expression, output_amount::text, availability_status, unavailable_reason
		FROM dre_calculation_steps WHERE snapshot_id = $1 AND statement_line_id = $2 ORDER BY step_order`, snapshotID, lineID)
	if err != nil {
		return nil, fmt.Errorf("read DRE calculation steps: %w", err)
	}
	defer rows.Close()
	var steps []Step
	for rows.Next() {
		var id uuid.UUID
		var step Step
		var amount *string
		var availability string
		if err := rows.Scan(&id, &step.Operation, &step.Formula, &amount, &availability, &step.UnavailableReason); err != nil {
			return nil, fmt.Errorf("scan DRE calculation step: %w", err)
		}
		step.Available = availability == "AVAILABLE"
		step.Output, err = parseOptionalMoney(amount)
		if err != nil {
			return nil, err
		}
		step.Operands, err = loadOperands(ctx, tx, snapshotID, id)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func loadOperands(ctx context.Context, tx pgx.Tx, snapshotID, stepID uuid.UUID) ([]Operand, error) {
	rows, err := tx.Query(ctx, `
		SELECT operand_code, operand_amount::text, availability_status, unavailable_reason
		FROM dre_calculation_operands WHERE snapshot_id = $1 AND calculation_step_id = $2 ORDER BY operand_order`, snapshotID, stepID)
	if err != nil {
		return nil, fmt.Errorf("read DRE calculation operands: %w", err)
	}
	defer rows.Close()
	var result []Operand
	for rows.Next() {
		var operand Operand
		var amount *string
		var availability string
		if err := rows.Scan(&operand.Code, &amount, &availability, &operand.UnavailableReason); err != nil {
			return nil, fmt.Errorf("scan DRE calculation operand: %w", err)
		}
		operand.Available = availability == "AVAILABLE"
		operand.Amount, err = parseOptionalMoney(amount)
		if err != nil {
			return nil, err
		}
		result = append(result, operand)
	}
	return result, rows.Err()
}

func formulaExpression(step financial.DRECalculationStep) (string, error) {
	operands := make([]string, len(step.OperandCodes))
	if len(step.OperandCodes) != len(step.OperandValues) || len(operands) == 0 {
		return "", ErrInvalidExecution
	}
	for i, code := range step.OperandCodes {
		operands[i] = string(code)
	}
	switch step.Operation {
	case "SUM":
		return strings.Join(operands, " + "), nil
	case "IDENTITY":
		if len(operands) == 1 {
			return operands[0], nil
		}
	case "SUBTRACT_SIGNED_EXPENSE":
		if len(operands) == 2 {
			return operands[0] + " - (" + operands[1] + ")", nil
		}
	}
	return "", ErrInvalidExecution
}

func statementStatus(status domain.DataStatus) string {
	switch status {
	case domain.StatusConfirmed:
		return "CONFIRMED"
	case domain.StatusPartiallyConfirmed:
		return "PARTIALLY_CONFIRMED"
	case domain.StatusDivergent:
		return "DIVERGENT"
	default:
		return "PROVISIONAL"
	}
}

func parseOptionalMoney(value *string) (*domain.Money, error) {
	if value == nil {
		return nil, nil
	}
	amount, err := domain.ParseMoney(*value)
	if err != nil {
		return nil, fmt.Errorf("invalid amount in DRE memory: %w", err)
	}
	return &amount, nil
}

func isHex(value string) bool {
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func lineLabel(code financial.DRELineCode) string {
	labels := map[financial.DRELineCode]string{
		financial.DREGrossRevenueCode:             "Receita bruta",
		financial.DRERevenueDeductionsCode:        "Deduções da receita",
		financial.DRENetRevenueCode:               "Receita líquida",
		financial.DRECostOfGoodsSoldCode:          "Custo das mercadorias vendidas",
		financial.DREGrossProfitCode:              "Lucro bruto",
		financial.DRESellingExpensesCode:          "Despesas comerciais",
		financial.DREAdministrativeCode:           "Despesas administrativas",
		financial.DREOtherOperatingCode:           "Outras despesas operacionais",
		financial.DREOperatingResultCode:          "Resultado operacional",
		financial.DREFinancialResultCode:          "Resultado financeiro",
		financial.DREOtherIncomeExpenseCode:       "Outras receitas e despesas",
		financial.DREBeforeTaxCode:                "Resultado antes dos tributos sobre o lucro",
		financial.DREIncomeTaxesCode:              "Tributos sobre o lucro",
		financial.DRENetIncomeCode:                "Resultado líquido",
		financial.DREDepreciationAmortizationCode: "Depreciação e amortização",
		financial.DREBITCode:                      "EBIT",
		financial.DREBITDACode:                    "EBITDA",
	}
	if label, ok := labels[code]; ok {
		return label
	}
	return string(code)
}
