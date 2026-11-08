-- Immutable membership for one complete API snapshot of reconciled cash
-- movements. The collection response has no conciliation field, so the
-- official situacaoConciliacao=1 filter is the source of this membership.
CREATE TABLE bling_cash_movement_reconciled_snapshot_rows (
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_id text NOT NULL CHECK (source_id ~ '^[1-9][0-9]{0,63}$'),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (sync_batch_id, source_id)
);

CREATE INDEX ix_bling_cash_movement_reconciled_snapshot_source
  ON bling_cash_movement_reconciled_snapshot_rows(connection_id, source_id, sync_batch_id);

COMMENT ON TABLE bling_cash_movement_reconciled_snapshot_rows IS
  'Complete daily membership of GET /caixas filtered with situacaoConciliacao=1; absence from a completed snapshot means not reconciled at that snapshot time.';
