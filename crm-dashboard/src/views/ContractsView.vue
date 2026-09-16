<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  Plus,
  Search,
  Eye,
  Pencil,
  Trash2,
  X,
  FileText,
  CheckCircle2,
  AlertTriangle,
  Clock,
  RefreshCw,
  DollarSign,
  CalendarDays,
  Check,
  XCircle,
} from 'lucide-vue-next'
import type { Contract, ContractType, ContractStatus } from '@/types'
import { useContractsStore } from '@/stores/contracts'

function uid(): string {
  return Math.random().toString(36).slice(2, 11)
}

function formatSAR(v: number): string {
  return 'SAR ' + v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

const contractStore = useContractsStore()
onMounted(() => contractStore.fetchContracts())

const typeLabels: Record<ContractType, string> = {
  sales: 'Sales',
  maintenance: 'Maintenance',
  service: 'Service',
  project: 'Project',
  subscription: 'Subscription',
}

const typeBadge: Record<ContractType, string> = {
  sales: 'badge-primary',
  maintenance: 'badge-info',
  service: 'badge-success',
  project: 'badge-warning',
  subscription: 'badge-gray',
}

const statusLabels: Record<ContractStatus, string> = {
  draft: 'Draft',
  'pending-approval': 'Pending',
  active: 'Active',
  expired: 'Expired',
  terminated: 'Terminated',
  renewed: 'Renewed',
}

const statusBadge: Record<ContractStatus, string> = {
  draft: 'badge-gray',
  'pending-approval': 'badge-warning',
  active: 'badge-success',
  expired: 'badge-danger',
  terminated: 'badge-danger',
  renewed: 'badge-info',
}

const contracts = ref<Contract[]>([
  {
    id: uid(), contractNumber: 'CTR-2025-001', title: 'CCTV Maintenance Agreement',
    customerId: 'c1', customerName: 'Saudi Aramco', type: 'maintenance', status: 'active',
    startDate: '2025-04-01', endDate: '2026-03-31', value: 480000,
    autoRenew: true, renewalNoticeDays: 60,
    terms: 'Quarterly preventive maintenance for all CCTV systems across Dhahran and Ras Tanura sites. Includes 4-hour SLA for critical failures.',
    notes: 'Renewed from CTR-2024-003. Key account — assign senior technicians.',
    createdAt: '2025-03-15T08:00:00Z', updatedAt: '2026-01-20T10:00:00Z',
  },
  {
    id: uid(), contractNumber: 'CTR-2025-002', title: 'Access Control System Supply & Install',
    customerId: 'c5', customerName: 'Ministry of Interior', type: 'project', status: 'active',
    startDate: '2025-06-01', endDate: '2026-05-31', value: 1250000,
    autoRenew: false, renewalNoticeDays: 90,
    terms: 'Supply and installation of ZKTeco access control for 12 government buildings. Includes one-year warranty and training.',
    notes: 'Phase 1 of 3. Government procurement — strict delivery milestones.',
    createdAt: '2025-05-20T08:00:00Z', updatedAt: '2026-02-10T14:00:00Z',
  },
  {
    id: uid(), contractNumber: 'CTR-2026-001', title: 'Security Monitoring Subscription',
    customerId: 'c4', customerName: 'Saudi Telecom Company (STC)', type: 'subscription', status: 'pending-approval',
    startDate: '2026-03-01', endDate: '2027-02-28', value: 360000,
    autoRenew: true, renewalNoticeDays: 30,
    terms: '24/7 remote monitoring of all STC sites. Monthly reports and quarterly reviews.',
    notes: 'Awaiting procurement VP signature. Expected approval by end of Feb.',
    createdAt: '2026-01-15T08:00:00Z', updatedAt: '2026-02-18T09:00:00Z',
  },
  {
    id: uid(), contractNumber: 'CTR-2024-003', title: 'Fire Alarm Maintenance',
    customerId: 'c2', customerName: 'King Faisal Specialist Hospital', type: 'maintenance', status: 'expired',
    startDate: '2024-01-01', endDate: '2024-12-31', value: 195000,
    autoRenew: false, renewalNoticeDays: 60,
    terms: 'Semi-annual inspection and testing of Honeywell fire alarm panels. Emergency call-out within 2 hours.',
    notes: 'Contract expired — renewal proposal sent Jan 2025.',
    createdAt: '2023-11-01T08:00:00Z', updatedAt: '2025-01-15T08:00:00Z',
  },
  {
    id: uid(), contractNumber: 'CTR-2026-002', title: 'Integrated Security Systems',
    customerId: 'c6', customerName: 'Al Rajhi Bank', type: 'sales', status: 'draft',
    startDate: '2026-04-01', endDate: '2027-03-31', value: 875000,
    autoRenew: true, renewalNoticeDays: 45,
    terms: 'Supply of CCTV, access control, and intrusion detection for 50 bank branches. Phased rollout over 12 months.',
    notes: 'Draft — pricing under review with procurement team.',
    createdAt: '2026-02-01T08:00:00Z', updatedAt: '2026-02-22T16:00:00Z',
  },
])

const searchQuery = ref('')
const filterStatus = ref<ContractStatus | ''>('')
const filterType = ref<ContractType | ''>('')

const filteredContracts = computed(() => {
  let list = contracts.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter(c =>
      c.contractNumber.toLowerCase().includes(q) ||
      c.title.toLowerCase().includes(q) ||
      c.customerName.toLowerCase().includes(q),
    )
  }
  if (filterStatus.value) list = list.filter(c => c.status === filterStatus.value)
  if (filterType.value) list = list.filter(c => c.type === filterType.value)
  return list
})

