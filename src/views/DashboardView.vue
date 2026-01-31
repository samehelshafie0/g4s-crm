<template>
  <div class="dashboard-view">
    <div class="kpi-cards">
      <div class="kpi-card">
        <div class="kpi-icon">👥</div>
        <div class="kpi-content">
          <div class="kpi-label">Total Customers</div>
          <div class="kpi-value">{{ customersStore.customers.length }}</div>
          <div class="kpi-change positive">{{ customersStore.activeCustomers.length }} active</div>
        </div>
      </div>

      <div class="kpi-card">
        <div class="kpi-icon">🎯</div>
        <div class="kpi-content">
          <div class="kpi-label">Active Opportunities</div>
          <div class="kpi-value">{{ opportunitiesStore.activeOpportunities.length }}</div>
          <div class="kpi-change positive">{{ formatCurrency(opportunitiesStore.totalOpportunityValue) }}</div>
        </div>
      </div>

      <div class="kpi-card">
        <div class="kpi-icon">📝</div>
        <div class="kpi-content">
          <div class="kpi-label">Total Quotes</div>
          <div class="kpi-value">{{ quotesStore.quotes.length }}</div>
          <div class="kpi-change" :class="approvedQuotesPercent >= 50 ? 'positive' : 'neutral'">
            {{ approvedQuotesPercent }}% approved
          </div>
        </div>
      </div>

      <div class="kpi-card">
        <div class="kpi-icon">💰</div>
        <div class="kpi-content">
          <div class="kpi-label">Total Quote Value</div>
          <div class="kpi-value">{{ formatCurrency(totalQuoteValue) }}</div>
          <div class="kpi-change positive">{{ quotesStore.quotes.length }} quotes</div>
        </div>
      </div>

      <div class="kpi-card">
        <div class="kpi-icon">🏆</div>
        <div class="kpi-content">
          <div class="kpi-label">Win Rate</div>
          <div class="kpi-value">{{ opportunitiesStore.winRate }}%</div>
          <div class="kpi-change positive">{{ opportunitiesStore.wonOpportunities.length }} won</div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <div class="card chart-card">
        <div class="card-header">
          <h3>Opportunity Pipeline</h3>
          <select class="filter-select">
            <option>All Stages</option>
            <option>Active Only</option>
          </select>
        </div>
        <div class="pipeline-overview">
          <div
            v-for="stage in opportunityStages"
            :key="stage.key"
            class="pipeline-stage"
          >
            <div class="stage-header">
              <span class="stage-icon">{{ stage.icon }}</span>
              <span class="stage-name">{{ stage.label }}</span>
            </div>
            <div class="stage-count">{{ getOpportunitiesByStage(stage.key).length }}</div>
            <div class="stage-value">{{ formatCurrency(getTotalValueByStage(stage.key)) }}</div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>Quote Distribution</h3>
        </div>
        <div class="quote-distribution">
          <div class="distribution-item approved">
            <span class="distribution-label">Approved</span>
            <span class="distribution-value">{{ approvedQuotes.length }}</span>
            <span class="distribution-percent">{{ approvedQuotesPercent }}%</span>
          </div>
          <div class="distribution-item pending">
            <span class="distribution-label">Pending Review</span>
            <span class="distribution-value">{{ pendingQuotes.length }}</span>
            <span class="distribution-percent">{{ pendingQuotesPercent }}%</span>
          </div>
          <div class="distribution-item draft">
            <span class="distribution-label">Draft</span>
            <span class="distribution-value">{{ draftQuotes.length }}</span>
            <span class="distribution-percent">{{ draftQuotesPercent }}%</span>
          </div>
          <div class="distribution-item rejected">
            <span class="distribution-label">Rejected</span>
            <span class="distribution-value">{{ rejectedQuotes.length }}</span>
            <span class="distribution-percent">{{ rejectedQuotesPercent }}%</span>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <div class="card">
        <div class="card-header">
          <h3>Recent Opportunities</h3>
          <button class="view-all-btn" @click="navigateTo('/leads')">View All</button>
        </div>
        <div class="updates-list">
          <div
            v-for="opp in recentOpportunities"
            :key="opp.id"
            class="update-item"
          >
            <span class="update-icon">{{ getStageIcon(opp.stage) }}</span>
            <div class="update-content">
              <div class="update-title">{{ opp.name }}</div>
              <div class="update-time">
                {{ getCompanyName(opp.companyId) }} • {{ formatCurrency(opp.estimatedValue) }}
              </div>
            </div>
          </div>
          <div v-if="recentOpportunities.length === 0" class="empty-state">
            No recent opportunities
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>Inventory Alerts</h3>
          <span class="alert-badge">{{ inventoryAlerts.length }}</span>
        </div>
        <div class="alerts-list">
          <div
            v-for="alert in inventoryAlerts"
            :key="alert.id"
            class="alert-item warning"
          >
            <span class="alert-icon">⚠️</span>
            <div class="alert-content">
              <div class="alert-title">Low Stock: {{ getProductName(alert.productId) }}</div>
              <div class="alert-detail">Only {{ alert.quantityAvailable }} units at {{ alert.warehouseLocation }}</div>
            </div>
          </div>
          <div v-if="inventoryAlerts.length === 0" class="alert-item info">
            <span class="alert-icon">✅</span>
            <div class="alert-content">
              <div class="alert-title">All Stock Levels Normal</div>
              <div class="alert-detail">No inventory alerts at this time</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-grid">
      <div class="card">
        <div class="card-header">
          <h3>Top Customers by Revenue</h3>
        </div>
        <div class="customers-list">
          <div
            v-for="customer in topCustomers"
            :key="customer.id"
            class="customer-item"
          >
            <div class="customer-avatar">{{ customer.name.charAt(0) }}</div>
            <div class="customer-info">
              <div class="customer-name">{{ customer.name }}</div>
              <div class="customer-sector">{{ customer.sector }}</div>
            </div>
            <div class="customer-revenue">
              {{ formatCurrency(getCustomerRevenue(customer.id)) }}
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>Recent Quotes</h3>
          <button class="view-all-btn" @click="navigateTo('/quoting')">View All</button>
        </div>
        <div class="quotes-list">
          <div
            v-for="quote in recentQuotes"
            :key="quote.id"
            class="quote-item"
          >
            <div class="quote-info">
              <div class="quote-number">{{ quote.quoteNumber }}</div>
              <div class="quote-customer">{{ getCompanyName(quote.companyId) }}</div>
            </div>
            <div class="quote-value">{{ formatCurrency(quote.total) }}</div>
          </div>
          <div v-if="recentQuotes.length === 0" class="empty-state">
            No recent quotes
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCustomersStore } from '@/stores/customers'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { useQuotesStore } from '@/stores/quotes'
import { useWarehouseStore } from '@/stores/warehouse'
import { useProductsStore } from '@/stores/products'
import type { OpportunityStage } from '@/types'

