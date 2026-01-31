<template>
  <div class="pricebook-builder-modern">
    <!-- Progress Steps -->
    <div class="progress-bar">
      <div
        v-for="(step, index) in steps"
        :key="step.id"
        class="step-item"
        :class="{ active: currentStep === index + 1, completed: currentStep > index + 1 }"
      >
        <div class="step-badge">{{ index + 1 }}</div>
        <span class="step-label">{{ step.label }}</span>
      </div>
    </div>

    <!-- Step Content -->
    <div class="steps-container">
      <!-- Step 1: Contract Information -->
      <div v-show="currentStep === 1" class="step-content">
        <div class="modern-card">
          <div class="card-header">
            <h3 class="card-title">Contract Information</h3>
            <p class="card-subtitle">Define the basic details for this price book contract</p>
          </div>
          <div class="card-body">
            <div class="form-grid">
              <div class="form-group full-width">
                <label class="form-label">Contract Name *</label>
                <input
                  v-model="priceBookData.name"
                  type="text"
                  class="form-input"
                  placeholder="e.g., Standard Pricing 2024, VIP Customer Contract"
                  required
                />
              </div>

              <div class="form-group full-width">
                <label class="form-label">Description</label>
                <textarea
                  v-model="priceBookData.description"
                  class="form-input"
                  rows="3"
                  placeholder="Describe the purpose and scope of this contract..."
                ></textarea>
              </div>

              <div class="form-group">
                <label class="form-label">Contract Type *</label>
                <select v-model="priceBookData.type" class="form-input" required>
                  <option value="standard">Standard</option>
                  <option value="volume">Volume Discount</option>
                  <option value="contract">Contract Pricing</option>
                  <option value="promotional">Promotional</option>
                  <option value="customer_specific">Customer Specific</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">Currency *</label>
                <select v-model="priceBookData.currency" class="form-input" required>
                  <option value="SAR">SAR</option>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="AED">AED</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">Valid From *</label>
                <input
                  v-model="priceBookData.validFrom"
                  type="date"
                  class="form-input"
                  required
                />
              </div>

              <div class="form-group">
                <label class="form-label">Valid Until</label>
                <input
                  v-model="priceBookData.validUntil"
                  type="date"
                  class="form-input"
                  placeholder="Leave blank for no expiry"
                />
                <small class="form-hint">Leave blank for ongoing contracts</small>
              </div>

              <div class="form-group">
                <label class="form-label">Status *</label>
                <select v-model="priceBookData.status" class="form-input" required>
                  <option value="draft">Draft</option>
                  <option value="pending_approval">Pending Approval</option>
                  <option value="active">Active</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 2: Customer Assignment -->
      <div v-show="currentStep === 2" class="step-content">
        <div class="modern-card">
          <div class="card-header">
            <h3 class="card-title">Customer Assignment</h3>
            <p class="card-subtitle">Choose which customers this contract applies to</p>
          </div>
          <div class="card-body">
            <div class="assignment-options">
              <label class="radio-card">
                <input
                  type="radio"
                  :value="true"
                  v-model="assignToAllCustomers"
                  name="assignment"
                />
                <div class="radio-card-content">
                  <div class="radio-icon">🌍</div>
                  <div class="radio-text">
                    <div class="radio-title">All Customers</div>
                    <div class="radio-subtitle">This contract will be available to all customers</div>
                  </div>
                </div>
              </label>

              <label class="radio-card">
                <input
                  type="radio"
                  :value="false"
                  v-model="assignToAllCustomers"
                  name="assignment"
                />
                <div class="radio-card-content">
                  <div class="radio-icon">👥</div>
                  <div class="radio-text">
                    <div class="radio-title">Specific Customers</div>
                    <div class="radio-subtitle">Select which customers can use this contract</div>
                  </div>
                </div>
              </label>
            </div>

            <div v-if="!assignToAllCustomers" class="customer-selection">
              <label class="form-label">Select Customers *</label>
              <div class="customer-search">
                <input
                  v-model="customerSearch"
                  type="text"
                  class="form-input"
                  placeholder="🔍 Search customers by name..."
                  @focus="showCustomerDropdown = true"
                  @blur="hideCustomerDropdown"
                  autocomplete="off"
                />
                <div v-if="showCustomerDropdown" class="dropdown-list">
                  <div v-if="filteredCustomers.length === 0" class="dropdown-empty">
                    No customers found
                  </div>
                  <div
                    v-for="customer in filteredCustomers"
                    :key="customer.id"
                    class="dropdown-item"
                    @mousedown.prevent="toggleCustomer(customer)"
                  >
                    <input
                      type="checkbox"
                      :checked="priceBookData.customerIds.includes(customer.id)"
                      class="dropdown-checkbox"
                      readonly
                    />
                    <div class="customer-info">
                      <div class="customer-name">{{ customer.name }}</div>
                      <div class="customer-email">{{ customer.email }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="selectedCustomers.length > 0" class="selected-customers">
                <div
                  v-for="customer in selectedCustomers"
                  :key="customer.id"
                  class="customer-chip"
                >
                  {{ customer.name }}
                  <button type="button" @click="removeCustomer(customer.id)" class="chip-remove">×</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 3: Product Pricing -->
      <div v-show="currentStep === 3" class="step-content">
        <div class="modern-card">
          <div class="card-header">
            <h3 class="card-title">Product Pricing</h3>
            <p class="card-subtitle">Define custom prices for products in this contract</p>
          </div>
          <div class="card-body">
            <button type="button" class="btn btn-add-item" @click="addNewItemCard">
              ➕ Add Item
            </button>

            <!-- Item Cards List -->
            <div v-if="priceBookData.priceBookEntries.length > 0" class="items-list">
              <div v-for="(entry, index) in priceBookData.priceBookEntries" :key="entry.tempId || index" class="item-card">
                <div class="item-number">{{ index + 1 }}</div>

                <div class="item-body">
                  <!-- Item Type Selector -->
                  <div class="type-row">
                    <div class="type-tabs">
                      <label class="type-tab" :class="{ active: entry.itemType === 'product' }">
                        <input type="radio" v-model="entry.itemType" value="product" @change="onItemTypeChange(entry)" />
                        <span>📦 Product</span>
                      </label>
                      <label class="type-tab" :class="{ active: entry.itemType === 'service' }">
                        <input type="radio" v-model="entry.itemType" value="service" @change="onItemTypeChange(entry)" />
                        <span>👷 Service</span>
                      </label>
                      <label class="type-tab" :class="{ active: entry.itemType === 'recurring' }">
                        <input type="radio" v-model="entry.itemType" value="recurring" @change="onItemTypeChange(entry)" />
                        <span>🔄 Recurring</span>
                      </label>
                    </div>

                    <!-- Autocomplete Product Search -->
                    <div class="autocomplete-container">
                      <input
                        v-model="entry.searchQuery"
                        type="text"
                        class="item-select"
                        :placeholder="`🔍 Search ${entry.itemType === 'recurring' ? 'recurring service' : entry.itemType === 'service' ? 'labor/service' : 'product'} by name or SKU...`"
                        @input="handleItemSearch(entry)"
                        @focus="handleItemSearchFocus(entry)"
                        @blur="handleItemSearchBlur(entry)"
                        autocomplete="off"
                      />
                      <div v-if="entry.showSuggestions && getItemSuggestions(entry).length > 0" class="autocomplete-dropdown">
                        <div class="autocomplete-header">
                          <span class="suggestion-count">{{ getItemSuggestions(entry).length }} results</span>
                        </div>
                        <div
                          v-for="suggestion in getItemSuggestions(entry)"
                          :key="suggestion.id"
                          class="autocomplete-item"
                          @mousedown.prevent="selectItemProduct(entry, suggestion)"
                        >
                          <div class="suggestion-main">
                            <span class="suggestion-icon">{{ suggestion.icon }}</span>
                            <div class="suggestion-details">
                              <div class="suggestion-title">
                                <span v-if="suggestion.sku" class="suggestion-sku">{{ suggestion.sku }}</span>
                                <span class="suggestion-name">{{ suggestion.name }}</span>
                              </div>
                              <div v-if="suggestion.manufacturer" class="suggestion-manufacturer">
                                🏭 {{ suggestion.manufacturer }}
                              </div>
                              <div v-if="entry.itemType === 'product'" class="suggestion-inventory">
                                <span v-if="suggestion.hasStock" class="stock-badge in-stock">
                                  ✓ {{ suggestion.stockAvailable }} in stock
                                </span>
                                <span v-else class="stock-badge out-stock">
                                  ⏳ Lead time: {{ suggestion.leadTime }} days
                                </span>
                              </div>
                              <div v-if="suggestion.department" class="suggestion-department">
                                {{ suggestion.department }}
                              </div>
                            </div>
                          </div>
                          <div class="suggestion-meta">
                            <span class="suggestion-price">{{ suggestion.priceLabel }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Product Details Grid -->
                  <div class="details-grid">
                    <div class="detail-box">
                      <label>Standard Price</label>
                      <div class="detail-display">{{ formatCurrency(entry.standardPrice || 0) }}</div>
                    </div>
                    <div class="detail-box">
                      <label>Custom Price</label>
                      <input
                        v-model.number="entry.customPrice"
                        type="number"
                        step="0.01"
                        min="0"
                        class="detail-input"
                        @input="calculateEntryDiscount(entry)"
                      />
                    </div>
                    <div class="detail-box">
                      <label>Discount %</label>
                      <input
                        v-model.number="entry.discountPercent"
                        type="number"
                        step="0.01"
                        min="0"
                        max="100"
                        class="detail-input"
                        @input="calculateEntryPrice(entry)"
                      />
                    </div>
                    <div class="detail-box">
                      <label>
                        <span v-if="entry.itemType === 'recurring'">Min. Months</span>
                        <span v-else-if="entry.itemType === 'service'">Min. Hours</span>
                        <span v-else>Min. Qty</span>
                      </label>
                      <input
                        v-model.number="entry.minimumQuantity"
                        type="number"
                        min="1"
                        class="detail-input"
                      />
                    </div>
                    <div class="detail-box">
                      <label>Change</label>
                      <div class="change-display" :class="getEntryChangeClass(entry)">
                        {{ getEntryChangePercent(entry) }}
                      </div>
                    </div>
                  </div>
                </div>

                <button class="remove-btn" @click="removeEntry(index)" title="Remove item">🗑️</button>
              </div>
            </div>

            <div v-else class="empty-state">
              <p>No items added yet. Click "Add Item" above to start building your price book.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 4: Review & Submit -->
      <div v-show="currentStep === 4" class="step-content">
        <div class="modern-card">
          <div class="card-header">
            <h3 class="card-title">Review Contract</h3>
            <p class="card-subtitle">Review and finalize your price book contract</p>
          </div>
          <div class="card-body">
            <div class="review-section">
              <h4 class="review-title">Contract Information</h4>
              <div class="review-grid">
                <div class="review-item">
                  <span class="review-label">Name:</span>
                  <span class="review-value">{{ priceBookData.name }}</span>
                </div>
                <div class="review-item">
                  <span class="review-label">Type:</span>
                  <span class="review-value">{{ formatType(priceBookData.type) }}</span>
                </div>
                <div class="review-item">
                  <span class="review-label">Currency:</span>
                  <span class="review-value">{{ priceBookData.currency }}</span>
                </div>
                <div class="review-item">
                  <span class="review-label">Valid From:</span>
                  <span class="review-value">{{ priceBookData.validFrom }}</span>
                </div>
                <div class="review-item">
                  <span class="review-label">Valid Until:</span>
                  <span class="review-value">{{ priceBookData.validUntil || 'No Expiry' }}</span>
                </div>
                <div class="review-item">
                  <span class="review-label">Status:</span>
                  <span class="review-value">{{ formatStatus(priceBookData.status) }}</span>
                </div>
              </div>
            </div>

            <div class="review-section">
              <h4 class="review-title">Customer Assignment</h4>
              <div v-if="priceBookData.customerIds.length === 0" class="review-info">
                🌍 Available to all customers
              </div>
              <div v-else class="review-customers">
                <div v-for="customerId in priceBookData.customerIds" :key="customerId" class="customer-chip">
                  {{ getCustomerName(customerId) }}
                </div>
              </div>
            </div>

            <div class="review-section">
              <h4 class="review-title">Products & Pricing</h4>
              <div class="review-summary">
                <div class="summary-stat">
                  <div class="stat-value">{{ priceBookData.priceBookEntries.length }}</div>
                  <div class="stat-label">Total Products</div>
                </div>
                <div class="summary-stat">
                  <div class="stat-value">{{ averageDiscount.toFixed(1) }}%</div>
                  <div class="stat-label">Avg. Discount</div>
                </div>
                <div class="summary-stat">
                  <div class="stat-value">{{ formatCurrency(totalStandardValue) }}</div>
                  <div class="stat-label">Standard Value</div>
                </div>
                <div class="summary-stat">
                  <div class="stat-value">{{ formatCurrency(totalContractValue) }}</div>
                  <div class="stat-label">Contract Value</div>
                </div>
              </div>
            </div>

            <!-- Approval Warning -->
            <div v-if="requiresApproval" class="approval-warning">
              <div class="warning-icon">⚠️</div>
              <div class="warning-content">
                <div class="warning-title">Approval Required</div>
                <div class="warning-message">{{ approvalReason }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Navigation Buttons -->
    <div class="navigation-buttons">
      <button
        v-if="currentStep > 1"
        type="button"
        class="btn btn-secondary"
        @click="previousStep"
      >
        ← Previous
      </button>
      <div class="spacer"></div>
      <button
        v-if="currentStep < 4"
        type="button"
        class="btn btn-primary"
        @click="nextStep"
        :disabled="!canProceed"
      >
        Next →
      </button>
      <button
        v-if="currentStep === 4"
        type="button"
        class="btn btn-success"
        @click="savePriceBook"
      >
        💾 Save Contract
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useCustomersStore } from '@/stores/customers'
import { useProductsStore } from '@/stores/products'
import { useRecurringServicesStore } from '@/stores/recurringServices'
import { useLaborPositionsStore } from '@/stores/laborPositions'
import { usePriceBooksStore } from '@/stores/priceBooks'
import { useWarehouseStockStore } from '@/stores/warehouseStock'
import { useToast } from '@/composables/useToast'
import type { PriceBook, PriceBookEntry, PriceBookStatus, PriceBookType } from '@/types'

const customersStore = useCustomersStore()
const productsStore = useProductsStore()
const recurringStore = useRecurringServicesStore()
const laborStore = useLaborPositionsStore()
const priceBooksStore = usePriceBooksStore()
const warehouseStockStore = useWarehouseStockStore()
const { success, error: showError } = useToast()

const steps = [
  { id: 1, label: 'Contract Info' },
  { id: 2, label: 'Customers' },
  { id: 3, label: 'Pricing' },
  { id: 4, label: 'Review' }
]

const currentStep = ref(1)
const assignToAllCustomers = ref(true)
const customerSearch = ref('')
const showCustomerDropdown = ref(false)

const priceBookData = ref({
  name: '',
  description: '',
  type: 'standard' as PriceBookType,
  currency: 'SAR',
  validFrom: new Date().toISOString().split('T')[0],
  validUntil: '',
  status: 'draft' as PriceBookStatus,
  customerIds: [] as string[],
  priceBookEntries: [] as PriceBookEntry[]
})

const newEntry = ref({
  category: '' as 'product' | 'service' | 'recurring' | '',
  productId: '',
  customPrice: 0,
  discountPercent: 0,
  minimumQuantity: 1
})

const productSearchQuery = ref('')
const showProductSuggestions = ref(false)

// Computed
const filteredCustomers = computed(() => {
  if (!customerSearch.value) return customersStore.customers
  const search = customerSearch.value.toLowerCase()
  return customersStore.customers.filter(c =>
    c.name.toLowerCase().includes(search)
  )
})

const selectedCustomers = computed(() => {
  return customersStore.customers.filter(c =>
    priceBookData.value.customerIds.includes(c.id)
  )
})

const availableProducts = computed(() => {
  const addedProductIds = priceBookData.value.priceBookEntries.map(e => e.productId)
  return productsStore.products.filter(p => !addedProductIds.includes(p.id))
})

// All available items (products, services, recurring)
const availableItems = computed(() => {
  const addedIds = priceBookData.value.priceBookEntries.map(e => e.productId)
  const items: any[] = []

  // Add products
  productsStore.products
    .filter(p => !addedIds.includes(p.id))
    .forEach(p => {
      // Get inventory information
      const totalAvailable = warehouseStockStore.getTotalAvailableForProduct(p.id)
      const stockInfo = warehouseStockStore.getStockByProductId(p.id)
      const hasStock = totalAvailable > 0

      items.push({
        id: p.id,
        type: 'product',
        name: p.name,
        sku: p.sku,
        sellingPrice: p.sellingPrice,
        landedCostSAR: p.landedCostSAR,
        unitOfMeasure: p.unitOfMeasure || 'Unit',
        icon: '📦',
        priceLabel: formatCurrency(p.sellingPrice),
        manufacturer: p.manufacturer,
        stockAvailable: totalAvailable,
        leadTime: p.leadTimeDays,
        hasStock: hasStock,
        stockLocations: stockInfo.length
      })
    })

  // Add labor/service positions
  laborStore.getActivePositions()
    .filter(pos => !addedIds.includes(pos.id))
    .forEach(pos => {
      items.push({
        id: pos.id,
        type: 'service',
        name: pos.name,
        sku: pos.id,
        sellingPrice: pos.sellingRatePerHour,
        landedCostSAR: pos.costPerHour,
        unitOfMeasure: 'Hour',
        department: pos.department,
        icon: '👷',
        priceLabel: `${formatCurrency(pos.sellingRatePerHour)}/hr`
      })
    })

  // Add recurring services
  recurringStore.getActiveRecurringServices()
    .filter(s => !addedIds.includes(s.id))
    .forEach(s => {
      items.push({
        id: s.id,
        type: 'recurring',
        name: s.name,
        sku: s.id,
        sellingPrice: s.monthlyPrice,
        landedCostSAR: s.monthlyCost,
        unitOfMeasure: 'Month',
        icon: '🔄',
        priceLabel: `${formatCurrency(s.monthlyPrice)}/mo`
      })
    })

  return items
})

// Filtered items by category
const filteredAvailableItems = computed(() => {
  if (!newEntry.value.category) return []
  return availableItems.value.filter(item => item.type === newEntry.value.category)
})

// Product suggestions for autocomplete
const productSuggestions = computed(() => {
  if (!productSearchQuery.value || productSearchQuery.value.length < 2 || !newEntry.value.category) {
    return []
  }

  const query = productSearchQuery.value.toLowerCase()
  const items = filteredAvailableItems.value

  return items.filter(item =>
    item.name.toLowerCase().includes(query) ||
    (item.sku && item.sku.toLowerCase().includes(query)) ||
    (item.department && item.department.toLowerCase().includes(query))
  ).slice(0, 10)
})

const selectedProduct = computed(() => {
  if (!newEntry.value.productId) return null

  // Check all stores
  const product = productsStore.getProductById(newEntry.value.productId)
  if (product) {
    return {
      ...product,
      unitOfMeasure: product.unitOfMeasure || 'Unit'
    }
  }

  const labor = laborStore.getPositionById(newEntry.value.productId)
  if (labor) {
    return {
      id: labor.id,
      name: labor.name,
      sellingPrice: labor.sellingRatePerHour,
      landedCostSAR: labor.costPerHour,
      unitOfMeasure: 'Hour'
    }
  }

  const recurring = recurringStore.getRecurringServiceById(newEntry.value.productId)
  if (recurring) {
    return {
      id: recurring.id,
      name: recurring.name,
      sellingPrice: recurring.monthlyPrice,
      landedCostSAR: recurring.monthlyCost,
      unitOfMeasure: 'Month'
    }
  }

  return null
})

const priceChangePercent = computed(() => {
  if (!selectedProduct.value || !newEntry.value.customPrice) return 0
  const standard = selectedProduct.value.sellingPrice
  return ((newEntry.value.customPrice - standard) / standard) * 100
})

const priceChangeClass = computed(() => {
  if (priceChangePercent.value < 0) return 'text-success'
  if (priceChangePercent.value > 0) return 'text-danger'
  return ''
})

const averageDiscount = computed(() => {
  if (priceBookData.value.priceBookEntries.length === 0) return 0
  const totalDiscount = priceBookData.value.priceBookEntries.reduce((sum, e) => sum + (e.discountPercent || 0), 0)
  return totalDiscount / priceBookData.value.priceBookEntries.length
})

const totalStandardValue = computed(() => {
  return priceBookData.value.priceBookEntries.reduce((sum, entry) => {
    const product = productsStore.getProductById(entry.productId)
    return sum + (product?.sellingPrice || 0)
  }, 0)
})

const totalContractValue = computed(() => {
  return priceBookData.value.priceBookEntries.reduce((sum, entry) => {
    return sum + entry.customPrice
  }, 0)
})

const requiresApproval = computed(() => {
  // Requires approval if average discount > 20% or any product has > 30% discount
  if (averageDiscount.value > 20) return true
  return priceBookData.value.priceBookEntries.some(e => (e.discountPercent || 0) > 30)
})

const approvalReason = computed(() => {
  if (averageDiscount.value > 20) {
    return `High average discount: ${averageDiscount.value.toFixed(1)}% (threshold: 20%)`
  }
  const highDiscounts = priceBookData.value.priceBookEntries.filter(e => (e.discountPercent || 0) > 30)
  if (highDiscounts.length > 0) {
    return `${highDiscounts.length} product(s) with discount > 30%`
  }
  return ''
})

const canProceed = computed(() => {
  switch (currentStep.value) {
    case 1:
      return priceBookData.value.name && priceBookData.value.validFrom
    case 2:
      return assignToAllCustomers.value || priceBookData.value.customerIds.length > 0
    case 3:
      return priceBookData.value.priceBookEntries.length > 0
    default:
      return true
  }
})

// Watch
watch(assignToAllCustomers, (newVal) => {
  if (newVal) {
    priceBookData.value.customerIds = []
  }
})

watch(() => newEntry.value.customPrice, (newPrice) => {
  if (selectedProduct.value && newPrice) {
    const standard = selectedProduct.value.sellingPrice
    newEntry.value.discountPercent = ((standard - newPrice) / standard) * 100
  }
})

watch(() => newEntry.value.discountPercent, (discount) => {
  if (selectedProduct.value && discount !== undefined) {
    const standard = selectedProduct.value.sellingPrice
    newEntry.value.customPrice = standard * (1 - discount / 100)
  }
})

// Methods
function nextStep() {
  if (canProceed.value && currentStep.value < 4) {
    currentStep.value++
  }
}

function previousStep() {
  if (currentStep.value > 1) {
    currentStep.value--
  }
}

function hideCustomerDropdown() {
  setTimeout(() => {
    showCustomerDropdown.value = false
  }, 200)
}

function toggleCustomer(customer: any) {
  const index = priceBookData.value.customerIds.indexOf(customer.id)
  if (index > -1) {
    priceBookData.value.customerIds.splice(index, 1)
  } else {
    priceBookData.value.customerIds.push(customer.id)
  }
}

function removeCustomer(customerId: string) {
  const index = priceBookData.value.customerIds.indexOf(customerId)
  if (index > -1) {
    priceBookData.value.customerIds.splice(index, 1)
  }
}

function onCategoryChange() {
  newEntry.value.productId = ''
  newEntry.value.customPrice = 0
  newEntry.value.discountPercent = 0
  productSearchQuery.value = ''
  showProductSuggestions.value = false
}

function handleProductSearch() {
  if (productSearchQuery.value.length >= 2) {
    showProductSuggestions.value = true
  } else {
    showProductSuggestions.value = false
  }
  // Clear selection if user types after selecting
  if (newEntry.value.productId) {
    newEntry.value.productId = ''
    newEntry.value.customPrice = 0
  }
}

function handleProductSearchFocus() {
  if (productSearchQuery.value.length >= 2) {
    showProductSuggestions.value = true
  }
}

function handleProductSearchBlur() {
  setTimeout(() => {
    showProductSuggestions.value = false
  }, 200)
}

function selectProductItem(item: any) {
  productSearchQuery.value = `${item.sku ? item.sku + ' - ' : ''}${item.name}`
  newEntry.value.productId = item.id
  showProductSuggestions.value = false
  onProductSelect()
}

function onProductSelect() {
  if (selectedProduct.value) {
    newEntry.value.customPrice = selectedProduct.value.sellingPrice
    newEntry.value.discountPercent = 0
  }
}

function getCategoryLabel(): string {
  switch (newEntry.value.category) {
    case 'product':
      return 'Product'
    case 'service':
      return 'Service/Labor'
    case 'recurring':
      return 'Recurring Service'
    default:
      return 'Item'
  }
}

function addProductEntry() {
  if (!newEntry.value.productId || !newEntry.value.customPrice) return

  if (!selectedProduct.value) return

  priceBookData.value.priceBookEntries.push({
    productId: newEntry.value.productId,
    customPrice: newEntry.value.customPrice,
    discountPercent: newEntry.value.discountPercent,
    minimumQuantity: newEntry.value.minimumQuantity
  })

  success(`${selectedProduct.value.name} added to contract`)

  // Reset form
  newEntry.value = {
    category: newEntry.value.category, // Keep category selected
    productId: '',
    customPrice: 0,
    discountPercent: 0,
    minimumQuantity: 1
  }
  productSearchQuery.value = ''
  showProductSuggestions.value = false
}

function removeEntry(index: number) {
  priceBookData.value.priceBookEntries.splice(index, 1)
}

function calculateEntryDiscount(entry: PriceBookEntry) {
  const standardPrice = getProductStandardPrice(entry.productId)
  if (standardPrice > 0) {
    entry.discountPercent = ((standardPrice - entry.customPrice) / standardPrice) * 100
  }
}

function calculateEntryPrice(entry: PriceBookEntry) {
  const standardPrice = getProductStandardPrice(entry.productId)
  if (standardPrice > 0) {
    entry.customPrice = standardPrice * (1 - (entry.discountPercent || 0) / 100)
  }
}

function getEntryChangePercent(entry: PriceBookEntry): string {
  const standardPrice = getProductStandardPrice(entry.productId)
  if (standardPrice === 0) return '0%'
  const change = ((entry.customPrice - standardPrice) / standardPrice) * 100
  return `${change > 0 ? '+' : ''}${change.toFixed(1)}%`
}

function getEntryChangeClass(entry: PriceBookEntry): string {
  const standardPrice = getProductStandardPrice(entry.productId)
  if (standardPrice === 0) return ''
  if (entry.customPrice < standardPrice) return 'text-success'
  if (entry.customPrice > standardPrice) return 'text-danger'
  return ''
}

function getProductName(productId: string): string {
  const product = productsStore.getProductById(productId)
  if (product) return product.name

  const labor = laborStore.getPositionById(productId)
  if (labor) return `${labor.name} 👷`

  const recurring = recurringStore.getRecurringServiceById(productId)
  if (recurring) return `${recurring.name} 🔄`

  return 'Unknown Product'
}

function getProductStandardPrice(productId: string): number {
  const product = productsStore.getProductById(productId)
  if (product) return product.sellingPrice

  const labor = laborStore.getPositionById(productId)
  if (labor) return labor.sellingRatePerHour

  const recurring = recurringStore.getRecurringServiceById(productId)
  if (recurring) return recurring.monthlyPrice

  return 0
}

function getProductUnitLabel(productId: string): string {
  const product = productsStore.getProductById(productId)
  if (product) return ''

  const labor = laborStore.getPositionById(productId)
  if (labor) return '/hr'

  const recurring = recurringStore.getRecurringServiceById(productId)
  if (recurring) return '/mo'

  return ''
}

function getCustomerName(customerId: string): string {
  const customer = customersStore.customers.find(c => c.id === customerId)
  return customer?.name || 'Unknown Customer'
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: priceBookData.value.currency || 'SAR',
    minimumFractionDigits: 2
  }).format(value)
}

