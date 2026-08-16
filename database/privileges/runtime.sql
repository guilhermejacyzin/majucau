-- Execute as the schema owner after every migration.
-- The installer creates majucau_runtime as a non-owner LOGIN role with SCRAM.

REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO majucau_runtime;

REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO majucau_runtime;

-- RAW and audit are append-only for the runtime account. Approved state
-- transitions execute through the SECURITY DEFINER functions below.
REVOKE UPDATE, DELETE ON TABLE raw_records FROM majucau_runtime;
REVOKE UPDATE, DELETE ON TABLE audit_events FROM majucau_runtime;

GRANT EXECUTE ON FUNCTION retire_current_raw_record(uuid) TO majucau_runtime;
GRANT EXECUTE ON FUNCTION redact_raw_record(uuid, jsonb, uuid, text) TO majucau_runtime;
