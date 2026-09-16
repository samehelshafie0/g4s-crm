<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus,
  Search,
  Eye,
  Pencil,
  Trash2,
  X,
  Building2,
  MapPin,
  Users,
  Phone,
  Mail,
  FileText,
} from 'lucide-vue-next'
import type { Customer, CustomerStatus, CustomerType, Sector } from '@/types'
import { useCustomersStore } from '@/stores/customers'

const customerStore = useCustomersStore()

const searchQuery = ref('')
const sectorFilter = ref<Sector | 'all'>('all')
const statusFilter = ref<CustomerStatus | 'all'>('all')
const typeFilter = ref<CustomerType | 'all'>('all')

const showAddModal = ref(false)
const showViewModal = ref(false)
const editingCustomer = ref<Customer | null>(null)
const viewingCustomer = ref<Customer | null>(null)

const defaultForm = (): Omit<Customer, 'id' | 'createdAt' | 'updatedAt'> => ({
  companyName: '',
  sector: 'government',
  region: '',
  status: 'prospect',
  type: 'get',
  crNumber: '',
  vatNumber: '',
  sites: [],
  contacts: [],
  notes: '',
})

const form = ref(defaultForm())

const customers = computed(() => customerStore.customers)
const loading = computed(() => customerStore.loading)

onMounted(() => {
  customerStore.fetchCustomers()
})

const sectorOptions: { value: Sector; label: string }[] = [
  { value: 'government', label: 'Government' },
  { value: 'healthcare', label: 'Healthcare' },
  { value: 'education', label: 'Education' },
  { value: 'retail', label: 'Retail' },
  { value: 'banking', label: 'Banking' },
  { value: 'oil-gas', label: 'Oil & Gas' },
  { value: 'telecom', label: 'Telecom' },
  { value: 'hospitality', label: 'Hospitality' },
  { value: 'real-estate', label: 'Real Estate' },
  { value: 'other', label: 'Other' },
]

const sectorLabel = (s: Sector) => sectorOptions.find((o) => o.value === s)?.label ?? s

const statusBadgeClass = (s: CustomerStatus) =>
  ({ active: 'badge-success', inactive: 'badge-danger', prospect: 'badge-warning' })[s]

const typeBadgeClass = (t: CustomerType) =>
  ({ get: 'badge-primary', grow: 'badge-success' })[t]

const filteredCustomers = computed(() =>
  customers.value.filter((c) => {
    if (searchQuery.value && !c.companyName.toLowerCase().includes(searchQuery.value.toLowerCase())) return false
    if (sectorFilter.value !== 'all' && c.sector !== sectorFilter.value) return false
    if (statusFilter.value !== 'all' && c.status !== statusFilter.value) return false
    if (typeFilter.value !== 'all' && c.type !== typeFilter.value) return false
    return true
  }),
)

function openAddModal() {
  editingCustomer.value = null
  form.value = defaultForm()
  showAddModal.value = true
}

function openEditModal(c: Customer) {
  editingCustomer.value = c
  form.value = {
    companyName: c.companyName,
    sector: c.sector,
    region: c.region,
    status: c.status,
    type: c.type,
    crNumber: c.crNumber,
    vatNumber: c.vatNumber,
    sites: c.sites,
    contacts: c.contacts,
    notes: c.notes,
  }
  showAddModal.value = true
}

function openViewModal(c: Customer) {
  viewingCustomer.value = c
  showViewModal.value = true
}

async function saveCustomer() {
  try {
    if (editingCustomer.value) {
      await customerStore.updateCustomer(editingCustomer.value.id, form.value)
    } else {
      await customerStore.addCustomer(form.value)
    }
    showAddModal.value = false
  } catch (e) {
    console.error('Failed to save customer', e)
  }
}

async function deleteCustomer(id: string) {
  try {
    await customerStore.deleteCustomer(id)
  } catch (e) {
    console.error('Failed to delete customer', e)
  }
}
</script>