function formatType(type: string): string {
  const types: Record<string, string> = {
    standard: 'Standard',
    volume: 'Volume Discount',
    contract: 'Contract Pricing',
    promotional: 'Promotional',
    customer_specific: 'Customer Specific'
  }
  return types[type] || type
}

function formatStatus(status: string): string {
  const statuses: Record<string, string> = {
    draft: 'Draft',
    pending_approval: 'Pending Approval',
    active: 'Active',
    inactive: 'Inactive',
    expired: 'Expired'
  }
  return statuses[status] || status
}

// New Item Card Methods
function addNewItemCard() {
  const newItem: any = {
    tempId: `temp-${Date.now()}`,
    itemType: 'product',
    productId: '',
    searchQuery: '',
    showSuggestions: false,
    standardPrice: 0,
    customPrice: 0,
    discountPercent: 0,
    minimumQuantity: 1
  }
  priceBookData.value.priceBookEntries.push(newItem)
}

function onItemTypeChange(item: any) {
  // Reset search when item type changes
  item.productId = ''
  item.searchQuery = ''
  item.showSuggestions = false
  item.standardPrice = 0
  item.customPrice = 0
  item.discountPercent = 0
}

function handleItemSearch(item: any) {
  if (item.searchQuery && item.searchQuery.length >= 2) {
    item.showSuggestions = true
  } else {
    item.showSuggestions = false
  }
}

