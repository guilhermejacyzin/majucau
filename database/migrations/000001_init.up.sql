-- Majucau Financial Intelligence - PostgreSQL foundation
-- Migration policy: forward-only in production. Do not add secrets here.
-- Monetary values are exact and use numeric(19,4); percentages use numeric(20,8).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = clock_timestamp();
  RETURN NEW;
END;
$$;

-- ---------------------------------------------------------------------------
-- RBAC and local identity
-- ---------------------------------------------------------------------------

CREATE TABLE roles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE CHECK (code IN ('ADMIN', 'DIRECTOR', 'FINANCE', 'ACCOUNTING', 'VIEWER')),
  display_name text NOT NULL,
  description text NOT NULL DEFAULT '',
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE permissions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  description text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE role_permissions (
  role_id uuid NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id uuid NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE system_users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  windows_sid text NOT NULL UNIQUE,
  display_name text NOT NULL,
  role_id uuid NOT NULL REFERENCES roles(id),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX ix_system_users_role_active ON system_users(role_id, active);

INSERT INTO roles (code, display_name, description) VALUES
  ('ADMIN', 'Administrador', 'Configuração, credenciais, usuários, backup/restore e operações materiais'),
  ('DIRECTOR', 'Diretor', 'Visão executiva, cenários, aprovações e relatórios'),
  ('FINANCE', 'Financeiro', 'Tesouraria, integrações operacionais e conciliação'),
  ('ACCOUNTING', 'Contabilidade', 'Plano de contas, períodos, DRE/P&L e ajustes contábeis'),
  ('VIEWER', 'Visualizador', 'Leitura de telas e drill-down permitido')
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions (code, description) VALUES
  ('DASHBOARD_READ', 'Visualizar o dashboard executivo'),
  ('INTEGRATION_READ', 'Visualizar integrações e status'),
  ('CREDENTIAL_WRITE', 'Salvar, reconectar ou revogar credenciais'),
  ('INTEGRATION_SYNC', 'Iniciar ou cancelar sincronização'),
  ('FINANCE_WRITE', 'Criar ajustes financeiros autorizados'),
  ('ACCOUNTING_WRITE', 'Criar lançamentos e ajustes contábeis'),
  ('PERIOD_CLOSE', 'Fechar período contábil'),
  ('BACKUP_EXPORT', 'Exportar backup'),
  ('RESTORE', 'Restaurar backup'),
  ('UPDATE', 'Executar atualização do aplicativo'),
  ('USER_ADMIN', 'Administrar usuários e perfis')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.code = 'ADMIN'
ON CONFLICT DO NOTHING;

WITH role_permission_codes(role_code, permission_code) AS (
  VALUES
    ('DIRECTOR', 'DASHBOARD_READ'), ('DIRECTOR', 'INTEGRATION_READ'),
    ('DIRECTOR', 'INTEGRATION_SYNC'), ('DIRECTOR', 'BACKUP_EXPORT'),
    ('FINANCE', 'DASHBOARD_READ'), ('FINANCE', 'INTEGRATION_READ'),
    ('FINANCE', 'INTEGRATION_SYNC'), ('FINANCE', 'FINANCE_WRITE'),
    ('ACCOUNTING', 'DASHBOARD_READ'), ('ACCOUNTING', 'INTEGRATION_READ'),
    ('ACCOUNTING', 'ACCOUNTING_WRITE'), ('ACCOUNTING', 'PERIOD_CLOSE'),
    ('VIEWER', 'DASHBOARD_READ'), ('VIEWER', 'INTEGRATION_READ')
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
  FROM role_permission_codes m
  JOIN roles r ON r.code = m.role_code
  JOIN permissions p ON p.code = m.permission_code
ON CONFLICT DO NOTHING;

-- ---------------------------------------------------------------------------
-- Integrations, OAuth metadata, batches, cursors and events
-- ---------------------------------------------------------------------------

CREATE TABLE integration_connections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  provider text NOT NULL UNIQUE CHECK (provider IN ('BLING', 'NUVEMSHOP', 'NUVEM_PAGO', 'PAYROLL')),
  configured_by uuid REFERENCES system_users(id),
  status text NOT NULL DEFAULT 'NOT_CONFIGURED' CHECK (status IN (
    'NOT_CONFIGURED', 'AUTHORIZING', 'CONNECTED', 'TOKEN_EXPIRING', 'AUTH_ERROR',
    'SCHEMA_MISMATCH', 'SYNCING', 'STALE', 'PARTIALLY_AVAILABLE', 'UNAVAILABLE'
  )),
  external_account_id text,
  external_account_name text,
  client_id text,
  redirect_uri text,
  secret_ref text,
  secret_store_scope text NOT NULL DEFAULT 'DPAPI_CURRENT_USER'
    CHECK (secret_store_scope IN ('DPAPI_CURRENT_USER')),
  authorized_scopes text[] NOT NULL DEFAULT ARRAY[]::text[],
  authorized_at timestamptz,
  token_expires_at timestamptz,
  refresh_token_expires_at timestamptz,
  credentials_updated_at timestamptz,
  revoked_at timestamptz,
  last_test_at timestamptz,
  last_test_status text CHECK (last_test_status IS NULL OR last_test_status IN ('SUCCESS', 'FAILED')),
  last_success_at timestamptz,
  last_attempt_at timestamptz,
  last_error_code text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

COMMENT ON COLUMN integration_connections.secret_ref IS
  'Opaque reference to the worker vault. DPAPI current-user means the NT SERVICE\\MajucauWorker identity because only the worker may encrypt or decrypt credentials.';

CREATE TABLE sync_cursors (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE CASCADE,
  resource text NOT NULL,
  cursor_value text,
  watermark_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, resource)
);

CREATE TABLE integration_sync_batches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  resource text NOT NULL,
  started_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  finished_at timestamptz,
  status text NOT NULL DEFAULT 'RUNNING' CHECK (status IN ('RUNNING', 'SUCCESS', 'PARTIAL', 'FAILED', 'CANCELLED')),
  records_read integer NOT NULL DEFAULT 0 CHECK (records_read >= 0),
  records_created integer NOT NULL DEFAULT 0 CHECK (records_created >= 0),
  records_updated integer NOT NULL DEFAULT 0 CHECK (records_updated >= 0),
  records_failed integer NOT NULL DEFAULT 0 CHECK (records_failed >= 0),
  retry_count integer NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
  error_code text,
  error_message_sanitized text,
  correlation_id uuid NOT NULL DEFAULT gen_random_uuid(),
  CHECK ((status = 'RUNNING' AND finished_at IS NULL) OR (status <> 'RUNNING'))
);

