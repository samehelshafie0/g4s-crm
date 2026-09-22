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

## Reference catalog seed

`make seed-catalog` (Docker) or `make dev-seed-catalog` (local development database) loads the reference catalog from `crm-api/internal/seed/catalog.json`: manufacturers and their categories, products with landed-cost and selling-price calculations, catalog services, recurring services and equipment-rental packages, plus starting SAR exchange rates.

The seed is idempotent. Records are matched by manufacturer code, product SKU, service SKU, recurring-service name or currency pair and updated in place, so it can be re-run after editing `catalog.json` without creating duplicates. Exchange-rate rows that already exist are never overwritten, so operator-entered and live-fetched rates survive a re-seed. Product costs are recalculated from the exchange rate held in the database at seed time.

Rental packages are recurring services of type `rental` built from seeded products: each line charges the product's landed cost divided by the package's contract term, plus the target margin, so a rental line never charges the full purchase price monthly.

`POST /api/v1/exchange-rates/refresh` (`exchange-rates:update`) pulls SAR rates from `https://open.er-api.com/v6/latest/SAR` and appends a history row per changed pair. Set `FX_PROVIDER_URL` to use a different provider with the same response shape. Rates are informational; quotes still store their own currency and totals.

## Importing a vendor price list or quotation

**Products → Import from vendor file** reads an Excel, CSV or text-based PDF sent by a vendor and turns its rows into catalog products. The browser parses the file and detects the part number, description, price, quantity and lead-time columns; `POST /api/v1/products/import` then owns validation, costing and the audit trail.

Column detection handles what vendor files actually look like: a title row above the header, a leading spacer column, and category banners between product groups (dropped, not imported). When a sheet prices each line several times — MSRP, distributor and discounted — the payable price wins over the list price.

The import form sets what the file cannot say: vendor name, manufacturer and category, source currency and its SAR rate, freight/customs/clearance percentages and the target margin. Landed cost and selling price are previewed per row before importing and recalculated server side on save, using the same formulas as the product editor. Rows with no price (a vendor's "Call Us") arrive unticked so they can be priced by hand or left out.

Each imported row upserts the product by SKU, refreshes its vendor entry, and writes a price-history record carrying the file name, so any figure can be traced back to the vendor document. `updateExisting` decides whether a known SKU is repriced or reported as skipped. The whole import runs in one transaction and returns per-row outcomes.

## Following a vendor file through to the customer

A supplier quotation is raised to fulfil a customer quotation, usually inside a project. The chain is recorded on both procurement records:

- **Procurement → Supplier Quotes → new quote** has *Requested for customer quote* and *Project* selectors. Both are optional and validated against existing records.
- **Convert to PO** carries those links onto the purchase order, so the order knows which customer quote and project it serves. The purchase-order form keeps its own selectors for orders raised directly.
- Purchase orders and supplier quotes return `sourceQuoteNumber` and `projectName` alongside their IDs, and both are shown on the detail views.

The original vendor document is kept with the record rather than only as a file name. **Attached files** on a supplier quote, purchase order and quote uploads the file to the document library once and links it to that record; the same document can sit on both the supplier quote and the order it produced. Document links accept `supplier-quote` in addition to the existing entity types.

### How a PDF is read

Excel and CSV parse in the browser. A PDF is read **two ways at once**, because neither method wins on every vendor layout:

- `POST /api/v1/extract/pdf-tables` uploads the PDF to the API, which runs MuPDF's table finder (`pymupdf`, a runtime dependency) in a private temporary directory and returns each table as a matrix of cleaned cells. The file is deleted immediately and never stored; attaching the document to a record stays a separate, deliberate step. Two extractions run at once, larger requests get HTTP 429, and the file is capped at 25 MB with a 60-second budget.
- The browser's own text-grouping parser runs in parallel.

Each reading is mapped with the same column detection used for spreadsheets, then scored by how many rows carry a usable price. The best is shown first and the rest are offered as buttons, so a wrong guess is one click to correct. Measured on the sample supplier quotations, server extraction recovers files the browser reads as empty (a Dell quotation goes from no rows to its three priced lines), while the browser still wins on some price lists — which is why both are kept.

Tables are ranked by money-formatted cells rather than by any number, so an address or banking block does not outrank the real price table.

Every parsed row stays editable before import, and a file that cannot be read at all can be pasted or typed. A scanned PDF holds no text layer and cannot be parsed by either method — attach it to the record and enter the lines by hand.
