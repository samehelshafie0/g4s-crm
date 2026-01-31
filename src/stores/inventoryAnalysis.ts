import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { useStockMovementsStore } from './stockMovements'
import { useWarehouseStockStore } from './warehouseStock'
import { useProductsStore } from './products'

export interface ProductMovementAnalysis {
  productId: string
  productSku: string
  productName: string
  category: string

  // Movement Statistics (last 90 days)
  totalIssued: number
  totalReceived: number
  totalTransferred: number
  netMovement: number

  // Frequency
  movementCount: number
  avgDailyMovement: number
  avgWeeklyMovement: number
  avgMonthlyMovement: number

  // Current Stock
  currentStock: number
  availableStock: number
  reservedStock: number

  // Classification
  movementClass: 'fast_moving' | 'medium_moving' | 'slow_moving' | 'non_moving'
  abcClass: 'A' | 'B' | 'C' | 'D'

  // Velocity Metrics
  velocityScore: number // 0-100
  turnoverRate: number // times per year
  daysSinceLastMovement: number

  // Stock Health
  stockoutRisk: 'high' | 'medium' | 'low'
  overstockRisk: 'high' | 'medium' | 'low'

  // Recommendations
  reorderPoint: number
  economicOrderQuantity: number
  safetyStock: number

  // Value
  unitCost: number
  totalValue: number
  movementValue: number
}

