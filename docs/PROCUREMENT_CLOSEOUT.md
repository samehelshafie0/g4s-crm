# Procurement and Inventory closeout review — 7 October 2026

The requested workflow fixes are implemented on `feature/opportunity-flow` and verified locally. They are **uncommitted and not deployed**. The live database was inspected read-only; all writes and browser exercises used disposable test databases. This closes the listed code defects, with catalog corrections still requiring input. It does not establish whole-project production readiness.

## Results against the requested fixes

| Issue | Result | Evidence / remaining limit |
|---|---|---|
| Inventory holds have no quote-line link | Fixed for new job holds | The form loads the accepted version and sends its exact line ID; the API derives the quote version/customer reference and rejects unlinked job holds. Browser-created hold for `HIK-DS-K1T343MFWX` persisted against the correct line, quantity 4, reference `QT-2026-0002 v1`, on the isolated copy. Manual stock holds remain available. |
| Ordering/holding exceeds a quote line | Fixed | Combined active/fulfilled holds plus unreceived quantities on non-cancelled POs cannot exceed `ceil(quantity × multiplier)`. Draft POs consume demand. Direct PO create/edit, supplier-quote conversion, hold creation and submission/ordering enforce the rule. Duplicate lines within a request count together. Product locks serialize competing stock/order writes. A concurrent hold-versus-PO regression allows exactly one request. |
| Damaged receipts have no follow-up | Fixed for new receipts | Damaged quantities create pending return records and a linked replacement **draft** PO within the same transaction. Only good units enter stock and job holds. Partial damage requires the exact damaged count. A replacement follows normal approval and ordering; it does not automatically purchase anything. Returning damaged goods records a dispatch reference without reducing usable stock again. |
| Blank Inventory manufacturer/category | Display defect fixed; catalog incomplete | Removed sample-SKU mappings. API loads manufacturer/category relations and Inventory uses those values. Browser confirmed Hikvision / IP Cameras, Dahua Technology / Recorders & Storage, and Riyadh Cables Group / Structured Cabling. Missing catalog values show “Not assigned.” **29 of 98 live products lack both classifications.** |
| Vendor `in` and duplicate alfanar names | alfanar fixed; `in` identity unresolved | There is one alfanar master, code `ALFANAR`; its two catalog vendor entries used `alfanar Electric (Riyadh)`. Migration 17, seed data and name matching use `alfanar Electric`; browser shows one supplier choice. `in` belongs to SKU `NSI- Prod`, “HID Proximity cards – Staff,” imported from `ZGT- Card printing.xlsx`. Its actual vendor is not established. It is excluded from procurement suggestions, new product/import/procurement writes reject it, and the original record/history are preserved pending correction. |

Additional correction found during review: goods-receipt unit prices now display the PO currency (e.g. USD 138.00); landed values remain SAR.

## Reviewed flow and operating rules

1. Customer → opportunity → quote versions → accepted version. An accepted quote is the demand basis for purchasing and holds. Job POs cannot be raised against an unaccepted quote; stock purchasing remains available without a customer source.
2. Fulfil from available stock or order the remaining demand. Holds and draft POs share the same quantity budget. Released holds, cancelled POs and deleted drafts free demand; fulfilled holds continue counting because those units have already been issued.
3. Supplier-quote conversion automatically links a product only when it maps to one accepted customer line. If the same product occurs on several lines/sites, use a direct PO with explicit line IDs. The system refuses to guess.
4. Submit → approve → order → receive. Good receipt quantities become usable stock and are automatically held for the linked job. Receiving replaces outstanding quantities with holds; these are not counted twice.
5. Damaged quantities remain outside usable inventory. Record their return from Goods Receipts; inspect the generated replacement draft in Purchase Orders, confirm supplier credit/pricing, then follow normal approval. Partial damage can contain both good and damaged quantities on the same receipt line.
6. Fulfil the hold with a dispatch reference. The integration flow verifies stock reduction and preserved demand accounting; the existing suite also exercises quote acceptance, contract conversion and delivery updates.

