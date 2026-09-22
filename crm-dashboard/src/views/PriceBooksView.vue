<script setup lang="ts">
import { onMounted, ref, computed, watch, reactive } from 'vue'
import {
  Plus,
  Search,
  Eye,
  Pencil,
  X,
  BookOpen,
  Tag,
  Calendar,
  Save,
  Trash2,
  ArrowLeft,
  CheckCircle2,
  Package,
  Users,
  RefreshCw,
  GripVertical,
  Percent,
} from 'lucide-vue-next'
import type { PriceBook, PriceBookType, PriceBookEntry } from '@/types'
import { usePriceBooksStore } from '@/stores/priceBooks'

function uid(): string { return Math.random().toString(36).slice(2, 11) }

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

import { priceBooksService, customersService, http } from '@/services'
import type { ApiResponse } from '@/services/http'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'

const priceBookStore = usePriceBooksStore()
async function reloadBooks() { priceBooks.value = (await allPages(priceBooksService.list)).map(book => ({...book, validFrom:book.validFrom?.slice(0,10) ?? '', validTo:book.validTo?.slice(0,10) ?? '', entries:book.entries.map(entry => ({...entry,kind:entry.kind ?? 'product'}))})) }
onMounted(async () => { try {
 await reloadBooks(); customers.value = (await allPages(customersService.list)).map(c => ({id:c.id,name:c.companyName}))
 const catalog = (await http.get<ApiResponse<{products:{id:string;sku:string;name:string;manufacturer:string;unitPrice:number}[];services:{id:string;sku:string;name:string;unitPrice:number}[];recurring:{id:string;sku:string;name:string;monthlyPrice:number}[]}>>('/catalog')).data.data
 catalogItems.push(...catalog.products.map(p => ({...p,kind:'product' as const,standardPrice:p.unitPrice})),...catalog.services.map(p => ({...p,kind:'service' as const,standardPrice:p.unitPrice})),...catalog.recurring.map(p => ({...p,kind:'recurring' as const,standardPrice:p.monthlyPrice})))
 } catch (e) { window.alert(errorMessage(e)) } })

const typeLabels: Record<PriceBookType, string> = {
  standard: 'Standard',
  volume: 'Volume',
  contract: 'Contract',
  promotional: 'Promotional',
  'customer-specific': 'Customer Specific',
}

const typeBadge: Record<PriceBookType, string> = {
  standard: 'badge-primary',
  volume: 'badge-info',
  contract: 'badge-warning',
  promotional: 'badge-success',
  'customer-specific': 'badge-gray',
}

// ── Catalog: Products, Services, Recurring ──────────────────
type CatalogItemKind = 'product' | 'service' | 'recurring'

interface CatalogItem {
  id: string
  sku: string
  name: string
  kind: CatalogItemKind
  manufacturer?: string
  standardPrice: number
}

const catalogItems = reactive<CatalogItem[]>([])

const kindIcons = { product: Package, service: Users, recurring: RefreshCw } as const
const kindLabels: Record<CatalogItemKind, string> = { product: 'Product', service: 'Service', recurring: 'Recurring' }
const kindBadge: Record<CatalogItemKind, string> = { product: 'badge-primary', service: 'badge-warning', recurring: 'badge-success' }

// ── Customer directory ──────────────────────────────────────────
const customers = ref<{id:string;name:string}[]>([])

// ── Extended entry type (adds kind) ─────────────────────────
interface PBEntry extends PriceBookEntry {
  kind: CatalogItemKind
}

// ── Price Book Data ─────────────────────────────────────────
const priceBooks = ref<(Omit<PriceBook, 'entries'> & { entries: PBEntry[] })[]>([])
// ── View Mode: list vs builder ──────────────────────────────
const mode = ref<'list' | 'builder'>('list')
const activeBook = ref<(Omit<PriceBook, 'entries'> & { entries: PBEntry[] }) | null>(null)
const saveMessage = ref('')

