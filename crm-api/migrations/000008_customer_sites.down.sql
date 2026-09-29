DROP INDEX IF EXISTS idx_customer_sites_customer;
ALTER TABLE customer_sites
    DROP COLUMN IF EXISTS is_active, DROP COLUMN IF EXISTS contact_phone, DROP COLUMN IF EXISTS contact_name,
    DROP COLUMN IF EXISTS longitude, DROP COLUMN IF EXISTS latitude, DROP COLUMN IF EXISTS site_type, DROP COLUMN IF EXISTS code;
