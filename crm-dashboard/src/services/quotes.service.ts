import type { BuilderQuote, BuilderInput, Activity } from './workflowDtos'
import { editable } from './payload'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Quote } from '@/types'

export const quotesService = {
  activity: (id: string) => http.get<ApiResponse<Activity[]>>(`/quotes/${id}/activity`).then(r => r.data),
  builder: (id: string) => http.get<ApiResponse<BuilderQuote>>(`/quotes/${id}/builder`).then(r => r.data),
  saveBuilder: (id: string, data: BuilderInput) => http.put<ApiResponse<BuilderQuote>>(`/quotes/${id}/builder`, data).then(r => r.data),
  duplicate: (id: string) => http.post<ApiResponse<BuilderQuote>>(`/quotes/${id}/duplicate`).then(r => r.data),
  revise: (id: string) => http.post<ApiResponse<BuilderQuote>>(`/quotes/${id}/revisions`).then(r => r.data),
  convertToContract: (id: string, data: { title: string; type: string; startDate: string; endDate: string; terms?: string }) => http.post<ApiResponse<import('@/types').Contract>>(`/quotes/${id}/convert-to-contract`, data).then(r => r.data),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Quote[]>>('/quotes', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Quote>>(`/quotes/${id}`).then((r) => r.data),

  create: (data: Partial<Quote>) =>
    http.post<ApiResponse<Quote>>('/quotes', editable(data, ['priceBookId', 'customerId', 'opportunityId', 'currency', 'validUntil', 'discountPercent', 'vatPercent', 'notes', 'lineItems'])).then((r) => r.data),

  update: (id: string, data: Partial<Quote>) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}`, editable(data, ['discountPercent', 'vatPercent', 'validUntil', 'notes'])).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/quotes/${id}`).then((r) => r.data),

  submit: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/submit`).then((r) => r.data),

  approve: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/approve`).then((r) => r.data),

  reject: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/reject`).then((r) => r.data),

  send: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/send`).then((r) => r.data),

  accept: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/accept`).then((r) => r.data),

  decline: (id: string) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}/decline`).then((r) => r.data),

  recalculate: (id: string) =>
    http.post<ApiResponse<Quote>>(`/quotes/${id}/recalculate`).then((r) => r.data),
}
