<template>
  <div class="contracts-view">
    <BaseCard title="Contracts Management">
      <template #actions>
        <div class="header-stats">
          <div class="stat-item">
            <span class="stat-label">Total:</span>
            <span class="stat-value">{{ contractsStore.contracts.length }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Active:</span>
            <span class="stat-value active">{{ contractsStore.activeContracts.length }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Expiring:</span>
            <span class="stat-value warning">{{ contractsStore.expiringContracts.length }}</span>
          </div>
        </div>
      </template>

      <!-- Filters -->
      <div class="filters-bar">
        <div class="filter-group">
          <label>Search:</label>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search contracts..."
            class="filter-input"
          />
        </div>

        <div class="filter-group">
          <label>Status:</label>
          <select v-model="statusFilter" class="filter-select">
            <option value="all">All Status</option>
            <option value="draft">Draft</option>
            <option value="pending_approval">Pending Approval</option>
            <option value="active">Active</option>
            <option value="expired">Expired</option>
            <option value="terminated">Terminated</option>
            <option value="renewed">Renewed</option>
          </select>
        </div>

        <div class="filter-group">
          <label>Type:</label>
          <select v-model="typeFilter" class="filter-select">
            <option value="all">All Types</option>
            <option value="sales">Sales</option>
            <option value="maintenance">Maintenance</option>
            <option value="service">Service</option>
            <option value="project">Project</option>
            <option value="subscription">Subscription</option>
          </select>
        </div>
      </div>

      <!-- Contracts Table -->
      <BaseTable :columns="columns" :data="filteredContracts" @row-click="viewContractDetails">
        <template #cell-contractNumber="{ row }">
          <span class="contract-number">{{ row.contractNumber }}</span>
        </template>

        <template #cell-status="{ row }">
          <BaseBadge :variant="getStatusVariant(row.status)" size="sm">
            {{ formatStatus(row.status) }}
          </BaseBadge>
        </template>

        <template #cell-type="{ row }">
          <span class="contract-type">{{ formatType(row.type) }}</span>
        </template>

        <template #cell-totalValue="{ row }">
          <span class="amount">{{ formatCurrency(row.totalValue) }} {{ row.currency }}</span>
        </template>

        <template #cell-dates="{ row }">
          <div class="dates">
            <div>{{ formatDate(row.startDate) }}</div>
            <div class="date-separator">→</div>
            <div>{{ formatDate(row.endDate) }}</div>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="View Details" @click.stop="viewContractDetails(row)">
              👁️
            </button>
            <button class="action-btn" title="Download PDF" @click.stop="downloadContractPDF(row)">
              📄
            </button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Contract Details Modal -->
    <BaseModal
      v-model="showDetailsModal"
      :title="`Contract ${selectedContract?.contractNumber}`"
      size="xl"
      hide-confirm
      cancel-text="Close"
    >
      <div v-if="selectedContract" class="contract-details">
        <div class="details-header">
          <div class="detail-item">
            <span class="detail-label">Contract Number:</span>
            <span class="detail-value">{{ selectedContract.contractNumber }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Status:</span>
            <BaseBadge :variant="getStatusVariant(selectedContract.status)">
              {{ formatStatus(selectedContract.status) }}
            </BaseBadge>
          </div>
          <div class="detail-item">
            <span class="detail-label">Type:</span>
            <span class="detail-value">{{ formatType(selectedContract.type) }}</span>
          </div>
        </div>

        <div class="details-grid">
          <div class="detail-item">
            <span class="detail-label">Customer:</span>
            <span class="detail-value">{{ getCustomerName(selectedContract.customerId) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Total Value:</span>
            <span class="detail-value">{{ formatCurrency(selectedContract.totalValue) }} {{ selectedContract.currency }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Start Date:</span>
            <span class="detail-value">{{ formatDate(selectedContract.startDate) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">End Date:</span>
            <span class="detail-value">{{ formatDate(selectedContract.endDate) }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Auto Renewal:</span>
            <span class="detail-value">{{ selectedContract.autoRenewal ? 'Yes' : 'No' }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Payment Terms:</span>
            <span class="detail-value">{{ selectedContract.paymentTerms }}</span>
          </div>
        </div>

        <div v-if="selectedContract.terms" class="detail-section">
          <h4>Terms & Conditions</h4>
          <p>{{ selectedContract.terms }}</p>
        </div>

        <div v-if="selectedContract.deliverables" class="detail-section">
          <h4>Deliverables</h4>
          <p>{{ selectedContract.deliverables }}</p>
        </div>

        <div v-if="selectedContract.sla" class="detail-section">
          <h4>Service Level Agreement</h4>
          <p>{{ selectedContract.sla }}</p>
        </div>

        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">Close</button>
          <button class="btn btn-primary" @click="downloadContractPDF(selectedContract)">
            📄 Download PDF
          </button>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useContractsStore } from '@/stores/contracts'
import { useCustomersStore } from '@/stores/customers'
import { usePdfExport } from '@/composables/usePdfExport'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'
import type { Contract } from '@/types'

const contractsStore = useContractsStore()
const customersStore = useCustomersStore()
const { generateContractPdf } = usePdfExport()
const { success } = useToast()

const searchQuery = ref('')
const statusFilter = ref('all')
const typeFilter = ref('all')
const showDetailsModal = ref(false)
const selectedContract = ref<Contract | null>(null)

const columns: TableColumn[] = [
  { key: 'contractNumber', label: 'Contract #', sortable: true },
  { key: 'name', label: 'Name', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'totalValue', label: 'Value', sortable: true },
  { key: 'dates', label: 'Contract Period' },
  { key: 'actions', label: 'Actions' }
]

const filteredContracts = computed(() => {
  let contracts = [...contractsStore.contracts]

  // Status filter
  if (statusFilter.value !== 'all') {
    contracts = contracts.filter(c => c.status === statusFilter.value)
  }

  // Type filter
  if (typeFilter.value !== 'all') {
    contracts = contracts.filter(c => c.type === typeFilter.value)
  }

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    contracts = contracts.filter(c =>
      c.contractNumber.toLowerCase().includes(query) ||
      c.name.toLowerCase().includes(query) ||
      getCustomerName(c.customerId).toLowerCase().includes(query)
    )
  }

  return contracts
})

function getCustomerName(customerId: string): string {
  const customer = customersStore.getCustomerById(customerId)
  return customer?.name || 'Unknown Customer'
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatStatus(status: string): string {
  const statusMap: Record<string, string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    active: 'Active',
    expired: 'Expired',
    terminated: 'Terminated',
    renewed: 'Renewed'
  }
  return statusMap[status] || status
}

function formatType(type: string): string {
  const typeMap: Record<string, string> = {
    sales: 'Sales',
    maintenance: 'Maintenance',
    service: 'Service',
    project: 'Project',
    subscription: 'Subscription'
  }
  return typeMap[type] || type
}

function getStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'info' | 'default' {
  const variantMap: Record<string, 'success' | 'warning' | 'danger' | 'info' | 'default'> = {
    active: 'success',
    draft: 'default',
    pending_approval: 'warning',
    expired: 'danger',
    terminated: 'danger',
    renewed: 'info'
  }
  return variantMap[status] || 'default'
}

function viewContractDetails(contract: Contract) {
  selectedContract.value = contract
  showDetailsModal.value = true
}

function closeDetailsModal() {
  showDetailsModal.value = false
  selectedContract.value = null
}

function downloadContractPDF(contract: Contract) {
  const customerName = getCustomerName(contract.customerId)
  generateContractPdf(contract, customerName)
  success(`Generating PDF for contract ${contract.contractNumber}`)
}
</script>

<style scoped>
.contracts-view {
  padding: 1.5rem;
}

.header-stats {
  display: flex;
  gap: 2rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stat-label {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.stat-value.active {
  color: var(--color-success);
}

.stat-value.warning {
  color: var(--color-warning);
}

.filters-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 200px;
}

.filter-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.filter-input,
.filter-select {
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.contract-number {
  font-weight: 600;
  color: var(--color-primary);
}

.contract-type {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.amount {
  font-weight: 600;
  color: var(--color-success);
}

.dates {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.date-separator {
  color: var(--color-text-secondary);
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.action-btn:hover {
  background: var(--color-surface);
}

.contract-details {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.details-header {
  display: flex;
  gap: 2rem;
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 8px;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-value {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.detail-section {
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 8px;
}

.detail-section h4 {
  color: var(--color-primary);
  margin-bottom: 0.75rem;
  font-size: 1rem;
}

.detail-section p {
  color: var(--color-text-secondary);
  line-height: 1.6;
  font-size: 0.95rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
}

.btn-secondary {
  background: var(--color-surface);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-border);
}
</style>
