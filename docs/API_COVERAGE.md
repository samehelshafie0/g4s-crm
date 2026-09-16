# Frontend/backend coverage — 2026-09-16

The main CRM views now use the backend instead of in-memory sample records. This batch adds missing workflow endpoints, fixes request/response mismatches, and persists quote-builder, procurement, inventory and supporting-module data. **This is a verified integration milestone, not production release approval.**

Use [STATUS.md](../STATUS.md) for the completion backlog, [the original review](CRM_REVIEW.md) for historical findings, [local setup](LOCAL_DEVELOPMENT.md) to run the system, and [the endpoint inventory](API_ENDPOINTS.md) for all 179 registered routes and permissions. All business routes are under `/api/v1`.

## Coverage by screen

| Screen / capability | Backend and frontend implemented | Remaining acceptance gaps |
|---|---|---|
| Login / session | Login, current-user profile, bounded coordinated refresh, logout; active-user/current-role enforcement. Verified administrator login. | Password recovery email is not configured; recovery endpoints return 501. Profile/password APIs exist but the header profile control is not yet a management screen. |
| Customers | Header create/edit/delete and complete collection loading. Nested site/contact list/create/update/delete; parent ownership, single primary contact and contact/site consistency enforced. | Nested site/contact editing UI still needs to be built; current detail displays these records. |
| Opportunities | Live customer/active-user lookups, owners, service types, dates, costs, margins, stage changes and quote relationships. | Agreed assignment/data visibility rules; separate-role browser acceptance and accessible form refinements. |
| Quotations | Live collection, creation, server-assigned numbers, duplicate/revision, deletion restrictions and activity. | Final list/export/permission UX review. |
| Quote Builder | Real catalog and pricing; atomic header/row save and reload; fractional quantities, multipliers, line/global discounts, optional rows, notes, addresses and terms. Optimistic lock version; submitted documents read-only. Submit → approve/reject → sent → accepted/declined; conversion to contract. | Approval thresholds remain provisional; sending records a status, not email delivery. Acceptance evidence/signature capture, dirty-navigation guard and full print/CSV review remain. |
| Contracts | Create/edit drafts, dates, renewal notice, accepted-quote amount/currency snapshot, activation, termination and transactional renewal. Conversion and renewal retries reuse the existing result. | Scheduled expiry/auto-renew jobs and approval policy/UAT. Auto-renew is a stored preference, not an executing scheduled job. |
| Dashboard | Live KPI, pipeline, activity, recent/expiring quotes, top customers and quarterly salesperson results. Removed sample rows. | Role/data visibility and final metric sign-off. Accepted quote value is not accounting revenue. Activity “View All” and global notifications/search are not implemented workflows. |
| Manufacturers | Headers and nested category CRUD; category deletion blocked while products use it. | Header/category form saves use separate requests, so a category failure can follow a successful header save. |
| Products | Header/cost fields, manufacturer/category references, vendors, manual/imported vendor rows, price history, document upload/attachment/download. Calculated costs stay server-owned. | Cost/price policy and historical selling-price snapshots need refinement. Vendor import and header writes are separate requests. |
| Catalog services | `/services` CRUD; combined `/catalog` supplies products, services and recurring items to builder and price books. | No dedicated service-catalog administration screen yet. Services can be managed through the API. |
| Price Books | Headers, dates, eligibility, mixed product/service/recurring entries, atomic entry replacement; builder applies eligible overrides. | Form saves header and entries in two requests. Volume tiers and price-book versioning are not implemented. |
| Recurring Services | Header, active flag, component lines, monthly costs, annual calculations and duplication persist. | Catalog pricing totals are not realized subscription revenue. Subscription billing/invoicing is not implemented. |
| Exchange Rates | Real pairs/history, create and dated rate updates; explicit Add exchange rate action for a fresh database. | Enter approved rates; there is no automatic external rate feed. Effective-date/backdating business policy remains open. |
| Inventory | Warehouse stock, reservations, movement history, reserve/release/fulfill, transfers, adjustments and reorder-level API. All stock writers serialize on product and stock rows. | Basic UI supports holds/releases and bulk movements; fulfillment/reorder editing need dedicated controls. No reservation of not-yet-received PO quantities. Source references are descriptive, not relational links. |
| Procurement | PO/SQ full-line persistence, actual PATCH/delete, SQ acceptance and idempotent PO conversion, PO approval/ordering and partial goods receipts. Receipts reconcile quantities and stock; foreign costs converted to SAR using an effective rate. | Receipt UI is basic; damaged-goods quarantine/returns and cancellation policy remain. Supplier-item catalog API exists; no automatic catalog update from a supplier quote. |
| Projects | Accepted-quote conversion, customer/totals/currency, manager IDs, priorities, dates, status and persisted records. | Procurement linkage/detail and project delivery workflow remain limited; one-project-per-quote is not enforced by a unique constraint. |
| Teams / users | Real directory and membership, leader validation, team CRUD; typed user creation/edit/deactivation/password reset APIs and last-admin protection. | No complete administrator user-management UI. Role-aware navigation and all cross-role workflows still need review. |
| Documents | Real multipart upload, authenticated download, metadata, soft deletion, versions and validated entity links. Product documents reference uploaded files. | Current UI provides basic upload/download/delete/customer link; version browsing/replacement and advanced link editing need controls. Malware scanning and retention/purge policy are not configured. |

## Important API contracts

