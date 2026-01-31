<template>
  <div class="products-view">
    <!-- Stats Bar -->
    <div class="stats-bar">
      <div class="stat-card">
        <div class="stat-label">Total Products</div>
        <div class="stat-value">{{ stats.totalProducts }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Import Products</div>
        <div class="stat-value info">{{ stats.importProducts }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Local Products</div>
        <div class="stat-value">{{ stats.localProducts }}</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">Avg Margin</div>
        <div class="stat-value" :class="getMarginClass(stats.avgMargin)">{{ stats.avgMargin.toFixed(1) }}%</div>
      </div>
    </div>

    <!-- Header -->
    <div class="view-header">
      <div class="header-actions">
        <div class="autocomplete-wrapper">
          <input
            v-model="searchQuery"
            type="search"
            placeholder="Search products by SKU, name, manufacturer..."
            class="search-input"
            @input="handleSearchInput"
            @focus="showSuggestions = true"
            @blur="handleBlur"
          />
          <div v-if="showSuggestions && filteredSuggestions.length > 0" class="autocomplete-dropdown">
            <div class="autocomplete-header">
              <span class="suggestion-count">{{ filteredSuggestions.length }} results found</span>
            </div>
            <div
              v-for="(product, index) in filteredSuggestions"
              :key="product.id"
              class="autocomplete-item"
              @mousedown.prevent="selectProduct(product)"
            >
              <div class="suggestion-main">
                <span class="suggestion-sku">{{ product.sku }}</span>
                <span class="suggestion-name">{{ product.name }}</span>
              </div>
              <div class="suggestion-meta">
                <span class="suggestion-category">{{ product.category }}</span>
                <span class="suggestion-manufacturer">{{ product.manufacturer }}</span>
              </div>
            </div>
          </div>
        </div>
        <button class="btn btn-primary" @click="showAddProductModal">
          <span>➕</span> Add Product
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <div class="filter-group">
        <label>Category:</label>
        <select v-model="categoryFilter" class="filter-select">
          <option value="all">All Categories</option>
          <option value="Security Equipment">Security Equipment</option>
          <option value="Access Control Systems">Access Control Systems</option>
          <option value="CCTV Systems">CCTV Systems</option>
          <option value="Alarm Systems">Alarm Systems</option>
          <option value="Fire Safety Equipment">Fire Safety Equipment</option>
          <option value="Uniforms & Apparel">Uniforms & Apparel</option>
          <option value="Communication Equipment">Communication Equipment</option>
          <option value="Vehicles">Vehicles</option>
          <option value="Software & Licenses">Software & Licenses</option>
          <option value="Professional Services">Professional Services</option>
          <option value="Servers">Servers</option>
          <option value="Networking">Networking</option>
          <option value="Storage">Storage</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="import">Import Items</option>
          <option value="local">Local Items</option>
        </select>
      </div>

    </div>

    <!-- Products Table -->
    <BaseCard no-padding>
      <div class="table-container">
        <table class="products-table">
          <thead>
            <tr>
              <th>Product</th>
              <th>Category</th>
              <th>Type</th>
              <th>Costing</th>
              <th>Margin</th>
              <th>Lead Time</th>
              <th class="actions-col">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="product in filteredProducts" :key="product.id" class="product-row" @click="viewProductDetails(product)">
              <td>
                <div class="product-cell">
                  <div class="product-sku">{{ product.sku }}</div>
                  <div class="product-name">{{ product.name }}</div>
                  <div class="product-manufacturer">{{ product.manufacturer }}</div>
                </div>
              </td>
              <td>
                <BaseBadge variant="default" size="sm">{{ product.category }}</BaseBadge>
              </td>
              <td>
                <BaseBadge :variant="product.isImport ? 'info' : 'success'" size="sm">
                  {{ product.isImport ? 'Import' : 'Local' }}
                </BaseBadge>
              </td>
              <td>
                <div class="costing-cell">
                  <div v-if="product.isImport" class="cost-row-sm">
                    <span class="cost-label-sm">Origin:</span>
                    <span class="cost-value-sm">{{ formatCurrencyWithCode(product.unitCost, product.originCurrency) }}</span>
                  </div>
                  <div class="cost-row-sm">
                    <span class="cost-label-sm">Cost (SAR):</span>
                    <span class="cost-value-sm">{{ formatCurrency(product.costInSAR) }}</span>
                  </div>
                  <div class="cost-row-sm highlight">
                    <span class="cost-label-sm">Landed:</span>
                    <span class="cost-value-sm">{{ formatCurrency(product.landedCostSAR) }}</span>
                  </div>
                  <div class="cost-row-sm highlight">
                    <span class="cost-label-sm">Selling:</span>
                    <span class="cost-value-sm primary">{{ formatCurrency(product.sellingPrice) }}</span>
                  </div>
                </div>
              </td>
              <td>
                <div class="margin-display">
                  <div class="margin-percent" :class="getMarginClass(productsStore.calculateMarginPercent(product.landedCostSAR, product.sellingPrice))">
                    {{ productsStore.calculateMarginPercent(product.landedCostSAR, product.sellingPrice).toFixed(1) }}%
                  </div>
                  <div class="margin-amount">
                    {{ formatCurrency(productsStore.calculateMarginAmount(product.landedCostSAR, product.sellingPrice)) }}
                  </div>
                </div>
              </td>
              <td>
                <span class="lead-time">{{ product.leadTimeDays }} days</span>
              </td>
              <td class="actions-col">
                <div class="action-buttons">
                  <button class="action-btn" title="View Details" @click.stop="viewProductDetails(product)">👁️</button>
                  <button class="action-btn" title="Edit" @click.stop="editProduct(product)">✏️</button>
                  <button class="action-btn danger" title="Delete" @click.stop="deleteProduct(product)">🗑️</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredProducts.length === 0">
              <td colspan="8" class="empty-row">
                No products found matching your criteria
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </BaseCard>

    <!-- Add/Edit Product Modal -->
    <BaseModal
      v-model="showProductModal"
      :title="editingProduct ? 'Edit Product' : 'Add New Product'"
      size="xl"
      @confirm="saveProduct"
    >
      <div class="modal-form">
        <!-- Basic Information -->
        <div class="form-section">
          <h4>Basic Information</h4>

          <!-- Manufacturer First - Primary Source -->
          <div class="primary-source-section">
            <h5><span class="primary-icon">🏭</span> Manufacturer - Primary Ordering Source</h5>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Manufacturer *</label>
                <select v-model="productForm.manufacturerId" class="form-input" @change="onManufacturerChange">
                  <option value="">Select manufacturer</option>
                  <option v-for="mfg in manufacturersStore.activeManufacturers" :key="mfg.id" :value="mfg.id">
                    {{ mfg.name }} ({{ mfg.code }})
                  </option>
                </select>
                <small class="form-hint">Direct manufacturer - primary ordering source for POs</small>
              </div>
              <div class="form-group">
                <label class="form-label">Product Category</label>
                <select
                  v-model="productForm.manufacturerCategory"
                  class="form-input"
                  :disabled="!productForm.manufacturerId"
                  @change="onManufacturerCategoryChange"
                >
                  <option value="">Select category</option>
                  <option v-for="cat in availableManufacturerCategories" :key="cat.id" :value="cat.name">
                    {{ cat.name }}
                  </option>
                </select>
                <small class="form-hint">Category of product from manufacturer</small>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Manufacturer Product</label>
                <select
                  v-model="productForm.manufacturerProductId"
                  class="form-input"
                  :disabled="!productForm.manufacturerCategory"
                  @change="onManufacturerProductChange"
                >
                  <option value="">Select product (optional)</option>
                  <option v-for="prod in availableManufacturerProducts" :key="prod.id" :value="prod.id">
                    {{ prod.modelNumber }} - {{ prod.name }}
                  </option>
                </select>
                <small class="form-hint">Select from manufacturer catalog or leave blank for custom product</small>
              </div>
              <div class="form-group">
                <label class="form-label">Manufacturer Part Number</label>
                <input v-model="productForm.supplierPartNumber" type="text" class="form-input" placeholder="e.g., DS-2CD2385G1-I" />
                <small class="form-hint">Original manufacturer part/model number</small>
              </div>
            </div>
          </div>

          <!-- Product Identification -->
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Internal SKU *</label>
              <input v-model="productForm.sku" type="text" class="form-input" placeholder="Your internal SKU code" />
            </div>
            <div class="form-group">
              <label class="form-label">Product Name *</label>
              <input v-model="productForm.name" type="text" class="form-input" placeholder="Product name" />
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Description *</label>
            <textarea v-model="productForm.description" class="form-input" rows="2" placeholder="Detailed product description"></textarea>
          </div>

          <!-- Category & Classification -->
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Category *</label>
              <select v-model="productForm.category" class="form-input">
                <option value="">Select category</option>
                <option value="Security Equipment">Security Equipment</option>
                <option value="Access Control Systems">Access Control Systems</option>
                <option value="CCTV Systems">CCTV Systems</option>
                <option value="Alarm Systems">Alarm Systems</option>
                <option value="Fire Safety Equipment">Fire Safety Equipment</option>
                <option value="Uniforms & Apparel">Uniforms & Apparel</option>
                <option value="Communication Equipment">Communication Equipment</option>
                <option value="Vehicles">Vehicles</option>
                <option value="Software & Licenses">Software & Licenses</option>
                <option value="Professional Services">Professional Services</option>
                <option value="Servers">Servers</option>
                <option value="Networking">Networking</option>
                <option value="Storage">Storage</option>
                <option value="Other">Other</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">Unit of Measure *</label>
              <select v-model="productForm.unitOfMeasure" class="form-input">
                <option value="Unit">Unit</option>
                <option value="Set">Set</option>
                <option value="Box">Box</option>
                <option value="License">License</option>
                <option value="Hour">Hour</option>
                <option value="Day">Day</option>
                <option value="Month">Month</option>
                <option value="Year">Year</option>
              </select>
            </div>
          </div>

          <!-- Optional Supplier (Alternative Source) -->
          <div class="alternative-source-section">
            <h5><span class="alt-icon">🔄</span> Alternative Supplier (Optional)</h5>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Supplier/Distributor</label>
                <input v-model="productForm.supplier" type="text" class="form-input" placeholder="Optional distributor/reseller" />
                <small class="form-hint">Alternative ordering source if not directly from manufacturer</small>
              </div>
              <div class="form-group">
                <label class="form-label">Lead Time (days) *</label>
                <input v-model.number="productForm.leadTimeDays" type="number" class="form-input" placeholder="30" min="0" />
              </div>
            </div>
          </div>
        </div>

        <!-- Costing Information -->
        <div class="form-section">
          <h4>Costing Information</h4>
          <div class="form-group">
            <label class="form-label">Product Type *</label>
            <div class="radio-group">
              <label class="radio-label">
                <input v-model="productForm.isImport" type="radio" :value="false" />
                <span>Local Product</span>
              </label>
              <label class="radio-label">
                <input v-model="productForm.isImport" type="radio" :value="true" />
                <span>Import Product</span>
              </label>
            </div>
          </div>

          <div v-if="!productForm.isImport">
            <div class="form-group">
              <label class="form-label">Unit Cost (SAR) *</label>
              <input v-model.number="productForm.unitCost" type="number" step="0.01" class="form-input" placeholder="0.00" />
            </div>
          </div>

          <div v-if="productForm.isImport">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Origin Currency *</label>
                <select v-model="productForm.originCurrency" class="form-input" @change="updateFxRate">
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="AED">AED</option>
                  <option value="CNY">CNY</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Unit Cost (Origin Currency) *</label>
                <input v-model.number="productForm.unitCost" type="number" step="0.01" class="form-input" placeholder="0.00" />
              </div>
              <div class="form-group">
                <label class="form-label">Exchange Rate *</label>
                <input v-model.number="productForm.fxRate" type="number" step="0.0001" class="form-input" placeholder="3.75" />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Freight % *</label>
                <input v-model.number="productForm.freightPercent" type="number" step="0.1" class="form-input" placeholder="5" min="0" />
              </div>
              <div class="form-group">
                <label class="form-label">Customs % *</label>
                <input v-model.number="productForm.customsPercent" type="number" step="0.1" class="form-input" placeholder="5" min="0" />
              </div>
              <div class="form-group">
                <label class="form-label">Clearance % *</label>
                <input v-model.number="productForm.clearancePercent" type="number" step="0.1" class="form-input" placeholder="2" min="0" />
              </div>
            </div>
          </div>
        </div>

        <!-- Pricing & Margins -->
        <div class="form-section">
          <h4>Pricing & Margins</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Target Margin % *</label>
              <input v-model.number="productForm.targetMarginPercent" type="number" step="0.1" class="form-input" placeholder="20" min="0" max="100" />
            </div>
          </div>

          <!-- Cost Preview -->
          <div class="cost-preview">
            <h5>Cost Preview</h5>
            <div class="preview-row">
              <span class="preview-label">Cost in SAR:</span>
              <span class="preview-value">{{ formatCurrency(calculatedCostInSAR) }}</span>
            </div>
            <div v-if="productForm.isImport" class="preview-row">
              <span class="preview-label">Import Costs ({{ totalImportPercent }}%):</span>
              <span class="preview-value">{{ formatCurrency(calculatedImportCosts) }}</span>
            </div>
            <div class="preview-row preview-total">
              <span class="preview-label">Landed Cost:</span>
              <span class="preview-value">{{ formatCurrency(calculatedLandedCost) }}</span>
            </div>
            <div class="preview-row preview-total">
              <span class="preview-label">Selling Price ({{ productForm.targetMarginPercent }}% margin):</span>
              <span class="preview-value primary">{{ formatCurrency(calculatedSellingPrice) }}</span>
            </div>
            <div class="preview-row">
              <span class="preview-label">Margin Amount:</span>
              <span class="preview-value success">{{ formatCurrency(calculatedMarginAmount) }}</span>
            </div>
          </div>
        </div>

        <!-- Additional Information -->
        <div class="form-section">
          <h4>Additional Information</h4>
          <div class="form-group">
            <label class="form-label">Status</label>
            <select v-model="productForm.status" class="form-input">
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
              <option value="discontinued">Discontinued</option>
              <option value="out_of_stock">Out of Stock</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">Notes</label>
            <textarea v-model="productForm.notes" class="form-input" rows="3" placeholder="Additional notes..."></textarea>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- Delete Confirmation Modal -->
    <BaseModal
      v-model="showDeleteModal"
      title="Confirm Delete"
      size="sm"
      confirm-text="Delete"
      @confirm="confirmDelete"
    >
      <div class="delete-modal-content">
        <div class="delete-icon">🗑️</div>
        <p class="delete-message">Are you sure you want to delete <strong>{{ deleteItem?.sku }}</strong>?</p>
        <p class="delete-warning">This action cannot be undone.</p>
      </div>
    </BaseModal>

    <!-- Product Details Modal -->
    <BaseModal
      v-if="selectedProduct"
      v-model="showDetailsModal"
      :title="`${selectedProduct.sku} - ${selectedProduct.name}`"
      size="xl"
    >
      <div class="product-details">
        <!-- Product Information -->
        <div class="details-section">
          <h3>Product Information</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">SKU:</span>
              <span class="detail-value">{{ selectedProduct.sku }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Product Name:</span>
              <span class="detail-value">{{ selectedProduct.name }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Category:</span>
              <span class="detail-value">{{ selectedProduct.category }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Manufacturer:</span>
              <span class="detail-value">{{ selectedProduct.manufacturer }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Supplier:</span>
              <span class="detail-value">{{ selectedProduct.supplier }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Supplier Part #:</span>
              <span class="detail-value">{{ selectedProduct.supplierPartNumber || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Unit of Measure:</span>
              <span class="detail-value">{{ selectedProduct.unitOfMeasure }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Lead Time:</span>
              <span class="detail-value">{{ selectedProduct.leadTimeDays }} days</span>
            </div>
            <div class="detail-item full-width">
              <span class="detail-label">Description:</span>
              <span class="detail-value">{{ selectedProduct.description }}</span>
            </div>
          </div>
        </div>

        <!-- Cost Breakdown -->
        <div class="details-section">
          <h3>Cost Breakdown</h3>
          <div class="breakdown-table">
            <div v-if="selectedProduct.isImport">
              <div class="breakdown-row">
                <span class="breakdown-label">Unit Cost ({{ selectedProduct.originCurrency }}):</span>
                <span class="breakdown-value">{{ formatCurrencyWithCode(selectedProduct.unitCost, selectedProduct.originCurrency) }}</span>
              </div>
              <div class="breakdown-row">
                <span class="breakdown-label">Exchange Rate ({{ selectedProduct.originCurrency }}/SAR):</span>
                <span class="breakdown-value">{{ selectedProduct.fxRate }}</span>
              </div>
              <div class="breakdown-row">
                <span class="breakdown-label">Cost in SAR:</span>
                <span class="breakdown-value">{{ formatCurrency(selectedProduct.costInSAR) }}</span>
              </div>
              <div class="breakdown-row">
                <span class="breakdown-label">Freight ({{ selectedProduct.freightPercent }}%):</span>
                <span class="breakdown-value">{{ formatCurrency(selectedProduct.costInSAR * (selectedProduct.freightPercent || 0) / 100) }}</span>
              </div>
              <div class="breakdown-row">
                <span class="breakdown-label">Customs ({{ selectedProduct.customsPercent }}%):</span>
                <span class="breakdown-value">{{ formatCurrency(selectedProduct.costInSAR * (selectedProduct.customsPercent || 0) / 100) }}</span>
              </div>
              <div class="breakdown-row">
                <span class="breakdown-label">Clearance ({{ selectedProduct.clearancePercent }}%):</span>
                <span class="breakdown-value">{{ formatCurrency(selectedProduct.costInSAR * (selectedProduct.clearancePercent || 0) / 100) }}</span>
              </div>
            </div>
            <div v-else>
              <div class="breakdown-row">
                <span class="breakdown-label">Unit Cost (SAR):</span>
                <span class="breakdown-value">{{ formatCurrency(selectedProduct.unitCost) }}</span>
              </div>
            </div>
            <div class="breakdown-row highlight">
              <span class="breakdown-label">Landed Cost (SAR):</span>
              <span class="breakdown-value">{{ formatCurrency(selectedProduct.landedCostSAR) }}</span>
            </div>
            <div class="breakdown-row highlight">
              <span class="breakdown-label">Selling Price (SAR):</span>
              <span class="breakdown-value primary">{{ formatCurrency(selectedProduct.sellingPrice) }}</span>
            </div>
            <div class="breakdown-row success-row">
              <span class="breakdown-label">Margin:</span>
              <span class="breakdown-value">
                {{ productsStore.calculateMarginPercent(selectedProduct.landedCostSAR, selectedProduct.sellingPrice).toFixed(2) }}%
                ({{ formatCurrency(productsStore.calculateMarginAmount(selectedProduct.landedCostSAR, selectedProduct.sellingPrice)) }})
              </span>
            </div>
            <div class="breakdown-row">
              <span class="breakdown-label">Target Margin:</span>
              <span class="breakdown-value">{{ selectedProduct.targetMarginPercent }}%</span>
            </div>
          </div>
        </div>

        <!-- Metadata -->
        <div class="details-section">
          <h3>Additional Information</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Created By:</span>
              <span class="detail-value">{{ selectedProduct.createdBy }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Created At:</span>
              <span class="detail-value">{{ formatDate(selectedProduct.createdAt) }}</span>
            </div>
            <div v-if="selectedProduct.notes" class="detail-item full-width">
              <span class="detail-label">Notes:</span>
              <span class="detail-value">{{ selectedProduct.notes }}</span>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">Close</button>
          <button class="btn btn-primary" @click="editProduct(selectedProduct)">Edit Product</button>
        </div>
      </template>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProductsStore } from '@/stores/products'
import { useManufacturersStore } from '@/stores/manufacturers'
import { useExchangeRatesStore } from '@/stores/exchangeRates'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { Product, ProductCategory } from '@/types'

const productsStore = useProductsStore()
const manufacturersStore = useManufacturersStore()
const exchangeRatesStore = useExchangeRatesStore()
const { success, info } = useToast()

const searchQuery = ref('')
const categoryFilter = ref('all')
const typeFilter = ref('all')
const showSuggestions = ref(false)

const showProductModal = ref(false)
const editingProduct = ref<Product | null>(null)
const showDeleteModal = ref(false)
const deleteItem = ref<Product | null>(null)
const showDetailsModal = ref(false)
const selectedProduct = ref<Product | null>(null)

// Product form
const productForm = ref({
  sku: '',
  name: '',
  description: '',
  category: '' as ProductCategory | '',
  supplier: '',
  supplierPartNumber: '',
  manufacturer: '',
  manufacturerId: '',
  manufacturerProductId: '',
  manufacturerCategory: '' as ProductCategory | '',
  isImport: false,
  originCurrency: 'SAR' as any,
  unitCost: 0,
  fxRate: 1,
  freightPercent: 0,
  customsPercent: 0,
  clearancePercent: 0,
  targetMarginPercent: 20,
  unitOfMeasure: 'Unit' as any,
  leadTimeDays: 30,
  notes: ''
})

// Manufacturer selection helpers
const selectedManufacturer = computed(() => {
  if (!productForm.value.manufacturerId) return null
  return manufacturersStore.getManufacturerById(productForm.value.manufacturerId)
})

const availableManufacturerCategories = computed(() => {
  if (!selectedManufacturer.value) return []
  return selectedManufacturer.value.categories
})

const availableManufacturerProducts = computed(() => {
  if (!productForm.value.manufacturerId || !productForm.value.manufacturerCategory) return []
  return manufacturersStore.getProductsByCategory(
    productForm.value.manufacturerId,
    productForm.value.manufacturerCategory
  )
})

// Stats
const stats = computed(() => productsStore.getProductStats())

// Autocomplete suggestions - limited to 10 results for better UX
const filteredSuggestions = computed(() => {
  if (!searchQuery.value || searchQuery.value.length < 2) {
    return []
  }

  const query = searchQuery.value.toLowerCase()
  let products = productsStore.products

  // Apply category filter to suggestions
  if (categoryFilter.value !== 'all') {
    products = products.filter(p => p.category === categoryFilter.value)
  }

  // Apply type filter to suggestions
  if (typeFilter.value === 'import') {
    products = products.filter(p => p.isImport)
  } else if (typeFilter.value === 'local') {
    products = products.filter(p => !p.isImport)
  }

  // Search by SKU, name, or manufacturer
  products = products.filter(p =>
    p.sku.toLowerCase().includes(query) ||
    p.name.toLowerCase().includes(query) ||
    p.manufacturer.toLowerCase().includes(query) ||
    (p.supplier && p.supplier.toLowerCase().includes(query))
  )

  // Limit to 10 suggestions for autocomplete
  return products.slice(0, 10)
})

// Filtered products for table display
const filteredProducts = computed(() => {
  let products = productsStore.products

  if (categoryFilter.value !== 'all') {
    products = products.filter(p => p.category === categoryFilter.value)
  }

  if (typeFilter.value === 'import') {
    products = products.filter(p => p.isImport)
  } else if (typeFilter.value === 'local') {
    products = products.filter(p => !p.isImport)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    products = products.filter(p =>
      p.sku.toLowerCase().includes(query) ||
      p.name.toLowerCase().includes(query) ||
      p.manufacturer.toLowerCase().includes(query) ||
      (p.supplier && p.supplier.toLowerCase().includes(query))
    )
  }

  return products
})

// Form calculations
const calculatedCostInSAR = computed(() => {
  return productsStore.calculateCostInSAR(productForm.value.originCurrency, productForm.value.unitCost)
})

const totalImportPercent = computed(() => {
  return (productForm.value.freightPercent || 0) +
         (productForm.value.customsPercent || 0) +
         (productForm.value.clearancePercent || 0)
})

const calculatedImportCosts = computed(() => {
  if (!productForm.value.isImport) return 0
  return calculatedCostInSAR.value * (totalImportPercent.value / 100)
})

const calculatedLandedCost = computed(() => {
  return productsStore.calculateLandedCost(
    calculatedCostInSAR.value,
    productForm.value.isImport,
    productForm.value.freightPercent || 0,
    productForm.value.customsPercent || 0,
    productForm.value.clearancePercent || 0
  )
})

const calculatedSellingPrice = computed(() => {
  return productsStore.calculateSellingPrice(
    calculatedLandedCost.value,
    productForm.value.targetMarginPercent
  )
})

const calculatedMarginAmount = computed(() => {
  return productsStore.calculateMarginAmount(
    calculatedLandedCost.value,
    calculatedSellingPrice.value
  )
})

// Helper functions
function formatCurrency(amount: number) {
  return new Intl.NumberFormat('en-SA', {
    style: 'currency',
    currency: 'SAR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount)
}

function formatCurrencyWithCode(amount: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(amount)
}

function formatDate(dateString: string) {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function getMarginClass(margin: number) {
  if (margin >= 30) return 'success'
  if (margin >= 20) return 'warning'
  return 'danger'
}

function resetProductForm() {
  productForm.value = {
    sku: '',
    name: '',
    description: '',
    category: '' as ProductCategory | '',
    supplier: '',
    supplierPartNumber: '',
    manufacturer: '',
    manufacturerId: '',
    manufacturerProductId: '',
    manufacturerCategory: '' as ProductCategory | '',
    isImport: false,
    originCurrency: 'SAR' as any,
    unitCost: 0,
    fxRate: 1,
    freightPercent: 0,
    customsPercent: 0,
    clearancePercent: 0,
    targetMarginPercent: 20,
    unitOfMeasure: 'Unit' as any,
    leadTimeDays: 30,
    notes: ''
  }
}

function updateFxRate() {
  // Get rate from centralized exchange rates store
  productForm.value.fxRate = exchangeRatesStore.getCurrentRate(productForm.value.originCurrency)
}

// Manufacturer selection handlers
function onManufacturerChange() {
  // Reset category and product when manufacturer changes
  productForm.value.manufacturerCategory = ''
  productForm.value.manufacturerProductId = ''

  // Update manufacturer name for legacy compatibility
  const manufacturer = manufacturersStore.getManufacturerById(productForm.value.manufacturerId)
  if (manufacturer) {
    productForm.value.manufacturer = manufacturer.name
  }
}

function onManufacturerCategoryChange() {
  // Reset product when category changes
  productForm.value.manufacturerProductId = ''

  // Auto-fill category if not set
  if (!productForm.value.category && productForm.value.manufacturerCategory) {
    productForm.value.category = productForm.value.manufacturerCategory
  }
}

function onManufacturerProductChange() {
  if (!productForm.value.manufacturerProductId) return

  const result = manufacturersStore.getProductById(productForm.value.manufacturerProductId)
  if (result) {
    // Auto-fill product details from manufacturer catalog
    productForm.value.name = result.product.name
    productForm.value.description = result.product.description || ''
    productForm.value.supplierPartNumber = result.product.modelNumber

    // Auto-fill category
    if (!productForm.value.category) {
      productForm.value.category = result.product.category
    }
  }
}

// Autocomplete functions
function handleSearchInput() {
  if (searchQuery.value.length >= 2) {
    showSuggestions.value = true
  } else {
    showSuggestions.value = false
  }
}

function handleBlur() {
  // Delay hiding to allow click events on suggestions to fire
  setTimeout(() => {
    showSuggestions.value = false
  }, 200)
}

function selectProduct(product: Product) {
  searchQuery.value = product.sku
  showSuggestions.value = false
  viewProductDetails(product)
}

// Actions
function showAddProductModal() {
  editingProduct.value = null
  resetProductForm()
  showProductModal.value = true
}

function viewProductDetails(product: Product) {
  selectedProduct.value = product
  showDetailsModal.value = true
}

function closeDetailsModal() {
  showDetailsModal.value = false
  selectedProduct.value = null
}

function editProduct(product: Product) {
  closeDetailsModal()
  editingProduct.value = product
  productForm.value = {
    sku: product.sku,
    name: product.name,
    description: product.description,
    category: product.category,
    supplier: product.supplier,
    supplierPartNumber: product.supplierPartNumber || '',
    manufacturer: product.manufacturer,
    manufacturerId: product.manufacturerId || '',
    manufacturerProductId: product.manufacturerProductId || '',
    manufacturerCategory: product.category, // Use product category as manufacturer category
    isImport: product.isImport,
    originCurrency: product.originCurrency,
    unitCost: product.unitCost,
    fxRate: product.fxRate,
    freightPercent: product.freightPercent || 0,
    customsPercent: product.customsPercent || 0,
    clearancePercent: product.clearancePercent || 0,
    targetMarginPercent: product.targetMarginPercent,
    unitOfMeasure: product.unitOfMeasure,
    leadTimeDays: product.leadTimeDays,
    notes: product.notes || ''
  }
  showProductModal.value = true
}

function saveProduct() {
  // Validation
  if (!productForm.value.sku || !productForm.value.name || !productForm.value.description ||
      !productForm.value.category || !productForm.value.manufacturerId) {
    info('Please fill in all required fields (SKU, Name, Description, Category, Manufacturer)')
    return
  }

  const productData = {
    sku: productForm.value.sku,
    name: productForm.value.name,
    description: productForm.value.description,
    category: productForm.value.category,
    supplier: productForm.value.supplier,
    supplierPartNumber: productForm.value.supplierPartNumber,
    manufacturer: productForm.value.manufacturer,
    manufacturerId: productForm.value.manufacturerId,
    manufacturerProductId: productForm.value.manufacturerProductId || undefined,
    isImport: productForm.value.isImport,
    originCurrency: productForm.value.isImport ? productForm.value.originCurrency : 'SAR' as any,
    unitCost: productForm.value.unitCost,
    fxRate: productForm.value.isImport ? productForm.value.fxRate : 1,
    costInSAR: calculatedCostInSAR.value,
    freightPercent: productForm.value.isImport ? productForm.value.freightPercent : 0,
    customsPercent: productForm.value.isImport ? productForm.value.customsPercent : 0,
    clearancePercent: productForm.value.isImport ? productForm.value.clearancePercent : 0,
    landedCostSAR: calculatedLandedCost.value,
    targetMarginPercent: productForm.value.targetMarginPercent,
    sellingPrice: calculatedSellingPrice.value,
    unitOfMeasure: productForm.value.unitOfMeasure,
    leadTimeDays: productForm.value.leadTimeDays,
    notes: productForm.value.notes
  }

  if (editingProduct.value) {
    productsStore.updateProduct(editingProduct.value.id, productData)
    success('Product updated successfully')
  } else {
    productsStore.addProduct(productData as any)
    success('Product created successfully')
  }

  showProductModal.value = false
  editingProduct.value = null
  resetProductForm()
}

function deleteProduct(product: Product) {
  deleteItem.value = product
  showDeleteModal.value = true
}

function confirmDelete() {
  if (!deleteItem.value) return

  productsStore.deleteProduct(deleteItem.value.id)
  success('Product deleted successfully')

  showDeleteModal.value = false
  deleteItem.value = null
}

</script>

<style scoped>
.products-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

/* Stats Bar */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.stat-value.success {
  color: var(--color-success);
}

.stat-value.info {
  color: var(--color-info);
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

.autocomplete-wrapper {
  position: relative;
  flex: 1;
}

.search-input {
  width: 100%;
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

/* Autocomplete Dropdown */
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
  gap: 1rem;
}

.suggestion-sku {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 0.95rem;
  min-width: 120px;
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
  padding-left: 136px;
}

.suggestion-category {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  background: var(--color-background);
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
  font-weight: 500;
}

.suggestion-manufacturer {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
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

.btn-secondary {
  background: var(--color-border);
  color: var(--color-text-primary);
}

.btn-secondary:hover {
  background: var(--color-surface-hover);
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

.products-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}

.products-table thead {
  background: var(--color-background);
  border-bottom: 2px solid var(--color-border);
}

.products-table th {
  padding: 1.25rem 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  font-size: 0.8rem;
  letter-spacing: 0.5px;
  white-space: nowrap;
}

.products-table tbody tr {
  border-bottom: 1px solid var(--color-border);
  transition: background-color 0.2s;
  cursor: pointer;
}

.products-table tbody tr:hover {
  background: var(--color-surface-hover);
}

.products-table td {
  padding: 1rem;
}

.product-cell {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 250px;
}

.product-sku {
  font-weight: 700;
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
  font-size: 0.95rem;
  letter-spacing: 0.5px;
}

.product-name {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 1rem;
}

.product-manufacturer {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.costing-cell {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  min-width: 200px;
}

.cost-row-sm {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  font-size: 0.875rem;
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
}

.cost-row-sm.highlight {
  font-weight: 600;
  background: var(--color-primary-light);
  padding: 0.4rem 0.6rem;
}

.cost-label-sm {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.cost-value-sm {
  color: var(--color-text-primary);
  font-weight: 500;
}

.cost-value-sm.primary {
  color: var(--color-primary);
  font-weight: 700;
}

.margin-display {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  align-items: flex-start;
}

.margin-percent {
  font-weight: 700;
  font-size: 1.1rem;
  padding: 0.25rem 0.75rem;
  border-radius: 6px;
}

.margin-amount {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.lead-time {
  color: var(--color-text-secondary);
  font-size: 0.95rem;
}

.actions-col {
  text-align: right;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
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

.action-btn.danger:hover {
  filter: hue-rotate(340deg);
}

.empty-row {
  text-align: center;
  padding: 3rem;
  color: var(--color-text-secondary);
  font-size: 1.05rem;
}

/* Modal Form Styles */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.form-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.form-section h4 {
  margin: 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.form-section h5 {
  margin: 0 0 0.75rem 0;
  font-size: 1rem;
  color: var(--color-text-primary);
  font-weight: 600;
}

.primary-source-section {
  background: var(--color-primary-light);
  padding: 1.25rem;
  border-radius: 8px;
  border-left: 4px solid var(--color-primary);
  margin-bottom: 1rem;
}

.primary-source-section h5 {
  color: var(--color-primary);
  margin-bottom: 1rem;
}

.primary-icon {
  font-size: 1.2rem;
  margin-right: 0.5rem;
}

.alternative-source-section {
  background: var(--color-background);
  padding: 1.25rem;
  border-radius: 8px;
  border-left: 4px solid var(--color-border);
  margin-top: 1rem;
}

.alternative-source-section h5 {
  color: var(--color-text-secondary);
  margin-bottom: 1rem;
}

.alt-icon {
  font-size: 1.2rem;
  margin-right: 0.5rem;
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

.form-input:disabled {
  background: var(--color-background);
  opacity: 0.6;
  cursor: not-allowed;
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

.radio-group {
  display: flex;
  gap: 1.5rem;
  padding: 0.5rem 0;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.radio-label input[type="radio"] {
  cursor: pointer;
  width: 18px;
  height: 18px;
}

/* Cost Preview */
.cost-preview {
  background: var(--color-background);
  padding: 1.5rem;
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.preview-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
}

.preview-row:last-child {
  border-bottom: none;
}

.preview-row.preview-total {
  font-weight: 700;
  font-size: 1.05rem;
  background: var(--color-primary-light);
  padding: 0.75rem 1rem;
  margin: 0.5rem -1rem;
  border-radius: 6px;
  border-bottom: none;
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
  font-weight: 700;
}

.preview-value.success {
  color: var(--color-success);
  font-weight: 700;
}

/* Product Details Modal */
.product-details {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.details-section {
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 1.5rem;
}

.details-section:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.details-section h3 {
  margin: 0 0 1rem 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
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

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 0.95rem;
  color: var(--color-text-primary);
}

.breakdown-table {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-width: 600px;
}

.breakdown-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  background: var(--color-background);
}

.breakdown-row.highlight {
  background: var(--color-primary-light);
  font-weight: 700;
  font-size: 1.05rem;
}

.breakdown-row.success-row {
  background: var(--color-success-light);
  font-weight: 700;
  font-size: 1.05rem;
}

.breakdown-label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.breakdown-value {
  color: var(--color-text-primary);
  font-weight: 600;
}

.breakdown-value.primary {
  color: var(--color-primary);
  font-weight: 700;
}

.modal-actions {
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

/* Delete Modal */
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
