ALTER TABLE receivables
  ADD COLUMN source_payment_method_id bigint,
  ADD COLUMN source_financial_account_id bigint,
  ADD COLUMN source_accounting_account_id bigint;

ALTER TABLE payables
  ADD COLUMN source_payment_method_id bigint,
  ADD COLUMN source_financial_account_id bigint;

ALTER TABLE receivables
  ADD CONSTRAINT receivables_source_payment_method_id_nonnegative
    CHECK (source_payment_method_id IS NULL OR source_payment_method_id > 0),
  ADD CONSTRAINT receivables_source_financial_account_id_nonnegative
    CHECK (source_financial_account_id IS NULL OR source_financial_account_id > 0),
  ADD CONSTRAINT receivables_source_accounting_account_id_nonnegative
    CHECK (source_accounting_account_id IS NULL OR source_accounting_account_id > 0);

ALTER TABLE payables
  ADD CONSTRAINT payables_source_payment_method_id_nonnegative
    CHECK (source_payment_method_id IS NULL OR source_payment_method_id > 0),
  ADD CONSTRAINT payables_source_financial_account_id_nonnegative
    CHECK (source_financial_account_id IS NULL OR source_financial_account_id > 0);
