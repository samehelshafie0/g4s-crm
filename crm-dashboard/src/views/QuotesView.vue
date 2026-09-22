<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Plus,
  Search,
  Eye,
  Pencil,
  Trash2,
  X,
  FileText,
  Clock,
  CheckCircle2,
  TrendingUp,
  DollarSign,
  Percent,
  ExternalLink,
  Package,
  Users,
  Wrench,
  Boxes,
  Copy,
  Lock,
  History,
  Target,
  BookOpen,
  ChevronRight,
  Building2,
} from 'lucide-vue-next'
import type { Quote, QuoteStatus, QuoteLineItem, QuoteLineCategory } from '@/types'
import { useOpportunitiesStore } from '@/stores/opportunities'
import { usePriceBooksStore } from '@/stores/priceBooks'
import { useQuotesStore } from '@/stores/quotes'
import { useCustomersStore } from '@/stores/customers'

import { quotesService, customersService, opportunitiesService, priceBooksService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import type { Activity } from '@/services/workflowDtos'

const router = useRouter()
const route = useRoute()

function uid(): string {
  return Math.random().toString(36).slice(2, 11)
}

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function marginClass(pct: number): string {
  if (pct >= 25) return 'margin-high'
  if (pct >= 20) return 'margin-medium'
  return 'margin-low'
}

const statusConfig: Record<QuoteStatus, { label: string; badge: string }> = {
  draft: { label: 'Draft', badge: 'badge-gray' },
  'pending-approval': { label: 'Pending Approval', badge: 'badge-warning' },
  approved: { label: 'Approved', badge: 'badge-info' },
  sent: { label: 'Sent', badge: 'badge-primary' },
  accepted: { label: 'Accepted', badge: 'badge-success' },
  declined: { label: 'Declined', badge: 'badge-danger' },
  expired: { label: 'Expired', badge: 'badge-gray' },
}

const categoryIcons: Record<QuoteLineCategory, typeof Package> = {
  materials: Package,
  manpower: Users,
  miscellaneous: Wrench,
}

const categoryLabels: Record<QuoteLineCategory, string> = {
  materials: 'Materials',
  manpower: 'Manpower',
  miscellaneous: 'Miscellaneous',
}

const quotes = ref<Quote[]>([])

const searchQuery = ref('')
const statusFilter = ref<QuoteStatus | ''>('')
const dateFrom = ref('')
const dateTo = ref('')

const filteredQuotes = computed(() => {
  let list = quotes.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter(
      (qt) =>
        qt.quoteNumber.toLowerCase().includes(q) ||
        qt.customerName.toLowerCase().includes(q),
    )
  }
  if (statusFilter.value) {
    list = list.filter((qt) => qt.status === statusFilter.value)
  }
  if (dateFrom.value) {
    list = list.filter((qt) => qt.validUntil >= dateFrom.value)
  }
  if (dateTo.value) {
    list = list.filter((qt) => qt.validUntil <= dateTo.value)
  }
  return list
})

// ── KPI Stats ────────────────────────────────────────────────
const totalQuotes = computed(() => quotes.value.length)
const pendingApproval = computed(() =>
  quotes.value.filter((q) => q.status === 'pending-approval').length,
)
const acceptedQuotes = computed(() => {
  const accepted = quotes.value.filter((q) => q.status === 'accepted')
  return {
    count: accepted.length,
    value: accepted.reduce((s, q) => s + q.total, 0),
  }
})
const avgMargin = computed(() => {
  if (!quotes.value.length) return 0
  return quotes.value.reduce((s, q) => s + q.marginPercent, 0) / quotes.value.length
})

// ── View Modal ───────────────────────────────────────────────
const showViewModal = ref(false)
const viewingQuote = ref<Quote | null>(null)

async function openViewModal(q: Quote) {
  auditTrail.value = []
  try { const res = await quotesService.activity(q.id); auditTrail.value = res.data.map(a => ({ id: a.id, action: a.description, user: a.user ? `${a.user.firstName} ${a.user.lastName}` : 'System', timestamp: a.createdAt })) } catch (error) { window.alert(errorMessage(error)) }
  viewingQuote.value = q
  showViewModal.value = true
}

async function deleteQuote(id: string) {
  try { await quotesService.delete(id); quotes.value = quotes.value.filter(q => q.id !== id) }
  catch (error) { window.alert(errorMessage(error)) }
}

function linesByCategory(items: QuoteLineItem[], cat: QuoteLineCategory): QuoteLineItem[] {
  return items.filter((li) => li.category === cat)
}

// ── Stores ───────────────────────────────────────────────────
const oppStore = useOpportunitiesStore()
const pbStore = usePriceBooksStore()
const quotesStore = useQuotesStore()
const customerStore = useCustomersStore()

