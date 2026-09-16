# G4S CRM — current status and work plan

**Last updated:** 2026-09-16  
**Current phase:** baseline review complete; implementation fixes not started.  
**Release state:** not ready for production.  
**Primary objective:** make customers → opportunities → quotes → contracts work end to end, then finish remaining modules and deploy to the owner's server.  
**Detailed findings:** [docs/CRM_REVIEW.md](docs/CRM_REVIEW.md).

## Start here each work session

1. Read this file and the findings linked to the next task.
2. Inspect current changes and preserve existing work. No Git repository was detected during the baseline review; establish version control before substantive changes.
3. Mark one bounded task `IN_PROGRESS`; record assumptions and dependencies.
4. Implement and run the task's acceptance checks. An existing screen or endpoint is not completion.
5. Update this file with actual results, files changed, remaining issues and the next action. Mark `DONE` only with evidence. Keep review findings as historical context rather than silently rewriting them as fixed.

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED` (name the dependency), `DONE`, `DEFERRED` (record the decision). Review recommendations are not yet completed work.

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

## Verified baseline

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
| M1 — Reliable foundation | Checked builds, isolated setup, validation, safe updates, stable session/DTO contracts | Clean install; no approval bypass; refresh/customer smoke tests pass | TODO |
| M2 — Customers and opportunities | Persist customers/sites/contacts and real opportunities/assignees | Cross-session CRUD, pagination/search and stage tests pass | TODO |
| M3 — Quotes | API-backed catalog/list/builder with authoritative calculations, approvals and revisions | Save/reload/send/accept flow and role/concurrency tests pass | TODO |
| M4 — Contracts and sales reporting | Convert accepted revision to contract; lifecycle, files and truthful dashboard | Full sales journey passes with database persistence and audit trail | TODO |
| M5 — Remaining modules | Procurement, inventory, projects, teams, recurring/price-book depth | Each workflow passes its own functional and integrity gates | TODO |
| M6 — Server release | Harden, deploy and prove operations | TLS, migrations, backup+restore, monitoring and rollback verified | TODO |

M6 can follow M4 only if the owner chooses a sales-only first release and incomplete modules are explicitly disabled. Otherwise M5 must complete first. Sales-only release scope has not yet been confirmed; priority alone does not remove modules from scope.

## Ordered work backlog

| Task | Milestone | Concrete deliverable / acceptance | Findings | State |
|---|---|---|---|---|
| T01 | M1 | Establish Git baseline/root ignore rules and reproducible local environment; configurable ports that do not collide with other projects | R18, R24 | TODO |
| T02 | M1 | Fix 51 type diagnostics without disabling strict checks; use checked build in Docker/CI | R01 | TODO |
| T03 | M1 | Correct validation helper and nested field validation; consistent error/rows-affected handling; whitelist sort keys | R07–R09 | TODO |
| T04 | M1 | Remove generic status/approval/total updates; test each role against bypasses | R06, R11 | TODO |
| T05 | M1 | Real migration runner + first-admin bootstrap + reference data; match Go toolchain; verify clean startup/readiness | R18 | TODO |
| T06 | M1 | Shared user/collection DTOs; safe `/auth/me`; bounded refresh, session policy, logout reset and role-aware UI | R04, R05, R16 | TODO |
| T07 | M1 | Add test/CI structure; preserve reproduced failures as regression cases; triage and update dependencies | R19, R24 | TODO |
| T08 | M2 | Customer CRUD plus site/contact UI/API; optional CR handling; relationship/primary-contact constraints; visible errors | R03, R05, R14, R20 | TODO |
| T09 | M2 | Opportunity request mapping, live customer/user lookups, owners/services/dates/costs, stage events and validation | R02, R03, R11 | TODO |
| T10 | M2 | Shared server-driven list query pattern with total/page/filter/sort; lookup endpoints and 60-record tests | R08, R20 | TODO |
| T11 | M3 | Agree quote row/revision/decimal/FX DTO and add migration; resolve price-book/service representations | R10, R14 | TODO |
| T12 | M3 | Persist the catalog subset required for quotes: manufacturer/category/product/service costing and eligible pricing | R02, R03, R10 | TODO |
| T13 | M3 | Replace quote list/builder mocks; load by ID; atomic header+line save; reload and second-user persistence | R02, R03, R05, R09 | TODO |
| T14 | M3 | Server calculations and threshold policy; approval/rejection/send/accept/decline; immutable revision and stale-edit checks | R06, R10, R11 | TODO |
| T15 | M3 | Fix yearly number generation; test rollover and concurrency; remove frontend number assignment | R15 | TODO |
| T16 | M3 | Safe, accurate customer print/export; dirty-state/pending-save feedback; permission-controlled costs | R10, R17, R23 | TODO |
| T17 | M4 | Accepted quote → contract conversion once; dates/value snapshot; activate/terminate/renew with history and transactions | R03, R09–R11 | TODO |
| T18 | M4 | Real document upload/download/version/link behavior with limits and contained paths | R17 | TODO |
| T19 | M4 | Real dashboard activity/quotes/expiry and defined currency/date/metric semantics | R02, R22 | TODO |
| T20 | M4 | Fix mobile shell, common dialogs/form labels/keyboard access, table columns/actions and public login layout | R20, R21, R23 | TODO |
| T21 | M5 | PO/SQ item editing, acceptance, approvals/conversion and partial goods receipts; correct update/delete routes | R09, R13 | TODO |
| T22 | M5 | Stock receipt/reserve/release/fulfill/transfer/adjustment with concurrency, availability and movement reconciliation | R12, R14 | TODO |
| T23 | M5 | Complete projects, teams/user management, price books, FX history and recurring service behavior | R02, R10, R11, R14 | TODO |
| T24 | M6 | Server runbook/configuration/TLS/secrets/networking/readiness/logging; repeatable deploy and rollback | R16, R18, R24 | TODO |
| T25 | M6 | Full regression/UAT, role tests, migration upgrade/rollback, backup+upload restore and realistic load checks | All relevant findings | TODO |

Dependencies: T03/T04 protect all mutations. T05 enables repeatable environments. T06 precedes live UI integration. T11 precedes T13/T14. T17 requires an accepted persisted quote revision. T21/T22 depend on agreed catalog/stock contracts. T24/T25 require actual server details. Work on those independent foundations need not wait for optional design decisions.

## First implementation batch

**Recommended next work: T01–T07 (M1), starting with a baseline and build repair.**

1. Preserve the current tree in version control and establish isolated ports/configuration.
2. Restore a passing checked frontend build.
3. Fix validation, sort allowlists and approval bypasses with regression tests.
4. Repair first-admin/migration startup and session/response shapes.
5. Re-run login → reload → list/create customer with no contacts/sites, then proceed to T08/T09.

Do not wire the complex quote editor directly to the current underspecified quote model. T11 is a deliberate design task so existing builder features are not silently lost.

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

### Template for future entries

`Date | Task IDs | Changes/files | Checks and results | Remaining risks/decisions | Next task`
