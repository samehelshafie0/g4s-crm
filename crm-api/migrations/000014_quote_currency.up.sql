-- Suppliers quote in their own currency; customers are quoted in SAR or USD
-- only. A quote stores the SAR rate it was priced at, so a version reads as it
-- was sent even after the market rate moves.
ALTER TABLE quotes ADD COLUMN fx_rate numeric(12, 6) NOT NULL DEFAULT 1 CHECK (fx_rate > 0);
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM quotes WHERE currency NOT IN ('SAR', 'USD')) THEN
        RAISE EXCEPTION 'quotes exist in a currency other than SAR or USD; reprice them before restricting';
    END IF;
END $$;
ALTER TABLE quotes ADD CONSTRAINT quotes_customer_currency CHECK (currency IN ('SAR', 'USD'));
