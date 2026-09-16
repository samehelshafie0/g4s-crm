import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { exchangeRatesService } from '@/services'
import type { ExchangeRate } from '@/types'

export const useExchangeRatesStore = defineStore('exchangeRates', () => {
  const exchangeRates = ref<ExchangeRate[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const rateMap = computed(() => {
    const map: Record<string, number> = {}
    exchangeRates.value.forEach((r) => { map[r.fromCurrency] = r.currentRate })
    return map
  })

  async function fetchRates() {
    loading.value = true
    try {
      const res = await exchangeRatesService.list()
      if (res.success) exchangeRates.value = res.data
    } catch (e) { error.value = 'Failed to load exchange rates'; console.error(e) }
    finally { loading.value = false }
  }

  function getRate(currency: string): number {
    return rateMap.value[currency] ?? 1
  }

  async function updateRate(id: string, currentRate: number, effectiveDate?: string) {
    const res = await exchangeRatesService.update(id, currentRate, effectiveDate)
    if (res.success) {
      const idx = exchangeRates.value.findIndex((r) => r.id === id)
      if (idx !== -1) exchangeRates.value[idx] = res.data
    }
  }

  return { exchangeRates, loading, error, rateMap, fetchRates, getRate, updateRate }
})
