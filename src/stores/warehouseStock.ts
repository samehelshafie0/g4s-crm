import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { WarehouseStock, WarehouseLocation } from '@/types'

export const useWarehouseStockStore = defineStore('warehouseStock', () => {
  const stock = ref<WarehouseStock[]>([
    {
      id: 'WHS-001',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-001',
      quantityOnHand: 150,
      quantityReserved: 25,
      quantityAvailable: 125,
      unitCost: 756,
      totalValue: 113400,
      batchNumber: 'BATCH-2024-HK-001',
      supplierId: 'SUP-001',
      supplierSku: 'DS-2CD2385G1-I',
      lastPurchaseDate: '2024-01-15T00:00:00Z',
      lastPurchasePrice: 756,
      lastUpdated: '2024-01-15T00:00:00Z',
      notes: 'Good stock level'
    },
    {
      id: 'WHS-002',
      warehouseLocation: 'Jeddah Branch',
      productId: 'PRD-001',
      quantityOnHand: 80,
      quantityReserved: 10,
      quantityAvailable: 70,
      unitCost: 756,
      totalValue: 60480,
      batchNumber: 'BATCH-2024-HK-001',
      supplierId: 'SUP-001',
      supplierSku: 'DS-2CD2385G1-I',
      lastPurchaseDate: '2024-01-15T00:00:00Z',
      lastPurchasePrice: 756,
      lastUpdated: '2024-01-15T00:00:00Z',
      notes: 'Regional stock'
    },
    {
      id: 'WHS-003',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-002',
      quantityOnHand: 120,
      quantityReserved: 30,
      quantityAvailable: 90,
      unitCost: 924,
      totalValue: 110880,
      batchNumber: 'BATCH-2024-DH-001',
      supplierId: 'SUP-002',
      supplierSku: 'IPC-HFW8542E-ZE',
      lastPurchaseDate: '2024-01-15T00:00:00Z',
      lastPurchasePrice: 924,
      lastUpdated: '2024-01-15T00:00:00Z',
      notes: 'Popular model, fast-moving'
    },
    {
      id: 'WHS-004',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-003',
      quantityOnHand: 8,
      quantityReserved: 2,
      quantityAvailable: 6,
      unitCost: 15094,
      totalValue: 120752,
      batchNumber: 'BATCH-2024-HK-NVR',
      supplierId: 'SUP-001',
      supplierSku: 'DS-96256NI-I24',
      lastPurchaseDate: '2024-01-15T00:00:00Z',
      lastPurchasePrice: 15094,
      lastUpdated: '2024-01-15T00:00:00Z',
      notes: 'High-value item'
    },
    {
      id: 'WHS-005',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-004',
      quantityOnHand: 45,
      quantityReserved: 8,
      quantityAvailable: 37,
      unitCost: 1907,
      totalValue: 85815,
      batchNumber: 'BATCH-2024-ZK-001',
      supplierId: 'SUP-003',
      supplierSku: 'C3-400',
      lastPurchaseDate: '2024-01-16T00:00:00Z',
      lastPurchasePrice: 1907,
      lastUpdated: '2024-01-16T00:00:00Z'
    },
    {
      id: 'WHS-006',
      warehouseLocation: 'Dammam Branch',
      productId: 'PRD-004',
      quantityOnHand: 25,
      quantityReserved: 5,
      quantityAvailable: 20,
      unitCost: 1907,
      totalValue: 47675,
      batchNumber: 'BATCH-2024-ZK-001',
      supplierId: 'SUP-003',
      supplierSku: 'C3-400',
      lastPurchaseDate: '2024-01-16T00:00:00Z',
      lastPurchasePrice: 1907,
      lastUpdated: '2024-01-16T00:00:00Z'
    },
    {
      id: 'WHS-007',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-005',
      quantityOnHand: 500,
      quantityReserved: 75,
      quantityAvailable: 425,
      unitCost: 357,
      totalValue: 178500,
      batchNumber: 'BATCH-2024-HID-001',
      supplierId: 'SUP-004',
      supplierSku: '5365EGK00',
      lastPurchaseDate: '2024-01-16T00:00:00Z',
      lastPurchasePrice: 357,
      lastUpdated: '2024-01-16T00:00:00Z',
      notes: 'High-volume consumable item'
    },
    {
      id: 'WHS-008',
      warehouseLocation: 'Jeddah Branch',
      productId: 'PRD-005',
      quantityOnHand: 300,
      quantityReserved: 40,
      quantityAvailable: 260,
      unitCost: 357,
      totalValue: 107100,
      batchNumber: 'BATCH-2024-HID-001',
      supplierId: 'SUP-004',
      supplierSku: '5365EGK00',
      lastPurchaseDate: '2024-01-16T00:00:00Z',
      lastPurchasePrice: 357,
      lastUpdated: '2024-01-16T00:00:00Z'
    },
    {
      id: 'WHS-009',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-006',
      quantityOnHand: 15,
      quantityReserved: 3,
      quantityAvailable: 12,
      unitCost: 4160,
      totalValue: 62400,
      batchNumber: 'BATCH-2024-BOSCH-001',
      supplierId: 'SUP-005',
      supplierSku: 'B9512G',
      lastPurchaseDate: '2024-01-17T00:00:00Z',
      lastPurchasePrice: 4160,
      lastUpdated: '2024-01-17T00:00:00Z'
    },
    {
      id: 'WHS-010',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-007',
      quantityOnHand: 12,
      quantityReserved: 4,
      quantityAvailable: 8,
      unitCost: 12075,
      totalValue: 144900,
      batchNumber: 'BATCH-2024-NOTIFIER-001',
      supplierId: 'SUP-006',
      supplierSku: 'NFS2-3030',
      lastPurchaseDate: '2024-01-17T00:00:00Z',
      lastPurchasePrice: 12075,
      lastUpdated: '2024-01-17T00:00:00Z',
      notes: 'Critical safety equipment'
    },
    {
      id: 'WHS-011',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-008',
      quantityOnHand: 200,
      quantityReserved: 50,
      quantityAvailable: 150,
      unitCost: 180,
      totalValue: 36000,
      supplierId: 'SUP-007',
      lastPurchaseDate: '2024-01-18T00:00:00Z',
      lastPurchasePrice: 180,
      lastUpdated: '2024-01-18T00:00:00Z',
      notes: 'Local manufacture, quick replenishment'
    },
    {
      id: 'WHS-012',
      warehouseLocation: 'Jeddah Branch',
      productId: 'PRD-008',
      quantityOnHand: 150,
      quantityReserved: 30,
      quantityAvailable: 120,
      unitCost: 180,
      totalValue: 27000,
      supplierId: 'SUP-007',
      lastPurchaseDate: '2024-01-18T00:00:00Z',
      lastPurchasePrice: 180,
      lastUpdated: '2024-01-18T00:00:00Z'
    },
    {
      id: 'WHS-013',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-009',
      quantityOnHand: 85,
      quantityReserved: 15,
      quantityAvailable: 70,
      unitCost: 1344,
      totalValue: 114240,
      batchNumber: 'BATCH-2024-MOT-001',
      supplierId: 'SUP-008',
      supplierSku: 'DP4400e',
      lastPurchaseDate: '2024-01-18T00:00:00Z',
      lastPurchasePrice: 1344,
      lastUpdated: '2024-01-18T00:00:00Z'
    },
    {
      id: 'WHS-014',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-013',
      quantityOnHand: 5,
      quantityReserved: 1,
      quantityAvailable: 4,
      unitCost: 36019,
      totalValue: 180095,
      batchNumber: 'BATCH-2024-DELL-001',
      supplierId: 'SUP-010',
      supplierSku: 'PER750-001',
      lastPurchaseDate: '2024-01-21T00:00:00Z',
      lastPurchasePrice: 36019,
      lastUpdated: '2024-01-21T00:00:00Z',
      notes: 'High-value server inventory'
    },
    {
      id: 'WHS-015',
      warehouseLocation: 'Riyadh Main',
      productId: 'PRD-014',
      quantityOnHand: 60,
      quantityReserved: 12,
      quantityAvailable: 48,
      unitCost: 2730,
      totalValue: 163800,
      batchNumber: 'BATCH-2024-CISCO-001',
      supplierId: 'SUP-011',
      supplierSku: 'SG350-28P-K9',
      lastPurchaseDate: '2024-01-21T00:00:00Z',
      lastPurchasePrice: 2730,
      lastUpdated: '2024-01-21T00:00:00Z'
    },
    {
      id: 'WHS-016',
      warehouseLocation: 'Jeddah Branch',
      productId: 'PRD-014',
      quantityOnHand: 35,
      quantityReserved: 8,
      quantityAvailable: 27,
      unitCost: 2730,
      totalValue: 95550,
      batchNumber: 'BATCH-2024-CISCO-001',
      supplierId: 'SUP-011',
      supplierSku: 'SG350-28P-K9',
      lastPurchaseDate: '2024-01-21T00:00:00Z',
      lastPurchasePrice: 2730,
      lastUpdated: '2024-01-21T00:00:00Z'
    }
  ])

  // Computed getters
  const stockByLocation = computed(() => {
    const grouped: Record<WarehouseLocation, WarehouseStock[]> = {
      'Riyadh Main': [],
      'Jeddah Branch': [],
      'Dammam Branch': [],
      'Other': []
    }

    stock.value.forEach(item => {
      grouped[item.warehouseLocation].push(item)
    })

    return grouped
  })

  // Getters
  function getStockById(id: string): WarehouseStock | undefined {
    return stock.value.find(s => s.id === id)
  }

  function getStockByProductId(productId: string): WarehouseStock[] {
    return stock.value.filter(s => s.productId === productId)
  }

  function getStockByLocation(location: WarehouseLocation): WarehouseStock[] {
    return stock.value.filter(s => s.warehouseLocation === location)
  }

  function getStockByProductAndLocation(productId: string, location: WarehouseLocation): WarehouseStock | undefined {
    return stock.value.find(s => s.productId === productId && s.warehouseLocation === location)
  }

  function getTotalQuantityForProduct(productId: string): number {
    return stock.value
      .filter(s => s.productId === productId)
      .reduce((sum, s) => sum + s.quantityOnHand, 0)
  }

  function getTotalAvailableForProduct(productId: string): number {
    return stock.value
      .filter(s => s.productId === productId)
      .reduce((sum, s) => sum + s.quantityAvailable, 0)
  }

  function getLowStockItems(threshold: number = 10): WarehouseStock[] {
    return stock.value.filter(s => s.quantityAvailable < threshold)
  }

  // CRUD Operations
  function addStock(stockData: Omit<WarehouseStock, 'id' | 'lastUpdated'>): WarehouseStock {
    const newStock: WarehouseStock = {
      ...stockData,
      id: `WHS-${String(stock.value.length + 1).padStart(3, '0')}`,
      lastUpdated: new Date().toISOString()
    }
    stock.value.push(newStock)
    return newStock
  }

  function updateStock(id: string, updates: Partial<WarehouseStock>): boolean {
    const index = stock.value.findIndex(s => s.id === id)
    if (index !== -1) {
      const currentStock = stock.value[index]
      const updatedStock = {
        ...currentStock,
        ...updates,
        lastUpdated: new Date().toISOString()
      } as WarehouseStock

      // Recalculate derived values
      updatedStock.quantityAvailable = updatedStock.quantityOnHand - updatedStock.quantityReserved
      updatedStock.totalValue = updatedStock.quantityOnHand * updatedStock.unitCost

      stock.value[index] = updatedStock
      return true
    }
    return false
  }

  function deleteStock(id: string): boolean {
    const index = stock.value.findIndex(s => s.id === id)
    if (index !== -1) {
      stock.value.splice(index, 1)
      return true
    }
    return false
  }

  // Stock Operations
  function reserveStock(id: string, quantity: number): boolean {
    const stockItem = stock.value.find(s => s.id === id)
    if (stockItem && stockItem.quantityAvailable >= quantity) {
      stockItem.quantityReserved += quantity
      stockItem.quantityAvailable -= quantity
      stockItem.lastUpdated = new Date().toISOString()
      return true
    }
    return false
  }

  function releaseReservedStock(id: string, quantity: number): boolean {
    const stockItem = stock.value.find(s => s.id === id)
    if (stockItem && stockItem.quantityReserved >= quantity) {
      stockItem.quantityReserved -= quantity
      stockItem.quantityAvailable += quantity
      stockItem.lastUpdated = new Date().toISOString()
      return true
    }
    return false
  }

  function adjustQuantity(id: string, quantityChange: number, reason?: string): boolean {
    const stockItem = stock.value.find(s => s.id === id)
    if (stockItem) {
      stockItem.quantityOnHand += quantityChange
      stockItem.quantityAvailable = stockItem.quantityOnHand - stockItem.quantityReserved
      stockItem.totalValue = stockItem.quantityOnHand * stockItem.unitCost
      stockItem.lastUpdated = new Date().toISOString()
      if (reason) {
        stockItem.notes = reason
      }
      return true
    }
    return false
  }

  // Analytics
  function getWarehouseStats() {
    const totalValue = stock.value.reduce((sum, s) => sum + s.totalValue, 0)
    const totalItems = stock.value.reduce((sum, s) => sum + s.quantityOnHand, 0)
    const totalAvailable = stock.value.reduce((sum, s) => sum + s.quantityAvailable, 0)
    const totalReserved = stock.value.reduce((sum, s) => sum + s.quantityReserved, 0)
    const lowStockCount = getLowStockItems().length

    const valueByLocation: Record<WarehouseLocation, number> = {
      'Riyadh Main': 0,
      'Jeddah Branch': 0,
      'Dammam Branch': 0,
      'Other': 0
    }

    stock.value.forEach(item => {
      valueByLocation[item.warehouseLocation] += item.totalValue
    })

    return {
      totalValue,
      totalItems,
      totalAvailable,
      totalReserved,
      lowStockCount,
      valueByLocation,
      stockLocations: Object.keys(valueByLocation).length,
      uniqueProducts: new Set(stock.value.map(s => s.productId)).size
    }
  }

  function getLocationStats(location: WarehouseLocation) {
    const locationStock = getStockByLocation(location)
    const totalValue = locationStock.reduce((sum, s) => sum + s.totalValue, 0)
    const totalItems = locationStock.reduce((sum, s) => sum + s.quantityOnHand, 0)
    const totalAvailable = locationStock.reduce((sum, s) => sum + s.quantityAvailable, 0)
    const uniqueProducts = new Set(locationStock.map(s => s.productId)).size

    return {
      location,
      totalValue,
      totalItems,
      totalAvailable,
      uniqueProducts,
      stockCount: locationStock.length
    }
  }

  return {
    stock,
    stockByLocation,
    getStockById,
    getStockByProductId,
    getStockByLocation,
    getStockByProductAndLocation,
    getTotalQuantityForProduct,
    getTotalAvailableForProduct,
    getLowStockItems,
    addStock,
    updateStock,
    deleteStock,
    reserveStock,
    releaseReservedStock,
    adjustQuantity,
    getWarehouseStats,
    getLocationStats
  }
})
