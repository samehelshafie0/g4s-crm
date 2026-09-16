<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  TrendingUp,
  TrendingDown,
  RefreshCw,
  X,
  Save,
  DollarSign,
  History,
} from 'lucide-vue-next'
import type { ExchangeRate, Currency, ExchangeRateHistory } from '@/types'
import { useExchangeRatesStore } from '@/stores/exchangeRates'

function formatRate(v: number): string {
  return v.toFixed(4)
}

function genHistory(base: number, months: number): ExchangeRateHistory[] {
  const hist: ExchangeRateHistory[] = []
  const now = new Date()
  for (let i = months; i >= 0; i--) {
    const d = new Date(now)
    d.setMonth(d.getMonth() - i)
    const variance = (Math.random() - 0.5) * 0.04 * base
    hist.push({
      rate: Math.round((base + variance) * 10000) / 10000,
      effectiveDate: d.toISOString().slice(0, 10),
    })
  }
  return hist
}

const fxStore = useExchangeRatesStore()
onMounted(() => fxStore.fetchRates())

const rates = ref<ExchangeRate[]>([
  { id: 'fx1', fromCurrency: 'USD', toCurrency: 'SAR', currentRate: 3.7500, effectiveDate: '2026-02-20', history: genHistory(3.75, 3), createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'fx2', fromCurrency: 'EUR', toCurrency: 'SAR', currentRate: 4.1000, effectiveDate: '2026-02-20', history: genHistory(4.10, 3), createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'fx3', fromCurrency: 'GBP', toCurrency: 'SAR', currentRate: 4.7200, effectiveDate: '2026-02-20', history: genHistory(4.72, 3), createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'fx4', fromCurrency: 'AED', toCurrency: 'SAR', currentRate: 1.0210, effectiveDate: '2026-02-20', history: genHistory(1.021, 3), createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'fx5', fromCurrency: 'CNY', toCurrency: 'SAR', currentRate: 0.5200, effectiveDate: '2026-02-20', history: genHistory(0.52, 3), createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
])

const currencySymbols: Record<string, string> = {
  USD: '$', EUR: '€', GBP: '£', AED: 'د.إ', CNY: '¥',
}

const currencyNames: Record<string, string> = {
  USD: 'US Dollar', EUR: 'Euro', GBP: 'British Pound', AED: 'UAE Dirham', CNY: 'Chinese Yuan',
}

const selectedCurrency = ref<string>('USD')
const showUpdateModal = ref(false)
const updateForm = ref({ currency: '' as Currency, newRate: 0, effectiveDate: '' })

const selectedRate = computed(() => rates.value.find(r => r.fromCurrency === selectedCurrency.value))

const selectedHistory = computed(() => {
  const r = selectedRate.value
  if (!r) return []
  return [...r.history].sort((a, b) => b.effectiveDate.localeCompare(a.effectiveDate))
})

function lastChange(rate: ExchangeRate): { pct: number; up: boolean } {
  if (rate.history.length < 2) return { pct: 0, up: true }
  const sorted = [...rate.history].sort((a, b) => b.effectiveDate.localeCompare(a.effectiveDate))
  const curr = sorted[0]!.rate
  const prev = sorted[1]!.rate
  const pct = ((curr - prev) / prev) * 100
  return { pct: Math.abs(Math.round(pct * 100) / 100), up: pct >= 0 }
}

function openUpdateModal(rate: ExchangeRate) {
  updateForm.value = {
    currency: rate.fromCurrency,
    newRate: rate.currentRate,
    effectiveDate: new Date().toISOString().slice(0, 10),
  }
  showUpdateModal.value = true
}

function saveRate() {
  const r = rates.value.find(r => r.fromCurrency === updateForm.value.currency)
  if (!r) return
  r.history.push({ rate: r.currentRate, effectiveDate: r.effectiveDate })
  r.currentRate = updateForm.value.newRate
  r.effectiveDate = updateForm.value.effectiveDate
  r.updatedAt = new Date().toISOString()
  showUpdateModal.value = false
}
</script>

<template>
  <div class="exchange-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Exchange Rates</h1>
        <p class="page-header-subtitle">Currency rates against SAR</p>
      </div>
    </div>

    <!-- Currency Cards Grid -->
    <div class="fx-grid">
      <button
        v-for="rate in rates"
        :key="rate.id"
        :class="['fx-card', selectedCurrency === rate.fromCurrency && 'fx-card--selected']"
        @click="selectedCurrency = rate.fromCurrency"
      >
        <div class="fx-card-top">
          <div class="fx-pair">
            <span class="fx-symbol">{{ currencySymbols[rate.fromCurrency] }}</span>
            <div>
              <div class="fx-pair-code">{{ rate.fromCurrency }} → SAR</div>
              <div class="fx-pair-name">{{ currencyNames[rate.fromCurrency] }}</div>
            </div>
          </div>
          <button class="btn btn-ghost btn-icon btn-sm" @click.stop="openUpdateModal(rate)" title="Update rate">
            <RefreshCw :size="14" />
          </button>
        </div>
        <div class="fx-rate">{{ formatRate(rate.currentRate) }}</div>
        <div class="fx-card-bottom">
          <span class="fx-effective">Effective {{ rate.effectiveDate }}</span>
          <span :class="['fx-trend', lastChange(rate).up ? 'fx-trend--up' : 'fx-trend--down']">
            <component :is="lastChange(rate).up ? TrendingUp : TrendingDown" :size="14" />
            {{ lastChange(rate).pct }}%
          </span>
        </div>
      </button>
    </div>

    <!-- Rate History -->
    <div class="card mt-6">
      <div class="card-header">
        <div class="history-header">
          <History :size="18" />
          <h3 class="history-title">{{ selectedCurrency }} → SAR Rate History</h3>
        </div>
      </div>
      <div v-if="selectedHistory.length" class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>Date</th>
              <th class="text-right">Rate</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(h, i) in selectedHistory" :key="i">
              <td>{{ h.effectiveDate }}</td>
              <td class="text-right text-mono font-medium">{{ formatRate(h.rate) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">
        <DollarSign :size="36" class="empty-state-icon" />
        <p class="empty-state-text">No history available.</p>
      </div>
    </div>

    <!-- Update Rate Modal -->
    <Teleport to="body">
      <div v-if="showUpdateModal" class="modal-backdrop" @click.self="showUpdateModal = false">
        <div class="modal">
          <div class="modal-header">
            <h2 class="modal-title">Update Exchange Rate</h2>
            <button class="modal-close" @click="showUpdateModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Currency</label>
              <input :value="updateForm.currency + ' → SAR'" type="text" class="form-input" disabled />
            </div>
            <div class="form-group">
              <label class="form-label">New Rate</label>
              <input v-model.number="updateForm.newRate" type="number" step="0.0001" min="0" class="form-input" />
            </div>
            <div class="form-group">
              <label class="form-label">Effective Date</label>
              <input v-model="updateForm.effectiveDate" type="date" class="form-input" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showUpdateModal = false">Cancel</button>
            <button class="btn btn-primary" @click="saveRate">
              <Save :size="14" />
              Update Rate
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.exchange-page {
  padding: var(--space-6);
}

.fx-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: var(--space-4);
}

.fx-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: var(--space-5);
  background: var(--content-surface);
  border: 2px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
  width: 100%;
}

.fx-card:hover { border-color: var(--color-primary); }
.fx-card--selected { border-color: var(--color-primary); box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1); }

.fx-card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.fx-pair {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.fx-symbol {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: var(--color-neutral-100);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-lg, 1.125rem);
  font-weight: var(--font-bold, 700);
  color: var(--color-neutral-700);
}

.fx-pair-code {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.fx-pair-name {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.fx-rate {
  font-size: 1.75rem;
  font-weight: var(--font-bold, 700);
  color: var(--color-neutral-900);
  font-family: var(--font-mono);
}

.fx-card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fx-effective {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.fx-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
}

.fx-trend--up { color: #059669; }
.fx-trend--down { color: #dc2626; }

.history-header {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.history-title {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
  margin: 0;
}

@media (max-width: 768px) {
  .fx-grid { grid-template-columns: 1fr 1fr; }
}

@media (max-width: 480px) {
  .fx-grid { grid-template-columns: 1fr; }
}
</style>
