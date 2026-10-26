-- Run as the disposable PostgreSQL schema owner after runtime.sql.
-- This script validates privileges only and cannot change business data.
DO $verify_runtime_privileges$
DECLARE
  runtime_oid oid;
  relation_row record;
  function_row record;
  sequence_row record;
  expected_select boolean;
  expected_insert boolean;
  expected_update boolean;
  expected_delete boolean;
BEGIN
  SELECT oid INTO runtime_oid
    FROM pg_roles
   WHERE rolname = 'majucau_runtime';

  IF runtime_oid IS NULL THEN
    RAISE EXCEPTION 'runtime role is missing';
  END IF;

  IF EXISTS (
    SELECT 1
      FROM pg_roles
     WHERE oid = runtime_oid
       AND (NOT rolcanlogin OR rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)
  ) THEN
    RAISE EXCEPTION 'runtime role has unsafe role attributes';
  END IF;

  IF EXISTS (SELECT 1 FROM pg_auth_members WHERE member = runtime_oid) THEN
    RAISE EXCEPTION 'runtime role inherits privileges from another role';
  END IF;

  IF NOT has_schema_privilege(runtime_oid, 'public', 'USAGE') OR
     has_schema_privilege(runtime_oid, 'public', 'CREATE') OR
     (SELECT nspowner = runtime_oid FROM pg_namespace WHERE nspname = 'public') THEN
    RAISE EXCEPTION 'runtime schema privileges are incorrect';
  END IF;

  FOR relation_row IN
    SELECT format('%I.%I', namespace.nspname, relation.relname) AS qualified_name,
           relation.relname,
           relation.relowner
      FROM pg_class AS relation
      JOIN pg_namespace AS namespace ON namespace.oid = relation.relnamespace
     WHERE namespace.nspname = 'public'
       AND relation.relkind IN ('r', 'p', 'v', 'm', 'f')
  LOOP
    expected_select := relation_row.qualified_name = ANY (ARRAY[
      'public.integration_connections',
      'public.integration_status',
      'public.integration_sync_batches',
      'public.bling_sync_page_checkpoints',
      'public.google_drive_sync_page_checkpoints',
      'public.google_drive_sync_folders',
      'public.google_drive_file_versions',
      'public.google_drive_files',
      'public.google_drive_source_validation_runs',
      'public.sync_cursors',
      'public.raw_records',
      'public.payable_sync_snapshot_rows',
      'public.receipts',
      'public.receivables',
      'public.payables',
      'public.payments',
      'public.contacts'
    ]::text[]);
    expected_insert := relation_row.qualified_name = ANY (ARRAY[
      'public.integration_connections',
      'public.integration_sync_batches',
      'public.google_drive_source_validation_runs',
      'public.bling_sync_page_checkpoints',
      'public.google_drive_sync_page_checkpoints',
      'public.google_drive_sync_folders',
      'public.google_drive_file_versions',
      'public.google_drive_files',
      'public.sync_cursors',
      'public.raw_records',
      'public.payable_sync_snapshot_rows',
      'public.payable_sources',
      'public.payables',
      'public.receipts',
      'public.receivables'
    ]::text[]);
    expected_update := relation_row.qualified_name = ANY (ARRAY[
      'public.integration_connections',
      'public.integration_sync_batches',
      'public.sync_cursors',
      'public.google_drive_sync_folders',
      'public.google_drive_files',
      'public.google_drive_source_validation_runs',
      'public.payables',
      'public.receipts',
      'public.receivables'
    ]::text[]);
    expected_delete := relation_row.qualified_name = ANY (ARRAY[
      'public.bling_sync_page_checkpoints',
      'public.google_drive_sync_page_checkpoints',
      'public.google_drive_sync_folders'
    ]::text[]);

    IF relation_row.relowner = runtime_oid OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'SELECT') IS DISTINCT FROM expected_select OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'INSERT') IS DISTINCT FROM expected_insert OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'UPDATE') IS DISTINCT FROM expected_update OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'DELETE') IS DISTINCT FROM expected_delete OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'TRUNCATE') OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'REFERENCES') OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'TRIGGER') OR
       has_table_privilege(runtime_oid, relation_row.qualified_name, 'MAINTAIN') OR
       has_any_column_privilege(runtime_oid, relation_row.qualified_name, 'SELECT') IS DISTINCT FROM expected_select OR
       has_any_column_privilege(runtime_oid, relation_row.qualified_name, 'INSERT') IS DISTINCT FROM expected_insert OR
       has_any_column_privilege(runtime_oid, relation_row.qualified_name, 'UPDATE') IS DISTINCT FROM expected_update OR
       has_any_column_privilege(runtime_oid, relation_row.qualified_name, 'REFERENCES') THEN
      RAISE EXCEPTION 'runtime role has unexpected relation privileges on %', relation_row.qualified_name;
    END IF;
  END LOOP;

  FOR function_row IN
    SELECT function_entry.oid, function_entry.proowner
      FROM pg_proc AS function_entry
      JOIN pg_namespace AS namespace ON namespace.oid = function_entry.pronamespace
     WHERE namespace.nspname = 'public'
       AND function_entry.prokind IN ('f', 'p', 'a', 'w')
  LOOP
    IF function_row.proowner = runtime_oid OR
       has_function_privilege(runtime_oid, function_row.oid, 'EXECUTE') IS DISTINCT FROM (
         function_row.oid = ANY (ARRAY[
           'public.retire_current_raw_record(uuid)'::regprocedure,
           'public.audit_google_drive_credential_configured(text,uuid,boolean)'::regprocedure,
           'public.upsert_bling_supplier_contact(uuid,text,text,uuid)'::regprocedure
         ])
       ) THEN
      RAISE EXCEPTION 'runtime role has unexpected function privileges on %', function_row.oid::regprocedure;
    END IF;
  END LOOP;

  IF EXISTS (
    SELECT 1
      FROM pg_proc AS function_entry
      JOIN pg_namespace AS namespace ON namespace.oid = function_entry.pronamespace
      CROSS JOIN LATERAL aclexplode(coalesce(function_entry.proacl, acldefault('f', function_entry.proowner))) AS acl
     WHERE namespace.nspname = 'public'
       AND function_entry.prokind IN ('f', 'p', 'a', 'w')
       AND acl.grantee = 0
       AND acl.privilege_type = 'EXECUTE'
  ) THEN
    RAISE EXCEPTION 'a public-schema routine is executable by PUBLIC';
  END IF;

  FOR sequence_row IN
    SELECT relation.oid, format('%I.%I', namespace.nspname, relation.relname) AS qualified_name
      FROM pg_class AS relation
      JOIN pg_namespace AS namespace ON namespace.oid = relation.relnamespace
     WHERE namespace.nspname = 'public'
       AND relation.relkind = 'S'
  LOOP
    IF has_sequence_privilege(runtime_oid, sequence_row.oid, 'USAGE') OR
       has_sequence_privilege(runtime_oid, sequence_row.oid, 'SELECT') OR
       has_sequence_privilege(runtime_oid, sequence_row.oid, 'UPDATE') THEN
      RAISE EXCEPTION 'runtime role has unexpected sequence privileges on %', sequence_row.qualified_name;
    END IF;
  END LOOP;