// ── Builder: Search & Filter ────────────────────────────────
const searchQuery = ref('')
const catalogOpen = ref(true)
const kindFilter = ref<CatalogItemKind | 'all'>('all')
const entryKindFilter = ref<CatalogItemKind | 'all'>('all')

const searchResults = computed(() => {
  // An empty query browses the catalog; typing narrows it.
  if (!activeBook.value) return []
  const q = searchQuery.value.toLowerCase().trim()
  const existingIds = new Set(activeBook.value.entries.map(e => e.productId))
  return catalogItems.filter(item => {
    if (existingIds.has(item.id)) return false
    if (kindFilter.value !== 'all' && item.kind !== kindFilter.value) return false
    return !q || item.sku.toLowerCase().includes(q) || item.name.toLowerCase().includes(q) || (item.manufacturer?.toLowerCase().includes(q) ?? false)
  }).slice(0, 25)
})

const filteredEntries = computed(() => {
  if (!activeBook.value) return []
  if (entryKindFilter.value === 'all') return activeBook.value.entries
  return activeBook.value.entries.filter(e => e.kind === entryKindFilter.value)
})

const entryCounts = computed(() => {
  if (!activeBook.value) return { all: 0, product: 0, service: 0, recurring: 0 }
  const entries = activeBook.value.entries
  return {
    all: entries.length,
    product: entries.filter(e => e.kind === 'product').length,
    service: entries.filter(e => e.kind === 'service').length,
    recurring: entries.filter(e => e.kind === 'recurring').length,
  }
})

// ── Builder: Entry Actions ──────────────────────────────────
function addCatalogItem(item: CatalogItem) {
  if (!activeBook.value) return
  activeBook.value.entries.push({
    id: uid(),
    productId: item.id,
    productSku: item.sku,
    productName: item.name,
    standardPrice: item.standardPrice,
    customPrice: item.standardPrice,
    discountPercent: 0,
    kind: item.kind,
  })
  searchQuery.value = ''
}

function removeEntry(entryId: string) {
  if (!activeBook.value) return
  activeBook.value.entries = activeBook.value.entries.filter(e => e.id !== entryId)
}

function onCustomPriceChange(entry: PBEntry) {
  if (entry.standardPrice > 0) {
    entry.discountPercent = Math.round(((entry.standardPrice - entry.customPrice) / entry.standardPrice) * 10000) / 100
  }
}

function onDiscountChange(entry: PBEntry) {
  entry.customPrice = Math.round((entry.standardPrice * (1 - entry.discountPercent / 100)) * 100) / 100
}

function applyBulkDiscount() {
  if (!activeBook.value) return
  const pct = bulkDiscount.value
  if (pct <= 0 || pct > 100) return
  for (const entry of activeBook.value.entries) {
    entry.discountPercent = pct
    entry.customPrice = Math.round((entry.standardPrice * (1 - pct / 100)) * 100) / 100
  }
}

const bulkDiscount = ref(0)

// ── Summary Computeds ───────────────────────────────────────
const summaryStats = computed(() => {
  if (!activeBook.value) return { totalStandard: 0, totalCustom: 0, avgDiscount: 0, count: 0 }
  const entries = activeBook.value.entries
  const totalStandard = entries.reduce((s, e) => s + e.standardPrice, 0)
  const totalCustom = entries.reduce((s, e) => s + e.customPrice, 0)
  const avgDiscount = entries.length > 0 ? entries.reduce((s, e) => s + e.discountPercent, 0) / entries.length : 0
  return { totalStandard, totalCustom, avgDiscount, count: entries.length }
})

// ── Navigation ──────────────────────────────────────────────
function openBuilder(book: Omit<PriceBook, 'entries'> & { entries: PBEntry[] }) {
  activeBook.value = JSON.parse(JSON.stringify(book))
  entryKindFilter.value = 'all'
  searchQuery.value = ''
  kindFilter.value = 'all'
  bulkDiscount.value = 0
  mode.value = 'builder'
}

