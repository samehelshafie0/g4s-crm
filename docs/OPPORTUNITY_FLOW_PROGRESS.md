# Opportunity-centred flow — build progress

**Last updated:** 2026-10-07
**Design doc:** [Opportunity-Centred CRM Flow](https://claude.ai/code/artifact/5421df6f-acb3-4fe5-b2ef-9f10fd45b728) (Claude Doc — the plan, rules, diagrams, open questions)
**State:** backend and frontend committed on `feature/opportunity-flow`; additional October 7 procurement/inventory and full-cycle/import fixes are verified locally and uncommitted. Nothing deployed or migrated on the live database. See [full-cycle acceptance](FULL_CYCLE_ACCEPTANCE.md) and [closeout report](PROCUREMENT_CLOSEOUT.md) for current results and remaining data/release blockers.

## Rules being implemented (confirmed by owner, 22 Sep 2026)

1. The quote number is bound to the opportunity for the life of the job.
2. One quote per opportunity, with many versions; every version is kept.
3. The project **is** the opportunity. No separate project entity.
4. A maintenance contract is a period; several quotes can be raised inside it.
5. Sites belong to the customer; an opportunity chooses from its customer's sites.
6. Customers record their KSA branches (city, region, type, contact).
7. Customer quotes go out in SAR or USD only; supplier quotes arrive in any currency.

## Defaults taken for the doc's open questions

The owner had not answered these; each can be changed later.

| Question | Default implemented |
|---|---|
| Two priced alternatives | Two opportunities (strict one-quote rule, enforced by a unique index) |
| Contract quotes and the pipeline | A quote raised inside a contract has no opportunity |
| Renewals | Unchanged: renewal creates a linked successor contract |
| Existing 3 quotes | Kept; `QT-2026-0003` had no opportunity, so the migration creates one for it |
| Delivery screen or filter | Separate `GET /delivery` list (screen still to build) |
| Who sets delivery status | `projects:update` holders (admin, project manager); project managers gain `opportunities:read` |
| USD rate | Live rate from the exchange-rate table, pinned to the version when the quote is created or its currency changed |

## Done — backend (all 7 phases)

### Migrations `crm-api/migrations/000008`–`000014`

| # | File | Change |
|---|---|---|
| 8 | `customer_sites` | code, site_type, lat/long, site contact, is_active |
| 9 | `opportunity_absorbs_project` | delivery_status (renamed `project_status` enum), project_manager_id, start/target/actual dates, actual value/cost; CHECK delivery only when won; SQ/PO `project_id` → `opportunity_id`; **drops `projects`** (refuses if it holds live rows) |
| 10 | `opportunity_sites` | join table; `quote_line_items.site_id` |
| 11 | `quote_versions` | orphan quotes get an opportunity; revision chains renumbered to the root's number with versions by depth; `is_current`, `revision_reason`, `superseded_at`; unique `(quote_number, version)`, one current version per number, one current quote per opportunity; refuses if an opportunity holds two chains |
| 12 | `contract_quotes` | `quotes.contract_id`; CHECK a quote has exactly one parent (opportunity XOR contract); opportunity FK → RESTRICT |
| 13 | `fulfilment_links` | `quote_line_items.fulfilment_source` (stock/purchase/service); `quote_line_item_id` on reservations and PO items |
| 14 | `quote_currency` | `quotes.fx_rate` (SAR per unit, pinned); CHECK currency in SAR/USD (refuses if others exist) |

**Verified:** applied cleanly to a copy of the live database, and to a synthetic copy with a 3-number revision chain, an orphan chain and a soft-deleted orphan — all renumbered and parented correctly. That test caught one ordering bug (renumbering before dropping the old unique key), since fixed.

### API

- **Sites:** `POST/PATCH /customers/:id/sites` take the new fields; deleting a site used by an opportunity or quote line is refused (deactivate instead).
- **Opportunities:** `siteIds` on create/update (must belong to the customer, active when newly chosen); customer cannot change once a quote exists; stage and delivery kept consistent (`syncDelivery`); response includes `sites`, `projectManagerName`, `currentQuote`.
- **Delivery:** `GET /delivery` (won opportunities, `?open=true`), `PATCH /opportunities/:id/delivery`, `GET /opportunities/:id/fulfilment` (per line: source, reserved, ordered, received, POs, state).
- **Quotes:** create requires exactly one of `opportunityId` / `contractId`; second quote on an opportunity → 409 naming the existing one; currency SAR/USD with pinned `fxRate`; list shows current versions only (`?versions=all`, `?opportunityId`, `?contractId`); every mutation requires the current version; deleting a draft revision restores the previous version; accepting a job's quote moves the opportunity to closed-won + delivery planning.
- **Versions:** `POST /quotes/:id/revisions` (reason required; only from approved/sent/declined/expired; same number, version+1, old one frozen); `GET /quotes/:id/versions`; `POST /quotes/:id/currency` converts all lines at old/new pinned rates. `/duplicate` removed.
- **Builder:** cannot move a quote to another customer/opportunity or change currency (use the currency endpoint); rows accept `siteId` (must be a site the job covers) and `fulfilmentSource`.
- **Contracts:** `GET /contracts/:id/quotes` (origin + quotes raised inside).
- **Fulfilment links:** reservations and PO lines accept `quoteLineItemId` — only against an accepted version, product must match; PO takes the job's opportunity automatically.
- **Dashboard:** current versions only, money converted with `total * fx_rate`, new `wonNotDelivered` KPI.
- **PDF:** reference printed as `QT-2026-0001 v3`; filename `QT-2026-0001-v3.pdf`.
- **Removed:** `ProjectHandler`, `/projects` routes, `Project` model; document links no longer accept `project`.
- **Fix:** `recordActivity` truncates actions over 50 characters (column is `varchar(50)`); full text stays in the description.

### Tests

`./scripts/test-integration.sh` — **passing**. Updated: every test quote now has an opportunity; the revision test now revises a *sent* quote, checks the number is kept, the old version is frozen, accepting wins the job, and an accepted version cannot be revised; the project test is now a delivery test; the appendix test revises after approval. Frontend unit tests not yet re-run (frontend unchanged).

## Done — frontend (all 8 phases, committed on `feature/opportunity-flow`)

1. **Customers → Sites tab** (`CustomerSites.vue`) with the KSA city list.
2. **Opportunities:** site multi-select; opportunity page `/opportunities/:id` (overview, quote and versions, fulfilment, procurement, documents, activity).
3. **Quotes list:** one row per job with version; a new quote starts from an opportunity.
4. **Quote builder:** version banner, revise with reason, version history, SAR/USD switch with pinned rate, site and fulfilment source per line.
5. **Delivery screen** `/delivery` replaces Projects (`/projects` redirects). `ProjectsView` and `projectsService` removed.
6. **Contracts:** Quotes section in the contract view and a "Raise quote" action.
7. **Procurement:** opportunity replaces project on POs and supplier quotes; PO lines carry `quoteLineItemId` from the source quote.
8. **Dashboard:** "Won, not delivered" tile.

`npm test` (18 pass) and `npm run build` (vue-tsc clean) pass. `docs/API_ENDPOINTS.md` is updated.

**Still to do:** rebuild release images, complete broader UAT/operational gates, then the release below. Browser checks against a migrated copy are recorded in the closeout report. Nothing has been migrated or deployed.

## Procurement fixes from the live-data test (2026-09-29)

Tested on a migrated copy of the live database with real catalog SKUs.

- **Fulfilment** counts only stock held for a line. Good receipts against a job's quote line are held for that job automatically, up to what the line still needs. Damaged units are reported separately and never count as covering the line. States: covered = held ≥ need; on order = held + outstanding ≥ need.
- **Supplier quote → PO** links each line to the job's accepted quote line (by product; products quoted on two lines are left unlinked) and takes the job's opportunity. A supplier quote with a source quote takes that quote's opportunity.
- **Migration 000015** fills the opportunity on existing POs and supplier quotes from their source quote.
- **PO wizard:**
  - Picking a won job brings in each material line with the quantity still to buy.
  - Lines are priced from the supplier's quote, the supplier catalog, then the product's vendor price, converted to the order's currency.
  - Catalog search covers all products.
  - Vendors named on products are offered as suppliers.
  - POs and supplier quotes have a currency choice.
- **Supplier quote form:** choosing a customer quote loads its material lines and opportunity.

**Live migration blocker:** migration 11 stops on live data because opportunity *op2* has two quotes, QT-2026-0002 (22 Sep) and QT-2026-0004 (29 Sep), both empty drafts. The owner must choose which one stays (delete the other) before migrating.

## Before deploying

- Back up the live database first; migration 9 drops a table and 11 renumbers quotes.
- Rebuild the API image and run `./g4s-crm-api migrate`, then deploy the frontend in the same release.
- Update the design doc if any default above is changed.
