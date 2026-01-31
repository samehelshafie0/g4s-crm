<template>
  <div class="workflows-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search workflows..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createWorkflow">
          <span>➕</span>
          New Workflow
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="approval">Approval</option>
          <option value="notification">Notification</option>
          <option value="automation">Automation</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total</span>
          <span class="stat-value">{{ filteredWorkflows.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active</span>
          <span class="stat-value text-success">{{ activeCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Pending</span>
          <span class="stat-value text-warning">{{ pendingCount }}</span>
        </div>
      </div>
    </div>

    <BaseTable
      :columns="tableColumns"
      :data="filteredWorkflows"
      :loading="false"
    >
      <template #cell-name="{ row }">
        <div class="name-cell">
          <div class="workflow-name">{{ row.name }}</div>
          <div class="workflow-desc">{{ row.description }}</div>
        </div>
      </template>

      <template #cell-type="{ row }">
        <BaseBadge :variant="getTypeVariant(row.type)">
          {{ formatType(row.type) }}
        </BaseBadge>
      </template>

      <template #cell-trigger="{ row }">
        <div class="trigger-cell">
          {{ row.trigger }}
        </div>
      </template>

      <template #cell-approvers="{ row }">
        <div class="approvers-cell">
          <span class="approver-count">{{ row.approvers.length }}</span>
          <span class="text-secondary">approvers</span>
        </div>
      </template>

      <template #cell-executed="{ row }">
        <div class="executed-cell">
          <div class="executed-count">{{ row.executedCount }}</div>
          <div class="text-secondary">times</div>
        </div>
      </template>

      <template #cell-status="{ row }">
        <BaseBadge :variant="getStatusVariant(row.status)">
          {{ formatStatus(row.status) }}
        </BaseBadge>
      </template>

      <template #cell-actions="{ row }">
        <div class="action-buttons">
          <button class="action-btn" @click="viewWorkflow(row)" title="View">
            👁️
          </button>
          <button class="action-btn" @click="editWorkflow(row)" title="Edit">
            ✏️
          </button>
          <button class="action-btn" @click="toggleStatus(row)" :title="row.status === 'active' ? 'Deactivate' : 'Activate'">
            {{ row.status === 'active' ? '⏸️' : '▶️' }}
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

interface Workflow {
  id: string
  name: string
  description: string
  type: 'approval' | 'notification' | 'automation'
  trigger: string
  approvers: string[]
  executedCount: number
  status: 'active' | 'inactive'
}

const searchQuery = ref('')
const typeFilter = ref('all')
const statusFilter = ref('all')

const workflows = ref<Workflow[]>([
  {
    id: '1',
    name: 'Quote Approval Workflow',
    description: 'Approval required for quotes over SAR 100,000',
    type: 'approval',
    trigger: 'Quote value > 100K',
    approvers: ['Ahmed Al-Rashid', 'Sarah Abdullah'],
    executedCount: 42,
    status: 'active'
  },
  {
    id: '2',
    name: 'New Customer Welcome',
    description: 'Send welcome email to new customers',
    type: 'notification',
    trigger: 'Customer created',
    approvers: [],
    executedCount: 158,
    status: 'active'
  },
  {
    id: '3',
    name: 'Purchase Order Approval',
    description: 'Multi-level approval for POs',
    type: 'approval',
    trigger: 'PO submitted',
    approvers: ['Nora Ahmed', 'Sarah Abdullah', 'Finance Manager'],
    executedCount: 89,
    status: 'active'
  },
  {
    id: '4',
    name: 'Opportunity Stage Update',
    description: 'Auto-update opportunity stages based on activity',
    type: 'automation',
    trigger: 'Activity logged',
    approvers: [],
    executedCount: 234,
    status: 'active'
  },
  {
    id: '5',
    name: 'Invoice Payment Reminder',
    description: 'Send payment reminders for overdue invoices',
    type: 'notification',
    trigger: 'Invoice overdue',
    approvers: [],
    executedCount: 67,
    status: 'active'
  },
  {
    id: '6',
    name: 'Contract Renewal Alert',
    description: 'Notify team 30 days before contract expiry',
    type: 'notification',
    trigger: '30 days before expiry',
    approvers: [],
    executedCount: 12,
    status: 'inactive'
  }
])

const tableColumns = [
  { key: 'name', label: 'Workflow', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'trigger', label: 'Trigger', sortable: false },
  { key: 'approvers', label: 'Approvers', sortable: true },
  { key: 'executed', label: 'Executed', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false }
]

const filteredWorkflows = computed(() => {
  let filtered = workflows.value

  if (typeFilter.value !== 'all') {
    filtered = filtered.filter(wf => wf.type === typeFilter.value)
  }

  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(wf => wf.status === statusFilter.value)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(wf =>
      wf.name.toLowerCase().includes(query) ||
      wf.description.toLowerCase().includes(query)
    )
  }

  return filtered
})

const activeCount = computed(() =>
  workflows.value.filter(wf => wf.status === 'active').length
)

const pendingCount = computed(() =>
  workflows.value.filter(wf => wf.type === 'approval' && wf.status === 'active').length
)

function formatType(type: Workflow['type']): string {
  const types: Record<Workflow['type'], string> = {
    approval: 'Approval',
    notification: 'Notification',
    automation: 'Automation'
  }
  return types[type]
}

function formatStatus(status: Workflow['status']): string {
  const statuses: Record<Workflow['status'], string> = {
    active: 'Active',
    inactive: 'Inactive'
  }
  return statuses[status]
}

function getTypeVariant(type: Workflow['type']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Workflow['type'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    approval: 'warning',
    notification: 'info',
    automation: 'primary'
  }
  return variants[type]
}

function getStatusVariant(status: Workflow['status']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Workflow['status'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    active: 'success',
    inactive: 'default'
  }
  return variants[status]
}

function createWorkflow() {
  info('Create new workflow functionality')
}

function viewWorkflow(workflow: Workflow) {
  info(`Viewing workflow: ${workflow.name}`)
}

function editWorkflow(workflow: Workflow) {
  info(`Editing workflow: ${workflow.name}`)
}

function toggleStatus(workflow: Workflow) {
  workflow.status = workflow.status === 'active' ? 'inactive' : 'active'
  success(`Workflow ${workflow.status === 'active' ? 'activated' : 'deactivated'}`)
}
</script>

<style scoped>
.workflows-view {
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

.text-warning {
  color: var(--color-warning) !important;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 350px;
}

.workflow-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.workflow-desc {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.trigger-cell {
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
  color: var(--color-text-primary);
  min-width: 220px;
}

.approvers-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
}

.approver-count {
  font-weight: 600;
  font-size: 1.15rem;
  color: var(--color-primary);
}

.executed-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 1rem 0.75rem;
}

.executed-count {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
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

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}
</style>
