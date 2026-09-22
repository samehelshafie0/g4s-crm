DROP INDEX IF EXISTS idx_purchase_orders_project;
DROP INDEX IF EXISTS idx_supplier_quotes_project;
DROP INDEX IF EXISTS idx_supplier_quotes_source_quote;
ALTER TABLE purchase_orders DROP COLUMN IF EXISTS project_id;
ALTER TABLE supplier_quotes DROP COLUMN IF EXISTS project_id;
ALTER TABLE supplier_quotes DROP COLUMN IF EXISTS source_quote_id;
