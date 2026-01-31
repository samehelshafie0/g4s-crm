import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export interface GoodsReceiptItem {
  id: string
  productId: string
  productSku: string
  productName: string
  poLineItemId: string
  orderedQuantity: number
  receivedQuantity: number
  acceptedQuantity: number
  rejectedQuantity: number
  rejectionReason?: string
  unitPrice: number
  totalValue: number
  batchNumber?: string
  expiryDate?: string
  notes?: string
}

export interface GoodsReceipt {
  id: string
  grNumber: string

  // PO Reference
  purchaseOrderId: string
  purchaseOrderNumber: string
  vendorId: string
  vendorName: string

  // Receipt Details
  receiptDate: string
  receiptType: 'full' | 'partial' | 'final'

  // Delivery Information
  warehouseLocation: string
  deliveryNoteNumber?: string
  carrierName?: string
  trackingNumber?: string
  vehicleNumber?: string
  driverName?: string

  // Items
  items: GoodsReceiptItem[]

  // Quality Control
  qualityCheckStatus: 'pending' | 'passed' | 'failed' | 'partial'
  qualityCheckedBy?: string
  qualityCheckedAt?: string
  qualityNotes?: string

  // Status
  status: 'draft' | 'in_progress' | 'completed' | 'rejected'

  // Financial
  totalValue: number
  currency: 'SAR' | 'USD' | 'EUR' | 'CNY'

  // Discrepancies
  hasDiscrepancy: boolean
  discrepancyNotes?: string

  // Documents
  attachments?: string[]

  // Notes
  notes?: string

  // Audit
  receivedBy: string
  createdAt: string
  completedAt?: string
  updatedAt?: string
}

