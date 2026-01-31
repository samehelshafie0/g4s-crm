import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { StockMovement, WarehouseLocation } from '@/types'

export const useStockMovementsStore = defineStore('stockMovements', () => {
  const movements = ref<StockMovement[]>([
    {
      id: 'MOV-001',
      movementNumber: 'REC-2024-001',
      productId: 'PRD-001',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 150,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-001',
      movementDate: '2024-01-15T10:30:00Z',
      performedBy: 'USR-001',
      notes: 'Initial stock receipt from Hikvision'
    },
    {
      id: 'MOV-002',
      movementNumber: 'TRANS-2024-001',
      productId: 'PRD-001',
      warehouseLocation: 'Jeddah Branch',
      movementType: 'transfer',
      quantity: 80,
      fromLocation: 'Riyadh Main',
      toLocation: 'Jeddah Branch',
      movementDate: '2024-01-16T14:00:00Z',
      performedBy: 'USR-002',
      notes: 'Transfer to Jeddah branch for regional project'
    },
    {
      id: 'MOV-003',
      movementNumber: 'ISS-2024-001',
      productId: 'PRD-001',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 25,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-001',
      movementDate: '2024-01-20T09:15:00Z',
      performedBy: 'USR-003',
      notes: 'Issued for Al-Rajhi Bank project'
    },
    {
      id: 'MOV-004',
      movementNumber: 'REC-2024-002',
      productId: 'PRD-002',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 120,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-002',
      movementDate: '2024-01-15T11:00:00Z',
      performedBy: 'USR-001',
      notes: 'Receipt from Dahua Technology'
    },
    {
      id: 'MOV-005',
      movementNumber: 'ISS-2024-002',
      productId: 'PRD-002',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 30,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-002',
      movementDate: '2024-01-22T10:30:00Z',
      performedBy: 'USR-003',
      notes: 'Issued for Kingdom Hospital security upgrade'
    },
    {
      id: 'MOV-006',
      movementNumber: 'REC-2024-003',
      productId: 'PRD-003',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 8,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-003',
      movementDate: '2024-01-15T15:30:00Z',
      performedBy: 'USR-001',
      notes: 'High-value NVR units'
    },
    {
      id: 'MOV-007',
      movementNumber: 'ISS-2024-003',
      productId: 'PRD-003',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 2,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-001',
      movementDate: '2024-01-20T09:15:00Z',
      performedBy: 'USR-003',
      notes: 'Issued with cameras for Al-Rajhi project'
    },
    {
      id: 'MOV-008',
      movementNumber: 'REC-2024-004',
      productId: 'PRD-004',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 50,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-004',
      movementDate: '2024-01-16T10:00:00Z',
      performedBy: 'USR-001',
      notes: 'ZKTeco access controllers'
    },
    {
      id: 'MOV-009',
      movementNumber: 'TRANS-2024-002',
      productId: 'PRD-004',
      warehouseLocation: 'Dammam Branch',
      movementType: 'transfer',
      quantity: 25,
      fromLocation: 'Riyadh Main',
      toLocation: 'Dammam Branch',
      movementDate: '2024-01-17T13:00:00Z',
      performedBy: 'USR-004',
      notes: 'Stock transfer to Eastern Province'
    },
    {
      id: 'MOV-010',
      movementNumber: 'ISS-2024-004',
      productId: 'PRD-004',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 8,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-003',
      movementDate: '2024-01-23T11:00:00Z',
      performedBy: 'USR-003',
      notes: 'Issued for SABIC facility'
    },
    {
      id: 'MOV-011',
      movementNumber: 'REC-2024-005',
      productId: 'PRD-005',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 500,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-005',
      movementDate: '2024-01-16T14:30:00Z',
      performedBy: 'USR-001',
      notes: 'Bulk order of HID card readers'
    },
    {
      id: 'MOV-012',
      movementNumber: 'TRANS-2024-003',
      productId: 'PRD-005',
      warehouseLocation: 'Jeddah Branch',
      movementType: 'transfer',
      quantity: 300,
      fromLocation: 'Riyadh Main',
      toLocation: 'Jeddah Branch',
      movementDate: '2024-01-18T10:00:00Z',
      performedBy: 'USR-002',
      notes: 'Stock distribution to Jeddah'
    },
    {
      id: 'MOV-013',
      movementNumber: 'ISS-2024-005',
      productId: 'PRD-005',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 75,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-001',
      movementDate: '2024-01-20T09:15:00Z',
      performedBy: 'USR-003',
      notes: 'Card readers for Al-Rajhi Bank branches'
    },
    {
      id: 'MOV-014',
      movementNumber: 'ADJ-2024-001',
      productId: 'PRD-005',
      warehouseLocation: 'Jeddah Branch',
      movementType: 'adjustment',
      quantity: -10,
      referenceType: 'adjustment',
      referenceNumber: 'ADJ-001',
      movementDate: '2024-01-25T16:00:00Z',
      performedBy: 'USR-002',
      notes: 'Stock count adjustment - damaged units'
    },
    {
      id: 'MOV-015',
      movementNumber: 'REC-2024-006',
      productId: 'PRD-008',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 200,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-006',
      movementDate: '2024-01-18T09:00:00Z',
      performedBy: 'USR-001',
      notes: 'Security guard uniforms - local supplier'
    },
    {
      id: 'MOV-016',
      movementNumber: 'TRANS-2024-004',
      productId: 'PRD-008',
      warehouseLocation: 'Jeddah Branch',
      movementType: 'transfer',
      quantity: 150,
      fromLocation: 'Riyadh Main',
      toLocation: 'Jeddah Branch',
      movementDate: '2024-01-19T11:00:00Z',
      performedBy: 'USR-002',
      notes: 'Uniform stock for Jeddah operations'
    },
    {
      id: 'MOV-017',
      movementNumber: 'ISS-2024-006',
      productId: 'PRD-008',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 50,
      referenceType: 'sales_order',
      referenceNumber: 'SO-2024-001',
      movementDate: '2024-01-21T10:00:00Z',
      performedBy: 'USR-003',
      notes: 'Uniforms for new guard deployment'
    },
    {
      id: 'MOV-018',
      movementNumber: 'REC-2024-007',
      productId: 'PRD-014',
      warehouseLocation: 'Riyadh Main',
      movementType: 'receipt',
      quantity: 60,
      referenceType: 'purchase_order',
      referenceNumber: 'PO-2024-007',
      movementDate: '2024-01-21T13:00:00Z',
      performedBy: 'USR-001',
      notes: 'Cisco PoE switches'
    },
    {
      id: 'MOV-019',
      movementNumber: 'TRANS-2024-005',
      productId: 'PRD-014',
      warehouseLocation: 'Jeddah Branch',
      movementType: 'transfer',
      quantity: 35,
      fromLocation: 'Riyadh Main',
      toLocation: 'Jeddah Branch',
      movementDate: '2024-01-22T14:00:00Z',
      performedBy: 'USR-002',
      notes: 'Network equipment for Jeddah projects'
    },
    {
      id: 'MOV-020',
      movementNumber: 'ISS-2024-007',
      productId: 'PRD-014',
      warehouseLocation: 'Riyadh Main',
      movementType: 'issue',
      quantity: 12,
      referenceType: 'quote',
      referenceNumber: 'QT-2024-002',
      movementDate: '2024-01-22T10:30:00Z',
      performedBy: 'USR-003',
      notes: 'Switches for Kingdom Hospital network infrastructure'
    }
  ])

  // Computed getters
  const movementsByType = computed(() => {
    const grouped: Record<string, StockMovement[]> = {
      receipt: [],
      issue: [],
      transfer: [],
      adjustment: []
    }

    movements.value.forEach(movement => {
      grouped[movement.movementType].push(movement)
    })

    return grouped
  })

  const recentMovements = computed(() => {
    return [...movements.value]
      .sort((a, b) => new Date(b.movementDate).getTime() - new Date(a.movementDate).getTime())
      .slice(0, 20)
  })

  // Getters
  function getMovementById(id: string): StockMovement | undefined {
    return movements.value.find(m => m.id === id)
  }

  function getMovementsByProduct(productId: string): StockMovement[] {
    return movements.value.filter(m => m.productId === productId)
  }

  function getMovementsByLocation(location: WarehouseLocation): StockMovement[] {
    return movements.value.filter(m => m.warehouseLocation === location)
  }

  function getMovementsByType(type: 'receipt' | 'issue' | 'transfer' | 'adjustment'): StockMovement[] {
    return movements.value.filter(m => m.movementType === type)
  }

  function getMovementsByDateRange(startDate: string, endDate: string): StockMovement[] {
    const start = new Date(startDate).getTime()
    const end = new Date(endDate).getTime()

    return movements.value.filter(m => {
      const movementTime = new Date(m.movementDate).getTime()
      return movementTime >= start && movementTime <= end
    })
  }

  function getMovementsByReference(referenceType: string, referenceNumber: string): StockMovement[] {
    return movements.value.filter(
      m => m.referenceType === referenceType && m.referenceNumber === referenceNumber
    )
  }

  // CRUD Operations
  function addMovement(movementData: Omit<StockMovement, 'id' | 'movementNumber' | 'movementDate'>): StockMovement {
    // Generate movement number based on type
    const typePrefix = {
      receipt: 'REC',
      issue: 'ISS',
      transfer: 'TRANS',
      adjustment: 'ADJ'
    }[movementData.movementType]

    const year = new Date().getFullYear()
    const typeMovements = movements.value.filter(m => m.movementType === movementData.movementType)
    const sequenceNumber = String(typeMovements.length + 1).padStart(3, '0')

    const newMovement: StockMovement = {
      ...movementData,
      id: `MOV-${String(movements.value.length + 1).padStart(3, '0')}`,
      movementNumber: `${typePrefix}-${year}-${sequenceNumber}`,
      movementDate: new Date().toISOString()
    }

    movements.value.push(newMovement)
    return newMovement
  }

  function updateMovement(id: string, updates: Partial<StockMovement>): boolean {
    const index = movements.value.findIndex(m => m.id === id)
    if (index !== -1) {
      movements.value[index] = {
        ...movements.value[index],
        ...updates
      } as StockMovement
      return true
    }
    return false
  }

  function deleteMovement(id: string): boolean {
    const index = movements.value.findIndex(m => m.id === id)
    if (index !== -1) {
      movements.value.splice(index, 1)
      return true
    }
    return false
  }

  // Analytics
  function getMovementStats() {
    const totalReceipts = movements.value.filter(m => m.movementType === 'receipt').length
    const totalIssues = movements.value.filter(m => m.movementType === 'issue').length
    const totalTransfers = movements.value.filter(m => m.movementType === 'transfer').length
    const totalAdjustments = movements.value.filter(m => m.movementType === 'adjustment').length

    const totalReceiptQty = movements.value
      .filter(m => m.movementType === 'receipt')
      .reduce((sum, m) => sum + m.quantity, 0)

    const totalIssueQty = movements.value
      .filter(m => m.movementType === 'issue')
      .reduce((sum, m) => sum + m.quantity, 0)

    return {
      totalMovements: movements.value.length,
      totalReceipts,
      totalIssues,
      totalTransfers,
      totalAdjustments,
      totalReceiptQty,
      totalIssueQty,
      netMovement: totalReceiptQty - totalIssueQty
    }
  }

  function getProductMovementHistory(productId: string) {
    const productMovements = getMovementsByProduct(productId)

    let totalReceived = 0
    let totalIssued = 0
    let totalAdjustments = 0

    productMovements.forEach(m => {
      if (m.movementType === 'receipt') {
        totalReceived += m.quantity
      } else if (m.movementType === 'issue') {
        totalIssued += m.quantity
      } else if (m.movementType === 'adjustment') {
        totalAdjustments += m.quantity
      }
    })

    return {
      productId,
      totalMovements: productMovements.length,
      totalReceived,
      totalIssued,
      totalAdjustments,
      netQuantity: totalReceived - totalIssued + totalAdjustments,
      movements: productMovements.sort(
        (a, b) => new Date(b.movementDate).getTime() - new Date(a.movementDate).getTime()
      )
    }
  }

  function getLocationActivity(location: WarehouseLocation) {
    const locationMovements = getMovementsByLocation(location)

    const receipts = locationMovements.filter(m => m.movementType === 'receipt').length
    const issues = locationMovements.filter(m => m.movementType === 'issue').length
    const transfersIn = movements.value.filter(
      m => m.movementType === 'transfer' && m.toLocation === location
    ).length
    const transfersOut = movements.value.filter(
      m => m.movementType === 'transfer' && m.fromLocation === location
    ).length

    return {
      location,
      totalMovements: locationMovements.length,
      receipts,
      issues,
      transfersIn,
      transfersOut,
      recentActivity: locationMovements
        .sort((a, b) => new Date(b.movementDate).getTime() - new Date(a.movementDate).getTime())
        .slice(0, 10)
    }
  }

  return {
    movements,
    movementsByType,
    recentMovements,
    getMovementById,
    getMovementsByProduct,
    getMovementsByLocation,
    getMovementsByType,
    getMovementsByDateRange,
    getMovementsByReference,
    addMovement,
    updateMovement,
    deleteMovement,
    getMovementStats,
    getProductMovementHistory,
    getLocationActivity
  }
})
