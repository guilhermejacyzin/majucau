-- Stores explicitly approved balance checkpoints outside source code and Git.
-- The worker may read checkpoints but cannot create or change them.
CREATE TABLE bling_cash_balance_checkpoints (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  account_source_id text NOT NULL CHECK (account_source_id ~ '^[1-9][0-9]{0,18}$'),
  reference_date date NOT NULL,
  reconciled_amount numeric(19,4) NOT NULL,
  source text NOT NULL CHECK (source = 'USER_APPROVED_BLING_SCREENSHOT'),
  approved_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, account_source_id, reference_date)
);

CREATE INDEX ix_bling_cash_balance_checkpoints_latest
  ON bling_cash_balance_checkpoints(connection_id, account_source_id, reference_date DESC);

COMMENT ON TABLE bling_cash_balance_checkpoints IS
  'Approved, dated opening points used to calculate subsequent balances from Bling API movements; not part of source code.';
