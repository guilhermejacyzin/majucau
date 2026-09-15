-- name: InsertRawRecord :one
WITH inserted AS (
  INSERT INTO raw_records (
    sync_batch_id,
    connection_id,
    source_system,
    source_entity,
    source_id,
    payload_hash,
    payload,
    source_updated_at,
    is_current
  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
  ON CONFLICT (connection_id, source_system, source_entity, source_id, payload_hash)
  DO NOTHING
  RETURNING id, sync_batch_id, connection_id, source_system, source_entity,
            source_id, payload_hash, payload, source_updated_at, imported_at, is_current
)
SELECT id, sync_batch_id, connection_id, source_system, source_entity,
       source_id, payload_hash, payload, source_updated_at, imported_at, is_current
FROM inserted
UNION ALL
SELECT r.id, r.sync_batch_id, r.connection_id, r.source_system, r.source_entity,
       r.source_id, r.payload_hash, r.payload, r.source_updated_at, r.imported_at, r.is_current
FROM raw_records r
WHERE r.connection_id = $2
  AND r.source_system = $3
  AND r.source_entity = $4
  AND r.source_id = $5
  AND r.payload_hash = $6
  AND NOT EXISTS (SELECT 1 FROM inserted)
LIMIT 1;

-- name: RetireCurrentRawRecord :exec
SELECT retire_current_raw_record($1);

-- name: GetCurrentRawRecord :one
SELECT id, sync_batch_id, connection_id, source_system, source_entity,
       source_id, payload_hash, payload, source_updated_at, imported_at, is_current
FROM raw_records
WHERE connection_id = $1
  AND source_system = $2
  AND source_entity = $3
  AND source_id = $4
  AND is_current = true;
