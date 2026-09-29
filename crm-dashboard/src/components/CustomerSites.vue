<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Plus, Pencil, Trash2, MapPin } from 'lucide-vue-next'
import { customersService } from '@/services'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import { KSA_CITIES, KSA_REGIONS, regionForCity } from '@/utils/ksa'
import type { CustomerSite, SiteType } from '@/types'

const props = defineProps<{ customerId: string }>()
const emit = defineEmits<{ (e: 'changed', sites: CustomerSite[]): void }>()

const auth = useAuthStore()
const canEdit = computed(() => auth.can('customers:update'))

const siteTypes: { value: SiteType; label: string }[] = [
  { value: 'head-office', label: 'Head office' },
  { value: 'branch', label: 'Branch' },
  { value: 'warehouse', label: 'Warehouse' },
  { value: 'plant', label: 'Plant' },
  { value: 'retail', label: 'Retail outlet' },
  { value: 'other', label: 'Other' },
]
const typeLabel = (t: SiteType) => siteTypes.find((o) => o.value === t)?.label ?? t

const sites = ref<CustomerSite[]>([])
const loading = ref(false)
const error = ref('')
const saving = ref(false)
const editingId = ref<string | null>(null)
const showForm = ref(false)

const blank = () => ({
  name: '', code: '', siteType: 'branch' as SiteType, address: '', city: '', region: '',
  contactName: '', contactPhone: '', isActive: true,
})
const form = ref(blank())