// ── New Quote Modal (multi-step) ─────────────────────────────
const showNewQuoteModal = ref(false)
type NewQuoteStep = 'customer' | 'opportunity' | 'pricebook' | 'confirm'
const newQuoteStep = ref<NewQuoteStep>('customer')

const customers = ref<{ id: string; name: string; storeId: string }[]>([])

const newQuoteForm = ref({
  customerId: '',
  customerName: '',
  customerStoreId: '',
  opportunityId: '',
  opportunityTitle: '',
  priceBookId: '',
  priceBookName: '',
  notes: '',
  validUntil: '',
})

const customerSearch = ref('')

const filteredCustomers = computed(() => {
  const q = customerSearch.value.toLowerCase().trim()
  if (!q) return customers.value
  return customers.value.filter(c => c.name.toLowerCase().includes(q))
})

const customerOpportunities = computed(() => {
  if (!newQuoteForm.value.customerStoreId) return []
  return oppStore.opportunities.filter(
    o => o.customerId === newQuoteForm.value.customerStoreId && !['closed-lost'].includes(o.stage),
  )
})

const customerPriceBooks = computed(() => {
  if (!newQuoteForm.value.customerStoreId) return []
  const customerSpecific = pbStore.priceBooks.filter(
    pb => pb.isActive && pb.customerId === newQuoteForm.value.customerStoreId,
  )
  const general = pbStore.priceBooks.filter(
    pb => pb.isActive && !pb.customerId,
  )
  return [...customerSpecific, ...general]
})

function resetNewQuoteForm() {
  newQuoteForm.value = { customerId: '', customerName: '', customerStoreId: '', opportunityId: '', opportunityTitle: '', priceBookId: '', priceBookName: '', notes: '', validUntil: '' }
  newQuoteStep.value = 'customer'
  customerSearch.value = ''
}

function selectCustomer(cust: { id: string; name: string; storeId: string }) {
  newQuoteForm.value.customerId = cust.id
  newQuoteForm.value.customerName = cust.name
  newQuoteForm.value.customerStoreId = cust.storeId
  newQuoteForm.value.opportunityId = ''
  newQuoteForm.value.opportunityTitle = ''
  newQuoteForm.value.priceBookId = ''
  newQuoteForm.value.priceBookName = ''
  newQuoteStep.value = 'opportunity'
}

function selectOpportunity(oppId: string, oppTitle: string) {
  newQuoteForm.value.opportunityId = oppId
  newQuoteForm.value.opportunityTitle = oppTitle
  newQuoteStep.value = 'pricebook'
}

function skipOpportunity() {
  newQuoteForm.value.opportunityId = ''
  newQuoteForm.value.opportunityTitle = ''
  newQuoteStep.value = 'pricebook'
}

function selectPriceBook(pbId: string, pbName: string) {
  newQuoteForm.value.priceBookId = pbId
  newQuoteForm.value.priceBookName = pbName
  newQuoteStep.value = 'confirm'
}

function skipPriceBook() {
  newQuoteForm.value.priceBookId = ''
  newQuoteForm.value.priceBookName = ''
  newQuoteStep.value = 'confirm'
}

function goBackToStep(step: NewQuoteStep) {
  newQuoteStep.value = step
}

const creating = ref(false)
async function createAndOpenQuote() {
  if (!newQuoteForm.value.customerId || creating.value) return
  creating.value = true
  try {
    const quote = await quotesStore.addQuote({
      customerId: newQuoteForm.value.customerId,
      opportunityId: newQuoteForm.value.opportunityId || undefined,
      priceBookId: newQuoteForm.value.priceBookId || undefined,
      currency: 'SAR', notes: newQuoteForm.value.notes,
      validUntil: newQuoteForm.value.validUntil || new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10),
    })
    showNewQuoteModal.value = false
    await router.push(`/quotes/${quote.id}/builder`)
  } catch (error) { window.alert(errorMessage(error)) }
  finally { creating.value = false }
}

function openNewQuoteFromOpp(oppId: string) {
  const opp = oppStore.opportunities.find(o => o.id === oppId)
  if (!opp) return
  const cust = customers.value.find(c => c.storeId === opp.customerId)
  if (!cust) return

  showNewQuoteModal.value = true
  selectCustomer(cust)
  selectOpportunity(opp.id, opp.title)
}

function openBuilder(quoteId: string) {
  router.push(`/quotes/${quoteId}/builder`)
}

onMounted(async () => {
  try {
    const [quoteRows, customerRows, opportunityRows, bookRows] = await Promise.all([
      allPages(quotesService.list), auth.can('customers:read') ? allPages(customersService.list) : [], auth.can('opportunities:read') ? allPages(opportunitiesService.list) : [], auth.can('price-books:read') ? allPages(priceBooksService.list) : [],
    ])
    quotes.value = quoteRows
    customers.value = customerRows.map(c => ({ id: c.id, name: c.companyName, storeId: c.id }))
    oppStore.opportunities = opportunityRows
    pbStore.priceBooks = bookRows
    const oppId = route.query.newFromOpp as string | undefined
    if (oppId) { openNewQuoteFromOpp(oppId); await router.replace({ path: '/quotes' }) }
  } catch (error) { window.alert(errorMessage(error)) }
})

