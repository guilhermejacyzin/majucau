-- Persist a bounded folder-by-folder traversal for the three approved Drive
-- trees. File pages remain staged until every descendant folder was read.
DO $drop_legacy_page_token_constraint$
DECLARE
  constraint_name text;
BEGIN
  SELECT conname INTO constraint_name
    FROM pg_constraint
   WHERE conrelid = 'google_drive_sync_page_checkpoints'::regclass
     AND contype = 'u'
     AND pg_get_constraintdef(oid) LIKE '%start_page_token_hash%';
  IF constraint_name IS NULL THEN
    RAISE EXCEPTION 'legacy Google Drive page-token constraint is missing';
  END IF;
  EXECUTE format('ALTER TABLE google_drive_sync_page_checkpoints DROP CONSTRAINT %I', constraint_name);
END
$drop_legacy_page_token_constraint$;

ALTER TABLE google_drive_sync_page_checkpoints
  ADD COLUMN listing_folder_id text,
  ADD COLUMN root_folder_id text,
  ADD CONSTRAINT google_drive_sync_page_checkpoints_listing_folder_id_check
    CHECK (listing_folder_id IS NULL OR listing_folder_id ~ '^[A-Za-z0-9_-]{1,256}$'),
  ADD CONSTRAINT google_drive_sync_page_checkpoints_root_folder_id_check
    CHECK (root_folder_id IS NULL OR root_folder_id ~ '^[A-Za-z0-9_-]{1,256}$'),
  ADD CONSTRAINT google_drive_sync_page_checkpoints_folder_pair_check
    CHECK ((listing_folder_id IS NULL) = (root_folder_id IS NULL));

CREATE UNIQUE INDEX uq_google_drive_sync_page_folder_start
  ON google_drive_sync_page_checkpoints(
    connection_id, resource, correlation_id, listing_folder_id, start_page_token_hash
  )
  WHERE listing_folder_id IS NOT NULL;

CREATE TABLE google_drive_sync_folders (
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
  resource text NOT NULL CHECK (resource IN (
    'google_drive.files.01_perdas_operacionais',
    'google_drive.files.02_nuvem_pago/recebimentos_futuros',
    'google_drive.files.03_nuvem_envio'
  )),
  correlation_id uuid NOT NULL,
  root_folder_id text NOT NULL CHECK (root_folder_id ~ '^[A-Za-z0-9_-]{1,256}$'),
  folder_id text NOT NULL CHECK (folder_id ~ '^[A-Za-z0-9_-]{1,256}$'),
  parent_folder_id text CHECK (parent_folder_id IS NULL OR parent_folder_id ~ '^[A-Za-z0-9_-]{1,256}$'),
  depth smallint NOT NULL CHECK (depth BETWEEN 0 AND 32),
  status text NOT NULL CHECK (status IN ('PENDING', 'ACTIVE', 'DONE')),
  next_page_token text NOT NULL DEFAULT '' CHECK (octet_length(next_page_token) <= 4096),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (connection_id, resource, correlation_id, folder_id),
  CHECK (
    (depth = 0 AND folder_id = root_folder_id AND parent_folder_id IS NULL) OR
    (depth > 0 AND folder_id <> root_folder_id AND parent_folder_id IS NOT NULL)
  ),
  CHECK ((status = 'PENDING' AND next_page_token = '') OR status <> 'PENDING')
);

CREATE INDEX ix_google_drive_sync_folders_pending
  ON google_drive_sync_folders(connection_id, resource, correlation_id, status, depth, folder_id);

COMMENT ON TABLE google_drive_sync_folders IS
  'Bounded per-folder traversal state for an approved Drive tree. Only descendants discovered directly from an already approved folder are added; page tokens are private worker state.';
COMMENT ON COLUMN google_drive_sync_folders.next_page_token IS
  'Sensitive Google Drive pagination token for this folder. Read and write access is restricted to the worker runtime role.';
COMMENT ON COLUMN google_drive_sync_page_checkpoints.listing_folder_id IS
  'The authorized folder whose direct children produced this metadata page; null is reserved for legacy checkpoints that are discarded before a new tree scan.';
