-- Execute as the schema owner after every migration.
-- The installer creates majucau_runtime as a non-owner LOGIN role with SCRAM.

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO majucau_runtime;

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM majucau_runtime;

-- The worker currently reads only the integration and dashboard relations.
-- New modules must add the exact relations they use instead of inheriting
-- access to every financial and identity table.
GRANT SELECT ON TABLE
  integration_connections,
  integration_status,
  integration_sync_batches,
  bling_sync_page_checkpoints,
  google_drive_sync_page_checkpoints,
  google_drive_sync_folders,
  google_drive_file_versions,
  google_drive_files,
  google_drive_source_validation_runs,
  sync_cursors,
  raw_records,
  receipts,
  receivables,
  payables,
  payments,
  contacts
TO majucau_runtime;

-- Writes are restricted to the configuration, sync ledger, RAW ingestion,
-- cursors, imported receivable/receipt facts, and temporary page checkpoints.
-- The worker cannot delete published Drive inventory or mutate other domain tables.
GRANT INSERT, UPDATE ON TABLE
  integration_connections,
  integration_sync_batches,
  google_drive_source_validation_runs,
  sync_cursors,
  receipts,
  receivables
TO majucau_runtime;
GRANT INSERT ON TABLE raw_records, bling_sync_page_checkpoints TO majucau_runtime;
GRANT DELETE ON TABLE bling_sync_page_checkpoints TO majucau_runtime;
GRANT INSERT, DELETE ON TABLE google_drive_sync_page_checkpoints TO majucau_runtime;
GRANT INSERT, UPDATE, DELETE ON TABLE google_drive_sync_folders TO majucau_runtime;
GRANT INSERT ON TABLE google_drive_file_versions TO majucau_runtime;
GRANT INSERT, UPDATE ON TABLE google_drive_files TO majucau_runtime;

-- RAW and audit are append-only for the runtime account. The sync version
-- rollover is restricted to the SECURITY DEFINER function below.
REVOKE ALL ON TABLE audit_events FROM PUBLIC, majucau_runtime;

-- User-defined routines default to PUBLIC EXECUTE in PostgreSQL. Close that
-- path for functions and procedures, then grant only the RAW version-retirement
-- function and the sanitized Drive-credential audit writer to the worker.
REVOKE ALL ON ALL ROUTINES IN SCHEMA public FROM PUBLIC, majucau_runtime;
GRANT EXECUTE ON FUNCTION retire_current_raw_record(uuid) TO majucau_runtime;
GRANT EXECUTE ON FUNCTION audit_google_drive_credential_configured(text, uuid, boolean) TO majucau_runtime;

-- The schema uses UUID defaults and does not need runtime sequence access.
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM majucau_runtime;
