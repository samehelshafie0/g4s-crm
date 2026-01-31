<template>
  <div class="exchange-rates-view">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h2>Exchange Rates Management</h2>
        <p class="subtitle">Manage currency exchange rates with live updates and historical tracking</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="showHistoryModal = true">
          <span>📊</span> View History
        </button>
        <button class="btn btn-primary" @click="fetchOnlineRates" :disabled="fetching">
          <span>{{ fetching ? '⏳' : '🌐' }}</span> {{ fetching ? 'Fetching...' : 'Fetch Live Rates' }}
        </button>
      </div>
    </div>

    <!-- Last Update Info -->
    <BaseCard v-if="lastUpdate">
      <div class="last-update-info">
        <span class="info-icon">🕒</span>
        <span class="info-text">Last updated: {{ formatDateTime(lastUpdate) }}</span>
      </div>
    </BaseCard>

    <!-- Exchange Rates Cards -->
    <div class="rates-grid">
      <BaseCard
        v-for="rate in activeCurrencies"
        :key="rate.id"
        :class="['rate-card', { 'base-currency': rate.currency === 'SAR' }]"
      >
        <div class="rate-header">
          <div class="currency-info">
            <h3 class="currency-code">{{ rate.currency }}</h3>
            <span class="currency-name">{{ getCurrencyName(rate.currency) }}</span>
            <BaseBadge v-if="rate.currency === 'SAR'" variant="success" size="sm">Base Currency</BaseBadge>
          </div>
          <button
            v-if="rate.currency !== 'SAR'"
            class="btn-icon"
            title="Edit Rate"
            @click="editRate(rate)"
          >
            ✏️
          </button>
        </div>

        <div class="rate-body">
          <div class="rate-display">
            <span class="rate-label">1 {{ rate.currency }} =</span>
            <span class="rate-value">{{ rate.rate.toFixed(4) }} SAR</span>
          </div>

          <div class="rate-meta">
            <div class="meta-item">
              <span class="meta-label">Effective Date:</span>
              <span class="meta-value">{{ formatDate(rate.effectiveDate) }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">Set At:</span>
              <span class="meta-value">{{ formatDateTime(rate.setAt) }}</span>
            </div>
            <div class="meta-item" v-if="rate.notes">
              <span class="meta-label">Notes:</span>
              <span class="meta-value">{{ rate.notes }}</span>
            </div>
          </div>
        </div>

        <div class="rate-footer">
          <button class="btn-link" @click="viewCurrencyHistory(rate.currency)">
            View History →
          </button>
        </div>
      </BaseCard>
    </div>

    <!-- Edit Rate Modal -->
    <BaseModal
      v-model="showEditModal"
      :title="`Edit ${editingRate?.currency} Exchange Rate`"
      size="md"
      @confirm="saveRateEdit"
    >
      <div class="modal-form" v-if="editingRate">
        <div class="form-section">
          <div class="current-rate-display">
            <span class="label">Current Rate:</span>
            <span class="value">1 {{ editingRate.currency }} = {{ editingRate.rate.toFixed(4) }} SAR</span>
          </div>

          <div class="form-group">
            <label class="form-label">New Exchange Rate *</label>
            <div class="rate-input-group">
              <span class="input-prefix">1 {{ editingRate.currency }} =</span>
              <input
                v-model.number="rateForm.rate"
                type="number"
                step="0.0001"
                class="form-input"
                placeholder="0.0000"
              />
              <span class="input-suffix">SAR</span>
            </div>
            <small class="form-hint">Enter the rate with up to 4 decimal places</small>
          </div>

          <div class="form-group">
            <label class="form-label">Effective Date *</label>
            <input
              v-model="rateForm.effectiveDate"
              type="date"
              class="form-input"
            />
          </div>

          <div class="form-group">
            <label class="form-label">Notes</label>
            <textarea
              v-model="rateForm.notes"
              class="form-input"
              rows="3"
              placeholder="Optional notes about this rate change..."
            ></textarea>
          </div>

          <div class="rate-preview" v-if="rateForm.rate">
            <h4>Preview:</h4>
            <div class="preview-row">
              <span>1 {{ editingRate.currency }}</span>
              <span>=</span>
              <span class="highlight">{{ rateForm.rate.toFixed(4) }} SAR</span>
            </div>
            <div class="preview-row">
              <span>100 {{ editingRate.currency }}</span>
              <span>=</span>
              <span class="highlight">{{ (rateForm.rate * 100).toFixed(2) }} SAR</span>
            </div>
            <div class="preview-row">
              <span>1,000 {{ editingRate.currency }}</span>
              <span>=</span>
              <span class="highlight">{{ (rateForm.rate * 1000).toFixed(2) }} SAR</span>
            </div>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- Currency History Modal -->
    <BaseModal
      v-model="showHistoryModal"
      :title="selectedHistoryCurrency ? `${selectedHistoryCurrency} Rate History` : 'Exchange Rate History'"
      size="lg"
    >
      <div class="history-content">
        <div class="currency-selector" v-if="!selectedHistoryCurrency">
          <label class="form-label">Select Currency:</label>
          <select v-model="selectedHistoryCurrency" class="form-input">
            <option value="">Choose a currency</option>
            <option v-for="currency in availableCurrencies" :key="currency" :value="currency">
              {{ currency }} - {{ getCurrencyName(currency) }}
            </option>
          </select>
        </div>

        <div v-if="selectedHistoryCurrency">
          <div class="history-header">
            <h3>{{ selectedHistoryCurrency }} - {{ getCurrencyName(selectedHistoryCurrency) }}</h3>
            <button class="btn-link" @click="selectedHistoryCurrency = null">← Back to selection</button>
          </div>

          <div class="history-timeline">
            <div
              v-for="(rate, index) in currencyHistory"
              :key="rate.id"
              :class="['history-item', { 'current': index === 0 }]"
            >
              <div class="history-marker">
                <span class="marker-dot"></span>
                <span class="marker-line" v-if="index < currencyHistory.length - 1"></span>
              </div>
              <div class="history-card">
                <div class="history-header-row">
                  <div class="history-date">
                    <span class="date-label">{{ formatDate(rate.effectiveDate) }}</span>
                    <BaseBadge v-if="index === 0" variant="success" size="sm">Current</BaseBadge>
                  </div>
                  <div class="history-rate">
                    {{ rate.rate.toFixed(4) }} SAR
                  </div>
                </div>
                <div class="history-meta">
                  <span class="meta-text">Updated: {{ formatDateTime(rate.setAt) }}</span>
                  <span class="meta-text" v-if="rate.notes">{{ rate.notes }}</span>
                </div>
                <div class="history-change" v-if="index < currencyHistory.length - 1">
                  <span :class="['change-badge', getChangeClass(rate.rate, currencyHistory[index + 1].rate)]">
                    {{ calculateChange(rate.rate, currencyHistory[index + 1].rate) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExchangeRatesStore } from '@/stores/exchangeRates'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { ExchangeRate, Currency } from '@/types'

const exchangeRatesStore = useExchangeRatesStore()
const { success, error, info } = useToast()

const showEditModal = ref(false)
const showHistoryModal = ref(false)
const editingRate = ref<ExchangeRate | null>(null)
const fetching = ref(false)
const selectedHistoryCurrency = ref<Currency | null>(null)

const rateForm = ref({
  rate: 0,
  effectiveDate: new Date().toISOString().split('T')[0],
  notes: ''
})

const activeCurrencies = computed(() => exchangeRatesStore.exchangeRates)
const availableCurrencies = computed(() => exchangeRatesStore.getAllCurrencies)

const lastUpdate = computed(() => {
  const dates = activeCurrencies.value.map(r => new Date(r.setAt).getTime())
  return dates.length > 0 ? new Date(Math.max(...dates)).toISOString() : null
})

const currencyHistory = computed(() => {
  if (!selectedHistoryCurrency.value) return []
  return exchangeRatesStore.getHistoryByCurrency(selectedHistoryCurrency.value)
})

function getCurrencyName(currency: Currency): string {
  const names: Record<Currency, string> = {
    'SAR': 'Saudi Riyal',
    'USD': 'US Dollar',
    'EUR': 'Euro',
    'GBP': 'British Pound',
    'AED': 'UAE Dirham',
    'CNY': 'Chinese Yuan'
  }
  return names[currency] || currency
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function editRate(rate: ExchangeRate) {
  editingRate.value = rate
  rateForm.value = {
    rate: rate.rate,
    effectiveDate: new Date().toISOString().split('T')[0],
    notes: ''
  }
  showEditModal.value = true
}

function saveRateEdit() {
  if (!editingRate.value) return

  if (!rateForm.value.rate || rateForm.value.rate <= 0) {
    error('Please enter a valid exchange rate')
    return
  }

  exchangeRatesStore.updateRate(
    editingRate.value.currency,
    rateForm.value.rate,
    rateForm.value.effectiveDate,
    'USR-001', // Should be current user
    rateForm.value.notes
  )

  success(`${editingRate.value.currency} exchange rate updated successfully`)
  showEditModal.value = false
  editingRate.value = null
}

async function fetchOnlineRates() {
  fetching.value = true

  try {
    const updated = await exchangeRatesStore.updateRatesFromOnline('USR-001')

    if (updated) {
      success('Exchange rates updated from online source')
    } else {
      error('Failed to fetch live exchange rates. Please try again.')
    }
  } catch (err) {
    error('Error fetching exchange rates')
  } finally {
    fetching.value = false
  }
}

function viewCurrencyHistory(currency: Currency) {
  selectedHistoryCurrency.value = currency
  showHistoryModal.value = true
}

function calculateChange(current: number, previous: number): string {
  const change = ((current - previous) / previous) * 100
  const prefix = change > 0 ? '+' : ''
  return `${prefix}${change.toFixed(2)}%`
}

function getChangeClass(current: number, previous: number): string {
  if (current > previous) return 'increase'
  if (current < previous) return 'decrease'
  return 'neutral'
}
</script>

<style scoped>
.exchange-rates-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

/* Header */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1rem;
  width: 100%;
  min-width: 0;
}

.view-header h2 {
  margin: 0;
  font-size: 1.75rem;
  color: var(--color-text-primary);
}

.subtitle {
  margin: 0.5rem 0 0 0;
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.btn {
  padding: 1rem 2rem;
  border: none;
  border-radius: 10px;
  font-weight: 500;
  font-size: 1.05rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.625rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-hover);
  transform: translateY(-2px);
}

.btn-secondary {
  background: var(--color-border);
  color: var(--color-text-primary);
}

.btn-secondary:hover {
  background: var(--color-surface-hover);
}

.btn-link {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: 0.9rem;
  padding: 0;
  transition: opacity 0.2s;
}

.btn-link:hover {
  opacity: 0.8;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.5rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.btn-icon:hover {
  transform: scale(1.2);
  opacity: 1;
}

/* Last Update Info */
.last-update-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
}

.info-icon {
  font-size: 1.5rem;
}

.info-text {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

/* Rates Grid */
.rates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
  width: 100%;
}

.rate-card {
  transition: transform 0.2s, box-shadow 0.2s;
}

.rate-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.rate-card.base-currency {
  border: 2px solid var(--color-success);
  background: linear-gradient(135deg, var(--color-success-light) 0%, var(--color-surface) 100%);
}

.rate-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.currency-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.currency-code {
  margin: 0;
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  font-family: 'Courier New', monospace;
}

.currency-name {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}

.rate-body {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.rate-display {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  border-left: 4px solid var(--color-primary);
}

.rate-label {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.rate-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
}

.rate-meta {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.meta-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.meta-value {
  font-size: 0.9rem;
  color: var(--color-text-primary);
}

.rate-footer {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

/* Modal Form */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.current-rate-display {
  background: var(--color-info-light);
  padding: 1rem;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.current-rate-display .label {
  font-weight: 600;
  color: var(--color-text-primary);
}

.current-rate-display .value {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 1.1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-input {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.rate-input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.input-prefix,
.input-suffix {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.95rem;
  white-space: nowrap;
}

.rate-input-group .form-input {
  flex: 1;
  text-align: center;
  font-family: 'Courier New', monospace;
  font-weight: 700;
  font-size: 1.2rem;
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

.rate-preview {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 8px;
  border: 2px solid var(--color-primary);
}

.rate-preview h4 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.preview-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
  font-size: 1rem;
}

.preview-row:last-child {
  border-bottom: none;
}

.preview-row .highlight {
  font-weight: 700;
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
}

/* History Modal */
.history-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.currency-selector {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.history-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--color-text-primary);
}

.history-timeline {
  position: relative;
  padding-left: 2rem;
}

.history-item {
  position: relative;
  margin-bottom: 2rem;
}

.history-marker {
  position: absolute;
  left: -2rem;
  top: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.marker-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--color-border);
  border: 3px solid var(--color-surface);
  box-shadow: 0 0 0 2px var(--color-border);
}

.history-item.current .marker-dot {
  background: var(--color-success);
  box-shadow: 0 0 0 2px var(--color-success);
}

.marker-line {
  width: 2px;
  height: 100%;
  min-height: 60px;
  background: var(--color-border);
  margin-top: 4px;
}

.history-card {
  background: var(--color-surface);
  padding: 1.25rem;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.history-item.current .history-card {
  border-color: var(--color-success);
  background: var(--color-success-light);
}

.history-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.history-date {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.date-label {
  font-weight: 600;
  color: var(--color-text-primary);
}

.history-rate {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
}

.history-meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
}

.meta-text {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.history-change {
  margin-top: 0.5rem;
}

.change-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.85rem;
  font-weight: 600;
}

.change-badge.increase {
  background: var(--color-success-light);
  color: var(--color-success);
}

.change-badge.decrease {
  background: var(--color-danger-light);
  color: var(--color-danger);
}

.change-badge.neutral {
  background: var(--color-background);
  color: var(--color-text-secondary);
}
</style>