const activeCount = computed(() => contracts.value.filter(c => c.status === 'active').length)
const totalValue = computed(() => contracts.value.filter(c => c.status === 'active').reduce((a, c) => a + c.value, 0))
const expiringSoon = computed(() => {
  const cutoff = new Date()
  cutoff.setDate(cutoff.getDate() + 30)
  return contracts.value.filter(c => c.status === 'active' && new Date(c.endDate) <= cutoff).length
})
const renewedThisYear = computed(() => contracts.value.filter(c => c.status === 'renewed').length)

// Modal state
const showModal = ref(false)
const showViewModal = ref(false)
const editingId = ref<string | null>(null)
const viewingContract = ref<Contract | null>(null)

interface ContractForm {
  contractNumber: string
  title: string
  customerId: string
  customerName: string
  type: ContractType
  status: ContractStatus
  startDate: string
  endDate: string
  value: number
  autoRenew: boolean
  renewalNoticeDays: number
  terms: string
  notes: string
}

const defaultForm = (): ContractForm => ({
  contractNumber: '',
  title: '',
  customerId: '',
  customerName: '',
  type: 'service',
  status: 'draft',
  startDate: new Date().toISOString().slice(0, 10),
  endDate: new Date(Date.now() + 365 * 86400000).toISOString().slice(0, 10),
  value: 0,
  autoRenew: false,
  renewalNoticeDays: 30,
  terms: '',
  notes: '',
})

const form = ref<ContractForm>(defaultForm())

const customerOptions = [
  { id: 'c1', name: 'Saudi Aramco' },
  { id: 'c2', name: 'King Faisal Specialist Hospital' },
  { id: 'c4', name: 'Saudi Telecom Company (STC)' },
  { id: 'c5', name: 'Ministry of Interior' },
  { id: 'c6', name: 'Al Rajhi Bank' },
  { id: 'c3', name: 'NEOM' },
]

function openAddModal() {
  editingId.value = null
  form.value = defaultForm()
  showModal.value = true
}

function openEditModal(c: Contract) {
  editingId.value = c.id
  form.value = {
    contractNumber: c.contractNumber,
    title: c.title,
    customerId: c.customerId,
    customerName: c.customerName,
    type: c.type,
    status: c.status,
    startDate: c.startDate,
    endDate: c.endDate,
    value: c.value,
    autoRenew: c.autoRenew,
    renewalNoticeDays: c.renewalNoticeDays,
    terms: c.terms,
    notes: c.notes,
  }
  showModal.value = true
}

