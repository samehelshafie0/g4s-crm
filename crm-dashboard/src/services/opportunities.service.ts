import { editable } from './payload'
import { http, type ApiResponse, type PaginationParams } from './http'
import type { Opportunity, Fulfilment, DeliveryStatus } from '@/types'
import type { Activity } from './workflowDtos'

export interface DeliveryInput {
  deliveryStatus: DeliveryStatus
  projectManagerId?: string | null
  startDate?: string | null
  targetEndDate?: string | null
  actualEndDate?: string | null
  actualValue?: number | null
  actualCost?: number | null
}

export type OpportunityInput = Partial<Opportunity> & { siteIds?: string[] }

export const opportunitiesService = {
  list: (params?: PaginationParams) =>
    http.get<ApiResponse<Opportunity[]>>('/opportunities', { params }).then((r) => r.data),

  get: (id: string) =>
    http.get<ApiResponse<Opportunity>>(`/opportunities/${id}`).then((r) => r.data),

  pipeline: () =>
    http
      .get<ApiResponse<{ stage: string; count: number; value: number }[]>>('/opportunities/pipeline')
      .then((r) => r.data),

  create: (data: OpportunityInput) =>
    http.post<ApiResponse<Opportunity>>('/opportunities', editable(data, ['title', 'customerId', 'stage', 'serviceTypes', 'estimatedValue', 'estimatedCost', 'winProbability', 'salesExecutiveId', 'preSalesId', 'expectedCloseDate', 'notes', 'siteIds'])).then((r) => r.data),

  update: (id: string, data: OpportunityInput) =>
    http.patch<ApiResponse<Opportunity>>(`/opportunities/${id}`, editable(data, ['title', 'customerId', 'stage', 'serviceTypes', 'estimatedValue', 'estimatedCost', 'winProbability', 'salesExecutiveId', 'preSalesId', 'expectedCloseDate', 'notes', 'siteIds'])).then((r) => r.data),

  updateStage: (id: string, stage: string, notes?: string) =>
    http
      .patch<ApiResponse<Opportunity>>(`/opportunities/${id}/stage`, { stage, notes })
      .then((r) => r.data),

  delete: (id: string) =>
    http.delete<ApiResponse<null>>(`/opportunities/${id}`).then((r) => r.data),

  fulfilment: (id: string) =>
    http.get<ApiResponse<Fulfilment>>(`/opportunities/${id}/fulfilment`).then((r) => r.data),

  updateDelivery: (id: string, data: DeliveryInput) =>
    http.patch<ApiResponse<Opportunity>>(`/opportunities/${id}/delivery`, data).then((r) => r.data),

  activity: (id: string) =>
    http.get<ApiResponse<Activity[]>>(`/opportunities/${id}/activity`).then((r) => r.data),

  delivery: (params?: PaginationParams & Record<string, unknown>) =>
    http.get<ApiResponse<Opportunity[]>>('/delivery', { params }).then((r) => r.data),
}
