-- The content hash is evidence about a payload, not an identity constraint.
-- A provider may legitimately return an earlier payload again after a value
-- changes and later returns to the prior value. Keep each immutable version;
-- the separate partial unique index still permits only one current row per
-- source identity.
ALTER TABLE raw_records
  DROP CONSTRAINT raw_records_connection_id_source_system_source_entity_sourc_key;

CREATE INDEX ix_raw_records_source_version_hash
  ON raw_records(connection_id, source_system, source_entity, source_id, payload_hash);

COMMENT ON INDEX ix_raw_records_source_version_hash IS
  'Non-unique lookup index: append-only RAW history may contain recurring payload hashes as distinct imported versions.';
