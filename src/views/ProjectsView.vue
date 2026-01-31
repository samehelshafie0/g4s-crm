<template>
  <div class="projects-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search projects..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createProject">
          <span>➕</span>
          New Project
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="planning">Planning</option>
          <option value="in-progress">In Progress</option>
          <option value="on-hold">On Hold</option>
          <option value="completed">Completed</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Priority:</label>
        <select v-model="priorityFilter" class="filter-select">
          <option value="all">All Priorities</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total</span>
          <span class="stat-value">{{ filteredProjects.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active</span>
          <span class="stat-value text-success">{{ activeCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Budget</span>
          <span class="stat-value">{{ formatCurrency(totalBudget) }}</span>
        </div>
      </div>
    </div>

    <BaseTable
      :columns="tableColumns"
      :data="filteredProjects"
      :loading="false"
    >
      <template #cell-name="{ row }">
        <div class="name-cell">
          <div class="project-name">{{ row.name }}</div>
          <div class="project-code">{{ row.code }}</div>
        </div>
      </template>

      <template #cell-client="{ row }">
        <div class="client-cell">
          <div class="client-name">{{ row.clientName }}</div>
          <div class="text-secondary">{{ row.industry }}</div>
        </div>
      </template>

      <template #cell-timeline="{ row }">
        <div class="timeline-cell">
          <div class="date-row">
            <span class="date-label">Start:</span>
            <span>{{ formatDate(row.startDate) }}</span>
          </div>
          <div class="date-row">
            <span class="date-label">End:</span>
            <span>{{ formatDate(row.endDate) }}</span>
          </div>
        </div>
      </template>

      <template #cell-progress="{ row }">
        <div class="progress-cell">
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: row.progress + '%' }"
              :class="getProgressClass(row.progress)"
            ></div>
          </div>
          <span class="progress-text">{{ row.progress }}%</span>
        </div>
      </template>

      <template #cell-budget="{ row }">
        <div class="budget-cell">
          <div class="budget-value">{{ formatCurrency(row.budget) }}</div>
          <div class="spent-value" :class="{ 'over-budget': row.spent > row.budget }">
            {{ formatCurrency(row.spent) }} spent
          </div>
        </div>
      </template>

      <template #cell-team="{ row }">
        <div class="team-cell">
          <span class="team-count">{{ row.teamSize }}</span>
          <span class="text-secondary">members</span>
        </div>
      </template>

      <template #cell-status="{ row }">
        <BaseBadge :variant="getStatusVariant(row.status)">
          {{ formatStatus(row.status) }}
        </BaseBadge>
      </template>

      <template #cell-priority="{ row }">
        <BaseBadge :variant="getPriorityVariant(row.priority)">
          {{ formatPriority(row.priority) }}
        </BaseBadge>
      </template>

      <template #cell-actions="{ row }">
        <div class="action-buttons">
          <button class="action-btn" @click="viewProject(row)" title="View Details">
            👁️
          </button>
          <button class="action-btn" @click="editProject(row)" title="Edit">
            ✏️
          </button>
          <button class="action-btn" @click="viewTasks(row)" title="Tasks">
            ✅
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

interface Project {
  id: string
  code: string
  name: string
  clientName: string
  industry: string
  status: 'planning' | 'in-progress' | 'on-hold' | 'completed'
  priority: 'high' | 'medium' | 'low'
  startDate: string
  endDate: string
  progress: number
  budget: number
  spent: number
  teamSize: number
}

// Filters
const searchQuery = ref('')
const statusFilter = ref('all')
const priorityFilter = ref('all')

// Sample data
const projects = ref<Project[]>([
  {
    id: '1',
    code: 'PRJ-2024-001',
    name: 'ERP Implementation Phase 1',
    clientName: 'Acme Corporation',
    industry: 'Manufacturing',
    status: 'in-progress',
    priority: 'high',
    startDate: '2024-01-15',
    endDate: '2024-06-30',
    progress: 65,
    budget: 850000,
    spent: 520000,
    teamSize: 12
  },
  {
    id: '2',
    code: 'PRJ-2024-002',
    name: 'Website Redesign',
    clientName: 'TechStart Solutions',
    industry: 'Technology',
    status: 'in-progress',
    priority: 'medium',
    startDate: '2024-02-01',
    endDate: '2024-04-15',
    progress: 80,
    budget: 120000,
    spent: 95000,
    teamSize: 5
  },
  {
    id: '3',
    code: 'PRJ-2024-003',
    name: 'Cloud Migration',
    clientName: 'Global Finance Inc',
    industry: 'Finance',
    status: 'planning',
    priority: 'high',
    startDate: '2024-04-01',
    endDate: '2024-10-31',
    progress: 15,
    budget: 1200000,
    spent: 50000,
    teamSize: 15
  },
  {
    id: '4',
    code: 'PRJ-2024-004',
    name: 'Mobile App Development',
    clientName: 'Retail Plus',
    industry: 'Retail',
    status: 'in-progress',
    priority: 'medium',
    startDate: '2024-01-20',
    endDate: '2024-05-20',
    progress: 70,
    budget: 280000,
    spent: 195000,
    teamSize: 8
  },
  {
    id: '5',
    code: 'PRJ-2023-089',
    name: 'Data Center Setup',
    clientName: 'MedHealth Systems',
    industry: 'Healthcare',
    status: 'completed',
    priority: 'high',
    startDate: '2023-08-01',
    endDate: '2024-01-31',
    progress: 100,
    budget: 950000,
    spent: 920000,
    teamSize: 10
  },
  {
    id: '6',
    code: 'PRJ-2024-005',
    name: 'CRM Customization',
    clientName: 'Sales Force Pro',
    industry: 'Sales',
    status: 'on-hold',
    priority: 'low',
    startDate: '2024-02-15',
    endDate: '2024-07-15',
    progress: 35,
    budget: 150000,
    spent: 60000,
    teamSize: 4
  },
  {
    id: '7',
    code: 'PRJ-2024-006',
    name: 'Security Audit & Compliance',
    clientName: 'SecureBank',
    industry: 'Banking',
    status: 'in-progress',
    priority: 'high',
    startDate: '2024-03-01',
    endDate: '2024-05-31',
    progress: 55,
    budget: 320000,
    spent: 180000,
    teamSize: 6
  }
])

// Table columns
const tableColumns = [
  { key: 'name', label: 'Project', sortable: true },
  { key: 'client', label: 'Client', sortable: true },
  { key: 'timeline', label: 'Timeline', sortable: false },
  { key: 'progress', label: 'Progress', sortable: true },
  { key: 'budget', label: 'Budget', sortable: true },
  { key: 'team', label: 'Team', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'priority', label: 'Priority', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false }
]

// Computed
const filteredProjects = computed(() => {
  let filtered = projects.value

  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(project => project.status === statusFilter.value)
  }

  if (priorityFilter.value !== 'all') {
    filtered = filtered.filter(project => project.priority === priorityFilter.value)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(project =>
      project.name.toLowerCase().includes(query) ||
      project.code.toLowerCase().includes(query) ||
      project.clientName.toLowerCase().includes(query)
    )
  }

  return filtered
})

const activeCount = computed(() =>
  projects.value.filter(p => p.status === 'in-progress').length
)

const totalBudget = computed(() =>
  projects.value.reduce((sum, p) => sum + p.budget, 0)
)

// Helper functions
function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function formatStatus(status: Project['status']): string {
  const statuses: Record<Project['status'], string> = {
    planning: 'Planning',
    'in-progress': 'In Progress',
    'on-hold': 'On Hold',
    completed: 'Completed'
  }
  return statuses[status]
}

function formatPriority(priority: Project['priority']): string {
  const priorities: Record<Project['priority'], string> = {
    high: 'High',
    medium: 'Medium',
    low: 'Low'
  }
  return priorities[priority]
}

function getStatusVariant(status: Project['status']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Project['status'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    planning: 'info',
    'in-progress': 'primary',
    'on-hold': 'warning',
    completed: 'success'
  }
  return variants[status]
}

function getPriorityVariant(priority: Project['priority']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<Project['priority'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    high: 'danger',
    medium: 'warning',
    low: 'default'
  }
  return variants[priority]
}

function getProgressClass(progress: number): string {
  if (progress >= 75) return 'high'
  if (progress >= 50) return 'medium'
  return 'low'
}

// Actions
function createProject() {
  info('Create new project functionality')
}

function viewProject(project: Project) {
  info(`Viewing project: ${project.name}`)
}

function editProject(project: Project) {
  info(`Editing project: ${project.name}`)
}

function viewTasks(project: Project) {
  info(`Viewing tasks for: ${project.name}`)
}
</script>

<style scoped>
.projects-view {
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
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 300px;
}

.project-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.project-code {
  font-size: 0.95rem;
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
}

.client-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 240px;
}

.client-name {
  font-size: 1.05rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.timeline-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  min-width: 200px;
}

.date-row {
  font-size: 1rem;
  display: flex;
  gap: 0.75rem;
}

.date-label {
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 45px;
}

.progress-cell {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0.75rem;
  min-width: 200px;
}

.progress-bar {
  flex: 1;
  height: 12px;
  background: var(--color-border);
  border-radius: 6px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 6px;
  transition: width 0.3s ease;
}

.progress-fill.high {
  background: var(--color-success);
}

.progress-fill.medium {
  background: var(--color-warning);
}

.progress-fill.low {
  background: var(--color-danger);
}

.progress-text {
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--color-text-primary);
  min-width: 50px;
}

.budget-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 200px;
}

.budget-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-primary-light);
  padding: 0.5rem 1rem;
  border-radius: 6px;
}

.spent-value {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

.spent-value.over-budget {
  color: var(--color-danger);
  font-weight: 600;
}

.team-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
}

.team-count {
  font-weight: 600;
  font-size: 1.15rem;
  color: var(--color-primary);
}

.text-secondary {
  color: var(--color-text-secondary);
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
