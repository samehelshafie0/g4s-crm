-- One opportunity has one quote, and the quote number is bound to it for the
-- life of the job. Revising raises the version, never the number, and every
-- version is kept. Only the newest version is current and editable.

-- A quote chain is a root quote and its revisions. Resolve each quote's root.
CREATE TEMP TABLE quote_roots ON COMMIT DROP AS
WITH RECURSIVE chain AS (
    SELECT id, id AS root_id, 1 AS depth FROM quotes WHERE parent_quote_id IS NULL
    UNION ALL
    SELECT q.id, chain.root_id, chain.depth + 1 FROM quotes q JOIN chain ON q.parent_quote_id = chain.id
)
SELECT id, root_id, depth FROM chain;

-- An opportunity with two separate chains cannot be merged automatically.
DO $$
BEGIN
    IF EXISTS (
        SELECT opportunity_id FROM quotes
        WHERE parent_quote_id IS NULL AND opportunity_id IS NOT NULL AND deleted_at IS NULL
        GROUP BY opportunity_id HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'an opportunity holds more than one quote chain; decide which is the quote before migrating';
    END IF;
END $$;

-- Every quote now belongs to a job. A quote raised without an opportunity gets
-- one, named after it, at the stage its status implies. A soft-deleted quote
-- gets a soft-deleted opportunity, so the parent rule holds for every row.
ALTER TABLE opportunities ADD COLUMN migrated_from_quote uuid;
INSERT INTO opportunities (title, customer_id, stage, estimated_value, estimated_cost, deleted_at, migrated_from_quote)
SELECT 'Quotation ' || q.quote_number, q.customer_id,
       (CASE q.status
            WHEN 'accepted' THEN 'closed-won'
            WHEN 'declined' THEN 'closed-lost'
            WHEN 'expired' THEN 'closed-lost'
            WHEN 'draft' THEN 'qualification'
            ELSE 'proposal'
        END)::opportunity_stage,
       q.total, q.total_cost, q.deleted_at, q.id
FROM quotes q
WHERE q.parent_quote_id IS NULL AND q.opportunity_id IS NULL;
UPDATE opportunities SET delivery_status = 'planning'
WHERE migrated_from_quote IS NOT NULL AND stage = 'closed-won';
UPDATE quotes q SET opportunity_id = o.id FROM opportunities o WHERE o.migrated_from_quote = q.id;
ALTER TABLE opportunities DROP COLUMN migrated_from_quote;

-- Revisions inherit the root's opportunity and its number; versions renumber by
-- depth. The one-number-per-row rule goes first, since revisions now share one.
ALTER TABLE quotes DROP CONSTRAINT quotes_quote_number_key;
UPDATE quotes q
SET opportunity_id = root.opportunity_id,
    quote_number = root.quote_number,
    version = r.depth
FROM quote_roots r JOIN quotes root ON root.id = r.root_id
WHERE q.id = r.id;

ALTER TABLE quotes ALTER COLUMN version SET NOT NULL;
ALTER TABLE quotes
    ADD COLUMN is_current boolean NOT NULL DEFAULT true,
    ADD COLUMN revision_reason text NOT NULL DEFAULT '',
    ADD COLUMN superseded_at timestamptz;
UPDATE quotes q SET is_current = false, superseded_at = now()
WHERE EXISTS (SELECT 1 FROM quotes c WHERE c.parent_quote_id = q.id AND c.deleted_at IS NULL);

CREATE UNIQUE INDEX quotes_number_version ON quotes(quote_number, version);
CREATE UNIQUE INDEX quotes_one_current_version ON quotes(quote_number)
    WHERE is_current AND deleted_at IS NULL;
CREATE UNIQUE INDEX quotes_one_chain_per_opportunity ON quotes(opportunity_id)
    WHERE is_current AND deleted_at IS NULL AND opportunity_id IS NOT NULL;