<template>
  <div class="customers-view">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Customers</h1>
        <p class="page-header-subtitle">{{ filteredCustomers.length }} of {{ customers.length }} customers</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <Plus :size="16" />
        Add Customer
      </button>
    </div>

    <!-- Filter Toolbar -->
    <div class="card mb-6">
      <div class="card-body toolbar">
        <div class="search-input">
          <Search :size="16" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Search by company name…"
          />
        </div>
        <select v-model="sectorFilter" class="form-select">
          <option value="all">All Sectors</option>
          <option v-for="s in sectorOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
        <select v-model="statusFilter" class="form-select">
          <option value="all">All Statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="prospect">Prospect</option>
        </select>
        <select v-model="typeFilter" class="form-select">
          <option value="all">All Types</option>
          <option value="get">Get</option>
          <option value="grow">Grow</option>
        </select>
      </div>
    </div>

    <!-- Customer Table -->
    <div class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Company Name</th>
            <th>Sector</th>
            <th>Region</th>
            <th>Status</th>
            <th>Type</th>
            <th class="text-center">Contacts</th>
            <th class="text-center">Sites</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in filteredCustomers" :key="c.id">
            <td>
              <button class="company-link" @click="openViewModal(c)">{{ c.companyName }}</button>
            </td>
            <td><span class="badge badge-neutral">{{ sectorLabel(c.sector) }}</span></td>
            <td>{{ c.region }}</td>
            <td>
              <span class="badge badge-dot" :class="statusBadgeClass(c.status)">{{ c.status }}</span>
            </td>
            <td>
              <span class="badge" :class="typeBadgeClass(c.type)">{{ c.type.toUpperCase() }}</span>
            </td>
            <td class="text-center">{{ c.contacts.length }}</td>
            <td class="text-center">{{ c.sites.length }}</td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="View" @click="openViewModal(c)">
                  <Eye :size="15" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEditModal(c)">
                  <Pencil :size="15" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm text-danger" title="Delete" @click="deleteCustomer(c.id)">
                  <Trash2 :size="15" />
                </button>
              </div>
            </td>
          </tr>
          <tr v-if="filteredCustomers.length === 0">
            <td colspan="8">
              <div class="table-empty">
                <div class="empty-icon"><Building2 :size="40" /></div>
                <div class="empty-text">No customers found</div>
                <div class="empty-subtext">Try adjusting your search or filters.</div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Modal -->
    <Teleport to="body">
      <div v-if="showAddModal" class="modal-backdrop" @click.self="showAddModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h3 class="modal-title">{{ editingCustomer ? 'Edit Customer' : 'Add Customer' }}</h3>
            <button class="modal-close" @click="showAddModal = false"><X :size="18" /></button>
          </div>
          <div class="modal-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Company Name <span class="required">*</span></label>
                <input v-model="form.companyName" type="text" class="form-input" placeholder="Enter company name" />
              </div>
              <div class="form-group">
                <label class="form-label">Sector</label>
                <select v-model="form.sector" class="form-select">
                  <option v-for="s in sectorOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Region</label>
                <input v-model="form.region" type="text" class="form-input" placeholder="e.g. Riyadh" />
              </div>
              <div class="form-group">
                <label class="form-label">Status</label>
                <select v-model="form.status" class="form-select">
                  <option value="prospect">Prospect</option>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Type</label>
                <select v-model="form.type" class="form-select">
                  <option value="get">Get</option>
                  <option value="grow">Grow</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">CR Number</label>
                <input v-model="form.crNumber" type="text" class="form-input" placeholder="Commercial Registration" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">VAT Number</label>
              <input v-model="form.vatNumber" type="text" class="form-input" placeholder="VAT Registration Number" />
            </div>
            <div class="form-group">
              <label class="form-label">Notes</label>
              <textarea v-model="form.notes" class="form-textarea" placeholder="Additional notes…" rows="3" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showAddModal = false">Cancel</button>
            <button class="btn btn-primary" :disabled="!form.companyName" @click="saveCustomer">
              {{ editingCustomer ? 'Save Changes' : 'Add Customer' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- View Customer Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && viewingCustomer" class="modal-backdrop" @click.self="showViewModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h3 class="modal-title">{{ viewingCustomer.companyName }}</h3>
            <button class="modal-close" @click="showViewModal = false"><X :size="18" /></button>
          </div>
          <div class="modal-body">
            <!-- Top meta -->
            <div class="view-meta-row">
              <span class="badge badge-dot" :class="statusBadgeClass(viewingCustomer.status)">{{ viewingCustomer.status }}</span>
              <span class="badge" :class="typeBadgeClass(viewingCustomer.type)">{{ viewingCustomer.type.toUpperCase() }}</span>
              <span class="badge badge-neutral">{{ sectorLabel(viewingCustomer.sector) }}</span>
            </div>

            <!-- Details grid -->
            <div class="view-grid">
              <div class="view-field">
                <span class="view-label"><MapPin :size="14" /> Region</span>
                <span class="view-value">{{ viewingCustomer.region }}</span>
              </div>
              <div class="view-field">
                <span class="view-label"><FileText :size="14" /> CR Number</span>
                <span class="view-value text-mono">{{ viewingCustomer.crNumber || '—' }}</span>
              </div>
              <div class="view-field">
                <span class="view-label"><FileText :size="14" /> VAT Number</span>
                <span class="view-value text-mono">{{ viewingCustomer.vatNumber || '—' }}</span>
              </div>
            </div>

            <!-- Notes -->
            <div v-if="viewingCustomer.notes" class="view-notes">
              <p class="view-section-title">Notes</p>
              <p class="view-notes-text">{{ viewingCustomer.notes }}</p>
            </div>

            <!-- Contacts -->
            <div class="view-section">
              <p class="view-section-title"><Users :size="15" /> Contacts ({{ viewingCustomer.contacts.length }})</p>
              <div v-if="viewingCustomer.contacts.length" class="view-contacts-list">
                <div v-for="ct in viewingCustomer.contacts" :key="ct.id" class="view-contact-card">
                  <div class="view-contact-name">
                    {{ ct.name }}
                    <span v-if="ct.isPrimary" class="badge badge-primary">Primary</span>
                  </div>
                  <div class="view-contact-detail"><Mail :size="13" /> {{ ct.email }}</div>
                  <div class="view-contact-detail"><Phone :size="13" /> {{ ct.phone }}</div>
                  <div class="view-contact-detail text-muted">{{ ct.position }}</div>
                </div>
              </div>
              <p v-else class="text-muted text-sm">No contacts recorded.</p>
            </div>

            <!-- Sites -->
            <div class="view-section">
              <p class="view-section-title"><Building2 :size="15" /> Sites ({{ viewingCustomer.sites.length }})</p>
              <div v-if="viewingCustomer.sites.length" class="view-sites-list">
                <div v-for="site in viewingCustomer.sites" :key="site.id" class="view-site-card">
                  <div class="font-semibold">{{ site.name }}</div>
                  <div class="text-sm text-muted">{{ site.address }}, {{ site.city }} — {{ site.region }}</div>
                </div>
              </div>
              <p v-else class="text-muted text-sm">No sites recorded.</p>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showViewModal = false">Close</button>
            <button class="btn btn-primary" @click="showViewModal = false; openEditModal(viewingCustomer!)">Edit</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.customers-view {
  padding: var(--space-6);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.toolbar .search-input {
  flex: 1;
  min-width: 220px;
}

.toolbar .form-select {
  width: auto;
  min-width: 150px;
}

.company-link {
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
.company-link:hover {
  color: var(--color-primary-hover);
  text-decoration: underline;
}

/* View modal styles */
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
  display: flex;
  align-items: center;
  gap: var(--space-1);
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

.view-notes {
  margin-bottom: var(--space-5);
  padding: var(--space-4);
  background-color: var(--color-warning-light);
  border-radius: var(--radius-lg);
  border-left: 3px solid var(--color-warning);
}
.view-notes-text {
  font-size: var(--text-sm);
  color: var(--color-neutral-700);
  line-height: var(--leading-relaxed);
}

.view-section {
  margin-bottom: var(--space-5);
}

.view-section-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-neutral-200);
}

.view-contacts-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--space-3);
}

.view-contact-card {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  background-color: var(--content-surface);
}

.view-contact-name {
  font-weight: var(--font-semibold);
  font-size: var(--text-sm);
  color: var(--color-neutral-900);
  margin-bottom: var(--space-2);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.view-contact-detail {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-neutral-600);
  margin-bottom: 2px;
}

.view-sites-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.view-site-card {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  background-color: var(--content-surface);
}

@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
  }
  .toolbar .search-input,
  .toolbar .form-select {
    width: 100%;
    min-width: 0;
  }
  .view-grid {
    grid-template-columns: 1fr;
  }
}
</style>
