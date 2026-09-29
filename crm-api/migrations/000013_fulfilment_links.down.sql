DROP INDEX IF EXISTS idx_purchase_order_items_quote_line;
DROP INDEX IF EXISTS idx_stock_reservations_quote_line;
ALTER TABLE purchase_order_items DROP COLUMN IF EXISTS quote_line_item_id;
ALTER TABLE stock_reservations DROP COLUMN IF EXISTS quote_line_item_id;
ALTER TABLE quote_line_items DROP COLUMN IF EXISTS fulfilment_source;
