<template>
  <div class="manpower-view">
    <BaseCard>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <h2>Labor Positions & Rate Cards</h2>
            <p class="header-subtitle">Manage labor positions, hourly rates, and cost structures for quoting</p>
          </div>
          <div class="header-actions">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search positions..."
              class="search-input"
            />
            <select v-model="departmentFilter" class="filter-select">
              <option value="all">All Departments</option>
              <option value="Security Operations">Security Operations</option>
              <option value="Technical Services">Technical Services</option>
              <option value="Project Management">Project Management</option>
              <option value="Consulting">Consulting</option>
              <option value="Maintenance">Maintenance</option>
              <option value="Monitoring Services">Monitoring Services</option>
              <option value="Training">Training</option>
            </select>
            <select v-model="statusFilter" class="filter-select">
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
            <button class="btn btn-primary" @click="createPosition">
              <span>➕</span>
              Add Position
            </button>
          </div>
        </div>
      </template>

      <!-- Stats Summary -->
      <div class="stats-bar">
        <div class="stat-card">
          <div class="stat-label">Total Positions</div>
          <div class="stat-value">{{ filteredPositions.length }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Active</div>
          <div class="stat-value success">{{ activePositionsCount }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Avg Cost Rate</div>
          <div class="stat-value">{{ formatCurrency(averageCostRate) }}/hr</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Avg Selling Rate</div>
          <div class="stat-value primary">{{ formatCurrency(averageSellingRate) }}/hr</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Avg Margin</div>
          <div class="stat-value" :class="getMarginClass(averageMargin)">{{ averageMargin.toFixed(1) }}%</div>
        </div>
      </div>

      <!-- Positions Table -->
      <BaseTable
        :columns="tableColumns"
        :data="filteredPositions"
        :loading="false"
      >
        <template #cell-name="{ row }">
          <div class="name-cell">
            <div class="position-name">{{ row.name }}</div>
            <div class="position-description">{{ row.description }}</div>
          </div>
        </template>

        <template #cell-department="{ row }">
          <BaseBadge variant="info" size="sm">
            {{ row.department }}
          </BaseBadge>
        </template>

        <template #cell-rates="{ row }">
          <div class="rates-cell">
            <div class="rate-row">
              <span class="rate-label">Cost:</span>
              <span class="rate-value cost">{{ formatCurrency(row.costPerHour) }}/hr</span>
            </div>
            <div class="rate-row">
              <span class="rate-label">Selling:</span>
              <span class="rate-value selling">{{ formatCurrency(row.sellingRatePerHour) }}/hr</span>
            </div>
          </div>
        </template>

        <template #cell-monthly="{ row }">
          <div class="monthly-cell">
            <div class="monthly-row">
              <span class="monthly-label">Cost:</span>
              <span class="monthly-value">{{ formatCurrency(laborStore.calculateMonthlyCost(row.costPerHour)) }}</span>
            </div>
            <div class="monthly-row">
              <span class="monthly-label">Selling:</span>
              <span class="monthly-value primary">{{ formatCurrency(laborStore.calculateMonthlyCost(row.sellingRatePerHour)) }}</span>
            </div>
            <div class="monthly-note">(160 hrs/month)</div>
          </div>
        </template>

        <template #cell-margin="{ row }">
          <div class="margin-cell">
            <div class="margin-value" :class="getMarginClass(laborStore.calculateMarginPercent(row.costPerHour, row.sellingRatePerHour))">
              {{ laborStore.calculateMarginPercent(row.costPerHour, row.sellingRatePerHour).toFixed(1) }}%
            </div>
            <div class="margin-amount">
              {{ formatCurrency(row.sellingRatePerHour - row.costPerHour) }}/hr
            </div>
          </div>
        </template>

        <template #cell-status="{ row }">
          <BaseBadge :variant="row.status === 'active' ? 'success' : 'default'" size="sm">
            {{ row.status }}
          </BaseBadge>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" @click="viewPosition(row)" title="View Details">👁️</button>
            <button class="action-btn" @click="editPosition(row)" title="Edit">✏️</button>
            <button
              class="action-btn"
              @click="toggleStatus(row)"
              :title="row.status === 'active' ? 'Deactivate' : 'Activate'"
            >
              {{ row.status === 'active' ? '🔒' : '🔓' }}
            </button>
            <button class="action-btn action-btn-danger" @click="deletePosition(row)" title="Delete">🗑️</button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Add/Edit Position Modal -->
    <BaseModal
      v-model="showPositionModal"
      :title="editingPosition ? 'Edit Labor Position' : 'Add Labor Position'"
      size="lg"
    >
      <form @submit.prevent="savePosition" class="position-form">
        <!-- Position Details -->
        <div class="form-section">
          <h4 class="section-title">Position Details</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="required">Position Name</label>
              <input
                v-model="positionForm.name"
                type="text"
                class="form-control"
                placeholder="e.g., Security Guard - Grade A"
                required
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="required">Department</label>
              <select v-model="positionForm.department" class="form-control" required>
                <option value="">Select Department</option>
                <option value="Security Operations">Security Operations</option>
                <option value="Technical Services">Technical Services</option>
                <option value="Project Management">Project Management</option>
                <option value="Consulting">Consulting</option>
                <option value="Maintenance">Maintenance</option>
                <option value="Monitoring Services">Monitoring Services</option>
                <option value="Training">Training</option>
              </select>
            </div>

            <div class="form-group">
              <label class="required">Status</label>
              <select v-model="positionForm.status" class="form-control" required>
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group full-width">
              <label class="required">Description</label>
              <textarea
                v-model="positionForm.description"
                class="form-control"
                rows="3"
                placeholder="Detailed description of the position, requirements, and responsibilities"
                required
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Rate Information -->
        <div class="form-section">
          <h4 class="section-title">Rate Information</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="required">Cost Per Hour (SAR)</label>
              <input
                v-model.number="positionForm.costPerHour"
                type="number"
                class="form-control"
                step="0.01"
                min="0"
                placeholder="0.00"
                required
              />
              <small class="form-hint">Internal cost including overheads</small>
            </div>

            <div class="form-group">
              <label class="required">Selling Rate Per Hour (SAR)</label>
              <input
                v-model.number="positionForm.sellingRatePerHour"
                type="number"
                class="form-control"
                step="0.01"
                min="0"
                placeholder="0.00"
                required
              />
              <small class="form-hint">Client-facing rate</small>
            </div>
          </div>

          <!-- Rate Preview -->
          <div class="rate-preview">
            <div class="preview-row">
              <span class="preview-label">Monthly Cost (160 hrs):</span>
              <span class="preview-value">{{ formatCurrency(positionForm.costPerHour * 160) }}</span>
            </div>
            <div class="preview-row">
              <span class="preview-label">Monthly Selling (160 hrs):</span>
              <span class="preview-value primary">{{ formatCurrency(positionForm.sellingRatePerHour * 160) }}</span>
            </div>
            <div class="preview-row preview-total">
              <span class="preview-label">Margin:</span>
              <span class="preview-value" :class="getMarginClass(laborStore.calculateMarginPercent(positionForm.costPerHour, positionForm.sellingRatePerHour))">
                {{ laborStore.calculateMarginPercent(positionForm.costPerHour, positionForm.sellingRatePerHour).toFixed(1) }}%
                ({{ formatCurrency(positionForm.sellingRatePerHour - positionForm.costPerHour) }}/hr)
              </span>
            </div>
          </div>
        </div>

        <!-- Notes -->
        <div class="form-section">
          <h4 class="section-title">Additional Information</h4>
          <div class="form-row">
            <div class="form-group full-width">
              <label>Notes</label>
              <textarea
                v-model="positionForm.notes"
                class="form-control"
                rows="3"
                placeholder="Internal notes, requirements, certifications, etc."
              ></textarea>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" @click="closePositionModal">Cancel</button>
          <button type="button" class="btn btn-primary" @click="savePosition">
            {{ editingPosition ? 'Update Position' : 'Add Position' }}
          </button>
        </div>
      </template>
    </BaseModal>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model="showDeleteModal"
      title="Delete Labor Position"
      size="sm"
    >
      <div class="delete-modal-content">
        <div class="delete-icon">🗑️</div>
        <p class="delete-message">Are you sure you want to delete <strong>{{ positionToDelete?.name }}</strong>?</p>
        <p class="delete-warning">This action cannot be undone.</p>
      </div>

      <template #footer>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">Cancel</button>
          <button type="button" class="btn btn-danger" @click="confirmDelete">Delete</button>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useLaborPositionsStore } from '@/stores/laborPositions'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'
import type { LaborPosition, Status } from '@/types'

const laborStore = useLaborPositionsStore()
const { success, info } = useToast()

// Filters
const searchQuery = ref('')
const departmentFilter = ref('all')
const statusFilter = ref('all')

// Modal states
const showPositionModal = ref(false)
const showDeleteModal = ref(false)
const editingPosition = ref<LaborPosition | null>(null)
const positionToDelete = ref<LaborPosition | null>(null)

// Form data
const positionForm = ref({
  name: '',
  description: '',
  department: '',
  costPerHour: 0,
  sellingRatePerHour: 0,
  status: 'active' as Status,
  notes: ''
})

// Table columns
const tableColumns: TableColumn[] = [
  { key: 'name', label: 'Position', sortable: true },
  { key: 'department', label: 'Department', sortable: true },
  { key: 'rates', label: 'Hourly Rates', sortable: false },
  { key: 'monthly', label: 'Monthly (160 hrs)', sortable: false },
  { key: 'margin', label: 'Margin', sortable: false },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'actions', label: 'Actions', align: 'right' }
]

