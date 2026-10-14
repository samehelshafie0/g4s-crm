ALTER TABLE goods_receipt_items
  ADD COLUMN damaged_qty integer NOT NULL DEFAULT 0 CHECK (damaged_qty >= 0 AND damaged_qty <= received_qty),
  ADD COLUMN return_status varchar(20) NOT NULL DEFAULT 'not-required' CHECK (return_status IN ('not-required','pending','returned')),
  ADD COLUMN return_reference text NOT NULL DEFAULT '',
  ADD COLUMN returned_at timestamptz;
UPDATE goods_receipt_items SET damaged_qty = received_qty, return_status = 'pending'
WHERE condition IN ('damaged','partial-damage');
ALTER TABLE goods_receipts ADD COLUMN replacement_po_id uuid REFERENCES purchase_orders(id);
CREATE INDEX idx_receipt_pending_returns ON goods_receipt_items(return_status) WHERE return_status = 'pending';
