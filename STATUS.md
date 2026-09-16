# G4S CRM — current status and work plan

**Last updated:** 2026-09-16  
**Current phase:** API/frontend integration batch implemented and verified; M1–M5 acceptance work remains.
**Release state:** not ready for production.  
**Primary objective:** make customers → opportunities → quotes → contracts work end to end, then finish remaining modules and deploy to the owner's server.  
**Detailed findings:** [docs/CRM_REVIEW.md](docs/CRM_REVIEW.md).
**Current API coverage:** [docs/API_COVERAGE.md](docs/API_COVERAGE.md); [179-route inventory](docs/API_ENDPOINTS.md).

## Start here each work session

1. Read this file and the findings linked to the next task.
2. Inspect current changes and preserve existing work. Git baseline now exists; compare changes against the baseline commit before substantive work.
3. Mark one bounded task `IN_PROGRESS`; record assumptions and dependencies.
4. Implement and run the task's acceptance checks. An existing screen or endpoint is not completion.
5. Update this file with actual results, files changed, remaining issues and the next action. Mark `DONE` only with evidence. Keep review findings as historical context rather than silently rewriting them as fixed.

Status values: `TODO`, `IN_PROGRESS`, `PARTIAL` (implemented subset; see remaining acceptance gaps), `BLOCKED` (name the dependency), `DONE`, `DEFERRED` (record the decision). Review recommendations are not completion; the implementation notes below distinguish finished fixes from partially addressed tasks.

## Confirmed direction and open decisions

| Decision | State |
|---|---|
| First working delivery | **Confirmed by owner:** sales flow: customers → opportunities → quotes → contracts. |
| Deployment | **Confirmed by owner:** a server; not limited to the old local-server assumption. |
| Preserve frontend direction | Working approach: keep existing visual design and improve behavior/table usability. |
| Single company | Current model/old plan assumption; no multitenancy implementation or requirement confirmed. |
| Server/provider/domain/access model | Open; required before deployment configuration is finalized. |
| Roles/data visibility/approval policy | Open; current RBAC is a starting point, not an approved policy. |
| Currency/tax/rounding | Existing SAR/15% defaults observed; business confirmation needed. |
| Existing real data and launch date | Unknown. Do not assume a production DB is empty. |

## Historical baseline — before implementation

| Area | State |
|---|---|
| Frontend | 17 view files; extensive UI; many local sample datasets and local-only mutations. |
| Frontend checked build | **FAIL** — 51 TypeScript diagnostics in 7 files. |
| Frontend bundle only | PASS; parser chunk 845.47 kB / 265.38 kB gzip warning. |
| Backend | Go API compiles; `go vet ./...` passes; no Go test files. |
| Database | 36 tables; initial migration applied successfully to isolated PostgreSQL 16. |
| Core sales creation | **FAIL** — valid opportunity, quote and contract create requests returned 500. |
| Authorization | **FAIL** — sales executive could directly approve a draft through generic PATCH. |
| Session reload | **FAIL** — `/auth/me` omits names required by frontend. |
| Customer table | **FAIL** — empty contacts/sites omitted by API cause render exception. |
| Validation | **FAIL** — required validation tag ignored by existing binding helper. |
| Number rollover | **FAIL** — prior-year sequence causes duplicate primary key error. |
| Mobile | **FAIL** — 390 px viewport retains desktop sidebar; document width 1686 px. |
| npm audit | 13 affected-package entries; reachability triage pending. |
| Production installation/restore/load testing | Not verified. |

Logs and reproduction details: [review evidence](docs/review-evidence/README.md). Go compilation success is not a passing integration test suite.

## Delivery milestones

