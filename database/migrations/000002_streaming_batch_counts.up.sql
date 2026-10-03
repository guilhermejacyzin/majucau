BEGIN;

DROP VIEW integration_sync_logs;

ALTER TABLE integration_sync_batches
  ALTER COLUMN records_read TYPE bigint,
  ALTER COLUMN records_created TYPE bigint,
  ALTER COLUMN records_updated TYPE bigint,
  ALTER COLUMN records_failed TYPE bigint;

CREATE VIEW integration_sync_logs AS
SELECT id, connection_id, resource, started_at, finished_at, status,
       records_read, records_created, records_updated, records_failed,
       retry_count, error_code, error_message_sanitized, correlation_id
  FROM integration_sync_batches;

COMMENT ON VIEW integration_sync_logs IS 'Compatibility/reporting view over integration_sync_batches; no duplicated log source.';

COMMIT;
