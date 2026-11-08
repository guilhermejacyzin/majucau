-- name: InsertBlingReconciledCashMovementSnapshotRow :exec
INSERT INTO bling_cash_movement_reconciled_snapshot_rows (
  sync_batch_id,
  connection_id,
  source_id
) VALUES ($1, $2, $3)
ON CONFLICT (sync_batch_id, source_id) DO NOTHING;
