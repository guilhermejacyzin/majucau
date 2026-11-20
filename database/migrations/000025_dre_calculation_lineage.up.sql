-- DRE audit trail. These relations preserve the immutable path from a result
-- line to its source rows, inventory calculations, and approved formula steps.

ALTER TABLE statement_snapshots
  ADD CONSTRAINT uq_statement_snapshots_id_run UNIQUE (id, calculation_run_id);

ALTER TABLE statement_lines
  ADD COLUMN availability_status text NOT NULL DEFAULT 'AVAILABLE'
    CHECK (availability_status IN ('AVAILABLE', 'UNAVAILABLE')),
  ADD COLUMN unavailable_reason text,
  ADD CONSTRAINT uq_statement_lines_id_snapshot UNIQUE (id, snapshot_id),
  ALTER COLUMN amount DROP NOT NULL,
  ADD CONSTRAINT ck_statement_lines_availability_reason CHECK (
    (availability_status = 'AVAILABLE' AND amount IS NOT NULL AND unavailable_reason IS NULL) OR
    (availability_status = 'UNAVAILABLE' AND amount IS NULL AND length(trim(coalesce(unavailable_reason, ''))) > 0)
  );

CREATE OR REPLACE VIEW working_capital_snapshots AS
SELECT s.id, s.calculation_run_id, s.period_start, s.period_end, s.scenario_id, s.status,
       s.calculated_at,
       (CASE
         WHEN COUNT(l.id) FILTER (WHERE l.line_code = 'CURRENT_ASSETS' AND l.availability_status = 'AVAILABLE') = 0
           OR BOOL_OR(l.line_code = 'CURRENT_ASSETS' AND l.availability_status = 'UNAVAILABLE')
         THEN NULL
         ELSE SUM(l.amount) FILTER (WHERE l.line_code = 'CURRENT_ASSETS' AND l.availability_status = 'AVAILABLE')
       END)::numeric(19,4) AS current_assets,
       (CASE
         WHEN COUNT(l.id) FILTER (WHERE l.line_code = 'CURRENT_LIABILITIES' AND l.availability_status = 'AVAILABLE') = 0
           OR BOOL_OR(l.line_code = 'CURRENT_LIABILITIES' AND l.availability_status = 'UNAVAILABLE')
         THEN NULL
         ELSE SUM(l.amount) FILTER (WHERE l.line_code = 'CURRENT_LIABILITIES' AND l.availability_status = 'AVAILABLE')
       END)::numeric(19,4) AS current_liabilities,
       (CASE
         WHEN COUNT(l.id) FILTER (WHERE l.line_code = 'CURRENT_ASSETS' AND l.availability_status = 'AVAILABLE') = 0
           OR COUNT(l.id) FILTER (WHERE l.line_code = 'CURRENT_LIABILITIES' AND l.availability_status = 'AVAILABLE') = 0
           OR BOOL_OR(l.line_code IN ('CURRENT_ASSETS', 'CURRENT_LIABILITIES') AND l.availability_status = 'UNAVAILABLE')
         THEN NULL
         ELSE SUM(CASE
           WHEN l.line_code = 'CURRENT_ASSETS' THEN l.amount
           WHEN l.line_code = 'CURRENT_LIABILITIES' THEN -l.amount
           ELSE 0
         END) FILTER (WHERE l.line_code IN ('CURRENT_ASSETS', 'CURRENT_LIABILITIES'))
       END)::numeric(19,4) AS working_capital
  FROM statement_snapshots s
  LEFT JOIN statement_lines l ON l.snapshot_id = s.id
 WHERE s.statement_type = 'WORKING_CAPITAL'
 GROUP BY s.id, s.calculation_run_id, s.period_start, s.period_end, s.scenario_id, s.status, s.calculated_at;

CREATE TABLE dre_source_rows (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  source_table text NOT NULL,
  source_record_id text NOT NULL,
  source_item_key text NOT NULL DEFAULT '',
  source_document_ref text,
  source_row_number bigint CHECK (source_row_number IS NULL OR source_row_number > 0),
  competence_date date,
  category_id text,
  original_amount numeric(19,4),
  signed_amount numeric(19,4),
  movement_type text,
  transaction_type text,
  source_status text,
  origin_type text,
  origin_id text,
  order_id text,
  invoice_id text,
  product_id text,
  quantity numeric(19,6),
  included boolean NOT NULL,
  exclusion_reason text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (calculation_run_id, source_table, source_record_id, source_item_key),
  CHECK (length(trim(source_table)) > 0 AND length(trim(source_record_id)) > 0),
  CHECK (
    (included AND exclusion_reason IS NULL) OR
    (NOT included AND length(trim(coalesce(exclusion_reason, ''))) > 0)
  ),
  UNIQUE (id, calculation_run_id)
);