function createNewBook() {
  activeBook.value = {
    id: '', name: '', type: 'standard', description: '',
    validFrom: new Date().toISOString().slice(0, 10),
    validTo: new Date(Date.now() + 365 * 86400000).toISOString().slice(0, 10),
    isActive: true, entries: [],
    createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
  }
  entryKindFilter.value = 'all'
  searchQuery.value = ''
  kindFilter.value = 'all'
  bulkDiscount.value = 0
  mode.value = 'builder'
}

function goBack() {
  mode.value = 'list'
  activeBook.value = null
}

const saving = ref(false)
async function saveBook() {
 if (!activeBook.value?.name || saving.value) return
 saving.value = true
 try {
  const book = activeBook.value
  const result = book.id ? await priceBooksService.update(book.id, book) : await priceBooksService.create(book)
  book.id = result.data.id
  const saved = await priceBooksService.replaceEntries(book.id, book.entries)
  activeBook.value = { ...saved.data, validFrom:saved.data.validFrom?.slice(0,10) ?? '',validTo:saved.data.validTo?.slice(0,10) ?? '',entries:saved.data.entries.map(entry => ({...entry,kind:entry.kind ?? 'product'})) }
  await reloadBooks(); saveMessage.value = 'Price book saved'; window.setTimeout(() => {saveMessage.value=''},2500)
 } catch (e) { window.alert(errorMessage(e)) } finally { saving.value=false }
}

function onCustomerChange() {
  if (!activeBook.value) return
  const c = customers.value.find(c => c.id === activeBook.value!.customerId)
  if (c) activeBook.value.customerName = c.name
}

watch(searchQuery, () => {
  catalogOpen.value = true
})
</script>

