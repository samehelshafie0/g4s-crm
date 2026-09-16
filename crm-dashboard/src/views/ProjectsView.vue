<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  FolderKanban, Search, Plus, Eye, Pencil, Trash2, X, Package, Users,
  Calendar, DollarSign, AlertCircle, CheckCircle, Clock, Pause, XCircle,
  ShoppingCart, TrendingUp, BarChart3, ArrowRight,
} from 'lucide-vue-next'
import { useQuotesStore } from '@/stores/quotes'
import { useProcurementStore } from '@/stores/procurement'
import { useRouter } from 'vue-router'
import type { Project, ProjectStatus, ProjectPriority, Quote, QuoteLineItem, PurchaseOrder } from '@/types'

const quotesStore = useQuotesStore()
const procStore = useProcurementStore()
const router = useRouter()

import { projectsService, quotesService, usersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
const managers = ref<{id:string;name:string}[]>([])
onMounted(async () => {try {const [items,quotes,users] = await Promise.all([allPages(projectsService.list), allPages(quotesService.list), usersService.lookup()]); projects.value=items;quotesStore.quotes=quotes;managers.value=users.data.map(u=>({id:u.id,name:`${u.firstName} ${u.lastName}`}))}catch(e){window.alert(errorMessage(e))}})

function uid(): string { return Math.random().toString(36).slice(2, 11) }
function formatSAR(v: number): string { return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }
function formatDate(d: string): string { return new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) }
function daysFromNow(d: string): number { return Math.ceil((new Date(d).getTime() - Date.now()) / 86400000) }

// ── Project seed data from accepted quotes ───────────────────
const projects = ref<Project[]>([])
// ── KPIs ─────────────────────────────────────────────────────
const totalProjects = computed(() => projects.value.length)
const activeProjects = computed(() => projects.value.filter(p => p.status === 'in-progress').length)
const totalProjectValue = computed(() => projects.value.reduce((s, p) => s + p.totalValue, 0))
const avgMargin = computed(() => {
  if (projects.value.length === 0) return 0
  return projects.value.reduce((s, p) => s + p.marginPercent, 0) / projects.value.length
})

// ── Filters ──────────────────────────────────────────────────
const searchQuery = ref('')
const filterStatus = ref<'' | ProjectStatus>('')
const filterPriority = ref<'' | ProjectPriority>('')

const filteredProjects = computed(() => {
  let list = projects.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) list = list.filter(p => p.projectNumber.toLowerCase().includes(q) || p.name.toLowerCase().includes(q) || p.customerName.toLowerCase().includes(q))
  if (filterStatus.value) list = list.filter(p => p.status === filterStatus.value)
  if (filterPriority.value) list = list.filter(p => p.priority === filterPriority.value)
  return list
})

// ── Status / Priority config ─────────────────────────────────
const statusConfig: Record<ProjectStatus, { label: string; badge: string; icon: typeof CheckCircle }> = {
  planning: { label: 'Planning', badge: 'badge-gray', icon: Clock },
  'in-progress': { label: 'In Progress', badge: 'badge-primary', icon: TrendingUp },
  'on-hold': { label: 'On Hold', badge: 'badge-warning', icon: Pause },
  completed: { label: 'Completed', badge: 'badge-success', icon: CheckCircle },
  cancelled: { label: 'Cancelled', badge: 'badge-danger', icon: XCircle },
}
const priorityConfig: Record<ProjectPriority, { label: string; badge: string }> = {
  low: { label: 'Low', badge: 'badge-gray' },
  medium: { label: 'Medium', badge: 'badge-info' },
  high: { label: 'High', badge: 'badge-warning' },
  critical: { label: 'Critical', badge: 'badge-danger' },
}

// ── Project Detail Modal ─────────────────────────────────────
const showDetailModal = ref(false)
const detailProject = ref<Project | null>(null)
type DetailTab = 'overview' | 'items' | 'purchase-orders'
const detailTab = ref<DetailTab>('overview')

function openDetail(p: Project) { detailProject.value = p; detailTab.value = 'overview'; showDetailModal.value = true }