export const useInventoryAnalysisStore = defineStore('inventoryAnalysis', () => {
  const movementsStore = useStockMovementsStore()
  const warehouseStore = useWarehouseStockStore()
  const productsStore = useProductsStore()

  // Analysis period in days
  const analysisPeriodDays = ref(90)

  // Movement thresholds for classification
  const movementThresholds = ref({
    fast_moving: 50, // More than 50 movements in period
    medium_moving: 20, // 20-50 movements
    slow_moving: 5, // 5-20 movements
    non_moving: 0 // Less than 5 movements
  })

  // ABC Analysis thresholds (based on value contribution)
  const abcThresholds = ref({
    A: 80, // Top 80% of value
    B: 95, // Next 15% (80-95%)
    C: 100 // Bottom 5% (95-100%)
  })

  // Get movements for a specific period
  function getMovementsForPeriod(days: number) {
    const cutoffDate = new Date()
    cutoffDate.setDate(cutoffDate.getDate() - days)

    return movementsStore.movements.filter(m => {
      const movementDate = new Date(m.movementDate)
      return movementDate >= cutoffDate
    })
  }

  // Calculate movement statistics for a product
  function calculateProductMovements(productId: string, days: number = analysisPeriodDays.value) {
    const movements = getMovementsForPeriod(days).filter(m => m.productId === productId)

    const issued = movements
      .filter(m => m.movementType === 'issue')
      .reduce((sum, m) => sum + Math.abs(m.quantity), 0)

    const received = movements
      .filter(m => m.movementType === 'receipt')
      .reduce((sum, m) => sum + m.quantity, 0)

    const transferred = movements
      .filter(m => m.movementType === 'transfer')
      .reduce((sum, m) => sum + Math.abs(m.quantity), 0)

    const netMovement = received - issued

    // Find last movement date
    const lastMovement = movements.length > 0
      ? movements.sort((a, b) => new Date(b.movementDate).getTime() - new Date(a.movementDate).getTime())[0]
      : null

    const daysSinceLastMovement = lastMovement
      ? Math.floor((Date.now() - new Date(lastMovement.movementDate).getTime()) / (1000 * 60 * 60 * 24))
      : 999

    return {
      totalIssued: issued,
      totalReceived: received,
      totalTransferred: transferred,
      netMovement,
      movementCount: movements.length,
      avgDailyMovement: issued / days,
      avgWeeklyMovement: (issued / days) * 7,
      avgMonthlyMovement: (issued / days) * 30,
      daysSinceLastMovement
    }
  }

  // Classify movement speed
  function classifyMovement(movementCount: number): ProductMovementAnalysis['movementClass'] {
    if (movementCount >= movementThresholds.value.fast_moving) return 'fast_moving'
    if (movementCount >= movementThresholds.value.medium_moving) return 'medium_moving'
    if (movementCount >= movementThresholds.value.slow_moving) return 'slow_moving'
    return 'non_moving'
  }

  // Calculate velocity score (0-100)
  function calculateVelocityScore(movementStats: any, currentStock: number): number {
    const { avgDailyMovement, daysSinceLastMovement, movementCount } = movementStats

    // Factors: daily movement rate, recency, frequency
    const movementScore = Math.min((avgDailyMovement / 5) * 40, 40) // Max 40 points
    const recencyScore = Math.max(0, 30 - (daysSinceLastMovement / 3)) // Max 30 points
    const frequencyScore = Math.min((movementCount / 50) * 30, 30) // Max 30 points

    return Math.round(movementScore + recencyScore + frequencyScore)
  }

  // Calculate turnover rate (annual)
  function calculateTurnoverRate(issuedQuantity: number, avgStock: number): number {
    if (avgStock === 0) return 0
    const annualIssued = (issuedQuantity / analysisPeriodDays.value) * 365
    return Number((annualIssued / avgStock).toFixed(2))
  }

  // Calculate reorder point
  function calculateReorderPoint(avgDailyDemand: number, leadTimeDays: number, safetyStock: number): number {
    return Math.ceil((avgDailyDemand * leadTimeDays) + safetyStock)
  }

  // Calculate Economic Order Quantity (EOQ)
  function calculateEOQ(annualDemand: number, orderCost: number, holdingCost: number): number {
    if (holdingCost === 0) return Math.ceil(annualDemand / 12) // Default to monthly order
    return Math.ceil(Math.sqrt((2 * annualDemand * orderCost) / holdingCost))
  }

  // Calculate safety stock
  function calculateSafetyStock(avgDailyDemand: number, leadTimeDays: number): number {
    // Using simple method: 50% of lead time demand
    return Math.ceil(avgDailyDemand * leadTimeDays * 0.5)
  }

  // Assess stockout risk
  function assessStockoutRisk(currentStock: number, avgDailyDemand: number, leadTimeDays: number): 'high' | 'medium' | 'low' {
    const daysOfStock = avgDailyDemand > 0 ? currentStock / avgDailyDemand : 999

    if (daysOfStock < leadTimeDays) return 'high'
    if (daysOfStock < leadTimeDays * 2) return 'medium'
    return 'low'
  }

  // Assess overstock risk
  function assessOverstockRisk(currentStock: number, avgMonthlyDemand: number): 'high' | 'medium' | 'low' {
    if (avgMonthlyDemand === 0) return 'low'
    const monthsOfStock = currentStock / avgMonthlyDemand

    if (monthsOfStock > 6) return 'high'
    if (monthsOfStock > 3) return 'medium'
    return 'low'
  }

  // Complete product analysis
  const productAnalysis = computed((): ProductMovementAnalysis[] => {
    const products = productsStore.getActiveProducts()
    const analyses: ProductMovementAnalysis[] = []

    products.forEach(product => {
      const movementStats = calculateProductMovements(product.id)
      const stockItems = warehouseStore.stock.filter(s => s.productId === product.id)

      const currentStock = stockItems.reduce((sum, s) => sum + s.quantityOnHand, 0)
      const availableStock = stockItems.reduce((sum, s) => sum + s.quantityAvailable, 0)
      const reservedStock = stockItems.reduce((sum, s) => sum + s.quantityReserved, 0)

      const movementClass = classifyMovement(movementStats.movementCount)
      const velocityScore = calculateVelocityScore(movementStats, currentStock)
      const turnoverRate = calculateTurnoverRate(movementStats.totalIssued, currentStock || 1)

      const leadTimeDays = product.leadTimeDays || 30
      const safetyStock = calculateSafetyStock(movementStats.avgDailyMovement, leadTimeDays)
      const reorderPoint = calculateReorderPoint(movementStats.avgDailyMovement, leadTimeDays, safetyStock)

      const annualDemand = (movementStats.totalIssued / analysisPeriodDays.value) * 365
      const orderCost = 500 // Assumed order cost in SAR
      const holdingCostPercent = 0.20 // 20% of product value
      const holdingCost = product.landedCostSAR * holdingCostPercent
      const eoq = calculateEOQ(annualDemand, orderCost, holdingCost)

      const stockoutRisk = assessStockoutRisk(currentStock, movementStats.avgDailyMovement, leadTimeDays)
      const overstockRisk = assessOverstockRisk(currentStock, movementStats.avgMonthlyMovement)

      const totalValue = currentStock * product.landedCostSAR
      const movementValue = movementStats.totalIssued * product.landedCostSAR

      analyses.push({
        productId: product.id,
        productSku: product.sku,
        productName: product.name,
        category: product.category,
        totalIssued: movementStats.totalIssued,
        totalReceived: movementStats.totalReceived,
        totalTransferred: movementStats.totalTransferred,
        netMovement: movementStats.netMovement,
        movementCount: movementStats.movementCount,
        avgDailyMovement: Number(movementStats.avgDailyMovement.toFixed(2)),
        avgWeeklyMovement: Number(movementStats.avgWeeklyMovement.toFixed(2)),
        avgMonthlyMovement: Number(movementStats.avgMonthlyMovement.toFixed(2)),
        currentStock,
        availableStock,
        reservedStock,
        movementClass,
        abcClass: 'C', // Will be calculated separately
        velocityScore,
        turnoverRate,
        daysSinceLastMovement: movementStats.daysSinceLastMovement,
        stockoutRisk,
        overstockRisk,
        reorderPoint: Math.max(reorderPoint, 0),
        economicOrderQuantity: eoq,
        safetyStock,
        unitCost: product.landedCostSAR,
        totalValue,
        movementValue
      })
    })

    // Perform ABC analysis based on movement value
    const sortedByValue = [...analyses].sort((a, b) => b.movementValue - a.movementValue)
    const totalMovementValue = sortedByValue.reduce((sum, a) => sum + a.movementValue, 0)

    let cumulativeValue = 0
    sortedByValue.forEach(analysis => {
      cumulativeValue += analysis.movementValue
      const cumulativePercent = (cumulativeValue / totalMovementValue) * 100

      if (cumulativePercent <= 80) {
        analysis.abcClass = 'A'
      } else if (cumulativePercent <= 95) {
        analysis.abcClass = 'B'
      } else {
        analysis.abcClass = 'C'
      }
    })

    return analyses
  })

  // Filtered views
  const fastMovingProducts = computed(() =>
    productAnalysis.value.filter(p => p.movementClass === 'fast_moving')
  )

  const mediumMovingProducts = computed(() =>
    productAnalysis.value.filter(p => p.movementClass === 'medium_moving')
  )

  const slowMovingProducts = computed(() =>
    productAnalysis.value.filter(p => p.movementClass === 'slow_moving')
  )

  const nonMovingProducts = computed(() =>
    productAnalysis.value.filter(p => p.movementClass === 'non_moving')
  )

  const highStockoutRisk = computed(() =>
    productAnalysis.value.filter(p => p.stockoutRisk === 'high')
  )

  const highOverstockRisk = computed(() =>
    productAnalysis.value.filter(p => p.overstockRisk === 'high')
  )

  const classAProducts = computed(() =>
    productAnalysis.value.filter(p => p.abcClass === 'A')
  )

  const classBProducts = computed(() =>
    productAnalysis.value.filter(p => p.abcClass === 'B')
  )

  const classCProducts = computed(() =>
    productAnalysis.value.filter(p => p.abcClass === 'C')
  )

  // Summary statistics
  const analysisStats = computed(() => {
    const total = productAnalysis.value.length

    return {
      totalProducts: total,
      fastMoving: fastMovingProducts.value.length,
      mediumMoving: mediumMovingProducts.value.length,
      slowMoving: slowMovingProducts.value.length,
      nonMoving: nonMovingProducts.value.length,
      classA: classAProducts.value.length,
      classB: classBProducts.value.length,
      classC: classCProducts.value.length,
      highStockoutRisk: highStockoutRisk.value.length,
      highOverstockRisk: highOverstockRisk.value.length,
      totalStockValue: productAnalysis.value.reduce((sum, p) => sum + p.totalValue, 0),
      totalMovementValue: productAnalysis.value.reduce((sum, p) => sum + p.movementValue, 0),
      avgTurnoverRate: productAnalysis.value.reduce((sum, p) => sum + p.turnoverRate, 0) / total,
      avgVelocityScore: productAnalysis.value.reduce((sum, p) => sum + p.velocityScore, 0) / total
    }
  })

  // Actions
  function getProductAnalysis(productId: string): ProductMovementAnalysis | undefined {
    return productAnalysis.value.find(p => p.productId === productId)
  }

  function getProductsBelowReorderPoint(): ProductMovementAnalysis[] {
    return productAnalysis.value.filter(p => p.currentStock <= p.reorderPoint)
  }

  function getSuggestedReorders(): Array<ProductMovementAnalysis & { suggestedQuantity: number }> {
    return getProductsBelowReorderPoint().map(p => ({
      ...p,
      suggestedQuantity: p.economicOrderQuantity
    }))
  }

  function updateAnalysisPeriod(days: number) {
    analysisPeriodDays.value = days
  }

  function updateMovementThresholds(thresholds: Partial<typeof movementThresholds.value>) {
    movementThresholds.value = { ...movementThresholds.value, ...thresholds }
  }

  return {
    analysisPeriodDays,
    movementThresholds,
    productAnalysis,
    fastMovingProducts,
    mediumMovingProducts,
    slowMovingProducts,
    nonMovingProducts,
    highStockoutRisk,
    highOverstockRisk,
    classAProducts,
    classBProducts,
    classCProducts,
    analysisStats,
    getProductAnalysis,
    getProductsBelowReorderPoint,
    getSuggestedReorders,
    updateAnalysisPeriod,
    updateMovementThresholds
  }
})
