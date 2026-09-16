<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Save, Send, CheckCircle2, Printer, FileSpreadsheet, Search, X, Trash2,
  Plus, GripVertical, Package, Users, Wrench, ChevronDown, Calendar,
  Building2, AlertCircle, CircleDot, ArrowLeft, RefreshCw, Eye, EyeOff,
  MessageSquare, Minus, Hash, Zap, FileText, Settings, Copy, MapPin, Clipboard,
  ShoppingCart, Truck, AlertTriangle, PackageCheck, History, DollarSign, TrendingDown, TrendingUp, Check,
} from 'lucide-vue-next'
import type { QuoteLineCategory, QuoteStatus, Currency } from '@/types'
import { useProcurementStore } from '@/stores/procurement'

const router = useRouter()
const route = useRoute()

function uid(): string { return Math.random().toString(36).slice(2, 11) }

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function marginClass(pct: number): string {
  if (pct >= 25) return 'margin-high'
  if (pct >= 20) return 'margin-medium'
  return 'margin-low'
}

// ── Enums & Config ───────────────────────────────────────────
type ItemSource = 'product' | 'service' | 'recurring' | 'write-in'
type RowType = 'item' | 'heading' | 'comment' | 'subtotal' | 'discount'

const sourceLabels: Record<ItemSource, string> = { product: 'Product', service: 'Service', recurring: 'Recurring', 'write-in': 'Write-in' }
const sourceIcons = { product: Package, service: Users, recurring: RefreshCw, 'write-in': FileText } as const

const statusConfig: Record<QuoteStatus, { label: string; badge: string }> = {
  draft: { label: 'Draft', badge: 'badge-gray' },
  'pending-approval': { label: 'Pending Approval', badge: 'badge-warning' },
  approved: { label: 'Approved', badge: 'badge-info' },
  sent: { label: 'Sent', badge: 'badge-primary' },
  accepted: { label: 'Accepted', badge: 'badge-success' },
  declined: { label: 'Declined', badge: 'badge-danger' },
  expired: { label: 'Expired', badge: 'badge-gray' },
}

// ── Price History / Batches ───────────────────────────────────
interface PriceEntry {
  id: string
  source: string
  supplier: string
  batchRef?: string
  date: string
  cost: number
  price: number
  qty: number
  modifier?: string
  customer?: string
  quoteRef?: string
}

// ── Catalog Data (Products, Services, Recurring) ─────────────
interface CatalogProduct {
  id: string; sku: string; name: string; manufacturer: string
  unitCost: number; unitPrice: number; stockAvailable: number; leadTimeDays: number
  priceHistory: PriceEntry[]
}

interface CatalogService {
  id: string; sku: string; name: string; department: string
  rateType: string; unitCost: number; unitPrice: number
}

interface CatalogRecurring {
  id: string; sku: string; name: string; billingCycle: string
  monthlyCost: number; monthlyPrice: number
}

function ph(source: string, supplier: string, date: string, cost: number, price: number, qty: number, batchRef?: string, modifier?: string, customer?: string, quoteRef?: string): PriceEntry {
  return { id: uid(), source, supplier, date, cost, price, qty, batchRef, modifier, customer, quoteRef }
}

const productCatalog: CatalogProduct[] = [
  { id: 'cp1', sku: 'HIK-DS2CD2143', name: 'DS-2CD2143G2-IU 4MP Dome', manufacturer: 'Hikvision', unitCost: 367, unitPrice: 522, stockAvailable: 60, leadTimeDays: 21, priceHistory: [
    ph('Distributor', 'Hikvision Saudi', '2026-01-15', 367, 522, 100, 'B-2601', 'Standard'),
    ph('Distributor', 'Hikvision Saudi', '2025-09-20', 385, 548, 50, 'B-2509', 'Standard', 'Saudi Aramco', 'QT-2025-0098'),
    ph('Direct Import', 'Hikvision HQ', '2025-06-10', 342, 522, 200, 'IMP-2506', '-7% volume'),
    ph('Past Quote', 'Hikvision Saudi', '2025-11-05', 367, 490, 120, undefined, 'Aramco discount', 'Saudi Aramco', 'QT-2025-0134'),
  ]},
  { id: 'cp2', sku: 'HIK-DS2CD2T87', name: 'DS-2CD2T87G2-L 8MP Bullet', manufacturer: 'Hikvision', unitCost: 862, unitPrice: 1199, stockAvailable: 45, leadTimeDays: 21, priceHistory: [
    ph('Distributor', 'Hikvision Saudi', '2026-02-01', 862, 1199, 64, 'B-2602', 'Standard'),
    ph('Distributor', 'Hikvision Saudi', '2025-10-15', 890, 1240, 32, 'B-2510', 'Standard', 'SABIC', 'QT-2025-0112'),
    ph('Direct Import', 'Hikvision HQ', '2025-07-20', 810, 1199, 100, 'IMP-2507', '-6% volume'),
  ]},
  { id: 'cp3', sku: 'HIK-DS7732NI', name: 'DS-7732NI-K4 32CH NVR', manufacturer: 'Hikvision', unitCost: 1820, unitPrice: 2436, stockAvailable: 12, leadTimeDays: 28, priceHistory: [
    ph('Distributor', 'Hikvision Saudi', '2026-01-10', 1820, 2436, 10, 'B-2601'),
    ph('Distributor', 'Al-Jazirah Tech', '2025-12-01', 1880, 2500, 8, 'AJT-2512'),
    ph('Past Quote', 'Hikvision Saudi', '2025-08-15', 1780, 2380, 8, undefined, '-2.5% project', 'Saudi Aramco', 'QT-2025-0090'),
  ]},
  { id: 'cp4', sku: 'DH-IPC-HFW5442', name: 'IPC-HFW5442T-ASE 4MP AI Bullet', manufacturer: 'Dahua', unitCost: 474, unitPrice: 676, stockAvailable: 35, leadTimeDays: 25, priceHistory: [
    ph('Distributor', 'Dahua MEA', '2026-01-20', 474, 676, 36, 'B-2601'),
    ph('Distributor', 'SecureTech KSA', '2025-11-10', 490, 700, 20, 'ST-2511'),
  ]},
  { id: 'cp5', sku: 'DH-NVR5432-EI', name: 'NVR5432-EI 32CH AI NVR', manufacturer: 'Dahua', unitCost: 3720, unitPrice: 4960, stockAvailable: 0, leadTimeDays: 30, priceHistory: [
    ph('Distributor', 'Dahua MEA', '2025-12-05', 3720, 4960, 4, 'B-2512'),
    ph('Direct Import', 'Dahua HQ', '2025-08-20', 3500, 4960, 10, 'IMP-2508', '-6% direct'),
  ]},
  { id: 'cp6', sku: 'DH-ASI7214Y', name: 'ASI7214Y Face Recognition Terminal', manufacturer: 'Dahua', unitCost: 1440, unitPrice: 2118, stockAvailable: 0, leadTimeDays: 30, priceHistory: [
    ph('Distributor', 'Dahua MEA', '2025-11-15', 1440, 2118, 10, 'B-2511'),
  ]},
  { id: 'cp7', sku: 'AXIS-P3265LVE', name: 'P3265-LVE 2MP Dome', manufacturer: 'Axis', unitCost: 1720, unitPrice: 2656, stockAvailable: 20, leadTimeDays: 35, priceHistory: [
    ph('Distributor', 'Axis Partner KSA', '2026-01-05', 1720, 2656, 20, 'AX-2601'),
    ph('Distributor', 'Axis Partner KSA', '2025-07-10', 1680, 2600, 48, 'AX-2507', undefined, 'Ministry of Interior', 'QT-2025-0078'),
  ]},
  { id: 'cp8', sku: 'AXIS-Q6135LE', name: 'Q6135-LE PTZ Camera', manufacturer: 'Axis', unitCost: 19044, unitPrice: 24600, stockAvailable: 0, leadTimeDays: 42, priceHistory: [
    ph('Distributor', 'Axis Partner KSA', '2025-12-20', 19044, 24600, 8, 'AX-2512'),
    ph('Direct Import', 'Axis Sweden', '2025-05-15', 18200, 24600, 16, 'IMP-2505', '-4.5% direct'),
    ph('Past Quote', 'Axis Partner KSA', '2025-09-01', 19044, 23500, 32, undefined, 'Volume deal', 'NEOM', 'QT-2025-0105'),
  ]},
  { id: 'cp9', sku: 'HON-MNPDS2', name: 'Morley-IAS Fire Panel 2L', manufacturer: 'Honeywell', unitCost: 5438, unitPrice: 6510, stockAvailable: 5, leadTimeDays: 45, priceHistory: [
    ph('Distributor', 'Honeywell MEA', '2026-01-25', 5438, 6510, 8, 'HW-2601'),
    ph('Distributor', 'Gulf Security Dist.', '2025-10-01', 5600, 6800, 4, 'GSD-2510'),
  ]},
  { id: 'cp10', sku: 'HON-MAXPRO', name: 'MAXPRO Access 4-Door Controller', manufacturer: 'Honeywell', unitCost: 2100, unitPrice: 2800, stockAvailable: 8, leadTimeDays: 14, priceHistory: [
    ph('Distributor', 'Honeywell MEA', '2026-02-01', 2100, 2800, 20, 'HW-2602'),
    ph('Past Quote', 'Honeywell MEA', '2025-12-10', 2100, 2650, 12, undefined, '-5.4% SABIC', 'SABIC', 'QT-2025-0140'),
  ]},
  { id: 'cp11', sku: 'BOSCH-NDV3503', name: 'FLEXIDOME IP 3000i 5MP', manufacturer: 'Bosch', unitCost: 1252, unitPrice: 1660, stockAvailable: 18, leadTimeDays: 28, priceHistory: [
    ph('Distributor', 'Bosch KSA', '2026-01-18', 1252, 1660, 18, 'BSH-2601'),
  ]},
  { id: 'cp12', sku: 'BOSCH-FPA5000', name: 'FPA-5000 Fire Panel', manufacturer: 'Bosch', unitCost: 8424, unitPrice: 10694, stockAvailable: 3, leadTimeDays: 56, priceHistory: [
    ph('Distributor', 'Bosch KSA', '2025-11-20', 8424, 10694, 4, 'BSH-2511'),
    ph('Distributor', 'Gulf Security Dist.', '2025-08-05', 8600, 11000, 2, 'GSD-2508'),
  ]},
  { id: 'cp13', sku: 'ZKT-INBIO460', name: 'InBio460 4-Door Controller', manufacturer: 'ZKTeco', unitCost: 858, unitPrice: 1320, stockAvailable: 15, leadTimeDays: 35, priceHistory: [
    ph('Distributor', 'ZKTeco Gulf', '2026-01-08', 858, 1320, 15, 'ZK-2601'),
    ph('Distributor', 'ZKTeco Gulf', '2025-06-15', 820, 1280, 30, 'ZK-2506', '-5% promo'),
  ]},
  { id: 'cp14', sku: 'ZKT-SPEEDFACE', name: 'SpeedFace-V5L Facial Terminal', manufacturer: 'ZKTeco', unitCost: 1508, unitPrice: 2917, stockAvailable: 0, leadTimeDays: 30, priceHistory: [
    ph('Distributor', 'ZKTeco Gulf', '2025-12-01', 1508, 2917, 24, 'ZK-2512'),
    ph('Past Quote', 'ZKTeco Gulf', '2025-10-20', 1508, 2750, 24, undefined, '-5.7%', 'SABIC', 'QT-2025-0125'),
  ]},
  { id: 'cp15', sku: 'CBL-CAT6A-305', name: 'Cat6A UTP Cable 305m Box', manufacturer: 'Belden', unitCost: 420, unitPrice: 580, stockAvailable: 30, leadTimeDays: 14, priceHistory: [
    ph('Distributor', 'Belden MEA', '2026-02-05', 420, 580, 30, 'BLD-2602'),
    ph('Local Vendor', 'Al-Salam Cables', '2025-11-01', 440, 600, 50, 'ASC-2511'),
    ph('Local Vendor', 'Al-Salam Cables', '2025-07-01', 400, 560, 100, 'ASC-2507', '-7% bulk'),
  ]},
  { id: 'cp16', sku: 'CBL-FIBER-OM3', name: 'OM3 Fiber Optic Cable 1000m', manufacturer: 'Corning', unitCost: 1850, unitPrice: 2450, stockAvailable: 6, leadTimeDays: 21, priceHistory: [
    ph('Distributor', 'Corning Gulf', '2026-01-12', 1850, 2450, 6, 'CRN-2601'),
    ph('Direct Import', 'Corning US', '2025-09-10', 1750, 2450, 12, 'IMP-2509', '-5.4% direct'),
  ]},
]

const serviceCatalog: CatalogService[] = [
  { id: 'cs1', sku: 'SVC-INSTALL-SR', name: 'Senior Installation Engineer', department: 'Technical', rateType: 'per day', unitCost: 800, unitPrice: 1200 },
  { id: 'cs2', sku: 'SVC-INSTALL-JR', name: 'Technician', department: 'Technical', rateType: 'per day', unitCost: 450, unitPrice: 700 },
  { id: 'cs3', sku: 'SVC-PM-MONTH', name: 'Project Manager', department: 'Management', rateType: 'per month', unitCost: 12000, unitPrice: 18000 },
  { id: 'cs4', sku: 'SVC-AC-SPEC', name: 'Access Control Specialist', department: 'Technical', rateType: 'per day', unitCost: 900, unitPrice: 1350 },
  { id: 'cs5', sku: 'SVC-DESIGN', name: 'System Design & Engineering', department: 'Pre-Sales', rateType: 'fixed', unitCost: 5500, unitPrice: 8500 },
  { id: 'cs6', sku: 'SVC-COMMISSION', name: 'System Commissioning', department: 'Technical', rateType: 'per day', unitCost: 1000, unitPrice: 1500 },
  { id: 'cs7', sku: 'SVC-TRAINING', name: 'End-User Training', department: 'Support', rateType: 'per session', unitCost: 600, unitPrice: 950 },
]

