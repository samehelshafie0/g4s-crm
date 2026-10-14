// ─── Base Types ─────────────────────────────────────────────
export interface BaseEntity {
  id: string
  createdAt: string
  updatedAt: string
}

// ─── Customer Management ────────────────────────────────────
export type CustomerStatus = 'active' | 'inactive' | 'prospect'
export type CustomerType = 'get' | 'grow'
export type Sector =
  | 'government'
  | 'healthcare'
  | 'education'
  | 'retail'
  | 'banking'
  | 'oil-gas'
  | 'telecom'
  | 'hospitality'
  | 'real-estate'
  | 'other'

export interface CustomerContact {
  id: string
  name: string
  email: string
  phone: string
  position: string
  isPrimary: boolean
}

export type SiteType = 'head-office' | 'branch' | 'warehouse' | 'plant' | 'retail' | 'other'

export interface CustomerSite {
  id: string
  customerId?: string
  name: string
  code: string
  siteType: SiteType
  address: string
  city: string
  region: string
  latitude?: number | null
  longitude?: number | null
  contactName: string
  contactPhone: string
  isActive: boolean
  contacts?: CustomerContact[]
}

export interface Customer extends BaseEntity {
  companyName: string
  sector: Sector
  region: string
  status: CustomerStatus
  type: CustomerType
  crNumber: string
  vatNumber: string
  sites: CustomerSite[]
  contacts: CustomerContact[]
  notes: string
}

// ─── Opportunities & Leads ──────────────────────────────────
export type OpportunityStage =
  | 'qualification'
  | 'proposal'
  | 'negotiation'
  | 'closed-won'
  | 'closed-lost'

export type ServiceType =
  | 'cctv'
  | 'access-control'
  | 'intrusion-detection'
  | 'fire-alarm'
  | 'networking'
  | 'it-solutions'
  | 'guarding'
  | 'monitoring'
  | 'maintenance'
  | 'consulting'

export interface Opportunity extends BaseEntity {
  title: string
  customerId: string
  customerName: string
  stage: OpportunityStage
  serviceTypes: ServiceType[]
  estimatedValue: number
  estimatedCost: number
  estimatedMargin: number
  winProbability: number
  salesExecutiveId: string
  salesExecutiveName: string
  preSalesId: string
  preSalesName: string
  expectedCloseDate: string
  quoteIds: string[]
  notes: string
  sites: CustomerSite[]
  currentQuote?: OpportunityQuoteRef | null
  deliveryStatus?: DeliveryStatus
  projectManagerId?: string
  projectManagerName?: string
  startDate?: string
  targetEndDate?: string
  actualEndDate?: string
  actualValue?: number
  actualCost?: number
}

export interface OpportunityQuoteRef {
  id: string
  quoteNumber: string
  version: number
  status: QuoteStatus
  total: number
  currency: Currency
  fxRate: number
}

// ─── Delivery (a won opportunity is the project) ────────────
export type DeliveryStatus = 'planning' | 'in-progress' | 'on-hold' | 'completed' | 'cancelled'

export type FulfilmentSource = 'stock' | 'purchase' | 'service'

export interface FulfilmentOrder {
  purchaseOrderId: string
  poNumber: string
  status: string
  quantity: number
  received: number
  expectedDelivery?: string
}

export interface FulfilmentLine {
  quoteLineItemId: string
  sku: string
  description: string
  siteId?: string
  quantity: number
  source: FulfilmentSource | 'unassigned'
  reserved: number
  ordered: number
  received: number
  damaged: number
  outstanding: number
  orders: FulfilmentOrder[]
  state: 'covered' | 'on-order' | 'partial' | 'open' | 'not-applicable'
}

export interface Fulfilment {
  basis: 'accepted' | 'current' | 'none'
  quoteId?: string
  quoteNumber?: string
  version?: number
  lines: FulfilmentLine[]
}

// ─── Manufacturers ──────────────────────────────────────────
export interface ManufacturerCategory {
  id: string
  name: string
  description: string
}

export type VendorType = 'manufacturer' | 'supplier' | 'both'

export interface Manufacturer extends BaseEntity {
  name: string
  code: string
  country: string
  contactEmail: string
  contactPhone: string
  website: string
  categories: ManufacturerCategory[]
  vendorType: VendorType
  isActive: boolean
}

// ─── Products ───────────────────────────────────────────────
export type ProductType = 'import' | 'local'
export type Currency = 'SAR' | 'USD' | 'EUR' | 'GBP' | 'AED' | 'CNY'

