import { http, type ApiResponse, type PaginationParams } from './http'
import type { Quote } from '@/types'

export const quotesService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Quote[]>>('/quotes', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Quote>>(`/quotes/${id}`).then((r) => r.data),

  create: (data: Partial<Quote>) =>
    http.post<ApiResponse<Quote>>('/quotes', data).then((r) => r.data),

  update: (id: string, data: Partial<Quote>) =>
    http.patch<ApiResponse<Quote>>(`/quotes/${id}`, data).then((r) => r.data),

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