// Computed
const filteredPositions = computed(() => {
  let filtered = laborStore.positions

  // Filter by department
  if (departmentFilter.value !== 'all') {
    filtered = filtered.filter(p => p.department === departmentFilter.value)
  }

  // Filter by status
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(p => p.status === statusFilter.value)
  }

  // Filter by search
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(p =>
      p.name.toLowerCase().includes(query) ||
      p.description.toLowerCase().includes(query) ||
      p.department.toLowerCase().includes(query)
    )
  }

  return filtered
})

const activePositionsCount = computed(() =>
  laborStore.positions.filter(p => p.status === 'active').length
)

const averageCostRate = computed(() => {
  const active = laborStore.getActivePositions()
  if (active.length === 0) return 0
  const total = active.reduce((sum, p) => sum + p.costPerHour, 0)
  return total / active.length
})

const averageSellingRate = computed(() => {
  const active = laborStore.getActivePositions()
  if (active.length === 0) return 0
  const total = active.reduce((sum, p) => sum + p.sellingRatePerHour, 0)
  return total / active.length
})

const averageMargin = computed(() => {
  const active = laborStore.getActivePositions()
  if (active.length === 0) return 0
  const margins = active.map(p => laborStore.calculateMarginPercent(p.costPerHour, p.sellingRatePerHour))
  const total = margins.reduce((sum, m) => sum + m, 0)
  return total / active.length
})