export interface Product extends BaseEntity {
  vendorEntries?: ProductVendorEntry[]
  priceHistory?: ProductPriceRecord[]
  documents?: ProductDocument[]
  sku: string
  name: string
  description: string
  manufacturerId: string
  manufacturerName: string
  categoryId: string
  categoryName: string
  productType: ProductType
  originCurrency: Currency
  unitCostOrigin: number
  fxRate: number
  costInSAR: number
  freightPercent: number
  customsPercent: number
  clearancePercent: number
  landedCostSAR: number
  targetMarginPercent: number
  sellingPrice: number
  marginAmount: number
  leadTimeDays: number
  supplierName: string
  isActive: boolean
}

// ─── Product Vendor Catalog Entry ────────────────────────────
export interface ProductVendorEntry {
  id: string
  vendorName: string
  vendorSku?: string
  unitCost: number
  currency: Currency
  moq: number
  leadTimeDays: number
  lastQuoteDate: string
  catalogSource?: string
  notes?: string
}

// ─── Product Price History ──────────────────────────────────
export type PriceSource = 'vendor-catalog' | 'purchase-order' | 'supplier-quote' | 'goods-receipt' | 'manual'
export interface ProductPriceRecord {
  id: string
  date: string
  source: PriceSource
  sourceRef?: string
  vendorName: string
  unitCost: number
  currency: Currency
  landingCost?: number
  qty: number
  notes?: string
}

// ─── Product Attachment / Document ──────────────────────────
export type ProductDocType = 'datasheet' | 'manual' | 'certificate' | 'vendor-quote' | 'catalog' | 'image' | 'other'
export interface ProductDocument {
  documentId?: string
  id: string
  name: string
  docType: ProductDocType
  fileName: string
  fileSize: string
  uploadedBy: string
  uploadedAt: string
  url?: string
  notes?: string
}

// ─── Warehouse & Inventory ──────────────────────────────────
export type WarehouseLocation = 'riyadh-main' | 'jeddah-branch' | 'dammam-branch'

export interface WarehouseStock extends BaseEntity {
  manufacturerName?: string
  categoryName?: string
  productId: string
  productSku: string
  productName: string
  warehouseLocation: WarehouseLocation
  onHandQty: number
  reservedQty: number
  availableQty: number
  unitCost: number
  totalValue: number
  reorderLevel?: number
}

// ─── Quotation System ───────────────────────────────────────
export type QuoteStatus =
  | 'draft'
  | 'pending-approval'
  | 'approved'
  | 'sent'
  | 'accepted'
  | 'declined'
  | 'expired'

export type QuoteLineCategory = 'materials' | 'manpower' | 'miscellaneous'

export interface QuoteLineItem {
  rowType?: string
  multiplier?: number
  isOptional?: boolean
  isSelected?: boolean
  id: string
  category: QuoteLineCategory
  productId?: string
  sku?: string
  description: string
  manufacturerName?: string
  stockAvailable?: number
  leadTimeDays?: number
  quantity: number
  unitCost: number
  unitPrice: number
  lineTotal: number
  marginPercent: number
  siteId?: string
  fulfilmentSource?: FulfilmentSource
}

export interface Quote extends BaseEntity {
  priceBookId?: string
  quoteNumber: string
  opportunityId?: string
  contractId?: string
  isCurrent: boolean
  revisionReason?: string
  supersededAt?: string
  fxRate: number
  customerId: string
  customerName: string
  version: number
  status: QuoteStatus
  lineItems: QuoteLineItem[]
  subtotal: number
  discountPercent: number
  discountAmount: number
  subtotalAfterDiscount: number
  vatPercent: number
  vatAmount: number
  total: number
  totalCost: number
  marginAmount: number
  marginPercent: number
  validUntil: string
  currency: Currency
  notes: string
  approvedBy?: string
  approvedAt?: string
}

export type QuoteCurrency = 'SAR' | 'USD'

export interface ContractQuote {
  id: string
  quoteNumber: string
  version: number
  status: QuoteStatus
  total: number
  currency: Currency
  fxRate: number
  createdAt: string
  role: 'origin' | 'raised-inside'
}

// ─── Price Books ────────────────────────────────────────────
export type PriceBookType =
  | 'standard'
  | 'volume'
  | 'contract'
  | 'promotional'
  | 'customer-specific'