CREATE TABLE integration_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  provider_event_id text,
  event_type text NOT NULL,
  occurred_at timestamptz,
  received_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-fA-F]{64}$'),
  processing_status text NOT NULL DEFAULT 'RECEIVED'
    CHECK (processing_status IN ('RECEIVED', 'PROCESSING', 'PROCESSED', 'FAILED', 'IGNORED')),
  raw_record_id uuid,
  UNIQUE (connection_id, provider_event_id)
);

CREATE UNIQUE INDEX ux_integration_events_hash
  ON integration_events(connection_id, event_type, payload_hash)
  WHERE provider_event_id IS NULL;

-- ---------------------------------------------------------------------------
-- Immutable RAW and audit
-- ---------------------------------------------------------------------------

CREATE TABLE raw_records (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system IN ('BLING', 'NUVEMSHOP', 'NUVEM_PAGO', 'PAYROLL', 'MANUAL_VALIDATED')),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  payload_hash text NOT NULL CHECK (payload_hash ~ '^[0-9a-fA-F]{64}$'),
  payload jsonb NOT NULL,
  source_updated_at timestamptz,
  imported_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  is_current boolean NOT NULL DEFAULT true,
  redacted_at timestamptz,
  redacted_by uuid,
  redaction_reason text,
  redaction_previous_hash text,
  redaction_payload_hash text,
  UNIQUE (connection_id, source_system, source_entity, source_id, payload_hash),
  CHECK ((redacted_at IS NULL AND redacted_by IS NULL AND redaction_reason IS NULL)
      OR (redacted_at IS NOT NULL AND redacted_by IS NOT NULL AND length(trim(redaction_reason)) > 0))
);

CREATE UNIQUE INDEX ux_raw_records_current
  ON raw_records(connection_id, source_system, source_entity, source_id)
  WHERE is_current;
CREATE INDEX ix_raw_records_batch ON raw_records(sync_batch_id);
CREATE INDEX ix_raw_records_source ON raw_records(source_system, source_entity, source_id);
CREATE INDEX ix_raw_records_imported_at ON raw_records(imported_at);

ALTER TABLE integration_events
  ADD CONSTRAINT fk_integration_events_raw
  FOREIGN KEY (raw_record_id) REFERENCES raw_records(id) ON DELETE RESTRICT;

CREATE TABLE audit_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id uuid REFERENCES system_users(id) ON DELETE RESTRICT,
  actor_windows_sid text,
  action_code text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  correlation_id uuid,
  before_sanitized jsonb,
  after_sanitized jsonb,
  reason text,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  CHECK (before_sanitized IS NULL OR jsonb_typeof(before_sanitized) = 'object'),
  CHECK (after_sanitized IS NULL OR jsonb_typeof(after_sanitized) = 'object')
);

CREATE INDEX ix_audit_events_entity ON audit_events(entity_type, entity_id, created_at DESC);
CREATE INDEX ix_audit_events_actor ON audit_events(actor_user_id, created_at DESC);