const router = useRouter()
const customersStore = useCustomersStore()
const opportunitiesStore = useOpportunitiesStore()
const quotesStore = useQuotesStore()
const warehouseStore = useWarehouseStore()
const productsStore = useProductsStore()

const opportunityStages = [
  { key: 'qualification' as OpportunityStage, label: 'Qualification', icon: '🔍' },
  { key: 'proposal' as OpportunityStage, label: 'Proposal', icon: '📋' },
  { key: 'negotiation' as OpportunityStage, label: 'Negotiation', icon: '💬' },
  { key: 'closed_won' as OpportunityStage, label: 'Closed Won', icon: '✅' },
  { key: 'closed_lost' as OpportunityStage, label: 'Closed Lost', icon: '❌' }
]

// Quote distributions
const approvedQuotes = computed(() =>
  quotesStore.quotes.filter(q => q.status === 'approved')
)

const pendingQuotes = computed(() =>
  quotesStore.quotes.filter(q => q.status === 'pending_approval')
)

const draftQuotes = computed(() =>
  quotesStore.quotes.filter(q => q.status === 'draft')
)

const rejectedQuotes = computed(() =>
  quotesStore.quotes.filter(q => q.status === 'declined' || q.status === 'expired')
)

const totalQuotes = computed(() => quotesStore.quotes.length)

const approvedQuotesPercent = computed(() =>
  totalQuotes.value > 0
    ? Math.round((approvedQuotes.value.length / totalQuotes.value) * 100)
    : 0
)

const pendingQuotesPercent = computed(() =>
  totalQuotes.value > 0
    ? Math.round((pendingQuotes.value.length / totalQuotes.value) * 100)
    : 0
)

const draftQuotesPercent = computed(() =>
  totalQuotes.value > 0
    ? Math.round((draftQuotes.value.length / totalQuotes.value) * 100)
    : 0
)

const rejectedQuotesPercent = computed(() =>
  totalQuotes.value > 0
    ? Math.round((rejectedQuotes.value.length / totalQuotes.value) * 100)
    : 0
)

const totalQuoteValue = computed(() =>
  quotesStore.quotes.reduce((sum, q) => sum + q.total, 0)
)

// Recent opportunities (last 5)
const recentOpportunities = computed(() =>
  Array.from(opportunitiesStore.activeOpportunities)
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5)
)

// Recent quotes (last 5)
const recentQuotes = computed(() =>
  Array.from(quotesStore.quotes)
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5)
)