export interface PriceBookEntry {
  kind?: 'product' | 'service' | 'recurring'
  serviceId?: string
  recurringServiceId?: string
  id: string
  productId: string
  productSku: string
  productName: string
  standardPrice: number
  customPrice: number
  discountPercent: number
}

export interface PriceBook extends BaseEntity {
  name: string
  type: PriceBookType
  description: string
  customerId?: string
  customerName?: string
  contractId?: string
  validFrom: string
  validTo: string
  isActive: boolean
  entries: PriceBookEntry[]
}

// ─── Exchange Rates ─────────────────────────────────────────
export interface ExchangeRateHistory {
  rate: number
  effectiveDate: string
}

export interface ExchangeRate extends BaseEntity {
  fromCurrency: Currency
  toCurrency: 'SAR'
  currentRate: number
  effectiveDate: string
  history: ExchangeRateHistory[]
}

// ─── Contracts ──────────────────────────────────────────────
export type ContractType =
  | 'sales'
  | 'maintenance'
  | 'service'
  | 'project'
  | 'subscription'

export type ContractStatus =
  | 'draft'
  | 'pending-approval'
  | 'active'
  | 'expired'
  | 'terminated'
  | 'renewed'

export interface Contract extends BaseEntity {
  contractNumber: string
  title: string
  customerId: string
  customerName: string
  type: ContractType
  status: ContractStatus
  startDate: string
  endDate: string
  value: number
  autoRenew: boolean
  renewalNoticeDays: number
  renewedFromId?: string
  quoteId?: string
  terms: string
  notes: string
}

// ─── Recurring Services ─────────────────────────────────────
export type RecurringServiceType =
  | 'guarding'
  | 'maintenance'
  | 'monitoring'
  | 'patrol'
  | 'facility-management'
  | 'rental'

export type BillingFrequency = 'monthly' | 'quarterly' | 'annually'

export interface RecurringServiceLine {
 id: string; source: 'product' | 'service' | 'vendor' | 'write-in'; sourceId?: string; sku: string; name: string; description: string; qty: number; unitCost: number; unitPrice: number; vendorName?: string; vendorFile?: string
}
export interface RecurringService extends BaseEntity {
 lineItems?: RecurringServiceLine[]
  name: string
  serviceType: RecurringServiceType
  description: string
  monthlyCost: number
  monthlyPrice: number
  annualCost: number
  annualPrice: number
  targetMarginPercent: number
  billingFrequency: BillingFrequency
  isActive: boolean
}

// ─── Teams ──────────────────────────────────────────────────
export type Department =
  | 'sales'
  | 'pre-sales'
  | 'technical'
  | 'support'
  | 'marketing'
  | 'management'
  | 'operations'

export interface TeamMember {
  id: string
  name: string
  email: string
  phone: string
  role: string
  department: Department
  isLeader: boolean
}

export interface Team extends BaseEntity {
  name: string
  department: Department
  description: string
  leaderId: string
  leaderName: string
  members: TeamMember[]
  isActive: boolean
}

// ─── Documents ──────────────────────────────────────────────
export type DocumentCategory =
  | 'contract'
  | 'quote'
  | 'general'
  | 'compliance'
  | 'legal'

export type DocumentType =
  | 'terms'
  | 'delivery'
  | 'technical'
  | 'warranty'
  | 'sla'

export interface Document extends BaseEntity {
  fileName?: string
  name: string
  category: DocumentCategory
  documentType: DocumentType
  tags: string[]
  version: string
  fileSize: string
  fileType: string
  linkedEntities: { type: string; id: string; name: string }[]
  links?: { id: string; documentId: string; entityType: string; entityId: string; entityName: string }[]
  uploadedBy: string
}

// ─── Procurement / Purchase Orders ──────────────────────────
export type PurchaseOrderStatus =
  | 'draft'
  | 'pending-approval'
  | 'approved'
  | 'ordered'
  | 'partial-received'
  | 'received'
  | 'cancelled'

export interface PurchaseOrderItem {
  id: string
  productId: string
  productSku: string
  productName: string
  manufacturerName: string
  quantity: number
  unitCost: number
  total: number
  receivedQty: number
  quoteLineItemId?: string
  leadTimeDays: number
}

