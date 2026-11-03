-- name: UpsertBlingPaymentMethodTerms :exec
INSERT INTO bling_payment_method_terms (
  connection_id, sync_batch_id, source_id, destination,
  fee_rate, fee_rate_known, fixed_fee, fixed_fee_known,
  settlement_days, settlement_known, use_business_days, business_days_known,
  raw_record_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (connection_id, source_id)
DO UPDATE SET
  sync_batch_id = EXCLUDED.sync_batch_id,
  destination = EXCLUDED.destination,
  fee_rate = EXCLUDED.fee_rate,
  fee_rate_known = EXCLUDED.fee_rate_known,
  fixed_fee = EXCLUDED.fixed_fee,
  fixed_fee_known = EXCLUDED.fixed_fee_known,
  settlement_days = EXCLUDED.settlement_days,
  settlement_known = EXCLUDED.settlement_known,
  use_business_days = EXCLUDED.use_business_days,
  business_days_known = EXCLUDED.business_days_known,
  raw_record_id = EXCLUDED.raw_record_id,
  updated_at = clock_timestamp();
