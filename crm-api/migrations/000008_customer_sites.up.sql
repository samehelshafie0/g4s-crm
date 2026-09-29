-- Sites are a customer's branches across the Kingdom. An opportunity chooses
-- which of them it covers, so a site needs enough detail to identify a place
-- and the person to meet there.
ALTER TABLE customer_sites
    ADD COLUMN code varchar(50) NOT NULL DEFAULT '',
    ADD COLUMN site_type varchar(30) NOT NULL DEFAULT 'branch'
        CHECK (site_type IN ('head-office', 'branch', 'warehouse', 'plant', 'retail', 'other')),
    ADD COLUMN latitude numeric(9, 6),
    ADD COLUMN longitude numeric(9, 6),
    ADD COLUMN contact_name varchar(255) NOT NULL DEFAULT '',
    ADD COLUMN contact_phone varchar(50) NOT NULL DEFAULT '',
    ADD COLUMN is_active boolean NOT NULL DEFAULT true;
CREATE INDEX idx_customer_sites_customer ON customer_sites(customer_id);