| Milestone | Objective | Exit gate | State |
|---|---|---|---|
| M0 — Review | Establish actual implementation state and organized plan | Evidence, prioritized findings and status file exist | **DONE** |
| M1 — Reliable foundation | Checked builds, isolated setup, validation, safe updates, stable session/DTO contracts | Clean install; no approval bypass; refresh/customer smoke tests pass | IN_PROGRESS |
| M2 — Customers and opportunities | Persist customers/sites/contacts and real opportunities/assignees | Cross-session CRUD, pagination/search and stage tests pass | PARTIAL |
| M3 — Quotes | API-backed catalog/list/builder with authoritative calculations, approvals and revisions | Save/reload/send/accept flow and role/concurrency tests pass | PARTIAL |
| M4 — Contracts and sales reporting | Convert accepted revision to contract; lifecycle, files and truthful dashboard | Full sales journey passes with database persistence and audit trail | PARTIAL |
| M5 — Remaining modules | Procurement, inventory, projects, teams, recurring/price-book depth | Each workflow passes its own functional and integrity gates | PARTIAL |
| M6 — Server release | Harden, deploy and prove operations | TLS, migrations, backup+restore, monitoring and rollback verified | TODO |

M6 can follow M4 only if the owner chooses a sales-only first release and incomplete modules are explicitly disabled. Otherwise M5 must complete first. Sales-only release scope has not yet been confirmed; priority alone does not remove modules from scope.

## Ordered work backlog

| Task | Milestone | Concrete deliverable / acceptance | Findings | State |
|---|---|---|---|---|
| T01 | M1 | Establish Git baseline/root ignore rules and reproducible local environment; configurable ports that do not collide with other projects | R18, R24 | DONE |
| T02 | M1 | Fix 51 type diagnostics without disabling strict checks; use checked build in Docker/CI | R01 | DONE |
| T03 | M1 | Correct validation helper and nested field validation; consistent error/rows-affected handling; whitelist sort keys | R07–R09 | IN_PROGRESS |
| T04 | M1 | Remove generic status/approval/total updates; test each role against bypasses | R06, R11 | PARTIAL |
| T05 | M1 | Real migration runner + first-admin bootstrap + reference data; match Go toolchain; verify clean startup/readiness | R18 | PARTIAL |
| T06 | M1 | Shared user/collection DTOs; safe `/auth/me`; bounded refresh, session policy, logout reset and role-aware UI | R04, R05, R16 | PARTIAL |
| T07 | M1 | Add test/CI structure; preserve reproduced failures as regression cases; triage and update dependencies | R19, R24 | PARTIAL |
| T08 | M2 | Customer CRUD plus site/contact UI/API; optional CR handling; relationship/primary-contact constraints; visible errors | R03, R05, R14, R20 | PARTIAL |
| T09 | M2 | Opportunity request mapping, live customer/user lookups, owners/services/dates/costs, stage events and validation | R02, R03, R11 | PARTIAL |
| T10 | M2 | Shared server-driven list query pattern with total/page/filter/sort; lookup endpoints and 60-record tests | R08, R20 | PARTIAL |
| T11 | M3 | Agree quote row/revision/decimal/FX DTO and add migration; resolve price-book/service representations | R10, R14 | PARTIAL |
| T12 | M3 | Persist the catalog subset required for quotes: manufacturer/category/product/service costing and eligible pricing | R02, R03, R10 | PARTIAL |
| T13 | M3 | Replace quote list/builder mocks; load by ID; atomic header+line save; reload and second-user persistence | R02, R03, R05, R09 | PARTIAL |
| T14 | M3 | Server calculations and threshold policy; approval/rejection/send/accept/decline; immutable revision and stale-edit checks | R06, R10, R11 | PARTIAL |
| T15 | M3 | Fix yearly number generation; test rollover and concurrency; remove frontend number assignment | R15 | PARTIAL |
| T16 | M3 | Safe, accurate customer print/export; dirty-state/pending-save feedback; permission-controlled costs | R10, R17, R23 | PARTIAL |
| T17 | M4 | Accepted quote → contract conversion once; dates/value snapshot; activate/terminate/renew with history and transactions | R03, R09–R11 | PARTIAL |
| T18 | M4 | Real document upload/download/version/link behavior with limits and contained paths | R17 | PARTIAL |
| T19 | M4 | Real dashboard activity/quotes/expiry and defined currency/date/metric semantics | R02, R22 | PARTIAL |
| T20 | M4 | Fix mobile shell, common dialogs/form labels/keyboard access, table columns/actions and public login layout | R20, R21, R23 | PARTIAL |
| T21 | M5 | PO/SQ item editing, acceptance, approvals/conversion and partial goods receipts; correct update/delete routes | R09, R13 | PARTIAL |
| T22 | M5 | Stock receipt/reserve/release/fulfill/transfer/adjustment with concurrency, availability and movement reconciliation | R12, R14 | PARTIAL |
| T23 | M5 | Complete projects, teams/user management, price books, FX history and recurring service behavior | R02, R10, R11, R14 | PARTIAL |
| T24 | M6 | Server runbook/configuration/TLS/secrets/networking/readiness/logging; repeatable deploy and rollback | R16, R18, R24 | TODO |
| T25 | M6 | Full regression/UAT, role tests, migration upgrade/rollback, backup+upload restore and realistic load checks | All relevant findings | TODO |

