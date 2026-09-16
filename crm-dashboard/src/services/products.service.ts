import { http, type ApiResponse, type PaginationParams } from './http'
import type { Product } from '@/types'

export const productsService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Product[]>>('/products', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Product>>(`/products/${id}`).then((r) => r.data),

  create: (data: Partial<Product>) =>
    http.post<ApiResponse<Product>>('/products', data).then((r) => r.data),

  update: (id: string, data: Partial<Product>) =>
    http.patch<ApiResponse<Product>>(`/products/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/products/${id}`).then((r) => r.data),

  recalculateCosts: () =>
    http.post<ApiResponse<{ updated: number }>>('/products/recalculate-costs').then((r) => r.data),
}
