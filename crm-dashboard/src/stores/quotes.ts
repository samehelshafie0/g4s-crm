import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { quotesService } from '@/services'
import type { Quote } from '@/types'

export const useQuotesStore = defineStore('quotes', () => {
  const quotes = ref<Quote[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  const draftQuotes = computed(() => quotes.value.filter((q) => q.status === 'draft'))
  const activeQuotes = computed(() =>
    quotes.value.filter((q) => ['approved', 'sent', 'pending-approval'].includes(q.status)),
  )

  async function fetchQuotes(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await quotesService.list(params)
      if (res.success) {
        quotes.value = res.data
        if (res.meta) meta.value = res.meta
      }
    } catch (e) {
      error.value = 'Failed to load quotes'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getByStatus(status: string) {
    return quotes.value.filter((q) => q.status === status)
  }

  function getByCustomer(customerId: string) {
    return quotes.value.filter((q) => q.customerId === customerId)
  }

  async function addQuote(data: Partial<Quote>) {
    const res = await quotesService.create(data)
    if (res.success) {
      quotes.value.unshift(res.data)
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to create quote')
  }

  async function updateQuote(id: string, data: Partial<Quote>) {
    const res = await quotesService.update(id, data)
    if (res.success) {
      const idx = quotes.value.findIndex((q) => q.id === id)
      if (idx !== -1) quotes.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update quote')
  }

  async function deleteQuote(id: string) {
    await quotesService.delete(id)
    quotes.value = quotes.value.filter((q) => q.id !== id)
  }

  async function submitQuote(id: string) {
    const res = await quotesService.submit(id)
    if (res.success) {
      const idx = quotes.value.findIndex((q) => q.id === id)
      if (idx !== -1) quotes.value[idx] = res.data
    }
  }

  async function approveQuote(id: string) {
    const res = await quotesService.approve(id)
    if (res.success) {
      const idx = quotes.value.findIndex((q) => q.id === id)
      if (idx !== -1) quotes.value[idx] = res.data
    }
  }

  return {
    quotes,
    loading,
    error,
    meta,
    draftQuotes,
    activeQuotes,
    fetchQuotes,
    getByStatus,
    getByCustomer,
    addQuote,
    updateQuote,
    deleteQuote,
    submitQuote,
    approveQuote,
  }
})
