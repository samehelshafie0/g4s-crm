<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus, Search, Eye, Trash2, X, FileText, Clock, CheckCircle2,
  TrendingUp, TrendingDown, DollarSign, Package, Truck, AlertTriangle,
  ShoppingCart, ExternalLink, ArrowUpRight, Download, History,
  Minus, ChevronDown, Building2, Star, BarChart3, Clipboard,
  PackageCheck, RefreshCw, ArrowRight, AlertCircle, MapPin,
  Upload, Pencil, Check, Copy, FileUp, Table2, FileSpreadsheet,
} from 'lucide-vue-next'
import { parseFile, parseCSVText, downloadTemplate, getMappedValue, getMappedNumber, getMappedInt, type FileType, type ParseResult } from '@/utils/fileParser'
import RecordAttachments from '@/components/RecordAttachments.vue'
import { useProcurementStore } from '@/stores/procurement'
import { useManufacturersStore } from '@/stores/manufacturers'
import { useQuotesStore } from '@/stores/quotes'
import type {
  PurchaseOrder, PurchaseOrderStatus,
  SupplierQuote, SupplierQuoteStatus, SupplierQuoteLineItem,
  SupplierItemEntry, GoodsReceipt, Manufacturer,
  Quote, Project, QuoteLineItem,
} from '@/types'

import { procurementService, productsService, manufacturersService, quotesService, inventoryService, projectsService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import type { Product } from '@/types'
const catalogProducts = ref<Product[]>([])
const availableStock = ref<Record<string,number>>({})

const store = useProcurementStore()
const mfrStore = useManufacturersStore()
const quotesStore = useQuotesStore()
const projectList = ref<{ id: string; projectNumber: string; name: string }[]>([])

async function reloadProcurement() {await Promise.all([store.fetchPurchaseOrders({limit:100}),store.fetchSupplierQuotes({limit:100}),store.fetchGoodsReceipts({limit:100})]);store.supplierItems=(await procurementService.supplierItems()).data}
onMounted(async()=>{try{await reloadProcurement();catalogProducts.value=await allPages(productsService.list);mfrStore.manufacturers=await allPages(manufacturersService.list);quotesStore.quotes=await allPages(quotesService.list);projectList.value=await allPages(projectsService.list);const stock=await allPages(inventoryService.listStock);for(const row of stock) availableStock.value[row.productId]=(availableStock.value[row.productId]??0)+row.availableQty}catch(e){window.alert(errorMessage(e))}})

function uid(): string { return Math.random().toString(36).slice(2, 11) }

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

const statusConfig: Record<PurchaseOrderStatus, { label: string; badge: string }> = {
  draft: { label: 'Draft', badge: 'badge-gray' },
  'pending-approval': { label: 'Pending Approval', badge: 'badge-warning' },
  approved: { label: 'Approved', badge: 'badge-info' },
  ordered: { label: 'Ordered', badge: 'badge-primary' },
  'partial-received': { label: 'Partial Received', badge: 'badge-warning' },
  received: { label: 'Received', badge: 'badge-success' },
  cancelled: { label: 'Cancelled', badge: 'badge-danger' },
}

const sqStatusConfig: Record<SupplierQuoteStatus, { label: string; badge: string }> = {
  received: { label: 'Received', badge: 'badge-info' },
  'under-review': { label: 'Under Review', badge: 'badge-warning' },
  accepted: { label: 'Accepted', badge: 'badge-success' },
  expired: { label: 'Expired', badge: 'badge-gray' },
  rejected: { label: 'Rejected', badge: 'badge-danger' },
}

// ── Main Tabs ────────────────────────────────────────────────
type ProcTab = 'purchase-orders' | 'supplier-quotes' | 'goods-receipts' | 'supplier-catalog'
const activeTab = ref<ProcTab>('purchase-orders')

// ────────────────────────────────────────────────────────────
// PURCHASE ORDERS TAB
// ────────────────────────────────────────────────────────────
const poSearch = ref('')
const poStatusFilter = ref<PurchaseOrderStatus | ''>('')

const filteredPOs = computed(() => {
  let list = store.purchaseOrders
  const q = poSearch.value.toLowerCase().trim()
  if (q) {
    list = list.filter(po =>
      po.poNumber.toLowerCase().includes(q) ||
      po.supplierName.toLowerCase().includes(q) ||
      (po.sourceQuoteNumber?.toLowerCase().includes(q) ?? false),
    )
  }
  if (poStatusFilter.value) {
    list = list.filter(po => po.status === poStatusFilter.value)
  }
  return list
})

const showViewPOModal = ref(false)
const viewingPO = ref<PurchaseOrder | null>(null)

function openViewPO(po: PurchaseOrder) {
  viewingPO.value = po
  showViewPOModal.value = true
}

async function deletePO(id: string) {try{await store.deletePurchaseOrder(id)}catch(e){window.alert(errorMessage(e))}}

function totalReceived(po: PurchaseOrder): number {
  return po.items.reduce((s, i) => s + i.receivedQty, 0)
}

function totalOrdered(po: PurchaseOrder): number {
  return po.items.reduce((s, i) => s + i.quantity, 0)
}

function receiptProgress(po: PurchaseOrder): number {
  const ordered = totalOrdered(po)
  if (ordered === 0) return 0
  return (totalReceived(po) / ordered) * 100
}

// ── Enhanced PO Creation (4 steps: source → supplier → items → review) ──
const showCreatePOModal = ref(false)
const createPOStep = ref<'source' | 'supplier' | 'items' | 'review'>('source')
const selectedSupplierId = ref('')
const newPOSupplier = ref('')
const newPOExpected = ref('')
const newPONotes = ref('')
const newPOSourceSQ = ref('')
const supplierSearchQuery = ref('')

// Step 1: Source — project or quote
type POSourceType = 'project' | 'quote' | 'manual'
const poSourceType = ref<POSourceType>('project')
const selectedSourceQuote = ref<Quote | null>(null)
const sourceSearchQuery = ref('')

const availableProjects = computed(() => {
  const q = sourceSearchQuery.value.toLowerCase().trim()
  const accepted = quotesStore.quotes.filter(qt => qt.status === 'accepted')
  if (!q) return accepted
  return accepted.filter(qt => qt.quoteNumber.toLowerCase().includes(q) || qt.customerName.toLowerCase().includes(q) || qt.notes.toLowerCase().includes(q))
})

const availableQuotes = computed(() => {
  const q = sourceSearchQuery.value.toLowerCase().trim()
  const all = quotesStore.quotes.filter(qt => qt.status !== 'declined' && qt.status !== 'expired')
  if (!q) return all
  return all.filter(qt => qt.quoteNumber.toLowerCase().includes(q) || qt.customerName.toLowerCase().includes(q) || qt.notes.toLowerCase().includes(q))
})

function selectSource(quote: Quote) {
  selectedSourceQuote.value = quote
  createPOStep.value = 'supplier'
}

function skipSource() {
  selectedSourceQuote.value = null
  poSourceType.value = 'manual'
  createPOStep.value = 'supplier'
}

// Source items extracted from the selected quote/project
const sourceItems = computed(() => {
  if (!selectedSourceQuote.value) return []
  return selectedSourceQuote.value.lineItems.filter(li => li.category === 'materials')
})

// Step 2: Supplier (with catalog/quote requirement)
const selectedSupplier = computed(() => {
  if (!selectedSupplierId.value) return null
  return mfrStore.manufacturers.find(m => m.id === selectedSupplierId.value) ?? null
})

function supplierHasCatalog(supplierName: string): boolean {
  const hasItems = store.supplierItems.some(si => si.supplierName === supplierName)
  const hasQuotes = store.supplierQuotes.some(sq => sq.supplierName === supplierName && sq.status !== 'expired' && sq.status !== 'rejected')
  return hasItems || hasQuotes
}

interface SupplierCoverage {
  matchedSkus: string[]
  matchedNames: string[]
  totalSourceItems: number
  matchCount: number
  coveragePercent: number
}

function getSupplierCoverage(supplierName: string): SupplierCoverage {
  const items = sourceItems.value
  if (items.length === 0) return { matchedSkus: [], matchedNames: [], totalSourceItems: 0, matchCount: 0, coveragePercent: 0 }

  const name = supplierName.toLowerCase()
  const matchedSkus: string[] = []
  const matchedNames: string[] = []

  for (const li of items) {
    const sku = li.sku || ''
    const hasCatalog = sku && store.supplierItems.some(si => si.productSku === sku && si.supplierName.toLowerCase() === name)
    const hasSQ = sku && store.supplierQuotes.some(sq =>
      sq.supplierName.toLowerCase() === name && sq.status !== 'expired' && sq.status !== 'rejected' && sq.items.some(i => i.productSku === sku),
    )
    if (hasCatalog || hasSQ) {
      matchedSkus.push(sku)
      matchedNames.push(li.description.length > 30 ? li.description.slice(0, 28) + '...' : li.description)
    }
  }

  return {
    matchedSkus,
    matchedNames,
    totalSourceItems: items.length,
    matchCount: matchedSkus.length,
    coveragePercent: items.length > 0 ? Math.round((matchedSkus.length / items.length) * 100) : 0,
  }
}

const filteredSupplierList = computed(() => {
  const q = supplierSearchQuery.value.toLowerCase().trim()
  let list = [...mfrStore.suppliers]
  if (q) list = list.filter(s => s.name.toLowerCase().includes(q) || s.code.toLowerCase().includes(q))

  if (sourceItems.value.length > 0) {
    list.sort((a, b) => {
      const ca = getSupplierCoverage(a.name)
      const cb = getSupplierCoverage(b.name)
      if (cb.matchCount !== ca.matchCount) return cb.matchCount - ca.matchCount
      const aCat = supplierHasCatalog(a.name) ? 1 : 0
      const bCat = supplierHasCatalog(b.name) ? 1 : 0
      return bCat - aCat
    })
  }

  return list
})

const supplierAvailableQuotes = computed(() => {
  if (!selectedSupplier.value) return []
  const name = selectedSupplier.value.name
  return store.supplierQuotes.filter(sq =>
    sq.supplierName === name && sq.status !== 'expired' && sq.status !== 'rejected',
  )
})

const supplierCatalogItems = computed(() => {
  if (!selectedSupplier.value) return []
  return store.supplierItems.filter(si => si.supplierName === selectedSupplier.value!.name)
})

function selectSupplierForPO(mfr: Manufacturer) {
  selectedSupplierId.value = mfr.id
  newPOSupplier.value = mfr.name
  autoMatchSourceItems()
  createPOStep.value = 'items'
}

function enterNewSupplier() {
  selectedSupplierId.value = ''
  if (supplierSearchQuery.value.trim()) {
    newPOSupplier.value = supplierSearchQuery.value.trim()
  }
  createPOStep.value = 'items'
}

// Step 3: Items — auto-match then manual
interface NewPOItem {
  sku: string; name: string; manufacturer: string; qty: number; unitCost: number;
  leadTimeDays: number; onHand: number; bestSupplierPrice: number | null;
  bestSupplierName: string; landingCost: number | null; lastReceiveDate: string;
  fromSQ: boolean; sqRef: string; matched: boolean;
}
const newPOItems = ref<NewPOItem[]>([])

const newPOItemSearch = ref('')
const showPOItemDropdown = ref(false)

function autoMatchSourceItems() {
  newPOItems.value = []
  if (!selectedSourceQuote.value) return
  const materialItems = selectedSourceQuote.value.lineItems.filter(li => li.category === 'materials' && li.sku)
  const supplierName = newPOSupplier.value.toLowerCase()

  for (const li of materialItems) {
    const catalogMatch = store.supplierItems.find(si => si.productSku === li.sku && si.supplierName.toLowerCase() === supplierName)
    const sqMatch = findSQItemMatch(li.sku || '', newPOSupplier.value)

    if (catalogMatch) {
      const receipts = store.getReceiptHistoryForProduct(catalogMatch.productId)
      const lastReceipt = receipts.length > 0 ? receipts[0] : null
      newPOItems.value.push({
        sku: li.sku || '', name: li.description, manufacturer: li.manufacturerName || '',
        qty: li.quantity, unitCost: catalogMatch.latestCost, leadTimeDays: catalogMatch.leadTimeDays,
        onHand: li.stockAvailable || 0,
        bestSupplierPrice: catalogMatch.latestCost, bestSupplierName: catalogMatch.supplierName,
        landingCost: lastReceipt ? lastReceipt.item.landingCost : null,
        lastReceiveDate: lastReceipt ? lastReceipt.gr.receiveDate : '',
        fromSQ: false, sqRef: '', matched: true,
      })
    } else if (sqMatch) {
      newPOItems.value.push({
        sku: li.sku || '', name: li.description, manufacturer: li.manufacturerName || '',
        qty: li.quantity, unitCost: sqMatch.item.unitCost, leadTimeDays: sqMatch.item.leadTimeDays,
        onHand: li.stockAvailable || 0,
        bestSupplierPrice: sqMatch.item.unitCost, bestSupplierName: sqMatch.sq.supplierName,
        landingCost: null, lastReceiveDate: '',
        fromSQ: true, sqRef: sqMatch.sq.sqNumber, matched: true,
      })
    }
  }
}

function findSQItemMatch(sku: string, supplierName: string): { sq: SupplierQuote; item: SupplierQuoteLineItem } | null {
  for (const sq of store.supplierQuotes) {
    if (sq.supplierName !== supplierName || sq.status === 'expired' || sq.status === 'rejected') continue
    const item = sq.items.find(i => i.productSku === sku)
    if (item) return { sq, item }
  }
  return null
}

const poItemSearchResults = computed(() => {
  const q = newPOItemSearch.value.toLowerCase().trim()
  if (!q) return []
  const existingSkus = new Set(newPOItems.value.map(i => i.sku))
  return store.supplierItems
    .filter(si => {
      if (newPOSupplier.value && si.supplierName.toLowerCase() !== newPOSupplier.value.toLowerCase()) return false
      if (existingSkus.has(si.productSku)) return false
      return si.productSku.toLowerCase().includes(q) || si.productName.toLowerCase().includes(q) || si.manufacturerName.toLowerCase().includes(q)
    })
    .slice(0, 10)
})

function selectPOItem(si: SupplierItemEntry) {
  const receipts = store.getReceiptHistoryForProduct(si.productId)
  const lastReceipt = receipts.length > 0 ? receipts[0] : null
  const best = store.getBestSupplierPrice(si.productSku)
  const onHandMap = availableStock.value
  newPOItems.value.push({
    sku: si.productSku, name: si.productName, manufacturer: si.manufacturerName,
    qty: si.moq, unitCost: si.latestCost, leadTimeDays: si.leadTimeDays,
    onHand: onHandMap[si.productId] ?? 0,
    bestSupplierPrice: best ? best.cost : null, bestSupplierName: best ? best.supplierName : '',
    landingCost: lastReceipt ? lastReceipt.item.landingCost : null,
    lastReceiveDate: lastReceipt ? lastReceipt.gr.receiveDate : '',
    fromSQ: false, sqRef: '', matched: false,
  })
  newPOItemSearch.value = ''
  showPOItemDropdown.value = false
}

function importFromSupplierQuote(sq: SupplierQuote) {
  const mfr = mfrStore.manufacturers.find(m => m.name === sq.supplierName)
  if (mfr) selectedSupplierId.value = mfr.id
  newPOSupplier.value = sq.supplierName
  newPOSourceSQ.value = sq.sqNumber
  for (const item of sq.items) {
    const receipts = store.getReceiptHistoryForProduct(item.productId ?? '')
    const lastReceipt = receipts.length > 0 ? receipts[0] : null
    const best = store.getBestSupplierPrice(item.productSku)
    const onHandMap = availableStock.value
    newPOItems.value.push({
      sku: item.productSku, name: item.productName, manufacturer: item.manufacturerName,
      qty: item.quantity, unitCost: item.unitCost, leadTimeDays: item.leadTimeDays,
      onHand: onHandMap[item.productId ?? ''] ?? 0,
      bestSupplierPrice: best ? best.cost : null, bestSupplierName: best ? best.supplierName : '',
      landingCost: lastReceipt ? lastReceipt.item.landingCost : null,
      lastReceiveDate: lastReceipt ? lastReceipt.gr.receiveDate : '',
      fromSQ: true, sqRef: sq.sqNumber, matched: false,
    })
  }
  showCreatePOModal.value = true
  createPOStep.value = 'items'
}

function addBlankPOItem() {
  newPOItems.value.push({
    sku: '', name: '', manufacturer: '', qty: 1, unitCost: 0, leadTimeDays: 30,
    onHand: 0, bestSupplierPrice: null, bestSupplierName: '', landingCost: null,
    lastReceiveDate: '', fromSQ: false, sqRef: '', matched: false,
  })
}

function removePOItem(idx: number) { newPOItems.value.splice(idx, 1) }

// ── PO File Upload (CSV / Excel / PDF) ──────────────────────
const poFileRef = ref<HTMLInputElement | null>(null)
const poFileName = ref('')
const poFileType = ref<FileType | ''>('')
const poFileError = ref('')
const poPasteText = ref('')
const poFileParsing = ref(false)

interface POFileRow {
  sku: string; name: string; qty: number; unitCost: number; manufacturer: string; leadTimeDays: number; selected: boolean
}
const poFileRows = ref<POFileRow[]>([])
const showPOFileUpload = ref(false)

function triggerPOFileUpload() { poFileRef.value?.click() }

function openPOFileUpload() {
  poFileName.value = ''
  poFileType.value = ''
  poFileError.value = ''
  poFileRows.value = []
  poPasteText.value = ''
  poFileParsing.value = false
  showPOFileUpload.value = true
}

function closePOFileUpload() {
  showPOFileUpload.value = false
  poFileRows.value = []
  poFileName.value = ''
}

function applyPOParseResult(result: ParseResult) {
  const { headers, rows, mappedColumns } = result
  if (mappedColumns.name === undefined && mappedColumns.sku === undefined) {
    poFileError.value = 'Could not detect SKU or Product Name columns. Check your file headers.'
    return
  }
  poFileRows.value = []
  for (const row of rows) {
    const sku = getMappedValue(row, 'sku', headers, mappedColumns)
    const name = getMappedValue(row, 'name', headers, mappedColumns) || sku
    if (!sku && !name) continue
    poFileRows.value.push({
      sku, name,
      qty: getMappedInt(row, 'qty', headers, mappedColumns, 1),
      unitCost: getMappedNumber(row, 'unitCost', headers, mappedColumns, 0),
      manufacturer: getMappedValue(row, 'manufacturer', headers, mappedColumns),
      leadTimeDays: getMappedInt(row, 'leadTimeDays', headers, mappedColumns, 30),
      selected: true,
    })
  }
  if (poFileRows.value.length === 0) { poFileError.value = 'No data rows could be parsed.'; return }
  poFileError.value = ''
}

async function handlePOFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await readPOFile(file)
  input.value = ''
}

