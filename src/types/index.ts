// ============================================================================
// Core Business Types
// ============================================================================

// Common types
export type Status = 'active' | 'inactive'
export type Currency = 'SAR' | 'USD' | 'EUR' | 'GBP' | 'AED' | 'CNY'
export type Region = 'Riyadh' | 'Jeddah' | 'Dammam' | 'Makkah' | 'Madinah' | 'Eastern Province' | 'Other'
export type Priority = 'low' | 'medium' | 'high' | 'critical'

// Sectors
export type Sector =
  | 'Banking & Finance'
  | 'Healthcare'
  | 'Retail & Manufacturing'
  | 'Real Estate'
  | 'Government'
  | 'Education'
  | 'Hospitality'
  | 'Technology'
  | 'Energy'
  | 'Transportation'
  | 'Other'

// ============================================================================
// Teams & Personnel
// ============================================================================

export interface TeamMember {
  id: string
  name: string
  email: string
  role: string
  phone?: string
  status: Status
  joinedAt: string
}

export interface Team {
  id: string
  name: string
  description: string
  department: 'Sales' | 'Pre-Sales' | 'Technical' | 'Support' | 'Marketing' | 'Management' | 'Operations' | 'Other'
  members: TeamMember[]
  leaderId?: string  // Team leader's member ID
  status: Status
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

// ============================================================================
// Company & Contacts
// ============================================================================

export interface Company {
  id: string
  name: string
  sector: Sector
  regions: string[] // Multiple regions in KSA
  cities: string[] // Multiple cities in KSA
  status: 'get' | 'grow' // Get: new customer, Grow: existing customer to expand

  // Contact Information
  phone?: string
  email?: string
  website?: string
  address?: string

  // Legal Information
  crNumber?: string
  vatNumber?: string

  // Business Information
  contactPerson?: string

  // Attached Files
  attachedFile?: {
    id: string
    name: string
    url: string
    uploadedAt: string
    uploadedBy: string
  }

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

export interface CompanySite {
  id: string
  companyId: string
  name: string
  region: Region
  address: string
  city: string
  isPrimary: boolean
  contactPerson?: string
  phone?: string
  email?: string
  notes?: string
  createdAt: string
}

export interface Contact {
  id: string
  companyId: string
  siteId?: string

  // Personal Information
  firstName: string
  lastName: string
  fullName: string
  title: string
  department: string

  // Contact Information
  email: string
  phone: string
  mobile?: string

  // Status
  isPrimary: boolean
  status: Status

  // Metadata
  createdAt: string
  createdBy: string
  notes?: string
}

// ============================================================================
// Opportunities & Leads
// ============================================================================

export type OpportunityStage =
  | 'qualification'
  | 'proposal'
  | 'negotiation'
  | 'closed_won'
  | 'closed_lost'

export type ServiceType =
  | 'Security Systems'
  | 'CCTV & Surveillance'
  | 'Access Control Systems'
  | 'Alarm Systems'
  | 'Fire Safety Systems'
  | 'Security Guards'
  | 'Mobile Patrols'
  | 'Event Security'
  | 'Facility Management'
  | 'Maintenance Services'
  | 'Installation Services'
  | 'Consulting Services'
  | 'Training Services'
  | 'Rent Service'
  | 'Integration Services'
  | 'Monitoring Services'
  | 'Risk Assessment'
  | 'Other'

export interface Opportunity {
  id: string
  opportunityNumber: string

  // Basic Information
  name: string
  description: string
  companyId: string
  contactId?: string
  contractId?: string  // Link to existing contract if this is a renewal/extension

  // Classification
  serviceType: ServiceType | ServiceType[]  // Can be single or multiple service types
  stage: OpportunityStage
  priority: Priority

  // Financial
  estimatedValue: number
  targetMargin: number
  estimatedCost: number
  currency: Currency

  // Probability & Dates
  winProbability: number
  requestDate: string
  actualCloseDate?: string

