import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { inventoryService } from '@/services'

export const useWarehouseStockStore = defineStore('warehouseStock', () => {
  const stock = ref<unknown[]>([])
  const reservations = ref<unknown[]>([])
  const movements = ref<unknown[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const totalInventoryValue = computed(() =>
    (stock.value as Array<{ totalValue?: number }>).reduce((s, i) => s + (i.totalValue ?? 0), 0),
  )

  const lowStockItems = computed(() =>
    (stock.value as Array<{ reorderLevel?: number; onHandQty?: number }>).filter(
      (i) => i.reorderLevel != null && (i.onHandQty ?? 0) <= i.reorderLevel,
    ),
  )

  async function fetchStock(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await inventoryService.listStock(params)
      if (res.success) stock.value = res.data
    } catch (e) { error.value = 'Failed to load stock'; console.error(e) }
    finally { loading.value = false }
  }

  async function fetchLowStock() {
    try {
      const res = await inventoryService.lowStock()
      if (res.success) stock.value = res.data
    } catch (e) { console.error(e) }
  }

  async function fetchMovements(params?: Record<string, unknown>) {
    try {
      const res = await inventoryService.listMovements(params)
      if (res.success) movements.value = res.data
    } catch (e) { console.error(e) }
  }

  async function createReservation(data: Record<string, unknown>) {
    const res = await inventoryService.createReservation(data)
    if (res.success) return res.data
    throw new Error('Failed to create reservation')
  }

  async function transfer(data: Record<string, unknown>) {
    const res = await inventoryService.transfer(data)
    if (res.success) return res.data
    throw new Error('Failed to transfer')
  }

  return {
    stock,
    reservations,
    movements,
    loading,
    error,
    totalInventoryValue,
    lowStockItems,
    fetchStock,
    fetchLowStock,
    fetchMovements,
    createReservation,
    transfer,
  }
})
