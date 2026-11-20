-- Preserve Bling's category identity separately from the display name so the
-- DRE source trace can map the original category without guessing by text.
ALTER TABLE receivables
  ADD COLUMN source_category_id text,
  ADD CONSTRAINT ck_receivables_source_category_id
    CHECK (source_category_id IS NULL OR source_category_id ~ '^[1-9][0-9]{0,18}$');

ALTER TABLE payables
  ADD COLUMN source_category_id text,
  ADD CONSTRAINT ck_payables_source_category_id
    CHECK (source_category_id IS NULL OR source_category_id ~ '^[1-9][0-9]{0,18}$');

CREATE INDEX ix_receivables_source_category_id
  ON receivables(connection_id, source_category_id)
  WHERE source_category_id IS NOT NULL;

CREATE INDEX ix_payables_source_category_id
  ON payables(connection_id, source_category_id)
  WHERE source_category_id IS NOT NULL;
