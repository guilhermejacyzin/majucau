-- Typed, read-only copy of the Bling payment-method fields used by its
-- cash-flow forecast. Unknown source values remain NULL and are not inferred.
CREATE TABLE bling_payment_method_terms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  source_id text NOT NULL CHECK (source_id ~ '^[1-9][0-9]{0,18}$'),
  destination smallint CHECK (destination IS NULL OR destination BETWEEN 1 AND 3),
  fee_rate numeric(20,8) NOT NULL CHECK (fee_rate >= 0),
  fee_rate_known boolean NOT NULL,
  fixed_fee numeric(19,4) NOT NULL CHECK (fixed_fee >= 0),
  fixed_fee_known boolean NOT NULL,
  settlement_days integer CHECK (settlement_days IS NULL OR settlement_days BETWEEN 0 AND 3650),
  settlement_known boolean NOT NULL,
  use_business_days boolean,
  business_days_known boolean NOT NULL,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_id),
  CHECK ((settlement_known AND settlement_days IS NOT NULL) OR (NOT settlement_known AND settlement_days IS NULL)),
  CHECK ((business_days_known AND use_business_days IS NOT NULL) OR (NOT business_days_known AND use_business_days IS NULL))
);

CREATE INDEX ix_bling_payment_method_terms_batch
  ON bling_payment_method_terms(sync_batch_id, source_id);

CREATE TRIGGER trg_bling_payment_method_terms_updated_at BEFORE UPDATE ON bling_payment_method_terms
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE bling_payment_method_terms IS
  'Typed read-only Bling payment-method settings used to reproduce the official cash-flow report; *_known flags distinguish source values from stored zero defaults.';
