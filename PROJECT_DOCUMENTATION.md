# CRM Dashboard - Project Documentation

> Historical design document. For the verified implementation state and current work plan, read [STATUS.md](STATUS.md) and the [September 2026 review](docs/CRM_REVIEW.md). Some features and file paths below describe earlier designs rather than working functionality.

## Project Overview

A comprehensive Customer Relationship Management (CRM) system designed for security systems and IT solutions businesses. The application manages the entire sales lifecycle from lead generation to quote creation, contract management, and inventory tracking.

## Technology Stack

- **Frontend Framework**: Vue 3 with TypeScript
- **State Management**: Pinia
- **Routing**: Vue Router
- **Build Tool**: Vite
- **Styling**: Custom CSS with CSS variables for theming

## Core Modules

### 1. Customer Management

**Purpose**: Manage customer companies, sites, and contacts.

**Features**:
- Company profile management (name, sector, regions, status)
- Multiple site locations per company
- Contact management with primary contact designation
- Document attachments and legal information (CR Number, VAT Number)
- Customer categorization: "Get" (new customers) vs "Grow" (existing customers)

**Key Files**:
- [CustomersView.vue](src/views/CustomersView.vue)
- [customers.ts](src/stores/customers.ts)

**Workflow**:
1. Add new company with basic information
2. Add multiple site locations
3. Add contacts for each site or company level
4. Attach relevant documents
5. Track customer status and interactions

---

### 2. Opportunities & Leads Management

**Purpose**: Track sales opportunities from initial lead to closure.

**Features**:
- Opportunity pipeline tracking
- Multiple service types per opportunity
- Win probability estimation
- Financial forecasting (estimated value, cost, margin)
- Team assignment (sales executive, pre-sales)
- Quote linking and tracking
- Stage-based workflow (Qualification → Proposal → Negotiation → Closed Won/Lost)

**Key Files**:
- [LeadsView.vue](src/views/LeadsView.vue)
- [opportunities.ts](src/stores/opportunities.ts)

**Workflow**:
1. Create opportunity from customer inquiry
2. Assign to sales and pre-sales team members
3. Estimate value and probability
4. Move through pipeline stages
5. Create quotes from opportunity
6. Track to won/lost closure

---

### 3. Manufacturers Management

**Purpose**: Maintain manufacturer database with product categories.

**Features**:
- Manufacturer profile (name, code, country, contact info)
- Product category management per manufacturer
- Category-based product catalog organization
- Active/Inactive status tracking

**Key Files**:
- [ManufacturersView.vue](src/views/ManufacturersView.vue)
- [manufacturers.ts](src/stores/manufacturers.ts)

**Workflow**:
1. Add manufacturer with basic details
2. Define product categories for manufacturer
3. Categories become available for product creation
4. Link products to manufacturer categories

**Category Structure**:
```
Manufacturer (e.g., Hikvision)
├── Category 1: CCTV Systems
│   ├── Product 1
│   ├── Product 2
│   └── Product 3
└── Category 2: Access Control Systems
    ├── Product 1
    └── Product 2
```

---

### 4. Product Management

**Purpose**: Comprehensive product catalog with cost tracking and inventory integration.

**Features**:
- Product types: Import vs Local products
- Multi-currency support with exchange rates
- Automatic cost calculations:
  - Unit cost in origin currency
  - Exchange rate conversion
  - Import costs (freight, customs, clearance)
  - Landed cost calculation
  - Target margin and selling price
- Category selection from manufacturer
- Lead time tracking
- Supplier/distributor information
- Warehouse stock integration

**Key Files**:
- [ProductsView.vue](src/views/ProductsView.vue)
- [products.ts](src/stores/products.ts)

**Costing Flow**:

**For Import Products**:
```
Unit Cost (USD) × FX Rate = Cost in SAR
Cost in SAR + Freight% + Customs% + Clearance% = Landed Cost
Landed Cost ÷ (1 - Target Margin%) = Selling Price
```

**For Local Products**:
```
Unit Cost (SAR) = Landed Cost
Landed Cost ÷ (1 - Target Margin%) = Selling Price
```

**Workflow**:
1. Select manufacturer (required first)
2. Choose category from manufacturer's categories
3. Enter product details (SKU, name, description)
4. Define costing:
   - For imports: origin currency, cost, FX rate, import percentages
   - For local: cost in SAR
5. Set target margin
6. System calculates selling price automatically
7. Product ready for quoting

