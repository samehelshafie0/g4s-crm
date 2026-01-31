<template>
  <div class="integrations-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search integrations..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="addIntegration">
          <span>➕</span>
          Add Integration
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Category:</label>
        <select v-model="categoryFilter" class="filter-select">
          <option value="all">All Categories</option>
          <option value="accounting">Accounting/ERP</option>
          <option value="hr">Human Resources</option>
          <option value="email">Email</option>
          <option value="payment">Payment</option>
          <option value="auth">Authentication</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="connected">Connected</option>
          <option value="disconnected">Disconnected</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total</span>
          <span class="stat-value">{{ filteredIntegrations.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active</span>
          <span class="stat-value text-success">{{ connectedCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Syncs</span>
          <span class="stat-value">{{ totalSyncs }}</span>
        </div>
      </div>
    </div>

    <BaseTable
      :columns="tableColumns"
      :data="filteredIntegrations"
      :loading="false"
    >
      <template #cell-name="{ row }">
        <div class="name-cell">
          <span class="integration-icon">{{ row.icon }}</span>
          <div>
            <div class="integration-name">{{ row.name }}</div>
            <div class="integration-desc">{{ row.description }}</div>
          </div>
        </div>
      </template>

      <template #cell-category="{ row }">
        <BaseBadge :variant="getCategoryVariant(row.category)">
          {{ formatCategory(row.category) }}
        </BaseBadge>
      </template>

      <template #cell-lastSync="{ row }">
        <div class="sync-cell">
          <div v-if="row.lastSync">{{ formatDate(row.lastSync) }}</div>
          <div v-else class="text-secondary">Never</div>
        </div>
      </template>

      <template #cell-syncCount="{ row }">
        <div class="count-cell">
          <span class="count-value">{{ row.syncCount }}</span>
          <span class="text-secondary">syncs</span>
        </div>
      </template>

      <template #cell-status="{ row }">
        <BaseBadge :variant="getStatusVariant(row.status)">
          {{ formatStatus(row.status) }}
        </BaseBadge>
      </template>

      <template #cell-actions="{ row }">
        <div class="action-buttons">
          <button class="action-btn" @click="configureIntegration(row)" title="Configure">
            ⚙️
          </button>
          <button
            class="action-btn"
            @click="syncIntegration(row)"
            :title="row.status === 'connected' ? 'Sync Now' : 'Connect First'"
            :disabled="row.status !== 'connected'"
          >
            🔄
          </button>
          <button
            class="action-btn"
            @click="toggleConnection(row)"
            :title="row.status === 'connected' ? 'Disconnect' : 'Connect'"
          >
            {{ row.status === 'connected' ? '🔌' : '🔗' }}
          </button>
        </div>
      </template>
    </BaseTable>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import { useToast } from '@/composables/useToast'

const { info, success } = useToast()

interface Integration {
  id: string
  name: string
  description: string
  icon: string
  category: 'accounting' | 'hr' | 'email' | 'payment' | 'auth'
  status: 'connected' | 'disconnected'
  lastSync: string | null
  syncCount: number
}

const searchQuery = ref('')
const categoryFilter = ref('all')
const statusFilter = ref('all')

const integrations = ref<Integration[]>([
  {
    id: '1',
    name: 'QuickBooks Online',
    description: 'Sync invoices and financial data',
    icon: '💼',
    category: 'accounting',
    status: 'connected',
    lastSync: '2024-03-18T14:30:00',
    syncCount: 1247
  },
  {
    id: '2',
    name: 'Microsoft Office 365',
    description: 'Email and calendar integration',
    icon: '📧',
    category: 'email',
    status: 'connected',
    lastSync: '2024-03-18T16:45:00',
    syncCount: 3421
  },
  {
    id: '3',
    name: 'Azure AD',
    description: 'Single Sign-On authentication',
    icon: '🔐',
    category: 'auth',
    status: 'connected',
    lastSync: '2024-03-18T15:20:00',
    syncCount: 892
  },
  {
    id: '4',
    name: 'Stripe',
    description: 'Online payment processing',
    icon: '💳',
    category: 'payment',
    status: 'connected',
    lastSync: '2024-03-18T12:10:00',
    syncCount: 567
  },
  {
    id: '5',
    name: 'SAP Business One',
    description: 'ERP system integration',
    icon: '🏢',
    category: 'accounting',
    status: 'disconnected',
    lastSync: null,
    syncCount: 0
  },
  {
    id: '6',
    name: 'BambooHR',
    description: 'Employee data synchronization',
    icon: '👥',
    category: 'hr',
    status: 'disconnected',
    lastSync: null,
    syncCount: 0
  },
  {
    id: '7',
    name: 'Xe Currency Data',
    description: 'Real-time exchange rates',
    icon: '💱',
    category: 'accounting',
    status: 'connected',
    lastSync: '2024-03-18T17:00:00',
    syncCount: 2156
  }
])

const tableColumns = [
  { key: 'name', label: 'Integration', sortable: true },
  { key: 'category', label: 'Category', sortable: true },
  { key: 'lastSync', label: 'Last Sync', sortable: true },
  { key: 'syncCount', label: 'Sync Count', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false }
]

const filteredIntegrations = computed(() => {
  let filtered = integrations.value

  if (categoryFilter.value !== 'all') {
    filtered = filtered.filter(int => int.category === categoryFilter.value)
  }

  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(int => int.status === statusFilter.value)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(int =>
      int.name.toLowerCase().includes(query) ||
      int.description.toLowerCase().includes(query)
    )
  }

  return filtered
})

const connectedCount = computed(() =>
  integrations.value.filter(int => int.status === 'connected').length
)

const totalSyncs = computed(() =>
  integrations.value.reduce((sum, int) => sum + int.syncCount, 0)
)

function formatCategory(category: Integration['category']): string {
  const categories: Record<Integration['category'], string> = {
    accounting: 'Accounting/ERP',
    hr: 'Human Resources',
    email: 'Email',
    payment: 'Payment',
    auth: 'Authentication'
  }
  return categories[category]
}

function formatStatus(status: Integration['status']): string {
  const statuses: Record<Integration['status'], string> = {
    connected: 'Connected',
    disconnected: 'Disconnected'
  }
  return statuses[status]
}

function getCategoryVariant(category: Integration['category']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Integration['category'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    accounting: 'primary',
    hr: 'info',
    email: 'success',
    payment: 'warning',
    auth: 'danger'
  }
  return variants[category]
}

function getStatusVariant(status: Integration['status']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Integration['status'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    connected: 'success',
    disconnected: 'default'
  }
  return variants[status]
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 60) return `${diffMins}m ago`
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours}h ago`
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function addIntegration() {
  info('Add integration functionality')
}

function configureIntegration(integration: Integration) {
  info(`Configuring: ${integration.name}`)
}

function syncIntegration(integration: Integration) {
  if (integration.status !== 'connected') return
  integration.lastSync = new Date().toISOString()
  integration.syncCount++
  success(`${integration.name} synced successfully`)
}

function toggleConnection(integration: Integration) {
  integration.status = integration.status === 'connected' ? 'disconnected' : 'connected'
  success(`${integration.name} ${integration.status === 'connected' ? 'connected' : 'disconnected'}`)
}
</script>

<style scoped>
.integrations-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: 100%;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1.25rem;
  width: 100%;
}

.search-input {
  flex: 1;
  padding: 1rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 10px;
  font-size: 1.05rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  padding: 1rem 2rem;
  border: none;
  border-radius: 10px;
  font-size: 1.05rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.625rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
  transform: translateY(-2px);
}

.filters-bar {
  display: flex;
  gap: 2.5rem;
  align-items: center;
  flex-wrap: wrap;
  padding: 1.75rem 2rem;
  background: var(--color-surface);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-select {
  padding: 0.75rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
  min-width: 200px;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

.stats-summary {
  display: flex;
  gap: 2.5rem;
  margin-left: auto;
  padding-left: 2.5rem;
  border-left: 1px solid var(--color-border);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-primary);
}

.text-success {
  color: var(--color-success) !important;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0.75rem;
  min-width: 350px;
}

.integration-icon {
  font-size: 2.5rem;
  flex-shrink: 0;
}

.integration-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.integration-desc {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.sync-cell {
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
  color: var(--color-text-primary);
}

.count-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
}

.count-value {
  font-weight: 600;
  font-size: 1.15rem;
  color: var(--color-primary);
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 0.95rem;
}

.action-buttons {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 0.5rem;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.3rem;
  cursor: pointer;
  padding: 0.35rem 0.65rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover:not(:disabled) {
  transform: scale(1.2);
  opacity: 1;
}

.action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
</style>