CREATE INDEX ix_dre_source_rows_competence_category
  ON dre_source_rows(calculation_run_id, competence_date, category_id);
CREATE INDEX ix_dre_source_rows_raw_record
  ON dre_source_rows(raw_record_id);

CREATE TABLE dre_inventory_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  event_id text NOT NULL,
  source_row_id uuid REFERENCES dre_source_rows(id) ON DELETE RESTRICT,
  event_date date NOT NULL,
  priority integer NOT NULL CHECK (priority >= 0),
  product_id text NOT NULL,
  quantity_delta numeric(19,6) NOT NULL,
  unit_cost_before numeric(19,8),
  unit_cost_after numeric(19,8),
  value_effect numeric(19,4),
  bom_version text,
  anchor_id text,
  warning text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (calculation_run_id, event_id),
  UNIQUE (id, calculation_run_id),
  FOREIGN KEY (source_row_id, calculation_run_id)
    REFERENCES dre_source_rows(id, calculation_run_id) ON DELETE RESTRICT
);

CREATE INDEX ix_dre_inventory_events_product_date
  ON dre_inventory_events(calculation_run_id, product_id, event_date, priority);

CREATE TABLE dre_sale_costs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  order_id text NOT NULL,
  item_id text NOT NULL,
  product_id text NOT NULL,
  event_date date NOT NULL,
  quantity numeric(19,6) NOT NULL,
  cmp_before numeric(19,8),
  cmv numeric(19,4),
  inventory_event_id uuid,
  cost_completeness text NOT NULL
    CHECK (cost_completeness IN ('COMPLETE', 'PARTIAL', 'UNAVAILABLE')),
  unavailable_reason text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (calculation_run_id, order_id, item_id),
  UNIQUE (id, calculation_run_id),
  FOREIGN KEY (inventory_event_id, calculation_run_id)
    REFERENCES dre_inventory_events(id, calculation_run_id) ON DELETE RESTRICT,
  CHECK (
    (cost_completeness = 'COMPLETE' AND cmp_before IS NOT NULL AND cmv IS NOT NULL AND unavailable_reason IS NULL) OR
    (cost_completeness <> 'COMPLETE' AND length(trim(coalesce(unavailable_reason, ''))) > 0)
  )
);

CREATE INDEX ix_dre_sale_costs_run_date
  ON dre_sale_costs(calculation_run_id, event_date);

CREATE TABLE dre_adjustments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  target_line_code text NOT NULL,
  source_line_code text,
  amount_signed numeric(19,4) NOT NULL,
  reason_code text NOT NULL,
  evidence_ref text NOT NULL,
  approved_by uuid NOT NULL REFERENCES system_users(id) ON DELETE RESTRICT,
  approved_at timestamptz NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (id, calculation_run_id)
);

