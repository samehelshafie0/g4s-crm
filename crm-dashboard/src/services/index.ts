import { editable } from './payload'
import type { ServiceItem, DocumentVersion, UserLookup, WarehouseStock, StockReservation, InventoryMovement, Team, Project, CrmDocument } from './workflowDtos'
import type { PurchaseOrderDto, SupplierQuoteDto, GoodsReceiptDto } from './procurementDtos'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Contract, PriceBook, ExchangeRate, RecurringService } from '@/types'

// ─── Contracts ──────────────────────────────────────────────
export const contractsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Contract[]>>('/contracts', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<Contract>>(`/contracts/${id}`).then((r) => r.data),
  create: (data: Partial<Contract>) =>
    http.post<ApiResponse<Contract>>('/contracts', editable(data, ['title', 'customerId', 'type', 'value', 'currency', 'startDate', 'endDate', 'autoRenew', 'renewalNoticeDays', 'terms', 'notes'])).then((r) => r.data),
  update: (id: string, data: Partial<Contract>) =>
    http.patch<ApiResponse<Contract>>(`/contracts/${id}`, editable(data, ['title', 'type', 'value', 'startDate', 'endDate', 'autoRenew', 'renewalNoticeDays', 'terms', 'notes'])).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/contracts/${id}`).then((r) => r.data),
  activate: (id: string) =>
    http.patch<ApiResponse<Contract>>(`/contracts/${id}/activate`).then((r) => r.data),
  terminate: (id: string) =>
    http.patch<ApiResponse<Contract>>(`/contracts/${id}/terminate`).then((r) => r.data),
  renew: (id: string) =>
    http.post<ApiResponse<Contract>>(`/contracts/${id}/renew`).then((r) => r.data),
  expiring: () =>
    http.get<ApiResponse<Contract[]>>('/contracts/expiring').then((r) => r.data),
}

// ─── Price Books ────────────────────────────────────────────
export const priceBooksService = {
  replaceEntries: (id: string, entries: import('@/types').PriceBookEntry[]) => http.put<ApiResponse<PriceBook>>(`/price-books/${id}/entries`, { entries: entries.map(entry => ({ kind:entry.kind ?? 'product', productId:(!entry.kind || entry.kind === 'product') ? entry.productId : undefined, serviceId:entry.kind === 'service' ? entry.productId : undefined, recurringServiceId:entry.kind === 'recurring' ? entry.productId : undefined, customPrice:entry.customPrice })) }).then(r => r.data),
  entries: (id: string) => http.get<ApiResponse<import('@/types').PriceBookEntry[]>>(`/price-books/${id}/entries`).then(r => r.data),
  saveEntry: (id: string, data: { productId: string; customPrice: number }, entryId?: string) => http.request({ method: entryId ? 'PATCH' : 'POST', url: `/price-books/${id}/entries${entryId ? '/' + entryId : ''}`, data }).then(r => r.data),
  deleteEntry: (id: string, entryId: string) => http.delete(`/price-books/${id}/entries/${entryId}`),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<PriceBook[]>>('/price-books', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<PriceBook>>(`/price-books/${id}`).then((r) => r.data),
  create: (data: Partial<PriceBook>) =>
    http.post<ApiResponse<PriceBook>>('/price-books', editable(data, ['name', 'type', 'description', 'customerId', 'contractId', 'validFrom', 'validTo', 'isActive'])).then((r) => r.data),
  update: (id: string, data: Partial<PriceBook>) =>
    http.patch<ApiResponse<PriceBook>>(`/price-books/${id}`, editable(data, ['name', 'type', 'description', 'customerId', 'contractId', 'validFrom', 'validTo', 'isActive'])).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/price-books/${id}`).then((r) => r.data),
}

// ─── Exchange Rates ─────────────────────────────────────────
export const exchangeRatesService = {
  create: (data: { fromCurrency: string; toCurrency: string; currentRate: number; effectiveDate: string }) => http.post<ApiResponse<ExchangeRate>>('/exchange-rates', data).then(r => r.data),
  list: () =>
    http.get<ApiResponse<ExchangeRate[]>>('/exchange-rates').then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<ExchangeRate>>(`/exchange-rates/${id}`).then((r) => r.data),
  update: (id: string, currentRate: number, effectiveDate?: string) =>
    http
      .patch<ApiResponse<ExchangeRate>>(`/exchange-rates/${id}`, { currentRate, effectiveDate })
      .then((r) => r.data),
}

