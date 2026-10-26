-- name: UpsertBlingSupplierContact :one
SELECT upsert_bling_supplier_contact($1, $2, $3, $4)::uuid;

-- name: UpsertBlingPayable :one
INSERT INTO payables (
  connection_id, source_system, source_entity, source_id, supplier_id,
  document, category, competence_date, due_date, payment_date,
  original_amount, open_balance, currency_code, status, raw_record_id
) VALUES (
  $1, 'BLING', 'bling.conta_pagar', $2, $3,
  $4, $5, $6, $7, $8,
  $9, $10, 'BRL', $11, $12
)
ON CONFLICT (connection_id, source_system, source_entity, source_id)
DO UPDATE SET
  supplier_id = EXCLUDED.supplier_id,
  document = EXCLUDED.document,
  category = EXCLUDED.category,
  competence_date = EXCLUDED.competence_date,
  due_date = EXCLUDED.due_date,
  payment_date = EXCLUDED.payment_date,
  original_amount = EXCLUDED.original_amount,
  open_balance = EXCLUDED.open_balance,
  currency_code = EXCLUDED.currency_code,
  status = EXCLUDED.status,
  raw_record_id = EXCLUDED.raw_record_id,
  updated_at = clock_timestamp()
RETURNING id;

-- name: InsertBlingPayableSource :exec
INSERT INTO payable_sources (payable_id, raw_record_id)
VALUES ($1, $2)
ON CONFLICT (payable_id, raw_record_id) DO NOTHING;

-- name: InsertPayableSyncSnapshotRow :execrows
INSERT INTO payable_sync_snapshot_rows (sync_batch_id, payable_id)
VALUES ($1, $2)
ON CONFLICT (sync_batch_id, payable_id) DO NOTHING;

-- name: MarkIntegrationSyncBatchNormalized :execrows
UPDATE integration_sync_batches
SET data_stage = 'NORMALIZED'
WHERE id = $1 AND status = 'RUNNING' AND data_stage = 'RAW_ONLY';