// Helper functions
function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

function getMarginClass(margin: number): string {
  if (margin >= 30) return 'high'
  if (margin >= 20) return 'medium'
  return 'low'
}

// CRUD Functions
function createPosition() {
  editingPosition.value = null
  resetPositionForm()
  showPositionModal.value = true
}

function viewPosition(position: LaborPosition) {
  info(`Viewing ${position.name} - Cost: ${formatCurrency(position.costPerHour)}/hr, Selling: ${formatCurrency(position.sellingRatePerHour)}/hr`)
}

function editPosition(position: LaborPosition) {
  editingPosition.value = position
  positionForm.value = {
    name: position.name,
    description: position.description,
    department: position.department,
    costPerHour: position.costPerHour,
    sellingRatePerHour: position.sellingRatePerHour,
    status: position.status,
    notes: position.notes || ''
  }
  showPositionModal.value = true
}

function savePosition() {
  // Validation
  if (!positionForm.value.name || !positionForm.value.description || !positionForm.value.department) {
    info('Please fill in all required fields')
    return
  }

  if (positionForm.value.costPerHour <= 0 || positionForm.value.sellingRatePerHour <= 0) {
    info('Cost and selling rates must be greater than zero')
    return
  }

  if (positionForm.value.sellingRatePerHour < positionForm.value.costPerHour) {
    info('Warning: Selling rate is lower than cost rate (negative margin)')
  }

  if (editingPosition.value) {
    // Update existing
    laborStore.updatePosition(editingPosition.value.id, positionForm.value)
    success('Position updated successfully')
  } else {
    // Create new
    laborStore.addPosition(positionForm.value)
    success('Position created successfully')
  }

  closePositionModal()
}

