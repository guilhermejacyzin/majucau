-- Record protected Drive credential changes without exposing the credential or
-- folder IDs to PostgreSQL. Only the authenticated worker can call this writer.
CREATE OR REPLACE FUNCTION audit_google_drive_credential_configured(
  p_actor_windows_sid text,
  p_correlation_id uuid,
  p_replaced_existing boolean
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  IF p_actor_windows_sid IS NULL
     OR length(p_actor_windows_sid) > 184
     OR p_actor_windows_sid !~ '^S-[0-9]+(-[0-9]+){1,15}$'
     OR p_correlation_id IS NULL
     OR p_replaced_existing IS NULL THEN
    RAISE EXCEPTION 'invalid Google Drive credential audit context';
  END IF;

  INSERT INTO public.audit_events (
    actor_windows_sid,
    action_code,
    entity_type,
    correlation_id,
    before_sanitized,
    after_sanitized,
    reason
  ) VALUES (
    p_actor_windows_sid,
    'GOOGLE_DRIVE_CREDENTIAL_CONFIGURED',
    'google_drive_credentials',
    p_correlation_id,
    jsonb_build_object('configured', p_replaced_existing),
    jsonb_build_object('configured', true, 'scope', 'drive.readonly', 'folder_count', 3),
    'Configured the read-only Google Drive identity in the protected Windows vault.'
  );
END;
$$;

REVOKE ALL ON FUNCTION audit_google_drive_credential_configured(text, uuid, boolean) FROM PUBLIC;
