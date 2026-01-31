import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { WarehouseStock, StockMovement, WarehouseLocation } from '@/types'

export const useWarehouseStore = defineStore('warehouse', () => {
  const stock = ref<WarehouseStock[]>([
    {
      id: 'STK-001',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-001',
      quantityOnHand: 25,
      quantityReserved: 10,
      quantityAvailable: 15,
      unitCost: 13125,
      totalValue: 328125,
      batchNumber: 'BATCH-2024-001',
      supplierId: 'SUP-001',
      supplierSku: 'P20174-B21',
      lastPurchaseDate: '2024-01-15',
      lastPurchasePrice: 13125,
      lastUpdated: '2024-01-20T10:00:00Z',
      notes: 'Good stock levels'
    },
    {
      id: 'STK-002',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-002',
      quantityOnHand: 15,
      quantityReserved: 5,
      quantityAvailable: 10,
      unitCost: 31875,
      totalValue: 478125,
      batchNumber: 'BATCH-2024-002',
      supplierId: 'SUP-002',
      supplierSku: 'C9300-48P',
      lastPurchaseDate: '2024-01-16',
      lastPurchasePrice: 31875,
      lastUpdated: '2024-01-20T10:00:00Z'
    },
    {
      id: 'STK-003',
      warehouseLocation: 'Jeddah Branch',
      productId: 'PRD-001',
      quantityOnHand: 12,
      quantityReserved: 2,
      quantityAvailable: 10,
      unitCost: 13125,
      totalValue: 157500,
      batchNumber: 'BATCH-2024-001',
      lastUpdated: '2024-01-18T14:00:00Z'
    },
    {
      id: 'STK-004',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-003',
      quantityOnHand: 8,
      quantityReserved: 0,
      quantityAvailable: 8,
      unitCost: 45000,
      totalValue: 360000,
      lastPurchaseDate: '2024-01-17',
      lastPurchasePrice: 45000,
      lastUpdated: '2024-01-20T10:00:00Z',
      notes: 'Local supplier, fast delivery'
    },
    {
      id: 'STK-005',
      warehouseLocation: 'Dammam Branch',
      productId: 'PRD-002',
      quantityOnHand: 5,
      quantityReserved: 0,
      quantityAvailable: 5,
      unitCost: 31875,
      totalValue: 159375,
      batchNumber: 'BATCH-2024-002',
      lastUpdated: '2024-01-19T09:00:00Z'
    }
  ])

  const movements = ref<StockMovement[]>([
    {
      id: 'MOV-001',
      movementNumber: 'MVT-2024-001',
      productId: 'PRD-001',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 25,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-001',
      movementDate: '2024-01-15',
      performedBy: 'USR-005',
      notes: 'Initial stock receipt from supplier'
    },
    {
      id: 'MOV-002',
      movementNumber: 'MVT-2024-002',
      productId: 'PRD-002',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 20,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-002',
      movementDate: '2024-01-16',
      performedBy: 'USR-005',
      notes: 'Bulk order from Cisco distributor'
    },
    {
      id: 'MOV-003',
      movementNumber: 'MVT-2024-003',
      productId: 'PRD-002',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 5,
      referenceType: 'sales_order',
      referenceNumber: 'SO-2024-001',
      movementDate: '2024-01-18',
      performedBy: 'USR-006',
      notes: 'Issued for customer delivery'
    },
    {
      id: 'MOV-004',
      movementNumber: 'MVT-2024-004',
      productId: 'PRD-001',
      warehouseLocation: 'Riyadh Main',
      movementType: 'transfer',
      quantity: 10,
      fromLocation: 'Riyadh Main',
      toLocation: 'Jeddah Branch',
      movementDate: '2024-01-18',
      performedBy: 'USR-005',
      notes: 'Stock transfer to Jeddah branch'
    },
    {
      id: 'MOV-005',
      movementNumber: 'MVT-2024-005',
      productId: 'PRD-001',
      warehouseLocation: 'Riyadh Main',
      movementType: 'adjustment',
      quantity: -2,
      referenceType: 'adjustment',
      movementDate: '2024-01-19',
      performedBy: 'USR-005',
      notes: 'Damaged units removed from inventory'
    }
  ])

  // Computed
  const totalInventoryValue = computed(() => {
    return stock.value.reduce((sum, s) => sum + s.totalValue, 0)
  })

  const totalItems = computed(() => {
    return stock.value.reduce((sum, s) => sum + s.quantityOnHand, 0)
  })

  const lowStockAlerts = computed(() => {
    return stock.value.filter(s => s.quantityAvailable < 5)
  })

  const outOfStockItems = computed(() => {
    return stock.value.filter(s => s.quantityAvailable === 0)
  })

  const stockByLocation = computed(() => {
    const byLocation: Record<WarehouseLocation, WarehouseStock[]> = {
      'Riyadh Main': [],
      'Jeddah Branch': [],
      'Dammam Branch': [],
      'Other': []
    }

    stock.value.forEach(s => {
      byLocation[s.warehouseLocation].push(s)
    })

    return byLocation
  })

  // Actions
  function addStock(stockItem: WarehouseStock) {
    stock.value.push(stockItem)
  }

  function updateStock(id: string, updates: Partial<WarehouseStock>) {
    const index = stock.value.findIndex(s => s.id === id)
    if (index !== -1) {
      stock.value[index] = {
        ...stock.value[index],
        ...updates,
        lastUpdated: new Date().toISOString()
      }
    }
  }

  function deleteStock(id: string) {
    const index = stock.value.findIndex(s => s.id === id)
    if (index !== -1) {
      stock.value.splice(index, 1)
    }
  }

  function getStockById(id: string) {
    return stock.value.find(s => s.id === id)
  }

  function getStockByProduct(productId: string) {
    return stock.value.filter(s => s.productId === productId)
  }

  function getStockByLocation(location: WarehouseLocation) {
    return stock.value.filter(s => s.warehouseLocation === location)
  }

  function addMovement(movement: StockMovement) {
    movements.value.push(movement)

    // Update stock levels based on movement
    const stockItem = stock.value.find(s =>
      s.productId === movement.productId &&
      s.warehouseLocation === movement.warehouseLocation
    )

    if (stockItem) {
      switch (movement.movementType) {
        case 'receipt':
          stockItem.quantityOnHand += movement.quantity
          stockItem.quantityAvailable += movement.quantity
          break
        case 'issue':
          stockItem.quantityOnHand -= movement.quantity
          stockItem.quantityAvailable -= movement.quantity
          stockItem.quantityReserved -= movement.quantity
          break
        case 'adjustment':
          stockItem.quantityOnHand += movement.quantity
          stockItem.quantityAvailable += movement.quantity
          break
      }

      stockItem.totalValue = stockItem.quantityOnHand * stockItem.unitCost
      stockItem.lastUpdated = new Date().toISOString()
    }
  }

  function getMovementsByProduct(productId: string) {
    return movements.value.filter(m => m.productId === productId)
  }

  function getMovementsByLocation(location: WarehouseLocation) {
    return movements.value.filter(m => m.warehouseLocation === location)
  }

  function reserveStock(productId: string, location: WarehouseLocation, quantity: number): boolean {
    const stockItem = stock.value.find(s =>
      s.productId === productId &&
      s.warehouseLocation === location
    )

    if (stockItem && stockItem.quantityAvailable >= quantity) {
      stockItem.quantityReserved += quantity
      stockItem.quantityAvailable -= quantity
      stockItem.lastUpdated = new Date().toISOString()
      return true
    }

    return false
  }

  function releaseReservedStock(productId: string, location: WarehouseLocation, quantity: number) {
    const stockItem = stock.value.find(s =>
      s.productId === productId &&
      s.warehouseLocation === location
    )

    if (stockItem) {
      stockItem.quantityReserved -= quantity
      stockItem.quantityAvailable += quantity
      stockItem.lastUpdated = new Date().toISOString()
    }
  }

  function getTotalStockForProduct(productId: string): number {
    return stock.value
      .filter(s => s.productId === productId)
      .reduce((sum, s) => sum + s.quantityOnHand, 0)
  }

  function getAvailableStockForProduct(productId: string): number {
    return stock.value
      .filter(s => s.productId === productId)
      .reduce((sum, s) => sum + s.quantityAvailable, 0)
  }

  return {
    stock,
    movements,
    totalInventoryValue,
    totalItems,
    lowStockAlerts,
    outOfStockItems,
    stockByLocation,
    addStock,
    updateStock,
    deleteStock,
    getStockById,
    getStockByProduct,
    getStockByLocation,
    addMovement,
    getMovementsByProduct,
    getMovementsByLocation,
    reserveStock,
    releaseReservedStock,
    getTotalStockForProduct,
    getAvailableStockForProduct
  }
})
