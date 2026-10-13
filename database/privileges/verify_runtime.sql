-- Run as the disposable PostgreSQL schema owner after runtime.sql.
-- This script validates only privileges; it does not read or change business data.
DO $verify_runtime_privileges$
DECLARE
  runtime_oid oid;
  table_row record;
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
     has_schema_privilege(runtime_oid, 'public', 'CREATE') THEN
    RAISE EXCEPTION 'runtime schema privileges are incorrect';
  END IF;

  FOR table_row IN
    SELECT format('%I.%I', namespace.nspname, relation.relname) AS qualified_name
      FROM pg_class AS relation
      JOIN pg_namespace AS namespace ON namespace.oid = relation.relnamespace
     WHERE namespace.nspname = 'public'
       AND relation.relkind IN ('r', 'p', 'v', 'm', 'f')
  LOOP
    IF NOT has_table_privilege(runtime_oid, table_row.qualified_name, 'SELECT') THEN
      RAISE EXCEPTION 'runtime role cannot read required relation %', table_row.qualified_name;
    END IF;
  END LOOP;

  IF NOT has_table_privilege(runtime_oid, 'public.raw_records', 'INSERT') OR
     has_table_privilege(runtime_oid, 'public.raw_records', 'UPDATE') OR
     has_table_privilege(runtime_oid, 'public.raw_records', 'DELETE') OR
     NOT has_table_privilege(runtime_oid, 'public.audit_events', 'INSERT') OR
     has_table_privilege(runtime_oid, 'public.audit_events', 'UPDATE') OR
     has_table_privilege(runtime_oid, 'public.audit_events', 'DELETE') THEN
    RAISE EXCEPTION 'runtime RAW or audit append-only privileges are incorrect';
  END IF;

  IF NOT has_function_privilege(runtime_oid, 'public.retire_current_raw_record(uuid)', 'EXECUTE') OR
     NOT has_function_privilege(runtime_oid, 'public.redact_raw_record(uuid,jsonb,uuid,text)', 'EXECUTE') THEN
    RAISE EXCEPTION 'runtime role cannot execute approved RAW transition functions';
  END IF;

  IF EXISTS (
    SELECT 1
      FROM pg_proc AS function_entry
      CROSS JOIN LATERAL aclexplode(coalesce(function_entry.proacl, acldefault('f', function_entry.proowner))) AS acl
     WHERE function_entry.oid IN (
       'public.retire_current_raw_record(uuid)'::regprocedure,
       'public.redact_raw_record(uuid,jsonb,uuid,text)'::regprocedure
     )
       AND acl.grantee = 0
       AND acl.privilege_type = 'EXECUTE'
  ) THEN
    RAISE EXCEPTION 'approved RAW transition functions are executable by PUBLIC';
  END IF;
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
END
$verify_runtime_write_guards$;
ROLLBACK;

SELECT 'runtime database privileges: PASS' AS verification_result;