// Top customers by revenue
const topCustomers = computed(() => {
  const customersWithRevenue = customersStore.customers.map(customer => ({
    ...customer,
    revenue: getCustomerRevenue(customer.id)
  }))
  return customersWithRevenue
    .sort((a, b) => b.revenue - a.revenue)
    .slice(0, 5)
})

// Inventory alerts (low stock items - using quantityAvailable < 10 as threshold)
const inventoryAlerts = computed(() =>
  warehouseStore.stock
    .filter(item => item.quantityAvailable < 10)
    .slice(0, 3)
)

// Helper functions
function getOpportunitiesByStage(stage: OpportunityStage) {
  return opportunitiesStore.opportunities.filter(o => o.stage === stage)
}

function getTotalValueByStage(stage: OpportunityStage) {
  return getOpportunitiesByStage(stage).reduce((sum, o) => sum + o.estimatedValue, 0)
}

function getStageIcon(stage: OpportunityStage) {
  const stageData = opportunityStages.find(s => s.key === stage)
  return stageData?.icon || '📋'
}

function getCompanyName(companyId: string) {
  const company = customersStore.getCustomerById(companyId)
  return company?.name || 'Unknown'
}

function getProductName(productId: string) {
  const product = productsStore.getProductById(productId)
  return product?.description || 'Unknown Product'
}

function getCustomerRevenue(customerId: string) {
  const history = customersStore.getCustomerHistory(customerId)
  return history?.totalRevenue || 0
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function navigateTo(path: string) {
  router.push(path)
}
</script>

<style scoped>
.dashboard-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.kpi-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1.25rem;
}

.kpi-card {
  background: var(--color-surface);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  gap: 1rem;
  box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s, background-color 0.3s;
}

.kpi-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.kpi-icon {
  font-size: 2.5rem;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
}

.kpi-content {
  flex: 1;
}

.kpi-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin-bottom: 0.5rem;
}

.kpi-value {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
  transition: color 0.3s;
}

.kpi-change {
  font-size: 0.8rem;
  font-weight: 500;
}

.kpi-change.positive {
  color: var(--color-success);
}

.kpi-change.negative {
  color: var(--color-danger);
}

.kpi-change.neutral {
  color: var(--color-text-secondary);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
}

.card {
  background: var(--color-surface);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--shadow-sm);
  transition: background-color 0.3s;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.card-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  transition: color 0.3s;
}

.filter-select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 6px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

.chart-card {
  grid-column: 1 / -1;
}

.pipeline-overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.pipeline-stage {
  background: var(--color-background);
  padding: 1.25rem;
  border-radius: 8px;
  text-align: center;
  transition: background-color 0.3s;
}

.stage-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.stage-icon {
  font-size: 2rem;
}

.stage-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.stage-count {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.stage-value {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.quote-distribution {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.distribution-item {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-radius: 8px;
  gap: 1rem;
}

.distribution-item.approved {
  background: var(--color-success-light);
}

.distribution-item.pending {
  background: var(--color-warning-light);
}

.distribution-item.draft {
  background: var(--color-info-light);
}

.distribution-item.rejected {
  background: var(--color-danger-light);
}

.distribution-label {
  flex: 1;
  font-weight: 500;
  color: var(--color-text-primary);
}

.distribution-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  min-width: 60px;
  text-align: right;
}

.distribution-percent {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  min-width: 50px;
  text-align: right;
}

.view-all-btn {
  background: none;
  border: none;
  color: var(--color-primary);
  font-size: 0.875rem;
  cursor: pointer;
  font-weight: 500;
  transition: opacity 0.2s;
}

.view-all-btn:hover {
  opacity: 0.8;
}

.updates-list,
.alerts-list,
.customers-list,
.quotes-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.update-item,
.alert-item,
.customer-item,
.quote-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  border-radius: 8px;
  transition: background 0.2s;
  align-items: center;
}

.update-item:hover,
.alert-item:hover,
.customer-item:hover,
.quote-item:hover {
  background: var(--color-surface-hover);
}

.update-icon,
.alert-icon {
  font-size: 1.5rem;
}

.update-content,
.alert-content,
.customer-info,
.quote-info {
  flex: 1;
}

.update-title,
.alert-title,
.customer-name,
.quote-number {
  font-weight: 500;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.update-time,
.alert-detail,
.customer-sector,
.quote-customer {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.alert-badge {
  background: var(--color-danger);
  color: #fff;
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  border-radius: 12px;
  font-weight: 600;
}

.alert-item.warning {
  border-left: 3px solid var(--color-warning);
}

.alert-item.info {
  border-left: 3px solid var(--color-info);
}

.customer-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
}

.customer-revenue,
.quote-value {
  font-weight: 600;
  color: var(--color-primary);
  white-space: nowrap;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}
</style>
