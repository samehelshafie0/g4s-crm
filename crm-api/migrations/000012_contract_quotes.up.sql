-- A maintenance contract is a period, and work is quoted inside it as it is
-- agreed. A quote therefore belongs to exactly one parent: the opportunity it
-- prices, or the contract it was raised inside.
ALTER TABLE quotes ADD COLUMN contract_id uuid REFERENCES contracts(id);
CREATE INDEX idx_quotes_contract ON quotes(contract_id);
ALTER TABLE quotes ADD CONSTRAINT quotes_belong_to_one_parent
    CHECK ((opportunity_id IS NULL) <> (contract_id IS NULL));

-- Deleting a job must not strand its quote without a parent.
ALTER TABLE quotes DROP CONSTRAINT quotes_opportunity_id_fkey;
ALTER TABLE quotes ADD CONSTRAINT quotes_opportunity_id_fkey
    FOREIGN KEY (opportunity_id) REFERENCES opportunities(id) ON DELETE RESTRICT;