  // Lead Assignment - Team that got/created the opportunity
  teamId?: string  // Team that acquired this lead (can be Marketing, Sales, etc.)
  createdByPersonId?: string  // Specific person who created/got this lead

  // Quote Team Assignment - Required for creating quotes
  quoteSalesPersonId?: string  // Sales person assigned for quoting (from Sales team)
  quotePresalesPersonId?: string  // Pre-sales person assigned for quoting (from Pre-Sales team)

  // Legacy fields - kept for backward compatibility
  salesExecutiveId: string
  presalesId?: string

  // Quote Tracking
  quoteIds?: string[]  // Associated quote IDs
  quoteDates?: {
    quoteId: string
    sentDate: string
    respondedDate?: string
  }[]

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  updatedBy?: string
  notes?: string

  // Closure
  lostReason?: string
  competitorName?: string
}

// ============================================================================
// Manufacturers & Product Catalog
// ============================================================================

// Manufacturer Product - represents a specific product model from a manufacturer
export interface ManufacturerProduct {
  id: string
  name: string
  modelNumber: string
  description?: string
  category: ProductCategory
  specifications?: string
  createdAt: string
  notes?: string
}

// Manufacturer Category - groups products by type
export interface ManufacturerCategory {
  id: string
  name: ProductCategory
  description?: string
  products: ManufacturerProduct[]
}

// Manufacturer - represents a product manufacturer/brand
export interface Manufacturer {
  id: string
  name: string
  code: string  // Short code for manufacturer (e.g., "HIK" for Hikvision)
  description?: string
  country?: string
  website?: string
  contactEmail?: string
  contactPhone?: string
  categories: ManufacturerCategory[]  // Categories of products they manufacture
  status: Status
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

export type ProductCategory =
  | 'Security Equipment'
  | 'Access Control Systems'
  | 'CCTV Systems'
  | 'Alarm Systems'
  | 'Fire Safety Equipment'
  | 'Uniforms & Apparel'
  | 'Communication Equipment'
  | 'Vehicles'
  | 'Software & Licenses'
  | 'Professional Services'
  | 'Servers'
  | 'Networking'
  | 'Storage'
  | 'Other'

export type UnitOfMeasure = 'Unit' | 'Set' | 'Box' | 'License' | 'Hour' | 'Day' | 'Month' | 'Year'

export interface Product {
  id: string
  sku: string
  name: string
  description: string
  category: ProductCategory

  // Manufacturer & Supplier Information
  manufacturerId?: string  // Reference to manufacturer in manufacturers store
  manufacturer: string  // Manufacturer name (for legacy compatibility)
  manufacturerProductId?: string  // Reference to specific manufacturer product
  supplier?: string  // Optional - supplier/vendor
  vendorId?: string
  supplierPartNumber?: string

  // Costing
  isImport: boolean
  originCurrency: Currency
  unitCost: number
  fxRate: number
  costInSAR: number

  // Import Costs (percentages)
  freightPercent?: number
  customsPercent?: number
  clearancePercent?: number

  // Calculated Landed Cost
  landedCostSAR: number

  // Pricing
  targetMarginPercent: number
  sellingPrice: number  // Always ceiled (rounded up)

  // Inventory
  unitOfMeasure: UnitOfMeasure
  leadTimeDays: number

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

// Exchange Rate Management
export interface ExchangeRate {
  id: string
  currency: Currency
  rate: number  // 1 unit of currency = X SAR
  effectiveDate: string  // Date this rate became effective
  setBy: string  // User who set this rate
  setAt: string  // Timestamp when rate was set
  notes?: string
}

export interface ExchangeRateHistory {
  currency: Currency
  currentRate: number
  currentEffectiveDate: string
  history: ExchangeRate[]
}

// ============================================================================
// Quotations
// ============================================================================

export type QuoteStatus =
  | 'draft'
  | 'pending_approval'
  | 'approved'
  | 'sent'
  | 'accepted'
  | 'declined'
  | 'expired'

export type LineItemCategory = 'materials' | 'manpower' | 'miscellaneous'

export interface ApprovalHistoryEntry {
  date: string
  status: QuoteStatus
  userId: string
  userName: string
  comments: string
}

export interface Quote {
  id: string
  quoteNumber: string
  version: number