<template>
  <div class="pb-page">
    <!-- ═══════ LIST MODE ═══════ -->
    <template v-if="mode === 'list'">
      <div class="page-header">
        <div>
          <h1 class="page-header-title">Price Books</h1>
          <p class="page-header-subtitle">{{ priceBooks.length }} price book{{ priceBooks.length !== 1 ? 's' : '' }}</p>
        </div>
        <button class="btn btn-primary" @click="createNewBook">
          <Plus :size="18" />
          Create Price Book
        </button>
      </div>

      <div class="pb-grid">
        <div v-for="book in priceBooks" :key="book.id" class="card pb-card">
          <div class="card-body">
            <div class="pb-card-top">
              <h3 class="pb-card-name">{{ book.name }}</h3>
              <span :class="['badge', typeBadge[book.type]]">{{ typeLabels[book.type] }}</span>
            </div>
            <p class="pb-card-desc">{{ book.description }}</p>

            <div class="pb-card-meta">
              <div v-if="book.customerName" class="pb-meta-item">
                <Tag :size="14" />
                <span>{{ book.customerName }}</span>
              </div>
              <div class="pb-meta-item">
                <Calendar :size="14" />
                <span>{{ book.validFrom }} &rarr; {{ book.validTo }}</span>
              </div>
              <div class="pb-meta-item">
                <BookOpen :size="14" />
                <span>{{ book.entries.length }} entr{{ book.entries.length !== 1 ? 'ies' : 'y' }}</span>
              </div>
            </div>

            <div class="pb-card-footer">
              <span :class="['badge badge-dot', book.isActive ? 'badge-success' : 'badge-danger']">
                {{ book.isActive ? 'Active' : 'Inactive' }}
              </span>
              <button class="btn btn-primary btn-sm" @click="openBuilder(book)">
                <Pencil :size="14" />
                Open Builder
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ═══════ BUILDER MODE ═══════ -->
    <template v-if="mode === 'builder' && activeBook">
      <!-- Save Toast -->
      <transition name="toast">
        <div v-if="saveMessage" class="save-toast">
          <CheckCircle2 :size="16" />
          {{ saveMessage }}
        </div>
      </transition>

      <!-- Builder Header -->
      <div class="builder-header">
        <div class="builder-header-left">
          <button class="btn btn-ghost btn-icon" @click="goBack">
            <ArrowLeft :size="20" />
          </button>
          <div>
            <input
              v-model="activeBook.name"
              type="text"
              class="builder-title-input"
              placeholder="Price Book Name..."
            />
            <div class="builder-subtitle">
              <span :class="['badge', typeBadge[activeBook.type]]">{{ typeLabels[activeBook.type] }}</span>
              <span v-if="activeBook.customerName" class="builder-customer">{{ activeBook.customerName }}</span>
              <span class="text-muted">{{ activeBook.entries.length }} items</span>
            </div>
          </div>
        </div>
        <div class="builder-header-actions">
          <button class="btn btn-primary" @click="saveBook">
            <Save :size="16" />
            Save Price Book
          </button>
        </div>
      </div>

      <!-- Info Bar -->
      <div class="info-bar">
        <div class="info-field">
          <label class="info-label">Type</label>
          <select v-model="activeBook.type" class="form-select info-select">
            <option v-for="(label, key) in typeLabels" :key="key" :value="key">{{ label }}</option>
          </select>
        </div>
        <div class="info-field">
          <label class="info-label">Customer</label>
          <select v-model="activeBook.customerId" class="form-select info-select" @change="onCustomerChange">
            <option value="">None (General)</option>
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>
        <div class="info-field">
          <label class="info-label">Valid From</label>
          <input v-model="activeBook.validFrom" type="date" class="form-input info-date" />
        </div>
        <div class="info-field">
          <label class="info-label">Valid To</label>
          <input v-model="activeBook.validTo" type="date" class="form-input info-date" />
        </div>
        <div class="info-field">
          <label class="info-label">Status</label>
          <label class="toggle-label">
            <input v-model="activeBook.isActive" type="checkbox" class="toggle-input" />
            <span class="toggle-text">{{ activeBook.isActive ? 'Active' : 'Inactive' }}</span>
          </label>
        </div>
      </div>

      <!-- Builder Content: Main + Summary -->
      <div class="builder-content">
        <!-- Main Panel -->
        <div class="builder-main">
          <!-- Add Item Card -->
          <div class="card add-item-card">
            <div class="card-body">
              <h3 class="add-item-title">
                <Plus :size="16" />
                Add Items
              </h3>
              <div class="add-item-row">
                <!-- Kind filter chips -->
                <div class="kind-chips">
                  <button
                    :class="['kind-chip', kindFilter === 'all' && 'kind-chip--active']"
                    @click="kindFilter = 'all'"
                  >All</button>
                  <button
                    :class="['kind-chip', kindFilter === 'product' && 'kind-chip--active']"
                    @click="kindFilter = 'product'"
                  >
                    <Package :size="13" />
                    Products
                  </button>
                  <button
                    :class="['kind-chip', kindFilter === 'service' && 'kind-chip--active']"
                    @click="kindFilter = 'service'"
                  >
                    <Users :size="13" />
                    Services
                  </button>
                  <button
                    :class="['kind-chip', kindFilter === 'recurring' && 'kind-chip--active']"
                    @click="kindFilter = 'recurring'"
                  >
                    <RefreshCw :size="13" />
                    Recurring
                  </button>
                </div>

                <!-- Search -->
                <div class="add-search-wrapper">
                  <div class="search-input">
                    <Search :size="16" class="search-icon" />
                    <input
                      v-model="searchQuery"
                      type="text"
                      class="form-input"
                      placeholder="Search by SKU, name, or manufacturer..."
                    />
                  </div>

                  <!-- Search Dropdown -->
                  <p class="catalog-count">{{ catalogItems.length }} catalog items available</p>
                  <div v-if="catalogOpen && searchResults.length" class="product-dropdown product-dropdown--inline">
                    <button
                      v-for="item in searchResults"
                      :key="item.id"
                      type="button"
                      class="product-dropdown-item"
                      @click="addCatalogItem(item)"
                    >
                      <div class="dropdown-item-main">
                        <span :class="['kind-badge', kindBadge[item.kind]]">{{ kindLabels[item.kind] }}</span>
                        <span class="dropdown-sku">{{ item.sku }}</span>
                        <span class="dropdown-name">{{ item.name }}</span>
                        <span v-if="item.manufacturer" class="dropdown-mfr">{{ item.manufacturer }}</span>
                      </div>
                      <span class="dropdown-price">SAR {{ formatSAR(item.standardPrice) }}</span>
                    </button>
                  </div>

                  <div v-if="catalogOpen && !searchResults.length" class="product-dropdown product-dropdown--inline product-dropdown--empty">
                    <div class="dropdown-empty">
                      <Search :size="16" />
                      <span v-if="searchQuery.trim()">No items found for "{{ searchQuery }}"</span>
                      <span v-else>No catalog items left to add. Run the catalog seed, or every item is already in this book.</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Entry Kind Tabs -->
          <div class="tabs">
            <button
              :class="['tab', entryKindFilter === 'all' && 'tab--active']"
              @click="entryKindFilter = 'all'"
            >
              All
              <span class="tab-count">{{ entryCounts.all }}</span>
            </button>
            <button
              :class="['tab', entryKindFilter === 'product' && 'tab--active']"
              @click="entryKindFilter = 'product'"
            >
              <Package :size="14" />
              Products
              <span class="tab-count">{{ entryCounts.product }}</span>
            </button>
            <button
              :class="['tab', entryKindFilter === 'service' && 'tab--active']"
              @click="entryKindFilter = 'service'"
            >
              <Users :size="14" />
              Services
              <span class="tab-count">{{ entryCounts.service }}</span>
            </button>
            <button
              :class="['tab', entryKindFilter === 'recurring' && 'tab--active']"
              @click="entryKindFilter = 'recurring'"
            >
              <RefreshCw :size="14" />
              Recurring
              <span class="tab-count">{{ entryCounts.recurring }}</span>
            </button>
          </div>

          <!-- Excel-Like Entries Table -->
          <div v-if="filteredEntries.length" class="card">
            <div class="table-container table-container--flush">
              <table class="table builder-table">
                <thead>
                  <tr>
                    <th class="col-drag"></th>
                    <th class="col-num">#</th>
                    <th class="col-kind">Type</th>
                    <th class="col-sku">SKU</th>
                    <th>Item Name</th>
                    <th class="text-right">Standard Price</th>
                    <th class="text-right col-price">Custom Price</th>
                    <th class="text-right col-disc">Discount %</th>
                    <th class="text-right">Savings</th>
                    <th class="col-actions"></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, idx) in filteredEntries" :key="entry.id" class="entry-row">
                    <td class="col-drag">
                      <GripVertical :size="14" class="drag-handle" />
                    </td>
                    <td class="col-num text-muted">{{ idx + 1 }}</td>
                    <td class="col-kind">
                      <component :is="kindIcons[entry.kind]" :size="14" :class="'kind-icon kind-icon--' + entry.kind" :title="kindLabels[entry.kind]" />
                    </td>
                    <td class="col-sku">
                      <span class="text-mono text-xs">{{ entry.productSku }}</span>
                    </td>
                    <td>
                      <span class="entry-name">{{ entry.productName }}</span>
                    </td>
                    <td class="text-right whitespace-nowrap text-muted">
                      SAR {{ formatSAR(entry.standardPrice) }}
                    </td>
                    <td class="text-right col-price">
                      <input
                        v-model.number="entry.customPrice"
                        type="number"
                        step="0.01"
                        min="0"
                        class="inline-input inline-input--price"
                        @change="onCustomPriceChange(entry)"
                      />
                    </td>
                    <td class="text-right col-disc">
                      <div class="disc-input-wrap">
                        <input
                          v-model.number="entry.discountPercent"
                          type="number"
                          step="0.5"
                          min="0"
                          max="100"
                          class="inline-input inline-input--disc"
                          @change="onDiscountChange(entry)"
                        />
                        <span class="disc-symbol">%</span>
                      </div>
                    </td>
                    <td class="text-right whitespace-nowrap">
                      <span :class="entry.discountPercent > 0 ? 'text-success font-medium' : 'text-muted'">
                        {{ entry.discountPercent > 0 ? '- SAR ' + formatSAR(entry.standardPrice - entry.customPrice) : '—' }}
                      </span>
                    </td>
                    <td class="col-actions">
                      <button class="btn btn-ghost btn-icon btn-sm" title="Remove" @click="removeEntry(entry.id)">
                        <Trash2 :size="14" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-else class="empty-state">
            <BookOpen :size="40" class="empty-state-icon" />
            <h3 class="empty-state-title">No items yet</h3>
            <p class="empty-state-text">Search for products, services, or recurring items above to build your price book.</p>
          </div>
        </div>

        <!-- Summary Panel -->
        <div class="builder-summary">
          <div class="card summary-card">
            <div class="card-body">
              <h3 class="summary-title">Price Book Summary</h3>

              <div class="summary-rows">
                <div class="summary-row">
                  <span>Total Items</span>
                  <span class="font-semibold">{{ summaryStats.count }}</span>
                </div>
                <div class="summary-row">
                  <span>Total Standard</span>
                  <span class="text-mono">SAR {{ formatSAR(summaryStats.totalStandard) }}</span>
                </div>
                <div class="summary-row">
                  <span>Total Custom</span>
                  <span class="text-mono font-medium">SAR {{ formatSAR(summaryStats.totalCustom) }}</span>
                </div>
                <div class="summary-row summary-row--highlight">
                  <span class="font-medium">Total Savings</span>
                  <span class="text-mono text-success font-semibold">
                    SAR {{ formatSAR(summaryStats.totalStandard - summaryStats.totalCustom) }}
                  </span>
                </div>
                <div class="summary-row">
                  <span>Avg Discount</span>
                  <span :class="['font-semibold', summaryStats.avgDiscount > 0 ? 'text-success' : 'text-muted']">
                    {{ summaryStats.avgDiscount.toFixed(1) }}%
                  </span>
                </div>
              </div>

              <!-- Bulk Discount -->
              <div class="bulk-section">
                <h4 class="bulk-title">Apply Bulk Discount</h4>
                <div class="bulk-row">
                  <div class="bulk-input-wrap">
                    <input
                      v-model.number="bulkDiscount"
                      type="number"
                      step="0.5"
                      min="0"
                      max="100"
                      class="form-input bulk-input"
                      placeholder="0"
                    />
                    <span class="bulk-symbol">%</span>
                  </div>
                  <button class="btn btn-secondary btn-sm" :disabled="bulkDiscount <= 0" @click="applyBulkDiscount">
                    Apply to All
                  </button>
                </div>
              </div>

              <!-- Actions -->
              <div class="summary-actions">
                <button class="btn btn-primary btn-lg summary-action-btn" @click="saveBook">
                  <Save :size="16" />
                  Save Price Book
                </button>
                <button class="btn btn-secondary summary-action-btn" @click="goBack">
                  <ArrowLeft :size="16" />
                  Back to List
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.pb-page {
  padding: var(--space-6);
  position: relative;
}

