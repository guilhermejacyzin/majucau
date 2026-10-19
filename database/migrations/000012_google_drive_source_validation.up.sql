-- Persist only bounded validation counts for approved Drive sources whose
-- financial contracts have not been approved. File names and contents stay
-- out of PostgreSQL; replaying a scheduled run updates the same source row.
CREATE TABLE google_drive_source_validation_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  folder_scope text NOT NULL CHECK (folder_scope IN (
    '01_perdas_operacionais',
    '03_nuvem_envio'
  )),
  scheduled_at timestamptz NOT NULL,
  inventory_batch_id uuid REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  status text NOT NULL CHECK (status IN ('PENDING', 'PARTIAL', 'FAILED')),
  files_read bigint NOT NULL DEFAULT 0 CHECK (files_read >= 0),
  files_validated bigint NOT NULL DEFAULT 0 CHECK (files_validated >= 0),
  files_pending bigint NOT NULL DEFAULT 0 CHECK (files_pending >= 0),
  files_invalid bigint NOT NULL DEFAULT 0 CHECK (files_invalid >= 0),
  files_failed bigint NOT NULL DEFAULT 0 CHECK (files_failed >= 0),
  files_stale bigint NOT NULL DEFAULT 0 CHECK (files_stale >= 0),
  files_deferred bigint NOT NULL DEFAULT 0 CHECK (files_deferred >= 0),
  files_unsupported bigint NOT NULL DEFAULT 0 CHECK (files_unsupported >= 0),
  source_missing boolean NOT NULL DEFAULT false,
  error_code text CHECK (error_code IN (
    'GOOGLE_DRIVE_SOURCE_CONTRACT_PENDING',
    'GOOGLE_DRIVE_SOURCE_VALIDATION_FAILED',
    'GOOGLE_DRIVE_SOURCE_INVENTORY_UNAVAILABLE'
  )),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, folder_scope, scheduled_at)
);

CREATE INDEX ix_google_drive_source_validation_latest
  ON google_drive_source_validation_runs(connection_id, folder_scope, scheduled_at DESC);

COMMENT ON TABLE google_drive_source_validation_runs IS
  'Daily, idempotent structural-validation summary for the two approved Drive folders that remain pending financial contract approval. Stores counts only, never file names or contents.';