  // References
  opportunityId: string
  companyId: string
  contactId: string
  priceBookId?: string

  // Dates
  quoteDate: string
  validUntil: string
  expiryDate: string

  // Currency
  currency: Currency
  exchangeRate: number

  // Status & Workflow
  status: QuoteStatus

  // Financial Summary
  subtotal: number
  discountPercent: number
  discountAmount: number
  subtotalAfterDiscount: number
  vatPercent: number
  vatAmount: number
  total: number

  // Costing
  totalCost: number
  marginAmount: number
  marginPercent: number

  // Terms
  paymentTerms: string
  deliveryTerms: string
  internalNotes?: string
  customerNotes?: string

  // Contract Information (for recurring contracts)
  isRecurring?: boolean
  contractDurationYears?: number
  contractStartDate?: string
  contractEndDate?: string

  // Approval
  requiresApproval: boolean
  approvalReason?: string
  approvedBy?: string
  approvedAt?: string
  rejectionReason?: string

  // Team
  createdBy: string
  createdAt: string
  updatedAt?: string
  sentAt?: string

  // Customer Response
  acceptedAt?: string
  declinedAt?: string
  declineReason?: string


  // Attachments
  attachments?: QuoteAttachment[]
  // Approval History
  approvalHistory?: ApprovalHistoryEntry[]
}

export interface QuoteLineItem {
  id: string
  quoteId: string
  lineNumber: number

  // Product or Labor
  type: 'product' | 'labor' | 'service'
  category: LineItemCategory

  // Product Information (if type = 'product')
  productId?: string
  productSku?: string
  productName: string
  description: string

  // Labor Information (if type = 'labor')
  laborPosition?: string
  hoursPerMonth?: number
  months?: number

  // Pricing
  quantity: number
  unitCost: number
  unitPrice: number
  discountPercent: number
  discountAmount: number
  lineTotal: number

  // Costing
  totalCost: number
  marginAmount: number
  marginPercent: number

  // Metadata
  notes?: string
  comments?: string
}

// ============================================================================
// Labor & Manpower
// ============================================================================

export interface LaborPosition {
  id: string
  name: string
  description: string
  department: string

  // Rates
  costPerHour: number
  sellingRatePerHour: number

  // Metadata
  status: Status
  createdAt: string
  createdBy: string
  notes?: string
}

// ============================================================================
// Price Books
// ============================================================================

export type PriceBookType = 'standard' | 'volume' | 'contract' | 'promotional' | 'customer_specific'

export interface PriceBookEntry {
  id: string
  priceBookId: string
  productId: string
  productSku: string
  productName: string

  // Original pricing
  standardPrice: number

  // Custom pricing
  customPrice: number
  discountPercent: number
  discountAmount: number

  // Quantity breaks for volume pricing
  minQuantity?: number
  maxQuantity?: number

  // Metadata
  notes?: string
}

export interface PriceBook {
  id: string
  name: string
  description: string
  type: PriceBookType
  status: Status

  // Validity
  validFrom: string
  validUntil?: string

  // Assignment
  customerIds: string[]  // Empty array means applies to all customers
  contractIds: string[]  // Can be linked to specific contracts

  // Pricing entries
  priceBookEntries: PriceBookEntry[]

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

// ============================================================================
// Contracts
// ============================================================================

export type ContractStatus = 'draft' | 'pending_approval' | 'active' | 'expired' | 'terminated' | 'renewed'
export type ContractType = 'sales' | 'maintenance' | 'service' | 'project' | 'subscription'

export interface Contract {
  id: string
  contractNumber: string
  name: string
  type: ContractType
  status: ContractStatus

  // References
  quoteId?: string  // Generated from quote
  opportunityId?: string
  customerId: string
  contactId?: string
  priceBookId?: string

  // Dates
  startDate: string
  endDate: string
  signedDate?: string

