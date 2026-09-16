-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── ENUMS ─────────────────────────────────────────────────

CREATE TYPE user_role AS ENUM ('admin','sales_manager','sales_executive','pre_sales','procurement_manager','procurement_officer','warehouse_manager','project_manager','viewer');
CREATE TYPE department AS ENUM ('sales','pre-sales','technical','support','marketing','management','operations');
CREATE TYPE customer_status AS ENUM ('active','inactive','prospect');
CREATE TYPE customer_type AS ENUM ('get','grow');
CREATE TYPE sector AS ENUM ('government','healthcare','education','retail','banking','oil-gas','telecom','hospitality','real-estate','other');
CREATE TYPE opportunity_stage AS ENUM ('qualification','proposal','negotiation','closed-won','closed-lost');
CREATE TYPE vendor_type AS ENUM ('manufacturer','supplier','both');
CREATE TYPE product_type AS ENUM ('import','local');
CREATE TYPE currency AS ENUM ('SAR','USD','EUR','GBP','AED','CNY');
CREATE TYPE price_source AS ENUM ('vendor-catalog','purchase-order','supplier-quote','goods-receipt','manual');
CREATE TYPE product_doc_type AS ENUM ('datasheet','manual','certificate','vendor-quote','catalog','image','other');
CREATE TYPE warehouse_location AS ENUM ('riyadh-main','jeddah-branch','dammam-branch');
CREATE TYPE quote_status AS ENUM ('draft','pending-approval','approved','sent','accepted','declined','expired');
CREATE TYPE quote_line_category AS ENUM ('materials','manpower','miscellaneous');
CREATE TYPE project_status AS ENUM ('planning','in-progress','on-hold','completed','cancelled');
CREATE TYPE project_priority AS ENUM ('low','medium','high','critical');
CREATE TYPE price_book_type AS ENUM ('standard','volume','contract','promotional','customer-specific');
CREATE TYPE contract_type AS ENUM ('sales','maintenance','service','project','subscription');
CREATE TYPE contract_status AS ENUM ('draft','pending-approval','active','expired','terminated','renewed');
CREATE TYPE recurring_service_type AS ENUM ('guarding','maintenance','monitoring','patrol','facility-management');
CREATE TYPE billing_frequency AS ENUM ('monthly','quarterly','annually');
CREATE TYPE po_status AS ENUM ('draft','pending-approval','approved','ordered','partial-received','received','cancelled');
CREATE TYPE sq_status AS ENUM ('received','under-review','accepted','expired','rejected');
CREATE TYPE reservation_source AS ENUM ('quote','project','manual');
CREATE TYPE reservation_status AS ENUM ('active','released','fulfilled');
CREATE TYPE movement_type AS ENUM ('transfer','adjustment','allocation','receipt','release','write-off');
CREATE TYPE document_category AS ENUM ('contract','quote','general','compliance','legal');
CREATE TYPE document_type AS ENUM ('terms','delivery','technical','warranty','sla');
CREATE TYPE goods_condition AS ENUM ('good','damaged','partial-damage');

-- ─── TEAMS (before users because users FK → teams) ─────────

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    department department NOT NULL,
    description TEXT,
    leader_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── USERS ─────────────────────────────────────────────────

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    role user_role NOT NULL DEFAULT 'viewer',
    department department,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE teams ADD CONSTRAINT fk_teams_leader FOREIGN KEY (leader_id) REFERENCES users(id) ON DELETE SET NULL;

-- ─── REFRESH TOKENS ────────────────────────────────────────

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);

-- ─── CUSTOMERS ─────────────────────────────────────────────

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    sector sector NOT NULL,
    region VARCHAR(100) NOT NULL,
    status customer_status NOT NULL DEFAULT 'prospect',
    type customer_type NOT NULL DEFAULT 'get',
    cr_number VARCHAR(50) UNIQUE,
    vat_number VARCHAR(50),
    notes TEXT,
    created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_customers_sector ON customers(sector);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customers_company_name ON customers(company_name);

