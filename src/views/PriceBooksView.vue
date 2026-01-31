<template>
  <div class="pricebooks-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search contracts and price books by name or description..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="createNewPriceBook">
          <span>➕</span> New Contract
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Status:</label>
        <select v-model="statusFilter" class="filter-select">
          <option value="all">All Status</option>
          <option value="draft">Draft</option>
          <option value="pending_approval">Pending Approval</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="expired">Expired</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="standard">Standard</option>
          <option value="volume">Volume Discount</option>
          <option value="contract">Contract Pricing</option>
          <option value="promotional">Promotional</option>
          <option value="customer_specific">Customer Specific</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total Contracts:</span>
          <span class="stat-value">{{ filteredPriceBooks.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Active:</span>
          <span class="stat-value text-success">{{ activePriceBooksCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Pending Approval:</span>
          <span class="stat-value text-warning">{{ pendingApprovalCount }}</span>
        </div>
      </div>
    </div>

    <BaseCard no-padding>
      <BaseTable
        :columns="priceBookColumns"
        :data="filteredPriceBooks"
        :clickable="true"
        @row-click="viewPriceBookDetails"
      >
        <template #cell-name="{ row }">
          <div class="name-cell">
            <div class="name">{{ row.name }}</div>
            <div class="description">{{ row.description }}</div>
          </div>
        </template>

        <template #cell-type="{ value }">
          <BaseBadge :variant="getTypeVariant(value)" size="sm">
            {{ formatType(value) }}
          </BaseBadge>
        </template>

        <template #cell-validity="{ row }">
          <div class="validity-cell">
            <div class="date-row">
              <span class="date-label">From:</span>
              <span>{{ formatDate(row.validFrom) }}</span>
            </div>
            <div v-if="row.validUntil" class="date-row">
              <span class="date-label">Until:</span>
              <span :class="{ 'text-danger': isExpiringSoon(row.validUntil) }">
                {{ formatDate(row.validUntil) }}
              </span>
            </div>
            <div v-else class="date-row">
              <span class="date-label">Until:</span>
              <span class="text-success">No Expiry</span>
            </div>
          </div>
        </template>

        <template #cell-customers="{ row }">
          <div class="customers-cell">
            <span v-if="row.customerIds.length === 0" class="text-secondary">All Customers</span>
            <span v-else class="customer-count">{{ row.customerIds.length }} Customer(s)</span>
          </div>
        </template>

        <template #cell-items="{ row }">
          <div class="items-cell">
            <span class="item-count">{{ row.priceBookEntries.length }} Item(s)</span>
          </div>
        </template>

        <template #cell-status="{ value }">
          <BaseBadge :variant="getStatusVariant(value)" size="sm">
            {{ formatStatus(value) }}
          </BaseBadge>
        </template>

        <template #cell-approval="{ row }">
          <div v-if="row.requiresApproval" class="approval-cell">
            <span class="approval-icon">⚠️</span>
            <span class="approval-text">{{ row.approvalReason }}</span>
          </div>
          <div v-else class="approval-cell">
            <span class="approval-icon">✓</span>
            <span class="approval-text">Standard</span>
          </div>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="View Details" @click.stop="viewPriceBookDetails(row)">👁️</button>
            <button class="action-btn" title="Edit" @click.stop="editPriceBook(row)" :disabled="!canEdit(row)">✏️</button>
            <button class="action-btn" title="Duplicate" @click.stop="duplicatePriceBook(row)">📋</button>
            <button
              v-if="row.status === 'draft' || row.status === 'pending_approval'"
              class="action-btn"
              title="Submit for Approval"
              @click.stop="submitForApproval(row)"
            >
              📤
            </button>
            <button
              v-if="row.status === 'pending_approval'"
              class="action-btn"
              title="Approve"
              @click.stop="approvePriceBook(row)"
            >
              ✅
            </button>
            <button
              v-if="row.status === 'pending_approval'"
              class="action-btn action-btn-danger"
              title="Reject"
              @click.stop="rejectPriceBook(row)"
            >
              ❌
            </button>
            <button
              v-if="row.status === 'active'"
              class="action-btn"
              title="Deactivate"
              @click.stop="deactivatePriceBook(row)"
            >
              ⏸️
            </button>
            <button
              v-if="row.status === 'inactive'"
              class="action-btn"
              title="Reactivate"
              @click.stop="reactivatePriceBook(row)"
            >
              ▶️
            </button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Price Book Details Modal -->
    <BaseModal
      v-if="selectedPriceBook"
      v-model="showDetailsModal"
      :title="selectedPriceBook.name"
      size="xl"
    >
      <div class="pricebook-details-modern">
        <!-- Workflow Progress Indicator -->
        <div v-if="selectedPriceBook.status !== 'active'" class="modern-progress-card">
          <div class="progress-steps">
            <div class="progress-step" :class="{ active: selectedPriceBook.status === 'draft' || selectedPriceBook.status === 'pending_approval' || selectedPriceBook.status === 'active', completed: selectedPriceBook.status !== 'draft' }">
              <div class="step-icon">📝</div>
              <div class="step-label">Draft</div>
            </div>
            <div class="progress-line" :class="{ active: selectedPriceBook.status !== 'draft' }"></div>
            <div class="progress-step" :class="{ active: selectedPriceBook.status === 'pending_approval' || selectedPriceBook.status === 'active', completed: selectedPriceBook.status === 'active' }">
              <div class="step-icon">⏳</div>
              <div class="step-label">Pending</div>
            </div>
            <div class="progress-line" :class="{ active: selectedPriceBook.status === 'active' }"></div>
            <div class="progress-step" :class="{ active: selectedPriceBook.status === 'active', completed: selectedPriceBook.status === 'active' }">
              <div class="step-icon">✅</div>
              <div class="step-label">Active</div>
            </div>
          </div>
        </div>

        <!-- Approval Information Card -->
        <div v-if="selectedPriceBook.requiresApproval || selectedPriceBook.approvedBy || selectedPriceBook.rejectionReason" class="modern-info-card approval-card">
          <div class="card-header-mini">
            <h4>⚠️ Approval Information</h4>
          </div>
          <div class="card-content">
            <div v-if="selectedPriceBook.requiresApproval" class="alert-box warning">
              <div class="alert-content">
                <div class="alert-title">Approval Required</div>
                <div class="alert-message">{{ selectedPriceBook.approvalReason }}</div>
              </div>
            </div>
            <div v-if="selectedPriceBook.approvedBy" class="info-row">
              <span class="info-label">Approved By:</span>
              <span class="info-value">{{ selectedPriceBook.approvedBy }}</span>
            </div>
            <div v-if="selectedPriceBook.approvedAt" class="info-row">
              <span class="info-label">Approved At:</span>
              <span class="info-value">{{ formatDate(selectedPriceBook.approvedAt) }}</span>
            </div>
            <div v-if="selectedPriceBook.rejectionReason" class="alert-box danger">
              <div class="alert-content">
                <div class="alert-title">Rejected</div>
                <div class="alert-message">{{ selectedPriceBook.rejectionReason }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Contract Information Card -->
        <div class="modern-info-card">
          <div class="card-header-mini">
            <h4>📋 Contract Information</h4>
          </div>
          <div class="card-content">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">Name</span>
                <span class="info-value highlight">{{ selectedPriceBook.name }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Type</span>
                <BaseBadge :variant="getTypeVariant(selectedPriceBook.type)" size="sm">
                  {{ formatType(selectedPriceBook.type) }}
                </BaseBadge>
              </div>
              <div class="info-item">
                <span class="info-label">Status</span>
                <BaseBadge :variant="getStatusVariant(selectedPriceBook.status)" size="sm">
                  {{ formatStatus(selectedPriceBook.status) }}
                </BaseBadge>
              </div>
              <div class="info-item">
                <span class="info-label">Currency</span>
                <span class="info-value">{{ selectedPriceBook.currency }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Valid From</span>
                <span class="info-value">{{ formatDate(selectedPriceBook.validFrom) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Valid Until</span>
                <span class="info-value">
                  {{ selectedPriceBook.validUntil ? formatDate(selectedPriceBook.validUntil) : '∞ No Expiry' }}
                </span>
              </div>
            </div>
            <div v-if="selectedPriceBook.description" class="description-box">
              <span class="info-label">Description</span>
              <p class="description-text">{{ selectedPriceBook.description }}</p>
            </div>
          </div>
        </div>

        <!-- Assigned Customers Card -->
        <div class="modern-info-card">
          <div class="card-header-mini">
            <h4>👥 Assigned Customers</h4>
            <span class="count-badge">{{ selectedPriceBook.customerIds.length || 'All' }}</span>
          </div>
          <div class="card-content">
            <div v-if="selectedPriceBook.customerIds.length === 0" class="empty-state-mini">
              <span class="empty-icon">🌍</span>
              <span class="empty-text">Available to all customers</span>
            </div>
            <div v-else class="chips-container">
              <div
                v-for="customerId in selectedPriceBook.customerIds"
                :key="customerId"
                class="modern-chip"
              >
                {{ getCustomerName(customerId) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Price Book Entries Card -->
        <div class="modern-info-card items-card">
          <div class="card-header-mini">
            <h4>💰 Price Book Items</h4>
            <span class="count-badge">{{ selectedPriceBook.priceBookEntries.length }}</span>
          </div>
          <div class="card-content">
            <div v-if="selectedPriceBook.priceBookEntries.length === 0" class="empty-state-mini">
              <span class="empty-icon">📦</span>
              <span class="empty-text">No items added yet</span>
            </div>
            <div v-else class="items-list-modern">
              <div v-for="(entry, index) in selectedPriceBook.priceBookEntries" :key="entry.id" class="item-card-mini">
                <div class="item-number-mini">{{ index + 1 }}</div>
                <div class="item-details">
                  <div class="item-header-row">
                    <div class="item-name">
                      <span class="item-code">{{ getProductCode(entry.productId) }}</span>
                      <span class="item-title">{{ getProductName(entry.productId) }}</span>
                    </div>
                    <div class="item-final-price">
                      {{ formatCurrency(calculateFinalPrice(entry)) }}
                    </div>
                  </div>
                  <div class="item-pricing-row">
                    <div class="pricing-detail">
                      <span class="pricing-label">List Price:</span>
                      <span class="pricing-value">{{ formatCurrency(entry.listPrice) }}</span>
                    </div>
                    <div v-if="entry.specialPrice" class="pricing-detail highlight">
                      <span class="pricing-label">Special Price:</span>
                      <span class="pricing-value special">{{ formatCurrency(entry.specialPrice) }}</span>
                    </div>
                    <div v-if="entry.discountPercent" class="pricing-detail">
                      <span class="pricing-label">Discount:</span>
                      <span class="discount-pill">{{ entry.discountPercent }}%</span>
                    </div>
                    <div class="pricing-detail">
                      <span class="pricing-label">Min Qty:</span>
                      <span class="pricing-value">{{ entry.minimumQuantity || 1 }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Notes Card -->
        <div v-if="selectedPriceBook.notes" class="modern-info-card">
          <div class="card-header-mini">
            <h4>📝 Notes</h4>
          </div>
          <div class="card-content">
            <div class="notes-box">
              {{ selectedPriceBook.notes }}
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">Close</button>

          <!-- Workflow Actions -->
          <button
            v-if="selectedPriceBook.status === 'draft' || selectedPriceBook.status === 'pending_approval'"
            class="btn btn-primary"
            @click="submitForApproval(selectedPriceBook); closeDetailsModal()"
          >
            📤 Submit for Approval
          </button>

          <button
            v-if="selectedPriceBook.status === 'pending_approval'"
            class="btn btn-success"
            @click="approvePriceBook(selectedPriceBook); closeDetailsModal()"
          >
            ✅ Approve
          </button>

          <button
            v-if="selectedPriceBook.status === 'pending_approval'"
            class="btn btn-danger"
            @click="rejectPriceBook(selectedPriceBook); closeDetailsModal()"
          >
            ❌ Reject
          </button>

          <button
            v-if="canEdit(selectedPriceBook)"
            class="btn btn-primary"
            @click="editPriceBook(selectedPriceBook)"
          >
            ✏️ Edit Contract
          </button>
        </div>
      </template>
    </BaseModal>

    <!-- Create/Edit Price Book/Contract Modal -->
    <BaseModal
      v-model="showCreateModal"
      :title="editingPriceBook ? 'Edit Price Book/Contract' : 'Create New Price Book/Contract'"
      size="xl"
      hide-confirm
      cancel-text="Cancel"
    >
      <PriceBookBuilderModern />
    </BaseModal>

    <!-- Add Document Modal -->
    <BaseModal
      v-model="showAddDocumentModal"
      title="Attach Document to Price Book"
      size="md"
      hide-confirm
      cancel-text="Cancel"
    >
      <form @submit.prevent="attachDocument" class="document-form">
        <div class="form-group">
          <label class="form-label">Select Document *</label>
          <select v-model="documentForm.documentId" class="form-input" required>
            <option value="">-- Select Document --</option>
            <option v-for="doc in availableDocuments" :key="doc.id" :value="doc.id">
              {{ doc.name }} ({{ doc.type }})
            </option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">Custom Name (Optional)</label>
          <input
            v-model="documentForm.customName"
            type="text"
            class="form-input"
            placeholder="Leave blank to use document's original name"
          />
        </div>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="documentForm.includeInPdf" />
            Include this document in PDF export
          </label>
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-secondary" @click="closeDocumentModal">Cancel</button>
          <button type="submit" class="btn btn-primary">Attach Document</button>
        </div>
      </form>
    </BaseModal>

    <!-- Add Entry Modal -->
    <BaseModal
      v-model="showAddEntryModal"
      title="Add Product/Service to Price Book"
      size="lg"
      hide-confirm
      cancel-text="Cancel"
    >
      <form @submit.prevent="saveEntry" class="entry-form">
        <div class="form-group">
          <label class="form-label">Category *</label>
          <select v-model="entryForm.category" class="form-input" required @change="onCategoryChange">
            <option value="">-- Select Category --</option>
            <option value="product">📦 Products</option>
            <option value="recurring">🔄 Recurring Services</option>
            <option value="labor">👷 Labor Positions</option>
          </select>
        </div>

        <div v-if="entryForm.category" class="form-group autocomplete-container">
          <label class="form-label">Search {{ getCategoryLabel() }} *</label>
          <div class="autocomplete-wrapper">
            <input
              v-model="productSearchQuery"
              type="text"
              class="form-input"
              :placeholder="`Search by name or SKU...`"
              required
              @input="handleProductSearchInput"
              @focus="showProductSuggestions = true"
              @blur="handleProductSearchBlur"
              autocomplete="off"
            />
            <div v-if="showProductSuggestions && filteredProductSuggestions.length > 0" class="autocomplete-dropdown">
              <div class="autocomplete-header">
                <span class="suggestion-count">{{ filteredProductSuggestions.length }} results found</span>
              </div>
              <div
                v-for="item in filteredProductSuggestions"
                :key="item.id"
                class="autocomplete-item"
                @mousedown.prevent="selectProductItem(item)"
              >
                <div class="suggestion-main">
                  <span class="suggestion-icon">{{ item.icon }}</span>
                  <span class="suggestion-sku" v-if="item.sku">{{ item.sku }}</span>
                  <span class="suggestion-name">{{ item.name }}</span>
                </div>
                <div class="suggestion-meta">
                  <span class="suggestion-price">{{ item.priceLabel }}</span>
                  <span v-if="item.department" class="suggestion-department">{{ item.department }}</span>
                </div>
              </div>
            </div>
          </div>
          <small class="form-hint">
            {{ getFilteredItems().length }} items available
          </small>
        </div>

        <div v-if="selectedProduct" class="product-details">
          <div class="detail-row">
            <span class="label">Type:</span>
            <span class="value">
              <span v-if="selectedProduct.unitOfMeasure === 'Month'">🔄 Recurring Service</span>
              <span v-else-if="selectedProduct.unitOfMeasure === 'Hour'">👷 Labor Position</span>
              <span v-else>📦 Product</span>
              ({{ selectedProduct.unitOfMeasure }})
            </span>
          </div>
          <div class="detail-row">
            <span class="label">Landed Cost:</span>
            <span class="value cost-value">
              {{ formatCurrency(selectedProduct.landedCostSAR) }}{{ selectedProduct.unitOfMeasure === 'Month' ? '/month' : selectedProduct.unitOfMeasure === 'Hour' ? '/hr' : '' }}
            </span>
          </div>
          <div class="detail-row">
            <span class="label">Standard Price:</span>
            <span class="value price-value">
              {{ formatCurrency(selectedProduct.sellingPrice) }}{{ selectedProduct.unitOfMeasure === 'Month' ? '/month' : selectedProduct.unitOfMeasure === 'Hour' ? '/hr' : '' }}
            </span>
          </div>
          <div class="detail-row">
            <span class="label">Standard Margin:</span>
            <span class="value margin-value">
              {{ calculateStandardMargin() }}%
            </span>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">
            Custom Price *
            <span v-if="selectedProduct && selectedProduct.unitOfMeasure === 'Month'">/month</span>
            <span v-else-if="selectedProduct && selectedProduct.unitOfMeasure === 'Hour'">/hr</span>
          </label>
          <input
            v-model.number="entryForm.customPrice"
            type="number"
            class="form-input"
            min="0"
            step="0.01"
            required
          />
          <small v-if="selectedProduct" class="form-hint">
            Discount: {{ calculateDiscount() }}% | Margin: {{ calculateCustomMargin() }}%
            <span v-if="selectedProduct.unitOfMeasure === 'Month'" class="service-badge"> (Monthly Service)</span>
            <span v-else-if="selectedProduct.unitOfMeasure === 'Hour'" class="service-badge"> (Hourly Rate)</span>
          </small>
        </div>

        <div class="form-group">
          <label class="form-label">
            <span v-if="selectedProduct && selectedProduct.unitOfMeasure === 'Month'">Minimum Duration (Months)</span>
            <span v-else-if="selectedProduct && selectedProduct.unitOfMeasure === 'Hour'">Minimum Hours</span>
            <span v-else>Minimum Quantity</span>
          </label>
          <input
            v-model.number="entryForm.minQuantity"
            type="number"
            class="form-input"
            min="1"
            :placeholder="selectedProduct && selectedProduct.unitOfMeasure === 'Month' ? 'e.g., 3 months minimum' : selectedProduct && selectedProduct.unitOfMeasure === 'Hour' ? 'e.g., 160 hours minimum' : 'Optional'"
          />
        </div>

        <div class="form-group full-width">
          <label class="form-label">Notes</label>
          <textarea
            v-model="entryForm.notes"
            class="form-input"
            rows="2"
            placeholder="Special notes about this pricing..."
          ></textarea>
        </div>

        <div class="form-actions">
          <button type="button" class="btn btn-secondary" @click="closeEntryModal">Cancel</button>
          <button type="submit" class="btn btn-primary">Add to Price Book</button>
        </div>
      </form>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useCustomersStore } from '@/stores/customers'
import { useProductsStore } from '@/stores/products'
import { useDocumentsStore } from '@/stores/documents'
import { useLaborPositionsStore } from '@/stores/laborPositions'
import { useRecurringServicesStore } from '@/stores/recurringServices'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import PriceBookBuilderModern from '@/components/PriceBookBuilderModern.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'

interface PriceBookEntry {
  id: string
  productId: string
  listPrice: number
  specialPrice?: number
  discountPercent?: number
  minimumQuantity?: number
}

interface PriceBook {
  id: string
  name: string
  description: string
  type: 'standard' | 'volume' | 'contract' | 'promotional' | 'customer_specific'
  status: 'draft' | 'pending_approval' | 'active' | 'inactive' | 'expired'
  currency: string
  validFrom: string
  validUntil?: string
  customerIds: string[]
  priceBookEntries: PriceBookEntry[]
  notes?: string
  createdAt: string
  updatedAt: string

  // Contract workflow fields
  requiresApproval?: boolean
  approvalReason?: string
  approvedBy?: string
  approvedAt?: string
  rejectionReason?: string

  // Contract-specific fields
  contractDuration?: string
  autoRenewal?: boolean
  renewalNoticeDays?: number
  paymentTerms?: string
  deliveryTerms?: string
  terms?: string
  sla?: string

  // Progress tracking
  completionPercent?: number
  workflowStage?: 'creation' | 'review' | 'approval' | 'active' | 'completed'
}

const customersStore = useCustomersStore()
const productsStore = useProductsStore()
const documentsStore = useDocumentsStore()
const laborStore = useLaborPositionsStore()
const recurringStore = useRecurringServicesStore()
const { success, info } = useToast()

const searchQuery = ref('')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')
const typeFilter = ref<'all' | PriceBook['type']>('all')
const showDetailsModal = ref(false)
const selectedPriceBook = ref<PriceBook | null>(null)

// Create/Edit Modal State
const showCreateModal = ref(false)
const editingPriceBook = ref<PriceBook | null>(null)
const selectedCustomers = ref<string[]>([])

// Price Book Form
const priceBookForm = ref({
  name: '',
  description: '',
  type: 'standard' as PriceBook['type'],
  status: 'draft' as 'active' | 'inactive' | 'draft' | 'pending_approval',
  currency: 'SAR',
  validFrom: new Date().toISOString().split('T')[0],
  validUntil: '',
  contractDuration: '',
  autoRenewal: false,
  renewalNoticeDays: 30,
  paymentTerms: '',
  deliveryTerms: '',
  terms: '',
  sla: ''
})

// Line Items State
const priceBookEntries = ref<PriceBookEntry[]>([])
const showAddEntryModal = ref(false)
const editingEntryIndex = ref<number | null>(null)

const entryForm = ref({
  category: '' as 'product' | 'recurring' | 'labor' | '',
  productId: '',
  customPrice: 0,
  minQuantity: 1,
  notes: ''
})

// Product autocomplete state
const productSearchQuery = ref('')
const showProductSuggestions = ref(false)

// Documents State
const attachedDocuments = ref<Array<{
  documentId: string
  customName?: string
  includeInPdf: boolean
}>>([])
const showAddDocumentModal = ref(false)

const documentForm = ref({
  documentId: '',
  customName: '',
  includeInPdf: true
})

// Helper function - defined before computed properties that use it
function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(amount)
}

// Get customers list
const customers = computed(() => customersStore.customers || [])

// Available products for selection
const availableProducts = computed(() => productsStore.getActiveProducts())

// Available items - combines products, recurring services, and labor positions
const availableItems = computed(() => {
  const items: Array<{
    id: string
    type: 'product' | 'recurring' | 'labor'
    name: string
    sku?: string
    price: number
    priceLabel: string
    unitOfMeasure?: string
    department?: string
    icon: string
  }> = []

  // Add products from the products store
  productsStore.getActiveProducts().forEach(product => {
    items.push({
      id: product.id,
      type: 'product',
      name: product.name,
      sku: product.sku,
      price: product.sellingPrice,
      priceLabel: formatCurrency(product.sellingPrice),
      unitOfMeasure: product.unitOfMeasure,
      icon: '📦'
    })
  })

  // Add recurring services
  recurringStore.getActiveRecurringServices().forEach(service => {
    items.push({
      id: service.id,
      type: 'recurring',
      name: service.name,
      sku: service.id,
      price: service.monthlyPrice,
      priceLabel: `${formatCurrency(service.monthlyPrice)}/month`,
      unitOfMeasure: 'Month',
      icon: '🔄'
    })
  })

  // Add labor positions
  laborStore.getActivePositions().forEach(position => {
    items.push({
      id: position.id,
      type: 'labor',
      name: position.name,
      price: position.sellingRatePerHour,
      priceLabel: `${formatCurrency(position.sellingRatePerHour)}/hr`,
      department: position.department,
      icon: '👷'
    })
  })

  return items
})

// Grouped items for the dropdown
const groupedItems = computed(() => {
  const groups: Record<string, typeof availableItems.value> = {
    'Products': [],
    'Recurring Services': [],
    'Labor Positions': []
  }

  availableItems.value.forEach(item => {
    if (item.type === 'product') {
      groups['Products'].push(item)
    } else if (item.type === 'service') {
      groups['Recurring Services'].push(item)
    } else if (item.type === 'labor') {
      groups['Labor Positions'].push(item)
    }
  })

  // Remove empty groups
  Object.keys(groups).forEach(key => {
    if (groups[key].length === 0) {
      delete groups[key]
    }
  })

  return groups
})

// Selected product from entry form
const selectedProduct = computed(() => {
  if (!entryForm.value.productId) return null

  // Try to get from products store first
  const product = productsStore.getProductById(entryForm.value.productId)
  if (product) return product

  // Try recurring services store
  const recurring = recurringStore.getRecurringServiceById(entryForm.value.productId)
  if (recurring) {
    // Convert recurring service to product-like object
    return {
      id: recurring.id,
      sku: recurring.id,
      name: recurring.name,
      sellingPrice: recurring.monthlyPrice,
      unitOfMeasure: 'Month',
      landedCostSAR: recurring.monthlyCost
    }
  }

  // Try labor store
  const labor = laborStore.getPositionById(entryForm.value.productId)
  if (labor) {
    // Convert labor position to product-like object
    return {
      id: labor.id,
      sku: '',
      name: labor.name,
      sellingPrice: labor.sellingRatePerHour,
      unitOfMeasure: 'Hour',
      landedCostSAR: labor.costPerHour
    }
  }

  return null
})

// Available documents for selection
const availableDocuments = computed(() => documentsStore.documents || [])

// Mock price books data
const priceBooks = ref<PriceBook[]>([
  {
    id: 'pb1',
    name: 'Standard Retail Pricing',
    description: 'Standard pricing for all retail customers',
    type: 'standard',
    status: 'active',
    currency: 'SAR',
    validFrom: '2025-01-01',
    customerIds: [],
    priceBookEntries: [
      {
        id: 'pbe1',
        productId: 'p1',
        listPrice: 15000,
        specialPrice: 14000
      },
      {
        id: 'pbe2',
        productId: 'p2',
        listPrice: 8500
      },
      {
        id: 'pbe3',
        productId: 'p3',
        listPrice: 2500,
        discountPercent: 5
      }
    ],
    notes: 'Default price book for all customers',
    createdAt: '2025-01-01T10:00:00Z',
    updatedAt: '2025-01-15T14:30:00Z'
  },
  {
    id: 'pb2',
    name: 'VIP Customer Pricing',
    description: 'Special pricing for VIP tier customers',
    type: 'contract',
    status: 'active',
    currency: 'SAR',
    validFrom: '2025-01-01',
    validUntil: '2025-12-31',
    customerIds: ['c1', 'c2'],
    priceBookEntries: [
      {
        id: 'pbe4',
        productId: 'p1',
        listPrice: 15000,
        specialPrice: 13500,
        discountPercent: 10
      },
      {
        id: 'pbe5',
        productId: 'p2',
        listPrice: 8500,
        specialPrice: 7800
      }
    ],
    notes: 'Annual contract pricing for VIP customers',
    createdAt: '2025-01-05T09:00:00Z',
    updatedAt: '2025-01-05T09:00:00Z'
  },
  {
    id: 'pb3',
    name: 'Volume Discount - Q1 2025',
    description: 'Volume-based pricing for bulk orders',
    type: 'volume',
    status: 'active',
    currency: 'SAR',
    validFrom: '2025-01-01',
    validUntil: '2025-03-31',
    customerIds: [],
    priceBookEntries: [
      {
        id: 'pbe6',
        productId: 'p1',
        listPrice: 15000,
        minimumQuantity: 10,
        discountPercent: 15
      },
      {
        id: 'pbe7',
        productId: 'p2',
        listPrice: 8500,
        minimumQuantity: 20,
        discountPercent: 12
      },
      {
        id: 'pbe8',
        productId: 'p3',
        listPrice: 2500,
        minimumQuantity: 50,
        discountPercent: 10
      }
    ],
    notes: 'Minimum quantity requirements apply',
    createdAt: '2024-12-20T11:00:00Z',
    updatedAt: '2025-01-02T16:00:00Z'
  },
  {
    id: 'pb4',
    name: 'Ramadan Promotion 2025',
    description: 'Special promotional pricing for Ramadan season',
    type: 'promotional',
    status: 'inactive',
    currency: 'SAR',
    validFrom: '2025-03-01',
    validUntil: '2025-04-15',
    customerIds: [],
    priceBookEntries: [
      {
        id: 'pbe9',
        productId: 'p1',
        listPrice: 15000,
        discountPercent: 20
      },
      {
        id: 'pbe10',
        productId: 'p3',
        listPrice: 2500,
        discountPercent: 25
      }
    ],
    notes: 'Promotional pricing valid during Ramadan period only',
    createdAt: '2025-01-10T08:00:00Z',
    updatedAt: '2025-01-10T08:00:00Z'
  },
  {
    id: 'pb5',
    name: 'Government Contract Pricing',
    description: 'Special pricing for government entities',
    type: 'contract',
    status: 'active',
    currency: 'SAR',
    validFrom: '2024-06-01',
    validUntil: '2026-05-31',
    customerIds: ['c4'],
    priceBookEntries: [
      {
        id: 'pbe11',
        productId: 'p1',
        listPrice: 15000,
        specialPrice: 14250,
        discountPercent: 5
      },
      {
        id: 'pbe12',
        productId: 'p2',
        listPrice: 8500,
        specialPrice: 8075,
        discountPercent: 5
      },
      {
        id: 'pbe13',
        productId: 'p3',
        listPrice: 2500,
        specialPrice: 2375,
        discountPercent: 5
      }
    ],
    notes: '2-year government contract with fixed pricing',
    createdAt: '2024-05-15T10:00:00Z',
    updatedAt: '2024-05-15T10:00:00Z'
  }
])

// Table columns
const priceBookColumns: TableColumn[] = [
  { key: 'name', label: 'Contract/Price Book Name', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'validity', label: 'Validity Period', sortable: false },
  { key: 'customers', label: 'Customers', sortable: false },
  { key: 'items', label: 'Items', sortable: false },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'approval', label: 'Approval', sortable: false },
  { key: 'actions', label: 'Actions', align: 'right' }
]

// Filtered price books
const filteredPriceBooks = computed(() => {
  let filtered = priceBooks.value

  // Filter by status
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(pb => pb.status === statusFilter.value)
  }

  // Filter by type
  if (typeFilter.value !== 'all') {
    filtered = filtered.filter(pb => pb.type === typeFilter.value)
  }

  // Filter by search query
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(pb =>
      pb.name.toLowerCase().includes(query) ||
      pb.description.toLowerCase().includes(query)
    )
  }

  return filtered
})

const activePriceBooksCount = computed(() =>
  priceBooks.value.filter(pb => pb.status === 'active').length
)

const pendingApprovalCount = computed(() =>
  priceBooks.value.filter(pb => pb.status === 'pending_approval').length
)

// Helper functions
function getCustomerName(customerId: string) {
  const customer = customersStore.getCustomerById(customerId)
  return customer?.name || 'Unknown Customer'
}

function getProductCode(productId: string) {
  const product = productsStore.getProductById(productId)
  return product?.code || 'N/A'
}

function getProductName(productId: string) {
  const product = productsStore.getProductById(productId)
  return product?.name || 'Unknown Product'
}

function getProductSku(productId: string) {
  const product = productsStore.getProductById(productId)
  return product?.sku || 'N/A'
}

function getDocumentName(documentId: string) {
  const doc = documentsStore.getDocumentById(documentId)
  return doc?.name || 'Unknown Document'
}

function getDocumentType(documentId: string) {
  const doc = documentsStore.getDocumentById(documentId)
  return doc?.type || 'N/A'
}

function isServiceEntry(productId: string) {
  const product = productsStore.getProductById(productId)
  return product?.unitOfMeasure === 'Month'
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function formatType(type: PriceBook['type']) {
  const types: Record<PriceBook['type'], string> = {
    standard: 'Standard',
    volume: 'Volume Discount',
    contract: 'Contract Pricing',
    promotional: 'Promotional',
    customer_specific: 'Customer Specific'
  }
  return types[type]
}

function getTypeVariant(type: PriceBook['type']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<PriceBook['type'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    standard: 'default',
    volume: 'info',
    contract: 'warning',
    promotional: 'success',
    customer_specific: 'primary'
  }
  return variants[type]
}

function formatStatus(status: PriceBook['status']): string {
  const statusMap: Record<PriceBook['status'], string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    active: 'Active',
    inactive: 'Inactive',
    expired: 'Expired'
  }
  return statusMap[status]
}

function getStatusVariant(status: PriceBook['status']): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variantMap: Record<PriceBook['status'], 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    draft: 'default',
    pending_approval: 'warning',
    active: 'success',
    inactive: 'default',
    expired: 'danger'
  }
  return variantMap[status]
}

function canEdit(priceBook: PriceBook): boolean {
  // Can only edit if draft or pending approval (before active)
  return priceBook.status === 'draft' || priceBook.status === 'pending_approval'
}

function isExpiringSoon(validUntil: string) {
  const daysUntilExpiry = Math.ceil(
    (new Date(validUntil).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24)
  )
  return daysUntilExpiry <= 30 && daysUntilExpiry >= 0
}

function calculateFinalPrice(entry: PriceBookEntry): number {
  let price = entry.specialPrice || entry.listPrice
  if (entry.discountPercent) {
    price = price * (1 - entry.discountPercent / 100)
  }
  return price
}

// Actions
function createNewPriceBook() {
  resetPriceBookForm()
  showCreateModal.value = true
}

function resetPriceBookForm() {
  priceBookForm.value = {
    name: '',
    description: '',
    type: 'standard',
    status: 'draft',
    currency: 'SAR',
    validFrom: new Date().toISOString().split('T')[0],
    validUntil: '',
    contractDuration: '',
    autoRenewal: false,
    renewalNoticeDays: 30,
    paymentTerms: '',
    deliveryTerms: '',
    terms: '',
    sla: ''
  }
  selectedCustomers.value = []
  priceBookEntries.value = []
  attachedDocuments.value = []
  editingPriceBook.value = null
}

// Two-dropdown helper functions
function getCategoryLabel(): string {
  switch (entryForm.value.category) {
    case 'product':
      return 'Product'
    case 'recurring':
      return 'Recurring Service'
    case 'labor':
      return 'Labor Position'
    default:
      return 'Item'
  }
}

function getFilteredItems() {
  return availableItems.value.filter(item => item.type === entryForm.value.category)
}

// Filtered product suggestions for autocomplete - limited to 10 results
const filteredProductSuggestions = computed(() => {
  if (!productSearchQuery.value || productSearchQuery.value.length < 2 || !entryForm.value.category) {
    return []
  }

  const query = productSearchQuery.value.toLowerCase()
  const items = getFilteredItems()

  // Search by name or SKU
  const filtered = items.filter(item =>
    item.name.toLowerCase().includes(query) ||
    (item.sku && item.sku.toLowerCase().includes(query))
  )

  // Limit to 10 suggestions for autocomplete
  return filtered.slice(0, 10)
})

// Autocomplete handlers
function handleProductSearchInput() {
  if (productSearchQuery.value.length >= 2) {
    showProductSuggestions.value = true
  } else {
    showProductSuggestions.value = false
  }
  // Clear selection if user types after selecting
  if (entryForm.value.productId) {
    entryForm.value.productId = ''
    entryForm.value.customPrice = 0
  }
}

function handleProductSearchBlur() {
  // Delay hiding to allow click events on suggestions to fire
  setTimeout(() => {
    showProductSuggestions.value = false
  }, 200)
}

function selectProductItem(item: any) {
  productSearchQuery.value = `${item.sku ? item.sku + ' - ' : ''}${item.name}`
  entryForm.value.productId = item.id
  showProductSuggestions.value = false
  onProductSelect()
}

function onCategoryChange() {
  // Reset product selection when category changes
  entryForm.value.productId = ''
  entryForm.value.customPrice = 0
  productSearchQuery.value = ''
  showProductSuggestions.value = false
}

// Line Items / Entry Management Functions
function onProductSelect() {
  if (!entryForm.value.productId) return

  if (selectedProduct.value) {
    entryForm.value.customPrice = selectedProduct.value.sellingPrice

    // Determine item type
    const itemType = selectedProduct.value.unitOfMeasure === 'Month'
      ? 'Recurring Service'
      : selectedProduct.value.unitOfMeasure === 'Hour'
        ? 'Labor Position'
        : 'Product'

    success(`${itemType} loaded: ${selectedProduct.value.name} - ${formatCurrency(selectedProduct.value.sellingPrice)}${selectedProduct.value.unitOfMeasure === 'Month' ? '/month' : selectedProduct.value.unitOfMeasure === 'Hour' ? '/hr' : ''}`)
  }
}

function calculateDiscount(): string {
  if (!selectedProduct.value || !entryForm.value.customPrice) return '0.00'
  const listPrice = selectedProduct.value.sellingPrice
  const customPrice = entryForm.value.customPrice
  if (listPrice === 0) return '0.00'
  const discount = ((listPrice - customPrice) / listPrice) * 100
  return discount.toFixed(2)
}

function calculateStandardMargin(): string {
  if (!selectedProduct.value) return '0.00'
  const sellingPrice = selectedProduct.value.sellingPrice
  const landedCost = selectedProduct.value.landedCostSAR
  if (sellingPrice === 0) return '0.00'
  const margin = ((sellingPrice - landedCost) / sellingPrice) * 100
  return margin.toFixed(2)
}

function calculateCustomMargin(): string {
  if (!selectedProduct.value || !entryForm.value.customPrice) return '0.00'
  const customPrice = entryForm.value.customPrice
  const landedCost = selectedProduct.value.landedCostSAR
  if (customPrice === 0) return '0.00'
  const margin = ((customPrice - landedCost) / customPrice) * 100
  return margin.toFixed(2)
}

function saveEntry() {
  if (!entryForm.value.productId || !entryForm.value.customPrice) {
    info('Please select a product and enter a custom price')
    return
  }

  const product = selectedProduct.value
  if (!product) {
    info('Product not found')
    return
  }

  const newEntry: PriceBookEntry = {
    id: `entry-${Date.now()}`,
    productId: entryForm.value.productId,
    listPrice: product.sellingPrice,
    specialPrice: entryForm.value.customPrice,
    discountPercent: parseFloat(calculateDiscount()),
    minimumQuantity: entryForm.value.minQuantity
  }

  if (editingEntryIndex.value !== null) {
    // Update existing entry
    priceBookEntries.value[editingEntryIndex.value] = newEntry
    success('Entry updated successfully')
  } else {
    // Add new entry
    priceBookEntries.value.push(newEntry)
    success('Product added to price book')
  }

  closeEntryModal()
}

function editEntry(index: number) {
  const entry = priceBookEntries.value[index]
  const product = productsStore.getProductById(entry.productId)

  if (product) {
    entryForm.value = {
      productId: entry.productId,
      customPrice: entry.specialPrice || entry.listPrice,
      minQuantity: entry.minimumQuantity || 1,
      notes: ''
    }
    editingEntryIndex.value = index
    showAddEntryModal.value = true
  }
}

function removeEntry(index: number) {
  priceBookEntries.value.splice(index, 1)
  success('Entry removed')
}

function closeEntryModal() {
  showAddEntryModal.value = false
  editingEntryIndex.value = null
  entryForm.value = {
    category: '',
    productId: '',
    customPrice: 0,
    minQuantity: 1,
    notes: ''
  }
}

// Document Management Functions
function attachDocument() {
  if (!documentForm.value.documentId) {
    info('Please select a document')
    return
  }

  // Check if document is already attached
  const alreadyAttached = attachedDocuments.value.some(
    doc => doc.documentId === documentForm.value.documentId
  )
  if (alreadyAttached) {
    info('This document is already attached')
    return
  }

  attachedDocuments.value.push({
    documentId: documentForm.value.documentId,
    customName: documentForm.value.customName || undefined,
    includeInPdf: documentForm.value.includeInPdf
  })

  success('Document attached successfully')
  closeDocumentModal()
}

function removeDocument(index: number) {
  attachedDocuments.value.splice(index, 1)
  success('Document removed')
}

function moveDocumentUp(index: number) {
  if (index > 0) {
    const temp = attachedDocuments.value[index]
    attachedDocuments.value[index] = attachedDocuments.value[index - 1]
    attachedDocuments.value[index - 1] = temp
  }
}

function moveDocumentDown(index: number) {
  if (index < attachedDocuments.value.length - 1) {
    const temp = attachedDocuments.value[index]
    attachedDocuments.value[index] = attachedDocuments.value[index + 1]
    attachedDocuments.value[index + 1] = temp
  }
}

function closeDocumentModal() {
  showAddDocumentModal.value = false
  documentForm.value = {
    documentId: '',
    customName: '',
    includeInPdf: true
  }
}

function closeCreateModal() {
  showCreateModal.value = false
  resetPriceBookForm()
}

function savePriceBook() {
  // Validation
  if (!priceBookForm.value.name || !priceBookForm.value.description) {
    info('Please fill in all required fields')
    return
  }

  if (editingPriceBook.value) {
    // Update existing
    const index = priceBooks.value.findIndex(pb => pb.id === editingPriceBook.value!.id)
    if (index !== -1) {
      priceBooks.value[index] = {
        ...priceBooks.value[index],
        name: priceBookForm.value.name,
        description: priceBookForm.value.description,
        type: priceBookForm.value.type,
        status: priceBookForm.value.status as 'active' | 'inactive',
        currency: priceBookForm.value.currency,
        validFrom: priceBookForm.value.validFrom,
        validUntil: priceBookForm.value.validUntil || undefined,
        customerIds: selectedCustomers.value,
        priceBookEntries: [...priceBookEntries.value],
        updatedAt: new Date().toISOString()
      }
      success('Price book updated successfully')
    }
  } else {
    // Create new
    const newPriceBook: PriceBook = {
      id: `pb${Date.now()}`,
      name: priceBookForm.value.name,
      description: priceBookForm.value.description,
      type: priceBookForm.value.type,
      status: priceBookForm.value.status as 'active' | 'inactive',
      currency: priceBookForm.value.currency,
      validFrom: priceBookForm.value.validFrom,
      validUntil: priceBookForm.value.validUntil || undefined,
      customerIds: selectedCustomers.value,
      priceBookEntries: [...priceBookEntries.value],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    priceBooks.value.push(newPriceBook)

    if (priceBookEntries.value.length > 0) {
      success(`Price book created with ${priceBookEntries.value.length} product(s)!`)
    } else {
      success('Price book created successfully!')
    }
  }

  closeCreateModal()
}

function viewPriceBookDetails(priceBook: PriceBook) {
  selectedPriceBook.value = priceBook
  showDetailsModal.value = true
}

function closeDetailsModal() {
  showDetailsModal.value = false
  selectedPriceBook.value = null
}

function editPriceBook(priceBook: PriceBook) {
  editingPriceBook.value = priceBook
  priceBookForm.value = {
    name: priceBook.name,
    description: priceBook.description,
    type: priceBook.type,
    status: priceBook.status as any,
    currency: priceBook.currency,
    validFrom: priceBook.validFrom,
    validUntil: priceBook.validUntil || '',
    contractDuration: '',
    autoRenewal: false,
    renewalNoticeDays: 30,
    paymentTerms: '',
    deliveryTerms: '',
    terms: '',
    sla: ''
  }
  selectedCustomers.value = [...priceBook.customerIds]
  priceBookEntries.value = priceBook.priceBookEntries ? [...priceBook.priceBookEntries] : []
  closeDetailsModal()
  showCreateModal.value = true
}

function duplicatePriceBook(priceBook: PriceBook) {
  const duplicate: PriceBook = {
    ...priceBook,
    id: `pb${Date.now()}`,
    name: `${priceBook.name} (Copy)`,
    status: 'inactive',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  }
  priceBooks.value.push(duplicate)
  success(`Price book duplicated: ${duplicate.name}`)
}

function toggleStatus(priceBook: PriceBook) {
  const newStatus = priceBook.status === 'active' ? 'inactive' : 'active'
  priceBook.status = newStatus
  success(`Price book ${newStatus === 'active' ? 'activated' : 'deactivated'}`)
}

// Workflow action functions
function submitForApproval(priceBook: PriceBook) {
  if (priceBook.status !== 'draft' && priceBook.status !== 'pending_approval') {
    info('Only draft contracts can be submitted for approval')
    return
  }

  // Check if all required fields are filled
  if (!priceBook.name || !priceBook.description || priceBook.priceBookEntries.length === 0) {
    info('Please fill in all required fields and add at least one line item before submitting')
    return
  }

  priceBook.status = 'pending_approval'
  priceBook.requiresApproval = true
  priceBook.approvalReason = 'Awaiting approval for activation'
  priceBook.updatedAt = new Date().toISOString()

  success(`Contract "${priceBook.name}" submitted for approval`)
}

function approvePriceBook(priceBook: PriceBook) {
  if (priceBook.status !== 'pending_approval') {
    info('Only pending contracts can be approved')
    return
  }

  priceBook.status = 'active'
  priceBook.approvedBy = 'USR-001' // Should come from auth
  priceBook.approvedAt = new Date().toISOString()
  priceBook.updatedAt = new Date().toISOString()

  success(`Contract "${priceBook.name}" has been approved and activated`)
}

function rejectPriceBook(priceBook: PriceBook) {
  if (priceBook.status !== 'pending_approval') {
    info('Only pending contracts can be rejected')
    return
  }

  const reason = prompt('Please enter reason for rejection:')
  if (!reason) return

  priceBook.status = 'draft'
  priceBook.rejectionReason = reason
  priceBook.requiresApproval = false
  priceBook.updatedAt = new Date().toISOString()

  info(`Contract "${priceBook.name}" has been rejected and returned to draft`)
}

function deactivatePriceBook(priceBook: PriceBook) {
  if (priceBook.status !== 'active') {
    info('Only active contracts can be deactivated')
    return
  }

  priceBook.status = 'inactive'
  priceBook.updatedAt = new Date().toISOString()

  success(`Contract "${priceBook.name}" has been deactivated`)
}

function reactivatePriceBook(priceBook: PriceBook) {
  if (priceBook.status !== 'inactive') {
    info('Only inactive contracts can be reactivated')
    return
  }

  priceBook.status = 'active'
  priceBook.updatedAt = new Date().toISOString()

  success(`Contract "${priceBook.name}" has been reactivated`)
}
</script>

<style scoped>
.pricebooks-view {
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

.btn-secondary {
  background: var(--color-border);
  color: var(--color-text-primary);
}

.btn-secondary:hover {
  background: var(--color-surface-hover);
}

.btn-success {
  background: var(--color-success, #4caf50);
  color: #fff;
}

.btn-success:hover {
  background: var(--color-success-hover, #45a049);
  transform: translateY(-2px);
}

.btn-danger {
  background: var(--color-danger, #f44336);
  color: #fff;
}

.btn-danger:hover {
  background: var(--color-danger-hover, #d32f2f);
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

.name-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 350px;
}

.name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.description {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.4;
}

.validity-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 0.75rem;
  min-width: 280px;
}

.date-row {
  font-size: 1rem;
  display: flex;
  gap: 0.75rem;
}

.date-label {
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 60px;
}

.customers-cell,
.items-cell {
  font-size: 1.05rem;
  padding: 1rem 0.75rem;
}

.customer-count {
  color: var(--color-primary);
  font-weight: 600;
  font-size: 1.1rem;
}

.item-count {
  color: var(--color-text-primary);
  font-weight: 600;
  font-size: 1.1rem;
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

.text-danger {
  color: var(--color-danger) !important;
}

.text-success {
  color: var(--color-success) !important;
}

.text-warning {
  color: var(--color-warning, #ff9800) !important;
}

.text-right {
  text-align: right;
}

/* Modal Styles */
.pricebook-details {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.details-section {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 2rem;
}

.details-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.details-section h3 {
  margin: 0 0 1.25rem 0;
  font-size: 1.25rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 1.05rem;
  color: var(--color-text-primary);
}

.empty-message {
  padding: 1.25rem;
  background: var(--color-background);
  border-radius: 8px;
  color: var(--color-text-secondary);
  font-style: italic;
  font-size: 1rem;
}

.customer-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.customer-badge {
  padding: 0.65rem 1.25rem;
  background: var(--color-primary);
  color: #fff;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 500;
}

.entries-table {
  overflow-x: auto;
}

.entries-table table {
  width: 100%;
  border-collapse: collapse;
}

.entries-table th,
.entries-table td {
  padding: 1rem 1.25rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}

.entries-table th {
  background: var(--color-background);
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.entries-table td {
  font-size: 1.05rem;
  color: var(--color-text-primary);
}

.special-price {
  color: var(--color-success);
  font-weight: 600;
  font-size: 1.1rem;
}

.discount-badge {
  background: var(--color-danger);
  color: #fff;
  padding: 0.35rem 0.65rem;
  border-radius: 5px;
  font-size: 0.9rem;
  font-weight: 600;
}

.final-price {
  font-weight: 700;
  color: var(--color-primary);
  font-size: 1.15rem;
  background: var(--color-primary-light);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  display: inline-block;
}

.notes-content {
  padding: 1.25rem;
  background: var(--color-background);
  border-radius: 8px;
  color: var(--color-text-primary);
  line-height: 1.6;
  font-size: 1rem;
}

.modal-actions {
  display: flex;
  gap: 1.25rem;
  justify-content: flex-end;
}

/* Price Book Form */
.pricebook-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-border);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.form-input {
  padding: 0.625rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.form-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  cursor: pointer;
  width: 18px;
  height: 18px;
}

/* Line Items / Entries Styles */
.line-items-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.entries-list {
  border: 1px solid var(--color-border);
  border-radius: 8px;
  overflow: hidden;
}

.entries-table {
  width: 100%;
  border-collapse: collapse;
}

.entries-table thead {
  background: var(--color-background);
}

.entries-table th {
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.95rem;
  border-bottom: 2px solid var(--color-border);
}

.entries-table th.text-right {
  text-align: right;
}

.entries-table th.text-center {
  text-align: center;
}

.entries-table tbody tr {
  border-bottom: 1px solid var(--color-border);
}

.entries-table tbody tr:last-child {
  border-bottom: none;
}

.entries-table tbody tr:hover {
  background: var(--color-background);
}

.entries-table td {
  padding: 1rem;
  color: var(--color-text-primary);
}

.entries-table td.text-right {
  text-align: right;
}

.entries-table td.text-center {
  text-align: center;
}

.product-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.product-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.product-sku {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}

.price-value {
  font-weight: 600;
  color: var(--color-text-primary);
}

.discount-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  background: var(--color-success-light, #e8f5e9);
  color: var(--color-success, #4caf50);
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.9rem;
}

.entry-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

.entry-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  font-size: 1.2rem;
  opacity: 0.7;
  transition: all 0.2s;
}

.entry-action-btn:hover {
  opacity: 1;
  transform: scale(1.1);
}

.empty-state {
  padding: 2rem;
  text-align: center;
  color: var(--color-text-secondary);
  border: 1px dashed var(--color-border);
  border-radius: 8px;
}

/* Entry Form Modal */
.entry-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Autocomplete styles */
.autocomplete-container {
  position: relative;
}

.autocomplete-wrapper {
  position: relative;
}

.autocomplete-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: var(--shadow-lg);
  max-height: 400px;
  overflow-y: auto;
  z-index: 1000;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.autocomplete-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
  border-radius: 10px 10px 0 0;
}

.suggestion-count {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.autocomplete-item {
  padding: 1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: background-color 0.15s;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.autocomplete-item:last-child {
  border-bottom: none;
  border-radius: 0 0 10px 10px;
}

.autocomplete-item:hover {
  background: var(--color-surface-hover);
}

.suggestion-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.suggestion-icon {
  font-size: 1.2rem;
}

.suggestion-sku {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 0.95rem;
  min-width: 100px;
}

.suggestion-name {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 0.95rem;
  flex: 1;
}

.suggestion-meta {
  display: flex;
  gap: 1rem;
  align-items: center;
  padding-left: 2rem;
}

.suggestion-price {
  font-size: 0.875rem;
  color: var(--color-success);
  font-weight: 600;
}

.suggestion-department {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  background: var(--color-background);
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
}

.form-hint {
  display: block;
  margin-top: 0.5rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.discount-highlight {
  color: var(--color-success);
  font-weight: 600;
}

/* Product Details in Entry Form */
.product-details {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
}

.detail-row .label {
  font-weight: 600;
  color: var(--color-text-secondary);
}

.detail-row .value {
  font-weight: 600;
  color: var(--color-text-primary);
}

.detail-row .cost-value {
  color: #ef4444;
  font-weight: 700;
}

.detail-row .price-value {
  color: #10b981;
  font-weight: 700;
}

.detail-row .margin-value {
  color: #3b82f6;
  font-weight: 700;
  font-size: 1.1em;
}

.service-badge {
  color: var(--color-primary);
  font-weight: 600;
}

/* Service/Product Type Badges */
.service-type-badge,
.product-type-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.85rem;
  font-weight: 600;
}

.service-type-badge {
  background: var(--color-primary-light, #e3f2fd);
  color: var(--color-primary, #1976d2);
}

.product-type-badge {
  background: var(--color-success-light, #e8f5e9);
  color: var(--color-success, #4caf50);
}

.price-suffix {
  display: block;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  margin-top: 0.15rem;
}

/* Documents Styles */
.documents-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.documents-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.document-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-background);
}

.document-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
}

.document-icon {
  font-size: 1.5rem;
}

.document-details {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.document-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.document-meta {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  text-transform: capitalize;
}

.document-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.checkbox-inline {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.checkbox-inline input[type="checkbox"] {
  cursor: pointer;
  width: 16px;
  height: 16px;
}

.document-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Workflow Progress Indicator */
.workflow-progress {
  background: var(--color-background);
  border-radius: 12px;
  padding: 2rem;
  margin-bottom: 2rem;
  border: 2px solid var(--color-border);
}

.progress-steps {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
}

.progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  flex: 0 0 auto;
  z-index: 1;
}

.step-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: var(--color-surface);
  border: 3px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.75rem;
  transition: all 0.3s;
}

.progress-step.active .step-icon {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.progress-step.completed .step-icon {
  background: var(--color-success);
  border-color: var(--color-success);
  color: white;
}

.step-label {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.progress-step.active .step-label {
  color: var(--color-primary);
}

.progress-step.completed .step-label {
  color: var(--color-success);
}

.progress-line {
  flex: 1;
  height: 3px;
  background: var(--color-border);
  margin: 0 1rem;
  position: relative;
  top: -15px;
  transition: all 0.3s;
}

.progress-line.active {
  background: var(--color-primary);
}

/* Approval Information */
.approval-info {
  background: var(--color-background);
  border-radius: 8px;
  padding: 1.5rem;
}

.approval-notice {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 8px;
  border-left: 4px solid;
}

.approval-notice.warning {
  background: rgba(255, 152, 0, 0.1);
  border-left-color: var(--color-warning, #ff9800);
}

.approval-notice.danger {
  background: rgba(244, 67, 54, 0.1);
  border-left-color: var(--color-danger, #f44336);
}

.approval-notice .approval-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.approval-title {
  font-weight: 600;
  font-size: 1.05rem;
  color: var(--color-text-primary);
  margin-bottom: 0.35rem;
}

.approval-reason {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* Approval Cell in Table */
.approval-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.95rem;
}

.approval-icon {
  font-size: 1.1rem;
}

.approval-text {
  color: var(--color-text-secondary);
}

/* Action button danger variant */
.action-btn-danger {
  color: var(--color-danger, #f44336);
}

.action-btn-danger:hover {
  color: var(--color-danger, #d32f2f);
  transform: scale(1.2);
}

.action-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
  pointer-events: none;
}

/* Modern Premium Details View */
.pricebook-details-modern {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Modern Progress Card */
.modern-progress-card {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 16px;
  padding: 2rem;
  border: 1px solid rgba(59, 130, 246, 0.1);
}

.modern-progress-card .progress-steps {
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
}

.modern-progress-card .progress-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  flex: 0 0 auto;
  z-index: 1;
}

.modern-progress-card .step-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: var(--color-surface);
  border: 3px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.modern-progress-card .progress-step.active .step-icon {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-color: transparent;
  transform: scale(1.15);
  box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4);
}

.modern-progress-card .progress-step.completed .step-icon {
  background: linear-gradient(135deg, #10b981, #059669);
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.modern-progress-card .step-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.modern-progress-card .progress-step.active .step-label {
  color: var(--color-primary);
  font-weight: 700;
}

.modern-progress-card .progress-step.completed .step-label {
  color: #10b981;
}

.modern-progress-card .progress-line {
  flex: 1;
  height: 3px;
  background: var(--color-border);
  margin: 0 1rem;
  position: relative;
  top: -20px;
  transition: all 0.3s;
}

.modern-progress-card .progress-line.active {
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

/* Modern Info Card */
.modern-info-card {
  background: var(--color-surface);
  border-radius: 16px;
  border: 1px solid var(--color-border);
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.modern-info-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}

.card-header-mini {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.03), rgba(139, 92, 246, 0.03));
  border-bottom: 1px solid var(--color-border);
}

.card-header-mini h4 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.count-badge {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 0.375rem 0.875rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.card-content {
  padding: 1.5rem;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.25rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.info-value.highlight {
  font-size: 1.125rem;
  color: var(--color-primary);
  font-weight: 700;
}

.description-box {
  margin-top: 1.25rem;
  padding: 1rem 1.25rem;
  background: var(--color-background);
  border-radius: 12px;
  border: 1px solid var(--color-border);
}

.description-text {
  margin: 0.5rem 0 0 0;
  line-height: 1.6;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
}

/* Alert Boxes */
.alert-box {
  padding: 1rem 1.25rem;
  border-radius: 12px;
  border-left: 4px solid;
  margin-bottom: 1rem;
}

.alert-box.warning {
  background: rgba(251, 146, 60, 0.08);
  border-left-color: #fb923c;
}

.alert-box.danger {
  background: rgba(239, 68, 68, 0.08);
  border-left-color: #ef4444;
}

.alert-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.alert-title {
  font-weight: 700;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.alert-message {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
}

.info-row:last-child {
  border-bottom: none;
}

/* Chips Container */
.chips-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.modern-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.625rem 1.125rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 24px;
  color: var(--color-primary);
  font-size: 0.9375rem;
  font-weight: 600;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.modern-chip:hover {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(139, 92, 246, 0.15));
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
}

/* Empty State Mini */
.empty-state-mini {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2rem;
  background: var(--color-background);
  border-radius: 12px;
  border: 2px dashed var(--color-border);
}

.empty-icon {
  font-size: 2.5rem;
  opacity: 0.5;
}

.empty-text {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

/* Items List Modern */
.items-list-modern {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.item-card-mini {
  display: flex;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.item-card-mini:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.1);
  transform: translateX(4px);
}

.item-number-mini {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.item-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.item-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.item-name {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.item-code {
  font-family: 'Courier New', monospace;
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--color-primary);
  opacity: 0.8;
}

.item-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.4;
}

.item-final-price {
  font-size: 1.25rem;
  font-weight: 800;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 0.5rem 1rem;
  border-radius: 10px;
  white-space: nowrap;
}

.item-pricing-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
}

.pricing-detail {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pricing-label {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.pricing-value {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.pricing-value.special {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 0.25rem 0.625rem;
  border-radius: 6px;
}

.discount-pill {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.8125rem;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
}

/* Notes Box */
.notes-box {
  padding: 1rem 1.25rem;
  background: var(--color-background);
  border-radius: 12px;
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  line-height: 1.7;
  font-size: 0.9375rem;
  font-style: italic;
}
</style>