CREATE OR REPLACE FUNCTION prevent_raw_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF TG_OP = 'UPDATE'
     AND current_setting('app.raw_rollover', true) = 'approved'
     AND NEW.id = OLD.id
     AND NEW.is_current = false
     AND OLD.is_current = true
     AND NEW.payload = OLD.payload
     AND NEW.payload_hash = OLD.payload_hash
     AND NEW.source_id = OLD.source_id
     AND NEW.source_entity = OLD.source_entity
     AND NEW.connection_id = OLD.connection_id THEN
    RETURN NEW;
  END IF;
  IF current_setting('app.raw_redaction', true) IS DISTINCT FROM 'approved' THEN
    RAISE EXCEPTION 'raw_records is append-only; use the approved LGPD redaction procedure';
  END IF;
  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'raw_records cannot be deleted';
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_raw_records_no_mutation
BEFORE UPDATE OR DELETE ON raw_records
FOR EACH ROW EXECUTE FUNCTION prevent_raw_mutation();

CREATE OR REPLACE FUNCTION redact_raw_record(
  p_raw_record_id uuid,
  p_redacted_payload jsonb,
  p_actor_user_id uuid,
  p_reason text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  v_previous_hash text;
BEGIN
  IF p_redacted_payload IS NULL OR p_actor_user_id IS NULL OR length(trim(coalesce(p_reason, ''))) = 0 THEN
    RAISE EXCEPTION 'redaction requires payload, actor and reason';
  END IF;
  SELECT payload_hash INTO v_previous_hash FROM raw_records WHERE id = p_raw_record_id FOR UPDATE;
  IF v_previous_hash IS NULL THEN
    RAISE EXCEPTION 'raw record % not found', p_raw_record_id;
  END IF;
  PERFORM set_config('app.raw_redaction', 'approved', true);
  UPDATE raw_records
     SET payload = p_redacted_payload,
         redacted_at = clock_timestamp(),
         redacted_by = p_actor_user_id,
         redaction_reason = p_reason,
         redaction_previous_hash = v_previous_hash,
         redaction_payload_hash = encode(digest(p_redacted_payload::text, 'sha256'), 'hex')
   WHERE id = p_raw_record_id;
  INSERT INTO audit_events(actor_user_id, action_code, entity_type, entity_id, before_sanitized, after_sanitized, reason)
  VALUES (p_actor_user_id, 'RAW_REDACTION', 'raw_records', p_raw_record_id,
          jsonb_build_object('payload_hash', v_previous_hash),
          jsonb_build_object('payload_hash', encode(digest(p_redacted_payload::text, 'sha256'), 'hex')),
          p_reason);
END;
$$;

REVOKE ALL ON FUNCTION redact_raw_record(uuid, jsonb, uuid, text) FROM PUBLIC;

CREATE OR REPLACE FUNCTION retire_current_raw_record(p_raw_record_id uuid)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
  PERFORM set_config('app.raw_rollover', 'approved', true);
  UPDATE raw_records
     SET is_current = false
   WHERE id = p_raw_record_id AND is_current = true;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'current raw record % not found', p_raw_record_id;
  END IF;
END;
$$;

REVOKE ALL ON FUNCTION retire_current_raw_record(uuid) FROM PUBLIC;

-- ---------------------------------------------------------------------------
-- Rules, scenarios and calculation runs
-- ---------------------------------------------------------------------------

CREATE TABLE financial_rule_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_code text NOT NULL,
  version integer NOT NULL CHECK (version > 0),
  description text NOT NULL,
  parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'RETIRED')),
  approved_by uuid REFERENCES system_users(id) ON DELETE RESTRICT,
  approved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (rule_code, version),
  CHECK ((status = 'APPROVED' AND approved_by IS NOT NULL AND approved_at IS NOT NULL) OR status <> 'APPROVED')
);

CREATE TABLE scenarios (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE CHECK (code IN ('BASE', 'CONSERVATIVE', 'STRESS', 'OPTIMISTIC')),
  name text NOT NULL,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

INSERT INTO scenarios (code, name) VALUES
  ('BASE', 'Base'), ('CONSERVATIVE', 'Conservador'), ('STRESS', 'Stress'), ('OPTIMISTIC', 'Otimista')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE calculation_runs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_type text NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  scenario_id uuid REFERENCES scenarios(id) ON DELETE RESTRICT,
  reference_date date NOT NULL,
  started_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  finished_at timestamptz,
  status text NOT NULL DEFAULT 'RUNNING' CHECK (status IN ('RUNNING', 'SUCCESS', 'PARTIAL', 'FAILED', 'CANCELLED')),
  input_hash text NOT NULL CHECK (input_hash ~ '^[0-9a-fA-F]{64}$'),
  error_code text,
  UNIQUE (calculation_type, rule_version_id, scenario_id, reference_date, input_hash),
  CHECK ((status = 'RUNNING' AND finished_at IS NULL) OR status <> 'RUNNING')
);

-- ---------------------------------------------------------------------------
-- Normalized financial entities
-- ---------------------------------------------------------------------------

