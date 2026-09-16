import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { customersService } from '@/services'
import type { Customer } from '@/types'

export const useCustomersStore = defineStore('customers', () => {
  const customers = ref<Customer[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  const activeCustomers = computed(() => customers.value.filter((c) => c.status === 'active'))
  const prospectCustomers = computed(() => customers.value.filter((c) => c.status === 'prospect'))

  async function fetchCustomers(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await customersService.list(params)
      if (res.success) {
        customers.value = res.data
        if (res.meta) meta.value = res.meta
      }
    } catch (e: unknown) {
      error.value = 'Failed to load customers'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getCustomerById(id: string) {
    return customers.value.find((c) => c.id === id)
  }

  async function addCustomer(data: Partial<Customer>) {
    const res = await customersService.create(data)
    if (res.success) {
      customers.value.unshift(res.data)
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to create customer')
  }

  async function updateCustomer(id: string, data: Partial<Customer>) {
    const res = await customersService.update(id, data)
    if (res.success) {
      const idx = customers.value.findIndex((c) => c.id === id)
      if (idx !== -1) customers.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update customer')
  }

  async function deleteCustomer(id: string) {
    await customersService.delete(id)
    customers.value = customers.value.filter((c) => c.id !== id)
  }

  return {
    customers,
    loading,
    error,
    meta,
    activeCustomers,
    prospectCustomers,
    fetchCustomers,
    getCustomerById,
    addCustomer,
    updateCustomer,
    deleteCustomer,
  }
})