/* ── List Mode ────────────────────────────────────────────── */
.pb-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: var(--space-4);
}

.pb-card .card-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.pb-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}

.pb-card-name {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
  margin: 0;
}

.pb-card-desc {
  font-size: var(--text-sm);
  color: var(--color-neutral-500);
  margin: 0;
}

.pb-card-meta {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.pb-meta-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
}

.pb-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-neutral-200);
}

/* ── Builder Mode ─────────────────────────────────────────── */
.save-toast {
  position: fixed;
  top: var(--space-4);
  right: var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  background-color: var(--color-success);
  color: white;
  border-radius: var(--radius-lg);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-12px);
}

/* Builder Header */
.builder-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-5);
}

.builder-header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.builder-title-input {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
  border: none;
  background: transparent;
  outline: none;
  padding: 0;
  width: 400px;
  font-family: inherit;
}

.builder-title-input::placeholder {
  color: var(--color-neutral-300);
}

.builder-title-input:focus {
  border-bottom: 2px solid var(--color-primary);
}

.builder-subtitle {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-top: var(--space-1);
  font-size: var(--text-sm);
}

.builder-customer {
  font-weight: var(--font-medium);
  color: var(--color-neutral-700);
}

.builder-header-actions {
  display: flex;
  gap: var(--space-2);
}