---

### 5. Warehouse & Inventory Management

**Purpose**: Track product stock across multiple warehouse locations.

**Features**:
- Multiple warehouse locations (Riyadh Main, Jeddah Branch, Dammam Branch)
- Quantity tracking:
  - On-hand quantity
  - Reserved quantity
  - Available quantity
- Stock valuation
- Integration with quote builder (real-time stock display)
- Lead time display when out of stock

**Key Files**:
- [InventoryView.vue](src/views/InventoryView.vue)
- [warehouseStock.ts](src/stores/warehouseStock.ts)

**Workflow**:
1. Stock received into warehouse location
2. Quantities updated (on-hand, available)
3. When creating quotes:
   - System checks available stock
   - Shows green badge if in stock
   - Shows orange badge with lead time if out of stock
4. Stock reserved when quote becomes order

---

### 6. Quotation System

**Purpose**: Create, manage, and track customer quotations with approval workflows.

**Features**:
- Quote creation from opportunities
- Line item management:
  - Products (from catalog)
  - Labor/Services
  - Miscellaneous items
- Three categories: Materials, Manpower, Miscellaneous
- Advanced builder interface with autocomplete
- Real-time inventory display:
  - Manufacturer name
  - Stock availability
  - Lead time if unavailable
- Automatic calculations:
  - Line totals
  - Subtotal, discounts, VAT
  - Total cost and margins
- Multi-currency support
- Quote versioning
- Approval workflows based on margin thresholds
- Quote status tracking
- Print functionality with professional layout
- Excel export with full item details

**Key Files**:
- [QuotingView.vue](src/views/QuotingView.vue)
- [QuoteBuilderModern.vue](src/components/QuoteBuilderModern.vue)
- [quotes.ts](src/stores/quotes.ts)

**Quote Statuses**:
- Draft
- Pending Approval
- Approved
- Sent
- Accepted
- Declined
- Expired

**Quote Builder Features**:
1. **Product Search with Autocomplete**:
   - Search by SKU or name
   - Shows real-time inventory info
   - Displays manufacturer
   - Stock status indicator (green/orange)
   - Lead time for out-of-stock items

2. **Line Item Information Display**:
   - Product info bar on each line showing:
     - Manufacturer
     - Stock availability
     - Lead time (if not in stock)

3. **Categorization**:
   - Materials: Physical products and equipment
   - Manpower: Labor, installation, services
   - Miscellaneous: Accessories, project management, training

**Workflow**:
1. Create quote from opportunity
2. Add line items:
   - Search and select products (see stock info)
   - Enter quantities and pricing
   - Add labor/service items
   - Add miscellaneous costs
3. System calculates totals automatically
4. Review margin calculations
5. If margin below threshold → requires approval
6. Approve (if needed)
7. Print or export quote
8. Send to customer
9. Track response (accepted/declined)

**Export Features**:
- **Print**: Professional HTML layout with company branding
- **Excel**: Comprehensive export with:
  - All line items organized by category
  - Manufacturer information
  - Stock availability
  - Lead time
  - Unit costs, prices, margins
  - Quote totals and summary

---

### 7. Price Books Management

**Purpose**: Create special pricing agreements for specific customers or contracts.

**Features**:
- Multiple price book types:
  - Standard
  - Volume (quantity-based discounts)
  - Contract-specific
  - Promotional
  - Customer-specific
- Product-level custom pricing
- Validity period management
- Customer/contract assignment
- Price book builder with product search

**Key Files**:
- [PriceBooksView.vue](src/views/PriceBooksView.vue)
- [PriceBookBuilderModern.vue](src/components/PriceBookBuilderModern.vue)
- [priceBooks.ts](src/stores/priceBooks.ts)

**Workflow**:
1. Create price book with type and validity
2. Assign to customers/contracts
3. Add products with custom pricing
4. Set discounts or special prices
5. Use price book when creating quotes for assigned customers

---

### 8. Exchange Rates Management

**Purpose**: Centralized currency conversion management.

**Features**:
- Support for multiple currencies (USD, EUR, GBP, AED, CNY)
- Historical rate tracking
- Effective date management
- Current rate vs historical rates
- Used automatically in product costing

**Key Files**:
- [ExchangeRatesView.vue](src/views/ExchangeRatesView.vue)
- [exchangeRates.ts](src/stores/exchangeRates.ts)

