<script setup lang="ts">
import { computed, ref } from 'vue'
import { Plus, Pencil, Trash2, Search, Check, X } from 'lucide-vue-next'
import { manufacturersService } from '@/services'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import { sameMaster } from '@/services/catalogMasters'

interface CategoryRef { id: string; name: string }
interface ManufacturerRef { id: string; name: string; code: string; categories: CategoryRef[] }

// Categories belong to a manufacturer; products, the import template and the
// import itself match them by name within that manufacturer.
const props = defineProps<{ manufacturers: ManufacturerRef[]; productCounts: Record<string, number> }>()
const emit = defineEmits<{ changed: [] }>()

const auth = useAuthStore()
const canEdit = computed(() => auth.can('manufacturers:update'))

const search = ref('')
const manufacturerFilter = ref('')
const error = ref('')
const notice = ref('')

const rows = computed(() => {
  const q = search.value.trim().toLowerCase()
  return props.manufacturers
    .filter(m => !manufacturerFilter.value || m.id === manufacturerFilter.value)
    .flatMap(m => m.categories.map(c => ({ manufacturer: m, category: c, products: props.productCounts[c.id] ?? 0 })))
    .filter(r => !q || r.category.name.toLowerCase().includes(q) || r.manufacturer.name.toLowerCase().includes(q))
    .sort((a, b) => a.manufacturer.name.localeCompare(b.manufacturer.name) || a.category.name.localeCompare(b.category.name))
})
const emptyManufacturers = computed(() => props.manufacturers.filter(m => !m.categories.length).map(m => m.name))

// ── Add: one or several names (one per line or comma-separated) ──
const adding = ref(false)
const addForm = ref({ manufacturerId: '', names: '', busy: false })
function openAdd() {
  addForm.value = { manufacturerId: manufacturerFilter.value, names: '', busy: false }
  error.value = ''
  adding.value = true
}
const newNames = computed(() => {
  const owner = props.manufacturers.find(m => m.id === addForm.value.manufacturerId)
  const seen: string[] = []
  for (const raw of addForm.value.names.split(/[\n,]/)) {
    const name = raw.trim().replace(/\s+/g, ' ')
    if (name && !seen.some(n => sameMaster(n, name))) seen.push(name)
  }
  return seen.map(name => ({ name, exists: !!owner?.categories.some(c => sameMaster(c.name, name)) }))
})
async function saveAdd() {
  const f = addForm.value
  const toCreate = newNames.value.filter(n => !n.exists)
  if (!f.manufacturerId || !toCreate.length || f.busy) return
  f.busy = true
  error.value = ''
  try {
    for (const { name } of toCreate) await manufacturersService.addCategory(f.manufacturerId, { name })
    notice.value = `${toCreate.length} categor${toCreate.length === 1 ? 'y' : 'ies'} added.`
    adding.value = false
    emit('changed')
  } catch (e) {
    error.value = errorMessage(e)
    emit('changed')
  } finally {
    f.busy = false
  }
}

// ── Rename ──
const editingId = ref('')
const editName = ref('')
function startEdit(category: CategoryRef) { editingId.value = category.id; editName.value = category.name; error.value = '' }
async function saveEdit(manufacturer: ManufacturerRef, category: CategoryRef) {
  const name = editName.value.trim().replace(/\s+/g, ' ')
  if (!name || name === category.name) { editingId.value = ''; return }
  if (manufacturer.categories.some(c => c.id !== category.id && sameMaster(c.name, name))) {
    error.value = `${manufacturer.name} already has a category named ${name}.`
    return
  }
  try {
    await manufacturersService.updateCategory(manufacturer.id, category.id, { name })
    editingId.value = ''
    emit('changed')
  } catch (e) {
    error.value = errorMessage(e)
  }
}

// ── Delete (only when no product uses it) ──
async function remove(manufacturer: ManufacturerRef, category: CategoryRef) {
  if (!window.confirm(`Delete the category "${category.name}" from ${manufacturer.name}?`)) return
  error.value = ''
  try {
    await manufacturersService.deleteCategory(manufacturer.id, category.id)
    emit('changed')
  } catch (e) {
    error.value = errorMessage(e)
  }
}
</script>

