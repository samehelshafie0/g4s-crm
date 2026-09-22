<script setup lang="ts">
import { exchangeRatesService } from '@/services'
import { errorMessage } from '@/services/payload'
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
import type { ExchangeRate, Currency } from '@/types'

function formatRate(v: number): string {
  return v.toFixed(4)
}

async function loadRates() {
  try { rates.value = (await exchangeRatesService.list()).data } catch (e) { window.alert(errorMessage(e)) }
}

onMounted(loadRates)

const refreshing = ref(false)
const refreshNote = ref('')

async function refreshLiveRates() {
  if (refreshing.value) return
  refreshing.value = true
  refreshNote.value = ''
  try {
    const result = await exchangeRatesService.refresh()
    await loadRates()
    const count = result.data.updated.length
    refreshNote.value = count
      ? `Updated ${count} ${count === 1 ? 'rate' : 'rates'} for ${result.data.effectiveDate} from ${result.data.provider}.`
      : `Rates were already current for ${result.data.effectiveDate}.`
  } catch (e) {
    refreshNote.value = errorMessage(e)
  } finally {
    refreshing.value = false
  }
}

const rates = ref<ExchangeRate[]>([])
const currencySymbols: Record<string, string> = {
  USD: '$', EUR: '€', GBP: '£', AED: 'د.إ', CNY: '¥',
}

const currencyNames: Record<string, string> = {
  USD: 'US Dollar', EUR: 'Euro', GBP: 'British Pound', AED: 'UAE Dirham', CNY: 'Chinese Yuan',
}

const selectedCurrency = ref<string>('USD')
const showUpdateModal = ref(false)
const creating = ref(false)
const saving = ref(false)
function openCreateModal() {
 creating.value = true
 updateForm.value = {currency: 'USD', newRate: 0, effectiveDate: new Date().toISOString().slice(0,10)}
 showUpdateModal.value = true
}
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
  creating.value = false
  updateForm.value = {
    currency: rate.fromCurrency,
    newRate: rate.currentRate,
    effectiveDate: new Date().toISOString().slice(0, 10),
  }
  showUpdateModal.value = true
}

async function saveRate() {
 if (saving.value) return
 saving.value = true
 try {
   if (creating.value) await exchangeRatesService.create({fromCurrency:updateForm.value.currency,toCurrency:'SAR',currentRate:updateForm.value.newRate,effectiveDate:updateForm.value.effectiveDate})
   else {
     const rate = rates.value.find(r => r.fromCurrency === updateForm.value.currency)
     if (!rate) throw new Error('Select an exchange rate')
     await exchangeRatesService.update(rate.id, updateForm.value.newRate, updateForm.value.effectiveDate)
   }
   rates.value = (await exchangeRatesService.list()).data
   selectedCurrency.value = updateForm.value.currency
   showUpdateModal.value = false
 } catch (e) { window.alert(errorMessage(e)) }
 finally { saving.value = false }
}
</script>

<template>
  <div class="exchange-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Exchange Rates</h1>
        <p class="page-header-subtitle">Currency rates against SAR</p>
      </div>
      <div class="fx-actions">
        <button class="btn btn-sm" :disabled="refreshing" @click="refreshLiveRates">
          <RefreshCw :size="14" :class="{ 'fx-spin': refreshing }" />
          {{ refreshing ? 'Fetching…' : 'Fetch live rates' }}
        </button>
        <button class="btn btn-primary btn-sm" @click="openCreateModal">Add exchange rate</button>
      </div>
    </div>

    <p v-if="refreshNote" class="fx-note" role="status">{{ refreshNote }}</p>
    <p v-if="!rates.length">Add your approved exchange rates to start recording their history.</p>

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
              <select v-if="creating" v-model="updateForm.currency" class="select form-input" aria-label="Currency"><option v-for="code in ['USD','EUR','GBP','AED','CNY']" :key="code" :value="code">{{ code }} → SAR</option></select>
              <input v-else :value="updateForm.currency + ' → SAR'" type="text" class="form-input" disabled />
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
            <button class="btn btn-primary" @click="saveRate" :disabled="saving || updateForm.newRate <= 0">
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
.fx-actions { display: flex; gap: var(--space-2); align-items: center; }
.fx-note { font-size: var(--text-sm); color: var(--color-neutral-600); margin-bottom: var(--space-3); }
.fx-spin { animation: fx-spin 1s linear infinite; }
@keyframes fx-spin { to { transform: rotate(360deg); } }
@media (prefers-reduced-motion: reduce) { .fx-spin { animation: none; } }

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