const detailMaterialItems = computed(() => detailProject.value?.lineItems.filter(li => li.category === 'materials') || [])
const detailManpowerItems = computed(() => detailProject.value?.lineItems.filter(li => li.category === 'manpower') || [])
const detailMiscItems = computed(() => detailProject.value?.lineItems.filter(li => li.category === 'miscellaneous') || [])
const detailPOs = computed<PurchaseOrder[]>(() => {
  if (!detailProject.value) return []
  return procStore.purchaseOrders.filter(po =>
    (detailProject.value!.purchaseOrders ?? []).includes(po.poNumber) ||
    po.sourceQuoteNumber === detailProject.value!.quoteNumber,
  )
})

async function updateStatus(status: ProjectStatus) {
 if (!detailProject.value) return
 try { const result=await projectsService.update(detailProject.value.id,{status});detailProject.value=result.data;projects.value=await allPages(projectsService.list) }catch(e){window.alert(errorMessage(e))}
}

// ── Convert Won Quote to Project ─────────────────────────────
const showConvertModal = ref(false)
const convertQuote = ref<Quote | null>(null)
const convertForm = ref({ name: '', startDate: '', targetEndDate: '', projectManager: '', priority: 'medium' as ProjectPriority, notes: '' })

const wonQuotes = computed(() => {
  const existingQuoteIds = new Set(projects.value.map(p => p.quoteId))
  return quotesStore.quotes.filter(q => q.status === 'accepted' && !existingQuoteIds.has(q.id))
})

function openConvertQuote(quote?: Quote) {
  const q = quote || wonQuotes.value[0]
  if (!q) return
  convertQuote.value = q
  const today = new Date().toISOString().slice(0, 10)
  convertForm.value = {
    name: `${q.customerName} — ${q.notes.slice(0, 50)}`,
    startDate: today,
    targetEndDate: new Date(Date.now() + 180 * 86400000).toISOString().slice(0, 10),
    projectManager: '',
    priority: 'medium',
    notes: `Converted from quote ${q.quoteNumber}`,
  }
  showConvertModal.value = true
}

async function confirmConvert() {
 if (!convertQuote.value || !convertForm.value.name) return
 try { await projectsService.create({ ...convertForm.value, projectManagerId:convertForm.value.projectManager || null,customerId:convertQuote.value.customerId,quoteId:convertQuote.value.id });projects.value=await allPages(projectsService.list);showConvertModal.value=false }catch(e){window.alert(errorMessage(e))}
}

function goToProcurement() {
  router.push('/procurement')
}

async function deleteProject(id: string) {try{await projectsService.delete(id);projects.value=projects.value.filter(p=>p.id!==id)}catch(e){window.alert(errorMessage(e))}}
</script>