export interface PurchaseOrder extends BaseEntity {
  poNumber: string
  supplierName: string
  supplierId?: string
  status: PurchaseOrderStatus
  items: PurchaseOrderItem[]
  subtotal: number
  shippingCost: number
  customsDuty: number
  total: number
  currency: Currency
  expectedDelivery: string
  actualDelivery?: string
  sourceQuoteId?: string
  sourceQuoteNumber?: string
  opportunityId?: string
  opportunityTitle?: string
  notes: string
  approvedBy?: string
  approvedAt?: string
}

// ─── Supplier Quotes (Vendor RFQ Responses) ────────────────
export type SupplierQuoteStatus = 'received' | 'under-review' | 'accepted' | 'expired' | 'rejected'

export interface SupplierQuoteLineItem {
  id: string
  productId?: string
  productSku: string
  productName: string
  manufacturerName: string
  quantity: number
  unitCost: number
  total: number
  leadTimeDays: number
  moq?: number
  validUntil?: string
  notes?: string
}

export interface SupplierQuote extends BaseEntity {
  sqNumber: string
  supplierName: string
  supplierId?: string
  supplierRef?: string
  sourceQuoteId?: string
  sourceQuoteNumber?: string
  opportunityId?: string
  opportunityTitle?: string
  status: SupplierQuoteStatus
  items: SupplierQuoteLineItem[]
  subtotal: number
  currency: Currency
  validFrom: string
  validUntil: string
  contactName?: string
  contactEmail?: string
  paymentTerms?: string
  deliveryTerms?: string
  notes: string
}

// ─── Supplier-Item Catalog ─────────────────────────────────
export interface SupplierItemEntry {
  id: string
  supplierId: string
  supplierName: string
  productId: string
  productSku: string
  productName: string
  manufacturerName: string
  latestCost: number
  previousCost?: number
  costTrend: 'up' | 'down' | 'stable'
  moq: number
  leadTimeDays: number
  lastQuoteDate: string
  lastPODate?: string
  reliability: number
}

// ─── Goods Receipt ─────────────────────────────────────────
export interface GoodsReceiptItem {
  damagedQty?: number
  returnStatus?: 'not-required' | 'pending' | 'returned'
  returnReference?: string
  id: string
  poItemId: string
  productId: string
  productSku: string
  productName: string
  receivedQty: number
  unitCost: number
  shippingAlloc: number
  customsAlloc: number
  landingCost: number
  serialNumbers?: string[]
  storageLocation?: string
  condition: 'good' | 'damaged' | 'partial-damage'
  notes?: string
}

export interface GoodsReceipt extends BaseEntity {
  sourceCurrency?: Currency
  replacementPoId?: string
  replacementPo?: { id: string; poNumber: string; status: string }
  grNumber: string
  poId: string
  poNumber: string
  supplierName: string
  receiveDate: string
  items: GoodsReceiptItem[]
  totalItems: number
  totalLandingCost: number
  receivedBy: string
  notes: string
}

// ─── Stock Reservations (Holds) ─────────────────────────────
export type ReservationSource = 'quote' | 'project' | 'manual'
export type ReservationStatus = 'active' | 'released' | 'fulfilled'

export interface StockReservation {
  quoteLineItemId?: string
  id: string
  productId: string
  productSku: string
  productName: string
  warehouseLocation: WarehouseLocation
  qty: number
  source: ReservationSource
  sourceRef: string
  sourceLabel: string
  customerName: string
  reservedBy: string
  reservedAt: string
  releaseDate?: string
  releasedBy?: string
  releaseReason?: string
  status: ReservationStatus
  notes?: string
}

// ─── Inventory Movements ────────────────────────────────────
export type MovementType = 'transfer' | 'adjustment' | 'allocation' | 'receipt' | 'release' | 'write-off'

export interface InventoryMovement {
  id: string
  productId: string
  productSku: string
  productName: string
  movementType: MovementType
  qty: number
  fromWarehouse?: WarehouseLocation
  toWarehouse?: WarehouseLocation
  reference?: string
  reason: string
  performedBy: string
  performedAt: string
  notes?: string
}

// ─── Dashboard ──────────────────────────────────────────────
export interface DashboardKPI {
  label: string
  value: string | number
  change: number
  changeLabel: string
  icon: string
}

export interface ActivityItem {
  id: string
  type: 'quote' | 'opportunity' | 'contract' | 'customer'
  action: string
  description: string
  timestamp: string
  user: string
}

// ─── Navigation ─────────────────────────────────────────────
export interface NavItem {
  name: string
  path: string
  icon: string
  badge?: number
}

export interface NavGroup {
  title: string
  items: NavItem[]
}
