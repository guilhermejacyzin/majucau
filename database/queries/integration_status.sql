-- name: ListIntegrationStatus :many
SELECT
  id,
  provider,
  status,
  external_account_id,
  external_account_name,
  authorized_scopes,
  authorized_at,
  token_expires_at,
  last_test_at,
  last_test_status,
  last_success_at,
  last_attempt_at,
  last_error_code,
  updated_at
FROM integration_status
ORDER BY provider;

-- name: GetIntegrationStatus :one
SELECT
  id,
  provider,
  status,
  external_account_id,
  external_account_name,
  authorized_scopes,
  authorized_at,
  token_expires_at,
  last_test_at,
  last_test_status,
  last_success_at,
  last_attempt_at,
  last_error_code,
  updated_at
FROM integration_status
WHERE id = $1;

-- name: MarkIntegrationRevoked :exec
UPDATE integration_connections
SET status = 'AUTH_ERROR',
    revoked_at = clock_timestamp(),
    credentials_updated_at = clock_timestamp(),
    updated_at = clock_timestamp(),
    last_error_code = 'CREDENTIALS_REVOKED'
WHERE id = $1;
