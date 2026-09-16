import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { contractsService } from '@/services'
import type { Contract } from '@/types'

export const useContractsStore = defineStore('contracts', () => {
  const contracts = ref<Contract[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  const activeContracts = computed(() => contracts.value.filter((c) => c.status === 'active'))
  const expiringContracts = computed(() => {
    const soon = new Date()
    soon.setDate(soon.getDate() + 30)
    return contracts.value.filter(
      (c) => c.status === 'active' && c.endDate && new Date(c.endDate) <= soon,
    )
  })

  async function fetchContracts(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await contractsService.list(params)
      if (res.success) {
        contracts.value = res.data
        if (res.meta) meta.value = res.meta
      }
    } catch (e) {
      error.value = 'Failed to load contracts'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getByCustomer(customerId: string) {
    return contracts.value.filter((c) => c.customerId === customerId)
  }

  function getByStatus(status: string) {
    return contracts.value.filter((c) => c.status === status)
  }

  async function addContract(data: Partial<Contract>) {
    const res = await contractsService.create(data)
    if (res.success) { contracts.value.unshift(res.data); return res.data }
    throw new Error(res.error?.message ?? 'Failed to create')
  }

  async function updateContract(id: string, data: Partial<Contract>) {
    const res = await contractsService.update(id, data)
    if (res.success) {
      const idx = contracts.value.findIndex((c) => c.id === id)
      if (idx !== -1) contracts.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update')
  }

  return {
    contracts,
    loading,
    error,
    meta,
    activeContracts,
    expiringContracts,
    fetchContracts,
    getByCustomer,
    getByStatus,
    addContract,
    updateContract,
  }
})
