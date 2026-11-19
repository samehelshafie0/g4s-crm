# Meeting fixes — 13 November 2026

**Follow-up:** [Durable supplier offers, accepted-sale comparisons and lifecycle review](QUOTATION_PRICE_HISTORY.md) adds migration 000020 and supersedes the earlier price-history behavior described below.

The 13 November review used local, uncommitted implementation and an isolated database. On 19 November the owner requested commands to deploy this batch from `feature/opportunity-flow`. Server execution and verification remain pending. Migrations `000019_meeting_workflows.up.sql` and `000020_quotation_price_history.up.sql` must run before the new API starts.

## Confirmed decisions

- Product hierarchy is **Manufacturer → Category**, with no additional subcategory.
- Team-assigned quotations are approved by their team leader. Administrators can also approve; there is no mandatory second admin stage or threshold stage.
- Quotation actions have per-user Allow / Deny / Role default settings in User Management. The API enforces them using current database permissions on every request.

## Implemented and checked

| Meeting item | Result |
| --- | --- |
| Customer contact count and relationship | Add/edit/delete real customer contacts with name, position, email, phone, primary flag and optional site. Counts update after save. |
| Add sites and contacts in their columns | Separate count/open and Add controls in the Sites and Contacts columns. Add opens the relevant editor immediately. |
| Supplier quotation linked to the project/customer quote | Builder → Documents can link an existing supplier quotation and show its original attachments. Supplier quotations retain opportunity and source quote references. |
| Product pricing provenance | Products → Project prices lists supplier offers with customer, project, customer quote number/version, supplier and currency. This does not overwrite the product's standard catalog price. |
| Fixed unit cost, editable price or margin | Saved line cost stays fixed on the server. Editing margin recalculates selling price, including line discount. Editing price recalculates margin. New supplier lines resolve their cost from the linked supplier offer on the server. |
| Different price for the same product by project | Independent quote lines and their snapshots retain their own selling price and supplier cost. Duplicate lines can have different selling prices. |
| Customer-specific books | Catalog/search/quick-add show only members of the chosen book. The API rejects new rows outside that book, including write-in bypass attempts. Previously saved lines remain when switching books. |
| Customer-specific price history | Customer detail and builder history show that customer's actual quotation line prices, currency, status and version. Applying history changes selling price, preserving saved cost. Supplier purchasing history remains in the product module. |
| Revision/update controls | Drafts can be saved; pending quotes can be returned to draft by an authorized approver; approved/sent/declined/expired current versions can be revised. Old versions remain frozen. |
| Quotation permissions | View, create, edit, pricing, cost/margin visibility, submit, approve/return, mark sent, accept/decline, revise, export, document changes, and delete have per-user overrides. View permission and team membership apply before quote actions. |
| Team membership and approvals | A person can belong to several teams, and a leader can lead several teams. Removing one membership preserves the others. Assigned quotes are visible to members of their active team and administrators. Approve/return also requires the approve capability and team-leader/admin identity. |
| Quotation document grant | Authorized quote users can upload and link supporting files without needing general document-library edit rights. The grant does not allow changes to unrelated customer/product document links. |
| Vendor duplicates | Case/whitespace-normalized vendor/manufacturer creation and product-vendor duplicate entries are rejected. Vendor/customer form suggestions show existing names. Existing duplicates are not automatically merged. |
| Manufacturer/category and imports | Existing manufacturer/category masters, dependent import-template lists and import validation are retained. No subcategory layer was introduced. Real unclassified imported items still require the client's choices. |
| Dark dropdowns and scrolling | Select options use theme surface/text colors and color-scheme; scrollbar colors and stable gutters follow the theme. Dark price-book dropdown checked in Chrome. |
| Hide optional fields | Builder's Customize columns persists cost, margin, discount and multiplier visibility in this browser. Customization cannot reveal permission-denied quote costs/margins. |
| Product forms closing | Removed accidental backdrop dismissal for product entry dialogs and added warnings when leaving open product/import forms. No browser-process crash was reproduced. |
| Exchange-rate refresh | Exchange Rates refreshes on opening and hourly while visible for users with update rights. Other users reload stored rates. The existing provider publishes daily; this is not a streaming feed or a server background schedule. |

