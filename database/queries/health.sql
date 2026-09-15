-- name: Health :one
SELECT
  current_database()::text AS database_name,
  current_user::text AS database_user,
  clock_timestamp() AS checked_at,
  (NOT pg_is_in_recovery()) AS writable;

-- name: SchemaHealth :one
SELECT
  to_regclass('public.integration_connections') IS NOT NULL AS has_integrations,
  to_regclass('public.raw_records') IS NOT NULL AS has_raw,
  to_regclass('public.audit_events') IS NOT NULL AS has_audit,
  to_regclass('public.daily_balances') IS NOT NULL AS has_daily_balances;
