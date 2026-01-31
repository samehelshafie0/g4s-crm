<template>
  <div class="security-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search roles & users..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createRole">
          <span>➕</span>
          New Role
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="role">Roles</option>
          <option value="user">Users</option>
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
          <span class="stat-label">Roles</span>
          <span class="stat-value">{{ rolesCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Users</span>
          <span class="stat-value text-success">{{ usersCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active</span>
          <span class="stat-value text-primary">{{ activeCount }}</span>
        </div>
      </div>
    </div>

    <BaseTable
      :columns="tableColumns"
      :data="filteredItems"
      :loading="false"
    >
      <template #cell-name="{ row }">
        <div class="name-cell">
          <div class="item-name">{{ row.name }}</div>
          <div class="item-desc">{{ row.description }}</div>
        </div>
      </template>

      <template #cell-type="{ row }">
        <BaseBadge :variant="getTypeVariant(row.type)">
          {{ formatType(row.type) }}
        </BaseBadge>
      </template>

      <template #cell-permissions="{ row }">
        <div class="permissions-cell">
          <span class="perm-count">{{ row.permissions.length }}</span>
          <span class="text-secondary">permissions</span>
        </div>
      </template>

      <template #cell-users="{ row }">
        <div class="users-cell">
          <span class="user-count">{{ row.usersCount || 0 }}</span>
          <span class="text-secondary">users</span>
        </div>
      </template>

      <template #cell-status="{ row }">
        <BaseBadge :variant="getStatusVariant(row.status)">
          {{ formatStatus(row.status) }}
        </BaseBadge>
      </template>

      <template #cell-actions="{ row }">
        <div class="action-buttons">
          <button class="action-btn" @click="viewItem(row)" title="View">
            👁️
          </button>
          <button class="action-btn" @click="editItem(row)" title="Edit">
            ✏️
          </button>
          <button class="action-btn" @click="managePermissions(row)" title="Permissions">
            🔐
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

const { info } = useToast()

interface SecurityItem {
  id: string
  name: string
  description: string
  type: 'role' | 'user'
  permissions: string[]
  usersCount?: number
  status: 'active' | 'inactive'
}

const searchQuery = ref('')
const typeFilter = ref('all')
const statusFilter = ref('all')

const securityItems = ref<SecurityItem[]>([
  {
    id: '1',
    name: 'Administrator',
    description: 'Full system access and configuration',
    type: 'role',
    permissions: ['all'],
    usersCount: 3,
    status: 'active'
  },
  {
    id: '2',
    name: 'Sales Manager',
    description: 'Manage sales pipeline, quotes, and customers',
    type: 'role',
    permissions: ['customers.read', 'customers.write', 'opportunities.all', 'quotes.all', 'reports.sales'],
    usersCount: 5,
    status: 'active'
  },
  {
    id: '3',
    name: 'Finance Team',
    description: 'Financial operations and reporting',
    type: 'role',
    permissions: ['invoices.all', 'procurement.read', 'reports.finance', 'price-books.read'],
    usersCount: 4,
    status: 'active'
  },
  {
    id: '4',
    name: 'Operations Coordinator',
    description: 'Project and warehouse management',
    type: 'role',
    permissions: ['projects.all', 'inventory.all', 'procurement.all', 'manpower.read'],
    usersCount: 8,
    status: 'active'
  },
  {
    id: '5',
    name: 'Sales Representative',
    description: 'Basic sales operations',
    type: 'role',
    permissions: ['customers.read', 'opportunities.read', 'opportunities.write', 'quotes.create'],
    usersCount: 12,
    status: 'active'
  },
  {
    id: '6',
    name: 'Guest / Read-Only',
    description: 'View-only access to selected modules',
    type: 'role',
    permissions: ['customers.read', 'opportunities.read', 'reports.view'],
    usersCount: 2,
    status: 'inactive'
  }
])

const tableColumns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'permissions', label: 'Permissions', sortable: true },
  { key: 'users', label: 'Users', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false }
]

const filteredItems = computed(() => {
  let filtered = securityItems.value

  if (typeFilter.value !== 'all') {
    filtered = filtered.filter(item => item.type === typeFilter.value)
  }

  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(item => item.status === statusFilter.value)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(item =>
      item.name.toLowerCase().includes(query) ||
      item.description.toLowerCase().includes(query)
    )
  }

  return filtered
})

const rolesCount = computed(() =>
  securityItems.value.filter(item => item.type === 'role').length
)

const usersCount = computed(() =>
  securityItems.value.reduce((sum, item) => sum + (item.usersCount || 0), 0)
)

const activeCount = computed(() =>
  securityItems.value.filter(item => item.status === 'active').length
)

function formatType(type: SecurityItem['type']): string {
  const types: Record<SecurityItem['type'], string> = {
    role: 'Role',
    user: 'User'
  }
  return types[type]
}

function formatStatus(status: SecurityItem['status']): string {
  const statuses: Record<SecurityItem['status'], string> = {
    active: 'Active',
    inactive: 'Inactive'
  }
  return statuses[status]
}

function getTypeVariant(type: SecurityItem['type']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<SecurityItem['type'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    role: 'primary',
    user: 'info'
  }
  return variants[type]
}

function getStatusVariant(status: SecurityItem['status']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<SecurityItem['status'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    active: 'success',
    inactive: 'default'
  }
  return variants[status]
}

function createRole() {
  info('Create new role functionality')
}

function viewItem(item: SecurityItem) {
  info(`Viewing: ${item.name}`)
}

function editItem(item: SecurityItem) {
  info(`Editing: ${item.name}`)
}

function managePermissions(item: SecurityItem) {
  info(`Managing permissions for: ${item.name}`)
}
</script>

<style scoped>
.security-view {
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

.text-primary {
  color: var(--color-primary) !important;
}

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 300px;
}

.item-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.item-desc {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.permissions-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
}

.perm-count {
  font-weight: 600;
  font-size: 1.15rem;
  color: var(--color-primary);
}

.users-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
}

.user-count {
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

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}
</style>