CREATE TABLE contacts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system IN ('BLING', 'NUVEMSHOP', 'NUVEM_PAGO', 'PAYROLL', 'MANUAL_VALIDATED')),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('CUSTOMER', 'SUPPLIER', 'BOTH', 'EMPLOYEE', 'OWNER')),
  name text NOT NULL,
  document_hash text,
  status text NOT NULL,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id)
);

CREATE TABLE financial_accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system = 'BLING'),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  name text NOT NULL,
  account_type text NOT NULL,
  status text NOT NULL,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id)
);

CREATE TABLE financial_balance_snapshots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  financial_account_id uuid NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
  reference_date date NOT NULL,
  reconciled_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  status text NOT NULL CHECK (status IN ('PENDING_RECONCILIATION', 'CONFIRMED', 'DIVERGENT')),
  calculated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, financial_account_id, reference_date, rule_version_id)
);

CREATE TABLE receivables (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system IN ('BLING', 'NUVEMSHOP', 'NUVEM_PAGO')),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  business_type text NOT NULL CHECK (business_type IN ('B2B', 'B2C', 'B2C_EXCLUDED_FROM_BLING_FUTURE', 'UNCLASSIFIED')),
  customer_id uuid REFERENCES contacts(id) ON DELETE RESTRICT,
  external_document text,
  issue_date date,
  competence_date date,
  due_date date,
  expected_receipt_date date,
  gross_amount numeric(19,4) NOT NULL CHECK (gross_amount >= 0),
  fee_amount numeric(19,4) NOT NULL DEFAULT 0 CHECK (fee_amount >= 0),
  net_amount numeric(19,4) CHECK (net_amount IS NULL OR net_amount >= 0),
  open_balance numeric(19,4) CHECK (open_balance IS NULL OR open_balance >= 0),
  currency_code char(3) NOT NULL DEFAULT 'BRL',
  status text NOT NULL,
  payment_method text,
  installment_number integer CHECK (installment_number IS NULL OR installment_number >= 1),
  installment_count integer CHECK (installment_count IS NULL OR installment_count >= 1),
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id),
  CHECK (installment_number IS NULL OR installment_count IS NULL OR installment_number <= installment_count)
);

CREATE TABLE receipts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system = 'BLING'),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  financial_account_id uuid REFERENCES financial_accounts(id) ON DELETE RESTRICT,
  receipt_date date NOT NULL,
  amount numeric(19,4) NOT NULL CHECK (amount >= 0),
  interest numeric(19,4) NOT NULL DEFAULT 0 CHECK (interest >= 0),
  penalty numeric(19,4) NOT NULL DEFAULT 0 CHECK (penalty >= 0),
  discount numeric(19,4) NOT NULL DEFAULT 0 CHECK (discount >= 0),
  currency_code char(3) NOT NULL DEFAULT 'BRL',
  status text NOT NULL,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id)
);

CREATE TABLE receipt_allocations (
  receivable_id uuid NOT NULL REFERENCES receivables(id) ON DELETE RESTRICT,
  receipt_id uuid NOT NULL REFERENCES receipts(id) ON DELETE RESTRICT,
  allocated_amount numeric(19,4) NOT NULL CHECK (allocated_amount > 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (receivable_id, receipt_id)
);

CREATE TABLE payables (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system = 'BLING'),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  supplier_id uuid REFERENCES contacts(id) ON DELETE RESTRICT,
  document text,
  category text,
  competence_date date,
  due_date date NOT NULL,
  payment_date date,
  original_amount numeric(19,4) NOT NULL CHECK (original_amount >= 0),
  interest numeric(19,4) NOT NULL DEFAULT 0 CHECK (interest >= 0),
  penalty numeric(19,4) NOT NULL DEFAULT 0 CHECK (penalty >= 0),
  discount numeric(19,4) NOT NULL DEFAULT 0 CHECK (discount >= 0),
  open_balance numeric(19,4) NOT NULL DEFAULT 0 CHECK (open_balance >= 0),
  currency_code char(3) NOT NULL DEFAULT 'BRL',
  status text NOT NULL,
  cost_center text,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id)
);

CREATE TABLE payments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  connection_id uuid NOT NULL REFERENCES integration_connections(id) ON DELETE RESTRICT,
  source_system text NOT NULL CHECK (source_system = 'BLING'),
  source_entity text NOT NULL,
  source_id text NOT NULL,
  financial_account_id uuid REFERENCES financial_accounts(id) ON DELETE RESTRICT,
  payment_date date NOT NULL,
  amount numeric(19,4) NOT NULL CHECK (amount >= 0),
  currency_code char(3) NOT NULL DEFAULT 'BRL',
  status text NOT NULL,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (connection_id, source_system, source_entity, source_id)
);

