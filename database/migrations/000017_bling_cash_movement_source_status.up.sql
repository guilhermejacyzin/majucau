-- The public Bling /caixas schema describes situacao as an optional string
-- and documents only an example. Store the source value without assigning
-- financial meaning; dashboard rules decide which status belongs in totals.
ALTER TABLE bling_cash_movements
  DROP CONSTRAINT bling_cash_movements_source_status_check;

ALTER TABLE bling_cash_movements
  ADD CONSTRAINT bling_cash_movements_source_status_check
  CHECK (octet_length(source_status) <= 128);

COMMENT ON COLUMN bling_cash_movements.source_status IS
  'Optional source status from Bling; blank or unknown values are retained as metadata and are not implicitly treated as realized.';
