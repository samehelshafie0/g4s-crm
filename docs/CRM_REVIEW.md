# G4S CRM — implementation review

**Reviewed:** 16 September 2026  
**Scope:** frontend behavior and tables, API, business logic, database design, deployment, and completion planning.  
**Delivery priority confirmed by the owner:** customers → opportunities → quotes → contracts.  
**Deployment intent:** a server. Provider, operating system, public/private access, and domain are still unspecified.  
**Live progress:** [STATUS.md](../STATUS.md). This document records the baseline; fixes must update the status file and retain the original finding IDs.

## Overall finding

The project has a substantial interface and a useful backend foundation, but it is **not ready for real business data or deployment**. The main work is completing integration and enforcing business rules. A visual rebuild is unnecessary.

The frontend is a mixture of working API calls and independent demonstration data. Calling a store's fetch method does not mean the screen renders that store. Many screens continue to display and edit their own local arrays. The quote builder does not load the route's quote or persist its changes.

The Go API compiles, and the initial migration creates a clean PostgreSQL database successfully. However, critical create handlers omit required relationships, validation tags are not enforced by the binding helper, and generic updates bypass approval permissions. These are functional defects, not just missing enhancements.

**24 grouped findings:** 6 P0, 16 P1, 2 P2. P0 blocks the core workflow or a critical authorization boundary; P1 must be addressed before the affected feature is released; P2 improves correctness or usability after the blockers. Grouped findings contain multiple related examples; this is not a count of every individual defect.

## What was actually verified

| Check | Result | Evidence / limitation |
|---|---|---|
| `npm run build` | Failed | Type checking failed; no claim of a passing production build. |
| `npm run type-check` | Failed: 51 diagnostics across 7 files | [Raw output](review-evidence/frontend-typecheck.log). |
| `npm run build-only` | Passed | [Bundler output](review-evidence/frontend-bundle.log). This command skips type checking. |
| Frontend bundle | Import parser chunk 845.47 kB, 265.38 kB gzip | Vite emitted its large-chunk warning. No load/performance benchmark was run. |
| `go test ./...` | Passed compilation; every package reports `[no test files]` | This is not evidence of behavioral test coverage. |
| `go vet ./...` | Passed | Static checks only. |
| Initial SQL migration | Passed on disposable PostgreSQL 16 | 36 tables created; no existing project database was modified. |
| Real router and PostgreSQL probes | Multiple defects reproduced | [API probe output](review-evidence/api-probes.log), [probe source](review-evidence/api-probe-source.go.txt). |
| Browser review | Login, dashboard, quotations, customers, session refresh | Actual frontend connected to the isolated API; desktop and 390 × 844 viewport. |
| `npm audit --json` | 13 affected-package entries: 1 critical, 8 high, 2 moderate, 2 low | [Audit snapshot](review-evidence/npm-audit.json). Reachability and production impact still require triage. |
| Existing deployment | Not tested | No server was supplied. Docker image build, TLS, restoration, load, and migration rollback are not verified. |

Environment: Node 23.11.0, npm 10.9.2, Go 1.25.4. Review infrastructure used loopback ports 55432 / 18080 / 5174 because other projects occupy the standard API/database ports. The temporary API, database, keys, and dev servers were removed after review. No deployment, application fix, dependency upgrade, or real-data migration was performed.

The Impeccable context launcher returned permission denied. No PRODUCT.md, DESIGN.md, or applicable AGENTS.md was found in the inspected project/inherited locations. The bundled design detector did not run; interface conclusions below use source inspection and browser observations.

## Current module coverage

“API scaffold” means routes/models/handlers exist; it does not mean the module works end to end.

