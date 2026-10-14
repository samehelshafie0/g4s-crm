# Business acceptance — fresh-data sales-to-delivery cycle

Date: 2026-10-07. Branch: `feature/opportunity-flow`. **Functional acceptance: PASS for the cycle below.** Changes remain local and uncommitted. This is test execution evidence, not owner approval of business policy or a production release.

## What was tested

A fresh PostgreSQL 16 database (`business_uat`) was migrated from empty. It contains only fictitious UAT records. Customer/site/contact, manufacturer/category, one manual product, one vendor-imported product, a draft quote and the initial stock PO were prepared through authenticated APIs. No business state was seeded by SQL.

The browser then performed the workflow with separate sessions for sales executive, sales manager, procurement officer, procurement manager, warehouse manager and project manager. The continuous integration scenario independently creates its own records and runs business operations with separate role tokens; admin only provisions its accounts. Its supplier-quotation-to-PO branch complements the direct-PO browser branch.

| Acceptance step | Evidence and result |
|---|---|
| Customer and catalog | Fresh company, site and primary contact; manual recorder and imported camera with vendor/manufacturer/category. API PASS. Earlier browser creation/import evidence remains in [the workbook report](FULL_CYCLE_ACCEPTANCE.md). |
| Quote | Sales submitted QT-2026-0001; sales could not approve or edit the submitted quote. Manager approved, recorded sent/accepted status. Browser PASS. Sending is a status change, not email. |
| Money | 5 cameras × SAR 150 + 2 recorders × SAR 100 = SAR 950; VAT 142.50; total **SAR 1,092.50**. Same amount persisted in the activated contract. |
| Existing inventory | Manager approved/ordered PO-2026-0001. Warehouse received 3 cameras and held exactly those 3 against the accepted quote line. Manufacturer/category visible. Browser PASS. |
| Buy shortfall | Purchasing officer selected the accepted job without sales-quote permissions. Wizard proposed only **2 cameras + 2 recorders**. Created PO-2026-0002, submitted for manager approval; officer had no approval button. Manager approved/ordered. Browser PASS. |
| Partial/damaged receipt | Warehouse received 2 cameras (1 damaged) and 1 recorder. PO became partial-received. Good units became job holds; damage created a pending return and replacement draft PO-2026-0003 for 1 camera. Browser PASS. |
| Return | Warehouse recorded **UAT-RETURN-001**. Receipt changed to Returned; damaged unit never entered usable inventory. Browser PASS. |
| Remaining receipt/replacement | Warehouse received the final recorder. Procurement manager submitted/approved/ordered the replacement; warehouse received it. All 3 POs received, 4 GRs total. Browser PASS. |
| Dispatch | Warehouse fulfilled all 5 holds under **UAT-DISPATCH-001**: 5 cameras and 2 recorders. Both SKUs ended with **0 on-hand, 0 reserved, 0 available**, and demand lookup reports zero remaining. Browser and independent API reconciliation PASS. |
| Contract | Sales manager converted the accepted quote to CNT-2026-0001 and activated it (7 Oct 2026–7 Oct 2027). Browser PASS. |
| Delivery | Project manager assigned the job, set In progress, then Completed with actual end date. Completed row verified with the completed-jobs filter enabled. Browser PASS. |

Machine-readable final reconciliation: [fresh UAT result](business-acceptance-20261007.json). Supplier credits/refunds are outside this cycle: PO totals include the original damaged item and replacement. Delivery completion is still operator-set, not an enforced dispatch gate.

## Issues found and fixed in this run

- **Operational job access:** purchasing roles could not load sales quotes/opportunities; warehouse had no accepted quotes for holds. Added paginated accepted-job lookup with product identities, quantities and remaining demand. It excludes sales prices, margins, internal notes and customer objects. Full quote APIs remain forbidden to these roles.
- **Warehouse receiving/returns:** receipt/return endpoints and controls now accept warehouse inventory-update authority without granting PO creation or approval. Added one receiving form for date, warehouse, per-line received/damaged quantities, skipping lines and validation.
- **Permission-aware loading/actions:** procurement skips forbidden reference calls. Create/edit/delete/approve/order controls follow their actual permissions. Warehouse sees purchasing-needed text instead of unusable order buttons.
- **Hold defaults:** unavailable rows stay unselected, so the warehouse can reserve available units immediately. Server remaining demand supplements the local planner and server quantity caps remain authoritative.
- **Sales dashboard access:** warehouse and procurement accounts previously received sales totals/margins through dashboard APIs despite having no quote-read permission. Sales dashboard endpoints now require quote-read authority; these users land on their operational workspace.
- **Stock detail freshness:** the open product panel now follows current stock after fulfillment, instead of retaining the earlier on-hand total.
- **History lookups:** receipt/supplier-item history accepts product IDs as well as SKUs, matching callers from Procurement and Inventory.
- **Sales display:** missing quote expiry renders a dash; hiding purchasing controls no longer causes a false “all items in stock” message.

The earlier fixes remain covered: exact quote-line holds; combined hold/order caps and concurrency; over-receipt rejection; damage/replacement follow-up; manufacturer/category display; canonical alfanar vendor aliases and rejection of incomplete new vendor names.

## Verification

- `./scripts/test-integration.sh`: **PASS**, disposable PostgreSQL 16, Go race detector; final integration package **132.381s**. Includes separate-role cycle and dashboard/quote access denials, returns, replacement, import rollback/FX/history, demand/stock reconciliation, concurrency and existing module regressions.
- `npm test`: **24 PASS**, including the populated shipped workbook parser and hold-planning regression.
- `npm run build`: **PASS**, TypeScript/Vue checks and production build. Existing file-parser chunk-size warning remains.
- `go vet ./...` and `git diff --check`: **PASS**.
- Desktop Chrome: workflow in the table above; receiving dialog visually checked; reload and separate sessions used. This is not exhaustive mobile/Safari UAT.

## Review environment and remaining launch work

The isolated review environment is retained at **http://127.0.0.1:5179** (API loopback 18109). Local test accounts are listed in ignored `crm-api/tmp/business-uat/LOGIN.txt` (mode 0600). The separate Docker container is `g4s-business-uat-20261007`; no live records were reset or migrated. Temporary regression containers were removed by the test script.

Fresh data avoids the legacy migration blocker; it does not repair it. If retaining the existing live dataset, resolve the two empty draft quotes on op2, identify vendor `in` for SKU `NSI- Prod`, and classify the 29 products with missing manufacturer/category. These were left intact rather than guessed or deleted. Verified alfanar aliases are handled by the pending code/migration.

The fillable Excel template and mapping/import path are implemented and parser-tested. The earlier automated native-file-picker attempt was blocked by Chrome extension file access; native upload remains a manual check. Paste import is a tested alternative. See [workbook instructions](FULL_CYCLE_ACCEPTANCE.md#how-to-use-the-workbook).

Before production: owner confirmation of permissions/approval and SAR/VAT policy, review/commit of the local batch, remaining T16 export/unsaved-edit work and the broader backlog in [STATUS.md](../STATUS.md), then deployment details and backup/restore/migration/load checks. No deployment, commit, push, external supplier order or real contract action occurred in this run.
