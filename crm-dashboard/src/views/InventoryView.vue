<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Search, Package, Boxes, Lock, DollarSign, AlertTriangle, Eye,
  TrendingUp, TrendingDown, History, X, Truck, FileText, Building2,
  ShoppingCart, PackageCheck, MapPin, BarChart3, ArrowUpRight,
  ChevronDown, ChevronUp, Minus, AlertCircle, RefreshCw,
  Plus, Check, Unlock, ArrowRightLeft, ClipboardList, User, ShoppingBag,
} from 'lucide-vue-next'
import { useProcurementStore } from '@/stores/procurement'
import { useQuotesStore } from '@/stores/quotes'
import { useManufacturersStore } from '@/stores/manufacturers'
import { useWarehouseStockStore } from '@/stores/warehouseStock'
import type {
  WarehouseStock, WarehouseLocation, Currency,
  PurchaseOrder, PurchaseOrderItem,
  SupplierQuote, SupplierQuoteLineItem,
  GoodsReceipt, GoodsReceiptItem,
  SupplierItemEntry,
  StockReservation, ReservationSource, ReservationStatus,
  InventoryMovement, MovementType,
  Quote,
} from '@/types'

const procStore = useProcurementStore()
const quotesStore = useQuotesStore()
const mfrStore = useManufacturersStore()
const stockStore = useWarehouseStockStore()

onMounted(() => {
  stockStore.fetchStock()
  procStore.fetchPurchaseOrders()
  procStore.fetchSupplierQuotes()
  procStore.fetchGoodsReceipts()
  quotesStore.fetchQuotes()
  mfrStore.fetchManufacturers()
})

function formatSAR(v: number): string {
  return v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}
function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

const warehouseLabels: Record<WarehouseLocation, string> = {
  'riyadh-main': 'Riyadh Main',
  'jeddah-branch': 'Jeddah Branch',
  'dammam-branch': 'Dammam Branch',
}
const warehouseBadge: Record<WarehouseLocation, string> = {
  'riyadh-main': 'badge-primary',
  'jeddah-branch': 'badge-info',
  'dammam-branch': 'badge-warning',
}

