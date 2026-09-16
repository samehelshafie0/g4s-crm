import { http, type ApiResponse, type PaginationParams } from './http'
import type { Customer, CustomerSite, CustomerContact } from '@/types'

// Send only editable header fields; contacts and sites have their own endpoints.
function customerPayload(data: Partial<Customer>) {
  const { companyName, sector, region, status, type, crNumber, vatNumber, notes } = data
  return { companyName, sector, region, status, type, crNumber, vatNumber, notes }
}

export const customersService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Customer[]>>('/customers', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Customer>>(`/customers/${id}`).then((r) => r.data),

  create: (data: Partial<Customer>) =>
    http.post<ApiResponse<Customer>>('/customers', customerPayload(data)).then((r) => r.data),

  update: (id: string, data: Partial<Customer>) =>
    http.patch<ApiResponse<Customer>>(`/customers/${id}`, customerPayload(data)).then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/customers/${id}`).then((r) => r.data),

  // Sites
  addSite: (customerId: string, data: Partial<CustomerSite>) =>
    http.post<ApiResponse<CustomerSite>>(`/customers/${customerId}/sites`, data).then((r) => r.data),

  updateSite: (customerId: string, siteId: string, data: Partial<CustomerSite>) =>
    http
      .patch<ApiResponse<CustomerSite>>(`/customers/${customerId}/sites/${siteId}`, data)
      .then((r) => r.data),

  deleteSite: (customerId: string, siteId: string) =>
    http.delete(`/customers/${customerId}/sites/${siteId}`),

  // Contacts
  addContact: (customerId: string, data: Partial<CustomerContact>) =>
    http
      .post<ApiResponse<CustomerContact>>(`/customers/${customerId}/contacts`, data)
      .then((r) => r.data),

  updateContact: (customerId: string, contactId: string, data: Partial<CustomerContact>) =>
    http
      .patch<ApiResponse<CustomerContact>>(
        `/customers/${customerId}/contacts/${contactId}`,
        data,
      )
      .then((r) => r.data),

  deleteContact: (customerId: string, contactId: string) =>
    http.delete(`/customers/${customerId}/contacts/${contactId}`),
}