CREATE TABLE payment_allocations (
  payable_id uuid NOT NULL REFERENCES payables(id) ON DELETE RESTRICT,
  payment_id uuid NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
  allocated_amount numeric(19,4) NOT NULL CHECK (allocated_amount > 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (payable_id, payment_id)
);

CREATE TABLE financial_adjustments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  effective_date date NOT NULL,
  amount numeric(19,4) NOT NULL CHECK (amount >= 0),
  direction text NOT NULL CHECK (direction IN ('INFLOW', 'OUTFLOW')),
  reason text NOT NULL,
  authorized_by uuid NOT NULL REFERENCES system_users(id) ON DELETE RESTRICT,
  authorized_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  status text NOT NULL DEFAULT 'AUTHORIZED' CHECK (status IN ('AUTHORIZED', 'CANCELLED')),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX ix_receivables_due_date ON receivables(due_date, status);
CREATE INDEX ix_receivables_expected_date ON receivables(expected_receipt_date, status);
CREATE INDEX ix_payables_due_date ON payables(due_date, status);
CREATE INDEX ix_receipts_date ON receipts(receipt_date);
CREATE INDEX ix_payments_date ON payments(payment_date);

-- ---------------------------------------------------------------------------
-- Typed lineage/source links (real FKs; no polymorphic integrity)
-- ---------------------------------------------------------------------------

CREATE TABLE contact_sources (
  contact_id uuid NOT NULL REFERENCES contacts(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (contact_id, raw_record_id)
);

CREATE TABLE financial_account_sources (
  financial_account_id uuid NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (financial_account_id, raw_record_id)
);

CREATE TABLE receivable_sources (
  receivable_id uuid NOT NULL REFERENCES receivables(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (receivable_id, raw_record_id)
);

CREATE TABLE receipt_sources (
  receipt_id uuid NOT NULL REFERENCES receipts(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (receipt_id, raw_record_id)
);

CREATE TABLE payable_sources (
  payable_id uuid NOT NULL REFERENCES payables(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (payable_id, raw_record_id)
);

CREATE TABLE payment_sources (
  payment_id uuid NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (payment_id, raw_record_id)
);

CREATE TABLE financial_balance_snapshot_sources (
  snapshot_id uuid NOT NULL REFERENCES financial_balance_snapshots(id) ON DELETE RESTRICT,
  financial_account_id uuid NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
  raw_record_id uuid NOT NULL REFERENCES raw_records(id) ON DELETE RESTRICT,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  PRIMARY KEY (snapshot_id, financial_account_id, raw_record_id)
);

CREATE TABLE daily_balances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  scenario_id uuid NOT NULL REFERENCES scenarios(id) ON DELETE RESTRICT,
  balance_date date NOT NULL,
  opening_balance numeric(19,4) NOT NULL,
  inflows numeric(19,4) NOT NULL DEFAULT 0 CHECK (inflows >= 0),
  outflows numeric(19,4) NOT NULL DEFAULT 0 CHECK (outflows >= 0),
  adjustments numeric(19,4) NOT NULL DEFAULT 0,
  closing_balance numeric(19,4) NOT NULL,
  balance_type text NOT NULL CHECK (balance_type IN ('REALIZED', 'PROVISIONAL', 'PROJECTED')),
  calculated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (calculation_run_id, balance_date)
);

CREATE INDEX ix_daily_balances_date ON daily_balances(scenario_id, balance_date);

CREATE TABLE daily_balance_receivable_contributions (
  daily_balance_id uuid NOT NULL REFERENCES daily_balances(id) ON DELETE RESTRICT,
  receivable_id uuid NOT NULL REFERENCES receivables(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  PRIMARY KEY (daily_balance_id, receivable_id)
);

CREATE TABLE daily_balance_receipt_contributions (
  daily_balance_id uuid NOT NULL REFERENCES daily_balances(id) ON DELETE RESTRICT,
  receipt_id uuid NOT NULL REFERENCES receipts(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  PRIMARY KEY (daily_balance_id, receipt_id)
);

CREATE TABLE daily_balance_payable_contributions (
  daily_balance_id uuid NOT NULL REFERENCES daily_balances(id) ON DELETE RESTRICT,
  payable_id uuid NOT NULL REFERENCES payables(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  PRIMARY KEY (daily_balance_id, payable_id)
);

CREATE TABLE daily_balance_payment_contributions (
  daily_balance_id uuid NOT NULL REFERENCES daily_balances(id) ON DELETE RESTRICT,
  payment_id uuid NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  PRIMARY KEY (daily_balance_id, payment_id)
);

CREATE TABLE daily_balance_adjustment_contributions (
  daily_balance_id uuid NOT NULL REFERENCES daily_balances(id) ON DELETE RESTRICT,
  adjustment_id uuid NOT NULL REFERENCES financial_adjustments(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  PRIMARY KEY (daily_balance_id, adjustment_id)
);

CREATE VIEW record_lineage AS
SELECT 'contacts'::text AS target_entity, cs.contact_id AS target_id, cs.raw_record_id,
       cs.rule_version_id, cs.created_at
  FROM contact_sources cs
UNION ALL
SELECT 'financial_accounts', fas.financial_account_id, fas.raw_record_id,
       fas.rule_version_id, fas.created_at
  FROM financial_account_sources fas
UNION ALL
SELECT 'receivables', rs.receivable_id, rs.raw_record_id,
       rs.rule_version_id, rs.created_at
  FROM receivable_sources rs
UNION ALL
SELECT 'receipts', rs.receipt_id, rs.raw_record_id,
       rs.rule_version_id, rs.created_at
  FROM receipt_sources rs
UNION ALL
SELECT 'payables', ps.payable_id, ps.raw_record_id,
       ps.rule_version_id, ps.created_at
  FROM payable_sources ps
UNION ALL
SELECT 'payments', ps.payment_id, ps.raw_record_id,
       ps.rule_version_id, ps.created_at
  FROM payment_sources ps;

-- ---------------------------------------------------------------------------
-- Planning and accounting foundations
-- ---------------------------------------------------------------------------

CREATE TABLE forecast_versions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  scenario_id uuid NOT NULL REFERENCES scenarios(id) ON DELETE RESTRICT,
  assumptions jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_by uuid NOT NULL REFERENCES system_users(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  status text NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'FINAL', 'ARCHIVED')),
  UNIQUE (name, scenario_id)
);

CREATE TABLE forecast_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  forecast_version_id uuid NOT NULL REFERENCES forecast_versions(id) ON DELETE RESTRICT,
  account_or_category text NOT NULL,
  period date NOT NULL,
  amount numeric(19,4) NOT NULL,
  UNIQUE (forecast_version_id, account_or_category, period)
);

CREATE TABLE chart_of_accounts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  nature text NOT NULL CHECK (nature IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE', 'FINANCIAL')),
  parent_id uuid REFERENCES chart_of_accounts(id) ON DELETE RESTRICT,
  active boolean NOT NULL DEFAULT true
);

CREATE TABLE budget_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  chart_account_id uuid REFERENCES chart_of_accounts(id) ON DELETE RESTRICT,
  account_or_category text,
  period date NOT NULL,
  budget_amount numeric(19,4) NOT NULL,
  actual_amount numeric(19,4) NOT NULL DEFAULT 0,
  variance_amount numeric(19,4) NOT NULL DEFAULT 0,
  variance_percent numeric(20,8),
  favorability text NOT NULL DEFAULT 'NOT_APPLICABLE'
    CHECK (favorability IN ('FAVORABLE', 'UNFAVORABLE', 'NEUTRAL', 'NOT_APPLICABLE')),
  CHECK (chart_account_id IS NOT NULL OR account_or_category IS NOT NULL)
);

CREATE TABLE accounting_periods (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  start_date date NOT NULL,
  end_date date NOT NULL,
  status text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'REOPENED')),
  closed_at timestamptz,
  closed_by uuid REFERENCES system_users(id) ON DELETE RESTRICT,
  CHECK (end_date >= start_date),
  UNIQUE (start_date, end_date)
);

CREATE TABLE accounting_entries (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  period_id uuid NOT NULL REFERENCES accounting_periods(id) ON DELETE RESTRICT,
  entry_date date NOT NULL,
  description text NOT NULL,
  source_system text,
  source_id text,
  status text NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'POSTED', 'VOIDED')),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE accounting_entry_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id uuid NOT NULL REFERENCES accounting_entries(id) ON DELETE RESTRICT,
  chart_account_id uuid NOT NULL REFERENCES chart_of_accounts(id) ON DELETE RESTRICT,
  debit_amount numeric(19,4) NOT NULL DEFAULT 0 CHECK (debit_amount >= 0),
  credit_amount numeric(19,4) NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
  raw_record_id uuid REFERENCES raw_records(id) ON DELETE RESTRICT,
  CHECK (num_nonnulls(debit_amount, credit_amount) = 2),
  CHECK ((debit_amount > 0 AND credit_amount = 0) OR (credit_amount > 0 AND debit_amount = 0))
);