  // Financial
  totalValue: number
  currency: Currency
  paymentTerms: string

  // Renewal
  autoRenewal: boolean
  renewalNoticeDays?: number
  renewedFromId?: string  // If this is a renewal
  renewedToId?: string    // If this was renewed

  // Documents
  documentIds: string[]

  // Terms & Conditions
  terms?: string
  deliverables?: string
  sla?: string

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  approvedBy?: string
  approvedAt?: string
  signedBy?: string  // Customer signatory
  notes?: string
}

// ============================================================================
// Warehouse & Inventory
// ============================================================================

export type WarehouseLocation = 'Riyadh Main' | 'Jeddah Branch' | 'Dammam Branch' | 'Other'

export interface WarehouseStock {
  id: string
  warehouseLocation: WarehouseLocation
  productId: string

  // Stock Levels
  quantityOnHand: number
  quantityReserved: number
  quantityAvailable: number

  // Costing
  unitCost: number
  totalValue: number

  // Batch Information
  batchNumber?: string
  expiryDate?: string

  // Supplier
  supplierId?: string
  supplierSku?: string
  lastPurchaseDate?: string
  lastPurchasePrice?: number

  // Metadata
  lastUpdated: string
  notes?: string
}

export interface StockMovement {
  id: string
  movementNumber: string
  productId: string
  warehouseLocation: WarehouseLocation

  // Movement Details
  movementType: 'receipt' | 'issue' | 'transfer' | 'adjustment'
  quantity: number

  // References
  referenceType?: 'purchase_order' | 'sales_order' | 'quote' | 'adjustment'
  referenceNumber?: string

  // Transfer (if type = 'transfer')
  fromLocation?: WarehouseLocation
  toLocation?: WarehouseLocation

  // Metadata
  movementDate: string
  performedBy: string
  notes?: string
}

// ============================================================================
// Pricing History & Intelligence
// ============================================================================

export interface PricingHistory {
  id: string

  // References
  quoteId: string
  quoteNumber: string
  companyId: string
  companyName: string
  productId: string
  productSku: string
  productName: string
  opportunityName?: string

  // Pricing Details
  quoteDate: string
  quantity: number
  unitPrice: number
  discountPercent: number
  finalUnitPrice: number
  lineTotal: number

  // Costing
  unitCost: number
  marginPercent: number
  marginAmount: number

  // Outcome
  quoteStatus: QuoteStatus
  wasWon: boolean
  wasLost: boolean
  isPending: boolean

  // Team
  salesExecutiveId: string
  salesExecutiveName: string
  presalesId?: string
  presalesName?: string

  // Metadata
  createdAt: string
}

// ============================================================================
// Pricing Rules
// ============================================================================

export type RuleType = 'volume_discount' | 'customer_segment' | 'product_category' | 'minimum_margin'

export interface PricingRule {
  id: string
  name: string
  description: string
  ruleType: RuleType

  // Applicability
  appliesTo: 'all' | 'category' | 'specific_product'
  productCategory?: ProductCategory
  productId?: string

  // Conditions
  minQuantity?: number
  minValue?: number
  customerSegment?: string

  // Actions
  discountPercent?: number
  minimumMarginPercent?: number

  // Approval
  requiresApproval: boolean
  approvalThreshold?: number

  // Priority & Status
  priority: number
  status: Status
  effectiveFrom: string
  effectiveTo?: string

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
}

// ============================================================================
// Users & Authentication
// ============================================================================

export type UserRole =
  | 'admin'
  | 'presales_manager'
  | 'sales_executive'
  | 'presales'
  | 'finance_reviewer'

export interface User {
  id: string
  email: string

  // Personal Information
  firstName: string
  lastName: string
  fullName: string
  phone?: string

  // Organization
  role: UserRole
  department: string
  region?: Region

  // Status
  status: Status
  lastLogin?: string

  // Metadata
  createdAt: string
  createdBy?: string
}

// ============================================================================
// Approvals & Workflow
// ============================================================================

export type ApprovalStatus = 'pending' | 'approved' | 'rejected'
export type ApprovalEntityType = 'quote' | 'pricing_rule' | 'discount'

export interface Approval {
  id: string

