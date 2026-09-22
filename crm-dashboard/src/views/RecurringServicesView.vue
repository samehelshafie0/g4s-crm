<script setup lang="ts">
import { recurringServicesService, productsService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import { onMounted, ref, computed, watch } from 'vue'
import {
  Plus, Pencil, Trash2, X, Shield, Wrench, Eye as Monitor, Car, Building,
  Search, DollarSign, TrendingUp, RefreshCw,
  CheckCircle2, Copy, Package, Layers, Upload, FileText, PenLine,
  ChevronDown, ArrowRight, AlertTriangle,
} from 'lucide-vue-next'
import type { RecurringService, RecurringServiceType, BillingFrequency, Product } from '@/types'
import { useProductsStore } from '@/stores/products'
import { useRecurringServicesStore } from '@/stores/recurringServices'

const recurringStore = useRecurringServicesStore()
async function reloadServices() { services.value = await allPages(recurringServicesService.list); for (const service of services.value) serviceItems.value[service.id] = service.lineItems ?? [] }
onMounted(async () => { try { await reloadServices(); productsStore.products = await allPages(productsService.list) } catch (e) { window.alert(errorMessage(e)) } })

const productsStore = useProductsStore()

function uid(): string { return Math.random().toString(36).slice(2, 11) }

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function fmtCompact(v: number): string {
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(1) + 'M'
  if (v >= 1_000) return (v / 1_000).toFixed(1) + 'K'
  return v.toFixed(0)
}

const serviceTypeLabels: Record<RecurringServiceType, string> = {
  guarding: 'Guarding', maintenance: 'Maintenance', monitoring: 'Monitoring',
  patrol: 'Patrol', 'facility-management': 'Facility Mgmt', rental: 'Equipment Rental',
}

const serviceTypeColors: Record<RecurringServiceType, string> = {
  guarding: 'type-guarding', maintenance: 'type-maintenance', monitoring: 'type-monitoring',
  patrol: 'type-patrol', 'facility-management': 'type-fm', rental: 'type-rental',
}

const serviceTypeIcons: Record<RecurringServiceType, typeof Shield> = {
  guarding: Shield, maintenance: Wrench, monitoring: Monitor, patrol: Car, 'facility-management': Building, rental: Package,
}

const billingLabels: Record<BillingFrequency, string> = { monthly: 'Monthly', quarterly: 'Quarterly', annually: 'Annually' }

// ── Line Item Type ──────────────────────────────────────────
type ItemSource = 'product' | 'service' | 'vendor' | 'write-in'

interface ServiceLineItem {
  id: string
  source: ItemSource
  sourceId?: string
  sku: string
  name: string
  description: string
  qty: number
  unitCost: number
  unitPrice: number
  vendorName?: string
  vendorFile?: string
}

function lineTotal(item: ServiceLineItem): number { return item.qty * item.unitPrice }
function lineCost(item: ServiceLineItem): number { return item.qty * item.unitCost }
function lineMargin(item: ServiceLineItem): number {
  const t = lineTotal(item)
  return t > 0 ? ((t - lineCost(item)) / t) * 100 : 0
}

// ── Data ─────────────────────────────────────────────────────
const services = ref<RecurringService[]>([])
// Service-level line items map (serviceId -> items)
const serviceItems = ref<Record<string, ServiceLineItem[]>>({})

// ── Filters ──────────────────────────────────────────────────
const searchQuery = ref('')
const typeFilter = ref<RecurringServiceType | 'all'>('all')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')

const filteredServices = computed(() => {
  let list = services.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) list = list.filter(s => s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q))
  if (typeFilter.value !== 'all') list = list.filter(s => s.serviceType === typeFilter.value)
  if (statusFilter.value === 'active') list = list.filter(s => s.isActive)
  else if (statusFilter.value === 'inactive') list = list.filter(s => !s.isActive)
  return list
})

// ── KPIs ─────────────────────────────────────────────────────
const activeServices = computed(() => services.value.filter(s => s.isActive))
const totalMonthlyRevenue = computed(() => activeServices.value.reduce((s, x) => s + x.monthlyPrice, 0))
const totalAnnualRevenue = computed(() => activeServices.value.reduce((s, x) => s + x.annualPrice, 0))
const avgMargin = computed(() => {
  if (!activeServices.value.length) return 0
  return activeServices.value.reduce((s, x) => s + x.targetMarginPercent, 0) / activeServices.value.length
})

// ── Modal Form ───────────────────────────────────────────────
const showModal = ref(false)
const editingId = ref<string | null>(null)
const modalStep = ref<'info' | 'items' | 'pricing'>('info')

interface ServiceForm {
  name: string; serviceType: RecurringServiceType; description: string
  monthlyCost: number; monthlyPrice: number
  targetMarginPercent: number; billingFrequency: BillingFrequency; isActive: boolean
}

const defaultForm = (): ServiceForm => ({
  name: '', serviceType: 'guarding', description: '',
  monthlyCost: 0, monthlyPrice: 0,
  targetMarginPercent: 30, billingFrequency: 'monthly', isActive: true,
})

const form = ref<ServiceForm>(defaultForm())
const lineItems = ref<ServiceLineItem[]>([])

// ── Item Source Picker ───────────────────────────────────────
const activeSource = ref<ItemSource>('product')
const itemSearch = ref('')
const showItemDropdown = ref(false)

const productResults = computed(() => {
  const q = itemSearch.value.toLowerCase().trim()
  if (!q) return productsStore.activeProducts.slice(0, 12)
  return productsStore.activeProducts.filter(p =>
    p.name.toLowerCase().includes(q) || p.sku.toLowerCase().includes(q) ||
    p.manufacturerName.toLowerCase().includes(q) || p.categoryName.toLowerCase().includes(q)
  ).slice(0, 12)
})

const serviceResults = computed(() => {
  const q = itemSearch.value.toLowerCase().trim()
  const list = services.value.filter(s => s.isActive)
  if (!q) return list.slice(0, 12)
  return list.filter(s => s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q)).slice(0, 12)
})

