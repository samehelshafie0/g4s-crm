import type { Quote, QuoteLineItem, Department, WarehouseStock, StockReservation, InventoryMovement, Team, Project, Document as CrmDocument } from '@/types'
export type { WarehouseStock, StockReservation, InventoryMovement, Team, Project, CrmDocument }
export interface ServiceItem {
  id: string; sku: string; name: string; description: string; department: Department
  rateType: string; unitCost: number; unitPrice: number; isActive: boolean
}
export interface QuoteAddress {
  contactName: string; company: string; address: string; city: string; country: string; phone: string; email: string
}
export interface BuilderLine extends QuoteLineItem {
  rowType: 'item' | 'heading' | 'comment' | 'subtotal' | 'discount'
  source: 'product' | 'service' | 'recurring' | 'write-in'
  serviceId?: string; recurringServiceId?: string; multiplier: number; discountPercent: number
  isOptional: boolean; isSelected: boolean; isPrintable: boolean
  headingText?: string; commentText?: string; rateType?: string; billingCycle?: string
}
export interface BuilderQuote extends Quote {
  lockVersion: number; parentQuoteId?: string; paymentTerms: string; deliveryTerms: string
  introductionText: string; closingText: string; internalNotes: string; purchasingNotes: string; statementOfWork: string
  soldTo: QuoteAddress; shipTo: QuoteAddress; lineItems: BuilderLine[]
}
export interface BuilderInput {
  lockVersion: number; customerId: string; opportunityId?: string | null; currency: string
  validUntil?: string | null; discountPercent: number; vatPercent: number; notes: string
  paymentTerms: string; deliveryTerms: string; introductionText: string; closingText: string
  internalNotes: string; purchasingNotes: string; statementOfWork: string; soldTo: QuoteAddress; shipTo: QuoteAddress
  rows: Array<Omit<BuilderLine, 'quoteId' | 'category' | 'lineTotal' | 'marginPercent' | 'manufacturerName'> & { manufacturer?: string }>
}
export interface DocumentVersion {
  id: string; documentId: string; version: string; fileName: string; fileSize: number; fileType: string; createdAt: string
}
export interface UserLookup { id: string; firstName: string; lastName: string; email: string; role: string }
export interface Activity { id: string; entityType: string; entityId: string; action: string; description: string; createdAt: string; user?: { firstName: string; lastName: string } }
