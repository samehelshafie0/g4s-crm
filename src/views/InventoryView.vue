<template>
  <div class="inventory-view">
    <!-- Stats Bar -->
    <div class="stats-bar">
      <div class="stat-card">
        <div class="stat-label">Total Stock Value</div>
        <div class="stat-value">{{ formatCurrency(stats.totalValue) }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Total Items</div>
        <div class="stat-value">{{ stats.totalItems }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Available</div>
        <div class="stat-value success">{{ stats.totalAvailable }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Reserved</div>
        <div class="stat-value warning">{{ stats.totalReserved }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Low Stock Items</div>
        <div class="stat-value danger">{{ stats.lowStockCount }}</div>
      </div>
    </div>

    <!-- Header -->
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search by product SKU or name..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="showAddMovementModal = true">
          <span>📦</span> Record Movement
        </button>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button
        class="tab"
        :class="{ active: activeTab === 'stock' }"
        @click="activeTab = 'stock'"
      >
        <span class="tab-icon">📊</span>
        Stock Levels
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'movements' }"
        @click="activeTab = 'movements'"
      >
        <span class="tab-icon">🔄</span>
        Stock Movements
      </button>
    </div>

    <!-- Stock Levels Tab -->
    <div v-if="activeTab === 'stock'">
      <div class="filters-bar">
        <div class="filter-group">
          <label>Location:</label>
          <select v-model="locationFilter" class="filter-select">
            <option value="all">All Locations</option>
            <option value="Riyadh Main">Riyadh Main</option>
            <option value="Jeddah Branch">Jeddah Branch</option>
            <option value="Dammam Branch">Dammam Branch</option>
          </select>
        </div>

        <div class="filter-group">
          <label>Stock Status:</label>
          <select v-model="stockStatusFilter" class="filter-select">
            <option value="all">All Items</option>
            <option value="low">Low Stock (< 10)</option>
            <option value="available">In Stock</option>
            <option value="reserved">Has Reservations</option>
          </select>
        </div>
      </div>

      <BaseCard no-padding>
        <div class="table-container">
          <table class="inventory-table">
            <thead>
              <tr>
                <th>Product</th>
                <th>Location</th>
                <th class="text-right">On Hand</th>
                <th class="text-right">Reserved</th>
                <th class="text-right">Available</th>
                <th class="text-right">Unit Cost</th>
                <th class="text-right">Total Value</th>
                <th>Last Updated</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in filteredStock" :key="item.id" class="stock-row">
                <td>
                  <div class="product-info">
                    <div class="product-name">{{ getProductName(item.productId) }}</div>
                    <div class="product-sku">{{ getProductSKU(item.productId) }}</div>
                    <div v-if="item.batchNumber" class="product-batch">Batch: {{ item.batchNumber }}</div>
                  </div>
                </td>
                <td>
                  <BaseBadge variant="info" size="sm">{{ item.warehouseLocation }}</BaseBadge>
                </td>
                <td class="text-right">{{ item.quantityOnHand }}</td>
                <td class="text-right">
                  <span v-if="item.quantityReserved > 0" class="reserved-qty">{{ item.quantityReserved }}</span>
                  <span v-else>-</span>
                </td>
                <td class="text-right">
                  <span :class="getStockClass(item.quantityAvailable)">{{ item.quantityAvailable }}</span>
                </td>
                <td class="text-right">{{ formatCurrency(item.unitCost) }}</td>
                <td class="text-right">{{ formatCurrency(item.totalValue) }}</td>
                <td>{{ formatDate(item.lastUpdated) }}</td>
                <td>
                  <div class="action-buttons">
                    <button class="action-btn" title="Adjust" @click="adjustStock(item)">⚙️</button>
                    <button class="action-btn" title="View History" @click="viewHistory(item)">📋</button>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredStock.length === 0">
                <td colspan="9" class="empty-row">
                  No stock found matching your criteria
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </BaseCard>
    </div>

    <!-- Stock Movements Tab -->
    <div v-if="activeTab === 'movements'">
      <div class="filters-bar">
        <div class="filter-group">
          <label>Movement Type:</label>
          <select v-model="movementTypeFilter" class="filter-select">
            <option value="all">All Types</option>
            <option value="receipt">Receipts</option>
            <option value="issue">Issues</option>
            <option value="transfer">Transfers</option>
            <option value="adjustment">Adjustments</option>
          </select>
        </div>

        <div class="filter-group">
          <label>Location:</label>
          <select v-model="movementLocationFilter" class="filter-select">
            <option value="all">All Locations</option>
            <option value="Riyadh Main">Riyadh Main</option>
            <option value="Jeddah Branch">Jeddah Branch</option>
            <option value="Dammam Branch">Dammam Branch</option>
          </select>
        </div>
      </div>

      <BaseCard no-padding>
        <div class="table-container">
          <table class="inventory-table">
            <thead>
              <tr>
                <th>Movement #</th>
                <th>Type</th>
                <th>Product</th>
                <th>Location</th>
                <th class="text-right">Quantity</th>
                <th>Reference</th>
                <th>Date</th>
                <th>Performed By</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="movement in filteredMovements" :key="movement.id" class="movement-row">
                <td>
                  <span class="movement-number">{{ movement.movementNumber }}</span>
                </td>
                <td>
                  <BaseBadge :variant="getMovementTypeVariant(movement.movementType)" size="sm">
                    {{ formatMovementType(movement.movementType) }}
                  </BaseBadge>
                </td>
                <td>
                  <div class="product-info-sm">
                    <div>{{ getProductName(movement.productId) }}</div>
                    <div class="product-sku-sm">{{ getProductSKU(movement.productId) }}</div>
                  </div>
                </td>
                <td>
                  <div v-if="movement.movementType === 'transfer'">
                    <div class="transfer-info">
                      <span class="from-location">{{ movement.fromLocation }}</span>
                      <span class="transfer-arrow">→</span>
                      <span class="to-location">{{ movement.toLocation }}</span>
                    </div>
                  </div>
                  <div v-else>
                    {{ movement.warehouseLocation }}
                  </div>
                </td>
                <td class="text-right">
                  <span :class="getQuantityClass(movement.movementType, movement.quantity)">
                    {{ getQuantityDisplay(movement.movementType, movement.quantity) }}
                  </span>
                </td>
                <td>
                  <div v-if="movement.referenceNumber">
                    <div class="reference-type">{{ formatReferenceType(movement.referenceType) }}</div>
                    <div class="reference-number">{{ movement.referenceNumber }}</div>
                  </div>
                  <span v-else>-</span>
                </td>
                <td>{{ formatDate(movement.movementDate) }}</td>
                <td>{{ movement.performedBy }}</td>
              </tr>
              <tr v-if="filteredMovements.length === 0">
                <td colspan="8" class="empty-row">
                  No movements found matching your criteria
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </BaseCard>
    </div>

    <!-- Add Movement Modal -->
    <BaseModal
      v-model="showAddMovementModal"
      title="Record Stock Movement"
      size="lg"
      @confirm="saveMovement"
    >
      <div class="modal-form">
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Movement Type *</label>
            <select v-model="movementForm.movementType" class="form-input" required>
              <option value="receipt">Receipt (Goods In)</option>
              <option value="issue">Issue (Goods Out)</option>
              <option value="transfer">Transfer</option>
              <option value="adjustment">Adjustment</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Product *</label>
            <select v-model="movementForm.productId" class="form-input" required>
              <option value="">Select Product</option>
              <option
                v-for="product in productsStore.getActiveProducts()"
                :key="product.id"
                :value="product.id"
              >
                {{ product.sku }} - {{ product.name }}
              </option>
            </select>
          </div>
        </div>

        <div v-if="movementForm.movementType !== 'transfer'" class="form-row">
          <div class="form-group">
            <label class="form-label">Location *</label>
            <select v-model="movementForm.warehouseLocation" class="form-input" required>
              <option value="Riyadh Main">Riyadh Main</option>
              <option value="Jeddah Branch">Jeddah Branch</option>
              <option value="Dammam Branch">Dammam Branch</option>
            </select>
          </div>
        </div>

        <div v-if="movementForm.movementType === 'transfer'" class="form-row">
          <div class="form-group">
            <label class="form-label">From Location *</label>
            <select v-model="movementForm.fromLocation" class="form-input" required>
              <option value="Riyadh Main">Riyadh Main</option>
              <option value="Jeddah Branch">Jeddah Branch</option>
              <option value="Dammam Branch">Dammam Branch</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">To Location *</label>
            <select v-model="movementForm.toLocation" class="form-input" required>
              <option value="Riyadh Main">Riyadh Main</option>
              <option value="Jeddah Branch">Jeddah Branch</option>
              <option value="Dammam Branch">Dammam Branch</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Quantity *</label>
            <input
              v-model.number="movementForm.quantity"
              type="number"
              class="form-input"
              min="1"
              required
            />
          </div>
        </div>

        <div v-if="movementForm.movementType !== 'adjustment'" class="form-row">
          <div class="form-group">
            <label class="form-label">Reference Type</label>
            <select v-model="movementForm.referenceType" class="form-input">
              <option value="">None</option>
              <option value="purchase_order">Purchase Order</option>
              <option value="sales_order">Sales Order</option>
              <option value="quote">Quote</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">Reference Number</label>
            <input
              v-model="movementForm.referenceNumber"
              type="text"
              class="form-input"
              placeholder="e.g., PO-2024-001"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Notes</label>
            <textarea
              v-model="movementForm.notes"
              class="form-input"
              rows="3"
              placeholder="Additional notes about this movement..."
            ></textarea>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- Stock Adjustment Modal -->
    <BaseModal
      v-model="showAdjustmentModal"
      title="Adjust Stock Quantity"
      size="md"
      @confirm="saveAdjustment"
    >
      <div v-if="adjustingStock" class="modal-form">
        <div class="adjustment-info">
          <h4>{{ getProductName(adjustingStock.productId) }}</h4>
          <p class="sku">{{ getProductSKU(adjustingStock.productId) }}</p>
          <p class="location">Location: {{ adjustingStock.warehouseLocation }}</p>
          <p class="current-qty">Current Quantity: <strong>{{ adjustingStock.quantityOnHand }}</strong></p>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Adjustment Type *</label>
            <select v-model="adjustmentType" class="form-input" required>
              <option value="add">Add Stock (+)</option>
              <option value="remove">Remove Stock (-)</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Quantity *</label>
            <input
              v-model.number="adjustmentQuantity"
              type="number"
              class="form-input"
              min="1"
              required
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Reason *</label>
            <textarea
              v-model="adjustmentReason"
              class="form-input"
              rows="3"
              placeholder="Explain the reason for this adjustment..."
              required
            ></textarea>
          </div>
        </div>

        <div class="adjustment-preview">
          <div class="preview-row">
            <span>Current:</span>
            <span>{{ adjustingStock.quantityOnHand }}</span>
          </div>
          <div class="preview-row">
            <span>Adjustment:</span>
            <span :class="adjustmentType === 'add' ? 'success' : 'danger'">
              {{ adjustmentType === 'add' ? '+' : '-' }}{{ adjustmentQuantity }}
            </span>
          </div>
          <div class="preview-row preview-total">
            <span>New Quantity:</span>
            <span>{{ newQuantityAfterAdjustment }}</span>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- Item History Modal -->
    <BaseModal
      v-model="showHistoryModal"
      :title="`Stock Movement History`"
      size="xl"
      hide-confirm
      cancel-text="Close"
    >
      <div v-if="viewingHistoryStock" class="history-modal">
        <div class="history-header">
          <div class="history-product-info">
            <div class="product-name">{{ getProductName(viewingHistoryStock.productId) }}</div>
            <div class="product-sku">SKU: {{ getProductSKU(viewingHistoryStock.productId) }}</div>
            <div class="product-location">
              <BaseBadge variant="info" size="sm">{{ viewingHistoryStock.warehouseLocation }}</BaseBadge>
            </div>
          </div>
          <div class="history-stats">
            <div class="history-stat">
              <span class="stat-label">Current Stock</span>
              <span class="stat-value">{{ viewingHistoryStock.quantityOnHand }}</span>
            </div>
            <div class="history-stat">
              <span class="stat-label">Available</span>
              <span class="stat-value success">{{ viewingHistoryStock.quantityAvailable }}</span>
            </div>
            <div class="history-stat">
              <span class="stat-label">Reserved</span>
              <span class="stat-value warning">{{ viewingHistoryStock.quantityReserved }}</span>
            </div>
          </div>
        </div>

        <div v-if="historyMovements.length === 0" class="empty-history">
          No movement history found for this item
        </div>

        <div v-else class="history-timeline">
          <div
            v-for="movement in historyMovements"
            :key="movement.id"
            class="timeline-item"
          >
            <div class="timeline-marker" :class="getMovementTypeVariant(movement.movementType)"></div>
            <div class="timeline-content">
              <div class="timeline-header">
                <div class="timeline-type">
                  <BaseBadge :variant="getMovementTypeVariant(movement.movementType)" size="sm">
                    {{ formatMovementType(movement.movementType) }}
                  </BaseBadge>
                  <span class="timeline-number">{{ movement.movementNumber }}</span>
                </div>
                <div class="timeline-date">{{ formatDate(movement.movementDate) }}</div>
              </div>

              <div class="timeline-details">
                <div v-if="movement.movementType === 'transfer'" class="timeline-detail">
                  <span class="detail-label">Transfer:</span>
                  <span class="detail-value">
                    {{ movement.fromLocation }} → {{ movement.toLocation }}
                  </span>
                </div>
                <div v-else class="timeline-detail">
                  <span class="detail-label">Location:</span>
                  <span class="detail-value">{{ movement.warehouseLocation }}</span>
                </div>

                <div class="timeline-detail">
                  <span class="detail-label">Quantity:</span>
                  <span
                    class="detail-value"
                    :class="getQuantityClass(movement.movementType, movement.quantity)"
                  >
                    {{ getQuantityDisplay(movement.movementType, movement.quantity) }}
                  </span>
                </div>

                <div v-if="movement.referenceNumber" class="timeline-detail">
                  <span class="detail-label">Reference:</span>
                  <span class="detail-value">
                    {{ formatReferenceType(movement.referenceType) }} #{{ movement.referenceNumber }}
                  </span>
                </div>

                <div class="timeline-detail">
                  <span class="detail-label">Performed By:</span>
                  <span class="detail-value">{{ movement.performedBy }}</span>
                </div>

                <div v-if="movement.notes" class="timeline-notes">
                  <span class="detail-label">Notes:</span>
                  <span class="detail-value">{{ movement.notes }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useWarehouseStockStore } from '@/stores/warehouseStock'
import { useStockMovementsStore } from '@/stores/stockMovements'
import { useProductsStore } from '@/stores/products'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'

const warehouseStore = useWarehouseStockStore()
const movementsStore = useStockMovementsStore()
const productsStore = useProductsStore()
const { success, info } = useToast()

const activeTab = ref('stock')
const searchQuery = ref('')
const locationFilter = ref('all')
const stockStatusFilter = ref('all')
const movementTypeFilter = ref('all')
const movementLocationFilter = ref('all')

const showAddMovementModal = ref(false)
const showAdjustmentModal = ref(false)
const showHistoryModal = ref(false)
const adjustingStock = ref<any>(null)
const viewingHistoryStock = ref<any>(null)
const adjustmentType = ref('add')
const adjustmentQuantity = ref(1)
const adjustmentReason = ref('')

const movementForm = ref({
  movementType: 'receipt' as 'receipt' | 'issue' | 'transfer' | 'adjustment',
  productId: '',
  warehouseLocation: 'Riyadh Main' as any,
  fromLocation: '' as any,
  toLocation: '' as any,
  quantity: 1,
  referenceType: '' as any,
  referenceNumber: '',
  notes: ''
})

// Stats
const stats = computed(() => warehouseStore.getWarehouseStats())

// Filtered Stock
const filteredStock = computed(() => {
  let stock = warehouseStore.stock

  if (locationFilter.value !== 'all') {
    stock = stock.filter(s => s.warehouseLocation === locationFilter.value)
  }

  if (stockStatusFilter.value === 'low') {
    stock = stock.filter(s => s.quantityAvailable < 10)
  } else if (stockStatusFilter.value === 'available') {
    stock = stock.filter(s => s.quantityAvailable > 0)
  } else if (stockStatusFilter.value === 'reserved') {
    stock = stock.filter(s => s.quantityReserved > 0)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    stock = stock.filter(s => {
      const product = productsStore.getProductById(s.productId)
      if (!product) return false
      return product.sku.toLowerCase().includes(query) ||
             product.name.toLowerCase().includes(query)
    })
  }

  return stock
})

// Filtered Movements
const filteredMovements = computed(() => {
  let movements = movementsStore.recentMovements

  if (movementTypeFilter.value !== 'all') {
    movements = movements.filter(m => m.movementType === movementTypeFilter.value)
  }

  if (movementLocationFilter.value !== 'all') {
    movements = movements.filter(m =>
      m.warehouseLocation === movementLocationFilter.value ||
      m.fromLocation === movementLocationFilter.value ||
      m.toLocation === movementLocationFilter.value
    )
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    movements = movements.filter(m => {
      const product = productsStore.getProductById(m.productId)
      if (!product) return false
      return product.sku.toLowerCase().includes(query) ||
             product.name.toLowerCase().includes(query) ||
             m.movementNumber.toLowerCase().includes(query)
    })
  }

  return movements
})

const newQuantityAfterAdjustment = computed(() => {
  if (!adjustingStock.value) return 0
  const current = adjustingStock.value.quantityOnHand
  const change = adjustmentQuantity.value
  return adjustmentType.value === 'add' ? current + change : current - change
})

// History Movements for selected item
const historyMovements = computed(() => {
  if (!viewingHistoryStock.value) return []
  return movementsStore.getProductMovementHistory(viewingHistoryStock.value.productId)
    .filter(m => m.warehouseLocation === viewingHistoryStock.value.warehouseLocation ||
                 m.fromLocation === viewingHistoryStock.value.warehouseLocation ||
                 m.toLocation === viewingHistoryStock.value.warehouseLocation)
})

// Helper functions
function getProductName(productId: string): string {
  const product = productsStore.getProductById(productId)
  return product?.name || 'Unknown Product'
}

function getProductSKU(productId: string): string {
  const product = productsStore.getProductById(productId)
  return product?.sku || ''
}

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount)
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function getStockClass(quantity: number) {
  if (quantity === 0) return 'stock-out'
  if (quantity < 10) return 'stock-low'
  return 'stock-ok'
}

function getMovementTypeVariant(type: string) {
  switch (type) {
    case 'receipt': return 'success'
    case 'issue': return 'warning'
    case 'transfer': return 'info'
    case 'adjustment': return 'default'
    default: return 'default'
  }
}

function formatMovementType(type: string) {
  const types = {
    receipt: 'Receipt',
    issue: 'Issue',
    transfer: 'Transfer',
    adjustment: 'Adjustment'
  }
  return types[type as keyof typeof types] || type
}

function formatReferenceType(type?: string) {
  if (!type) return ''
  const types = {
    purchase_order: 'PO',
    sales_order: 'SO',
    quote: 'Quote',
    adjustment: 'Adjustment'
  }
  return types[type as keyof typeof types] || type
}

function getQuantityClass(type: string, quantity: number) {
  if (type === 'receipt') return 'qty-in'
  if (type === 'issue') return 'qty-out'
  return ''
}

function getQuantityDisplay(type: string, quantity: number) {
  if (type === 'receipt') return `+${quantity}`
  if (type === 'issue') return `-${quantity}`
  return quantity
}

// Actions
function adjustStock(stock: any) {
  adjustingStock.value = stock
  adjustmentType.value = 'add'
  adjustmentQuantity.value = 1
  adjustmentReason.value = ''
  showAdjustmentModal.value = true
}

function saveAdjustment() {
  if (!adjustingStock.value || !adjustmentReason.value) {
    info('Please fill in all required fields')
    return
  }

  const change = adjustmentType.value === 'add' ? adjustmentQuantity.value : -adjustmentQuantity.value

  warehouseStore.adjustQuantity(adjustingStock.value.id, change, adjustmentReason.value)

  // Record movement
  movementsStore.addMovement({
    productId: adjustingStock.value.productId,
    warehouseLocation: adjustingStock.value.warehouseLocation,
    movementType: 'adjustment',
    quantity: change,
    referenceType: 'adjustment',
    performedBy: 'USR-001',
    notes: adjustmentReason.value
  })

  success('Stock adjusted successfully')
  showAdjustmentModal.value = false
  adjustingStock.value = null
}

function saveMovement() {
  if (!movementForm.value.productId || !movementForm.value.quantity) {
    info('Please fill in all required fields')
    return
  }

  movementsStore.addMovement({
    productId: movementForm.value.productId,
    warehouseLocation: movementForm.value.warehouseLocation,
    movementType: movementForm.value.movementType,
    quantity: movementForm.value.quantity,
    referenceType: movementForm.value.referenceType || undefined,
    referenceNumber: movementForm.value.referenceNumber || undefined,
    fromLocation: movementForm.value.fromLocation || undefined,
    toLocation: movementForm.value.toLocation || undefined,
    performedBy: 'USR-001',
    notes: movementForm.value.notes || undefined
  })

  success('Stock movement recorded successfully')
  showAddMovementModal.value = false
  resetMovementForm()
}

function resetMovementForm() {
  movementForm.value = {
    movementType: 'receipt',
    productId: '',
    warehouseLocation: 'Riyadh Main',
    fromLocation: '' as any,
    toLocation: '' as any,
    quantity: 1,
    referenceType: '' as any,
    referenceNumber: '',
    notes: ''
  }
}

function viewHistory(stock: any) {
  viewingHistoryStock.value = stock
  showHistoryModal.value = true
}
</script>

<style scoped>
.inventory-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

/* Stats Bar */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
  margin-bottom: 0.5rem;
}

.stat-card {
  background: var(--color-surface);
  padding: 1.75rem;
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  transition: transform 0.2s, box-shadow 0.2s;
  min-width: 0;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  word-break: break-word;
  line-height: 1.2;
}

.stat-value.success {
  color: var(--color-success);
}

.stat-value.warning {
  color: var(--color-warning);
}

.stat-value.danger {
  color: var(--color-danger);
}

/* Header */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1rem;
  width: 100%;
  align-items: center;
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
  font-weight: 500;
  font-size: 1.05rem;
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

/* Tabs */
.tabs {
  display: flex;
  gap: 1rem;
  border-bottom: 2px solid var(--color-border);
  padding: 0 1rem;
}

.tab {
  padding: 1rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 3px solid transparent;
  color: var(--color-text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tab:hover {
  color: var(--color-text-primary);
}

.tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-icon {
  font-size: 1.1rem;
}

/* Filters */
.filters-bar {
  display: flex;
  gap: 2rem;
  align-items: center;
  padding: 1.5rem 2rem;
  background: var(--color-surface);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-select {
  padding: 0.75rem 1.25rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  cursor: pointer;
  min-width: 180px;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

/* Table */
.table-container {
  overflow-x: auto;
}

.inventory-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}

.inventory-table thead {
  background: var(--color-background);
  border-bottom: 2px solid var(--color-border);
}

.inventory-table th {
  padding: 1.25rem 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  font-size: 0.8rem;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.inventory-table th.text-right {
  text-align: right;
}

.inventory-table tbody tr {
  border-bottom: 1px solid var(--color-border);
  transition: background-color 0.2s;
}

.inventory-table tbody tr:hover {
  background: var(--color-surface-hover);
}

.inventory-table td {
  padding: 1rem;
}

.inventory-table td.text-right {
  text-align: right;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.product-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

.product-sku {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-family: 'Courier New', monospace;
}

.product-batch {
  font-size: 0.8rem;
  color: var(--color-info);
}

.product-info-sm {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.product-sku-sm {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.reserved-qty {
  color: var(--color-warning);
  font-weight: 600;
}

.stock-ok {
  color: var(--color-success);
  font-weight: 600;
}

.stock-low {
  color: var(--color-warning);
  font-weight: 600;
}

.stock-out {
  color: var(--color-danger);
  font-weight: 600;
}

.movement-number {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: var(--color-primary);
}

.transfer-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.from-location {
  color: var(--color-text-secondary);
}

.transfer-arrow {
  color: var(--color-primary);
  font-weight: bold;
}

.to-location {
  color: var(--color-text-primary);
  font-weight: 600;
}

.reference-type {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
}

.reference-number {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  color: var(--color-text-primary);
}

.qty-in {
  color: var(--color-success);
  font-weight: 600;
}

.qty-out {
  color: var(--color-danger);
  font-weight: 600;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-start;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

.empty-row {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-secondary);
  font-size: 1.05rem;
}

/* Modal Form */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-input {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

/* Adjustment Info */
.adjustment-info {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.adjustment-info h4 {
  margin: 0 0 0.5rem 0;
  color: var(--color-text-primary);
}

.adjustment-info .sku {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  font-family: 'Courier New', monospace;
  margin: 0.25rem 0;
}

.adjustment-info .location,
.adjustment-info .current-qty {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  margin: 0.25rem 0;
}

.adjustment-info .current-qty strong {
  color: var(--color-text-primary);
  font-size: 1.1rem;
}

.adjustment-preview {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1rem;
}

.preview-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--color-border);
}

.preview-row:last-child {
  border-bottom: none;
}

.preview-row.preview-total {
  font-weight: 700;
  font-size: 1.1rem;
  padding-top: 1rem;
  border-top: 2px solid var(--color-border);
}

.preview-row .success {
  color: var(--color-success);
}

.preview-row .danger {
  color: var(--color-danger);
}

/* History Modal */
.history-modal {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 2px solid var(--color-border);
}

.history-product-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-product-info .product-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.history-product-info .product-sku {
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  font-family: 'Courier New', monospace;
}

.history-stats {
  display: flex;
  gap: 2rem;
}

.history-stat {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 100px;
}

.history-stat .stat-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.history-stat .stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.history-stat .stat-value.success {
  color: var(--color-success);
}

.history-stat .stat-value.warning {
  color: var(--color-warning);
}

.empty-history {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-secondary);
  font-size: 1.05rem;
}

/* Timeline */
.history-timeline {
  display: flex;
  flex-direction: column;
  gap: 0;
  position: relative;
  padding-left: 2rem;
}

.timeline-item {
  display: flex;
  gap: 1.5rem;
  position: relative;
  padding-bottom: 2rem;
}

.timeline-item:last-child {
  padding-bottom: 0;
}

.timeline-item::before {
  content: '';
  position: absolute;
  left: 0.5rem;
  top: 1.5rem;
  bottom: -0.5rem;
  width: 2px;
  background: var(--color-border);
}

.timeline-item:last-child::before {
  display: none;
}

.timeline-marker {
  width: 1rem;
  height: 1rem;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 0.25rem;
  z-index: 1;
  position: relative;
  border: 3px solid var(--color-background);
}

.timeline-marker.success {
  background: var(--color-success);
}

.timeline-marker.warning {
  background: var(--color-warning);
}

.timeline-marker.info {
  background: var(--color-info);
}

.timeline-marker.default {
  background: var(--color-text-secondary);
}

.timeline-content {
  flex: 1;
  background: var(--color-surface);
  padding: 1.25rem;
  border-radius: 8px;
  box-shadow: var(--shadow-sm);
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.timeline-type {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.timeline-number {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  font-family: 'Courier New', monospace;
}

.timeline-date {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}

.timeline-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.timeline-detail {
  display: flex;
  gap: 0.75rem;
  font-size: 0.95rem;
}

.timeline-detail .detail-label {
  font-weight: 500;
  color: var(--color-text-secondary);
  min-width: 100px;
}

.timeline-detail .detail-value {
  color: var(--color-text-primary);
}

.timeline-detail .detail-value.qty-in {
  color: var(--color-success);
  font-weight: 600;
}

.timeline-detail .detail-value.qty-out {
  color: var(--color-danger);
  font-weight: 600;
}

.timeline-notes {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
  margin-top: 0.25rem;
}

.timeline-notes .detail-label {
  font-weight: 600;
  color: var(--color-text-primary);
}

.timeline-notes .detail-value {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
}
</style>
