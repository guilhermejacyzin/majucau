# Database foundation

This directory is the SQL-first persistence boundary for Majucau Financial Intelligence.

## Layout

- `migrations/*.up.sql`: ordered, forward-only schema changes. `000001_init.up.sql` creates the schema; `000002_streaming_batch_counts.up.sql` widens sync-batch counters to `bigint` for large streaming imports; `000003_director_report_export_permission.up.sql` adds the approved `REPORT_EXPORT` capability to `ADMIN` and `DIRECTOR`; `000004_v1_access_catalog_alignment.up.sql` aligns the role catalog with the approved V1 matrix, separates cancellation from synchronization, and stops without changing data if a user still has a retired role.
- `queries/`: SQL consumed by `sqlc` for health, integration status/revocation, RAW ingestion and idempotent receipt batches.
- `privileges/runtime.sql`: least-privilege grants applied after migrations.
- `sqlc.yaml`: PostgreSQL + `pgx/v5` generation configuration.
- `gen/`: generated Go code (created by `sqlc generate`; do not edit manually).

## Invariants

- Money is `numeric(19,4)`; percentages are `numeric(20,8)`.
- External identity is `(connection_id, source_system, source_entity, source_id)`.
- RAW rows are append-only. In one transaction, call `retire_current_raw_record` for the old version, insert the new payload version, and commit the cursor only after the whole batch succeeds; rollback restores the previous current version.
- `payload_hash` identifies an exact canonical payload version. `raw_records` never stores client secrets, access tokens or refresh tokens.
- `integration_connections.secret_ref` is only an opaque reference to the worker vault. `DPAPI_CURRENT_USER` means the `NT SERVICE\MajucauWorker` identity: the UI sends credentials only through the authenticated Named Pipe and only the worker encrypts/decrypts them. Client secrets/tokens remain outside PostgreSQL.
- `integration_sync_logs` and `integration_status` are read views; `integration_sync_batches` and `integration_connections` are the sources of truth.
- Lineage uses typed tables with real foreign keys. `record_lineage` is a read-only convenience view and is not an integrity boundary.
- Production migrations are forward-only. There is intentionally no destructive `down` migration for the initial schema; a disposable development database may be dropped and recreated.
- Migrations run as a non-login schema owner. The worker connects as the separate `majucau_runtime` SCRAM role. It can read only the integration and dashboard relations currently used by the worker, and can write only integration configuration/status, sync batches/cursors, RAW inserts, receipts and projected receivables. It cannot delete rows, mutate RAW/audit directly, write the access catalog or financial statements, own database objects, or execute arbitrary public routines. It can execute the controlled RAW version-retirement function used by sync; redaction stays owner-only until it has an authenticated application path.
- Apply the explicit privilege script after every migration. New modules must update `privileges/runtime.sql` and `privileges/verify_runtime.sql` with the exact relations and operations they need; no future table or function receives runtime access implicitly.

## Apply and validate

Run against a disposable PostgreSQL 14+ database first:

```text
for migration in migrations/*.up.sql; do
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$migration"
done
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f privileges/runtime.sql
sqlc generate
go test ./...
```

Then inspect:

```sql
SELECT current_database(), current_user, clock_timestamp(), NOT pg_is_in_recovery();
SELECT * FROM integration_status;
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;
SELECT table_name FROM information_schema.views WHERE table_schema = 'public' ORDER BY table_name;
```

The repository root owns the Go module, so this task does not change `go.mod`. The generated package is configured under `database/gen` and should only be created when the repository's `sqlc` toolchain is available.

## Required migration checks

1. Apply with `ON_ERROR_STOP=1` to an empty disposable database.
2. Apply the migration a second time only if the migration runner correctly records version state; the SQL itself is an initial migration and is not intended to be re-run directly.
3. Insert two identical RAW payloads and verify the unique identity/version constraint returns one row.
4. In one transaction retire the previous current row, insert a changed payload, and verify exactly one `is_current` row.
5. Attempt direct RAW `UPDATE`/`DELETE` and verify both fail.
6. As the schema owner in the disposable database, run `redact_raw_record` with an approved actor/reason and verify the tombstone metadata and `audit_events` row; confirm that `majucau_runtime` cannot execute it.
7. Verify a statement contribution cannot have both source FKs or neither source FK.
8. Verify `sqlc generate` produces no uncommitted diff after generation.
9. `SET ROLE majucau_runtime` and verify direct RAW/audit mutation fails and the sync retirement function works; redaction remains owner-only until an authenticated application flow exists.
10. Run `privileges/verify_runtime.sql` after `privileges/runtime.sql` to check the runtime role's exact read/write allowlist (including column-level grants), role attributes, object ownership, routine execution and append-only protections. CI performs this check against its disposable PostgreSQL service.

No real credentials or personal payloads belong in this directory or its fixtures.
