import { editable } from './payload'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Product, ProductVendorEntry, ProductPriceRecord, ProductDocument } from '@/types'

export interface ProductImportRow { sku: string; name: string; description?: string; unitCost: number; qty?: number; leadTimeDays?: number; vendorSku?: string }
export interface ProductImportRequest {
  vendorName: string
  manufacturerId?: string
  categoryId?: string
  sourceRef?: string
  source: 'vendor-catalog' | 'supplier-quote' | 'manual'
  productType: 'import' | 'local'
  originCurrency: string
  fxRate: number
  freightPercent: number
  customsPercent: number
  clearancePercent: number
  targetMarginPercent: number
  updateExisting: boolean
  rows: ProductImportRow[]
}
export interface ProductImportOutcome { sku: string; status: string; reason?: string; landedCostSAR?: number; sellingPrice?: number; productId?: string }
export interface ProductImportResult { created: number; updated: number; skipped: number; failed: number; rows: ProductImportOutcome[] }

export interface ExtractedTable { page: number; lineItemRows: number; rows: string[][] }

/** Reads tables out of a vendor PDF on the server; nothing is stored. */
export const pdfExtractService = {
  tables: (file: File) => {
    const form = new FormData()
    form.append('file', file)
    return http
      .post<ApiResponse<{ fileName: string; tables: ExtractedTable[] }>>('/extract/pdf-tables', form, {
        timeout: 90_000,
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then(r => r.data)
  },
}

export const productsService = {
  importProducts: (data: ProductImportRequest) =>
    http.post<ApiResponse<ProductImportResult>>('/products/import', data, { timeout: 120_000 }).then(r => r.data),
  documents: (id: string) => http.get<ApiResponse<ProductDocument[]>>(`/products/${id}/documents`).then(r => r.data),
  addDocument: (id: string, data: {documentId:string;name:string;docType:string;notes?:string}) => http.post<ApiResponse<ProductDocument>>(`/products/${id}/documents`, data).then(r => r.data),
  deleteDocument: (id: string, documentId: string) => http.delete(`/products/${id}/documents/${documentId}`),
  vendors: (id: string) => http.get<ApiResponse<ProductVendorEntry[]>>(`/products/${id}/vendors`).then(r => r.data),
  saveVendor: (id: string, data: Partial<ProductVendorEntry>, vendorId?: string) => http.request<ApiResponse<ProductVendorEntry>>({ method: vendorId ? 'PATCH' : 'POST', url: `/products/${id}/vendors${vendorId ? '/' + vendorId : ''}`, data: editable(data, ['vendorName', 'vendorSku', 'currency', 'unitCost', 'moq', 'leadTimeDays', 'catalogSource', 'notes']) }).then(r => r.data),
  deleteVendor: (id: string, vendorId: string) => http.delete(`/products/${id}/vendors/${vendorId}`),
  priceHistory: (id: string) => http.get<ApiResponse<ProductPriceRecord[]>>(`/products/${id}/price-history`).then(r => r.data),
  addPrice: (id: string, data: Partial<ProductPriceRecord>) => http.post<ApiResponse<ProductPriceRecord>>(`/products/${id}/price-history`, editable(data, ['date', 'unitCost', 'currency', 'qty', 'vendorName', 'notes'])).then(r => r.data),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Product[]>>('/products', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Product>>(`/products/${id}`).then((r) => r.data),

  create: (data: Partial<Product>) =>
    http.post<ApiResponse<Product>>('/products', editable(data, ['sku', 'name', 'description', 'manufacturerId', 'categoryId', 'productType', 'originCurrency', 'unitCostOrigin', 'fxRate', 'freightPercent', 'customsPercent', 'clearancePercent', 'targetMarginPercent', 'sellingPrice', 'leadTimeDays', 'supplierName', 'isActive'])).then((r) => r.data),

  update: (id: string, data: Partial<Product>) =>
    http.patch<ApiResponse<Product>>(`/products/${id}`, editable(data, ['sku', 'name', 'description', 'manufacturerId', 'categoryId', 'productType', 'originCurrency', 'unitCostOrigin', 'fxRate', 'freightPercent', 'customsPercent', 'clearancePercent', 'targetMarginPercent', 'sellingPrice', 'leadTimeDays', 'supplierName', 'isActive'])).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/products/${id}`).then((r) => r.data),

  recalculateCosts: () =>
    http.post<ApiResponse<{ updated: number }>>('/products/recalculate-costs').then((r) => r.data),
}
