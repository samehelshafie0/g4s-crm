ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_customer_currency;
ALTER TABLE quotes DROP COLUMN IF EXISTS fx_rate;