// ── Stock Data ───────────────────────────────────────────────
const stockItems = ref<WarehouseStock[]>([
  { id: 'ws1', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', warehouseLocation: 'riyadh-main', onHandQty: 120, reservedQty: 35, availableQty: 85, unitCost: 365.63, totalValue: 43875.00, reorderLevel: 40, createdAt: '2024-03-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'ws2', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', warehouseLocation: 'jeddah-branch', onHandQty: 45, reservedQty: 10, availableQty: 35, unitCost: 365.63, totalValue: 16453.35, reorderLevel: 15, createdAt: '2024-03-01T08:00:00Z', updatedAt: '2026-02-18T08:00:00Z' },
  { id: 'ws3', productId: 'cp2', productSku: 'HIK-DS2CD2T87', productName: 'DS-2CD2T87G2-L 8MP Bullet', warehouseLocation: 'riyadh-main', onHandQty: 60, reservedQty: 22, availableQty: 38, unitCost: 839.06, totalValue: 50343.75, reorderLevel: 20, createdAt: '2024-04-10T08:00:00Z', updatedAt: '2026-02-19T08:00:00Z' },
  { id: 'ws4', productId: 'cp2', productSku: 'HIK-DS2CD2T87', productName: 'DS-2CD2T87G2-L 8MP Bullet', warehouseLocation: 'dammam-branch', onHandQty: 18, reservedQty: 5, availableQty: 13, unitCost: 839.06, totalValue: 15103.13, reorderLevel: 10, createdAt: '2024-04-10T08:00:00Z', updatedAt: '2026-02-15T08:00:00Z' },
  { id: 'ws5', productId: 'cp3', productSku: 'HIK-DS7732NI', productName: 'DS-7732NI-K4 32CH NVR', warehouseLocation: 'riyadh-main', onHandQty: 15, reservedQty: 8, availableQty: 7, unitCost: 1827.00, totalValue: 27405.00, reorderLevel: 5, createdAt: '2024-05-01T08:00:00Z', updatedAt: '2026-02-21T08:00:00Z' },
  { id: 'ws6', productId: 'cp4', productSku: 'DH-IPC-HFW5442', productName: 'IPC-HFW5442T-ASE 4MP AI Bullet', warehouseLocation: 'riyadh-main', onHandQty: 80, reservedQty: 30, availableQty: 50, unitCost: 473.44, totalValue: 37875.00, reorderLevel: 25, createdAt: '2024-06-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'ws7', productId: 'cp4', productSku: 'DH-IPC-HFW5442', productName: 'IPC-HFW5442T-ASE 4MP AI Bullet', warehouseLocation: 'jeddah-branch', onHandQty: 25, reservedQty: 25, availableQty: 0, unitCost: 473.44, totalValue: 11836.00, reorderLevel: 10, createdAt: '2024-06-01T08:00:00Z', updatedAt: '2026-02-17T08:00:00Z' },
  { id: 'ws8', productId: 'cp5', productSku: 'DH-NVR5432-EI', productName: 'NVR5432-EI 32CH AI NVR', warehouseLocation: 'dammam-branch', onHandQty: 8, reservedQty: 3, availableQty: 5, unitCost: 2958.00, totalValue: 23664.00, reorderLevel: 3, createdAt: '2024-07-01T08:00:00Z', updatedAt: '2026-02-16T08:00:00Z' },
  { id: 'ws9', productId: 'cp7', productSku: 'AXIS-P3265LVE', productName: 'P3265-LVE 2MP Dome', warehouseLocation: 'riyadh-main', onHandQty: 40, reservedQty: 12, availableQty: 28, unitCost: 1729.40, totalValue: 69176.00, reorderLevel: 15, createdAt: '2024-08-01T08:00:00Z', updatedAt: '2026-02-19T08:00:00Z' },
  { id: 'ws10', productId: 'cp8', productSku: 'AXIS-Q6135LE', productName: 'Q6135-LE PTZ Camera', warehouseLocation: 'riyadh-main', onHandQty: 4, reservedQty: 2, availableQty: 2, unitCost: 19118.00, totalValue: 76472.00, reorderLevel: 2, createdAt: '2024-09-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'ws11', productId: 'cp9', productSku: 'HON-MNPDS2', productName: 'Morley-IAS Fire Panel 2L', warehouseLocation: 'jeddah-branch', onHandQty: 6, reservedQty: 1, availableQty: 5, unitCost: 5437.50, totalValue: 32625.00, reorderLevel: 3, createdAt: '2024-10-01T08:00:00Z', updatedAt: '2026-02-18T08:00:00Z' },
  { id: 'ws12', productId: 'cp10', productSku: 'HON-MAXPRO', productName: 'MAXPRO Access 4-Door', warehouseLocation: 'riyadh-main', onHandQty: 22, reservedQty: 10, availableQty: 12, unitCost: 2100.00, totalValue: 46200.00, reorderLevel: 8, createdAt: '2024-10-15T08:00:00Z', updatedAt: '2026-02-21T08:00:00Z' },
  { id: 'ws13', productId: 'cp13', productSku: 'ZKT-INBIO460', productName: 'InBio460 4-Door Controller', warehouseLocation: 'dammam-branch', onHandQty: 10, reservedQty: 10, availableQty: 0, unitCost: 1029.60, totalValue: 10296.00, reorderLevel: 5, createdAt: '2024-11-01T08:00:00Z', updatedAt: '2026-02-14T08:00:00Z' },
  { id: 'ws14', productId: 'cp14', productSku: 'ZKT-SPEEDFACE', productName: 'SpeedFace-V5L Facial Terminal', warehouseLocation: 'riyadh-main', onHandQty: 30, reservedQty: 8, availableQty: 22, unitCost: 1808.80, totalValue: 54264.00, reorderLevel: 10, createdAt: '2024-11-15T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z' },
  { id: 'ws15', productId: 'cp12', productSku: 'BOSCH-FPA5000', productName: 'FPA-5000 Fire Panel', warehouseLocation: 'jeddah-branch', onHandQty: 3, reservedQty: 0, availableQty: 3, unitCost: 8413.20, totalValue: 25239.60, reorderLevel: 2, createdAt: '2024-12-01T08:00:00Z', updatedAt: '2026-02-15T08:00:00Z' },
  { id: 'ws16', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', warehouseLocation: 'riyadh-main', onHandQty: 65, reservedQty: 20, availableQty: 45, unitCost: 420.00, totalValue: 27300.00, reorderLevel: 30, createdAt: '2024-12-10T08:00:00Z', updatedAt: '2026-01-15T08:00:00Z' },
  { id: 'ws17', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', warehouseLocation: 'dammam-branch', onHandQty: 32, reservedQty: 10, availableQty: 22, unitCost: 420.00, totalValue: 13440.00, reorderLevel: 15, createdAt: '2024-12-10T08:00:00Z', updatedAt: '2025-12-12T08:00:00Z' },
  { id: 'ws18', productId: 'cp6', productSku: 'DH-ASI7214Y', productName: 'ASI7214Y Face Recognition Terminal', warehouseLocation: 'riyadh-main', onHandQty: 14, reservedQty: 6, availableQty: 8, unitCost: 1440.00, totalValue: 20160.00, reorderLevel: 5, createdAt: '2025-01-10T08:00:00Z', updatedAt: '2026-01-20T08:00:00Z' },
  { id: 'ws19', productId: 'cp11', productSku: 'BOSCH-NDV3503', productName: 'FLEXIDOME IP 3000i 5MP', warehouseLocation: 'riyadh-main', onHandQty: 18, reservedQty: 4, availableQty: 14, unitCost: 1252.00, totalValue: 22536.00, reorderLevel: 8, createdAt: '2025-02-01T08:00:00Z', updatedAt: '2026-01-18T08:00:00Z' },
  { id: 'ws20', productId: 'cp16', productSku: 'CBL-FIBER-OM3', productName: 'OM3 Fiber Optic Cable 1000m', warehouseLocation: 'riyadh-main', onHandQty: 8, reservedQty: 2, availableQty: 6, unitCost: 1850.00, totalValue: 14800.00, reorderLevel: 4, createdAt: '2025-03-01T08:00:00Z', updatedAt: '2026-01-12T08:00:00Z' },
])

// ── Manufacturer Mapping ─────────────────────────────────────
const productManufacturer: Record<string, string> = {
  'HIK-DS2CD2143': 'Hikvision', 'HIK-DS2CD2T87': 'Hikvision', 'HIK-DS7732NI': 'Hikvision',
  'DH-IPC-HFW5442': 'Dahua', 'DH-NVR5432-EI': 'Dahua', 'DH-ASI7214Y': 'Dahua',
  'AXIS-P3265LVE': 'Axis', 'AXIS-Q6135LE': 'Axis',
  'HON-MNPDS2': 'Honeywell', 'HON-MAXPRO': 'Honeywell',
  'ZKT-INBIO460': 'ZKTeco', 'ZKT-SPEEDFACE': 'ZKTeco',
  'BOSCH-FPA5000': 'Bosch', 'BOSCH-NDV3503': 'Bosch',
  'CBL-CAT6A-305': 'Belden', 'CBL-FIBER-OM3': 'Corning',
}

const productCategory: Record<string, string> = {
  'HIK-DS2CD2143': 'CCTV', 'HIK-DS2CD2T87': 'CCTV', 'HIK-DS7732NI': 'NVR',
  'DH-IPC-HFW5442': 'CCTV', 'DH-NVR5432-EI': 'NVR', 'DH-ASI7214Y': 'Access Control',
  'AXIS-P3265LVE': 'CCTV', 'AXIS-Q6135LE': 'CCTV',
  'HON-MNPDS2': 'Fire Alarm', 'HON-MAXPRO': 'Access Control',
  'ZKT-INBIO460': 'Access Control', 'ZKT-SPEEDFACE': 'Access Control',
  'BOSCH-FPA5000': 'Fire Alarm', 'BOSCH-NDV3503': 'CCTV',
  'CBL-CAT6A-305': 'Cabling', 'CBL-FIBER-OM3': 'Cabling',
}

// ── Stock Reservations (Holds) ───────────────────────────────
const reservations = ref<StockReservation[]>([
  { id: 'res-1', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', warehouseLocation: 'riyadh-main', qty: 20, source: 'quote', sourceRef: 'QT-2026-0148', sourceLabel: 'Saudi Aramco CCTV Phase 2', customerName: 'Saudi Aramco', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-10T09:00:00Z', status: 'active' },
  { id: 'res-2', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', warehouseLocation: 'riyadh-main', qty: 15, source: 'project', sourceRef: 'PRJ-2026-003', sourceLabel: 'KAIA Airport Terminal 5', customerName: 'GACA', reservedBy: 'Khalid Al-Rashid', reservedAt: '2026-02-15T10:00:00Z', status: 'active' },
  { id: 'res-3', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', warehouseLocation: 'jeddah-branch', qty: 10, source: 'quote', sourceRef: 'QT-2026-0152', sourceLabel: 'Madinah Hotel Security', customerName: 'Al-Madinah Hotels', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-18T14:00:00Z', status: 'active' },
  { id: 'res-4', productId: 'cp2', productSku: 'HIK-DS2CD2T87', productName: 'DS-2CD2T87G2-L 8MP Bullet', warehouseLocation: 'riyadh-main', qty: 22, source: 'quote', sourceRef: 'QT-2026-0148', sourceLabel: 'Saudi Aramco CCTV Phase 2', customerName: 'Saudi Aramco', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-10T09:00:00Z', status: 'active' },
  { id: 'res-5', productId: 'cp3', productSku: 'HIK-DS7732NI', productName: 'DS-7732NI-K4 32CH NVR', warehouseLocation: 'riyadh-main', qty: 4, source: 'quote', sourceRef: 'QT-2026-0148', sourceLabel: 'Saudi Aramco CCTV Phase 2', customerName: 'Saudi Aramco', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-10T09:00:00Z', status: 'active' },
  { id: 'res-6', productId: 'cp3', productSku: 'HIK-DS7732NI', productName: 'DS-7732NI-K4 32CH NVR', warehouseLocation: 'riyadh-main', qty: 4, source: 'project', sourceRef: 'PRJ-2026-003', sourceLabel: 'KAIA Airport Terminal 5', customerName: 'GACA', reservedBy: 'Khalid Al-Rashid', reservedAt: '2026-02-15T10:00:00Z', status: 'active' },
  { id: 'res-7', productId: 'cp4', productSku: 'DH-IPC-HFW5442', productName: 'IPC-HFW5442T-ASE 4MP AI Bullet', warehouseLocation: 'riyadh-main', qty: 30, source: 'project', sourceRef: 'PRJ-2026-005', sourceLabel: 'MOI Headquarters Upgrade', customerName: 'Ministry of Interior', reservedBy: 'Mohammed Al-Zahrani', reservedAt: '2026-02-12T08:00:00Z', status: 'active' },
  { id: 'res-8', productId: 'cp4', productSku: 'DH-IPC-HFW5442', productName: 'IPC-HFW5442T-ASE 4MP AI Bullet', warehouseLocation: 'jeddah-branch', qty: 25, source: 'quote', sourceRef: 'QT-2026-0155', sourceLabel: 'KAEC Smart City Phase 1', customerName: 'KAEC', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-17T11:00:00Z', status: 'active' },
  { id: 'res-9', productId: 'cp10', productSku: 'HON-MAXPRO', productName: 'MAXPRO Access 4-Door', warehouseLocation: 'riyadh-main', qty: 10, source: 'quote', sourceRef: 'QT-2026-0149', sourceLabel: 'SABIC Office Access Control', customerName: 'SABIC', reservedBy: 'Khalid Al-Rashid', reservedAt: '2026-02-14T13:00:00Z', status: 'active' },
  { id: 'res-10', productId: 'cp13', productSku: 'ZKT-INBIO460', productName: 'InBio460 4-Door Controller', warehouseLocation: 'dammam-branch', qty: 10, source: 'project', sourceRef: 'PRJ-2026-001', sourceLabel: 'Jubail Industrial Access', customerName: 'Royal Commission Jubail', reservedBy: 'Abdullah Al-Qahtani', reservedAt: '2026-01-20T08:00:00Z', status: 'active' },
  { id: 'res-11', productId: 'cp14', productSku: 'ZKT-SPEEDFACE', productName: 'SpeedFace-V5L Facial Terminal', warehouseLocation: 'riyadh-main', qty: 8, source: 'quote', sourceRef: 'QT-2026-0150', sourceLabel: 'NEOM Staff Access Phase 1', customerName: 'NEOM', reservedBy: 'Ahmed bin Saleh', reservedAt: '2026-02-19T09:00:00Z', status: 'active' },
  { id: 'res-12', productId: 'cp7', productSku: 'AXIS-P3265LVE', productName: 'P3265-LVE 2MP Dome', warehouseLocation: 'riyadh-main', qty: 12, source: 'project', sourceRef: 'PRJ-2026-004', sourceLabel: 'Riyadh Metro Station CCTV', customerName: 'Riyadh Metro', reservedBy: 'Mohammed Al-Zahrani', reservedAt: '2026-02-08T10:00:00Z', status: 'active' },
  { id: 'res-13', productId: 'cp8', productSku: 'AXIS-Q6135LE', productName: 'Q6135-LE PTZ Camera', warehouseLocation: 'riyadh-main', qty: 2, source: 'quote', sourceRef: 'QT-2026-0145', sourceLabel: 'Ministry Perimeter Security', customerName: 'Ministry of Interior', reservedBy: 'Khalid Al-Rashid', reservedAt: '2026-02-16T14:00:00Z', status: 'active' },
  // Released reservations
  { id: 'res-14', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', warehouseLocation: 'dammam-branch', qty: 15, source: 'quote', sourceRef: 'QT-2025-0098', sourceLabel: 'SABIC Cabling Upgrade', customerName: 'SABIC', reservedBy: 'Fahad Al-Mutairi', reservedAt: '2025-11-10T08:00:00Z', releaseDate: '2025-12-15T10:00:00Z', releasedBy: 'Fahad Al-Mutairi', releaseReason: 'Quote expired — customer did not proceed', status: 'released' },
  { id: 'res-15', productId: 'cp9', productSku: 'HON-MNPDS2', productName: 'Morley-IAS Fire Panel 2L', warehouseLocation: 'jeddah-branch', qty: 1, source: 'manual', sourceRef: 'DEMO-001', sourceLabel: 'Demo unit for Jeddah showroom', customerName: 'Internal', reservedBy: 'Khalid Al-Rashid', reservedAt: '2025-12-01T08:00:00Z', releaseDate: '2026-01-15T12:00:00Z', releasedBy: 'Khalid Al-Rashid', releaseReason: 'Demo completed, returned to stock', status: 'released' },
  // Fulfilled
  { id: 'res-16', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', warehouseLocation: 'riyadh-main', qty: 20, source: 'project', sourceRef: 'PRJ-2025-012', sourceLabel: 'SABIC HQ Network Build', customerName: 'SABIC', reservedBy: 'Ahmed bin Saleh', reservedAt: '2025-10-01T08:00:00Z', releaseDate: '2025-11-20T10:00:00Z', releasedBy: 'Ahmed bin Saleh', releaseReason: 'Dispatched to project site', status: 'fulfilled' },
])

// ── Inventory Movements ──────────────────────────────────────
const movements = ref<InventoryMovement[]>([
  { id: 'mv-1', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', movementType: 'receipt', qty: 50, toWarehouse: 'riyadh-main', reference: 'GR-2026-0003', reason: 'Goods receipt from PO-2025-0092', performedBy: 'Mohammed Al-Zahrani', performedAt: '2025-10-05T09:30:00Z' },
  { id: 'mv-2', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', movementType: 'transfer', qty: 20, fromWarehouse: 'riyadh-main', toWarehouse: 'jeddah-branch', reference: 'TRF-2025-015', reason: 'Replenish Jeddah stock for upcoming projects', performedBy: 'Abdullah Al-Qahtani', performedAt: '2025-11-12T10:00:00Z' },
  { id: 'mv-3', productId: 'cp1', productSku: 'HIK-DS2CD2143', productName: 'DS-2CD2143G2-IU 4MP Dome', movementType: 'allocation', qty: 20, fromWarehouse: 'riyadh-main', reference: 'QT-2026-0148', reason: 'Reserved for Saudi Aramco CCTV Phase 2', performedBy: 'Ahmed bin Saleh', performedAt: '2026-02-10T09:00:00Z' },
  { id: 'mv-4', productId: 'cp2', productSku: 'HIK-DS2CD2T87', productName: 'DS-2CD2T87G2-L 8MP Bullet', movementType: 'receipt', qty: 30, toWarehouse: 'riyadh-main', reference: 'GR-2026-0003', reason: 'Goods receipt from PO-2025-0092', performedBy: 'Mohammed Al-Zahrani', performedAt: '2025-10-05T09:30:00Z' },
  { id: 'mv-5', productId: 'cp2', productSku: 'HIK-DS2CD2T87', productName: 'DS-2CD2T87G2-L 8MP Bullet', movementType: 'transfer', qty: 10, fromWarehouse: 'riyadh-main', toWarehouse: 'dammam-branch', reference: 'TRF-2026-002', reason: 'Dammam project needs', performedBy: 'Fahad Al-Mutairi', performedAt: '2026-01-05T11:00:00Z' },
  { id: 'mv-6', productId: 'cp4', productSku: 'DH-IPC-HFW5442', productName: 'IPC-HFW5442T-ASE 4MP AI Bullet', movementType: 'allocation', qty: 30, fromWarehouse: 'riyadh-main', reference: 'PRJ-2026-005', reason: 'Reserved for MOI Headquarters Upgrade', performedBy: 'Mohammed Al-Zahrani', performedAt: '2026-02-12T08:00:00Z' },
  { id: 'mv-7', productId: 'cp10', productSku: 'HON-MAXPRO', productName: 'MAXPRO Access 4-Door', movementType: 'receipt', qty: 20, toWarehouse: 'riyadh-main', reference: 'GR-2026-0005', reason: 'Goods receipt from PO-2026-0011', performedBy: 'Mohammed Al-Zahrani', performedAt: '2026-02-20T14:00:00Z' },
  { id: 'mv-8', productId: 'cp13', productSku: 'ZKT-INBIO460', productName: 'InBio460 4-Door Controller', movementType: 'receipt', qty: 15, toWarehouse: 'dammam-branch', reference: 'GR-2026-0004', reason: 'Goods receipt from PO-2026-0010', performedBy: 'Abdullah Al-Qahtani', performedAt: '2026-02-14T10:00:00Z' },
  { id: 'mv-9', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', movementType: 'receipt', qty: 50, toWarehouse: 'dammam-branch', reference: 'GR-2026-0002', reason: 'Goods receipt from PO-2025-0088', performedBy: 'Fahad Al-Mutairi', performedAt: '2025-12-12T11:00:00Z' },
  { id: 'mv-10', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', movementType: 'release', qty: 15, toWarehouse: 'dammam-branch', reference: 'QT-2025-0098', reason: 'Hold released — quote expired', performedBy: 'Fahad Al-Mutairi', performedAt: '2025-12-15T10:00:00Z' },
  { id: 'mv-11', productId: 'cp15', productSku: 'CBL-CAT6A-305', productName: 'Cat6A UTP Cable 305m Box', movementType: 'adjustment', qty: -3, fromWarehouse: 'riyadh-main', reason: 'Stock count correction — 3 boxes damaged in storage', performedBy: 'Abdullah Al-Qahtani', performedAt: '2026-01-10T08:30:00Z', notes: 'Water damage from roof leak, Zone A' },
  { id: 'mv-12', productId: 'cp8', productSku: 'AXIS-Q6135LE', productName: 'Q6135-LE PTZ Camera', movementType: 'write-off', qty: -1, fromWarehouse: 'riyadh-main', reason: 'Unit defective on arrival — returned to supplier pending credit', performedBy: 'Mohammed Al-Zahrani', performedAt: '2025-09-15T14:00:00Z', notes: 'RMA #AX-RMA-2025-018' },
])

function uid(): string { return Math.random().toString(36).slice(2, 11) }

// ── Aggregated Product View ──────────────────────────────────
interface AggregatedProduct {
  productId: string
  productSku: string
  productName: string
  manufacturer: string
  category: string
  totalOnHand: number
  totalReserved: number
  totalAvailable: number
  avgUnitCost: number
  totalValue: number
  reorderLevel: number
  warehouseStock: WarehouseStock[]
  stockStatus: 'in-stock' | 'low-stock' | 'out-of-stock'
  incomingQty: number
}

const aggregatedProducts = computed<AggregatedProduct[]>(() => {
  const grouped: Record<string, WarehouseStock[]> = {}
  for (const s of stockItems.value) {
    (grouped[s.productSku] ??= []).push(s)
  }
  return Object.entries(grouped).map(([sku, items]) => {
    const totalOnHand = items.reduce((s, i) => s + i.onHandQty, 0)
    const totalReserved = items.reduce((s, i) => s + i.reservedQty, 0)
    const totalAvailable = items.reduce((s, i) => s + i.availableQty, 0)
    const totalValue = items.reduce((s, i) => s + i.totalValue, 0)
    const avgUnitCost = totalOnHand > 0 ? totalValue / totalOnHand : items[0].unitCost
    const maxReorder = Math.max(...items.map(i => i.reorderLevel ?? 0))

    const incomingQty = procStore.purchaseOrders
      .filter(po => ['ordered', 'approved'].includes(po.status))
      .reduce((sum, po) => {
        for (const item of po.items) {
          if (item.productSku === sku) sum += (item.quantity - item.receivedQty)
        }
        return sum
      }, 0)

    let stockStatus: 'in-stock' | 'low-stock' | 'out-of-stock' = 'in-stock'
    if (totalAvailable <= 0) stockStatus = 'out-of-stock'
    else if (totalAvailable <= maxReorder) stockStatus = 'low-stock'

    return {
      productId: items[0].productId,
      productSku: sku,
      productName: items[0].productName,
      manufacturer: productManufacturer[sku] || '',
      category: productCategory[sku] || '',
      totalOnHand, totalReserved, totalAvailable,
      avgUnitCost, totalValue,
      reorderLevel: maxReorder,
      warehouseStock: items,
      stockStatus,
      incomingQty,
    }
  })
})

// ── Filters ──────────────────────────────────────────────────
const searchQuery = ref('')
const activeWarehouse = ref<'all' | WarehouseLocation>('all')
const categoryFilter = ref('')
const stockStatusFilter = ref<'' | 'in-stock' | 'low-stock' | 'out-of-stock'>('')

const allCategories = computed(() => [...new Set(aggregatedProducts.value.map(p => p.category))].sort())

const filteredProducts = computed(() => {
  let list = aggregatedProducts.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter(p =>
      p.productSku.toLowerCase().includes(q) ||
      p.productName.toLowerCase().includes(q) ||
      p.manufacturer.toLowerCase().includes(q),
    )
  }
  if (categoryFilter.value) {
    list = list.filter(p => p.category === categoryFilter.value)
  }
  if (stockStatusFilter.value) {
    list = list.filter(p => p.stockStatus === stockStatusFilter.value)
  }
  if (activeWarehouse.value !== 'all') {
    list = list.filter(p => p.warehouseStock.some(w => w.warehouseLocation === activeWarehouse.value))
  }
  return list
})

// ── KPIs ─────────────────────────────────────────────────────
const totalSKUs = computed(() => aggregatedProducts.value.length)
const totalOnHand = computed(() => aggregatedProducts.value.reduce((a, p) => a + p.totalOnHand, 0))
const totalReserved = computed(() => aggregatedProducts.value.reduce((a, p) => a + p.totalReserved, 0))
const totalValue = computed(() => aggregatedProducts.value.reduce((a, p) => a + p.totalValue, 0))
const lowStockCount = computed(() => aggregatedProducts.value.filter(p => p.stockStatus === 'low-stock').length)
const outOfStockCount = computed(() => aggregatedProducts.value.filter(p => p.stockStatus === 'out-of-stock').length)

function stockStatusClass(status: string): string {
  if (status === 'in-stock') return 'stock-green'
  if (status === 'low-stock') return 'stock-orange'
  return 'stock-red'
}

function stockStatusLabel(status: string): string {
  if (status === 'in-stock') return 'In Stock'
  if (status === 'low-stock') return 'Low Stock'
  return 'Out of Stock'
}

function stockPercent(onHand: number, available: number): number {
  if (onHand === 0) return 0
  return Math.round((available / onHand) * 100)
}

// ── Item Detail Modal ────────────────────────────────────────
const showDetailModal = ref(false)
const detailProduct = ref<AggregatedProduct | null>(null)
type DetailTab = 'overview' | 'reservations' | 'price-history' | 'orders' | 'supplier-quotes' | 'receiving' | 'suppliers' | 'movements'
const detailTab = ref<DetailTab>('overview')

function openItemDetail(product: AggregatedProduct) {
  detailProduct.value = product
  detailTab.value = 'overview'
  showDetailModal.value = true
}

// Price History: combine data from POs, SQs, and GRs
interface PriceHistoryEntry {
  date: string
  source: 'Purchase Order' | 'Supplier Quote' | 'Goods Receipt'
  sourceRef: string
  supplier: string
  unitCost: number
  landingCost?: number
  qty: number
}

const detailPriceHistory = computed<PriceHistoryEntry[]>(() => {
  if (!detailProduct.value) return []
  const sku = detailProduct.value.productSku
  const pid = detailProduct.value.productId
  const entries: PriceHistoryEntry[] = []

  for (const po of procStore.purchaseOrders) {
    for (const item of po.items) {
      if (item.productSku === sku) {
        entries.push({
          date: po.createdAt.slice(0, 10),
          source: 'Purchase Order',
          sourceRef: po.poNumber,
          supplier: po.supplierName,
          unitCost: item.unitCost,
          qty: item.quantity,
        })
      }
    }
  }

  for (const sq of procStore.supplierQuotes) {
    for (const item of sq.items) {
      if (item.productSku === sku) {
        entries.push({
          date: sq.validFrom,
          source: 'Supplier Quote',
          sourceRef: sq.sqNumber,
          supplier: sq.supplierName,
          unitCost: item.unitCost,
          qty: item.quantity,
        })
      }
    }
  }

  for (const gr of procStore.goodsReceipts) {
    for (const item of gr.items) {
      if (item.productId === pid || item.productSku === sku) {
        entries.push({
          date: gr.receiveDate,
          source: 'Goods Receipt',
          sourceRef: gr.grNumber,
          supplier: gr.supplierName,
          unitCost: item.unitCost,
          landingCost: item.landingCost,
          qty: item.receivedQty,
        })
      }
    }
  }

  return entries.sort((a, b) => b.date.localeCompare(a.date))
})

const detailAvgCost = computed(() => {
  if (detailPriceHistory.value.length === 0) return 0
  const costs = detailPriceHistory.value.map(e => e.unitCost)
  return costs.reduce((s, c) => s + c, 0) / costs.length
})

const detailLowestCost = computed(() => {
  if (detailPriceHistory.value.length === 0) return 0
  return Math.min(...detailPriceHistory.value.map(e => e.unitCost))
})

const detailHighestCost = computed(() => {
  if (detailPriceHistory.value.length === 0) return 0
  return Math.max(...detailPriceHistory.value.map(e => e.unitCost))
})

const detailLastLandingCost = computed(() => {
  const grEntries = detailPriceHistory.value.filter(e => e.source === 'Goods Receipt' && e.landingCost)
  if (grEntries.length === 0) return null
  return grEntries[0].landingCost
})

// Past Orders
const detailPastOrders = computed<{ po: PurchaseOrder; item: PurchaseOrderItem }[]>(() => {
  if (!detailProduct.value) return []
  const sku = detailProduct.value.productSku
  const results: { po: PurchaseOrder; item: PurchaseOrderItem }[] = []
  for (const po of procStore.purchaseOrders) {
    for (const item of po.items) {
      if (item.productSku === sku) results.push({ po, item })
    }
  }
  return results.sort((a, b) => b.po.createdAt.localeCompare(a.po.createdAt))
})

// Supplier Quotes for item
const detailSQs = computed<{ sq: SupplierQuote; item: SupplierQuoteLineItem }[]>(() => {
  if (!detailProduct.value) return []
  return procStore.getSupplierQuotesForProduct(detailProduct.value.productSku)
})

// Receive History
const detailReceiveHistory = computed<{ gr: GoodsReceipt; item: GoodsReceiptItem }[]>(() => {
  if (!detailProduct.value) return []
  return procStore.getReceiptHistoryForProduct(detailProduct.value.productId)
})

// Suppliers carrying this item
const detailSuppliers = computed<SupplierItemEntry[]>(() => {
  if (!detailProduct.value) return []
  return procStore.getSupplierItemsForProduct(detailProduct.value.productId)
})

const poStatusConfig: Record<string, { label: string; badge: string }> = {
  draft: { label: 'Draft', badge: 'badge-gray' },
  'pending-approval': { label: 'Pending', badge: 'badge-warning' },
  approved: { label: 'Approved', badge: 'badge-info' },
  ordered: { label: 'Ordered', badge: 'badge-primary' },
  'partial-received': { label: 'Partial', badge: 'badge-warning' },
  received: { label: 'Received', badge: 'badge-success' },
  cancelled: { label: 'Cancelled', badge: 'badge-danger' },
}

const sqStatusConfig: Record<string, { label: string; badge: string }> = {
  received: { label: 'Received', badge: 'badge-info' },
  'under-review': { label: 'Under Review', badge: 'badge-warning' },
  accepted: { label: 'Accepted', badge: 'badge-success' },
  expired: { label: 'Expired', badge: 'badge-gray' },
  rejected: { label: 'Rejected', badge: 'badge-danger' },
}

function priceSourceBadge(source: string): string {
  if (source === 'Purchase Order') return 'badge-primary'
  if (source === 'Supplier Quote') return 'badge-info'
  return 'badge-success'
}

// ── Reservation & Movement helpers ───────────────────────────
const detailReservations = computed(() => {
  if (!detailProduct.value) return []
  return reservations.value.filter(r => r.productSku === detailProduct.value!.productSku)
})

const detailActiveReservations = computed(() => detailReservations.value.filter(r => r.status === 'active'))

const detailMovements = computed(() => {
  if (!detailProduct.value) return []
  return movements.value
    .filter(m => m.productSku === detailProduct.value!.productSku)
    .sort((a, b) => b.performedAt.localeCompare(a.performedAt))
})

const totalActiveHolds = computed(() => reservations.value.filter(r => r.status === 'active').reduce((s, r) => s + r.qty, 0))

function getActiveHoldsForProduct(sku: string): { count: number; qty: number } {
  const active = reservations.value.filter(r => r.productSku === sku && r.status === 'active')
  return { count: active.length, qty: active.reduce((s, r) => s + r.qty, 0) }
}

function reservationSourceBadge(source: ReservationSource): string {
  if (source === 'quote') return 'badge-primary'
  if (source === 'project') return 'badge-success'
  return 'badge-gray'
}
function reservationSourceLabel(source: ReservationSource): string {
  if (source === 'quote') return 'Won Quote'
  if (source === 'project') return 'Project'
  return 'Manual'
}
function reservationStatusBadge(status: ReservationStatus): string {
  if (status === 'active') return 'badge-warning'
  if (status === 'released') return 'badge-info'
  return 'badge-success'
}

function movementTypeBadge(type: MovementType): string {
  const map: Record<MovementType, string> = {
    transfer: 'badge-info', adjustment: 'badge-warning', allocation: 'badge-primary',
    receipt: 'badge-success', release: 'badge-gray', 'write-off': 'badge-danger',
  }
  return map[type]
}
function movementTypeLabel(type: MovementType): string {
  const map: Record<MovementType, string> = {
    transfer: 'Transfer', adjustment: 'Adjustment', allocation: 'Allocation',
    receipt: 'Receipt', release: 'Release', 'write-off': 'Write-Off',
  }
  return map[type]
}

// ── Release Hold ─────────────────────────────────────────────
const showReleaseModal = ref(false)
const releasingReservation = ref<StockReservation | null>(null)
const releaseReason = ref('')

function openReleaseHold(res: StockReservation) {
  releasingReservation.value = res
  releaseReason.value = ''
  showReleaseModal.value = true
}

function confirmRelease() {
  if (!releasingReservation.value) return
  const res = releasingReservation.value
  const idx = reservations.value.findIndex(r => r.id === res.id)
  if (idx !== -1) {
    reservations.value[idx] = {
      ...res,
      status: 'released',
      releaseDate: new Date().toISOString(),
      releasedBy: 'Current User',
      releaseReason: releaseReason.value || 'Released manually',
    }
  }
  movements.value.unshift({
    id: uid(), productId: res.productId, productSku: res.productSku,
    productName: res.productName, movementType: 'release', qty: res.qty,
    toWarehouse: res.warehouseLocation, reference: res.sourceRef,
    reason: `Hold released: ${releaseReason.value || 'Manual release'}`,
    performedBy: 'Current User', performedAt: new Date().toISOString(),
  })
  // Return qty to available
  const ws = stockItems.value.find(s => s.productId === res.productId && s.warehouseLocation === res.warehouseLocation)
  if (ws) {
    ws.reservedQty -= res.qty
    ws.availableQty += res.qty
  }
  showReleaseModal.value = false
  releasingReservation.value = null
}

// ── Bulk Create Hold ─────────────────────────────────────────
const showCreateHoldModal = ref(false)

interface BulkHoldItem {
  productId: string; productSku: string; productName: string
  warehouseLocation: WarehouseLocation
  availableQty: number
  holdQty: number
  selected: boolean
  poCreated?: string
}

const holdMeta = ref({
  source: 'quote' as ReservationSource,
  sourceRef: '',
  sourceLabel: '',
  customerName: '',
  notes: '',
})
const holdItems = ref<BulkHoldItem[]>([])
const holdItemSearch = ref('')
const showHoldItemDropdown = ref(false)

function openCreateHold(product?: AggregatedProduct) {
  holdMeta.value = { source: 'quote', sourceRef: '', sourceLabel: '', customerName: '', notes: '' }
  holdItems.value = []
  holdItemSearch.value = ''
  if (product) {
    for (const ws of product.warehouseStock) {
      if (ws.availableQty > 0) {
        holdItems.value.push({
          productId: product.productId, productSku: product.productSku, productName: product.productName,
          warehouseLocation: ws.warehouseLocation, availableQty: ws.availableQty,
          holdQty: 1, selected: true,
        })
      }
    }
  }
  showCreateHoldModal.value = true
}

function openBulkHold() {
  holdMeta.value = { source: 'quote', sourceRef: '', sourceLabel: '', customerName: '', notes: '' }
  holdItems.value = []
  holdItemSearch.value = ''
  quoteSearchQuery.value = ''
  selectedQuote.value = null
  showCreateHoldModal.value = true
}

// ── Quote search for hold ────────────────────────────────────
const quoteSearchQuery = ref('')
const showQuoteDropdown = ref(false)
const selectedQuote = ref<Quote | null>(null)

const acceptedQuotes = computed(() => quotesStore.quotes.filter(q => q.status === 'accepted'))

const quoteSearchResults = computed(() => {
  const q = quoteSearchQuery.value.toLowerCase().trim()
  if (!q) return acceptedQuotes.value.slice(0, 8)
  return acceptedQuotes.value.filter(qt =>
    qt.quoteNumber.toLowerCase().includes(q) ||
    qt.customerName.toLowerCase().includes(q) ||
    qt.notes.toLowerCase().includes(q),
  ).slice(0, 8)
})

function selectQuoteForHold(quote: Quote) {
  selectedQuote.value = quote
  holdMeta.value.sourceRef = quote.quoteNumber
  holdMeta.value.sourceLabel = `${quote.customerName} — ${quote.notes.slice(0, 60)}`
  holdMeta.value.customerName = quote.customerName
  quoteSearchQuery.value = ''
  showQuoteDropdown.value = false
  loadItemsFromQuote(quote)
}

function loadItemsFromQuote(quote: Quote) {
  holdItems.value = []
  const materialItems = quote.lineItems.filter(li => li.category === 'materials' && li.sku)
  for (const li of materialItems) {
    const matchingStock = stockItems.value.filter(ws => ws.productSku === li.sku && ws.availableQty > 0)
    if (matchingStock.length > 0) {
      let remaining = li.quantity
      for (const ws of matchingStock) {
        if (remaining <= 0) break
        const holdQty = Math.min(remaining, ws.availableQty)
        holdItems.value.push({
          productId: ws.productId, productSku: ws.productSku, productName: ws.productName,
          warehouseLocation: ws.warehouseLocation, availableQty: ws.availableQty,
          holdQty, selected: true,
        })
        remaining -= holdQty
      }
    } else {
      holdItems.value.push({
        productId: li.productId || '', productSku: li.sku || '', productName: li.description,
        warehouseLocation: 'riyadh-main', availableQty: 0,
        holdQty: li.quantity, selected: true,
      })
    }
  }
}

function clearSelectedQuote() {
  selectedQuote.value = null
  holdMeta.value.sourceRef = ''
  holdMeta.value.sourceLabel = ''
  holdMeta.value.customerName = ''
  holdItems.value = []
}

function delayHideQuoteDropdown() { window.setTimeout(() => { showQuoteDropdown.value = false }, 200) }

const holdSearchResults = computed(() => {
  const q = holdItemSearch.value.toLowerCase().trim()
  if (!q) return []
  const existingKeys = new Set(holdItems.value.map(i => i.productSku + '|' + i.warehouseLocation))
  return stockItems.value
    .filter(ws => {
      if (ws.availableQty <= 0) return false
      if (existingKeys.has(ws.productSku + '|' + ws.warehouseLocation)) return false
      return ws.productSku.toLowerCase().includes(q) || ws.productName.toLowerCase().includes(q)
    })
    .slice(0, 8)
})

function addHoldItem(ws: WarehouseStock) {
  holdItems.value.push({
    productId: ws.productId, productSku: ws.productSku, productName: ws.productName,
    warehouseLocation: ws.warehouseLocation, availableQty: ws.availableQty,
    holdQty: 1, selected: true,
  })
  holdItemSearch.value = ''
  showHoldItemDropdown.value = false
}

function removeHoldItem(idx: number) { holdItems.value.splice(idx, 1) }

const holdSelectedItems = computed(() => holdItems.value.filter(i => i.selected && i.holdQty > 0))
const holdTotalUnits = computed(() => holdSelectedItems.value.reduce((s, i) => s + i.holdQty, 0))
const holdHasErrors = computed(() => holdSelectedItems.value.some(i => i.holdQty > i.availableQty && !i.poCreated))

function confirmBulkHold() {
  if (!holdMeta.value.sourceRef || !holdMeta.value.customerName || holdSelectedItems.value.length === 0 || holdHasErrors.value) return
  const now = new Date().toISOString()
  for (const item of holdSelectedItems.value) {
    const poRef = item.poCreated
    const shortfall = getShortfall(item)
    const fromStockQty = Math.min(item.holdQty, item.availableQty)

    reservations.value.unshift({
      id: uid(), productId: item.productId, productSku: item.productSku,
      productName: item.productName,
      warehouseLocation: item.warehouseLocation,
      qty: item.holdQty,
      source: holdMeta.value.source,
      sourceRef: holdMeta.value.sourceRef,
      sourceLabel: holdMeta.value.sourceLabel || holdMeta.value.sourceRef,
      customerName: holdMeta.value.customerName,
      reservedBy: 'Current User', reservedAt: now,
      status: 'active',
      notes: poRef ? `${fromStockQty} from stock, ${shortfall} on ${poRef}` : (holdMeta.value.notes || undefined),
    })

    if (fromStockQty > 0) {
      movements.value.unshift({
        id: uid(), productId: item.productId, productSku: item.productSku,
        productName: item.productName, movementType: 'allocation',
        qty: fromStockQty, fromWarehouse: item.warehouseLocation,
        reference: holdMeta.value.sourceRef,
        reason: `Hold created for ${holdMeta.value.customerName}: ${holdMeta.value.sourceLabel || holdMeta.value.sourceRef}`,
        performedBy: 'Current User', performedAt: now,
      })
      const ws = stockItems.value.find(s => s.productId === item.productId && s.warehouseLocation === item.warehouseLocation)
      if (ws) { ws.reservedQty += fromStockQty; ws.availableQty -= fromStockQty }
    }
  }
  showCreateHoldModal.value = false
}

function delayHideHoldDropdown() { window.setTimeout(() => { showHoldItemDropdown.value = false }, 200) }

// ── Quick PO from Hold (for zero-stock items) ────────────────
const showQuickPOModal = ref(false)

interface QuickPOItem {
  productId: string; productSku: string; productName: string
  manufacturerName: string
  orderQty: number; unitCost: number; leadTimeDays: number
}

const quickPOMeta = ref({
  supplierId: '', supplierName: '',
  currency: 'SAR' as Currency,
  notes: '',
  expectedDelivery: '',
})
const quickPOItems = ref<QuickPOItem[]>([])

const holdMissingItems = computed(() => holdItems.value.filter(i => i.selected && i.holdQty > i.availableQty))

function getShortfall(item: BulkHoldItem): number {
  return Math.max(0, item.holdQty - item.availableQty)
}

function openQuickPO(singleItem?: BulkHoldItem) {
  const items = singleItem ? [singleItem] : holdMissingItems.value
  if (items.length === 0) return

  quickPOItems.value = items.map(i => {
    const siMatch = procStore.supplierItems.find(si => si.productSku === i.productSku)
    return {
      productId: i.productId, productSku: i.productSku, productName: i.productName,
      manufacturerName: siMatch?.manufacturerName || '',
      orderQty: getShortfall(i), unitCost: siMatch?.latestCost || 0, leadTimeDays: siMatch?.leadTimeDays || 21,
    }
  })

  const today = new Date()
  const maxLead = Math.max(...quickPOItems.value.map(i => i.leadTimeDays))
  const deliveryDate = new Date(today.getTime() + maxLead * 86400000)
  quickPOMeta.value = {
    supplierId: '', supplierName: '',
    currency: 'SAR',
    notes: holdMeta.value.sourceRef ? `Items for hold: ${holdMeta.value.sourceRef} — ${holdMeta.value.customerName}` : '',
    expectedDelivery: deliveryDate.toISOString().slice(0, 10),
  }
  showQuickPOModal.value = true
}

function selectQuickPOSupplier(e: Event) {
  const id = (e.target as HTMLSelectElement).value
  const found = mfrStore.suppliers.find(s => s.id === id)
  if (found) { quickPOMeta.value.supplierId = found.id; quickPOMeta.value.supplierName = found.name }
}

const quickPOSubtotal = computed(() => quickPOItems.value.reduce((s, i) => s + i.orderQty * i.unitCost, 0))

function removeQuickPOItem(idx: number) { quickPOItems.value.splice(idx, 1) }

function confirmQuickPO() {
  if (!quickPOMeta.value.supplierName || quickPOItems.value.length === 0) return
  const now = new Date().toISOString()
  const poNumber = procStore.generatePoNumber()
  const poItems: PurchaseOrderItem[] = quickPOItems.value.map(i => ({
    id: uid(), productId: i.productId, productSku: i.productSku,
    productName: i.productName, manufacturerName: i.manufacturerName,
    quantity: i.orderQty, unitCost: i.unitCost, total: i.orderQty * i.unitCost,
    receivedQty: 0, leadTimeDays: i.leadTimeDays,
  }))
  const subtotal = poItems.reduce((s, i) => s + i.total, 0)
  const po: PurchaseOrder = {
    id: uid(), poNumber,
    supplierName: quickPOMeta.value.supplierName,
    supplierId: quickPOMeta.value.supplierId || undefined,
    status: 'draft',
    items: poItems,
    subtotal,
    shippingCost: 0, customsDuty: 0,
    total: subtotal,
    currency: quickPOMeta.value.currency,
    expectedDelivery: quickPOMeta.value.expectedDelivery,
    sourceQuoteId: selectedQuote.value?.id || undefined,
    sourceQuoteNumber: selectedQuote.value?.quoteNumber || undefined,
    notes: quickPOMeta.value.notes,
    createdAt: now, updatedAt: now,
  }
  procStore.addPurchaseOrder(po)

  for (const item of quickPOItems.value) {
    const holdItem = holdItems.value.find(hi => hi.productSku === item.productSku && hi.holdQty > hi.availableQty && !hi.poCreated)
    if (holdItem) {
      holdItem.poCreated = poNumber
    }
  }

  showQuickPOModal.value = false
}

// ── Bulk Movement ────────────────────────────────────────────
const showBulkMoveModal = ref(false)

interface BulkMoveItem {
  productId: string; productSku: string; productName: string
  currentWarehouse: WarehouseLocation; currentAvailable: number
  moveQty: number; selected: boolean
}

const moveMeta = ref({
  movementType: 'transfer' as MovementType,
  toWarehouse: 'jeddah-branch' as WarehouseLocation,
  fromWarehouse: 'riyadh-main' as WarehouseLocation,
  reason: '',
  notes: '',
})
const moveItems = ref<BulkMoveItem[]>([])
const moveItemSearch = ref('')
const showMoveItemDropdown = ref(false)

function openBulkMove() {
  moveMeta.value = { movementType: 'transfer', toWarehouse: 'jeddah-branch', fromWarehouse: 'riyadh-main', reason: '', notes: '' }
  moveItems.value = []
  moveItemSearch.value = ''
  showBulkMoveModal.value = true
}

const moveSearchResults = computed(() => {
  const q = moveItemSearch.value.toLowerCase().trim()
  if (!q) return []
  const from = moveMeta.value.fromWarehouse
  const existingKeys = new Set(moveItems.value.map(i => i.productSku))
  return stockItems.value
    .filter(ws => {
      if (ws.warehouseLocation !== from) return false
      if (ws.availableQty <= 0) return false
      if (existingKeys.has(ws.productSku)) return false
      return ws.productSku.toLowerCase().includes(q) || ws.productName.toLowerCase().includes(q)
    })
    .slice(0, 8)
})

function addMoveItem(ws: WarehouseStock) {
  moveItems.value.push({
    productId: ws.productId, productSku: ws.productSku, productName: ws.productName,
    currentWarehouse: ws.warehouseLocation, currentAvailable: ws.availableQty,
    moveQty: 1, selected: true,
  })
  moveItemSearch.value = ''
  showMoveItemDropdown.value = false
}

function removeMoveItem(idx: number) { moveItems.value.splice(idx, 1) }

function loadFromWarehouseItems() {
  const from = moveMeta.value.fromWarehouse
  const existingKeys = new Set(moveItems.value.map(i => i.productSku))
  for (const ws of stockItems.value) {
    if (ws.warehouseLocation !== from || ws.availableQty <= 0 || existingKeys.has(ws.productSku)) continue
    moveItems.value.push({
      productId: ws.productId, productSku: ws.productSku, productName: ws.productName,
      currentWarehouse: ws.warehouseLocation, currentAvailable: ws.availableQty,
      moveQty: 0, selected: false,
    })
  }
}

const moveSelectedItems = computed(() => moveItems.value.filter(i => i.selected && i.moveQty > 0))
const moveTotalUnits = computed(() => moveSelectedItems.value.reduce((s, i) => s + i.moveQty, 0))
const moveHasErrors = computed(() => moveSelectedItems.value.some(i => i.moveQty > i.currentAvailable))

function confirmBulkMove() {
  if (moveSelectedItems.value.length === 0 || moveHasErrors.value) return
  const now = new Date().toISOString()
  const refId = `MV-${Date.now().toString(36).toUpperCase()}`
  for (const item of moveSelectedItems.value) {
    const mt = moveMeta.value.movementType
    movements.value.unshift({
      id: uid(), productId: item.productId, productSku: item.productSku,
      productName: item.productName,
      movementType: mt,
      qty: mt === 'adjustment' || mt === 'write-off' ? -item.moveQty : item.moveQty,
      fromWarehouse: moveMeta.value.fromWarehouse,
      toWarehouse: mt === 'transfer' ? moveMeta.value.toWarehouse : undefined,
      reference: refId,
      reason: moveMeta.value.reason || movementTypeLabel(mt),
      performedBy: 'Current User', performedAt: now,
      notes: moveMeta.value.notes || undefined,
    })
    const fromWs = stockItems.value.find(s => s.productId === item.productId && s.warehouseLocation === moveMeta.value.fromWarehouse)
    if (fromWs) {
      fromWs.onHandQty -= item.moveQty
      fromWs.availableQty -= item.moveQty
      fromWs.totalValue = fromWs.onHandQty * fromWs.unitCost
    }
    if (mt === 'transfer') {
      let toWs = stockItems.value.find(s => s.productId === item.productId && s.warehouseLocation === moveMeta.value.toWarehouse)
      if (!toWs) {
        const newWs: WarehouseStock = {
          id: uid(), productId: item.productId, productSku: item.productSku, productName: item.productName,
          warehouseLocation: moveMeta.value.toWarehouse, onHandQty: 0, reservedQty: 0, availableQty: 0,
          unitCost: fromWs?.unitCost ?? 0, totalValue: 0,
          createdAt: now, updatedAt: now,
        }
        stockItems.value.push(newWs)
        toWs = newWs
      }
      toWs.onHandQty += item.moveQty
      toWs.availableQty += item.moveQty
      toWs.totalValue = toWs.onHandQty * toWs.unitCost
    }
  }
  showBulkMoveModal.value = false
}

function delayHideMoveDropdown() { window.setTimeout(() => { showMoveItemDropdown.value = false }, 200) }
</script>

<template>
  <div class="inventory-page">
    <!-- Page Header -->
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Inventory</h1>
        <p class="page-header-subtitle">Warehouse stock, price history &amp; reorder tracking</p>
      </div>
      <div class="page-header-actions">
        <button class="btn btn-secondary btn-sm" @click="openBulkMove"><ArrowRightLeft :size="15" /> Bulk Movement</button>
        <button class="btn btn-primary btn-sm" @click="openBulkHold"><Lock :size="15" /> Create Hold</button>
      </div>
    </div>

    <!-- KPI Row -->
    <div class="kpi-row kpi-row--6">
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--primary"><Package :size="20" /></div>
        <div>
          <div class="kpi-label">Total SKUs</div>
          <div class="kpi-value">{{ totalSKUs }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--info"><Boxes :size="20" /></div>
        <div>
          <div class="kpi-label">Total On-Hand</div>
          <div class="kpi-value">{{ totalOnHand.toLocaleString() }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--neutral"><Lock :size="20" /></div>
        <div>
          <div class="kpi-label">Reserved</div>
          <div class="kpi-value">{{ totalReserved.toLocaleString() }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--success"><DollarSign :size="20" /></div>
        <div>
          <div class="kpi-label">Total Value</div>
          <div class="kpi-value">SAR {{ formatSAR(totalValue) }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--warning"><AlertTriangle :size="20" /></div>
        <div>
          <div class="kpi-label">Low Stock</div>
          <div class="kpi-value">{{ lowStockCount }}</div>
        </div>
      </div>
      <div class="kpi-card">
        <div class="kpi-icon kpi-icon--danger"><AlertCircle :size="20" /></div>
        <div>
          <div class="kpi-label">Out of Stock</div>
          <div class="kpi-value">{{ outOfStockCount }}</div>
        </div>
      </div>
    </div>

    <!-- Toolbar: Warehouse tabs + search + filters -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="tab-row">
          <button :class="['tab-btn', activeWarehouse === 'all' && 'tab-btn--active']" @click="activeWarehouse = 'all'">All Locations</button>
          <button v-for="(label, key) in warehouseLabels" :key="key" :class="['tab-btn', activeWarehouse === key && 'tab-btn--active']" @click="activeWarehouse = key as WarehouseLocation">{{ label }}</button>
        </div>
        <div class="toolbar-filters">
          <div class="search-input toolbar-search">
            <Search :size="18" class="search-icon" />
            <input v-model="searchQuery" type="text" class="form-input" placeholder="Search by SKU, name, or manufacturer..." />
          </div>
          <select v-model="categoryFilter" class="form-select toolbar-select">
            <option value="">All Categories</option>
            <option v-for="cat in allCategories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
          <select v-model="stockStatusFilter" class="form-select toolbar-select">
            <option value="">All Stock Status</option>
            <option value="in-stock">In Stock</option>
            <option value="low-stock">Low Stock</option>
            <option value="out-of-stock">Out of Stock</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Inventory Table -->
    <div v-if="filteredProducts.length" class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>SKU</th>
            <th>Product Name</th>
            <th>Manufacturer</th>
            <th>Category</th>
            <th class="text-right">On-Hand</th>
            <th class="text-right">Reserved</th>
            <th class="text-center">Holds</th>
            <th class="text-right">Available</th>
            <th>Stock Level</th>
            <th class="text-right">Avg Cost</th>
            <th class="text-right">Total Value</th>
            <th class="text-center">Incoming</th>
            <th class="text-center">Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in filteredProducts" :key="product.productSku" class="inv-row" @click="openItemDetail(product)">
            <td class="text-mono font-medium">{{ product.productSku }}</td>
            <td class="font-medium text-dark">{{ product.productName }}</td>
            <td class="text-muted" style="font-size:0.78rem">{{ product.manufacturer }}</td>
            <td><span class="badge badge-gray">{{ product.category }}</span></td>
            <td class="text-right">{{ product.totalOnHand }}</td>
            <td class="text-right text-muted">{{ product.totalReserved }}</td>
            <td class="text-center">
              <span v-if="getActiveHoldsForProduct(product.productSku).count > 0" class="inv-hold-badge" :title="getActiveHoldsForProduct(product.productSku).count + ' active hold(s)'">
                <Lock :size="10" /> {{ getActiveHoldsForProduct(product.productSku).qty }}
              </span>
              <span v-else class="text-muted">—</span>
            </td>
            <td class="text-right">
              <span :class="['font-semibold', stockStatusClass(product.stockStatus)]">{{ product.totalAvailable }}</span>
            </td>
            <td>
              <div class="stock-bar-wrapper">
                <div class="stock-bar">
                  <div class="stock-bar-fill" :class="stockStatusClass(product.stockStatus)" :style="{ width: stockPercent(product.totalOnHand, product.totalAvailable) + '%' }" />
                </div>
                <span class="stock-bar-label">{{ stockPercent(product.totalOnHand, product.totalAvailable) }}%</span>
              </div>
            </td>
            <td class="text-right whitespace-nowrap" style="font-size:0.78rem">SAR {{ formatSAR(product.avgUnitCost) }}</td>
            <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(product.totalValue) }}</td>
            <td class="text-center">
              <span v-if="product.incomingQty > 0" class="badge badge-info">
                <Truck :size="11" /> {{ product.incomingQty }}
              </span>
              <span v-else class="text-muted">—</span>
            </td>
            <td class="text-center">
              <span :class="['badge badge-dot', stockStatusClass(product.stockStatus) === 'stock-green' ? 'badge-success' : stockStatusClass(product.stockStatus) === 'stock-orange' ? 'badge-warning' : 'badge-danger']">
                {{ stockStatusLabel(product.stockStatus) }}
              </span>
            </td>
            <td>
              <button class="btn btn-ghost btn-icon btn-sm" title="View Details" @click.stop="openItemDetail(product)"><Eye :size="14" /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="empty-state">
      <AlertTriangle :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No inventory found</h3>
      <p class="empty-state-text">Try adjusting your filters or search query.</p>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- Item Detail Modal                                      -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="showDetailModal && detailProduct" class="modal-backdrop" @click.self="showDetailModal = false">
      <div class="modal inv-detail-modal">
        <!-- Header -->
        <div class="inv-detail-header">
          <div class="inv-detail-header-left">
            <div class="inv-detail-icon"><Package :size="24" /></div>
            <div>
              <h2 class="inv-detail-title">{{ detailProduct.productName }}</h2>
              <div class="inv-detail-meta">
                <span class="inv-detail-sku">{{ detailProduct.productSku }}</span>
                <span class="inv-detail-sep">|</span>
                <span class="inv-detail-mfr">{{ detailProduct.manufacturer }}</span>
                <span class="inv-detail-sep">|</span>
                <span class="badge badge-gray" style="font-size:0.72rem">{{ detailProduct.category }}</span>
                <span :class="['badge badge-dot', detailProduct.stockStatus === 'in-stock' ? 'badge-success' : detailProduct.stockStatus === 'low-stock' ? 'badge-warning' : 'badge-danger']" style="font-size:0.72rem">
                  {{ stockStatusLabel(detailProduct.stockStatus) }}
                </span>
              </div>
            </div>
          </div>
          <!-- Quick-glance numbers in header -->
          <div class="inv-detail-header-right">
            <div class="inv-hdr-num">
              <span class="inv-hdr-num-val" :class="stockStatusClass(detailProduct.stockStatus)">{{ detailProduct.totalAvailable }}</span>
              <span class="inv-hdr-num-label">Available</span>
            </div>
            <div class="inv-hdr-num">
              <span class="inv-hdr-num-val">{{ detailProduct.totalOnHand }}</span>
              <span class="inv-hdr-num-label">On-Hand</span>
            </div>
            <div class="inv-hdr-num">
              <span class="inv-hdr-num-val">SAR {{ formatSAR(detailProduct.avgUnitCost) }}</span>
              <span class="inv-hdr-num-label">Avg Cost</span>
            </div>
            <button class="modal-close" @click="showDetailModal = false"><X :size="20" /></button>
          </div>
        </div>

        <!-- Detail Tabs -->
        <div class="inv-detail-tabs">
          <button :class="['inv-dtab', detailTab === 'overview' && 'inv-dtab--active']" @click="detailTab = 'overview'"><Boxes :size="15" /> Overview</button>
          <button :class="['inv-dtab', detailTab === 'reservations' && 'inv-dtab--active']" @click="detailTab = 'reservations'"><Lock :size="15" /> Holds <span :class="['inv-dtab-count', detailActiveReservations.length > 0 ? 'inv-dtab-count--warn' : '']">{{ detailActiveReservations.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'movements' && 'inv-dtab--active']" @click="detailTab = 'movements'"><ArrowRightLeft :size="15" /> Movements <span class="inv-dtab-count">{{ detailMovements.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'price-history' && 'inv-dtab--active']" @click="detailTab = 'price-history'"><History :size="15" /> Price History <span class="inv-dtab-count">{{ detailPriceHistory.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'orders' && 'inv-dtab--active']" @click="detailTab = 'orders'"><ShoppingCart :size="15" /> Orders <span class="inv-dtab-count">{{ detailPastOrders.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'supplier-quotes' && 'inv-dtab--active']" @click="detailTab = 'supplier-quotes'"><FileText :size="15" /> Supplier Quotes <span class="inv-dtab-count">{{ detailSQs.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'receiving' && 'inv-dtab--active']" @click="detailTab = 'receiving'"><PackageCheck :size="15" /> Receiving <span class="inv-dtab-count">{{ detailReceiveHistory.length }}</span></button>
          <button :class="['inv-dtab', detailTab === 'suppliers' && 'inv-dtab--active']" @click="detailTab = 'suppliers'"><Building2 :size="15" /> Suppliers <span class="inv-dtab-count">{{ detailSuppliers.length }}</span></button>
        </div>

        <div class="modal-body inv-detail-body">

          <!-- ── OVERVIEW TAB ──────────────────────────── -->
          <div v-if="detailTab === 'overview'">
            <!-- Stock Summary Cards -->
            <div class="inv-summary-cards">
              <div class="inv-sum-card">
                <div class="inv-sum-label">Total On-Hand</div>
                <div class="inv-sum-value">{{ detailProduct.totalOnHand }}</div>
              </div>
              <div class="inv-sum-card">
                <div class="inv-sum-label">Reserved</div>
                <div class="inv-sum-value">{{ detailProduct.totalReserved }}</div>
              </div>
              <div class="inv-sum-card">
                <div class="inv-sum-label">Available</div>
                <div class="inv-sum-value" :class="stockStatusClass(detailProduct.stockStatus)">{{ detailProduct.totalAvailable }}</div>
              </div>
              <div class="inv-sum-card">
                <div class="inv-sum-label">Avg Unit Cost</div>
                <div class="inv-sum-value">SAR {{ formatSAR(detailProduct.avgUnitCost) }}</div>
              </div>
              <div class="inv-sum-card">
                <div class="inv-sum-label">Total Value</div>
                <div class="inv-sum-value">SAR {{ formatSAR(detailProduct.totalValue) }}</div>
              </div>
              <div class="inv-sum-card">
                <div class="inv-sum-label">Incoming (PO)</div>
                <div class="inv-sum-value" :class="detailProduct.incomingQty > 0 ? 'stock-green' : ''">{{ detailProduct.incomingQty }}</div>
              </div>
            </div>

            <!-- Warehouse Breakdown -->
            <h4 class="inv-section-title"><MapPin :size="14" /> Stock by Warehouse</h4>
            <div class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Warehouse</th>
                    <th class="text-right">On-Hand</th>
                    <th class="text-right">Reserved</th>
                    <th class="text-right">Available</th>
                    <th>Stock Level</th>
                    <th class="text-right">Reorder Level</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-right">Value</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="ws in detailProduct.warehouseStock" :key="ws.id">
                    <td><span :class="['badge', warehouseBadge[ws.warehouseLocation]]">{{ warehouseLabels[ws.warehouseLocation] }}</span></td>
                    <td class="text-right">{{ ws.onHandQty }}</td>
                    <td class="text-right text-muted">{{ ws.reservedQty }}</td>
                    <td class="text-right font-semibold">{{ ws.availableQty }}</td>
                    <td>
                      <div class="stock-bar-wrapper">
                        <div class="stock-bar"><div class="stock-bar-fill" :class="ws.availableQty > (ws.reorderLevel ?? 0) ? 'stock-green' : ws.availableQty > 0 ? 'stock-orange' : 'stock-red'" :style="{ width: stockPercent(ws.onHandQty, ws.availableQty) + '%' }" /></div>
                        <span class="stock-bar-label">{{ stockPercent(ws.onHandQty, ws.availableQty) }}%</span>
                      </div>
                    </td>
                    <td class="text-right">
                      <span :class="ws.availableQty <= (ws.reorderLevel ?? 0) ? 'text-danger font-semibold' : 'text-muted'">{{ ws.reorderLevel ?? '—' }}</span>
                    </td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(ws.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(ws.totalValue) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Cost Snapshot -->
            <h4 class="inv-section-title"><BarChart3 :size="14" /> Cost Snapshot</h4>
            <div class="inv-cost-grid">
              <div class="inv-cost-item">
                <span class="inv-cost-label">Average Cost (all sources)</span>
                <span class="inv-cost-value">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailAvgCost) : '—' }}</span>
              </div>
              <div class="inv-cost-item">
                <span class="inv-cost-label">Lowest Quoted / Ordered</span>
                <span class="inv-cost-value stock-green">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailLowestCost) : '—' }}</span>
              </div>
              <div class="inv-cost-item">
                <span class="inv-cost-label">Highest Quoted / Ordered</span>
                <span class="inv-cost-value stock-red">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailHighestCost) : '—' }}</span>
              </div>
              <div class="inv-cost-item">
                <span class="inv-cost-label">Last Landing Cost</span>
                <span class="inv-cost-value">{{ detailLastLandingCost ? 'SAR ' + formatSAR(detailLastLandingCost) : 'No GR data' }}</span>
              </div>
              <div class="inv-cost-item">
                <span class="inv-cost-label">Best Current Supplier Price</span>
                <span v-if="procStore.getBestSupplierPrice(detailProduct.productSku)" class="inv-cost-value stock-green">
                  SAR {{ formatSAR(procStore.getBestSupplierPrice(detailProduct.productSku)!.cost) }}
                  <span class="text-muted" style="font-size:0.7rem">({{ procStore.getBestSupplierPrice(detailProduct.productSku)!.supplierName }})</span>
                </span>
                <span v-else class="inv-cost-value text-muted">No active SQ</span>
              </div>
              <div class="inv-cost-item">
                <span class="inv-cost-label">Price Records</span>
                <span class="inv-cost-value">{{ detailPriceHistory.length }} entries</span>
              </div>
            </div>
          </div>

          <!-- ── RESERVATIONS (HOLDS) TAB ─────────────── -->
          <div v-if="detailTab === 'reservations'">
            <div class="inv-holds-toolbar">
              <div>
                <span class="inv-holds-summary">
                  <Lock :size="14" /> <strong>{{ detailActiveReservations.length }}</strong> active hold{{ detailActiveReservations.length !== 1 ? 's' : '' }}
                  totalling <strong>{{ detailActiveReservations.reduce((s, r) => s + r.qty, 0) }}</strong> units
                </span>
              </div>
              <button class="btn btn-primary btn-sm" @click="openCreateHold(detailProduct!)">
                <Plus :size="14" /> Create Hold
              </button>
            </div>

            <div v-if="detailReservations.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Status</th>
                    <th>Source</th>
                    <th>Reference</th>
                    <th>Description</th>
                    <th>Customer</th>
                    <th>Warehouse</th>
                    <th class="text-center">Qty</th>
                    <th>Reserved By</th>
                    <th>Date</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="res in detailReservations" :key="res.id" :class="res.status !== 'active' ? 'inv-row--faded' : ''">
                    <td>
                      <span :class="['badge badge-dot', reservationStatusBadge(res.status)]">
                        {{ res.status === 'active' ? 'Active' : res.status === 'released' ? 'Released' : 'Fulfilled' }}
                      </span>
                    </td>
                    <td><span :class="['badge', reservationSourceBadge(res.source)]">{{ reservationSourceLabel(res.source) }}</span></td>
                    <td class="text-mono font-medium" style="font-size:0.78rem">{{ res.sourceRef }}</td>
                    <td class="font-medium" style="font-size:0.78rem; max-width: 200px;">{{ res.sourceLabel }}</td>
                    <td class="font-medium">{{ res.customerName }}</td>
                    <td><span :class="['badge', warehouseBadge[res.warehouseLocation]]">{{ warehouseLabels[res.warehouseLocation] }}</span></td>
                    <td class="text-center font-bold">{{ res.qty }}</td>
                    <td class="text-muted" style="font-size:0.75rem"><User :size="12" /> {{ res.reservedBy }}</td>
                    <td class="whitespace-nowrap text-muted" style="font-size:0.75rem">{{ formatDate(res.reservedAt) }}</td>
                    <td>
                      <div class="table-actions">
                        <button v-if="res.status === 'active'" class="btn btn-ghost btn-sm" style="font-size:0.72rem; gap:4px" title="Release Hold" @click="openReleaseHold(res)">
                          <Unlock :size="13" /> Release
                        </button>
                        <span v-else-if="res.releaseDate" class="text-muted" style="font-size:0.68rem; display: flex; flex-direction: column; align-items:flex-end;">
                          <span>{{ formatDate(res.releaseDate) }}</span>
                          <span style="font-size:0.62rem; max-width:120px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap">{{ res.releaseReason }}</span>
                        </span>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <Lock :size="36" class="text-muted" />
              <p class="text-muted">No holds or reservations for this item.</p>
              <button class="btn btn-primary btn-sm" @click="openCreateHold(detailProduct!)"><Plus :size="14" /> Create Hold</button>
            </div>
          </div>

          <!-- ── MOVEMENTS / ACTIVITY TAB ──────────────── -->
          <div v-if="detailTab === 'movements'">
            <div v-if="detailMovements.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Type</th>
                    <th class="text-center">Qty</th>
                    <th>From</th>
                    <th>To</th>
                    <th>Reference</th>
                    <th>Reason</th>
                    <th>Performed By</th>
                    <th>Notes</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="mv in detailMovements" :key="mv.id">
                    <td class="whitespace-nowrap" style="font-size:0.78rem">{{ formatDate(mv.performedAt) }}</td>
                    <td><span :class="['badge', movementTypeBadge(mv.movementType)]">{{ movementTypeLabel(mv.movementType) }}</span></td>
                    <td class="text-center font-bold" :class="mv.qty < 0 ? 'stock-red' : 'stock-green'">
                      {{ mv.qty > 0 ? '+' : '' }}{{ mv.qty }}
                    </td>
                    <td>
                      <span v-if="mv.fromWarehouse" :class="['badge', warehouseBadge[mv.fromWarehouse]]" style="font-size:0.68rem">{{ warehouseLabels[mv.fromWarehouse] }}</span>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td>
                      <span v-if="mv.toWarehouse" :class="['badge', warehouseBadge[mv.toWarehouse]]" style="font-size:0.68rem">{{ warehouseLabels[mv.toWarehouse] }}</span>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td class="text-mono" style="font-size:0.72rem">{{ mv.reference || '—' }}</td>
                    <td style="font-size:0.78rem; max-width: 250px;">{{ mv.reason }}</td>
                    <td class="text-muted" style="font-size:0.75rem"><User :size="12" /> {{ mv.performedBy }}</td>
                    <td class="text-muted" style="font-size:0.72rem; max-width: 150px;">{{ mv.notes || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <ArrowRightLeft :size="36" class="text-muted" />
              <p class="text-muted">No inventory movements recorded for this item.</p>
            </div>
          </div>

          <!-- ── PRICE HISTORY TAB ─────────────────────── -->
          <div v-if="detailTab === 'price-history'">
            <!-- Stats -->
            <div class="inv-ph-stats">
              <div class="inv-ph-stat">
                <div class="inv-ph-stat-label">Average Cost</div>
                <div class="inv-ph-stat-value">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailAvgCost) : '—' }}</div>
              </div>
              <div class="inv-ph-stat">
                <div class="inv-ph-stat-label">Lowest</div>
                <div class="inv-ph-stat-value stock-green">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailLowestCost) : '—' }}</div>
              </div>
              <div class="inv-ph-stat">
                <div class="inv-ph-stat-label">Highest</div>
                <div class="inv-ph-stat-value stock-red">SAR {{ detailPriceHistory.length > 0 ? formatSAR(detailHighestCost) : '—' }}</div>
              </div>
              <div class="inv-ph-stat">
                <div class="inv-ph-stat-label">Last Landing Cost</div>
                <div class="inv-ph-stat-value">{{ detailLastLandingCost ? 'SAR ' + formatSAR(detailLastLandingCost) : '—' }}</div>
              </div>
            </div>

            <div v-if="detailPriceHistory.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Source</th>
                    <th>Reference</th>
                    <th>Supplier</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-right">Landing Cost</th>
                    <th class="text-center">Qty</th>
                    <th class="text-right">vs Avg</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, idx) in detailPriceHistory" :key="idx">
                    <td class="whitespace-nowrap">{{ formatDate(entry.date) }}</td>
                    <td><span :class="['badge', priceSourceBadge(entry.source)]">{{ entry.source }}</span></td>
                    <td class="text-mono font-medium" style="font-size:0.75rem">{{ entry.sourceRef }}</td>
                    <td class="font-medium" style="font-size:0.78rem">{{ entry.supplier }}</td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(entry.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap">
                      <span v-if="entry.landingCost" class="font-medium">SAR {{ formatSAR(entry.landingCost) }}</span>
                      <span v-else class="text-muted">—</span>
                    </td>
                    <td class="text-center">{{ entry.qty }}</td>
                    <td class="text-right whitespace-nowrap">
                      <span v-if="detailAvgCost > 0" :class="['inv-ph-delta', entry.unitCost < detailAvgCost ? 'stock-green' : entry.unitCost > detailAvgCost ? 'stock-red' : '']">
                        <TrendingDown v-if="entry.unitCost < detailAvgCost" :size="12" />
                        <TrendingUp v-else-if="entry.unitCost > detailAvgCost" :size="12" />
                        <Minus v-else :size="12" />
                        {{ Math.abs(Math.round(((entry.unitCost - detailAvgCost) / detailAvgCost) * 100)) }}%
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <History :size="36" class="text-muted" />
              <p class="text-muted">No price history available for this item.</p>
            </div>
          </div>

          <!-- ── ORDERS TAB ────────────────────────────── -->
          <div v-if="detailTab === 'orders'">
            <div v-if="detailPastOrders.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>PO #</th>
                    <th>Supplier</th>
                    <th>Status</th>
                    <th>Date</th>
                    <th class="text-center">Qty Ordered</th>
                    <th class="text-center">Qty Received</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-right">Line Total</th>
                    <th>Expected Delivery</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, idx) in detailPastOrders" :key="idx">
                    <td class="text-mono font-bold">{{ entry.po.poNumber }}</td>
                    <td class="font-medium">{{ entry.po.supplierName }}</td>
                    <td>
                      <span :class="['badge badge-dot', (poStatusConfig[entry.po.status] || {badge:'badge-gray'}).badge]">
                        {{ (poStatusConfig[entry.po.status] || {label:entry.po.status}).label }}
                      </span>
                    </td>
                    <td class="whitespace-nowrap">{{ formatDate(entry.po.createdAt) }}</td>
                    <td class="text-center">{{ entry.item.quantity }}</td>
                    <td class="text-center">
                      <span :class="entry.item.receivedQty >= entry.item.quantity ? 'stock-green font-semibold' : entry.item.receivedQty > 0 ? 'stock-orange font-semibold' : ''">
                        {{ entry.item.receivedQty }}
                      </span>
                    </td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(entry.item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(entry.item.total) }}</td>
                    <td class="whitespace-nowrap text-muted">{{ entry.po.expectedDelivery ? formatDate(entry.po.expectedDelivery) : '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <ShoppingCart :size="36" class="text-muted" />
              <p class="text-muted">No purchase orders found for this item.</p>
            </div>
          </div>

          <!-- ── SUPPLIER QUOTES TAB ───────────────────── -->
          <div v-if="detailTab === 'supplier-quotes'">
            <div v-if="detailSQs.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Quote #</th>
                    <th>Supplier</th>
                    <th>Status</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-center">Qty</th>
                    <th class="text-right">Total</th>
                    <th class="text-center">Lead Time</th>
                    <th class="text-center">MOQ</th>
                    <th>Valid Until</th>
                    <th>Terms</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, idx) in detailSQs" :key="idx">
                    <td class="text-mono font-bold">{{ entry.sq.sqNumber }}</td>
                    <td class="font-medium">{{ entry.sq.supplierName }}</td>
                    <td>
                      <span :class="['badge badge-dot', (sqStatusConfig[entry.sq.status] || {badge:'badge-gray'}).badge]">
                        {{ (sqStatusConfig[entry.sq.status] || {label:entry.sq.status}).label }}
                      </span>
                    </td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(entry.item.unitCost) }}</td>
                    <td class="text-center">{{ entry.item.quantity }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(entry.item.total) }}</td>
                    <td class="text-center">{{ entry.item.leadTimeDays }}d</td>
                    <td class="text-center">{{ entry.item.moq ?? '—' }}</td>
                    <td class="whitespace-nowrap">{{ formatDate(entry.sq.validUntil) }}</td>
                    <td class="text-muted" style="font-size:0.75rem">{{ entry.sq.paymentTerms || '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <FileText :size="36" class="text-muted" />
              <p class="text-muted">No supplier quotes found for this item.</p>
            </div>
          </div>

          <!-- ── RECEIVING TAB ─────────────────────────── -->
          <div v-if="detailTab === 'receiving'">
            <div v-if="detailReceiveHistory.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>GR #</th>
                    <th>PO #</th>
                    <th>Supplier</th>
                    <th>Receive Date</th>
                    <th class="text-center">Qty</th>
                    <th class="text-right">Unit Cost</th>
                    <th class="text-right">Shipping</th>
                    <th class="text-right">Customs</th>
                    <th class="text-right">Landing Cost</th>
                    <th>Location</th>
                    <th>Condition</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, idx) in detailReceiveHistory" :key="idx">
                    <td class="text-mono font-bold">{{ entry.gr.grNumber }}</td>
                    <td class="text-mono" style="font-size:0.75rem">{{ entry.gr.poNumber }}</td>
                    <td class="font-medium" style="font-size:0.78rem">{{ entry.gr.supplierName }}</td>
                    <td class="whitespace-nowrap">{{ formatDate(entry.gr.receiveDate) }}</td>
                    <td class="text-center font-medium">{{ entry.item.receivedQty }}</td>
                    <td class="text-right whitespace-nowrap">SAR {{ formatSAR(entry.item.unitCost) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(entry.item.shippingAlloc) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">{{ formatSAR(entry.item.customsAlloc) }}</td>
                    <td class="text-right whitespace-nowrap font-bold">SAR {{ formatSAR(entry.item.landingCost) }}</td>
                    <td class="text-muted" style="font-size:0.72rem">{{ entry.item.storageLocation || '—' }}</td>
                    <td>
                      <span :class="['badge', entry.item.condition === 'good' ? 'badge-success' : entry.item.condition === 'damaged' ? 'badge-danger' : 'badge-warning']">
                        {{ entry.item.condition }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <PackageCheck :size="36" class="text-muted" />
              <p class="text-muted">No goods receipts found for this item.</p>
            </div>
          </div>

          <!-- ── SUPPLIERS TAB ─────────────────────────── -->
          <div v-if="detailTab === 'suppliers'">
            <div v-if="detailSuppliers.length" class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Supplier</th>
                    <th>Manufacturer</th>
                    <th class="text-right">Latest Cost</th>
                    <th class="text-right">Previous Cost</th>
                    <th class="text-center">Trend</th>
                    <th class="text-center">MOQ</th>
                    <th class="text-center">Lead Time</th>
                    <th>Reliability</th>
                    <th>Last Quote</th>
                    <th>Last PO</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="si in detailSuppliers" :key="si.id">
                    <td class="font-bold">{{ si.supplierName }}</td>
                    <td class="text-muted">{{ si.manufacturerName }}</td>
                    <td class="text-right whitespace-nowrap font-medium">SAR {{ formatSAR(si.latestCost) }}</td>
                    <td class="text-right whitespace-nowrap text-muted">SAR {{ formatSAR(si.previousCost) }}</td>
                    <td class="text-center">
                      <span v-if="si.costTrend === 'down'" class="stock-green"><TrendingDown :size="14" /></span>
                      <span v-else-if="si.costTrend === 'up'" class="stock-red"><TrendingUp :size="14" /></span>
                      <span v-else class="text-muted"><Minus :size="14" /></span>
                    </td>
                    <td class="text-center">{{ si.moq }}</td>
                    <td class="text-center">{{ si.leadTimeDays }}d</td>
                    <td>
                      <div class="reliability-bar-wrap">
                        <div class="reliability-bar">
                          <div class="reliability-fill" :class="si.reliability >= 90 ? 'stock-green' : si.reliability >= 80 ? 'stock-orange' : 'stock-red'" :style="{ width: si.reliability + '%' }" />
                        </div>
                        <span class="reliability-label">{{ si.reliability }}%</span>
                      </div>
                    </td>
                    <td class="whitespace-nowrap text-muted" style="font-size:0.75rem">{{ si.lastQuoteDate ? formatDate(si.lastQuoteDate) : '—' }}</td>
                    <td class="whitespace-nowrap text-muted" style="font-size:0.75rem">{{ si.lastPODate ? formatDate(si.lastPODate) : '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-empty-tab">
              <Building2 :size="36" class="text-muted" />
              <p class="text-muted">No supplier catalog entries found for this item.</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- Release Hold Modal                                     -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="showReleaseModal && releasingReservation" class="modal-backdrop" @click.self="showReleaseModal = false">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h2 class="modal-title"><Unlock :size="18" /> Release Hold</h2>
          <button class="modal-close" @click="showReleaseModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body">
          <div class="inv-release-info">
            <div class="inv-release-row"><span class="inv-release-label">Product</span><span class="font-medium">{{ releasingReservation.productName }}</span></div>
            <div class="inv-release-row"><span class="inv-release-label">Qty Held</span><span class="font-bold">{{ releasingReservation.qty }} units</span></div>
            <div class="inv-release-row"><span class="inv-release-label">Warehouse</span><span :class="['badge', warehouseBadge[releasingReservation.warehouseLocation]]">{{ warehouseLabels[releasingReservation.warehouseLocation] }}</span></div>
            <div class="inv-release-row"><span class="inv-release-label">Source</span><span class="text-mono" style="font-size:0.78rem">{{ releasingReservation.sourceRef }}</span></div>
            <div class="inv-release-row"><span class="inv-release-label">Customer</span><span>{{ releasingReservation.customerName }}</span></div>
          </div>
          <div class="form-group" style="margin-top: var(--space-4)">
            <label class="form-label">Reason for release</label>
            <textarea v-model="releaseReason" class="form-input" rows="3" placeholder="e.g. Quote expired, customer cancelled, stock no longer needed..." />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showReleaseModal = false">Cancel</button>
          <button class="btn btn-primary" @click="confirmRelease">
            <Unlock :size="14" /> Release {{ releasingReservation.qty }} Units
          </button>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- Bulk Create Hold Modal                                 -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="showCreateHoldModal" class="modal-backdrop" @click.self="showCreateHoldModal = false">
      <div class="modal inv-detail-modal">
        <div class="modal-header">
          <h2 class="modal-title"><Lock :size="18" /> Create Hold</h2>
          <button class="modal-close" @click="showCreateHoldModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body" style="max-height:75vh; overflow-y:auto">
          <!-- Source info -->
          <div class="inv-bulk-section">
            <h4 class="inv-bulk-section-title"><ClipboardList :size="14" /> Hold Details</h4>
            <div class="create-form-row">
              <div class="form-group" style="width:160px">
                <label class="form-label">Source <span class="text-danger">*</span></label>
                <select v-model="holdMeta.source" class="form-select" @change="clearSelectedQuote()">
                  <option value="quote">Won Quote</option>
                  <option value="project">Project</option>
                  <option value="manual">Manual</option>
                </select>
              </div>

              <!-- QUOTE SEARCH (shown when source = quote) -->
              <div v-if="holdMeta.source === 'quote'" class="form-group" style="flex:2">
                <label class="form-label">Select Accepted Quote <span class="text-danger">*</span></label>
                <div v-if="!selectedQuote" class="inv-quote-search-wrap">
                  <div class="search-input" style="width:100%">
                    <Search :size="14" class="search-icon" />
                    <input v-model="quoteSearchQuery" type="text" class="form-input" placeholder="Search by quote #, customer, or description..." @focus="showQuoteDropdown = true" @blur="delayHideQuoteDropdown" @input="showQuoteDropdown = true" />
                  </div>
                  <div v-if="showQuoteDropdown && quoteSearchResults.length" class="inv-quote-dropdown">
                    <button v-for="qt in quoteSearchResults" :key="qt.id" class="inv-quote-dd-item" @mousedown.prevent="selectQuoteForHold(qt)">
                      <div class="inv-quote-dd-top">
                        <span class="text-mono font-bold">{{ qt.quoteNumber }}</span>
                        <span class="badge badge-success" style="font-size:0.62rem">Accepted</span>
                        <span class="font-medium" style="font-size:0.78rem">SAR {{ formatSAR(qt.total) }}</span>
                      </div>
                      <div class="inv-quote-dd-bottom">
                        <span class="font-medium">{{ qt.customerName }}</span>
                        <span class="text-muted">{{ qt.lineItems.filter(l => l.category === 'materials').length }} material items</span>
                      </div>
                    </button>
                    <div v-if="quoteSearchResults.length === 0" class="inv-quote-dd-empty">No accepted quotes found</div>
                  </div>
                </div>
                <!-- Selected quote card -->
                <div v-else class="inv-selected-quote-card">
                  <div class="inv-sq-card-left">
                    <span class="text-mono font-bold">{{ selectedQuote.quoteNumber }}</span>
                    <span class="font-medium">{{ selectedQuote.customerName }}</span>
                    <span class="text-muted" style="font-size:0.72rem">{{ selectedQuote.lineItems.filter(l => l.category === 'materials').length }} material items &middot; SAR {{ formatSAR(selectedQuote.total) }}</span>
                  </div>
                  <button class="btn btn-ghost btn-sm" @click="clearSelectedQuote"><X :size="14" /> Change</button>
                </div>
              </div>

              <!-- MANUAL REF (shown when source != quote) -->
              <div v-else class="form-group" style="flex:1">
                <label class="form-label">Reference # <span class="text-danger">*</span></label>
                <input v-model="holdMeta.sourceRef" type="text" class="form-input" :placeholder="holdMeta.source === 'project' ? 'e.g. PRJ-2026-010' : 'e.g. DEMO-002'" />
              </div>

              <div class="form-group" style="flex:1">
                <label class="form-label">Customer <span class="text-danger">*</span></label>
                <input v-model="holdMeta.customerName" type="text" class="form-input" placeholder="Customer name" :readonly="holdMeta.source === 'quote' && !!selectedQuote" />
              </div>
            </div>
            <div class="create-form-row">
              <div class="form-group" style="flex:2">
                <label class="form-label">Description / Project Name</label>
                <input v-model="holdMeta.sourceLabel" type="text" class="form-input" placeholder="e.g. Saudi Aramco CCTV Phase 3" />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label">Notes</label>
                <input v-model="holdMeta.notes" type="text" class="form-input" placeholder="Optional" />
              </div>
            </div>
          </div>

          <!-- Items -->
          <div class="inv-bulk-section">
            <div class="inv-bulk-items-header">
              <h4 class="inv-bulk-section-title"><Package :size="14" /> Items to Hold</h4>
              <div class="inv-bulk-items-actions">
                <div class="inv-bulk-search-wrap">
                  <div class="search-input" style="width:280px">
                    <Search :size="14" class="search-icon" />
                    <input v-model="holdItemSearch" type="text" class="form-input" style="font-size:0.78rem" placeholder="Search items to add..." @focus="showHoldItemDropdown = holdItemSearch.trim().length > 0" @blur="delayHideHoldDropdown" @input="showHoldItemDropdown = holdItemSearch.trim().length > 0" />
                  </div>
                  <div v-if="showHoldItemDropdown && holdSearchResults.length" class="inv-bulk-dropdown">
                    <button v-for="ws in holdSearchResults" :key="ws.id" class="inv-bulk-dd-item" @mousedown.prevent="addHoldItem(ws)">
                      <div class="inv-bulk-dd-left">
                        <span class="text-mono font-medium" style="font-size:0.72rem">{{ ws.productSku }}</span>
                        <span style="font-size:0.75rem">{{ ws.productName }}</span>
                      </div>
                      <div class="inv-bulk-dd-right">
                        <span :class="['badge', warehouseBadge[ws.warehouseLocation]]" style="font-size:0.62rem">{{ warehouseLabels[ws.warehouseLocation] }}</span>
                        <span class="font-medium" style="font-size:0.72rem">{{ ws.availableQty }} avail</span>
                      </div>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="holdItems.length" class="table-container table-container--embedded" style="max-height:340px">
              <table class="table">
                <thead>
                  <tr>
                    <th style="width:40px"><input type="checkbox" :checked="holdItems.every(i => i.selected)" @change="holdItems.forEach(i => i.selected = ($event.target as HTMLInputElement).checked)" /></th>
                    <th>SKU</th>
                    <th>Product</th>
                    <th>Warehouse</th>
                    <th class="text-center">Available</th>
                    <th class="text-center">Hold Qty</th>
                    <th>Status</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in holdItems" :key="idx" :class="[!item.selected ? 'inv-row--faded' : '', item.holdQty > item.availableQty && !item.poCreated ? 'inv-row--no-stock' : '']">
                    <td><input type="checkbox" v-model="item.selected" /></td>
                    <td class="text-mono font-medium" style="font-size:0.78rem">{{ item.productSku }}</td>
                    <td class="font-medium" style="font-size:0.78rem">{{ item.productName }}</td>
                    <td><span :class="['badge', warehouseBadge[item.warehouseLocation]]" style="font-size:0.65rem">{{ warehouseLabels[item.warehouseLocation] }}</span></td>
                    <td class="text-center">
                      <span :class="item.availableQty > 0 ? 'stock-green' : 'stock-red'" class="font-semibold">{{ item.availableQty }}</span>
                    </td>
                    <td class="text-center">
                      <input v-model.number="item.holdQty" type="number" min="0" class="form-input form-input--sm form-input--center" style="width:70px; margin:0 auto" :class="item.holdQty > item.availableQty && !item.poCreated ? 'inv-input-error' : ''" />
                    </td>
                    <td>
                      <template v-if="item.holdQty <= item.availableQty">
                        <span class="badge badge-success" style="font-size:0.62rem">In Stock</span>
                      </template>
                      <template v-else-if="item.poCreated">
                        <span class="badge badge-primary" style="font-size:0.62rem"><Truck :size="10" style="margin-right:3px" /> {{ item.poCreated }}</span>
                        <span class="text-muted" style="font-size:0.62rem; margin-left:4px">{{ getShortfall(item) }} ordered</span>
                      </template>
                      <template v-else>
                        <button class="btn btn-warning btn-sm inv-create-po-btn" @click="openQuickPO(item)">
                          <ShoppingBag :size="12" /> Order {{ getShortfall(item) }}
                        </button>
                      </template>
                    </td>
                    <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeHoldItem(idx)"><X :size="13" /></button></td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-bulk-empty">
              <Package :size="28" class="text-muted" />
              <p class="text-muted">Search and add items, or click "Load All Inventory" to see all available stock.</p>
            </div>
          </div>

          <div v-if="holdMissingItems.length > 0 && !holdMissingItems.every(i => !!i.poCreated)" class="inv-hold-po-banner">
            <div class="inv-hold-po-banner-left">
              <AlertTriangle :size="16" />
              <div>
                <strong>{{ holdMissingItems.filter(i => !i.poCreated).length }} item{{ holdMissingItems.filter(i => !i.poCreated).length !== 1 ? 's' : '' }} need more stock than available</strong>
                <span class="text-muted" style="font-size:0.75rem; margin-left:6px">Create a Purchase Order to cover the shortfall</span>
              </div>
            </div>
            <button class="btn btn-warning btn-sm" @click="openQuickPO()"><ShoppingBag :size="14" /> Create PO for Shortfall</button>
          </div>
          <div v-if="holdMissingItems.length > 0 && holdMissingItems.every(i => !!i.poCreated)" class="inv-hold-po-success">
            <Check :size="16" /> All shortfall items have POs created
          </div>
          <div v-if="holdHasErrors" class="inv-hold-warning">
            <AlertTriangle :size="14" /> One or more items exceed available stock
          </div>
        </div>
        <div class="modal-footer">
          <div class="inv-bulk-footer-info">
            <span v-if="holdSelectedItems.length > 0">
              <strong>{{ holdSelectedItems.length }}</strong> item{{ holdSelectedItems.length !== 1 ? 's' : '' }},
              <strong>{{ holdTotalUnits }}</strong> units total
            </span>
          </div>
          <button class="btn btn-secondary" @click="showCreateHoldModal = false">Cancel</button>
          <button class="btn btn-primary" :disabled="!holdMeta.sourceRef || !holdMeta.customerName || holdSelectedItems.length === 0 || holdHasErrors" @click="confirmBulkHold">
            <Lock :size="14" /> Reserve {{ holdTotalUnits }} Units
          </button>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- Quick PO from Hold Modal                               -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="showQuickPOModal" class="modal-backdrop" @click.self="showQuickPOModal = false">
      <div class="modal modal-lg">
        <div class="modal-header">
          <h2 class="modal-title"><ShoppingBag :size="18" /> Create Purchase Order</h2>
          <button class="modal-close" @click="showQuickPOModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body" style="max-height:70vh; overflow-y:auto">
          <div class="inv-quick-po-note">
            <AlertTriangle :size="14" />
            Creating a PO for <strong>{{ quickPOItems.length }} item{{ quickPOItems.length !== 1 ? 's' : '' }}</strong> to cover stock shortfall.
            <span v-if="holdMeta.sourceRef" class="text-muted"> Hold ref: {{ holdMeta.sourceRef }}</span>
          </div>

          <h4 class="inv-bulk-section-title" style="margin-top:var(--space-4)"><Building2 :size="14" /> Supplier</h4>
          <div class="create-form-row">
            <div class="form-group" style="flex:2">
              <label class="form-label">Supplier <span class="text-danger">*</span></label>
              <select :value="quickPOMeta.supplierId" class="form-select" @change="selectQuickPOSupplier($event)">
                <option value="" disabled>Select supplier...</option>
                <option v-for="s in mfrStore.suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </div>
            <div class="form-group" style="width:120px">
              <label class="form-label">Currency</label>
              <select v-model="quickPOMeta.currency" class="form-select">
                <option value="SAR">SAR</option><option value="USD">USD</option><option value="EUR">EUR</option><option value="GBP">GBP</option><option value="CNY">CNY</option><option value="AED">AED</option>
              </select>
            </div>
            <div class="form-group" style="width:160px">
              <label class="form-label">Expected Delivery</label>
              <input v-model="quickPOMeta.expectedDelivery" type="date" class="form-input" />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">Notes</label>
            <input v-model="quickPOMeta.notes" type="text" class="form-input" placeholder="e.g. Urgent - items needed for project hold" />
          </div>

          <h4 class="inv-bulk-section-title" style="margin-top:var(--space-5)"><Package :size="14" /> Items</h4>
          <div v-if="quickPOItems.length" class="table-container table-container--embedded" style="max-height:300px">
            <table class="table">
              <thead><tr><th>SKU</th><th>Product</th><th class="text-center">Order Qty</th><th class="text-right">Unit Cost</th><th class="text-right">Total</th><th class="text-center">Lead Time</th><th></th></tr></thead>
              <tbody>
                <tr v-for="(item, idx) in quickPOItems" :key="idx">
                  <td class="text-mono font-medium" style="font-size:0.78rem">{{ item.productSku }}</td>
                  <td class="font-medium" style="font-size:0.78rem">{{ item.productName }}</td>
                  <td class="text-center"><input v-model.number="item.orderQty" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:70px; margin:0 auto" /></td>
                  <td class="text-right"><input v-model.number="item.unitCost" type="number" step="0.01" min="0" class="form-input form-input--sm" style="width:100px; margin-left:auto" /></td>
                  <td class="text-right whitespace-nowrap font-medium">{{ formatSAR(item.orderQty * item.unitCost) }}</td>
                  <td class="text-center"><input v-model.number="item.leadTimeDays" type="number" min="1" class="form-input form-input--sm form-input--center" style="width:60px; margin:0 auto" /></td>
                  <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeQuickPOItem(idx)"><X :size="13" /></button></td>
                </tr>
              </tbody>
              <tfoot><tr><td colspan="4" class="text-right font-bold">Subtotal</td><td class="text-right font-bold">{{ formatSAR(quickPOSubtotal) }}</td><td colspan="2"></td></tr></tfoot>
            </table>
          </div>
        </div>
        <div class="modal-footer">
          <div class="inv-bulk-footer-info">
            <span v-if="quickPOItems.length"><strong>{{ quickPOItems.length }}</strong> item{{ quickPOItems.length !== 1 ? 's' : '' }}, total <strong>{{ formatSAR(quickPOSubtotal) }}</strong></span>
          </div>
          <button class="btn btn-secondary" @click="showQuickPOModal = false">Cancel</button>
          <button class="btn btn-primary" :disabled="!quickPOMeta.supplierName || quickPOItems.length === 0" @click="confirmQuickPO">
            <ShoppingBag :size="14" /> Create Draft PO
          </button>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════════════════ -->
    <!-- Bulk Movement Modal                                    -->
    <!-- ═══════════════════════════════════════════════════════ -->
    <div v-if="showBulkMoveModal" class="modal-backdrop" @click.self="showBulkMoveModal = false">
      <div class="modal inv-detail-modal">
        <div class="modal-header">
          <h2 class="modal-title"><ArrowRightLeft :size="18" /> Bulk Inventory Movement</h2>
          <button class="modal-close" @click="showBulkMoveModal = false"><X :size="20" /></button>
        </div>
        <div class="modal-body" style="max-height:75vh; overflow-y:auto">
          <!-- Movement info -->
          <div class="inv-bulk-section">
            <h4 class="inv-bulk-section-title"><ClipboardList :size="14" /> Movement Details</h4>
            <div class="create-form-row">
              <div class="form-group" style="width:180px">
                <label class="form-label">Type</label>
                <select v-model="moveMeta.movementType" class="form-select">
                  <option value="transfer">Transfer</option>
                  <option value="adjustment">Adjustment (reduce)</option>
                  <option value="write-off">Write-Off</option>
                </select>
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label">From Warehouse</label>
                <select v-model="moveMeta.fromWarehouse" class="form-select" @change="moveItems = []">
                  <option v-for="(label, key) in warehouseLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
              <div v-if="moveMeta.movementType === 'transfer'" class="form-group" style="flex:1">
                <label class="form-label">To Warehouse</label>
                <select v-model="moveMeta.toWarehouse" class="form-select">
                  <option v-for="(label, key) in warehouseLabels" :key="key" :value="key" :disabled="key === moveMeta.fromWarehouse">{{ label }}</option>
                </select>
              </div>
            </div>
            <div class="create-form-row">
              <div class="form-group" style="flex:2">
                <label class="form-label">Reason</label>
                <input v-model="moveMeta.reason" type="text" class="form-input" placeholder="e.g. Replenish branch stock, Damaged goods, Stock count correction..." />
              </div>
              <div class="form-group" style="flex:1">
                <label class="form-label">Notes</label>
                <input v-model="moveMeta.notes" type="text" class="form-input" placeholder="Optional" />
              </div>
            </div>
          </div>

          <!-- Items -->
          <div class="inv-bulk-section">
            <div class="inv-bulk-items-header">
              <h4 class="inv-bulk-section-title"><Package :size="14" /> Items</h4>
              <div class="inv-bulk-items-actions">
                <button class="btn btn-ghost btn-sm" @click="loadFromWarehouseItems">
                  <RefreshCw :size="13" /> Load {{ warehouseLabels[moveMeta.fromWarehouse] }} Items
                </button>
                <div class="inv-bulk-search-wrap">
                  <div class="search-input" style="width:280px">
                    <Search :size="14" class="search-icon" />
                    <input v-model="moveItemSearch" type="text" class="form-input" style="font-size:0.78rem" placeholder="Search items..." @focus="showMoveItemDropdown = moveItemSearch.trim().length > 0" @blur="delayHideMoveDropdown" @input="showMoveItemDropdown = moveItemSearch.trim().length > 0" />
                  </div>
                  <div v-if="showMoveItemDropdown && moveSearchResults.length" class="inv-bulk-dropdown">
                    <button v-for="ws in moveSearchResults" :key="ws.id" class="inv-bulk-dd-item" @mousedown.prevent="addMoveItem(ws)">
                      <div class="inv-bulk-dd-left">
                        <span class="text-mono font-medium" style="font-size:0.72rem">{{ ws.productSku }}</span>
                        <span style="font-size:0.75rem">{{ ws.productName }}</span>
                      </div>
                      <div class="inv-bulk-dd-right">
                        <span class="font-medium" style="font-size:0.72rem">{{ ws.availableQty }} avail</span>
                      </div>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="moveItems.length" class="table-container table-container--embedded" style="max-height:340px">
              <table class="table">
                <thead>
                  <tr>
                    <th style="width:40px"><input type="checkbox" :checked="moveItems.every(i => i.selected)" @change="moveItems.forEach(i => i.selected = ($event.target as HTMLInputElement).checked)" /></th>
                    <th>SKU</th>
                    <th>Product</th>
                    <th class="text-center">Available</th>
                    <th class="text-center">{{ moveMeta.movementType === 'transfer' ? 'Transfer Qty' : moveMeta.movementType === 'adjustment' ? 'Adjust Qty' : 'Write-Off Qty' }}</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(item, idx) in moveItems" :key="idx" :class="!item.selected ? 'inv-row--faded' : ''">
                    <td><input type="checkbox" v-model="item.selected" /></td>
                    <td class="text-mono font-medium" style="font-size:0.78rem">{{ item.productSku }}</td>
                    <td class="font-medium" style="font-size:0.78rem">{{ item.productName }}</td>
                    <td class="text-center font-semibold" :class="item.currentAvailable > 0 ? 'stock-green' : 'stock-red'">{{ item.currentAvailable }}</td>
                    <td class="text-center">
                      <input v-model.number="item.moveQty" type="number" min="0" :max="item.currentAvailable" class="form-input form-input--sm form-input--center" style="width:70px; margin:0 auto" :class="item.moveQty > item.currentAvailable ? 'inv-input-error' : ''" />
                    </td>
                    <td><button class="btn btn-ghost btn-icon btn-sm" @click="removeMoveItem(idx)"><X :size="13" /></button></td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="inv-bulk-empty">
              <ArrowRightLeft :size="28" class="text-muted" />
              <p class="text-muted">Search items or click "Load {{ warehouseLabels[moveMeta.fromWarehouse] }} Items" to populate.</p>
            </div>
          </div>

          <div v-if="moveHasErrors" class="inv-hold-warning">
            <AlertTriangle :size="14" /> One or more items exceed available stock at {{ warehouseLabels[moveMeta.fromWarehouse] }}
          </div>
        </div>
        <div class="modal-footer">
          <div class="inv-bulk-footer-info">
            <span v-if="moveSelectedItems.length > 0">
              <strong>{{ moveSelectedItems.length }}</strong> item{{ moveSelectedItems.length !== 1 ? 's' : '' }},
              <strong>{{ moveTotalUnits }}</strong> units total
            </span>
          </div>
          <button class="btn btn-secondary" @click="showBulkMoveModal = false">Cancel</button>
          <button class="btn btn-primary" :disabled="moveSelectedItems.length === 0 || moveHasErrors" @click="confirmBulkMove">
            <Check :size="14" /> {{ moveMeta.movementType === 'transfer' ? 'Transfer' : moveMeta.movementType === 'adjustment' ? 'Adjust' : 'Write Off' }} {{ moveTotalUnits }} Units
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.inventory-page { padding: var(--space-6); }

/* ── KPI Grid ─────────────────────────────────────────────── */
.kpi-row--6 {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.kpi-card {
  display: flex; align-items: center; gap: var(--space-4);
  padding: var(--space-5);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
}

.kpi-icon {
  width: 44px; height: 44px; border-radius: var(--radius-md);
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.kpi-icon--primary { background: var(--color-primary-light, #e0e7ff); color: var(--color-primary); }
.kpi-icon--info { background: #dbeafe; color: #2563eb; }
.kpi-icon--neutral { background: #f1f5f9; color: #64748b; }
.kpi-icon--success { background: #d1fae5; color: #059669; }
.kpi-icon--warning { background: #fef3c7; color: #d97706; }
.kpi-icon--danger { background: #fee2e2; color: #dc2626; }

.kpi-label {
  font-size: var(--text-xs); color: var(--color-neutral-500);
  font-weight: var(--font-medium); text-transform: uppercase; letter-spacing: 0.04em;
}
.kpi-value {
  font-size: var(--text-xl, 1.25rem); font-weight: var(--font-bold, 700);
  color: var(--color-neutral-900);
}

/* ── Toolbar ──────────────────────────────────────────────── */
.toolbar {
  display: flex; flex-direction: column; gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
}
.tab-row { display: flex; gap: var(--space-1); }
.tab-btn {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  background: var(--content-surface);
  font-size: var(--text-sm); font-weight: var(--font-medium);
  color: var(--color-neutral-600); cursor: pointer; transition: all 0.15s;
}
.tab-btn:hover { background: var(--color-neutral-50); }
.tab-btn--active { background: var(--color-primary); color: #fff; border-color: var(--color-primary); }
.toolbar-filters { display: flex; gap: var(--space-3); flex-wrap: wrap; }
.toolbar-search { flex: 1; min-width: 220px; }
.toolbar-select { width: 160px; }

/* ── Table ────────────────────────────────────────────────── */
.inv-row { cursor: pointer; transition: background 80ms; }
.inv-row:hover { background: var(--color-primary-50, #eef2ff); }
.inv-hold-badge {
  display: inline-flex; align-items: center; gap: 3px;
  padding: 2px 8px; border-radius: 10px;
  background: #fef3c7; color: #92400e;
  font-size: 0.72rem; font-weight: 700;
}

.stock-green { color: #059669; }
.stock-orange { color: #d97706; }
.stock-red { color: #dc2626; }

.stock-bar-wrapper { display: flex; align-items: center; gap: var(--space-2); }
.stock-bar {
  flex: 1; height: 8px; background: var(--color-neutral-100);
  border-radius: 4px; overflow: hidden; min-width: 60px;
}
.stock-bar-fill { height: 100%; border-radius: 4px; transition: width 0.3s ease; }
.stock-bar-fill.stock-green { background: #059669; }
.stock-bar-fill.stock-orange { background: #d97706; }
.stock-bar-fill.stock-red { background: #dc2626; }
.stock-bar-label { font-size: var(--text-xs); color: var(--color-neutral-500); min-width: 32px; text-align: right; }

/* ── Item Detail Modal ────────────────────────────────────── */
.inv-detail-modal {
  width: 100%;
  max-width: 1280px;
  max-height: calc(100vh - var(--space-8));
}

.inv-detail-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--color-neutral-200);
  gap: var(--space-4);
}
.inv-detail-header-left { display: flex; align-items: center; gap: var(--space-4); }
.inv-detail-icon {
  width: 48px; height: 48px; border-radius: var(--radius-lg);
  background: var(--color-primary-light, #e0e7ff);
  color: var(--color-primary);
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.inv-detail-title {
  font-size: 1.15rem; font-weight: 700; color: var(--color-neutral-900);
  margin: 0; line-height: 1.3;
}
.inv-detail-meta {
  display: flex; align-items: center; gap: var(--space-2);
  margin-top: 4px; font-size: 0.8rem; flex-wrap: wrap;
}
.inv-detail-sku {
  font-family: var(--font-mono, 'SF Mono', 'Fira Code', monospace);
  font-weight: 600; color: var(--color-primary);
  background: var(--color-primary-50, #eef2ff);
  padding: 1px 8px; border-radius: var(--radius-sm);
  font-size: 0.75rem;
}
.inv-detail-mfr { font-weight: 500; color: var(--color-neutral-600); }
.inv-detail-sep { color: var(--color-neutral-300); }

.inv-detail-header-right {
  display: flex; align-items: center; gap: var(--space-5);
}
.inv-hdr-num { text-align: center; }
.inv-hdr-num-val { display: block; font-size: 1.2rem; font-weight: 800; color: var(--color-neutral-800); line-height: 1.2; }
.inv-hdr-num-label { font-size: 0.65rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.05em; }

.inv-detail-tabs {
  display: flex; gap: 0; padding: 0 var(--space-6);
  border-bottom: 2px solid var(--color-neutral-200);
  overflow-x: auto; background: var(--color-neutral-50, #f8fafc);
}
.inv-dtab {
  display: flex; align-items: center; gap: 6px;
  padding: var(--space-3) var(--space-5);
  font-size: 0.82rem; font-weight: 500;
  color: var(--color-neutral-500);
  border-bottom: 3px solid transparent;
  margin-bottom: -2px;
  cursor: pointer; background: none; border-top: none; border-left: none; border-right: none;
  transition: all 120ms; white-space: nowrap;
}
.inv-dtab:hover { color: var(--color-neutral-700); background: var(--color-neutral-100, #f1f5f9); }
.inv-dtab--active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
  font-weight: 700;
  background: transparent;
}
.inv-dtab-count {
  font-size: 0.68rem; font-weight: 700;
  background: var(--color-neutral-200); color: var(--color-neutral-600);
  padding: 1px 7px; border-radius: 10px; min-width: 20px; text-align: center;
}
.inv-dtab--active .inv-dtab-count {
  background: var(--color-primary-100, #c7d2fe); color: var(--color-primary);
}

.inv-detail-body { padding: var(--space-6); max-height: 70vh; overflow-y: auto; }

/* ── Summary Cards ────────────────────────────────────────── */
.inv-summary-cards {
  display: grid; grid-template-columns: repeat(6, 1fr);
  gap: var(--space-4); margin-bottom: var(--space-6);
}
.inv-sum-card {
  padding: var(--space-5);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  text-align: center;
}
.inv-sum-label { font-size: 0.7rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 4px; }
.inv-sum-value { font-size: 1.25rem; font-weight: 800; color: var(--color-neutral-800); }

.inv-section-title {
  display: flex; align-items: center; gap: var(--space-2);
  font-size: 0.9rem; font-weight: 700; color: var(--color-neutral-700);
  margin: var(--space-6) 0 var(--space-4) 0;
}

/* ── Cost Grid ────────────────────────────────────────────── */
.inv-cost-grid {
  display: grid; grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}
.inv-cost-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: var(--space-4) var(--space-5);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
}
.inv-cost-label { font-size: 0.78rem; color: var(--color-neutral-500); }
.inv-cost-value { font-size: 0.88rem; font-weight: 700; color: var(--color-neutral-800); }

/* ── Price History Stats ──────────────────────────────────── */
.inv-ph-stats {
  display: grid; grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4); margin-bottom: var(--space-5);
}
.inv-ph-stat {
  padding: var(--space-5); text-align: center;
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
}
.inv-ph-stat-label { font-size: 0.7rem; color: var(--color-neutral-500); text-transform: uppercase; letter-spacing: 0.04em; }
.inv-ph-stat-value { font-size: 1.15rem; font-weight: 800; margin-top: 4px; }

.inv-ph-delta {
  display: inline-flex; align-items: center; gap: 2px;
  font-size: 0.75rem; font-weight: 600;
}

/* ── Supplier Reliability ─────────────────────────────────── */
.reliability-bar-wrap { display: flex; align-items: center; gap: var(--space-2); }
.reliability-bar {
  width: 60px; height: 6px; background: var(--color-neutral-100);
  border-radius: 3px; overflow: hidden;
}
.reliability-fill { height: 100%; border-radius: 3px; }
.reliability-fill.stock-green { background: #059669; }
.reliability-fill.stock-orange { background: #d97706; }
.reliability-fill.stock-red { background: #dc2626; }
.reliability-label { font-size: 0.72rem; color: var(--color-neutral-600); }

/* ── Tab count badge variant ──────────────────────────────── */
.inv-dtab-count--warn {
  background: #fef3c7; color: #b45309;
}
.inv-dtab--active .inv-dtab-count--warn {
  background: #fef3c7; color: #b45309;
}

/* ── Holds toolbar ────────────────────────────────────────── */
.inv-holds-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: var(--space-4); flex-wrap: wrap; gap: var(--space-3);
}
.inv-holds-summary {
  display: flex; align-items: center; gap: var(--space-2);
  font-size: 0.85rem; color: var(--color-neutral-600);
}
.inv-row--faded { opacity: 0.55; }

/* ── Release Hold Modal ───────────────────────────────────── */
.inv-release-info {
  display: flex; flex-direction: column; gap: var(--space-2);
  padding: var(--space-4); background: var(--color-neutral-50);
  border-radius: var(--radius-md); border: 1px solid var(--color-neutral-200);
}
.inv-release-row {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 0.82rem;
}
.inv-release-label { color: var(--color-neutral-500); font-weight: 500; }

/* ── Bulk Modals (Hold + Movement) ─────────────────────────── */
.create-form-row {
  display: flex; gap: var(--space-4); margin-bottom: var(--space-3);
}
.inv-bulk-section {
  margin-bottom: var(--space-5); padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-neutral-200);
}
.inv-bulk-section:last-of-type { border-bottom: none; }
.inv-bulk-section-title {
  display: flex; align-items: center; gap: var(--space-2);
  font-size: 0.85rem; font-weight: 700; color: var(--color-neutral-700);
  margin-bottom: var(--space-3);
}
.inv-bulk-items-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: var(--space-3); flex-wrap: wrap; gap: var(--space-2);
}
.inv-bulk-items-actions { display: flex; align-items: center; gap: var(--space-2); }
.inv-bulk-search-wrap { position: relative; }
.inv-bulk-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 30;
  background: var(--content-surface); border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-md); max-height: 260px; overflow-y: auto;
  box-shadow: var(--shadow-lg);
}
.inv-bulk-dd-item {
  display: flex; align-items: center; justify-content: space-between;
  width: 100%; padding: var(--space-2) var(--space-3);
  background: none; border: none; cursor: pointer;
  border-bottom: 1px solid var(--color-neutral-100);
  text-align: left; transition: background 80ms;
}
.inv-bulk-dd-item:hover { background: var(--color-primary-50, #eef2ff); }
.inv-bulk-dd-item:last-child { border-bottom: none; }
.inv-bulk-dd-left { display: flex; flex-direction: column; gap: 1px; }
.inv-bulk-dd-right { display: flex; align-items: center; gap: var(--space-2); }
.inv-bulk-empty {
  display: flex; flex-direction: column; align-items: center;
  gap: var(--space-2); padding: var(--space-7) 0;
}
.inv-bulk-empty p { font-size: 0.78rem; }
.inv-bulk-footer-info {
  flex: 1; font-size: 0.82rem; color: var(--color-neutral-600);
}
.inv-input-error {
  border-color: #dc2626 !important; background: #fef2f2 !important;
}
.form-input--sm { padding: 4px 8px; font-size: 0.78rem; }
.form-input--center { text-align: center; }
.inv-hold-warning {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4); margin-bottom: var(--space-3);
  background: #fef3c7; border: 1px solid #fde68a;
  border-radius: var(--radius-md);
  font-size: 0.78rem; color: #92400e;
}
.page-header-actions { display: flex; gap: var(--space-2); }

/* ── Quote search in hold modal ───────────────────────────── */
.inv-quote-search-wrap { position: relative; }
.inv-quote-dropdown {
  position: absolute; top: 100%; left: 0; right: 0; z-index: 30;
  background: var(--content-surface); border: 1px solid var(--color-neutral-300);
  border-radius: var(--radius-md); max-height: 300px; overflow-y: auto;
  box-shadow: var(--shadow-lg);
}
.inv-quote-dd-item {
  display: flex; flex-direction: column; gap: 3px;
  width: 100%; padding: var(--space-3) var(--space-4);
  background: none; border: none; cursor: pointer;
  border-bottom: 1px solid var(--color-neutral-100);
  text-align: left; transition: background 80ms;
}
.inv-quote-dd-item:hover { background: var(--color-primary-50, #eef2ff); }
.inv-quote-dd-item:last-child { border-bottom: none; }
.inv-quote-dd-top {
  display: flex; align-items: center; gap: var(--space-2);
}
.inv-quote-dd-bottom {
  display: flex; align-items: center; gap: var(--space-2);
  font-size: 0.75rem;
}
.inv-quote-dd-empty {
  padding: var(--space-4); text-align: center;
  font-size: 0.78rem; color: var(--color-neutral-500);
}
.inv-selected-quote-card {
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  background: var(--color-success-50, #f0fdf4);
  border: 1px solid var(--color-success-200, #bbf7d0);
  border-radius: var(--radius-md);
}
.inv-sq-card-left {
  display: flex; flex-direction: column; gap: 2px;
  font-size: 0.82rem;
}

/* ── Empty Tab ────────────────────────────────────────────── */
.inv-empty-tab {
  display: flex; flex-direction: column; align-items: center;
  gap: var(--space-3); padding: var(--space-10) 0;
}
.inv-empty-tab p { font-size: 0.88rem; }

/* ── Embedded table container ─────────────────────────────── */
.table-container--embedded {
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
  overflow: auto; max-height: 420px;
}
.table-container--embedded .table { margin-bottom: 0; }
.table-container--embedded .table th {
  position: sticky; top: 0; z-index: 2;
  background: var(--color-neutral-50, #f8fafc);
}

/* ── Quick PO from Hold ────────────────────────────────────── */
.inv-row--no-stock { background: var(--color-warning-50, #fffbeb); }
[data-theme="dark"] .inv-row--no-stock { background: rgba(234, 179, 8, 0.08); }

.inv-create-po-btn { font-size: 0.68rem !important; padding: 2px 8px !important; gap: 4px; white-space: nowrap; }

.inv-hold-po-banner {
  display: flex; align-items: center; justify-content: space-between; gap: var(--space-3);
  padding: var(--space-3) var(--space-4); margin-top: var(--space-3);
  background: var(--color-warning-50, #fffbeb); border: 1px solid var(--color-warning-200, #fde68a);
  border-radius: var(--radius-lg); color: var(--color-warning-700, #a16207);
}
[data-theme="dark"] .inv-hold-po-banner { background: rgba(234, 179, 8, 0.1); border-color: rgba(234, 179, 8, 0.3); color: #fbbf24; }
.inv-hold-po-banner-left { display: flex; align-items: center; gap: var(--space-3); font-size: 0.82rem; }
.inv-hold-po-banner-left strong { font-weight: 700; }

.inv-hold-po-success {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4); margin-top: var(--space-3);
  background: var(--color-success-50, #f0fdf4); border: 1px solid var(--color-success-200, #bbf7d0);
  border-radius: var(--radius-lg); color: var(--color-success-700, #15803d);
  font-size: 0.82rem; font-weight: 600;
}
[data-theme="dark"] .inv-hold-po-success { background: rgba(34, 197, 94, 0.1); border-color: rgba(34, 197, 94, 0.3); color: #4ade80; }

.inv-quick-po-note {
  display: flex; align-items: center; gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: var(--color-warning-50, #fffbeb); border: 1px solid var(--color-warning-200, #fde68a);
  border-radius: var(--radius-lg); color: var(--color-warning-700, #a16207);
  font-size: 0.82rem;
}
[data-theme="dark"] .inv-quick-po-note { background: rgba(234, 179, 8, 0.1); border-color: rgba(234, 179, 8, 0.3); color: #fbbf24; }

/* ── Responsive ───────────────────────────────────────────── */
@media (max-width: 1200px) {
  .kpi-row--6 { grid-template-columns: repeat(3, 1fr); }
  .inv-summary-cards { grid-template-columns: repeat(3, 1fr); }
  .inv-cost-grid { grid-template-columns: repeat(2, 1fr); }
  .inv-detail-header { flex-direction: column; align-items: flex-start; }
  .inv-detail-header-right { width: 100%; justify-content: flex-start; }
}
@media (max-width: 768px) {
  .kpi-row--6 { grid-template-columns: repeat(2, 1fr); }
  .toolbar { flex-direction: column; }
  .tab-row { flex-wrap: wrap; }
  .toolbar-filters { flex-direction: column; }
  .toolbar-select { width: 100%; }
  .inv-summary-cards { grid-template-columns: repeat(2, 1fr); }
  .inv-cost-grid { grid-template-columns: 1fr; }
  .inv-ph-stats { grid-template-columns: repeat(2, 1fr); }
  .inv-detail-modal { max-width: calc(100vw - var(--space-4)); }
  .inv-detail-tabs { padding: 0 var(--space-3); }
  .inv-dtab { padding: var(--space-2) var(--space-3); font-size: 0.75rem; }
}
</style>
