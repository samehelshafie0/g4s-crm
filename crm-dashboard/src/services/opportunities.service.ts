import { editable } from './payload'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Opportunity } from '@/types'

export const opportunitiesService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Opportunity[]>>('/opportunities', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Opportunity>>(`/opportunities/${id}`).then((r) => r.data),

  pipeline: () =>
    http
      .get<ApiResponse<{ stage: string; count: number; value: number }[]>>('/opportunities/pipeline')
      .then((r) => r.data),

  create: (data: Partial<Opportunity>) =>
    http.post<ApiResponse<Opportunity>>('/opportunities', editable(data, ['title', 'customerId', 'stage', 'serviceTypes', 'estimatedValue', 'estimatedCost', 'winProbability', 'salesExecutiveId', 'preSalesId', 'expectedCloseDate', 'notes'])).then((r) => r.data),

  update: (id: string, data: Partial<Opportunity>) =>
    http.patch<ApiResponse<Opportunity>>(`/opportunities/${id}`, editable(data, ['title', 'customerId', 'stage', 'serviceTypes', 'estimatedValue', 'estimatedCost', 'winProbability', 'salesExecutiveId', 'preSalesId', 'expectedCloseDate', 'notes'])).then((r) => r.data),

  updateStage: (id: string, stage: string, notes?: string) =>
    http
      .patch<ApiResponse<Opportunity>>(`/opportunities/${id}/stage`, { stage, notes })
      .then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/opportunities/${id}`).then((r) => r.data),
}