function addProductItem(p: Product) {
  lineItems.value.push({
    id: uid(), source: 'product', sourceId: p.id,
    sku: p.sku, name: p.name, description: p.categoryName + ' - ' + p.manufacturerName,
    qty: 1, unitCost: p.landedCostSAR, unitPrice: p.sellingPrice,
  })
  itemSearch.value = ''
  showItemDropdown.value = false
  recalcTotals()
}

function addServiceItem(s: RecurringService) {
  lineItems.value.push({
    id: uid(), source: 'service', sourceId: s.id,
    sku: '', name: s.name, description: serviceTypeLabels[s.serviceType] + ' service',
    qty: 1, unitCost: s.monthlyCost, unitPrice: s.monthlyPrice,
  })
  itemSearch.value = ''
  showItemDropdown.value = false
  recalcTotals()
}

function addWriteInItem() {
  lineItems.value.push({
    id: uid(), source: 'write-in', sku: '', name: '', description: '',
    qty: 1, unitCost: 0, unitPrice: 0,
  })
  recalcTotals()
}

// ── Vendor Quote Upload ──────────────────────────────────────
const vendorFile = ref<File | null>(null)
const vendorFileName = ref('')
const vendorItems = ref<{ name: string; description: string; qty: number; unitCost: number; unitPrice: number }[]>([])
const showVendorPreview = ref(false)

function handleVendorUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  vendorFile.value = file
  vendorFileName.value = file.name

  const reader = new FileReader()
  reader.onload = () => {
    const text = reader.result as string
    parseVendorCSV(text)
  }
  reader.readAsText(file)
}

function parseVendorCSV(text: string) {
  const lines = text.split('\n').filter(l => l.trim())
  const items: typeof vendorItems.value = []

  for (let i = 1; i < lines.length; i++) {
    const cols = (lines[i] ?? '').split(',').map(c => c.trim().replace(/^"|"$/g, ''))
    if (cols.length >= 3) {
      items.push({
        name: cols[0] || `Item ${i}`,
        description: cols[1] || '',
        qty: parseInt(cols[2] ?? '0') || 1,
        unitCost: parseFloat(cols[3] ?? '0') || 0,
        unitPrice: parseFloat(cols[4] ?? '0') || parseFloat(cols[3] ?? '0') || 0,
      })
    }
  }

  if (items.length === 0) {
    items.push({ name: '', description: '', qty: 1, unitCost: 0, unitPrice: 0 })
  }

  vendorItems.value = items
  showVendorPreview.value = true
}

function addVendorManualRow() {
  vendorItems.value.push({ name: '', description: '', qty: 1, unitCost: 0, unitPrice: 0 })
}

function removeVendorRow(idx: number) {
  vendorItems.value.splice(idx, 1)
}

function importVendorItems() {
  for (const vi of vendorItems.value) {
    if (!vi.name) continue
    lineItems.value.push({
      id: uid(), source: 'vendor', sku: '', name: vi.name, description: vi.description,
      qty: vi.qty, unitCost: vi.unitCost, unitPrice: vi.unitPrice,
      vendorName: vendorFileName.value,
    })
  }
  vendorItems.value = []
  vendorFile.value = null
  vendorFileName.value = ''
  showVendorPreview.value = false
  recalcTotals()
}

function startManualVendorEntry() {
  vendorItems.value = [{ name: '', description: '', qty: 1, unitCost: 0, unitPrice: 0 }]
  vendorFileName.value = 'Manual Entry'
  showVendorPreview.value = true
}

// ── Line Item Management ─────────────────────────────────────
function removeLineItem(id: string) {
  lineItems.value = lineItems.value.filter(i => i.id !== id)
  recalcTotals()
}

function recalcTotals() {
  if (lineItems.value.length === 0) return
  form.value.monthlyCost = lineItems.value.reduce((s, i) => s + lineCost(i), 0)
  form.value.monthlyPrice = lineItems.value.reduce((s, i) => s + lineTotal(i), 0)
  if (form.value.monthlyPrice > 0) {
    form.value.targetMarginPercent = Math.round(((form.value.monthlyPrice - form.value.monthlyCost) / form.value.monthlyPrice) * 1000) / 10
  }
}

const itemsTotalCost = computed(() => lineItems.value.reduce((s, i) => s + lineCost(i), 0))
const itemsTotalPrice = computed(() => lineItems.value.reduce((s, i) => s + lineTotal(i), 0))
const itemsTotalMargin = computed(() => itemsTotalPrice.value > 0 ? ((itemsTotalPrice.value - itemsTotalCost.value) / itemsTotalPrice.value) * 100 : 0)

// ── Computed Pricing ─────────────────────────────────────────
const formAnnualCost = computed(() => form.value.monthlyCost * 12)
const formAnnualPrice = computed(() => form.value.monthlyPrice * 12)
const formMargin = computed(() => form.value.monthlyPrice > 0 ? ((form.value.monthlyPrice - form.value.monthlyCost) / form.value.monthlyPrice) * 100 : 0)
const formMarginAmount = computed(() => form.value.monthlyPrice - form.value.monthlyCost)
const formAnnualMargin = computed(() => formAnnualPrice.value - formAnnualCost.value)

function calcPriceFromMargin() {
  if (form.value.targetMarginPercent > 0 && form.value.targetMarginPercent < 100 && form.value.monthlyCost > 0) {
    form.value.monthlyPrice = Math.ceil(form.value.monthlyCost / (1 - form.value.targetMarginPercent / 100))
  }
}

function marginColor(pct: number): string {
  if (pct >= 30) return 'margin-high'
  if (pct >= 20) return 'margin-mid'
  return 'margin-low'
}

function delayHideDropdown() { window.setTimeout(() => { showItemDropdown.value = false }, 200) }

// ── CRUD ─────────────────────────────────────────────────────
function openAddModal() {
  editingId.value = null
  form.value = defaultForm()
  lineItems.value = []
  modalStep.value = 'info'
  vendorItems.value = []
  showVendorPreview.value = false
  vendorFileName.value = ''
  showModal.value = true
}