Dependencies: T03/T04 protect all mutations. T05 enables repeatable environments. T06 precedes live UI integration. T11 precedes T13/T14. T17 requires an accepted persisted quote revision. T21/T22 depend on agreed catalog/stock contracts. T24/T25 require actual server details. Work on those independent foundations need not wait for optional design decisions.

## Current API/frontend integration batch — 2026-09-16

**Batch state: DONE for the implementation and checks below. Overall modules remain PARTIAL.** The [coverage matrix](docs/API_COVERAGE.md) is the current source for implemented behavior versus unfinished UI and release work.

- **Sales:** live opportunities/customer/user lookups; actual quotation list/catalog; atomic full builder saves, saved row order/false flags, authoritative rounded totals, stale-edit protection, revisions and lifecycle actions. Browser verified save/reload → submit → approve → sent → accepted. Submitted inputs are disabled. Accepted quotes convert once to a dated contract with amount/currency snapshot; activation, termination and renewal persist.
- **Relationships/catalog:** typed writes for manufacturers/categories, products/vendors/history/files, service catalog, price-book headers/mixed entries, recurring component costing, teams/membership/users and project managers. Site/contact APIs enforce parent ownership and primary-contact consistency; their edit UI remains open.
- **Inventory/procurement:** corrected PO/SQ PATCH/delete routing and real line persistence; approval/ordering/SQ conversion, partial receipts, stock valuation, transfer/adjustment/reserve/release/fulfill. Concurrency tests confirm no over-reservation; foreign-currency receipt test confirms SAR landed value. Generated document numbers come from the server.
- **Documents/reporting:** actual uploads, authenticated downloads, version snapshots and validated entity links; truthful dashboard collections/activity. Basic document UI connected; advanced versions/link controls remain open. Dashboard accepted-value sums are SAR-only and labeled accordingly.
- **API contracts:** added/updated missing routes and presentation DTOs, safe empty arrays and editable-field adapters. Locally filtered tables load all pages; full server-driven table pagination is still required for scale. Added migrations 000003/000004 and applied locally; no production database changed.
- **Verified checks:** 22 PostgreSQL integration scenario groups with race detector; 13 frontend tests; checked production build and Go vet pass. Parser bundle size warning remains. See coverage document for exact verification limits.
- **Local QA records:** the clearly named `API Browser Verification` customer and its accepted `QT-2026-0001` quote (SAR 230) remain available to inspect the tested flow; these are verification data, not real business records.
- **Local access:** development database, API and frontend remain running. URL `http://localhost:5174`; administrator `admin@crm.local`. Password supplied directly to the owner and retained only in ignored `.env.local-login` (mode 0600), not in tracked documentation. This account exists in local development only.

### Next active work

Complete **T08/T23** customer sites/contacts and user/profile management UI, then finish **T06/T14/T16** separate-role sales acceptance, approval policy and exports. Follow the ordered remaining work in [API_COVERAGE.md](docs/API_COVERAGE.md). Server deployment (T24/T25) is still pending; server provider/domain and business rules have not been supplied. No milestone is marked complete based solely on route availability.

## Earlier foundation batch — historical evidence, 2026-09-16

