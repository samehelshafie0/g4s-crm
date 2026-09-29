-- An opportunity covers one or more of its customer's sites, and a quote line
-- can name the site it serves so a multi-site job prices per site in one quote.
CREATE TABLE opportunity_sites (
    opportunity_id uuid NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
    site_id uuid NOT NULL REFERENCES customer_sites(id),
    PRIMARY KEY (opportunity_id, site_id)
);
CREATE INDEX idx_opportunity_sites_site ON opportunity_sites(site_id);
ALTER TABLE quote_line_items ADD COLUMN site_id uuid REFERENCES customer_sites(id);