function handleItemSearchFocus(item: any) {
  if (item.searchQuery && item.searchQuery.length >= 2) {
    item.showSuggestions = true
  }
}

function handleItemSearchBlur(item: any) {
  setTimeout(() => {
    item.showSuggestions = false
  }, 200)
}

function getItemSuggestions(item: any) {
  if (!item.searchQuery || item.searchQuery.length < 2) return []

  const query = item.searchQuery.toLowerCase()
  const addedIds = priceBookData.value.priceBookEntries
    .filter((e: any) => e.productId && e !== item)
    .map((e: any) => e.productId)

  // Filter by item type
  let items = availableItems.value.filter((i: any) => i.type === item.itemType)

  // Exclude already added items
  items = items.filter((i: any) => !addedIds.includes(i.id))

  // Search by name, SKU, or department
  return items.filter((i: any) =>
    i.name.toLowerCase().includes(query) ||
    (i.sku && i.sku.toLowerCase().includes(query)) ||
    (i.department && i.department.toLowerCase().includes(query))
  ).slice(0, 10)
}

function selectItemProduct(item: any, suggestion: any) {
  item.productId = suggestion.id
  item.searchQuery = `${suggestion.sku ? suggestion.sku + ' - ' : ''}${suggestion.name}`
  item.showSuggestions = false
  item.standardPrice = suggestion.sellingPrice
  item.customPrice = suggestion.sellingPrice
  item.discountPercent = 0

  success(`${suggestion.name} selected`)
}