CREATE TABLE statement_snapshots (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  calculation_run_id uuid NOT NULL REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  statement_type text NOT NULL CHECK (statement_type IN ('DRE', 'PNL', 'EBITDA', 'BALANCE_SHEET', 'TRIAL_BALANCE', 'WORKING_CAPITAL')),
  period_start date NOT NULL,
  period_end date NOT NULL,
  scenario_id uuid REFERENCES scenarios(id) ON DELETE RESTRICT,
  status text NOT NULL CHECK (status IN ('CONFIRMED', 'PARTIALLY_CONFIRMED', 'PROVISIONAL', 'PROJECTED', 'DIVERGENT')),
  calculated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  CHECK (period_end >= period_start),
  UNIQUE (calculation_run_id, statement_type, period_start, period_end, scenario_id)
);

CREATE TABLE statement_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  snapshot_id uuid NOT NULL REFERENCES statement_snapshots(id) ON DELETE RESTRICT,
  line_code text NOT NULL,
  label text NOT NULL,
  amount numeric(19,4) NOT NULL,
  display_order integer NOT NULL CHECK (display_order >= 0),
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  UNIQUE (snapshot_id, line_code)
);

CREATE TABLE statement_line_contributions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  statement_line_id uuid NOT NULL REFERENCES statement_lines(id) ON DELETE RESTRICT,
  accounting_entry_line_id uuid REFERENCES accounting_entry_lines(id) ON DELETE RESTRICT,
  daily_balance_id uuid REFERENCES daily_balances(id) ON DELETE RESTRICT,
  contribution_amount numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  CHECK (num_nonnulls(accounting_entry_line_id, daily_balance_id) = 1)
);

