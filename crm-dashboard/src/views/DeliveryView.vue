<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Pencil, X, ExternalLink } from 'lucide-vue-next'
import { opportunitiesService, usersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import type { DeliveryStatus, Opportunity } from '@/types'

const router = useRouter()
const auth = useAuthStore()
const canEdit = computed(() => auth.can('projects:update'))

const statuses: { value: DeliveryStatus; label: string; badge: string }[] = [
  { value: 'planning', label: 'Planning', badge: 'badge-neutral' },
  { value: 'in-progress', label: 'In progress', badge: 'badge-primary' },
  { value: 'on-hold', label: 'On hold', badge: 'badge-warning' },
  { value: 'completed', label: 'Completed', badge: 'badge-success' },
  { value: 'cancelled', label: 'Cancelled', badge: 'badge-danger' },
]
const statusInfo = (s?: DeliveryStatus) => statuses.find((x) => x.value === s)

const jobs = ref<Opportunity[]>([])
const managers = ref<{ id: string; name: string }[]>([])
const loading = ref(false)
const error = ref('')
const search = ref('')
const statusFilter = ref<DeliveryStatus | ''>('')
const openOnly = ref(true)

const money = (v?: number) => (v == null ? '—' : `SAR ${v.toLocaleString('en-SA', { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`)
const day = (d?: string | null) => (d ? new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) : '—')
const isLate = (o: Opportunity) =>
  !!o.targetEndDate && !['completed', 'cancelled'].includes(o.deliveryStatus ?? '') && new Date(o.targetEndDate).getTime() < Date.now()

async function load() {
  loading.value = true
  error.value = ''
  try {
    jobs.value = await allPages(opportunitiesService.delivery, openOnly.value ? { open: true } : undefined)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await load()
  try {
    const res = await usersService.lookup()
    managers.value = res.data.map((u) => ({ id: u.id, name: `${u.firstName} ${u.lastName}` }))
  } catch { managers.value = [] }
})

const filtered = computed(() => {
  const q = search.value.toLowerCase().trim()
  return jobs.value.filter((o) => {
    if (statusFilter.value && o.deliveryStatus !== statusFilter.value) return false
    if (!q) return true
    return o.title.toLowerCase().includes(q) || o.customerName.toLowerCase().includes(q) || (o.currentQuote?.quoteNumber ?? '').toLowerCase().includes(q)
  })
})

const counts = computed(() => {
  const by = (s: DeliveryStatus) => jobs.value.filter((o) => o.deliveryStatus === s).length
  return { planning: by('planning'), inProgress: by('in-progress'), onHold: by('on-hold'), late: jobs.value.filter(isLate).length }
})

// ── Edit ────────────────────────────────────────────────────
const editing = ref<Opportunity | null>(null)
const saving = ref(false)
const form = ref({
  deliveryStatus: 'planning' as DeliveryStatus,
  projectManagerId: '',
  startDate: '',
  targetEndDate: '',
  actualEndDate: '',
  actualValue: null as number | null,
  actualCost: null as number | null,
})

function openEdit(o: Opportunity) {
  editing.value = o
  error.value = ''
  form.value = {
    deliveryStatus: o.deliveryStatus ?? 'planning',
    projectManagerId: o.projectManagerId ?? '',
    startDate: o.startDate?.slice(0, 10) ?? '',
    targetEndDate: o.targetEndDate?.slice(0, 10) ?? '',
    actualEndDate: o.actualEndDate?.slice(0, 10) ?? '',
    actualValue: o.actualValue ?? null,
    actualCost: o.actualCost ?? null,
  }
}

// Completing a job is the operator's call, but not silently while quoted items
// are still undispatched.
async function confirmCompletion(job: Opportunity): Promise<boolean> {
  try {
    const { lines } = (await opportunitiesService.fulfilment(job.id)).data
    const open = lines.filter((l) => l.source !== 'service' && l.dispatched < Math.ceil(l.quantity))
    if (!open.length) return true
    const list = open.map((l) => `• ${l.sku || l.description}: ${l.dispatched} of ${Math.ceil(l.quantity)} dispatched`).join('\n')
    return window.confirm(`${open.length} line(s) of ${job.title} are not fully dispatched:\n\n${list}\n\nMark the delivery completed anyway?`)
  } catch {
    return window.confirm('Dispatch status could not be checked. Mark the delivery completed anyway?')
  }
}