/* Info Bar */
.info-bar {
  display: flex;
  align-items: flex-end;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-5);
  flex-wrap: wrap;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  color: var(--color-neutral-500);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.info-select {
  width: 180px;
  padding: 6px 10px;
  font-size: var(--text-sm);
}

.info-date {
  width: 150px;
  padding: 6px 10px;
  font-size: var(--text-sm);
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
}

.toggle-input {
  width: 16px;
  height: 16px;
  accent-color: var(--color-primary);
}

.toggle-text {
  color: var(--color-neutral-700);
}

/* Builder Content Layout */
.builder-content {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: var(--space-5);
  align-items: start;
}

.builder-main {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

/* Add Item Card */
.add-item-card {
  border: 2px dashed var(--color-neutral-200);
  background: var(--color-neutral-50);
}

.add-item-card .card-body {
  padding: var(--space-4) var(--space-5);
}

.add-item-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-700);
  margin-bottom: var(--space-3);
}

.add-item-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.kind-chips {
  display: flex;
  gap: var(--space-2);
}

.kind-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-full);
  background: var(--content-surface);
  color: var(--color-neutral-600);
  cursor: pointer;
  transition: all 0.15s;
}

.kind-chip:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.kind-chip--active {
  background: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.add-search-wrapper {
  position: relative;
}

/* Product Dropdown */
.catalog-count { font-size: var(--text-xs); color: var(--color-neutral-500); margin-top: var(--space-2); }
.product-dropdown--inline { position: static; margin-top: var(--space-2); box-shadow: none; max-height: 300px; }
.product-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: var(--space-1);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  max-height: 360px;
  overflow-y: auto;
  z-index: 50;
}