export const useGoodsReceiptsStore = defineStore('goodsReceipts', () => {
  const receipts = ref<GoodsReceipt[]>([
    {
      id: 'GR-001',
      grNumber: 'GRN-2024-001',
      purchaseOrderId: 'PO-001',
      purchaseOrderNumber: 'PO-2024-001',
      vendorId: 'VEN-001',
      vendorName: 'Hikvision MENA',
      receiptDate: '2024-11-28T10:30:00Z',
      receiptType: 'full',
      warehouseLocation: 'Riyadh Main',
      deliveryNoteNumber: 'DN-HIK-45678',
      carrierName: 'DHL Express',
      trackingNumber: 'DHL-1234567890',
      items: [
        {
          id: 'GRI-001',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          poLineItemId: 'POL-001',
          orderedQuantity: 50,
          receivedQuantity: 50,
          acceptedQuantity: 50,
          rejectedQuantity: 0,
          unitPrice: 675,
          totalValue: 33750,
          batchNumber: 'HK-2024-B1234'
        },
        {
          id: 'GRI-002',
          productId: 'PRD-003',
          productSku: 'NVR-HK-7732',
          productName: 'Hikvision 32CH NVR',
          poLineItemId: 'POL-002',
          orderedQuantity: 10,
          receivedQuantity: 10,
          acceptedQuantity: 10,
          rejectedQuantity: 0,
          unitPrice: 2800,
          totalValue: 28000,
          batchNumber: 'HK-2024-B1235'
        }
      ],
      qualityCheckStatus: 'passed',
      qualityCheckedBy: 'USR-003',
      qualityCheckedAt: '2024-11-28T12:00:00Z',
      qualityNotes: 'All items inspected and passed QC',
      status: 'completed',
      totalValue: 61750,
      currency: 'SAR',
      hasDiscrepancy: false,
      receivedBy: 'USR-001',
      createdAt: '2024-11-28T10:30:00Z',
      completedAt: '2024-11-28T14:00:00Z'
    },
    {
      id: 'GR-002',
      grNumber: 'GRN-2024-002',
      purchaseOrderId: 'PO-002',
      purchaseOrderNumber: 'PO-2024-002',
      vendorId: 'VEN-002',
      vendorName: 'Dahua Technology Middle East',
      receiptDate: '2024-12-01T09:15:00Z',
      receiptType: 'partial',
      warehouseLocation: 'Jeddah Branch',
      deliveryNoteNumber: 'DN-DAH-98765',
      carrierName: 'Aramex',
      trackingNumber: 'ARX-9876543210',
      items: [
        {
          id: 'GRI-003',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          poLineItemId: 'POL-003',
          orderedQuantity: 100,
          receivedQuantity: 60,
          acceptedQuantity: 58,
          rejectedQuantity: 2,
          rejectionReason: 'Physical damage to housing',
          unitPrice: 825,
          totalValue: 49500,
          batchNumber: 'DH-2024-C5678',
          notes: '2 units rejected due to shipping damage'
        }
      ],
      qualityCheckStatus: 'partial',
      qualityCheckedBy: 'USR-003',
      qualityCheckedAt: '2024-12-01T11:00:00Z',
      qualityNotes: '2 units damaged, vendor notified for replacement',
      status: 'completed',
      totalValue: 49500,
      currency: 'SAR',
      hasDiscrepancy: true,
      discrepancyNotes: 'Received 60 out of 100 ordered. Remaining 40 expected in 7 days. 2 units rejected.',
      receivedBy: 'USR-004',
      createdAt: '2024-12-01T09:15:00Z',
      completedAt: '2024-12-01T13:30:00Z'
    },
    {
      id: 'GR-003',
      grNumber: 'GRN-2024-003',
      purchaseOrderId: 'PO-002',
      purchaseOrderNumber: 'PO-2024-002',
      vendorId: 'VEN-002',
      vendorName: 'Dahua Technology Middle East',
      receiptDate: '2024-12-08T14:20:00Z',
      receiptType: 'final',
      warehouseLocation: 'Jeddah Branch',
      deliveryNoteNumber: 'DN-DAH-98766',
      carrierName: 'Aramex',
      trackingNumber: 'ARX-1122334455',
      items: [
        {
          id: 'GRI-004',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          poLineItemId: 'POL-003',
          orderedQuantity: 42,
          receivedQuantity: 42,
          acceptedQuantity: 42,
          rejectedQuantity: 0,
          unitPrice: 825,
          totalValue: 34650,
          batchNumber: 'DH-2024-C5679',
          notes: 'Second batch + 2 replacement units for damaged items'
        }
      ],
      qualityCheckStatus: 'passed',
      qualityCheckedBy: 'USR-003',
      qualityCheckedAt: '2024-12-08T15:30:00Z',
      qualityNotes: 'All units passed inspection. PO now complete.',
      status: 'completed',
      totalValue: 34650,
      currency: 'SAR',
      hasDiscrepancy: false,
      receivedBy: 'USR-004',
      createdAt: '2024-12-08T14:20:00Z',
      completedAt: '2024-12-08T16:00:00Z'
    }
  ])

  // Computed
  const pendingReceipts = computed(() =>
    receipts.value.filter(r => r.status === 'draft' || r.status === 'in_progress')
  )

  const completedReceipts = computed(() =>
    receipts.value.filter(r => r.status === 'completed')
  )

  const receiptsWithDiscrepancies = computed(() =>
    receipts.value.filter(r => r.hasDiscrepancy)
  )

  const receiptsByPO = computed(() => {
    const grouped: Record<string, GoodsReceipt[]> = {}
    receipts.value.forEach(receipt => {
      if (!grouped[receipt.purchaseOrderId]) {
        grouped[receipt.purchaseOrderId] = []
      }
      grouped[receipt.purchaseOrderId].push(receipt)
    })
    return grouped
  })

  const receiptStats = computed(() => {
    const total = receipts.value.length
    const pending = receipts.value.filter(r => r.status === 'draft' || r.status === 'in_progress').length
    const completed = receipts.value.filter(r => r.status === 'completed').length
    const withDiscrepancies = receipts.value.filter(r => r.hasDiscrepancy).length
    const totalValue = receipts.value.reduce((sum, r) => sum + r.totalValue, 0)

    const qualityPassed = receipts.value.filter(r => r.qualityCheckStatus === 'passed').length
    const qualityFailed = receipts.value.filter(r => r.qualityCheckStatus === 'failed').length
    const qualityPartial = receipts.value.filter(r => r.qualityCheckStatus === 'partial').length

    return {
      total,
      pending,
      completed,
      withDiscrepancies,
      totalValue,
      qualityPassed,
      qualityFailed,
      qualityPartial
    }
  })

  const recentReceipts = computed(() => {
    return [...receipts.value]
      .sort((a, b) => new Date(b.receiptDate).getTime() - new Date(a.receiptDate).getTime())
      .slice(0, 10)
  })

  // Actions
  function getReceiptById(id: string): GoodsReceipt | undefined {
    return receipts.value.find(r => r.id === id)
  }

  function getReceiptByNumber(grNumber: string): GoodsReceipt | undefined {
    return receipts.value.find(r => r.grNumber === grNumber)
  }

  function getReceiptsByPO(poId: string): GoodsReceipt[] {
    return receipts.value.filter(r => r.purchaseOrderId === poId)
  }

  function getReceiptsByVendor(vendorId: string): GoodsReceipt[] {
    return receipts.value.filter(r => r.vendorId === vendorId)
  }

  function getReceiptsByWarehouse(warehouse: string): GoodsReceipt[] {
    return receipts.value.filter(r => r.warehouseLocation === warehouse)
  }

  function addReceipt(receipt: Omit<GoodsReceipt, 'id' | 'grNumber' | 'createdAt'>) {
    const receiptCount = receipts.value.length + 1
    const newId = `GR-${String(receiptCount).padStart(3, '0')}`
    const year = new Date().getFullYear()
    const grNumber = `GRN-${year}-${String(receiptCount).padStart(3, '0')}`

    const newReceipt: GoodsReceipt = {
      ...receipt,
      id: newId,
      grNumber,
      createdAt: new Date().toISOString()
    }

    receipts.value.push(newReceipt)
    return newReceipt
  }

  function updateReceipt(id: string, updates: Partial<GoodsReceipt>) {
    const index = receipts.value.findIndex(r => r.id === id)
    if (index !== -1) {
      receipts.value[index] = {
        ...receipts.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return receipts.value[index]
    }
    return null
  }

  function completeReceipt(id: string, qualityCheckedBy: string, qualityNotes?: string) {
    const receipt = receipts.value.find(r => r.id === id)
    if (receipt) {
      receipt.status = 'completed'
      receipt.completedAt = new Date().toISOString()
      receipt.qualityCheckedBy = qualityCheckedBy
      receipt.qualityCheckedAt = new Date().toISOString()
      if (qualityNotes) receipt.qualityNotes = qualityNotes

      // Determine quality check status
      const totalRejected = receipt.items.reduce((sum, item) => sum + item.rejectedQuantity, 0)
      const totalAccepted = receipt.items.reduce((sum, item) => sum + item.acceptedQuantity, 0)

      if (totalRejected === 0) {
        receipt.qualityCheckStatus = 'passed'
      } else if (totalAccepted === 0) {
        receipt.qualityCheckStatus = 'failed'
      } else {
        receipt.qualityCheckStatus = 'partial'
      }

      receipt.updatedAt = new Date().toISOString()
    }
  }

  function updateReceiptItem(receiptId: string, itemId: string, updates: Partial<GoodsReceiptItem>) {
    const receipt = receipts.value.find(r => r.id === receiptId)
    if (!receipt) return null

    const itemIndex = receipt.items.findIndex(i => i.id === itemId)
    if (itemIndex === -1) return null

    receipt.items[itemIndex] = {
      ...receipt.items[itemIndex],
      ...updates
    }

    // Recalculate total value
    receipt.totalValue = receipt.items.reduce((sum, item) => sum + item.totalValue, 0)
    receipt.updatedAt = new Date().toISOString()

    return receipt
  }

  function flagDiscrepancy(id: string, discrepancyNotes: string) {
    const receipt = receipts.value.find(r => r.id === id)
    if (receipt) {
      receipt.hasDiscrepancy = true
      receipt.discrepancyNotes = discrepancyNotes
      receipt.updatedAt = new Date().toISOString()
    }
  }

  function deleteReceipt(id: string) {
    const index = receipts.value.findIndex(r => r.id === id)
    if (index !== -1) {
      receipts.value.splice(index, 1)
      return true
    }
    return false
  }

  return {
    receipts,
    pendingReceipts,
    completedReceipts,
    receiptsWithDiscrepancies,
    receiptsByPO,
    receiptStats,
    recentReceipts,
    getReceiptById,
    getReceiptByNumber,
    getReceiptsByPO,
    getReceiptsByVendor,
    getReceiptsByWarehouse,
    addReceipt,
    updateReceipt,
    completeReceipt,
    updateReceiptItem,
    flagDiscrepancy,
    deleteReceipt
  }
})