const recurringCatalog: CatalogRecurring[] = [
  { id: 'cr1', sku: 'REC-GUARD-24', name: 'Security Guard Service (24/7)', billingCycle: 'Monthly', monthlyCost: 8000, monthlyPrice: 12000 },
  { id: 'cr2', sku: 'REC-MAINT-STD', name: 'System Maintenance', billingCycle: 'Monthly', monthlyCost: 2200, monthlyPrice: 3500 },
  { id: 'cr3', sku: 'REC-MON-247', name: '24/7 Remote Monitoring', billingCycle: 'Monthly', monthlyCost: 3200, monthlyPrice: 5000 },
  { id: 'cr4', sku: 'REC-PATROL', name: 'Mobile Patrol Service', billingCycle: 'Monthly', monthlyCost: 5500, monthlyPrice: 8000 },
  { id: 'cr5', sku: 'REC-FM', name: 'Facility Management', billingCycle: 'Monthly', monthlyCost: 9500, monthlyPrice: 15000 },
  { id: 'cr6', sku: 'REC-ALARM', name: 'Alarm Response Service', billingCycle: 'Monthly', monthlyCost: 1500, monthlyPrice: 2500 },
]

// ── Row Interface ────────────────────────────────────────────
interface QuoteRow {
  id: string
  rowType: RowType
  source: ItemSource
  productId?: string
  sku: string
  description: string
  manufacturer?: string
  stockAvailable?: number
  leadTimeDays?: number
  rateType?: string
  billingCycle?: string
  quantity: number
  multiplier: number
  unitCost: number
  unitPrice: number
  discountPercent: number
  lineTotal: number
  marginPercent: number
  isOptional: boolean
  isSelected: boolean
  isPrintable: boolean
  headingText?: string
  commentText?: string
}

// ── Quote State ──────────────────────────────────────────────
const quoteNumber = ref('QT-2026-0148')
const customerName = ref('Saudi Aramco')
const customerId = ref('c1')
const quoteStatus = ref<QuoteStatus>('draft')
const validUntil = ref('2026-04-15')
const currency = ref<Currency>('SAR')
const notes = ref('')
const discountPercent = ref(0)
const paymentTerms = ref('Net 30')
const deliveryTerms = ref('Ex-Works')

// ── Builder Tabs ─────────────────────────────────────────────
type BuilderTab = 'items' | 'addresses' | 'notes'
const builderTab = ref<BuilderTab>('items')

// ── Sold To / Ship To ────────────────────────────────────────
const soldTo = ref({
  contactName: 'Mohammed Al-Qahtani',
  company: 'Saudi Aramco',
  address: 'P.O. Box 5000, Dhahran 31311',
  city: 'Dhahran',
  country: 'Saudi Arabia',
  phone: '+966 13 872 0000',
  email: 'm.qahtani@aramco.com',
})

const shipTo = ref({
  contactName: '',
  company: '',
  address: '',
  city: '',
  country: 'Saudi Arabia',
  phone: '',
  email: '',
})

function copySoldToShipTo() {
  shipTo.value = { ...soldTo.value }
}

// ── Notes Fields ─────────────────────────────────────────────
const introductionText = ref('We are pleased to submit the following quotation for your review and consideration.')
const closingText = ref('This quotation is valid for the period indicated above. We look forward to your favorable response.')
const internalNotes = ref('')
const purchasingNotes = ref('')
const statementOfWork = ref('')

// ── Spotlight Search ─────────────────────────────────────────
const spotlightQuery = ref('')
const spotlightActive = computed(() => spotlightQuery.value.trim().length > 0)

function isSpotlightMatch(row: QuoteRow): boolean {
  if (!spotlightActive.value) return true
  const q = spotlightQuery.value.toLowerCase().trim()
  return (
    row.description.toLowerCase().includes(q) ||
    row.sku.toLowerCase().includes(q) ||
    (row.manufacturer?.toLowerCase().includes(q) ?? false) ||
    (row.headingText?.toLowerCase().includes(q) ?? false) ||
    (row.commentText?.toLowerCase().includes(q) ?? false)
  )
}

// ── Procurement ──────────────────────────────────────────────
const procurementStore = useProcurementStore()

const outOfStockRows = computed(() =>
  itemRows.value.filter(r =>
    r.source === 'product' &&
    (r.stockAvailable == null || r.stockAvailable === 0 || (r.stockAvailable != null && r.stockAvailable < r.quantity)),
  ),
)

const stockSufficient = computed(() =>
  itemRows.value.filter(r =>
    r.source === 'product' && r.stockAvailable != null && r.stockAvailable >= r.quantity,
  ),
)

function stockStatus(row: QuoteRow): 'in-stock' | 'low-stock' | 'out-of-stock' | 'na' {
  if (row.source !== 'product') return 'na'
  if (row.stockAvailable == null) return 'out-of-stock'
  if (row.stockAvailable === 0) return 'out-of-stock'
  if (row.stockAvailable < row.quantity) return 'low-stock'
  return 'in-stock'
}

function stockLabel(row: QuoteRow): string {
  const status = stockStatus(row)
  if (status === 'in-stock') return `${row.stockAvailable} in stock`
  if (status === 'low-stock') return `${row.stockAvailable} avail (need ${row.quantity})`
  if (status === 'out-of-stock' && row.leadTimeDays) return `${row.leadTimeDays}d lead time`
  if (status === 'out-of-stock') return 'Out of stock'
  return ''
}

const showPOModal = ref(false)

function generatePOFromQuote() {
  if (outOfStockRows.value.length === 0) return

  const supplierGroups: Record<string, typeof outOfStockRows.value> = {}
  for (const row of outOfStockRows.value) {
    const supplier = row.manufacturer || 'Unknown Supplier'
    if (!supplierGroups[supplier]) supplierGroups[supplier] = []
    supplierGroups[supplier].push(row)
  }

  for (const [supplier, items] of Object.entries(supplierGroups)) {
    const poItems = items.map(item => {
      const needQty = item.stockAvailable != null ? Math.max(0, item.quantity - item.stockAvailable) : item.quantity
      return {
        id: uid(),
        productId: item.productId || uid(),
        productSku: item.sku,
        productName: item.description,
        manufacturerName: item.manufacturer || '',
        quantity: needQty,
        unitCost: item.unitCost,
        total: needQty * item.unitCost,
        receivedQty: 0,
        leadTimeDays: item.leadTimeDays || 30,
      }
    })
    const subtotal = poItems.reduce((s, i) => s + i.total, 0)
    const shippingCost = Math.round(subtotal * 0.04)
    const customsDuty = Math.round(subtotal * 0.05)
    procurementStore.addPurchaseOrder({
      id: uid(),
      poNumber: procurementStore.generatePoNumber(),
      supplierName: supplier,
      status: 'draft',
      items: poItems,
      subtotal,
      shippingCost,
      customsDuty,
      total: subtotal + shippingCost + customsDuty,
      currency: currency.value,
      expectedDelivery: new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10),
      sourceQuoteId: 'current',
      sourceQuoteNumber: quoteNumber.value,
      notes: `Auto-generated from ${quoteNumber.value} for out-of-stock items.`,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    })
  }
  showSaveToast(`${Object.keys(supplierGroups).length} PO(s) created for ${outOfStockRows.value.length} items`)
  showPOModal.value = false
}

// ── Price History Modal ──────────────────────────────────────
const showPriceHistoryModal = ref(false)
const priceHistoryProductId = ref('')
const priceHistoryTargetRowId = ref<string | null>(null)
const priceHistorySortField = ref<'date' | 'cost' | 'price' | 'qty' | 'supplier'>('date')
const priceHistorySortAsc = ref(false)
const priceHistoryFilterSource = ref('')

const priceHistoryProduct = computed(() =>
  productCatalog.find(p => p.id === priceHistoryProductId.value),
)

const priceHistoryEntries = computed(() => {
  const product = priceHistoryProduct.value
  if (!product) return []
  let entries = [...product.priceHistory]

  if (priceHistoryFilterSource.value) {
    entries = entries.filter(e => e.source === priceHistoryFilterSource.value)
  }

  entries.sort((a, b) => {
    const dir = priceHistorySortAsc.value ? 1 : -1
    switch (priceHistorySortField.value) {
      case 'date': return dir * a.date.localeCompare(b.date)
      case 'cost': return dir * (a.cost - b.cost)
      case 'price': return dir * (a.price - b.price)
      case 'qty': return dir * (a.qty - b.qty)
      case 'supplier': return dir * a.supplier.localeCompare(b.supplier)
      default: return 0
    }
  })
  return entries
})

const priceHistorySources = computed(() => {
  const product = priceHistoryProduct.value
  if (!product) return []
  return [...new Set(product.priceHistory.map(e => e.source))]
})

const priceHistoryAvgCost = computed(() => {
  const entries = priceHistoryEntries.value
  if (!entries.length) return 0
  return entries.reduce((s, e) => s + e.cost, 0) / entries.length
})

const priceHistoryAvgPrice = computed(() => {
  const entries = priceHistoryEntries.value
  if (!entries.length) return 0
  return entries.reduce((s, e) => s + e.price, 0) / entries.length
})

const priceHistoryLowest = computed(() => {
  const entries = priceHistoryEntries.value
  if (!entries.length) return { cost: 0, price: 0 }
  return {
    cost: Math.min(...entries.map(e => e.cost)),
    price: Math.min(...entries.map(e => e.price)),
  }
})

const priceHistoryHighest = computed(() => {
  const entries = priceHistoryEntries.value
  if (!entries.length) return { cost: 0, price: 0 }
  return {
    cost: Math.max(...entries.map(e => e.cost)),
    price: Math.max(...entries.map(e => e.price)),
  }
})

function togglePriceHistorySort(field: typeof priceHistorySortField.value) {
  if (priceHistorySortField.value === field) {
    priceHistorySortAsc.value = !priceHistorySortAsc.value
  } else {
    priceHistorySortField.value = field
    priceHistorySortAsc.value = false
  }
}

function openPriceHistory(productId: string, targetRowId?: string) {
  priceHistoryProductId.value = productId
  priceHistoryTargetRowId.value = targetRowId ?? null
  priceHistoryFilterSource.value = ''
  priceHistorySortField.value = 'date'
  priceHistorySortAsc.value = false
  showPriceHistoryModal.value = true
}

function applyPriceFromHistory(entry: PriceEntry) {
  if (priceHistoryTargetRowId.value) {
    const row = rows.value.find(r => r.id === priceHistoryTargetRowId.value)
    if (row) {
      row.unitCost = entry.cost
      row.unitPrice = entry.price
      recalcRow(row)
      showSaveToast(`Applied ${entry.supplier} pricing (${entry.date})`)
    }
  } else {
    const product = priceHistoryProduct.value
    if (product) {
      rows.value.push(makeProductRow(
        product.id, product.sku, product.name, product.manufacturer,
        product.stockAvailable, product.stockAvailable === 0 ? product.leadTimeDays : undefined,
        1, entry.cost, entry.price,
      ))
      showSaveToast(`Added ${product.sku} with ${entry.supplier} pricing`)
    }
  }
  showPriceHistoryModal.value = false
}

function addWithDefaultPrice() {
  const product = priceHistoryProduct.value
  if (!product) return
  rows.value.push(makeProductRow(
    product.id, product.sku, product.name, product.manufacturer,
    product.stockAvailable, product.stockAvailable === 0 ? product.leadTimeDays : undefined,
    1, product.unitCost, product.unitPrice,
  ))
  showSaveToast(`Added ${product.sku} with default pricing`)
  showPriceHistoryModal.value = false
}

// ── Rows ─────────────────────────────────────────────────────
const rows = ref<QuoteRow[]>([
  makeProductRow('cp2', 'HIK-DS2CD2T87', 'DS-2CD2T87G2-L 8MP Bullet Camera', 'Hikvision', 45, undefined, 32, 862, 1199),
  makeProductRow('cp3', 'HIK-DS7732NI', 'DS-7732NI-K4 32CH NVR', 'Hikvision', 12, undefined, 4, 1820, 2436),
  makeProductRow('cp10', 'HON-MAXPRO', 'MAXPRO Access 4-Door Controller', 'Honeywell', 8, undefined, 6, 2100, 2800),
  makeProductRow('cp14', 'ZKT-SPEEDFACE', 'SpeedFace-V5L Facial Terminal', 'ZKTeco', 0, 30, 12, 1508, 2917),
  makeProductRow('cp15', 'CBL-CAT6A-305', 'Cat6A UTP Cable 305m Box', 'Belden', 30, undefined, 8, 420, 580),
  makeServiceRow('cs1', 'SVC-INSTALL-SR', 'Senior Installation Engineer (per day)', 'per day', 20, 800, 1200),
  makeServiceRow('cs2', 'SVC-INSTALL-JR', 'Technician (per day)', 'per day', 30, 450, 700),
  makeServiceRow('cs3', 'SVC-PM-MONTH', 'Project Manager (per month)', 'per month', 2, 12000, 18000),
])

function makeProductRow(pid: string, sku: string, desc: string, mfr: string, stock: number, lead: number | undefined, qty: number, cost: number, price: number): QuoteRow {
  const lt = qty * price
  return { id: uid(), rowType: 'item', source: 'product', productId: pid, sku, description: desc, manufacturer: mfr, stockAvailable: stock > 0 ? stock : undefined, leadTimeDays: lead, quantity: qty, multiplier: 1, unitCost: cost, unitPrice: price, discountPercent: 0, lineTotal: lt, marginPercent: price > 0 ? ((price - cost) / price) * 100 : 0, isOptional: false, isSelected: true, isPrintable: true }
}