// ── Document Locking ─────────────────────────────────────────
function isLocked(q: Quote): boolean {
  return q.status !== 'draft'
}

// ── Duplicate Quote ──────────────────────────────────────────
async function duplicateQuote(q: Quote) {
  try { const res = await quotesService.duplicate(q.id); quotes.value.unshift(res.data) }
  catch (error) { window.alert(errorMessage(error)) }
}

// ── Audit Trail ──────────────────────────────────────────────
interface AuditEntry {
  id: string
  action: string
  field?: string
  oldValue?: string
  newValue?: string
  user: string
  timestamp: string
}

const viewModalTab = ref<'details' | 'audit'>('details')

const auditTrail = ref<AuditEntry[]>([])

function formatTimestamp(ts: string): string {
  return new Date(ts).toLocaleString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="quotes-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Quotations</h1>
        <p class="page-header-subtitle">
          {{ filteredQuotes.length }} quote{{ filteredQuotes.length !== 1 ? 's' : '' }}
        </p>
      </div>
      <button v-if="auth.can('quotes:create')" class="btn btn-primary" @click="showNewQuoteModal = true">
        <Plus :size="18" />
        New Quote
      </button>
    </div>

    <!-- KPI Stats Row -->
    <div class="kpi-grid">
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Total Quotes</p>
            <p class="kpi-value">{{ totalQuotes }}</p>
          </div>
          <div class="kpi-icon kpi-icon--primary">
            <FileText :size="22" />
          </div>
        </div>
      </div>

      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Pending Approval</p>
            <p class="kpi-value">{{ pendingApproval }}</p>
          </div>
          <div class="kpi-icon kpi-icon--warning">
            <Clock :size="22" />
          </div>
        </div>
      </div>

      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Accepted</p>
            <p class="kpi-value">{{ acceptedQuotes.count }}</p>
            <p class="kpi-subtitle">SAR {{ formatSAR(acceptedQuotes.value) }}</p>
          </div>
          <div class="kpi-icon kpi-icon--success">
            <CheckCircle2 :size="22" />
          </div>
        </div>
      </div>

      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Average Margin</p>
            <p class="kpi-value" :class="marginClass(avgMargin)">{{ avgMargin.toFixed(1) }}%</p>
          </div>
          <div class="kpi-icon kpi-icon--info">
            <TrendingUp :size="22" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filter Toolbar -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search">
          <Search :size="18" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Search by quote # or customer…"
          />
        </div>

        <select v-model="statusFilter" class="form-select toolbar-select">
          <option value="">All Statuses</option>
          <option v-for="(cfg, key) in statusConfig" :key="key" :value="key">
            {{ cfg.label }}
          </option>
        </select>

        <input
          v-model="dateFrom"
          type="date"
          class="form-input toolbar-date"
          title="Valid from"
        />
        <input
          v-model="dateTo"
          type="date"
          class="form-input toolbar-date"
          title="Valid to"
        />
      </div>
    </div>

    <!-- Quotes Table -->
    <div v-if="filteredQuotes.length" class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Quote #</th>
            <th>Customer</th>
            <th>Status</th>
            <th class="text-center">Items</th>
            <th class="text-right">Subtotal</th>
            <th class="text-right">Disc %</th>
            <th class="text-right">VAT</th>
            <th class="text-right">Total</th>
            <th class="text-right">Margin</th>
            <th>Valid Until</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="q in filteredQuotes" :key="q.id">
            <td class="text-mono font-bold">{{ q.quoteNumber }}</td>
            <td class="font-medium text-dark">{{ q.customerName }}</td>
            <td>
              <span :class="['badge badge-dot', statusConfig[q.status].badge]">
                {{ statusConfig[q.status].label }}
              </span>
            </td>
            <td class="text-center">{{ q.lineItems.length }}</td>
            <td class="text-right whitespace-nowrap">SAR {{ formatSAR(q.subtotal) }}</td>
            <td class="text-right">{{ q.discountPercent.toFixed(1) }}%</td>
            <td class="text-right whitespace-nowrap">SAR {{ formatSAR(q.vatAmount) }}</td>
            <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(q.total) }}</td>
            <td class="text-right">
              <span :class="['font-semibold', marginClass(q.marginPercent)]">
                {{ q.marginPercent.toFixed(1) }}%
              </span>
            </td>
            <td class="text-muted whitespace-nowrap">{{ formatDate(q.validUntil) }}</td>
            <td>
              <div class="table-actions">
                <Lock v-if="isLocked(q)" :size="13" class="lock-indicator" title="Document locked" />
                <button class="btn btn-ghost btn-icon btn-sm" title="View" @click="openViewModal(q)">
                  <Eye :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" v-if="auth.can('quotes:update')" title="Edit" @click="openBuilder(q.id)" :disabled="isLocked(q)">
                  <Pencil :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" v-if="auth.can('quotes:create')" title="Duplicate" @click="duplicateQuote(q)">
                  <Copy :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Open Builder" @click="openBuilder(q.id)">
                  <ExternalLink :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" v-if="auth.can('quotes:delete')" title="Delete" @click="deleteQuote(q.id)" :disabled="isLocked(q)">
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="empty-state">
      <FileText :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No quotations found</h3>
      <p class="empty-state-text">
        {{ searchQuery || statusFilter || dateFrom || dateTo
          ? 'Try adjusting your filters.'
          : 'Create your first quotation to get started.' }}
      </p>
    </div>

    <!-- View Quote Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && viewingQuote" class="modal-backdrop" @click.self="showViewModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <div class="modal-header-left">
              <h2 class="modal-title">{{ viewingQuote.quoteNumber }}</h2>
              <span :class="['badge badge-dot', statusConfig[viewingQuote.status].badge]">
                {{ statusConfig[viewingQuote.status].label }}
              </span>
            </div>
            <button class="modal-close" @click="showViewModal = false">
              <X :size="20" />
            </button>
          </div>

          <div class="modal-body">
            <!-- Modal Tabs -->
            <div class="view-modal-tabs">
              <button :class="['view-modal-tab', viewModalTab === 'details' && 'view-modal-tab--active']" @click="viewModalTab = 'details'">
                <FileText :size="14" /> Details
              </button>
              <button :class="['view-modal-tab', viewModalTab === 'audit' && 'view-modal-tab--active']" @click="viewModalTab = 'audit'">
                <History :size="14" /> Audit Trail
              </button>
            </div>

            <div v-show="viewModalTab === 'details'">
            <!-- Customer & Quote Info -->
            <div class="view-info-grid">
              <div class="view-info-item">
                <span class="view-info-label">Customer</span>
                <span class="view-info-value font-semibold">{{ viewingQuote.customerName }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Currency</span>
                <span class="view-info-value">{{ viewingQuote.currency }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Valid Until</span>
                <span class="view-info-value">{{ formatDate(viewingQuote.validUntil) }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Version</span>
                <span class="view-info-value">v{{ viewingQuote.version }}</span>
              </div>
            </div>

            <!-- Notes -->
            <div v-if="viewingQuote.notes" class="view-notes">
              <p>{{ viewingQuote.notes }}</p>
            </div>

            <!-- Line Items by Category -->
            <template v-for="cat in (['materials', 'manpower', 'miscellaneous'] as QuoteLineCategory[])" :key="cat">
              <div v-if="linesByCategory(viewingQuote.lineItems, cat).length" class="view-category-section">
                <h4 class="view-category-title">
                  <component :is="categoryIcons[cat]" :size="16" />
                  {{ categoryLabels[cat] }}
                  <span class="view-category-count">{{ linesByCategory(viewingQuote.lineItems, cat).length }}</span>
                </h4>
                <div class="table-container table-container--embedded">
                  <table class="table">
                    <thead>
                      <tr>
                        <th>#</th>
                        <th>Description</th>
                        <th v-if="cat === 'materials'">SKU</th>
                        <th v-if="cat === 'materials'">Manufacturer</th>
                        <th class="text-center">Qty</th>
                        <th class="text-right">Unit Price</th>
                        <th class="text-right">Total</th>
                        <th class="text-right">Margin</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="(item, idx) in linesByCategory(viewingQuote.lineItems, cat)"
                        :key="item.id"
                      >
                        <td class="text-muted">{{ idx + 1 }}</td>
                        <td class="font-medium">{{ item.description }}</td>
                        <td v-if="cat === 'materials'" class="text-mono">{{ item.sku ?? '—' }}</td>
                        <td v-if="cat === 'materials'">{{ item.manufacturerName ?? '—' }}</td>
                        <td class="text-center">{{ item.quantity }}</td>
                        <td class="text-right whitespace-nowrap">SAR {{ formatSAR(item.unitPrice) }}</td>
                        <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(item.lineTotal) }}</td>
                        <td class="text-right">
                          <span :class="['font-semibold', marginClass(item.marginPercent)]">
                            {{ item.marginPercent.toFixed(1) }}%
                          </span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </template>

            <!-- Totals Breakdown -->
            <div class="view-totals-card">
              <div class="view-totals-row">
                <span>Subtotal</span>
                <span class="text-mono">SAR {{ formatSAR(viewingQuote.subtotal) }}</span>
              </div>
              <div v-if="viewingQuote.discountPercent > 0" class="view-totals-row">
                <span>Discount ({{ viewingQuote.discountPercent.toFixed(1) }}%)</span>
                <span class="text-mono text-danger">- SAR {{ formatSAR(viewingQuote.discountAmount) }}</span>
              </div>
              <div v-if="viewingQuote.discountPercent > 0" class="view-totals-row">
                <span>Subtotal After Discount</span>
                <span class="text-mono">SAR {{ formatSAR(viewingQuote.subtotalAfterDiscount) }}</span>
              </div>
              <div class="view-totals-row">
                <span>VAT ({{ viewingQuote.vatPercent }}%)</span>
                <span class="text-mono">SAR {{ formatSAR(viewingQuote.vatAmount) }}</span>
              </div>
              <div class="view-totals-row view-totals-row--grand">
                <span class="font-bold">Grand Total</span>
                <span class="font-bold text-mono view-grand-total">SAR {{ formatSAR(viewingQuote.total) }}</span>
              </div>
              <div class="view-totals-divider" />
              <div class="view-totals-row">
                <span>Total Cost</span>
                <span class="text-mono">SAR {{ formatSAR(viewingQuote.totalCost) }}</span>
              </div>
              <div class="view-totals-row">
                <span>Margin Amount</span>
                <span class="text-mono text-success">SAR {{ formatSAR(viewingQuote.marginAmount) }}</span>
              </div>
              <div class="view-totals-row">
                <span>Margin %</span>
                <span :class="['font-bold', marginClass(viewingQuote.marginPercent)]">
                  {{ viewingQuote.marginPercent.toFixed(1) }}%
                </span>
              </div>
            </div>

            <!-- Approval Info -->
            <div v-if="viewingQuote.approvedBy" class="view-approval">
              <CheckCircle2 :size="16" class="text-success" />
              <span>Approved by <strong>{{ viewingQuote.approvedBy }}</strong></span>
              <span v-if="viewingQuote.approvedAt" class="text-muted">
                on {{ formatDate(viewingQuote.approvedAt) }}
              </span>
            </div>
            </div>

            <!-- Audit Trail Tab -->
            <div v-show="viewModalTab === 'audit'" class="audit-trail">
              <div class="audit-timeline">
                <div v-for="entry in auditTrail" :key="entry.id" class="audit-entry">
                  <div class="audit-dot" />
                  <div class="audit-content">
                    <div class="audit-header">
                      <span class="audit-action">{{ entry.action }}</span>
                      <span class="audit-time">{{ formatTimestamp(entry.timestamp) }}</span>
                    </div>
                    <div v-if="entry.field" class="audit-detail">
                      <span class="audit-field">{{ entry.field }}:</span>
                      <span v-if="entry.oldValue" class="audit-old">{{ entry.oldValue }}</span>
                      <span v-if="entry.oldValue && entry.newValue" class="audit-arrow">&rarr;</span>
                      <span v-if="entry.newValue" class="audit-new">{{ entry.newValue }}</span>
                    </div>
                    <div class="audit-user">by {{ entry.user }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showViewModal = false">Close</button>
            <button class="btn btn-primary" @click="showViewModal = false; openBuilder(viewingQuote!.id)">
              <ExternalLink :size="14" />
              Open in Builder
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- New Quote Modal (Multi-Step) -->
    <Teleport to="body">
      <div v-if="showNewQuoteModal" class="modal-backdrop" @click.self="showNewQuoteModal = false; resetNewQuoteForm()">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">Create New Quote</h2>
            <button class="modal-close" @click="showNewQuoteModal = false; resetNewQuoteForm()"><X :size="20" /></button>
          </div>

          <!-- Step Indicator -->
          <div class="nq-steps">
            <div :class="['nq-step', newQuoteStep === 'customer' && 'nq-step--active', newQuoteForm.customerId && 'nq-step--done']" @click="newQuoteForm.customerId ? goBackToStep('customer') : undefined">
              <span class="nq-step-num">1</span> Customer
            </div>
            <ChevronRight :size="14" class="nq-step-arrow" />
            <div :class="['nq-step', newQuoteStep === 'opportunity' && 'nq-step--active', (newQuoteStep === 'pricebook' || newQuoteStep === 'confirm') && 'nq-step--done']" @click="newQuoteForm.customerId ? goBackToStep('opportunity') : undefined">
              <span class="nq-step-num">2</span> Opportunity
            </div>
            <ChevronRight :size="14" class="nq-step-arrow" />
            <div :class="['nq-step', newQuoteStep === 'pricebook' && 'nq-step--active', newQuoteStep === 'confirm' && 'nq-step--done']">
              <span class="nq-step-num">3</span> Price Book
            </div>
            <ChevronRight :size="14" class="nq-step-arrow" />
            <div :class="['nq-step', newQuoteStep === 'confirm' && 'nq-step--active']">
              <span class="nq-step-num">4</span> Confirm
            </div>
          </div>

          <div class="modal-body">
            <!-- Step 1: Select Customer -->
            <div v-if="newQuoteStep === 'customer'">
              <div class="nq-section-title">Select Customer</div>
              <div class="search-input mb-4">
                <Search :size="16" class="search-icon" />
                <input v-model="customerSearch" type="text" class="form-input" placeholder="Search customers..." />
              </div>
              <div class="nq-list">
                <button v-for="c in filteredCustomers" :key="c.id" class="nq-list-item" @click="selectCustomer(c)">
                  <div class="nq-list-icon"><Building2 :size="16" /></div>
                  <div class="nq-list-info">
                    <span class="nq-list-name">{{ c.name }}</span>
                    <span class="nq-list-meta">
                      {{ oppStore.opportunities.filter(o => o.customerId === c.storeId && !['closed-lost'].includes(o.stage)).length }} active opportunities
                    </span>
                  </div>
                  <ChevronRight :size="16" class="nq-list-chevron" />
                </button>
              </div>
            </div>

            <!-- Step 2: Select Opportunity (optional) -->
            <div v-else-if="newQuoteStep === 'opportunity'">
              <div class="nq-section-title">
                Link to Opportunity
                <span class="nq-section-subtitle">(optional)</span>
              </div>
              <div class="nq-selected-badge">
                <Building2 :size="14" /> {{ newQuoteForm.customerName }}
              </div>

              <div v-if="customerOpportunities.length" class="nq-list">
                <button v-for="opp in customerOpportunities" :key="opp.id" class="nq-list-item" @click="selectOpportunity(opp.id, opp.title)">
                  <div class="nq-list-icon nq-list-icon--opp"><Target :size="16" /></div>
                  <div class="nq-list-info">
                    <span class="nq-list-name">{{ opp.title }}</span>
                    <span class="nq-list-meta">
                      {{ opp.stage.replace('-', ' ') }} · SAR {{ (opp.estimatedValue / 1000).toFixed(0) }}K · {{ opp.quoteIds.length }} quote{{ opp.quoteIds.length !== 1 ? 's' : '' }}
                    </span>
                  </div>
                  <ChevronRight :size="16" class="nq-list-chevron" />
                </button>
              </div>
              <div v-else class="nq-empty">
                <Target :size="24" class="text-muted" />
                <p>No active opportunities for this customer</p>
              </div>

              <button class="btn btn-ghost nq-skip-btn" @click="skipOpportunity">
                Skip — Create without an opportunity
              </button>
            </div>

            <!-- Step 3: Select Price Book (optional) -->
            <div v-else-if="newQuoteStep === 'pricebook'">
              <div class="nq-section-title">
                Select Price Book
                <span class="nq-section-subtitle">(optional)</span>
              </div>
              <div class="nq-selected-badges">
                <div class="nq-selected-badge"><Building2 :size="14" /> {{ newQuoteForm.customerName }}</div>
                <div v-if="newQuoteForm.opportunityTitle" class="nq-selected-badge nq-selected-badge--opp"><Target :size="14" /> {{ newQuoteForm.opportunityTitle }}</div>
              </div>

              <div v-if="customerPriceBooks.length" class="nq-list">
                <button v-for="pb in customerPriceBooks" :key="pb.id" class="nq-list-item" @click="selectPriceBook(pb.id, pb.name)">
                  <div class="nq-list-icon nq-list-icon--pb"><BookOpen :size="16" /></div>
                  <div class="nq-list-info">
                    <span class="nq-list-name">{{ pb.name }}</span>
                    <span class="nq-list-meta">
                      {{ pb.type.replace('-', ' ') }}{{ pb.customerName ? ` · ${pb.customerName}` : '' }} · {{ pb.entries.length }} items · Valid {{ pb.validFrom }} to {{ pb.validTo }}
                    </span>
                  </div>
                  <span v-if="pb.customerId" class="nq-pb-badge">Customer-specific</span>
                  <ChevronRight :size="16" class="nq-list-chevron" />
                </button>
              </div>
              <div v-else class="nq-empty">
                <BookOpen :size="24" class="text-muted" />
                <p>No price books available</p>
              </div>

              <button class="btn btn-ghost nq-skip-btn" @click="skipPriceBook">
                Skip — Use standard pricing
              </button>
            </div>

            <!-- Step 4: Confirm & Create -->
            <div v-else-if="newQuoteStep === 'confirm'">
              <div class="nq-section-title">Review & Create</div>

              <div class="nq-confirm-grid">
                <div class="nq-confirm-item" @click="goBackToStep('customer')">
                  <span class="nq-confirm-label">Customer</span>
                  <span class="nq-confirm-value">{{ newQuoteForm.customerName }}</span>
                </div>
                <div class="nq-confirm-item" @click="goBackToStep('opportunity')">
                  <span class="nq-confirm-label">Opportunity</span>
                  <span class="nq-confirm-value">{{ newQuoteForm.opportunityTitle || 'None (standalone quote)' }}</span>
                </div>
                <div class="nq-confirm-item" @click="goBackToStep('pricebook')">
                  <span class="nq-confirm-label">Price Book</span>
                  <span class="nq-confirm-value">{{ newQuoteForm.priceBookName || 'Standard pricing' }}</span>
                </div>
              </div>

              <div class="form-group mb-4">
                <label class="form-label">Valid Until</label>
                <input v-model="newQuoteForm.validUntil" type="date" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Notes</label>
                <textarea v-model="newQuoteForm.notes" class="form-input" rows="3" placeholder="Brief description of this quotation..." />
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showNewQuoteModal = false; resetNewQuoteForm()">Cancel</button>
            <button v-if="newQuoteStep === 'confirm'" class="btn btn-primary" @click="createAndOpenQuote">
              <Plus :size="16" /> Create & Open Builder
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.quotes-page {
  padding: var(--space-6);
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-5);
  margin-bottom: var(--space-6);
}

.kpi-card {
  padding: var(--space-5);
}

.kpi-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.kpi-label {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-500);
  margin-bottom: var(--space-1);
}

