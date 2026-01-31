<template>
  <div class="customers-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search companies or contacts..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="showAddModal">
          <span>➕</span> Add {{ activeTab === 'accounts' ? 'Company' : 'Contact' }}
        </button>
      </div>
    </div>

    <div class="tabs">
      <button
        class="tab"
        :class="{ active: activeTab === 'accounts' }"
        @click="activeTab = 'accounts'"
      >
        <span class="tab-icon">🏢</span>
        Accounts ({{ customersStore.customers.length }})
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'contacts' }"
        @click="activeTab = 'contacts'"
      >
        <span class="tab-icon">👤</span>
        Contacts ({{ contactsStore.contacts.length }})
      </button>
    </div>

    <!-- Accounts Tab -->
    <BaseCard v-if="activeTab === 'accounts'" no-padding>
      <BaseTable
        :columns="accountsColumns"
        :data="filteredCustomers"
        :clickable="true"
        @row-click="viewCustomer"
      >
        <template #cell-name="{ row }">
          <div class="customer-name">
            <span class="avatar">{{ row.name.charAt(0) }}</span>
            <div>
              <div class="name-text">{{ row.name }}</div>
              <div class="name-subtitle">{{ row.email }}</div>
            </div>
          </div>
        </template>

        <template #cell-sector="{ value }">
          <BaseBadge variant="info" size="sm">{{ value }}</BaseBadge>
        </template>

        <template #cell-regions="{ value }">
          <div class="regions-cell">
            <BaseBadge v-for="region in value" :key="region" variant="default" size="sm" style="margin: 2px;">
              {{ region }}
            </BaseBadge>
            <div v-if="value.length === 0" class="text-muted">-</div>
          </div>
        </template>

        <template #cell-status="{ value }">
          <BaseBadge :variant="value === 'get' ? 'success' : 'info'" size="sm">
            {{ value === 'get' ? 'Get (New)' : 'Grow (Expand)' }}
          </BaseBadge>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="View" @click.stop="viewCustomer(row)">👁️</button>
            <button class="action-btn" title="Edit" @click.stop="editCustomer(row)">✏️</button>
            <button class="action-btn" title="History" @click.stop="viewHistory(row)">📊</button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>

    <!-- Contacts Tab -->
    <BaseCard v-if="activeTab === 'contacts'" no-padding>
      <BaseTable
        :columns="contactsColumns"
        :data="filteredContacts"
        :clickable="true"
        @row-click="viewContact"
      >
        <template #cell-fullName="{ row }">
          <div class="customer-name">
            <span class="avatar">{{ row.firstName.charAt(0) }}</span>
            <div>
              <div class="name-text">{{ row.fullName }}</div>
              <div class="name-subtitle">{{ getCompanyName(row.companyId) }}</div>
            </div>
          </div>
        </template>

        <template #cell-title="{ row }">
          <div>
            <div>{{ row.title }}</div>
            <div class="contact-detail">{{ row.department }}</div>
          </div>
        </template>

        <template #cell-email="{ row }">
          <div>
            <div>{{ row.email }}</div>
            <div class="contact-detail">{{ row.phone }}</div>
          </div>
        </template>

        <template #cell-isPrimary="{ value }">
          <BaseBadge :variant="value ? 'success' : 'default'" size="sm">
            {{ value ? 'Primary' : 'Secondary' }}
          </BaseBadge>
        </template>

        <template #cell-status="{ value }">
          <BaseBadge :variant="value === 'active' ? 'success' : 'default'" size="sm">
            {{ value }}
          </BaseBadge>
        </template>

        <template #cell-actions="{ row }">
          <div class="action-buttons">
            <button class="action-btn" title="Email" @click.stop="emailContact(row)">✉️</button>
            <button class="action-btn" title="Edit" @click.stop="editContact(row)">✏️</button>
            <button class="action-btn" title="Delete" @click.stop="deleteContact(row)">🗑️</button>
          </div>
        </template>
      </BaseTable>
    </BaseCard>
  </div>

  <!-- Add/Edit Company Modal -->
  <BaseModal
    v-model="showCompanyModal"
    :title="editingCompany ? 'Edit Company' : 'Add New Company'"
    size="lg"
    @confirm="saveCompany"
  >
    <div class="modal-form">
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Company Name *</label>
          <input v-model="companyForm.name" type="text" class="form-input" placeholder="Enter company name" />
        </div>
        <div class="form-group">
          <label class="form-label">Email *</label>
          <input v-model="companyForm.email" type="email" class="form-input" placeholder="company@example.com" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Sector *</label>
          <select v-model="companyForm.sector" class="form-input">
            <option value="">Select sector</option>
            <option value="Banking">Banking</option>
            <option value="Healthcare">Healthcare</option>
            <option value="Retail">Retail</option>
            <option value="Manufacturing">Manufacturing</option>
            <option value="Real Estate">Real Estate</option>
            <option value="Government">Government</option>
            <option value="Education">Education</option>
            <option value="Technology">Technology</option>
            <option value="Other">Other</option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">Customer Type *</label>
          <select v-model="companyForm.status" class="form-input">
            <option value="get">Get (New Customer)</option>
            <option value="grow">Grow (Existing to Expand)</option>
          </select>
        </div>
      </div>

      <!-- Regions Selection -->
      <div class="form-group">
        <label class="form-label">Regions * <small>({{ companyForm.regions.length }} selected)</small></label>
        <div class="checkbox-grid">
          <label
            v-for="region in ksaRegionNames"
            :key="region"
            class="checkbox-item"
          >
            <input
              type="checkbox"
              :value="region"
              v-model="companyForm.regions"
              @change="onRegionChange(region, $event)"
            />
            <span>{{ region }}</span>
          </label>
        </div>
      </div>

      <!-- Cities Selection -->
      <div class="form-group" v-if="companyForm.regions.length > 0">
        <label class="form-label">
          Cities * <small>({{ companyForm.cities.length }} selected)</small>
          <button
            type="button"
            class="btn-link"
            @click="selectAllCities"
            style="margin-left: 10px;"
          >
            Select All Cities
          </button>
          <button
            type="button"
            class="btn-link"
            @click="clearAllCities"
            style="margin-left: 5px;"
          >
            Clear All
          </button>
        </label>
        <div class="cities-selection">
          <div v-for="region in selectedRegionsData" :key="region.name" class="region-cities-group">
            <div class="region-header">
              <strong>{{ region.name }}</strong>
              <button
                type="button"
                class="btn-link-sm"
                @click="toggleRegionCities(region.name)"
              >
                {{ areAllRegionCitiesSelected(region.name) ? 'Deselect All' : 'Select All' }}
              </button>
            </div>
            <div class="checkbox-grid">
              <label
                v-for="city in region.cities"
                :key="city"
                class="checkbox-item"
              >
                <input
                  type="checkbox"
                  :value="city"
                  v-model="companyForm.cities"
                />
                <span>{{ city }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="form-group">
        <small class="form-hint-warning">Please select at least one region to choose cities</small>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Phone</label>
          <input v-model="companyForm.phone" type="tel" class="form-input" placeholder="+966 XX XXX XXXX" />
        </div>
        <div class="form-group">
          <label class="form-label">Website</label>
          <input v-model="companyForm.website" type="url" class="form-input" placeholder="https://example.com" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">CR Number</label>
          <input v-model="companyForm.crNumber" type="text" class="form-input" placeholder="Commercial Registration Number" />
        </div>
        <div class="form-group">
          <label class="form-label">VAT Number</label>
          <input v-model="companyForm.vatNumber" type="text" class="form-input" placeholder="VAT Registration Number" />
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">Attach File</label>
        <input type="file" class="form-input" @change="onFileSelect" accept=".pdf,.doc,.docx,.xls,.xlsx,.jpg,.jpeg,.png" />
        <small v-if="companyForm.attachedFile" class="form-hint">
          Selected: {{ companyForm.attachedFile.name }}
        </small>
      </div>

      <div class="form-group">
        <label class="form-label">Address</label>
        <textarea v-model="companyForm.address" class="form-input" rows="3" placeholder="Full address"></textarea>
      </div>
    </div>
  </BaseModal>

  <!-- Add/Edit Contact Modal -->
  <BaseModal
    v-model="showContactModal"
    :title="editingContact ? 'Edit Contact' : 'Add New Contact'"
    size="lg"
    @confirm="saveContact"
  >
    <div class="modal-form">
      <div class="form-group">
        <label class="form-label">Company *</label>
        <select v-model="contactForm.companyId" class="form-input">
          <option value="">Select company</option>
          <option v-for="company in customersStore.customers" :key="company.id" :value="company.id">
            {{ company.name }}
          </option>
        </select>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">First Name *</label>
          <input v-model="contactForm.firstName" type="text" class="form-input" placeholder="First name" />
        </div>
        <div class="form-group">
          <label class="form-label">Last Name *</label>
          <input v-model="contactForm.lastName" type="text" class="form-input" placeholder="Last name" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Email *</label>
          <input v-model="contactForm.email" type="email" class="form-input" placeholder="email@example.com" />
        </div>
        <div class="form-group">
          <label class="form-label">Phone *</label>
          <input v-model="contactForm.phone" type="tel" class="form-input" placeholder="+966 XX XXX XXXX" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Title *</label>
          <input v-model="contactForm.title" type="text" class="form-input" placeholder="Job title" />
        </div>
        <div class="form-group">
          <label class="form-label">Department</label>
          <input v-model="contactForm.department" type="text" class="form-input" placeholder="Department" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Mobile</label>
          <input v-model="contactForm.mobile" type="tel" class="form-input" placeholder="+966 XX XXX XXXX" />
        </div>
        <div class="form-group">
          <label class="form-label">Contact Type *</label>
          <select v-model="contactForm.isPrimary" class="form-input">
            <option :value="true">Primary</option>
            <option :value="false">Secondary</option>
          </select>
        </div>
      </div>

      <div class="form-group">
        <label class="form-label">Status *</label>
        <select v-model="contactForm.status" class="form-input">
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
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
      <p class="delete-message">Are you sure you want to delete <strong>{{ deleteItem?.name || deleteItem?.fullName }}</strong>?</p>
      <p class="delete-warning">This action cannot be undone.</p>
    </div>
  </BaseModal>

  <!-- View Detail Modal -->
  <BaseModal
    v-model="showDetailModal"
    :title="detailTitle"
    size="lg"
    hide-footer
  >
    <div class="detail-view" v-if="viewingItem">
      <div class="detail-section" v-if="activeTab === 'accounts'">
        <div class="detail-row">
          <div class="detail-label">Company Name:</div>
          <div class="detail-value">{{ viewingItem.name }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Email:</div>
          <div class="detail-value">{{ viewingItem.email }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Phone:</div>
          <div class="detail-value">{{ viewingItem.phone || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Website:</div>
          <div class="detail-value">{{ viewingItem.website || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Sector:</div>
          <div class="detail-value">
            <BaseBadge variant="info">{{ viewingItem.sector }}</BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Region:</div>
          <div class="detail-value">
            <BaseBadge variant="default">{{ viewingItem.region }}</BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">CR Number:</div>
          <div class="detail-value">{{ viewingItem.crNumber || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">VAT Number:</div>
          <div class="detail-value">{{ viewingItem.vatNumber || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">SLA Hours:</div>
          <div class="detail-value">{{ viewingItem.slaHours }}h</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Status:</div>
          <div class="detail-value">
            <BaseBadge :variant="viewingItem.status === 'active' ? 'success' : 'default'">
              {{ viewingItem.status }}
            </BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Address:</div>
          <div class="detail-value">{{ viewingItem.address || 'N/A' }}</div>
        </div>
      </div>

      <div class="detail-section" v-if="activeTab === 'sites'">
        <div class="detail-row">
          <div class="detail-label">Site Name:</div>
          <div class="detail-value">{{ viewingItem.name }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Company:</div>
          <div class="detail-value">{{ getCompanyName(viewingItem.companyId) }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Type:</div>
          <div class="detail-value">
            <BaseBadge :variant="viewingItem.isPrimary ? 'success' : 'default'">
              {{ viewingItem.isPrimary ? 'Primary' : 'Branch' }}
            </BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Region:</div>
          <div class="detail-value">
            <BaseBadge variant="info">{{ viewingItem.region }}</BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">City:</div>
          <div class="detail-value">{{ viewingItem.city }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Address:</div>
          <div class="detail-value">{{ viewingItem.address }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Contact Person:</div>
          <div class="detail-value">{{ viewingItem.contactPerson || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Phone:</div>
          <div class="detail-value">{{ viewingItem.phone || 'N/A' }}</div>
        </div>
      </div>

      <div class="detail-section" v-if="activeTab === 'contacts'">
        <div class="detail-row">
          <div class="detail-label">Full Name:</div>
          <div class="detail-value">{{ viewingItem.fullName }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Company:</div>
          <div class="detail-value">{{ getCompanyName(viewingItem.companyId) }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Title:</div>
          <div class="detail-value">{{ viewingItem.title }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Department:</div>
          <div class="detail-value">{{ viewingItem.department }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Email:</div>
          <div class="detail-value">{{ viewingItem.email }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Phone:</div>
          <div class="detail-value">{{ viewingItem.phone }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Mobile:</div>
          <div class="detail-value">{{ viewingItem.mobile || 'N/A' }}</div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Contact Type:</div>
          <div class="detail-value">
            <BaseBadge :variant="viewingItem.isPrimary ? 'success' : 'default'">
              {{ viewingItem.isPrimary ? 'Primary' : 'Secondary' }}
            </BaseBadge>
          </div>
        </div>
        <div class="detail-row">
          <div class="detail-label">Status:</div>
          <div class="detail-value">
            <BaseBadge :variant="viewingItem.status === 'active' ? 'success' : 'default'">
              {{ viewingItem.status }}
            </BaseBadge>
          </div>
        </div>
      </div>

      <div class="detail-actions">
        <button class="btn btn-primary" @click="editFromDetail">
          Edit {{ activeTab === 'accounts' ? 'Company' : activeTab === 'sites' ? 'Site' : 'Contact' }}
        </button>
      </div>
    </div>
  </BaseModal>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useCustomersStore } from '@/stores/customers'
import { useContactsStore } from '@/stores/contacts'
import { useToast } from '@/composables/useToast'
import { getAllRegionNames, getCitiesForRegions } from '@/constants/ksaLocations'
import BaseCard from '@/components/UI/BaseCard.vue'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import type { TableColumn } from '@/components/UI/BaseTable.vue'

const customersStore = useCustomersStore()
const contactsStore = useContactsStore()
const { success, info } = useToast()

const activeTab = ref<'accounts' | 'contacts'>('accounts')
const searchQuery = ref('')

// Modal states
const showCompanyModal = ref(false)
const showSiteModal = ref(false)
const showContactModal = ref(false)
const showDeleteModal = ref(false)
const showDetailModal = ref(false)

// Editing states
const editingCompany = ref<any>(null)
const editingSite = ref<any>(null)
const editingContact = ref<any>(null)
const deleteItem = ref<any>(null)
const viewingItem = ref<any>(null)

// Form data
const companyForm = ref({
  name: '',
  email: '',
  phone: '',
  website: '',
  sector: '',
  regions: [] as string[],
  cities: [] as string[],
  crNumber: '',
  vatNumber: '',
  status: 'get' as 'get' | 'grow',
  address: '',
  attachedFile: null as File | null
})

const siteForm = ref({
  companyId: '',
  name: '',
  isPrimary: false,
  region: '',
  city: '',
  address: '',
  contactPerson: '',
  phone: ''
})

const contactForm = ref({
  companyId: '',
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  mobile: '',
  title: '',
  department: '',
  isPrimary: false,
  status: 'active' as 'active' | 'inactive'
})

// Computed
const detailTitle = computed(() => {
  if (!viewingItem.value) return ''
  if (activeTab.value === 'accounts') return viewingItem.value.name
  if (activeTab.value === 'sites') return viewingItem.value.name
  return viewingItem.value.fullName
})

// Table columns
const accountsColumns: TableColumn[] = [
  { key: 'name', label: 'Company Name', sortable: true },
  { key: 'sector', label: 'Sector', sortable: true },
  { key: 'regions', label: 'Regions', sortable: false },
  { key: 'status', label: 'Type', sortable: true },
  { key: 'actions', label: 'Actions', align: 'right' }
]

const sitesColumns: TableColumn[] = [
  { key: 'name', label: 'Site Name', sortable: true },
  { key: 'isPrimary', label: 'Type', sortable: true },
  { key: 'region', label: 'Region', sortable: true },
  { key: 'city', label: 'City', sortable: true },
  { key: 'contactPerson', label: 'Contact', sortable: true },
  { key: 'actions', label: 'Actions', align: 'right' }
]

const contactsColumns: TableColumn[] = [
  { key: 'fullName', label: 'Name', sortable: true },
  { key: 'title', label: 'Title & Department', sortable: true },
  { key: 'email', label: 'Contact Info', sortable: true },
  { key: 'isPrimary', label: 'Type', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'actions', label: 'Actions', align: 'right' }
]

// KSA Regions and Cities
import { KSA_REGIONS } from '@/constants/ksaLocations'

const ksaRegionNames = getAllRegionNames()

// Get region data with cities for selected regions
const selectedRegionsData = computed(() => {
  return KSA_REGIONS.filter(region => companyForm.value.regions.includes(region.name))
})

const availableCities = computed(() => {
  if (companyForm.value.regions.length === 0) {
    return []
  }
  return getCitiesForRegions(companyForm.value.regions)
})

// Region and City selection helpers
const onRegionChange = (region: string, event: Event) => {
  const isChecked = (event.target as HTMLInputElement).checked

  if (isChecked) {
    // When a region is checked, automatically select all its cities
    const regionData = KSA_REGIONS.find(r => r.name === region)
    if (regionData) {
      // Add all cities from this region
      regionData.cities.forEach(city => {
        if (!companyForm.value.cities.includes(city)) {
          companyForm.value.cities.push(city)
        }
      })
    }
  } else {
    // When a region is unchecked, remove all its cities
    const regionData = KSA_REGIONS.find(r => r.name === region)
    if (regionData) {
      companyForm.value.cities = companyForm.value.cities.filter(
        city => !regionData.cities.includes(city)
      )
    }
  }
}

const selectAllCities = () => {
  companyForm.value.cities = getCitiesForRegions(companyForm.value.regions)
}

const clearAllCities = () => {
  companyForm.value.cities = []
}

const toggleRegionCities = (regionName: string) => {
  const regionData = KSA_REGIONS.find(r => r.name === regionName)

  if (!regionData) return

  const allSelected = regionData.cities.every(city =>
    companyForm.value.cities.includes(city)
  )

  if (allSelected) {
    // Deselect all cities from this region
    companyForm.value.cities = companyForm.value.cities.filter(
      city => !regionData.cities.includes(city)
    )
  } else {
    // Select all cities from this region
    regionData.cities.forEach(city => {
      if (!companyForm.value.cities.includes(city)) {
        companyForm.value.cities.push(city)
      }
    })
  }
}

const areAllRegionCitiesSelected = (regionName: string): boolean => {
  const regionData = KSA_REGIONS.find(r => r.name === regionName)

  if (!regionData) return false

  return regionData.cities.every(city =>
    companyForm.value.cities.includes(city)
  )
}

// Filtered data
const filteredCustomers = computed(() => {
  if (!searchQuery.value) return customersStore.customers

  const query = searchQuery.value.toLowerCase()
  return customersStore.customers.filter(c =>
    c.name.toLowerCase().includes(query) ||
    c.sector.toLowerCase().includes(query) ||
    c.region.toLowerCase().includes(query) ||
    c.email?.toLowerCase().includes(query)
  )
})

const filteredSites = computed(() => {
  if (!searchQuery.value) return customersStore.sites

  const query = searchQuery.value.toLowerCase()
  return customersStore.sites.filter(s =>
    s.name.toLowerCase().includes(query) ||
    s.city.toLowerCase().includes(query) ||
    s.address.toLowerCase().includes(query) ||
    getCompanyName(s.companyId).toLowerCase().includes(query)
  )
})

const filteredContacts = computed(() => {
  if (!searchQuery.value) return contactsStore.contacts

  const query = searchQuery.value.toLowerCase()
  return contactsStore.contacts.filter(c =>
    c.fullName.toLowerCase().includes(query) ||
    c.email.toLowerCase().includes(query) ||
    c.title.toLowerCase().includes(query) ||
    c.department.toLowerCase().includes(query) ||
    getCompanyName(c.companyId).toLowerCase().includes(query)
  )
})

// Helper functions
const getCompanyName = (companyId: string) => {
  const company = customersStore.getCustomerById(companyId)
  return company?.name || 'Unknown'
}

// Helper functions to reset forms
const resetCompanyForm = () => {
  companyForm.value = {
    name: '',
    email: '',
    phone: '',
    website: '',
    sector: '',
    regions: [],
    cities: [],
    crNumber: '',
    vatNumber: '',
    status: 'get',
    address: '',
    attachedFile: null
  }
}

const resetSiteForm = () => {
  siteForm.value = {
    companyId: '',
    name: '',
    isPrimary: false,
    region: '',
    city: '',
    address: '',
    contactPerson: '',
    phone: ''
  }
}

const resetContactForm = () => {
  contactForm.value = {
    companyId: '',
    firstName: '',
    lastName: '',
    email: '',
    phone: '',
    mobile: '',
    title: '',
    department: '',
    isPrimary: false,
    status: 'active'
  }
}

// File upload handler
const onFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    companyForm.value.attachedFile = file
  }
}

// Actions - Add/Edit Modal Handlers
const showAddModal = () => {
  if (activeTab.value === 'accounts') {
    editingCompany.value = null
    resetCompanyForm()
    showCompanyModal.value = true
  } else if (activeTab.value === 'contacts') {
    editingContact.value = null
    resetContactForm()
    showContactModal.value = true
  }
}

// Company CRUD
const viewCustomer = (customer: any) => {
  viewingItem.value = customer
  showDetailModal.value = true
}

const editCustomer = (customer: any) => {
  editingCompany.value = customer
  companyForm.value = {
    name: customer.name,
    email: customer.email,
    phone: customer.phone || '',
    website: customer.website || '',
    sector: customer.sector,
    regions: customer.regions || [],
    cities: customer.cities || [],
    crNumber: customer.crNumber || '',
    vatNumber: customer.vatNumber || '',
    status: customer.status,
    address: customer.address || '',
    attachedFile: null
  }
  showCompanyModal.value = true
}

const saveCompany = () => {
  // Validation
  if (!companyForm.value.name || !companyForm.value.email || !companyForm.value.sector || companyForm.value.regions.length === 0 || companyForm.value.cities.length === 0) {
    info('Please fill in all required fields')
    return
  }

  // Prepare company data with file info if file is attached
  const companyData: any = {
    ...companyForm.value
  }

  // If there's an attached file, create file metadata
  if (companyForm.value.attachedFile) {
    companyData.attachedFile = {
      id: `file-${Date.now()}`,
      name: companyForm.value.attachedFile.name,
      url: URL.createObjectURL(companyForm.value.attachedFile), // In real app, this would be uploaded to server
      uploadedAt: new Date().toISOString(),
      uploadedBy: 'USR-001' // Should come from auth
    }
  }

  // Remove the File object before saving (we only save metadata)
  delete companyData.attachedFile

  if (editingCompany.value) {
    // Update existing
    customersStore.updateCustomer(editingCompany.value.id, companyData)
    success('Company updated successfully')
  } else {
    // Create new
    customersStore.addCustomer(companyData)
    success('Company created successfully')
  }

  showCompanyModal.value = false
  editingCompany.value = null
  resetCompanyForm()
}

const viewHistory = (customer: any) => {
  const history = customersStore.getCustomerHistory(customer.id)
  if (history) {
    info(`${customer.name} - Total Revenue: ${history.totalRevenue.toLocaleString()} SAR`)
  }
}

// Site CRUD
const viewSite = (site: any) => {
  viewingItem.value = site
  showDetailModal.value = true
}

const editSite = (site: any) => {
  editingSite.value = site
  siteForm.value = {
    companyId: site.companyId,
    name: site.name,
    isPrimary: site.isPrimary,
    region: site.region,
    city: site.city,
    address: site.address,
    contactPerson: site.contactPerson || '',
    phone: site.phone || ''
  }
  showSiteModal.value = true
}

const saveSite = () => {
  // Validation
  if (!siteForm.value.companyId || !siteForm.value.name || !siteForm.value.region || !siteForm.value.city || !siteForm.value.address) {
    info('Please fill in all required fields')
    return
  }

  if (editingSite.value) {
    // Update existing
    customersStore.updateSite(editingSite.value.id, siteForm.value)
    success('Site updated successfully')
  } else {
    // Create new
    customersStore.addSite(siteForm.value)
    success('Site created successfully')
  }

  showSiteModal.value = false
  editingSite.value = null
  resetSiteForm()
}

const deleteSite = (site: any) => {
  deleteItem.value = site
  showDeleteModal.value = true
}

// Contact CRUD
const viewContact = (contact: any) => {
  viewingItem.value = contact
  showDetailModal.value = true
}

const editContact = (contact: any) => {
  editingContact.value = contact
  contactForm.value = {
    companyId: contact.companyId,
    firstName: contact.firstName,
    lastName: contact.lastName,
    email: contact.email,
    phone: contact.phone,
    mobile: contact.mobile || '',
    title: contact.title,
    department: contact.department,
    isPrimary: contact.isPrimary,
    status: contact.status
  }
  showContactModal.value = true
}

const saveContact = () => {
  // Validation
  if (!contactForm.value.companyId || !contactForm.value.firstName || !contactForm.value.lastName ||
      !contactForm.value.email || !contactForm.value.phone || !contactForm.value.title) {
    info('Please fill in all required fields')
    return
  }

  if (editingContact.value) {
    // Update existing
    contactsStore.updateContact(editingContact.value.id, contactForm.value)
    success('Contact updated successfully')
  } else {
    // Create new
    contactsStore.addContact(contactForm.value)
    success('Contact created successfully')
  }

  showContactModal.value = false
  editingContact.value = null
  resetContactForm()
}

const emailContact = (contact: any) => {
  window.location.href = `mailto:${contact.email}`
}

const deleteContact = (contact: any) => {
  deleteItem.value = contact
  showDeleteModal.value = true
}

// Delete confirmation
const confirmDelete = () => {
  if (!deleteItem.value) return

  if (activeTab.value === 'sites') {
    customersStore.deleteSite(deleteItem.value.id)
    success('Site deleted successfully')
  } else if (activeTab.value === 'contacts') {
    contactsStore.deleteContact(deleteItem.value.id)
    success('Contact deleted successfully')
  }

  showDeleteModal.value = false
  deleteItem.value = null
}

// Edit from detail view
const editFromDetail = () => {
  showDetailModal.value = false

  if (activeTab.value === 'accounts') {
    editCustomer(viewingItem.value)
  } else if (activeTab.value === 'sites') {
    editSite(viewingItem.value)
  } else if (activeTab.value === 'contacts') {
    editContact(viewingItem.value)
  }

  viewingItem.value = null
}
</script>

<style scoped>
.customers-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1rem;
  width: 100%;
}

.search-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
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

.tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 2px solid var(--color-border);
}

.tab {
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  font-weight: 500;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tab:hover {
  color: var(--color-primary);
  background: var(--color-surface-hover);
}

.tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-icon {
  font-size: 1.1rem;
}

.customer-name,
.site-name {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.1rem;
}

.site-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--color-primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
}

.name-text {
  font-weight: 500;
  color: var(--color-text-primary);
}

.name-subtitle {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  margin-top: 0.125rem;
}

.contact-detail {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  margin-top: 0.125rem;
}

.sla-badge {
  padding: 0.25rem 0.75rem;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
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

/* Modal Form Styles */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
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

.form-input::placeholder {
  color: var(--color-text-secondary);
}

textarea.form-input {
  resize: vertical;
  font-family: inherit;
}

/* Delete Modal Styles */
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

/* Detail View Styles */
.detail-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.detail-row {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-weight: 600;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.detail-value {
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.detail-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}

/* Checkbox Grid for Regions/Cities Selection */
.checkbox-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.5rem;
  max-height: 300px;
  overflow-y: auto;
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
  user-select: none;
}

.checkbox-item:hover {
  background: var(--color-surface-hover);
}

.checkbox-item input[type="checkbox"] {
  cursor: pointer;
  width: 16px;
  height: 16px;
}

.checkbox-item span {
  font-size: 0.9rem;
  color: var(--color-text-primary);
}

/* Cities Selection by Region */
.cities-selection {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-height: 400px;
  overflow-y: auto;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
}

.region-cities-group {
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 0.75rem;
  background: var(--color-background);
}

.region-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--color-border);
}

.region-header strong {
  color: var(--color-primary);
  font-size: 0.95rem;
}

.btn-link,
.btn-link-sm {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0.25rem 0.5rem;
  text-decoration: underline;
  transition: color 0.2s;
}

.btn-link:hover,
.btn-link-sm:hover {
  color: var(--color-primary-hover);
}

.btn-link-sm {
  font-size: 0.8rem;
}

.form-hint-warning {
  color: #f59e0b;
  font-size: 0.85rem;
  font-style: italic;
  display: block;
  padding: 0.75rem;
  background: rgba(245, 158, 11, 0.1);
  border-radius: 4px;
  border-left: 3px solid #f59e0b;
}

.regions-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  align-items: center;
}

.text-muted {
  color: var(--color-text-secondary);
  font-style: italic;
}
</style>
