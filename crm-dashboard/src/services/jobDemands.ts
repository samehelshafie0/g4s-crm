import type { QuoteLineItem } from '@/types'

export type JobDemandLine = Pick<QuoteLineItem, 'id' | 'productId' | 'sku' | 'description' | 'category' | 'quantity' | 'multiplier'> & { remainingQty: number }
export interface JobDemand {
  id: string; quoteNumber: string; version: number; status: 'accepted'
  customerName: string; opportunityId?: string; opportunityTitle: string
  lineItems: JobDemandLine[]
}
