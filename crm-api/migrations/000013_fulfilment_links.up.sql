-- A job is filled partly from stock and partly by ordering. Each quote line
-- says which, and reservations and order lines point at the line they fulfil
-- instead of naming it in free text.
ALTER TABLE quote_line_items ADD COLUMN fulfilment_source varchar(20)
    CHECK (fulfilment_source IN ('stock', 'purchase', 'service'));
ALTER TABLE stock_reservations ADD COLUMN quote_line_item_id uuid REFERENCES quote_line_items(id);
ALTER TABLE purchase_order_items ADD COLUMN quote_line_item_id uuid REFERENCES quote_line_items(id);
CREATE INDEX idx_stock_reservations_quote_line ON stock_reservations(quote_line_item_id);
CREATE INDEX idx_purchase_order_items_quote_line ON purchase_order_items(quote_line_item_id);