| Module | Frontend reality | Backend reality | Current assessment |
|---|---|---|---|
| Authentication | Login is wired; reload receives incomplete profile | Login, refresh, RBAC exist; no usable first-admin bootstrap or reset workflow | Partially working |
| Customers | List/create/update/delete use the API; nested empty arrays crash rendering | CRUD and nested site/contact routes exist; optional CR uniqueness issue | Closest to integration, still blocked |
| Opportunities | API-backed list/write; customer and assignee choices are mocks | Create omits customer/date and other UI fields | Blocked |
| Quotes list | Fetches API but renders local sample quotes | List and lifecycle routes exist | Demonstration UI plus API scaffold |
| Quote builder | Static customer/number/catalog/rows; save and submit only update local state | No complete line editing/version/snapshot contract | Not persisted |
| Contracts | Fetches API but renders and mutates local contracts | Create omits customer/dates; lifecycle is incomplete | Not integrated |
| Manufacturers | Local sample records/categories despite fetch | Basic CRUD; category management absent from router | Not integrated |
| Products | Local catalog, vendor entries, prices, attachments despite fetch | Basic CRUD; create drops manufacturer/category/cost inputs | Not integrated |
| Price books | Local books, mixed product/service catalog, local save | Header CRUD; entry writes/customer assignment incomplete | Not integrated |
| Exchange rates | Local rates and history despite fetch | List/get/update; no creation/seed path; effective date ignored | Not integrated |
| Recurring services | Local services/calculations despite fetch | Basic CRUD, incomplete calculation consistency | Not integrated |
| Inventory | Local stock/reservations/movements despite fetch | Reservation/transfer defects; release/fulfillment absent | Not integrated |
| Procurement | Uses the store, but has no initial fetch on direct navigation | Header-only PO/SQ creates; receipt processing incomplete | Partially wired, unsafe to use |
| Projects | Local sample projects and mutations | Basic CRUD from quote; missing accepted-state validation | Not integrated |
| Teams / users | Teams use local data; no dedicated user-management route | CRUD and roles exist; member/leader workflow incomplete | Not integrated |
| Documents | Local document metadata/upload UI despite fetch | Disk upload/download exists; safety and linking gaps | Not integrated |
| Dashboard | Live KPIs/pipeline mixed with static quotes/activity/expiries/rep results | Aggregates exist; no activity-log writes found | Misleading mix of real and sample data |

Source examples: `QuotesView.vue:163`, `ContractsView.vue:72`, `ProductsView.vue:69`, `ManufacturersView.vue:39`, `PriceBooksView.vue:114`, `ExchangeRatesView.vue:37`, `RecurringServicesView.vue:71`, `InventoryView.vue:58`, `TeamsView.vue:47`, `DocumentsView.vue:72`, and `ProjectsView.vue:23` under `crm-dashboard/src/views/`.

## Detailed findings

### R01 · P0 · The checked frontend build fails

**Evidence:** 51 TypeScript diagnostics: QuoteBuilder 13, PriceBooks 12, Procurement 10, Inventory 5, Products 5, RecurringServices 5, and useTheme 1. They include missing discriminant fields, unsafe array indexing, untyped procurement items, and `DataTransfer` referenced incorrectly from templates. `crm-dashboard/Dockerfile` runs `build-only`, masking this gate.

**Impact:** The current build cannot serve as a reliable release baseline; some diagnostics correspond to reachable runtime faults.

**Action / acceptance:** Correct the types and real edge cases without disabling strictness or replacing everything with `any`. Both `npm run build` and the frontend image build must pass the checked build.

### R02 · P0 · Core screens do not persist the data they show

**Evidence:** The module matrix above. `QuoteBuilderView.vue:200` fixes the quote number/customer in local refs; lines 686–687 only set a status and show a toast. `QuotesView.vue:382` creates a quote locally and navigates to a generated ID. The builder never loads that ID. Contract create/edit/delete are local functions at `ContractsView.vue:228` onward.

**Impact:** Users see “saved” records that disappear on reload; different modules show different versions of the business. The browser showed six sample quotes while the database contained one review fixture.

**Action / acceptance:** One API-backed source per module; remove production sample arrays, replace mock lookups, fetch detail by route ID, await successful saves, and retain input on failure. A second session and a hard reload must show the same saved record. Keep demonstration fixtures only behind an explicit development/demo mode if needed.

### R03 · P0 · Create handlers discard required relationships

**Evidence:** `crm-api/internal/handlers/opportunities.go:48`, `quotes.go:49`, and `contracts.go:39` accept `customerId` without assigning the model's `CustomerID`. The quote handler also drops opportunity/valid-until and line product/SKU data. Product create drops manufacturer/category and several costing inputs. Reservation/movement/receipt construction similarly omits required IDs.

**Verified:** Creating an opportunity, quote, or contract with a valid existing customer returned HTTP 500 against the migrated database.

