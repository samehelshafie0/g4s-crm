ALTER TABLE purchase_orders ADD COLUMN supplier_quote_id uuid REFERENCES supplier_quotes(id);
CREATE UNIQUE INDEX uq_po_supplier_quote ON purchase_orders(supplier_quote_id) WHERE supplier_quote_id IS NOT NULL AND deleted_at IS NULL;
ALTER TABLE warehouse_stocks ADD CONSTRAINT stock_reserved_within_on_hand CHECK (reserved_qty <= on_hand_qty);
CREATE TABLE document_versions (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 document_id uuid NOT NULL REFERENCES documents(id), version varchar(50) NOT NULL, file_name text NOT NULL, file_size bigint NOT NULL, file_type text NOT NULL, file_path text NOT NULL, uploaded_by_id uuid REFERENCES users(id),
 UNIQUE(document_id,version)
);
INSERT INTO document_versions(document_id,version,file_name,file_size,file_type,file_path,uploaded_by_id) SELECT id,COALESCE(NULLIF(version,''),'1'),file_name,file_size,file_type,file_path,uploaded_by_id FROM documents;
UPDATE documents SET version='1' WHERE version IS NULL OR version='';
ALTER TABLE quotes ADD COLUMN purchasing_notes text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN statement_of_work text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN price_book_id uuid REFERENCES price_books(id);
ALTER TABLE recurring_services ADD COLUMN line_items jsonb NOT NULL DEFAULT '[]';
ALTER TABLE price_book_entries ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE price_book_entries ADD COLUMN kind text NOT NULL DEFAULT 'product' CHECK(kind IN ('product','service','recurring'));
ALTER TABLE price_book_entries ADD COLUMN service_id uuid REFERENCES catalog_services(id);
ALTER TABLE price_book_entries ADD COLUMN recurring_service_id uuid REFERENCES recurring_services(id);
ALTER TABLE price_book_entries ADD CONSTRAINT price_entry_one_item CHECK ((kind='product' AND product_id IS NOT NULL AND service_id IS NULL AND recurring_service_id IS NULL) OR (kind='service' AND service_id IS NOT NULL AND product_id IS NULL AND recurring_service_id IS NULL) OR (kind='recurring' AND recurring_service_id IS NOT NULL AND product_id IS NULL AND service_id IS NULL));
ALTER TABLE product_documents ADD COLUMN document_id uuid REFERENCES documents(id);
ALTER TABLE goods_receipts ADD COLUMN fx_rate numeric(18,8) NOT NULL DEFAULT 1;
ALTER TABLE contracts ADD COLUMN currency currency NOT NULL DEFAULT 'SAR';
ALTER TABLE projects ADD COLUMN currency currency NOT NULL DEFAULT 'SAR';
UPDATE contracts c SET currency=q.currency FROM quotes q WHERE c.quote_id=q.id;
UPDATE projects p SET currency=q.currency FROM quotes q WHERE p.quote_id=q.id;
