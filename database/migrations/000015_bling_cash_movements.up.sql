-- Normalized read model for Bling cash/bank movements. Direction remains
-- explicit and amount is a non-negative magnitude; the cash-flow calculation
-- applies the negative sign to outflows without confusing movements with an
-- account balance.
CREATE TABLE bling_cash_movements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  source_id text NOT NULL CHECK (source_id ~ '^[1-9][0-9]{0,63}$'),
  account_source_id text CHECK (account_source_id IS NULL OR account_source_id ~ '^[1-9][0-9]{0,63}$'),
  account_name text CHECK (account_name IS NULL OR octet_length(account_name) <= 512),
  direction text NOT NULL CHECK (direction IN ('INFLOW', 'OUTFLOW')),
  amount numeric(19,4) NOT NULL CHECK (amount >= 0),
  movement_date date NOT NULL,
  source_status text NOT NULL CHECK (source_status IN ('R', 'E', 'H', 'N', 'P', 'C')),
  origin_type text CHECK (origin_type IS NULL OR octet_length(origin_type) <= 128),
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_id)
);

CREATE INDEX ix_bling_cash_movements_snapshot_date
  ON bling_cash_movements(sync_batch_id, movement_date, direction, source_status);

CREATE TRIGGER trg_bling_cash_movements_updated_at BEFORE UPDATE ON bling_cash_movements
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE bling_cash_movements IS
  'Current normalized read model for cash/bank movements from a completed Bling batch; signed cash impact is derived from direction, never from payable title balance.';
