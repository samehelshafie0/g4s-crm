# Sales acceptance and supporting workflows

Updated: 2026-09-17. Customer contacts/sites are deferred by owner. This document records verified behavior and the remaining manual acceptance steps; it is not production release approval.

## Automated acceptance — PASS

Run `make test-integration` from the repository root. It creates a disposable PostgreSQL database, applies all migrations and runs Go tests with the race detector. Test accounts exist only in that database and are removed with it; they are not server login credentials.

The suite now has 26 scenario groups. `crm-api/internal/integration/role_workflows_test.go` uses separate authenticated administrator, sales executive, sales manager and warehouse manager sessions.

| Actor | Verified behavior |
|---|---|
| Administrator | Create users, assign teams, reset passwords, deactivate/reactivate. Password resets invalidate refresh sessions and old-password login; disabled users immediately lose API access. |
| Sales executive | Edit own profile without role escalation; create customer/opportunity; save a two-unit service quote; reload persisted lines and totals. Stale edits return 409, unauthorized approval/conversion/activation return 403. |
| Sales manager | Read the sales user's quote, reject, approve resubmission, convert the accepted quote once and activate the contract. Repeat conversion returns the existing contract. |
| Sales executive | Send and accept the approved quote, then read the activated contract from a separate session. Submitted quote edits are rejected. |
| Warehouse manager | Adjust stock, reserve two units, change reorder level, fulfill with a reason, reject duplicate fulfillment and reconcile remaining on-hand/reserved/available quantities. Sales cannot fulfill. |

The sales fixture has SAR 200 subtotal, SAR 30 VAT and SAR 230 total. Contract value remains SAR 230. Generic status PATCH cannot bypass approval. Current-user permissions are supplied by backend RBAC rather than duplicated frontend role rules. Existing grants were not expanded.

`modules_test.go` additionally verifies service edits, inactive state and deletion; document metadata updates; separate downloads of original and replacement file versions; entity link creation and unlinking. Existing tests cover invalid paths, partial receipts, over-reservation and other module integrity gates.

Frontend: 14 tests pass, including password confirmation and UTF-8 byte-limit validation. Checked production build and Go vet pass. The parser bundle warning remains.

## Manual owner acceptance checklist

Use the local application at `http://localhost:5174`. The existing administrator is `admin@crm.local`; its password remains in the ignored `.env.local-login`, not this document. Create real staff accounts through **User Management** with unique passwords and the intended roles. Passwords are not emailed automatically.

1. As administrator, create a Sales Executive and Sales Manager. Confirm their teams, roles and active state. Use separate browser profiles/sessions for the following steps.
2. As sales, create a customer and opportunity; start a quote, add a catalog service/product, save, navigate away and reopen. Check row order, quantities, discounts, currency and totals.
3. Submit the quote. Sales should see read-only content without approval controls. Manager should see Approve/Reject. Reject once, edit/resubmit as sales, then approve as manager.
4. As sales, mark the quote Sent, then Accepted. These actions record status only; they do not deliver email or capture a customer signature.
5. As manager, select **Create Contract**, enter valid start/end dates and submit. Confirm one contract with the accepted quote value/currency, then activate it. Reopen as sales and confirm persisted status.
6. In **Documents**, upload a clearly named test file. Open **Details & versions**, upload different content, download each version, edit metadata, link it to a record and unlink it. Confirm errors are visible for an invalid/oversized file.
7. In **Service Catalog**, add/edit/deactivate a test service. Verify eligibility for new catalog selection and retention of already-saved quote snapshots.
8. As warehouse manager, select a stock item, set a reorder level, reserve available units and fulfill the reservation with a dispatch reason. Verify stock and movement history. Use test stock only.
9. In **My Profile**, edit personal details and reload. For a test account, change the password and sign back in with the new password. As admin, verify reset and deactivation on that test account. Never deactivate the only active administrator.

## Browser evidence and remaining limits

Captured evidence: [desktop profile/navigation](review-evidence/2026-09-17/profile-desktop.png), [mobile user editor](review-evidence/2026-09-17/users-mobile.png), [mobile service editor](review-evidence/2026-09-17/services-mobile.png).

This batch inspected administrator user/service editors and profile forms on desktop and 390px mobile. Native dialog Escape, account-menu Space, mobile navigation and no page-width overflow at 390px were verified. A separate code/screenshot review scored its two navigation accessibility findings resolved. The complete checklist above has automated API coverage for the core path, but has not yet been executed as a full separate-role browser UAT session.

Existing SAR/15% defaults, approval policy, money/FX rounding and data visibility still require business confirmation. Current role grants are authoritative but are not a newly approved policy. Access tokens can survive password changes until expiry; refresh sessions are revoked, and deactivation blocks requests immediately. Recovery email is not configured.

Next work: quote print/CSV and unsaved-change protection, remaining module exceptions, scalable table pagination, then target-server configuration, migration upgrade/rollback, backup restore, load checks and production UAT. No production deployment was performed.
