-- name: UpsertBlingCustomerContact :one
SELECT upsert_bling_customer_contact($1, $2, $3, $4)::uuid;

-- name: UpsertBlingReceivable :one
INSERT INTO receivables (
  connection_id, source_system, source_entity, source_id, business_type,
  customer_id, external_document, issue_date, competence_date, due_date,
  gross_amount, open_balance, currency_code, status, raw_record_id,
  source_payment_method_id, source_financial_account_id, source_accounting_account_id
) VALUES (
  $1, 'BLING', 'bling.conta_receber', $2, 'UNCLASSIFIED',
  $3, $4, $5, $6, $7,
  $8, $9, 'BRL', $10, $11, $12, $13, $14
)
ON CONFLICT (connection_id, source_system, source_entity, source_id)
DO UPDATE SET
  business_type = 'UNCLASSIFIED',
  customer_id = EXCLUDED.customer_id,
  external_document = EXCLUDED.external_document,
  issue_date = EXCLUDED.issue_date,
  competence_date = EXCLUDED.competence_date,
  due_date = EXCLUDED.due_date,
  gross_amount = EXCLUDED.gross_amount,
  open_balance = EXCLUDED.open_balance,
  currency_code = EXCLUDED.currency_code,
  status = EXCLUDED.status,
  raw_record_id = EXCLUDED.raw_record_id,
  source_payment_method_id = EXCLUDED.source_payment_method_id,
  source_financial_account_id = EXCLUDED.source_financial_account_id,
  source_accounting_account_id = EXCLUDED.source_accounting_account_id,
  updated_at = clock_timestamp()
RETURNING id;

-- name: InsertBlingReceivableSource :exec
INSERT INTO receivable_sources (receivable_id, raw_record_id)
VALUES ($1, $2)
ON CONFLICT (receivable_id, raw_record_id) DO NOTHING;

-- name: InsertReceivableSyncSnapshotRow :execrows
INSERT INTO receivable_sync_snapshot_rows (sync_batch_id, receivable_id)
VALUES ($1, $2)
ON CONFLICT (sync_batch_id, receivable_id) DO NOTHING;
