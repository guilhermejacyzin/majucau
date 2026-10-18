-- Keep Google Drive file listings staged by page until a complete snapshot is
-- available. Publishing a snapshot then atomically updates the worker's
-- current-file index, batch status and cursor without writing file bytes here.
ALTER TABLE integration_connections
  DROP CONSTRAINT integration_connections_provider_check;
ALTER TABLE integration_connections
  ADD CONSTRAINT integration_connections_provider_check
  CHECK (provider IN ('BLING', 'NUVEMSHOP', 'NUVEM_PAGO', 'PAYROLL', 'GOOGLE_DRIVE'));

CREATE TABLE google_drive_sync_page_checkpoints (
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
  resource text NOT NULL CHECK (resource IN (
    'google_drive.files.01_perdas_operacionais',
    'google_drive.files.02_nuvem_pago/recebimentos_futuros',
    'google_drive.files.03_nuvem_envio'
  )),
  correlation_id uuid NOT NULL,
  page_number integer NOT NULL CHECK (page_number >= 1),
  start_page_token text NOT NULL CHECK (octet_length(start_page_token) <= 4096),
  start_page_token_hash text NOT NULL CHECK (start_page_token_hash ~ '^[0-9a-f]{64}$'),
  next_page_token text NOT NULL CHECK (octet_length(next_page_token) <= 4096),
  has_next boolean NOT NULL,
  file_count integer NOT NULL CHECK (file_count BETWEEN 0 AND 100),
  payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-f]{64}$'),
  payload jsonb NOT NULL CHECK (
    jsonb_typeof(payload) = 'array' AND
    jsonb_array_length(payload) = file_count AND
    octet_length(payload::text) <= 1048576
  ),
  checkpointed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (connection_id, resource, correlation_id, page_number),
  UNIQUE (connection_id, resource, correlation_id, start_page_token_hash),
  CHECK (has_next = (next_page_token <> ''))
);

COMMENT ON TABLE google_drive_sync_page_checkpoints IS
  'Bounded file metadata pages for one Drive scan. The page token and payload are checkpointed together and removed only after atomic snapshot publication; file contents are never stored here.';
COMMENT ON COLUMN google_drive_sync_page_checkpoints.start_page_token IS
  'Sensitive Google Drive pagination token. Read and write access is restricted to the worker runtime role.';
COMMENT ON COLUMN google_drive_sync_page_checkpoints.next_page_token IS
  'Sensitive Google Drive pagination token. Read and write access is restricted to the worker runtime role.';
COMMENT ON COLUMN google_drive_sync_page_checkpoints.payload IS
  'Potentially sensitive file metadata, bounded to 1 MiB per page. Access is restricted to the worker runtime role and rows are deleted after publication or when superseded.';

CREATE TABLE google_drive_file_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  folder_scope text NOT NULL CHECK (folder_scope IN (
    '01_perdas_operacionais',
    '02_nuvem_pago/recebimentos_futuros',
    '03_nuvem_envio'
  )),
  file_id text NOT NULL CHECK (length(file_id) BETWEEN 1 AND 256),
  payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-f]{64}$'),
  metadata jsonb NOT NULL CHECK (jsonb_typeof(metadata) = 'object'),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, folder_scope, file_id, payload_hash),
  UNIQUE (id, connection_id, folder_scope, file_id)
);

COMMENT ON TABLE google_drive_file_versions IS
  'Append-only, deduplicated metadata versions for approved Drive files. File bytes remain in secure temporary staging and are not stored in PostgreSQL.';

CREATE INDEX ix_google_drive_file_versions_file
  ON google_drive_file_versions(connection_id, folder_scope, file_id, created_at DESC);

CREATE TABLE google_drive_files (
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  folder_scope text NOT NULL CHECK (folder_scope IN (
    '01_perdas_operacionais',
    '02_nuvem_pago/recebimentos_futuros',
    '03_nuvem_envio'
  )),
  file_id text NOT NULL CHECK (length(file_id) BETWEEN 1 AND 256),
  current_version_id uuid NOT NULL,
  last_seen_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  last_seen_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (connection_id, folder_scope, file_id),
  FOREIGN KEY (current_version_id, connection_id, folder_scope, file_id)
    REFERENCES google_drive_file_versions(id, connection_id, folder_scope, file_id)
    ON DELETE RESTRICT
);

COMMENT ON TABLE google_drive_files IS
  'Current inventory pointer for a Drive file. Consumers use the completed batch in sync_cursors to expose a whole folder snapshot; files absent from a new snapshot are not deleted.';

CREATE INDEX ix_google_drive_files_snapshot
  ON google_drive_files(connection_id, folder_scope, last_seen_batch_id, file_id);