function savePriceBook() {
  try {
    priceBooksStore.addPriceBook(priceBookData.value as any)
    success('Contract saved successfully!')
    // Emit event or reset form
  } catch (err) {
    showError('Failed to save contract')
  }
}
</script>

<style scoped>
.pricebook-builder-modern {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem;
}

/* Progress Bar */
.progress-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 16px;
  position: relative;
}

.progress-bar::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 10%;
  right: 10%;
  height: 3px;
  background: var(--color-border);
  z-index: 0;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  position: relative;
  z-index: 1;
}

.step-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--color-surface);
  border: 3px solid var(--color-border);
  color: var(--color-text-secondary);
  font-size: 1rem;
  font-weight: 700;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.step-item.active .step-badge {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-color: transparent;
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transform: scale(1.1);
}

.step-item.completed .step-badge {
  background: #10b981;
  border-color: transparent;
  color: white;
}

.step-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  transition: color 0.2s;
}

.step-item.active .step-label {
  color: var(--color-primary);
}

/* Modern Card */
.modern-card {
  background: var(--color-surface);
  border-radius: 20px;
  overflow: visible;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.03), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-bottom: 1.5rem;
}

.card-header {
  padding: 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.03), rgba(139, 92, 246, 0.03));
  border-bottom: 1px solid var(--color-border);
}

