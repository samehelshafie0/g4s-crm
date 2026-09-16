import type { PurchaseOrder, SupplierQuote, GoodsReceipt, SupplierItemEntry } from '@/types'
import { toPurchaseOrder, toSupplierQuote, toGoodsReceipt } from '@/services/procurementDtos'
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { procurementService } from '@/services'

export const useProcurementStore = defineStore('procurement', () => {
  const purchaseOrders = ref<PurchaseOrder[]>([])
  const supplierQuotes = ref<SupplierQuote[]>([])
  const goodsReceipts = ref<GoodsReceipt[]>([])
  const supplierItems = ref<SupplierItemEntry[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalPOs = computed(() => purchaseOrders.value.length)
  const pendingPOs = computed(() =>
    purchaseOrders.value.filter((p) => p.status === 'pending-approval').length,
  )
  const inTransitPOs = computed(() =>
    purchaseOrders.value.filter((p) => p.status === 'ordered' || p.status === 'partial-received').length,
  )
  const totalValue = computed(() =>
    purchaseOrders.value.reduce((sum: number, p: any) => sum + (p.total ?? 0), 0),
  )
  const activeSQCount = computed(() =>
    supplierQuotes.value.filter((s) => s.status === 'received' || s.status === 'under-review').length,
  )
  const sqTotalValue = computed(() =>
    supplierQuotes.value.reduce((sum: number, s: any) => sum + (s.subtotal ?? 0), 0),
  )
  const totalGRs = computed(() => goodsReceipts.value.length)
  const totalReceivedValue = computed(() =>
    goodsReceipts.value.reduce((sum: number, g: any) => sum + (g.totalLandingCost ?? 0), 0),
  )

  async function fetchPurchaseOrders(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await procurementService.listPOs(params)
      if (res.success) purchaseOrders.value = res.data.map(toPurchaseOrder)
    } catch (e) { error.value = 'Failed to load POs'; console.error(e) }
    finally { loading.value = false }
  }

  async function fetchSupplierQuotes(params?: Record<string, unknown>) {
    try {
      const res = await procurementService.listSQs(params)
      if (res.success) supplierQuotes.value = res.data.map(toSupplierQuote)
    } catch (e) { console.error(e) }
  }

  async function fetchGoodsReceipts(params?: Record<string, unknown>) {
    try {
      const res = await procurementService.listGRs(params)
      if (res.success) goodsReceipts.value = res.data.map(toGoodsReceipt)
    } catch (e) { console.error(e) }
  }

  async function addPurchaseOrder(data: any) {
    const res = await procurementService.createPO(data as Record<string, unknown>)
    if (res.success) { purchaseOrders.value.unshift(toPurchaseOrder(res.data)); return toPurchaseOrder(res.data) }
    throw new Error('Failed to create PO')
  }

  async function addSupplierQuote(data: any) {
    const res = await procurementService.createSQ(data as Record<string, unknown>)
    if (res.success) { supplierQuotes.value.unshift(toSupplierQuote(res.data)); return toSupplierQuote(res.data) }
    throw new Error('Failed to create SQ')
  }

  async function addGoodsReceipt(data: any) {
    const res = await procurementService.createGR(data as Record<string, unknown>)
    if (res.success) { goodsReceipts.value.unshift(toGoodsReceipt(res.data)); return toGoodsReceipt(res.data) }
    throw new Error('Failed to create GR')
  }

  // Helper methods (retained for InventoryView compatibility)
  function getSupplierQuotesForProduct(productSku: string) {
    return supplierQuotes.value.flatMap(sq =>
      sq.items.filter(item => item.productSku === productSku).map(item => ({ sq, item })),
    )
  }

  function getReceiptHistoryForProduct(productSku: string) {
    return goodsReceipts.value.flatMap(gr =>
      gr.items.filter(item => item.productSku === productSku).map(item => ({ gr, item })),
    ).sort((a, b) => b.gr.receiveDate.localeCompare(a.gr.receiveDate))
  }

  function getSupplierItemsForProduct(productSku: string) {
    return supplierItems.value.filter((si) => si.productSku === productSku)
  }

  function generatePoNumber(): string {
    const count = purchaseOrders.value.length + 1
    return `PO-${new Date().getFullYear()}-${String(count).padStart(4, '0')}`
  }

  function getBestSupplierPrice(productSku: string): { cost: number; supplierName: string } | null {
    const quotes = getSupplierQuotesForProduct(productSku)
    if (!quotes.length) return null
    let best: { cost: number; supplierName: string } | null = null
    for (const { sq, item } of quotes) {
      if (best === null || item.unitCost < best.cost) {
        best = { cost: item.unitCost, supplierName: sq.supplierName }
      }
    }
    return best
  }

  function generateSqNumber(): string {
    const count = supplierQuotes.value.length + 1
    return `SQ-${new Date().getFullYear()}-${String(count).padStart(4, '0')}`
  }

  async function updatePurchaseOrder(id: string, data: any) {
    const res = await procurementService.createPO({ ...data, id })
    if (res.success) {
      const idx = purchaseOrders.value.findIndex((p) => p.id === id)
      if (idx !== -1) purchaseOrders.value[idx] = toPurchaseOrder(res.data)
    }
  }

  async function deletePurchaseOrder(id: string) {
    purchaseOrders.value = purchaseOrders.value.filter((p) => p.id !== id)
  }

  async function updateSupplierQuote(id: string, data: any) {
    const res = await procurementService.createSQ({ ...data, id })
    if (res.success) {
      const idx = supplierQuotes.value.findIndex((s) => s.id === id)
      if (idx !== -1) supplierQuotes.value[idx] = toSupplierQuote(res.data)
    }
  }

  async function deleteSupplierQuote(id: string) {
    supplierQuotes.value = supplierQuotes.value.filter((s) => s.id !== id)
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