.kpi-value {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
  line-height: var(--leading-tight);
}

.kpi-subtitle {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  margin-top: var(--space-1);
}

.kpi-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.kpi-icon--primary {
  background-color: var(--color-primary-light);
  color: var(--color-primary);
}

.kpi-icon--warning {
  background-color: var(--color-warning-light);
  color: var(--color-warning);
}

.kpi-icon--success {
  background-color: var(--color-success-light);
  color: var(--color-success);
}

.kpi-icon--info {
  background-color: var(--color-primary-100);
  color: var(--color-primary-700);
}

/* Toolbar */
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  flex-wrap: wrap;
}

.toolbar-search {
  flex: 1;
  min-width: 220px;
}

.toolbar-select {
  width: 180px;
  flex-shrink: 0;
}

.toolbar-date {
  width: 150px;
  flex-shrink: 0;
}

/* View Modal */
.modal-header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.view-info-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  padding: var(--space-4);
  background-color: var(--color-neutral-50);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-5);
}

.view-info-label {
  display: block;
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.03em;
  margin-bottom: var(--space-1);
}

.view-info-value {
  font-size: var(--text-sm);
  color: var(--color-neutral-800);
}

.view-notes {
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-warning-light);
  border-left: 3px solid var(--color-warning);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-5);
  font-size: var(--text-sm);
  color: var(--color-neutral-700);
  line-height: var(--leading-relaxed);
}