CREATE INDEX ix_statement_lines_code ON statement_lines(line_code);

CREATE VIEW dre_snapshots AS
SELECT * FROM statement_snapshots WHERE statement_type = 'DRE';
CREATE VIEW pnl_snapshots AS
SELECT * FROM statement_snapshots WHERE statement_type = 'PNL';
CREATE VIEW ebitda_snapshots AS
SELECT * FROM statement_snapshots WHERE statement_type = 'EBITDA';
CREATE VIEW working_capital_snapshots AS
SELECT s.id, s.calculation_run_id, s.period_start, s.period_end, s.scenario_id, s.status,
       s.calculated_at,
       COALESCE(SUM(CASE WHEN l.line_code = 'CURRENT_ASSETS' THEN l.amount ELSE 0 END), 0)::numeric(19,4) AS current_assets,
       COALESCE(SUM(CASE WHEN l.line_code = 'CURRENT_LIABILITIES' THEN l.amount ELSE 0 END), 0)::numeric(19,4) AS current_liabilities,
       (COALESCE(SUM(CASE WHEN l.line_code = 'CURRENT_ASSETS' THEN l.amount ELSE 0 END), 0)
        - COALESCE(SUM(CASE WHEN l.line_code = 'CURRENT_LIABILITIES' THEN l.amount ELSE 0 END), 0))::numeric(19,4) AS working_capital
  FROM statement_snapshots s
  LEFT JOIN statement_lines l ON l.snapshot_id = s.id
 WHERE s.statement_type = 'WORKING_CAPITAL'
 GROUP BY s.id, s.calculation_run_id, s.period_start, s.period_end, s.scenario_id, s.status, s.calculated_at;

CREATE TABLE investment_parameters (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  effective_from date NOT NULL,
  minimum_reserve numeric(19,4) NOT NULL CHECK (minimum_reserve >= 0),
  eligible_profit_cutoff_months integer NOT NULL DEFAULT 2 CHECK (eligible_profit_cutoff_months >= 0),
  projection_horizon_days integer NOT NULL DEFAULT 60 CHECK (projection_horizon_days > 0),
  include_expected_receipts boolean NOT NULL DEFAULT false,
  approved_by uuid NOT NULL REFERENCES system_users(id) ON DELETE RESTRICT,
  approved_at timestamptz NOT NULL,
  UNIQUE (rule_version_id, effective_from)
);

CREATE TABLE investment_calculations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  statement_snapshot_id uuid REFERENCES statement_snapshots(id) ON DELETE RESTRICT,
  daily_balance_run_id uuid REFERENCES calculation_runs(id) ON DELETE RESTRICT,
  eligible_accumulated_profit numeric(19,4) NOT NULL,
  already_invested numeric(19,4) NOT NULL,
  available_profit numeric(19,4) NOT NULL,
  worst_week_cash numeric(19,4) NOT NULL,
  minimum_reserve numeric(19,4) NOT NULL,
  financial_limit numeric(19,4) NOT NULL,
  maximum_investment numeric(19,4) NOT NULL,
  rule_version_id uuid NOT NULL REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'PROVISIONAL' CHECK (status IN ('PROVISIONAL', 'CONFIRMED', 'DIVERGENT')),
  calculated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  CHECK (already_invested >= 0),
  CHECK (available_profit >= 0),
  CHECK (minimum_reserve >= 0),
  CHECK (maximum_investment >= 0)
);

-- ---------------------------------------------------------------------------
-- Alerts, quality and operational views
-- ---------------------------------------------------------------------------

CREATE TABLE alerts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  alert_type text NOT NULL CHECK (alert_type IN (
    'CASH_BELOW_MINIMUM_RESERVE', 'OVERDUE_PAYABLES', 'OVERDUE_RECEIVABLES',
    'LARGE_PAYMENT_DUE', 'EBITDA_RELEVANT_DROP', 'SYNC_ERROR',
    'ACCOUNTING_DIVERGENCE', 'STALE_DATA'
  )),
  severity text NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
  status text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED', 'DISMISSED')),
  message text NOT NULL,
  rule_version_id uuid REFERENCES financial_rule_versions(id) ON DELETE RESTRICT,
  reference_entity text,
  reference_id uuid,
  opened_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  resolved_at timestamptz,
  resolved_by uuid REFERENCES system_users(id) ON DELETE RESTRICT,
  resolution_note text
);