async function load() {
  if (!props.customerId) return
  loading.value = true
  error.value = ''
  try {
    const res = await customersService.sites(props.customerId)
    sites.value = res.data ?? []
    emit('changed', sites.value)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
watch(() => props.customerId, load, { immediate: true })

function openAdd() {
  editingId.value = null
  form.value = blank()
  showForm.value = true
}

function openEdit(site: CustomerSite) {
  editingId.value = site.id
  form.value = {
    name: site.name, code: site.code ?? '', siteType: site.siteType, address: site.address ?? '',
    city: site.city ?? '', region: site.region ?? '', contactName: site.contactName ?? '',
    contactPhone: site.contactPhone ?? '', isActive: site.isActive,
  }
  showForm.value = true
}

function onCityInput() {
  const region = regionForCity(form.value.city)
  if (region) form.value.region = region
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    if (editingId.value) await customersService.updateSite(props.customerId, editingId.value, form.value)
    else await customersService.addSite(props.customerId, form.value)
    showForm.value = false
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}

async function toggleActive(site: CustomerSite) {
  error.value = ''
  try {
    await customersService.updateSite(props.customerId, site.id, { isActive: !site.isActive })
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function remove(site: CustomerSite) {
  if (!window.confirm(`Delete site "${site.name}"? A site used by an opportunity or quote line cannot be deleted; deactivate it instead.`)) return
  error.value = ''
  try {
    await customersService.deleteSite(props.customerId, site.id)
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  }
}
</script>

<template>
  <div class="sites">
    <p v-if="error" class="sites-error" role="alert">{{ error }}</p>

    <div class="sites-head">
      <span class="text-sm text-muted">Branches and locations. An opportunity chooses from these.</span>
      <button v-if="canEdit && !showForm" class="btn btn-primary btn-sm" @click="openAdd"><Plus :size="14" /> Add site</button>
    </div>

    <form v-if="showForm" class="sites-form" @submit.prevent="save">
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Site name <span class="required">*</span></label>
          <input v-model="form.name" class="form-input" required maxlength="255" placeholder="e.g. Jeddah branch" />
        </div>
        <div class="form-group">
          <label class="form-label">Code</label>
          <input v-model="form.code" class="form-input" maxlength="50" placeholder="e.g. JED-01" />
        </div>
        <div class="form-group">
          <label class="form-label">Type</label>
          <select v-model="form.siteType" class="form-select">
            <option v-for="t in siteTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">City</label>
          <input v-model="form.city" class="form-input" list="ksa-cities" maxlength="100" placeholder="Choose or type a city" @input="onCityInput" />
          <datalist id="ksa-cities">
            <option v-for="city in Object.keys(KSA_CITIES)" :key="city" :value="city" />
          </datalist>
        </div>
        <div class="form-group">
          <label class="form-label">Region</label>
          <select v-model="form.region" class="form-select">
            <option value="">—</option>
            <option v-for="r in KSA_REGIONS" :key="r" :value="r">{{ r }}</option>
            <option v-if="form.region && !KSA_REGIONS.includes(form.region as never)" :value="form.region">{{ form.region }}</option>
          </select>
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">Address</label>
        <input v-model="form.address" class="form-input" maxlength="2000" />
      </div>
      <div class="form-row">
        <div class="form-group">
          <label class="form-label">Site contact</label>
          <input v-model="form.contactName" class="form-input" maxlength="255" />
        </div>
        <div class="form-group">
          <label class="form-label">Contact phone</label>
          <input v-model="form.contactPhone" class="form-input" maxlength="50" />
        </div>
      </div>
      <label class="sites-check"><input v-model="form.isActive" type="checkbox" /> Active</label>
      <div class="sites-actions">
        <button type="button" class="btn btn-secondary btn-sm" @click="showForm = false">Cancel</button>
        <button class="btn btn-primary btn-sm" :disabled="saving || !form.name.trim()">{{ saving ? 'Saving…' : editingId ? 'Save site' : 'Add site' }}</button>
      </div>
    </form>

    <p v-if="loading" class="text-muted text-sm">Loading sites…</p>
    <div v-else-if="sites.length" class="table-container">
      <table class="table table-compact">
        <thead>
          <tr><th>Site</th><th>Type</th><th>City / region</th><th>Contact</th><th>Status</th><th v-if="canEdit">Actions</th></tr>
        </thead>
        <tbody>
          <tr v-for="site in sites" :key="site.id" :class="{ 'site-inactive': !site.isActive }">
            <td>
              <div class="font-semibold">{{ site.name }}</div>
              <div v-if="site.code" class="text-xs text-muted text-mono">{{ site.code }}</div>
            </td>
            <td>{{ typeLabel(site.siteType) }}</td>
            <td>{{ [site.city, site.region].filter(Boolean).join(' · ') || '—' }}</td>
            <td>
              <div>{{ site.contactName || '—' }}</div>
              <div v-if="site.contactPhone" class="text-xs text-muted">{{ site.contactPhone }}</div>
            </td>
            <td><span class="badge badge-dot" :class="site.isActive ? 'badge-success' : 'badge-neutral'">{{ site.isActive ? 'active' : 'inactive' }}</span></td>
            <td v-if="canEdit">
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEdit(site)"><Pencil :size="14" /></button>
                <button class="btn btn-ghost btn-sm" @click="toggleActive(site)">{{ site.isActive ? 'Deactivate' : 'Activate' }}</button>
                <button class="btn btn-ghost btn-icon btn-sm text-danger" title="Delete" @click="remove(site)"><Trash2 :size="14" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else-if="!showForm" class="table-empty">
      <div class="empty-icon"><MapPin :size="32" /></div>
      <div class="empty-text">No sites yet</div>
      <div class="empty-subtext">Add the customer's branches so opportunities can name where the work happens.</div>
    </div>
  </div>
</template>

<style scoped>
.sites-head { display: flex; align-items: center; justify-content: space-between; gap: var(--space-3); margin-bottom: var(--space-3); }
.sites-form { padding: var(--space-4); margin-bottom: var(--space-4); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); background: var(--color-neutral-50); }
.sites-check { display: inline-flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); }
.sites-actions { display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-3); }
.sites-error { color: var(--color-danger); font-size: var(--text-sm); margin-bottom: var(--space-3); }
.site-inactive { opacity: 0.6; }
</style>