// ─── Recurring Services ─────────────────────────────────────
export const recurringServicesService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<RecurringService[]>>('/recurring-services', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<RecurringService>>(`/recurring-services/${id}`).then((r) => r.data),
  create: (data: Partial<RecurringService>) =>
    http.post<ApiResponse<RecurringService>>('/recurring-services', editable(data, ['name', 'serviceType', 'description', 'lineItems', 'monthlyCost', 'monthlyPrice', 'targetMarginPercent', 'billingFrequency', 'isActive'])).then((r) => r.data),
  update: (id: string, data: Partial<RecurringService>) =>
    http
      .patch<ApiResponse<RecurringService>>(`/recurring-services/${id}`, editable(data, ['name', 'serviceType', 'description', 'lineItems', 'monthlyCost', 'monthlyPrice', 'targetMarginPercent', 'billingFrequency', 'isActive']))
      .then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/recurring-services/${id}`).then((r) => r.data),
}

// ─── Inventory ──────────────────────────────────────────────
export const inventoryService = {
  release: (id: string, reason: string) => http.patch<ApiResponse<StockReservation>>(`/inventory/reservations/${id}/release`, { reason }).then(r => r.data),
  fulfill: (id: string, reason: string) => http.patch<ApiResponse<StockReservation>>(`/inventory/reservations/${id}/fulfill`, { reason }).then(r => r.data),
  adjust: (data: { productId: string; warehouseLocation: string; qty: number; unitCost?: number; reason: string }) => http.post<ApiResponse<WarehouseStock>>('/inventory/movements/adjustment', data).then(r => r.data),
  setReorderLevel: (id: string, reorderLevel: number | null) => http.patch(`/inventory/stock/${id}/reorder-level`, { reorderLevel }).then(r => r.data),
  listStock: (params?: PaginationParams) =>
    http.get<ApiResponse<WarehouseStock[]>>('/inventory/stock', { params }).then((r) => r.data),
  lowStock: () =>
    http.get<ApiResponse<WarehouseStock[]>>('/inventory/stock/low-stock').then((r) => r.data),
  listReservations: () =>
    http.get<ApiResponse<StockReservation[]>>('/inventory/reservations').then((r) => r.data),
  createReservation: (data: Record<string, unknown>) =>
    http.post<ApiResponse<StockReservation>>('/inventory/reservations', data).then((r) => r.data),
  transfer: (data: Record<string, unknown>) =>
    http.post<ApiResponse<WarehouseStock>>('/inventory/movements/transfer', data).then((r) => r.data),
  listMovements: (params?: PaginationParams) =>
    http.get<ApiResponse<InventoryMovement[]>>('/inventory/movements', { params }).then((r) => r.data),
}

// ─── Procurement ────────────────────────────────────────────
export const procurementService = {
  supplierItems: (params?: {productId?:string; supplierId?:string}) => http.get<ApiResponse<import('@/types').SupplierItemEntry[]>>('/procurement/supplier-items', { params }).then(r => r.data),
  saveSupplierItem: (data: Record<string, unknown>, id?: string) => http.request({method:id ? 'PATCH' : 'POST', url:`/procurement/supplier-items${id ? '/' + id : ''}`, data:editable(data, ['supplierId', 'productId', 'latestCost', 'moq', 'leadTimeDays', 'reliability'])}).then(r => r.data),
  deleteSupplierItem: (id: string) => http.delete(`/procurement/supplier-items/${id}`),
  updatePO: (id: string, data: Record<string, unknown>) => http.patch<ApiResponse<PurchaseOrderDto>>(`/procurement/purchase-orders/${id}`, poPayload(data)).then(r => r.data),
  deletePO: (id: string) => http.delete(`/procurement/purchase-orders/${id}`),
  updatePOStatus: (id: string, status: 'pending-approval' | 'ordered' | 'cancelled') => http.patch(`/procurement/purchase-orders/${id}/status`, { status }).then(r => r.data),
  getSQ: (id: string) => http.get<ApiResponse<SupplierQuoteDto>>(`/procurement/supplier-quotes/${id}`).then(r => r.data),
  updateSQ: (id: string, data: Record<string, unknown>) => http.patch<ApiResponse<SupplierQuoteDto>>(`/procurement/supplier-quotes/${id}`, sqPayload(data)).then(r => r.data),
  deleteSQ: (id: string) => http.delete(`/procurement/supplier-quotes/${id}`),
  updateSQStatus: (id: string, status: 'under-review' | 'accepted' | 'rejected' | 'expired') => http.patch<ApiResponse<SupplierQuoteDto>>(`/procurement/supplier-quotes/${id}/status`, { status }).then(r => r.data),
  getGR: (id: string) => http.get<ApiResponse<GoodsReceiptDto>>(`/procurement/goods-receipts/${id}`).then(r => r.data),
  listPOs: (params?: PaginationParams) =>
    http.get<ApiResponse<PurchaseOrderDto[]>>('/procurement/purchase-orders', { params }).then((r) => r.data),
  getPO: (id: string) =>
    http.get<ApiResponse<PurchaseOrderDto>>(`/procurement/purchase-orders/${id}`).then((r) => r.data),
  createPO: (data: Record<string, unknown>) =>
    http.post<ApiResponse<PurchaseOrderDto>>('/procurement/purchase-orders', poPayload(data)).then((r) => r.data),
  approvePO: (id: string) =>
    http.patch<ApiResponse<PurchaseOrderDto>>(`/procurement/purchase-orders/${id}/approve`).then((r) => r.data),
  listSQs: (params?: PaginationParams) =>
    http.get<ApiResponse<SupplierQuoteDto[]>>('/procurement/supplier-quotes', { params }).then((r) => r.data),
  createSQ: (data: Record<string, unknown>) =>
    http.post<ApiResponse<SupplierQuoteDto>>('/procurement/supplier-quotes', sqPayload(data)).then((r) => r.data),
  convertToPO: (id: string) =>
    http.post<ApiResponse<PurchaseOrderDto>>(`/procurement/supplier-quotes/${id}/convert-to-po`).then((r) => r.data),
  listGRs: (params?: PaginationParams) =>
    http.get<ApiResponse<GoodsReceiptDto[]>>('/procurement/goods-receipts', { params }).then((r) => r.data),
  createGR: (data: Record<string, unknown>) =>
    http.post<ApiResponse<GoodsReceiptDto>>('/procurement/goods-receipts', grPayload(data)).then((r) => r.data),
}

// ─── Documents ──────────────────────────────────────────────
export const documentsService = {
  update: (id: string, data: Partial<CrmDocument>) => http.patch<ApiResponse<CrmDocument>>(`/documents/${id}`, editable(data, ['name', 'category', 'documentType', 'tags'])).then(r => r.data),
  versions: (id: string) => http.get<ApiResponse<DocumentVersion[]>>(`/documents/${id}/versions`).then(r => r.data),
  uploadVersion: (id: string, data: FormData) => http.post<ApiResponse<CrmDocument>>(`/documents/${id}/versions`, data).then(r => r.data),
  addLink: (id: string, entityType: string, entityId: string) => http.post(`/documents/${id}/links`, { entityType, entityId }).then(r => r.data),
  deleteLink: (id: string, linkId: string) => http.delete(`/documents/${id}/links/${linkId}`),
  download: async (id: string, fileName: string, versionId?: string) => {
    const res = await http.get<Blob>(`/documents/${id}${versionId ? '/versions/' + versionId : ''}/download`, { responseType: 'blob' })
    const url = URL.createObjectURL(res.data)
    const link = document.createElement('a'); link.href = url; link.download = fileName; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
  },
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<CrmDocument[]>>('/documents', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<CrmDocument>>(`/documents/${id}`).then((r) => r.data),
  upload: (formData: FormData) =>
    http
      .post<ApiResponse<CrmDocument>>('/documents', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/documents/${id}`).then((r) => r.data),
  downloadUrl: (id: string) => `/api/v1/documents/${id}/download`,
}