function openEditModal(s: RecurringService) {
  editingId.value = s.id
  form.value = {
    name: s.name, serviceType: s.serviceType, description: s.description,
    monthlyCost: s.monthlyCost, monthlyPrice: s.monthlyPrice,
    targetMarginPercent: s.targetMarginPercent, billingFrequency: s.billingFrequency, isActive: s.isActive,
  }
  lineItems.value = serviceItems.value[s.id] ? JSON.parse(JSON.stringify(serviceItems.value[s.id])) : []
  modalStep.value = 'info'
  vendorItems.value = []
  showVendorPreview.value = false
  vendorFileName.value = ''
  showModal.value = true
}

const saving = ref(false)
async function saveService() {
 if (!form.value.name || saving.value) return
 saving.value = true
 try { const data = { ...form.value, lineItems:lineItems.value }; if (editingId.value) await recurringServicesService.update(editingId.value, data); else await recurringServicesService.create(data); await reloadServices(); showModal.value = false }
 catch (e) { window.alert(errorMessage(e)) } finally { saving.value = false }
}
async function deleteService(id: string) { try { await recurringServicesService.delete(id); await reloadServices() } catch (e) { window.alert(errorMessage(e)) } }
async function duplicateService(s: RecurringService) { try { await recurringServicesService.create({ ...s, name:s.name + ' (Copy)' }); await reloadServices() } catch (e) { window.alert(errorMessage(e)) } }
async function toggleActive(s: RecurringService) { try { await recurringServicesService.update(s.id, { isActive:!s.isActive }); await reloadServices() } catch (e) { window.alert(errorMessage(e)) } }

function getItemCount(sId: string): number { return serviceItems.value[sId]?.length ?? 0 }

const sourceLabels: Record<ItemSource, string> = { product: 'Product', service: 'Service', vendor: 'Vendor', 'write-in': 'Write-in' }
const sourceColors: Record<ItemSource, string> = { product: 'src-product', service: 'src-service', vendor: 'src-vendor', 'write-in': 'src-writein' }
</script>