.card-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.card-subtitle {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin: 0;
}

.card-body {
  padding: 1.5rem;
}

/* Form Styles */
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.full-width {
  grid-column: span 2;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.form-input {
  padding: 0.75rem;
  border: 2px solid var(--color-border);
  border-radius: 12px;
  background: var(--color-background);
  color: var(--color-text-primary);
  font-size: 0.95rem;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  margin-top: -0.25rem;
}

.text-success {
  color: #10b981;
}

.text-danger {
  color: #ef4444;
}

/* Radio Cards */
.assignment-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.radio-card {
  position: relative;
  cursor: pointer;
  display: block;
}

.radio-card input[type="radio"] {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.radio-card-content {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.radio-card input:checked + .radio-card-content {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.05);
}

.radio-icon {
  font-size: 2rem;
}

.radio-title {
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 0.25rem;
}

.radio-subtitle {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* Customer Selection */
.customer-search {
  position: relative;
}

.dropdown-list {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  max-height: 300px;
  overflow-y: auto;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  margin-top: 0.5rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.dropdown-item {
  padding: 0.875rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  transition: background 0.2s;
  border-bottom: 1px solid var(--color-border);
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-item:hover {
  background: rgba(59, 130, 246, 0.05);
}

.dropdown-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  flex-shrink: 0;
}

.customer-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.customer-name {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
}

.customer-email {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.dropdown-empty {
  padding: 1rem;
  text-align: center;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.selected-customers {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-top: 1rem;
}

.customer-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.chip-remove {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  font-size: 1.25rem;
  cursor: pointer;
  padding: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.chip-remove:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #ef4444;
}

/* Entries List */
.add-product-section {
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: rgba(59, 130, 246, 0.03);
  border-radius: 12px;
  border: 1px dashed rgba(59, 130, 246, 0.2);
}

.entries-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.entries-header {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 0.8fr 0.8fr 1fr 80px;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: var(--color-background);
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.entry-row {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 0.8fr 0.8fr 1fr 80px;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  align-items: center;
  transition: all 0.2s;
}

.entry-row:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.entry-product strong {
  color: var(--color-text-primary);
  font-weight: 600;
}

.entry-value {
  color: var(--color-text-primary);
  font-weight: 500;
}

.entry-change {
  font-weight: 600;
  font-size: 0.875rem;
}

.inline-input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: 0.9rem;
}

.entry-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: rgba(0, 0, 0, 0.05);
}

.btn-icon.btn-danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

/* Review Section */
.review-section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--color-border);
}

.review-section:last-child {
  border-bottom: none;
}

.review-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 1rem;
}

.review-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.review-item {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem;
  background: var(--color-background);
  border-radius: 8px;
}

.review-label {
  font-weight: 600;
  color: var(--color-text-secondary);
}

.review-value {
  font-weight: 500;
  color: var(--color-text-primary);
}

.review-info {
  padding: 1rem;
  background: rgba(59, 130, 246, 0.05);
  border-radius: 8px;
  color: var(--color-text-primary);
  font-weight: 500;
}

.review-customers {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.review-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.summary-stat {
  padding: 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border-radius: 12px;
  text-align: center;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-primary);
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

/* Approval Warning */
.approval-warning {
  display: flex;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: rgba(251, 146, 60, 0.1);
  border: 1px solid rgba(251, 146, 60, 0.3);
  border-radius: 12px;
  margin-top: 1.5rem;
}

.warning-icon {
  font-size: 1.5rem;
}

.warning-title {
  font-weight: 600;
  color: #ea580c;
  margin-bottom: 0.25rem;
}

.warning-message {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* Navigation Buttons */
.navigation-buttons {
  display: flex;
  gap: 1rem;
  padding-top: 1.5rem;
  border-top: 2px solid var(--color-border);
  margin-top: 2rem;
}

.spacer {
  flex: 1;
}

.btn {
  padding: 0.875rem 1.75rem;
  border: none;
  border-radius: 12px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(59, 130, 246, 0.4);
}

.btn-secondary {
  background: var(--color-surface);
  color: var(--color-text-primary);
  border: 2px solid var(--color-border);
}

.btn-secondary:hover {
  background: var(--color-background);
}

.btn-success {
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.btn-success:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(16, 185, 129, 0.4);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none !important;
}

/* Empty State */
.empty-state {
  padding: 3rem 2rem;
  text-align: center;
  color: var(--color-text-secondary);
  background: var(--color-background);
  border: 2px dashed var(--color-border);
  border-radius: 12px;
}

/* Autocomplete Styles */
.autocomplete-wrapper {
  position: relative;
}

.autocomplete-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  left: 0;
  right: 0;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
  max-height: 350px;
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
  border-radius: 12px 12px 0 0;
  position: sticky;
  top: 0;
  z-index: 1;
}

.suggestion-count {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  font-weight: 600;
}

.autocomplete-item {
  padding: 1rem;
  cursor: pointer;
  border-bottom: 1px solid var(--color-border);
  transition: all 0.15s;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.autocomplete-item:last-child {
  border-bottom: none;
  border-radius: 0 0 12px 12px;
}

.autocomplete-item:hover {
  background: rgba(59, 130, 246, 0.05);
  padding-left: 1.25rem;
}

.suggestion-main {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  width: 100%;
  justify-content: space-between;
}

.suggestion-icon {
  font-size: 1.125rem;
  flex-shrink: 0;
  margin-top: 2px;
}

.suggestion-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.suggestion-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.suggestion-sku {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 0.8125rem;
  padding: 0.125rem 0.375rem;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 4px;
}

.suggestion-name {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 0.9375rem;
}

.suggestion-manufacturer {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.suggestion-inventory {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.stock-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.stock-badge.in-stock {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.stock-badge.out-stock {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.suggestion-department {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.suggestion-meta {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  flex-shrink: 0;
  margin-top: 2px;
}

.suggestion-price {
  font-size: 0.875rem;
  color: #10b981;
  font-weight: 700;
  white-space: nowrap;
}

/* Product Details Box */
.product-details-box {
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05), rgba(139, 92, 246, 0.05));
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.product-details-box .detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.product-details-box .label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.product-details-box .value {
  font-weight: 600;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
}

.product-details-box .price-value {
  color: #10b981;
  font-size: 1.125rem;
  font-weight: 700;
}

/* Steps Container */
.steps-container {
  min-height: 400px;
}

.step-content {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Add Item Button */
.btn-add-item {
  width: 100%;
  padding: 1rem;
  margin-bottom: 1.5rem;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1));
  border: 2px dashed rgba(59, 130, 246, 0.3);
  border-radius: 12px;
  color: var(--color-primary);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-add-item:hover {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(139, 92, 246, 0.15));
  border-color: rgba(59, 130, 246, 0.5);
  transform: translateY(-2px);
}

/* Items List */
.items-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Item Card */
.item-card {
  position: relative;
  display: flex;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: 16px;
  transition: all 0.3s;
}

.item-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

.item-number {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  flex-shrink: 0;
  font-size: 1rem;
}

.item-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Type Tabs */
.type-row {
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.type-tabs {
  display: flex;
  gap: 0.5rem;
  flex-shrink: 0;
}

.type-tab {
  position: relative;
  cursor: pointer;
  padding: 0.75rem 1.25rem;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 10px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.type-tab input[type="radio"] {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.type-tab:hover {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.05);
}

.type-tab.active {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  border-color: transparent;
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

/* Autocomplete Container */
.autocomplete-container {
  flex: 1;
  position: relative;
}

.item-select {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 2px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: 0.95rem;
  transition: all 0.2s;
}

.item-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* Details Grid */
.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
}

.detail-box {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-box label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-input {
  padding: 0.75rem;
  border: 2px solid var(--color-border);
  border-radius: 8px;
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: 0.95rem;
  font-weight: 600;
  transition: all 0.2s;
}

.detail-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.detail-display {
  padding: 0.75rem;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  min-height: 46px;
}

.change-display {
  padding: 0.75rem;
  background: var(--color-surface);
  border: 2px solid var(--color-border);
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 46px;
}

.change-display.text-success {
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.3);
}

.change-display.text-danger {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
}

/* Remove Button */
.remove-btn {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: rgba(239, 68, 68, 0.1);
  border: 2px solid transparent;
  color: #ef4444;
  font-size: 1.25rem;
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.remove-btn:hover {
  background: #ef4444;
  color: white;
  transform: scale(1.1);
}
</style>
