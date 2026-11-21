-- DAS is a revenue deduction in the official Majucau monthly DRE.
-- Keep its immutable resolution on the same snapshot line that carries it.

ALTER TABLE statement_lines
  DROP CONSTRAINT ck_statement_lines_tax_memory;

ALTER TABLE statement_lines
  ADD CONSTRAINT ck_statement_lines_tax_memory CHECK (
    (tax_value_kind IS NULL AND tax_memory IS NULL) OR
    (
      line_code = 'REVENUE_DEDUCTIONS'
      AND tax_value_kind IN ('ACTUAL', 'ESTIMATED', 'UNAVAILABLE')
      AND tax_memory IS NOT NULL
    )
  );

COMMENT ON COLUMN statement_lines.tax_value_kind IS
  'Identifies whether the DAS included in revenue deductions is actual, estimated, or unavailable.';
COMMENT ON COLUMN statement_lines.tax_memory IS
  'Immutable sources, rate, weekly bases, fallback and reason for the DAS value in revenue deductions.';
