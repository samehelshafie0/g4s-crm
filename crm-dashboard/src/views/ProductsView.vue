<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  Plus, Search, Pencil, Trash2, X, Package, Calculator, Eye,
  Building2, History, FileText, Upload, TrendingUp, TrendingDown,
  Minus, DollarSign, ExternalLink, Download, FileUp, Check, AlertTriangle,
  Table2, FileSpreadsheet, RefreshCw,
} from 'lucide-vue-next'
import { parseFile, parseCSVText, downloadTemplate, getMappedValue, getMappedNumber, getMappedInt, type FileType, type ParseResult } from '@/utils/fileParser'
import { useProcurementStore } from '@/stores/procurement'
import { useProductsStore } from '@/stores/products'
import { useManufacturersStore } from '@/stores/manufacturers'
import type {
  Product, ProductType, Currency, ManufacturerCategory,
  ProductVendorEntry, ProductPriceRecord, ProductDocument, ProductDocType, PriceSource,
  SupplierItemEntry,
} from '@/types'

const procStore = useProcurementStore()
const productsStore = useProductsStore()
const mfrStore = useManufacturersStore()

onMounted(() => {
  productsStore.fetchProducts()
  mfrStore.fetchManufacturers()
  procStore.fetchPurchaseOrders()
})

function uid(): string { return Math.random().toString(36).slice(2, 11) }

