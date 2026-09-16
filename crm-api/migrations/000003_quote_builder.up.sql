ALTER TABLE quotes ADD COLUMN lock_version integer NOT NULL DEFAULT 1;
ALTER TABLE quotes ADD COLUMN parent_quote_id uuid REFERENCES quotes(id);
ALTER TABLE quotes ADD COLUMN payment_terms text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN delivery_terms text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN introduction_text text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN closing_text text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN internal_notes text NOT NULL DEFAULT '';
ALTER TABLE quotes ADD COLUMN sold_to jsonb NOT NULL DEFAULT '{}';
ALTER TABLE quotes ADD COLUMN ship_to jsonb NOT NULL DEFAULT '{}';
ALTER TABLE quote_line_items ALTER COLUMN quantity TYPE numeric(16,4);
ALTER TABLE quote_line_items ADD COLUMN row_type text NOT NULL DEFAULT 'item' CHECK (row_type IN ('item','heading','comment','subtotal','discount'));
ALTER TABLE quote_line_items ADD COLUMN source text NOT NULL DEFAULT 'product' CHECK (source IN ('product','service','recurring','write-in'));
ALTER TABLE quote_line_items ADD COLUMN multiplier numeric(16,4) NOT NULL DEFAULT 1 CHECK (multiplier > 0);
ALTER TABLE quote_line_items ADD COLUMN discount_percent numeric(7,4) NOT NULL DEFAULT 0 CHECK(discount_percent BETWEEN 0 AND 100);
ALTER TABLE quote_line_items ADD COLUMN is_optional boolean NOT NULL DEFAULT false;
ALTER TABLE quote_line_items ADD COLUMN is_selected boolean NOT NULL DEFAULT true;
ALTER TABLE quote_line_items ADD COLUMN is_printable boolean NOT NULL DEFAULT true;
ALTER TABLE quote_line_items ADD COLUMN heading_text text NOT NULL DEFAULT '';
ALTER TABLE quote_line_items ADD COLUMN comment_text text NOT NULL DEFAULT '';
ALTER TABLE quote_line_items ADD COLUMN rate_type text NOT NULL DEFAULT '';
ALTER TABLE quote_line_items ADD COLUMN billing_cycle text NOT NULL DEFAULT '';
CREATE TABLE catalog_services (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 sku varchar(100) NOT NULL UNIQUE, name varchar(255) NOT NULL, description text NOT NULL DEFAULT '', department department NOT NULL DEFAULT 'technical', rate_type varchar(100) NOT NULL DEFAULT 'per day',
 unit_cost numeric(18,4) NOT NULL DEFAULT 0 CHECK(unit_cost>=0), unit_price numeric(18,4) NOT NULL DEFAULT 0 CHECK(unit_price>=0), is_active boolean NOT NULL DEFAULT true
);
ALTER TABLE quote_line_items ADD COLUMN service_id uuid REFERENCES catalog_services(id);
ALTER TABLE quote_line_items ADD COLUMN recurring_service_id uuid REFERENCES recurring_services(id);
CREATE UNIQUE INDEX uq_contract_quote ON contracts(quote_id) WHERE quote_id IS NOT NULL AND deleted_at IS NULL;
