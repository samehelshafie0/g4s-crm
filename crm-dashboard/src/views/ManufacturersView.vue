<script setup lang="ts">
import { manufacturersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import { onMounted, ref, computed } from 'vue'
import {
  Plus,
  Search,
  Pencil,
  Trash2,
  X,
  Globe,
  Mail,
  Phone,
  ExternalLink,
} from 'lucide-vue-next'
import type { Manufacturer, ManufacturerCategory, VendorType } from '@/types'
import { useManufacturersStore } from '@/stores/manufacturers'

const mfrStore = useManufacturersStore()
onMounted(async () => { try { manufacturers.value = await allPages(manufacturersService.list) } catch (e) { window.alert(errorMessage(e)) } })

const countryFlags: Record<string, string> = {
  China: '🇨🇳',
  Sweden: '🇸🇪',
  USA: '🇺🇸',
  Germany: '🇩🇪',
}

function uid(): string {
  return Math.random().toString(36).slice(2, 11)
}

const vendorTypeConfig: Record<VendorType, { label: string; badge: string }> = {
  manufacturer: { label: 'Manufacturer', badge: 'badge-info' },
  supplier: { label: 'Supplier', badge: 'badge-primary' },
  both: { label: 'Mfr + Supplier', badge: 'badge-success' },
}

const typeFilter = ref<VendorType | ''>('')

const manufacturers = ref<Manufacturer[]>([])
const searchQuery = ref('')
const showModal = ref(false)
const editingId = ref<string | null>(null)
const newCategoryName = ref('')

const defaultForm = (): Omit<Manufacturer, 'id' | 'createdAt' | 'updatedAt'> => ({
  name: '',
  code: '',
  country: '',
  contactEmail: '',
  contactPhone: '',
  website: '',
  categories: [],
  vendorType: 'manufacturer',
  isActive: true,
})

const form = ref<Omit<Manufacturer, 'id' | 'createdAt' | 'updatedAt'>>(defaultForm())

const filteredManufacturers = computed(() => {
  let list = manufacturers.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter((m) => m.name.toLowerCase().includes(q) || m.code.toLowerCase().includes(q))
  }
  if (typeFilter.value) {
    list = list.filter((m) => m.vendorType === typeFilter.value)
  }
  return list
})

function openAddModal() {
  editingId.value = null
  form.value = defaultForm()
  newCategoryName.value = ''
  showModal.value = true
}

function openEditModal(m: Manufacturer) {
  editingId.value = m.id
  form.value = {
    name: m.name,
    code: m.code,
    country: m.country,
    contactEmail: m.contactEmail,
    contactPhone: m.contactPhone,
    website: m.website,
    categories: m.categories.map((c) => ({ ...c })),
    vendorType: m.vendorType,
    isActive: m.isActive,
  }
  newCategoryName.value = ''
  showModal.value = true
}

function addCategory() {
  const name = newCategoryName.value.trim()
  if (!name) return
  form.value.categories.push({ id: uid(), name, description: '' })
  newCategoryName.value = ''
}

function removeCategory(catId: string) {
  form.value.categories = form.value.categories.filter((c) => c.id !== catId)
}

const saving = ref(false)
async function saveManufacturer() {
 if (saving.value) return
 saving.value = true
 try {
   const old = editingId.value ? (await manufacturersService.get(editingId.value)).data : null
   const result = editingId.value ? await manufacturersService.update(editingId.value, form.value) : await manufacturersService.create(form.value)
   editingId.value = result.data.id
   for (const category of form.value.categories) {
     if (old?.categories.some(c => c.id === category.id)) await manufacturersService.updateCategory(result.data.id, category.id, category)
     else category.id = (await manufacturersService.addCategory(result.data.id, category)).data.id
   }
   for (const removed of old?.categories ?? []) if (!form.value.categories.some(c => c.id === removed.id)) await manufacturersService.deleteCategory(result.data.id, removed.id)
   manufacturers.value = await allPages(manufacturersService.list)
   showModal.value = false
 } catch (e) { window.alert(errorMessage(e)) } finally { saving.value = false }
}
async function deleteManufacturer(id: string) {
 try { await manufacturersService.delete(id); manufacturers.value = manufacturers.value.filter(m => m.id !== id) }
 catch (e) { window.alert(errorMessage(e)) }
}
</script>