- Success responses use `{success, data, meta?}`. Collection data is always an array, including empty results. Related display names are returned alongside canonical IDs. Dates returned by Go may be RFC3339; date inputs send `YYYY-MM-DD`.
- Writes use explicit editable fields. Display names, local UI IDs, generated numbers, computed totals and approval/status fields must not be copied into arbitrary PATCH bodies. Frontend service adapters filter these fields.
- Paginated lists support the existing allowlisted sort parameters. Most locally filtered screens now fetch every page via `allPages`; this avoids silently dropping records after page one. This is an interim correctness fix, not the final scalable server-driven table design. Lookups can still be bounded; for example customer lookup returns at most 100 results per query.
- `PUT /quotes/:id/builder` replaces all rows and editable header content in one transaction. Supply `lockVersion` from GET; stale writes return 409, non-draft writes 422. Service/recurring references use `serviceId` / `recurringServiceId`, not `productId`. Row order and false optional/print flags survive reload.
- Quote totals round to cents on the server. Non-item rows and unselected optional items do not contribute. Multiplier, line discount, quote discount, VAT and pre-tax margin are applied explicitly. The existing Go money representation is still floating point; final decimal/rounding/currency policy is a release gate.
- Quote lifecycle mutations use named endpoints. A quote must contain a selected item to submit. All submissions currently require authorized manager approval; low-margin warnings in the UI are advisory until the owner confirms thresholds.
- `POST /quotes/:id/convert-to-contract` accepts title, type and start/end dates; only an accepted quote converts. A second call returns the existing contract. Contract PATCH cannot rewrite quote-derived amounts/currency.
- `PATCH /procurement/purchase-orders/:id` and supplier-quote PATCH save complete editable headers/items. Clients send all editable items; these are not partial line patches. Only draft POs are editable/deletable. Supplier-quote conversion requires an accepted quote and catalog product IDs for every line.
- Goods receipts use real `poItemId`, `productId`, positive `receivedQty`, warehouse and condition. Receipt locks reject over-receipt. Good stock enters inventory at weighted landed SAR cost; damaged quantities count as received but are not available stock. FX rate is snapshotted on the receipt. Shipping/customs start at zero in the UI; previous automatic 4%/5% estimates were removed.
- Stock adjustments require a reason and nonzero signed quantity. Transfers cannot move reserved stock. Reservation release/fulfill are state transitions, not record deletion. Concurrent stock writes preserve `0 ≤ reserved ≤ on-hand`.
- `PUT /price-books/:id/entries` validates/replaces the full entry list transactionally; duplicate entries reject without removing existing entries. Each entry identifies exactly one product, service or recurring service.
- Documents are uploaded as multipart `file`, maximum 50 MB, with metadata fields. Download through authenticated Axios/blob requests; a plain browser link lacks the bearer token. Version uploads are immutable file snapshots. Storage paths are never public API fields and downloads are contained under the configured upload root. Soft deletion retains files for recovery.
- `PATCH /users/:id/password` is an administrator operation. Password changes revoke refresh sessions; already-issued access tokens remain valid until expiry unless the user is deactivated. No public registration or working email-reset flow is provided.

## Database changes

`000003_quote_builder.up.sql` adds builder row/header fields, service catalog, revisions/lock version, fractional quantities and uniqueness for quote-to-contract conversion.

`000004_module_workflows.up.sql` adds supplier-quote-to-PO linkage, inventory constraints, document version snapshots, mixed price-book references, recurring component storage, quote notes/pricing references, product document references, receipt FX snapshots and contract/project currencies.

Both are applied to the isolated local development database. Integration tests apply all four migrations to a fresh disposable PostgreSQL 16 database. **Do not edit applied migrations**; add a new one. An existing production dataset has not been migrated or tested.

## Verification and limits

- PostgreSQL integration suite with race detector: 22 scenario groups covering foundations and module workflows, including builder save/reload/stale edits, immutable states, conversion/renewal retries, nested relationships, mixed price books, recurring components, document versions/download containment, partial/over-receipts, SAR valuation of foreign-currency receipts, concurrent reservations and all frontend collection routes.
- Frontend: 13 tests passing, covering session refresh, file imports, complete-page loading, failed-page handling, write-field filtering and procurement display adapters. Checked production build passes with no TypeScript errors; the file-parser chunk size warning remains.
- `go vet ./...` passes. No live server deployment, hosted CI run, backup restore or full role-by-role UAT has been performed.
- Browser: administrator login, customer creation, quote line save/reload (SAR 200 subtotal, SAR 30 VAT, SAR 230 total), submit/approve/sent/accepted and read-only controls verified. Dashboard reflects persisted quote value/activity. Main module pages were opened against the live local API; this is smoke coverage, not exhaustive form/edge-case testing.

## Next work, in order

1. Complete customer sites/contacts and administrator user/profile/password screens; expose service-catalog management and document versions. Finish visible controls that currently have only API support.
2. Confirm role visibility and approval policy, then run the full sales journey with separate sales/manager accounts, failure handling and business acceptance evidence. Hide unauthorized actions and finish mobile/accessibility behavior.
3. Finalize money/FX/rounding, cross-currency UI labels, exports and recurring reporting semantics; implement scalable table pagination/search and broad mutation error/concurrency consistency.
4. Complete procurement exception handling, project associations, stock reconciliation/fulfillment UI, expiry jobs and audit coverage. Some legacy list handlers still need consistent database-error propagation; not all writes have audit events or optimistic locks.
5. Configure the target server, production secrets/TLS/storage, recovery email, backups and tested restore/rollback; run migration upgrade and load/UAT checks before publishing.