- **T01 DONE:** local Git baseline and ignores; `compose.dev.yml`, `scripts/dev-api.sh`, configurable loopback ports, isolated test PostgreSQL with automatic cleanup. Clean local migration/startup verified. See [local development guide](docs/LOCAL_DEVELOPMENT.md).
- **T02 DONE:** all 51 original diagnostics repaired without weakening strict checks. Full `npm run build` passes. Docker now runs the checked build. Procurement DTO adapters and item/history helpers replace untyped collection assumptions.
- **T03 PARTIAL:** `validate` tags are active, nested quote/receipt validation is enabled, all paginated sort input is allowlisted, strict quote/customer/contract/PO requests added. Other handlers still require typed mutation DTOs, nested validation and consistent DB error handling.
- **T04 PARTIAL:** quote generic PATCH cannot change status/approval/totals; quote create/edit/recalculate/transitions are transactional and serialize on the quote row. PO status cannot approve/receive directly and ordering requires approval. Contract PATCH cannot activate or rewrite protected fields. Full lifecycle/audit/relationship review of other modules remains.
- **T05 PARTIAL:** embedded transactional/checksummed migration runner, first-admin bootstrap, database readiness and matching Go Docker toolchain implemented. Fresh migrations run twice successfully; bootstrap refuses a second active administrator. Existing-schema adoption, reference catalog/FX data, full server configuration and restore/rollback remain.
- **T06 PARTIAL:** `/auth/me` returns the complete current user; middleware checks current role and active status; refresh rotation is one-use under concurrency; password changes revoke refresh sessions. Frontend refresh is bounded/coordinated and logout reloads all stores. Customer empty collections serialize as arrays. Role-aware navigation, consistent DTOs across other modules and final server session policy remain. Existing access tokens survive logout/password change until expiry; disabled users are rejected immediately.
- **T07 PARTIAL:** PostgreSQL integration suite with race detection, 9 frontend session/import tests and GitHub Actions checks added. Compatible npm dependency updates plus official SheetJS 0.20.3 distribution reduce the audit from 13 affected packages to **zero reported vulnerabilities** ([audit result](docs/foundation-npm-audit.json)). XLS/XLSX imports with spacer columns now retain correct values; PDF parsing loads on demand. Hosted CI and a separate Go vulnerability scan have not run.
- **T08 PARTIAL:** multiple customers without CR now work; header edits map camelCase correctly, validate fields and preserve related arrays; read-only fields are rejected. Site/contact CRUD, relationship/primary-contact rules, visible errors and server-driven pagination remain.
- **T15 PARTIAL:** sequence rollover updates the existing name-keyed row atomically; 24 concurrent allocations and prior-year rollover verified. Missing/future-year sequences fail explicitly. Frontend mock quote/PO number generators still exist and must be removed when those screens are integrated.
- **T17 PARTIAL:** contract creation now persists customer/start/end dates and checks date order; generic updates are limited to editable draft fields. Accepted-quote conversion, transactional renewal, lifecycle rules and history remain.
- **T20 PARTIAL:** public login no longer renders the authenticated application shell. Mobile shell, form/dialog accessibility and table UX remain.

### Verification

- Frontend session/import tests: **9 passing**; checked production build: **PASS**, zero TypeScript errors. Parser chunk reduced from 845.47 kB to 507.96 kB; the size warning remains.
- PostgreSQL integration tests: **12 scenarios PASS with Go race detector**, including validation, empty collections/optional CR, customer editing, atomic quote creation, all nine roles attempting approval bypass, PO ordering prerequisites, contract create mapping, immediate role/deactivation checks, concurrent refresh, sequence rollover/concurrency and migration readiness/checksum checks.
- Both Docker images: **build PASS**, including checked frontend compilation. No production deployment was performed.
- `go vet ./...`: **PASS**. Ordinary `go test ./...` skips database integration unless `CRM_TEST_DATABASE_URL` is supplied; `make test-integration` always supplies a disposable database.
- Browser: login → create customer with blank CR → reload → retained profile and customer row with zero contacts/sites → edit customer → logout verified against local PostgreSQL. Public login checked at 390 px with no horizontal overflow; the authenticated mobile shell remains open work.
- This does **not** establish end-to-end sales completion. Opportunity creation, mock quote editor/list, contract conversion and remaining modules are still release blockers.

