# Review evidence — 2026-09-16

These are baseline observations, not a passing regression suite. All API/browser data used for testing was synthetic and isolated from existing projects.

| Artifact | Description |
|---|---|
| [frontend-typecheck.log](frontend-typecheck.log) | Direct `npm run type-check`; 51 diagnostics. |
| [frontend-bundle.log](frontend-bundle.log) | `npm run build-only`; successful bundling and size warning. |
| [npm-audit.json](npm-audit.json) | Registry audit snapshot; triage needed before drawing exploitability conclusions. |
| [api-probes.log](api-probes.log) | Real Gin router + PostgreSQL responses; includes normal request logger output. |
| [api-probe-source.go.txt](api-probe-source.go.txt) | Temporary probe source retained as text for reproducibility; not an application command or production seed. |
| [browser-observations.md](browser-observations.md) | Desktop/mobile and console observations. |

## Commands and results

- In `crm-dashboard`: `npm run build` failed when its type-check process exited 2. A direct type-check was captured separately. `npm run build-only` passed but does not validate types.
- In `crm-api`: `go test ./... && go vet ./...` exited 0. All 12 packages reported `[no test files]`; vet emitted no diagnostics. No application tests existed before or were added during this review.
- Applied `crm-api/migrations/000001_init_schema.up.sql` using `psql -v ON_ERROR_STOP=1` on a fresh PostgreSQL 16 container: success, 36 tables. Migration rollback and production migration-runner execution were not tested.
- `npm audit --json` exited nonzero for reported advisories; its JSON is retained without performing `audit fix`.

## API probe method

The source uses GORM against a disposable database and the actual application's router with real signed access tokens and existing RBAC. It creates synthetic admin/sales accounts, logs in, calls endpoints through `httptest`, and reads results. The validation helper is also tested in a minimal Gin route. The sort probe uses SQL dry-run with a harmless expression; no destructive injection was executed.

To reproduce, use a **fresh isolated database only**: create a disposable PostgreSQL 16 container bound to `127.0.0.1:55432` with database `crm_review`, user `review`, and the synthetic local password found in the probe source. Apply the initial migration. Copy the `.go.txt` file to `crm-api/.review-runtime/main.go`, generate a temporary RSA pair at `.review-runtime/private.pem` and `.review-runtime/public.pem`, then run `go run ./.review-runtime` from `crm-api`. The fixed loopback port and credentials are solely for this disposable review harness; do not reuse them for deployment. Reusing a populated review database will cause fixture conflicts.

For the original browser pass, `REVIEW_SERVE=1` kept the same router running at `127.0.0.1:18080`. Vite was started programmatically with its existing config plus a temporary `/api` proxy override to that port, on `127.0.0.1:5174`; no project config edit was needed. Browsing used the seeded review account, not a mocked authentication response.

Probe successes also matter: initial customer create, login, and camelCase `companyName` PATCH succeeded. Therefore this review does **not** claim every camelCase update fails. Generic updates remain unsafe because they accept fields they should not permit.

The `expiring contracts` probe returned 200/null on an empty table. It does not by itself distinguish an empty result from a suppressed query error; do not treat it as proof of a working expiry query.

The temporary container, API/dev processes, fixtures and private keys were removed after inspection. The retained source and logs contain only synthetic fixture identities, not real credentials or records. Logs capture the tested baseline and will become historical after fixes.