async function readPOFile(file?: File) {
  if (!file) return
  poFileName.value = file.name
  poFileError.value = ''
  poFileRows.value = []
  poPasteText.value = ''
  poFileParsing.value = true

  const out = await parseFile(file)
  poFileParsing.value = false
  if ('error' in out) { poFileError.value = out.error; return }
  poFileType.value = out.fileType
  applyPOParseResult(out.result)
}

function parsePOPastedText() {
  if (!poPasteText.value.trim()) return
  const result = parseCSVText(poPasteText.value)
  if (result.rows.length === 0) { poFileError.value = 'No data could be parsed from pasted text.'; return }
  applyPOParseResult(result)
}

const poFileSelectedCount = computed(() => poFileRows.value.filter(r => r.selected).length)

function importPOFileItems() {
  const selected = poFileRows.value.filter(r => r.selected)
  for (const row of selected) {
    const catalogMatch = newPOSupplier.value ? store.supplierItems.find(si => si.supplierName.toLowerCase() === newPOSupplier.value.toLowerCase() && (si.productSku.toLowerCase() === row.sku.toLowerCase() || si.productName.toLowerCase() === row.name.toLowerCase())) : null
    const onHandMap = availableStock.value
    if (catalogMatch) {
      const receipts = store.getReceiptHistoryForProduct(catalogMatch.productId)
      const lastReceipt = receipts.length > 0 ? receipts[0] : null
      newPOItems.value.push({
        sku: catalogMatch.productSku, name: catalogMatch.productName, manufacturer: catalogMatch.manufacturerName,
        qty: row.qty, unitCost: row.unitCost || catalogMatch.latestCost, leadTimeDays: row.leadTimeDays || catalogMatch.leadTimeDays,
        onHand: onHandMap[catalogMatch.productId] ?? 0,
        bestSupplierPrice: catalogMatch.latestCost, bestSupplierName: catalogMatch.supplierName,
        landingCost: lastReceipt ? lastReceipt.item.landingCost : null,
        lastReceiveDate: lastReceipt ? lastReceipt.gr.receiveDate : '',
        fromSQ: false, sqRef: '', matched: true,
      })
    } else {
      newPOItems.value.push({
        sku: row.sku, name: row.name, manufacturer: row.manufacturer,
        qty: row.qty, unitCost: row.unitCost, leadTimeDays: row.leadTimeDays,
        onHand: 0, bestSupplierPrice: null, bestSupplierName: '',
        landingCost: null, lastReceiveDate: '',
        fromSQ: false, sqRef: '', matched: false,
      })
    }
  }
  closePOFileUpload()
}

const newPOSubtotal = computed(() => newPOItems.value.reduce((s, i) => s + i.qty * i.unitCost, 0))
const matchedCount = computed(() => newPOItems.value.filter(i => i.matched).length)
const unmatchedSourceItems = computed(() => {
  if (!selectedSourceQuote.value) return []
  const matchedSkus = new Set(newPOItems.value.filter(i => i.matched).map(i => i.sku))
  return selectedSourceQuote.value.lineItems.filter(li => li.category === 'materials' && li.sku && !matchedSkus.has(li.sku))
})

function resetCreatePO() {
  selectedSupplierId.value = ''
  supplierSearchQuery.value = ''
  newPOSupplier.value = ''
  newPOExpected.value = ''
  newPONotes.value = ''
  newPOSourceSQ.value = ''
  newPOItems.value = []
  createPOStep.value = 'source'
  poSourceType.value = 'project'
  selectedSourceQuote.value = null
  sourceSearchQuery.value = ''
  showPOFileUpload.value = false
  poFileRows.value = []
  poFileName.value = ''
  poFileType.value = ''
  poFileError.value = ''
  poPasteText.value = ''
}

async function createPO() {
  try {
  if (!newPOSupplier.value || newPOItems.value.length === 0) return
  const items = newPOItems.value.map(i => ({
    id: uid(), productId: resolveProduct(i.sku), productSku: i.sku, productName: i.name,
    manufacturerName: i.manufacturer, quantity: i.qty, unitCost: i.unitCost,
    total: i.qty * i.unitCost, receivedQty: 0, leadTimeDays: i.leadTimeDays,
  }))
  const subtotal = items.reduce((s, i) => s + i.total, 0)
  const shippingCost = 0
  const customsDuty = 0
  const po: PurchaseOrder = {
    id: uid(), poNumber: '', supplierName: newPOSupplier.value,
    status: 'draft', items, subtotal, shippingCost, customsDuty,
    total: subtotal + shippingCost + customsDuty, currency: 'SAR',
    expectedDelivery: newPOExpected.value || new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10),
    sourceQuoteId: selectedSourceQuote.value?.id,
    sourceQuoteNumber: selectedSourceQuote.value?.quoteNumber,
    notes: newPONotes.value,
    createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
  }
  await store.addPurchaseOrder(po)
  showCreatePOModal.value = false
  resetCreatePO()
  } catch(e){window.alert(errorMessage(e))}
}

// ────────────────────────────────────────────────────────────
// SUPPLIER QUOTES TAB
// ────────────────────────────────────────────────────────────
const sqSearch = ref('')
const sqStatusFilter = ref<SupplierQuoteStatus | ''>('')

const filteredSQs = computed(() => {
  let list = store.supplierQuotes
  const q = sqSearch.value.toLowerCase().trim()
  if (q) {
    list = list.filter(sq =>
      sq.sqNumber.toLowerCase().includes(q) ||
      sq.supplierName.toLowerCase().includes(q) ||
      (sq.supplierRef?.toLowerCase().includes(q) ?? false),
    )
  }
  if (sqStatusFilter.value) {
    list = list.filter(sq => sq.status === sqStatusFilter.value)
  }
  return list
})

const showViewSQModal = ref(false)
const viewingSQ = ref<SupplierQuote | null>(null)

function openViewSQ(sq: SupplierQuote) {
  viewingSQ.value = sq
  showViewSQModal.value = true
}

function daysUntilExpiry(sq: SupplierQuote): number {
  return Math.ceil((new Date(sq.validUntil).getTime() - Date.now()) / 86400000)
}

// ────────────────────────────────────────────────────────────
// GOODS RECEIPTS TAB
// ────────────────────────────────────────────────────────────
const grSearch = ref('')

const filteredGRs = computed(() => {
  const q = grSearch.value.toLowerCase().trim()
  if (!q) return store.goodsReceipts
  return store.goodsReceipts.filter(gr =>
    gr.grNumber.toLowerCase().includes(q) ||
    gr.poNumber.toLowerCase().includes(q) ||
    gr.supplierName.toLowerCase().includes(q),
  )
})

const showViewGRModal = ref(false)
const viewingGR = ref<GoodsReceipt | null>(null)

function openViewGR(gr: GoodsReceipt) {
  viewingGR.value = gr
  showViewGRModal.value = true
}

// ────────────────────────────────────────────────────────────
// SUPPLIER CATALOG TAB
// ────────────────────────────────────────────────────────────
const catSearch = ref('')
const catSupplierFilter = ref('')

const supplierNames = computed(() => [...new Set(store.supplierItems.map(si => si.supplierName))])

const filteredCatalog = computed(() => {
  let list = store.supplierItems
  const q = catSearch.value.toLowerCase().trim()
  if (q) {
    list = list.filter(si =>
      si.productSku.toLowerCase().includes(q) ||
      si.productName.toLowerCase().includes(q) ||
      si.supplierName.toLowerCase().includes(q) ||
      si.manufacturerName.toLowerCase().includes(q),
    )
  }
  if (catSupplierFilter.value) {
    list = list.filter(si => si.supplierName === catSupplierFilter.value)
  }
  return list
})

const showItemDetailModal = ref(false)
const detailItem = ref<SupplierItemEntry | null>(null)

function openItemDetail(item: SupplierItemEntry) {
  detailItem.value = item
  showItemDetailModal.value = true
}

const detailReceiptHistory = computed(() => {
  if (!detailItem.value) return []
  return store.getReceiptHistoryForProduct(detailItem.value.productId)
})

const detailSupplierQuotes = computed(() => {
  if (!detailItem.value) return []
  return store.getSupplierQuotesForProduct(detailItem.value.productSku)
})

const detailAllSuppliers = computed(() => {
  if (!detailItem.value) return []
  return store.getSupplierItemsForProduct(detailItem.value.productId)
})

function delayHidePODropdown() { window.setTimeout(() => { showPOItemDropdown.value = false }, 200) }

// ────────────────────────────────────────────────────────────
// CREATE / EDIT SUPPLIER QUOTE
// ────────────────────────────────────────────────────────────
const showCreateSQModal = ref(false)
const editingSQId = ref<string | null>(null)

interface SQFormItem {
  sku: string; name: string; manufacturer: string
  qty: number; unitCost: number; leadTimeDays: number; moq: number
}

const sqForm = ref({
  supplierName: '',
  supplierRef: '',
  sourceQuoteId: '',
  projectId: '',
  contactName: '',
  contactEmail: '',
  validFrom: '',
  validUntil: '',
  paymentTerms: '',
  deliveryTerms: '',
  notes: '',
  items: [] as SQFormItem[],
})

function resetSQForm() {
  editingSQId.value = null
  sqForm.value = {
    supplierName: '', supplierRef: '', sourceQuoteId: '', projectId: '', contactName: '', contactEmail: '',
    validFrom: new Date().toISOString().slice(0, 10),
    validUntil: new Date(Date.now() + 90 * 86400000).toISOString().slice(0, 10),
    paymentTerms: '', deliveryTerms: '', notes: '', items: [],
  }
  csvPreviewItems.value = []
  csvFileName.value = ''
  csvFileType.value = ''
  sqPasteText.value = ''
}

function openCreateSQ() {
  resetSQForm()
  showCreateSQModal.value = true
}

function openEditSQ(sq: SupplierQuote) {
  editingSQId.value = sq.id
  sqForm.value = {
    supplierName: sq.supplierName,
    supplierRef: sq.supplierRef || '',
    sourceQuoteId: sq.sourceQuoteId || '',
    projectId: sq.projectId || '',
    contactName: sq.contactName || '',
    contactEmail: sq.contactEmail || '',
    validFrom: sq.validFrom?.slice(0, 10) ?? '',
    validUntil: sq.validUntil?.slice(0, 10) ?? '',
    paymentTerms: sq.paymentTerms || '',
    deliveryTerms: sq.deliveryTerms || '',
    notes: sq.notes,
    items: sq.items.map(i => ({
      sku: i.productSku, name: i.productName, manufacturer: i.manufacturerName,
      qty: i.quantity, unitCost: i.unitCost, leadTimeDays: i.leadTimeDays, moq: i.moq ?? 1,
    })),
  }
  csvPreviewItems.value = []
  csvFileName.value = ''
  csvFileType.value = ''
  sqPasteText.value = ''
  showCreateSQModal.value = true
}

function addBlankSQItem() {
  sqForm.value.items.push({ sku: '', name: '', manufacturer: '', qty: 1, unitCost: 0, leadTimeDays: 30, moq: 1 })
}

function removeSQItem(idx: number) { sqForm.value.items.splice(idx, 1) }

const sqFormSubtotal = computed(() => sqForm.value.items.reduce((s, i) => s + i.qty * i.unitCost, 0))

function selectSQSupplier(mfr: Manufacturer) {
  sqForm.value.supplierName = mfr.name
  sqForm.value.contactEmail = mfr.contactEmail
}

function resolveProduct(sku: string): string {const product=catalogProducts.value.find(p=>p.sku.toLowerCase()===sku.trim().toLowerCase());if(!product)throw new Error(`Add SKU ${sku || '(empty)'} to Products before creating a purchase order.`);return product.id}
async function saveSQ() {
 if(!sqForm.value.supplierName||sqForm.value.items.length===0)return
 try {const items=sqForm.value.items.map(i=>({productId:catalogProducts.value.find(p=>p.sku.toLowerCase()===i.sku.toLowerCase())?.id,productSku:i.sku,productName:i.name,manufacturerName:i.manufacturer,quantity:i.qty,unitCost:i.unitCost,leadTimeDays:i.leadTimeDays,moq:i.moq}));const data={...sqForm.value,items,sourceQuoteId:sqForm.value.sourceQuoteId||null,projectId:sqForm.value.projectId||null,currency:store.supplierQuotes.find(s => s.id === editingSQId.value)?.currency ?? 'SAR'}
 if(editingSQId.value)await store.updateSupplierQuote(editingSQId.value,data);else await store.addSupplierQuote(data)
 showCreateSQModal.value=false;resetSQForm()
 }catch(e){window.alert(errorMessage(e))}
}
async function deleteSQ(id:string){try{await store.deleteSupplierQuote(id)}catch(e){window.alert(errorMessage(e))}}
async function poAction(po:PurchaseOrder,action:'pending-approval'|'approve'|'ordered'|'cancelled'){
 try{if(action==='approve')await procurementService.approvePO(po.id);else await procurementService.updatePOStatus(po.id,action);await reloadProcurement();viewingPO.value=store.purchaseOrders.find(p=>p.id===po.id)??null}catch(e){window.alert(errorMessage(e))}
}
async function acceptSQ(sq:SupplierQuote){try{await procurementService.updateSQStatus(sq.id,'accepted');await reloadProcurement();viewingSQ.value=store.supplierQuotes.find(s=>s.id===sq.id)??null}catch(e){window.alert(errorMessage(e))}}
async function convertSQ(sq:SupplierQuote){try{await procurementService.convertToPO(sq.id);await reloadProcurement();showViewSQModal.value=false;activeTab.value='purchase-orders'}catch(e){window.alert(errorMessage(e))}}
async function receivePO(po:PurchaseOrder){
 const warehouse=window.prompt('Receiving warehouse: riyadh-main, jeddah-branch or dammam-branch','riyadh-main');if(!warehouse)return
 const items=[]
 for(const line of po.items.filter(i=>i.receivedQty<i.quantity)){const answer=window.prompt(`Receive quantity for ${line.productSku} (${line.quantity-line.receivedQty} outstanding; 0 to skip)`,String(line.quantity-line.receivedQty));if(answer===null)return;const qty=Number(answer);if(!Number.isInteger(qty)||qty<0){window.alert('Enter a non-negative whole quantity.');return};if(qty>0)items.push({poItemId:line.id,productId:line.productId,receivedQty:qty,storageLocation:warehouse,condition:'good'})}
 if(items.length===0)return
 try{await store.addGoodsReceipt({poId:po.id,receiveDate:new Date().toISOString().slice(0,10),items});await reloadProcurement();showViewPOModal.value=false;activeTab.value='goods-receipts'}catch(e){window.alert(errorMessage(e))}
}

// ── CSV / Excel / PDF Catalog Upload ─────────────────────────
const csvFileName = ref('')
const csvFileType = ref<FileType | ''>('')
const csvPreviewItems = ref<SQFormItem[]>([])
const csvParseError = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)
const sqPasteText = ref('')
const sqFileParsing = ref(false)

function triggerFileUpload() {
  fileInputRef.value?.click()
}