.view-category-section {
  margin-bottom: var(--space-5);
}

.view-category-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-700);
  margin-bottom: var(--space-3);
  padding-bottom: var(--space-2);
  border-bottom: 1px solid var(--color-neutral-200);
}

.view-category-count {
  background-color: var(--color-neutral-100);
  color: var(--color-neutral-600);
  font-size: var(--text-xs);
  padding: 1px 8px;
  border-radius: var(--radius-full);
  font-weight: var(--font-medium);
}

.table-container--embedded {
  border: none;
  border-radius: 0;
  box-shadow: none;
}

/* Totals Card */
.view-totals-card {
  background-color: var(--color-neutral-50);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  padding: var(--space-4) var(--space-5);
  margin-bottom: var(--space-4);
}

.view-totals-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) 0;
  font-size: var(--text-sm);
  color: var(--color-neutral-700);
}

.view-totals-row--grand {
  border-top: 2px solid var(--color-neutral-300);
  padding-top: var(--space-3);
  margin-top: var(--space-1);
}

.view-grand-total {
  font-size: var(--text-lg);
  color: var(--color-primary);
}

.view-totals-divider {
  height: 1px;
  background: var(--color-neutral-200);
  margin: var(--space-2) 0;
}

.view-approval {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-success-light);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  color: var(--color-neutral-700);
}