**Action / acceptance:** Explicit request-to-model mapping, UUID/date validation, relationship existence and consistency checks. Valid creates must persist every supported field; invalid references must return a useful 4xx and leave no partial records.

### R04 · P0 · Reloading a session breaks the user profile

**Evidence:** `crm-api/internal/handlers/auth.go:143` returns only ID/role/email for `/auth/me`. `crm-dashboard/src/stores/auth.ts:17` accesses `firstName[0]` / `lastName[0]`. Login returns a different, fuller shape.

**Verified:** Reload produced `undefined undefined` in the header and `Cannot read properties of undefined (reading '0')` in the console.

**Action / acceptance:** Share a safe user response shape between login/refresh/me; load the active user from the database, guard missing initials, and test cold navigation with valid, expired, and revoked sessions.

### R05 · P0 · API response shapes do not match frontend types

**Evidence:** Customer model arrays have `omitempty` (`crm-api/internal/models/customer.go:55` onward), while `CustomersView.vue:208` accesses `c.contacts.length` and `c.sites.length`. Opportunity/product/quote models return nested relationships; frontend types expect flattened fields such as `customerName`, `manufacturerName`, and `quoteIds`. Several other services use `unknown[]` or `any[]` instead of an agreed contract.

**Verified:** A real customer with no contacts/sites caused `Cannot read properties of undefined (reading 'length')` and a blank customer content area.

**Action / acceptance:** Define explicit list/detail/create/update DTOs, normalize empty collections to `[]`, define nullable dates/relationships, and use response adapters or shared/generated types. Cover empty and populated records with API-to-UI contract tests. A TypeScript cast alone does not transform the server response.

### R06 · P0 · Generic updates bypass approvals

**Evidence:** `crm-api/internal/handlers/quotes.go:124` accepts arbitrary map fields through `Updates(req)`. A draft check does not prevent changing the draft directly to approved or overwriting totals. `procurement.go:85` permits any status via the ordinary update permission, separate from the approval permission. Contracts also permit unrestricted field updates.

**Verified:** An authenticated **sales_executive** sent `{"status":"approved","total":999}` to generic quote PATCH, received 200, and a subsequent GET confirmed both changes persisted.

**Action / acceptance:** Use field allowlists, keep status/computed totals/approval metadata out of generic PATCH, and enforce transitions in a shared service. Ordinary sales/procurement roles must be unable to approve through any route. Approval must record actor, timestamp, policy result, and the exact approved revision.

### R07 · P1 · Declared request validation is inactive

**Evidence:** `crm-api/pkg/validator/validator.go:31` only calls `ShouldBindJSON`. Request structs use `validate:"..."`; the installed Gin binder uses the `binding` tag and Setup does not change it or explicitly invoke validation. Nested item slices also need item-level traversal.

**Verified:** Passing `{}` to the existing helper with a `validate:"required,email"` field returned success.

**Action / acceptance:** Select one validation mechanism and apply it consistently. Test missing/blank fields, malformed UUIDs/dates, invalid enums, negative values, percentages outside their range, and invalid nested items. Distinguish omitted values from valid zero values, including an explicitly supplied zero tax rate.

### R08 · P1 · Sort parameters reach SQL unescaped

**Evidence:** `crm-api/pkg/pagination/pagination.go:28` accepts arbitrary `sort`; handlers pass `params.Sort + " " + params.Order` to GORM `Order`. The review's harmless dry-run probe preserved a subquery in the generated ORDER BY SQL.

