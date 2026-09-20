-- name: InsertIntegrationSyncBatch :one
INSERT INTO integration_sync_batches (connection_id, resource)
VALUES ($1, $2)
RETURNING id, connection_id, resource, started_at, finished_at, status,
          records_read, records_created, records_updated, records_failed,
          retry_count, error_code, error_message_sanitized, correlation_id;

-- name: GetIntegrationConnectionByProvider :one
SELECT id, provider, configured_by, status, external_account_id,
       external_account_name, client_id, redirect_uri, secret_ref,
       secret_store_scope, authorized_scopes, authorized_at, token_expires_at,
       refresh_token_expires_at, credentials_updated_at, revoked_at,
       last_test_at, last_test_status, last_success_at, last_attempt_at,
       last_error_code, created_at, updated_at
FROM integration_connections
WHERE provider = $1;

-- name: FinishIntegrationSyncBatch :exec
UPDATE integration_sync_batches
SET finished_at = clock_timestamp(),
    status = $2,
    records_read = $3,
    records_created = $4,
    records_updated = $5,
    records_failed = $6,
    error_code = $7,
    error_message_sanitized = $8
WHERE id = $1;

-- name: UpsertReceipt :one
INSERT INTO receipts (
  connection_id,
  source_system,
  source_entity,
  source_id,
  receipt_date,
  amount,
  currency_code,
  status,
  raw_record_id
) VALUES ($1, 'BLING', $2, $3, $4, $5, 'BRL', $6, $7)
ON CONFLICT (connection_id, source_system, source_entity, source_id)
DO UPDATE SET
  receipt_date = EXCLUDED.receipt_date,
  amount = EXCLUDED.amount,
  currency_code = EXCLUDED.currency_code,
  status = EXCLUDED.status,
  raw_record_id = EXCLUDED.raw_record_id
RETURNING id, connection_id, source_system, source_entity, source_id,
          financial_account_id, receipt_date, amount, interest, penalty,
          discount, currency_code, status, raw_record_id, created_at;

-- name: ReceiptExists :one
SELECT EXISTS (
  SELECT 1
  FROM receipts
  WHERE connection_id = $1
    AND source_system = 'BLING'
    AND source_entity = $2
    AND source_id = $3
);