// ─── Teams ──────────────────────────────────────────────────
export const teamsService = {
  addMember: (id: string, userId: string) => http.post<ApiResponse<Team>>(`/teams/${id}/members`, { userId }).then(r => r.data),
  removeMember: (id: string, userId: string) => http.delete(`/teams/${id}/members/${userId}`),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Team[]>>('/teams', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<Team>>(`/teams/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<Team>>('/teams', editable(data, ['name', 'department', 'description', 'leaderId', 'isActive'])).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<Team>>(`/teams/${id}`, editable(data, ['name', 'department', 'description', 'leaderId', 'isActive'])).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/teams/${id}`).then((r) => r.data),
}

// ─── Projects ───────────────────────────────────────────────
export const projectsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Project[]>>('/projects', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<Project>>(`/projects/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<Project>>('/projects', editable(data, ['name', 'customerId', 'quoteId', 'status', 'priority', 'startDate', 'targetEndDate', 'actualEndDate', 'projectManagerId', 'notes'])).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<Project>>(`/projects/${id}`, editable(data, ['name', 'customerId', 'quoteId', 'status', 'priority', 'startDate', 'targetEndDate', 'actualEndDate', 'projectManagerId', 'notes'])).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/projects/${id}`).then((r) => r.data),
}

// ─── Dashboard ──────────────────────────────────────────────
export const dashboardService = {
  recentQuotes: () => http.get<ApiResponse<import('@/types').Quote[]>>('/dashboard/recent-quotes').then(r => r.data),
  expiringQuotes: () => http.get<ApiResponse<import('@/types').Quote[]>>('/dashboard/expiring-quotes').then(r => r.data),
  salesPerformance: () => http.get<ApiResponse<{ name:string; quotesCreated:number; quotesWon:number; revenue:number; winRate:number }[]>>('/dashboard/sales-performance').then(r => r.data),
  kpis: () =>
    http.get<ApiResponse<unknown>>('/dashboard/kpis').then((r) => r.data),
  pipeline: () =>
    http.get<ApiResponse<unknown>>('/dashboard/pipeline').then((r) => r.data),
  recentActivity: () =>
    http.get<ApiResponse<unknown[]>>('/dashboard/recent-activity').then((r) => r.data),
  topCustomers: () =>
    http.get<ApiResponse<unknown[]>>('/dashboard/top-customers').then((r) => r.data),
  alerts: () =>
    http.get<ApiResponse<unknown[]>>('/dashboard/alerts').then((r) => r.data),
}

// ─── Users ──────────────────────────────────────────────────
export const usersService = {
  resetPassword: (id: string, password: string) => http.patch(`/users/${id}/password`, { password }),
  updateProfile: (data: { firstName?: string; lastName?: string; email?: string; phone?: string }) => http.patch('/auth/me', data).then(r => r.data),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/users', { params }).then((r) => r.data),
  lookup: (role?: string, department?: string) =>
    http.get<ApiResponse<UserLookup[]>>('/users/lookup', { params: { role, department } }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<unknown>>(`/users/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/users', editable(data, ['firstName', 'lastName', 'email', 'password', 'role', 'department', 'phone'])).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<unknown>>(`/users/${id}`, editable(data, ['firstName', 'lastName', 'email', 'role', 'department', 'phone', 'teamId', 'isActive'])).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/users/${id}`).then((r) => r.data),
}

// Re-export individual services
export { authService } from './auth.service'
export { customersService } from './customers.service'
export { opportunitiesService } from './opportunities.service'
export { productsService } from './products.service'
export { quotesService } from './quotes.service'
export { manufacturersService } from './manufacturers.service'
export { http, tokenStorage } from './http'

function linePayloads(data: Record<string, unknown>, fields: string[]) {
 return Array.isArray(data.items) ? data.items.map(item => editable(item, fields)) : []
}
export function poPayload(data: Record<string, unknown>) {
 return { ...editable(data, ['supplierId', 'supplierName', 'currency', 'shippingCost', 'customsDuty', 'expectedDelivery', 'sourceQuoteId', 'notes']), items: linePayloads(data, ['productId', 'quantity', 'unitCost', 'leadTimeDays']) }
}
export function sqPayload(data: Record<string, unknown>) {
 return { ...editable(data, ['supplierId', 'supplierName', 'supplierRef', 'currency', 'validFrom', 'validUntil', 'contactName', 'contactEmail', 'paymentTerms', 'deliveryTerms', 'notes']), items: linePayloads(data, ['productId', 'productSku', 'productName', 'manufacturerName', 'quantity', 'unitCost', 'leadTimeDays', 'moq', 'validUntil', 'notes']) }
}
export function grPayload(data: Record<string, unknown>) {
 return { ...editable(data, ['poId', 'receiveDate', 'notes']), items: linePayloads(data, ['poItemId', 'productId', 'receivedQty', 'storageLocation', 'condition', 'notes']) }
}
export const catalogServicesService = {
 list: (params?: PaginationParams) => http.get<ApiResponse<ServiceItem[]>>('/services', { params }).then(r => r.data),
 get: (id: string) => http.get<ApiResponse<ServiceItem>>(`/services/${id}`).then(r => r.data),
 create: (data: Omit<ServiceItem, 'id'>) => http.post<ApiResponse<ServiceItem>>('/services', data).then(r => r.data),
 update: (id: string, data: Partial<Omit<ServiceItem, 'id'>>) => http.patch<ApiResponse<ServiceItem>>(`/services/${id}`, data).then(r => r.data),
 delete: (id: string) => http.delete(`/services/${id}`),
}