CREATE TABLE data_quality_issues (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_code text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  severity text NOT NULL CHECK (severity IN ('INFO', 'WARNING', 'CRITICAL')),
  status text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED', 'DISMISSED')),
  message text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  resolved_at timestamptz,
  resolved_by uuid REFERENCES system_users(id) ON DELETE RESTRICT
);

CREATE VIEW integration_sync_logs AS
SELECT id, connection_id, resource, started_at, finished_at, status,
       records_read, records_created, records_updated, records_failed,
       retry_count, error_code, error_message_sanitized, correlation_id
  FROM integration_sync_batches;

CREATE VIEW integration_status AS
SELECT c.id, c.provider, c.status, c.external_account_id, c.external_account_name,
       c.authorized_scopes, c.authorized_at, c.token_expires_at,
       c.last_test_at, c.last_test_status, c.last_success_at, c.last_attempt_at,
       c.last_error_code, c.updated_at
  FROM integration_connections c;

-- Operational indexes and updated_at triggers.
CREATE INDEX ix_sync_batches_connection_resource ON integration_sync_batches(connection_id, resource, started_at DESC);
CREATE INDEX ix_sync_cursors_updated ON sync_cursors(updated_at);
CREATE INDEX ix_integration_connections_status ON integration_connections(status);
CREATE INDEX ix_alerts_open ON alerts(status, severity, opened_at DESC) WHERE status IN ('OPEN', 'ACKNOWLEDGED');
CREATE INDEX ix_quality_open ON data_quality_issues(status, severity, created_at DESC) WHERE status IN ('OPEN', 'ACKNOWLEDGED');

CREATE TRIGGER trg_system_users_updated_at BEFORE UPDATE ON system_users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_integration_connections_updated_at BEFORE UPDATE ON integration_connections
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_sync_cursors_updated_at BEFORE UPDATE ON sync_cursors
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_contacts_updated_at BEFORE UPDATE ON contacts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_financial_accounts_updated_at BEFORE UPDATE ON financial_accounts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_receivables_updated_at BEFORE UPDATE ON receivables
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_payables_updated_at BEFORE UPDATE ON payables
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE raw_records IS 'Append-only external payloads. Synchronization never updates or deletes rows; use redact_raw_record for approved LGPD redaction.';
COMMENT ON VIEW integration_sync_logs IS 'Compatibility/reporting view over integration_sync_batches; no duplicated log source.';
COMMENT ON VIEW record_lineage IS 'Read-only union of typed lineage tables. Typed tables, not this view, provide referential integrity.';

CREATE OR REPLACE VIEW record_lineage AS
SELECT 'contacts'::text AS target_entity, cs.contact_id AS target_id, cs.raw_record_id,
       cs.rule_version_id, cs.created_at
  FROM contact_sources cs
UNION ALL
SELECT 'financial_accounts', fas.financial_account_id, fas.raw_record_id,
       fas.rule_version_id, fas.created_at
  FROM financial_account_sources fas
UNION ALL
SELECT 'receivables', rs.receivable_id, rs.raw_record_id,
       rs.rule_version_id, rs.created_at
  FROM receivable_sources rs
UNION ALL
SELECT 'receipts', rs.receipt_id, rs.raw_record_id,
       rs.rule_version_id, rs.created_at
  FROM receipt_sources rs
UNION ALL
SELECT 'payables', ps.payable_id, ps.raw_record_id,
       ps.rule_version_id, ps.created_at
  FROM payable_sources ps
UNION ALL
SELECT 'payments', ps.payment_id, ps.raw_record_id,
       ps.rule_version_id, ps.created_at
  FROM payment_sources ps
UNION ALL
SELECT 'financial_balance_snapshots', fbs.snapshot_id, fbs.raw_record_id,
       fbs.rule_version_id, NULL::timestamptz
  FROM financial_balance_snapshot_sources fbs
UNION ALL
SELECT 'statement_lines', slc.statement_line_id, ael.raw_record_id,
       slc.rule_version_id, NULL::timestamptz
  FROM statement_line_contributions slc
  JOIN accounting_entry_lines ael ON ael.id = slc.accounting_entry_line_id
 WHERE slc.accounting_entry_line_id IS NOT NULL
UNION ALL
SELECT 'daily_balances', dbc.daily_balance_id, rs.raw_record_id,
       dbc.rule_version_id, NULL::timestamptz
  FROM daily_balance_receivable_contributions dbc
  JOIN receivable_sources rs ON rs.receivable_id = dbc.receivable_id
UNION ALL
SELECT 'daily_balances', dbc.daily_balance_id, ps.raw_record_id,
       dbc.rule_version_id, NULL::timestamptz
  FROM daily_balance_payable_contributions dbc
  JOIN payable_sources ps ON ps.payable_id = dbc.payable_id;
