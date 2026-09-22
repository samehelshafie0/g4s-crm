# Local development and verification

This setup uses a separate `g4s-crm-dev` Compose project, database volume and loopback-only ports. It does not use another project's PostgreSQL or Redis. Redis is not needed for the implemented API paths.

Prerequisites: Docker with Compose, Go matching `crm-api/go.mod`, Node 22.12+ and npm. No separately installed migration CLI is required.

## First startup

From the repository root:

```sh
make dev-db
make dev-migrate
cd crm-dashboard && npm ci
```

Create the first administrator by supplying `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_FIRST_NAME` and `ADMIN_LAST_NAME` in your shell, then run `make dev-admin` from the root. The password must be 12–72 bytes. Avoid putting passwords directly in command history; use your shell's hidden-input prompt or a secret manager. Bootstrap refuses to create another administrator when an active administrator exists and never resets an existing account.

Run these in separate terminals:

```sh
make dev-api
make dev-web
```

Open http://127.0.0.1:5174. The API is on http://127.0.0.1:18080; PostgreSQL is on `127.0.0.1:55432`. The database credentials in `compose.dev.yml` are for this isolated development environment only. The script generates ignored development RSA keys once and preserves them on subsequent runs.

Optional overrides: `CRM_DB_PORT`, `CRM_API_PORT`, `CRM_WEB_PORT`. When changing the API port, also set `CRM_API_URL=http://127.0.0.1:<port>` for the frontend. Vite fails clearly if its port is occupied.

Stop the terminals with Ctrl+C. Stop local PostgreSQL with:

```sh
docker compose -f compose.dev.yml down
```

The named database volume survives this command. Do not add `-v` unless deliberately erasing development data.

## Checks

```sh
cd crm-dashboard && npm test && npm run build
cd ../crm-api && go test ./... && go vet ./...
cd .. && make test-integration
```

`make test-integration` creates a temporary PostgreSQL container on an automatically assigned loopback port, runs Go tests with the race detector, then removes that container. Tests run real migrations, handlers, permissions and transactions. Ordinary `go test` skips the PostgreSQL suite if `CRM_TEST_DATABASE_URL` is absent; CI supplies a disposable database and runs it.

The 13 frontend tests cover refresh coordination, one-retry limits, bad login handling, missing-token recovery, logout during refresh, XLS/XLSX column alignment, quoted CSV, unsupported files, multi-page loading, failed-page handling, editable-field filtering and procurement DTOs. The checked production build also runs TypeScript validation. CI configuration is in `.github/workflows/checks.yml`; its hosted run is not verified until this repository is pushed.

## Migrations and server preparation

The API binary supports `migrate` and `bootstrap-admin` before loading JWT keys. Migrations are embedded, transactionally applied, serialized by a PostgreSQL advisory lock and recorded with checksums in `crm_schema_migrations`. Applied SQL files must not be edited; add a new migration. `/health` is liveness; `/ready` verifies that all embedded migrations are applied with matching checksums. Normal startup refuses an unmigrated database.

An existing database created manually or through a different migration tool needs a baseline review and backup before adoption. The runner deliberately does not infer that an existing schema matches the source. There is no automatic down/force command; restore and rollback procedures remain a release task.

Root Docker commands can build the API image and run `make migrate` / `make seed` as one-off containers. Start PostgreSQL before migration. Full server deployment still needs secret/key ownership, TLS, backups, restore testing, ongoing dependency checks and the remaining functional work recorded in `STATUS.md`. Do not use the development database password or temporary QA accounts on a server.

## Dependency changes in this batch

The lockfile now resolves Axios 1.20.0 and Vite 7.3.6, along with compatible patched transitive dependencies. SheetJS uses the pinned 0.20.3 tarball from its [official distribution](https://docs.sheetjs.com/docs/getting-started/installation/nodejs/), whose documentation identifies the public npm `xlsx` package as outdated. The lockfile records its integrity hash. `npm audit` reports zero vulnerabilities on 2026-09-16; this is a dependency advisory check, not a claim that the application is production-secure.

## Current local account and integration work

A local administrator `admin@crm.local` was created and browser-verified on 2026-09-16. Its password is provided to the owner and stored in the ignored, mode-0600 `.env.local-login` file; it is intentionally absent from tracked documentation. Do not rerun bootstrap to reset it. For an existing account use the authenticated password-change/admin-reset API.

The frontend is available at `http://localhost:5174` while `make dev-web` runs. API routes are listed in [API_ENDPOINTS.md](API_ENDPOINTS.md); implemented workflows and remaining limitations are in [API_COVERAGE.md](API_COVERAGE.md). Migrations through 000004 are required. User-created records survive service restarts in the development volume. A fresh database starts with an empty business catalog; add your manufacturers/products and approved FX rates through the app/API.

## Quote PDF and Office conversion

Run `make dev-pdf` once to install the pinned PDF/test dependencies in an ignored virtual environment. Install LibreOffice (`brew install --cask libreoffice` on macOS, or `libreoffice-writer libreoffice-calc fonts-dejavu-core` on Debian/Ubuntu). Restart `make dev-api`; it detects this Python environment automatically. `make test-pdf` exercises actual PDF/image/Word/Excel conversion. `PDF_PYTHON` and `SOFFICE_PATH` can override the executables. API Docker runtime includes these tools; rebuild it before using the new export on a server. See [quote appendices](QUOTE_PDF_APPENDICES.md) for limits and verification.