  // Entity
  entityType: ApprovalEntityType
  entityId: string
  entityNumber: string

  // Approval Chain
  level: number
  approverRole: UserRole
  approverId?: string

  // Status
  status: ApprovalStatus

  // Reason
  requestReason: string

  // Response
  comments?: string
  rejectionReason?: string

  // Dates
  requestedAt: string
  requestedBy: string
  respondedAt?: string

  // Metadata
  notificationSent: boolean
}

// ============================================================================
// Audit & Compliance
// ============================================================================

export type AuditAction =
  | 'created'
  | 'updated'
  | 'deleted'
  | 'approved'
  | 'rejected'
  | 'sent'
  | 'exported'
  | 'login'
  | 'logout'

export type AuditEntityType =
  | 'company'
  | 'contact'
  | 'opportunity'
  | 'quote'
  | 'product'
  | 'user'
  | 'approval'
  | 'pricing_rule'
  | 'warehouse_stock'

export interface AuditLog {
  id: string

  // Entity
  entityType: AuditEntityType
  entityId: string
  entityIdentifier?: string // Human readable identifier

  // Action
  action: AuditAction

  // User
  userId: string
  userEmail: string
  userRole: UserRole

  // Details
  changes?: Record<string, { before: any; after: any }>
  metadata?: Record<string, any>

  // Technical
  ipAddress?: string
  userAgent?: string
  timestamp: string
}

// ============================================================================
// Reports & Analytics
// ============================================================================

export interface KPIMetrics {
  // Companies
  totalCompanies: number
  activeCompanies: number

  // Opportunities
  totalOpportunities: number
  activeOpportunities: number
  totalOpportunityValue: number

  // Quotes
  totalQuotes: number
  pendingApprovals: number
  quotesThisMonth: number
  totalQuoteValue: number

  // Win Rates
  wonQuotes: number
  lostQuotes: number
  winRate: number

  // Financial
  totalRevenue: number
  averageMargin: number
  averageDealSize: number
}

export interface SalesPipelineMetrics {
  stage: OpportunityStage
  count: number
  totalValue: number
  averageValue: number
  conversionRate?: number
}

export interface UserPerformanceMetrics {
  userId: string
  userName: string

  // Activity
  opportunitiesCreated: number
  quotesCreated: number

  // Success
  wonDeals: number
  lostDeals: number
  winRate: number

  // Financial
  totalRevenue: number
  averageMargin: number
  averageDealSize: number
}

// ============================================================================
// UI State & Preferences
// ============================================================================

export interface UserPreferences {
  theme: 'light' | 'dark'
  language: string
  dateFormat: string
  currency: Currency
  notifications: {
    email: boolean
    browser: boolean
  }
}

// ============================================================================
// Documents & Attachments
// ============================================================================

export type DocumentType = 'terms' | 'delivery' | 'technical' | 'warranty' | 'sla' | 'other'
export type DocumentCategory = 'contract' | 'quote' | 'general' | 'compliance' | 'legal'

export interface Document {
  id: string
  name: string
  fileName: string
  fileType: string // mime type
  fileSize: number // bytes
  category: DocumentCategory
  type: DocumentType
  description?: string
  fileUrl: string // In real app, this would be a storage URL
  fileData?: string // Base64 data for demo purposes
  uploadedBy: string
  uploadedAt: string
  tags?: string[]
  version?: number
  // Associations
  linkedToQuotes?: string[] // Quote IDs
  linkedToContracts?: string[] // Contract IDs
  linkedToCustomers?: string[] // Customer IDs
}

export interface QuoteAttachment {
  id: string
  documentId: string
  documentName: string // Custom name for this attachment
  documentType: DocumentType
  includeInPdf: boolean // Whether to append to PDF export
  order: number // Display order
}

// Add to Quote type
export interface QuoteWithAttachments extends Quote {
  attachments?: QuoteAttachment[]
}
