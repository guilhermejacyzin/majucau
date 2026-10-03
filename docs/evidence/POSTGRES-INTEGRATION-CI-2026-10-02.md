# PostgreSQL integration tests in CI — 2026-10-02

## Evidence

- Workflow run: [37079852944](https://github.com/guilhermejacyzin/majucau/actions/runs/37079852944)
- Commit: `26ec678` (`Executar integrações PostgreSQL descartáveis na CI`)
- Result: both `verify-postgres` and `verify-windows` completed successfully.
- PostgreSQL service: disposable PostgreSQL 18.6 container; the CI job applied `database/migrations/000001_init.up.sql` before running the opt-in integration tests.
- Test data: repository fixtures under `internal/integrations/bling/testdata/e2e` and `internal/integrations/nuvempago/testdata/e2e`; no real account credentials or customer data are required.

## Coverage established

- `TestReceiptImportServicePostgresE2E`: imports the sanitized Bling receipt report into PostgreSQL, checks the partial batch outcome, and verifies that replay updates the current RAW record without duplicating it.
- `TestFutureImportServicePostgresE2E`: imports sanitized future B2C/`PROJECTED` rows, records a rejected row, verifies idempotent replay and current RAW records, and confirms that these rows do not enter the Bling `receipts` ledger.
- `verify-windows`: the same workflow run also passed Go and frontend checks, desktop and worker builds, unsigned NSIS installer build, installer/uninstaller smoke, and centralized output verification.

## Limits

This run validates the existing import paths against PostgreSQL using sanitized fixtures. It does not establish live-provider reconciliation, an official Nuvem B2C source contract, Bling API incremental-cursor behavior, failed-batch audit count semantics, or end-to-end report integration. Those requirements remain partial or blocked in `REQUIREMENTS-TRACEABILITY.md`.
