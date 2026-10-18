-- Store bounded Bling pages durably so a daily read can resume after a worker
-- restart without publishing an incomplete snapshot as current RAW data.
CREATE TABLE bling_sync_page_checkpoints (
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
  resource text NOT NULL,
  correlation_id uuid NOT NULL,
  page_number integer NOT NULL CHECK (page_number >= 1),
  page_limit integer NOT NULL CHECK (page_limit >= 0),
  has_next boolean NOT NULL,
  record_count integer NOT NULL CHECK (record_count >= 0),
  payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-f]{64}$'),
  payload jsonb NOT NULL,
  checkpointed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (connection_id, resource, correlation_id, page_number)
);

CREATE INDEX ix_bling_sync_page_checkpoints_run
  ON bling_sync_page_checkpoints(connection_id, resource, correlation_id);

COMMENT ON TABLE bling_sync_page_checkpoints IS
  'Temporary, bounded RAW pages for one Bling load. Pages become current RAW only in the final atomic publish transaction; stale runs are removed when a new correlation starts.';

COMMENT ON COLUMN bling_sync_page_checkpoints.payload IS
  'Sensitive provider response page. Access is restricted to the worker runtime role and deleted after successful atomic publish or when superseded.';
