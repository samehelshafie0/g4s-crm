import { http, type ApiResponse, type PaginationParams } from './http'
import type { Manufacturer } from '@/types'

export const manufacturersService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Manufacturer[]>>('/manufacturers', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Manufacturer>>(`/manufacturers/${id}`).then((r) => r.data),

  create: (data: Partial<Manufacturer>) =>
    http.post<ApiResponse<Manufacturer>>('/manufacturers', data).then((r) => r.data),

  update: (id: string, data: Partial<Manufacturer>) =>
    http.patch<ApiResponse<Manufacturer>>(`/manufacturers/${id}`, data).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/manufacturers/${id}`).then((r) => r.data),
}
