-- Preserve only the official Bling financial-account ID and description needed
-- to match the two accounts selected for the cash-flow report. This is account
-- metadata, not a balance or a transaction.
CREATE TABLE bling_financial_account_identities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  source_id text NOT NULL CHECK (source_id ~ '^[1-9][0-9]{0,18}$'),
  description text NOT NULL CHECK (octet_length(description) BETWEEN 1 AND 512),
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_id)
);

CREATE INDEX ix_bling_financial_account_identities_batch
  ON bling_financial_account_identities(sync_batch_id, source_id);

CREATE TRIGGER trg_bling_financial_account_identities_updated_at
  BEFORE UPDATE ON bling_financial_account_identities
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE bling_financial_account_identities IS
  'Read-only Bling account identities used to match report-selected accounts; contains no balance or movement.';
