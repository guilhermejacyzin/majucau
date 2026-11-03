-- name: UpsertBlingFinancialAccountIdentity :exec
INSERT INTO bling_financial_account_identities (
  connection_id, sync_batch_id, source_id, description, raw_record_id
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (connection_id, source_id)
DO UPDATE SET
  sync_batch_id = EXCLUDED.sync_batch_id,
  description = EXCLUDED.description,
  raw_record_id = EXCLUDED.raw_record_id,
  updated_at = clock_timestamp();