function makeServiceRow(pid: string, sku: string, desc: string, rateType: string, qty: number, cost: number, price: number): QuoteRow {
  const lt = qty * price
  return { id: uid(), rowType: 'item', source: 'service', productId: pid, sku, description: desc, rateType, quantity: qty, multiplier: 1, unitCost: cost, unitPrice: price, discountPercent: 0, lineTotal: lt, marginPercent: price > 0 ? ((price - cost) / price) * 100 : 0, isOptional: false, isSelected: true, isPrintable: true }
}

// ── Row Computations ─────────────────────────────────────────
function recalcRow(row: QuoteRow) {
  if (row.rowType !== 'item') return
  const gross = row.quantity * row.multiplier * row.unitPrice
  const disc = gross * (row.discountPercent / 100)
  row.lineTotal = gross - disc
  const costTotal = row.quantity * row.multiplier * row.unitCost
  row.marginPercent = row.lineTotal > 0 ? ((row.lineTotal - costTotal) / row.lineTotal) * 100 : 0
}

// ── Item rows (for totals, exclude headings/comments/subtotals) ──
const itemRows = computed(() => rows.value.filter(r => r.rowType === 'item'))
const activeItemRows = computed(() => itemRows.value.filter(r => !r.isOptional || r.isSelected))

// ── Totals ───────────────────────────────────────────────────
const subtotal = computed(() => activeItemRows.value.reduce((s, r) => s + r.lineTotal, 0))
const discountAmount = computed(() => subtotal.value * (discountPercent.value / 100))
const subtotalAfterDiscount = computed(() => subtotal.value - discountAmount.value)
const vatPercent = 15
const vatAmount = computed(() => subtotalAfterDiscount.value * (vatPercent / 100))
const total = computed(() => subtotalAfterDiscount.value + vatAmount.value)
const totalCost = computed(() => activeItemRows.value.reduce((s, r) => s + r.quantity * r.multiplier * r.unitCost, 0))
const marginAmount = computed(() => subtotalAfterDiscount.value - totalCost.value)
const overallMargin = computed(() => subtotalAfterDiscount.value > 0 ? (marginAmount.value / subtotalAfterDiscount.value) * 100 : 0)

// Category breakdowns
const categoryBreakdown = computed(() => {
  const map: Record<string, { count: number; total: number }> = { product: { count: 0, total: 0 }, service: { count: 0, total: 0 }, recurring: { count: 0, total: 0 }, 'write-in': { count: 0, total: 0 } }
  for (const r of activeItemRows.value) {
    map[r.source].count++
    map[r.source].total += r.lineTotal
  }
  return map
})

// ── Approval Logic ───────────────────────────────────────────
const requiresApproval = computed(() => overallMargin.value < 20 || discountPercent.value > 10 || total.value > 500000)
const approvalReasons = computed(() => {
  const reasons: string[] = []
  if (overallMargin.value < 20) reasons.push(`Low margin (${overallMargin.value.toFixed(1)}%)`)
  if (discountPercent.value > 10) reasons.push(`High discount (${discountPercent.value}%)`)
  if (total.value > 500000) reasons.push(`High value (SAR ${formatSAR(total.value)})`)
  return reasons
})

// ── Filter Tabs ──────────────────────────────────────────────
type TabFilter = 'all' | ItemSource
const activeTab = ref<TabFilter>('all')

const filteredRows = computed(() => {
  if (activeTab.value === 'all') return rows.value
  return rows.value.filter(r => r.rowType !== 'item' || r.source === activeTab.value)
})

const tabCounts = computed(() => ({
  all: itemRows.value.length,
  product: itemRows.value.filter(r => r.source === 'product').length,
  service: itemRows.value.filter(r => r.source === 'service').length,
  recurring: itemRows.value.filter(r => r.source === 'recurring').length,
  'write-in': itemRows.value.filter(r => r.source === 'write-in').length,
}))

// ── Add Items: Search ────────────────────────────────────────
const addSource = ref<ItemSource>('product')
const searchQuery = ref('')
const showDropdown = ref(false)

function delayHideDropdown() { window.setTimeout(() => { showDropdown.value = false }, 200) }

watch(searchQuery, (val) => { showDropdown.value = val.trim().length > 0 })

const searchResults = computed(() => {
  const q = searchQuery.value.toLowerCase().trim()
  if (!q) return [] as { id: string; sku: string; name: string; meta: string; price: string; stockLabel?: string; inStock?: boolean }[]

  if (addSource.value === 'product') {
    return productCatalog.filter(p => p.sku.toLowerCase().includes(q) || p.name.toLowerCase().includes(q) || p.manufacturer.toLowerCase().includes(q)).slice(0, 10).map(p => ({
      id: p.id, sku: p.sku, name: p.name, meta: p.manufacturer,
      price: `SAR ${formatSAR(p.unitPrice)}`,
      stockLabel: p.stockAvailable > 0 ? `${p.stockAvailable} in stock` : `${p.leadTimeDays}d lead`,
      inStock: p.stockAvailable > 0,
    }))
  }
  if (addSource.value === 'service') {
    return serviceCatalog.filter(s => s.sku.toLowerCase().includes(q) || s.name.toLowerCase().includes(q) || s.department.toLowerCase().includes(q)).slice(0, 10).map(s => ({
      id: s.id, sku: s.sku, name: s.name, meta: `${s.department} · ${s.rateType}`,
      price: `SAR ${formatSAR(s.unitPrice)}/${s.rateType}`,
    }))
  }
  if (addSource.value === 'recurring') {
    return recurringCatalog.filter(r => r.sku.toLowerCase().includes(q) || r.name.toLowerCase().includes(q)).slice(0, 10).map(r => ({
      id: r.id, sku: r.sku, name: r.name, meta: r.billingCycle,
      price: `SAR ${formatSAR(r.monthlyPrice)}/mo`,
    }))
  }
  return []
})

function selectSearchItem(item: { id: string }) {
  if (addSource.value === 'product') {
    const p = productCatalog.find(x => x.id === item.id)
    if (!p) return
    rows.value.push(makeProductRow(p.id, p.sku, p.name, p.manufacturer, p.stockAvailable, p.stockAvailable === 0 ? p.leadTimeDays : undefined, 1, p.unitCost, p.unitPrice))
  } else if (addSource.value === 'service') {
    const s = serviceCatalog.find(x => x.id === item.id)
    if (!s) return
    rows.value.push(makeServiceRow(s.id, s.sku, s.name, s.rateType, 1, s.unitCost, s.unitPrice))
  } else if (addSource.value === 'recurring') {
    const r = recurringCatalog.find(x => x.id === item.id)
    if (!r) return
    const row: QuoteRow = { id: uid(), rowType: 'item', source: 'recurring', productId: r.id, sku: r.sku, description: r.name, billingCycle: r.billingCycle, quantity: 1, multiplier: 12, unitCost: r.monthlyCost, unitPrice: r.monthlyPrice, discountPercent: 0, lineTotal: 12 * r.monthlyPrice, marginPercent: r.monthlyPrice > 0 ? ((r.monthlyPrice - r.monthlyCost) / r.monthlyPrice) * 100 : 0, isOptional: false, isSelected: true, isPrintable: true }
    rows.value.push(row)
  }
  searchQuery.value = ''
  showDropdown.value = false
}

// ── Quick Add Bar ────────────────────────────────────────────
const quickAddValue = ref('')
function handleQuickAdd() {
  const val = quickAddValue.value.trim()
  if (!val) return
  const parts = val.split(/[\s,]+/)
  const skuPart = parts[0].toUpperCase()
  const qty = parts.length > 1 ? parseInt(parts[1]) || 1 : 1

  const found = productCatalog.find(p => p.sku.toUpperCase() === skuPart) ||
    serviceCatalog.find(s => s.sku.toUpperCase() === skuPart)
  if (found && 'manufacturer' in found) {
    const p = found as CatalogProduct
    rows.value.push(makeProductRow(p.id, p.sku, p.name, p.manufacturer, p.stockAvailable, p.stockAvailable === 0 ? p.leadTimeDays : undefined, qty, p.unitCost, p.unitPrice))
  } else if (found) {
    const s = found as CatalogService
    rows.value.push(makeServiceRow(s.id, s.sku, s.name, s.rateType, qty, s.unitCost, s.unitPrice))
  } else {
    const recFound = recurringCatalog.find(r => r.sku.toUpperCase() === skuPart)
    if (recFound) {
      rows.value.push({ id: uid(), rowType: 'item', source: 'recurring', productId: recFound.id, sku: recFound.sku, description: recFound.name, billingCycle: recFound.billingCycle, quantity: qty, multiplier: 12, unitCost: recFound.monthlyCost, unitPrice: recFound.monthlyPrice, discountPercent: 0, lineTotal: qty * 12 * recFound.monthlyPrice, marginPercent: recFound.monthlyPrice > 0 ? ((recFound.monthlyPrice - recFound.monthlyCost) / recFound.monthlyPrice) * 100 : 0, isOptional: false, isSelected: true, isPrintable: true })
    } else {
      showSaveToast(`SKU "${skuPart}" not found`)
      return
    }
  }
  quickAddValue.value = ''
  showSaveToast(`Added ${qty}× ${skuPart}`)
}

// ── Add Special Rows ─────────────────────────────────────────
function addWriteInRow() {
  rows.value.push({ id: uid(), rowType: 'item', source: 'write-in', sku: '', description: '', quantity: 1, multiplier: 1, unitCost: 0, unitPrice: 0, discountPercent: 0, lineTotal: 0, marginPercent: 0, isOptional: false, isSelected: true, isPrintable: true })
}

function addHeading() {
  rows.value.push({ id: uid(), rowType: 'heading', source: 'write-in', sku: '', description: '', headingText: 'Section Title', quantity: 0, multiplier: 1, unitCost: 0, unitPrice: 0, discountPercent: 0, lineTotal: 0, marginPercent: 0, isOptional: false, isSelected: true, isPrintable: true })
}

function addComment() {
  rows.value.push({ id: uid(), rowType: 'comment', source: 'write-in', sku: '', description: '', commentText: '', quantity: 0, multiplier: 1, unitCost: 0, unitPrice: 0, discountPercent: 0, lineTotal: 0, marginPercent: 0, isOptional: false, isSelected: true, isPrintable: true })
}

function addSubtotalLine() {
  rows.value.push({ id: uid(), rowType: 'subtotal', source: 'write-in', sku: '', description: 'Subtotal', quantity: 0, multiplier: 1, unitCost: 0, unitPrice: 0, discountPercent: 0, lineTotal: 0, marginPercent: 0, isOptional: false, isSelected: true, isPrintable: true })
}

function removeRow(id: string) { rows.value = rows.value.filter(r => r.id !== id) }

function duplicateRow(row: QuoteRow) {
  const clone: QuoteRow = { ...row, id: uid() }
  const idx = rows.value.findIndex(r => r.id === row.id)
  rows.value.splice(idx + 1, 0, clone)
}

function toggleOptional(row: QuoteRow) {
  row.isOptional = !row.isOptional
  if (row.isOptional) row.isSelected = false
  else row.isSelected = true
}

function computeRunningSubtotal(upToIndex: number): number {
  let sum = 0
  for (let i = 0; i <= upToIndex; i++) {
    const r = rows.value[i]
    if (r.rowType === 'item' && (!r.isOptional || r.isSelected)) sum += r.lineTotal
  }
  return sum
}

// ── Save Actions ─────────────────────────────────────────────
const saveMessage = ref('')
function showSaveToast(msg: string) { saveMessage.value = msg; window.setTimeout(() => { saveMessage.value = '' }, 2500) }
function saveDraft() { quoteStatus.value = 'draft'; showSaveToast('Draft saved') }
function submitForApproval() { quoteStatus.value = 'pending-approval'; showSaveToast('Submitted for approval') }