function applySQParseResult(result: ParseResult) {
  const { headers, rows, mappedColumns } = result
  if (mappedColumns.name === undefined && mappedColumns.sku === undefined) {
    csvParseError.value = 'Could not find "SKU" or "Name/Description" column in the file.'
    return
  }
  const parsed: SQFormItem[] = []
  for (const row of rows) {
    const sku = getMappedValue(row, 'sku', headers, mappedColumns)
    const name = getMappedValue(row, 'name', headers, mappedColumns) || sku
    if (!name && !sku) continue
    parsed.push({
      sku: sku || '', name: name || sku,
      manufacturer: getMappedValue(row, 'manufacturer', headers, mappedColumns),
      qty: getMappedInt(row, 'qty', headers, mappedColumns, 1),
      unitCost: getMappedNumber(row, 'unitCost', headers, mappedColumns, 0),
      leadTimeDays: getMappedInt(row, 'leadTimeDays', headers, mappedColumns, 30),
      moq: getMappedInt(row, 'moq', headers, mappedColumns, 1),
    })
  }
  if (parsed.length === 0) { csvParseError.value = 'No valid items found in file.'; return }
  csvPreviewItems.value = parsed
  csvParseError.value = ''
}

async function handleFileUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await readSQFile(file)
  input.value = ''
}

async function readSQFile(file?: File) {
  if (!file) return
  csvFileName.value = file.name
  csvParseError.value = ''
  csvPreviewItems.value = []
  sqPasteText.value = ''
  sqFileParsing.value = true

  const out = await parseFile(file)
  sqFileParsing.value = false
  if ('error' in out) { csvParseError.value = out.error; return }
  csvFileType.value = out.fileType
  applySQParseResult(out.result)
}

function parseSQPastedText() {
  if (!sqPasteText.value.trim()) return
  const result = parseCSVText(sqPasteText.value)
  if (result.rows.length === 0) { csvParseError.value = 'No data could be parsed from pasted text.'; return }
  applySQParseResult(result)
}

function importCSVItems() {
  sqForm.value.items.push(...csvPreviewItems.value)
  csvPreviewItems.value = []
  csvFileName.value = ''
}

function clearCSVPreview() {
  csvPreviewItems.value = []
  csvFileName.value = ''
  csvParseError.value = ''
}

// ── SQ Item search from catalog ──────────────────────────────
const sqItemSearch = ref('')
const showSQItemDropdown = ref(false)

const sqItemSearchResults = computed(() => {
  const q = sqItemSearch.value.toLowerCase().trim()
  if (!q) return []
  const supplierName = sqForm.value.supplierName.toLowerCase()
  return store.supplierItems
    .filter(si => {
      if (supplierName && si.supplierName.toLowerCase() !== supplierName) return false
      return si.productSku.toLowerCase().includes(q) || si.productName.toLowerCase().includes(q)
    })
    .slice(0, 8)
})

function selectSQCatalogItem(si: SupplierItemEntry) {
  sqForm.value.items.push({
    sku: si.productSku, name: si.productName, manufacturer: si.manufacturerName,
    qty: si.moq, unitCost: si.latestCost, leadTimeDays: si.leadTimeDays, moq: si.moq,
  })
  sqItemSearch.value = ''
  showSQItemDropdown.value = false
}

function delayHideSQDropdown() { window.setTimeout(() => { showSQItemDropdown.value = false }, 200) }
</script>

