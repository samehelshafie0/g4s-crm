-- A vendor quotation and its purchase order are raised to fulfil a customer
-- quotation, usually within a project. Without these links the chain from the
-- customer's quote to the supplier's order cannot be followed in either direction.
ALTER TABLE supplier_quotes ADD COLUMN source_quote_id uuid REFERENCES quotes(id);
ALTER TABLE supplier_quotes ADD COLUMN project_id uuid REFERENCES projects(id);
ALTER TABLE purchase_orders ADD COLUMN project_id uuid REFERENCES projects(id);
CREATE INDEX idx_supplier_quotes_source_quote ON supplier_quotes(source_quote_id);
CREATE INDEX idx_supplier_quotes_project ON supplier_quotes(project_id);
CREATE INDEX idx_purchase_orders_project ON purchase_orders(project_id);