<template>
  <div class="manufacturers-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Manufacturers & Suppliers</h1>
        <p class="page-header-subtitle">
          Manage equipment manufacturers, suppliers, and distributors
        </p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <Plus :size="18" />
        Add Vendor
      </button>
    </div>

    <div class="search-filter-row mb-6">
      <div class="search-input" style="flex:1; max-width:420px">
        <Search :size="18" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          class="form-input"
          placeholder="Search by name or code…"
        />
      </div>
      <select v-model="typeFilter" class="form-select" style="width:180px">
        <option value="">All Types</option>
        <option value="manufacturer">Manufacturers</option>
        <option value="supplier">Suppliers</option>
        <option value="both">Both</option>
      </select>
    </div>

    <div v-if="filteredManufacturers.length" class="manufacturer-grid">
      <div v-for="m in filteredManufacturers" :key="m.id" class="card manufacturer-card">
        <div class="card-header">
          <div class="mfr-title-block">
            <h3 class="card-title">{{ m.name }}</h3>
            <span class="badge badge-neutral text-mono">{{ m.code }}</span>
            <span :class="['badge', vendorTypeConfig[m.vendorType].badge]">{{ vendorTypeConfig[m.vendorType].label }}</span>
          </div>
          <span
            :class="['badge badge-dot', m.isActive ? 'badge-success' : 'badge-danger']"
          >
            {{ m.isActive ? 'Active' : 'Inactive' }}
          </span>
        </div>

        <div class="card-body">
          <div class="mfr-country mb-3">
            <Globe :size="15" class="text-muted" />
            <span>{{ countryFlags[m.country] || '🌍' }} {{ m.country }}</span>
          </div>

          <div class="mfr-contact-list">
            <div class="mfr-contact-item">
              <Mail :size="14" class="text-muted" />
              <a :href="'mailto:' + m.contactEmail">{{ m.contactEmail }}</a>
            </div>
            <div class="mfr-contact-item">
              <Phone :size="14" class="text-muted" />
              <span>{{ m.contactPhone }}</span>
            </div>
            <div class="mfr-contact-item">
              <ExternalLink :size="14" class="text-muted" />
              <a :href="m.website" target="_blank" rel="noopener">{{ m.website.replace(/^https?:\/\//, '') }}</a>
            </div>
          </div>

          <div class="mfr-categories mt-4">
            <span class="form-label mb-2" style="margin-bottom: var(--space-2)">Categories</span>
            <div class="flex flex-wrap gap-2">
              <span v-for="cat in m.categories" :key="cat.id" class="badge badge-primary">
                {{ cat.name }}
              </span>
            </div>
          </div>
        </div>

        <div class="card-footer">
          <button class="btn btn-ghost btn-sm" @click="openEditModal(m)">
            <Pencil :size="14" /> Edit
          </button>
          <button class="btn btn-ghost btn-sm text-danger" @click="deleteManufacturer(m.id)">
            <Trash2 :size="14" /> Delete
          </button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <div class="empty-state-icon">🏭</div>
      <h3 class="empty-state-title">No manufacturers found</h3>
      <p class="empty-state-text">
        {{ searchQuery ? 'Try adjusting your search terms.' : 'Add your first manufacturer to get started.' }}
      </p>
      <button v-if="!searchQuery" class="btn btn-primary" @click="openAddModal">
        <Plus :size="18" /> Add Manufacturer
      </button>
    </div>

    <!-- Add / Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">
              {{ editingId ? 'Edit Vendor' : 'Add Vendor' }}
            </h2>
            <button class="modal-close" @click="showModal = false">
              <X :size="20" />
            </button>
          </div>

          <div class="modal-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Name <span class="required">*</span></label>
                <input v-model="form.name" type="text" class="form-input" placeholder="e.g. Hikvision" />
              </div>
              <div class="form-group">
                <label class="form-label">Code <span class="required">*</span></label>
                <input v-model="form.code" type="text" class="form-input" placeholder="e.g. HIK" />
              </div>
              <div class="form-group">
                <label class="form-label">Type <span class="required">*</span></label>
                <select v-model="form.vendorType" class="form-select">
                  <option value="manufacturer">Manufacturer</option>
                  <option value="supplier">Supplier</option>
                  <option value="both">Manufacturer + Supplier</option>
                </select>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Country</label>
                <input v-model="form.country" type="text" class="form-input" placeholder="e.g. China" />
              </div>
              <div class="form-group">
                <label class="form-label">Contact Email</label>
                <input v-model="form.contactEmail" type="email" class="form-input" placeholder="sales@example.com" />
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Phone</label>
                <input v-model="form.contactPhone" type="text" class="form-input" placeholder="+1 800 123 4567" />
              </div>
              <div class="form-group">
                <label class="form-label">Website</label>
                <input v-model="form.website" type="url" class="form-input" placeholder="https://www.example.com" />
              </div>
            </div>

            <div class="divider" />

            <div class="form-group">
              <label class="form-label">Categories</label>
              <div v-if="form.categories.length" class="category-chips mb-3">
                <span v-for="cat in form.categories" :key="cat.id" class="chip">
                  {{ cat.name }}
                  <button class="chip-remove" @click="removeCategory(cat.id)">
                    <X :size="12" />
                  </button>
                </span>
              </div>
              <div class="flex gap-2">
                <input
                  v-model="newCategoryName"
                  type="text"
                  class="form-input"
                  placeholder="New category name…"
                  @keyup.enter="addCategory"
                />
                <button class="btn btn-secondary" @click="addCategory">Add</button>
              </div>
            </div>

            <div class="divider" />

            <label class="form-checkbox">
              <input v-model="form.isActive" type="checkbox" />
              Active
            </label>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showModal = false">Cancel</button>
            <button class="btn btn-primary" :disabled="!form.name || !form.code" @click="saveManufacturer">
              {{ editingId ? 'Save Changes' : 'Add Vendor' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.manufacturers-page {
  padding: var(--space-6);
}

.search-bar {
  max-width: 420px;
}

.manufacturer-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-5);
}

@media (max-width: 1024px) {
  .manufacturer-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .manufacturer-grid {
    grid-template-columns: 1fr;
  }
}

.manufacturer-card {
  display: flex;
  flex-direction: column;
}

.manufacturer-card .card-body {
  flex: 1;
}

.mfr-title-block {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.mfr-country {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
}

.mfr-contact-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.mfr-contact-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
}

.mfr-contact-item a {
  color: var(--color-primary);
  font-size: var(--text-sm);
}

.mfr-categories .form-label {
  display: block;
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-neutral-500);
  font-weight: var(--font-semibold);
}

.category-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.text-danger {
  color: var(--color-danger) !important;
}

.search-filter-row {
  display: flex;
  gap: var(--space-3);
  align-items: center;
}
</style>
