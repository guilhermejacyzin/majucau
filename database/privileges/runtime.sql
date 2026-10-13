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
  sync_cursors,
  raw_records,
  receipts,
  receivables,
  payables,
  payments,
  contacts
TO majucau_runtime;

-- Writes are restricted to the configuration, sync ledger, RAW ingestion,
-- cursors, and imported receivable/receipt facts used by the current worker.
-- No runtime code needs DELETE or direct writes to the other domain tables.
GRANT INSERT, UPDATE ON TABLE
  integration_connections,
  integration_sync_batches,
  sync_cursors,
  receipts,
  receivables
TO majucau_runtime;
GRANT INSERT ON TABLE raw_records TO majucau_runtime;

-- RAW and audit are append-only for the runtime account. The sync version
-- rollover is restricted to the SECURITY DEFINER function below.
REVOKE ALL ON TABLE audit_events FROM PUBLIC, majucau_runtime;

-- User-defined routines default to PUBLIC EXECUTE in PostgreSQL. Close that
-- path for functions and procedures, then grant only the RAW version-retirement
-- function to the worker.
REVOKE ALL ON ALL ROUTINES IN SCHEMA public FROM PUBLIC, majucau_runtime;
GRANT EXECUTE ON FUNCTION retire_current_raw_record(uuid) TO majucau_runtime;

-- The schema uses UUID defaults and does not need runtime sequence access.
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM majucau_runtime;
