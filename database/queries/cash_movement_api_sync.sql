-- name: UpsertBlingCashMovement :exec
INSERT INTO bling_cash_movements (
  connection_id,
  sync_batch_id,
  source_id,
  account_source_id,
  account_name,
  direction,
  amount,
  movement_date,
  source_status,
  origin_type,
  raw_record_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (connection_id, source_id)
DO UPDATE SET
  sync_batch_id = EXCLUDED.sync_batch_id,
  account_source_id = EXCLUDED.account_source_id,
  account_name = EXCLUDED.account_name,
  direction = EXCLUDED.direction,
  amount = EXCLUDED.amount,
  movement_date = EXCLUDED.movement_date,
  source_status = EXCLUDED.source_status,
  origin_type = EXCLUDED.origin_type,
  raw_record_id = EXCLUDED.raw_record_id,
  updated_at = clock_timestamp();
