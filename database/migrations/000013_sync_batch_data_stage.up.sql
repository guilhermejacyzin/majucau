ALTER TABLE integration_sync_batches
  ADD COLUMN data_stage text NOT NULL DEFAULT 'UNVERIFIED'
    CHECK (data_stage IN ('UNVERIFIED', 'RAW_ONLY', 'NORMALIZED'));

COMMENT ON COLUMN integration_sync_batches.data_stage IS
  'Describes whether an import produced normalized domain records. RAW_ONLY batches must not confirm financial dashboard metrics.';

CREATE OR REPLACE VIEW integration_sync_logs AS
SELECT id, connection_id, resource, started_at, finished_at, status,
       records_read, records_created, records_updated, records_failed,
       retry_count, error_code, error_message_sanitized, correlation_id,
       data_stage
  FROM integration_sync_batches;
