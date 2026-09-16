# G4S CRM — Backend Development Plan

> Historical plan. The backend now has an implementation, but several workflows remain incomplete. See [STATUS.md](../STATUS.md) and the [September 2026 review](../docs/CRM_REVIEW.md) for verified readiness and the active task list. The owner now intends server deployment; hosting details are still open.

> **Version:** 1.1  
> **Date:** April 5, 2026  
> **Status:** Planning  
> **Frontend:** Vue 3 + Vite + Pinia + TypeScript (existing)  
> **Backend:** Go (Gin) + PostgreSQL + GORM  
> **Deployment:** On-premise (client local server)

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Architecture](#2-architecture)
3. [Tech Stack](#3-tech-stack)
4. [Authentication & Authorization](#4-authentication--authorization)
5. [Database Schema](#5-database-schema)
6. [API Design Conventions](#6-api-design-conventions)
7. [API Reference](#7-api-reference)
   - 7.1 [Auth](#71-auth)
   - 7.2 [Users & Teams](#72-users--teams)
   - 7.3 [Customers](#73-customers)
   - 7.4 [Opportunities](#74-opportunities)
   - 7.5 [Manufacturers](#75-manufacturers)
   - 7.6 [Products](#76-products)
   - 7.7 [Inventory & Warehouse](#77-inventory--warehouse)
   - 7.8 [Quotes](#78-quotes)
   - 7.9 [Projects](#79-projects)
   - 7.10 [Price Books](#710-price-books)
   - 7.11 [Exchange Rates](#711-exchange-rates)
   - 7.12 [Contracts](#712-contracts)
   - 7.13 [Recurring Services](#713-recurring-services)
   - 7.14 [Procurement](#714-procurement)
   - 7.15 [Documents & File Upload](#715-documents--file-upload)
   - 7.16 [Dashboard & Analytics](#716-dashboard--analytics)
8. [Business Logic & Workflows](#8-business-logic--workflows)
9. [Error Handling](#9-error-handling)
10. [Security Considerations](#10-security-considerations)
11. [Deployment & Infrastructure](#11-deployment--infrastructure)
12. [Development Phases](#12-development-phases)

---

## 1. System Overview

The G4S CRM is an internal enterprise system for managing the full sales-to-delivery lifecycle of a security solutions company operating in Saudi Arabia. The system covers:

| Domain | Scope |
|--------|-------|
| **Customer Management** | Companies, sites, contacts; GET/GROW classification; sector-based segmentation |
| **Sales Pipeline** | Opportunities through qualification → proposal → negotiation → close |
| **Quotation Engine** | Multi-line quotes with materials/manpower/misc categories, margin tracking, approval workflow |
| **Product Catalog** | SKU-level items with full costing chain (origin cost → FX → freight → customs → landed cost → margin → selling price) |
| **Procurement** | Purchase orders, supplier quotes (RFQ responses), goods receipts, supplier-item catalog |
| **Inventory** | Multi-warehouse stock tracking, reservations, movements, reorder alerts |
| **Contracts** | Contract lifecycle with auto-renewal tracking |
| **Recurring Services** | Guarding, maintenance, monitoring services with billing frequency |
| **Projects** | Post-sale project tracking derived from accepted quotes |
| **Price Books** | Standard, volume, contract, promotional, and customer-specific pricing |
| **Exchange Rates** | Multi-currency with SAR as base; historical rate tracking |
| **Documents** | File management linked to any entity |
| **Teams** | Internal team structure with department-based organization |
| **Dashboard** | KPIs, pipeline analytics, activity feeds, alerts |

---

## 2. Architecture

### On-Premise Deployment Architecture

```
┌──────────────────────────────────────────────────────────┐
│                 Client Server (On-Premise)                │
│                                                          │
│  ┌──────────────┐    ┌────────────────────────────────┐  │
│  │    Nginx     │───►│  g4s-crm-api (Go binary)      │  │
│  │  (reverse    │    │  ┌────────┐  ┌─────────────┐   │  │
│  │   proxy,     │    │  │  Gin   │  │  Services   │   │  │
│  │   SSL,       │    │  │ Router │  │  (Business  │   │  │
│  │   SPA static │    │  │        │  │   Logic)    │   │  │
│  │   files)     │    │  └────────┘  └─────────────┘   │  │
│  └──────────────┘    └──────┬────────────┬────────────┘  │
│        │                    │            │                │
│  ┌─────┴──────┐      ┌─────┴──────┐  ┌──┴─────────────┐ │
│  │ Vue SPA    │      │ PostgreSQL │  │ Local FS       │ │
│  │ (static    │      │            │  │ /opt/g4s/      │ │
│  │  dist/)    │      │            │  │   storage/     │ │
│  └────────────┘      └────────────┘  └────────────────┘ │
│                      ┌────────────┐                      │
│                      │   Redis    │  (optional, for      │
│                      │            │   caching/sessions)  │
│                      └────────────┘                      │
└──────────────────────────────────────────────────────────┘
```

Nginx serves the Vue SPA as static files from the `dist/` folder and reverse-proxies `/api/*` requests to the Go binary. PostgreSQL runs locally on the server. Files are stored on the local filesystem.

### Go Project Structure

```
g4s-crm-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Environment, DB, storage config
│   ├── middleware/              # Auth, RBAC, logging, rate limiter, CORS
│   ├── models/                  # GORM models (DB structs)
│   ├── handlers/                # HTTP handlers (per module)
│   │   ├── auth.go
│   │   ├── users.go
│   │   ├── customers.go
│   │   ├── opportunities.go
│   │   ├── manufacturers.go
│   │   ├── products.go
│   │   ├── inventory.go
│   │   ├── quotes.go
│   │   ├── projects.go
│   │   ├── pricebooks.go
│   │   ├── exchangerates.go
│   │   ├── contracts.go
│   │   ├── recurringservices.go
│   │   ├── procurement.go
│   │   ├── documents.go
│   │   ├── teams.go
│   │   └── dashboard.go
│   ├── services/                # Business logic layer
│   ├── repository/              # Database access layer
│   ├── dto/                     # Request/response structs + validation tags
│   ├── router/                  # Route registration
│   ├── storage/                 # Local filesystem abstraction
│   └── scheduler/               # Cron jobs (contract renewal, etc.)
├── pkg/
│   ├── response/                # Standard JSON response helpers
│   ├── pagination/              # Pagination utilities
│   ├── validator/               # Custom validation rules
│   └── seqgen/                  # Auto-number generation
├── migrations/                  # SQL migration files
├── seeds/                       # Seed data
├── storage/                     # Local file storage root (runtime)
│   ├── documents/
│   ├── products/
│   └── temp/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## 3. Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Language** | Go 1.22+ | Single-binary deployment; low memory; no runtime deps on client server |
| **Framework** | Gin | Most popular Go HTTP framework; fast, well-documented, middleware ecosystem |
| **Database** | PostgreSQL 16 | Relational data with complex joins; JSONB for flexible fields |
| **ORM** | GORM | Migrations, relations, hooks, preloading; most mature Go ORM |
| **Auth** | golang-jwt/jwt v5 | JWT handling with RS256 support |
| **Password Hashing** | golang.org/x/crypto/bcrypt | Industry-standard password hashing |
| **Validation** | go-playground/validator v10 | Struct tag-based validation with custom rules |
| **Cache** | Redis 7 (go-redis/redis) | Optional — dashboard cache, rate limiting; can run without for simpler setups |
| **File Storage** | Local filesystem | On-premise; files stored at `/opt/g4s/storage/`; served through authenticated API |
| **API Docs** | swaggo/swag | Auto-generates OpenAPI/Swagger from Go doc comments |
| **Logging** | zerolog (or slog stdlib) | Structured JSON logging; zero-allocation |
| **Cron Jobs** | robfig/cron v3 | Contract expiration checks, scheduled reports |
| **Background Jobs** | Goroutine worker pool | PDF generation, FX recalculation; no external queue dependency |
| **Config** | spf13/viper | Environment variables + YAML config files |
| **Migration** | golang-migrate/migrate | Versioned SQL migrations; CLI + library |
| **UUID** | google/uuid | UUID v4 generation for primary keys |
| **CSV/XLSX** | excelize | Export/import spreadsheet support |
| **PDF** | jung-kurt/gofpdf or go-wkhtmltopdf | Quote/report PDF generation |
| **Testing** | testing (stdlib) + testify + httptest | Unit + integration + HTTP handler tests |
| **Build** | Go toolchain + Makefile | Single binary output; cross-compilation |
| **Containerization** | Docker + docker-compose | Local dev environment; optional for production |

### Why Go for On-Premise

- **Single binary (~15 MB):** Copy one file to the server, run it. No runtime, no `node_modules`, no package manager needed on the client server.
- **Low resource usage:** ~10–20 MB idle memory vs ~100–200 MB for Node.js. Critical for shared on-premise hardware.
- **Long-term stability:** Go services run for months without restarts, no memory leaks, no GC pauses.
- **No runtime dependencies:** The client doesn't need to install Go, Node.js, or any language runtime.
- **Fast startup:** < 1 second cold start; restarts are instant.

---

## 4. Authentication & Authorization

### 4.1 Auth Flow

```
Login  ──►  POST /api/auth/login  ──►  { accessToken, refreshToken }
                                            │
                 ┌──────────────────────────┘
                 ▼
Client stores tokens ──► Authorization: Bearer <accessToken>
                                            │
                 ┌──────────────────────────┘
                 ▼
Token expired? ──► POST /api/auth/refresh ──► new { accessToken, refreshToken }
```

- **Access token:** 15-minute expiry, signed with RS256
- **Refresh token:** 7-day expiry, stored in DB, single-use with rotation
- **Password hashing:** golang.org/x/crypto/bcrypt (cost factor 12)

### 4.2 Role-Based Access Control (RBAC)

| Role | Description | Permissions |
|------|-------------|-------------|
| `admin` | System administrator | Full access to all modules |
| `sales_manager` | Sales team lead | Manage opportunities, quotes, customers; approve quotes |
| `sales_executive` | Salesperson | Create/edit own opportunities and quotes; view customers |
| `pre_sales` | Pre-sales engineer | View opportunities; create/edit quotes; manage products |
| `procurement_manager` | Procurement lead | Full procurement module access; approve POs |
| `procurement_officer` | Buyer | Create POs, supplier quotes, goods receipts |
| `warehouse_manager` | Inventory lead | Full inventory access; stock adjustments |
| `project_manager` | Delivery lead | Manage projects; view quotes and POs |
| `viewer` | Read-only user | View access to all modules, no create/edit/delete |

### 4.3 Permission Matrix

Each API endpoint enforces permissions via Gin middleware:

```go
// Route registration with RBAC middleware
quotes := api.Group("/quotes")
quotes.Use(middleware.Auth())
{
    quotes.POST("", middleware.Authorize("quotes:create"), handlers.CreateQuote)
    quotes.PATCH("/:id/approve", middleware.Authorize("quotes:approve"), handlers.ApproveQuote)
}
```

```go
// middleware/auth.go — JWT extraction + role check
func Authorize(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetCurrentUser(c)
        if !user.HasPermission(permission) {
            c.AbortWithStatusJSON(403, response.Error("FORBIDDEN", "Insufficient permissions"))
            return
        }
        c.Next()
    }
}
```

Permissions follow the pattern: `{module}:{action}` where action is `read`, `create`, `update`, `delete`, `approve`, `export`.

---

## 5. Database Schema

### 5.1 Core Tables

#### `users`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK, default gen |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL |
| `password_hash` | VARCHAR(255) | NOT NULL |
| `first_name` | VARCHAR(100) | NOT NULL |
| `last_name` | VARCHAR(100) | NOT NULL |
| `phone` | VARCHAR(20) | |
| `role` | ENUM(roles) | NOT NULL, default 'viewer' |
| `department` | ENUM(departments) | |
| `team_id` | UUID | FK → teams.id |
| `is_active` | BOOLEAN | default true |
| `last_login_at` | TIMESTAMPTZ | |
| `created_at` | TIMESTAMPTZ | default NOW() |
| `updated_at` | TIMESTAMPTZ | auto-update |

#### `customers`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `company_name` | VARCHAR(255) | NOT NULL |
| `sector` | ENUM(sectors) | NOT NULL |
| `region` | VARCHAR(100) | NOT NULL |
| `status` | ENUM('active','inactive','prospect') | NOT NULL |
| `type` | ENUM('get','grow') | NOT NULL |
| `cr_number` | VARCHAR(50) | UNIQUE |
| `vat_number` | VARCHAR(50) | |
| `notes` | TEXT | |
| `created_by` | UUID | FK → users.id |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `customer_sites`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `customer_id` | UUID | FK → customers.id, ON DELETE CASCADE |
| `name` | VARCHAR(255) | NOT NULL |
| `address` | TEXT | |
| `city` | VARCHAR(100) | |
| `region` | VARCHAR(100) | |

#### `customer_contacts`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `customer_id` | UUID | FK → customers.id, ON DELETE CASCADE |
| `site_id` | UUID | FK → customer_sites.id, nullable |
| `name` | VARCHAR(255) | NOT NULL |
| `email` | VARCHAR(255) | |
| `phone` | VARCHAR(20) | |
| `position` | VARCHAR(100) | |
| `is_primary` | BOOLEAN | default false |

#### `opportunities`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `title` | VARCHAR(255) | NOT NULL |
| `customer_id` | UUID | FK → customers.id |
| `stage` | ENUM(stages) | NOT NULL, default 'qualification' |
| `service_types` | TEXT[] | PostgreSQL array |
| `estimated_value` | DECIMAL(15,2) | |
| `estimated_cost` | DECIMAL(15,2) | |
| `estimated_margin` | DECIMAL(15,2) | |
| `win_probability` | INTEGER | CHECK 0–100 |
| `sales_executive_id` | UUID | FK → users.id |
| `pre_sales_id` | UUID | FK → users.id |
| `expected_close_date` | DATE | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `manufacturers`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `name` | VARCHAR(255) | NOT NULL |
| `code` | VARCHAR(50) | UNIQUE |
| `country` | VARCHAR(100) | |
| `contact_email` | VARCHAR(255) | |
| `contact_phone` | VARCHAR(20) | |
| `website` | VARCHAR(500) | |
| `vendor_type` | ENUM('manufacturer','supplier','both') | |
| `is_active` | BOOLEAN | default true |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `manufacturer_categories`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `manufacturer_id` | UUID | FK → manufacturers.id, ON DELETE CASCADE |
| `name` | VARCHAR(255) | NOT NULL |
| `description` | TEXT | |

#### `products`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `sku` | VARCHAR(50) | UNIQUE, NOT NULL |
| `name` | VARCHAR(255) | NOT NULL |
| `description` | TEXT | |
| `manufacturer_id` | UUID | FK → manufacturers.id |
| `category_id` | UUID | FK → manufacturer_categories.id |
| `product_type` | ENUM('import','local') | NOT NULL |
| `origin_currency` | ENUM(currencies) | NOT NULL |
| `unit_cost_origin` | DECIMAL(15,4) | |
| `fx_rate` | DECIMAL(12,6) | |
| `cost_in_sar` | DECIMAL(15,4) | GENERATED |
| `freight_percent` | DECIMAL(5,2) | default 0 |
| `customs_percent` | DECIMAL(5,2) | default 0 |
| `clearance_percent` | DECIMAL(5,2) | default 0 |
| `landed_cost_sar` | DECIMAL(15,4) | GENERATED |
| `target_margin_percent` | DECIMAL(5,2) | |
| `selling_price` | DECIMAL(15,4) | |
| `margin_amount` | DECIMAL(15,4) | GENERATED |
| `lead_time_days` | INTEGER | |
| `supplier_name` | VARCHAR(255) | |
| `is_active` | BOOLEAN | default true |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

> **Note:** `cost_in_sar`, `landed_cost_sar`, and `margin_amount` are computed columns (PostgreSQL GENERATED ALWAYS AS).

#### `product_vendor_entries`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `vendor_name` | VARCHAR(255) | NOT NULL |
| `vendor_sku` | VARCHAR(50) | |
| `unit_cost` | DECIMAL(15,4) | |
| `currency` | ENUM(currencies) | |
| `moq` | INTEGER | default 1 |
| `lead_time_days` | INTEGER | |
| `last_quote_date` | DATE | |
| `catalog_source` | VARCHAR(255) | |
| `notes` | TEXT | |

#### `product_price_records`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `date` | DATE | NOT NULL |
| `source` | ENUM('vendor-catalog','purchase-order','supplier-quote','goods-receipt','manual') | |
| `source_ref` | VARCHAR(100) | |
| `vendor_name` | VARCHAR(255) | |
| `unit_cost` | DECIMAL(15,4) | |
| `currency` | ENUM(currencies) | |
| `landing_cost` | DECIMAL(15,4) | |
| `qty` | INTEGER | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |

#### `product_documents`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `name` | VARCHAR(255) | NOT NULL |
| `doc_type` | ENUM(doc_types) | NOT NULL |
| `file_name` | VARCHAR(255) | |
| `file_size` | BIGINT | bytes |
| `file_path` | VARCHAR(500) | Relative path on local filesystem |
| `uploaded_by` | UUID | FK → users.id |
| `uploaded_at` | TIMESTAMPTZ | |
| `notes` | TEXT | |

#### `warehouse_stock`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `warehouse_location` | ENUM('riyadh-main','jeddah-branch','dammam-branch') | NOT NULL |
| `on_hand_qty` | INTEGER | default 0, CHECK >= 0 |
| `reserved_qty` | INTEGER | default 0, CHECK >= 0 |
| `available_qty` | INTEGER | GENERATED (on_hand - reserved) |
| `unit_cost` | DECIMAL(15,4) | |
| `total_value` | DECIMAL(15,4) | GENERATED |
| `reorder_level` | INTEGER | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |
| | | UNIQUE(product_id, warehouse_location) |

#### `stock_reservations`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `warehouse_location` | ENUM(locations) | NOT NULL |
| `qty` | INTEGER | NOT NULL |
| `source` | ENUM('quote','project','manual') | NOT NULL |
| `source_ref` | VARCHAR(100) | |
| `source_label` | VARCHAR(255) | |
| `customer_name` | VARCHAR(255) | |
| `reserved_by` | UUID | FK → users.id |
| `reserved_at` | TIMESTAMPTZ | |
| `release_date` | TIMESTAMPTZ | |
| `released_by` | UUID | FK → users.id |
| `release_reason` | TEXT | |
| `status` | ENUM('active','released','fulfilled') | default 'active' |
| `notes` | TEXT | |

#### `inventory_movements`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `product_id` | UUID | FK → products.id |
| `movement_type` | ENUM('transfer','adjustment','allocation','receipt','release','write-off') | |
| `qty` | INTEGER | NOT NULL |
| `from_warehouse` | ENUM(locations) | nullable |
| `to_warehouse` | ENUM(locations) | nullable |
| `reference` | VARCHAR(100) | |
| `reason` | TEXT | |
| `performed_by` | UUID | FK → users.id |
| `performed_at` | TIMESTAMPTZ | default NOW() |
| `notes` | TEXT | |

#### `quotes`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `quote_number` | VARCHAR(50) | UNIQUE, auto-generated |
| `opportunity_id` | UUID | FK → opportunities.id |
| `customer_id` | UUID | FK → customers.id |
| `version` | INTEGER | default 1 |
| `status` | ENUM(quote_statuses) | default 'draft' |
| `subtotal` | DECIMAL(15,2) | |
| `discount_percent` | DECIMAL(5,2) | default 0 |
| `discount_amount` | DECIMAL(15,2) | |
| `subtotal_after_discount` | DECIMAL(15,2) | |
| `vat_percent` | DECIMAL(5,2) | default 15 (Saudi VAT) |
| `vat_amount` | DECIMAL(15,2) | |
| `total` | DECIMAL(15,2) | |
| `total_cost` | DECIMAL(15,2) | |
| `margin_amount` | DECIMAL(15,2) | |
| `margin_percent` | DECIMAL(5,2) | |
| `valid_until` | DATE | |
| `currency` | ENUM(currencies) | default 'SAR' |
| `notes` | TEXT | |
| `approved_by` | UUID | FK → users.id |
| `approved_at` | TIMESTAMPTZ | |
| `created_by` | UUID | FK → users.id |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `quote_line_items`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `quote_id` | UUID | FK → quotes.id, ON DELETE CASCADE |
| `category` | ENUM('materials','manpower','miscellaneous') | NOT NULL |
| `product_id` | UUID | FK → products.id, nullable |
| `sku` | VARCHAR(50) | |
| `description` | TEXT | NOT NULL |
| `manufacturer_name` | VARCHAR(255) | |
| `stock_available` | INTEGER | |
| `lead_time_days` | INTEGER | |
| `quantity` | INTEGER | NOT NULL |
| `unit_cost` | DECIMAL(15,4) | |
| `unit_price` | DECIMAL(15,4) | |
| `line_total` | DECIMAL(15,2) | GENERATED |
| `margin_percent` | DECIMAL(5,2) | GENERATED |
| `sort_order` | INTEGER | default 0 |

#### `projects`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `project_number` | VARCHAR(50) | UNIQUE |
| `name` | VARCHAR(255) | NOT NULL |
| `customer_id` | UUID | FK → customers.id |
| `quote_id` | UUID | FK → quotes.id |
| `status` | ENUM('planning','in-progress','on-hold','completed','cancelled') | |
| `priority` | ENUM('low','medium','high','critical') | default 'medium' |
| `start_date` | DATE | |
| `target_end_date` | DATE | |
| `actual_end_date` | DATE | |
| `project_manager_id` | UUID | FK → users.id |
| `total_value` | DECIMAL(15,2) | |
| `total_cost` | DECIMAL(15,2) | |
| `margin_percent` | DECIMAL(5,2) | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `price_books`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `name` | VARCHAR(255) | NOT NULL |
| `type` | ENUM('standard','volume','contract','promotional','customer-specific') | |
| `description` | TEXT | |
| `customer_id` | UUID | FK → customers.id, nullable |
| `contract_id` | UUID | FK → contracts.id, nullable |
| `valid_from` | DATE | |
| `valid_to` | DATE | |
| `is_active` | BOOLEAN | default true |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `price_book_entries`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `price_book_id` | UUID | FK → price_books.id, ON DELETE CASCADE |
| `product_id` | UUID | FK → products.id |
| `standard_price` | DECIMAL(15,4) | |
| `custom_price` | DECIMAL(15,4) | |
| `discount_percent` | DECIMAL(5,2) | |

#### `exchange_rates`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `from_currency` | ENUM(currencies) | NOT NULL |
| `to_currency` | VARCHAR(3) | default 'SAR' |
| `current_rate` | DECIMAL(12,6) | NOT NULL |
| `effective_date` | DATE | NOT NULL |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |
| | | UNIQUE(from_currency, to_currency) |

#### `exchange_rate_history`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `exchange_rate_id` | UUID | FK → exchange_rates.id |
| `rate` | DECIMAL(12,6) | NOT NULL |
| `effective_date` | DATE | NOT NULL |
| `created_at` | TIMESTAMPTZ | |

#### `contracts`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `contract_number` | VARCHAR(50) | UNIQUE |
| `title` | VARCHAR(255) | NOT NULL |
| `customer_id` | UUID | FK → customers.id |
| `type` | ENUM('sales','maintenance','service','project','subscription') | |
| `status` | ENUM('draft','pending-approval','active','expired','terminated','renewed') | |
| `start_date` | DATE | |
| `end_date` | DATE | |
| `value` | DECIMAL(15,2) | |
| `auto_renew` | BOOLEAN | default false |
| `renewal_notice_days` | INTEGER | default 30 |
| `renewed_from_id` | UUID | FK → contracts.id, nullable |
| `quote_id` | UUID | FK → quotes.id, nullable |
| `terms` | TEXT | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `recurring_services`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `name` | VARCHAR(255) | NOT NULL |
| `service_type` | ENUM('guarding','maintenance','monitoring','patrol','facility-management') | |
| `description` | TEXT | |
| `monthly_cost` | DECIMAL(15,2) | |
| `monthly_price` | DECIMAL(15,2) | |
| `annual_cost` | DECIMAL(15,2) | GENERATED |
| `annual_price` | DECIMAL(15,2) | GENERATED |
| `target_margin_percent` | DECIMAL(5,2) | |
| `billing_frequency` | ENUM('monthly','quarterly','annually') | |
| `is_active` | BOOLEAN | default true |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `purchase_orders`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `po_number` | VARCHAR(50) | UNIQUE |
| `supplier_id` | UUID | FK → manufacturers.id, nullable |
| `supplier_name` | VARCHAR(255) | NOT NULL |
| `status` | ENUM(po_statuses) | default 'draft' |
| `subtotal` | DECIMAL(15,2) | |
| `shipping_cost` | DECIMAL(15,2) | default 0 |
| `customs_duty` | DECIMAL(15,2) | default 0 |
| `total` | DECIMAL(15,2) | |
| `currency` | ENUM(currencies) | default 'SAR' |
| `expected_delivery` | DATE | |
| `actual_delivery` | DATE | |
| `source_quote_id` | UUID | FK → quotes.id, nullable |
| `notes` | TEXT | |
| `approved_by` | UUID | FK → users.id |
| `approved_at` | TIMESTAMPTZ | |
| `created_by` | UUID | FK → users.id |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `purchase_order_items`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `purchase_order_id` | UUID | FK → purchase_orders.id, ON DELETE CASCADE |
| `product_id` | UUID | FK → products.id |
| `quantity` | INTEGER | NOT NULL |
| `unit_cost` | DECIMAL(15,4) | |
| `total` | DECIMAL(15,2) | |
| `received_qty` | INTEGER | default 0 |
| `lead_time_days` | INTEGER | |

#### `supplier_quotes`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `sq_number` | VARCHAR(50) | UNIQUE |
| `supplier_id` | UUID | FK → manufacturers.id, nullable |
| `supplier_name` | VARCHAR(255) | NOT NULL |
| `supplier_ref` | VARCHAR(100) | |
| `status` | ENUM('received','under-review','accepted','expired','rejected') | |
| `subtotal` | DECIMAL(15,2) | |
| `currency` | ENUM(currencies) | |
| `valid_from` | DATE | |
| `valid_until` | DATE | |
| `contact_name` | VARCHAR(255) | |
| `contact_email` | VARCHAR(255) | |
| `payment_terms` | VARCHAR(255) | |
| `delivery_terms` | VARCHAR(255) | |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `supplier_quote_line_items`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `supplier_quote_id` | UUID | FK → supplier_quotes.id, ON DELETE CASCADE |
| `product_id` | UUID | FK → products.id, nullable |
| `product_sku` | VARCHAR(50) | |
| `product_name` | VARCHAR(255) | |
| `manufacturer_name` | VARCHAR(255) | |
| `quantity` | INTEGER | |
| `unit_cost` | DECIMAL(15,4) | |
| `total` | DECIMAL(15,2) | |
| `lead_time_days` | INTEGER | |
| `moq` | INTEGER | |
| `valid_until` | DATE | |
| `notes` | TEXT | |

#### `supplier_item_catalog`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `supplier_id` | UUID | FK → manufacturers.id |
| `product_id` | UUID | FK → products.id |
| `latest_cost` | DECIMAL(15,4) | |
| `previous_cost` | DECIMAL(15,4) | |
| `cost_trend` | ENUM('up','down','stable') | default 'stable' |
| `moq` | INTEGER | default 1 |
| `lead_time_days` | INTEGER | |
| `last_quote_date` | DATE | |
| `last_po_date` | DATE | |
| `reliability` | DECIMAL(3,2) | CHECK 0–1 |
| | | UNIQUE(supplier_id, product_id) |

#### `goods_receipts`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `gr_number` | VARCHAR(50) | UNIQUE |
| `po_id` | UUID | FK → purchase_orders.id |
| `supplier_name` | VARCHAR(255) | |
| `receive_date` | DATE | NOT NULL |
| `total_items` | INTEGER | |
| `total_landing_cost` | DECIMAL(15,2) | |
| `received_by` | UUID | FK → users.id |
| `notes` | TEXT | |
| `created_at` | TIMESTAMPTZ | |

#### `goods_receipt_items`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `goods_receipt_id` | UUID | FK → goods_receipts.id, ON DELETE CASCADE |
| `po_item_id` | UUID | FK → purchase_order_items.id |
| `product_id` | UUID | FK → products.id |
| `received_qty` | INTEGER | NOT NULL |
| `unit_cost` | DECIMAL(15,4) | |
| `shipping_alloc` | DECIMAL(15,4) | |
| `customs_alloc` | DECIMAL(15,4) | |
| `landing_cost` | DECIMAL(15,4) | |
| `serial_numbers` | TEXT[] | |
| `storage_location` | VARCHAR(100) | |
| `condition` | ENUM('good','damaged','partial-damage') | default 'good' |
| `notes` | TEXT | |

#### `teams`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `name` | VARCHAR(255) | NOT NULL |
| `department` | ENUM(departments) | NOT NULL |
| `description` | TEXT | |
| `leader_id` | UUID | FK → users.id |
| `is_active` | BOOLEAN | default true |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `documents`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `name` | VARCHAR(255) | NOT NULL |
| `category` | ENUM('contract','quote','general','compliance','legal') | |
| `document_type` | ENUM('terms','delivery','technical','warranty','sla') | |
| `tags` | TEXT[] | |
| `version` | VARCHAR(20) | |
| `file_name` | VARCHAR(255) | |
| `file_size` | BIGINT | |
| `file_type` | VARCHAR(50) | |
| `file_path` | VARCHAR(500) | Relative path on local filesystem |
| `uploaded_by` | UUID | FK → users.id |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

#### `document_entity_links`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `document_id` | UUID | FK → documents.id, ON DELETE CASCADE |
| `entity_type` | VARCHAR(50) | NOT NULL (e.g. 'customer', 'quote', 'contract') |
| `entity_id` | UUID | NOT NULL |
| `entity_name` | VARCHAR(255) | |

#### `activity_log`

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | UUID | PK |
| `user_id` | UUID | FK → users.id |
| `entity_type` | VARCHAR(50) | NOT NULL |
| `entity_id` | UUID | NOT NULL |
| `action` | VARCHAR(50) | NOT NULL (created, updated, deleted, status_changed, approved, etc.) |
| `description` | TEXT | |
| `old_values` | JSONB | nullable |
| `new_values` | JSONB | nullable |
| `created_at` | TIMESTAMPTZ | default NOW() |

### 5.2 Key Indexes

```sql
-- Customers
CREATE INDEX idx_customers_sector ON customers(sector);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_type ON customers(type);
CREATE INDEX idx_customers_company_name ON customers(company_name);

-- Opportunities
CREATE INDEX idx_opportunities_customer ON opportunities(customer_id);
CREATE INDEX idx_opportunities_stage ON opportunities(stage);
CREATE INDEX idx_opportunities_sales_exec ON opportunities(sales_executive_id);
CREATE INDEX idx_opportunities_close_date ON opportunities(expected_close_date);

-- Products
CREATE INDEX idx_products_manufacturer ON products(manufacturer_id);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_active ON products(is_active);

-- Quotes
CREATE INDEX idx_quotes_customer ON quotes(customer_id);
CREATE INDEX idx_quotes_opportunity ON quotes(opportunity_id);
CREATE INDEX idx_quotes_status ON quotes(status);
CREATE INDEX idx_quotes_number ON quotes(quote_number);

-- Purchase Orders
CREATE INDEX idx_po_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_source_quote ON purchase_orders(source_quote_id);

-- Warehouse Stock
CREATE INDEX idx_stock_product ON warehouse_stock(product_id);
CREATE INDEX idx_stock_location ON warehouse_stock(warehouse_location);

-- Activity Log
CREATE INDEX idx_activity_entity ON activity_log(entity_type, entity_id);
CREATE INDEX idx_activity_user ON activity_log(user_id);
CREATE INDEX idx_activity_created ON activity_log(created_at DESC);

-- Contracts
CREATE INDEX idx_contracts_customer ON contracts(customer_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_end_date ON contracts(end_date);
```

---

## 6. API Design Conventions

### 6.1 Base URL

```
Production:  https://<client-server-ip>/api/v1    (behind Nginx reverse proxy)
Development: http://localhost:8080/api/v1
```

### 6.2 Request / Response Format

All requests and responses use **JSON** (`Content-Type: application/json`).

### 6.3 Pagination

All list endpoints support cursor-based or offset pagination:

```
GET /api/v1/customers?page=1&limit=25&sort=company_name&order=asc
```

**Response envelope:**

```json
{
  "success": true,
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 25,
    "total": 142,
    "totalPages": 6
  }
}
```

### 6.4 Filtering

Filters are passed as query parameters:

```
GET /api/v1/customers?sector=healthcare&status=active&type=grow
GET /api/v1/opportunities?stage=proposal&salesExecutiveId=uuid
GET /api/v1/products?manufacturerId=uuid&isActive=true
```

### 6.5 Search

Full-text search via `q` parameter:

```
GET /api/v1/customers?q=aramco
GET /api/v1/products?q=hikvision+camera
```

### 6.6 Field Selection (Sparse Fields)

```
GET /api/v1/customers?fields=id,companyName,sector,status
```

### 6.7 Resource Expansion (Includes)

```
GET /api/v1/quotes/uuid?include=lineItems,customer,opportunity
GET /api/v1/purchase-orders/uuid?include=items,goodsReceipts
```

### 6.8 Bulk Operations

```
POST   /api/v1/products/bulk          { items: [...] }
PATCH  /api/v1/products/bulk-update   { ids: [...], update: {...} }
DELETE /api/v1/products/bulk-delete   { ids: [...] }
```

### 6.9 HTTP Status Codes

| Code | Usage |
|------|-------|
| `200` | Success (GET, PATCH, actions) |
| `201` | Created (POST) |
| `204` | No Content (DELETE) |
| `400` | Bad Request (validation error) |
| `401` | Unauthorized (missing/invalid token) |
| `403` | Forbidden (insufficient permissions) |
| `404` | Not Found |
| `409` | Conflict (duplicate CR number, SKU, etc.) |
| `422` | Unprocessable Entity (business rule violation) |
| `429` | Too Many Requests (rate limit) |
| `500` | Internal Server Error |

### 6.10 Versioning

API version is embedded in the URL path: `/api/v1/...`. Breaking changes result in a new version (`v2`).

---

## 7. API Reference

### 7.1 Auth

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/auth/register` | Register new user (admin only) | Admin |
| `POST` | `/auth/login` | Authenticate user | Public |
| `POST` | `/auth/refresh` | Refresh access token | Refresh Token |
| `POST` | `/auth/logout` | Revoke refresh token | Bearer |
| `POST` | `/auth/forgot-password` | Send password reset email | Public |
| `POST` | `/auth/reset-password` | Reset password with token | Public |
| `PATCH` | `/auth/change-password` | Change own password | Bearer |
| `GET` | `/auth/me` | Get current user profile | Bearer |

**POST `/auth/login`**

Request:
```json
{
  "email": "admin@g4s-crm.com",
  "password": "securePassword123"
}
```

Response `200`:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "admin@g4s-crm.com",
      "firstName": "Admin",
      "lastName": "User",
      "role": "admin",
      "department": "management"
    },
    "accessToken": "eyJhbGciOiJSUz...",
    "refreshToken": "dGhpcyBpcyBhI..."
  }
}
```

---

### 7.2 Users & Teams

#### Users

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/users` | List users (with filters) | `users:read` |
| `GET` | `/users/:id` | Get user details | `users:read` |
| `POST` | `/users` | Create user | `users:create` |
| `PATCH` | `/users/:id` | Update user | `users:update` |
| `DELETE` | `/users/:id` | Deactivate user (soft delete) | `users:delete` |
| `GET` | `/users/lookup` | Lookup users by role/department (for dropdowns) | Any authenticated |

#### Teams

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/teams` | List teams | `teams:read` |
| `GET` | `/teams/:id` | Get team with members | `teams:read` |
| `POST` | `/teams` | Create team | `teams:create` |
| `PATCH` | `/teams/:id` | Update team | `teams:update` |
| `DELETE` | `/teams/:id` | Delete team | `teams:delete` |
| `POST` | `/teams/:id/members` | Add member to team | `teams:update` |
| `DELETE` | `/teams/:id/members/:userId` | Remove member from team | `teams:update` |

---

### 7.3 Customers

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/customers` | List customers (paginated, filtered) | `customers:read` |
| `GET` | `/customers/:id` | Get customer with sites & contacts | `customers:read` |
| `POST` | `/customers` | Create customer | `customers:create` |
| `PATCH` | `/customers/:id` | Update customer | `customers:update` |
| `DELETE` | `/customers/:id` | Delete customer | `customers:delete` |
| `GET` | `/customers/:id/sites` | List sites for a customer | `customers:read` |
| `POST` | `/customers/:id/sites` | Add site | `customers:update` |
| `PATCH` | `/customers/:id/sites/:siteId` | Update site | `customers:update` |
| `DELETE` | `/customers/:id/sites/:siteId` | Delete site | `customers:update` |
| `GET` | `/customers/:id/contacts` | List contacts for a customer | `customers:read` |
| `POST` | `/customers/:id/contacts` | Add contact | `customers:update` |
| `PATCH` | `/customers/:id/contacts/:contactId` | Update contact | `customers:update` |
| `DELETE` | `/customers/:id/contacts/:contactId` | Delete contact | `customers:update` |
| `GET` | `/customers/:id/opportunities` | List opportunities for customer | `customers:read` |
| `GET` | `/customers/:id/quotes` | List quotes for customer | `customers:read` |
| `GET` | `/customers/:id/contracts` | List contracts for customer | `customers:read` |
| `GET` | `/customers/export` | Export customers to CSV/XLSX | `customers:export` |

**Query Parameters for `GET /customers`:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search company name, CR number |
| `sector` | enum | Filter by sector |
| `status` | enum | Filter by status |
| `type` | enum | Filter by GET/GROW |
| `region` | string | Filter by region |
| `page` | integer | Page number (default 1) |
| `limit` | integer | Items per page (default 25, max 100) |
| `sort` | string | Sort field (default `company_name`) |
| `order` | asc/desc | Sort order (default `asc`) |

**POST `/customers`** — Request Body:
```json
{
  "companyName": "Saudi Aramco",
  "sector": "oil-gas",
  "region": "Eastern Province",
  "status": "active",
  "type": "grow",
  "crNumber": "1010000001",
  "vatNumber": "300000000000003",
  "notes": "Key strategic account",
  "sites": [
    {
      "name": "Dhahran HQ",
      "address": "King Khalid Road",
      "city": "Dhahran",
      "region": "Eastern Province",
      "contacts": []
    }
  ],
  "contacts": [
    {
      "name": "Ahmed Al-Rashid",
      "email": "ahmed@aramco.com",
      "phone": "+966501234567",
      "position": "Security Director",
      "isPrimary": true
    }
  ]
}
```

---

### 7.4 Opportunities

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/opportunities` | List opportunities | `opportunities:read` |
| `GET` | `/opportunities/:id` | Get opportunity detail | `opportunities:read` |
| `POST` | `/opportunities` | Create opportunity | `opportunities:create` |
| `PATCH` | `/opportunities/:id` | Update opportunity | `opportunities:update` |
| `DELETE` | `/opportunities/:id` | Delete opportunity | `opportunities:delete` |
| `PATCH` | `/opportunities/:id/stage` | Move to a new pipeline stage | `opportunities:update` |
| `GET` | `/opportunities/:id/quotes` | List quotes linked to opportunity | `opportunities:read` |
| `GET` | `/opportunities/pipeline` | Get pipeline summary (counts & values per stage) | `opportunities:read` |
| `GET` | `/opportunities/export` | Export to CSV/XLSX | `opportunities:export` |

**Query Parameters for `GET /opportunities`:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Search title, customer name |
| `stage` | enum | Filter by stage |
| `customerId` | UUID | Filter by customer |
| `salesExecutiveId` | UUID | Filter by assigned sales exec |
| `preSalesId` | UUID | Filter by pre-sales engineer |
| `serviceTypes` | string (comma-separated) | Filter by service type |
| `minValue` | number | Minimum estimated value |
| `maxValue` | number | Maximum estimated value |
| `expectedCloseBefore` | date | Close date upper bound |
| `expectedCloseAfter` | date | Close date lower bound |

**PATCH `/opportunities/:id/stage`** — Request Body:
```json
{
  "stage": "proposal",
  "notes": "Client requested formal proposal"
}
```

---

### 7.5 Manufacturers

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/manufacturers` | List manufacturers/suppliers | `manufacturers:read` |
| `GET` | `/manufacturers/:id` | Get manufacturer detail | `manufacturers:read` |
| `POST` | `/manufacturers` | Create manufacturer | `manufacturers:create` |
| `PATCH` | `/manufacturers/:id` | Update manufacturer | `manufacturers:update` |
| `DELETE` | `/manufacturers/:id` | Delete manufacturer | `manufacturers:delete` |
| `GET` | `/manufacturers/:id/categories` | List categories | `manufacturers:read` |
| `POST` | `/manufacturers/:id/categories` | Add category | `manufacturers:update` |
| `PATCH` | `/manufacturers/:id/categories/:catId` | Update category | `manufacturers:update` |
| `DELETE` | `/manufacturers/:id/categories/:catId` | Delete category | `manufacturers:update` |
| `GET` | `/manufacturers/:id/products` | List products by manufacturer | `manufacturers:read` |

---

### 7.6 Products

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/products` | List products (paginated, filtered) | `products:read` |
| `GET` | `/products/:id` | Get product with vendor entries, price history, docs | `products:read` |
| `POST` | `/products` | Create product | `products:create` |
| `PATCH` | `/products/:id` | Update product | `products:update` |
| `DELETE` | `/products/:id` | Delete product (soft) | `products:delete` |
| `POST` | `/products/bulk` | Bulk import products (from CSV/XLSX) | `products:create` |
| `GET` | `/products/:id/vendor-entries` | List vendor catalog entries | `products:read` |
| `POST` | `/products/:id/vendor-entries` | Add vendor entry | `products:update` |
| `PATCH` | `/products/:id/vendor-entries/:entryId` | Update vendor entry | `products:update` |
| `DELETE` | `/products/:id/vendor-entries/:entryId` | Delete vendor entry | `products:update` |
| `GET` | `/products/:id/price-history` | List price records | `products:read` |
| `POST` | `/products/:id/price-history` | Add price record | `products:update` |
| `GET` | `/products/:id/documents` | List product documents | `products:read` |
| `POST` | `/products/:id/documents` | Upload product document | `products:update` |
| `DELETE` | `/products/:id/documents/:docId` | Delete product document | `products:update` |
| `POST` | `/products/recalculate-costs` | Recalculate landed costs for FX/freight changes | `products:update` |
| `GET` | `/products/export` | Export products to CSV/XLSX | `products:export` |
| `GET` | `/products/import-template` | Download CSV/XLSX import template | `products:read` |

**Product costing recalculation logic:**

```
cost_in_sar       = unit_cost_origin × fx_rate
freight_amount    = cost_in_sar × freight_percent / 100
customs_amount    = cost_in_sar × customs_percent / 100
clearance_amount  = cost_in_sar × clearance_percent / 100
landed_cost_sar   = cost_in_sar + freight_amount + customs_amount + clearance_amount
margin_amount     = selling_price - landed_cost_sar
margin_percent    = (margin_amount / selling_price) × 100
```

---

### 7.7 Inventory & Warehouse

#### Warehouse Stock

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/inventory/stock` | List all stock across warehouses | `inventory:read` |
| `GET` | `/inventory/stock/:id` | Get stock entry detail | `inventory:read` |
| `PATCH` | `/inventory/stock/:id` | Update stock levels / reorder level | `inventory:update` |
| `GET` | `/inventory/stock/low-stock` | Get items below reorder level | `inventory:read` |
| `GET` | `/inventory/stock/by-warehouse/:location` | Get stock at a specific warehouse | `inventory:read` |
| `GET` | `/inventory/stock/by-product/:productId` | Get stock for a product across warehouses | `inventory:read` |

#### Stock Reservations

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/inventory/reservations` | List reservations | `inventory:read` |
| `GET` | `/inventory/reservations/:id` | Get reservation detail | `inventory:read` |
| `POST` | `/inventory/reservations` | Create reservation | `inventory:create` |
| `PATCH` | `/inventory/reservations/:id/release` | Release reservation | `inventory:update` |
| `PATCH` | `/inventory/reservations/:id/fulfill` | Mark reservation fulfilled | `inventory:update` |

**POST `/inventory/reservations`:**
```json
{
  "productId": "uuid",
  "warehouseLocation": "riyadh-main",
  "qty": 10,
  "source": "quote",
  "sourceRef": "QT-2026-001",
  "sourceLabel": "Saudi Aramco CCTV Quote",
  "customerName": "Saudi Aramco",
  "notes": "Reserved for Q2 delivery"
}
```

#### Inventory Movements

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/inventory/movements` | List movements (filterable) | `inventory:read` |
| `POST` | `/inventory/movements/transfer` | Transfer stock between warehouses | `inventory:create` |
| `POST` | `/inventory/movements/adjustment` | Adjust stock (count correction) | `inventory:create` |
| `POST` | `/inventory/movements/write-off` | Write off damaged/lost stock | `inventory:create` |

**POST `/inventory/movements/transfer`:**
```json
{
  "productId": "uuid",
  "qty": 5,
  "fromWarehouse": "riyadh-main",
  "toWarehouse": "jeddah-branch",
  "reason": "Branch replenishment",
  "notes": "Requested by Jeddah ops team"
}
```

> **Side effect:** Transfer creates two stock updates (decrement source, increment destination) within a database transaction.

---

### 7.8 Quotes

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/quotes` | List quotes (paginated, filtered) | `quotes:read` |
| `GET` | `/quotes/:id` | Get quote with line items | `quotes:read` |
| `POST` | `/quotes` | Create quote (draft) | `quotes:create` |
| `PATCH` | `/quotes/:id` | Update quote (header fields) | `quotes:update` |
| `DELETE` | `/quotes/:id` | Delete quote (draft only) | `quotes:delete` |
| `POST` | `/quotes/:id/line-items` | Add line item | `quotes:update` |
| `PATCH` | `/quotes/:id/line-items/:itemId` | Update line item | `quotes:update` |
| `DELETE` | `/quotes/:id/line-items/:itemId` | Remove line item | `quotes:update` |
| `PATCH` | `/quotes/:id/line-items/reorder` | Reorder line items | `quotes:update` |
| `POST` | `/quotes/:id/duplicate` | Duplicate quote (new version) | `quotes:create` |
| `PATCH` | `/quotes/:id/submit` | Submit for approval | `quotes:update` |
| `PATCH` | `/quotes/:id/approve` | Approve quote | `quotes:approve` |
| `PATCH` | `/quotes/:id/reject` | Reject quote (back to draft) | `quotes:approve` |
| `PATCH` | `/quotes/:id/send` | Mark as sent to customer | `quotes:update` |
| `PATCH` | `/quotes/:id/accept` | Mark as accepted by customer | `quotes:update` |
| `PATCH` | `/quotes/:id/decline` | Mark as declined by customer | `quotes:update` |
| `POST` | `/quotes/:id/generate-pdf` | Generate PDF document | `quotes:read` |
| `POST` | `/quotes/:id/recalculate` | Recalculate totals from line items | `quotes:update` |
| `GET` | `/quotes/export` | Export quotes list to CSV/XLSX | `quotes:export` |

**Quote Status Transitions:**

```
draft → pending-approval → approved → sent → accepted
                         ↘ rejected (→ draft)     ↘ declined
                                                   ↘ expired
```

**POST `/quotes`** — Request Body:
```json
{
  "opportunityId": "uuid",
  "customerId": "uuid",
  "currency": "SAR",
  "validUntil": "2026-07-01",
  "discountPercent": 5,
  "vatPercent": 15,
  "notes": "Includes 1-year warranty",
  "lineItems": [
    {
      "category": "materials",
      "productId": "uuid",
      "description": "Hikvision 4MP IP Camera DS-2CD2143G2-I",
      "quantity": 50,
      "unitCost": 450.00,
      "unitPrice": 675.00
    },
    {
      "category": "manpower",
      "description": "Installation Engineer - 5 days",
      "quantity": 5,
      "unitCost": 800.00,
      "unitPrice": 1200.00
    }
  ]
}
```

**Recalculation Logic (server-side):**

```
For each lineItem:
  line_total = quantity × unit_price
  line_cost  = quantity × unit_cost
  margin_percent = ((unit_price - unit_cost) / unit_price) × 100

subtotal             = SUM(line_totals)
discount_amount      = subtotal × discount_percent / 100
subtotal_after_disc  = subtotal - discount_amount
vat_amount           = subtotal_after_disc × vat_percent / 100
total                = subtotal_after_disc + vat_amount
total_cost           = SUM(line_costs)
margin_amount        = subtotal_after_disc - total_cost
margin_percent       = (margin_amount / subtotal_after_disc) × 100
```

---

### 7.9 Projects

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/projects` | List projects | `projects:read` |
| `GET` | `/projects/:id` | Get project detail | `projects:read` |
| `POST` | `/projects` | Create project (from accepted quote) | `projects:create` |
| `PATCH` | `/projects/:id` | Update project | `projects:update` |
| `DELETE` | `/projects/:id` | Delete project | `projects:delete` |
| `PATCH` | `/projects/:id/status` | Change project status | `projects:update` |
| `GET` | `/projects/:id/purchase-orders` | List POs linked to project | `projects:read` |
| `POST` | `/projects/:id/purchase-orders` | Link PO to project | `projects:update` |

**POST `/projects`** — Auto-created from an accepted quote:
```json
{
  "quoteId": "uuid",
  "name": "Saudi Aramco CCTV Installation",
  "startDate": "2026-05-01",
  "targetEndDate": "2026-08-01",
  "projectManagerId": "uuid",
  "priority": "high",
  "notes": "Phase 1 - Dhahran campus"
}
```

The server copies the quote's line items, customer, value, and cost into the project record.

---

### 7.10 Price Books

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/price-books` | List price books | `price-books:read` |
| `GET` | `/price-books/:id` | Get price book with entries | `price-books:read` |
| `POST` | `/price-books` | Create price book | `price-books:create` |
| `PATCH` | `/price-books/:id` | Update price book | `price-books:update` |
| `DELETE` | `/price-books/:id` | Delete price book | `price-books:delete` |
| `POST` | `/price-books/:id/entries` | Add product entry | `price-books:update` |
| `PATCH` | `/price-books/:id/entries/:entryId` | Update product entry | `price-books:update` |
| `DELETE` | `/price-books/:id/entries/:entryId` | Remove product entry | `price-books:update` |
| `POST` | `/price-books/:id/entries/bulk` | Bulk add/update entries | `price-books:update` |
| `GET` | `/price-books/lookup` | Find applicable price book for customer/product | `price-books:read` |

---

### 7.11 Exchange Rates

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/exchange-rates` | List current exchange rates | `exchange-rates:read` |
| `GET` | `/exchange-rates/:id` | Get rate with history | `exchange-rates:read` |
| `POST` | `/exchange-rates` | Create exchange rate | `exchange-rates:create` |
| `PATCH` | `/exchange-rates/:id` | Update current rate (auto-archives previous) | `exchange-rates:update` |
| `GET` | `/exchange-rates/:id/history` | Get rate history | `exchange-rates:read` |
| `POST` | `/exchange-rates/sync` | Sync rates from external API (async job) | `exchange-rates:update` |

**Side Effect:** When an exchange rate is updated, a background job can optionally trigger product cost recalculation for all products using that currency.

---

### 7.12 Contracts

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/contracts` | List contracts | `contracts:read` |
| `GET` | `/contracts/:id` | Get contract detail | `contracts:read` |
| `POST` | `/contracts` | Create contract | `contracts:create` |
| `PATCH` | `/contracts/:id` | Update contract | `contracts:update` |
| `DELETE` | `/contracts/:id` | Delete contract | `contracts:delete` |
| `PATCH` | `/contracts/:id/activate` | Activate contract | `contracts:approve` |
| `PATCH` | `/contracts/:id/terminate` | Terminate contract | `contracts:approve` |
| `POST` | `/contracts/:id/renew` | Renew contract (creates new linked contract) | `contracts:create` |
| `GET` | `/contracts/expiring` | List contracts expiring within N days | `contracts:read` |
| `GET` | `/contracts/export` | Export contracts to CSV/XLSX | `contracts:export` |

**Query Parameters for `GET /contracts/expiring`:**

| Param | Type | Description |
|-------|------|-------------|
| `days` | integer | Number of days to look ahead (default 30) |

**Contract Renewal Logic:**

When `POST /contracts/:id/renew` is called:
1. Original contract status → `renewed`
2. New contract created with `renewed_from_id` pointing to original
3. Dates auto-set: `start_date` = original `end_date` + 1 day
4. Same customer, type, value (or updated value from request body)

---

### 7.13 Recurring Services

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/recurring-services` | List recurring services | `recurring-services:read` |
| `GET` | `/recurring-services/:id` | Get service detail | `recurring-services:read` |
| `POST` | `/recurring-services` | Create recurring service | `recurring-services:create` |
| `PATCH` | `/recurring-services/:id` | Update recurring service | `recurring-services:update` |
| `DELETE` | `/recurring-services/:id` | Delete recurring service | `recurring-services:delete` |

**Computed Fields:**
```
annual_cost  = monthly_cost × 12
annual_price = monthly_price × 12
margin_percent = ((monthly_price - monthly_cost) / monthly_price) × 100
```

---

### 7.14 Procurement

#### Purchase Orders

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/procurement/purchase-orders` | List POs | `procurement:read` |
| `GET` | `/procurement/purchase-orders/:id` | Get PO with items | `procurement:read` |
| `POST` | `/procurement/purchase-orders` | Create PO | `procurement:create` |
| `PATCH` | `/procurement/purchase-orders/:id` | Update PO | `procurement:update` |
| `DELETE` | `/procurement/purchase-orders/:id` | Delete PO (draft only) | `procurement:delete` |
| `POST` | `/procurement/purchase-orders/:id/items` | Add item to PO | `procurement:update` |
| `PATCH` | `/procurement/purchase-orders/:id/items/:itemId` | Update PO item | `procurement:update` |
| `DELETE` | `/procurement/purchase-orders/:id/items/:itemId` | Remove PO item | `procurement:update` |
| `PATCH` | `/procurement/purchase-orders/:id/submit` | Submit PO for approval | `procurement:update` |
| `PATCH` | `/procurement/purchase-orders/:id/approve` | Approve PO | `procurement:approve` |
| `PATCH` | `/procurement/purchase-orders/:id/reject` | Reject PO | `procurement:approve` |
| `PATCH` | `/procurement/purchase-orders/:id/mark-ordered` | Mark as ordered (sent to supplier) | `procurement:update` |
| `PATCH` | `/procurement/purchase-orders/:id/cancel` | Cancel PO | `procurement:update` |
| `POST` | `/procurement/purchase-orders/from-quote` | Auto-generate PO from quote line items | `procurement:create` |
| `GET` | `/procurement/purchase-orders/export` | Export POs to CSV/XLSX | `procurement:export` |

**PO Status Transitions:**

```
draft → pending-approval → approved → ordered → partial-received → received
                         ↘ rejected (→ draft)
                         → cancelled
```

#### Supplier Quotes (RFQ Responses)

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/procurement/supplier-quotes` | List supplier quotes | `procurement:read` |
| `GET` | `/procurement/supplier-quotes/:id` | Get supplier quote with items | `procurement:read` |
| `POST` | `/procurement/supplier-quotes` | Create supplier quote | `procurement:create` |
| `PATCH` | `/procurement/supplier-quotes/:id` | Update supplier quote | `procurement:update` |
| `DELETE` | `/procurement/supplier-quotes/:id` | Delete supplier quote | `procurement:delete` |
| `PATCH` | `/procurement/supplier-quotes/:id/accept` | Accept supplier quote | `procurement:update` |
| `PATCH` | `/procurement/supplier-quotes/:id/reject` | Reject supplier quote | `procurement:update` |
| `POST` | `/procurement/supplier-quotes/:id/convert-to-po` | Convert accepted SQ to PO | `procurement:create` |
| `POST` | `/procurement/supplier-quotes/import` | Import supplier quote from CSV/XLSX | `procurement:create` |

#### Goods Receipts

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/procurement/goods-receipts` | List goods receipts | `procurement:read` |
| `GET` | `/procurement/goods-receipts/:id` | Get goods receipt with items | `procurement:read` |
| `POST` | `/procurement/goods-receipts` | Create goods receipt (against PO) | `procurement:create` |
| `PATCH` | `/procurement/goods-receipts/:id` | Update goods receipt | `procurement:update` |

**Side Effects on Goods Receipt Creation:**
1. Update `purchase_order_items.received_qty` for each received item
2. Update PO status: if all items fully received → `received`; partial → `partial-received`
3. Create `inventory_movements` (type: `receipt`) for each item
4. Update `warehouse_stock.on_hand_qty` for each item at the receiving warehouse
5. Create `product_price_records` with source `goods-receipt` for cost tracking

#### Supplier-Item Catalog

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/procurement/supplier-catalog` | List supplier-item relationships | `procurement:read` |
| `GET` | `/procurement/supplier-catalog/:id` | Get catalog entry detail | `procurement:read` |
| `POST` | `/procurement/supplier-catalog` | Create catalog entry | `procurement:create` |
| `PATCH` | `/procurement/supplier-catalog/:id` | Update catalog entry | `procurement:update` |
| `DELETE` | `/procurement/supplier-catalog/:id` | Delete catalog entry | `procurement:delete` |
| `POST` | `/procurement/supplier-catalog/import` | Bulk import from CSV/XLSX | `procurement:create` |
| `GET` | `/procurement/supplier-catalog/by-product/:productId` | Get all suppliers for a product | `procurement:read` |
| `GET` | `/procurement/supplier-catalog/by-supplier/:supplierId` | Get all products from a supplier | `procurement:read` |

---

### 7.15 Documents & File Upload

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/documents` | List documents | `documents:read` |
| `GET` | `/documents/:id` | Get document metadata | `documents:read` |
| `POST` | `/documents` | Upload document (multipart/form-data) | `documents:create` |
| `PATCH` | `/documents/:id` | Update document metadata | `documents:update` |
| `DELETE` | `/documents/:id` | Delete document (removes file from disk) | `documents:delete` |
| `GET` | `/documents/:id/download` | Download file (streamed from local disk) | `documents:read` |
| `POST` | `/documents/:id/link` | Link document to an entity | `documents:update` |
| `DELETE` | `/documents/:id/link/:linkId` | Unlink document from entity | `documents:update` |

**Upload Flow (Local Filesystem):**

```
Client                    Go API Server            Local Filesystem
  │                          │                        │
  ├── POST /documents ──────►│                        │
  │   (multipart form-data)  │                        │
  │                          ├── os.Create() ────────►│
  │                          │   io.Copy(dst, src)    │
  │                          │◄── File saved ─────────┤
  │                          ├── Save metadata to DB  │
  │◄── 201 { document } ────┤                        │
  │                          │                        │
  ├── GET /documents/:id ───►│                        │
  │     /download            │                        │
  │                          ├── os.Open() ──────────►│
  │                          │◄── io.Reader ──────────┤
  │◄── 200 (file stream) ───┤                        │
```

**Local File Storage Layout:**

```
/opt/g4s/storage/
├── documents/
│   ├── contracts/           # Contract-related files
│   ├── quotes/              # Quote-related files
│   ├── general/             # General documents
│   ├── compliance/          # Compliance documents
│   └── legal/               # Legal documents
├── products/
│   ├── datasheets/          # Product datasheets
│   ├── catalogs/            # Vendor catalogs
│   ├── certificates/        # Product certificates
│   └── images/              # Product images
├── exports/                 # Generated CSV/XLSX exports (temporary)
├── reports/                 # Generated PDF reports
└── temp/                    # Temporary upload staging
```

Files are stored with a UUID-based name to avoid collisions: `{category}/{uuid}_{original_filename}`. The original filename is preserved in the database metadata. Files are **never served directly by Nginx** — all access goes through the API with auth checks.

**POST `/documents`** — Multipart Form Data:

| Field | Type | Description |
|-------|------|-------------|
| `file` | File | The file to upload (max 50MB) |
| `name` | string | Display name |
| `category` | enum | Document category |
| `documentType` | enum | Document type |
| `tags` | string (JSON array) | Tags |
| `version` | string | Version label |
| `linkedEntities` | string (JSON array) | `[{ type, id, name }]` |

---

### 7.16 Dashboard & Analytics

| Method | Endpoint | Description | Permissions |
|--------|----------|-------------|-------------|
| `GET` | `/dashboard/kpis` | Get KPI summary | Any authenticated |
| `GET` | `/dashboard/pipeline` | Get opportunity pipeline breakdown | Any authenticated |
| `GET` | `/dashboard/recent-activity` | Get activity feed | Any authenticated |
| `GET` | `/dashboard/top-customers` | Get top customers by revenue | Any authenticated |
| `GET` | `/dashboard/alerts` | Get system alerts | Any authenticated |
| `GET` | `/dashboard/revenue-trend` | Monthly revenue trend | Any authenticated |
| `GET` | `/dashboard/quote-stats` | Quote conversion metrics | Any authenticated |

**GET `/dashboard/kpis`** — Response:
```json
{
  "success": true,
  "data": {
    "totalRevenue": {
      "value": 12500000,
      "change": 12.5,
      "period": "vs last quarter"
    },
    "activeOpportunities": {
      "value": 47,
      "totalValue": 8300000,
      "change": 8.2
    },
    "openQuotes": {
      "value": 23,
      "totalValue": 4200000,
      "conversionRate": 68.5
    },
    "recurringRevenue": {
      "value": 1850000,
      "change": 15.3,
      "period": "monthly"
    },
    "activeContracts": {
      "value": 34,
      "expiringIn30Days": 3
    }
  }
}
```

**GET `/dashboard/pipeline`** — Response:
```json
{
  "success": true,
  "data": {
    "stages": [
      { "stage": "qualification", "count": 15, "value": 2100000 },
      { "stage": "proposal", "count": 12, "value": 3400000 },
      { "stage": "negotiation", "count": 8, "value": 1800000 },
      { "stage": "closed-won", "count": 10, "value": 4200000 },
      { "stage": "closed-lost", "count": 5, "value": 950000 }
    ],
    "totalActive": 35,
    "totalValue": 7300000,
    "weightedValue": 4850000
  }
}
```

**GET `/dashboard/recent-activity`** — Query Parameters:

| Param | Type | Description |
|-------|------|-------------|
| `limit` | integer | Number of items (default 20, max 50) |
| `types` | string | Comma-separated entity types to filter |

**GET `/dashboard/alerts`** — Response:
```json
{
  "success": true,
  "data": [
    {
      "type": "contract_expiring",
      "severity": "warning",
      "message": "3 contracts expiring in the next 30 days",
      "count": 3,
      "link": "/contracts?status=active&expiringDays=30"
    },
    {
      "type": "low_stock",
      "severity": "info",
      "message": "5 products below reorder level",
      "count": 5,
      "link": "/inventory?filter=low-stock"
    },
    {
      "type": "pending_approvals",
      "severity": "action",
      "message": "4 quotes awaiting approval",
      "count": 4,
      "link": "/quotes?status=pending-approval"
    }
  ]
}
```

---

## 8. Business Logic & Workflows

### 8.1 Quote-to-Project Workflow

```
1. Sales creates Opportunity (stage: qualification)
2. Pre-sales builds Quote (status: draft)
3. Quote submitted → pending-approval
4. Sales Manager reviews → approved / rejected
5. Approved quote sent to customer → sent
6. Customer accepts → accepted
   └── Auto-trigger: Create Project from accepted Quote
   └── Auto-trigger: Option to auto-generate POs for materials
7. Customer declines → declined
   └── Option to duplicate quote with new version
```

### 8.2 Procurement Workflow

```
1. PO created (draft) — manually or auto-generated from quote
2. PO submitted for approval → pending-approval
3. Procurement Manager approves → approved
4. PO sent to supplier → ordered
5. Goods arrive → Create Goods Receipt
   └── Partial delivery → PO status: partial-received
   └── Full delivery → PO status: received
   └── Side effects: stock updated, price history recorded
```

### 8.3 Supplier Quote → PO Conversion

```
1. Receive Supplier Quote (from vendor RFQ response)
2. Review line items, costs, lead times
3. Accept supplier quote
4. Convert to PO → auto-populates items from SQ
5. New PO enters approval workflow
```

### 8.4 Inventory Reservation Flow

```
1. Quote created with materials → optional stock reservation
2. Reservation holds stock (reserved_qty incremented, available_qty decremented)
3. Quote accepted → reservation status: fulfilled (stock allocated to project)
4. Quote declined/expired → reservation released (stock returned to available)
```

### 8.5 Contract Renewal Automation

```
Scheduled job (daily):
1. Query contracts WHERE end_date - NOW() <= renewal_notice_days AND auto_renew = true
2. For each:
   a. Send notification to sales exec and customer contact
   b. If auto_renew enabled, create renewal contract draft
   c. Log activity
```

### 8.6 Exchange Rate Impact

```
When exchange rate updated:
1. Archive old rate to history
2. Optionally trigger async job:
   a. Find all products using that currency
   b. Recalculate cost_in_sar, landed_cost_sar, margin for each
   c. Log all changes to product_price_records
   d. Notify affected users
```

### 8.7 Auto-Number Generation

| Entity | Pattern | Example |
|--------|---------|---------|
| Quote | `QT-{YYYY}-{SEQ:4}` | QT-2026-0042 |
| Purchase Order | `PO-{YYYY}-{SEQ:4}` | PO-2026-0015 |
| Supplier Quote | `SQ-{YYYY}-{SEQ:4}` | SQ-2026-0008 |
| Goods Receipt | `GR-{YYYY}-{SEQ:4}` | GR-2026-0003 |
| Contract | `CNT-{YYYY}-{SEQ:4}` | CNT-2026-0012 |
| Project | `PRJ-{YYYY}-{SEQ:4}` | PRJ-2026-0007 |

Sequence counters stored in a `sequences` table, incremented atomically.

---

## 9. Error Handling

### 9.1 Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "companyName",
        "message": "Company name is required",
        "code": "required"
      },
      {
        "field": "crNumber",
        "message": "CR number already exists",
        "code": "unique"
      }
    ]
  }
}
```

### 9.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request body/params failed validation |
| `UNAUTHORIZED` | 401 | Missing or invalid auth token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate resource (unique constraint) |
| `BUSINESS_RULE_VIOLATION` | 422 | Business logic prevented the operation |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

### 9.3 Business Rule Violation Examples

```json
{
  "success": false,
  "error": {
    "code": "BUSINESS_RULE_VIOLATION",
    "message": "Cannot delete quote in 'approved' status. Only draft quotes can be deleted."
  }
}
```

```json
{
  "success": false,
  "error": {
    "code": "BUSINESS_RULE_VIOLATION",
    "message": "Insufficient stock. Available: 8, Requested: 15. Warehouse: riyadh-main"
  }
}
```

---

## 10. Security Considerations

### 10.1 Authentication

- JWT signed with RS256 (asymmetric keys)
- Access tokens: short-lived (15 min)
- Refresh tokens: single-use rotation, stored hashed in DB
- Password requirements: min 8 chars, at least 1 uppercase, 1 number, 1 special char
- Account lockout after 5 failed login attempts (15-minute cooldown)
- All passwords hashed with bcrypt (cost factor 12)

### 10.2 API Security

- HTTPS enforced via Nginx SSL termination (Let's Encrypt or client-provided certificate)
- CORS configured via Gin CORS middleware to allow only the frontend origin
- Secure HTTP headers via custom Gin middleware (X-Content-Type-Options, X-Frame-Options, etc.)
- Rate limiting: 100 req/min per user, 1000 req/min per IP (via go-redis rate limiter or in-memory token bucket)
- Request size limit: 10MB default (50MB for file upload endpoints, configured per-route in Gin)
- SQL injection prevention via GORM parameterized queries (never raw string concatenation)
- Input validation on all endpoints via go-playground/validator struct tags

### 10.3 Data Security

- Sensitive fields (passwords) never returned in API responses (use `json:"-"` struct tags)
- Audit log for all CUD operations (activity_log table)
- File storage directory not exposed by Nginx — all file access routed through authenticated API endpoints
- File permissions: storage directory owned by the API process user only (chmod 700)
- Soft deletes for critical entities (customers, products, users)
- Database connection via SSL in production (sslmode=require in connection string)
- Uploaded files validated: MIME type check, file extension whitelist, max size enforcement

### 10.4 Saudi Arabia Compliance

- VAT at 15% (configurable per quote, default enforced)
- CR (Commercial Registration) number validation format
- VAT number validation format
- Data residency: all data stored on-premise on the client's local server (full data sovereignty)
- Arabic character support (UTF-8 throughout)

---

## 11. Deployment & Infrastructure

### 11.1 Environment Configuration

```env
# Server
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=g4s_crm
DB_USER=g4s_app
DB_PASSWORD=<secure-password>
DB_SSLMODE=require

# Redis (optional)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_PRIVATE_KEY_PATH=/opt/g4s/keys/private.pem
JWT_PUBLIC_KEY_PATH=/opt/g4s/keys/public.pem
JWT_REFRESH_SECRET=<secure-random-string>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# File Storage
STORAGE_ROOT=/opt/g4s/storage
MAX_UPLOAD_SIZE=52428800

# Email (for notifications)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=crm@g4s.com
SMTP_PASS=<password>
SMTP_FROM=G4S CRM <crm@g4s.com>

# Logging
LOG_LEVEL=info
LOG_FILE=/var/log/g4s/crm-api.log
```

### 11.2 Docker Setup (Development)

```yaml
# docker-compose.yml — local development
services:
  api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./storage:/opt/g4s/storage
      - ./.env:/app/.env
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: g4s_crm
      POSTGRES_USER: g4s_app
      POSTGRES_PASSWORD: devpassword
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  pgdata:
```

```dockerfile
# Dockerfile — multi-stage build producing ~15 MB binary
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o g4s-crm-api ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/g4s-crm-api .
COPY --from=builder /build/migrations ./migrations
EXPOSE 8080
CMD ["./g4s-crm-api"]
```

### 11.3 On-Premise Production Deployment

**Server Requirements:**

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| OS | Ubuntu 22.04 LTS / RHEL 9 | Ubuntu 24.04 LTS |
| CPU | 2 cores | 4 cores |
| RAM | 4 GB | 8 GB |
| Disk | 50 GB SSD | 200 GB SSD (depends on document volume) |
| Network | LAN access | LAN + optional VPN for remote access |

**Services on the Server:**

```
┌──────────────────────────────────────────────────────────┐
│                 Client Server (On-Premise)                │
│                                                          │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  systemd services                                    │ │
│  │                                                      │ │
│  │  ┌─────────────┐  ┌────────────┐  ┌──────────────┐  │ │
│  │  │   nginx     │  │ g4s-crm-   │  │  postgresql  │  │ │
│  │  │  (port 443) │  │ api        │  │  (port 5432) │  │ │
│  │  │  SSL +      │──│ (port 8080)│──│              │  │ │
│  │  │  SPA files  │  │            │  │              │  │ │
│  │  └─────────────┘  └─────┬──────┘  └──────────────┘  │ │
│  │                         │                            │ │
│  │                   ┌─────┴──────┐  ┌──────────────┐   │ │
│  │                   │ /opt/g4s/  │  │   redis      │   │ │
│  │                   │  storage/  │  │  (port 6379) │   │ │
│  │                   │            │  │  (optional)  │   │ │
│  │                   └────────────┘  └──────────────┘   │ │
│  └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

**Nginx Configuration:**

```nginx
server {
    listen 443 ssl http2;
    server_name crm.g4s-client.local;

    ssl_certificate     /etc/nginx/ssl/g4s-crm.crt;
    ssl_certificate_key /etc/nginx/ssl/g4s-crm.key;

    # Vue SPA — serve static files
    root /opt/g4s/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # API reverse proxy
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 50m;
    }
}
```

**systemd Service File:**

```ini
# /etc/systemd/system/g4s-crm-api.service
[Unit]
Description=G4S CRM API Server
After=network.target postgresql.service

[Service]
Type=simple
User=g4s
Group=g4s
WorkingDirectory=/opt/g4s
ExecStart=/opt/g4s/g4s-crm-api
EnvironmentFile=/opt/g4s/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Deployment Script:**

```bash
#!/bin/bash
# deploy.sh — run on the client server

set -e

echo "Stopping API service..."
sudo systemctl stop g4s-crm-api

echo "Backing up current binary..."
cp /opt/g4s/g4s-crm-api /opt/g4s/g4s-crm-api.bak

echo "Deploying new binary..."
cp ./g4s-crm-api /opt/g4s/g4s-crm-api
chmod +x /opt/g4s/g4s-crm-api

echo "Running database migrations..."
/opt/g4s/g4s-crm-api migrate

echo "Deploying frontend..."
rm -rf /opt/g4s/frontend/dist
cp -r ./frontend-dist /opt/g4s/frontend/dist

echo "Starting API service..."
sudo systemctl start g4s-crm-api

echo "Deployment complete."
```

### 11.4 Backup Strategy

| What | Method | Frequency |
|------|--------|-----------|
| PostgreSQL database | `pg_dump` to compressed file | Daily (keep 30 days) |
| File storage (`/opt/g4s/storage/`) | `rsync` to backup disk/NAS | Daily (incremental) |
| Application binary + config | Version-controlled releases | Per deployment |
| JWT keys | Encrypted backup to secure location | On key rotation |

```bash
# /opt/g4s/scripts/backup.sh — daily cron job
#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/opt/g4s/backups

pg_dump -U g4s_app -h localhost g4s_crm | gzip > "$BACKUP_DIR/db_$TIMESTAMP.sql.gz"
rsync -a /opt/g4s/storage/ "$BACKUP_DIR/storage_$TIMESTAMP/"

# Keep only last 30 days
find "$BACKUP_DIR" -name "db_*.sql.gz" -mtime +30 -delete
```

---

## 12. Development Phases

### Phase 1 — Foundation (Weeks 1–3)

| Task | Priority |
|------|----------|
| Go project scaffolding (Gin + GORM + project structure) | Critical |
| Database schema + golang-migrate migrations | Critical |
| Config management (viper + .env) | Critical |
| Auth module (register, login, refresh, JWT RS256, RBAC middleware) | Critical |
| User management CRUD handlers | Critical |
| Standard error handling + validation framework (validator v10) | Critical |
| Structured logging (zerolog) + health check endpoint | High |
| Docker + docker-compose for local dev (Go + PostgreSQL + Redis) | High |
| Makefile for build, run, migrate, seed commands | High |

### Phase 2 — Core CRM (Weeks 4–6)

| Task | Priority |
|------|----------|
| Customers module (CRUD + sites + contacts) | Critical |
| Opportunities module (CRUD + pipeline + stage transitions) | Critical |
| Manufacturers module (CRUD + categories) | High |
| Products module (CRUD + costing + vendor entries + price history) | Critical |
| Exchange rates module (CRUD + history + product recalculation) | High |
| Teams module (CRUD + member management) | Medium |
| Activity logging service (middleware + repository) | High |
| Pagination + filtering + search utilities | High |
| Frontend integration: wire Pinia stores to real APIs | Critical |

### Phase 3 — Quotation & Commercial (Weeks 7–9)

| Task | Priority |
|------|----------|
| Quotes module (CRUD + line items + recalculation + approval workflow) | Critical |
| Quote PDF generation (goroutine worker) | High |
| Price books module (CRUD + entries + lookup logic) | High |
| Contracts module (CRUD + status transitions + renewal) | High |
| Recurring services module (CRUD) | Medium |
| Quote-to-project auto-creation | High |
| Auto-number generation service (sequences table) | High |

### Phase 4 — Procurement & Inventory (Weeks 10–12)

| Task | Priority |
|------|----------|
| Purchase orders module (CRUD + approval workflow) | Critical |
| Supplier quotes module (CRUD + accept/reject + convert-to-PO) | High |
| Goods receipts module (CRUD + stock update side effects in DB transactions) | Critical |
| Supplier-item catalog (CRUD + import) | High |
| Inventory stock module (read + update + low-stock alerts) | Critical |
| Stock reservations (create + release + fulfill) | High |
| Inventory movements (transfer + adjustment + write-off) | High |
| CSV/XLSX import via excelize for supplier catalog + POs | Medium |

### Phase 5 — Documents, Dashboard & Polish (Weeks 13–15)

| Task | Priority |
|------|----------|
| Local file storage abstraction layer | High |
| Documents module (upload to local disk + metadata + entity linking) | High |
| Dashboard KPIs (aggregation queries) | High |
| Dashboard pipeline + activity feed + alerts | High |
| Contract expiration cron job (robfig/cron) | Medium |
| Export functionality (CSV/XLSX via excelize for all major entities) | Medium |
| Global search API | Medium |
| Swagger documentation (swaggo/swag) | High |
| Integration testing suite (httptest + testify) | High |

### Phase 6 — Hardening & On-Premise Deployment (Weeks 16–18)

| Task | Priority |
|------|----------|
| Performance optimization (query tuning, GORM preloading, indexes) | High |
| Rate limiting + security audit | Critical |
| Load testing with production-like data volume | High |
| Build pipeline (Makefile → binary + frontend dist → deploy script) | High |
| Nginx configuration (SSL, reverse proxy, SPA static serving) | High |
| systemd service setup + log rotation | High |
| Backup scripts (pg_dump + rsync) | High |
| On-premise deployment + server provisioning documentation | Critical |
| User acceptance testing (UAT) with stakeholders | Critical |
| Bug fixes and refinements | Ongoing |

---

## Appendix A — Enum Reference

All enum values used across the API, matching the frontend TypeScript types:

```
Sector:              government, healthcare, education, retail, banking, oil-gas, telecom, hospitality, real-estate, other
CustomerStatus:      active, inactive, prospect
CustomerType:        get, grow
OpportunityStage:    qualification, proposal, negotiation, closed-won, closed-lost
ServiceType:         cctv, access-control, intrusion-detection, fire-alarm, networking, it-solutions, guarding, monitoring, maintenance, consulting
VendorType:          manufacturer, supplier, both
ProductType:         import, local
Currency:            SAR, USD, EUR, GBP, AED, CNY
PriceSource:         vendor-catalog, purchase-order, supplier-quote, goods-receipt, manual
ProductDocType:      datasheet, manual, certificate, vendor-quote, catalog, image, other
WarehouseLocation:   riyadh-main, jeddah-branch, dammam-branch
QuoteStatus:         draft, pending-approval, approved, sent, accepted, declined, expired
QuoteLineCategory:   materials, manpower, miscellaneous
ProjectStatus:       planning, in-progress, on-hold, completed, cancelled
ProjectPriority:     low, medium, high, critical
PriceBookType:       standard, volume, contract, promotional, customer-specific
ContractType:        sales, maintenance, service, project, subscription
ContractStatus:      draft, pending-approval, active, expired, terminated, renewed
RecurringServiceType: guarding, maintenance, monitoring, patrol, facility-management
BillingFrequency:    monthly, quarterly, annually
PurchaseOrderStatus: draft, pending-approval, approved, ordered, partial-received, received, cancelled
SupplierQuoteStatus: received, under-review, accepted, expired, rejected
ReservationSource:   quote, project, manual
ReservationStatus:   active, released, fulfilled
MovementType:        transfer, adjustment, allocation, receipt, release, write-off
Department:          sales, pre-sales, technical, support, marketing, management, operations
DocumentCategory:    contract, quote, general, compliance, legal
DocumentType:        terms, delivery, technical, warranty, sla
GoodsCondition:      good, damaged, partial-damage
CostTrend:           up, down, stable
UserRole:            admin, sales_manager, sales_executive, pre_sales, procurement_manager, procurement_officer, warehouse_manager, project_manager, viewer
```

## Appendix B — Number Generation Sequences Table

```sql
CREATE TABLE sequences (
  name       VARCHAR(50) PRIMARY KEY,
  prefix     VARCHAR(10) NOT NULL,
  year       INTEGER NOT NULL,
  current    INTEGER NOT NULL DEFAULT 0
);

-- Seed data
INSERT INTO sequences (name, prefix, year, current) VALUES
  ('quote',          'QT',  2026, 0),
  ('purchase_order', 'PO',  2026, 0),
  ('supplier_quote', 'SQ',  2026, 0),
  ('goods_receipt',  'GR',  2026, 0),
  ('contract',       'CNT', 2026, 0),
  ('project',        'PRJ', 2026, 0);

-- Atomic increment function
CREATE OR REPLACE FUNCTION next_number(seq_name VARCHAR)
RETURNS VARCHAR AS $$
DECLARE
  seq RECORD;
  next_val INTEGER;
BEGIN
  UPDATE sequences SET current = current + 1
  WHERE name = seq_name AND year = EXTRACT(YEAR FROM NOW())
  RETURNING prefix, year, current INTO seq;

  IF NOT FOUND THEN
    INSERT INTO sequences (name, prefix, year, current)
    SELECT name, prefix, EXTRACT(YEAR FROM NOW())::INTEGER, 1
    FROM sequences WHERE name = seq_name
    ORDER BY year DESC LIMIT 1
    RETURNING prefix, year, current INTO seq;
  END IF;

  RETURN seq.prefix || '-' || seq.year || '-' || LPAD(seq.current::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;
```
