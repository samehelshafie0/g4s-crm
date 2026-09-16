import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { priceBooksService } from '@/services'
import type { PriceBook } from '@/types'

export const usePriceBooksStore = defineStore('priceBooks', () => {
  const priceBooks = ref<PriceBook[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activePriceBooks = computed(() => priceBooks.value.filter((p) => p.isActive))

  async function fetchPriceBooks(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await priceBooksService.list(params)
      if (res.success) priceBooks.value = res.data
    } catch (e) { error.value = 'Failed to load price books'; console.error(e) }
    finally { loading.value = false }
  }

  function getByCustomer(customerId: string) {
    return priceBooks.value.filter((p) => p.customerId === customerId || !p.customerId)
  }

  async function addPriceBook(data: Partial<PriceBook>) {
    const res = await priceBooksService.create(data)
    if (res.success) { priceBooks.value.unshift(res.data); return res.data }
    throw new Error('Failed to create')
  }

  async function updatePriceBook(id: string, data: Partial<PriceBook>) {
    const res = await priceBooksService.update(id, data)
    if (res.success) {
      const idx = priceBooks.value.findIndex((p) => p.id === id)
      if (idx !== -1) priceBooks.value[idx] = res.data
    }
  }

  return { priceBooks, loading, error, activePriceBooks, fetchPriceBooks, getByCustomer, addPriceBook, updatePriceBook }
})