**Impact:** Authenticated list endpoints expose an SQL injection surface. This review did not run destructive injection payloads. GORM explicitly lists `Order` among methods whose inputs are not escaped. [GORM security documentation](https://gorm.io/docs/security.html).

**Action / acceptance:** Map public sort keys to a per-resource allowlist of columns and construct ordering from that mapping; reject unknown keys. Add a stable ID tie-breaker. Tests must reject SQL expressions while preserving supported sorting.

### R09 · P1 · Database failures and partial writes can be reported as success

**Evidence:** Many handlers ignore `ShouldBindJSON`, `Create`, `Updates`, `Save`, `Delete`, `Count`, and `Find` errors. Quote header/lines/totals are written separately; renewal marks the old contract renewed before checking new-contract creation. PO/SQ creates ignore number-generation and insert errors.

**Impact:** A 200/201 can be returned for an operation that failed; multi-step writes can leave inconsistent records. Zero affected rows are often treated as success.

**Action / acceptance:** Shared error handling, explicit rows-affected checks, request contexts, and transactions around each business operation. Inject a failing line/renewal/receipt write and prove rollback plus an accurate response. Return 404 only for genuinely absent records, not arbitrary database failures.

### R10 · P1 · The quote schema cannot represent the existing builder

**Evidence:** `QuoteBuilderView.vue:179` supports row kinds, product/service/recurring/write-in sources, multiplier, line discounts, optional/selected/printable flags, and sold-to/ship-to details. `crm-api/internal/models/quote.go:88` and migration line 294 store a much smaller model. Price-book entries require a product FK even though the UI includes services and recurring services.

**Impact:** Simply wiring save would lose data or calculate different totals. A recurring row with quantity 1 and multiplier 12 is not equivalent to the backend's quantity × unit price. Quote PATCH does not provide a complete line replacement/edit operation.

**Action / acceptance:** Agree a versioned quote DTO and schema before integration. Persist supported row semantics and price/cost/FX/tax/contact snapshots. Calculate authoritative totals on the server using a defined decimal/rounding policy; mirror previews in the UI. Preserve margin on discounted revenue before tax (the current model and builder do this; the old project document does not). Test recurring multipliers, line/header discounts, optional rows, zero revenue, and currency changes. The displayed currency must correspond to converted values, not just a changed label.

**Versioning:** Add quote-family/revision linkage and unique revision numbering, immutable sent/approved snapshots, and optimistic concurrency. Existing `version: 1` alone is not revision management. Recalculate currently permits all states; locked documents must not silently change.

### R11 · P1 · Sales lifecycle rules are incomplete

**Evidence:** Quote submit always moves to pending approval while thresholds appear only in frontend calculations. There is no accepted-quote-to-contract operation. Contract activate/terminate lack prior-state/date checks; renewal is non-atomic and lacks a new end date. Opportunity stages are freely writable. Project create accepts a quote without requiring acceptance. `QuotesView.vue:460` allows editing approved/pending quotes although the API only allows draft editing.

**Action / acceptance:** Specify legal transitions and required fields centrally. Convert accepted quote → contract once, with preserved customer/revision/value; repeating a request must not duplicate it. Require loss/termination reasons, maintain stage history, implement expiry behavior, and test illegal and repeated transitions. Clarify whether quote acceptance automatically closes the opportunity or needs a sales decision.

### R12 · P1 · Inventory operations are not internally consistent

**Evidence:** `crm-api/internal/handlers/inventory.go:39` omits the reservation product ID and checks availability outside the transaction without a lock. Transfer omits movement product ID, ignores source rows affected, calls an UPDATE rather than the commented upsert at destination, and does not maintain available quantity/value. SQL stores on-hand, reserved, available, and total value independently.

**Impact:** Current foreign keys block some operations; after those mappings are repaired, overselling and quantity drift remain. Missing source stock must never create destination stock.

**Action / acceptance:** Product+warehouse identity, atomic conditional updates or row locks, explicit stock movement records, and one defined source for derived availability/value. Add reserve/release/fulfill/adjust/transfer flows with idempotency. Test two competing reservations, insufficient stock, new destination, same-warehouse transfer, release twice, and `available = on_hand - reserved`.

### R13 · P1 · Procurement does not complete a purchase/receipt cycle

**Evidence:** `crm-api/internal/handlers/procurement.go:44` and `:120` create headers without item payloads. Supplier quote conversion requires accepted status but no update/accept route exists. Goods receipt omits PO/product/PO-item IDs and updates stock by product across all warehouses; it does not reconcile ordered/received quantities. `crm-dashboard/src/stores/procurement.ts:130` updates by calling create, while deletes only remove local rows. ProcurementView has no initial fetch.

**Action / acceptance:** Add real edit/status/item APIs; persist lines/totals/source linkage; validate receipt quantities against outstanding approved PO lines and target warehouse. A partial receipt must update exactly that warehouse and PO, retain damaged quantities appropriately, and be safe to retry. A page opened directly must load the same data as one reached through Inventory.

### R14 · P1 · Database constraints need business-level consistency

**Verified:** Two customer creates with no CR number yielded 201 then 409. The handler writes `""` into a UNIQUE `cr_number` rather than storing absence as NULL.

**Other evidence:** Migration lines 85, 116, 159, 340, 355, 576 onward lack several cross-entity/range safeguards. Contacts can reference another customer's site; a product can reference a category from another manufacturer; `price_books.contract_id` has no FK. Multiple primary contacts and duplicate price-book product entries are possible. Available stock is unconstrained relative to reserved/on-hand. Most money/quantity/date ranges are not checked.

**Action / acceptance:** Apply the schema changes in the database section below through additive migrations. Decide natural-key reuse under soft deletion explicitly. Test invalid relationships and concurrent duplicates at the database boundary, not only through UI checks.

### R15 · P1 · Business-number generation fails at year rollover

**Evidence:** `crm-api/pkg/seqgen/seqgen.go:38` inserts a new row for a new year, but the SQL primary key is only `name`. Existing seeded sequences are dated 2026.

**Verified:** Moving the review project sequence to the prior year and requesting the next number failed with `sequences_pkey` duplicate-key error.

**Action / acceptance:** Use a `(name, year)` key with an atomic increment/upsert, or a deliberate locked rollover strategy. Test first number in a new year, parallel requests, formatting, and rollback/error propagation. Numbers must always be assigned by the server.

### R16 · P1 · Session and permission behavior needs hardening

**Evidence:** Refresh rotation reads/revokes/inserts without a transaction or replay guard (`auth_service.go:93`). Access middleware trusts token role until expiry and does not check active status. Password change does not revoke refresh sessions. No login throttling exists. The frontend interceptor applies refresh handling to login failures, has a no-refresh-token branch before its `finally`, and its refresh request has no explicit timeout. Routes/sidebar/actions have no role-aware UI gating.

**Action / acceptance:** Atomic refresh rotation and a defined logout/deactivation/role-change policy; bounded request timeouts; exclude login/reset from refresh retries; throttle authentication; clear user-scoped stores on logout/account switch. Permission-aware navigation must match server enforcement. Record the acceptable access-token revocation delay and whether users see all company records or assigned records only. Test two concurrent refreshes and a disabled user. Review token storage choice alongside the XSS findings; do not assume localStorage alone constitutes a demonstrated compromise.

### R17 · P1 · Files and exports require safety and persistence work

**Evidence:** `crm-api/internal/handlers/documents.go:54` uses multipart memory sizing, not a hard body-size limit, and ignores configured `MaxUploadSize`. `:132` includes unvalidated category text in a filesystem path before database enum validation. Type is taken from the client header; file/metadata deletion errors are ignored. No link/version-write API exists. The frontend's document download URL has no Bearer header handling. `QuoteBuilderView.vue:689` interpolates editable strings into HTML and writes them into a popup; CSV quoting does not neutralize formula-leading cells.

**Impact:** Arbitrary path segments can escape the intended document folder, uploads can consume excessive resources, and print/export can interpret untrusted input. Path write/deletion attempts are possible before an invalid category is rejected by the DB; this review did not exploit the filesystem.

**Action / acceptance:** Allowlisted categories/types, root-contained server-generated paths, hard request/file limits, authenticated blob downloads, metadata/file consistency, and real entity linking. Escape all print text and apply spreadsheet-formula-safe export handling. Test traversal, oversized/mislabelled files, lost files, authorization, and literal HTML/formula text. Coordinate the reverse-proxy upload limit with the API limit.

### R18 · P1 · Clean installation cannot bootstrap the application as documented

**Evidence:** Root `Makefile` runs `migrate` inside the API image, but `crm-api/Dockerfile` never installs that binary. `crm-api/Makefile` points seed at absent `seeds/main.go`. Registration requires an already-authenticated admin; the migration seeds only sequences. API health returns OK without schema readiness. Redis is deployed but no runtime client/use was found. Docker starts from Go 1.22 while go.mod requires 1.25.4; automatic toolchain download may hide this mismatch, so it is a reproducibility risk rather than a confirmed image-build failure.

**Action / acceptance:** A clean bootstrap path with migration runner, explicit first-admin provisioning, seed/reference-data policy, matching pinned toolchain, safe key generation that does not overwrite existing keys, and readiness that checks required dependencies/schema. Remove unused services or justify their purpose. Prove clean installation and an upgrade from the prior schema without losing data.

### R19 · P1 · Dependencies need a reviewed update pass

**Evidence:** The audit snapshot reports 13 affected-package entries. `xlsx` is direct and has no automatic fix reported; direct axios and Vite entries also need review. Some reports affect dev tooling or Node-only paths and must not be presented as proven browser exploits.

**Action / acceptance:** Triage production vs development reachability, update compatible packages, review the lockfile, and retest importing files and API calls. SheetJS identifies its CDN as the authoritative distribution source and the npm registry package as outdated; decide a maintained import path or replacement deliberately. [SheetJS installation guidance](https://docs.sheetjs.com/docs/getting-started/installation/nodejs/). No dependency changes were made during review. Go vulnerability scanning remains to be run.

### R20 · P1 · Tables search only a loaded page and hide failures

**Evidence:** API pagination defaults to 25 (`pkg/pagination/pagination.go:27`), but customer/opportunity views fetch once and filter locally without using `meta` for pagination. Customer UI says “No customers found” for empty data, including failed loading; save/delete errors only reach the console. Common labels/counts summarize the loaded array, not the total database.

**Impact:** Record 26 can exist but be impossible to find from the screen. Users cannot distinguish an empty dataset from an API failure. Repeated save clicks can create duplicates.

**Action / acceptance:** Server-side query/filter/sort/page state, debounced search, real total counts, dedicated lookup APIs, visible loading/error/retry/empty states, pending-save protection, and clear delete/archive confirmation. Test at least 60 records, a record outside page one, filtered page reset, 403, 500, and slow requests. KPIs must come from aggregates, not paginated lists.

### R21 · P1 · Mobile layout and keyboard accessibility are incomplete

**Verified:** At 390 × 844, the quotations page retained its full sidebar, leaving a narrow strip for content; document scroll width was 1686 px. Global responsive CSS targets `.app-sidebar`/`.content-area`/`.top-bar`, while actual components use `.sidebar`/`.app-content`/`.app-header` and scoped margins (`main.css:1991`, `MainLayout.vue:41`).

**Source evidence:** Modal divs lack a shared dialog/focus/Escape pattern; many labels are not associated with inputs; header/profile controls include unnamed buttons/clickable divs. Row icon buttons are 28 × 28 (`main.css:779`). The login password toggle has `tabindex="-1"`. No reduced-motion handling was found.

**Action / acceptance:** Responsive drawer navigation, actual-class breakpoints, named controls, associated labels, semantic dialogs with focus restoration, keyboard-operable menus, visible focus, and larger practical touch targets. Confine wide-table scrolling to the table. Test desktop/tablet/phone plus keyboard and zoom; contrast and screen-reader compliance still require a dedicated pass. Suggested follow-up: `$impeccable adapt` and `$impeccable harden`.

### R22 · P2 · Dashboard totals need clear business definitions

**Evidence:** Dashboard API calls accepted quote totals “revenue” and sums totals across currencies without conversion (`dashboard.go:14`). The UI labels the result SAR. Static activity, recent quotes, expiry examples, and rep performance remain in `DashboardView.vue:144` onward. Quarter changes are fixed at zero. No activity-log writer exists. Expiring-contract queries include past dates as well as the upcoming interval.

**Action / acceptance:** Label accepted quote totals as booked/accepted sales unless invoicing defines revenue; aggregate comparable currencies using an agreed reporting rate; define tax inclusion and date boundaries. Make sample widgets real or remove them. Use weighted aggregate margin instead of averaging quote percentages when reporting portfolio profitability. Fixture-backed totals must reconcile with list/detail records and period filters. This is a product metric recommendation, not accounting or tax advice.

### R23 · P2 · Preserve the visual system, simplify the busy tables and shell

**Observed:** Dark panels, consistent spacing, status text/badges, recognizable navigation, and clear module groupings provide a useful base. Quote IDs wrapped over three lines on desktop; five row actions and many financial columns compete for width. Login renders inside the signed-in shell (`App.vue:7`). Header search/⌘K, notifications, settings, profile/preferences and forgot-password affordances are incomplete.

**Action / acceptance:** Apply the table plan below, render login in a public layout, and connect or remove nonfunctional controls. Consolidate existing tokens; login hardcodes its palette while much of the app uses theme variables. Suggested follow-up: `$impeccable clarify`, then `$impeccable polish` after integration and responsive fixes.

### R24 · P1 · Release verification and server operations are not established

**Evidence:** No automated tests/CI were found. Root `git status` reports no repository; no child Git metadata was found in the inspected tree. The current server configuration exposes HTTP only and does not describe TLS termination, backups/restoration, monitoring, or a release/rollback procedure. SMTP/Redis configuration exists without completed workflows.

**Action / acceptance:** Establish version control and CI, meaningful sales-flow and permission tests, migration tests, server-specific configuration, TLS, private DB networking, secret provisioning, readiness, logs, database+upload backup/restore, and rollback. Keep Vite as a development tool. Provider-specific deployment comes after server details and release gates are known; this review is not a deployment authorization or a deployment result.

## Database/table design recommendations

Retain PostgreSQL, UUID identities, foreign keys, separate line-item tables, and NUMERIC storage. Complete the model rather than replacing it. SQL migrations should remain the schema authority; do not add production AutoMigrate as a shortcut around missing migrations.

| Area | Recommended design | Why / acceptance |
|---|---|---|
| Customers | Nullable normalized CR/VAT values; explicit active-record uniqueness policy | Multiple prospects may lack legal IDs. Define whether archived legal IDs can be reused. |
| Sites / contacts | FK/constraint or transactional validation ensuring site belongs to customer; defined primary-contact scope | Prevent mixed-customer contacts; enforce at most one primary per chosen scope. |
| Opportunities | Persist owner/pre-sales, service types, dates and cost; stage event/history rows | Support real assignment, reporting, and loss/win reasons. |
| Quotes / revisions | Family ID + revision number; unique revision key; frozen approved/sent snapshot; concurrency version | Preserve the actual commercial document approved/sent and detect stale edits. |
| Quote rows | Row kind, source kind, relevant product/service FK, quantity/unit/multiplier, discounts, flags, order and snapshot descriptions/costs | Persist existing builder behavior. Use constraints so only relevant source references are populated. |
| Pricing | Explicit decimal policy, currency and FX snapshot; server-owned derived totals | Round deliberately at agreed stages; historic quotes must not change with catalog prices. |
| Price books | Contract FK; real customer/date eligibility; entries capable of product/service/recurring references; tier quantities if volume pricing ships | Header “type” alone does not implement volume or customer pricing. Prevent overlapping/duplicate entries according to policy. |
| Contracts | Accepted revision FK, dates, currency/value snapshot, renewal linkage and activation/termination events | Prevent accidental duplicate conversions/renewals and preserve evidence. |
| Approvals / audit | Append-only approval and business-event records with actor/time/revision/reason | Existing activity table is unused; it should not be an editable generic CRUD record. |
| Stock | Product+warehouse uniqueness, nonnegative/availability constraints, immutable movements with source linkage | Derive availability or maintain all quantities transactionally; stock audit must reconcile. |
| Warehouses | Consider a reference table instead of a fixed three-value enum | Makes adding a real warehouse an administrative operation. Not a prerequisite if the three locations are fixed. |
| Procurement | Item-level PO/SQ/receipt linkage; received/accepted/damaged quantities; target warehouse; idempotency key | Partial deliveries and retry safety need durable references. |
| Sequences | Atomic generator with a `(name, year)` key or equivalent locked rollover | Prevent new-year failure and collisions. |
| Documents | Object metadata + version/links; contained storage key; lifecycle/retention status | Restore DB and files together; validate permitted entity links. |
| Query performance | Add indexes for actual FK/filter/date/sort paths, including missing child FKs and token lookup; inspect query plans with realistic data | Avoid adding speculative indexes everywhere; `%term%` search requires a deliberate plan. |

Add CHECK constraints for positive quantities, allowed percentage ranges, nonnegative prices/costs where appropriate, date ordering, positive FX rates, and `reserved <= on_hand`. Define whether fractional quantities are needed for labor/cable units before freezing integer quantity fields.

**Migration approach:** inventory any existing customer data first; add a new numbered migration; report and resolve invalid/duplicate data before constraints; backfill with a reversible plan; validate on a copy; test upgrade and rollback. Do not rewrite migration 000001 on a server that has already applied it. The review only used a disposable database and cannot establish whether another environment contains real data.

## Frontend table design plan

The interface is an operational CRM: prioritise finding a record, understanding its state, and taking the next allowed action.

| Table | Default visible information | Move to details / optional columns |
|---|---|---|
| Customers | Company, primary contact, sector/region, status, assigned owner, next action | Full address, legal IDs, all sites/contacts, notes |
| Opportunities | Title + customer, stage, owner, estimated value, probability, expected close, next activity | All service types, full notes, linked quote history |
| Quotes | Number + revision, customer, status, total with currency, margin if permitted, valid until, updated date | Subtotal, VAT, discount, line count, approval history |
| Contracts | Number/title + customer, status, type, term dates, value/currency, renewal due | Full terms, documents, renewal chain, detailed SLA |
| Inventory | Product/SKU, warehouse, on-hand, reserved, available, reorder indicator | Value/cost by permission, movement ledger, reservation owners |
| Procurement | Number, supplier, status, total/currency, expected delivery, received progress | Every line's costing, shipment allocations, detailed receipt history |

Use nowrap/minimum width for identifiers, a readable customer/title column, right-aligned tabular monetary values, sticky table headers, and an explicit Actions header. Put View/Edit as the primary row action and secondary actions in a named menu; show only allowed lifecycle actions. Use status text as well as color. Persist useful filters/columns, display active filters and real totals, and preserve list position after viewing a record. Column selection/saved views are improvements after query correctness, not substitutes for it.

On narrow screens show identity, status, total and the primary action first; expose secondary information in details or a table-local scroll region. A long quotation editor may remain a wide workspace, but navigation and Save/validation feedback must remain reachable. Add unsaved-change protection and display last-saved state only after server acknowledgment.

### Provisional interface health assessment

This is a scoped engineering assessment, not a WCAG certification or measured performance score.

| Dimension | Score / 4 | Basis |
|---|---:|---|
| Accessibility | 1 | Basic buttons/tables exist; dialog, labeling and keyboard gaps are repeated. |
| Performance | 2 | Route lazy loading exists; parser bundle is large; runtime profiling remains untested. |
| Responsive design | 1 | Breakpoints exist but the observed mobile shell is unusable. |
| Theming | 2 | Shared light/dark tokens exist; hardcoded exceptions and stale class rules remain. |
| Implementation integrity | 1 | UI is visually coherent but sample data and false persistence undermine operational trust. |
| **Total** | **7 / 20, provisional** | **Major functional/accessibility work; preserve the visual direction.** |

After fixes, use `$impeccable audit` for a bounded follow-up verification and `$impeccable polish` as the final visual pass. The owner can choose individual follow-ups or a grouped pass.

## What should be preserved

- Vue 3 + TypeScript + Pinia + service-layer separation and lazy-loaded routes.
- The existing navigation, visual tokens, table styling, and detailed quote-builder concepts.
- Go/Gin/PostgreSQL, authenticated route groups, explicit permission names, bcrypt password hashes, RSA-signed access tokens, and hashed refresh-token storage.
- SQL foreign keys, NUMERIC columns, transaction use where already present, and server-side sequence intent.
- The current pre-tax margin basis in quote calculations; replace the conflicting historical documentation formula.

## Decisions still needed, without blocking foundational repairs

1. Server OS/provider/domain and whether access is public internet, VPN, or an internal network.
2. Initial users/roles; all-company versus assigned-record visibility; who can see costs and margins.
3. Confirm the existing SAR/15% defaults, supported currencies, date/timezone behavior, rounding, and approval thresholds as business configuration. These values were observed in code, not validated as legal/tax requirements.
4. Approval delegation/self-approval rules, quote revision rules, acceptance evidence, and opportunity closure policy.
5. Existing data to import/preserve, real warehouse locations, and required quote branding/output.
6. Launch date and which non-sales modules must be enabled at first release.

The selected first delivery is the complete persisted sales workflow. Inventory, procurement, projects and remaining operations stay in the completion backlog with explicit gates; unfinished modules should not appear usable in a production release.
