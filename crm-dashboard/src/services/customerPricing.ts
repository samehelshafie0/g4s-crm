import { http, type ApiResponse } from './http'
export interface CustomerPrice {
 id: string; productId?: string; sku: string; description: string; quantity: number; unitPrice: number
 unitCost?: number; discountPercent: number; quoteId: string; quoteNumber: string; version: number
 currency: string; fxRate: number; status: string; date: string
}
export const customerPriceHistory = (id: string, productId?: string) => http.get<ApiResponse<CustomerPrice[]>>(`/customers/${id}/price-history`, { params: { productId } }).then(r => r.data.data)
export interface QuoteOffer {
 id: string; unitCost: number; quantity: number; currency: string; supplierQuoteId: string; supplierQuoteNumber: string
 supplierName: string; supplierRef: string; quoteId?: string; quoteNumber?: string; version?: number; customerId?: string; customerName?: string; projectName?: string
 revision: number; recordedAt: string; validUntil?: string; isCurrent: boolean
}
export const productQuoteOffers = (id: string) => http.get<ApiResponse<QuoteOffer[]>>(`/products/${id}/quote-offers`).then(r => r.data.data)
export interface SalePrice {
 id: string; productId: string; sku: string; description: string; quantity: number; multiplier: number
 unitPrice: number; discountPercent: number; quoteDiscountPercent: number; netUnitPrice: number
 quoteId: string; quoteNumber: string; version: number; currency: string; fxRate: number
 customerId: string; customerName: string; acceptedAt: string; acceptedAtEstimated: boolean
}
export const productSalesHistory = (productId: string, customerId: string, scope: 'customer' | 'others', page = 1) => http.get<ApiResponse<SalePrice[]>>(`/products/${productId}/sales-history`, { params: { customerId, scope, page } }).then(r => r.data)
export const customerSalesHistory = (id: string, page = 1) => http.get<ApiResponse<SalePrice[]>>(`/customers/${id}/price-history`, { params: { scope: 'sold', page } }).then(r => r.data)
