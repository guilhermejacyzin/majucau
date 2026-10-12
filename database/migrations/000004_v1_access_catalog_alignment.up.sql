-- Align the seeded access catalog with the approved version 1 matrix.
-- This catalogs permissions; the worker still must enforce them per command.
BEGIN;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
      FROM system_users AS u
      JOIN roles AS r ON r.id = u.role_id
     WHERE r.code IN ('FINANCE', 'ACCOUNTING', 'VIEWER')
  ) THEN
    RAISE EXCEPTION USING MESSAGE =
      'Cannot retire legacy roles while users still reference them; reassign each user explicitly first.';
  END IF;
END;
$$;

DELETE FROM role_permissions AS rp
USING roles AS r
WHERE rp.role_id = r.id
  AND r.code IN ('FINANCE', 'ACCOUNTING', 'VIEWER');

DELETE FROM roles
WHERE code IN ('FINANCE', 'ACCOUNTING', 'VIEWER');

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM pg_constraint
     WHERE conrelid = 'roles'::regclass
       AND conname = 'roles_v1_code_check'
  ) THEN
    EXECUTE 'ALTER TABLE roles ADD CONSTRAINT roles_v1_code_check CHECK (code IN (''ADMIN'', ''DIRECTOR''))';
  END IF;
END;
$$;

UPDATE permissions
   SET description = 'Sincronizar integrações e importar arquivos CSV manuais aprovados'
 WHERE code = 'INTEGRATION_SYNC';

INSERT INTO permissions (code, description)
VALUES ('INTEGRATION_CANCEL', 'Cancelar uma sincronização em andamento')
ON CONFLICT (code) DO UPDATE
  SET description = EXCLUDED.description;

-- Keep the director role limited to the approved version 1 capabilities.
DELETE FROM role_permissions AS rp
USING roles AS r, permissions AS p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.code = 'DIRECTOR'
  AND p.code NOT IN (
    'DASHBOARD_READ', 'INTEGRATION_READ', 'INTEGRATION_SYNC',
    'BACKUP_EXPORT', 'REPORT_EXPORT'
  );

-- Cancellation is reserved for ADMIN until Director cancellation is approved.
DELETE FROM role_permissions AS rp
USING roles AS r, permissions AS p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.code = 'DIRECTOR'
  AND p.code = 'INTEGRATION_CANCEL';

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
  FROM roles AS r
  CROSS JOIN permissions AS p
 WHERE r.code = 'ADMIN'
   AND p.code = 'INTEGRATION_CANCEL'
ON CONFLICT DO NOTHING;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM roles WHERE code NOT IN ('ADMIN', 'DIRECTOR')) THEN
    RAISE EXCEPTION 'The version 1 role catalog contains an unsupported role.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM role_permissions AS rp
    JOIN roles AS r ON r.id = rp.role_id
    JOIN permissions AS p ON p.id = rp.permission_id
    WHERE r.code = 'ADMIN' AND p.code = 'INTEGRATION_CANCEL'
  ) THEN
    RAISE EXCEPTION 'ADMIN must retain the integration cancellation capability.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM role_permissions AS rp
    JOIN roles AS r ON r.id = rp.role_id
    JOIN permissions AS p ON p.id = rp.permission_id
    WHERE r.code = 'DIRECTOR' AND p.code = 'INTEGRATION_CANCEL'
  ) THEN
    RAISE EXCEPTION 'DIRECTOR must not receive integration cancellation before approval.';
  END IF;
END;
$$;

COMMIT;
