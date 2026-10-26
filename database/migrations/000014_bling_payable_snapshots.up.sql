-- Publish only complete daily Bling payable snapshots to the dashboard.
-- Membership is append-only by batch so a failed refresh cannot retire the
-- last valid set of obligations.
CREATE TABLE payable_sync_snapshot_rows (
  sync_batch_id uuid NOT NULL REFERENCES integration_sync_batches(id) ON DELETE RESTRICT,
  payable_id uuid NOT NULL REFERENCES payables(id) ON DELETE RESTRICT,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  PRIMARY KEY (sync_batch_id, payable_id)
);

CREATE INDEX ix_payable_sync_snapshot_rows_payable
  ON payable_sync_snapshot_rows(payable_id, sync_batch_id);

COMMENT ON TABLE payable_sync_snapshot_rows IS
  'Records the payables included in a completed daily Bling snapshot; dashboard readers must use the latest successful NORMALIZED daily batch.';

-- Allow the worker to publish payable facts and lineage, while contact writes
-- are limited to a narrow, validated function and never expose document data.
CREATE OR REPLACE FUNCTION upsert_bling_supplier_contact(
  p_connection_id uuid,
  p_source_id text,
  p_name text,
  p_raw_record_id uuid
) RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
DECLARE
  resolved_id uuid;
  clean_name text := btrim(p_name);
BEGIN
  IF p_connection_id IS NULL OR p_raw_record_id IS NULL
     OR p_source_id IS NULL OR p_source_id !~ '^[1-9][0-9]{0,18}$'
     OR clean_name IS NULL OR clean_name = '' OR octet_length(clean_name) > 512 THEN
    RAISE EXCEPTION 'invalid Bling supplier identity';
  END IF;

  IF NOT EXISTS (
       SELECT 1 FROM public.integration_connections
        WHERE id = p_connection_id AND provider = 'BLING'
     ) OR NOT EXISTS (
       SELECT 1 FROM public.raw_records
        WHERE id = p_raw_record_id AND connection_id = p_connection_id
          AND source_system = 'BLING' AND source_entity = 'bling.contato'
          AND source_id = p_source_id AND is_current
     ) THEN
    RAISE EXCEPTION 'Bling supplier source does not match its connection';
  END IF;

  INSERT INTO public.contacts (
    connection_id, source_system, source_entity, source_id,
    kind, name, status, raw_record_id
  ) VALUES (
    p_connection_id, 'BLING', 'contatos', p_source_id,
    'SUPPLIER', clean_name, 'UNKNOWN', p_raw_record_id
  )
  ON CONFLICT (connection_id, source_system, source_entity, source_id)
  DO UPDATE SET
    kind = CASE
      WHEN contacts.kind = 'BOTH' THEN 'BOTH'
      WHEN contacts.kind = 'CUSTOMER' THEN 'BOTH'
      ELSE 'SUPPLIER'
    END,
    name = EXCLUDED.name,
    raw_record_id = EXCLUDED.raw_record_id,
    updated_at = clock_timestamp()
  RETURNING id INTO resolved_id;

  RETURN resolved_id;
END;
$$;

REVOKE ALL ON FUNCTION upsert_bling_supplier_contact(uuid, text, text, uuid) FROM PUBLIC;
