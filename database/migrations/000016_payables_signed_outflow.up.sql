-- The approved financial presentation treats payable balances as cash outflows.
-- Keep amounts signed negative in the normalized read model so totals, due-date
-- cards, and payable rows agree with the cash-flow direction.
ALTER TABLE payables
  DROP CONSTRAINT payables_open_balance_check;

UPDATE payables
   SET open_balance = -open_balance
 WHERE open_balance > 0;

COMMENT ON COLUMN payables.open_balance IS
  'Signed open cash outflow from a payable title; normalized negative regardless of the sign returned by Bling.';