function toggleStatus(position: LaborPosition) {
  laborStore.togglePositionStatus(position.id)
  success(`Position ${position.status === 'active' ? 'deactivated' : 'activated'}`)
}

function deletePosition(position: LaborPosition) {
  positionToDelete.value = position
  showDeleteModal.value = true
}

function confirmDelete() {
  if (positionToDelete.value) {
    laborStore.deletePosition(positionToDelete.value.id)
    success(`Position ${positionToDelete.value.name} deleted successfully`)
    positionToDelete.value = null
  }
  showDeleteModal.value = false
}

function resetPositionForm() {
  positionForm.value = {
    name: '',
    description: '',
    department: '',
    costPerHour: 0,
    sellingRatePerHour: 0,
    status: 'active',
    notes: ''
  }
}

function closePositionModal() {
  showPositionModal.value = false
  editingPosition.value = null
  resetPositionForm()
}
</script>

<style scoped>
.manpower-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 2rem;
  flex-wrap: wrap;
}

.header-left h2 {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.header-subtitle {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.search-input {
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  min-width: 250px;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.filter-select {
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  cursor: pointer;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.btn-secondary {
  background: var(--color-surface);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.btn-danger {
  background: var(--color-danger);
  color: white;
}

/* Stats Bar */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: var(--color-background);
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.stat-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 0.5rem;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.stat-value.success {
  color: var(--color-success);
}

.stat-value.primary {
  color: var(--color-primary);
}

.stat-value.high {
  color: var(--color-success);
}

.stat-value.medium {
  color: var(--color-warning);
}

.stat-value.low {
  color: var(--color-danger);
}

/* Table Cells */
.name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 250px;
}

.position-name {
  font-weight: 600;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.position-description {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.rates-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.rate-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.rate-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  min-width: 50px;
}

.rate-value {
  font-weight: 600;
  font-size: 0.95rem;
}

.rate-value.cost {
  color: var(--color-text-primary);
}

.rate-value.selling {
  color: var(--color-primary);
}

.monthly-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.monthly-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.monthly-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  min-width: 50px;
}

.monthly-value {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.monthly-value.primary {
  color: var(--color-primary);
}

.monthly-note {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  font-style: italic;
  margin-top: 0.25rem;
}

.margin-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  align-items: center;
}

.margin-value {
  font-weight: 700;
  font-size: 1.1rem;
}

.margin-value.high {
  color: var(--color-success);
}

.margin-value.medium {
  color: var(--color-warning);
}

.margin-value.low {
  color: var(--color-danger);
}

.margin-amount {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

.action-btn-danger:hover {
  opacity: 1;
  filter: brightness(1.2);
}

/* Form Styles */
.position-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.form-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.section-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-group label.required::after {
  content: ' *';
  color: var(--color-danger);
}

.form-control {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-control:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

textarea.form-control {
  resize: vertical;
  font-family: inherit;
}

.form-hint {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

/* Rate Preview */
.rate-preview {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.preview-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.95rem;
}

.preview-label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.preview-value {
  color: var(--color-text-primary);
  font-weight: 600;
}

.preview-value.primary {
  color: var(--color-primary);
}

.preview-value.high {
  color: var(--color-success);
}

.preview-value.medium {
  color: var(--color-warning);
}

.preview-value.low {
  color: var(--color-danger);
}

.preview-total {
  padding-top: 0.5rem;
  margin-top: 0.5rem;
  border-top: 2px solid var(--color-border);
  font-size: 1.05rem;
}

.preview-total .preview-label {
  color: var(--color-text-primary);
  font-weight: 600;
}

/* Modal Styles */
.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.delete-modal-content {
  text-align: center;
  padding: 1rem;
}

.delete-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.delete-message {
  font-size: 1.05rem;
  color: var(--color-text-primary);
  margin-bottom: 0.75rem;
}

.delete-warning {
  font-size: 0.9rem;
  color: var(--color-danger);
}
</style>
