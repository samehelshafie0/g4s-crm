-- Orders and supplier quotes raised from a customer quote before orders carried
-- the job take that quote's opportunity, so they show on the job.
UPDATE purchase_orders po
SET opportunity_id = q.opportunity_id
FROM quotes q
WHERE po.source_quote_id = q.id AND po.opportunity_id IS NULL AND q.opportunity_id IS NOT NULL;

UPDATE supplier_quotes sq
SET opportunity_id = q.opportunity_id
FROM quotes q
WHERE sq.source_quote_id = q.id AND sq.opportunity_id IS NULL AND q.opportunity_id IS NOT NULL;