<template>
  <div class="procurement-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Procurement</h1>
        <p class="page-header-subtitle">Purchase orders, supplier quotes, receiving & vendor catalog</p>
      </div>
      <div class="page-header-actions">
        <button class="btn btn-secondary" @click="activeTab = 'supplier-quotes'">
          <FileText :size="16" /> Supplier Quotes
        </button>
        <button class="btn btn-primary" @click="resetCreatePO(); showCreatePOModal = true">
          <Plus :size="18" /> New Purchase Order
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="kpi-grid kpi-grid--6">
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Purchase Orders</p>
            <p class="kpi-value">{{ store.totalPOs }}</p>
          </div>
          <div class="kpi-icon kpi-icon--primary"><ShoppingCart :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Pending Approval</p>
            <p class="kpi-value">{{ store.pendingPOs }}</p>
          </div>
          <div class="kpi-icon kpi-icon--warning"><Clock :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">In Transit</p>
            <p class="kpi-value">{{ store.inTransitPOs }}</p>
          </div>
          <div class="kpi-icon kpi-icon--info"><Truck :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">PO Total Value</p>
            <p class="kpi-value kpi-value--sm">SAR {{ formatSAR(store.totalValue) }}</p>
          </div>
          <div class="kpi-icon kpi-icon--success"><DollarSign :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Active Supplier Quotes</p>
            <p class="kpi-value">{{ store.activeSQCount }}</p>
          </div>
          <div class="kpi-icon kpi-icon--purple"><FileText :size="22" /></div>
        </div>
      </div>
      <div class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">Goods Received</p>
            <p class="kpi-value">{{ store.totalGRs }}</p>
          </div>
          <div class="kpi-icon kpi-icon--teal"><PackageCheck :size="22" /></div>
        </div>
      </div>
    </div>

    <!-- Main Tabs -->
    <div class="proc-tabs">
      <button :class="['proc-tab', activeTab === 'purchase-orders' && 'proc-tab--active']" @click="activeTab = 'purchase-orders'">
        <ShoppingCart :size="16" /> Purchase Orders <span class="proc-tab-count">{{ store.totalPOs }}</span>
      </button>
      <button :class="['proc-tab', activeTab === 'supplier-quotes' && 'proc-tab--active']" @click="activeTab = 'supplier-quotes'">
        <FileText :size="16" /> Supplier Quotes <span class="proc-tab-count">{{ store.supplierQuotes.length }}</span>
      </button>
      <button :class="['proc-tab', activeTab === 'goods-receipts' && 'proc-tab--active']" @click="activeTab = 'goods-receipts'">
        <PackageCheck :size="16" /> Goods Receipts <span class="proc-tab-count">{{ store.goodsReceipts.length }}</span>
      </button>
      <button :class="['proc-tab', activeTab === 'supplier-catalog' && 'proc-tab--active']" @click="activeTab = 'supplier-catalog'">
        <Building2 :size="16" /> Supplier Catalog <span class="proc-tab-count">{{ store.supplierItems.length }}</span>
      </button>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- TAB: Purchase Orders                                   -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="activeTab === 'purchase-orders'">
      <div class="card mb-5">
        <div class="toolbar">
          <div class="search-input toolbar-search">
            <Search :size="18" class="search-icon" />
            <input v-model="poSearch" type="text" class="form-input" placeholder="Search by PO #, supplier, or quote #..." />
          </div>
          <select v-model="poStatusFilter" class="form-select toolbar-select">
            <option value="">All Statuses</option>
            <option v-for="(cfg, key) in statusConfig" :key="key" :value="key">{{ cfg.label }}</option>
          </select>
        </div>
      </div>

      <div v-if="filteredPOs.length" class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>PO #</th>
              <th>Supplier</th>
              <th>Status</th>
              <th class="text-center">Items</th>
              <th class="text-right">Total</th>
              <th>Expected</th>
              <th>Source Quote</th>
              <th class="text-center">Receipt</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="po in filteredPOs" :key="po.id">
              <td class="text-mono font-bold">{{ po.poNumber }}</td>
              <td class="font-medium text-dark">{{ po.supplierName }}</td>
              <td><span :class="['badge badge-dot', statusConfig[po.status].badge]">{{ statusConfig[po.status].label }}</span></td>
              <td class="text-center">{{ po.items.length }}</td>
              <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(po.total) }}</td>
              <td class="text-muted whitespace-nowrap">{{ formatDate(po.expectedDelivery) }}</td>
              <td>
                <span v-if="po.sourceQuoteNumber" class="source-quote-link"><FileText :size="12" /> {{ po.sourceQuoteNumber }}</span>
                <span v-else class="text-muted">—</span>
              </td>
              <td class="text-center">
                <div class="receipt-cell">
                  <div class="receipt-bar-track">
                    <div class="receipt-bar-fill" :class="{ 'receipt-complete': receiptProgress(po) >= 100 }" :style="{ width: `${Math.min(receiptProgress(po), 100)}%` }" />
                  </div>
                  <span class="receipt-label">{{ totalReceived(po) }}/{{ totalOrdered(po) }}</span>
                </div>
              </td>
              <td>
                <div class="table-actions">
                  <button class="btn btn-ghost btn-icon btn-sm" title="View" @click="openViewPO(po)"><Eye :size="14" /></button>
                  <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deletePO(po.id)" :disabled="po.status === 'received'"><Trash2 :size="14" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">
        <ShoppingCart :size="48" class="empty-state-icon" />
        <h3 class="empty-state-title">No purchase orders found</h3>
        <p class="empty-state-text">{{ poSearch || poStatusFilter ? 'Try adjusting your filters.' : 'Create a PO or generate one from a quote.' }}</p>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- TAB: Supplier Quotes                                   -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="activeTab === 'supplier-quotes'">
      <div class="card mb-5">
        <div class="toolbar">
          <div class="search-input toolbar-search">
            <Search :size="18" class="search-icon" />
            <input v-model="sqSearch" type="text" class="form-input" placeholder="Search by quote #, supplier, or ref..." />
          </div>
          <select v-model="sqStatusFilter" class="form-select toolbar-select">
            <option value="">All Statuses</option>
            <option v-for="(cfg, key) in sqStatusConfig" :key="key" :value="key">{{ cfg.label }}</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="openCreateSQ">
            <Plus :size="15" /> Add Supplier Quote
          </button>
        </div>
      </div>

      <div v-if="filteredSQs.length" class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>Quote #</th>
              <th>Supplier</th>
              <th>Supplier Ref</th>
              <th>Status</th>
              <th class="text-center">Items</th>
              <th class="text-right">Value</th>
              <th>Valid Until</th>
              <th>Terms</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sq in filteredSQs" :key="sq.id">
              <td class="text-mono font-bold">{{ sq.sqNumber }}</td>
              <td class="font-medium text-dark">{{ sq.supplierName }}</td>
              <td class="text-mono text-muted" style="font-size:0.75rem">{{ sq.supplierRef || '—' }}</td>
              <td><span :class="['badge badge-dot', sqStatusConfig[sq.status].badge]">{{ sqStatusConfig[sq.status].label }}</span></td>
              <td class="text-center">{{ sq.items.length }}</td>
              <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(sq.subtotal) }}</td>
              <td>
                <span :class="['whitespace-nowrap', daysUntilExpiry(sq) <= 14 && daysUntilExpiry(sq) > 0 ? 'text-warning font-semibold' : daysUntilExpiry(sq) <= 0 ? 'text-danger font-semibold' : 'text-muted']">
                  {{ formatDate(sq.validUntil) }}
                  <span v-if="daysUntilExpiry(sq) > 0 && daysUntilExpiry(sq) <= 30" class="sq-days-left">({{ daysUntilExpiry(sq) }}d left)</span>
                </span>
              </td>
              <td class="text-muted" style="font-size:0.75rem">{{ sq.paymentTerms || '—' }}</td>
              <td>
                <div class="table-actions">
                  <button class="btn btn-ghost btn-icon btn-sm" title="View Details" @click="openViewSQ(sq)"><Eye :size="14" /></button>
                  <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click="openEditSQ(sq)"><Pencil :size="14" /></button>
                  <button v-if="sq.status === 'accepted'" class="btn btn-ghost btn-icon btn-sm" title="Create PO from Quote" @click="convertSQ(sq)"><ArrowRight :size="14" /></button>
                  <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deleteSQ(sq.id)"><Trash2 :size="14" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">
        <FileText :size="48" class="empty-state-icon" />
        <h3 class="empty-state-title">No supplier quotes found</h3>
        <p class="empty-state-text">{{ sqSearch || sqStatusFilter ? 'Try adjusting your filters.' : 'Add supplier quotes to compare vendor pricing.' }}</p>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- TAB: Goods Receipts                                    -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="activeTab === 'goods-receipts'">
      <div class="card mb-5">
        <div class="toolbar">
          <div class="search-input toolbar-search">
            <Search :size="18" class="search-icon" />
            <input v-model="grSearch" type="text" class="form-input" placeholder="Search by GR #, PO #, or supplier..." />
          </div>
        </div>
      </div>

      <div v-if="filteredGRs.length" class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>GR #</th>
              <th>PO #</th>
              <th>Supplier</th>
              <th>Receive Date</th>
              <th class="text-center">Items</th>
              <th class="text-right">Landing Cost</th>
              <th>Received By</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="gr in filteredGRs" :key="gr.id">
              <td class="text-mono font-bold">{{ gr.grNumber }}</td>
              <td class="text-mono">
                <span class="source-quote-link"><ShoppingCart :size="11" /> {{ gr.poNumber }}</span>
              </td>
              <td class="font-medium text-dark">{{ gr.supplierName }}</td>
              <td class="whitespace-nowrap">{{ formatDate(gr.receiveDate) }}</td>
              <td class="text-center">{{ gr.totalItems }}</td>
              <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(gr.totalLandingCost) }}</td>
              <td class="text-muted">{{ gr.receivedBy }}</td>
              <td>
                <button class="btn btn-ghost btn-icon btn-sm" title="View" @click="openViewGR(gr)"><Eye :size="14" /></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">
        <PackageCheck :size="48" class="empty-state-icon" />
        <h3 class="empty-state-title">No goods receipts found</h3>
        <p class="empty-state-text">Goods receipts track individual receiving events with landing costs.</p>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- TAB: Supplier Catalog                                  -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="activeTab === 'supplier-catalog'">
      <div class="card mb-5">
        <div class="toolbar">
          <div class="search-input toolbar-search">
            <Search :size="18" class="search-icon" />
            <input v-model="catSearch" type="text" class="form-input" placeholder="Search by SKU, product, supplier, or manufacturer..." />
          </div>
          <select v-model="catSupplierFilter" class="form-select toolbar-select">
            <option value="">All Suppliers</option>
            <option v-for="name in supplierNames" :key="name" :value="name">{{ name }}</option>
          </select>
        </div>
      </div>

      <div v-if="filteredCatalog.length" class="table-container">
        <table class="table">
          <thead>
            <tr>
              <th>SKU</th>
              <th>Product</th>
              <th>Supplier</th>
              <th>Manufacturer</th>
              <th class="text-right">Latest Cost</th>
              <th class="text-center">Trend</th>
              <th class="text-center">MOQ</th>
              <th class="text-center">Lead Time</th>
              <th class="text-center">Reliability</th>
              <th>Last Quote</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="si in filteredCatalog" :key="si.id">
              <td class="text-mono font-medium">{{ si.productSku }}</td>
              <td class="font-medium text-dark">{{ si.productName }}</td>
              <td>{{ si.supplierName }}</td>
              <td class="text-muted">{{ si.manufacturerName }}</td>
              <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(si.latestCost) }}</td>
              <td class="text-center">
                <span v-if="si.costTrend === 'down'" class="trend-badge trend-badge--down"><TrendingDown :size="12" /></span>
                <span v-else-if="si.costTrend === 'up'" class="trend-badge trend-badge--up"><TrendingUp :size="12" /></span>
                <span v-else class="trend-badge trend-badge--stable"><Minus :size="12" /></span>
              </td>
              <td class="text-center">{{ si.moq }}</td>
              <td class="text-center">{{ si.leadTimeDays }}d</td>
              <td class="text-center">
                <div class="reliability-bar">
                  <div class="reliability-fill" :class="si.reliability >= 90 ? 'reliability--good' : si.reliability >= 80 ? 'reliability--ok' : 'reliability--poor'" :style="{ width: `${si.reliability}%` }" />
                </div>
                <span class="reliability-label">{{ si.reliability }}%</span>
              </td>
              <td class="text-muted whitespace-nowrap">{{ formatDate(si.lastQuoteDate) }}</td>
              <td>
                <button class="btn btn-ghost btn-icon btn-sm" title="View Item Detail" @click="openItemDetail(si)"><Eye :size="14" /></button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="empty-state">
        <Building2 :size="48" class="empty-state-icon" />
        <h3 class="empty-state-title">No items found</h3>
        <p class="empty-state-text">The supplier catalog maps items to suppliers with pricing and lead times.</p>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- MODALS                                                 -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <!-- View PO Modal -->
      <div v-if="showViewPOModal && viewingPO" class="modal-backdrop" @click.self="showViewPOModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <div class="modal-header-left">
              <h2 class="modal-title">{{ viewingPO.poNumber }}</h2>
              <span :class="['badge badge-dot', statusConfig[viewingPO.status].badge]">{{ statusConfig[viewingPO.status].label }}</span>
            </div>
            <button class="modal-close" @click="showViewPOModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="view-info-grid">
              <div class="view-info-item">
                <span class="view-info-label">Supplier</span>
                <span class="view-info-value font-semibold">{{ viewingPO.supplierName }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Currency</span>
                <span class="view-info-value">{{ viewingPO.currency }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Expected Delivery</span>
                <span class="view-info-value">{{ formatDate(viewingPO.expectedDelivery) }}</span>
              </div>
              <div class="view-info-item" v-if="viewingPO.sourceQuoteNumber">
                <span class="view-info-label">Source Quote</span>
                <span class="view-info-value source-quote-link"><FileText :size="12" /> {{ viewingPO.sourceQuoteNumber }}</span>
              </div>
              <div class="view-info-item" v-if="viewingPO.projectName">
                <span class="view-info-label">Project</span>
                <span class="view-info-value">{{ viewingPO.projectName }}</span>
              </div>
            </div>
            <div v-if="viewingPO.notes" class="view-notes"><p>{{ viewingPO.notes }}</p></div>
            <div class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>#</th><th>SKU</th><th>Product</th><th>Manufacturer</th>
                    <th class="text-center">Qty</th><th class="text-right">Unit Cost</th>
                    <th class="text-right">Total</th><th class="text-center">Received</th><th class="text-center">Lead</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in viewingPO.items" :key="item.id">
                    <td class="text-muted">{{ idx + 1 }}</td>
                    <td class="text-mono font-medium">{{ item.productSku }}</td>
                    <td class="font-medium">{{ item.productName }}</td>
                    <td>{{ item.manufacturerName }}</td>
                    <td class="text-center">{{ item.quantity }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(item.total) }}</td>
                    <td class="text-center">
                      <span :class="item.receivedQty >= item.quantity ? 'text-success font-semibold' : 'text-warning'">{{ item.receivedQty }} / {{ item.quantity }}</span>
                    </td>
                    <td class="text-center text-muted">{{ item.leadTimeDays }}d</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="view-totals-card">
              <div class="view-totals-row"><span>Subtotal</span><span class="text-mono">SAR {{ formatSAR(viewingPO.subtotal) }}</span></div>
              <div class="view-totals-row"><span>Shipping</span><span class="text-mono">SAR {{ formatSAR(viewingPO.shippingCost) }}</span></div>
              <div class="view-totals-row"><span>Customs / Duty</span><span class="text-mono">SAR {{ formatSAR(viewingPO.customsDuty) }}</span></div>
              <div class="view-totals-row view-totals-row--grand">
                <span class="font-bold">Total</span>
                <span class="font-bold text-mono view-grand-total">SAR {{ formatSAR(viewingPO.total) }}</span>
              </div>
            </div>
            <div v-if="viewingPO.approvedBy" class="view-approval">
              <CheckCircle2 :size="16" class="text-success" />
              <span>Approved by <strong>{{ viewingPO.approvedBy }}</strong></span>
              <span v-if="viewingPO.approvedAt" class="text-muted"> on {{ formatDate(viewingPO.approvedAt) }}</span>
            </div>
            <RecordAttachments
              class="view-attachments"
              entity-type="purchase-order"
              :entity-id="viewingPO.id"
              :entity-label="viewingPO.poNumber"
            />
          </div>
          <div class="modal-footer">
            <button v-if="viewingPO.status === 'draft'" class="btn btn-sm" @click="poAction(viewingPO,'pending-approval')">Submit for approval</button>
            <button v-if="viewingPO.status === 'pending-approval'" class="btn btn-sm" @click="poAction(viewingPO,'approve')">Approve</button>
            <button v-if="viewingPO.status === 'approved'" class="btn btn-sm" @click="poAction(viewingPO,'ordered')">Mark ordered</button>
            <button v-if="['ordered','partial-received'].includes(viewingPO.status)" class="btn btn-sm" @click="receivePO(viewingPO)">Receive goods</button>
            <button class="btn btn-secondary" @click="showViewPOModal = false">Close</button>
          </div>
        </div>
      </div>

      <!-- View Supplier Quote Modal -->
      <div v-if="showViewSQModal && viewingSQ" class="modal-backdrop" @click.self="showViewSQModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <div class="modal-header-left">
              <h2 class="modal-title">{{ viewingSQ.sqNumber }}</h2>
              <span :class="['badge badge-dot', sqStatusConfig[viewingSQ.status].badge]">{{ sqStatusConfig[viewingSQ.status].label }}</span>
            </div>
            <button class="modal-close" @click="showViewSQModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="view-info-grid">
              <div class="view-info-item">
                <span class="view-info-label">Supplier</span>
                <span class="view-info-value font-semibold">{{ viewingSQ.supplierName }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Supplier Ref</span>
                <span class="view-info-value text-mono">{{ viewingSQ.supplierRef || '—' }}</span>
              </div>
              <div class="view-info-item" v-if="viewingSQ.sourceQuoteNumber">
                <span class="view-info-label">Customer Quote</span>
                <span class="view-info-value source-quote-link"><FileText :size="12" /> {{ viewingSQ.sourceQuoteNumber }}</span>
              </div>
              <div class="view-info-item" v-if="viewingSQ.projectName">
                <span class="view-info-label">Project</span>
                <span class="view-info-value">{{ viewingSQ.projectName }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Contact</span>
                <span class="view-info-value">{{ viewingSQ.contactName || '—' }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Email</span>
                <span class="view-info-value" style="font-size:0.75rem">{{ viewingSQ.contactEmail || '—' }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Valid Period</span>
                <span class="view-info-value">{{ formatDate(viewingSQ.validFrom) }} — {{ formatDate(viewingSQ.validUntil) }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Payment Terms</span>
                <span class="view-info-value">{{ viewingSQ.paymentTerms || '—' }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Delivery Terms</span>
                <span class="view-info-value">{{ viewingSQ.deliveryTerms || '—' }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Currency</span>
                <span class="view-info-value">{{ viewingSQ.currency }}</span>
              </div>
            </div>
            <div v-if="viewingSQ.notes" class="view-notes"><p>{{ viewingSQ.notes }}</p></div>

            <div class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>#</th><th>SKU</th><th>Product</th><th>Manufacturer</th>
                    <th class="text-center">Qty</th><th class="text-center">MOQ</th>
                    <th class="text-right">Unit Cost</th><th class="text-right">Total</th>
                    <th class="text-center">Lead Time</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in viewingSQ.items" :key="item.id">
                    <td class="text-muted">{{ idx + 1 }}</td>
                    <td class="text-mono font-medium">{{ item.productSku }}</td>
                    <td class="font-medium">{{ item.productName }}</td>
                    <td>{{ item.manufacturerName }}</td>
                    <td class="text-center">{{ item.quantity }}</td>
                    <td class="text-center text-muted">{{ item.moq ?? '—' }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(item.total) }}</td>
                    <td class="text-center">{{ item.leadTimeDays }}d</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="view-totals-card">
              <div class="view-totals-row view-totals-row--grand">
                <span class="font-bold">Total Value</span>
                <span class="font-bold text-mono view-grand-total">SAR {{ formatSAR(viewingSQ.subtotal) }}</span>
              </div>
            </div>
            <RecordAttachments
              class="view-attachments"
              entity-type="supplier-quote"
              :entity-id="viewingSQ.id"
              :entity-label="viewingSQ.sqNumber"
            />
          </div>
          <div class="modal-footer">
            <button v-if="['received','under-review'].includes(viewingSQ.status)" class="btn btn-sm" @click="acceptSQ(viewingSQ)">Accept supplier quote</button><button class="btn btn-secondary" @click="showViewSQModal = false">Close</button>
            <button v-if="viewingSQ.status === 'accepted'" class="btn btn-primary" @click="convertSQ(viewingSQ)">
              <ArrowRight :size="16" /> Create PO from This Quote
            </button>
          </div>
        </div>
      </div>

      <!-- View Goods Receipt Modal -->
      <div v-if="showViewGRModal && viewingGR" class="modal-backdrop" @click.self="showViewGRModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <h2 class="modal-title"><PackageCheck :size="20" /> {{ viewingGR.grNumber }}</h2>
            <button class="modal-close" @click="showViewGRModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="view-info-grid">
              <div class="view-info-item">
                <span class="view-info-label">PO Number</span>
                <span class="view-info-value source-quote-link"><ShoppingCart :size="12" /> {{ viewingGR.poNumber }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Supplier</span>
                <span class="view-info-value font-semibold">{{ viewingGR.supplierName }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Receive Date</span>
                <span class="view-info-value">{{ formatDate(viewingGR.receiveDate) }}</span>
              </div>
              <div class="view-info-item">
                <span class="view-info-label">Received By</span>
                <span class="view-info-value">{{ viewingGR.receivedBy }}</span>
              </div>
            </div>
            <div v-if="viewingGR.notes" class="view-notes"><p>{{ viewingGR.notes }}</p></div>

            <div class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>#</th><th>SKU</th><th>Product</th>
                    <th class="text-center">Qty</th><th class="text-right">Unit Cost</th>
                    <th class="text-right">Shipping</th><th class="text-right">Customs</th>
                    <th class="text-right">Landing Cost</th><th>Location</th><th>Condition</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in viewingGR.items" :key="item.id">
                    <td class="text-muted">{{ idx + 1 }}</td>
                    <td class="text-mono font-medium">{{ item.productSku }}</td>
                    <td class="font-medium">{{ item.productName }}</td>
                    <td class="text-center">{{ item.receivedQty }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(item.shippingAlloc) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(item.customsAlloc) }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(item.landingCost) }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ item.storageLocation || '—' }}</td>
                    <td>
                      <span :class="['badge badge-dot', item.condition === 'good' ? 'badge-success' : item.condition === 'damaged' ? 'badge-danger' : 'badge-warning']">
                        {{ item.condition === 'good' ? 'Good' : item.condition === 'damaged' ? 'Damaged' : 'Partial Damage' }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="view-totals-card">
              <div class="view-totals-row"><span>Total Items Received</span><span class="font-bold">{{ viewingGR.totalItems }}</span></div>
              <div class="view-totals-row view-totals-row--grand">
                <span class="font-bold">Total Landing Cost</span>
                <span class="font-bold text-mono view-grand-total">SAR {{ formatSAR(viewingGR.totalLandingCost) }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showViewGRModal = false">Close</button>
          </div>
        </div>
      </div>

      <!-- Item Detail Modal (Supplier Catalog) -->
      <div v-if="showItemDetailModal && detailItem" class="modal-backdrop" @click.self="showItemDetailModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <div class="modal-header-left">
              <h2 class="modal-title"><Package :size="20" /> Item Detail</h2>
              <div class="ph-product-badge">
                <span class="ph-sku">{{ detailItem.productSku }}</span>
                <span class="ph-name">{{ detailItem.productName }}</span>
              </div>
            </div>
            <button class="modal-close" @click="showItemDetailModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <!-- All suppliers for this item -->
            <h3 class="detail-section-title"><Building2 :size="15" /> Suppliers for This Item</h3>
            <div class="table-container table-container--embedded mb-5">
              <table class="table">
                <thead>
                  <tr>
                    <th>Supplier</th><th class="text-right">Latest Cost</th><th class="text-right">Previous Cost</th>
                    <th class="text-center">Trend</th><th class="text-center">MOQ</th><th class="text-center">Lead Time</th>
                    <th class="text-center">Reliability</th><th>Last Quote</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="si in detailAllSuppliers" :key="si.id" :class="{ 'detail-row-highlight': si.id === detailItem.id }">
                    <td class="font-medium">{{ si.supplierName }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(si.latestCost) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ si.previousCost ? 'SAR ' + formatSAR(si.previousCost) : '—' }}</td>
                    <td class="text-center">
                      <span v-if="si.costTrend === 'down'" class="trend-badge trend-badge--down"><TrendingDown :size="12" /></span>
                      <span v-else-if="si.costTrend === 'up'" class="trend-badge trend-badge--up"><TrendingUp :size="12" /></span>
                      <span v-else class="trend-badge trend-badge--stable"><Minus :size="12" /></span>
                    </td>
                    <td class="text-center">{{ si.moq }}</td>
                    <td class="text-center">{{ si.leadTimeDays }}d</td>
                    <td class="text-center">{{ si.reliability }}%</td>
                    <td class="text-muted whitespace-nowrap">{{ formatDate(si.lastQuoteDate) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Active supplier quotes for this item -->
            <h3 class="detail-section-title"><FileText :size="15" /> Active Supplier Quotes</h3>
            <div v-if="detailSupplierQuotes.length" class="table-container table-container--embedded mb-5">
              <table class="table">
                <thead>
                  <tr>
                    <th>Quote #</th><th>Supplier</th><th class="text-right">Unit Cost</th>
                    <th class="text-center">Qty</th><th class="text-center">Lead</th>
                    <th>Valid Until</th><th>Terms</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="entry in detailSupplierQuotes" :key="entry.sq.id + entry.item.id">
                    <td class="text-mono font-medium">{{ entry.sq.sqNumber }}</td>
                    <td class="font-medium">{{ entry.sq.supplierName }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(entry.item.unitCost) }}</td>
                    <td class="text-center">{{ entry.item.quantity }}</td>
                    <td class="text-center">{{ entry.item.leadTimeDays }}d</td>
                    <td class="text-muted whitespace-nowrap">{{ formatDate(entry.sq.validUntil) }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ entry.sq.paymentTerms || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="detail-empty">No active supplier quotes for this item.</div>

            <!-- Receipt history for this item -->
            <h3 class="detail-section-title"><History :size="15" /> Receive History & Landing Costs</h3>
            <div v-if="detailReceiptHistory.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>GR #</th><th>Receive Date</th><th>Supplier</th>
                    <th class="text-center">Qty</th><th class="text-right">Unit Cost</th>
                    <th class="text-right">Shipping</th><th class="text-right">Customs</th>
                    <th class="text-right">Landing Cost</th><th>Location</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="entry in detailReceiptHistory" :key="entry.gr.id + entry.item.id">
                    <td class="text-mono">{{ entry.gr.grNumber }}</td>
                    <td class="whitespace-nowrap">{{ formatDate(entry.gr.receiveDate) }}</td>
                    <td>{{ entry.gr.supplierName }}</td>
                    <td class="text-center">{{ entry.item.receivedQty }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(entry.item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(entry.item.shippingAlloc) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(entry.item.customsAlloc) }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(entry.item.landingCost) }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ entry.item.storageLocation || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="detail-empty">No receiving records for this item yet.</div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showItemDetailModal = false">Close</button>
          </div>
        </div>
      </div>

      <!-- Create / Edit Supplier Quote Modal -->
      <div v-if="showCreateSQModal" class="modal-backdrop" @click.self="showCreateSQModal = false; resetSQForm()">
        <div class="modal modal-xxl">
          <div class="modal-header">
            <h2 class="modal-title"><FileText :size="20" /> {{ editingSQId ? 'Edit Supplier Quote' : 'Add Supplier Quote' }}</h2>
            <button class="modal-close" @click="showCreateSQModal = false; resetSQForm()"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <!-- Supplier Info -->
            <div class="sq-form-section">
              <h4 class="sq-form-section-title"><Building2 :size="14" /> Supplier Details</h4>
              <div class="create-form-row">
                <div class="form-group" style="flex:2">
                  <label class="form-label">Supplier <span class="text-danger">*</span></label>
                  <div class="sq-supplier-pick">
                    <input v-model="sqForm.supplierName" type="text" class="form-input" placeholder="Type or select supplier..." />
                    <div v-if="!sqForm.supplierName" class="sq-supplier-chips">
                      <button v-for="mfr in mfrStore.suppliers.slice(0, 6)" :key="mfr.id" class="sq-supplier-chip" @click="selectSQSupplier(mfr)">
                        {{ mfr.name }}
                      </button>
                    </div>
                  </div>
                </div>
                <div class="form-group" style="flex:1">
                  <label class="form-label">Supplier Ref / Quote #</label>
                  <input v-model="sqForm.supplierRef" type="text" class="form-input" placeholder="e.g. HIK-QT-2026..." />
                </div>
              </div>
              <div class="create-form-row">
                <div class="form-group" style="flex:1">
                  <label class="form-label" for="sq-source-quote">Requested for customer quote</label>
                  <select id="sq-source-quote" v-model="sqForm.sourceQuoteId" class="form-input">
                    <option value="">Not linked</option>
                    <option v-for="qt in quotesStore.quotes" :key="qt.id" :value="qt.id">{{ qt.quoteNumber }} — {{ qt.customerName }}</option>
                  </select>
                </div>
                <div class="form-group" style="flex:1">
                  <label class="form-label" for="sq-project">Project</label>
                  <select id="sq-project" v-model="sqForm.projectId" class="form-input">
                    <option value="">Not linked</option>
                    <option v-for="pr in projectList" :key="pr.id" :value="pr.id">{{ pr.projectNumber }} — {{ pr.name }}</option>
                  </select>
                </div>
              </div>
              <div class="create-form-row">
                <div class="form-group" style="flex:1">
                  <label class="form-label">Contact Name</label>
                  <input v-model="sqForm.contactName" type="text" class="form-input" placeholder="Contact person" />
                </div>
                <div class="form-group" style="flex:1">
                  <label class="form-label">Contact Email</label>
                  <input v-model="sqForm.contactEmail" type="email" class="form-input" placeholder="email@supplier.com" />
                </div>
                <div class="form-group" style="width:150px">
                  <label class="form-label">Valid From</label>
                  <input v-model="sqForm.validFrom" type="date" class="form-input" />
                </div>
                <div class="form-group" style="width:150px">
                  <label class="form-label">Valid Until</label>
                  <input v-model="sqForm.validUntil" type="date" class="form-input" />
                </div>
              </div>
              <div class="create-form-row">
                <div class="form-group" style="flex:1">
                  <label class="form-label">Payment Terms</label>
                  <input v-model="sqForm.paymentTerms" type="text" class="form-input" placeholder="e.g. Net 30" />
                </div>
                <div class="form-group" style="flex:1">
                  <label class="form-label">Delivery Terms</label>
                  <input v-model="sqForm.deliveryTerms" type="text" class="form-input" placeholder="e.g. CIF Riyadh" />
                </div>
              </div>
            </div>

            <!-- Upload Section -->
            <div class="sq-form-section">
              <h4 class="sq-form-section-title"><Upload :size="14" /> Import from File</h4>
              <div class="sq-upload-area">
                <input ref="fileInputRef" type="file" accept=".csv,.tsv,.txt,.xlsx,.xls,.pdf" class="sq-file-input" @change="handleFileUpload" />
                <div class="sq-upload-box" @click="triggerFileUpload" @dragover.prevent @drop.prevent="readSQFile($event.dataTransfer?.files[0])">
                  <div v-if="!csvFileName" class="sq-upload-inner">
                    <Upload :size="24" />
                    <div class="sq-upload-text">
                      <span class="sq-upload-title">Drop file here or click to browse</span>
                      <span class="sq-upload-hint">Supports CSV, TSV, Excel (.xlsx), and PDF</span>
                    </div>
                    <div class="sq-upload-types">
                      <span class="sq-upload-type-tag"><Table2 :size="11" /> CSV / TSV</span>
                      <span class="sq-upload-type-tag"><FileSpreadsheet :size="11" /> Excel</span>
                      <span class="sq-upload-type-tag"><FileText :size="11" /> PDF</span>
                    </div>
                    <button class="btn-download-template" @click.stop="downloadTemplate('supplier-quote')"><Download :size="12" /> Download Excel Template</button>
                  </div>
                  <div v-else class="sq-upload-selected">
                    <Check :size="16" class="text-success" />
                    <span class="font-medium" style="font-size:0.82rem">{{ csvFileName }}</span>
                    <span class="badge badge-info" style="font-size:0.6rem">{{ csvFileType?.toUpperCase() }}</span>
                    <button class="btn btn-ghost btn-sm" @click.stop="csvFileName = ''; csvFileType = ''; csvPreviewItems = []; sqPasteText = ''; csvParseError = ''" style="margin-left:auto; font-size:0.7rem"><X :size="12" /> Remove</button>
                  </div>
                </div>
              </div>

              <!-- Parsing spinner -->
              <div v-if="sqFileParsing" class="sq-parsing-indicator">
                <RefreshCw :size="16" class="spin-icon" /> Parsing file...
              </div>

              <!-- Parse Error -->
              <div v-if="csvParseError && !sqFileParsing" class="sq-parse-error">
                <AlertCircle :size="14" /> {{ csvParseError }}
              </div>

              <!-- Paste fallback when parsing found nothing -->
              <div v-if="csvFileName && !sqFileParsing && csvPreviewItems.length === 0 && !csvParseError" class="sq-pdf-paste">
                <div class="sq-pdf-notice">
                  <AlertTriangle :size="14" />
                  <span>Could not extract structured data. Try pasting the content below, or <a href="#" class="fallback-template-link" @click.prevent.stop="downloadTemplate('supplier-quote')">download our Excel template</a> and re-upload.</span>
                </div>
                <textarea v-model="sqPasteText" class="form-input sq-paste-area" placeholder="Paste table data here (tab-separated or comma-separated)&#10;&#10;Example:&#10;SKU, Name, Manufacturer, Qty, Unit Cost&#10;CAM-001, IP Camera 4MP, Hikvision, 20, 245.00" rows="6"></textarea>
                <button class="btn btn-primary btn-sm" :disabled="!sqPasteText.trim()" @click="parseSQPastedText" style="margin-top:6px"><Table2 :size="13" /> Parse Pasted Data</button>
              </div>

              <!-- Preview -->
              <div v-if="csvPreviewItems.length" class="sq-csv-preview">
                <div class="sq-csv-preview-header">
                  <span class="font-semibold"><Check :size="14" /> {{ csvPreviewItems.length }} items parsed from <span class="text-mono">{{ csvFileName || 'pasted data' }}</span></span>
                  <div class="sq-csv-preview-actions">
                    <button class="btn btn-primary btn-sm" @click="importCSVItems">
                      <Plus :size="14" /> Import All Items
                    </button>
                    <button class="btn btn-ghost btn-sm" @click="clearCSVPreview">
                      <X :size="14" /> Dismiss
                    </button>
                  </div>
                </div>
                <div class="sq-csv-preview-table">
                  <table class="table">
                    <thead>
                      <tr>
                        <th>SKU</th><th>Name</th><th>Manufacturer</th>
                        <th class="text-center">Qty</th><th class="text-right">Unit Cost</th>
                        <th class="text-center">Lead</th><th class="text-center">MOQ</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(item, idx) in csvPreviewItems.slice(0, 10)" :key="idx">
                        <td class="text-mono" style="font-size:0.75rem">{{ item.sku || '—' }}</td>
                        <td style="font-size:0.75rem">{{ item.name }}</td>
                        <td class="text-muted" style="font-size:0.75rem">{{ item.manufacturer || '—' }}</td>
                        <td class="text-center">{{ item.qty }}</td>
                        <td class="text-right text-mono">{{ item.unitCost > 0 ? formatSAR(item.unitCost) : '—' }}</td>
                        <td class="text-center">{{ item.leadTimeDays }}d</td>
                        <td class="text-center">{{ item.moq }}</td>
                      </tr>
                    </tbody>
                  </table>
                  <div v-if="csvPreviewItems.length > 10" class="text-muted text-center" style="padding:var(--space-2); font-size:0.75rem">
                    ...and {{ csvPreviewItems.length - 10 }} more items
                  </div>
                </div>
              </div>
            </div>

            <!-- Line Items -->
            <div class="sq-form-section">
              <div class="sq-items-header">
                <h4 class="sq-form-section-title"><Package :size="14" /> Line Items</h4>
                <div class="sq-items-header-actions">
                  <!-- Catalog search -->
                  <div class="sq-item-search-wrap">
                    <div class="search-input" style="width:280px">
                      <Search :size="14" class="search-icon" />
                      <input v-model="sqItemSearch" type="text" class="form-input" style="font-size:0.75rem" placeholder="Search catalog..." @focus="showSQItemDropdown = sqItemSearch.trim().length > 0" @blur="delayHideSQDropdown" @input="showSQItemDropdown = sqItemSearch.trim().length > 0" />
                    </div>
                    <div v-if="showSQItemDropdown && sqItemSearchResults.length" class="sq-item-dropdown">
                      <button v-for="si in sqItemSearchResults" :key="si.id" class="po-dd-item" @mousedown.prevent="selectSQCatalogItem(si)">
                        <div class="po-dd-left">
                          <span class="po-dd-sku">{{ si.productSku }}</span>
                          <span class="po-dd-name">{{ si.productName }}</span>
                        </div>
                        <div class="po-dd-right">
                          <span class="po-dd-price">SAR {{ formatSAR(si.latestCost) }}</span>
                        </div>
                      </button>
                    </div>
                  </div>
                  <button class="btn btn-ghost btn-sm" @click="addBlankSQItem"><Plus :size="14" /> Manual Item</button>
                </div>
              </div>

              <div v-if="sqForm.items.length" class="table-container table-container--embedded sq-items-table">
                <table class="table">
                  <thead>
                    <tr>
                      <th>#</th><th>SKU</th><th>Product Name</th><th>Manufacturer</th>
                      <th class="text-center">Qty</th><th class="text-right">Unit Cost</th>
                      <th class="text-right">Total</th><th class="text-center">Lead</th>
                      <th class="text-center">MOQ</th><th></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(item, idx) in sqForm.items" :key="idx">
                      <td class="text-muted">{{ idx + 1 }}</td>
                      <td><input v-model="item.sku" type="text" class="form-input form-input--sm" placeholder="SKU" style="width:120px" /></td>
                      <td><input v-model="item.name" type="text" class="form-input form-input--sm" placeholder="Product name" style="min-width:160px" /></td>
                      <td><input v-model="item.manufacturer" type="text" class="form-input form-input--sm" placeholder="Mfr" style="width:120px" /></td>
                      <td class="text-center"><input v-model.number="item.qty" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:65px" /></td>
                      <td class="text-right"><input v-model.number="item.unitCost" type="number" step="0.01" min="0" class="form-input form-input--sm form-input--right" style="width:100px" /></td>
                      <td class="text-right whitespace-nowrap font-medium" style="font-size:0.75rem">SAR {{ formatSAR(item.qty * item.unitCost) }}</td>
                      <td class="text-center"><input v-model.number="item.leadTimeDays" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:55px" /></td>
                      <td class="text-center"><input v-model.number="item.moq" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:55px" /></td>
                      <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeSQItem(idx)"><Trash2 :size="14" /></button></td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr>
                      <td colspan="6" class="text-right font-bold" style="border-top:2px solid var(--color-neutral-300)">Total</td>
                      <td class="text-right font-bold whitespace-nowrap" style="border-top:2px solid var(--color-neutral-300)">SAR {{ formatSAR(sqFormSubtotal) }}</td>
                      <td colspan="3" style="border-top:2px solid var(--color-neutral-300)"></td>
                    </tr>
                  </tfoot>
                </table>
              </div>
              <div v-else class="create-empty">
                <Package :size="28" class="text-muted" />
                <p class="text-muted">Search the catalog, upload a vendor quote, or add items manually.</p>
              </div>
            </div>

            <!-- Notes -->
            <div class="form-group">
              <label class="form-label">Notes</label>
              <textarea v-model="sqForm.notes" class="form-input" rows="2" placeholder="Additional notes about this supplier quote..." />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showCreateSQModal = false; resetSQForm()">Cancel</button>
            <button class="btn btn-primary" :disabled="!sqForm.supplierName || sqForm.items.length === 0" @click="saveSQ">
              <Check :size="16" /> {{ editingSQId ? 'Save Changes' : 'Create Supplier Quote' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Enhanced Create PO Modal (4 Steps) -->
      <div v-if="showCreatePOModal" class="modal-backdrop" @click.self="showCreatePOModal = false; resetCreatePO()">
        <div class="modal modal-xxl">
          <div class="modal-header">
            <h2 class="modal-title"><ShoppingCart :size="20" /> Create Purchase Order</h2>
            <button class="modal-close" @click="showCreatePOModal = false; resetCreatePO()"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <!-- Step indicator -->
            <div class="po-steps">
              <div :class="['po-step', createPOStep === 'source' && 'po-step--active', createPOStep !== 'source' && 'po-step--done']">
                <span class="po-step-num">1</span> Source
              </div>
              <div class="po-step-divider" />
              <div :class="['po-step', createPOStep === 'supplier' && 'po-step--active', (createPOStep === 'items' || createPOStep === 'review') && 'po-step--done']">
                <span class="po-step-num">2</span> Supplier
              </div>
              <div class="po-step-divider" />
              <div :class="['po-step', createPOStep === 'items' && 'po-step--active', createPOStep === 'review' && 'po-step--done']">
                <span class="po-step-num">3</span> Items
              </div>
              <div class="po-step-divider" />
              <div :class="['po-step', createPOStep === 'review' && 'po-step--active']">
                <span class="po-step-num">4</span> Review
              </div>
            </div>

            <!-- Step 1: Select Source (Project / Quote / Manual) -->
            <div v-if="createPOStep === 'source'">
              <div class="po-source-tabs">
                <button :class="['po-source-tab', poSourceType === 'project' && 'po-source-tab--active']" @click="poSourceType = 'project'">Projects (Won Quotes)</button>
                <button :class="['po-source-tab', poSourceType === 'quote' && 'po-source-tab--active']" @click="poSourceType = 'quote'">All Quotes</button>
                <button :class="['po-source-tab', poSourceType === 'manual' && 'po-source-tab--active']" @click="poSourceType = 'manual'">No Source</button>
              </div>

              <template v-if="poSourceType !== 'manual'">
                <div class="search-input mb-4" style="max-width:400px">
                  <Search :size="16" class="search-icon" />
                  <input v-model="sourceSearchQuery" type="text" class="form-input" placeholder="Search by quote #, customer name..." />
                </div>
                <div v-if="(poSourceType === 'project' ? availableProjects : availableQuotes).length" class="po-source-grid">
                  <button v-for="qt in (poSourceType === 'project' ? availableProjects : availableQuotes)" :key="qt.id" class="po-source-card" @click="selectSource(qt)">
                    <div class="po-source-card-top">
                      <span class="text-mono font-bold" style="font-size:0.78rem">{{ qt.quoteNumber }}</span>
                      <span :class="['badge', qt.status === 'accepted' ? 'badge-success' : qt.status === 'sent' ? 'badge-info' : 'badge-gray']" style="font-size:0.6rem">{{ qt.status }}</span>
                    </div>
                    <div class="po-source-card-customer font-medium">{{ qt.customerName }}</div>
                    <div class="po-source-card-meta">
                      <span>{{ qt.lineItems.filter(l => l.category === 'materials').length }} material items</span>
                      <span class="font-bold" style="color:var(--color-primary)">SAR {{ formatSAR(qt.total) }}</span>
                    </div>
                    <div v-if="qt.lineItems.filter(l => l.category === 'materials').length" class="po-source-card-items">
                      <div v-for="li in qt.lineItems.filter(l => l.category === 'materials').slice(0, 4)" :key="li.id" class="po-src-inline-item">
                        <span v-if="li.sku" class="po-src-inline-sku">{{ li.sku }}</span>
                        <span class="po-src-inline-desc">{{ li.description.length > 30 ? li.description.slice(0, 28) + '...' : li.description }}</span>
                        <span class="po-src-inline-qty">× {{ li.quantity }}</span>
                      </div>
                      <span v-if="qt.lineItems.filter(l => l.category === 'materials').length > 4" class="po-src-inline-more">+{{ qt.lineItems.filter(l => l.category === 'materials').length - 4 }} more</span>
                    </div>
                    <div v-if="!qt.lineItems.filter(l => l.category === 'materials').length" class="po-source-card-desc text-muted">{{ qt.notes.slice(0, 80) }}</div>
                  </button>
                </div>
                <div v-else class="detail-empty">No {{ poSourceType === 'project' ? 'won quotes / projects' : 'quotes' }} found.</div>
              </template>
              <template v-else>
                <div class="po-skip-source">
                  <Package :size="32" class="text-muted" />
                  <p class="text-muted" style="font-size:0.85rem">Create a PO without linking to a project or quote.</p>
                  <button class="btn btn-primary" @click="skipSource">Continue to Supplier Selection <ArrowRight :size="16" /></button>
                </div>
              </template>
            </div>

            <!-- Step 2: Select Supplier -->
            <div v-if="createPOStep === 'supplier'">
              <div v-if="selectedSourceQuote" class="po-selected-source-bar">
                <div class="po-source-bar-header">
                  <span class="text-mono font-bold">{{ selectedSourceQuote.quoteNumber }}</span>
                  <span class="font-medium">{{ selectedSourceQuote.customerName }}</span>
                  <span class="text-muted">{{ sourceItems.length }} material item{{ sourceItems.length !== 1 ? 's' : '' }}</span>
                  <button class="btn btn-ghost btn-sm" @click="createPOStep = 'source'; selectedSourceQuote = null" style="font-size:0.68rem"><X :size="12" /> Change</button>
                </div>
                <!-- Source items preview -->
                <div v-if="sourceItems.length" class="po-source-items-preview">
                  <div v-for="li in sourceItems" :key="li.id" class="po-source-item-chip">
                    <span v-if="li.sku" class="po-src-sku">{{ li.sku }}</span>
                    <span class="po-src-name">{{ li.description }}</span>
                    <span class="po-src-qty">× {{ li.quantity }}</span>
                  </div>
                </div>
              </div>

              <p class="text-muted" style="font-size:0.82rem; margin-bottom: var(--space-3)">
                <template v-if="sourceItems.length">Select a supplier — items from your source are matched against each supplier's catalog. Best matches are shown first.</template>
                <template v-else>Select a supplier that has catalogs or quotes uploaded.</template>
              </p>

              <div class="supplier-select-section">
                <div class="search-input mb-4" style="max-width:400px">
                  <Search :size="16" class="search-icon" />
                  <input v-model="supplierSearchQuery" type="text" class="form-input" placeholder="Search suppliers by name or code..." />
                </div>

                <div v-if="filteredSupplierList.length" class="supplier-grid">
                  <button v-for="mfr in filteredSupplierList" :key="mfr.id" class="supplier-card" :class="{ 'supplier-card--has-catalog': supplierHasCatalog(mfr.name), 'supplier-card--full-match': sourceItems.length > 0 && getSupplierCoverage(mfr.name).coveragePercent === 100, 'supplier-card--partial-match': sourceItems.length > 0 && getSupplierCoverage(mfr.name).matchCount > 0 && getSupplierCoverage(mfr.name).coveragePercent < 100 }" @click="selectSupplierForPO(mfr)">
                    <div class="supplier-card-top">
                      <span class="supplier-card-name">{{ mfr.name }}</span>
                      <span class="badge badge-neutral text-mono" style="font-size:0.625rem">{{ mfr.code }}</span>
                    </div>
                    <div class="supplier-card-meta">
                      <span>{{ mfr.country }}</span>
                      <span :class="['badge', mfr.vendorType === 'supplier' ? 'badge-primary' : 'badge-success']" style="font-size:0.5625rem">
                        {{ mfr.vendorType === 'both' ? 'Mfr + Supplier' : 'Supplier' }}
                      </span>
                    </div>
                    <div class="supplier-card-cats">
                      <span v-for="cat in mfr.categories.slice(0, 3)" :key="cat.id" class="supplier-card-cat">{{ cat.name }}</span>
                    </div>

                    <!-- Coverage info when a source is selected -->
                    <template v-if="sourceItems.length > 0">
                      <div class="supplier-coverage">
                        <div class="supplier-coverage-bar-track">
                          <div class="supplier-coverage-bar-fill" :style="{ width: getSupplierCoverage(mfr.name).coveragePercent + '%' }" :class="{ 'cov-full': getSupplierCoverage(mfr.name).coveragePercent === 100, 'cov-partial': getSupplierCoverage(mfr.name).coveragePercent > 0 && getSupplierCoverage(mfr.name).coveragePercent < 100, 'cov-none': getSupplierCoverage(mfr.name).coveragePercent === 0 }"></div>
                        </div>
                        <span class="supplier-coverage-text" :class="{ 'cov-text-full': getSupplierCoverage(mfr.name).coveragePercent === 100 }">
                          {{ getSupplierCoverage(mfr.name).matchCount }} / {{ getSupplierCoverage(mfr.name).totalSourceItems }} items matched
                        </span>
                      </div>
                      <div v-if="getSupplierCoverage(mfr.name).matchedNames.length" class="supplier-matched-items">
                        <span v-for="(n, idx) in getSupplierCoverage(mfr.name).matchedNames.slice(0, 3)" :key="idx" class="supplier-matched-chip"><Check :size="9" /> {{ n }}</span>
                        <span v-if="getSupplierCoverage(mfr.name).matchedNames.length > 3" class="supplier-matched-more">+{{ getSupplierCoverage(mfr.name).matchedNames.length - 3 }} more</span>
                      </div>
                    </template>

                    <!-- Catalog badge (shown when no source or as secondary info) -->
                    <div v-if="sourceItems.length === 0 && supplierHasCatalog(mfr.name)" class="supplier-card-catalog-badge">
                      <Check :size="10" /> Has Catalog / Quotes
                    </div>
                    <div v-if="sourceItems.length === 0 && !supplierHasCatalog(mfr.name)" class="supplier-card-no-catalog">
                      <AlertCircle :size="10" /> No catalog uploaded
                    </div>
                  </button>
                </div>
                <div v-else class="detail-empty">No suppliers match your search.</div>

                <div class="supplier-manual-entry">
                  <span class="text-muted" style="font-size:var(--text-sm)">Can't find the supplier?</span>
                  <div class="supplier-manual-row">
                    <input v-model="supplierSearchQuery" type="text" class="form-input" placeholder="Enter new supplier name..." style="flex:1; max-width:300px" />
                    <button class="btn btn-secondary btn-sm" :disabled="!supplierSearchQuery.trim()" @click="enterNewSupplier">
                      <Plus :size="14" /> Use as New Supplier
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 3: Items -->
            <div v-if="createPOStep === 'items'">
              <div class="selected-supplier-bar">
                <div class="selected-supplier-info">
                  <Building2 :size="16" />
                  <span class="font-semibold">{{ newPOSupplier }}</span>
                  <span v-if="selectedSourceQuote" class="text-muted" style="font-size:0.72rem">| Source: {{ selectedSourceQuote.quoteNumber }} — {{ selectedSourceQuote.customerName }}</span>
                  <button class="btn btn-ghost btn-sm" @click="createPOStep = 'supplier'; newPOItems = []" style="font-size:0.6875rem; padding:2px 8px">Change Supplier</button>
                </div>
              </div>

              <!-- Auto-match results -->
              <div v-if="selectedSourceQuote && matchedCount > 0" class="po-match-banner po-match-banner--success">
                <Check :size="14" /> <strong>{{ matchedCount }} item{{ matchedCount !== 1 ? 's' : '' }}</strong> auto-matched from {{ newPOSupplier }}'s catalog/quotes
              </div>
              <div v-if="selectedSourceQuote && unmatchedSourceItems.length > 0" class="po-match-banner po-match-banner--warn">
                <AlertTriangle :size="14" /> <strong>{{ unmatchedSourceItems.length }} item{{ unmatchedSourceItems.length !== 1 ? 's' : '' }}</strong> from the source could not be matched in this supplier's catalog:
                <span class="po-unmatched-list">{{ unmatchedSourceItems.map(li => li.sku || li.description).join(', ') }}</span>
              </div>

              <!-- Available quotes from this supplier -->
              <div v-if="supplierAvailableQuotes.length" class="sq-available-section">
                <h4 class="sq-available-title"><FileText :size="14" /> Supplier Quotes & Catalogs</h4>
                <div class="sq-available-list">
                  <div v-for="sq in supplierAvailableQuotes" :key="sq.id" class="sq-available-card">
                    <div class="sq-available-left">
                      <span class="text-mono font-semibold" style="font-size:0.75rem">{{ sq.sqNumber }}</span>
                      <span class="text-muted" style="font-size:0.6875rem">{{ sq.items.length }} items · SAR {{ formatSAR(sq.subtotal) }}</span>
                      <span class="text-muted" style="font-size:0.625rem">Valid until {{ formatDate(sq.validUntil) }}</span>
                    </div>
                    <button class="btn btn-primary btn-sm" @click="convertSQ(sq)" style="font-size:0.6875rem">
                      <ArrowRight :size="13" /> Import Items
                    </button>
                  </div>
                </div>
              </div>

              <div class="create-form-row">
                <div class="form-group" style="width:200px"><label class="form-label">Expected Delivery</label><input v-model="newPOExpected" type="date" class="form-input" /></div>
                <div class="form-group" style="width:160px"><label class="form-label">Source SQ</label><input v-model="newPOSourceSQ" type="text" class="form-input" placeholder="SQ-..." readonly /></div>
              </div>

              <div class="po-item-search-row">
                <div class="po-item-search-wrapper">
                  <div class="search-input">
                    <Search :size="16" class="search-icon" />
                    <input v-model="newPOItemSearch" type="text" class="form-input" placeholder="Search catalog by SKU, name, or manufacturer..." @focus="showPOItemDropdown = newPOItemSearch.trim().length > 0" @blur="delayHidePODropdown" @input="showPOItemDropdown = newPOItemSearch.trim().length > 0" />
                  </div>
                  <div v-if="showPOItemDropdown && poItemSearchResults.length" class="po-item-dropdown">
                    <button v-for="si in poItemSearchResults" :key="si.id" class="po-dd-item" @mousedown.prevent="selectPOItem(si)">
                      <div class="po-dd-left"><span class="po-dd-sku">{{ si.productSku }}</span><span class="po-dd-name">{{ si.productName }}</span><span class="po-dd-meta">{{ si.supplierName }} · {{ si.manufacturerName }}</span></div>
                      <div class="po-dd-right"><span class="po-dd-price">SAR {{ formatSAR(si.latestCost) }}</span><span class="po-dd-lead">{{ si.leadTimeDays }}d lead · MOQ {{ si.moq }}</span></div>
                    </button>
                  </div>
                </div>
                <button class="btn btn-secondary btn-sm" @click="openPOFileUpload"><FileUp :size="14" /> Upload File</button>
                <button class="btn btn-ghost btn-sm" @click="addBlankPOItem"><Plus :size="14" /> Manual Item</button>
              </div>

              <!-- File upload area (inline, expandable) -->
              <div v-if="showPOFileUpload" class="po-file-upload-section">
                <div class="po-file-upload-header">
                  <h4 class="po-file-title"><FileUp :size="16" /> Import Items from File</h4>
                  <button class="btn btn-ghost btn-icon btn-sm" @click="closePOFileUpload"><X :size="16" /></button>
                </div>

                <div class="po-file-upload-zone" @click="triggerPOFileUpload" @dragover.prevent @drop.prevent="readPOFile($event.dataTransfer?.files[0])">
                  <input ref="poFileRef" type="file" accept=".csv,.tsv,.txt,.xlsx,.xls,.pdf" style="display:none" @change="handlePOFile" />
                  <div v-if="!poFileName" class="po-file-placeholder">
                    <Upload :size="24" class="text-muted" />
                    <span class="po-file-placeholder-text">Drop file here or click to browse</span>
                    <div class="po-file-types">
                      <span class="po-file-type-tag"><Table2 :size="11" /> CSV / TSV</span>
                      <span class="po-file-type-tag"><FileSpreadsheet :size="11" /> Excel</span>
                      <span class="po-file-type-tag"><FileText :size="11" /> PDF</span>
                    </div>
                    <button class="btn-download-template" @click.stop="downloadTemplate('purchase-order')"><Download :size="12" /> Download Excel Template</button>
                  </div>
                  <div v-else class="po-file-selected">
                    <Check :size="16" class="text-success" />
                    <span class="font-medium" style="font-size:0.82rem">{{ poFileName }}</span>
                    <span class="badge badge-info" style="font-size:0.6rem">{{ poFileType?.toUpperCase() }}</span>
                    <button class="btn btn-ghost btn-sm" @click.stop="poFileName = ''; poFileType = ''; poFileRows = []; poPasteText = ''; poFileError = ''" style="margin-left:auto; font-size:0.7rem"><X :size="12" /> Remove</button>
                  </div>
                </div>

                <div v-if="poFileParsing" class="po-file-parsing">
                  <RefreshCw :size="16" class="spin-icon" /> Parsing file...
                </div>

                <div v-if="poFileError && !poFileParsing" class="po-file-error">
                  <AlertTriangle :size="14" /> {{ poFileError }}
                </div>

                <!-- Paste fallback when parsing found nothing -->
                <div v-if="poFileName && !poFileParsing && poFileRows.length === 0 && !poFileError" class="po-file-pdf-paste">
                  <div class="po-file-pdf-notice">
                    <AlertTriangle :size="14" />
                    <span>Could not extract structured data. Try pasting the content below, or <a href="#" class="fallback-template-link" @click.prevent.stop="downloadTemplate('purchase-order')">download our Excel template</a> and re-upload.</span>
                  </div>
                  <textarea v-model="poPasteText" class="form-input po-file-paste-area" placeholder="Paste table data here (tab-separated or comma-separated)&#10;&#10;Example:&#10;SKU, Description, Qty, Unit Price&#10;CAM-001, IP Camera 4MP, 20, 245.00" rows="6"></textarea>
                  <button class="btn btn-primary btn-sm" :disabled="!poPasteText.trim()" @click="parsePOPastedText" style="margin-top:6px"><Table2 :size="13" /> Parse Data</button>
                </div>

                <!-- Preview parsed rows -->
                <div v-if="poFileRows.length > 0" class="po-file-preview">
                  <div class="po-file-preview-header">
                    <span class="font-medium" style="font-size:0.82rem">{{ poFileRows.length }} items detected</span>
                    <span class="text-muted" style="font-size:0.72rem">{{ poFileSelectedCount }} selected</span>
                    <div style="margin-left:auto; display:flex; gap:6px">
                      <button class="btn btn-ghost btn-sm" style="font-size:0.7rem" @click="poFileRows.forEach(r => r.selected = true)">All</button>
                      <button class="btn btn-ghost btn-sm" style="font-size:0.7rem" @click="poFileRows.forEach(r => r.selected = false)">None</button>
                    </div>
                  </div>
                  <div class="table-container table-container--embedded" style="max-height:250px; overflow-y:auto">
                    <table class="table">
                      <thead>
                        <tr>
                          <th style="width:36px"><input type="checkbox" :checked="poFileSelectedCount === poFileRows.length" @change="poFileRows.forEach(r => r.selected = ($event.target as HTMLInputElement).checked)" /></th>
                          <th>SKU</th><th>Product / Description</th><th class="text-center">Qty</th><th class="text-right">Unit Cost</th><th>Manufacturer</th><th class="text-center">Lead</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="(row, idx) in poFileRows" :key="idx" :class="{ 'po-file-row--off': !row.selected }">
                          <td><input type="checkbox" v-model="row.selected" /></td>
                          <td><input v-model="row.sku" type="text" class="form-input form-input--sm" style="width:120px" /></td>
                          <td><input v-model="row.name" type="text" class="form-input form-input--sm" style="min-width:160px" /></td>
                          <td class="text-center"><input v-model.number="row.qty" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:65px" /></td>
                          <td class="text-right"><input v-model.number="row.unitCost" type="number" step="0.01" class="form-input form-input--sm form-input--right" style="width:100px" /></td>
                          <td><input v-model="row.manufacturer" type="text" class="form-input form-input--sm" style="width:130px" /></td>
                          <td class="text-center"><input v-model.number="row.leadTimeDays" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:60px" /></td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <div class="po-file-preview-footer">
                    <button class="btn btn-primary btn-sm" :disabled="poFileSelectedCount === 0" @click="importPOFileItems">
                      <Check :size="14" /> Import {{ poFileSelectedCount }} Item{{ poFileSelectedCount !== 1 ? 's' : '' }} to PO
                    </button>
                  </div>
                </div>
              </div>

              <div v-if="newPOItems.length" class="table-container table-container--embedded po-items-table">
                <table class="table">
                  <thead><tr><th>#</th><th>SKU</th><th>Product</th><th class="text-center">On Hand</th><th class="text-center">Qty</th><th class="text-right">Unit Cost</th><th class="text-right">Best Price</th><th class="text-right">Last Landing</th><th>Last Received</th><th class="text-right">Line Total</th><th class="text-center">Lead</th><th>Source</th><th></th></tr></thead>
                  <tbody>
                    <tr v-for="(item, idx) in newPOItems" :key="idx" :class="item.matched ? 'po-row--matched' : ''">
                      <td class="text-muted">{{ idx + 1 }}</td>
                      <td><input v-model="item.sku" type="text" class="form-input form-input--sm" placeholder="SKU" style="width:130px" /></td>
                      <td><input v-model="item.name" type="text" class="form-input form-input--sm" placeholder="Product name" style="min-width:160px" /></td>
                      <td class="text-center"><span :class="['onhand-badge', item.onHand > 0 ? 'onhand-badge--has' : 'onhand-badge--zero']">{{ item.onHand }}</span></td>
                      <td class="text-center"><input v-model.number="item.qty" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:70px" /></td>
                      <td class="text-right"><input v-model.number="item.unitCost" type="number" step="0.01" min="0" class="form-input form-input--sm form-input--right" style="width:110px" /></td>
                      <td class="text-right"><span v-if="item.bestSupplierPrice !== null" class="best-price-badge" :class="{ 'best-price--better': item.bestSupplierPrice < item.unitCost, 'best-price--same': item.bestSupplierPrice === item.unitCost, 'best-price--worse': item.bestSupplierPrice > item.unitCost }" :title="item.bestSupplierName">SAR {{ formatSAR(item.bestSupplierPrice) }}</span><span v-else class="text-muted">—</span></td>
                      <td class="text-right"><span v-if="item.landingCost !== null" class="text-mono" style="font-size:0.75rem">SAR {{ formatSAR(item.landingCost) }}</span><span v-else class="text-muted">—</span></td>
                      <td><span v-if="item.lastReceiveDate" class="text-muted whitespace-nowrap" style="font-size:0.75rem">{{ formatDate(item.lastReceiveDate) }}</span><span v-else class="text-muted">Never</span></td>
                      <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(item.qty * item.unitCost) }}</td>
                      <td class="text-center"><input v-model.number="item.leadTimeDays" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:55px" /></td>
                      <td>
                        <span v-if="item.matched" class="badge badge-success" style="font-size:0.58rem">Matched</span>
                        <span v-else-if="item.fromSQ" class="badge badge-info" style="font-size:0.58rem">{{ item.sqRef }}</span>
                        <span v-else class="badge badge-gray" style="font-size:0.58rem">Manual</span>
                      </td>
                      <td><button class="btn btn-ghost btn-icon btn-sm" @click="removePOItem(idx)"><Trash2 :size="14" /></button></td>
                    </tr>
                  </tbody>
                  <tfoot><tr><td colspan="9" class="text-right font-bold" style="border-top:2px solid var(--color-neutral-300)">Subtotal</td><td class="text-right font-bold whitespace-nowrap" style="border-top:2px solid var(--color-neutral-300)">SAR {{ formatSAR(newPOSubtotal) }}</td><td colspan="3" style="border-top:2px solid var(--color-neutral-300)"></td></tr></tfoot>
                </table>
              </div>
              <div v-else class="create-empty"><Package :size="28" class="text-muted" /><p class="text-muted">Search the catalog to add items, or add manually.</p></div>

              <div class="form-group mt-4"><label class="form-label">Notes</label><textarea v-model="newPONotes" class="form-input" rows="2" placeholder="Purchase order notes..." /></div>
            </div>

            <!-- Step 4: Review -->
            <div v-if="createPOStep === 'review'" class="po-review">
              <div class="view-info-grid">
                <div class="view-info-item"><span class="view-info-label">Supplier</span><span class="view-info-value font-semibold">{{ newPOSupplier }}</span></div>
                <div class="view-info-item" v-if="selectedSourceQuote"><span class="view-info-label">Source Quote</span><span class="view-info-value source-quote-link"><FileText :size="12" /> {{ selectedSourceQuote.quoteNumber }} — {{ selectedSourceQuote.customerName }}</span></div>
                <div class="view-info-item"><span class="view-info-label">Expected Delivery</span><span class="view-info-value">{{ newPOExpected ? formatDate(newPOExpected) : 'Auto (30 days)' }}</span></div>
                <div class="view-info-item" v-if="newPOSourceSQ"><span class="view-info-label">Supplier Quote Ref</span><span class="view-info-value source-quote-link"><FileText :size="12" /> {{ newPOSourceSQ }}</span></div>
                <div class="view-info-item"><span class="view-info-label">Total Items</span><span class="view-info-value font-bold">{{ newPOItems.length }} ({{ newPOItems.reduce((s, i) => s + i.qty, 0) }} units)</span></div>
                <div v-if="matchedCount > 0" class="view-info-item"><span class="view-info-label">Auto-Matched</span><span class="view-info-value text-success font-bold">{{ matchedCount }} items</span></div>
              </div>
              <div class="view-totals-card mt-4">
                <div class="view-totals-row"><span>Subtotal</span><span class="text-mono">SAR {{ formatSAR(newPOSubtotal) }}</span></div>
                <div class="view-totals-row"><span>Shipping (not entered)</span><span class="text-mono">SAR {{ formatSAR(0) }}</span></div>
                <div class="view-totals-row"><span>Customs (not entered)</span><span class="text-mono">SAR {{ formatSAR(0) }}</span></div>
                <div class="view-totals-row view-totals-row--grand"><span class="font-bold">Estimated Total</span><span class="font-bold text-mono view-grand-total">SAR {{ formatSAR(newPOSubtotal + 0 + 0) }}</span></div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showCreatePOModal = false; resetCreatePO()">Cancel</button>
            <div class="modal-footer-right">
              <button v-if="createPOStep === 'supplier'" class="btn btn-secondary" @click="createPOStep = 'source'">Back</button>
              <button v-if="createPOStep === 'items'" class="btn btn-secondary" @click="createPOStep = 'supplier'; newPOItems = []">Back</button>
              <button v-if="createPOStep === 'review'" class="btn btn-secondary" @click="createPOStep = 'items'">Back</button>
              <button v-if="createPOStep === 'items'" class="btn btn-primary" :disabled="!newPOSupplier || newPOItems.length === 0" @click="createPOStep = 'review'">
                Review <ArrowRight :size="16" />
              </button>
              <button v-if="createPOStep === 'review'" class="btn btn-primary" @click="createPO">
                <Plus :size="16" /> Create Purchase Order
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.view-attachments { margin-top: var(--space-4); padding-top: var(--space-4); border-top: 1px solid var(--color-neutral-200); }

.procurement-page { padding: var(--space-6); }
.page-header-actions { display: flex; gap: var(--space-2); }

.kpi-grid--6 { display: grid; grid-template-columns: repeat(6, 1fr); gap: var(--space-4); margin-bottom: var(--space-6); }
.kpi-card { padding: var(--space-4); }
.kpi-top { display: flex; justify-content: space-between; align-items: flex-start; }
.kpi-label { font-size: var(--text-xs); font-weight: var(--font-medium); color: var(--color-neutral-500); margin-bottom: var(--space-1); }
.kpi-value { font-size: var(--text-xl); font-weight: var(--font-bold); color: var(--color-neutral-900); line-height: var(--leading-tight); }
.kpi-value--sm { font-size: var(--text-base); }
.kpi-icon { width: 40px; height: 40px; border-radius: var(--radius-lg); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.kpi-icon--primary { background-color: var(--color-primary-light); color: var(--color-primary); }
.kpi-icon--warning { background-color: var(--color-warning-light); color: var(--color-warning); }
.kpi-icon--info { background-color: var(--color-primary-100); color: var(--color-primary-700); }
.kpi-icon--success { background-color: var(--color-success-light); color: var(--color-success); }
.kpi-icon--purple { background: rgba(139, 92, 246, .1); color: #8b5cf6; }
.kpi-icon--teal { background: rgba(20, 184, 166, .1); color: #14b8a6; }

/* Tabs */
.proc-tabs { display: flex; gap: var(--space-1); margin-bottom: var(--space-5); background: var(--color-neutral-100); padding: 4px; border-radius: var(--radius-lg); }
.proc-tab {
  display: flex; align-items: center; gap: var(--space-2); padding: var(--space-2) var(--space-4);
  border: none; background: transparent; border-radius: var(--radius-md); cursor: pointer;
  font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-600); transition: all .15s;
}
.proc-tab:hover { background: var(--content-surface); color: var(--color-neutral-800); }
.proc-tab--active { background: var(--content-surface); color: var(--color-primary); box-shadow: var(--shadow-sm); font-weight: 600; }
.proc-tab-count { font-size: 0.625rem; background: var(--color-neutral-200); padding: 1px 6px; border-radius: 9999px; font-weight: 600; }
.proc-tab--active .proc-tab-count { background: var(--color-primary-50); color: var(--color-primary); }

.toolbar { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) var(--space-5); flex-wrap: wrap; }
.toolbar-search { flex: 1; min-width: 220px; }
.toolbar-select { width: 200px; flex-shrink: 0; }

.modal-header-left { display: flex; align-items: center; gap: var(--space-3); }

.source-quote-link {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: var(--text-xs); font-weight: 500; color: var(--color-primary);
  background: var(--color-primary-light, #eff6ff); padding: 2px 8px; border-radius: var(--radius-sm);
}

.receipt-cell { display: flex; align-items: center; gap: var(--space-2); justify-content: center; }
.receipt-bar-track { width: 50px; height: 6px; background: var(--color-neutral-200); border-radius: var(--radius-full); overflow: hidden; }
.receipt-bar-fill { height: 100%; background: var(--color-warning); border-radius: var(--radius-full); transition: width 0.4s ease; }
.receipt-bar-fill.receipt-complete { background: var(--color-success); }
.receipt-label { font-size: var(--text-xs); color: var(--color-neutral-500); font-weight: 500; white-space: nowrap; }

.view-info-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: var(--space-4); padding: var(--space-4); background-color: var(--color-neutral-50); border-radius: var(--radius-lg); margin-bottom: var(--space-5); }
.view-info-label { display: block; font-size: var(--text-xs); color: var(--color-neutral-500); font-weight: var(--font-medium); text-transform: uppercase; letter-spacing: 0.03em; margin-bottom: var(--space-1); }
.view-info-value { font-size: var(--text-sm); color: var(--color-neutral-800); }
.view-notes { padding: var(--space-3) var(--space-4); background-color: var(--color-warning-light); border-left: 3px solid var(--color-warning); border-radius: var(--radius-md); margin-bottom: var(--space-5); font-size: var(--text-sm); color: var(--color-neutral-700); line-height: var(--leading-relaxed); }
.table-container--embedded { border: none; border-radius: 0; box-shadow: none; margin-bottom: var(--space-5); }

.view-totals-card { background-color: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); padding: var(--space-4) var(--space-5); margin-bottom: var(--space-4); }
.view-totals-row { display: flex; justify-content: space-between; align-items: center; padding: var(--space-2) 0; font-size: var(--text-sm); color: var(--color-neutral-700); }
.view-totals-row--grand { border-top: 2px solid var(--color-neutral-300); padding-top: var(--space-3); margin-top: var(--space-1); }
.view-grand-total { font-size: var(--text-lg); color: var(--color-primary); }
.view-approval { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-3) var(--space-4); background-color: var(--color-success-light); border-radius: var(--radius-md); font-size: var(--text-sm); color: var(--color-neutral-700); }

/* Supplier Quotes */
.sq-days-left { font-size: 0.625rem; margin-left: 4px; }

/* Trend badges */
.trend-badge { display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; border-radius: var(--radius-full); }
.trend-badge--down { background: rgba(16, 185, 129, .1); color: var(--color-success); }
.trend-badge--up { background: rgba(239, 68, 68, .1); color: var(--color-danger); }
.trend-badge--stable { background: var(--color-neutral-100); color: var(--color-neutral-500); }

/* Reliability bar */
.reliability-bar { width: 40px; height: 4px; background: var(--color-neutral-200); border-radius: var(--radius-full); overflow: hidden; display: inline-block; vertical-align: middle; margin-right: 4px; }
.reliability-fill { height: 100%; border-radius: var(--radius-full); }
.reliability--good { background: var(--color-success); }
.reliability--ok { background: var(--color-warning); }
.reliability--poor { background: var(--color-danger); }
.reliability-label { font-size: 0.625rem; color: var(--color-neutral-500); vertical-align: middle; }

/* Item detail modal */
.detail-section-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: 600; color: var(--color-neutral-700); text-transform: uppercase; letter-spacing: .03em; margin-bottom: var(--space-3); padding-bottom: var(--space-2); border-bottom: 1px solid var(--color-neutral-200); }
.detail-row-highlight { background: var(--color-primary-50) !important; }
.detail-empty { text-align: center; padding: var(--space-6); color: var(--color-neutral-400); font-size: var(--text-sm); font-style: italic; margin-bottom: var(--space-5); }

.ph-product-badge { display: flex; align-items: center; gap: var(--space-3); font-size: var(--text-sm); }
.ph-sku { font-family: var(--font-mono); color: var(--color-primary); font-weight: 600; background: var(--color-primary-50); padding: 1px 8px; border-radius: var(--radius-sm); }
.ph-name { color: var(--color-neutral-700); font-weight: 500; }

/* Create PO Modal */
.modal-xxl { max-width: 1200px; width: 96vw; }
.create-form-row { display: flex; gap: var(--space-4); margin-bottom: var(--space-4); }

.po-steps { display: flex; align-items: center; gap: var(--space-3); margin-bottom: var(--space-5); padding: var(--space-3) var(--space-4); background: var(--color-neutral-50); border-radius: var(--radius-lg); }
.po-step { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); color: var(--color-neutral-400); font-weight: 500; }
.po-step--active { color: var(--color-primary); font-weight: 600; }
.po-step-num { width: 24px; height: 24px; display: flex; align-items: center; justify-content: center; border-radius: var(--radius-full); background: var(--color-neutral-200); font-size: 0.6875rem; font-weight: 700; }
.po-step--active .po-step-num { background: var(--color-primary); color: white; }
.po-step--done { color: var(--color-success); }
.po-step--done .po-step-num { background: var(--color-success); color: white; }
.po-step-divider { flex: 1; height: 1px; background: var(--color-neutral-300); }

/* Supplier Selection */
/* Source step */
.po-source-tabs { display: flex; gap: 0; margin-bottom: var(--space-5); border-bottom: 2px solid var(--color-neutral-200); }
.po-source-tab { padding: var(--space-3) var(--space-5); font-size: 0.82rem; font-weight: 500; color: var(--color-neutral-500); border-bottom: 3px solid transparent; margin-bottom: -2px; cursor: pointer; background: none; border-top: none; border-left: none; border-right: none; transition: all 120ms; }
.po-source-tab:hover { color: var(--color-neutral-700); background: var(--color-neutral-50); }
.po-source-tab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); font-weight: 700; }

.po-source-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-3); }
.po-source-card { display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); cursor: pointer; text-align: left; transition: all .15s; }
.po-source-card:hover { border-color: var(--color-primary); box-shadow: 0 0 0 2px var(--color-primary-50); }
.po-source-card-top { display: flex; align-items: center; gap: var(--space-2); }
.po-source-card-customer { font-size: 0.85rem; color: var(--color-neutral-800); }
.po-source-card-meta { display: flex; align-items: center; justify-content: space-between; font-size: 0.72rem; color: var(--color-neutral-500); }
.po-source-card-desc { font-size: 0.68rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 100%; }

/* Source card item list */
.po-source-card-items { display: flex; flex-direction: column; gap: 2px; margin-top: 4px; border-top: 1px solid var(--color-neutral-100); padding-top: 4px; }
.po-src-inline-item { display: flex; align-items: center; gap: 4px; font-size: 0.625rem; color: var(--color-neutral-600); }
.po-src-inline-sku { font-family: var(--font-mono); color: var(--color-primary-600); font-weight: 600; min-width: 60px; }
.po-src-inline-desc { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.po-src-inline-qty { font-weight: 600; color: var(--color-neutral-500); white-space: nowrap; }
.po-src-inline-more { font-size: 0.6rem; color: var(--color-neutral-400); font-style: italic; }
[data-theme="dark"] .po-source-card-items { border-top-color: rgba(255,255,255,0.08); }
[data-theme="dark"] .po-src-inline-item { color: var(--color-neutral-300); }
[data-theme="dark"] .po-src-inline-sku { color: #93c5fd; }

.po-skip-source { display: flex; flex-direction: column; align-items: center; gap: var(--space-3); padding: var(--space-8); }

.po-selected-source-bar { display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-3) var(--space-4); background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); margin-bottom: var(--space-4); font-size: 0.82rem; }
.po-source-bar-header { display: flex; align-items: center; gap: var(--space-3); }

/* Source items preview chips in step 2 */
.po-source-items-preview { display: flex; flex-wrap: wrap; gap: 4px; padding-top: var(--space-1); border-top: 1px solid var(--color-neutral-150, var(--color-neutral-200)); margin-top: var(--space-1); }
.po-source-item-chip { display: inline-flex; align-items: center; gap: 3px; padding: 2px 8px; background: var(--color-neutral-100); border-radius: var(--radius-full); font-size: 0.66rem; }
.po-src-sku { font-family: var(--font-mono); font-weight: 600; color: var(--color-primary-600); }
.po-src-name { color: var(--color-neutral-700); max-width: 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.po-src-qty { font-weight: 600; color: var(--color-neutral-500); }
[data-theme="dark"] .po-source-item-chip { background: rgba(255,255,255,0.06); }
[data-theme="dark"] .po-src-sku { color: #93c5fd; }
[data-theme="dark"] .po-src-name { color: var(--color-neutral-300); }
[data-theme="dark"] .po-source-items-preview { border-top-color: rgba(255,255,255,0.08); }

.po-match-banner { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-3) var(--space-4); border-radius: var(--radius-lg); font-size: 0.8rem; margin-bottom: var(--space-3); }
.po-match-banner--success { background: var(--color-success-50, #f0fdf4); border: 1px solid var(--color-success-200, #bbf7d0); color: var(--color-success-700, #15803d); }
.po-match-banner--warn { background: var(--color-warning-50, #fffbeb); border: 1px solid var(--color-warning-200, #fde68a); color: var(--color-warning-700, #a16207); flex-wrap: wrap; }
[data-theme="dark"] .po-match-banner--success { background: rgba(34,197,94,0.1); border-color: rgba(34,197,94,0.3); color: #4ade80; }
[data-theme="dark"] .po-match-banner--warn { background: rgba(234,179,8,0.1); border-color: rgba(234,179,8,0.3); color: #fbbf24; }
.po-unmatched-list { font-family: var(--font-mono); font-size: 0.72rem; margin-left: 4px; }

.po-row--matched { background: var(--color-success-50, #f0fdf4); }
[data-theme="dark"] .po-row--matched { background: rgba(34,197,94,0.05); }

.supplier-card--has-catalog { border-color: var(--color-success-300, #86efac); }
.supplier-card-catalog-badge { display: flex; align-items: center; gap: 4px; font-size: 0.62rem; color: var(--color-success-600, #16a34a); font-weight: 600; margin-top: 2px; }
.supplier-card-no-catalog { display: flex; align-items: center; gap: 4px; font-size: 0.62rem; color: var(--color-neutral-400); margin-top: 2px; }

/* Supplier card match state */
.supplier-card--full-match { border-color: var(--color-success-300) !important; box-shadow: 0 0 0 1px var(--color-success-200), 0 2px 8px rgba(34,197,94,0.1) !important; }
.supplier-card--full-match:hover { border-color: var(--color-success-400) !important; }
.supplier-card--partial-match { border-color: var(--color-primary-200) !important; }
[data-theme="dark"] .supplier-card--full-match { border-color: rgba(34,197,94,0.4) !important; box-shadow: 0 0 0 1px rgba(34,197,94,0.2), 0 2px 8px rgba(34,197,94,0.06) !important; }
[data-theme="dark"] .supplier-card--partial-match { border-color: rgba(59,130,246,0.3) !important; }

/* Coverage bar */
.supplier-coverage { margin-top: 6px; }
.supplier-coverage-bar-track { width: 100%; height: 4px; background: var(--color-neutral-100); border-radius: 2px; overflow: hidden; }
.supplier-coverage-bar-fill { height: 100%; border-radius: 2px; transition: width 0.3s ease; }
.supplier-coverage-bar-fill.cov-full { background: var(--color-success-500); }
.supplier-coverage-bar-fill.cov-partial { background: var(--color-primary-500); }
.supplier-coverage-bar-fill.cov-none { background: var(--color-neutral-300); }
.supplier-coverage-text { display: block; font-size: 0.6rem; margin-top: 2px; color: var(--color-neutral-500); }
.supplier-coverage-text.cov-text-full { color: var(--color-success-600); font-weight: 600; }
[data-theme="dark"] .supplier-coverage-bar-track { background: rgba(255,255,255,0.06); }
[data-theme="dark"] .supplier-coverage-text { color: var(--color-neutral-400); }
[data-theme="dark"] .supplier-coverage-text.cov-text-full { color: #86efac; }

/* Matched item chips on supplier card */
.supplier-matched-items { display: flex; flex-wrap: wrap; gap: 2px; margin-top: 3px; }
.supplier-matched-chip { display: inline-flex; align-items: center; gap: 2px; padding: 1px 5px; background: var(--color-success-50); color: var(--color-success-700); border-radius: var(--radius-full); font-size: 0.55rem; font-weight: 500; }
.supplier-matched-more { font-size: 0.55rem; color: var(--color-neutral-400); padding: 1px 4px; }
[data-theme="dark"] .supplier-matched-chip { background: rgba(34,197,94,0.1); color: #86efac; }

.supplier-select-section { padding-top: var(--space-2); }
.supplier-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-3); margin-bottom: var(--space-5); }
.supplier-card {
  display: flex; flex-direction: column; gap: var(--space-2); padding: var(--space-4);
  background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg);
  cursor: pointer; text-align: left; transition: all .15s;
}
.supplier-card:hover { border-color: var(--color-primary); box-shadow: 0 0 0 2px var(--color-primary-50); }
.supplier-card-top { display: flex; align-items: center; gap: var(--space-2); }
.supplier-card-name { font-weight: 600; font-size: var(--text-sm); color: var(--color-neutral-900); }
.supplier-card-meta { display: flex; align-items: center; gap: var(--space-2); font-size: 0.6875rem; color: var(--color-neutral-500); }
.supplier-card-cats { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 2px; }
.supplier-card-cat { font-size: 0.5625rem; padding: 1px 6px; background: var(--color-neutral-100); border-radius: 9999px; color: var(--color-neutral-600); }

.supplier-manual-entry { padding: var(--space-4); background: var(--color-neutral-50); border-radius: var(--radius-lg); border: 1px dashed var(--color-neutral-300); }
.supplier-manual-row { display: flex; gap: var(--space-2); align-items: center; margin-top: var(--space-2); }

/* Selected supplier bar */
.selected-supplier-bar { display: flex; align-items: center; justify-content: space-between; padding: var(--space-3) var(--space-4); background: var(--color-primary-50); border: 1px solid var(--color-primary-100); border-radius: var(--radius-lg); margin-bottom: var(--space-4); }
.selected-supplier-info { display: flex; align-items: center; gap: var(--space-2); color: var(--color-primary); }

/* Supplier quote cards in PO creation */
.sq-available-section { margin-bottom: var(--space-5); }
.sq-available-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: 600; color: var(--color-neutral-700); margin-bottom: var(--space-3); }
.sq-available-list { display: flex; gap: var(--space-3); flex-wrap: wrap; }
.sq-available-card { display: flex; align-items: center; justify-content: space-between; gap: var(--space-4); padding: var(--space-3) var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); min-width: 280px; flex: 1; }
.sq-available-left { display: flex; flex-direction: column; gap: 2px; }

.po-item-search-row { display: flex; gap: var(--space-3); align-items: flex-start; margin-bottom: var(--space-4); }
.po-item-search-wrapper { flex: 1; position: relative; }

.po-item-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 50;
  background: var(--content-surface); border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg); box-shadow: var(--shadow-lg); max-height: 300px; overflow-y: auto;
}
.po-dd-item {
  display: flex; justify-content: space-between; align-items: center; width: 100%;
  padding: var(--space-3) var(--space-4); border: none; background: transparent;
  cursor: pointer; text-align: left; transition: background .1s; border-bottom: 1px solid var(--color-neutral-100);
}
.po-dd-item:last-child { border-bottom: none; }
.po-dd-item:hover { background: var(--color-primary-50); }
.po-dd-left { display: flex; flex-direction: column; gap: 1px; }
.po-dd-sku { font-family: var(--font-mono); font-size: var(--text-xs); color: var(--color-primary); font-weight: 600; }
.po-dd-name { font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-800); }
.po-dd-meta { font-size: 0.6875rem; color: var(--color-neutral-400); }
.po-dd-right { text-align: right; }
.po-dd-price { font-size: var(--text-sm); font-weight: 700; color: var(--color-neutral-800); font-family: var(--font-mono); display: block; }
.po-dd-lead { font-size: 0.625rem; color: var(--color-neutral-400); }

.po-items-table { max-height: 400px; overflow-y: auto; }

.onhand-badge { display: inline-block; padding: 1px 8px; border-radius: 9999px; font-size: 0.6875rem; font-weight: 600; }
.onhand-badge--has { background: rgba(16, 185, 129, .1); color: var(--color-success); }
.onhand-badge--zero { background: rgba(239, 68, 68, .1); color: var(--color-danger); }

.best-price-badge { font-size: 0.6875rem; font-family: var(--font-mono); padding: 1px 6px; border-radius: var(--radius-sm); font-weight: 600; }
.best-price--better { background: rgba(16, 185, 129, .1); color: var(--color-success); }
.best-price--same { background: var(--color-neutral-100); color: var(--color-neutral-600); }
.best-price--worse { background: rgba(239, 68, 68, .1); color: var(--color-danger); }

.create-empty { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); padding: var(--space-8); }
.form-input--sm { padding: 4px 8px; font-size: var(--text-sm); }
.form-input--center { text-align: center; }
.form-input--right { text-align: right; }

.modal-footer-right { display: flex; gap: var(--space-2); }
.po-review { padding-top: var(--space-2); }

@media (max-width: 1400px) { .kpi-grid--6 { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 1200px) { .kpi-grid--6 { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .kpi-grid--6 { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-select { width: 100%; }
  .create-form-row { flex-direction: column; }
  .view-info-grid { grid-template-columns: repeat(2, 1fr); }
  .proc-tabs { flex-wrap: wrap; }
  .po-item-search-row { flex-direction: column; }
  .page-header-actions { flex-direction: column; }
  .supplier-grid { grid-template-columns: 1fr; }
}

/* ── Create/Edit Supplier Quote Modal ─────────────────────── */
.sq-form-section {
  margin-bottom: var(--space-5);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-neutral-200);
}
.sq-form-section:last-of-type { border-bottom: none; }
.sq-form-section-title {
  display: flex; align-items: center; gap: var(--space-2);
  font-size: 0.85rem; font-weight: 700; color: var(--color-neutral-700);
  margin-bottom: var(--space-3);
}

.sq-supplier-pick { position: relative; }
.sq-supplier-chips {
  display: flex; flex-wrap: wrap; gap: var(--space-1);
  margin-top: var(--space-2);
}
.sq-supplier-chip {
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-sm);
  padding: 2px 10px; font-size: 0.72rem; cursor: pointer;
  color: var(--color-neutral-600); transition: all 120ms;
}
.sq-supplier-chip:hover {
  background: var(--color-primary-50);
  border-color: var(--color-primary-400);
  color: var(--color-primary-700);
}

/* Upload area */
.sq-upload-area { position: relative; }
.sq-file-input { position: absolute; width: 0; height: 0; opacity: 0; pointer-events: none; }
.sq-upload-box {
  display: flex; align-items: center; gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border: 2px dashed var(--color-neutral-300);
  border-radius: var(--radius-lg);
  cursor: pointer; transition: all 150ms;
  background: var(--content-surface);
}
.sq-upload-box:hover {
  border-color: var(--color-primary-400);
  background: var(--color-primary-50);
}
.sq-upload-box svg { color: var(--color-neutral-400); }
.sq-upload-box:hover svg { color: var(--color-primary-500); }
.sq-upload-inner { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); text-align: center; width: 100%; }
.sq-upload-text { display: flex; flex-direction: column; gap: 2px; }
.sq-upload-title { font-size: 0.82rem; font-weight: 600; color: var(--color-neutral-700); }
.sq-upload-hint { font-size: 0.7rem; color: var(--color-neutral-500); }
.sq-upload-types { display: flex; gap: var(--space-2); margin-top: var(--space-1); }
.sq-upload-type-tag {
  display: inline-flex; align-items: center; gap: 3px;
  font-size: 0.62rem; font-weight: 500;
  padding: 2px 8px;
  background: var(--color-neutral-100);
  border-radius: var(--radius-full);
  color: var(--color-neutral-500);
}
.sq-upload-selected { display: flex; align-items: center; gap: var(--space-2); width: 100%; }
[data-theme="dark"] .sq-upload-box:hover { background: rgba(59,130,246,0.08); }

.sq-pdf-paste { margin-top: var(--space-3); }
.sq-pdf-notice {
  display: flex; align-items: flex-start; gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--color-primary-50); color: var(--color-primary-700);
  border-radius: var(--radius-sm); font-size: 0.75rem; margin-bottom: var(--space-2);
}
[data-theme="dark"] .sq-pdf-notice { background: rgba(59,130,246,0.1); color: #93c5fd; }
.sq-paste-area {
  width: 100%; font-family: var(--font-mono, monospace);
  font-size: 0.75rem; line-height: 1.5; resize: vertical; min-height: 90px;
}

.sq-parse-error {
  display: flex; align-items: center; gap: var(--space-2);
  margin-top: var(--space-3); padding: var(--space-3) var(--space-4);
  background: var(--color-danger-50, #fef2f2); border: 1px solid var(--color-danger-200, #fecaca);
  border-radius: var(--radius-md); font-size: 0.78rem; color: var(--color-danger-600, #dc2626);
}

/* CSV Preview */
.sq-csv-preview {
  margin-top: var(--space-3);
  border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-md);
  overflow: hidden; background: var(--content-surface);
}
.sq-csv-preview-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--color-success-50, #f0fdf4);
  border-bottom: 1px solid var(--color-neutral-200);
  font-size: 0.78rem;
}
.sq-csv-preview-header .font-semibold {
  display: flex; align-items: center; gap: var(--space-2);
  color: var(--color-success-700, #15803d);
}
.sq-csv-preview-actions { display: flex; gap: var(--space-2); }
.sq-csv-preview-table { max-height: 300px; overflow-y: auto; }
.sq-csv-preview-table .table { margin-bottom: 0; }

/* SQ Items section */
.sq-items-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: var(--space-3); flex-wrap: wrap; gap: var(--space-2);
}
.sq-items-header-actions { display: flex; align-items: center; gap: var(--space-2); }
.sq-item-search-wrap { position: relative; }
.sq-item-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 30;
  background: var(--content-surface); border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-md); max-height: 240px; overflow-y: auto;
  box-shadow: var(--shadow-lg);
}
.sq-items-table { max-height: 340px; overflow-y: auto; }
.sq-items-table .table { margin-bottom: 0; }
.form-input--sm { padding: 4px 8px; font-size: 0.75rem; }
.form-input--center { text-align: center; }
.form-input--right { text-align: right; }
.create-empty {
  display: flex; flex-direction: column; align-items: center;
  gap: var(--space-2); padding: var(--space-7) 0;
}
.create-empty p { font-size: 0.78rem; }

/* ─── PO File Upload ──────────────────────────────────── */
.po-file-upload-section {
  margin-bottom: var(--space-4);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  background: var(--content-surface);
  overflow: hidden;
}
.po-file-upload-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--color-neutral-50);
  border-bottom: 1px solid var(--color-neutral-200);
}
[data-theme="dark"] .po-file-upload-header { background: var(--color-neutral-800, #1e293b); }
.po-file-title { font-size: 0.82rem; font-weight: 600; display: flex; align-items: center; gap: var(--space-2); margin: 0; color: var(--color-neutral-800); }
.po-file-upload-zone {
  margin: var(--space-3) var(--space-4);
  border: 2px dashed var(--color-neutral-300);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}
.po-file-upload-zone:hover { border-color: var(--color-primary-500); background: var(--color-primary-50); }
[data-theme="dark"] .po-file-upload-zone:hover { background: rgba(59,130,246,0.08); }
.po-file-placeholder { display: flex; flex-direction: column; align-items: center; gap: var(--space-1); text-align: center; }
.po-file-placeholder-text { font-size: 0.82rem; font-weight: 500; color: var(--color-neutral-600); }
.po-file-types { display: flex; gap: var(--space-2); margin-top: var(--space-1); }
.po-file-type-tag {
  display: inline-flex; align-items: center; gap: 3px;
  font-size: 0.62rem; font-weight: 500;
  padding: 2px 8px;
  background: var(--color-neutral-100);
  border-radius: var(--radius-full);
  color: var(--color-neutral-500);
}
.po-file-selected { display: flex; align-items: center; gap: var(--space-2); }
.po-file-error {
  display: flex; align-items: center; gap: var(--space-2);
  margin: 0 var(--space-4) var(--space-3);
  padding: var(--space-2) var(--space-3);
  background: #fef3c7; color: #92400e;
  border-radius: var(--radius-sm); font-size: 0.78rem;
}
[data-theme="dark"] .po-file-error { background: rgba(251,191,36,0.15); color: #fbbf24; }
.po-file-pdf-paste { padding: 0 var(--space-4) var(--space-3); }
.po-file-pdf-notice {
  display: flex; align-items: flex-start; gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--color-primary-50); color: var(--color-primary-700);
  border-radius: var(--radius-sm); font-size: 0.75rem; margin-bottom: var(--space-2);
}
[data-theme="dark"] .po-file-pdf-notice { background: rgba(59,130,246,0.1); color: #93c5fd; }
.po-file-paste-area {
  width: 100%; font-family: var(--font-mono, monospace);
  font-size: 0.75rem; line-height: 1.5; resize: vertical; min-height: 80px;
}
.po-file-preview { padding: 0 var(--space-4) var(--space-3); }
.po-file-preview-header { display: flex; align-items: center; gap: var(--space-2); margin-bottom: var(--space-2); }
.po-file-preview-footer { display: flex; justify-content: flex-end; margin-top: var(--space-3); }
.po-file-row--off { opacity: 0.4; }
.po-file-parsing, .sq-parsing-indicator {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4); margin: var(--space-3) var(--space-4) 0;
  font-size: 0.82rem; font-weight: 500; color: var(--color-primary-600);
  background: var(--color-primary-50); border-radius: var(--radius-md);
}
.sq-parsing-indicator { margin: var(--space-3) 0 0; }
[data-theme="dark"] .po-file-parsing,
[data-theme="dark"] .sq-parsing-indicator { background: rgba(59,130,246,0.1); color: #93c5fd; }
@keyframes spin { to { transform: rotate(360deg); } }
.spin-icon { animation: spin 1s linear infinite; }
.btn-download-template {
  display: inline-flex; align-items: center; gap: 4px;
  margin-top: var(--space-2);
  padding: 4px 12px;
  font-size: 0.72rem; font-weight: 500;
  color: var(--color-primary-600);
  background: none; border: 1px dashed var(--color-primary-300);
  border-radius: var(--radius-full);
  cursor: pointer; transition: all 0.15s;
}
.btn-download-template:hover {
  background: var(--color-primary-50);
  border-color: var(--color-primary-500);
  color: var(--color-primary-700);
}
[data-theme="dark"] .btn-download-template { color: #93c5fd; border-color: rgba(147,197,253,0.3); }
[data-theme="dark"] .btn-download-template:hover { background: rgba(59,130,246,0.1); border-color: #60a5fa; }
.fallback-template-link { color: var(--color-primary-600); font-weight: 600; text-decoration: underline; cursor: pointer; }
.fallback-template-link:hover { color: var(--color-primary-700); }
[data-theme="dark"] .fallback-template-link { color: #93c5fd; }
[data-theme="dark"] .fallback-template-link:hover { color: #bfdbfe; }
</style>
