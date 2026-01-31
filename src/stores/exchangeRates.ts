import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { ExchangeRate, Currency } from '@/types'

export const useExchangeRatesStore = defineStore('exchangeRates', () => {
  // Current active rates
  const exchangeRates = ref<ExchangeRate[]>([
    {
      id: 'ER-001',
      currency: 'USD',
      rate: 3.75,
      effectiveDate: '2024-01-01',
      setBy: 'USR-001',
      setAt: '2024-01-01T08:00:00Z',
      notes: 'Initial rate for 2024'
    },
    {
      id: 'ER-002',
      currency: 'EUR',
      rate: 4.12,
      effectiveDate: '2024-01-01',
      setBy: 'USR-001',
      setAt: '2024-01-01T08:00:00Z',
      notes: 'Initial rate for 2024'
    },
    {
      id: 'ER-003',
      currency: 'GBP',
      rate: 4.75,
      effectiveDate: '2024-01-01',
      setBy: 'USR-001',
      setAt: '2024-01-01T08:00:00Z',
      notes: 'Initial rate for 2024'
    },
    {
      id: 'ER-004',
      currency: 'SAR',
      rate: 1.0,
      effectiveDate: '2024-01-01',
      setBy: 'USR-001',
      setAt: '2024-01-01T08:00:00Z',
      notes: 'Base currency'
    }
  ])

  // Historical rates (all previous rates)
  const rateHistory = ref<ExchangeRate[]>([
    {
      id: 'ER-H001',
      currency: 'USD',
      rate: 3.74,
      effectiveDate: '2023-12-01',
      setBy: 'USR-001',
      setAt: '2023-12-01T08:00:00Z',
      notes: 'December 2023 rate'
    },
    {
      id: 'ER-H002',
      currency: 'EUR',
      rate: 4.08,
      effectiveDate: '2023-12-01',
      setBy: 'USR-001',
      setAt: '2023-12-01T08:00:00Z',
      notes: 'December 2023 rate'
    },
    {
      id: 'ER-H003',
      currency: 'USD',
      rate: 3.76,
      effectiveDate: '2023-11-01',
      setBy: 'USR-001',
      setAt: '2023-11-01T08:00:00Z',
      notes: 'November 2023 rate'
    }
  ])

  // Getters
  const getCurrentRate = (currency: Currency): number => {
    const rate = exchangeRates.value.find(r => r.currency === currency)
    return rate?.rate || 1.0
  }

  const getRateById = (id: string): ExchangeRate | undefined => {
    return exchangeRates.value.find(r => r.id === id) ||
           rateHistory.value.find(r => r.id === id)
  }

  const getHistoryByCurrency = (currency: Currency): ExchangeRate[] => {
    const allRates = [...rateHistory.value, ...exchangeRates.value]
    return allRates
      .filter(r => r.currency === currency)
      .sort((a, b) => new Date(b.effectiveDate).getTime() - new Date(a.effectiveDate).getTime())
  }

  const getAllCurrencies = computed(() => {
    const currencies = new Set<Currency>()
    exchangeRates.value.forEach(r => currencies.add(r.currency))
    return Array.from(currencies)
  })

  // Actions
  function updateRate(currency: Currency, newRate: number, effectiveDate: string, setBy: string, notes?: string) {
    // Find current rate
    const currentRate = exchangeRates.value.find(r => r.currency === currency)

    if (currentRate) {
      // Move current rate to history
      rateHistory.value.push({ ...currentRate })

      // Update current rate
      currentRate.rate = newRate
      currentRate.effectiveDate = effectiveDate
      currentRate.setBy = setBy
      currentRate.setAt = new Date().toISOString()
      currentRate.notes = notes
    } else {
      // Add new currency rate
      const newRateEntry: ExchangeRate = {
        id: `ER-${String(exchangeRates.value.length + 1).padStart(3, '0')}`,
        currency,
        rate: newRate,
        effectiveDate,
        setBy,
        setAt: new Date().toISOString(),
        notes
      }
      exchangeRates.value.push(newRateEntry)
    }
  }

  function addCurrency(currency: Currency, rate: number, effectiveDate: string, setBy: string, notes?: string) {
    const exists = exchangeRates.value.some(r => r.currency === currency)
    if (exists) {
      throw new Error(`Currency ${currency} already exists`)
    }

    const newRate: ExchangeRate = {
      id: `ER-${String(exchangeRates.value.length + 1).padStart(3, '0')}`,
      currency,
      rate,
      effectiveDate,
      setBy,
      setAt: new Date().toISOString(),
      notes
    }
    exchangeRates.value.push(newRate)
  }

  // Fetch live rates from API
  async function fetchLiveRates(): Promise<{ [key: string]: number } | null> {
    try {
      // Using exchangerate-api.com (free tier available)
      // Base currency is SAR (Saudi Riyal)
      const response = await fetch('https://api.exchangerate-api.com/v4/latest/SAR')

      if (!response.ok) {
        throw new Error('Failed to fetch exchange rates')
      }

      const data = await response.json()

      // The API returns rates from SAR to other currencies
      // We need to invert them to get "1 unit of currency = X SAR"
      const rates: { [key: string]: number } = {}

      // Invert rates: if 1 SAR = 0.27 USD, then 1 USD = 1/0.27 = 3.75 SAR
      for (const [currency, rate] of Object.entries(data.rates)) {
        rates[currency] = 1 / (rate as number)
      }

      return rates
    } catch (error) {
      console.error('Error fetching live rates:', error)
      return null
    }
  }

  // Update rates from online source
  async function updateRatesFromOnline(setBy: string, selectedCurrencies?: Currency[]): Promise<boolean> {
    const liveRates = await fetchLiveRates()

    if (!liveRates) {
      return false
    }

    const now = new Date()
    const effectiveDate = now.toISOString().split('T')[0]

    // Update only the currencies we're tracking (or selected currencies)
    const currenciesToUpdate = selectedCurrencies || getAllCurrencies.value.filter(c => c !== 'SAR')

    for (const currency of currenciesToUpdate) {
      if (liveRates[currency]) {
        // Round to 4 decimal places
        const roundedRate = Math.round(liveRates[currency] * 10000) / 10000
        updateRate(currency, roundedRate, effectiveDate, setBy, 'Updated from online source')
      }
    }

    return true
  }

  return {
    exchangeRates,
    rateHistory,
    getCurrentRate,
    getRateById,
    getHistoryByCurrency,
    getAllCurrencies,
    updateRate,
    addCurrency,
    fetchLiveRates,
    updateRatesFromOnline
  }
})