## Permission and migration details

1. Role defaults remain until an administrator changes an override. An explicit Deny wins over that user's role, including admin. Only authorized user administrators can set overrides; profile updates cannot self-escalate.
2. Cost/margin denial removes quotation financial fields from API responses, including nested quotation summaries, customer price history, dashboard quote rows and the builder catalog. It is not merely CSS hiding. Product, Procurement and general Documents remain separately authorized modules; their own catalog prices or original supplier files may disclose commercial information. Assign those module roles consistently with the user's responsibilities.
3. `quotes:documents` controls changes to quotation attachments/appendices. General document-library reading still uses `documents:read`. Supplier pricing also requires quotation cost/margin visibility.
4. A team leader still needs `quotes:approve`. Being designated leader does not silently grant an approval capability. The builder loads the user's eligible quotation teams independently of access to the broader Teams directory.
5. **Unassigned quotations retain the existing role-based approval path.** Administrators should assign the appropriate team to drafts to use team-lead approval. Nonadmins cannot clear an assigned team. Making team assignment mandatory for every legacy/new quotation is not enforced in this batch.
6. Migration 000019 adds the membership join table, backfills primary memberships and team leaders, and preserves compatibility with the primary-team field. Changing primary team adds membership; use Teams to remove an old membership explicitly.
7. Supplier cost is a quotation snapshot. Later supplier changes do not rewrite existing quote lines. The source item ID is retained without a foreign key because the supplier editor replaces its item rows; current product offers reflect the supplier's current items.
8. Existing unrelated roles, profile behavior, customer data, VAT policy, and deployment decisions were not reconfigured.

## Verification

- Disposable PostgreSQL integration suite with race detector, including existing sales-to-delivery, imports, documents and approval workflows.
- New integration scenario covers multiple memberships, leader without Teams-directory access, role overrides, no profile self-escalation, fixed costs, customer price history, book-only selection, supplier cost provenance, quote-only document upload/link, unrelated-link denial, team-lead approval and read/export/pricing denials.
- Response-policy tests cover quotation costs, nested summaries, team scoping, builder catalogs, separate product access, and operational quantity reports.
- Frontend: 30 tests, production build and Vue/TypeScript checking.
- Go vet and whitespace verification.
- Native Chrome on a fresh isolated development database: adding a contact changed the customer count from 1 to 2; cost SAR 225 remained fixed while margin 40% yielded price SAR 375; save succeeded; dark dropdown readable; submit → approve → revise produced **QT-2026-0001 v2**, preserving the line cost and price. Revision-dialog and builder screenshots were shown in the conversation.

Final command results are recorded in the current batch in `STATUS.md`.

## Remaining input or separate work

- Team scheduling/time-slot conflicts need a scheduling model and rules: which slots exist, timezone, overlap policy and whether conflicts warn or block. Multiple team membership is implemented; scheduling is not.
- Browser/tab crash diagnosis needs reproducible steps or a crash report. The accidental form-close fix is verified; it is not proof of a browser crash fix.
- Client must resolve the pre-existing manufacturer/category gaps in imported data. No classification guesses or duplicate merges were made.
- Form-wide field customization beyond the quotation columns is not implemented.
- Wael's interface comments, requirements and introduction to the client's team remain human follow-ups. No messages were sent.
- Deployment and the pre-existing test-server/production release gates remain separate.

## Local review

The fresh review frontend runs at `http://127.0.0.1:5184`, with the API at `http://127.0.0.1:18089`. Its separate Postgres container is `g4s-crm-meeting-review` on port 55439. Temporary test credentials and fixture IDs are in the ignored `crm-api/tmp/meeting-review/` directory. This is not the live database.
