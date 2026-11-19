ALTER TABLE quotes ADD COLUMN accepted_at timestamptz;
-- Old acceptance times were not stored. Use the last recorded quote change,
-- explicitly flagged as an estimate in history responses.
ALTER TABLE quotes ADD COLUMN accepted_at_estimated boolean NOT NULL DEFAULT false;
UPDATE quotes SET accepted_at = updated_at, accepted_at_estimated = true WHERE status = 'accepted';

CREATE TABLE supplier_price_records (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 recorded_at timestamptz NOT NULL DEFAULT now(),
 supplier_quote_id uuid NOT NULL REFERENCES supplier_quotes(id),
 source_item_id uuid NOT NULL UNIQUE,
 revision integer NOT NULL,
 product_id uuid NOT NULL REFERENCES products(id),
 sku text NOT NULL DEFAULT '',
 product_name text NOT NULL DEFAULT '',
 supplier_name text NOT NULL,
 supplier_ref text NOT NULL DEFAULT '',
 supplier_quote_number text NOT NULL,
 currency text NOT NULL,
 unit_cost numeric(18,2) NOT NULL,
 quantity integer NOT NULL,
 valid_until timestamptz,
 source_quote_id uuid REFERENCES quotes(id)
);
CREATE INDEX supplier_price_product_date_idx ON supplier_price_records(product_id, recorded_at DESC);
CREATE INDEX supplier_price_quote_idx ON supplier_price_records(supplier_quote_id, revision);
CREATE INDEX quotes_accepted_prices_idx ON quotes(customer_id, accepted_at DESC) WHERE status = 'accepted' AND deleted_at IS NULL;
INSERT INTO supplier_price_records (recorded_at, supplier_quote_id, source_item_id, revision, product_id, sku, product_name, supplier_name, supplier_ref, supplier_quote_number, currency, unit_cost, quantity, valid_until, source_quote_id)
 SELECT sq.updated_at, sq.id, item.id, 1, item.product_id, COALESCE(item.product_sku,''), item.product_name, sq.supplier_name, COALESCE(sq.supplier_ref,''), sq.sq_number, sq.currency, item.unit_cost, item.quantity, COALESCE(item.valid_until,sq.valid_until), sq.source_quote_id
 FROM supplier_quote_items item JOIN supplier_quotes sq ON sq.id = item.supplier_quote_id
 WHERE item.product_id IS NOT NULL AND item.deleted_at IS NULL;