@media (max-width: 1200px) {
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Lock Indicator */
.lock-indicator {
  color: var(--color-neutral-400);
  flex-shrink: 0;
}

/* View Modal Tabs */
.view-modal-tabs {
  display: flex;
  gap: var(--space-1);
  border-bottom: 2px solid var(--color-neutral-200);
  margin-bottom: var(--space-5);
}

.view-modal-tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-neutral-500);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  cursor: pointer;
  transition: all .15s;
}

.view-modal-tab:hover { color: var(--color-neutral-700); }
.view-modal-tab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); }

/* Audit Trail */
.audit-trail { padding: var(--space-2) 0; }

.audit-timeline {
  position: relative;
  padding-left: var(--space-6);
}

.audit-timeline::before {
  content: '';
  position: absolute;
  left: 7px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: var(--color-neutral-200);
}

.audit-entry {
  position: relative;
  padding-bottom: var(--space-5);
}

.audit-entry:last-child { padding-bottom: 0; }

.audit-dot {
  position: absolute;
  left: calc(-1 * var(--space-6) + 3px);
  top: 6px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--color-primary);
  border: 2px solid var(--content-surface);
  box-shadow: 0 0 0 2px var(--color-primary-light, #eff6ff);
}

.audit-content { min-width: 0; }

.audit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: 2px;
}

