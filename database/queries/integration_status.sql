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

-- name: MarkBlingAuthorizing :exec
UPDATE integration_connections
SET status = 'AUTHORIZING',
    last_attempt_at = clock_timestamp(),
    last_error_code = NULL,
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

-- name: CompleteBlingOAuth :exec
UPDATE integration_connections
SET status = 'CONNECTED',
    authorized_scopes = $1,
    authorized_at = clock_timestamp(),
    token_expires_at = $2,
    refresh_token_expires_at = $3,
    revoked_at = NULL,
    last_error_code = NULL,
    last_attempt_at = clock_timestamp(),
    last_success_at = clock_timestamp(),
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

-- name: MarkBlingOAuthError :exec
UPDATE integration_connections
SET status = 'AUTH_ERROR',
    last_attempt_at = clock_timestamp(),
    last_error_code = $1,
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

-- name: RecordBlingTestSuccess :exec
UPDATE integration_connections
SET status = 'CONNECTED',
    last_test_at = clock_timestamp(),
    last_test_status = 'SUCCESS',
    last_success_at = clock_timestamp(),
    last_attempt_at = clock_timestamp(),
    last_error_code = NULL,
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

-- name: RecordBlingTestFailure :exec
UPDATE integration_connections
SET last_test_at = clock_timestamp(),
    last_test_status = 'FAILED',
    last_attempt_at = clock_timestamp(),
    last_error_code = $1,
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

-- name: MarkBlingDisconnected :exec
UPDATE integration_connections
SET status = 'NOT_CONFIGURED',
    external_account_id = NULL,
    external_account_name = NULL,
    authorized_scopes = ARRAY[]::text[],
    authorized_at = NULL,
    token_expires_at = NULL,
    refresh_token_expires_at = NULL,
    revoked_at = clock_timestamp(),
    last_error_code = NULL,
    last_attempt_at = clock_timestamp(),
    updated_at = clock_timestamp()
WHERE provider = 'BLING';