// ── Print ────────────────────────────────────────────────────
function printQuote() {
  const printableRows = rows.value.filter(r => r.isPrintable)
  let lineNum = 0

  const rowsHtml = printableRows.map(r => {
    if (r.rowType === 'heading') {
      return `<tr class="heading"><td colspan="7" style="background:#f1f5f9;font-weight:700;text-transform:uppercase;letter-spacing:.04em;padding:10px 12px;font-size:13px;border-bottom:2px solid #cbd5e1">${r.headingText || ''}</td></tr>`
    }
    if (r.rowType === 'comment') {
      return `<tr class="comment"><td colspan="7" style="background:#fffbeb;padding:8px 12px;font-style:italic;color:#78716c;font-size:12px">${r.commentText || ''}</td></tr>`
    }
    if (r.rowType === 'subtotal') {
      return `<tr class="subtotal"><td colspan="5" style="border-top:2px solid #94a3b8;padding:8px 12px;font-weight:600;text-transform:uppercase;font-size:11px;color:#64748b">Subtotal</td><td colspan="2" style="border-top:2px solid #94a3b8;text-align:right;padding:8px 12px;font-weight:700;font-family:monospace">SAR ${formatSAR(computeRunningSubtotal(rows.value.indexOf(r)))}</td></tr>`
    }
    lineNum++
    const optLabel = r.isOptional ? ' <span style="color:#f59e0b;font-size:10px">(OPTIONAL)</span>' : ''
    return `<tr>
      <td style="text-align:center;color:#94a3b8;width:30px">${lineNum}</td>
      <td><strong>${r.description}</strong>${optLabel}<br><span style="font-family:monospace;font-size:11px;color:#3b82f6">${r.sku || ''}</span> ${r.manufacturer ? `<span style="color:#94a3b8;font-size:11px">· ${r.manufacturer}</span>` : ''}</td>
      <td style="text-align:center">${r.quantity}</td>
      <td style="text-align:right;font-family:monospace">SAR ${formatSAR(r.unitPrice)}</td>
      <td style="text-align:right">${r.discountPercent > 0 ? r.discountPercent.toFixed(1) + '%' : '—'}</td>
      <td style="text-align:right;font-family:monospace;font-weight:600">SAR ${formatSAR(r.lineTotal)}</td>
    </tr>`
  }).join('')

  const soldToHtml = `<div style="flex:1">
    <h4 style="margin:0 0 8px;font-size:11px;text-transform:uppercase;letter-spacing:.06em;color:#64748b">Sold To</h4>
    <p style="margin:0;font-weight:600">${soldTo.value.contactName}</p>
    <p style="margin:0">${soldTo.value.company}</p>
    <p style="margin:0;color:#64748b">${soldTo.value.address}</p>
    <p style="margin:0;color:#64748b">${soldTo.value.city}, ${soldTo.value.country}</p>
    <p style="margin:4px 0 0;font-size:12px;color:#64748b">${soldTo.value.phone} · ${soldTo.value.email}</p>
  </div>`

  const shipToHtml = shipTo.value.contactName ? `<div style="flex:1">
    <h4 style="margin:0 0 8px;font-size:11px;text-transform:uppercase;letter-spacing:.06em;color:#64748b">Ship To</h4>
    <p style="margin:0;font-weight:600">${shipTo.value.contactName}</p>
    <p style="margin:0">${shipTo.value.company}</p>
    <p style="margin:0;color:#64748b">${shipTo.value.address}</p>
    <p style="margin:0;color:#64748b">${shipTo.value.city}, ${shipTo.value.country}</p>
    <p style="margin:4px 0 0;font-size:12px;color:#64748b">${shipTo.value.phone} · ${shipTo.value.email}</p>
  </div>` : ''

  const html = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>${quoteNumber.value}</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif; font-size:13px; color:#1e293b; padding:40px; line-height:1.5; }
  .header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:32px; padding-bottom:20px; border-bottom:3px solid #2563eb; }
  .logo { font-size:28px; font-weight:800; color:#2563eb; }
  .logo-sub { font-size:11px; color:#64748b; text-transform:uppercase; letter-spacing:.1em; }
  .quote-info { text-align:right; }
  .quote-num { font-size:20px; font-weight:700; color:#1e293b; }
  .quote-meta { font-size:12px; color:#64748b; margin-top:4px; }
  .addresses { display:flex; gap:40px; margin-bottom:28px; }
  .intro { margin-bottom:24px; padding:12px 16px; background:#f8fafc; border-left:3px solid #2563eb; font-size:13px; color:#475569; line-height:1.6; }
  table { width:100%; border-collapse:collapse; margin-bottom:24px; }
  th { background:#f1f5f9; padding:10px 12px; font-size:11px; text-transform:uppercase; letter-spacing:.04em; color:#64748b; border-bottom:2px solid #e2e8f0; text-align:left; }
  td { padding:8px 12px; border-bottom:1px solid #f1f5f9; }
  .totals { width:320px; margin-left:auto; margin-bottom:28px; }
  .totals td { padding:6px 12px; font-size:13px; }
  .totals .grand { border-top:2px solid #334155; font-size:16px; font-weight:700; }
  .totals .grand td:last-child { color:#2563eb; }
  .closing { margin-bottom:24px; padding:12px 16px; background:#f0fdf4; border-left:3px solid #22c55e; font-size:13px; color:#475569; line-height:1.6; }
  .footer { margin-top:40px; padding-top:16px; border-top:1px solid #e2e8f0; display:flex; justify-content:space-between; font-size:11px; color:#94a3b8; }
  @media print { body { padding:20px; } @page { margin:15mm; } }
</style></head><body>
<div class="header">
  <div><div class="logo">G4S</div><div class="logo-sub">Security Solutions</div></div>
  <div class="quote-info">
    <div class="quote-num">${quoteNumber.value}</div>
    <div class="quote-meta">Status: ${statusConfig[quoteStatus.value].label} · Valid Until: ${validUntil.value}<br>Payment: ${paymentTerms.value} · Delivery: ${deliveryTerms.value}</div>
  </div>
</div>
<div class="addresses">${soldToHtml}${shipToHtml}</div>
${introductionText.value ? `<div class="intro">${introductionText.value}</div>` : ''}
<table>
  <thead><tr><th style="width:30px;text-align:center">#</th><th>Description</th><th style="text-align:center;width:60px">Qty</th><th style="text-align:right;width:110px">Unit Price</th><th style="text-align:right;width:70px">Disc</th><th style="text-align:right;width:130px">Total</th></tr></thead>
  <tbody>${rowsHtml}</tbody>
</table>
<table class="totals">
  <tr><td>Subtotal</td><td style="text-align:right;font-family:monospace">SAR ${formatSAR(subtotal.value)}</td></tr>
  ${discountPercent.value > 0 ? `<tr><td>Discount (${discountPercent.value}%)</td><td style="text-align:right;font-family:monospace;color:#ef4444">- SAR ${formatSAR(discountAmount.value)}</td></tr>` : ''}
  ${discountPercent.value > 0 ? `<tr><td>After Discount</td><td style="text-align:right;font-family:monospace">SAR ${formatSAR(subtotalAfterDiscount.value)}</td></tr>` : ''}
  <tr><td>VAT (${vatPercent}%)</td><td style="text-align:right;font-family:monospace">SAR ${formatSAR(vatAmount.value)}</td></tr>
  <tr class="grand"><td>Grand Total</td><td style="text-align:right;font-family:monospace">SAR ${formatSAR(total.value)}</td></tr>
</table>
${closingText.value ? `<div class="closing">${closingText.value}</div>` : ''}
${statementOfWork.value ? `<div style="margin-bottom:24px"><h4 style="font-size:11px;text-transform:uppercase;letter-spacing:.06em;color:#64748b;margin-bottom:8px">Statement of Work</h4><p style="font-size:13px;color:#475569;line-height:1.6;white-space:pre-wrap">${statementOfWork.value}</p></div>` : ''}
<div class="footer"><span>Generated on ${new Date().toLocaleDateString('en-GB', { day:'numeric', month:'long', year:'numeric' })}</span><span>${quoteNumber.value} · ${customerName.value}</span></div>
</body></html>`

  const w = window.open('', '_blank', 'width=900,height=700')
  if (w) {
    w.document.write(html)
    w.document.close()
    w.onload = () => { w.print() }
  }
}

// ── Excel/CSV Export ─────────────────────────────────────────
function exportExcel() {
  const sep = ','
  const esc = (v: string) => `"${v.replace(/"/g, '""')}"`

  const headerRow = ['#', 'Type', 'SKU', 'Description', 'Manufacturer', 'Qty', 'Multiplier', 'Unit Cost (SAR)', 'Unit Price (SAR)', 'Discount %', 'Line Total (SAR)', 'Margin %', 'Optional', 'Stock Status'].join(sep)

  let lineNum = 0
  const dataRows = rows.value.map(r => {
    if (r.rowType === 'heading') return ['', 'HEADING', '', esc(r.headingText || ''), '', '', '', '', '', '', '', '', '', ''].join(sep)
    if (r.rowType === 'comment') return ['', 'COMMENT', '', esc(r.commentText || ''), '', '', '', '', '', '', '', '', '', ''].join(sep)
    if (r.rowType === 'subtotal') return ['', 'SUBTOTAL', '', '', '', '', '', '', '', '', formatSAR(computeRunningSubtotal(rows.value.indexOf(r))), '', '', ''].join(sep)
    lineNum++
    const status = stockStatus(r)
    const stockStr = status === 'in-stock' ? `In stock (${r.stockAvailable})` : status === 'low-stock' ? `Low (${r.stockAvailable}/${r.quantity})` : status === 'out-of-stock' ? (r.leadTimeDays ? `${r.leadTimeDays}d lead` : 'Out of stock') : ''
    return [
      lineNum,
      sourceLabels[r.source],
      esc(r.sku),
      esc(r.description),
      esc(r.manufacturer || ''),
      r.quantity,
      r.multiplier,
      r.unitCost.toFixed(2),
      r.unitPrice.toFixed(2),
      r.discountPercent.toFixed(1),
      r.lineTotal.toFixed(2),
      r.marginPercent.toFixed(1),
      r.isOptional ? 'Yes' : 'No',
      esc(stockStr),
    ].join(sep)
  })

  const summaryRows = [
    '',
    ['', '', '', '', '', '', '', '', '', 'Subtotal', subtotal.value.toFixed(2), '', '', ''].join(sep),
    discountPercent.value > 0 ? ['', '', '', '', '', '', '', '', '', `Discount (${discountPercent.value}%)`, (-discountAmount.value).toFixed(2), '', '', ''].join(sep) : null,
    ['', '', '', '', '', '', '', '', '', `VAT (${vatPercent}%)`, vatAmount.value.toFixed(2), '', '', ''].join(sep),
    ['', '', '', '', '', '', '', '', '', 'Grand Total', total.value.toFixed(2), '', '', ''].join(sep),
    '',
    ['', '', '', '', '', '', '', '', '', 'Total Cost', totalCost.value.toFixed(2), '', '', ''].join(sep),
    ['', '', '', '', '', '', '', '', '', 'Margin', marginAmount.value.toFixed(2), '', '', ''].join(sep),
    ['', '', '', '', '', '', '', '', '', 'Margin %', overallMargin.value.toFixed(1) + '%', '', '', ''].join(sep),
  ].filter(Boolean)

  const metaRows = [
    `Quote Number,${quoteNumber.value}`,
    `Customer,${esc(customerName.value)}`,
    `Status,${statusConfig[quoteStatus.value].label}`,
    `Valid Until,${validUntil.value}`,
    `Payment Terms,${esc(paymentTerms.value)}`,
    `Delivery Terms,${esc(deliveryTerms.value)}`,
    `Currency,${currency.value}`,
    '',
  ]

  const csv = [...metaRows, headerRow, ...dataRows, ...summaryRows].join('\n')
  const BOM = '\uFEFF'
  const blob = new Blob([BOM + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${quoteNumber.value}_${customerName.value.replace(/[^a-zA-Z0-9]/g, '_')}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  showSaveToast('Excel file downloaded')
}
</script>

<template>
  <div class="builder-page">
    <!-- Save Toast -->
    <Transition name="toast">
      <div v-if="saveMessage" class="save-toast"><CheckCircle2 :size="16" />{{ saveMessage }}</div>
    </Transition>

    <!-- Header -->
    <div class="builder-header">
      <div class="builder-header-left">
        <button class="btn btn-ghost btn-icon" @click="router.push('/quotes')"><ArrowLeft :size="20" /></button>
        <div>
          <h1 class="builder-title">{{ quoteNumber }}</h1>
          <div class="builder-subtitle">
            <span :class="['badge badge-dot', statusConfig[quoteStatus].badge]">{{ statusConfig[quoteStatus].label }}</span>
            <span class="text-muted">{{ customerName }}</span>
          </div>
        </div>
      </div>
      <div class="builder-header-actions">
        <button class="btn btn-secondary btn-sm" @click="saveDraft"><Save :size="14" /> Save Draft</button>
        <button class="btn btn-primary btn-sm" @click="submitForApproval"><Send :size="14" /> Submit</button>
      </div>
    </div>

    <!-- Info Bar -->
    <div class="info-bar">
      <div class="info-field">
        <label class="info-label">Customer</label>
        <span class="info-value"><Building2 :size="14" class="info-icon" /> {{ customerName }}</span>
      </div>
      <div class="info-field">
        <label class="info-label">Currency</label>
        <span class="info-value">{{ currency }}</span>
      </div>
      <div class="info-field">
        <label class="info-label">Valid Until</label>
        <input v-model="validUntil" type="date" class="info-date" />
      </div>
      <div class="info-field">
        <label class="info-label">Payment Terms</label>
        <select v-model="paymentTerms" class="info-select">
          <option>Net 30</option><option>Net 60</option><option>50% Advance, 50% on Delivery</option><option>100% Advance</option>
        </select>
      </div>
      <div class="info-field">
        <label class="info-label">Delivery</label>
        <select v-model="deliveryTerms" class="info-select">
          <option>Ex-Works</option><option>FOB</option><option>CIF</option><option>DDP</option>
        </select>
      </div>
    </div>

    <!-- Builder Tabs -->
    <div class="builder-tabs">
      <button :class="['builder-tab', builderTab === 'items' && 'builder-tab--active']" @click="builderTab = 'items'">
        <Package :size="16" /> Document Items <span class="builder-tab-count">{{ itemRows.length }}</span>
      </button>
      <button :class="['builder-tab', builderTab === 'addresses' && 'builder-tab--active']" @click="builderTab = 'addresses'">
        <MapPin :size="16" /> Sold To / Ship To
      </button>
      <button :class="['builder-tab', builderTab === 'notes' && 'builder-tab--active']" @click="builderTab = 'notes'">
        <FileText :size="16" /> Notes
      </button>
    </div>

    <!-- Tab: Document Items -->
    <div v-show="builderTab === 'items'" class="builder-content">
      <div class="builder-main">
        <!-- Add Items Panel -->
        <div class="card add-panel">
          <div class="add-panel-body">
            <!-- Source Tabs + Search -->
            <div class="add-top-row">
              <div class="source-tabs">
                <button v-for="src in (['product', 'service', 'recurring'] as ItemSource[])" :key="src" :class="['source-tab', addSource === src && 'source-tab--active']" @click="addSource = src">
                  <component :is="sourceIcons[src]" :size="14" /> {{ sourceLabels[src] }}
                </button>
              </div>
              <div class="add-special-btns">
                <button class="btn btn-ghost btn-sm" @click="addWriteInRow" title="Write-in (external item)"><FileText :size="14" /> Write-in</button>
                <button class="btn btn-ghost btn-sm" @click="addHeading" title="Section heading"><Hash :size="14" /> Heading</button>
                <button class="btn btn-ghost btn-sm" @click="addComment" title="Comment line"><MessageSquare :size="14" /> Comment</button>
                <button class="btn btn-ghost btn-sm" @click="addSubtotalLine" title="Subtotal line"><Minus :size="14" /> Subtotal</button>
              </div>
            </div>

            <div class="add-search-row">
              <!-- Main search -->
              <div class="add-search-wrapper">
                <div class="search-input">
                  <Search :size="16" class="search-icon" />
                  <input v-model="searchQuery" type="text" class="form-input" :placeholder="`Search ${sourceLabels[addSource].toLowerCase()}s by SKU or name...`" @focus="showDropdown = searchQuery.trim().length > 0" @blur="delayHideDropdown" />
                </div>

                <div v-if="showDropdown && searchResults.length" class="product-dropdown">
                  <div v-for="item in searchResults" :key="item.id" class="dd-item-wrap">
                    <button class="dd-item" @mousedown.prevent="selectSearchItem(item)">
                      <div class="dd-main">
                        <span class="dd-sku">{{ item.sku }}</span>
                        <span class="dd-name">{{ item.name }}</span>
                        <span v-if="item.meta" class="dd-meta">{{ item.meta }}</span>
                      </div>
                      <div class="dd-right">
                        <span class="dd-price">{{ item.price }}</span>
                        <span v-if="item.stockLabel" :class="['dd-stock', item.inStock ? 'dd-stock--in' : 'dd-stock--out']">
                          <span class="dd-stock-dot" /> {{ item.stockLabel }}
                        </span>
                      </div>
                    </button>
                    <button v-if="addSource === 'product'" class="dd-history-btn" title="View price history & batches" @mousedown.prevent="openPriceHistory(item.id)">
                      <History :size="13" />
                    </button>
                  </div>
                </div>

                <div v-if="showDropdown && searchQuery.trim() && !searchResults.length" class="product-dropdown dd-empty">
                  <div class="dd-empty-inner"><Search :size="16" /> No results for "{{ searchQuery }}"</div>
                </div>
              </div>

              <!-- Quick Add -->
              <div class="quick-add">
                <Zap :size="14" class="quick-icon" />
                <input v-model="quickAddValue" type="text" class="form-input quick-input" placeholder="SKU, qty  (e.g. HIK-DS2CD2143, 10)" @keydown.enter="handleQuickAdd" />
              </div>
            </div>
          </div>
        </div>

        <!-- Row Type Tabs -->
        <div class="tabs">
          <button :class="['tab', activeTab === 'all' && 'tab--active']" @click="activeTab = 'all'">All <span class="tab-count">{{ tabCounts.all }}</span></button>
          <button v-for="src in (['product', 'service', 'recurring', 'write-in'] as ItemSource[])" :key="src" :class="['tab', activeTab === src && 'tab--active']" @click="activeTab = src">
            <component :is="sourceIcons[src]" :size="14" /> {{ sourceLabels[src] }} <span class="tab-count">{{ tabCounts[src] }}</span>
          </button>
        </div>

        <!-- Spotlight Search -->
        <div class="spotlight-bar">
          <Search :size="14" class="spotlight-icon" />
          <input v-model="spotlightQuery" type="text" class="spotlight-input" placeholder="Spotlight: type to highlight matching items..." />
          <span v-if="spotlightActive" class="spotlight-count">
            {{ filteredRows.filter(r => isSpotlightMatch(r)).length }} / {{ filteredRows.length }} matching
          </span>
          <button v-if="spotlightActive" class="btn btn-ghost btn-icon btn-sm" @click="spotlightQuery = ''"><X :size="14" /></button>
        </div>

        <!-- Line Items Table -->
        <div v-if="filteredRows.length" class="card">
          <div class="table-wrap">
            <table class="builder-table">
              <thead>
                <tr>
                  <th class="col-grip"></th>
                  <th class="col-num">#</th>
                  <th class="col-type">Type</th>
                  <th>Description / SKU</th>
                  <th class="text-center col-qty">Qty</th>
                  <th class="text-center col-mult">×</th>
                  <th class="text-right col-price">Unit Cost</th>
                  <th class="text-right col-price">Unit Price</th>
                  <th class="text-right col-disc">Disc%</th>
                  <th class="text-right col-total">Total</th>
                  <th class="text-right col-margin">Margin</th>
                  <th class="col-opt">Opt</th>
                  <th class="col-actions"></th>
                </tr>
              </thead>
              <tbody>
                <template v-for="(row, idx) in filteredRows" :key="row.id">
                  <!-- HEADING ROW -->
                  <tr v-if="row.rowType === 'heading'" class="heading-row" :class="{ 'item-row--spotlight-dim': spotlightActive && !isSpotlightMatch(row) }">
                    <td class="col-grip"><GripVertical :size="14" class="drag-handle" /></td>
                    <td></td>
                    <td colspan="9">
                      <input v-model="row.headingText" type="text" class="heading-input" placeholder="Section Title..." />
                    </td>
                    <td></td>
                    <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeRow(row.id)"><Trash2 :size="14" /></button></td>
                  </tr>

                  <!-- COMMENT ROW -->
                  <tr v-else-if="row.rowType === 'comment'" class="comment-row" :class="{ 'item-row--spotlight-dim': spotlightActive && !isSpotlightMatch(row) }">
                    <td class="col-grip"><GripVertical :size="14" class="drag-handle" /></td>
                    <td></td>
                    <td colspan="9">
                      <input v-model="row.commentText" type="text" class="comment-input" placeholder="Add a comment or note..." />
                    </td>
                    <td></td>
                    <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeRow(row.id)"><Trash2 :size="14" /></button></td>
                  </tr>

                  <!-- SUBTOTAL ROW -->
                  <tr v-else-if="row.rowType === 'subtotal'" class="subtotal-row" :class="{ 'item-row--spotlight-dim': spotlightActive }">
                    <td class="col-grip"><GripVertical :size="14" class="drag-handle" /></td>
                    <td></td>
                    <td colspan="7" class="subtotal-label">
                      <Minus :size="14" /> Running Subtotal
                    </td>
                    <td class="text-right subtotal-value">SAR {{ formatSAR(computeRunningSubtotal(rows.indexOf(row))) }}</td>
                    <td></td>
                    <td></td>
                    <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeRow(row.id)"><Trash2 :size="14" /></button></td>
                  </tr>

                  <!-- ITEM ROW -->
                  <tr v-else class="item-row" :class="{ 'item-row--optional': row.isOptional && !row.isSelected, 'item-row--selected-optional': row.isOptional && row.isSelected, 'item-row--spotlight-dim': spotlightActive && !isSpotlightMatch(row), 'item-row--spotlight-match': spotlightActive && isSpotlightMatch(row) }">
                    <td class="col-grip"><GripVertical :size="14" class="drag-handle" /></td>
                    <td class="col-num text-muted">{{ idx + 1 }}</td>
                    <td class="col-type">
                      <component :is="sourceIcons[row.source]" :size="14" :class="'source-icon source-icon--' + row.source" :title="sourceLabels[row.source]" />
                    </td>
                    <td>
                      <div class="desc-cell">
                        <input v-model="row.description" type="text" class="inline-input inline-input--desc" :placeholder="row.source === 'write-in' ? 'Enter item description...' : 'Description'" />
                        <div class="desc-meta">
                          <span v-if="row.sku" class="cell-sku">{{ row.sku }}</span>
                          <span v-if="row.source === 'write-in'" class="cell-writein">Write-in</span>
                          <span v-if="row.manufacturer" class="cell-mfr">{{ row.manufacturer }}</span>
                          <span v-if="row.rateType" class="cell-rate">{{ row.rateType }}</span>
                          <span v-if="row.billingCycle" class="cell-rate">{{ row.billingCycle }} × {{ row.multiplier }}mo</span>
                          <span v-if="stockStatus(row) === 'in-stock'" class="stock-badge stock-badge--in">
                            <PackageCheck :size="10" /> {{ row.stockAvailable }} in stock
                          </span>
                          <span v-else-if="stockStatus(row) === 'low-stock'" class="stock-badge stock-badge--low">
                            <AlertTriangle :size="10" /> {{ row.stockAvailable }}/{{ row.quantity }} avail
                          </span>
                          <span v-else-if="stockStatus(row) === 'out-of-stock'" class="stock-badge stock-badge--out">
                            <Truck :size="10" /> {{ row.leadTimeDays ? `${row.leadTimeDays}d lead` : 'Out of stock' }}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td class="text-center col-qty">
                      <input v-model.number="row.quantity" type="number" min="1" class="inline-input inline-input--num" @change="recalcRow(row)" />
                    </td>
                    <td class="text-center col-mult">
                      <input v-model.number="row.multiplier" type="number" min="1" class="inline-input inline-input--num" @change="recalcRow(row)" />
                    </td>
                    <td class="text-right col-price">
                      <input v-model.number="row.unitCost" type="number" step="0.01" min="0" class="inline-input inline-input--price" @change="recalcRow(row)" />
                    </td>
                    <td class="text-right col-price">
                      <input v-model.number="row.unitPrice" type="number" step="0.01" min="0" class="inline-input inline-input--price" @change="recalcRow(row)" />
                    </td>
                    <td class="text-right col-disc">
                      <div class="disc-wrap">
                        <input v-model.number="row.discountPercent" type="number" step="0.5" min="0" max="100" class="inline-input inline-input--disc" @change="recalcRow(row)" />
                        <span class="disc-sym">%</span>
                      </div>
                    </td>
                    <td class="text-right col-total whitespace-nowrap font-medium">SAR {{ formatSAR(row.lineTotal) }}</td>
                    <td class="text-right col-margin">
                      <span :class="['font-semibold', marginClass(row.marginPercent)]">{{ row.marginPercent.toFixed(1) }}%</span>
                    </td>
                    <td class="col-opt">
                      <button class="opt-btn" :class="{ 'opt-btn--active': row.isOptional }" :title="row.isOptional ? 'Mark as required' : 'Mark as optional'" @click="toggleOptional(row)">
                        <EyeOff v-if="row.isOptional && !row.isSelected" :size="14" />
                        <Eye v-else :size="14" />
                      </button>
                    </td>
                    <td class="col-actions">
                      <div class="row-actions">
                        <button v-if="row.source === 'product' && row.productId" class="btn btn-ghost btn-icon btn-sm price-history-btn" title="Price History / Batches" @click="openPriceHistory(row.productId!, row.id)">
                          <History :size="13" />
                        </button>
                        <button class="btn btn-ghost btn-icon btn-sm" title="Duplicate" @click="duplicateRow(row)"><Copy :size="13" /></button>
                        <button class="btn btn-ghost btn-icon btn-sm" title="Remove" @click="removeRow(row.id)"><Trash2 :size="14" /></button>
                      </div>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else class="empty-state">
          <Package :size="40" class="empty-state-icon" />
          <h3 class="empty-state-title">No line items</h3>
          <p class="empty-state-text">Search for products, services, or recurring items, use quick-add, or add a write-in item.</p>
        </div>
      </div>

      <!-- Summary Panel -->
      <div class="builder-summary">
        <div class="card summary-card">
          <div class="card-body">
            <h3 class="summary-title">Quote Summary</h3>

            <!-- Category Breakdown -->
            <div class="breakdown">
              <div v-for="src in (['product', 'service', 'recurring', 'write-in'] as ItemSource[])" :key="src" class="breakdown-row">
                <div class="breakdown-label">
                  <component :is="sourceIcons[src]" :size="13" :class="'source-icon--' + src" />
                  {{ sourceLabels[src] }} ({{ categoryBreakdown[src].count }})
                </div>
                <span class="text-mono">SAR {{ formatSAR(categoryBreakdown[src].total) }}</span>
              </div>
            </div>

            <div class="summary-divider" />

            <!-- Totals -->
            <div class="summary-rows">
              <div class="summary-row">
                <span>Subtotal</span>
                <span class="text-mono">SAR {{ formatSAR(subtotal) }}</span>
              </div>
              <div class="summary-row summary-row--input">
                <div class="disc-label">
                  <span>Discount</span>
                  <div class="disc-input-wrap">
                    <input v-model.number="discountPercent" type="number" step="0.5" min="0" max="100" class="disc-input" />
                    <span class="disc-input-sym">%</span>
                  </div>
                </div>
                <span class="text-mono text-danger">- SAR {{ formatSAR(discountAmount) }}</span>
              </div>
              <div class="summary-row summary-row--sub">
                <span class="font-medium">After Discount</span>
                <span class="text-mono font-medium">SAR {{ formatSAR(subtotalAfterDiscount) }}</span>
              </div>
              <div class="summary-row">
                <span>VAT ({{ vatPercent }}%)</span>
                <span class="text-mono">SAR {{ formatSAR(vatAmount) }}</span>
              </div>
              <div class="summary-row summary-row--total">
                <span class="font-bold">Total</span>
                <span class="summary-grand">SAR {{ formatSAR(total) }}</span>
              </div>
            </div>

            <!-- Cost Breakdown -->
            <div class="cost-section">
              <h4 class="cost-title">Cost & Margin</h4>
              <div class="summary-rows">
                <div class="summary-row"><span>Total Cost</span><span class="text-mono">SAR {{ formatSAR(totalCost) }}</span></div>
                <div class="summary-row"><span>Margin</span><span class="text-mono text-success">SAR {{ formatSAR(marginAmount) }}</span></div>
                <div class="summary-row"><span class="font-medium">Margin %</span><span :class="['font-bold', marginClass(overallMargin)]">{{ overallMargin.toFixed(1) }}%</span></div>
              </div>
              <div class="margin-bar-track"><div class="margin-bar-fill" :class="marginClass(overallMargin)" :style="{ width: `${Math.min(overallMargin, 50) * 2}%` }" /></div>
            </div>

            <!-- Procurement Needed -->
            <div v-if="outOfStockRows.length > 0" class="procurement-section">
              <h4 class="procurement-title">
                <ShoppingCart :size="14" />
                Procurement Needed
              </h4>
              <div class="procurement-items">
                <div v-for="row in outOfStockRows" :key="row.id" class="procurement-item">
                  <div class="procurement-item-info">
                    <span class="procurement-item-name">{{ row.description }}</span>
                    <span class="procurement-item-meta">
                      {{ row.sku }} &middot;
                      <span v-if="row.stockAvailable != null && row.stockAvailable > 0">
                        Need {{ row.quantity - row.stockAvailable }} more
                      </span>
                      <span v-else>
                        {{ row.quantity }} units
                      </span>
                    </span>
                  </div>
                  <span :class="['stock-badge', stockStatus(row) === 'low-stock' ? 'stock-badge--low' : 'stock-badge--out']">
                    {{ row.leadTimeDays ? `${row.leadTimeDays}d` : 'N/A' }}
                  </span>
                </div>
              </div>
              <button class="btn btn-warning btn-sm procurement-btn" @click="showPOModal = true">
                <ShoppingCart :size="14" /> Generate PO ({{ outOfStockRows.length }} items)
              </button>
            </div>

            <div v-else-if="itemRows.filter(r => r.source === 'product').length > 0" class="stock-ok">
              <PackageCheck :size="14" />
              <span>All product items are in stock</span>
            </div>

            <!-- Approval Warning -->
            <div v-if="requiresApproval" class="approval-warn">
              <AlertCircle :size="16" />
              <div>
                <strong>Approval Required</strong>
                <p v-for="reason in approvalReasons" :key="reason" class="approval-reason">{{ reason }}</p>
              </div>
            </div>

            <!-- Actions -->
            <div class="summary-actions">
              <button class="btn btn-primary btn-lg summary-action-btn" @click="saveDraft"><Save :size="16" /> Save Draft</button>
              <button class="btn btn-success summary-action-btn" @click="submitForApproval"><CheckCircle2 :size="16" /> Submit for Approval</button>
              <div class="summary-secondary">
                <button class="btn btn-secondary btn-sm" @click="printQuote"><Printer :size="14" /> Print</button>
                <button class="btn btn-secondary btn-sm" @click="exportExcel"><FileSpreadsheet :size="14" /> Excel</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: Sold To / Ship To -->
    <div v-show="builderTab === 'addresses'" class="tab-content-panel">
      <div class="addresses-grid">
        <div class="card address-card">
          <div class="address-card-header">
            <h3 class="address-card-title"><Building2 :size="16" /> Sold To</h3>
          </div>
          <div class="address-card-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Contact Name</label>
                <input v-model="soldTo.contactName" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Company</label>
                <input v-model="soldTo.company" type="text" class="form-input" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Address</label>
              <input v-model="soldTo.address" type="text" class="form-input" />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">City</label>
                <input v-model="soldTo.city" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Country</label>
                <input v-model="soldTo.country" type="text" class="form-input" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Phone</label>
                <input v-model="soldTo.phone" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Email</label>
                <input v-model="soldTo.email" type="email" class="form-input" />
              </div>
            </div>
          </div>
        </div>

        <div class="card address-card">
          <div class="address-card-header">
            <h3 class="address-card-title"><MapPin :size="16" /> Ship To</h3>
            <button class="btn btn-ghost btn-sm" @click="copySoldToShipTo"><Copy :size="14" /> Copy from Sold To</button>
          </div>
          <div class="address-card-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Contact Name</label>
                <input v-model="shipTo.contactName" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Company</label>
                <input v-model="shipTo.company" type="text" class="form-input" />
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Address</label>
              <input v-model="shipTo.address" type="text" class="form-input" />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">City</label>
                <input v-model="shipTo.city" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Country</label>
                <input v-model="shipTo.country" type="text" class="form-input" />
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Phone</label>
                <input v-model="shipTo.phone" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label class="form-label">Email</label>
                <input v-model="shipTo.email" type="email" class="form-input" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab: Notes -->
    <div v-show="builderTab === 'notes'" class="tab-content-panel">
      <div class="notes-grid">
        <div class="card notes-card notes-card--full">
          <div class="notes-card-header">
            <h3 class="notes-card-title">Introduction</h3>
            <span class="notes-hint">Printed before the quote body</span>
          </div>
          <div class="notes-card-body">
            <textarea v-model="introductionText" class="form-input notes-textarea" rows="4" placeholder="Opening text for your quotation..." />
          </div>
        </div>

        <div class="card notes-card notes-card--full">
          <div class="notes-card-header">
            <h3 class="notes-card-title">Statement of Work</h3>
            <span class="notes-hint">Detailed scope description</span>
          </div>
          <div class="notes-card-body">
            <textarea v-model="statementOfWork" class="form-input notes-textarea notes-textarea--lg" rows="8" placeholder="Describe the scope of work, deliverables, timeline, and acceptance criteria..." />
          </div>
        </div>

        <div class="card notes-card">
          <div class="notes-card-header">
            <h3 class="notes-card-title">Closing</h3>
            <span class="notes-hint">Printed after the quote body</span>
          </div>
          <div class="notes-card-body">
            <textarea v-model="closingText" class="form-input notes-textarea" rows="4" placeholder="Closing remarks, terms, or next steps..." />
          </div>
        </div>

        <div class="card notes-card">
          <div class="notes-card-header">
            <h3 class="notes-card-title">Internal Notes</h3>
            <span class="notes-hint notes-hint--private">Private — not printed</span>
          </div>
          <div class="notes-card-body">
            <textarea v-model="internalNotes" class="form-input notes-textarea" rows="4" placeholder="Internal notes for your team..." />
          </div>
        </div>

        <div class="card notes-card notes-card--full">
          <div class="notes-card-header">
            <h3 class="notes-card-title">Purchasing Notes</h3>
            <span class="notes-hint">Used in PO workflow</span>
          </div>
          <div class="notes-card-body">
            <textarea v-model="purchasingNotes" class="form-input notes-textarea" rows="3" placeholder="Notes for procurement team..." />
          </div>
        </div>
      </div>
    </div>

    <!-- Generate PO Modal -->
    <Teleport to="body">
      <div v-if="showPOModal" class="modal-backdrop" @click.self="showPOModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title"><ShoppingCart :size="20" /> Generate Purchase Orders</h2>
            <button class="modal-close" @click="showPOModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <p class="po-modal-desc">
              The following <strong>{{ outOfStockRows.length }}</strong> items are out of stock or have insufficient quantity.
              Purchase orders will be grouped by manufacturer/supplier.
            </p>

            <div class="po-items-table">
              <table class="table">
                <thead>
                  <tr>
                    <th>Item</th>
                    <th>Manufacturer</th>
                    <th class="text-center">Needed</th>
                    <th class="text-center">In Stock</th>
                    <th class="text-center">To Order</th>
                    <th class="text-right">Est. Cost</th>
                    <th class="text-center">Lead Time</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in outOfStockRows" :key="row.id">
                    <td>
                      <div class="po-item-desc">
                        <span class="font-medium">{{ row.description }}</span>
                        <span class="text-mono text-muted" style="font-size:0.6875rem">{{ row.sku }}</span>
                      </div>
                    </td>
                    <td>{{ row.manufacturer || '—' }}</td>
                    <td class="text-center">{{ row.quantity }}</td>
                    <td class="text-center">{{ row.stockAvailable ?? 0 }}</td>
                    <td class="text-center font-bold">{{ row.stockAvailable != null ? Math.max(0, row.quantity - row.stockAvailable) : row.quantity }}</td>
                    <td class="text-right whitespace-nowrap text-mono">
                      SAR {{ formatSAR((row.stockAvailable != null ? Math.max(0, row.quantity - row.stockAvailable) : row.quantity) * row.unitCost) }}
                    </td>
                    <td class="text-center">
                      <span class="stock-badge stock-badge--out">{{ row.leadTimeDays || '?' }}d</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="po-modal-summary">
              <div class="po-modal-summary-row">
                <span>Total items to procure</span>
                <span class="font-bold">{{ outOfStockRows.reduce((s, r) => s + (r.stockAvailable != null ? Math.max(0, r.quantity - r.stockAvailable) : r.quantity), 0) }} units</span>
              </div>
              <div class="po-modal-summary-row">
                <span>Estimated procurement cost</span>
                <span class="font-bold text-mono">SAR {{ formatSAR(outOfStockRows.reduce((s, r) => s + (r.stockAvailable != null ? Math.max(0, r.quantity - r.stockAvailable) : r.quantity) * r.unitCost, 0)) }}</span>
              </div>
              <div class="po-modal-summary-row">
                <span>POs to create (grouped by supplier)</span>
                <span class="font-bold">{{ new Set(outOfStockRows.map(r => r.manufacturer || 'Unknown')).size }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showPOModal = false">Cancel</button>
            <button class="btn btn-warning" @click="generatePOFromQuote">
              <ShoppingCart :size="16" /> Create Purchase Orders
            </button>
          </div>
        </div>
      </div>
      <!-- Price History / Batch Selection Modal -->
      <div v-if="showPriceHistoryModal && priceHistoryProduct" class="modal-backdrop" @click.self="showPriceHistoryModal = false">
        <div class="modal modal-xl">
          <div class="modal-header">
            <div class="ph-modal-header-left">
              <h2 class="modal-title"><History :size="20" /> Product Price History</h2>
              <div class="ph-product-badge">
                <span class="ph-sku">{{ priceHistoryProduct.sku }}</span>
                <span class="ph-name">{{ priceHistoryProduct.name }}</span>
                <span class="ph-mfr">{{ priceHistoryProduct.manufacturer }}</span>
              </div>
            </div>
            <button class="modal-close" @click="showPriceHistoryModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <!-- Stats Cards -->
            <div class="ph-stats">
              <div class="ph-stat-card">
                <div class="ph-stat-label">Average Cost</div>
                <div class="ph-stat-value">SAR {{ formatSAR(priceHistoryAvgCost) }}</div>
              </div>
              <div class="ph-stat-card">
                <div class="ph-stat-label">Average Price</div>
                <div class="ph-stat-value">SAR {{ formatSAR(priceHistoryAvgPrice) }}</div>
              </div>
              <div class="ph-stat-card ph-stat-card--green">
                <div class="ph-stat-label"><TrendingDown :size="13" /> Lowest Cost</div>
                <div class="ph-stat-value">SAR {{ formatSAR(priceHistoryLowest.cost) }}</div>
              </div>
              <div class="ph-stat-card ph-stat-card--red">
                <div class="ph-stat-label"><TrendingUp :size="13" /> Highest Cost</div>
                <div class="ph-stat-value">SAR {{ formatSAR(priceHistoryHighest.cost) }}</div>
              </div>
              <div class="ph-stat-card">
                <div class="ph-stat-label">Current Default</div>
                <div class="ph-stat-value">SAR {{ formatSAR(priceHistoryProduct.unitPrice) }}</div>
              </div>
            </div>

            <!-- Filter Row -->
            <div class="ph-filters">
              <div class="ph-filter-group">
                <label class="ph-filter-label">Source</label>
                <select v-model="priceHistoryFilterSource" class="form-select form-select--sm">
                  <option value="">All Sources</option>
                  <option v-for="src in priceHistorySources" :key="src" :value="src">{{ src }}</option>
                </select>
              </div>
              <div class="ph-filter-info">
                {{ priceHistoryEntries.length }} {{ priceHistoryEntries.length === 1 ? 'entry' : 'entries' }}
              </div>
            </div>

            <!-- Price History Table -->
            <div class="ph-table-wrap">
              <table class="table ph-table">
                <thead>
                  <tr>
                    <th class="ph-th-sortable" @click="togglePriceHistorySort('date')">
                      Date
                      <ChevronDown v-if="priceHistorySortField === 'date'" :size="12" :class="{ 'sort-asc': priceHistorySortAsc }" />
                    </th>
                    <th class="ph-th-sortable" @click="togglePriceHistorySort('supplier')">
                      Supplier
                      <ChevronDown v-if="priceHistorySortField === 'supplier'" :size="12" :class="{ 'sort-asc': priceHistorySortAsc }" />
                    </th>
                    <th>Source</th>
                    <th>Batch / Ref</th>
                    <th class="text-right ph-th-sortable" @click="togglePriceHistorySort('cost')">
                      Cost
                      <ChevronDown v-if="priceHistorySortField === 'cost'" :size="12" :class="{ 'sort-asc': priceHistorySortAsc }" />
                    </th>
                    <th class="text-right ph-th-sortable" @click="togglePriceHistorySort('price')">
                      Price
                      <ChevronDown v-if="priceHistorySortField === 'price'" :size="12" :class="{ 'sort-asc': priceHistorySortAsc }" />
                    </th>
                    <th class="text-right">Margin</th>
                    <th class="text-center ph-th-sortable" @click="togglePriceHistorySort('qty')">
                      Qty
                      <ChevronDown v-if="priceHistorySortField === 'qty'" :size="12" :class="{ 'sort-asc': priceHistorySortAsc }" />
                    </th>
                    <th>Modifier</th>
                    <th>Customer / Quote</th>
                    <th class="text-center">Use</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="entry in priceHistoryEntries" :key="entry.id" class="ph-row">
                    <td class="whitespace-nowrap">{{ entry.date }}</td>
                    <td>{{ entry.supplier }}</td>
                    <td>
                      <span :class="['ph-source-badge', 'ph-source-badge--' + entry.source.toLowerCase().replace(/\s+/g, '-')]">
                        {{ entry.source }}
                      </span>
                    </td>
                    <td class="text-mono text-muted" style="font-size:0.75rem">{{ entry.batchRef || '—' }}</td>
                    <td class="text-right text-mono whitespace-nowrap">
                      <span :class="{ 'ph-cost-low': entry.cost === priceHistoryLowest.cost, 'ph-cost-high': entry.cost === priceHistoryHighest.cost }">
                        SAR {{ formatSAR(entry.cost) }}
                      </span>
                    </td>
                    <td class="text-right text-mono whitespace-nowrap font-medium">SAR {{ formatSAR(entry.price) }}</td>
                    <td class="text-right">
                      <span :class="['font-semibold', marginClass(entry.price > 0 ? ((entry.price - entry.cost) / entry.price) * 100 : 0)]">
                        {{ entry.price > 0 ? (((entry.price - entry.cost) / entry.price) * 100).toFixed(1) : '0.0' }}%
                      </span>
                    </td>
                    <td class="text-center">{{ entry.qty }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ entry.modifier || '—' }}</td>
                    <td>
                      <div v-if="entry.customer || entry.quoteRef" class="ph-customer-info">
                        <span v-if="entry.customer" class="ph-customer-name">{{ entry.customer }}</span>
                        <span v-if="entry.quoteRef" class="ph-quote-ref">{{ entry.quoteRef }}</span>
                      </div>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td class="text-center">
                      <button class="btn btn-primary btn-sm ph-apply-btn" title="Apply this price" @click="applyPriceFromHistory(entry)">
                        <Check :size="13" /> Use
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer ph-modal-footer">
            <div class="ph-footer-left">
              <span class="ph-footer-avg">Avg Cost: <strong>SAR {{ formatSAR(priceHistoryAvgCost) }}</strong></span>
              <span class="ph-footer-avg">Avg Price: <strong>SAR {{ formatSAR(priceHistoryAvgPrice) }}</strong></span>
            </div>
            <div class="ph-footer-right">
              <button class="btn btn-secondary" @click="showPriceHistoryModal = false">Cancel</button>
              <button v-if="!priceHistoryTargetRowId" class="btn btn-primary" @click="addWithDefaultPrice">
                <Plus :size="16" /> Add with Default Price
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.builder-page { padding: var(--space-6); position: relative; }

/* Toast */
.save-toast { position: fixed; top: var(--space-4); right: var(--space-6); display: flex; align-items: center; gap: var(--space-2); padding: var(--space-3) var(--space-5); background: var(--color-success); color: white; border-radius: var(--radius-lg); font-size: var(--text-sm); font-weight: 500; box-shadow: 0 4px 12px rgba(0,0,0,.15); z-index: 1000; }
.toast-enter-active, .toast-leave-active { transition: all .3s ease; }
.toast-enter-from, .toast-leave-to { opacity: 0; transform: translateY(-12px); }

/* Header */
.builder-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-5); }
.builder-header-left { display: flex; align-items: center; gap: var(--space-3); }
.builder-title { font-size: var(--text-2xl); font-weight: 700; color: var(--color-neutral-900); }
.builder-subtitle { display: flex; align-items: center; gap: var(--space-2); margin-top: 2px; font-size: var(--text-sm); }
.builder-header-actions { display: flex; gap: var(--space-2); }

/* Info Bar */
.info-bar { display: flex; align-items: flex-end; gap: var(--space-4); padding: var(--space-4) var(--space-5); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); margin-bottom: var(--space-5); flex-wrap: wrap; }
.info-field { display: flex; flex-direction: column; gap: 2px; }
.info-label { font-size: 0.6875rem; font-weight: 600; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: .03em; }
.info-value { display: flex; align-items: center; gap: var(--space-1); font-size: var(--text-sm); color: var(--color-neutral-800); font-weight: 500; }
.info-icon { color: var(--color-neutral-400); }
.info-date { font-size: var(--text-sm); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-sm); padding: 4px 8px; background: transparent; font-family: inherit; color: var(--color-neutral-800); }
.info-select { font-size: var(--text-sm); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-sm); padding: 4px 8px; background: transparent; font-family: inherit; color: var(--color-neutral-800); width: 180px; }

/* Content Layout */
.builder-content { display: grid; grid-template-columns: 1fr 320px; gap: var(--space-5); align-items: start; }
.builder-main { min-width: 0; display: flex; flex-direction: column; gap: var(--space-4); }

/* Add Panel */
.add-panel { border: 2px dashed var(--color-neutral-200); background: var(--color-neutral-50); }
.add-panel-body { padding: var(--space-4) var(--space-5); }
.add-top-row { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-3); flex-wrap: wrap; gap: var(--space-2); }
.source-tabs { display: flex; gap: var(--space-1); }
.source-tab { display: inline-flex; align-items: center; gap: 4px; padding: 5px 12px; font-size: var(--text-xs); font-weight: 500; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-full); background: var(--content-surface); color: var(--color-neutral-600); cursor: pointer; transition: all .15s; }
.source-tab:hover { border-color: var(--color-primary); color: var(--color-primary); }
.source-tab--active { background: var(--color-primary); color: white; border-color: var(--color-primary); }
.add-special-btns { display: flex; gap: var(--space-1); }
.add-search-row { display: flex; gap: var(--space-3); align-items: stretch; }
.add-search-wrapper { flex: 1; position: relative; }
.quick-add { display: flex; align-items: center; gap: var(--space-1); width: 280px; flex-shrink: 0; position: relative; }
.quick-icon { position: absolute; left: 10px; color: var(--color-warning); pointer-events: none; z-index: 1; }
.quick-input { padding-left: 30px !important; font-family: var(--font-mono); font-size: var(--text-xs); }

/* Dropdown */
.product-dropdown { position: absolute; top: 100%; left: 0; right: 0; margin-top: 4px; background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); box-shadow: 0 8px 24px rgba(0,0,0,.12); max-height: 360px; overflow-y: auto; z-index: 50; }
.dd-item { display: flex; align-items: center; justify-content: space-between; gap: var(--space-3); width: 100%; padding: var(--space-3) var(--space-4); border: none; background: none; cursor: pointer; text-align: left; transition: background .1s; }
.dd-item:hover { background: var(--color-primary-light, #eff6ff); }
.dd-item:not(:last-child) { border-bottom: 1px solid var(--color-neutral-100); }
.dd-main { display: flex; align-items: center; gap: var(--space-2); flex: 1; min-width: 0; }
.dd-sku { font-family: var(--font-mono); font-size: var(--text-xs); font-weight: 600; color: var(--color-primary); background: var(--color-primary-light, #eff6ff); padding: 2px 6px; border-radius: var(--radius-sm); white-space: nowrap; }
.dd-name { font-size: var(--text-sm); color: var(--color-neutral-800); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.dd-meta { font-size: var(--text-xs); color: var(--color-neutral-400); white-space: nowrap; }
.dd-right { display: flex; align-items: center; gap: var(--space-3); flex-shrink: 0; }
.dd-price { font-family: var(--font-mono); font-size: var(--text-xs); color: var(--color-neutral-600); white-space: nowrap; }
.dd-stock { display: inline-flex; align-items: center; gap: 4px; font-size: var(--text-xs); white-space: nowrap; }
.dd-stock-dot { width: 7px; height: 7px; border-radius: 50%; }
.dd-stock--in { color: var(--color-success); }
.dd-stock--in .dd-stock-dot { background: var(--color-success); }
.dd-stock--out { color: var(--color-warning); }
.dd-stock--out .dd-stock-dot { background: var(--color-warning); }
.dd-empty { padding: 0; }
.dd-empty-inner { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-4); font-size: var(--text-sm); color: var(--color-neutral-400); }

/* Tabs */
.tabs { display: flex; gap: var(--space-1); border-bottom: 2px solid var(--color-neutral-200); }
.tab { display: inline-flex; align-items: center; gap: 4px; padding: var(--space-2) var(--space-3); font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-500); background: none; border: none; border-bottom: 2px solid transparent; margin-bottom: -2px; cursor: pointer; transition: all .15s; white-space: nowrap; }
.tab:hover { color: var(--color-neutral-700); }
.tab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.tab-count { font-size: var(--text-xs); background: var(--color-neutral-100); color: var(--color-neutral-600); padding: 0 6px; border-radius: var(--radius-full); font-weight: 600; }
.tab--active .tab-count { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); }

/* Table */
.table-wrap { overflow-x: auto; }
.builder-table { width: 100%; border-collapse: separate; border-spacing: 0; }
.builder-table th { padding: 8px 10px; font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); background: var(--color-neutral-50); border-bottom: 1px solid var(--color-neutral-200); white-space: nowrap; }
.builder-table td { padding: 5px 10px; vertical-align: middle; border-bottom: 1px solid var(--color-neutral-100); font-size: var(--text-sm); }
.col-grip { width: 28px; text-align: center; }
.col-num { width: 32px; text-align: center; }
.col-type { width: 32px; text-align: center; }
.col-qty { width: 70px; }
.col-mult { width: 54px; }
.col-price { width: 110px; }
.col-disc { width: 80px; }
.col-total { width: 130px; }
.col-margin { width: 70px; }
.col-opt { width: 36px; text-align: center; }
.col-actions { width: 60px; }

.drag-handle { color: var(--color-neutral-300); cursor: grab; }
.item-row:hover .drag-handle { color: var(--color-neutral-500); }
.item-row:hover { background: var(--color-neutral-50); }

.item-row--optional { opacity: .55; }
.item-row--optional td { background: repeating-linear-gradient(45deg, transparent, transparent 10px, rgba(0,0,0,.02) 10px, rgba(0,0,0,.02) 20px); }
.item-row--selected-optional { background: var(--color-primary-light, #eff6ff); }

/* Source Icons */
.source-icon--product { color: var(--color-primary); }
.source-icon--service { color: var(--color-warning); }
.source-icon--recurring { color: var(--color-success); }
.source-icon--write-in { color: var(--color-neutral-500); }

/* Description Cell */
.desc-cell { display: flex; flex-direction: column; gap: 1px; }
.desc-meta { display: flex; align-items: center; gap: var(--space-2); flex-wrap: wrap; }
.cell-sku { font-family: var(--font-mono); font-size: 0.6875rem; color: var(--color-primary); background: var(--color-primary-light, #eff6ff); padding: 0 4px; border-radius: 3px; }
.cell-writein { font-size: 0.6875rem; color: var(--color-neutral-500); background: var(--color-neutral-100); padding: 0 4px; border-radius: 3px; font-style: italic; }
.cell-mfr { font-size: 0.6875rem; color: var(--color-neutral-400); }
.cell-rate { font-size: 0.6875rem; color: var(--color-warning); background: var(--color-warning-light); padding: 0 4px; border-radius: 3px; }
/* Stock Badges */
.stock-badge { display: inline-flex; align-items: center; gap: 3px; font-size: 0.6875rem; font-weight: 600; padding: 1px 6px; border-radius: var(--radius-full); white-space: nowrap; }
.stock-badge--in { background: var(--color-success-light); color: var(--color-success); }
.stock-badge--low { background: var(--color-warning-light); color: var(--color-warning-dark, #92400e); }
.stock-badge--out { background: var(--color-danger-light, #fef2f2); color: var(--color-danger); }

/* Inline Inputs */
.inline-input { width: 100%; padding: 3px 6px; font-size: var(--text-sm); font-family: var(--font-mono); border: 1px solid transparent; border-radius: var(--radius-sm); background: transparent; color: var(--color-neutral-800); transition: all .15s; outline: none; }
.inline-input:hover { border-color: var(--color-neutral-200); background: var(--content-surface); }
.inline-input:focus { border-color: var(--color-primary); background: var(--content-surface); box-shadow: 0 0 0 2px rgba(37,99,235,.15); }
.inline-input--desc { font-family: inherit; font-weight: 500; }
.inline-input--num { text-align: center; width: 60px; }
.inline-input--price { text-align: right; width: 100px; }
.inline-input--disc { text-align: right; width: 56px; padding-right: 16px; }
.disc-wrap { position: relative; display: inline-flex; align-items: center; }
.disc-sym { position: absolute; right: 4px; font-size: var(--text-xs); color: var(--color-neutral-400); pointer-events: none; }

/* Optional button */
.opt-btn { width: 28px; height: 28px; display: flex; align-items: center; justify-content: center; border: none; background: transparent; color: var(--color-neutral-300); cursor: pointer; border-radius: var(--radius-sm); transition: all .15s; }
.opt-btn:hover { background: var(--color-neutral-100); color: var(--color-neutral-600); }
.opt-btn--active { color: var(--color-warning); }

.row-actions { display: flex; gap: 2px; }

/* Special Rows */
.heading-row td { background: var(--color-neutral-100) !important; border-bottom-color: var(--color-neutral-200) !important; }
.heading-input { width: 100%; padding: 4px 8px; font-size: var(--text-sm); font-weight: 700; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-700); border: none; background: transparent; outline: none; }
.heading-input:focus { background: var(--content-surface); border-radius: var(--radius-sm); }

.comment-row td { background: var(--color-warning-light, #fffbeb) !important; }
.comment-input { width: 100%; padding: 4px 8px; font-size: var(--text-sm); font-style: italic; color: var(--color-neutral-600); border: none; background: transparent; outline: none; }
.comment-input:focus { background: var(--content-surface); border-radius: var(--radius-sm); }

.subtotal-row td { background: var(--color-neutral-50) !important; border-top: 2px solid var(--color-neutral-300) !important; font-weight: 600; }
.subtotal-label { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); color: var(--color-neutral-600); text-transform: uppercase; letter-spacing: .03em; }
.subtotal-value { font-family: var(--font-mono); font-size: var(--text-sm); font-weight: 700; color: var(--color-neutral-800); }

/* Summary Panel */
.builder-summary { position: sticky; top: calc(var(--header-height, 64px) + var(--space-6)); }
.summary-card .card-body { padding: var(--space-5); }
.summary-title { font-size: var(--text-sm); font-weight: 600; color: var(--color-neutral-700); text-transform: uppercase; letter-spacing: .04em; margin-bottom: var(--space-4); }

.breakdown { display: flex; flex-direction: column; gap: var(--space-1); margin-bottom: var(--space-3); }
.breakdown-row { display: flex; justify-content: space-between; align-items: center; font-size: var(--text-xs); color: var(--color-neutral-600); padding: 3px 0; }
.breakdown-label { display: flex; align-items: center; gap: 4px; }

.summary-divider { height: 1px; background: var(--color-neutral-200); margin: var(--space-2) 0; }
.summary-rows { display: flex; flex-direction: column; }
.summary-row { display: flex; justify-content: space-between; align-items: center; padding: var(--space-2) 0; font-size: var(--text-sm); color: var(--color-neutral-600); }
.summary-row--sub { padding-top: var(--space-3); border-top: 1px solid var(--color-neutral-200); }
.summary-row--total { padding-top: var(--space-3); margin-top: var(--space-1); border-top: 2px solid var(--color-neutral-300); }
.summary-row--input { flex-wrap: wrap; }
.summary-grand { font-size: var(--text-lg); font-weight: 700; color: var(--color-primary); font-family: var(--font-mono); }

.disc-label { display: flex; align-items: center; gap: var(--space-2); }
.disc-input-wrap { position: relative; display: inline-flex; align-items: center; }
.disc-input { width: 56px; padding: 3px 18px 3px 6px; font-size: var(--text-sm); font-family: var(--font-mono); text-align: right; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-sm); outline: none; }
.disc-input:focus { border-color: var(--color-primary); }
.disc-input-sym { position: absolute; right: 6px; font-size: var(--text-xs); color: var(--color-neutral-400); pointer-events: none; }

.cost-section { margin-top: var(--space-4); padding-top: var(--space-3); border-top: 1px solid var(--color-neutral-200); }
.cost-title { font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); margin-bottom: var(--space-2); }
.margin-bar-track { width: 100%; height: 6px; background: var(--color-neutral-200); border-radius: 3px; margin-top: var(--space-2); overflow: hidden; }
.margin-bar-fill { height: 100%; border-radius: 3px; transition: width .4s ease; }
.margin-bar-fill.margin-high { background: var(--color-success); }
.margin-bar-fill.margin-medium { background: var(--color-warning); }
.margin-bar-fill.margin-low { background: var(--color-danger); }

/* Approval Warning */
.approval-warn { display: flex; gap: var(--space-3); padding: var(--space-3) var(--space-4); background: var(--color-warning-light); border: 1px solid var(--color-warning); border-radius: var(--radius-md); margin-top: var(--space-4); font-size: var(--text-sm); color: var(--color-neutral-800); }
.approval-warn strong { display: block; margin-bottom: 2px; }
.approval-reason { font-size: var(--text-xs); color: var(--color-neutral-600); margin: 0; }

.summary-actions { margin-top: var(--space-5); display: flex; flex-direction: column; gap: var(--space-2); }
.summary-action-btn { width: 100%; justify-content: center; }
.summary-secondary { display: flex; gap: var(--space-2); }
.summary-secondary .btn { flex: 1; justify-content: center; }

/* Builder Tabs */
.builder-tabs { display: flex; gap: var(--space-1); border-bottom: 2px solid var(--color-neutral-200); margin-bottom: var(--space-5); }
.builder-tab { display: inline-flex; align-items: center; gap: var(--space-2); padding: var(--space-3) var(--space-4); font-size: var(--text-sm); font-weight: 500; color: var(--color-neutral-500); background: none; border: none; border-bottom: 2px solid transparent; margin-bottom: -2px; cursor: pointer; transition: all .15s; white-space: nowrap; }
.builder-tab:hover { color: var(--color-neutral-700); }
.builder-tab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); }
.builder-tab-count { font-size: var(--text-xs); background: var(--color-neutral-100); color: var(--color-neutral-600); padding: 0 6px; border-radius: var(--radius-full); font-weight: 600; }
.builder-tab--active .builder-tab-count { background: var(--color-primary-light, #eff6ff); color: var(--color-primary); }

/* Spotlight Search */
.spotlight-bar { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-2) var(--space-4); background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); margin-bottom: var(--space-3); }
.spotlight-icon { color: var(--color-neutral-400); flex-shrink: 0; }
.spotlight-input { flex: 1; border: none; background: transparent; font-size: var(--text-sm); color: var(--color-neutral-800); outline: none; font-family: inherit; }
.spotlight-input::placeholder { color: var(--color-neutral-400); }
.spotlight-count { font-size: var(--text-xs); color: var(--color-neutral-500); white-space: nowrap; font-weight: 500; }
.item-row--spotlight-dim { opacity: 0.2; transition: opacity .2s; }
.item-row--spotlight-match { background: var(--color-primary-light, #eff6ff) !important; }

/* Tab Content Panel */
.tab-content-panel { max-width: 1000px; }

/* Addresses */
.addresses-grid { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-5); }
.address-card-header { display: flex; align-items: center; justify-content: space-between; padding: var(--space-4) var(--space-5); border-bottom: 1px solid var(--color-neutral-200); }
.address-card-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: 600; color: var(--color-neutral-700); text-transform: uppercase; letter-spacing: .03em; }
.address-card-body { padding: var(--space-5); display: flex; flex-direction: column; gap: var(--space-4); }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-4); }

/* Notes */
.notes-grid { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-5); }
.notes-card--full { grid-column: 1 / -1; }
.notes-card-header { display: flex; align-items: center; justify-content: space-between; padding: var(--space-3) var(--space-5); border-bottom: 1px solid var(--color-neutral-200); }
.notes-card-title { font-size: var(--text-sm); font-weight: 600; color: var(--color-neutral-700); }
.notes-hint { font-size: var(--text-xs); color: var(--color-neutral-400); font-style: italic; }
.notes-hint--private { color: var(--color-warning); font-weight: 500; font-style: normal; }
.notes-card-body { padding: var(--space-4) var(--space-5); }
.notes-textarea { width: 100%; resize: vertical; font-family: inherit; line-height: 1.6; }
.notes-textarea--lg { min-height: 180px; }

/* Procurement Section (Summary Panel) */
.procurement-section { margin-top: var(--space-4); padding: var(--space-3) var(--space-4); background: var(--color-warning-light); border: 1px solid var(--color-warning); border-radius: var(--radius-md); }
.procurement-title { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: .03em; color: var(--color-warning-dark, #92400e); margin-bottom: var(--space-3); }
.procurement-items { display: flex; flex-direction: column; gap: var(--space-2); margin-bottom: var(--space-3); }
.procurement-item { display: flex; justify-content: space-between; align-items: center; gap: var(--space-2); }
.procurement-item-info { min-width: 0; flex: 1; }
.procurement-item-name { display: block; font-size: var(--text-xs); font-weight: 500; color: var(--color-neutral-800); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.procurement-item-meta { font-size: 0.625rem; color: var(--color-neutral-500); }
.procurement-btn { width: 100%; justify-content: center; }

.stock-ok { display: flex; align-items: center; gap: var(--space-2); padding: var(--space-3) var(--space-4); background: var(--color-success-light); border-radius: var(--radius-md); margin-top: var(--space-4); font-size: var(--text-xs); font-weight: 500; color: var(--color-success); }

/* PO Modal */
.po-modal-desc { font-size: var(--text-sm); color: var(--color-neutral-600); margin-bottom: var(--space-4); line-height: var(--leading-relaxed); }
.po-items-table { margin-bottom: var(--space-4); }
.po-item-desc { display: flex; flex-direction: column; gap: 1px; }
.po-modal-summary { background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); padding: var(--space-3) var(--space-4); }
.po-modal-summary-row { display: flex; justify-content: space-between; padding: var(--space-1) 0; font-size: var(--text-sm); color: var(--color-neutral-600); }

/* Dropdown item wrapper for price history button */
.dd-item-wrap { display: flex; align-items: stretch; }
.dd-item-wrap .dd-item { flex: 1; border-radius: var(--radius-md) 0 0 var(--radius-md); }
.dd-history-btn {
  display: flex; align-items: center; justify-content: center;
  width: 34px; background: var(--color-neutral-50); border: none; border-left: 1px solid var(--color-neutral-200);
  color: var(--color-neutral-500); cursor: pointer; transition: all .15s; flex-shrink: 0;
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
}
.dd-history-btn:hover { background: var(--color-primary-50); color: var(--color-primary); }

/* Price History Button on rows */
.price-history-btn { color: var(--color-neutral-400); }
.price-history-btn:hover { color: var(--color-primary) !important; background: var(--color-primary-50) !important; }

/* Price History Modal */
.modal-xl { max-width: 1100px; width: 95vw; }

.ph-modal-header-left { display: flex; flex-direction: column; gap: var(--space-2); }
.ph-product-badge { display: flex; align-items: center; gap: var(--space-3); font-size: var(--text-sm); }
.ph-sku { font-family: var(--font-mono); color: var(--color-primary); font-weight: 600; background: var(--color-primary-50); padding: 1px 8px; border-radius: var(--radius-sm); }
.ph-name { color: var(--color-neutral-700); font-weight: 500; }
.ph-mfr { color: var(--color-neutral-400); }

.ph-stats {
  display: grid; grid-template-columns: repeat(5, 1fr); gap: var(--space-3); margin-bottom: var(--space-5);
}
.ph-stat-card {
  background: var(--color-neutral-50); border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg); padding: var(--space-3) var(--space-4); text-align: center;
}
.ph-stat-card--green { border-color: var(--color-success); background: rgba(16, 185, 129, .05); }
.ph-stat-card--red { border-color: var(--color-danger); background: rgba(239, 68, 68, .05); }
.ph-stat-label { font-size: 0.6875rem; text-transform: uppercase; letter-spacing: .04em; color: var(--color-neutral-500); margin-bottom: 2px; display: flex; align-items: center; justify-content: center; gap: 4px; }
.ph-stat-value { font-size: var(--text-lg); font-weight: 700; color: var(--color-neutral-900); font-family: var(--font-mono); }
.ph-stat-card--green .ph-stat-value { color: var(--color-success); }
.ph-stat-card--red .ph-stat-value { color: var(--color-danger); }

.ph-filters { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-3); }
.ph-filter-group { display: flex; align-items: center; gap: var(--space-2); }
.ph-filter-label { font-size: var(--text-xs); font-weight: 600; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: .04em; }
.ph-filter-info { font-size: var(--text-sm); color: var(--color-neutral-500); }
.form-select--sm { padding: 4px 24px 4px 8px; font-size: var(--text-xs); }

.ph-table-wrap { max-height: 360px; overflow-y: auto; border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); }
.ph-table { font-size: var(--text-xs); }
.ph-table thead th { position: sticky; top: 0; z-index: 2; background: var(--color-neutral-50); font-size: 0.6875rem; text-transform: uppercase; letter-spacing: .04em; white-space: nowrap; }
.ph-th-sortable { cursor: pointer; user-select: none; }
.ph-th-sortable:hover { color: var(--color-primary); }
.ph-th-sortable svg { display: inline-block; vertical-align: middle; margin-left: 2px; transition: transform .15s; }
.ph-th-sortable .sort-asc { transform: rotate(180deg); }

.ph-row { transition: background .1s; }
.ph-row:hover { background: var(--color-primary-50); }

.ph-source-badge {
  display: inline-block; padding: 1px 8px; border-radius: 9999px;
  font-size: 0.625rem; font-weight: 600; text-transform: uppercase; letter-spacing: .03em;
}
.ph-source-badge--distributor { background: var(--color-primary-50); color: var(--color-primary); }
.ph-source-badge--direct-import { background: rgba(16, 185, 129, .1); color: var(--color-success); }
.ph-source-badge--past-quote { background: rgba(245, 158, 11, .1); color: var(--color-warning); }
.ph-source-badge--local-vendor { background: rgba(139, 92, 246, .1); color: #8b5cf6; }

.ph-cost-low { color: var(--color-success); font-weight: 600; }
.ph-cost-high { color: var(--color-danger); font-weight: 600; }

.ph-customer-info { display: flex; flex-direction: column; gap: 1px; }
.ph-customer-name { font-size: 0.6875rem; font-weight: 500; color: var(--color-neutral-700); }
.ph-quote-ref { font-size: 0.625rem; font-family: var(--font-mono); color: var(--color-primary); }

.ph-apply-btn { padding: 3px 10px; font-size: 0.6875rem; display: inline-flex; align-items: center; gap: 4px; }

.ph-modal-footer { display: flex; justify-content: space-between; align-items: center; }
.ph-footer-left { display: flex; gap: var(--space-4); }
.ph-footer-avg { font-size: var(--text-sm); color: var(--color-neutral-500); }
.ph-footer-avg strong { color: var(--color-neutral-800); font-family: var(--font-mono); }
.ph-footer-right { display: flex; gap: var(--space-2); }

/* Responsive */
@media (max-width: 1200px) {
  .builder-content { grid-template-columns: 1fr; }
  .builder-summary { position: static; }
}
@media (max-width: 768px) {
  .info-bar { flex-direction: column; }
  .add-search-row { flex-direction: column; }
  .quick-add { width: 100%; }
  .source-tabs { flex-wrap: wrap; }
  .addresses-grid { grid-template-columns: 1fr; }
  .notes-grid { grid-template-columns: 1fr; }
  .form-row { grid-template-columns: 1fr; }
}
</style>
