import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { manufacturersService } from '@/services'
import type { Manufacturer } from '@/types'

export const useManufacturersStore = defineStore('manufacturers', () => {
  const manufacturers = ref<Manufacturer[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeManufacturers = computed(() => manufacturers.value.filter((m) => m.isActive))
  const suppliers = computed(() =>
    manufacturers.value.filter((m) => m.vendorType === 'supplier' || m.vendorType === 'both'),
  )
  const pureManufacturers = computed(() =>
    manufacturers.value.filter((m) => m.vendorType === 'manufacturer' || m.vendorType === 'both'),
  )

  async function fetchManufacturers(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await manufacturersService.list(params)
      if (res.success) manufacturers.value = res.data
    } catch (e) {
      error.value = 'Failed to load manufacturers'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getCategoriesByManufacturer(id: string) {
    return manufacturers.value.find((m) => m.id === id)?.categories ?? []
  }

  async function addManufacturer(data: Partial<Manufacturer>) {
    const res = await manufacturersService.create(data)
    if (res.success) { manufacturers.value.unshift(res.data); return res.data }
    throw new Error(res.error?.message ?? 'Failed to create')
  }

  async function updateManufacturer(id: string, data: Partial<Manufacturer>) {
    const res = await manufacturersService.update(id, data)
    if (res.success) {
      const idx = manufacturers.value.findIndex((m) => m.id === id)
      if (idx !== -1) manufacturers.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update')
  }

  async function deleteManufacturer(id: string) {
    await manufacturersService.delete(id)
    manufacturers.value = manufacturers.value.filter((m) => m.id !== id)
  }

  return {
    manufacturers,
    loading,
    error,
    activeManufacturers,
    suppliers,
    pureManufacturers,
    fetchManufacturers,
    getCategoriesByManufacturer,
    addManufacturer,
    updateManufacturer,
    deleteManufacturer,
  }
})