function openViewModal(c: Contract) {
  viewingContract.value = c
  showViewModal.value = true
}

function saveContract() {
  const now = new Date().toISOString()
  const cust = customerOptions.find(o => o.id === form.value.customerId)

  const base = {
    contractNumber: form.value.contractNumber,
    title: form.value.title,
    customerId: form.value.customerId,
    customerName: cust?.name ?? form.value.customerName,
    type: form.value.type,
    status: form.value.status,
    startDate: form.value.startDate,
    endDate: form.value.endDate,
    value: form.value.value,
    autoRenew: form.value.autoRenew,
    renewalNoticeDays: form.value.renewalNoticeDays,
    terms: form.value.terms,
    notes: form.value.notes,
  }

  if (editingId.value) {
    const idx = contracts.value.findIndex(c => c.id === editingId.value)
    if (idx !== -1) {
      contracts.value[idx] = { ...contracts.value[idx], ...base, updatedAt: now } as Contract
    }
  } else {
    contracts.value.push({ id: uid(), ...base, createdAt: now, updatedAt: now })
  }
  showModal.value = false
}

function deleteContract(id: string) {
  contracts.value = contracts.value.filter(c => c.id !== id)
}

function daysRemaining(endDate: string): number {
  const diff = new Date(endDate).getTime() - Date.now()
  return Math.ceil(diff / 86400000)
}
</script>

