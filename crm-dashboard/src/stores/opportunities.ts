import { allPages } from '@/services/collections'
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { opportunitiesService } from '@/services'
import type { Opportunity } from '@/types'

export const useOpportunitiesStore = defineStore('opportunities', () => {
  const opportunities = ref<Opportunity[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  const activeOpportunities = computed(() =>
    opportunities.value.filter(
      (o) => o.stage !== 'closed-won' && o.stage !== 'closed-lost',
    ),
  )

  const pipelineValue = computed(() =>
    activeOpportunities.value.reduce((sum, o) => sum + (o.estimatedValue ?? 0), 0),
  )

  const wonDeals = computed(() => opportunities.value.filter((o) => o.stage === 'closed-won'))

  const wonDealsValue = computed(() =>
    wonDeals.value.reduce((sum, o) => sum + (o.estimatedValue ?? 0), 0),
  )

  async function fetchOpportunities(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      opportunities.value = await allPages(opportunitiesService.list, params)
      meta.value = {page: 1, limit: opportunities.value.length, total: opportunities.value.length, totalPages: 1}
    } catch (e) {
      error.value = 'Failed to load opportunities'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getByStage(stage: string) {
    return opportunities.value.filter((o) => o.stage === stage)
  }

  function getByCustomer(customerId: string) {
    return opportunities.value.filter((o) => o.customerId === customerId)
  }

  async function addOpportunity(data: Partial<Opportunity>) {
    const res = await opportunitiesService.create(data)
    if (res.success) {
      opportunities.value.unshift(res.data)
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to create opportunity')
  }

  async function updateOpportunity(id: string, data: Partial<Opportunity>) {
    const res = await opportunitiesService.update(id, data)
    if (res.success) {
      const idx = opportunities.value.findIndex((o) => o.id === id)
      if (idx !== -1) opportunities.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update opportunity')
  }

  async function updateStage(id: string, stage: string, notes?: string) {
    const res = await opportunitiesService.updateStage(id, stage, notes)
    if (res.success) {
      const idx = opportunities.value.findIndex((o) => o.id === id)
      if (idx !== -1) opportunities.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update stage')
  }

  async function deleteOpportunity(id: string) {
    await opportunitiesService.delete(id)
    opportunities.value = opportunities.value.filter((o) => o.id !== id)
  }

  return {
    opportunities,
    loading,
    error,
    meta,
    activeOpportunities,
    pipelineValue,
    wonDeals,
    wonDealsValue,
    fetchOpportunities,
    getByStage,
    getByCustomer,
    addOpportunity,
    updateOpportunity,
    updateStage,
    deleteOpportunity,
  }
})
