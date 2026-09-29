-- The opportunity is the job from enquiry to handover. The separate project
-- record held delivery fields that nothing ever filled in; the opportunity
-- carries them itself, and delivery only exists once the sale is won.
ALTER TYPE project_status RENAME TO delivery_status;

ALTER TABLE opportunities
    ADD COLUMN delivery_status delivery_status,
    ADD COLUMN project_manager_id uuid REFERENCES users(id),
    ADD COLUMN start_date date,
    ADD COLUMN target_end_date date,
    ADD COLUMN actual_end_date date,
    ADD COLUMN actual_value numeric(15, 2),
    ADD COLUMN actual_cost numeric(15, 2);
UPDATE opportunities SET delivery_status = 'planning' WHERE stage = 'closed-won';
ALTER TABLE opportunities ADD CONSTRAINT opportunities_delivery_requires_win
    CHECK (delivery_status IS NULL OR stage = 'closed-won');

-- Procurement is attributed to the job, which is now the opportunity.
ALTER TABLE supplier_quotes DROP COLUMN project_id;
ALTER TABLE purchase_orders DROP COLUMN project_id;
ALTER TABLE supplier_quotes ADD COLUMN opportunity_id uuid REFERENCES opportunities(id);
ALTER TABLE purchase_orders ADD COLUMN opportunity_id uuid REFERENCES opportunities(id);
CREATE INDEX idx_supplier_quotes_opportunity ON supplier_quotes(opportunity_id);
CREATE INDEX idx_purchase_orders_opportunity ON purchase_orders(opportunity_id);

UPDATE document_links SET entity_type = 'opportunity' WHERE entity_type = 'project';

-- Refuse rather than discard: this migration was written against an empty table.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM projects WHERE deleted_at IS NULL) THEN
        RAISE EXCEPTION 'projects holds live rows; migrate them to opportunities before retiring the table';
    END IF;
END $$;
DROP TABLE projects;
