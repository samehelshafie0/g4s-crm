import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export interface POLineItem {
  id: string
  productId: string
  productSku: string
  productName: string
  description: string
  quantity: number
  unitPrice: number
  discount: number
  subtotal: number
  taxPercent: number
  taxAmount: number
  total: number
  expectedDeliveryDate: string
  receivedQuantity: number
  remainingQuantity: number
  status: 'pending' | 'partial' | 'received' | 'cancelled'
}

export interface PurchaseOrder {
  id: string
  poNumber: string

  // Vendor Information
  vendorId: string
  vendorName: string
  vendorCode: string

  // Dates
  orderDate: string
  expectedDeliveryDate: string
  actualDeliveryDate?: string

  // Delivery Location
  deliveryWarehouse: string
  deliveryAddress?: string

  // Status
  status: 'draft' | 'pending_approval' | 'approved' | 'sent' | 'partial' | 'received' | 'cancelled' | 'closed'
  approvalStatus: 'pending' | 'approved' | 'rejected'
  approvedBy?: string
  approvedAt?: string

  // Line Items
  items: POLineItem[]

  // Financial Details
  subtotal: number
  discountAmount: number
  taxAmount: number
  freightCost: number
  otherCharges: number
  totalAmount: number
  currency: 'SAR' | 'USD' | 'EUR' | 'CNY'

  // Payment Terms
  paymentTerms: string
  paymentStatus: 'unpaid' | 'partial' | 'paid'

  // Reference
  referenceNumber?: string
  quoteReference?: string
  projectReference?: string

  // Notes
  notes?: string
  internalNotes?: string
  termsAndConditions?: string

  // Tracking
  receivedValue: number
  receivedPercentage: number

  // Audit
  createdAt: string
  createdBy: string
  updatedAt?: string
  updatedBy?: string
}