CREATE TABLE customer_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    region VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE customer_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    site_id UUID REFERENCES customer_sites(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    position VARCHAR(100),
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── MANUFACTURERS ─────────────────────────────────────────

CREATE TABLE manufacturers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,
    country VARCHAR(100),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    website VARCHAR(500),
    vendor_type vendor_type,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE manufacturer_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manufacturer_id UUID NOT NULL REFERENCES manufacturers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── PRODUCTS ──────────────────────────────────────────────

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    manufacturer_id UUID REFERENCES manufacturers(id) ON DELETE SET NULL,
    category_id UUID REFERENCES manufacturer_categories(id) ON DELETE SET NULL,
    product_type product_type NOT NULL,
    origin_currency currency NOT NULL DEFAULT 'USD',
    unit_cost_origin NUMERIC(15,4) DEFAULT 0,
    fx_rate NUMERIC(12,6) DEFAULT 1,
    cost_in_sar NUMERIC(15,4) DEFAULT 0,
    freight_percent NUMERIC(5,2) DEFAULT 0,
    customs_percent NUMERIC(5,2) DEFAULT 0,
    clearance_percent NUMERIC(5,2) DEFAULT 0,
    landed_cost_sar NUMERIC(15,4) DEFAULT 0,
    target_margin_percent NUMERIC(5,2) DEFAULT 0,
    selling_price NUMERIC(15,4) DEFAULT 0,
    margin_amount NUMERIC(15,4) DEFAULT 0,
    lead_time_days INTEGER DEFAULT 0,
    supplier_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_products_manufacturer ON products(manufacturer_id);
CREATE INDEX idx_products_sku ON products(sku);

CREATE TABLE product_vendor_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_sku VARCHAR(50),
    unit_cost NUMERIC(15,4),
    currency currency,
    moq INTEGER DEFAULT 1,
    lead_time_days INTEGER,
    last_quote_date DATE,
    catalog_source VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE product_price_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    source price_source,
    source_ref VARCHAR(100),
    vendor_name VARCHAR(255),
    unit_cost NUMERIC(15,4),
    currency currency,
    landing_cost NUMERIC(15,4),
    qty INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE product_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    doc_type product_doc_type NOT NULL,
    file_name VARCHAR(255),
    file_size BIGINT,
    file_path VARCHAR(500),
    uploaded_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── OPPORTUNITIES ─────────────────────────────────────────

CREATE TABLE opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    stage opportunity_stage NOT NULL DEFAULT 'qualification',
    service_types TEXT[],
    estimated_value NUMERIC(15,2) DEFAULT 0,
    estimated_cost NUMERIC(15,2) DEFAULT 0,
    estimated_margin NUMERIC(15,2) DEFAULT 0,
    win_probability INTEGER DEFAULT 0 CHECK (win_probability >= 0 AND win_probability <= 100),
    sales_executive_id UUID REFERENCES users(id) ON DELETE SET NULL,
    pre_sales_id UUID REFERENCES users(id) ON DELETE SET NULL,
    expected_close_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_opportunities_customer ON opportunities(customer_id);
CREATE INDEX idx_opportunities_stage ON opportunities(stage);
CREATE INDEX idx_opportunities_sales_exec ON opportunities(sales_executive_id);

-- ─── QUOTES ────────────────────────────────────────────────

CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    opportunity_id UUID REFERENCES opportunities(id) ON DELETE SET NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    version INTEGER DEFAULT 1,
    status quote_status DEFAULT 'draft',
    subtotal NUMERIC(15,2) DEFAULT 0,
    discount_percent NUMERIC(5,2) DEFAULT 0,
    discount_amount NUMERIC(15,2) DEFAULT 0,
    subtotal_after_discount NUMERIC(15,2) DEFAULT 0,
    vat_percent NUMERIC(5,2) DEFAULT 15,
    vat_amount NUMERIC(15,2) DEFAULT 0,
    total NUMERIC(15,2) DEFAULT 0,
    total_cost NUMERIC(15,2) DEFAULT 0,
    margin_amount NUMERIC(15,2) DEFAULT 0,
    margin_percent NUMERIC(5,2) DEFAULT 0,
    valid_until DATE,
    currency currency DEFAULT 'SAR',
    notes TEXT,
    approved_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_quotes_customer ON quotes(customer_id);
CREATE INDEX idx_quotes_opportunity ON quotes(opportunity_id);
CREATE INDEX idx_quotes_status ON quotes(status);

CREATE TABLE quote_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    category quote_line_category NOT NULL,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    sku VARCHAR(50),
    description TEXT NOT NULL,
    manufacturer_name VARCHAR(255),
    stock_available INTEGER,
    lead_time_days INTEGER,
    quantity INTEGER NOT NULL,
    unit_cost NUMERIC(15,4) DEFAULT 0,
    unit_price NUMERIC(15,4) DEFAULT 0,
    line_total NUMERIC(15,2) DEFAULT 0,
    margin_percent NUMERIC(5,2) DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── PROJECTS ──────────────────────────────────────────────

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_number VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    quote_id UUID REFERENCES quotes(id) ON DELETE SET NULL,
    status project_status DEFAULT 'planning',
    priority project_priority DEFAULT 'medium',
    start_date DATE,
    target_end_date DATE,
    actual_end_date DATE,
    project_manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
    total_value NUMERIC(15,2) DEFAULT 0,
    total_cost NUMERIC(15,2) DEFAULT 0,
    margin_percent NUMERIC(5,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── PRICE BOOKS ───────────────────────────────────────────

CREATE TABLE price_books (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type price_book_type NOT NULL,
    description TEXT,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    contract_id UUID,
    valid_from DATE,
    valid_to DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE price_book_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    price_book_id UUID NOT NULL REFERENCES price_books(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    standard_price NUMERIC(15,4) DEFAULT 0,
    custom_price NUMERIC(15,4) DEFAULT 0,
    discount_percent NUMERIC(5,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── EXCHANGE RATES ────────────────────────────────────────

CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency currency NOT NULL,
    to_currency VARCHAR(3) NOT NULL DEFAULT 'SAR',
    current_rate NUMERIC(12,6) NOT NULL,
    effective_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(from_currency, to_currency)
);

CREATE TABLE exchange_rate_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exchange_rate_id UUID NOT NULL REFERENCES exchange_rates(id) ON DELETE CASCADE,
    rate NUMERIC(12,6) NOT NULL,
    effective_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── CONTRACTS ─────────────────────────────────────────────

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_number VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    type contract_type NOT NULL,
    status contract_status DEFAULT 'draft',
    start_date DATE,
    end_date DATE,
    value NUMERIC(15,2) DEFAULT 0,
    auto_renew BOOLEAN DEFAULT FALSE,
    renewal_notice_days INTEGER DEFAULT 30,
    renewed_from_id UUID REFERENCES contracts(id) ON DELETE SET NULL,
    quote_id UUID REFERENCES quotes(id) ON DELETE SET NULL,
    terms TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_contracts_customer ON contracts(customer_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_end_date ON contracts(end_date);

-- ─── RECURRING SERVICES ────────────────────────────────────

CREATE TABLE recurring_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    service_type recurring_service_type,
    description TEXT,
    monthly_cost NUMERIC(15,2) DEFAULT 0,
    monthly_price NUMERIC(15,2) DEFAULT 0,
    annual_cost NUMERIC(15,2) DEFAULT 0,
    annual_price NUMERIC(15,2) DEFAULT 0,
    target_margin_percent NUMERIC(5,2) DEFAULT 0,
    billing_frequency billing_frequency,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── PURCHASE ORDERS ───────────────────────────────────────

CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    po_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id UUID REFERENCES manufacturers(id) ON DELETE SET NULL,
    supplier_name VARCHAR(255) NOT NULL,
    status po_status DEFAULT 'draft',
    subtotal NUMERIC(15,2) DEFAULT 0,
    shipping_cost NUMERIC(15,2) DEFAULT 0,
    customs_duty NUMERIC(15,2) DEFAULT 0,
    total NUMERIC(15,2) DEFAULT 0,
    currency currency DEFAULT 'SAR',
    expected_delivery DATE,
    actual_delivery DATE,
    source_quote_id UUID REFERENCES quotes(id) ON DELETE SET NULL,
    notes TEXT,
    approved_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_po_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_po_status ON purchase_orders(status);

CREATE TABLE purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL,
    unit_cost NUMERIC(15,4) DEFAULT 0,
    total NUMERIC(15,2) DEFAULT 0,
    received_qty INTEGER DEFAULT 0,
    lead_time_days INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── SUPPLIER QUOTES ───────────────────────────────────────

CREATE TABLE supplier_quotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sq_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id UUID REFERENCES manufacturers(id) ON DELETE SET NULL,
    supplier_name VARCHAR(255) NOT NULL,
    supplier_ref VARCHAR(100),
    status sq_status DEFAULT 'received',
    subtotal NUMERIC(15,2) DEFAULT 0,
    currency currency DEFAULT 'USD',
    valid_from DATE,
    valid_until DATE,
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    payment_terms VARCHAR(255),
    delivery_terms VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE supplier_quote_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_quote_id UUID NOT NULL REFERENCES supplier_quotes(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    product_sku VARCHAR(50),
    product_name VARCHAR(255),
    manufacturer_name VARCHAR(255),
    quantity INTEGER DEFAULT 1,
    unit_cost NUMERIC(15,4) DEFAULT 0,
    total NUMERIC(15,2) DEFAULT 0,
    lead_time_days INTEGER DEFAULT 0,
    moq INTEGER DEFAULT 1,
    valid_until DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── SUPPLIER ITEM CATALOG ─────────────────────────────────

CREATE TABLE supplier_item_catalogs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES manufacturers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    latest_cost NUMERIC(15,4) DEFAULT 0,
    previous_cost NUMERIC(15,4),
    cost_trend VARCHAR(10) DEFAULT 'stable',
    moq INTEGER DEFAULT 1,
    lead_time_days INTEGER DEFAULT 0,
    last_quote_date DATE,
    last_po_date DATE,
    reliability NUMERIC(3,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(supplier_id, product_id)
);

-- ─── GOODS RECEIPTS ────────────────────────────────────────

CREATE TABLE goods_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    gr_number VARCHAR(50) UNIQUE NOT NULL,
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE RESTRICT,
    supplier_name VARCHAR(255),
    receive_date DATE NOT NULL,
    total_items INTEGER DEFAULT 0,
    total_landing_cost NUMERIC(15,2) DEFAULT 0,
    received_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE goods_receipt_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goods_receipt_id UUID NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    po_item_id UUID NOT NULL REFERENCES purchase_order_items(id),
    product_id UUID NOT NULL REFERENCES products(id),
    received_qty INTEGER NOT NULL,
    unit_cost NUMERIC(15,4) DEFAULT 0,
    shipping_alloc NUMERIC(15,4) DEFAULT 0,
    customs_alloc NUMERIC(15,4) DEFAULT 0,
    landing_cost NUMERIC(15,4) DEFAULT 0,
    storage_location VARCHAR(100),
    condition goods_condition DEFAULT 'good',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── WAREHOUSE STOCK ───────────────────────────────────────

CREATE TABLE warehouse_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    warehouse_location warehouse_location NOT NULL,
    on_hand_qty INTEGER DEFAULT 0 CHECK (on_hand_qty >= 0),
    reserved_qty INTEGER DEFAULT 0 CHECK (reserved_qty >= 0),
    available_qty INTEGER DEFAULT 0,
    unit_cost NUMERIC(15,4) DEFAULT 0,
    total_value NUMERIC(15,4) DEFAULT 0,
    reorder_level INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(product_id, warehouse_location)
);
CREATE INDEX idx_stock_product ON warehouse_stocks(product_id);
CREATE INDEX idx_stock_location ON warehouse_stocks(warehouse_location);

CREATE TABLE stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    warehouse_location warehouse_location NOT NULL,
    qty INTEGER NOT NULL,
    source reservation_source NOT NULL,
    source_ref VARCHAR(100),
    source_label VARCHAR(255),
    customer_name VARCHAR(255),
    reserved_by_id UUID REFERENCES users(id),
    reserved_at TIMESTAMPTZ DEFAULT NOW(),
    release_date TIMESTAMPTZ,
    released_by_id UUID REFERENCES users(id),
    release_reason TEXT,
    status reservation_status DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    movement_type movement_type NOT NULL,
    qty INTEGER NOT NULL,
    from_warehouse warehouse_location,
    to_warehouse warehouse_location,
    reference VARCHAR(100),
    reason TEXT,
    performed_by_id UUID REFERENCES users(id),
    performed_at TIMESTAMPTZ DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── DOCUMENTS ─────────────────────────────────────────────

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category document_category,
    document_type document_type,
    tags TEXT[],
    version VARCHAR(20),
    file_name VARCHAR(255),
    file_size BIGINT,
    file_type VARCHAR(50),
    file_path VARCHAR(500),
    uploaded_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE document_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    entity_name VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── ACTIVITY LOG ──────────────────────────────────────────

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_activity_entity ON activity_logs(entity_type, entity_id);
CREATE INDEX idx_activity_user ON activity_logs(user_id);
CREATE INDEX idx_activity_created ON activity_logs(created_at DESC);

-- ─── SEQUENCES ─────────────────────────────────────────────

CREATE TABLE sequences (
    name VARCHAR(50) PRIMARY KEY,
    prefix VARCHAR(10) NOT NULL,
    year INTEGER NOT NULL,
    current INTEGER NOT NULL DEFAULT 0
);

INSERT INTO sequences (name, prefix, year, current) VALUES
    ('quote',          'QT',  2026, 0),
    ('purchase_order', 'PO',  2026, 0),
    ('supplier_quote', 'SQ',  2026, 0),
    ('goods_receipt',  'GR',  2026, 0),
    ('contract',       'CNT', 2026, 0),
    ('project',        'PRJ', 2026, 0);