The return record tracks dispatch of **all damaged units on that receipt line**, with a timestamp and audit entry. Credit notes, refunds, supplier settlement and partial return dispatches are not implemented. Replacement drafts copy the original unit price, use zero new freight/customs, and require buyer review. These are reviewable procurement drafts, not a claim that replacements are chargeable or free.

Historical damaged receipts are marked pending return by migration 16. The migration does **not** create retrospective replacement orders, because another order may already cover the shortage. Review their remaining demand before reordering. Older `partial-damage` records did not store a split; migration preserves the prior treatment of the whole row as damaged rather than inventing usable units.

Legacy job POs lacking line links are blocked at submission/ordering. Correct a draft against an accepted quote, or cancel and recreate a pending/approved legacy PO. Existing stock POs are unaffected. Live inspection found no existing stock-reservation rows to backfill.

## Verification completed

- Full `./scripts/test-integration.sh`: **PASS**, real disposable PostgreSQL 16, Go race detector, final run 73.452 seconds. Existing sales/module checks plus six new closeout groups cover classifications, linked holds, combined limits, repeated lines, draft edits, supplier conversion, partial damage, replacement receipt, duplicate receipt rejection, returns, concurrency, legacy bypass and incomplete vendors.
- Frontend: **22 tests PASS**; checked production build **PASS**. The existing ~508 kB parser-chunk warning remains.
- `go vet ./...`: **PASS**. `git diff --check`: **PASS**.
- Upgrade: cloned the previously migrated live-data test database into `g4s_closeout_20261007`; migrations 16–17 applied successfully and the migration command ran again without changes. Alias rows were normalized and the historical damaged receipt was marked pending.
- Browser on the isolated copy: real Inventory classifications; fully committed quote produces no additional hold; release/recreate four-unit hold and verify its persisted quote-line ID; damaged receipt pending-return indicator; return form disabled until a reference is supplied; confirmed return persists and displays `Returned`; USD unit-price label; one alfanar supplier option; no warning/error console entries at the final check.
- Cleanup: temporary preview/API stopped and `g4s_closeout_20261007` removed after verification; the pre-existing live-copy database and live services remain untouched.
- Not repeated here: full mobile/multi-browser acceptance, load testing, full operational restore, or every role/module mutation. API return access uses `procurement:update`; approval permissions are unchanged.

## What still prevents project sign-off

1. **Live migration blocker remains:** `op2` still has empty drafts `QT-2026-0002` and `QT-2026-0004` (read-only recheck today). Choose the quote number to retain, preserve a backup, and resolve the other before migration 11. Neither live quote was removed.
2. **Catalog corrections:** identify `in`; classify the 29 imported products. The subsequent [full-cycle batch](FULL_CYCLE_ACCEPTANCE.md) adds per-row import/classification and a fillable workbook, but does not infer historical manufacturer/category from vendor names or product descriptions.
3. **Business UAT and release gates:** the broader PARTIAL/TODO work in `STATUS.md` remains, including remaining module acceptance, role/policy confirmation, money/FX rules, server configuration, TLS, backup/upload restore, monitoring and rollback. Delivery completion is still an operator-set status; it is not proof that every material has been dispatched.
4. **Deploy together:** review/commit this batch, resolve migration prerequisites, back up, run migrations 8–17 on the live database, and release API/frontend together. No merge, push, production migration or deployment was performed.

## Main implementation files

- API: `handlers/fulfilment_links.go`, `inventory.go`, `procurement.go`, `receipt_returns.go`, `vendor_names.go`, product/import handlers, receipt/stock DTOs, and the authenticated return route.
- Database: `000016_receipt_damage_followup.up.sql`, `000017_catalog_vendor_alias.up.sql`.
- UI: `InventoryView.vue`, `ProcurementView.vue`, `fulfilmentPlanning.ts`, `vendorNames.ts`, shared types and procurement DTOs.
- Regression coverage: `fulfilment_closeout_test.go`, `fulfilmentPlanning.test.mjs`, and the foreign-currency receipt DTO assertion.
