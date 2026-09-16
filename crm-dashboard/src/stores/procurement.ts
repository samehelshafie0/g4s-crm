import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { procurementService } from '@/services'

export const useProcurementStore = defineStore('procurement', () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const purchaseOrders = ref<any[]>([])
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const supplierQuotes = ref<any[]>([])
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const goodsReceipts = ref<any[]>([])
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const supplierItems = ref<any[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalPOs = computed(() => purchaseOrders.value.length)
  const pendingPOs = computed(() =>
    purchaseOrders.value.filter((p: any) => p.status === 'pending-approval').length,
  )
  const inTransitPOs = computed(() =>
    purchaseOrders.value.filter((p: any) => p.status === 'in-transit' || p.status === 'shipped').length,
  )
  const totalValue = computed(() =>
    purchaseOrders.value.reduce((sum: number, p: any) => sum + (p.totalAmount ?? 0), 0),
  )
  const activeSQCount = computed(() =>
    supplierQuotes.value.filter((s: any) => s.status === 'submitted' || s.status === 'draft').length,
  )
  const sqTotalValue = computed(() =>
    supplierQuotes.value.reduce((sum: number, s: any) => sum + (s.totalAmount ?? 0), 0),
  )
  const totalGRs = computed(() => goodsReceipts.value.length)
  const totalReceivedValue = computed(() =>
    goodsReceipts.value.reduce((sum: number, g: any) => sum + (g.totalValue ?? 0), 0),
  )

  async function fetchPurchaseOrders(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await procurementService.listPOs(params)
      if (res.success) purchaseOrders.value = res.data
    } catch (e) { error.value = 'Failed to load POs'; console.error(e) }
    finally { loading.value = false }
  }

  async function fetchSupplierQuotes(params?: Record<string, unknown>) {
    try {
      const res = await procurementService.listSQs(params)
      if (res.success) supplierQuotes.value = res.data
    } catch (e) { console.error(e) }
  }

  async function fetchGoodsReceipts(params?: Record<string, unknown>) {
    try {
      const res = await procurementService.listGRs(params)
      if (res.success) goodsReceipts.value = res.data
    } catch (e) { console.error(e) }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async function addPurchaseOrder(data: any) {
    const res = await procurementService.createPO(data as Record<string, unknown>)
    if (res.success) { purchaseOrders.value.unshift(res.data); return res.data }
    throw new Error('Failed to create PO')
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async function addSupplierQuote(data: any) {
    const res = await procurementService.createSQ(data as Record<string, unknown>)
    if (res.success) { supplierQuotes.value.unshift(res.data); return res.data }
    throw new Error('Failed to create SQ')
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async function addGoodsReceipt(data: any) {
    const res = await procurementService.createGR(data as Record<string, unknown>)
    if (res.success) { goodsReceipts.value.unshift(res.data); return res.data }
    throw new Error('Failed to create GR')
  }

  // Helper methods (retained for InventoryView compatibility)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function getSupplierQuotesForProduct(productSku: string): any[] {
    return supplierQuotes.value.filter((sq: any) =>
      sq.items?.some((i: any) => i.productSku === productSku || i.product?.sku === productSku),
    )
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function getReceiptHistoryForProduct(productSku: string): any[] {
    return goodsReceipts.value.filter((gr: any) =>
      gr.items?.some((i: any) => i.product?.sku === productSku),
    )
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function getSupplierItemsForProduct(productSku: string): any[] {
    return supplierItems.value.filter((si: any) => si.productSku === productSku)
  }

  function generatePoNumber(): string {
    const count = purchaseOrders.value.length + 1
    return `PO-${new Date().getFullYear()}-${String(count).padStart(4, '0')}`
  }

  function getBestSupplierPrice(productSku: string): { cost: number; supplierName: string } | null {
    const quotes = getSupplierQuotesForProduct(productSku)
    if (!quotes.length) return null
    let best: { cost: number; supplierName: string } | null = null
    for (const sq of quotes) {
      for (const item of sq.items ?? []) {
        if (
          (item.productSku === productSku || item.product?.sku === productSku) &&
          (best === null || item.unitCost < best.cost)
        ) {
          best = { cost: item.unitCost as number, supplierName: sq.supplierName ?? '' }
        }
      }
    }
    return best
  }

  function generateSqNumber(): string {
    const count = supplierQuotes.value.length + 1
    return `SQ-${new Date().getFullYear()}-${String(count).padStart(4, '0')}`
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async function updatePurchaseOrder(id: string, data: any) {
    const res = await procurementService.createPO({ ...data, id })
    if (res.success) {
      const idx = purchaseOrders.value.findIndex((p: any) => p.id === id)
      if (idx !== -1) purchaseOrders.value[idx] = res.data
    }
  }

  async function deletePurchaseOrder(id: string) {
    purchaseOrders.value = purchaseOrders.value.filter((p: any) => p.id !== id)
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async function updateSupplierQuote(id: string, data: any) {
    const res = await procurementService.createSQ({ ...data, id })
    if (res.success) {
      const idx = supplierQuotes.value.findIndex((s: any) => s.id === id)
      if (idx !== -1) supplierQuotes.value[idx] = res.data
    }
  }

  async function deleteSupplierQuote(id: string) {
    supplierQuotes.value = supplierQuotes.value.filter((s: any) => s.id !== id)
  }

  return {
    purchaseOrders,
    supplierQuotes,
    goodsReceipts,
    supplierItems,
    loading,
    error,
    totalPOs,
    pendingPOs,
    inTransitPOs,
    totalValue,
    activeSQCount,
    sqTotalValue,
    totalGRs,
    totalReceivedValue,
    fetchPurchaseOrders,
    fetchSupplierQuotes,
    fetchGoodsReceipts,
    addPurchaseOrder,
    addSupplierQuote,
    addGoodsReceipt,
    getSupplierQuotesForProduct,
    getReceiptHistoryForProduct,
    getSupplierItemsForProduct,
    generatePoNumber,
    generateSqNumber,
    getBestSupplierPrice,
    updatePurchaseOrder,
    deletePurchaseOrder,
    updateSupplierQuote,
    deleteSupplierQuote,
  }
})
