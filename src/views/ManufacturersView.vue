<template>
  <div class="manufacturers-view">
    <!-- Header -->
    <div class="view-header">
      <h2>Manufacturers Database</h2>
      <button class="btn btn-primary" @click="showAddManufacturerModal">
        <span>➕</span> Add Manufacturer
      </button>
    </div>

    <!-- Manufacturers List -->
    <div class="manufacturers-grid">
      <BaseCard
        v-for="manufacturer in activeManufacturers"
        :key="manufacturer.id"
        class="manufacturer-card"
      >
        <div class="manufacturer-header">
          <div class="manufacturer-title">
            <h3>{{ manufacturer.name }}</h3>
            <BaseBadge variant="default" size="sm">{{ manufacturer.code }}</BaseBadge>
          </div>
          <div class="manufacturer-actions">
            <button class="action-btn" title="View Details" @click="viewManufacturer(manufacturer)">👁️</button>
            <button class="action-btn" title="Edit" @click="editManufacturer(manufacturer)">✏️</button>
            <button class="action-btn danger" title="Delete" @click="deleteManufacturer(manufacturer)">🗑️</button>
          </div>
        </div>

        <div class="manufacturer-meta">
          <div class="meta-item" v-if="manufacturer.country">
            <span class="meta-label">🌍 Country:</span>
            <span class="meta-value">{{ manufacturer.country }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">📦 Categories:</span>
            <span class="meta-value">{{ manufacturer.categories.length }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-label">🏷️ Products:</span>
            <span class="meta-value">{{ getTotalProducts(manufacturer) }}</span>
          </div>
        </div>

        <div class="manufacturer-categories">
          <div
            v-for="category in manufacturer.categories"
            :key="category.id"
            class="category-badge"
          >
            {{ category.name }} ({{ category.products.length }})
          </div>
        </div>
      </BaseCard>
    </div>

    <!-- Add/Edit Manufacturer Modal -->
    <BaseModal
      v-model="showManufacturerModal"
      :title="editingManufacturer ? 'Edit Manufacturer' : 'Add New Manufacturer'"
      size="lg"
      @confirm="saveManufacturer"
    >
      <div class="modal-form">
        <div class="form-section">
          <h4>Basic Information</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Manufacturer Name *</label>
              <input v-model="manufacturerForm.name" type="text" class="form-input" placeholder="e.g., Hikvision" />
            </div>
            <div class="form-group">
              <label class="form-label">Code *</label>
              <input v-model="manufacturerForm.code" type="text" class="form-input" placeholder="e.g., HIK" maxlength="10" />
              <small class="form-hint">Short code for manufacturer (max 10 characters)</small>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Description</label>
            <textarea v-model="manufacturerForm.description" class="form-input" rows="2" placeholder="Brief description of manufacturer"></textarea>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Country</label>
              <input v-model="manufacturerForm.country" type="text" class="form-input" placeholder="e.g., China, USA, Germany" />
            </div>
            <div class="form-group">
              <label class="form-label">Status *</label>
              <select v-model="manufacturerForm.status" class="form-input">
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </div>
          </div>
        </div>

        <div class="form-section">
          <h4>Contact Information</h4>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Website</label>
              <input v-model="manufacturerForm.website" type="url" class="form-input" placeholder="https://..." />
            </div>
            <div class="form-group">
              <label class="form-label">Contact Email</label>
              <input v-model="manufacturerForm.contactEmail" type="email" class="form-input" placeholder="info@example.com" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Contact Phone</label>
              <input v-model="manufacturerForm.contactPhone" type="tel" class="form-input" placeholder="+966..." />
            </div>
            <div class="form-group">
              <label class="form-label">Notes</label>
              <textarea v-model="manufacturerForm.notes" class="form-input" rows="2" placeholder="Additional notes"></textarea>
            </div>
          </div>
        </div>
      </div>
    </BaseModal>

    <!-- View Manufacturer Details Modal -->
    <BaseModal
      v-if="viewingManufacturer"
      v-model="showDetailsModal"
      :title="viewingManufacturer.name"
      size="xl"
    >
      <div class="manufacturer-details">
        <div class="details-section">
          <h3>Manufacturer Information</h3>
          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Code:</span>
              <span class="detail-value">{{ viewingManufacturer.code }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Country:</span>
              <span class="detail-value">{{ viewingManufacturer.country || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Status:</span>
              <BaseBadge :variant="viewingManufacturer.status === 'active' ? 'success' : 'danger'">
                {{ viewingManufacturer.status }}
              </BaseBadge>
            </div>
            <div class="detail-item full-width" v-if="viewingManufacturer.description">
              <span class="detail-label">Description:</span>
              <span class="detail-value">{{ viewingManufacturer.description }}</span>
            </div>
            <div class="detail-item" v-if="viewingManufacturer.website">
              <span class="detail-label">Website:</span>
              <span class="detail-value"><a :href="viewingManufacturer.website" target="_blank">{{ viewingManufacturer.website }}</a></span>
            </div>
            <div class="detail-item" v-if="viewingManufacturer.contactEmail">
              <span class="detail-label">Email:</span>
              <span class="detail-value">{{ viewingManufacturer.contactEmail }}</span>
            </div>
            <div class="detail-item" v-if="viewingManufacturer.contactPhone">
              <span class="detail-label">Phone:</span>
              <span class="detail-value">{{ viewingManufacturer.contactPhone }}</span>
            </div>
          </div>
        </div>

        <div class="details-section">
          <h3>Product Categories & Catalog</h3>
          <div
            v-for="category in viewingManufacturer.categories"
            :key="category.id"
            class="category-section"
          >
            <h4>{{ category.name }} ({{ category.products.length }} products)</h4>
            <p class="category-description" v-if="category.description">{{ category.description }}</p>
            <div class="products-list">
              <div
                v-for="product in category.products"
                :key="product.id"
                class="product-item"
              >
                <div class="product-main">
                  <span class="product-model">{{ product.modelNumber }}</span>
                  <span class="product-name">{{ product.name }}</span>
                </div>
                <div class="product-specs" v-if="product.specifications">
                  {{ product.specifications }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="closeDetailsModal">Close</button>
          <button class="btn btn-primary" @click="editManufacturer(viewingManufacturer)">Edit Manufacturer</button>
        </div>
      </template>
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
        <p class="delete-message">Are you sure you want to delete <strong>{{ deleteItem?.name }}</strong>?</p>
        <p class="delete-warning">This action cannot be undone and may affect existing products.</p>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useManufacturersStore } from '@/stores/manufacturers'
import { useToast } from '@/composables/useToast'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { Manufacturer } from '@/types'

const manufacturersStore = useManufacturersStore()
const { success, info } = useToast()

const showManufacturerModal = ref(false)
const editingManufacturer = ref<Manufacturer | null>(null)
const showDeleteModal = ref(false)
const deleteItem = ref<Manufacturer | null>(null)
const showDetailsModal = ref(false)
const viewingManufacturer = ref<Manufacturer | null>(null)

const manufacturerForm = ref({
  name: '',
  code: '',
  description: '',
  country: '',
  website: '',
  contactEmail: '',
  contactPhone: '',
  status: 'active' as 'active' | 'inactive',
  notes: ''
})

const activeManufacturers = computed(() => manufacturersStore.manufacturers)

function getTotalProducts(manufacturer: Manufacturer): number {
  return manufacturer.categories.reduce((total, cat) => total + cat.products.length, 0)
}

function showAddManufacturerModal() {
  editingManufacturer.value = null
  resetForm()
  showManufacturerModal.value = true
}

function viewManufacturer(manufacturer: Manufacturer) {
  viewingManufacturer.value = manufacturer
  showDetailsModal.value = true
}

function closeDetailsModal() {
  showDetailsModal.value = false
  viewingManufacturer.value = null
}

function editManufacturer(manufacturer: Manufacturer) {
  closeDetailsModal()
  editingManufacturer.value = manufacturer
  manufacturerForm.value = {
    name: manufacturer.name,
    code: manufacturer.code,
    description: manufacturer.description || '',
    country: manufacturer.country || '',
    website: manufacturer.website || '',
    contactEmail: manufacturer.contactEmail || '',
    contactPhone: manufacturer.contactPhone || '',
    status: manufacturer.status,
    notes: manufacturer.notes || ''
  }
  showManufacturerModal.value = true
}

function saveManufacturer() {
  // Validation
  if (!manufacturerForm.value.name || !manufacturerForm.value.code) {
    info('Please fill in all required fields (Name and Code)')
    return
  }

  const manufacturerData = {
    name: manufacturerForm.value.name,
    code: manufacturerForm.value.code.toUpperCase(),
    description: manufacturerForm.value.description,
    country: manufacturerForm.value.country,
    website: manufacturerForm.value.website,
    contactEmail: manufacturerForm.value.contactEmail,
    contactPhone: manufacturerForm.value.contactPhone,
    status: manufacturerForm.value.status,
    notes: manufacturerForm.value.notes,
    categories: editingManufacturer.value?.categories || [],
    createdBy: 'USR-001' // This should be current user
  }

  if (editingManufacturer.value) {
    manufacturersStore.updateManufacturer(editingManufacturer.value.id, manufacturerData)
    success('Manufacturer updated successfully')
  } else {
    manufacturersStore.addManufacturer(manufacturerData)
    success('Manufacturer created successfully')
  }

  showManufacturerModal.value = false
  editingManufacturer.value = null
  resetForm()
}

function deleteManufacturer(manufacturer: Manufacturer) {
  deleteItem.value = manufacturer
  showDeleteModal.value = true
}

function confirmDelete() {
  if (!deleteItem.value) return

  manufacturersStore.deleteManufacturer(deleteItem.value.id)
  success('Manufacturer deleted successfully')

  showDeleteModal.value = false
  deleteItem.value = null
}

function resetForm() {
  manufacturerForm.value = {
    name: '',
    code: '',
    description: '',
    country: '',
    website: '',
    contactEmail: '',
    contactPhone: '',
    status: 'active',
    notes: ''
  }
}
</script>

<style scoped>
.manufacturers-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  min-width: 0;
  max-width: 100%;
}

/* Header */
.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  width: 100%;
  min-width: 0;
}

.view-header h2 {
  margin: 0;
  font-size: 1.75rem;
  color: var(--color-text-primary);
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

/* Manufacturers Grid */
.manufacturers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(350px, 100%), 1fr));
  gap: 1.5rem;
  width: 100%;
  overflow: hidden;
}

.manufacturer-card {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: transform 0.2s, box-shadow 0.2s;
}

.manufacturer-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.manufacturer-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.manufacturer-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.manufacturer-title h3 {
  margin: 0;
  font-size: 1.25rem;
  color: var(--color-text-primary);
}

.manufacturer-actions {
  display: flex;
  gap: 0.5rem;
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

.manufacturer-meta {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.meta-item {
  display: flex;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.meta-label {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.meta-value {
  color: var(--color-text-primary);
}

.manufacturer-categories {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.category-badge {
  background: var(--color-background);
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
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

.form-hint {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

/* Details Modal */
.manufacturer-details {
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

.detail-value a {
  color: var(--color-primary);
  text-decoration: none;
}

.detail-value a:hover {
  text-decoration: underline;
}

.category-section {
  background: var(--color-background);
  padding: 1.25rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.category-section h4 {
  margin: 0 0 0.5rem 0;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.category-description {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin: 0 0 1rem 0;
}

.products-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.product-item {
  background: var(--color-surface);
  padding: 0.875rem;
  border-radius: 6px;
  border: 1px solid var(--color-border);
}

.product-main {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.product-model {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 0.9rem;
  min-width: 150px;
}

.product-name {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 0.9rem;
}

.product-specs {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  padding-left: 158px;
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