<template>
  <div class="contracts-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Contracts</h1>
        <p class="page-header-subtitle">{{ filteredContracts.length }} contract{{ filteredContracts.length !== 1 ? 's' : '' }}</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <Plus :size="18" />
        Add Contract
      </button>
    </div>

    <!-- KPI Row -->
    <div class="kpi-row">
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--success"><CheckCircle2 :size="20" /></div>
        <div>
          <div class="kpi-label">Active Contracts</div>
          <div class="kpi-value">{{ activeCount }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--primary"><DollarSign :size="20" /></div>
        <div>
          <div class="kpi-label">Total Value</div>
          <div class="kpi-value">{{ formatSAR(totalValue) }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--warning"><AlertTriangle :size="20" /></div>
        <div>
          <div class="kpi-label">Expiring Soon</div>
          <div class="kpi-value">{{ expiringSoon }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--info"><RefreshCw :size="20" /></div>
        <div>
          <div class="kpi-label">Renewed This Year</div>
          <div class="kpi-value">{{ renewedThisYear }}</div>
        </div>
      </div>
    </div>

    <!-- Filter Toolbar -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search">
          <Search :size="18" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Search contracts…"
          />
        </div>
        <select v-model="filterStatus" class="form-select toolbar-select">
          <option value="">All Statuses</option>
          <option v-for="(label, key) in statusLabels" :key="key" :value="key">{{ label }}</option>
        </select>
        <select v-model="filterType" class="form-select toolbar-select">
          <option value="">All Types</option>
          <option v-for="(label, key) in typeLabels" :key="key" :value="key">{{ label }}</option>
        </select>
      </div>
    </div>

    <!-- Contracts Table -->
    <div v-if="filteredContracts.length" class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Contract #</th>
            <th>Title</th>
            <th>Customer</th>
            <th>Type</th>
            <th>Status</th>
            <th>Start Date</th>
            <th>End Date</th>
            <th class="text-right">Value</th>
            <th class="text-center">Auto-Renew</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filteredContracts" :key="c.id">
            <td class="text-mono font-medium">{{ c.contractNumber }}</td>
            <td>
              <button class="link-btn" @click="openViewModal(c)">{{ c.title }}</button>
            </td>
            <td>{{ c.customerName }}</td>
            <td><span :class="['badge', typeBadge[c.type]]">{{ typeLabels[c.type] }}</span></td>
            <td><span :class="['badge badge-dot', statusBadge[c.status]]">{{ statusLabels[c.status] }}</span></td>
            <td class="whitespace-nowrap">{{ formatDate(c.startDate) }}</td>
            <td class="whitespace-nowrap">{{ formatDate(c.endDate) }}</td>
            <td class="text-right whitespace-nowrap font-medium">{{ formatSAR(c.value) }}</td>
            <td class="text-center">
              <Check v-if="c.autoRenew" :size="16" class="text-success" />
              <XCircle v-else :size="16" class="text-muted" />
            </td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="View" @click="openViewModal(c)">
                  <Eye :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEditModal(c)">
                  <Pencil :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deleteContract(c.id)">
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="empty-state">
      <FileText :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No contracts found</h3>
      <p class="empty-state-text">Try adjusting your filters or add a new contract.</p>
    </div>

    <!-- Add / Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">{{ editingId ? 'Edit Contract' : 'Add Contract' }}</h2>
            <button class="modal-close" @click="showModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Contract # <span class="required">*</span></label>
                <input v-model="form.contractNumber" type="text" class="form-input text-mono" placeholder="CTR-2026-XXX" />
              </div>
              <div class="form-group">
                <label class="form-label">Title <span class="required">*</span></label>
                <input v-model="form.title" type="text" class="form-input" placeholder="Contract title" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Customer <span class="required">*</span></label>
                <select v-model="form.customerId" class="form-select">
                  <option value="" disabled>Select customer</option>
                  <option v-for="c in customerOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Type</label>
                <select v-model="form.type" class="form-select">
                  <option v-for="(label, key) in typeLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Status</label>
                <select v-model="form.status" class="form-select">
                  <option v-for="(label, key) in statusLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Value (SAR)</label>
                <input v-model.number="form.value" type="number" step="1000" min="0" class="form-input" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Start Date</label>
                <input v-model="form.startDate" type="date" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">End Date</label>
                <input v-model="form.endDate" type="date" class="form-input" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-checkbox">
                  <input v-model="form.autoRenew" type="checkbox" />
                  Auto-Renew
                </label>
              </div>
              <div class="form-group">
                <label class="form-label">Renewal Notice (days)</label>
                <input v-model.number="form.renewalNoticeDays" type="number" min="0" class="form-input" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Terms & Conditions</label>
              <textarea v-model="form.terms" class="form-textarea" rows="3" placeholder="Contract terms…" />
            </div>
            <div class="form-group">
              <label class="form-label">Notes</label>
              <textarea v-model="form.notes" class="form-textarea" rows="2" placeholder="Internal notes…" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showModal = false">Cancel</button>
            <button
              class="btn btn-primary"
              :disabled="!form.contractNumber || !form.title || !form.customerId"
              @click="saveContract"
            >
              {{ editingId ? 'Save Changes' : 'Add Contract' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- View Contract Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && viewingContract" class="modal-backdrop" @click.self="showViewModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">{{ viewingContract.title }}</h2>
            <button class="modal-close" @click="showViewModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="view-meta-row">
              <span class="text-mono font-medium">{{ viewingContract.contractNumber }}</span>
              <span :class="['badge badge-dot', statusBadge[viewingContract.status]]">{{ statusLabels[viewingContract.status] }}</span>
              <span :class="['badge', typeBadge[viewingContract.type]]">{{ typeLabels[viewingContract.type] }}</span>
            </div>

            <div class="view-grid">
              <div class="view-field">
                <span class="view-label">Customer</span>
                <span class="view-value">{{ viewingContract.customerName }}</span>
              </div>
              <div class="view-field">
                <span class="view-label">Value</span>
                <span class="view-value font-semibold">{{ formatSAR(viewingContract.value) }}</span>
              </div>
              <div class="view-field">
                <span class="view-label">Auto-Renew</span>
                <span class="view-value">{{ viewingContract.autoRenew ? 'Yes' : 'No' }}</span>
              </div>
            </div>

            <!-- Timeline -->
            <div class="timeline-bar">
              <div class="timeline-segment">
                <CalendarDays :size="14" />
                <div>
                  <div class="timeline-label">Start Date</div>
                  <div class="timeline-date">{{ formatDate(viewingContract.startDate) }}</div>
                </div>
              </div>
              <div class="timeline-line" />
              <div class="timeline-segment">
                <Clock :size="14" />
                <div>
                  <div class="timeline-label">Days Remaining</div>
                  <div :class="['timeline-date', daysRemaining(viewingContract.endDate) <= 30 ? 'text-danger' : '']">
                    {{ daysRemaining(viewingContract.endDate) > 0 ? daysRemaining(viewingContract.endDate) + ' days' : 'Expired' }}
                  </div>
                </div>
              </div>
              <div class="timeline-line" />
              <div class="timeline-segment">
                <CalendarDays :size="14" />
                <div>
                  <div class="timeline-label">End Date</div>
                  <div class="timeline-date">{{ formatDate(viewingContract.endDate) }}</div>
                </div>
              </div>
            </div>

            <div v-if="viewingContract.terms" class="view-section">
              <p class="view-section-title">Terms & Conditions</p>
              <p class="view-text">{{ viewingContract.terms }}</p>
            </div>

            <div v-if="viewingContract.notes" class="view-notes">
              <p class="view-section-title">Notes</p>
              <p class="view-text">{{ viewingContract.notes }}</p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showViewModal = false">Close</button>
            <button class="btn btn-primary" @click="showViewModal = false; openEditModal(viewingContract!)">Edit</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.contracts-page {
  padding: var(--space-6);
}

.kpi-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.kpi-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
}

.kpi-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.kpi-icon--primary { background: var(--color-primary-light, #e0e7ff); color: var(--color-primary); }
.kpi-icon--info { background: #dbeafe; color: #2563eb; }
.kpi-icon--warning { background: #fef3c7; color: #d97706; }
.kpi-icon--success { background: #d1fae5; color: #059669; }

.kpi-label {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.kpi-value {
  font-size: var(--text-xl, 1.25rem);
  font-weight: var(--font-bold, 700);
  color: var(--color-neutral-900);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  flex-wrap: wrap;
}

.toolbar-search {
  flex: 1;
  min-width: 220px;
}

.toolbar-select {
  width: 180px;
  flex-shrink: 0;
}

.link-btn {
  font-weight: var(--font-semibold);
  color: var(--color-primary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  font-size: inherit;
  font-family: inherit;
  text-align: left;
}

.link-btn:hover {
  color: var(--color-primary-hover);
  text-decoration: underline;
}

.text-success { color: #059669; }
.text-muted { color: var(--color-neutral-400); }
.text-danger { color: #dc2626; }

/* View Modal */
.view-meta-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-bottom: var(--space-5);
}

.view-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
  margin-bottom: var(--space-5);
  padding: var(--space-4);
  background-color: var(--color-neutral-50);
  border-radius: var(--radius-lg);
}

.view-field .view-label {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin-bottom: var(--space-1);
}

.view-field .view-value {
  font-size: var(--text-sm);
  color: var(--color-neutral-800);
  font-weight: var(--font-medium);
}

.timeline-bar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-neutral-50);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-5);
}

.timeline-segment {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--color-neutral-600);
}

.timeline-line {
  flex: 1;
  height: 2px;
  background: var(--color-neutral-300);
  border-radius: 1px;
}

.timeline-label {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.timeline-date {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.view-section {
  margin-bottom: var(--space-4);
}

.view-section-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
  margin-bottom: var(--space-2);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-neutral-200);
}

.view-text {
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
  line-height: var(--leading-relaxed);
}

.view-notes {
  padding: var(--space-4);
  background-color: var(--color-warning-light);
  border-radius: var(--radius-lg);
  border-left: 3px solid var(--color-warning);
}

@media (max-width: 768px) {
  .kpi-row { grid-template-columns: repeat(2, 1fr); }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-select { width: 100%; }
  .view-grid { grid-template-columns: 1fr; }
  .timeline-bar { flex-direction: column; align-items: stretch; }
  .timeline-line { width: 100%; height: 2px; }
}
</style>
