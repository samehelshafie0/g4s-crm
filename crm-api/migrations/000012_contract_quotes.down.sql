ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_belong_to_one_parent;
DROP INDEX IF EXISTS idx_quotes_contract;
ALTER TABLE quotes DROP COLUMN IF EXISTS contract_id;