### Earlier implementation order — superseded by current API coverage

1. Finish typed mutation/relationship/error handling (T03/T04) and role/DTO and remaining verification gaps (T06/T07).
2. Finish real customers/sites/contacts and opportunity lookups/persistence/pagination (T08–T10).
3. Design the quote row/revision/decimal contract (T11) before wiring the full builder (T12–T16).
4. Complete accepted quote → contract and reporting; then remaining modules and server release gates.

Business rules and existing-data status were requested from the owner during implementation; no answer is assumed. Existing SAR/15%/approval-threshold values remain provisional.

## Definition of done for any module

- Visible data comes from the API; no unlabelled production demo records remain.
- Create/edit/archive survive reload and appear in a separate session.
- Valid/invalid/empty/not-found/forbidden/server-error cases produce accurate feedback.
- Relationships, validation, transactions, permissions and lifecycle rules are enforced on the server.
- List search/filter/sort/page work across the entire dataset; metrics reconcile independently of page size.
- Supported monetary/date/quantity rules are consistent in DB, API, UI and exports.
- Keyboard/responsive behavior works on the agreed devices; no new console or type errors.
- Tests target real business failures and relevant concurrency; checked builds pass.
- Documentation and this status file record evidence and remaining limits.

## First sales-release acceptance scenario

Using separate sales and manager accounts: create a customer with site/contact → create an assigned opportunity → create a quote using real catalog items → save/reload → change quantities/discounts → validate totals → submit → approve as manager → mark sent with a frozen revision → accept with recorded evidence → convert once to a contract → activate with valid dates → verify dashboard/audit/document links.

Repeat with invalid relationships, a low-margin quote, unauthorized approval, expired session, stale simultaneous edit, failed line write, and duplicate conversion request. Validate the policy for rejection/revision and contract renewal. No production claim until this scenario is repeatable on a fresh database and a migrated dataset.

## Session log

### 2026-09-16 — baseline review complete

- Reviewed 17 Vue views, service/store boundaries, Go routes/handlers/models, SQL schema and deployment files.
- Confirmed sales-first priority and server deployment intent with the owner.
- Ran frontend checks, Go compilation/vet, clean migration, isolated router/database probes and desktop/mobile browser inspection.
- Produced findings, evidence and this ordered work plan. Existing application code and schema were left unchanged.
- Temporary review infrastructure and private keys removed; existing projects' services/data were not changed.
- **Next:** T01 baseline and reproducible environment, then T02 checked build repair. No implementation task is marked started or complete.

### 2026-09-16 — foundation implementation

- Saved the reviewed tree in Git before application edits; added recoverable, testable local setup.
- Implemented the foundation fixes and partial task progress listed above. Original review evidence remains unchanged as historical evidence. Baseline commit: `e9672e3`.
- Added [LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md) and repeatable regression checks.
- Removed the temporary QA administrator, customer and refresh token after browser checks; stopped the local API, frontend and development PostgreSQL. The isolated development volume retains migrations but no QA records. Ignored development keys are preserved for subsequent sessions.
- **Next:** complete typed mutation/DTO boundaries, then customers and opportunities. M1 is not marked complete and production deployment remains blocked by functional and operational work.

### 2026-09-16 — frontend/backend integration

- Connected main module views to persistent APIs and added missing workflow routes, DTO mapping and migrations; see current batch above.
- Added `docs/API_COVERAGE.md`, `docs/API_ENDPOINTS.md`, module integration tests and frontend API-contract tests.
- Created and verified the requested local administrator; kept development services running for the owner.
- **Next:** customer/site/contact and user-management UI, role acceptance and remaining business/release gates. Historical foundation statements above describe the earlier state and are superseded by this batch.

### Template for future entries

`Date | Task IDs | Changes/files | Checks and results | Remaining risks/decisions | Next task`