function formatSAR(value: number): string {
  return value.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatCurrency(value: number, currency: Currency): string {
  const symbols: Record<Currency, string> = { SAR: 'SAR', USD: '$', EUR: '€', GBP: '£', AED: 'AED', CNY: '¥' }
  return `${symbols[currency]} ${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}
function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

// ── Manufacturer reference data ──────────────────────────────
interface MfrRef { id: string; name: string; code: string; categories: ManufacturerCategory[] }

const mfrList: MfrRef[] = [
  { id: 'mfr-hik', name: 'Hikvision', code: 'HIK', categories: [{ id: 'hik-cam', name: 'IP Cameras', description: '' }, { id: 'hik-nvr', name: 'NVRs', description: '' }, { id: 'hik-therm', name: 'Thermal Cameras', description: '' }] },
  { id: 'mfr-dh', name: 'Dahua Technology', code: 'DH', categories: [{ id: 'dh-cam', name: 'IP Cameras', description: '' }, { id: 'dh-nvr', name: 'NVRs', description: '' }, { id: 'dh-ac', name: 'Access Control', description: '' }] },
  { id: 'mfr-axis', name: 'Axis Communications', code: 'AXIS', categories: [{ id: 'axis-cam', name: 'IP Cameras', description: '' }, { id: 'axis-enc', name: 'Encoders', description: '' }, { id: 'axis-int', name: 'Intercoms', description: '' }] },
  { id: 'mfr-hon', name: 'Honeywell', code: 'HON', categories: [{ id: 'hon-fire', name: 'Fire Alarm Panels', description: '' }, { id: 'hon-intru', name: 'Intrusion Detection', description: '' }, { id: 'hon-ac', name: 'Access Control', description: '' }] },
  { id: 'mfr-bosch', name: 'Bosch Security', code: 'BOSCH', categories: [{ id: 'bosch-cam', name: 'IP Cameras', description: '' }, { id: 'bosch-fire', name: 'Fire Alarm Panels', description: '' }, { id: 'bosch-pa', name: 'Public Address', description: '' }] },
  { id: 'mfr-zkt', name: 'ZKTeco', code: 'ZKT', categories: [{ id: 'zkt-reader', name: 'Access Control Readers', description: '' }, { id: 'zkt-att', name: 'Time Attendance', description: '' }, { id: 'zkt-turn', name: 'Turnstiles', description: '' }] },
]
function mfr(id: string): MfrRef | undefined { return mfrList.find((m) => m.id === id) }

function calcLanded(cost: number, fx: number, freight: number, customs: number, clearance: number): number {
  return cost * fx * (1 + freight / 100 + customs / 100 + clearance / 100)
}
function calcSelling(landed: number, margin: number): number {
  if (margin >= 100) return landed * 2
  return landed / (1 - margin / 100)
}

function buildProduct(sku: string, name: string, desc: string, mfrId: string, catId: string, type: ProductType, currency: Currency, unitCost: number, fxRate: number, freight: number, customs: number, clearance: number, margin: number, leadDays: number, supplier: string, active: boolean): Product {
  const m = mfr(mfrId)!; const cat = m.categories.find((c) => c.id === catId)!
  const costSAR = unitCost * fxRate; const landed = calcLanded(unitCost, fxRate, freight, customs, clearance); const selling = calcSelling(landed, margin)
  return { id: uid(), sku, name, description: desc, manufacturerId: mfrId, manufacturerName: m.name, categoryId: catId, categoryName: cat.name, productType: type, originCurrency: currency, unitCostOrigin: unitCost, fxRate, costInSAR: costSAR, freightPercent: freight, customsPercent: customs, clearancePercent: clearance, landedCostSAR: landed, targetMarginPercent: margin, sellingPrice: selling, marginAmount: selling - landed, leadTimeDays: leadDays, supplierName: supplier, isActive: active, createdAt: '2024-03-01T08:00:00Z', updatedAt: '2024-07-01T08:00:00Z' }
}

const products = ref<Product[]>([
  buildProduct('HIK-DS2CD2143', 'DS-2CD2143G2-IU 4MP Dome', 'AcuSense 4MP dome with mic', 'mfr-hik', 'hik-cam', 'import', 'USD', 85, 3.75, 5, 7, 3, 30, 21, 'Al Futtaim Trading', true),
  buildProduct('HIK-DS2CD2T87', 'DS-2CD2T87G2-L 8MP Bullet', 'ColorVu 8MP bullet with strobe', 'mfr-hik', 'hik-cam', 'import', 'USD', 195, 3.75, 5, 7, 3, 28, 21, 'Al Futtaim Trading', true),
  buildProduct('HIK-DS7732NI', 'DS-7732NI-K4 32CH NVR', '32-channel 4K NVR with 4 SATA', 'mfr-hik', 'hik-nvr', 'import', 'USD', 420, 3.75, 6, 7, 3, 25, 28, 'Al Futtaim Trading', true),
  buildProduct('DH-IPC-HFW5442', 'IPC-HFW5442T-ASE 4MP AI Bullet', 'WizMind AI bullet with SMD', 'mfr-dh', 'dh-cam', 'import', 'USD', 110, 3.75, 5, 7, 3, 30, 25, 'Gulf Security Dist.', true),
  buildProduct('DH-NVR5432-EI', 'NVR5432-EI 32CH AI NVR', '32-channel AI NVR with analytics', 'mfr-dh', 'dh-nvr', 'import', 'USD', 680, 3.75, 6, 7, 3, 25, 30, 'Gulf Security Dist.', true),
  buildProduct('DH-ASI7214Y', 'ASI7214Y Face Recognition', 'Face recognition access terminal', 'mfr-dh', 'dh-ac', 'import', 'USD', 340, 3.75, 4, 7, 3, 32, 30, 'Gulf Security Dist.', false),
  buildProduct('AXIS-P3265LVE', 'P3265-LVE 2MP Dome', 'Outdoor dome with DLPU & Forensic WDR', 'mfr-axis', 'axis-cam', 'import', 'EUR', 380, 4.10, 4, 5, 2, 35, 35, 'Norden Communication ME', true),
  buildProduct('AXIS-Q6135LE', 'Q6135-LE PTZ Camera', '32x optical zoom PTZ with IR', 'mfr-axis', 'axis-cam', 'import', 'EUR', 4200, 4.10, 4, 5, 2, 30, 42, 'Norden Communication ME', true),
  buildProduct('HON-MNPDS2', 'Morley-IAS Fire Panel 2L', '2-loop addressable fire alarm panel', 'mfr-hon', 'hon-fire', 'import', 'USD', 1250, 3.75, 5, 8, 3, 28, 45, 'Honeywell Saudi LLC', true),
  buildProduct('HON-MAXPRO', 'MAXPRO Access 4-Door', '4-door access controller board', 'mfr-hon', 'hon-ac', 'local', 'SAR', 2100, 1, 0, 0, 0, 25, 14, 'Honeywell Saudi LLC', true),
  buildProduct('BOSCH-NDV3503', 'FLEXIDOME IP 3000i 5MP', 'Micro dome with EVA forensic', 'mfr-bosch', 'bosch-cam', 'import', 'EUR', 275, 4.10, 4, 5, 2, 32, 28, 'Bosch ME FZE', true),
  buildProduct('BOSCH-FPA5000', 'FPA-5000 Fire Panel', 'Modular fire alarm panel', 'mfr-bosch', 'bosch-fire', 'import', 'EUR', 1800, 4.10, 5, 6, 3, 27, 56, 'Bosch ME FZE', true),
  buildProduct('ZKT-INBIO460', 'InBio460 4-Door Controller', '4-door controller with web management', 'mfr-zkt', 'zkt-reader', 'import', 'CNY', 1650, 0.52, 6, 10, 4, 35, 35, 'ZKTeco ME DMCC', true),
  buildProduct('ZKT-SPEEDFACE', 'SpeedFace-V5L Facial Terminal', 'Visible light face recognition terminal', 'mfr-zkt', 'zkt-att', 'import', 'CNY', 2900, 0.52, 6, 10, 4, 38, 30, 'ZKTeco ME DMCC', true),
  buildProduct('ZKT-TS2122', 'TS2122 Tripod Turnstile', 'Full-height tripod turnstile', 'mfr-zkt', 'zkt-turn', 'import', 'CNY', 8500, 0.52, 8, 10, 4, 30, 60, 'ZKTeco ME DMCC', false),
])

// ── Vendor Catalog per product ───────────────────────────────
const vendorCatalogs = ref<Record<string, ProductVendorEntry[]>>({
  'HIK-DS2CD2143': [
    { id: uid(), vendorName: 'Al Futtaim Trading', vendorSku: 'AFT-HIK-2143', unitCost: 85, currency: 'USD', moq: 50, leadTimeDays: 21, lastQuoteDate: '2026-02-18', catalogSource: 'Annual Price List 2026' },
    { id: uid(), vendorName: 'Hikvision Saudi', vendorSku: 'HIK-SA-2143G2', unitCost: 92, currency: 'USD', moq: 20, leadTimeDays: 14, lastQuoteDate: '2026-01-10', catalogSource: 'SQ-2026-0015' },
    { id: uid(), vendorName: 'Gulf Security Dist.', vendorSku: 'GSD-2143-IU', unitCost: 88, currency: 'USD', moq: 25, leadTimeDays: 18, lastQuoteDate: '2025-11-05', catalogSource: 'Vendor Catalog PDF' },
  ],
  'HIK-DS2CD2T87': [
    { id: uid(), vendorName: 'Al Futtaim Trading', vendorSku: 'AFT-HIK-2T87', unitCost: 195, currency: 'USD', moq: 20, leadTimeDays: 21, lastQuoteDate: '2026-02-18', catalogSource: 'Annual Price List 2026' },
    { id: uid(), vendorName: 'Hikvision Saudi', vendorSku: 'HIK-SA-2T87', unitCost: 210, currency: 'USD', moq: 10, leadTimeDays: 18, lastQuoteDate: '2026-02-18', catalogSource: 'SQ-2026-0015' },
  ],
  'AXIS-Q6135LE': [
    { id: uid(), vendorName: 'Norden Communication ME', vendorSku: 'NOR-Q6135LE', unitCost: 4200, currency: 'EUR', moq: 4, leadTimeDays: 42, lastQuoteDate: '2026-01-15', catalogSource: 'Axis Partner Portal' },
    { id: uid(), vendorName: 'Axis Partner KSA', vendorSku: 'AX-KSA-Q6135', unitCost: 4050, currency: 'EUR', moq: 2, leadTimeDays: 38, lastQuoteDate: '2025-12-01', catalogSource: 'SQ-2026-0011' },
  ],
  'HON-MNPDS2': [
    { id: uid(), vendorName: 'Honeywell Saudi LLC', vendorSku: 'HW-MNPDS2-KSA', unitCost: 1250, currency: 'USD', moq: 2, leadTimeDays: 45, lastQuoteDate: '2026-01-20', catalogSource: 'SQ-2026-0013' },
  ],
  'ZKT-SPEEDFACE': [
    { id: uid(), vendorName: 'ZKTeco ME DMCC', vendorSku: 'ZK-SFV5L-ME', unitCost: 2900, currency: 'CNY', moq: 6, leadTimeDays: 30, lastQuoteDate: '2026-01-15', catalogSource: 'SQ-2026-0012' },
  ],
  'BOSCH-FPA5000': [
    { id: uid(), vendorName: 'Bosch ME FZE', vendorSku: 'B-FPA5000-ME', unitCost: 1800, currency: 'EUR', moq: 2, leadTimeDays: 56, lastQuoteDate: '2025-11-20', catalogSource: 'Framework Agreement 2025' },
  ],
})

// ── Price History per product ────────────────────────────────
const priceHistory = ref<Record<string, ProductPriceRecord[]>>({
  'HIK-DS2CD2143': [
    { id: uid(), date: '2026-02-18', source: 'supplier-quote', sourceRef: 'SQ-2026-0015', vendorName: 'Hikvision Saudi', unitCost: 355, currency: 'SAR', qty: 100 },
    { id: uid(), date: '2026-02-10', source: 'purchase-order', sourceRef: 'PO-2026-0012', vendorName: 'Al Futtaim Trading', unitCost: 85, currency: 'USD', qty: 50, notes: 'Aramco project order' },
    { id: uid(), date: '2025-10-05', source: 'goods-receipt', sourceRef: 'GR-2026-0003', vendorName: 'Hikvision Saudi', unitCost: 385, currency: 'SAR', landingCost: 422, qty: 50 },
    { id: uid(), date: '2025-07-15', source: 'vendor-catalog', vendorName: 'Al Futtaim Trading', unitCost: 90, currency: 'USD', qty: 0, notes: 'Catalog refresh H2 2025' },
    { id: uid(), date: '2025-03-01', source: 'purchase-order', sourceRef: 'PO-2025-0044', vendorName: 'Al Futtaim Trading', unitCost: 92, currency: 'USD', qty: 100 },
    { id: uid(), date: '2024-09-10', source: 'manual', vendorName: 'Gulf Security Dist.', unitCost: 95, currency: 'USD', qty: 0, notes: 'Price check by phone' },
  ],
  'HIK-DS2CD2T87': [
    { id: uid(), date: '2026-02-18', source: 'supplier-quote', sourceRef: 'SQ-2026-0015', vendorName: 'Hikvision Saudi', unitCost: 840, currency: 'SAR', qty: 80 },
    { id: uid(), date: '2026-02-10', source: 'purchase-order', sourceRef: 'PO-2026-0012', vendorName: 'Al Futtaim Trading', unitCost: 195, currency: 'USD', qty: 50 },
    { id: uid(), date: '2025-10-05', source: 'goods-receipt', sourceRef: 'GR-2026-0003', vendorName: 'Hikvision Saudi', unitCost: 890, currency: 'SAR', landingCost: 977, qty: 30 },
  ],
  'AXIS-Q6135LE': [
    { id: uid(), date: '2026-02-15', source: 'purchase-order', sourceRef: 'PO-2026-0008', vendorName: 'Axis Communications', unitCost: 19044, currency: 'SAR', qty: 8 },
    { id: uid(), date: '2025-12-01', source: 'supplier-quote', sourceRef: 'SQ-2026-0011', vendorName: 'Axis Partner KSA', unitCost: 18500, currency: 'SAR', qty: 16 },
    { id: uid(), date: '2025-06-15', source: 'vendor-catalog', vendorName: 'Norden Communication ME', unitCost: 4200, currency: 'EUR', qty: 0, notes: 'H2 2025 catalog' },
  ],
})

// ── Product Documents ────────────────────────────────────────
const productDocs = ref<Record<string, ProductDocument[]>>({
  'HIK-DS2CD2143': [
    { id: uid(), name: 'DS-2CD2143G2-IU Datasheet', docType: 'datasheet', fileName: 'DS-2CD2143G2-IU_Datasheet.pdf', fileSize: '1.2 MB', uploadedBy: 'Ahmed bin Saleh', uploadedAt: '2025-06-15T10:00:00Z', url: '#' },
    { id: uid(), name: 'Installation Guide', docType: 'manual', fileName: 'HIK-Dome-Install-Guide.pdf', fileSize: '3.8 MB', uploadedBy: 'Mohammed Al-Zahrani', uploadedAt: '2025-04-10T08:00:00Z', url: '#' },
    { id: uid(), name: 'SASO Compliance Certificate', docType: 'certificate', fileName: 'HIK-SASO-Cert-2025.pdf', fileSize: '450 KB', uploadedBy: 'Khalid Al-Rashid', uploadedAt: '2025-01-20T14:00:00Z', url: '#' },
    { id: uid(), name: 'Al Futtaim 2026 Price List', docType: 'catalog', fileName: 'AFT-HIK-PriceList-2026.xlsx', fileSize: '2.1 MB', uploadedBy: 'Ahmed bin Saleh', uploadedAt: '2026-01-05T09:00:00Z', url: '#' },
  ],
  'HIK-DS2CD2T87': [
    { id: uid(), name: 'DS-2CD2T87G2-L Datasheet', docType: 'datasheet', fileName: 'DS-2CD2T87G2-L_Datasheet.pdf', fileSize: '1.5 MB', uploadedBy: 'Ahmed bin Saleh', uploadedAt: '2025-06-15T10:00:00Z', url: '#' },
    { id: uid(), name: 'ColorVu Technology Whitepaper', docType: 'other', fileName: 'HIK-ColorVu-Whitepaper.pdf', fileSize: '890 KB', uploadedBy: 'Mohammed Al-Zahrani', uploadedAt: '2025-03-20T11:00:00Z', url: '#' },
  ],
  'AXIS-Q6135LE': [
    { id: uid(), name: 'Q6135-LE Product Sheet', docType: 'datasheet', fileName: 'Q6135-LE-ProductSheet.pdf', fileSize: '980 KB', uploadedBy: 'Ahmed bin Saleh', uploadedAt: '2025-08-10T09:00:00Z', url: '#' },
    { id: uid(), name: 'AXIS Camera Station Manual', docType: 'manual', fileName: 'AXIS-CamStation-Manual.pdf', fileSize: '12.4 MB', uploadedBy: 'Mohammed Al-Zahrani', uploadedAt: '2025-05-01T08:00:00Z', url: '#' },
    { id: uid(), name: 'Norden 2025 Partner Catalog', docType: 'catalog', fileName: 'Norden-Axis-Catalog-2025.pdf', fileSize: '5.6 MB', uploadedBy: 'Khalid Al-Rashid', uploadedAt: '2025-01-15T14:00:00Z', url: '#' },
  ],
  'BOSCH-FPA5000': [
    { id: uid(), name: 'FPA-5000 System Manual', docType: 'manual', fileName: 'Bosch-FPA5000-SysManual.pdf', fileSize: '18.2 MB', uploadedBy: 'Mohammed Al-Zahrani', uploadedAt: '2025-04-01T08:00:00Z', url: '#' },
    { id: uid(), name: 'Fire Panel Compliance Certificate', docType: 'certificate', fileName: 'Bosch-FPA5000-UL-FM.pdf', fileSize: '320 KB', uploadedBy: 'Khalid Al-Rashid', uploadedAt: '2025-02-10T10:00:00Z', url: '#' },
  ],
})

const docTypeConfig: Record<ProductDocType, { label: string; badge: string; icon: string }> = {
  datasheet: { label: 'Datasheet', badge: 'badge-primary', icon: '📄' },
  manual: { label: 'Manual', badge: 'badge-info', icon: '📘' },
  certificate: { label: 'Certificate', badge: 'badge-success', icon: '✅' },
  'vendor-quote': { label: 'Vendor Quote', badge: 'badge-warning', icon: '💰' },
  catalog: { label: 'Catalog', badge: 'badge-gray', icon: '📋' },
  image: { label: 'Image', badge: 'badge-info', icon: '🖼️' },
  other: { label: 'Other', badge: 'badge-gray', icon: '📎' },
}

const priceSourceConfig: Record<PriceSource, { label: string; badge: string }> = {
  'vendor-catalog': { label: 'Vendor Catalog', badge: 'badge-gray' },
  'purchase-order': { label: 'Purchase Order', badge: 'badge-primary' },
  'supplier-quote': { label: 'Supplier Quote', badge: 'badge-info' },
  'goods-receipt': { label: 'Goods Receipt', badge: 'badge-success' },
  manual: { label: 'Manual', badge: 'badge-warning' },
}

// ── Filters ──────────────────────────────────────────────────
const searchQuery = ref('')
const filterMfr = ref('')
const filterType = ref<'' | ProductType>('')
const filterStatus = ref<'' | 'active' | 'inactive'>('')

const filteredProducts = computed(() => {
  let list = products.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) list = list.filter(p => p.sku.toLowerCase().includes(q) || p.name.toLowerCase().includes(q) || p.description.toLowerCase().includes(q))
  if (filterMfr.value) list = list.filter(p => p.manufacturerId === filterMfr.value)
  if (filterType.value) list = list.filter(p => p.productType === filterType.value)
  if (filterStatus.value) list = list.filter(p => filterStatus.value === 'active' ? p.isActive : !p.isActive)
  return list
})

// ── Create/Edit Modal ────────────────────────────────────────
const showModal = ref(false)
const editingId = ref<string | null>(null)

interface ProductForm {
  sku: string; name: string; description: string; manufacturerId: string; categoryId: string
  productType: ProductType; originCurrency: Currency; unitCostOrigin: number; fxRate: number
  freightPercent: number; customsPercent: number; clearancePercent: number
  targetMarginPercent: number; leadTimeDays: number; supplierName: string; isActive: boolean
}

const defaultForm = (): ProductForm => ({
  sku: '', name: '', description: '', manufacturerId: '', categoryId: '',
  productType: 'import', originCurrency: 'USD', unitCostOrigin: 0, fxRate: 3.75,
  freightPercent: 5, customsPercent: 7, clearancePercent: 3,
  targetMarginPercent: 30, leadTimeDays: 21, supplierName: '', isActive: true,
})
const form = ref<ProductForm>(defaultForm())
const availableCategories = computed(() => { const m = mfrList.find(x => x.id === form.value.manufacturerId); return m ? m.categories : [] })
watch(() => form.value.manufacturerId, () => { form.value.categoryId = '' })
const costInSAR = computed(() => form.value.unitCostOrigin * form.value.fxRate)
const landedCostSAR = computed(() => form.value.productType === 'local' ? form.value.unitCostOrigin : calcLanded(form.value.unitCostOrigin, form.value.fxRate, form.value.freightPercent, form.value.customsPercent, form.value.clearancePercent))
const sellingPrice = computed(() => calcSelling(landedCostSAR.value, form.value.targetMarginPercent))
const marginAmount = computed(() => sellingPrice.value - landedCostSAR.value)

function marginClass(pct: number): string {
  if (pct >= 25) return 'text-success'
  if (pct >= 20) return 'text-warning'
  return 'text-danger'
}

function openAddModal() { editingId.value = null; form.value = defaultForm(); showModal.value = true }
function openEditModal(p: Product) {
  editingId.value = p.id
  form.value = { sku: p.sku, name: p.name, description: p.description, manufacturerId: p.manufacturerId, categoryId: p.categoryId, productType: p.productType, originCurrency: p.originCurrency, unitCostOrigin: p.unitCostOrigin, fxRate: p.fxRate, freightPercent: p.freightPercent, customsPercent: p.customsPercent, clearancePercent: p.clearancePercent, targetMarginPercent: p.targetMarginPercent, leadTimeDays: p.leadTimeDays, supplierName: p.supplierName, isActive: p.isActive }
  showModal.value = true
}

function saveProduct() {
  const now = new Date().toISOString()
  const m = mfr(form.value.manufacturerId); const cat = m?.categories.find(c => c.id === form.value.categoryId)
  const base = { sku: form.value.sku, name: form.value.name, description: form.value.description, manufacturerId: form.value.manufacturerId, manufacturerName: m?.name ?? '', categoryId: form.value.categoryId, categoryName: cat?.name ?? '', productType: form.value.productType, originCurrency: form.value.originCurrency, unitCostOrigin: form.value.unitCostOrigin, fxRate: form.value.fxRate, costInSAR: costInSAR.value, freightPercent: form.value.freightPercent, customsPercent: form.value.customsPercent, clearancePercent: form.value.clearancePercent, landedCostSAR: landedCostSAR.value, targetMarginPercent: form.value.targetMarginPercent, sellingPrice: sellingPrice.value, marginAmount: marginAmount.value, leadTimeDays: form.value.leadTimeDays, supplierName: form.value.supplierName, isActive: form.value.isActive }
  if (editingId.value) { const idx = products.value.findIndex(p => p.id === editingId.value); if (idx !== -1) products.value[idx] = { ...products.value[idx], ...base, updatedAt: now } as Product }
  else products.value.push({ id: uid(), ...base, createdAt: now, updatedAt: now })
  showModal.value = false
}
function deleteProduct(id: string) { products.value = products.value.filter(p => p.id !== id) }
const currencies: Currency[] = ['USD', 'EUR', 'GBP', 'AED', 'CNY', 'SAR']

// ── Product Detail Modal ─────────────────────────────────────
const showDetailModal = ref(false)
const detailProduct = ref<Product | null>(null)
type DetailTab = 'overview' | 'vendors' | 'price-history' | 'documents'
const detailTab = ref<DetailTab>('overview')

function openDetail(p: Product) { detailProduct.value = p; detailTab.value = 'overview'; showDetailModal.value = true }

const detailVendors = computed<ProductVendorEntry[]>(() => detailProduct.value ? (vendorCatalogs.value[detailProduct.value.sku] || []) : [])
const detailPriceHistory = computed<ProductPriceRecord[]>(() => detailProduct.value ? (priceHistory.value[detailProduct.value.sku] || []).sort((a, b) => b.date.localeCompare(a.date)) : [])
const detailDocs = computed<ProductDocument[]>(() => detailProduct.value ? (productDocs.value[detailProduct.value.sku] || []) : [])
const detailSupplierItems = computed<SupplierItemEntry[]>(() => {
  if (!detailProduct.value) return []
  return procStore.supplierItems.filter(si => si.productSku === detailProduct.value!.sku)
})

// Add vendor entry
const showAddVendorModal = ref(false)
const vendorForm = ref({ vendorName: '', vendorSku: '', unitCost: 0, currency: 'USD' as Currency, moq: 1, leadTimeDays: 21, catalogSource: '' })
function openAddVendor() { vendorForm.value = { vendorName: '', vendorSku: '', unitCost: 0, currency: 'USD', moq: 1, leadTimeDays: 21, catalogSource: '' }; showAddVendorModal.value = true }
function saveVendorEntry() {
  if (!detailProduct.value || !vendorForm.value.vendorName) return
  const sku = detailProduct.value.sku
  if (!vendorCatalogs.value[sku]) vendorCatalogs.value[sku] = []
  vendorCatalogs.value[sku].push({ id: uid(), vendorName: vendorForm.value.vendorName, vendorSku: vendorForm.value.vendorSku || undefined, unitCost: vendorForm.value.unitCost, currency: vendorForm.value.currency, moq: vendorForm.value.moq, leadTimeDays: vendorForm.value.leadTimeDays, lastQuoteDate: new Date().toISOString().slice(0, 10), catalogSource: vendorForm.value.catalogSource || undefined })
  showAddVendorModal.value = false
}
function removeVendorEntry(id: string) {
  if (!detailProduct.value) return
  const sku = detailProduct.value.sku
  if (vendorCatalogs.value[sku]) vendorCatalogs.value[sku] = vendorCatalogs.value[sku].filter(v => v.id !== id)
}

// ── Upload Catalog ───────────────────────────────────────────
const showUploadCatalogModal = ref(false)
const catalogFileRef = ref<HTMLInputElement | null>(null)
const catalogFileName = ref('')
const catalogFileType = ref<FileType | ''>('')
const catalogParseError = ref('')
const catalogVendorName = ref('')
const catalogCurrency = ref<Currency>('USD')
const catalogParsing = ref(false)

interface CatalogPreviewRow {
  vendorSku: string; productName: string; unitCost: number; moq: number; leadTimeDays: number
  selected: boolean
}
const catalogPreviewRows = ref<CatalogPreviewRow[]>([])

function openUploadCatalog() {
  catalogFileName.value = ''
  catalogFileType.value = ''
  catalogParseError.value = ''
  catalogVendorName.value = ''
  catalogCurrency.value = 'USD'
  catalogPreviewRows.value = []
  catalogParsing.value = false
  catalogPasteText.value = ''
  showUploadCatalogModal.value = true
}

function triggerCatalogUpload() { catalogFileRef.value?.click() }

function applyCatalogParseResult(result: ParseResult) {
  const { headers, rows, mappedColumns } = result
  if (mappedColumns.name === undefined && mappedColumns.sku === undefined) {
    catalogParseError.value = 'Could not detect SKU or Name/Description columns. Check the file headers.'
    return
  }
  catalogPreviewRows.value = []
  for (const row of rows) {
    const sku = getMappedValue(row, 'sku', headers, mappedColumns)
    const name = getMappedValue(row, 'name', headers, mappedColumns) || sku
    const cost = getMappedNumber(row, 'unitCost', headers, mappedColumns, 0)
    if (!name && !cost) continue
    catalogPreviewRows.value.push({
      vendorSku: sku,
      productName: name,
      unitCost: cost,
      moq: getMappedInt(row, 'moq', headers, mappedColumns, 1),
      leadTimeDays: getMappedInt(row, 'leadTimeDays', headers, mappedColumns, 21),
      selected: true,
    })
  }
  if (catalogPreviewRows.value.length === 0) { catalogParseError.value = 'No data rows could be parsed from the file.'; return }
  catalogParseError.value = ''
}

async function handleCatalogFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  await readCatalogFile(file)
  input.value = ''
}

async function readCatalogFile(file?: File) {
  if (!file) return
  catalogFileName.value = file.name
  catalogParseError.value = ''
  catalogPreviewRows.value = []
  catalogPasteText.value = ''
  catalogParsing.value = true

  const out = await parseFile(file)
  catalogParsing.value = false
  if ('error' in out) { catalogParseError.value = out.error; return }
  catalogFileType.value = out.fileType
  applyCatalogParseResult(out.result)
}

const catalogPasteText = ref('')
function parsePastedText() {
  if (!catalogPasteText.value.trim()) return
  const result = parseCSVText(catalogPasteText.value)
  if (result.rows.length === 0) { catalogParseError.value = 'No data could be parsed from pasted text.'; return }
  applyCatalogParseResult(result)
}

const catalogSelectedCount = computed(() => catalogPreviewRows.value.filter(r => r.selected).length)

function importCatalogItems() {
  if (!detailProduct.value || !catalogVendorName.value) return
  const sku = detailProduct.value.sku
  if (!vendorCatalogs.value[sku]) vendorCatalogs.value[sku] = []
  const selected = catalogPreviewRows.value.filter(r => r.selected)
  for (const row of selected) {
    vendorCatalogs.value[sku].push({
      id: uid(), vendorName: catalogVendorName.value, vendorSku: row.vendorSku || undefined,
      unitCost: row.unitCost, currency: catalogCurrency.value,
      moq: row.moq, leadTimeDays: row.leadTimeDays,
      lastQuoteDate: new Date().toISOString().slice(0, 10),
      catalogSource: catalogFileName.value || 'Pasted data',
    })
  }
  if (!priceHistory.value[sku]) priceHistory.value[sku] = []
  for (const row of selected) {
    priceHistory.value[sku].push({
      id: uid(), date: new Date().toISOString().slice(0, 10),
      source: 'vendor-catalog', vendorName: catalogVendorName.value,
      unitCost: row.unitCost, currency: catalogCurrency.value,
      qty: 0, notes: `Imported from ${catalogFileName.value || 'pasted catalog'}`,
    })
  }
  showUploadCatalogModal.value = false
}

// Add document
const showAddDocModal = ref(false)
const docForm = ref({ name: '', docType: 'datasheet' as ProductDocType, fileName: '', fileSize: '', notes: '' })
function openAddDoc() { docForm.value = { name: '', docType: 'datasheet', fileName: '', fileSize: '', notes: '' }; showAddDocModal.value = true }
function saveDoc() {
  if (!detailProduct.value || !docForm.value.name) return
  const sku = detailProduct.value.sku
  if (!productDocs.value[sku]) productDocs.value[sku] = []
  productDocs.value[sku].push({ id: uid(), name: docForm.value.name, docType: docForm.value.docType, fileName: docForm.value.fileName || `${docForm.value.name.replace(/\s+/g, '_')}.pdf`, fileSize: docForm.value.fileSize || '—', uploadedBy: 'Current User', uploadedAt: new Date().toISOString(), notes: docForm.value.notes || undefined })
  showAddDocModal.value = false
}
function removeDoc(id: string) {
  if (!detailProduct.value) return
  const sku = detailProduct.value.sku
  if (productDocs.value[sku]) productDocs.value[sku] = productDocs.value[sku].filter(d => d.id !== id)
}

// Add price record
const showAddPriceModal = ref(false)
const priceForm = ref({ date: '', source: 'manual' as PriceSource, sourceRef: '', vendorName: '', unitCost: 0, currency: 'SAR' as Currency, qty: 0, notes: '' })
function openAddPrice() { priceForm.value = { date: new Date().toISOString().slice(0, 10), source: 'manual', sourceRef: '', vendorName: '', unitCost: 0, currency: 'SAR', qty: 0, notes: '' }; showAddPriceModal.value = true }
function savePrice() {
  if (!detailProduct.value || !priceForm.value.vendorName) return
  const sku = detailProduct.value.sku
  if (!priceHistory.value[sku]) priceHistory.value[sku] = []
  priceHistory.value[sku].push({ id: uid(), date: priceForm.value.date, source: priceForm.value.source, sourceRef: priceForm.value.sourceRef || undefined, vendorName: priceForm.value.vendorName, unitCost: priceForm.value.unitCost, currency: priceForm.value.currency, qty: priceForm.value.qty, notes: priceForm.value.notes || undefined })
  showAddPriceModal.value = false
}
</script>

<template>
  <div class="products-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Products</h1>
        <p class="page-header-subtitle">{{ filteredProducts.length }} product{{ filteredProducts.length !== 1 ? 's' : '' }} in catalog</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal"><Plus :size="18" /> Add Product</button>
    </div>

    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search"><Search :size="18" class="search-icon" /><input v-model="searchQuery" type="text" class="form-input" placeholder="Search by SKU, name, or description..." /></div>
        <select v-model="filterMfr" class="form-select toolbar-select"><option value="">All Manufacturers</option><option v-for="m in mfrList" :key="m.id" :value="m.id">{{ m.name }}</option></select>
        <select v-model="filterType" class="form-select toolbar-select"><option value="">All Types</option><option value="import">Import</option><option value="local">Local</option></select>
        <select v-model="filterStatus" class="form-select toolbar-select"><option value="">All Statuses</option><option value="active">Active</option><option value="inactive">Inactive</option></select>
      </div>
    </div>

    <div v-if="filteredProducts.length" class="table-container">
      <table class="table">
        <thead><tr><th>SKU</th><th>Product Name</th><th>Manufacturer</th><th>Type</th><th class="text-right">Landed Cost</th><th class="text-right">Selling Price</th><th class="text-right">Margin</th><th class="text-center">Lead</th><th class="text-center">Vendors</th><th class="text-center">Docs</th><th>Status</th><th></th></tr></thead>
        <tbody>
          <tr v-for="p in filteredProducts" :key="p.id" class="prod-row" @click="openDetail(p)">
            <td class="text-mono font-medium">{{ p.sku }}</td>
            <td><span class="font-medium text-dark">{{ p.name }}</span><div class="text-xs text-muted">{{ p.description }}</div></td>
            <td><div>{{ p.manufacturerName }}</div><div class="text-xs text-muted">{{ p.categoryName }}</div></td>
            <td><span :class="['badge', p.productType === 'import' ? 'badge-primary' : 'badge-success']">{{ p.productType === 'import' ? 'Import' : 'Local' }}</span></td>
            <td class="text-right whitespace-nowrap">SAR {{ formatSAR(p.landedCostSAR) }}</td>
            <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(p.sellingPrice) }}</td>
            <td class="text-right"><span :class="['font-semibold', marginClass(p.targetMarginPercent)]">{{ p.targetMarginPercent.toFixed(1) }}%</span></td>
            <td class="text-center">{{ p.leadTimeDays }}d</td>
            <td class="text-center"><span v-if="(vendorCatalogs[p.sku] || []).length" class="badge badge-info">{{ (vendorCatalogs[p.sku]?.length ?? 0) }}</span><span v-else class="text-muted">—</span></td>
            <td class="text-center"><span v-if="(productDocs[p.sku] || []).length" class="badge badge-gray">{{ (productDocs[p.sku]?.length ?? 0) }}</span><span v-else class="text-muted">—</span></td>
            <td><span :class="['badge badge-dot', p.isActive ? 'badge-success' : 'badge-danger']">{{ p.isActive ? 'Active' : 'Inactive' }}</span></td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="View" @click.stop="openDetail(p)"><Eye :size="14" /></button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Edit" @click.stop="openEditModal(p)"><Pencil :size="14" /></button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click.stop="deleteProduct(p.id)"><Trash2 :size="14" /></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="empty-state"><Package :size="48" class="empty-state-icon" /><h3 class="empty-state-title">No products found</h3><p class="empty-state-text">{{ searchQuery || filterMfr || filterType || filterStatus ? 'Try adjusting your filters.' : 'Add your first product to get started.' }}</p></div>

    <!-- ═══ Product Detail Modal ══════════════════════════════ -->
    <div v-if="showDetailModal && detailProduct" class="modal-backdrop" @click.self="showDetailModal = false">
      <div class="modal prod-detail-modal">
        <div class="prod-detail-header">
          <div class="prod-detail-header-left">
            <div class="prod-detail-icon"><Package :size="24" /></div>
            <div>
              <h2 class="prod-detail-title">{{ detailProduct.name }}</h2>
              <div class="prod-detail-meta">
                <span class="prod-detail-sku">{{ detailProduct.sku }}</span><span class="prod-sep">|</span>
                <span>{{ detailProduct.manufacturerName }}</span><span class="prod-sep">|</span>
                <span class="badge badge-gray" style="font-size:0.7rem">{{ detailProduct.categoryName }}</span>
                <span :class="['badge badge-dot', detailProduct.isActive ? 'badge-success' : 'badge-danger']" style="font-size:0.7rem">{{ detailProduct.isActive ? 'Active' : 'Inactive' }}</span>
              </div>
            </div>
          </div>
          <div class="prod-detail-header-right">
            <div class="prod-hdr-num"><span class="prod-hdr-val">SAR {{ formatSAR(detailProduct.landedCostSAR) }}</span><span class="prod-hdr-lbl">Landed</span></div>
            <div class="prod-hdr-num"><span class="prod-hdr-val">SAR {{ formatSAR(detailProduct.sellingPrice) }}</span><span class="prod-hdr-lbl">Selling</span></div>
            <div class="prod-hdr-num"><span :class="['prod-hdr-val', marginClass(detailProduct.targetMarginPercent)]">{{ detailProduct.targetMarginPercent.toFixed(1) }}%</span><span class="prod-hdr-lbl">Margin</span></div>
            <button class="modal-close" @click="showDetailModal = false"><X :size="20" /></button>
          </div>
        </div>

        <div class="prod-detail-tabs">
          <button :class="['prod-dtab', detailTab === 'overview' && 'prod-dtab--active']" @click="detailTab = 'overview'"><Package :size="15" /> Overview</button>
          <button :class="['prod-dtab', detailTab === 'vendors' && 'prod-dtab--active']" @click="detailTab = 'vendors'"><Building2 :size="15" /> Vendor Catalog <span class="prod-dtab-ct">{{ detailVendors.length }}</span></button>
          <button :class="['prod-dtab', detailTab === 'price-history' && 'prod-dtab--active']" @click="detailTab = 'price-history'"><History :size="15" /> Price History <span class="prod-dtab-ct">{{ detailPriceHistory.length }}</span></button>
          <button :class="['prod-dtab', detailTab === 'documents' && 'prod-dtab--active']" @click="detailTab = 'documents'"><FileText :size="15" /> Documents <span class="prod-dtab-ct">{{ detailDocs.length }}</span></button>
        </div>

        <div class="modal-body prod-detail-body">
          <!-- OVERVIEW -->
          <div v-if="detailTab === 'overview'">
            <div class="prod-info-grid">
              <div class="prod-info-card"><span class="prod-info-lbl">SKU</span><span class="prod-info-val text-mono">{{ detailProduct.sku }}</span></div>
              <div class="prod-info-card"><span class="prod-info-lbl">Manufacturer</span><span class="prod-info-val">{{ detailProduct.manufacturerName }}</span></div>
              <div class="prod-info-card"><span class="prod-info-lbl">Category</span><span class="prod-info-val">{{ detailProduct.categoryName }}</span></div>
              <div class="prod-info-card"><span class="prod-info-lbl">Type</span><span class="prod-info-val">{{ detailProduct.productType === 'import' ? 'Import' : 'Local' }}</span></div>
              <div class="prod-info-card"><span class="prod-info-lbl">Lead Time</span><span class="prod-info-val">{{ detailProduct.leadTimeDays }} days</span></div>
              <div class="prod-info-card"><span class="prod-info-lbl">Supplier</span><span class="prod-info-val">{{ detailProduct.supplierName }}</span></div>
            </div>
            <p class="text-muted" style="margin: var(--space-4) 0; font-size:0.85rem">{{ detailProduct.description }}</p>
            <h4 class="prod-section-title"><Calculator :size="14" /> Cost Breakdown</h4>
            <div class="prod-cost-grid">
              <div class="prod-cost-item"><span class="prod-cost-lbl">Origin Cost</span><span class="prod-cost-val">{{ formatCurrency(detailProduct.unitCostOrigin, detailProduct.originCurrency) }}</span></div>
              <div v-if="detailProduct.productType === 'import'" class="prod-cost-item"><span class="prod-cost-lbl">FX Rate</span><span class="prod-cost-val">{{ detailProduct.fxRate }}</span></div>
              <div v-if="detailProduct.productType === 'import'" class="prod-cost-item"><span class="prod-cost-lbl">Cost in SAR</span><span class="prod-cost-val">SAR {{ formatSAR(detailProduct.costInSAR) }}</span></div>
              <div v-if="detailProduct.productType === 'import'" class="prod-cost-item"><span class="prod-cost-lbl">Freight</span><span class="prod-cost-val">{{ detailProduct.freightPercent }}%</span></div>
              <div v-if="detailProduct.productType === 'import'" class="prod-cost-item"><span class="prod-cost-lbl">Customs</span><span class="prod-cost-val">{{ detailProduct.customsPercent }}%</span></div>
              <div v-if="detailProduct.productType === 'import'" class="prod-cost-item"><span class="prod-cost-lbl">Clearance</span><span class="prod-cost-val">{{ detailProduct.clearancePercent }}%</span></div>
              <div class="prod-cost-item prod-cost-item--highlight"><span class="prod-cost-lbl">Landed Cost</span><span class="prod-cost-val font-bold">SAR {{ formatSAR(detailProduct.landedCostSAR) }}</span></div>
              <div class="prod-cost-item"><span class="prod-cost-lbl">Target Margin</span><span :class="['prod-cost-val font-bold', marginClass(detailProduct.targetMarginPercent)]">{{ detailProduct.targetMarginPercent.toFixed(1) }}%</span></div>
              <div class="prod-cost-item prod-cost-item--highlight"><span class="prod-cost-lbl">Selling Price</span><span class="prod-cost-val font-bold" style="color:var(--color-primary)">SAR {{ formatSAR(detailProduct.sellingPrice) }}</span></div>
              <div class="prod-cost-item"><span class="prod-cost-lbl">Margin Amount</span><span class="prod-cost-val font-bold text-success">SAR {{ formatSAR(detailProduct.marginAmount) }}</span></div>
            </div>

            <div v-if="detailSupplierItems.length">
              <h4 class="prod-section-title"><Building2 :size="14" /> Supplier Catalog Entries</h4>
              <div class="table-container table-container--embedded">
                <table class="table"><thead><tr><th>Supplier</th><th class="text-right">Latest Cost</th><th class="text-center">Trend</th><th class="text-center">MOQ</th><th class="text-center">Lead</th><th>Reliability</th></tr></thead>
                <tbody><tr v-for="si in detailSupplierItems" :key="si.id">
                  <td class="font-medium">{{ si.supplierName }}</td>
                  <td class="text-right font-medium">SAR {{ formatSAR(si.latestCost) }}</td>
                  <td class="text-center"><TrendingDown v-if="si.costTrend === 'down'" :size="14" class="text-success" /><TrendingUp v-else-if="si.costTrend === 'up'" :size="14" class="text-danger" /><Minus v-else :size="14" class="text-muted" /></td>
                  <td class="text-center">{{ si.moq }}</td><td class="text-center">{{ si.leadTimeDays }}d</td>
                  <td><div class="reliability-wrap"><div class="reliability-bar"><div class="reliability-fill" :class="si.reliability >= 90 ? 'bg-green' : si.reliability >= 80 ? 'bg-orange' : 'bg-red'" :style="{ width: si.reliability + '%' }" /></div><span class="reliability-lbl">{{ si.reliability }}%</span></div></td>
                </tr></tbody></table>
              </div>
            </div>
          </div>

          <!-- VENDORS -->
          <div v-if="detailTab === 'vendors'">
            <div class="prod-tab-toolbar"><span class="font-medium">{{ detailVendors.length }} vendor{{ detailVendors.length !== 1 ? 's' : '' }} in catalog</span><div class="prod-tab-actions"><button class="btn btn-secondary btn-sm" @click="openUploadCatalog"><FileUp :size="14" /> Upload Catalog</button><button class="btn btn-primary btn-sm" @click="openAddVendor"><Plus :size="14" /> Add Vendor</button></div></div>
            <div v-if="detailVendors.length" class="table-container table-container--embedded">
              <table class="table"><thead><tr><th>Vendor</th><th>Vendor SKU</th><th class="text-right">Unit Cost</th><th class="text-center">MOQ</th><th class="text-center">Lead Time</th><th>Last Quote</th><th>Source</th><th></th></tr></thead>
              <tbody><tr v-for="v in detailVendors" :key="v.id">
                <td class="font-bold">{{ v.vendorName }}</td>
                <td class="text-mono text-muted" style="font-size:0.75rem">{{ v.vendorSku || '—' }}</td>
                <td class="text-right whitespace-nowrap font-medium">{{ formatCurrency(v.unitCost, v.currency) }}</td>
                <td class="text-center">{{ v.moq }}</td><td class="text-center">{{ v.leadTimeDays }}d</td>
                <td class="whitespace-nowrap text-muted" style="font-size:0.75rem">{{ formatDate(v.lastQuoteDate) }}</td>
                <td class="text-muted" style="font-size:0.72rem; max-width:160px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap">{{ v.catalogSource || '—' }}</td>
                <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeVendorEntry(v.id)"><Trash2 :size="13" /></button></td>
              </tr></tbody></table>
            </div>
            <div v-else class="prod-empty-tab"><Building2 :size="36" class="text-muted" /><p class="text-muted">No vendor catalog entries yet.</p><div class="prod-tab-actions"><button class="btn btn-secondary btn-sm" @click="openUploadCatalog"><FileUp :size="14" /> Upload Catalog</button><button class="btn btn-primary btn-sm" @click="openAddVendor"><Plus :size="14" /> Add Vendor</button></div></div>
          </div>

          <!-- PRICE HISTORY -->
          <div v-if="detailTab === 'price-history'">
            <div class="prod-tab-toolbar"><span class="font-medium">{{ detailPriceHistory.length }} price record{{ detailPriceHistory.length !== 1 ? 's' : '' }}</span><button class="btn btn-primary btn-sm" @click="openAddPrice"><Plus :size="14" /> Add Price</button></div>
            <div v-if="detailPriceHistory.length" class="table-container table-container--embedded">
              <table class="table"><thead><tr><th>Date</th><th>Source</th><th>Reference</th><th>Vendor</th><th class="text-right">Unit Cost</th><th class="text-right">Landing</th><th class="text-center">Qty</th><th>Notes</th></tr></thead>
              <tbody><tr v-for="pr in detailPriceHistory" :key="pr.id">
                <td class="whitespace-nowrap" style="font-size:0.78rem">{{ formatDate(pr.date) }}</td>
                <td><span :class="['badge', priceSourceConfig[pr.source].badge]">{{ priceSourceConfig[pr.source].label }}</span></td>
                <td class="text-mono" style="font-size:0.72rem">{{ pr.sourceRef || '—' }}</td>
                <td class="font-medium" style="font-size:0.78rem">{{ pr.vendorName }}</td>
                <td class="text-right whitespace-nowrap font-medium">{{ formatCurrency(pr.unitCost, pr.currency) }}</td>
                <td class="text-right whitespace-nowrap"><span v-if="pr.landingCost" class="font-medium">SAR {{ formatSAR(pr.landingCost) }}</span><span v-else class="text-muted">—</span></td>
                <td class="text-center">{{ pr.qty || '—' }}</td>
                <td class="text-muted" style="font-size:0.72rem; max-width:180px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap">{{ pr.notes || '—' }}</td>
              </tr></tbody></table>
            </div>
            <div v-else class="prod-empty-tab"><History :size="36" class="text-muted" /><p class="text-muted">No price history records yet.</p><button class="btn btn-primary btn-sm" @click="openAddPrice"><Plus :size="14" /> Add Price Record</button></div>
          </div>

          <!-- DOCUMENTS -->
          <div v-if="detailTab === 'documents'">
            <div class="prod-tab-toolbar"><span class="font-medium">{{ detailDocs.length }} document{{ detailDocs.length !== 1 ? 's' : '' }}</span><button class="btn btn-primary btn-sm" @click="openAddDoc"><Upload :size="14" /> Add Document</button></div>
            <div v-if="detailDocs.length" class="prod-docs-grid">
              <div v-for="doc in detailDocs" :key="doc.id" class="prod-doc-card">
                <div class="prod-doc-card-top">
                  <span :class="['badge', docTypeConfig[doc.docType].badge]" style="font-size:0.65rem">{{ docTypeConfig[doc.docType].label }}</span>
                  <button class="btn btn-ghost btn-icon btn-sm" @click="removeDoc(doc.id)"><Trash2 :size="12" /></button>
                </div>
                <div class="prod-doc-name">{{ doc.name }}</div>
                <div class="prod-doc-meta">
                  <span class="text-mono">{{ doc.fileName }}</span>
                  <span>{{ doc.fileSize }}</span>
                </div>
                <div class="prod-doc-footer">
                  <span>{{ doc.uploadedBy }} &middot; {{ formatDate(doc.uploadedAt) }}</span>
                  <button class="btn btn-ghost btn-sm" style="font-size:0.68rem; gap:3px"><Download :size="12" /> Download</button>
                </div>
              </div>
            </div>
            <div v-else class="prod-empty-tab"><FileText :size="36" class="text-muted" /><p class="text-muted">No documents attached yet.</p><button class="btn btn-primary btn-sm" @click="openAddDoc"><Upload :size="14" /> Add Document</button></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ═══ Add Vendor Modal ══════════════════════════════════ -->
    <div v-if="showAddVendorModal" class="modal-backdrop" @click.self="showAddVendorModal = false">
      <div class="modal modal-lg"><div class="modal-header"><h2 class="modal-title"><Building2 :size="18" /> Add Vendor Entry</h2><button class="modal-close" @click="showAddVendorModal = false"><X :size="20" /></button></div>
      <div class="modal-body">
        <div class="form-row"><div class="form-group"><label class="form-label">Vendor Name <span class="required">*</span></label><input v-model="vendorForm.vendorName" type="text" class="form-input" placeholder="e.g. Al Futtaim Trading" /></div><div class="form-group"><label class="form-label">Vendor SKU</label><input v-model="vendorForm.vendorSku" type="text" class="form-input" placeholder="Vendor's part number" /></div></div>
        <div class="form-row"><div class="form-group"><label class="form-label">Unit Cost</label><input v-model.number="vendorForm.unitCost" type="number" step="0.01" class="form-input" /></div><div class="form-group"><label class="form-label">Currency</label><select v-model="vendorForm.currency" class="form-select"><option v-for="c in currencies" :key="c" :value="c">{{ c }}</option></select></div><div class="form-group"><label class="form-label">MOQ</label><input v-model.number="vendorForm.moq" type="number" min="1" class="form-input" /></div></div>
        <div class="form-row"><div class="form-group"><label class="form-label">Lead Time (days)</label><input v-model.number="vendorForm.leadTimeDays" type="number" min="1" class="form-input" /></div><div class="form-group"><label class="form-label">Catalog / Source</label><input v-model="vendorForm.catalogSource" type="text" class="form-input" placeholder="e.g. Annual Price List 2026" /></div></div>
      </div>
      <div class="modal-footer"><button class="btn btn-secondary" @click="showAddVendorModal = false">Cancel</button><button class="btn btn-primary" :disabled="!vendorForm.vendorName" @click="saveVendorEntry">Add Vendor</button></div></div>
    </div>

    <!-- ═══ Upload Catalog Modal ══════════════════════════════ -->
    <div v-if="showUploadCatalogModal" class="modal-backdrop" @click.self="showUploadCatalogModal = false">
      <div class="modal modal-xxl">
        <div class="modal-header">
          <h2 class="modal-title"><FileUp :size="20" /> Upload Vendor Catalog</h2>
          <button class="modal-close" @click="showUploadCatalogModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body" style="max-height: 70vh; overflow-y: auto">
          <!-- Vendor info row -->
          <div class="cat-vendor-row">
            <div class="form-group" style="flex:2">
              <label class="form-label">Vendor Name <span class="required">*</span></label>
              <input v-model="catalogVendorName" type="text" class="form-input" placeholder="e.g. Al Futtaim Trading" />
            </div>
            <div class="form-group" style="flex:1">
              <label class="form-label">Currency</label>
              <select v-model="catalogCurrency" class="form-select">
                <option v-for="c in currencies" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
          </div>

          <!-- Upload area -->
          <div class="cat-upload-zone" @click="triggerCatalogUpload" @dragover.prevent @drop.prevent="readCatalogFile($event.dataTransfer?.files[0])">
            <input ref="catalogFileRef" type="file" accept=".csv,.tsv,.txt,.xlsx,.xls,.pdf" style="display:none" @change="handleCatalogFile" />
            <div v-if="!catalogFileName" class="cat-upload-placeholder">
              <Upload :size="32" class="text-muted" />
              <p class="cat-upload-title">Drop file here or click to browse</p>
              <p class="cat-upload-subtitle">Supports CSV, Excel (.xlsx), TSV, PDF</p>
              <div class="cat-upload-types">
                <span class="cat-type-badge"><Table2 :size="12" /> CSV / TSV</span>
                <span class="cat-type-badge"><FileSpreadsheet :size="12" /> Excel</span>
                <span class="cat-type-badge"><FileText :size="12" /> PDF</span>
              </div>
              <button class="btn-download-template" @click.stop="downloadTemplate('product-catalog')"><Download :size="12" /> Download Excel Template</button>
            </div>
            <div v-else class="cat-upload-selected">
              <Check :size="18" class="text-success" />
              <span class="font-medium">{{ catalogFileName }}</span>
              <span class="badge badge-info" style="font-size:0.65rem">{{ catalogFileType?.toUpperCase() }}</span>
              <button class="btn btn-ghost btn-sm" @click.stop="catalogFileName = ''; catalogFileType = ''; catalogPreviewRows = []; catalogPasteText = ''; catalogParseError = ''" style="margin-left:auto"><X :size="14" /> Remove</button>
            </div>
          </div>

          <!-- Parsing spinner -->
          <div v-if="catalogParsing" class="cat-parsing-indicator">
            <RefreshCw :size="16" class="cat-spin-icon" /> Parsing file...
          </div>

          <!-- Parse error -->
          <div v-if="catalogParseError && !catalogParsing" class="cat-parse-error">
            <AlertTriangle :size="16" /> {{ catalogParseError }}
          </div>

          <!-- Paste fallback when parsing found nothing -->
          <div v-if="catalogFileName && !catalogParsing && catalogPreviewRows.length === 0 && !catalogParseError" class="cat-pdf-section">
            <div class="cat-pdf-notice">
              <AlertTriangle :size="16" />
              <span>Could not extract structured data. Try pasting the content below, or <a href="#" class="fallback-template-link" @click.prevent.stop="downloadTemplate('product-catalog')">download our Excel template</a> and re-upload.</span>
            </div>
            <textarea v-model="catalogPasteText" class="form-input cat-paste-area" placeholder="Paste table data here (tab-separated or comma-separated)&#10;&#10;Example:&#10;SKU, Description, Price, MOQ&#10;CAM-001, IP Camera 4MP, 245.00, 10&#10;CAM-002, IP Camera 8MP, 385.00, 5" rows="8"></textarea>
            <button class="btn btn-primary btn-sm" :disabled="!catalogPasteText.trim()" @click="parsePastedText" style="margin-top:8px"><Table2 :size="14" /> Parse Pasted Data</button>
          </div>

          <!-- Preview table -->
          <div v-if="catalogPreviewRows.length > 0" class="cat-preview-section">
            <div class="cat-preview-header">
              <span class="font-medium">{{ catalogPreviewRows.length }} items found</span>
              <span class="text-muted" style="font-size:0.75rem">{{ catalogSelectedCount }} selected for import</span>
              <div style="margin-left:auto; display:flex; gap:8px">
                <button class="btn btn-ghost btn-sm" @click="catalogPreviewRows.forEach(r => r.selected = true)">Select All</button>
                <button class="btn btn-ghost btn-sm" @click="catalogPreviewRows.forEach(r => r.selected = false)">Deselect All</button>
              </div>
            </div>
            <div class="table-container table-container--embedded" style="max-height:340px; overflow-y:auto">
              <table class="table">
                <thead>
                  <tr>
                    <th style="width:40px"><input type="checkbox" :checked="catalogSelectedCount === catalogPreviewRows.length" @change="catalogPreviewRows.forEach(r => r.selected = ($event.target as HTMLInputElement).checked)" /></th>
                    <th>Vendor SKU</th>
                    <th>Product / Description</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-center">MOQ</th>
                    <th class="text-center">Lead (days)</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in catalogPreviewRows" :key="idx" :class="{ 'cat-row--deselected': !row.selected }">
                    <td><input type="checkbox" v-model="row.selected" /></td>
                    <td class="text-mono" style="font-size:0.75rem"><input v-model="row.vendorSku" type="text" class="form-input form-input--mini" /></td>
                    <td><input v-model="row.productName" type="text" class="form-input form-input--mini" /></td>
                    <td class="text-right"><input v-model.number="row.unitCost" type="number" step="0.01" class="form-input form-input--mini" style="text-align:right; width:100px" /></td>
                    <td class="text-center"><input v-model.number="row.moq" type="number" min="1" class="form-input form-input--mini" style="text-align:center; width:70px" /></td>
                    <td class="text-center"><input v-model.number="row.leadTimeDays" type="number" min="1" class="form-input form-input--mini" style="text-align:center; width:70px" /></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showUploadCatalogModal = false">Cancel</button>
          <button class="btn btn-primary" :disabled="!catalogVendorName || catalogSelectedCount === 0" @click="importCatalogItems">
            <Check :size="16" /> Import {{ catalogSelectedCount }} Item{{ catalogSelectedCount !== 1 ? 's' : '' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ═══ Add Document Modal ════════════════════════════════ -->
    <div v-if="showAddDocModal" class="modal-backdrop" @click.self="showAddDocModal = false">
      <div class="modal modal-lg"><div class="modal-header"><h2 class="modal-title"><Upload :size="18" /> Add Document</h2><button class="modal-close" @click="showAddDocModal = false"><X :size="20" /></button></div>
      <div class="modal-body">
        <div class="form-row"><div class="form-group" style="flex:2"><label class="form-label">Document Name <span class="required">*</span></label><input v-model="docForm.name" type="text" class="form-input" placeholder="e.g. Product Datasheet" /></div><div class="form-group" style="flex:1"><label class="form-label">Type</label><select v-model="docForm.docType" class="form-select"><option value="datasheet">Datasheet</option><option value="manual">Manual</option><option value="certificate">Certificate</option><option value="vendor-quote">Vendor Quote</option><option value="catalog">Catalog</option><option value="image">Image</option><option value="other">Other</option></select></div></div>
        <div class="form-row"><div class="form-group"><label class="form-label">File Name</label><input v-model="docForm.fileName" type="text" class="form-input" placeholder="e.g. Product-Datasheet.pdf" /></div><div class="form-group" style="width:120px"><label class="form-label">File Size</label><input v-model="docForm.fileSize" type="text" class="form-input" placeholder="e.g. 2.5 MB" /></div></div>
        <div class="form-group"><label class="form-label">Notes</label><input v-model="docForm.notes" type="text" class="form-input" placeholder="Optional notes" /></div>
      </div>
      <div class="modal-footer"><button class="btn btn-secondary" @click="showAddDocModal = false">Cancel</button><button class="btn btn-primary" :disabled="!docForm.name" @click="saveDoc">Add Document</button></div></div>
    </div>

    <!-- ═══ Add Price Record Modal ════════════════════════════ -->
    <div v-if="showAddPriceModal" class="modal-backdrop" @click.self="showAddPriceModal = false">
      <div class="modal modal-lg"><div class="modal-header"><h2 class="modal-title"><DollarSign :size="18" /> Add Price Record</h2><button class="modal-close" @click="showAddPriceModal = false"><X :size="20" /></button></div>
      <div class="modal-body">
        <div class="form-row"><div class="form-group"><label class="form-label">Date</label><input v-model="priceForm.date" type="date" class="form-input" /></div><div class="form-group"><label class="form-label">Source</label><select v-model="priceForm.source" class="form-select"><option value="manual">Manual</option><option value="vendor-catalog">Vendor Catalog</option><option value="purchase-order">Purchase Order</option><option value="supplier-quote">Supplier Quote</option><option value="goods-receipt">Goods Receipt</option></select></div><div class="form-group"><label class="form-label">Reference</label><input v-model="priceForm.sourceRef" type="text" class="form-input" placeholder="e.g. PO-2026-0012" /></div></div>
        <div class="form-row"><div class="form-group"><label class="form-label">Vendor <span class="required">*</span></label><input v-model="priceForm.vendorName" type="text" class="form-input" placeholder="Vendor name" /></div><div class="form-group"><label class="form-label">Unit Cost</label><input v-model.number="priceForm.unitCost" type="number" step="0.01" class="form-input" /></div><div class="form-group" style="width:100px"><label class="form-label">Currency</label><select v-model="priceForm.currency" class="form-select"><option v-for="c in currencies" :key="c" :value="c">{{ c }}</option></select></div><div class="form-group" style="width:80px"><label class="form-label">Qty</label><input v-model.number="priceForm.qty" type="number" min="0" class="form-input" /></div></div>
        <div class="form-group"><label class="form-label">Notes</label><input v-model="priceForm.notes" type="text" class="form-input" placeholder="Optional" /></div>
      </div>
      <div class="modal-footer"><button class="btn btn-secondary" @click="showAddPriceModal = false">Cancel</button><button class="btn btn-primary" :disabled="!priceForm.vendorName" @click="savePrice">Add Record</button></div></div>
    </div>

    <!-- ═══ Create/Edit Product Modal (existing) ══════════════ -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal modal-lg">
          <div class="modal-header"><h2 class="modal-title">{{ editingId ? 'Edit Product' : 'Add Product' }}</h2><button class="modal-close" @click="showModal = false"><X :size="20" /></button></div>
          <div class="modal-body">
            <h4 class="section-title mb-4">Basic Information</h4>
            <div class="form-row"><div class="form-group"><label class="form-label">SKU <span class="required">*</span></label><input v-model="form.sku" type="text" class="form-input text-mono" placeholder="e.g. HIK-DS2CD2143" /></div><div class="form-group"><label class="form-label">Product Name <span class="required">*</span></label><input v-model="form.name" type="text" class="form-input" placeholder="Product name" /></div></div>
            <div class="form-group"><label class="form-label">Description</label><textarea v-model="form.description" class="form-textarea" rows="2" placeholder="Brief product description..." /></div>
            <div class="form-row"><div class="form-group"><label class="form-label">Manufacturer <span class="required">*</span></label><select v-model="form.manufacturerId" class="form-select"><option value="" disabled>Select manufacturer</option><option v-for="m in mfrList" :key="m.id" :value="m.id">{{ m.name }}</option></select></div><div class="form-group"><label class="form-label">Category <span class="required">*</span></label><select v-model="form.categoryId" class="form-select" :disabled="!form.manufacturerId"><option value="" disabled>{{ form.manufacturerId ? 'Select category' : 'Select manufacturer first' }}</option><option v-for="c in availableCategories" :key="c.id" :value="c.id">{{ c.name }}</option></select></div></div>
            <div class="divider" />
            <h4 class="section-title mb-4"><Calculator :size="16" style="vertical-align:-2px" /> Pricing</h4>
            <div class="form-group"><label class="form-label">Product Type</label><div class="flex gap-4 mt-1"><label class="form-radio"><input v-model="form.productType" type="radio" value="import" /> Import</label><label class="form-radio"><input v-model="form.productType" type="radio" value="local" /> Local</label></div></div>
            <div v-if="form.productType === 'import'" class="import-pricing">
              <div class="form-row"><div class="form-group"><label class="form-label">Origin Currency</label><select v-model="form.originCurrency" class="form-select"><option v-for="c in currencies.filter(x => x !== 'SAR')" :key="c" :value="c">{{ c }}</option></select></div><div class="form-group"><label class="form-label">Unit Cost ({{ form.originCurrency }})</label><input v-model.number="form.unitCostOrigin" type="number" step="0.01" min="0" class="form-input" /></div><div class="form-group"><label class="form-label">FX Rate → SAR</label><input v-model.number="form.fxRate" type="number" step="0.0001" min="0" class="form-input" /></div></div>
              <div class="form-group"><label class="form-label">Cost in SAR (auto)</label><input :value="formatSAR(costInSAR)" type="text" class="form-input" disabled /></div>
              <div class="form-row"><div class="form-group"><label class="form-label">Freight %</label><input v-model.number="form.freightPercent" type="number" step="0.1" min="0" class="form-input" /></div><div class="form-group"><label class="form-label">Customs %</label><input v-model.number="form.customsPercent" type="number" step="0.1" min="0" class="form-input" /></div><div class="form-group"><label class="form-label">Clearance %</label><input v-model.number="form.clearancePercent" type="number" step="0.1" min="0" class="form-input" /></div></div>
            </div>
            <div v-else class="local-pricing"><div class="form-row"><div class="form-group"><label class="form-label">Unit Cost (SAR)</label><input v-model.number="form.unitCostOrigin" type="number" step="0.01" min="0" class="form-input" /></div></div></div>
            <div class="form-row"><div class="form-group"><label class="form-label">Landed Cost SAR (auto)</label><input :value="'SAR ' + formatSAR(landedCostSAR)" type="text" class="form-input" disabled /></div><div class="form-group"><label class="form-label">Target Margin %</label><input v-model.number="form.targetMarginPercent" type="number" step="0.5" min="0" max="99" class="form-input" /></div><div class="form-group"><label class="form-label">Selling Price SAR (auto)</label><input :value="'SAR ' + formatSAR(sellingPrice)" type="text" class="form-input" disabled /></div></div>
            <div class="divider" />
            <h4 class="section-title mb-4">Additional Information</h4>
            <div class="form-row"><div class="form-group"><label class="form-label">Lead Time (days)</label><input v-model.number="form.leadTimeDays" type="number" min="0" class="form-input" /></div><div class="form-group"><label class="form-label">Supplier Name</label><input v-model="form.supplierName" type="text" class="form-input" placeholder="Supplier / Distributor" /></div></div>
            <label class="form-checkbox"><input v-model="form.isActive" type="checkbox" /> Active</label>
          </div>
          <div class="modal-footer"><button class="btn btn-secondary" @click="showModal = false">Cancel</button><button class="btn btn-primary" :disabled="!form.sku || !form.name || !form.manufacturerId || !form.categoryId" @click="saveProduct">{{ editingId ? 'Save Changes' : 'Add Product' }}</button></div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.products-page { padding: var(--space-6); }
.toolbar { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-4) var(--space-5); flex-wrap: wrap; }
.toolbar-search { flex: 1; min-width: 220px; }
.toolbar-select { width: 180px; flex-shrink: 0; }
.section-title { font-size: var(--text-sm); font-weight: var(--font-semibold); text-transform: uppercase; letter-spacing: 0.04em; color: var(--color-neutral-500); }
.import-pricing, .local-pricing { margin-bottom: var(--space-2); }

.prod-row { cursor: pointer; transition: background 80ms; }
.prod-row:hover { background: var(--color-primary-50, #eef2ff); }

/* Detail Modal */
.prod-detail-modal { width: 100%; max-width: 1280px; max-height: calc(100vh - var(--space-8)); }
.prod-detail-header { display: flex; align-items: center; justify-content: space-between; padding: var(--space-5) var(--space-6); border-bottom: 1px solid var(--color-neutral-200); gap: var(--space-4); }
.prod-detail-header-left { display: flex; align-items: center; gap: var(--space-4); }
.prod-detail-icon { width: 48px; height: 48px; border-radius: var(--radius-lg); background: var(--color-primary-light, #e0e7ff); color: var(--color-primary); display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.prod-detail-title { font-size: 1.15rem; font-weight: 700; color: var(--color-neutral-900); margin: 0; }
.prod-detail-meta { display: flex; align-items: center; gap: var(--space-2); margin-top: 4px; font-size: 0.8rem; flex-wrap: wrap; }
.prod-detail-sku { font-family: var(--font-mono); font-weight: 600; color: var(--color-primary); background: var(--color-primary-50, #eef2ff); padding: 1px 8px; border-radius: var(--radius-sm); font-size: 0.75rem; }
.prod-sep { color: var(--color-neutral-300); }
.prod-detail-header-right { display: flex; align-items: center; gap: var(--space-5); }
.prod-hdr-num { text-align: center; }
.prod-hdr-val { display: block; font-size: 1rem; font-weight: 800; color: var(--color-neutral-800); }
.prod-hdr-lbl { font-size: 0.62rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.05em; }

.prod-detail-tabs { display: flex; gap: 0; padding: 0 var(--space-6); border-bottom: 2px solid var(--color-neutral-200); overflow-x: auto; background: var(--color-neutral-50); }
.prod-dtab { display: flex; align-items: center; gap: 6px; padding: var(--space-3) var(--space-5); font-size: 0.82rem; font-weight: 500; color: var(--color-neutral-500); border-bottom: 3px solid transparent; margin-bottom: -2px; cursor: pointer; background: none; border-top: none; border-left: none; border-right: none; transition: all 120ms; white-space: nowrap; }
.prod-dtab:hover { color: var(--color-neutral-700); background: var(--color-neutral-100); }
.prod-dtab--active { color: var(--color-primary); border-bottom-color: var(--color-primary); font-weight: 700; background: transparent; }
.prod-dtab-ct { font-size: 0.68rem; font-weight: 700; background: var(--color-neutral-200); color: var(--color-neutral-600); padding: 1px 7px; border-radius: 10px; }
.prod-dtab--active .prod-dtab-ct { background: var(--color-primary-100, #c7d2fe); color: var(--color-primary); }
.prod-detail-body { padding: var(--space-6); max-height: 70vh; overflow-y: auto; }

/* Info grid */
.prod-info-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: var(--space-3); margin-bottom: var(--space-4); }
.prod-info-card { padding: var(--space-3) var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); }
.prod-info-lbl { display: block; font-size: 0.65rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 2px; }
.prod-info-val { font-size: 0.88rem; font-weight: 600; color: var(--color-neutral-800); }

.prod-section-title { display: flex; align-items: center; gap: var(--space-2); font-size: 0.85rem; font-weight: 700; color: var(--color-neutral-700); margin: var(--space-5) 0 var(--space-3) 0; }

/* Cost grid */
.prod-cost-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: var(--space-3); }
.prod-cost-item { display: flex; flex-direction: column; padding: var(--space-3) var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-md); }
.prod-cost-item--highlight { background: var(--color-neutral-50); border-color: var(--color-neutral-300); }
.prod-cost-lbl { font-size: 0.65rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; }
.prod-cost-val { font-size: 0.82rem; font-weight: 600; color: var(--color-neutral-800); margin-top: 2px; }

/* Tab toolbar */
.prod-tab-toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-4); }

/* Documents grid */
.prod-docs-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: var(--space-4); }
.prod-doc-card { padding: var(--space-4); background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); display: flex; flex-direction: column; gap: var(--space-2); }
.prod-doc-card-top { display: flex; align-items: center; justify-content: space-between; }
.prod-doc-name { font-size: 0.88rem; font-weight: 600; color: var(--color-neutral-800); }
.prod-doc-meta { display: flex; gap: var(--space-2); font-size: 0.72rem; color: var(--color-neutral-500); }
.prod-doc-footer { display: flex; align-items: center; justify-content: space-between; font-size: 0.68rem; color: var(--color-neutral-500); margin-top: var(--space-1); }

/* Reliability */
.reliability-wrap { display: flex; align-items: center; gap: var(--space-2); }
.reliability-bar { width: 60px; height: 6px; background: var(--color-neutral-100); border-radius: 3px; overflow: hidden; }
.reliability-fill { height: 100%; border-radius: 3px; }
.bg-green { background: #059669; } .bg-orange { background: #d97706; } .bg-red { background: #dc2626; }
.reliability-lbl { font-size: 0.72rem; color: var(--color-neutral-600); }

/* Embedded tables */
.table-container--embedded { border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg); overflow: auto; max-height: 400px; }
.table-container--embedded .table { margin-bottom: 0; }
.table-container--embedded .table th { position: sticky; top: 0; z-index: 2; background: var(--color-neutral-50); }

/* Empty tab */
.prod-empty-tab { display: flex; flex-direction: column; align-items: center; gap: var(--space-3); padding: var(--space-8) 0; }
.prod-empty-tab p { font-size: 0.82rem; }

/* Toolbar actions */
.prod-tab-actions { display: flex; gap: var(--space-2); }

/* ─── Upload Catalog Modal ─────────────────────────────── */
.cat-vendor-row { display: flex; gap: var(--space-4); margin-bottom: var(--space-4); }
.cat-upload-zone {
  border: 2px dashed var(--color-neutral-300);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
  background: var(--content-surface);
}
.cat-upload-zone:hover { border-color: var(--color-primary-500); background: var(--color-primary-50); }
[data-theme="dark"] .cat-upload-zone:hover { background: rgba(59,130,246,0.08); }
.cat-upload-placeholder { display: flex; flex-direction: column; align-items: center; gap: var(--space-2); text-align: center; }
.cat-upload-title { font-size: 0.92rem; font-weight: 600; color: var(--color-neutral-700); margin: 0; }
.cat-upload-subtitle { font-size: 0.78rem; color: var(--color-neutral-500); margin: 0; }
.cat-upload-types { display: flex; gap: var(--space-2); margin-top: var(--space-2); }
.cat-type-badge {
  display: inline-flex; align-items: center; gap: 4px;
  font-size: 0.68rem; font-weight: 500;
  padding: 3px 10px;
  background: var(--color-neutral-100);
  border-radius: var(--radius-full);
  color: var(--color-neutral-600);
}
.cat-upload-selected { display: flex; align-items: center; gap: var(--space-3); font-size: 0.85rem; }
.cat-parse-error {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  margin-top: var(--space-3);
  background: #fef3c7; color: #92400e;
  border-radius: var(--radius-md); font-size: 0.82rem; font-weight: 500;
}
[data-theme="dark"] .cat-parse-error { background: rgba(251,191,36,0.15); color: #fbbf24; }

.cat-pdf-section { margin-top: var(--space-4); }
.cat-pdf-notice {
  display: flex; align-items: flex-start; gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--color-primary-50); color: var(--color-primary-700);
  border-radius: var(--radius-md); font-size: 0.78rem; margin-bottom: var(--space-3);
}
[data-theme="dark"] .cat-pdf-notice { background: rgba(59,130,246,0.1); color: #93c5fd; }
.cat-paste-area {
  width: 100%; font-family: var(--font-mono, monospace); font-size: 0.78rem;
  line-height: 1.5; resize: vertical; min-height: 120px;
}

.cat-preview-section { margin-top: var(--space-4); }
.cat-preview-header {
  display: flex; align-items: center; gap: var(--space-3);
  margin-bottom: var(--space-3); font-size: 0.82rem;
}
.form-input--mini {
  padding: 3px 6px; font-size: 0.78rem; height: auto; min-height: unset;
  border: 1px solid var(--color-neutral-200); border-radius: var(--radius-sm);
  background: transparent; width: 100%;
}
.form-input--mini:focus { border-color: var(--color-primary-500); outline: none; }
.cat-row--deselected { opacity: 0.45; }
.cat-parsing-indicator {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4); margin-top: var(--space-3);
  font-size: 0.82rem; font-weight: 500; color: var(--color-primary-600);
  background: var(--color-primary-50); border-radius: var(--radius-md);
}
[data-theme="dark"] .cat-parsing-indicator { background: rgba(59,130,246,0.1); color: #93c5fd; }
@keyframes cat-spin { to { transform: rotate(360deg); } }
.cat-spin-icon { animation: cat-spin 1s linear infinite; }
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

@media (max-width: 1200px) { .prod-info-grid { grid-template-columns: repeat(2, 1fr); } .prod-cost-grid { grid-template-columns: repeat(3, 1fr); } .prod-detail-header { flex-direction: column; align-items: flex-start; } }
@media (max-width: 768px) { .toolbar { flex-direction: column; align-items: stretch; } .toolbar-select { width: 100%; } .prod-info-grid { grid-template-columns: 1fr; } .prod-cost-grid { grid-template-columns: repeat(2, 1fr); } .prod-docs-grid { grid-template-columns: 1fr; } .cat-vendor-row { flex-direction: column; } }
</style>
