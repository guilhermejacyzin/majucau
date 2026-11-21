-- Store the tax decision and its evidence in the same immutable DRE snapshot
-- as the result, so every screen reads one published calculation.

ALTER TABLE statement_lines
  ADD COLUMN tax_value_kind text,
  ADD COLUMN tax_memory jsonb,
  ADD CONSTRAINT ck_statement_lines_tax_memory CHECK (
    (tax_value_kind IS NULL AND tax_memory IS NULL) OR
    (
      line_code = 'INCOME_TAXES'
      AND tax_value_kind IN ('ACTUAL', 'ESTIMATED', 'UNAVAILABLE')
      AND tax_memory IS NOT NULL
    )
  );

COMMENT ON COLUMN statement_lines.tax_value_kind IS
  'Identifies the single tax value represented in this DRE snapshot: actual, estimated, or unavailable.';
COMMENT ON COLUMN statement_lines.tax_memory IS
  'Immutable source IDs, rate, weekly bases, fallback and reason used to resolve the tax line in this snapshot.';