<template>
  <div class="rs-page">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Recurring Services</h1>
        <p class="page-header-subtitle">{{ filteredServices.length }} of {{ services.length }} service{{ services.length !== 1 ? 's' : '' }}</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal"><Plus :size="18" /> Add Service</button>
    </div>

    <!-- KPI Cards -->
    <div class="kpi-grid">
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Active Services</p>
            <p class="kpi-value">{{ activeServices.length }}</p>
            <p class="kpi-sub">of {{ services.length }} total</p>
          </div>
          <div class="kpi-icon kpi-icon--primary"><RefreshCw :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Monthly Revenue</p>
            <p class="kpi-value">SAR {{ fmtCompact(totalMonthlyRevenue) }}</p>
            <p class="kpi-sub">{{ activeServices.length }} active services</p>
          </div>
          <div class="kpi-icon kpi-icon--success"><DollarSign :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Annual Revenue</p>
            <p class="kpi-value">SAR {{ fmtCompact(totalAnnualRevenue) }}</p>
            <p class="kpi-sub">projected annually</p>
          </div>
          <div class="kpi-icon kpi-icon--info"><TrendingUp :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Avg Margin</p>
            <p class="kpi-value" :class="marginColor(avgMargin)">{{ avgMargin.toFixed(1) }}%</p>
            <p class="kpi-sub">across active services</p>
          </div>
          <div class="kpi-icon kpi-icon--warning"><TrendingUp :size="22" /></div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="card filter-card">
      <div class="filter-row">
        <div class="search-input filter-search">
          <Search :size="16" class="search-icon" />
          <input v-model="searchQuery" type="text" class="form-input" placeholder="Search services..." />
        </div>
        <select v-model="typeFilter" class="form-select filter-select">
          <option value="all">All Types</option>
          <option v-for="(label, key) in serviceTypeLabels" :key="key" :value="key">{{ label }}</option>
        </select>
        <select v-model="statusFilter" class="form-select filter-select">
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
        </select>
      </div>
    </div>

    <!-- Services Table -->
    <div v-if="filteredServices.length" class="card">
      <div class="table-wrap">
        <table class="rs-table">
          <thead>
            <tr>
              <th class="col-type"></th>
              <th>Service Name</th>
              <th>Type</th>
              <th class="text-center">Items</th>
              <th class="text-right">Monthly Cost</th>
              <th class="text-right">Monthly Price</th>
              <th class="text-right">Annual Price</th>
              <th class="text-right">Margin</th>
              <th>Billing</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in filteredServices" :key="s.id" class="service-row" :class="{ 'service-row--inactive': !s.isActive }">
              <td class="col-type">
                <div :class="['type-icon', serviceTypeColors[s.serviceType]]">
                  <component :is="serviceTypeIcons[s.serviceType]" :size="16" />
                </div>
              </td>
              <td>
                <div class="name-cell">
                  <span class="name-main">{{ s.name }}</span>
                  <span class="name-desc">{{ s.description.substring(0, 80) }}{{ s.description.length > 80 ? '...' : '' }}</span>
                </div>
              </td>
              <td><span :class="['badge', 'badge-' + serviceTypeColors[s.serviceType]]">{{ serviceTypeLabels[s.serviceType] }}</span></td>
              <td class="text-center">
                <span v-if="getItemCount(s.id)" class="badge badge-primary">{{ getItemCount(s.id) }}</span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="text-right text-mono">SAR {{ formatSAR(s.monthlyCost) }}</td>
              <td class="text-right text-mono font-medium">SAR {{ formatSAR(s.monthlyPrice) }}</td>
              <td class="text-right text-mono font-medium">SAR {{ formatSAR(s.annualPrice) }}</td>
              <td class="text-right">
                <span :class="['font-semibold', marginColor(s.targetMarginPercent)]">{{ s.targetMarginPercent.toFixed(1) }}%</span>
              </td>
              <td><span class="badge badge-gray">{{ billingLabels[s.billingFrequency] }}</span></td>
              <td>
                <button :class="['status-toggle', s.isActive ? 'status-toggle--active' : 'status-toggle--inactive']" @click="toggleActive(s)">
                  <span class="toggle-dot" />
                  <span class="toggle-label">{{ s.isActive ? 'Active' : 'Inactive' }}</span>
                </button>
              </td>
              <td>
                <div class="row-actions">
                  <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEditModal(s)"><Pencil :size="14" /></button>
                  <button class="btn btn-ghost btn-icon btn-sm" title="Duplicate" @click="duplicateService(s)"><Copy :size="14" /></button>
                  <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deleteService(s.id)"><Trash2 :size="14" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-else class="empty-state">
      <RefreshCw :size="40" class="empty-state-icon" />
      <h3 class="empty-state-title">No services found</h3>
      <p class="empty-state-text">{{ searchQuery || typeFilter !== 'all' || statusFilter !== 'all' ? 'Try adjusting your filters.' : 'Add your first recurring service.' }}</p>
    </div>

    <!-- ═══════════════ ADD / EDIT MODAL ═══════════════ -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <h2 class="modal-title">{{ editingId ? 'Edit Service' : 'Add Recurring Service' }}</h2>
            <div class="modal-steps">
              <button :class="['step-btn', modalStep === 'info' && 'step-btn--active']" @click="modalStep = 'info'">
                <span class="step-num">1</span> Service Info
              </button>
              <ArrowRight :size="14" class="step-arrow" />
              <button :class="['step-btn', modalStep === 'items' && 'step-btn--active']" @click="modalStep = 'items'">
                <span class="step-num">2</span> Items & Sources
                <span v-if="lineItems.length" class="step-badge">{{ lineItems.length }}</span>
              </button>
              <ArrowRight :size="14" class="step-arrow" />
              <button :class="['step-btn', modalStep === 'pricing' && 'step-btn--active']" @click="modalStep = 'pricing'">
                <span class="step-num">3</span> Pricing
              </button>
            </div>
            <button class="modal-close" @click="showModal = false"><X :size="20" /></button>
          </div>

          <div class="modal-body modal-body--tall">
            <!-- ── STEP 1: Service Info ── -->
            <div v-if="modalStep === 'info'" class="step-panel">
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">Service Name <span class="required">*</span></label>
                  <input v-model="form.name" type="text" class="form-input" placeholder="e.g. 24/7 Security Guard Service" />
                </div>
                <div class="form-group">
                  <label class="form-label">Service Type</label>
                  <select v-model="form.serviceType" class="form-select">
                    <option v-for="(label, key) in serviceTypeLabels" :key="key" :value="key">{{ label }}</option>
                  </select>
                </div>
              </div>
              <div class="form-group">
                <label class="form-label">Description</label>
                <textarea v-model="form.description" class="form-textarea" rows="3" placeholder="Service scope and details..." />
              </div>
              <div class="form-row">
                <div class="form-group">
                  <label class="form-label">Billing Frequency</label>
                  <select v-model="form.billingFrequency" class="form-select">
                    <option v-for="(label, key) in billingLabels" :key="key" :value="key">{{ label }}</option>
                  </select>
                </div>
                <div class="form-group">
                  <label class="form-label">Status</label>
                  <label class="form-checkbox-label">
                    <input v-model="form.isActive" type="checkbox" class="form-checkbox" />
                    <span>Active</span>
                  </label>
                </div>
              </div>

              <div class="step-nav">
                <span />
                <button class="btn btn-primary" :disabled="!form.name" @click="modalStep = 'items'">
                  Next: Add Items <ArrowRight :size="16" />
                </button>
              </div>
            </div>

            <!-- ── STEP 2: Items & Sources ── -->
            <div v-if="modalStep === 'items'" class="step-panel">
              <!-- Source Tabs -->
              <div class="source-tabs">
                <button :class="['source-tab', activeSource === 'product' && 'source-tab--active']" @click="activeSource = 'product'; showVendorPreview = false">
                  <Package :size="16" /> Products
                </button>
                <button :class="['source-tab', activeSource === 'service' && 'source-tab--active']" @click="activeSource = 'service'; showVendorPreview = false">
                  <RefreshCw :size="16" /> Services
                </button>
                <button :class="['source-tab', activeSource === 'vendor' && 'source-tab--active']" @click="activeSource = 'vendor'">
                  <Upload :size="16" /> Vendor Quote
                </button>
                <button :class="['source-tab', activeSource === 'write-in' && 'source-tab--active']" @click="activeSource = 'write-in'; showVendorPreview = false">
                  <PenLine :size="16" /> Write-in
                </button>
              </div>

              <!-- ─ PRODUCT PICKER ─ -->
              <div v-if="activeSource === 'product'" class="source-panel">
                <div class="picker-search-wrap">
                  <Search :size="16" class="picker-search-icon" />
                  <input
                    v-model="itemSearch" type="text" class="form-input picker-search"
                    placeholder="Search by name, SKU, manufacturer, category..."
                    @focus="showItemDropdown = true" @blur="delayHideDropdown"
                  />
                </div>
                <div v-if="showItemDropdown && productResults.length" class="picker-dropdown">
                  <div
                    v-for="p in productResults" :key="p.id"
                    class="picker-option" @mousedown.prevent="addProductItem(p)"
                  >
                    <div class="picker-opt-main">
                      <span class="picker-opt-sku">{{ p.sku }}</span>
                      <span class="picker-opt-name">{{ p.name }}</span>
                    </div>
                    <div class="picker-opt-meta">
                      <span class="badge badge-gray">{{ p.categoryName }}</span>
                      <span class="text-mono">SAR {{ formatSAR(p.sellingPrice) }}</span>
                    </div>
                  </div>
                </div>
                <p class="picker-hint">Click an item to add it as a recurring line item. You can adjust quantity and pricing after adding.</p>
              </div>

              <!-- ─ SERVICE PICKER ─ -->
              <div v-if="activeSource === 'service'" class="source-panel">
                <div class="picker-search-wrap">
                  <Search :size="16" class="picker-search-icon" />
                  <input
                    v-model="itemSearch" type="text" class="form-input picker-search"
                    placeholder="Search existing recurring services..."
                    @focus="showItemDropdown = true" @blur="delayHideDropdown"
                  />
                </div>
                <div v-if="showItemDropdown && serviceResults.length" class="picker-dropdown">
                  <div
                    v-for="s in serviceResults" :key="s.id"
                    class="picker-option" @mousedown.prevent="addServiceItem(s)"
                  >
                    <div class="picker-opt-main">
                      <component :is="serviceTypeIcons[s.serviceType]" :size="14" />
                      <span class="picker-opt-name">{{ s.name }}</span>
                    </div>
                    <div class="picker-opt-meta">
                      <span :class="['badge', 'badge-' + serviceTypeColors[s.serviceType]]">{{ serviceTypeLabels[s.serviceType] }}</span>
                      <span class="text-mono">SAR {{ formatSAR(s.monthlyPrice) }}/mo</span>
                    </div>
                  </div>
                </div>
                <p class="picker-hint">Add existing recurring services as components of this service package.</p>
              </div>

              <!-- ─ VENDOR QUOTE UPLOAD ─ -->
              <div v-if="activeSource === 'vendor'" class="source-panel">
                <div v-if="!showVendorPreview" class="vendor-upload-area">
                  <div class="upload-zone">
                    <Upload :size="32" />
                    <p class="upload-title">Upload Vendor Quote</p>
                    <p class="upload-sub">CSV or Excel file with item details (Name, Description, Qty, Cost, Price)</p>
                    <label class="btn btn-secondary upload-btn">
                      <FileText :size="16" /> Choose File
                      <input type="file" accept=".csv,.xlsx,.xls,.txt" class="sr-only" @change="handleVendorUpload" />
                    </label>
                  </div>
                  <div class="upload-divider">
                    <span class="upload-divider-text">OR</span>
                  </div>
                  <button class="btn btn-secondary" @click="startManualVendorEntry">
                    <PenLine :size="16" /> Enter Items Manually
                  </button>
                </div>

                <!-- Vendor Items Preview / Manual Entry -->
                <div v-if="showVendorPreview" class="vendor-preview">
                  <div class="vendor-preview-header">
                    <div>
                      <h4 class="vendor-preview-title">
                        <FileText :size="16" />
                        {{ vendorFileName || 'Vendor Items' }}
                      </h4>
                      <p class="vendor-preview-sub">{{ vendorItems.length }} item{{ vendorItems.length !== 1 ? 's' : '' }} — edit below then import</p>
                    </div>
                    <button class="btn btn-ghost btn-sm" @click="showVendorPreview = false; vendorItems = []">
                      <X :size="14" /> Cancel
                    </button>
                  </div>

                  <div class="vendor-table-wrap">
                    <table class="vendor-table">
                      <thead>
                        <tr>
                          <th>Name</th>
                          <th>Description</th>
                          <th class="col-sm">Qty</th>
                          <th class="col-sm">Unit Cost</th>
                          <th class="col-sm">Unit Price</th>
                          <th class="col-xs"></th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(vi, idx) in vendorItems" :key="idx">
                          <td><input v-model="vi.name" type="text" class="form-input form-input--sm" placeholder="Item name" /></td>
                          <td><input v-model="vi.description" type="text" class="form-input form-input--sm" placeholder="Description" /></td>
                          <td class="col-sm"><input v-model.number="vi.qty" type="number" min="1" class="form-input form-input--sm" /></td>
                          <td class="col-sm"><input v-model.number="vi.unitCost" type="number" step="10" min="0" class="form-input form-input--sm" /></td>
                          <td class="col-sm"><input v-model.number="vi.unitPrice" type="number" step="10" min="0" class="form-input form-input--sm" /></td>
                          <td class="col-xs"><button class="btn btn-ghost btn-icon btn-sm" @click="removeVendorRow(idx)"><Trash2 :size="14" /></button></td>
                        </tr>
                      </tbody>
                    </table>
                  </div>

                  <div class="vendor-actions">
                    <button class="btn btn-secondary btn-sm" @click="addVendorManualRow"><Plus :size="14" /> Add Row</button>
                    <button class="btn btn-primary btn-sm" @click="importVendorItems" :disabled="!vendorItems.some(v => v.name)">
                      <CheckCircle2 :size="14" /> Import {{ vendorItems.filter(v => v.name).length }} Item{{ vendorItems.filter(v => v.name).length !== 1 ? 's' : '' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- ─ WRITE-IN ─ -->
              <div v-if="activeSource === 'write-in'" class="source-panel">
                <p class="picker-hint mb-3">Add a custom line item with your own description and pricing.</p>
                <button class="btn btn-secondary" @click="addWriteInItem"><Plus :size="16" /> Add Write-in Item</button>
              </div>

              <!-- ══ LINE ITEMS TABLE ══ -->
              <div class="items-section">
                <h4 class="items-title">
                  Line Items
                  <span v-if="lineItems.length" class="items-count">{{ lineItems.length }}</span>
                </h4>

                <div v-if="lineItems.length" class="items-table-wrap">
                  <table class="items-table">
                    <thead>
                      <tr>
                        <th class="col-src">Source</th>
                        <th>Name / Description</th>
                        <th class="col-num">Qty</th>
                        <th class="col-num">Unit Cost</th>
                        <th class="col-num">Unit Price</th>
                        <th class="col-num text-right">Total</th>
                        <th class="col-num text-right">Margin</th>
                        <th class="col-xs"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="item in lineItems" :key="item.id" class="item-row">
                        <td class="col-src"><span :class="['badge badge-sm', sourceColors[item.source]]">{{ sourceLabels[item.source] }}</span></td>
                        <td>
                          <template v-if="item.source === 'write-in'">
                            <input v-model="item.name" type="text" class="form-input form-input--sm" placeholder="Item name" @change="recalcTotals" />
                            <input v-model="item.description" type="text" class="form-input form-input--sm form-input--sub" placeholder="Description" />
                          </template>
                          <template v-else>
                            <div class="item-name">{{ item.name }}</div>
                            <div class="item-desc">{{ item.sku ? item.sku + ' — ' : '' }}{{ item.description }}</div>
                          </template>
                        </td>
                        <td class="col-num"><input v-model.number="item.qty" type="number" min="1" class="form-input form-input--sm" @change="recalcTotals" /></td>
                        <td class="col-num"><input v-model.number="item.unitCost" type="number" step="10" min="0" class="form-input form-input--sm" @change="recalcTotals" /></td>
                        <td class="col-num"><input v-model.number="item.unitPrice" type="number" step="10" min="0" class="form-input form-input--sm" @change="recalcTotals" /></td>
                        <td class="col-num text-right text-mono font-medium">SAR {{ formatSAR(lineTotal(item)) }}</td>
                        <td class="col-num text-right"><span :class="['font-semibold', marginColor(lineMargin(item))]">{{ lineMargin(item).toFixed(1) }}%</span></td>
                        <td class="col-xs"><button class="btn btn-ghost btn-icon btn-sm" @click="removeLineItem(item.id)"><Trash2 :size="14" /></button></td>
                      </tr>
                    </tbody>
                    <tfoot>
                      <tr class="items-totals">
                        <td colspan="5" class="text-right font-semibold">Totals</td>
                        <td class="text-right text-mono font-semibold">SAR {{ formatSAR(itemsTotalPrice) }}</td>
                        <td class="text-right"><span :class="['font-semibold', marginColor(itemsTotalMargin)]">{{ itemsTotalMargin.toFixed(1) }}%</span></td>
                        <td></td>
                      </tr>
                    </tfoot>
                  </table>
                </div>

                <div v-else class="items-empty">
                  <Layers :size="32" />
                  <p>No items added yet. Choose a source above to add line items.</p>
                </div>
              </div>

              <div class="step-nav">
                <button class="btn btn-secondary" @click="modalStep = 'info'">Back</button>
                <button class="btn btn-primary" @click="recalcTotals(); modalStep = 'pricing'">
                  Next: Review Pricing <ArrowRight :size="16" />
                </button>
              </div>
            </div>

            <!-- ── STEP 3: Pricing Review ── -->
            <div v-if="modalStep === 'pricing'" class="step-panel">
              <div v-if="lineItems.length" class="pricing-notice">
                <AlertTriangle :size="16" />
                <span>Pricing auto-calculated from {{ lineItems.length }} line item{{ lineItems.length !== 1 ? 's' : '' }}. You can override below.</span>
              </div>

              <div class="pricing-calc">
                <div class="calc-inputs">
                  <div class="form-row">
                    <div class="form-group">
                      <label class="form-label">Monthly Cost (SAR)</label>
                      <input v-model.number="form.monthlyCost" type="number" step="100" min="0" class="form-input" @change="calcPriceFromMargin" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">Target Margin %</label>
                      <div class="margin-input-wrap">
                        <input v-model.number="form.targetMarginPercent" type="number" step="0.5" min="0" max="100" class="form-input" @change="calcPriceFromMargin" />
                        <button class="calc-btn" @click="calcPriceFromMargin" title="Calculate price from cost + margin">Calc</button>
                      </div>
                    </div>
                    <div class="form-group">
                      <label class="form-label">Monthly Price (SAR)</label>
                      <input v-model.number="form.monthlyPrice" type="number" step="100" min="0" class="form-input" />
                    </div>
                  </div>
                </div>

                <div class="calc-preview">
                  <div class="preview-grid">
                    <div class="preview-item">
                      <span class="preview-label">Monthly</span>
                      <div class="preview-row">
                        <span class="preview-sub">Cost</span>
                        <span class="text-mono">SAR {{ formatSAR(form.monthlyCost) }}</span>
                      </div>
                      <div class="preview-row">
                        <span class="preview-sub">Price</span>
                        <span class="text-mono font-semibold">SAR {{ formatSAR(form.monthlyPrice) }}</span>
                      </div>
                      <div class="preview-row preview-row--highlight">
                        <span class="preview-sub">Profit</span>
                        <span :class="['text-mono font-semibold', formMarginAmount >= 0 ? 'text-success' : 'text-danger']">SAR {{ formatSAR(formMarginAmount) }}</span>
                      </div>
                    </div>
                    <div class="preview-item">
                      <span class="preview-label">Annual (×12)</span>
                      <div class="preview-row">
                        <span class="preview-sub">Cost</span>
                        <span class="text-mono">SAR {{ formatSAR(formAnnualCost) }}</span>
                      </div>
                      <div class="preview-row">
                        <span class="preview-sub">Revenue</span>
                        <span class="text-mono font-semibold">SAR {{ formatSAR(formAnnualPrice) }}</span>
                      </div>
                      <div class="preview-row preview-row--highlight">
                        <span class="preview-sub">Profit</span>
                        <span :class="['text-mono font-semibold', formAnnualMargin >= 0 ? 'text-success' : 'text-danger']">SAR {{ formatSAR(formAnnualMargin) }}</span>
                      </div>
                    </div>
                    <div class="preview-item preview-item--margin">
                      <span class="preview-label">Actual Margin</span>
                      <span :class="['preview-margin-value', marginColor(formMargin)]">{{ formMargin.toFixed(1) }}%</span>
                      <div class="preview-margin-bar">
                        <div class="preview-margin-fill" :class="marginColor(formMargin)" :style="{ width: Math.min(formMargin, 50) * 2 + '%' }" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="step-nav">
                <button class="btn btn-secondary" @click="modalStep = 'items'">Back</button>
                <button class="btn btn-primary" :disabled="!form.name || form.monthlyPrice <= 0" @click="saveService">
                  <CheckCircle2 :size="16" />
                  {{ editingId ? 'Save Changes' : 'Add Service' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.rs-page { padding: var(--space-6); }

/* KPI Grid */
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--space-4); margin-bottom: var(--space-6); }
.kpi-card { padding: var(--space-5); }
.kpi-top { display: flex; justify-content: space-between; align-items: flex-start; }
.kpi-label { font-size: var(--text-sm); color: var(--color-neutral-500); margin-bottom: 2px; }
.kpi-value { font-size: var(--text-2xl); font-weight: 700; color: var(--color-neutral-900); line-height: 1.2; }
.kpi-sub { font-size: var(--text-xs); color: var(--color-neutral-400); margin-top: 2px; }
.kpi-icon { width: 44px; height: 44px; border-radius: var(--radius-lg); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.kpi-icon--primary { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); }
.kpi-icon--success { background: #d1fae5; color: #059669; }
.kpi-icon--info { background: #dbeafe; color: #2563eb; }
.kpi-icon--warning { background: #fef3c7; color: #d97706; }

/* Filters */
.filter-card { padding: var(--space-4) var(--space-5); margin-bottom: var(--space-5); }
.filter-row { display: flex; align-items: center; gap: var(--space-3); flex-wrap: wrap; }
.filter-search { flex: 1; min-width: 220px; }
.filter-select { width: 160px; flex-shrink: 0; }

/* Table */
.table-wrap { overflow-x: auto; }
.rs-table { width: 100%; border-collapse: separate; border-spacing: 0; }
.rs-table th { padding: 10px 14px; font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); background: var(--color-neutral-50); border-bottom: 1px solid var(--color-neutral-200); white-space: nowrap; text-align: left; }
.rs-table td { padding: 12px 14px; border-bottom: 1px solid var(--color-neutral-100); font-size: var(--text-sm); vertical-align: middle; }
.service-row { transition: background .15s; }
.service-row:hover { background: var(--color-neutral-50); }
.service-row--inactive { opacity: .55; }
.col-type { width: 44px; }
.text-muted { color: var(--color-neutral-300); }

.type-icon { width: 34px; height: 34px; border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; }
.type-guarding { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); }
.type-maintenance { background: #fef3c7; color: #d97706; }
.type-monitoring { background: #dbeafe; color: #2563eb; }
.type-patrol { background: #d1fae5; color: #059669; }
.type-fm { background: var(--color-neutral-100); color: var(--color-neutral-600); }

.badge-type-guarding { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); }
.badge-type-maintenance { background: #fef3c7; color: #92400e; }
.badge-type-monitoring { background: #dbeafe; color: #1e40af; }
.badge-type-patrol { background: #d1fae5; color: #065f46; }
.badge-type-fm { background: var(--color-neutral-100); color: var(--color-neutral-600); }
.badge-type-rental { background: #ede9fe; color: #5b21b6; }

.name-cell { display: flex; flex-direction: column; gap: 1px; }
.name-main { font-weight: 500; color: var(--color-neutral-800); }
.name-desc { font-size: var(--text-xs); color: var(--color-neutral-400); max-width: 320px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.text-mono { font-family: var(--font-mono); }
.font-medium { font-weight: 500; }
.font-semibold { font-weight: 600; }

.status-toggle { display: inline-flex; align-items: center; gap: 6px; padding: 3px 10px 3px 6px; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-full); background: var(--content-surface); cursor: pointer; font-size: var(--text-xs); font-weight: 500; transition: all .2s; }
.toggle-dot { width: 8px; height: 8px; border-radius: 50%; transition: background .2s; }
.status-toggle--active { border-color: #d1fae5; background: #ecfdf5; }
.status-toggle--active .toggle-dot { background: #059669; }
.status-toggle--active .toggle-label { color: #065f46; }
.status-toggle--inactive { border-color: var(--color-neutral-200); }
.status-toggle--inactive .toggle-dot { background: var(--color-neutral-400); }
.status-toggle--inactive .toggle-label { color: var(--color-neutral-500); }
.row-actions { display: flex; gap: 2px; }

.margin-high { color: #059669; }
.margin-mid { color: #d97706; }
.margin-low { color: #dc2626; }
.text-success { color: #059669; }
.text-danger { color: #dc2626; }

/* ═══ MODAL ═══ */
.modal-xl { max-width: 960px; width: 95vw; }
.modal-body--tall { min-height: 440px; max-height: 72vh; overflow-y: auto; }
.required { color: #dc2626; }

/* Steps */
.modal-steps { display: flex; align-items: center; gap: var(--space-2); margin-left: auto; margin-right: var(--space-4); }
.step-btn { display: flex; align-items: center; gap: 6px; padding: 6px 12px; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); background: white; cursor: pointer; font-size: var(--text-xs); font-weight: 500; color: var(--color-neutral-500); transition: all .15s; }
.step-btn--active { background: var(--color-primary); color: white; border-color: var(--color-primary); }
.step-num { width: 20px; height: 20px; border-radius: 50%; background: var(--color-neutral-100); display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }
.step-btn--active .step-num { background: rgba(255,255,255,.25); color: white; }
.step-badge { background: var(--color-primary); color: white; border-radius: var(--radius-full); padding: 1px 7px; font-size: 10px; font-weight: 700; }
.step-btn--active .step-badge { background: rgba(255,255,255,.3); }
.step-arrow { color: var(--color-neutral-300); flex-shrink: 0; }

.step-panel { display: flex; flex-direction: column; gap: var(--space-4); }
.step-nav { display: flex; justify-content: space-between; align-items: center; padding-top: var(--space-4); border-top: 1px solid var(--color-neutral-200); margin-top: var(--space-2); }

/* Source Tabs */
.source-tabs { display: flex; gap: var(--space-2); flex-wrap: wrap; }
.source-tab { display: flex; align-items: center; gap: 6px; padding: 8px 16px; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); background: white; cursor: pointer; font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-600); transition: all .15s; }
.source-tab:hover { border-color: var(--color-primary); color: var(--color-primary); }
.source-tab--active { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); border-color: var(--color-primary); }

.source-panel { padding: var(--space-3) 0; }

/* Picker */
.picker-search-wrap { position: relative; }
.picker-search-icon { position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: var(--color-neutral-400); pointer-events: none; }
.picker-search { padding-left: 36px !important; }

.picker-dropdown { position: relative; background: white; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); max-height: 260px; overflow-y: auto; margin-top: var(--space-1); box-shadow: var(--shadow-lg); z-index: 50; }
.picker-option { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; cursor: pointer; transition: background .1s; border-bottom: 1px solid var(--color-neutral-50); }
.picker-option:hover { background: var(--color-primary-light, #eff6ff); }
.picker-opt-main { display: flex; align-items: center; gap: var(--space-2); flex: 1; min-width: 0; }
.picker-opt-sku { font-size: var(--text-xs); font-family: var(--font-mono); color: var(--color-neutral-500); flex-shrink: 0; }
.picker-opt-name { font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-800); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.picker-opt-meta { display: flex; align-items: center; gap: var(--space-2); flex-shrink: 0; margin-left: var(--space-3); }
.picker-hint { font-size: var(--text-xs); color: var(--color-neutral-400); margin-top: var(--space-2); }

/* Vendor Upload */
.vendor-upload-area { display: flex; flex-direction: column; align-items: center; gap: var(--space-4); padding: var(--space-4) 0; }
.upload-zone { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); padding: var(--space-6) var(--space-8); border: 2px dashed var(--color-neutral-300); border-radius: var(--radius-lg); color: var(--color-neutral-400); text-align: center; width: 100%; }
.upload-title { font-size: var(--text-base); font-weight: 600; color: var(--color-neutral-600); margin: 0; }
.upload-sub { font-size: var(--text-xs); color: var(--color-neutral-400); margin: 0; }
.upload-btn { margin-top: var(--space-2); cursor: pointer; }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0,0,0,0); }
.upload-divider { display: flex; align-items: center; gap: var(--space-3); width: 100%; color: var(--color-neutral-400); }
.upload-divider::before, .upload-divider::after { content: ''; flex: 1; height: 1px; background: var(--color-neutral-200); }
.upload-divider-text { font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; }

.vendor-preview { background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); padding: var(--space-4); }
.vendor-preview-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: var(--space-3); }
.vendor-preview-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: 600; margin: 0; color: var(--color-neutral-700); }
.vendor-preview-sub { font-size: var(--text-xs); color: var(--color-neutral-400); margin: 2px 0 0; }

.vendor-table-wrap { overflow-x: auto; }
.vendor-table { width: 100%; border-collapse: collapse; font-size: var(--text-sm); }
.vendor-table th { padding: 6px 8px; font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; color: var(--color-neutral-500); text-align: left; border-bottom: 1px solid var(--color-neutral-300); }
.vendor-table td { padding: 4px 8px; }
.vendor-table .col-sm { width: 90px; }
.vendor-table .col-xs { width: 36px; }

.vendor-actions { display: flex; justify-content: space-between; margin-top: var(--space-3); }

.form-input--sm { padding: 5px 8px; font-size: var(--text-sm); }
.form-input--sub { margin-top: 4px; font-size: var(--text-xs); }

/* Items Section */
.items-section { background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); padding: var(--space-4); margin-top: var(--space-2); }
.items-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: 600; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); margin: 0 0 var(--space-3); }
.items-count { background: var(--color-primary); color: white; border-radius: var(--radius-full); padding: 1px 8px; font-size: 11px; font-weight: 700; }

.items-table-wrap { overflow-x: auto; }
.items-table { width: 100%; border-collapse: collapse; font-size: var(--text-sm); }
.items-table th { padding: 8px 10px; font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; color: var(--color-neutral-500); text-align: left; border-bottom: 1px solid var(--color-neutral-300); }
.items-table td { padding: 8px 10px; border-bottom: 1px solid var(--color-neutral-100); vertical-align: middle; }
.items-table .col-src { width: 80px; }
.items-table .col-num { width: 100px; }
.items-table .col-xs { width: 36px; }

.item-name { font-weight: 500; color: var(--color-neutral-800); }
.item-desc { font-size: var(--text-xs); color: var(--color-neutral-400); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 260px; }

.items-totals td { padding-top: 12px; border-top: 2px solid var(--color-neutral-300); font-size: var(--text-sm); }

.items-empty { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); padding: var(--space-6); color: var(--color-neutral-400); text-align: center; }
.items-empty p { font-size: var(--text-sm); margin: 0; }

/* Source badges */
.badge-sm { font-size: 10px; padding: 2px 6px; }
.src-product { background: #dbeafe; color: #1e40af; }
.src-service { background: #d1fae5; color: #065f46; }
.src-vendor { background: #fef3c7; color: #92400e; }
.src-writein { background: var(--color-neutral-100); color: var(--color-neutral-600); }

/* Pricing */
.pricing-notice { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-3); background: #fffbeb; border: 1px solid #fde68a; border-radius: var(--radius-md); font-size: var(--text-sm); color: #92400e; margin-bottom: var(--space-3); }

.pricing-calc { display: flex; flex-direction: column; gap: var(--space-4); }
.calc-preview { background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); padding: var(--space-4); }
.preview-grid { display: grid; grid-template-columns: 1fr 1fr auto; gap: var(--space-5); }
.preview-item { display: flex; flex-direction: column; gap: var(--space-2); }
.preview-label { font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); margin-bottom: var(--space-1); }
.preview-row { display: flex; justify-content: space-between; font-size: var(--text-sm); color: var(--color-neutral-700); }
.preview-row--highlight { padding-top: var(--space-2); border-top: 1px solid var(--color-neutral-200); margin-top: var(--space-1); }
.preview-sub { color: var(--color-neutral-500); font-size: var(--text-xs); }

.preview-item--margin { display: flex; flex-direction: column; align-items: center; justify-content: center; min-width: 100px; }
.preview-margin-value { font-size: var(--text-2xl); font-weight: 700; line-height: 1.2; margin: var(--space-2) 0; }
.preview-margin-bar { width: 100%; height: 6px; background: var(--color-neutral-200); border-radius: 3px; overflow: hidden; }
.preview-margin-fill { height: 100%; border-radius: 3px; transition: width .4s ease; }
.preview-margin-fill.margin-high { background: #059669; }
.preview-margin-fill.margin-mid { background: #d97706; }
.preview-margin-fill.margin-low { background: #dc2626; }

.margin-input-wrap { display: flex; gap: var(--space-2); }
.margin-input-wrap .form-input { flex: 1; }
.calc-btn { padding: 6px 12px; font-size: var(--text-xs); font-weight: 600; background: var(--color-primary); color: white; border: none; border-radius: var(--radius-md); cursor: pointer; white-space: nowrap; transition: background .15s; }
.calc-btn:hover { background: var(--color-primary-dark, #1d4ed8); }

.form-checkbox-label { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); cursor: pointer; padding-top: var(--space-6); }
.form-checkbox { width: 16px; height: 16px; accent-color: var(--color-primary); }

.mb-3 { margin-bottom: var(--space-3); }
.mb-4 { margin-bottom: var(--space-4); }

@media (max-width: 1200px) { .kpi-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .kpi-grid { grid-template-columns: 1fr; }
  .filter-row { flex-direction: column; align-items: stretch; }
  .filter-select { width: 100%; }
  .preview-grid { grid-template-columns: 1fr; }
  .modal-steps { display: none; }
}
</style>