export const usePurchaseOrdersStore = defineStore('purchaseOrders', () => {
  const orders = ref<PurchaseOrder[]>([
    {
      id: 'PO-001',
      poNumber: 'PO-2024-001',
      vendorId: 'VEN-001',
      vendorName: 'Hikvision MENA',
      vendorCode: 'HIK-MENA',
      orderDate: '2024-11-01T00:00:00Z',
      expectedDeliveryDate: '2024-12-01T00:00:00Z',
      actualDeliveryDate: '2024-11-28T00:00:00Z',
      deliveryWarehouse: 'Riyadh Main',
      status: 'received',
      approvalStatus: 'approved',
      approvedBy: 'USR-002',
      approvedAt: '2024-11-02T00:00:00Z',
      items: [
        {
          id: 'POL-001',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          description: '8MP outdoor IP dome camera with IR',
          quantity: 50,
          unitPrice: 675,
          discount: 0,
          subtotal: 33750,
          taxPercent: 15,
          taxAmount: 5062.5,
          total: 38812.5,
          expectedDeliveryDate: '2024-12-01T00:00:00Z',
          receivedQuantity: 50,
          remainingQuantity: 0,
          status: 'received'
        },
        {
          id: 'POL-002',
          productId: 'PRD-003',
          productSku: 'NVR-HK-7732',
          productName: 'Hikvision 32CH NVR',
          description: '32 channel network video recorder',
          quantity: 10,
          unitPrice: 2800,
          discount: 0,
          subtotal: 28000,
          taxPercent: 15,
          taxAmount: 4200,
          total: 32200,
          expectedDeliveryDate: '2024-12-01T00:00:00Z',
          receivedQuantity: 10,
          remainingQuantity: 0,
          status: 'received'
        }
      ],
      subtotal: 61750,
      discountAmount: 0,
      taxAmount: 9262.5,
      freightCost: 2500,
      otherCharges: 500,
      totalAmount: 74012.5,
      currency: 'SAR',
      paymentTerms: 'Net 45',
      paymentStatus: 'paid',
      notes: 'Urgent order for government project',
      receivedValue: 74012.5,
      receivedPercentage: 100,
      createdAt: '2024-11-01T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'PO-002',
      poNumber: 'PO-2024-002',
      vendorId: 'VEN-002',
      vendorName: 'Dahua Technology Middle East',
      vendorCode: 'DAH-TECH',
      orderDate: '2024-11-15T00:00:00Z',
      expectedDeliveryDate: '2024-12-10T00:00:00Z',
      deliveryWarehouse: 'Jeddah Branch',
      status: 'partial',
      approvalStatus: 'approved',
      approvedBy: 'USR-002',
      approvedAt: '2024-11-16T00:00:00Z',
      items: [
        {
          id: 'POL-003',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          description: '5MP bullet camera with motorized lens',
          quantity: 100,
          unitPrice: 825,
          discount: 5,
          subtotal: 78375,
          taxPercent: 15,
          taxAmount: 11756.25,
          total: 90131.25,
          expectedDeliveryDate: '2024-12-10T00:00:00Z',
          receivedQuantity: 60,
          remainingQuantity: 40,
          status: 'partial'
        }
      ],
      subtotal: 78375,
      discountAmount: 4125,
      taxAmount: 11756.25,
      freightCost: 3000,
      otherCharges: 800,
      totalAmount: 90131.25,
      currency: 'SAR',
      paymentTerms: 'Net 30',
      paymentStatus: 'partial',
      notes: 'Shipment coming in two batches',
      receivedValue: 54078.75,
      receivedPercentage: 60,
      createdAt: '2024-11-15T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'PO-003',
      poNumber: 'PO-2024-003',
      vendorId: 'VEN-003',
      vendorName: 'Axis Communications Saudi Arabia',
      vendorCode: 'AXIS-KSA',
      orderDate: '2024-12-05T00:00:00Z',
      expectedDeliveryDate: '2025-01-10T00:00:00Z',
      deliveryWarehouse: 'Riyadh Main',
      status: 'approved',
      approvalStatus: 'approved',
      approvedBy: 'USR-002',
      approvedAt: '2024-12-06T00:00:00Z',
      items: [
        {
          id: 'POL-004',
          productId: 'PRD-005',
          productSku: 'CAM-AXIS-P3375',
          productName: 'Axis P3375-V Dome Camera',
          description: 'High-end dome camera for enterprise',
          quantity: 30,
          unitPrice: 3500,
          discount: 0,
          subtotal: 105000,
          taxPercent: 15,
          taxAmount: 15750,
          total: 120750,
          expectedDeliveryDate: '2025-01-10T00:00:00Z',
          receivedQuantity: 0,
          remainingQuantity: 30,
          status: 'pending'
        }
      ],
      subtotal: 105000,
      discountAmount: 0,
      taxAmount: 15750,
      freightCost: 4500,
      otherCharges: 1200,
      totalAmount: 126450,
      currency: 'SAR',
      paymentTerms: 'Net 60',
      paymentStatus: 'unpaid',
      projectReference: 'PROJ-2024-015',
      notes: 'Premium cameras for bank headquarters project',
      receivedValue: 0,
      receivedPercentage: 0,
      createdAt: '2024-12-05T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'PO-004',
      poNumber: 'PO-2024-004',
      vendorId: 'VEN-004',
      vendorName: 'Eltec Electronics - Riyadh',
      vendorCode: 'ELTEC-RY',
      orderDate: '2024-12-08T00:00:00Z',
      expectedDeliveryDate: '2024-12-15T00:00:00Z',
      deliveryWarehouse: 'Riyadh Main',
      status: 'sent',
      approvalStatus: 'approved',
      approvedBy: 'USR-001',
      approvedAt: '2024-12-08T00:00:00Z',
      items: [
        {
          id: 'POL-005',
          productId: 'PRD-010',
          productSku: 'CBL-CAT6-305M',
          productName: 'Cat6 UTP Cable - 305m Roll',
          description: 'Category 6 networking cable',
          quantity: 20,
          unitPrice: 450,
          discount: 0,
          subtotal: 9000,
          taxPercent: 15,
          taxAmount: 1350,
          total: 10350,
          expectedDeliveryDate: '2024-12-15T00:00:00Z',
          receivedQuantity: 0,
          remainingQuantity: 20,
          status: 'pending'
        },
        {
          id: 'POL-006',
          productId: 'PRD-011',
          productSku: 'PWR-12V5A',
          productName: '12V 5A Power Supply',
          description: 'Universal power adapter for cameras',
          quantity: 50,
          unitPrice: 45,
          discount: 10,
          subtotal: 2025,
          taxPercent: 15,
          taxAmount: 303.75,
          total: 2328.75,
          expectedDeliveryDate: '2024-12-15T00:00:00Z',
          receivedQuantity: 0,
          remainingQuantity: 50,
          status: 'pending'
        }
      ],
      subtotal: 11025,
      discountAmount: 225,
      taxAmount: 1653.75,
      freightCost: 0,
      otherCharges: 0,
      totalAmount: 12678.75,
      currency: 'SAR',
      paymentTerms: 'Net 15',
      paymentStatus: 'unpaid',
      notes: 'Local supplier - fast delivery expected',
      receivedValue: 0,
      receivedPercentage: 0,
      createdAt: '2024-12-08T00:00:00Z',
      createdBy: 'USR-001'
    }
  ])

  // Computed
  const pendingOrders = computed(() =>
    orders.value.filter(o => o.status === 'approved' || o.status === 'sent')
  )

  const partialOrders = computed(() =>
    orders.value.filter(o => o.status === 'partial')
  )

  const receivedOrders = computed(() =>
    orders.value.filter(o => o.status === 'received')
  )

  const ordersByVendor = computed(() => {
    const grouped: Record<string, PurchaseOrder[]> = {}
    orders.value.forEach(order => {
      if (!grouped[order.vendorId]) {
        grouped[order.vendorId] = []
      }
      grouped[order.vendorId].push(order)
    })
    return grouped
  })

  const orderStats = computed(() => {
    const total = orders.value.length
    const draft = orders.value.filter(o => o.status === 'draft').length
    const pending = orders.value.filter(o => o.status === 'approved' || o.status === 'sent').length
    const partial = orders.value.filter(o => o.status === 'partial').length
    const received = orders.value.filter(o => o.status === 'received').length
    const cancelled = orders.value.filter(o => o.status === 'cancelled').length

    const totalValue = orders.value.reduce((sum, o) => sum + o.totalAmount, 0)
    const receivedValue = orders.value.reduce((sum, o) => sum + o.receivedValue, 0)
    const pendingValue = totalValue - receivedValue

    return {
      total,
      draft,
      pending,
      partial,
      received,
      cancelled,
      totalValue,
      receivedValue,
      pendingValue
    }
  })

  const upcomingDeliveries = computed(() => {
    const today = new Date()
    const next7Days = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000)

    return orders.value
      .filter(o => {
        if (o.status === 'received' || o.status === 'cancelled') return false
        const deliveryDate = new Date(o.expectedDeliveryDate)
        return deliveryDate <= next7Days && deliveryDate >= today
      })
      .sort((a, b) => new Date(a.expectedDeliveryDate).getTime() - new Date(b.expectedDeliveryDate).getTime())
  })

  const overdueOrders = computed(() => {
    const today = new Date()
    return orders.value
      .filter(o => {
        if (o.status === 'received' || o.status === 'cancelled') return false
        const deliveryDate = new Date(o.expectedDeliveryDate)
        return deliveryDate < today
      })
  })

  // Actions
  function getOrderById(id: string): PurchaseOrder | undefined {
    return orders.value.find(o => o.id === id)
  }

  function getOrderByNumber(poNumber: string): PurchaseOrder | undefined {
    return orders.value.find(o => o.poNumber === poNumber)
  }

  function getOrdersByVendor(vendorId: string): PurchaseOrder[] {
    return orders.value.filter(o => o.vendorId === vendorId)
  }

  function getOrdersByStatus(status: PurchaseOrder['status']): PurchaseOrder[] {
    return orders.value.filter(o => o.status === status)
  }

  function addOrder(order: Omit<PurchaseOrder, 'id' | 'poNumber' | 'createdAt' | 'receivedValue' | 'receivedPercentage'>) {
    const orderCount = orders.value.length + 1
    const newId = `PO-${String(orderCount).padStart(3, '0')}`
    const year = new Date().getFullYear()
    const poNumber = `PO-${year}-${String(orderCount).padStart(3, '0')}`

    const newOrder: PurchaseOrder = {
      ...order,
      id: newId,
      poNumber,
      receivedValue: 0,
      receivedPercentage: 0,
      createdAt: new Date().toISOString()
    }

    orders.value.push(newOrder)
    return newOrder
  }

  function updateOrder(id: string, updates: Partial<PurchaseOrder>) {
    const index = orders.value.findIndex(o => o.id === id)
    if (index !== -1) {
      orders.value[index] = {
        ...orders.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return orders.value[index]
    }
    return null
  }

  function updateOrderStatus(id: string, status: PurchaseOrder['status']) {
    const order = orders.value.find(o => o.id === id)
    if (order) {
      order.status = status
      order.updatedAt = new Date().toISOString()

      if (status === 'received') {
        order.actualDeliveryDate = new Date().toISOString()
      }
    }
  }

  function approveOrder(id: string, approvedBy: string) {
    const order = orders.value.find(o => o.id === id)
    if (order) {
      order.approvalStatus = 'approved'
      order.approvedBy = approvedBy
      order.approvedAt = new Date().toISOString()
      order.status = 'approved'
      order.updatedAt = new Date().toISOString()
    }
  }

  function receiveOrderItem(orderId: string, itemId: string, receivedQuantity: number) {
    const order = orders.value.find(o => o.id === orderId)
    if (!order) return null

    const item = order.items.find(i => i.id === itemId)
    if (!item) return null

    item.receivedQuantity += receivedQuantity
    item.remainingQuantity = item.quantity - item.receivedQuantity

    if (item.receivedQuantity >= item.quantity) {
      item.status = 'received'
    } else if (item.receivedQuantity > 0) {
      item.status = 'partial'
    }

    // Update order status and received value
    const allReceived = order.items.every(i => i.status === 'received')
    const anyReceived = order.items.some(i => i.receivedQuantity > 0)

    if (allReceived) {
      order.status = 'received'
      order.actualDeliveryDate = new Date().toISOString()
    } else if (anyReceived) {
      order.status = 'partial'
    }

    // Calculate received value
    const totalReceived = order.items.reduce((sum, i) => {
      const itemReceivedValue = (i.receivedQuantity / i.quantity) * i.total
      return sum + itemReceivedValue
    }, 0)

    order.receivedValue = totalReceived
    order.receivedPercentage = Math.round((totalReceived / order.totalAmount) * 100)
    order.updatedAt = new Date().toISOString()

    return order
  }

  function cancelOrder(id: string, reason: string) {
    const order = orders.value.find(o => o.id === id)
    if (order) {
      order.status = 'cancelled'
      order.notes = (order.notes || '') + `\n[CANCELLED: ${new Date().toLocaleDateString()}] ${reason}`
      order.updatedAt = new Date().toISOString()
    }
  }

  function deleteOrder(id: string) {
    const index = orders.value.findIndex(o => o.id === id)
    if (index !== -1) {
      orders.value.splice(index, 1)
      return true
    }
    return false
  }

  return {
    orders,
    pendingOrders,
    partialOrders,
    receivedOrders,
    ordersByVendor,
    orderStats,
    upcomingDeliveries,
    overdueOrders,
    getOrderById,
    getOrderByNumber,
    getOrdersByVendor,
    getOrdersByStatus,
    addOrder,
    updateOrder,
    updateOrderStatus,
    approveOrder,
    receiveOrderItem,
    cancelOrder,
    deleteOrder
  }
})
