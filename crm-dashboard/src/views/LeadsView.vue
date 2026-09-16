<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus,
  Search,
  Pencil,
  Trash2,
  X,
  TrendingUp,
  Target,
  Trophy,
  DollarSign,
  LayoutGrid,
  List,
  Calendar,
  User,
  FileText,
} from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useOpportunitiesStore } from '@/stores/opportunities'
import type { Opportunity, OpportunityStage, ServiceType } from '@/types'

import { customersService, usersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'

const router = useRouter()
const oppStore = useOpportunitiesStore()

type ViewMode = 'pipeline' | 'table'

const viewMode = ref<ViewMode>('pipeline')
const searchQuery = ref('')
const stageFilter = ref<OpportunityStage | 'all'>('all')
const showAddModal = ref(false)
const editingOpportunity = ref<Opportunity | null>(null)

const stageConfig: Record<OpportunityStage, { label: string; color: string }> = {
  qualification: { label: 'Qualification', color: 'var(--color-primary)' },
  proposal: { label: 'Proposal', color: 'var(--color-warning)' },
  negotiation: { label: 'Negotiation', color: '#8b5cf6' },
  'closed-won': { label: 'Closed Won', color: 'var(--color-success)' },
  'closed-lost': { label: 'Closed Lost', color: 'var(--color-danger)' },
}

const stageBadgeClass = (s: OpportunityStage) =>
  ({
    qualification: 'badge-primary',
    proposal: 'badge-warning',
    negotiation: 'badge-negotiation',
    'closed-won': 'badge-success',
    'closed-lost': 'badge-danger',
  })[s]

const serviceTypeLabels: Record<ServiceType, string> = {
  cctv: 'CCTV',
  'access-control': 'Access Control',
  'intrusion-detection': 'Intrusion Detection',
  'fire-alarm': 'Fire Alarm',
  networking: 'Networking',
  'it-solutions': 'IT Solutions',
  guarding: 'Guarding',
  monitoring: 'Monitoring',
  maintenance: 'Maintenance',
  consulting: 'Consulting',
}

const allServiceTypes = Object.keys(serviceTypeLabels) as ServiceType[]

const customers = ref<{id: string; name: string}[]>([])
const salesTeam = ref<{id: string; name: string}[]>([])
const preSales = ref<{id: string; name: string}[]>([])

const defaultForm = (): Omit<Opportunity, 'id' | 'createdAt' | 'updatedAt' | 'estimatedMargin' | 'quoteIds'> => ({
  title: '',
  customerId: '',
  customerName: '',
  stage: 'qualification',
  serviceTypes: [],
  estimatedValue: 0,
  estimatedCost: 0,
  winProbability: 30,
  salesExecutiveId: '',
  salesExecutiveName: '',
  preSalesId: '',
  preSalesName: '',
  expectedCloseDate: '',
  notes: '',
})

const form = ref(defaultForm())

const opportunities = computed(() => oppStore.opportunities)
const loading = computed(() => oppStore.loading)

onMounted(async () => {
  try {
    const [customerRows, userRows] = await Promise.all([allPages(customersService.list), usersService.lookup()])
    customers.value = customerRows.map(c => ({ id: c.id, name: c.companyName }))
    const names = userRows.data.map(u => ({ id: u.id, name: `${u.firstName} ${u.lastName}`, role: u.role }))
    salesTeam.value = names.filter(u => ['admin', 'sales_manager', 'sales_executive'].includes(u.role))
    preSales.value = names.filter(u => ['admin', 'pre_sales'].includes(u.role))
    await oppStore.fetchOpportunities({ limit: 100 })
  } catch (error) { window.alert(errorMessage(error)) }
})

function formatCurrency(val: number): string {
  if (val >= 1_000_000) return `SAR ${(val / 1_000_000).toFixed(1)}M`
  if (val >= 1_000) return `SAR ${(val / 1_000).toFixed(0)}K`
  return `SAR ${val.toLocaleString()}`
}

function formatCurrencyFull(val: number): string {
  return `SAR ${val.toLocaleString()}`
}

const filteredOpportunities = computed(() =>
  opportunities.value.filter((o) => {
    if (searchQuery.value && !o.title.toLowerCase().includes(searchQuery.value.toLowerCase()) && !o.customerName.toLowerCase().includes(searchQuery.value.toLowerCase())) return false
    if (stageFilter.value !== 'all' && o.stage !== stageFilter.value) return false
    return true
  }),
)

const pipelineStages: OpportunityStage[] = ['qualification', 'proposal', 'negotiation', 'closed-won', 'closed-lost']

const pipelineColumns = computed(() =>
  pipelineStages.map((stage) => {
    const items = filteredOpportunities.value.filter((o) => o.stage === stage)
    return {
      stage,
      ...stageConfig[stage],
      count: items.length,
      total: items.reduce((s, o) => s + o.estimatedValue, 0),
      items,
    }
  }),
)

// KPIs
const totalPipelineValue = computed(() =>
  opportunities.value
    .filter((o) => !['closed-won', 'closed-lost'].includes(o.stage))
    .reduce((s, o) => s + o.estimatedValue, 0),
)
const qualifiedLeads = computed(() =>
  opportunities.value.filter((o) => !['closed-won', 'closed-lost'].includes(o.stage)).length,
)
const wonDeals = computed(() => opportunities.value.filter((o) => o.stage === 'closed-won'))
const closedDeals = computed(() => opportunities.value.filter((o) => o.stage === 'closed-won' || o.stage === 'closed-lost'))
const winRate = computed(() =>
  closedDeals.value.length ? Math.round((wonDeals.value.length / closedDeals.value.length) * 100) : 0,
)
const avgDealSize = computed(() => {
  const active = opportunities.value.filter((o) => o.stage !== 'closed-lost')
  return active.length ? active.reduce((s, o) => s + o.estimatedValue, 0) / active.length : 0
})

function openAddModal() {
  editingOpportunity.value = null
  form.value = defaultForm()
  showAddModal.value = true
}

function openEditModal(o: Opportunity) {
  editingOpportunity.value = o
  form.value = {
    title: o.title,
    customerId: o.customerId,
    customerName: o.customerName,
    stage: o.stage,
    serviceTypes: [...o.serviceTypes],
    estimatedValue: o.estimatedValue,
    estimatedCost: o.estimatedCost,
    winProbability: o.winProbability,
    salesExecutiveId: o.salesExecutiveId,
    salesExecutiveName: o.salesExecutiveName,
    preSalesId: o.preSalesId,
    preSalesName: o.preSalesName,
    expectedCloseDate: o.expectedCloseDate?.slice(0, 10),
    notes: o.notes,
  }
  showAddModal.value = true
}

function onCustomerChange() {
  const found = customers.value.find((c) => c.id === form.value.customerId)
  form.value.customerName = found?.name ?? ''
}
function onSalesChange() {
  const found = salesTeam.value.find((u) => u.id === form.value.salesExecutiveId)
  form.value.salesExecutiveName = found?.name ?? ''
}
function onPreSalesChange() {
  const found = preSales.value.find((u) => u.id === form.value.preSalesId)
  form.value.preSalesName = found?.name ?? ''
}

function toggleServiceType(st: ServiceType) {
  const idx = form.value.serviceTypes.indexOf(st)
  if (idx === -1) form.value.serviceTypes.push(st)
  else form.value.serviceTypes.splice(idx, 1)
}

async function saveOpportunity() {
  const margin = form.value.estimatedValue - form.value.estimatedCost
  try {
    if (editingOpportunity.value) {
      await oppStore.updateOpportunity(editingOpportunity.value.id, {
        ...form.value,
        estimatedMargin: margin,
      })
    } else {
      await oppStore.addOpportunity({
        ...form.value,
        estimatedMargin: margin,
      })
    }
    showAddModal.value = false
  } catch (e) {
    window.alert(errorMessage(e))
  }
}

async function deleteOpportunity(id: string) {
  try {
    await oppStore.deleteOpportunity(id)
  } catch (e) {
    window.alert(errorMessage(e))
  }
}

function createQuoteFromOpp(opp: Opportunity) {
  router.push({ path: '/quotes', query: { newFromOpp: opp.id } })
}

function probClass(prob: number): string {
  if (prob >= 70) return 'prob-high'
  if (prob >= 40) return 'prob-mid'
  return 'prob-low'
}
</script>

<template>
  <div class="leads-view">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Opportunities & Leads</h1>
        <p class="page-header-subtitle">{{ filteredOpportunities.length }} opportunities in pipeline</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <Plus :size="16" />
        New Opportunity
      </button>
    </div>

    <!-- KPI Row -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="card stat-card">
        <div class="stat-icon-wrap bg-primary-light">
          <DollarSign :size="20" class="text-primary" />
        </div>
        <div class="stat-label">Total Pipeline Value</div>
        <div class="stat-value">{{ formatCurrency(totalPipelineValue) }}</div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon-wrap bg-warning-light">
          <Target :size="20" class="text-warning" />
        </div>
        <div class="stat-label">Qualified Leads</div>
        <div class="stat-value">{{ qualifiedLeads }}</div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon-wrap bg-success-light">
          <Trophy :size="20" class="text-success" />
        </div>
        <div class="stat-label">Win Rate</div>
        <div class="stat-value">{{ winRate }}%</div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon-wrap bg-primary-light">
          <TrendingUp :size="20" class="text-primary" />
        </div>
        <div class="stat-label">Avg Deal Size</div>
        <div class="stat-value">{{ formatCurrency(avgDealSize) }}</div>
      </div>
    </div>

    <!-- Toolbar + Tabs -->
    <div class="card mb-6">
      <div class="card-body leads-toolbar">
        <div class="search-input toolbar-search">
          <Search :size="16" class="search-icon" />
          <input v-model="searchQuery" type="text" class="form-input" placeholder="Search opportunities…" />
        </div>
        <select v-model="stageFilter" class="form-select toolbar-filter">
          <option value="all">All Stages</option>
          <option v-for="s in pipelineStages" :key="s" :value="s">{{ stageConfig[s].label }}</option>
        </select>
        <div class="view-toggle">
          <button
            class="btn btn-sm" :class="viewMode === 'pipeline' ? 'btn-primary' : 'btn-secondary'"
            @click="viewMode = 'pipeline'"
          >
            <LayoutGrid :size="14" /> Pipeline
          </button>
          <button
            class="btn btn-sm" :class="viewMode === 'table' ? 'btn-primary' : 'btn-secondary'"
            @click="viewMode = 'table'"
          >
            <List :size="14" /> Table
          </button>
        </div>
      </div>
    </div>

    <!-- Pipeline / Kanban View -->
    <div v-if="viewMode === 'pipeline'" class="pipeline-board">
      <div v-for="col in pipelineColumns" :key="col.stage" class="pipeline-column">
        <div class="pipeline-column-header" :style="{ borderTopColor: col.color }">
          <div class="pipeline-column-title">
            <span>{{ col.label }}</span>
            <span class="tab-count">{{ col.count }}</span>
          </div>
          <div class="pipeline-column-total">{{ formatCurrency(col.total) }}</div>
        </div>
        <div class="pipeline-column-body">
          <div v-for="opp in col.items" :key="opp.id" class="pipeline-card card card-clickable" @click="openEditModal(opp)">
            <div class="pipeline-card-title">{{ opp.title }}</div>
            <div class="pipeline-card-customer">{{ opp.customerName }}</div>
            <div class="pipeline-card-meta">
              <span class="pipeline-card-value">{{ formatCurrency(opp.estimatedValue) }}</span>
              <span class="pipeline-card-prob" :class="probClass(opp.winProbability)">{{ opp.winProbability }}%</span>
            </div>
            <div class="pipeline-card-footer">
              <span class="pipeline-card-exec"><User :size="12" /> {{ opp.salesExecutiveName }}</span>
              <span v-if="opp.expectedCloseDate" class="pipeline-card-date"><Calendar :size="12" /> {{ opp.expectedCloseDate }}</span>
            </div>
          </div>
          <div v-if="col.items.length === 0" class="pipeline-empty">No opportunities</div>
        </div>
      </div>
    </div>

    <!-- Table View -->
    <div v-else class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Customer</th>
            <th>Stage</th>
            <th>Services</th>
            <th class="text-right">Value</th>
            <th class="text-center">Prob.</th>
            <th>Sales Exec</th>
            <th>Close Date</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in filteredOpportunities" :key="o.id">
            <td class="font-semibold">{{ o.title }}</td>
            <td>{{ o.customerName }}</td>
            <td>
              <span class="badge badge-dot" :class="stageBadgeClass(o.stage)">
                {{ stageConfig[o.stage].label }}
              </span>
            </td>
            <td>
              <div class="service-badges">
                <span v-for="st in o.serviceTypes.slice(0, 2)" :key="st" class="badge badge-neutral">
                  {{ serviceTypeLabels[st] }}
                </span>
                <span v-if="o.serviceTypes.length > 2" class="badge badge-neutral">+{{ o.serviceTypes.length - 2 }}</span>
              </div>
            </td>
            <td class="text-right font-semibold text-mono">{{ formatCurrencyFull(o.estimatedValue) }}</td>
            <td class="text-center">
              <span class="prob-pill" :class="probClass(o.winProbability)">{{ o.winProbability }}%</span>
            </td>
            <td>{{ o.salesExecutiveName }}</td>
            <td>{{ o.expectedCloseDate }}</td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="Create Quote" @click="createQuoteFromOpp(o)">
                  <FileText :size="15" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEditModal(o)">
                  <Pencil :size="15" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm text-danger" title="Delete" @click="deleteOpportunity(o.id)">
                  <Trash2 :size="15" />
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="filteredOpportunities.length === 0">
            <td colspan="9">
              <div class="table-empty">
                <div class="empty-icon"><Target :size="40" /></div>
                <div class="empty-text">No opportunities found</div>
                <div class="empty-subtext">Try adjusting your search or filters.</div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Opportunity Modal -->
    <Teleport to="body">
      <div v-if="showAddModal" class="modal-backdrop" @click.self="showAddModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h3 class="modal-title">{{ editingOpportunity ? 'Edit Opportunity' : 'New Opportunity' }}</h3>
            <button class="modal-close" @click="showAddModal = false"><X :size="18" /></button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Title <span class="required">*</span></label>
              <input v-model="form.title" type="text" class="form-input" placeholder="Opportunity title" />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Customer</label>
                <select v-model="form.customerId" class="form-select" @change="onCustomerChange">
                  <option value="">Select customer…</option>
                  <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Stage</label>
                <select v-model="form.stage" class="form-select">
                  <option v-for="s in pipelineStages" :key="s" :value="s">{{ stageConfig[s].label }}</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">Service Types</label>
              <div class="service-type-grid">
                <label v-for="st in allServiceTypes" :key="st" class="form-checkbox">
                  <input type="checkbox" :checked="form.serviceTypes.includes(st)" @change="toggleServiceType(st)" />
                  {{ serviceTypeLabels[st] }}
                </label>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Estimated Value (SAR)</label>
                <input v-model.number="form.estimatedValue" type="number" class="form-input" placeholder="0" />
              </div>
              <div class="form-group">
                <label class="form-label">Estimated Cost (SAR)</label>
                <input v-model.number="form.estimatedCost" type="number" class="form-input" placeholder="0" />
              </div>
              <div class="form-group">
                <label class="form-label">Win Probability (%)</label>
                <input v-model.number="form.winProbability" type="number" min="0" max="100" class="form-input" />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Sales Executive</label>
                <select v-model="form.salesExecutiveId" class="form-select" @change="onSalesChange">
                  <option value="">Select…</option>
                  <option v-for="u in salesTeam" :key="u.id" :value="u.id">{{ u.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Pre-Sales</label>
                <select v-model="form.preSalesId" class="form-select" @change="onPreSalesChange">
                  <option value="">Select…</option>
                  <option v-for="u in preSales" :key="u.id" :value="u.id">{{ u.name }}</option>
                </select>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">Expected Close Date</label>
              <input v-model="form.expectedCloseDate" type="date" class="form-input" />
            </div>

            <div class="form-group">
              <label class="form-label">Notes</label>
              <textarea v-model="form.notes" class="form-textarea" placeholder="Additional notes…" rows="3" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showAddModal = false">Cancel</button>
            <button class="btn btn-primary" :disabled="!form.title" @click="saveOpportunity">
              {{ editingOpportunity ? 'Save Changes' : 'Create Opportunity' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.leads-view {
  padding: var(--space-6);
}

/* KPI stat icon */
.stat-card {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.stat-icon-wrap {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--space-2);
}

/* Toolbar */
.leads-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}
.toolbar-search {
  flex: 1;
  min-width: 220px;
}
.toolbar-filter {
  width: auto;
  min-width: 160px;
}
.view-toggle {
  display: flex;
  gap: var(--space-1);
  margin-left: auto;
}

/* Pipeline Kanban */
.pipeline-board {
  display: flex;
  gap: var(--space-4);
  overflow-x: auto;
  padding-bottom: var(--space-4);
  align-items: flex-start;
}

.pipeline-column {
  flex: 1;
  min-width: 240px;
  max-width: 300px;
  background-color: var(--color-neutral-50);
  border-radius: var(--radius-xl);
  display: flex;
  flex-direction: column;
}

.pipeline-column-header {
  padding: var(--space-4);
  border-top: 3px solid transparent;
  border-radius: var(--radius-xl) var(--radius-xl) 0 0;
}

.pipeline-column-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.pipeline-column-total {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  margin-top: var(--space-1);
  font-weight: var(--font-medium);
}

.pipeline-column-body {
  padding: 0 var(--space-3) var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  min-height: 120px;
}

.pipeline-card {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.pipeline-card:hover {
  transform: translateY(-1px);
}

.pipeline-card-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
  margin-bottom: var(--space-1);
  line-height: var(--leading-tight);
}

.pipeline-card-customer {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  margin-bottom: var(--space-2);
}

.pipeline-card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.pipeline-card-value {
  font-size: var(--text-sm);
  font-weight: var(--font-bold);
  color: var(--color-neutral-800);
  font-family: var(--font-mono);
}

.pipeline-card-prob,
.prob-pill {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  padding: 1px var(--space-2);
  border-radius: var(--radius-full);
}
.prob-high {
  background-color: var(--color-success-light);
  color: var(--color-success-dark);
}
.prob-mid {
  background-color: var(--color-warning-light);
  color: var(--color-warning-dark);
}
.prob-low {
  background-color: var(--color-neutral-100);
  color: var(--color-neutral-600);
}

.pipeline-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 11px;
  color: var(--color-neutral-500);
  gap: var(--space-2);
}

.pipeline-card-exec,
.pipeline-card-date {
  display: flex;
  align-items: center;
  gap: 3px;
  white-space: nowrap;
}

.pipeline-empty {
  text-align: center;
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  padding: var(--space-6) 0;
}

/* Badge for negotiation stage */
.badge-negotiation {
  background-color: #ede9fe;
  color: #6d28d9;
}
.badge-negotiation.badge-dot::before {
  background-color: #8b5cf6;
}

/* Service badges */
.service-badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1);
}

/* Service type checkbox grid */
.service-type-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: var(--space-2);
  padding: var(--space-3);
  background-color: var(--color-neutral-50);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-neutral-200);
}

@media (max-width: 768px) {
  .leads-toolbar {
    flex-direction: column;
  }
  .toolbar-search,
  .toolbar-filter {
    width: 100%;
    min-width: 0;
  }
  .view-toggle {
    margin-left: 0;
    width: 100%;
    justify-content: stretch;
  }
  .view-toggle .btn {
    flex: 1;
  }
  .pipeline-board {
    flex-direction: column;
  }
  .pipeline-column {
    max-width: 100%;
    min-width: 0;
  }
}
</style>
