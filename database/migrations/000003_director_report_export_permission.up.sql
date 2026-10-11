BEGIN;

INSERT INTO permissions (code, description)
VALUES ('REPORT_EXPORT', 'Exportar relatórios aprovados')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles AS r
CROSS JOIN permissions AS p
WHERE r.code IN ('ADMIN', 'DIRECTOR')
  AND p.code = 'REPORT_EXPORT'
ON CONFLICT DO NOTHING;

COMMIT;