.product-dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-3) var(--space-4);
  border: none;
  background: none;
  cursor: pointer;
  text-align: left;
  transition: background-color 0.1s;
}

.product-dropdown-item:hover {
  background-color: var(--color-primary-light, #eff6ff);
}

.product-dropdown-item:not(:last-child) {
  border-bottom: 1px solid var(--color-neutral-100);
}

.dropdown-item-main {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex: 1;
  min-width: 0;
}

.kind-badge {
  font-size: 0.625rem;
  padding: 1px 6px;
  border-radius: var(--radius-full);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  white-space: nowrap;
  flex-shrink: 0;
}

.dropdown-sku {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  color: var(--color-primary);
  white-space: nowrap;
  background: var(--color-primary-light, #eff6ff);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.dropdown-name {
  font-size: var(--text-sm);
  color: var(--color-neutral-800);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.dropdown-mfr {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  white-space: nowrap;
}

.dropdown-price {
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--color-neutral-600);
  white-space: nowrap;
  flex-shrink: 0;
}

.product-dropdown--empty {
  padding: 0;
}

.dropdown-empty {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4);
  font-size: var(--text-sm);
  color: var(--color-neutral-400);
}

/* Tabs */
.tabs {
  display: flex;
  gap: var(--space-1);
  border-bottom: 2px solid var(--color-neutral-200);
}

.tab {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-2) var(--space-3);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-500);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}

.tab:hover {
  color: var(--color-neutral-700);
}

.tab--active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-count {
  font-size: var(--text-xs);
  background: var(--color-neutral-100);
  color: var(--color-neutral-600);
  padding: 0 6px;
  border-radius: var(--radius-full);
  font-weight: var(--font-semibold);
}