<template>
  <div class="cat-page">
    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search"><Search :size="18" class="search-icon" /><input v-model="search" type="text" class="form-input" placeholder="Search categories or manufacturers..." aria-label="Search categories" /></div>
        <select v-model="manufacturerFilter" class="form-select toolbar-select" aria-label="Filter by manufacturer">
          <option value="">All manufacturers</option>
          <option v-for="m in manufacturers" :key="m.id" :value="m.id">{{ m.name }}</option>
        </select>
        <button v-if="canEdit" class="btn btn-primary" @click="openAdd"><Plus :size="16" /> Add categories</button>
      </div>
    </div>

    <p v-if="error" class="cat-error" role="alert">{{ error }}</p>
    <p v-if="notice && !adding" class="cat-notice" role="status">{{ notice }}</p>

    <div v-if="adding" class="card cat-add">
      <div class="card-body">
        <h3 class="cat-add-title">Add categories</h3>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="cat-add-mfr">Manufacturer <span class="required">*</span></label>
            <select id="cat-add-mfr" v-model="addForm.manufacturerId" class="form-select">
              <option value="" disabled>Select manufacturer</option>
              <option v-for="m in manufacturers" :key="m.id" :value="m.id">{{ m.name }}</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label" for="cat-add-names">Category names <span class="required">*</span></label>
            <textarea id="cat-add-names" v-model="addForm.names" class="form-textarea" rows="3" placeholder="One per line, or separated by commas&#10;e.g. IP Cameras, Recorders, Accessories" />
          </div>
        </div>
        <p v-if="newNames.length" class="cat-preview">
          <span v-for="n in newNames" :key="n.name" class="badge" :class="n.exists ? 'badge-neutral' : 'badge-success'" :title="n.exists ? 'Already exists for this manufacturer' : 'Will be added'">{{ n.name }}{{ n.exists ? ' (exists)' : '' }}</span>
        </p>
        <div class="cat-add-actions">
          <button class="btn btn-secondary btn-sm" @click="adding = false">Cancel</button>
          <button class="btn btn-primary btn-sm" :disabled="!addForm.manufacturerId || !newNames.some(n => !n.exists) || addForm.busy" @click="saveAdd">
            {{ addForm.busy ? 'Adding…' : `Add ${newNames.filter(n => !n.exists).length || ''} categor${newNames.filter(n => !n.exists).length === 1 ? 'y' : 'ies'}` }}
          </button>
        </div>
      </div>
    </div>

    <div class="table-container">
      <table class="table">
        <thead><tr><th>Category</th><th>Manufacturer</th><th class="text-right">Products</th><th v-if="canEdit" class="text-right">Actions</th></tr></thead>
        <tbody>
          <tr v-for="row in rows" :key="row.category.id">
            <td>
              <div v-if="editingId === row.category.id" class="cat-edit">
                <input v-model="editName" class="form-input form-input--sm" :aria-label="`New name for ${row.category.name}`" maxlength="255" @keydown.enter.prevent="saveEdit(row.manufacturer, row.category)" @keydown.escape="editingId = ''" />
                <button class="btn btn-ghost btn-icon btn-sm" title="Save" @click="saveEdit(row.manufacturer, row.category)"><Check :size="14" /></button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Cancel" @click="editingId = ''"><X :size="14" /></button>
              </div>
              <span v-else class="font-medium">{{ row.category.name }}</span>
            </td>
            <td>{{ row.manufacturer.name }}</td>
            <td class="text-right">{{ row.products }}</td>
            <td v-if="canEdit" class="text-right">
              <div class="table-actions cat-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="Rename" @click="startEdit(row.category)"><Pencil :size="14" /></button>
                <button class="btn btn-ghost btn-icon btn-sm text-danger" :disabled="row.products > 0" :title="row.products > 0 ? 'In use by products — reassign them first' : 'Delete'" @click="remove(row.manufacturer, row.category)"><Trash2 :size="14" /></button>
              </div>
            </td>
          </tr>
          <tr v-if="!rows.length"><td :colspan="canEdit ? 4 : 3" class="text-muted">No categories match.</td></tr>
        </tbody>
      </table>
    </div>
    <p v-if="emptyManufacturers.length && !manufacturerFilter && !search" class="cat-note">
      No categories yet for: {{ emptyManufacturers.join(', ') }}.
    </p>
    <p v-if="!canEdit" class="cat-note">Adding or changing categories needs manufacturer editing rights (administrator or procurement manager).</p>
  </div>
</template>

<style scoped>
.cat-error { color: var(--color-danger); background: var(--color-danger-light); border-radius: var(--radius-md); padding: var(--space-2) var(--space-3); font-size: var(--text-sm); margin-bottom: var(--space-3); }
.cat-notice { color: var(--color-success-dark, #166534); font-size: var(--text-sm); margin-bottom: var(--space-3); }
.cat-add { margin-bottom: var(--space-4); }
.cat-add-title { font-size: var(--text-base); font-weight: 600; margin-bottom: var(--space-3); }
.cat-preview { display: flex; flex-wrap: wrap; gap: var(--space-1); margin-bottom: var(--space-3); }
.cat-add-actions { display: flex; justify-content: flex-end; gap: var(--space-2); }
.cat-edit { display: flex; align-items: center; gap: var(--space-1); }
.cat-actions { justify-content: flex-end; }
.cat-note { font-size: var(--text-xs); color: var(--color-neutral-500); margin-top: var(--space-3); }
</style>