END
$verify_runtime_privileges$;

-- Confirm direct writes fail under the runtime role. WHERE false ensures this
-- check cannot alter rows even if a grant is accidentally added.
BEGIN;
SET LOCAL ROLE majucau_runtime;
DO $verify_runtime_write_guards$
BEGIN
  BEGIN
    UPDATE public.raw_records SET is_current = false WHERE false;
    RAISE EXCEPTION 'runtime role unexpectedly updated RAW records';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;

  BEGIN
    DELETE FROM public.audit_events WHERE false;
    RAISE EXCEPTION 'runtime role unexpectedly deleted audit records';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;

  BEGIN
    INSERT INTO public.audit_events (action_code, entity_type, entity_id, reason)
    SELECT 'VERIFY_NOOP', 'verification', gen_random_uuid(), 'permission check only'
     WHERE false;
    RAISE EXCEPTION 'runtime role unexpectedly inserted audit records directly';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$verify_runtime_write_guards$;
ROLLBACK;

-- The runtime cannot write audit_events directly, but can call this one narrow
-- writer. The event must contain only the approved sanitized summary.
BEGIN;
SET LOCAL ROLE majucau_runtime;
SELECT public.audit_google_drive_credential_configured(
  'S-1-5-21-1-2-3-1001',
  'b36dd2ab-2954-4fe9-95cc-f2d55595b8de'::uuid,
  false
);
RESET ROLE;
DO $verify_drive_credential_audit$
BEGIN
  IF NOT EXISTS (
    SELECT 1
      FROM public.audit_events
     WHERE correlation_id = 'b36dd2ab-2954-4fe9-95cc-f2d55595b8de'::uuid
       AND actor_windows_sid = 'S-1-5-21-1-2-3-1001'
       AND action_code = 'GOOGLE_DRIVE_CREDENTIAL_CONFIGURED'
       AND entity_type = 'google_drive_credentials'
       AND before_sanitized = '{"configured": false}'::jsonb
       AND after_sanitized = '{"configured": true, "scope": "drive.readonly", "folder_count": 3}'::jsonb
       AND reason = 'Configured the read-only Google Drive identity in the protected Windows vault.'
  ) THEN
    RAISE EXCEPTION 'Google Drive credential audit event is missing or not sanitized';
  END IF;
END
$verify_drive_credential_audit$;
ROLLBACK;

SELECT 'runtime database privileges: PASS' AS verification_result;