.audit-action {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-neutral-800);
}

.audit-time {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  white-space: nowrap;
}

.audit-detail {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  margin-bottom: 2px;
}

.audit-field {
  font-weight: 500;
  color: var(--color-neutral-600);
}

.audit-old {
  color: var(--color-danger);
  text-decoration: line-through;
}

.audit-arrow {
  color: var(--color-neutral-400);
}

.audit-new {
  color: var(--color-success);
  font-weight: 500;
}

.audit-user {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
}

/* New Quote Steps */
.nq-steps {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid var(--color-neutral-200);
  background: var(--color-neutral-50);
}

.nq-step {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  font-weight: 500;
  color: var(--color-neutral-400);
  cursor: default;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  transition: all .15s;
}

.nq-step--done { color: var(--color-success); cursor: pointer; }
.nq-step--done:hover { background: var(--color-success-light); }
.nq-step--active { color: var(--color-primary); font-weight: 600; }
.nq-step-arrow { color: var(--color-neutral-300); flex-shrink: 0; }

.nq-step-num {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6875rem;
  font-weight: 700;
  background: var(--color-neutral-200);
  color: var(--color-neutral-500);
}

.nq-step--active .nq-step-num { background: var(--color-primary); color: white; }
.nq-step--done .nq-step-num { background: var(--color-success); color: white; }

