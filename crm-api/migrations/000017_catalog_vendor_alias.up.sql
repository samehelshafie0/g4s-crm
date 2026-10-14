-- Verified alias: the supplier dropdown previously showed the master and its
-- Riyadh catalog spelling separately. Keep historical quote/receipt snapshots.
UPDATE products SET supplier_name = 'alfanar Electric', updated_at = now()
WHERE lower(trim(supplier_name)) IN ('alfanar electric', 'alfanar electric (riyadh)');
UPDATE product_vendor_entries SET vendor_name = 'alfanar Electric', updated_at = now()
WHERE lower(trim(vendor_name)) IN ('alfanar electric', 'alfanar electric (riyadh)');
-- The imported vendor 'in' is intentionally preserved until its identity is
-- confirmed. New writes reject that incomplete name.
