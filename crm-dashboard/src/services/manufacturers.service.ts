import { editable } from './payload'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Manufacturer, ManufacturerCategory } from '@/types'

export const manufacturersService = {
  categories: (id: string) => http.get<ApiResponse<ManufacturerCategory[]>>(`/manufacturers/${id}/categories`).then(r => r.data),
  addCategory: (id: string, data: Partial<ManufacturerCategory>) => http.post<ApiResponse<ManufacturerCategory>>(`/manufacturers/${id}/categories`, editable(data, ['name', 'description'])).then(r => r.data),
  updateCategory: (id: string, categoryId: string, data: Partial<ManufacturerCategory>) => http.patch<ApiResponse<ManufacturerCategory>>(`/manufacturers/${id}/categories/${categoryId}`, editable(data, ['name', 'description'])).then(r => r.data),
  deleteCategory: (id: string, categoryId: string) => http.delete(`/manufacturers/${id}/categories/${categoryId}`),
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Manufacturer[]>>('/manufacturers', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Manufacturer>>(`/manufacturers/${id}`).then((r) => r.data),

  create: (data: Partial<Manufacturer>) =>
    http.post<ApiResponse<Manufacturer>>('/manufacturers', editable(data, ['name', 'code', 'country', 'contactEmail', 'contactPhone', 'website', 'vendorType', 'isActive'])).then((r) => r.data),

  update: (id: string, data: Partial<Manufacturer>) =>
    http.patch<ApiResponse<Manufacturer>>(`/manufacturers/${id}`, editable(data, ['name', 'code', 'country', 'contactEmail', 'contactPhone', 'website', 'vendorType', 'isActive'])).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/manufacturers/${id}`).then((r) => r.data),
}