<template>
  <div class="projects-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Projects</h1>
        <p class="page-header-subtitle">Won quotes converted to active projects</p>
      </div>
      <div class="page-header-actions">
        <button v-if="wonQuotes.length" class="btn btn-primary" @click="openConvertQuote()">
          <Plus :size="18" /> Convert Won Quote
          <span class="badge badge-light" style="margin-left:6px; font-size:0.65rem">{{ wonQuotes.length }}</span>
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="kpi-row kpi-row--4">
      <div class="kpi-card"><div class="kpi-icon kpi-icon--blue"><FolderKanban :size="20" /></div><div class="kpi-body"><span class="kpi-value">{{ totalProjects }}</span><span class="kpi-label">Total Projects</span></div></div>
      <div class="kpi-card"><div class="kpi-icon kpi-icon--green"><TrendingUp :size="20" /></div><div class="kpi-body"><span class="kpi-value">{{ activeProjects }}</span><span class="kpi-label">Active</span></div></div>
      <div class="kpi-card"><div class="kpi-icon kpi-icon--purple"><DollarSign :size="20" /></div><div class="kpi-body"><span class="kpi-value">SAR {{ formatSAR(totalProjectValue) }}</span><span class="kpi-label">Total Value</span></div></div>
      <div class="kpi-card"><div class="kpi-icon kpi-icon--orange"><BarChart3 :size="20" /></div><div class="kpi-body"><span class="kpi-value">{{ avgMargin.toFixed(1) }}%</span><span class="kpi-label">Avg Margin</span></div></div>
    </div>

    <!-- Filter toolbar -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search"><Search :size="18" class="search-icon" /><input v-model="searchQuery" type="text" class="form-input" placeholder="Search by project #, name, or customer..." /></div>
        <select v-model="filterStatus" class="form-select toolbar-select"><option value="">All Statuses</option><option value="planning">Planning</option><option value="in-progress">In Progress</option><option value="on-hold">On Hold</option><option value="completed">Completed</option><option value="cancelled">Cancelled</option></select>
        <select v-model="filterPriority" class="form-select toolbar-select"><option value="">All Priorities</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="critical">Critical</option></select>
      </div>
    </div>

    <!-- Projects table -->
    <div v-if="filteredProjects.length" class="table-container">
      <table class="table">
        <thead><tr><th>Project #</th><th>Name</th><th>Customer</th><th>Status</th><th>Priority</th><th class="text-right">Value</th><th class="text-right">Margin</th><th>Timeline</th><th>Manager</th><th class="text-center">POs</th><th></th></tr></thead>
        <tbody>
          <tr v-for="p in filteredProjects" :key="p.id" class="prj-row" @click="openDetail(p)">
            <td class="text-mono font-medium" style="color: var(--color-primary)">{{ p.projectNumber }}</td>
            <td><span class="font-medium text-dark">{{ p.name }}</span><div class="text-xs text-muted">Quote: {{ p.quoteNumber }}</div></td>
            <td class="font-medium">{{ p.customerName }}</td>
            <td><span :class="['badge badge-dot', statusConfig[p.status].badge]">{{ statusConfig[p.status].label }}</span></td>
            <td><span :class="['badge', priorityConfig[p.priority].badge]" style="font-size:0.65rem">{{ priorityConfig[p.priority].label }}</span></td>
            <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(p.totalValue) }}</td>
            <td class="text-right"><span :class="['font-semibold', p.marginPercent >= 25 ? 'text-success' : p.marginPercent >= 20 ? 'text-warning' : 'text-danger']">{{ p.marginPercent.toFixed(1) }}%</span></td>
            <td class="whitespace-nowrap">
              <div style="font-size:0.75rem">{{ formatDate(p.startDate) }}</div>
              <div class="text-xs text-muted">→ {{ formatDate(p.targetEndDate) }}</div>
            </td>
            <td class="text-muted" style="font-size:0.78rem">{{ p.projectManager || '—' }}</td>
            <td class="text-center">
              <span v-if="p.purchaseOrders.length" class="badge badge-info" style="font-size:0.65rem">{{ p.purchaseOrders.length }}</span>
              <span v-else class="text-muted">—</span>
            </td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" @click.stop="openDetail(p)"><Eye :size="14" /></button>
                <button class="btn btn-ghost btn-icon btn-sm" @click.stop="deleteProject(p.id)"><Trash2 :size="14" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-state">
      <FolderKanban :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No projects found</h3>
      <p class="empty-state-text">{{ searchQuery || filterStatus || filterPriority ? 'Try adjusting your filters.' : 'Convert a won quote to create your first project.' }}</p>
      <button v-if="wonQuotes.length" class="btn btn-primary mt-4" @click="openConvertQuote()"><Plus :size="16" /> Convert Won Quote</button>
    </div>

    <!-- ═══ Project Detail Modal ══════════════════════════════ -->
    <div v-if="showDetailModal && detailProject" class="modal-backdrop" @click.self="showDetailModal = false">
      <div class="modal prj-detail-modal">
        <div class="prj-detail-header">
          <div class="prj-detail-header-left">
            <div class="prj-detail-icon"><FolderKanban :size="24" /></div>
            <div>
              <h2 class="prj-detail-title">{{ detailProject.name }}</h2>
              <div class="prj-detail-meta">
                <span class="prj-detail-num">{{ detailProject.projectNumber }}</span>
                <span class="prj-sep">|</span>
                <span class="font-medium">{{ detailProject.customerName }}</span>
                <span class="prj-sep">|</span>
                <span :class="['badge badge-dot', statusConfig[detailProject.status].badge]" style="font-size:0.7rem">{{ statusConfig[detailProject.status].label }}</span>
                <span :class="['badge', priorityConfig[detailProject.priority].badge]" style="font-size:0.65rem">{{ priorityConfig[detailProject.priority].label }}</span>
              </div>
            </div>
          </div>
          <div class="prj-detail-header-right">
            <div class="prj-hdr-num"><span class="prj-hdr-val">SAR {{ formatSAR(detailProject.totalValue) }}</span><span class="prj-hdr-lbl">Value</span></div>
            <div class="prj-hdr-num"><span :class="['prj-hdr-val', detailProject.marginPercent >= 25 ? 'text-success' : 'text-warning']">{{ detailProject.marginPercent.toFixed(1) }}%</span><span class="prj-hdr-lbl">Margin</span></div>
            <button class="modal-close" @click="showDetailModal = false"><X :size="20" /></button>
          </div>
        </div>

        <div class="prj-detail-tabs">
          <button :class="['prj-dtab', detailTab === 'overview' && 'prj-dtab--active']" @click="detailTab = 'overview'"><FolderKanban :size="15" /> Overview</button>
          <button :class="['prj-dtab', detailTab === 'items' && 'prj-dtab--active']" @click="detailTab = 'items'"><Package :size="15" /> Line Items <span class="prj-dtab-ct">{{ detailProject.lineItems.length }}</span></button>
          <button :class="['prj-dtab', detailTab === 'purchase-orders' && 'prj-dtab--active']" @click="detailTab = 'purchase-orders'"><ShoppingCart :size="15" /> Purchase Orders <span class="prj-dtab-ct">{{ detailPOs.length }}</span></button>
        </div>

        <div class="modal-body prj-detail-body">
          <!-- OVERVIEW -->
          <div v-if="detailTab === 'overview'">
            <div class="prj-info-grid">
              <div class="prj-info-card"><span class="prj-info-lbl">Project #</span><span class="prj-info-val text-mono">{{ detailProject.projectNumber }}</span></div>
              <div class="prj-info-card"><span class="prj-info-lbl">Quote #</span><span class="prj-info-val text-mono">{{ detailProject.quoteNumber }}</span></div>
              <div class="prj-info-card"><span class="prj-info-lbl">Customer</span><span class="prj-info-val">{{ detailProject.customerName }}</span></div>
              <div class="prj-info-card"><span class="prj-info-lbl">Project Manager</span><span class="prj-info-val">{{ detailProject.projectManager || '—' }}</span></div>
              <div class="prj-info-card"><span class="prj-info-lbl">Start Date</span><span class="prj-info-val">{{ formatDate(detailProject.startDate) }}</span></div>
              <div class="prj-info-card"><span class="prj-info-lbl">Target End</span><span class="prj-info-val">{{ formatDate(detailProject.targetEndDate) }}
                <span v-if="daysFromNow(detailProject.targetEndDate) > 0" class="text-muted" style="font-size:0.68rem; margin-left:4px">({{ daysFromNow(detailProject.targetEndDate) }}d left)</span>
                <span v-else class="text-danger" style="font-size:0.68rem; margin-left:4px">(Overdue)</span>
              </span></div>
            </div>
            <div class="prj-cost-grid">
              <div class="prj-cost-item"><span class="prj-cost-lbl">Total Value</span><span class="prj-cost-val font-bold" style="color:var(--color-primary)">SAR {{ formatSAR(detailProject.totalValue) }}</span></div>
              <div class="prj-cost-item"><span class="prj-cost-lbl">Total Cost</span><span class="prj-cost-val">SAR {{ formatSAR(detailProject.totalCost) }}</span></div>
              <div class="prj-cost-item"><span class="prj-cost-lbl">Margin</span><span :class="['prj-cost-val font-bold', detailProject.marginPercent >= 25 ? 'text-success' : 'text-warning']">{{ detailProject.marginPercent.toFixed(1) }}% (SAR {{ formatSAR(detailProject.totalValue - detailProject.totalCost) }})</span></div>
              <div class="prj-cost-item"><span class="prj-cost-lbl">Materials</span><span class="prj-cost-val">{{ detailMaterialItems.length }} items</span></div>
              <div class="prj-cost-item"><span class="prj-cost-lbl">Purchase Orders</span><span class="prj-cost-val">{{ detailPOs.length }}</span></div>
            </div>

            <div v-if="detailProject.notes" style="margin-top: var(--space-4)">
              <h4 class="prj-section-title">Notes</h4>
              <p class="text-muted" style="font-size:0.82rem">{{ detailProject.notes }}</p>
            </div>

            <div class="prj-status-actions">
              <h4 class="prj-section-title">Update Status</h4>
              <div class="prj-status-btns">
                <button v-for="(cfg, st) in statusConfig" :key="st" :class="['btn btn-sm', detailProject.status === st ? 'btn-primary' : 'btn-secondary']" @click="updateStatus(st as ProjectStatus)">
                  <component :is="cfg.icon" :size="13" /> {{ cfg.label }}
                </button>
              </div>
            </div>

            <div class="prj-action-bar">
              <button class="btn btn-primary" @click="showDetailModal = false; goToProcurement()"><ShoppingCart :size="14" /> Create Purchase Order</button>
            </div>
          </div>

          <!-- LINE ITEMS -->
          <div v-if="detailTab === 'items'">
            <div v-if="detailMaterialItems.length">
              <h4 class="prj-section-title"><Package :size="14" /> Materials ({{ detailMaterialItems.length }})</h4>
              <div class="table-container table-container--embedded">
                <table class="table"><thead><tr><th>SKU</th><th>Description</th><th>Manufacturer</th><th class="text-center">Qty</th><th class="text-right">Unit Cost</th><th class="text-right">Unit Price</th><th class="text-right">Total</th><th class="text-right">Margin</th></tr></thead>
                <tbody>
                  <tr v-for="li in detailMaterialItems" :key="li.id">
                    <td class="text-mono" style="font-size:0.75rem">{{ li.sku || '—' }}</td>
                    <td class="font-medium" style="font-size:0.8rem">{{ li.description }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ li.manufacturerName || '—' }}</td>
                    <td class="text-center">{{ li.quantity }}</td>
                    <td class="text-right whitespace-nowrap" style="font-size:0.78rem">SAR {{ formatSAR(li.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap" style="font-size:0.78rem">SAR {{ formatSAR(li.unitPrice) }}</td>
                    <td class="text-right whitespace-nowrap font-medium" style="font-size:0.78rem">SAR {{ formatSAR(li.lineTotal) }}</td>
                    <td class="text-right"><span :class="['font-semibold', li.marginPercent >= 25 ? 'text-success' : 'text-warning']">{{ li.marginPercent.toFixed(1) }}%</span></td>
                  </tr>
                </tbody></table>
              </div>
            </div>
            <div v-if="detailManpowerItems.length" style="margin-top: var(--space-5)">
              <h4 class="prj-section-title"><Users :size="14" /> Manpower ({{ detailManpowerItems.length }})</h4>
              <div class="table-container table-container--embedded">
                <table class="table"><thead><tr><th>Description</th><th class="text-center">Qty</th><th class="text-right">Cost</th><th class="text-right">Price</th><th class="text-right">Total</th></tr></thead>
                <tbody><tr v-for="li in detailManpowerItems" :key="li.id"><td class="font-medium">{{ li.description }}</td><td class="text-center">{{ li.quantity }}</td><td class="text-right">SAR {{ formatSAR(li.unitCost) }}</td><td class="text-right">SAR {{ formatSAR(li.unitPrice) }}</td><td class="text-right font-medium">SAR {{ formatSAR(li.lineTotal) }}</td></tr></tbody></table>
              </div>
            </div>
            <div v-if="detailMiscItems.length" style="margin-top: var(--space-5)">
              <h4 class="prj-section-title">Miscellaneous ({{ detailMiscItems.length }})</h4>
              <div class="table-container table-container--embedded">
                <table class="table"><thead><tr><th>Description</th><th class="text-center">Qty</th><th class="text-right">Cost</th><th class="text-right">Price</th><th class="text-right">Total</th></tr></thead>
                <tbody><tr v-for="li in detailMiscItems" :key="li.id"><td class="font-medium">{{ li.description }}</td><td class="text-center">{{ li.quantity }}</td><td class="text-right">SAR {{ formatSAR(li.unitCost) }}</td><td class="text-right">SAR {{ formatSAR(li.unitPrice) }}</td><td class="text-right font-medium">SAR {{ formatSAR(li.lineTotal) }}</td></tr></tbody></table>
              </div>
            </div>
          </div>

          <!-- PURCHASE ORDERS -->
          <div v-if="detailTab === 'purchase-orders'">
            <div class="prj-tab-toolbar">
              <span class="font-medium">{{ detailPOs.length }} purchase order{{ detailPOs.length !== 1 ? 's' : '' }}</span>
              <button class="btn btn-primary btn-sm" @click="showDetailModal = false; goToProcurement()"><Plus :size="14" /> Create PO</button>
            </div>
            <div v-if="detailPOs.length" class="table-container table-container--embedded">
              <table class="table"><thead><tr><th>PO #</th><th>Supplier</th><th>Status</th><th class="text-center">Items</th><th class="text-right">Total</th><th>Expected</th></tr></thead>
              <tbody><tr v-for="po in detailPOs" :key="po.id">
                <td class="text-mono font-medium" style="color: var(--color-primary)">{{ po.poNumber }}</td>
                <td class="font-medium">{{ po.supplierName }}</td>
                <td><span class="badge badge-primary" style="font-size:0.65rem">{{ po.status }}</span></td>
                <td class="text-center">{{ po.items.length }}</td>
                <td class="text-right font-medium">SAR {{ formatSAR(po.total) }}</td>
                <td class="text-muted" style="font-size:0.78rem">{{ formatDate(po.expectedDelivery) }}</td>
              </tr></tbody></table>
            </div>
            <div v-else class="prj-empty-tab">
              <ShoppingCart :size="36" class="text-muted" />
              <p class="text-muted">No purchase orders linked to this project yet.</p>
              <button class="btn btn-primary btn-sm" @click="showDetailModal = false; goToProcurement()"><Plus :size="14" /> Create PO</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ═══ Convert Won Quote Modal ═══════════════════════════ -->
    <div v-if="showConvertModal && convertQuote" class="modal-backdrop" @click.self="showConvertModal = false">
      <div class="modal modal-lg">
        <div class="modal-header">
          <h2 class="modal-title"><ArrowRight :size="18" /> Convert Quote to Project</h2>
          <button class="modal-close" @click="showConvertModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body">
          <div class="prj-convert-quote-info">
            <span class="text-mono font-bold">{{ convertQuote.quoteNumber }}</span>
            <span class="font-medium">{{ convertQuote.customerName }}</span>
            <span class="badge badge-success" style="font-size:0.65rem">Accepted</span>
            <span class="font-medium" style="color:var(--color-primary)">SAR {{ formatSAR(convertQuote.total) }}</span>
          </div>

          <div v-if="wonQuotes.length > 1" class="form-group" style="margin-top: var(--space-4)">
            <label class="form-label">Select Quote</label>
            <select class="form-select" :value="convertQuote.id" @change="openConvertQuote(wonQuotes.find(q => q.id === ($event.target as HTMLSelectElement).value))">
              <option v-for="q in wonQuotes" :key="q.id" :value="q.id">{{ q.quoteNumber }} — {{ q.customerName }} — SAR {{ formatSAR(q.total) }}</option>
            </select>
          </div>

          <div class="form-group" style="margin-top: var(--space-4)">
            <label class="form-label">Project Name <span class="required">*</span></label>
            <input v-model="convertForm.name" type="text" class="form-input" placeholder="e.g. Al Rajhi Bank — CCTV Phase 2" />
          </div>
          <div class="form-row">
            <div class="form-group"><label class="form-label">Start Date</label><input v-model="convertForm.startDate" type="date" class="form-input" /></div>
            <div class="form-group"><label class="form-label">Target End Date</label><input v-model="convertForm.targetEndDate" type="date" class="form-input" /></div>
          </div>
          <div class="form-row">
            <div class="form-group"><label class="form-label">Project Manager</label><select v-model="convertForm.projectManager" class="select"><option value="">Unassigned</option><option v-for="manager in managers" :key="manager.id" :value="manager.id">{{ manager.name }}</option></select></div>
            <div class="form-group" style="width:160px"><label class="form-label">Priority</label><select v-model="convertForm.priority" class="form-select"><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option><option value="critical">Critical</option></select></div>
          </div>
          <div class="form-group"><label class="form-label">Notes</label><textarea v-model="convertForm.notes" class="form-textarea" rows="2" placeholder="Project notes..." /></div>

          <div class="prj-convert-items-preview">
            <h4 class="prj-section-title" style="margin-top:0"><Package :size="14" /> Items from Quote ({{ convertQuote.lineItems.length }})</h4>
            <div class="table-container table-container--embedded" style="max-height:200px">
              <table class="table"><thead><tr><th>Description</th><th>Category</th><th class="text-center">Qty</th><th class="text-right">Total</th></tr></thead>
              <tbody><tr v-for="li in convertQuote.lineItems" :key="li.id">
                <td class="font-medium" style="font-size:0.78rem">{{ li.description }}</td>
                <td><span class="badge badge-gray" style="font-size:0.6rem">{{ li.category }}</span></td>
                <td class="text-center">{{ li.quantity }}</td>
                <td class="text-right font-medium" style="font-size:0.78rem">SAR {{ formatSAR(li.lineTotal) }}</td>
              </tr></tbody></table>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showConvertModal = false">Cancel</button>
          <button class="btn btn-primary" :disabled="!convertForm.name" @click="confirmConvert"><FolderKanban :size="14" /> Create Project</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.projects-page { padding: var(--space-6); }

.page-header-actions { display: flex; gap: var(--space-3); }
.badge-light { background: rgba(255,255,255,0.3); color: inherit; }

/* KPI */
.kpi-row--4 { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--space-4); margin-bottom: var(--space-6); }
.kpi-card { display: flex; align-items: center; gap: var(--space-4); padding: var(--space-5); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); }
.kpi-icon { width: 44px; height: 44px; border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.kpi-icon--blue { background: #dbeafe; color: #2563eb; }
.kpi-icon--green { background: #dcfce7; color: #16a34a; }
.kpi-icon--purple { background: #ede9fe; color: #7c3aed; }
.kpi-icon--orange { background: #ffedd5; color: #ea580c; }
[data-theme="dark"] .kpi-icon--blue { background: rgba(37,99,235,0.15); }
[data-theme="dark"] .kpi-icon--green { background: rgba(22,163,74,0.15); }
[data-theme="dark"] .kpi-icon--purple { background: rgba(124,58,237,0.15); }
[data-theme="dark"] .kpi-icon--orange { background: rgba(234,88,12,0.15); }
.kpi-body { display: flex; flex-direction: column; }
.kpi-value { font-size: 1.25rem; font-weight: 800; color: var(--color-neutral-900); }
.kpi-label { font-size: 0.72rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; }

/* Toolbar */
.toolbar { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) var(--space-5); flex-wrap: wrap; }
.toolbar-search { flex: 1; min-width: 220px; }
.toolbar-select { width: 180px; flex-shrink: 0; }

.prj-row { cursor: pointer; transition: background 80ms; }
.prj-row:hover { background: var(--color-primary-50, #eef2ff); }

/* Detail Modal */
.prj-detail-modal { width: 100%; max-width: 1200px; max-height: calc(100vh - var(--space-8)); }
.prj-detail-header { display: flex; align-items: center; justify-content: space-between; padding: var(--space-5) var(--space-6); border-bottom: 1px solid var(--color-neutral-200); gap: var(--space-4); }
.prj-detail-header-left { display: flex; align-items: center; gap: var(--space-4); }
.prj-detail-icon { width: 48px; height: 48px; border-radius: var(--radius-lg); background: var(--color-primary-light, #e0e7ff); color: var(--color-primary); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.prj-detail-title { font-size: 1.1rem; font-weight: 700; color: var(--color-neutral-900); margin: 0; }
.prj-detail-meta { display: flex; align-items: center; gap: var(--space-2); margin-top: 4px; font-size: 0.78rem; flex-wrap: wrap; }
.prj-detail-num { font-family: var(--font-mono); font-weight: 600; color: var(--color-primary); background: var(--color-primary-50, #eef2ff); padding: 1px 8px; border-radius: var(--radius-sm); font-size: 0.72rem; }
.prj-sep { color: var(--color-neutral-300); }
.prj-detail-header-right { display: flex; align-items: center; gap: var(--space-5); }
.prj-hdr-num { text-align: center; }
.prj-hdr-val { display: block; font-size: 1rem; font-weight: 800; color: var(--color-neutral-800); }
.prj-hdr-lbl { font-size: 0.62rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.05em; }

.prj-detail-tabs { display: flex; gap: 0; padding: 0 var(--space-6); border-bottom: 2px solid var(--color-neutral-200); overflow-x: auto; background: var(--color-neutral-50); }
.prj-dtab { display: flex; align-items: center; gap: 6px; padding: var(--space-3) var(--space-5); font-size: 0.82rem; font-weight: 500; color: var(--color-neutral-500); border-bottom: 3px solid transparent; margin-bottom: -2px; cursor: pointer; background: none; border-top: none; border-left: none; border-right: none; transition: all 120ms; white-space: nowrap; }
.prj-dtab:hover { color: var(--color-neutral-700); background: var(--color-neutral-100); }
.prj-dtab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); font-weight: 700; background: transparent; }
.prj-dtab-ct { font-size: 0.68rem; font-weight: 700; background: var(--color-neutral-200); color: var(--color-neutral-600); padding: 1px 7px; border-radius: 10px; }
.prj-dtab--active .prj-dtab-ct { background: var(--color-primary-100, #c7d2fe); color: var(--color-primary); }
.prj-detail-body { padding: var(--space-6); max-height: 70vh; overflow-y: auto; }

/* Info grid */
.prj-info-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-3); margin-bottom: var(--space-4); }
.prj-info-card { padding: var(--space-3) var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); }
.prj-info-lbl { display: block; font-size: 0.65rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 2px; }
.prj-info-val { font-size: 0.85rem; font-weight: 600; color: var(--color-neutral-800); }

.prj-section-title { display: flex; align-items: center; gap: var(--space-2); font-size: 0.82rem; font-weight: 700; color: var(--color-neutral-700); margin: var(--space-5) 0 var(--space-3) 0; }

.prj-cost-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: var(--space-3); }
.prj-cost-item { display: flex; flex-direction: column; padding: var(--space-3) var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); }
.prj-cost-lbl { font-size: 0.65rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; }
.prj-cost-val { font-size: 0.82rem; font-weight: 600; color: var(--color-neutral-800); margin-top: 2px; }

.prj-status-actions { margin-top: var(--space-5); }
.prj-status-btns { display: flex; gap: var(--space-2); flex-wrap: wrap; }

.prj-action-bar { display: flex; gap: var(--space-3); margin-top: var(--space-5); padding-top: var(--space-4); border-top: 1px solid var(--color-neutral-200); }

.prj-tab-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }
.prj-empty-tab { display: flex; flex-direction: column; align-items: center; gap: var(--space-3); padding: var(--space-8) 0; }

/* Convert modal */
.prj-convert-quote-info { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); font-size: 0.85rem; }
.prj-convert-items-preview { margin-top: var(--space-4); padding-top: var(--space-4); border-top: 1px solid var(--color-neutral-200); }

/* Embedded tables */
.table-container--embedded { border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); overflow: auto; max-height: 400px; }
.table-container--embedded .table { margin-bottom: 0; }
.table-container--embedded .table th { position: sticky; top: 0; z-index: 2; background: var(--color-neutral-50); }

@media (max-width: 1200px) {
  .kpi-row--4 { grid-template-columns: repeat(2, 1fr); }
  .prj-info-grid { grid-template-columns: repeat(2, 1fr); }
  .prj-cost-grid { grid-template-columns: repeat(3, 1fr); }
  .prj-detail-header { flex-direction: column; align-items: flex-start; }
}
@media (max-width: 768px) {
  .kpi-row--4 { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-select { width: 100%; }
  .prj-info-grid { grid-template-columns: 1fr; }
  .prj-cost-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
