-- Opportunities are numbered, and a job's quote shares the number:
-- OPP-2026-0034 is quoted as QT-2026-0034 (versions v1, v2, ...). Quotes raised
-- inside a contract count up under the contract: QT-CNT-2026-0003-01.

ALTER TABLE opportunities ADD COLUMN opportunity_number varchar(30);

WITH numbered AS (
    SELECT id,
           extract(year FROM created_at)::int AS yr,
           row_number() OVER (PARTITION BY extract(year FROM created_at) ORDER BY created_at, id) AS n
    FROM opportunities
)
UPDATE opportunities o
SET opportunity_number = 'OPP-' || numbered.yr || '-' || lpad(numbered.n::text, 4, '0')
FROM numbered
WHERE o.id = numbered.id;

ALTER TABLE opportunities ALTER COLUMN opportunity_number SET NOT NULL;
CREATE UNIQUE INDEX opportunities_number ON opportunities(opportunity_number);

INSERT INTO sequences (name, prefix, year, current)
VALUES ('opportunity', 'OPP', extract(year FROM now())::int,
        (SELECT count(*) FROM opportunities
         WHERE opportunity_number LIKE 'OPP-' || extract(year FROM now())::int || '-%'));

-- Deleted quotes no longer reserve their number and version, so a job whose
-- draft was deleted can be quoted again, and a deleted draft revision does not
-- block the next revision.
DROP INDEX quotes_number_version;
CREATE UNIQUE INDEX quotes_number_version ON quotes(quote_number, version) WHERE deleted_at IS NULL;

-- Renumber live quote chains to their opportunity or contract.
CREATE TEMP TABLE quote_renumber (old_number varchar(50) PRIMARY KEY, new_number varchar(50) NOT NULL);

INSERT INTO quote_renumber (old_number, new_number)
SELECT DISTINCT q.quote_number, 'QT-' || substring(o.opportunity_number FROM 5)
FROM quotes q
JOIN opportunities o ON o.id = q.opportunity_id
WHERE q.deleted_at IS NULL;

INSERT INTO quote_renumber (old_number, new_number)
SELECT chains.quote_number,
       'QT-' || c.contract_number || '-' || lpad(row_number() OVER (PARTITION BY chains.contract_id ORDER BY chains.first_created)::text, 2, '0')
FROM (
    SELECT quote_number, contract_id, min(created_at) AS first_created
    FROM quotes
    WHERE deleted_at IS NULL AND contract_id IS NOT NULL AND opportunity_id IS NULL
    GROUP BY quote_number, contract_id
) chains
JOIN contracts c ON c.id = chains.contract_id
ON CONFLICT (old_number) DO NOTHING;

-- Two steps, so a new number never meets an old one that has not moved yet.
UPDATE quotes q SET quote_number = 'TMP-' || r.new_number
FROM quote_renumber r
WHERE q.quote_number = r.old_number AND q.deleted_at IS NULL;

UPDATE quotes SET quote_number = substring(quote_number FROM 5)
WHERE quote_number LIKE 'TMP-%' AND deleted_at IS NULL;

-- A deleted quote keeps its old number unless a live quote now uses it.
UPDATE quotes d SET quote_number = d.quote_number || '-deleted'
WHERE d.deleted_at IS NOT NULL
  AND EXISTS (SELECT 1 FROM quotes l WHERE l.deleted_at IS NULL AND l.quote_number = d.quote_number);

-- Stock held for a job names its quote ("QT-2026-0002 v1"); keep that readable.
UPDATE stock_reservations s
SET source_ref = r.new_number || substring(s.source_ref FROM length(r.old_number) + 1)
FROM quote_renumber r
WHERE s.source_ref LIKE r.old_number || ' v%' AND r.old_number <> r.new_number;

DROP TABLE quote_renumber;
