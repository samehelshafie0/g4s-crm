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
    http.post<ApiResponse<Contract>>('/contracts', data).then((r) => r.data),
  update: (id: string, data: Partial<Contract>) =>
    http.patch<ApiResponse<Contract>>(`/contracts/${id}`, data).then((r) => r.data),
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
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<PriceBook[]>>('/price-books', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<PriceBook>>(`/price-books/${id}`).then((r) => r.data),
  create: (data: Partial<PriceBook>) =>
    http.post<ApiResponse<PriceBook>>('/price-books', data).then((r) => r.data),
  update: (id: string, data: Partial<PriceBook>) =>
    http.patch<ApiResponse<PriceBook>>(`/price-books/${id}`, data).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/price-books/${id}`).then((r) => r.data),
}

// ─── Exchange Rates ─────────────────────────────────────────
export const exchangeRatesService = {
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
    http.post<ApiResponse<RecurringService>>('/recurring-services', data).then((r) => r.data),
  update: (id: string, data: Partial<RecurringService>) =>
    http
      .patch<ApiResponse<RecurringService>>(`/recurring-services/${id}`, data)
      .then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/recurring-services/${id}`).then((r) => r.data),
}

// ─── Inventory ──────────────────────────────────────────────
export const inventoryService = {
  listStock: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/inventory/stock', { params }).then((r) => r.data),
  lowStock: () =>
    http.get<ApiResponse<unknown[]>>('/inventory/stock/low-stock').then((r) => r.data),
  listReservations: () =>
    http.get<ApiResponse<unknown[]>>('/inventory/reservations').then((r) => r.data),
  createReservation: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/inventory/reservations', data).then((r) => r.data),
  transfer: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/inventory/movements/transfer', data).then((r) => r.data),
  listMovements: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/inventory/movements', { params }).then((r) => r.data),
}

// ─── Procurement ────────────────────────────────────────────
export const procurementService = {
  listPOs: (params?: PaginationParams) =>
    http.get<ApiResponse<PurchaseOrderDto[]>>('/procurement/purchase-orders', { params }).then((r) => r.data),
  getPO: (id: string) =>
    http.get<ApiResponse<PurchaseOrderDto>>(`/procurement/purchase-orders/${id}`).then((r) => r.data),
  createPO: (data: Record<string, unknown>) =>
    http.post<ApiResponse<PurchaseOrderDto>>('/procurement/purchase-orders', data).then((r) => r.data),
  approvePO: (id: string) =>
    http.patch<ApiResponse<PurchaseOrderDto>>(`/procurement/purchase-orders/${id}/approve`).then((r) => r.data),
  listSQs: (params?: PaginationParams) =>
    http.get<ApiResponse<SupplierQuoteDto[]>>('/procurement/supplier-quotes', { params }).then((r) => r.data),
  createSQ: (data: Record<string, unknown>) =>
    http.post<ApiResponse<SupplierQuoteDto>>('/procurement/supplier-quotes', data).then((r) => r.data),
  convertToPO: (id: string) =>
    http.post<ApiResponse<PurchaseOrderDto>>(`/procurement/supplier-quotes/${id}/convert-to-po`).then((r) => r.data),
  listGRs: (params?: PaginationParams) =>
    http.get<ApiResponse<GoodsReceiptDto[]>>('/procurement/goods-receipts', { params }).then((r) => r.data),
  createGR: (data: Record<string, unknown>) =>
    http.post<ApiResponse<GoodsReceiptDto>>('/procurement/goods-receipts', data).then((r) => r.data),
}

// ─── Documents ──────────────────────────────────────────────
export const documentsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/documents', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<unknown>>(`/documents/${id}`).then((r) => r.data),
  upload: (formData: FormData) =>
    http
      .post<ApiResponse<unknown>>('/documents', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      .then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/documents/${id}`).then((r) => r.data),
  downloadUrl: (id: string) => `/api/v1/documents/${id}/download`,
}

// ─── Teams ──────────────────────────────────────────────────
export const teamsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/teams', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<unknown>>(`/teams/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/teams', data).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<unknown>>(`/teams/${id}`, data).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/teams/${id}`).then((r) => r.data),
}

// ─── Projects ───────────────────────────────────────────────
export const projectsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/projects', { params }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<unknown>>(`/projects/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/projects', data).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<unknown>>(`/projects/${id}`, data).then((r) => r.data),
  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/projects/${id}`).then((r) => r.data),
}

// ─── Dashboard ──────────────────────────────────────────────
export const dashboardService = {
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
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<unknown[]>>('/users', { params }).then((r) => r.data),
  lookup: (role?: string, department?: string) =>
    http.get<ApiResponse<unknown[]>>('/users/lookup', { params: { role, department } }).then((r) => r.data),
  get: (id: string) =>
    http.get<ApiResponse<unknown>>(`/users/${id}`).then((r) => r.data),
  create: (data: Record<string, unknown>) =>
    http.post<ApiResponse<unknown>>('/users', data).then((r) => r.data),
  update: (id: string, data: Record<string, unknown>) =>
    http.patch<ApiResponse<unknown>>(`/users/${id}`, data).then((r) => r.data),
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