async function save() {
  if (!editing.value) return
  if (form.value.deliveryStatus === 'completed' && editing.value.deliveryStatus !== 'completed' && !(await confirmCompletion(editing.value))) return
  saving.value = true
  error.value = ''
  try {
    const f = form.value
    await opportunitiesService.updateDelivery(editing.value.id, {
      deliveryStatus: f.deliveryStatus,
      projectManagerId: f.projectManagerId || null,
      startDate: f.startDate || null,
      targetEndDate: f.targetEndDate || null,
      actualEndDate: f.actualEndDate || null,
      actualValue: f.actualValue,
      actualCost: f.actualCost,
    })
    editing.value = null
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="delivery-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Delivery</h1>
        <p class="page-header-subtitle">Won opportunities being delivered. The opportunity is the project.</p>
      </div>
    </div>

    <div class="kpi-grid">
      <div class="kpi-card card"><p class="kpi-label">Planning</p><p class="kpi-value">{{ counts.planning }}</p></div>
      <div class="kpi-card card"><p class="kpi-label">In progress</p><p class="kpi-value">{{ counts.inProgress }}</p></div>
      <div class="kpi-card card"><p class="kpi-label">On hold</p><p class="kpi-value">{{ counts.onHold }}</p></div>
      <div class="kpi-card card"><p class="kpi-label">Past target date</p><p class="kpi-value" :class="{ 'text-danger': counts.late }">{{ counts.late }}</p></div>
    </div>

    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search">
          <Search :size="18" class="search-icon" />
          <input v-model="search" type="text" class="form-input" placeholder="Search by job, customer or quote #…" />
        </div>
        <select v-model="statusFilter" class="form-select toolbar-select">
          <option value="">All statuses</option>
          <option v-for="s in statuses" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
        <label class="open-only"><input v-model="openOnly" type="checkbox" @change="load" /> Hide completed and cancelled</label>
      </div>
    </div>

    <p v-if="error && !editing" class="delivery-error" role="alert">{{ error }}</p>

    <div class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Job</th><th>Customer</th><th>Quote</th><th>Status</th><th>Project manager</th>
            <th>Start</th><th>Target end</th><th class="text-right">Quoted</th><th class="text-right">Actual</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading"><td colspan="10" class="text-muted">Loading…</td></tr>
          <tr v-for="o in filtered" :key="o.id">
            <td class="font-semibold"><a href="#" class="row-link" @click.prevent="router.push(`/opportunities/${o.id}`)">{{ o.title }}</a><div class="text-xs text-muted text-mono">{{ o.opportunityNumber }}</div></td>
            <td>{{ o.customerName }}</td>
            <td class="text-mono">{{ o.currentQuote ? `${o.currentQuote.quoteNumber} v${o.currentQuote.version}` : '—' }}</td>
            <td><span class="badge badge-dot" :class="statusInfo(o.deliveryStatus)?.badge">{{ statusInfo(o.deliveryStatus)?.label ?? '—' }}</span></td>
            <td>{{ o.projectManagerName || '—' }}</td>
            <td class="whitespace-nowrap">{{ day(o.startDate) }}</td>
            <td class="whitespace-nowrap" :class="{ 'text-danger': isLate(o) }">{{ day(o.targetEndDate) }}</td>
            <td class="text-right text-mono">{{ money(o.currentQuote?.total) }}</td>
            <td class="text-right text-mono">{{ money(o.actualValue) }}</td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="Open opportunity" @click="router.push(`/opportunities/${o.id}`)"><ExternalLink :size="15" /></button>
                <button v-if="canEdit" class="btn btn-ghost btn-icon btn-sm" title="Update delivery" @click="openEdit(o)"><Pencil :size="15" /></button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && !filtered.length">
            <td colspan="10">
              <div class="table-empty">
                <div class="empty-text">Nothing to deliver</div>
                <div class="empty-subtext">A job appears here when its quote is accepted and the opportunity is won.</div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Teleport to="body">
      <div v-if="editing" class="modal-backdrop" @click.self="editing = null">
        <div class="modal">
          <div class="modal-header">
            <h3 class="modal-title">Delivery — {{ editing.title }}</h3>
            <button class="modal-close" @click="editing = null"><X :size="18" /></button>
          </div>
          <div class="modal-body">
            <p v-if="error" class="delivery-error" role="alert">{{ error }}</p>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Status</label>
                <select v-model="form.deliveryStatus" class="form-select">
                  <option v-for="s in statuses" :key="s.value" :value="s.value">{{ s.label }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Project manager</label>
                <select v-model="form.projectManagerId" class="form-select">
                  <option value="">Unassigned</option>
                  <option v-for="m in managers" :key="m.id" :value="m.id">{{ m.name }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group"><label class="form-label">Start date</label><input v-model="form.startDate" type="date" class="form-input" /></div>
              <div class="form-group"><label class="form-label">Target end date</label><input v-model="form.targetEndDate" type="date" class="form-input" /></div>
              <div class="form-group"><label class="form-label">Actual end date</label><input v-model="form.actualEndDate" type="date" class="form-input" /></div>
            </div>
            <div class="form-row">
              <div class="form-group"><label class="form-label">Actual value (SAR)</label><input v-model.number="form.actualValue" type="number" min="0" step="0.01" class="form-input" /></div>
              <div class="form-group"><label class="form-label">Actual cost (SAR)</label><input v-model.number="form.actualCost" type="number" min="0" step="0.01" class="form-input" /></div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="editing = null">Cancel</button>
            <button class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? 'Saving…' : 'Save' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.delivery-page { padding: var(--space-6); }
.open-only { display: inline-flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); white-space: nowrap; }
.row-link { color: inherit; text-decoration: none; }
.row-link:hover { text-decoration: underline; }
.delivery-error { color: var(--color-danger); font-size: var(--text-sm); margin-bottom: var(--space-3); }
</style>