**Workflow**:
1. Set/update exchange rate for currency
2. System tracks history with effective dates
3. Rates automatically used in:
   - Product cost calculations
   - Quote currency conversions
   - Import cost calculations

---

### 9. Contracts Management

**Purpose**: Manage customer contracts and service agreements.

**Features**:
- Contract types: Sales, Maintenance, Service, Project, Subscription
- Contract lifecycle tracking
- Auto-renewal options
- Renewal notice management
- Contract-to-contract relationships (renewals)
- Financial tracking
- Document management
- SLA and terms storage

**Key Files**:
- [ContractsView.vue](src/views/ContractsView.vue)
- [contracts.ts](src/stores/contracts.ts)

**Contract Statuses**:
- Draft
- Pending Approval
- Active
- Expired
- Terminated
- Renewed

**Workflow**:
1. Create contract (often from accepted quote)
2. Define terms, duration, value
3. Set renewal options
4. Link documents
5. Activate contract
6. Track renewals
7. Manage terminations or renewals

---

### 10. Recurring Services Management

**Purpose**: Manage recurring service offerings and subscriptions.

**Features**:
- Service definition with costs
- Monthly/annual pricing
- Service types (guards, maintenance, monitoring, patrols)
- Target margin calculations
- Integration with quotes and contracts

**Key Files**:
- [RecurringServicesView.vue](src/views/RecurringServicesView.vue)
- [recurringServices.ts](src/stores/recurringServices.ts)

**Use Cases**:
- Security guard services
- System maintenance contracts
- 24/7 monitoring services
- Mobile patrol services
- Facility management

---

### 11. Team & Personnel Management

**Purpose**: Manage internal teams and team members.

**Features**:
- Team organization by department
- Member management with roles
- Team leader assignment
- Status tracking
- Integration with opportunities (assignment)

**Key Files**:
- [TeamsView.vue](src/views/TeamsView.vue)
- [teams.ts](src/stores/teams.ts)

**Departments**:
- Sales
- Pre-Sales
- Technical
- Support
- Marketing
- Management
- Operations

---

### 12. Documents Management

**Purpose**: Central document repository for contracts, quotes, and compliance.

**Features**:
- Document categorization (Contract, Quote, General, Compliance, Legal)
- Document types (Terms, Delivery, Technical, Warranty, SLA)
- Tagging system
- Version control
- Multi-entity linking (quotes, contracts, customers)
- File metadata tracking

**Key Files**:
- [DocumentsView.vue](src/views/DocumentsView.vue)
- [documents.ts](src/stores/documents.ts)

---

## Key Business Flows

### Complete Sales Cycle

```
1. Lead Generation
   ↓
2. Create Opportunity
   ↓
3. Assign Sales Team
   ↓
4. Move Through Pipeline (Qualification → Proposal → Negotiation)
   ↓
5. Create Quote
   ├─ Add Products (check inventory)
   ├─ Add Services
   ├─ Calculate Margins
   └─ Get Approvals (if needed)
   ↓
6. Send Quote to Customer
   ↓
7. Customer Response
   ├─ Accepted → Create Contract
   ├─ Declined → Mark Lost
   └─ Negotiate → Revise Quote
   ↓
8. Contract Execution
   ↓
9. Recurring Revenue (if applicable)
```

### Product-to-Quote Flow

```
1. Manufacturer Setup
   ├─ Add manufacturer details
   └─ Define product categories
   ↓
2. Product Creation
   ├─ Select manufacturer
   ├─ Choose category (from manufacturer)
   ├─ Define costing (import/local)
   └─ Set margin and pricing
   ↓
3. Inventory Management
   ├─ Receive stock
   └─ Track across warehouses
   ↓
4. Quote Creation
   ├─ Search products (autocomplete)
   ├─ View stock availability
   ├─ See manufacturer info
   ├─ Check lead times
   └─ Add to quote
   ↓
5. Quote Approval & Sending
   ├─ Auto-approve or manual review
   └─ Print/Export/Email
   ↓
6. Order Processing
   └─ Reserve inventory
```

### Pricing Strategy Flow

```
1. Base Pricing (Products)
   └─ Standard margins from product setup
   ↓
2. Price Books (Optional)
   └─ Special pricing for customers/contracts
   ↓
3. Quote-Level Adjustments
   ├─ Line-item discounts
   └─ Quote-level discount percentage
   ↓
4. Approval Workflow
   └─ Low margins trigger approval
   ↓
5. Final Pricing to Customer
```

