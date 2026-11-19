# Quotation price history and lifecycle review — 13 November 2026

## Confirmed behavior

The owner confirmed that a customer **accepting the quotation** is the event that counts as a sale. Delivery completion is not required for sale-price history.

The quote builder's item-history action now opens three views:

- **This customer:** accepted sales of that product to the quotation's selected customer.
- **Other customers:** accepted sales to other customers whose quotations the user may access.
- **Supplier offers:** retained manufacturer/supplier price offers, with supplier quotation, offer revision, original currency, quantity, validity and customer/project context where linked.

Customer detail → Price history defaults to accepted sales; an explicit selector also shows all quoted prices and earlier revisions. Products → Supplier offer history shows the retained supplier offers.

## Price records and calculations

Migration **000020_quotation_price_history.up.sql** adds `supplier_price_records`, quotation acceptance timestamps, and lookup indexes.

Each supplier quotation save records its linked product lines within the same transaction. Editing the supplier quotation adds another record instead of erasing old evidence. Soft deletion of the supplier quotation also preserves its records. Linking an existing supplier quotation associates previously unassigned snapshots with that customer quotation. A linked supplier quotation cannot be moved to a different customer quotation by editing it.

The migration preserves the supplier item rows that still exist when it runs. It cannot recover supplier prices overwritten before history was introduced. Supplier lines must identify a catalog product to appear in that product's history. Attaching an unstructured PDF alone does not extract item prices.

Accepted customer quotations and their lines already become immutable through the application workflow, so they remain the source of sale-price records. Draft, approved, sent, declined and expired quotations are not sales. Unselected optional items are excluded. The acceptance timestamp is recorded on transition to Accepted. Legacy accepted quotations use an explicitly labeled estimated date from their last stored update.

The comparison shows original unit price, line discount, quotation discount, net unit price before VAT, currency, quantity/multiplier, customer and quotation version. Net unit price is rounded to currency cents; aggregate invoice-style rounding may differ by a cent.

**Use price** converts the historical net price through the historical and current quotations' pinned exchange rates and adjusts the editable unit price for the current line and quotation discounts. It preserves those discounts and the existing unit cost. A 100% discount blocks application with an explanation. New items must still satisfy the selected price book.

Example: a USD 100 line with 10% line discount and 5% quotation discount records **USD 85.50 net**. At a pinned test rate of 3.75, that is SAR 320.625 before rounding, not SAR 375. A quotation without discounts can use SAR 320.63.

## Permissions

- All customer sale comparisons require quotation, product and customer reading rights.
- Both views retain the quotation's team visibility rules. A user cannot obtain another team's sale prices through the comparison endpoint.
- `quotes:compare-prices` controls the **Other customers** comparison. Its role default inherits quotation reading and can be overridden per user in User Management.
- Supplier offers require cost/margin visibility as well as quotation/product access.
- Applying a sale price requires editable draft and quotation pricing permissions.
- The sale-history response contains selling prices and discounts, without internal costs or margins.
- These controls supplement existing module permissions; they do not replace separate access rights for the customer directory, product catalog or document library.

## Visual changes

The previous popup mixed customer prices with supplier/batch labels and showed cost summaries as zero when cost fields were hidden. It has been replaced with a native modal dialog, clear source buttons, accepted-sale labels, original currencies, readable loading/error/empty states, pagination and explicit Use price actions.

The quote builder now lets the catalog search area shrink without pushing Quick add into the summary, keeps long descriptions usable, exposes horizontal item-table scrolling to the keyboard, wraps narrow layouts and adjusts running-subtotal alignment when optional columns are hidden. The new history selection/action buttons use theme colors with readable foregrounds.

## Verification results

- **PASS:** full disposable PostgreSQL integration suite with race detector, **160.999 seconds**. Existing sales-to-delivery tests ran alongside the new history checks.
- **PASS:** new coverage for accepted-only sales, discount calculations, excluding unselected optional items, acceptance timestamps, cross-customer comparison, permission denial, team scoping, supplier edits/deletion retaining both offers, and accepted cost snapshots staying fixed.
- **PASS:** 31 frontend tests, including pinned currency conversion and both discount levels. Production build/type checking and Go vet passed. The pre-existing roughly 509 kB file-parser bundle warning remains.
- **PASS, native Chrome:** pending approval → return to draft → resubmit → approve → sent → declined → revise → submit → approve → sent → accepted. The local quotation reached **QT-2026-0001 v3**, preserving cost SAR 225 and unit price SAR 375. Earlier versions remained in history.
- **PASS, native Chrome:** dark price-history dialog structure and empty state; Escape closes the native modal.
- **PASS, local API fixtures:** a future quote retrieves the selected customer's SAR 375 accepted sale; another customer's discounted USD 85.50 sale is returned separately; supplier offer revisions retain SAR 225 and SAR 230.
- **PASS, export review:** downloaded the accepted quote PDF through the real API, rendered its page and inspected the line, totals and footer. SAR 3,750 subtotal + SAR 562.50 VAT = SAR 4,312.50. No internal cost/margin columns appear.
- **PASS, native Chrome follow-up:** populated customer and other-customer history. Applied SAR 375, then the discounted USD 85.50 comparison converted to SAR 320.63. Unit cost remained SAR 240 in both cases. Saved the draft and confirmed SAR 320.63 / SAR 240 persisted when opened in Safari.
- **PASS, native Safari follow-up:** light desktop history, both supplier revisions (SAR 230 latest / SAR 225 earlier), dark payment-term dropdown, resized browser and enlarged-page layouts, modal body scrolling with a reachable fixed close action, and keyboard focus scrolling the history table to its Use price action (Safari Option-Tab). These were native resized/zoomed browser checks, not mobile device emulation or touch-device testing. The automation's horizontal wheel gestures did not move the table; keyboard access to the offscreen action was verified instead.
- **FIXED during visual review:** narrow catalog results overlapped manufacturer, price and stock labels and pushed the history button out of view. Result details now wrap, SKU/name support long text, and the history button retains its space. Rechecked the corrected row in narrow light and desktop dark layouts. Production build/type checking passed again after this CSS-only correction (10.00 seconds); `git diff --check` passed.

## Local review data

The 13 November review used local, uncommitted changes. Migration 000020 was applied only to the isolated review database and disposable integration databases. On 19 November the owner requested droplet commands and will execute the deployment; server migrations and health checks remain pending.

- Frontend: `http://127.0.0.1:5184`
- Future quote with comparison history: `/quotes/2a1ec6cf-0660-4683-b115-85a91f8d51ba/builder` (**QT-2026-0003**)
- Accepted quote: `/quotes/3ea51f41-fe0a-48a6-a661-3068b59310ff/builder` (**QT-2026-0001 v3**)
- Local fixture IDs: ignored `crm-api/tmp/meeting-review/price-history-fixtures.json`.
- PDF and rendered review image: ignored `crm-api/tmp/meeting-review/accepted-quotation.pdf` and `accepted-quotation-1.png`.

The requested browser follow-up is complete locally. The next step is to review the batch before deployment. Earlier meeting items about mandatory team assignment, scheduling, general form customization and the original browser crash remain as documented in `MEETING_FIXES.md`.
