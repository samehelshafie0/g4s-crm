import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { recurringServicesService } from '@/services'
import type { RecurringService } from '@/types'

export const useRecurringServicesStore = defineStore('recurringServices', () => {
  const services = ref<RecurringService[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeServices = computed(() => services.value.filter((s) => s.isActive))
  const totalMonthlyRevenue = computed(() =>
    activeServices.value.reduce((sum, s) => sum + (s.monthlyPrice ?? 0), 0),
  )

  async function fetchServices(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await recurringServicesService.list(params)
      if (res.success) services.value = res.data
    } catch (e) { error.value = 'Failed to load services'; console.error(e) }
    finally { loading.value = false }
  }

  async function addService(data: Partial<RecurringService>) {
    const res = await recurringServicesService.create(data)
    if (res.success) { services.value.unshift(res.data); return res.data }
    throw new Error('Failed to create')
  }

  async function updateService(id: string, data: Partial<RecurringService>) {
    const res = await recurringServicesService.update(id, data)
    if (res.success) {
      const idx = services.value.findIndex((s) => s.id === id)
      if (idx !== -1) services.value[idx] = res.data
    }
  }

  async function deleteService(id: string) {
    await recurringServicesService.delete(id)
    services.value = services.value.filter((s) => s.id !== id)
  }

  return { services, loading, error, activeServices, totalMonthlyRevenue, fetchServices, addService, updateService, deleteService }
})