## Data Models

### Product Cost Calculation

```typescript
// Import Product
originCost (USD) = 100
fxRate = 3.75
costInSAR = 100 × 3.75 = 375 SAR

freightPercent = 5%
customsPercent = 5%
clearancePercent = 2%
totalImportCosts = 375 × (5% + 5% + 2%) = 45 SAR

landedCostSAR = 375 + 45 = 420 SAR

targetMarginPercent = 25%
sellingPrice = 420 ÷ (1 - 0.25) = 560 SAR

marginAmount = 560 - 420 = 140 SAR
marginPercent = (140 ÷ 560) × 100 = 25%
```

### Quote Totals Calculation

```typescript
// Line Items
item1: quantity × unitPrice = lineTotal
item2: quantity × unitPrice = lineTotal
...

subtotal = sum of all lineTotals

discountAmount = subtotal × discountPercent
subtotalAfterDiscount = subtotal - discountAmount

vatAmount = subtotalAfterDiscount × vatPercent (15%)
total = subtotalAfterDiscount + vatAmount

totalCost = sum of all (quantity × unitCost)
marginAmount = total - totalCost
marginPercent = (marginAmount ÷ total) × 100
```

## User Interface Features

### Modern Quote Builder
- Real-time autocomplete product search
- Visual inventory indicators
- Drag-and-drop line reordering
- Inline editing
- Category-based organization
- Collapsible sections
- Summary cards with financial metrics

### Dashboard Views
- KPI cards
- Pipeline visualization
- Recent activity feeds
- Quick actions
- Filters and search

### Responsive Tables
- Sortable columns
- Inline actions
- Status badges
- Color-coded margins
- Expandable details

## Business Rules

### Approval Requirements
- Quotes with margin < 20% require approval
- Large value quotes (> 500,000 SAR) require approval
- High discount percentages (> 10%) require approval

### Inventory Management
- Products show available quantity in real-time
- Out-of-stock products display lead time
- Stock levels color-coded (green=available, orange=lead time)

### Pricing Rules
- All prices rounded up (ceiling function)
- VAT automatically calculated at 15%
- Margins calculated on selling price (not cost)

### Quote Workflow
- Draft quotes can be edited
- Sent quotes cannot be modified (must create new version)
- Accepted quotes can be converted to contracts
- Expired quotes can be renewed with new validity

## Future Enhancements

Potential areas for expansion:
- Purchase Order management
- Project management module
- Time tracking and resource allocation
- Customer portal for quote viewing
- Mobile application
- Email integration
- Reporting and analytics dashboard
- Integration with accounting systems
- Automated follow-up workflows
- E-signature integration

## Project Structure

```
crm-dashboard/
├── src/
│   ├── components/
│   │   ├── UI/              # Reusable UI components
│   │   ├── QuoteBuilderModern.vue
│   │   └── PriceBookBuilderModern.vue
│   ├── views/               # Main application views
│   │   ├── CustomersView.vue
│   │   ├── LeadsView.vue
│   │   ├── ProductsView.vue
│   │   ├── ManufacturersView.vue
│   │   ├── QuotingView.vue
│   │   ├── PriceBooksView.vue
│   │   ├── ContractsView.vue
│   │   ├── InventoryView.vue
│   │   ├── TeamsView.vue
│   │   └── ...
│   ├── stores/              # Pinia state management
│   │   ├── customers.ts
│   │   ├── opportunities.ts
│   │   ├── products.ts
│   │   ├── manufacturers.ts
│   │   ├── quotes.ts
│   │   ├── warehouseStock.ts
│   │   ├── exchangeRates.ts
│   │   └── ...
│   ├── types/
│   │   └── index.ts         # TypeScript interfaces
│   ├── composables/         # Vue composables
│   ├── router/              # Vue Router configuration
│   └── assets/              # Static assets
├── package.json
└── vite.config.ts
```

## Getting Started

### Prerequisites
- Node.js ^20.19.0 or >=22.12.0
- npm or yarn

### Installation
```bash
npm install
```

### Development
```bash
npm run dev
```

### Build
```bash
npm run build
```

### Type Check
```bash
npm run type-check
```

## Notes

- All currency values are in Saudi Riyals (SAR) unless specified
- The system uses client-side state management (Pinia) - no backend currently
- Sample data is included for demonstration purposes
- The application follows Vue 3 Composition API patterns with TypeScript