.nq-section-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-neutral-800);
  margin-bottom: var(--space-4);
}

.nq-section-subtitle { font-weight: 400; color: var(--color-neutral-400); }

.nq-selected-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background: var(--color-primary-light, #eff6ff);
  color: var(--color-primary);
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: 500;
  margin-bottom: var(--space-4);
}

.nq-selected-badge--opp {
  background: var(--color-warning-light);
  color: var(--color-warning-dark, #92400e);
}

.nq-selected-badges { display: flex; gap: var(--space-2); flex-wrap: wrap; }

.nq-list {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  overflow: hidden;
  max-height: 320px;
  overflow-y: auto;
}

.nq-list-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border: none;
  background: none;
  cursor: pointer;
  text-align: left;
  transition: background .1s;
  width: 100%;
}

.nq-list-item:hover { background: var(--color-neutral-50); }
.nq-list-item:not(:last-child) { border-bottom: 1px solid var(--color-neutral-100); }

.nq-list-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: var(--color-primary-light, #eff6ff);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nq-list-icon--opp { background: var(--color-warning-light); color: var(--color-warning); }
.nq-list-icon--pb { background: var(--color-success-light); color: var(--color-success); }

.nq-list-info { flex: 1; min-width: 0; }

.nq-list-name {
  display: block;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-neutral-800);
}

.nq-list-meta {
  display: block;
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  margin-top: 1px;
  text-transform: capitalize;
}

.nq-list-chevron { color: var(--color-neutral-300); flex-shrink: 0; }

.nq-pb-badge {
  font-size: 0.625rem;
  font-weight: 600;
  padding: 2px 8px;
  background: var(--color-success-light);
  color: var(--color-success);
  border-radius: var(--radius-full);
  white-space: nowrap;
}

.nq-skip-btn {
  width: 100%;
  justify-content: center;
  margin-top: var(--space-3);
  color: var(--color-neutral-500);
  font-size: var(--text-sm);
}

.nq-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-6);
  color: var(--color-neutral-400);
  font-size: var(--text-sm);
}

.nq-confirm-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: var(--space-3);
  margin-bottom: var(--space-5);
}

.nq-confirm-item {
  padding: var(--space-3) var(--space-4);
  background: var(--color-neutral-50);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all .15s;
}

.nq-confirm-item:hover { border-color: var(--color-primary); background: var(--color-primary-light, #eff6ff); }

.nq-confirm-label {
  display: block;
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-neutral-500);
  text-transform: uppercase;
  letter-spacing: .03em;
  margin-bottom: 4px;
}

.nq-confirm-value {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-neutral-800);
}

@media (max-width: 768px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
  .toolbar-select,
  .toolbar-date {
    width: 100%;
  }
  .view-info-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .nq-confirm-grid {
    grid-template-columns: 1fr;
  }
  .nq-steps { flex-wrap: wrap; }
}
</style>
