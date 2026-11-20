-- Allow the runtime to finalize only a DRE run that already has its immutable
-- snapshot. The function narrows the only mutable fields and transition.
CREATE FUNCTION finish_dre_calculation(p_run_id uuid, p_status text)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
  IF p_status IS NULL OR p_status NOT IN ('SUCCESS', 'PARTIAL') THEN
    RAISE EXCEPTION 'invalid DRE calculation status';
  END IF;

  UPDATE calculation_runs AS run
     SET status = p_status,
         finished_at = clock_timestamp()
   WHERE run.id = p_run_id
     AND run.calculation_type = 'DRE'
     AND run.status = 'RUNNING'
     AND EXISTS (
       SELECT 1 FROM statement_snapshots AS snapshot
        WHERE snapshot.calculation_run_id = run.id
          AND snapshot.statement_type = 'DRE'
     );
  IF NOT FOUND THEN
    RAISE EXCEPTION 'DRE run cannot be finalized';
  END IF;
END;
$$;

REVOKE ALL ON FUNCTION finish_dre_calculation(uuid, text) FROM PUBLIC;