.tab--active .tab-count {
  background: var(--color-primary-light, #eff6ff);
  color: var(--color-primary);
}

/* Builder Table */
.table-container--flush {
  border: none;
  border-radius: 0;
}

.builder-table th {
  padding: 8px 12px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-neutral-500);
  background: var(--color-neutral-50);
  border-bottom: 1px solid var(--color-neutral-200);
  white-space: nowrap;
}

.builder-table td {
  padding: 6px 12px;
  vertical-align: middle;
  border-bottom: 1px solid var(--color-neutral-100);
  font-size: var(--text-sm);
}

.col-drag { width: 30px; text-align: center; }
.col-num { width: 36px; text-align: center; }
.col-kind { width: 36px; text-align: center; }
.col-sku { width: 130px; }
.col-price { width: 130px; }
.col-disc { width: 110px; }
.col-actions { width: 40px; }

.drag-handle {
  color: var(--color-neutral-300);
  cursor: grab;
}

.drag-handle:hover {
  color: var(--color-neutral-500);
}

.entry-row:hover {
  background-color: var(--color-neutral-50);
}

.entry-row:hover .drag-handle {
  color: var(--color-neutral-500);
}

.entry-name {
  font-weight: var(--font-medium);
  color: var(--color-neutral-800);
}

.kind-icon { flex-shrink: 0; }
.kind-icon--product { color: var(--color-primary); }
.kind-icon--service { color: var(--color-warning); }
.kind-icon--recurring { color: var(--color-success); }

/* Inline Inputs */
.inline-input {
  width: 100%;
  padding: 4px 8px;
  font-size: var(--text-sm);
  font-family: var(--font-mono);
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  background: transparent;
  text-align: right;
  color: var(--color-neutral-800);
  transition: all 0.15s;
  outline: none;
}

.inline-input:hover {
  border-color: var(--color-neutral-200);
  background: var(--content-surface);
}

.inline-input:focus {
  border-color: var(--color-primary);
  background: var(--content-surface);
  box-shadow: 0 0 0 2px var(--color-primary-light, rgba(37, 99, 235, 0.15));
}

.inline-input--price { width: 120px; }
.inline-input--disc { width: 70px; padding-right: 20px; }

.disc-input-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.disc-symbol {
  position: absolute;
  right: 6px;
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  pointer-events: none;
}

/* Summary Panel */
.builder-summary {
  position: sticky;
  top: calc(var(--header-height, 64px) + var(--space-6));
}

.summary-card .card-body {
  padding: var(--space-5);
}

.summary-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-700);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: var(--space-4);
}

.summary-rows {
  display: flex;
  flex-direction: column;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) 0;
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
}

.summary-row--highlight {
  padding: var(--space-3) 0;
  margin-top: var(--space-1);
  border-top: 2px solid var(--color-neutral-200);
}

/* Bulk Discount */
.bulk-section {
  margin-top: var(--space-5);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-neutral-200);
}

.bulk-title {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-neutral-500);
  margin-bottom: var(--space-3);
}

.bulk-row {
  display: flex;
  gap: var(--space-2);
  align-items: center;
}

.bulk-input-wrap {
  position: relative;
  flex: 1;
}

.bulk-input {
  padding-right: 28px;
  text-align: right;
  font-family: var(--font-mono);
}

.bulk-symbol {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  font-size: var(--text-sm);
  color: var(--color-neutral-400);
  pointer-events: none;
}

/* Summary Actions */
.summary-actions {
  margin-top: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.summary-action-btn {
  width: 100%;
  justify-content: center;
}

/* ── Responsive ───────────────────────────────────────────── */
@media (max-width: 1200px) {
  .builder-content {
    grid-template-columns: 1fr;
  }
  .builder-summary {
    position: static;
  }
}

@media (max-width: 768px) {
  .pb-grid { grid-template-columns: 1fr; }
  .info-bar { flex-direction: column; }
  .info-select, .info-date { width: 100%; }
  .builder-title-input { width: 100%; }
  .kind-chips { flex-wrap: wrap; }
}
</style>