CREATE TABLE dre_line_sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  snapshot_id uuid NOT NULL,
  statement_line_id uuid NOT NULL,
  source_row_id uuid,
  inventory_event_id uuid,
  sale_cost_id uuid,
  adjustment_id uuid,
  category_id text,
  source_amount numeric(19,4),
  contribution_amount numeric(19,4),
  included boolean NOT NULL,
  inclusion_rule text NOT NULL,
  exclusion_reason text,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  CHECK (length(trim(inclusion_rule)) > 0),
  FOREIGN KEY (snapshot_id, calculation_run_id)
    REFERENCES statement_snapshots(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (statement_line_id, snapshot_id)
    REFERENCES statement_lines(id, snapshot_id) ON DELETE RESTRICT,
  FOREIGN KEY (source_row_id, calculation_run_id)
    REFERENCES dre_source_rows(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (inventory_event_id, calculation_run_id)
    REFERENCES dre_inventory_events(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (sale_cost_id, calculation_run_id)
    REFERENCES dre_sale_costs(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (adjustment_id, calculation_run_id)
    REFERENCES dre_adjustments(id, calculation_run_id) ON DELETE RESTRICT,
  CHECK (num_nonnulls(source_row_id, inventory_event_id, sale_cost_id, adjustment_id) = 1),
  CHECK (
    (included AND contribution_amount IS NOT NULL AND exclusion_reason IS NULL) OR
    (NOT included AND contribution_amount IS NULL AND length(trim(coalesce(exclusion_reason, ''))) > 0)
  )
);

CREATE INDEX ix_dre_line_sources_statement_line
  ON dre_line_sources(statement_line_id, included);
CREATE INDEX ix_dre_line_sources_source_row
  ON dre_line_sources(source_row_id) WHERE source_row_id IS NOT NULL;

CREATE TABLE dre_calculation_steps (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  snapshot_id uuid NOT NULL,
  statement_line_id uuid NOT NULL,
  step_order integer NOT NULL CHECK (step_order >= 0),
  operation_code text NOT NULL,
  formula_expression text NOT NULL,
  output_amount numeric(19,4),
  availability_status text NOT NULL CHECK (availability_status IN ('AVAILABLE', 'UNAVAILABLE')),
  unavailable_reason text,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (statement_line_id, step_order),
  UNIQUE (id, snapshot_id),
  CHECK (length(trim(operation_code)) > 0 AND length(trim(formula_expression)) > 0),
  FOREIGN KEY (snapshot_id, calculation_run_id)
    REFERENCES statement_snapshots(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (statement_line_id, snapshot_id)
    REFERENCES statement_lines(id, snapshot_id) ON DELETE RESTRICT,
  CHECK (
    (availability_status = 'AVAILABLE' AND output_amount IS NOT NULL AND unavailable_reason IS NULL) OR
    (availability_status = 'UNAVAILABLE' AND output_amount IS NULL AND length(trim(coalesce(unavailable_reason, ''))) > 0)
  )
);

CREATE TABLE dre_calculation_operands (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  snapshot_id uuid NOT NULL,
  calculation_step_id uuid NOT NULL,
  operand_order integer NOT NULL CHECK (operand_order >= 0),
  operand_code text NOT NULL,
  operand_type text NOT NULL CHECK (operand_type IN ('STATEMENT_LINE', 'CONSTANT')),
  statement_line_id uuid,
  constant_name text,
  operand_amount numeric(19,4),
  availability_status text NOT NULL CHECK (availability_status IN ('AVAILABLE', 'UNAVAILABLE')),
  unavailable_reason text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (calculation_step_id, operand_order),
  CHECK (length(trim(operand_code)) > 0),
  FOREIGN KEY (snapshot_id, calculation_run_id)
    REFERENCES statement_snapshots(id, calculation_run_id) ON DELETE RESTRICT,
  FOREIGN KEY (calculation_step_id, snapshot_id)
    REFERENCES dre_calculation_steps(id, snapshot_id) ON DELETE RESTRICT,
  FOREIGN KEY (statement_line_id, snapshot_id)
    REFERENCES statement_lines(id, snapshot_id) ON DELETE RESTRICT,
  CHECK (
    (operand_type = 'STATEMENT_LINE' AND statement_line_id IS NOT NULL AND constant_name IS NULL) OR
    (operand_type = 'CONSTANT' AND statement_line_id IS NULL AND length(trim(coalesce(constant_name, ''))) > 0)
  ),
  CHECK (
    (availability_status = 'AVAILABLE' AND operand_amount IS NOT NULL AND unavailable_reason IS NULL) OR
    (availability_status = 'UNAVAILABLE' AND operand_amount IS NULL AND length(trim(coalesce(unavailable_reason, ''))) > 0)
  )
);

CREATE INDEX ix_dre_calculation_steps_line
  ON dre_calculation_steps(statement_line_id, step_order);
CREATE INDEX ix_dre_calculation_operands_step
  ON dre_calculation_operands(calculation_step_id, operand_order);

CREATE FUNCTION prevent_dre_memory_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  RAISE EXCEPTION 'DRE calculation memory is immutable; create a new calculation run';
END;
$$;

CREATE TRIGGER trg_dre_source_rows_immutable
BEFORE UPDATE OR DELETE ON dre_source_rows
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_inventory_events_immutable
BEFORE UPDATE OR DELETE ON dre_inventory_events
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_sale_costs_immutable
BEFORE UPDATE OR DELETE ON dre_sale_costs
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_adjustments_immutable
BEFORE UPDATE OR DELETE ON dre_adjustments
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_line_sources_immutable
BEFORE UPDATE OR DELETE ON dre_line_sources
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_calculation_steps_immutable
BEFORE UPDATE OR DELETE ON dre_calculation_steps
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();
CREATE TRIGGER trg_dre_calculation_operands_immutable
BEFORE UPDATE OR DELETE ON dre_calculation_operands
FOR EACH ROW EXECUTE FUNCTION prevent_dre_memory_mutation();

COMMENT ON TABLE dre_source_rows IS
  'Normalized, non-payload source references for one DRE calculation run; raw source content remains in append-only raw_records.';
COMMENT ON TABLE dre_line_sources IS
  'Included and excluded source contributions for a published DRE statement line, with the applied rule and reason.';
COMMENT ON TABLE dre_calculation_steps IS
  'Immutable formula and result steps for a statement line; unavailable calculations have no numeric result.';
