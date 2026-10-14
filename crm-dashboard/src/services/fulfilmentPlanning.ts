import type { QuoteLineItem, WarehouseStock, StockReservation, PurchaseOrder } from '../types/index.ts'

export interface HoldPlanItem {
  productId: string; productSku: string; productName: string
  warehouseLocation: WarehouseStock['warehouseLocation']; availableQty: number
  holdQty: number; selected: boolean; quoteLineItemId?: string; remainingQty?: number; poCreated?: string
}

export function planQuoteHolds(lines: (Pick<QuoteLineItem, 'id' | 'productId' | 'sku' | 'description' | 'quantity' | 'fulfilmentSource'> & {rowType?: string; isOptional?: boolean; isSelected?: boolean; multiplier?: number; remainingQty?: number})[], stock: WarehouseStock[], holds: StockReservation[], orders: PurchaseOrder[]): HoldPlanItem[] {
  const result: HoldPlanItem[] = []
  const available = new Map(stock.map(row => [row.id, row.availableQty]))
  for (const line of lines) {
    if (!line.productId || (line.rowType && line.rowType !== 'item') || (line.isOptional && !line.isSelected) || line.fulfilmentSource === 'service') continue
    const held = holds.filter(h => h.quoteLineItemId === line.id && h.status !== 'released').reduce((sum, h) => sum + h.qty, 0)
    const ordered = orders.filter(po => po.status !== 'cancelled').flatMap(po => po.items).filter(i => i.quoteLineItemId === line.id).reduce((sum, i) => sum + Math.max(0, i.quantity - i.receivedQty), 0)
    const needed = line.remainingQty ?? Math.max(0, Math.ceil(line.quantity * (line.multiplier || 1)) - held - ordered)
    let remaining = needed
    for (const row of stock.filter(row => row.productId === line.productId)) {
      const free = available.get(row.id) || 0
      const qty = Math.min(free, remaining)
      if (qty <= 0) continue
      result.push({ productId: row.productId, productSku: row.productSku, productName: line.description, warehouseLocation: row.warehouseLocation, availableQty: qty, holdQty: qty, selected: true, quoteLineItemId: line.id, remainingQty: needed })
      available.set(row.id, free - qty)
      remaining -= qty
    }
    if (remaining > 0) result.push({ productId: line.productId, productSku: line.sku || '', productName: line.description, warehouseLocation: 'riyadh-main', availableQty: 0, holdQty: remaining, selected: false, quoteLineItemId: line.id, remainingQty: needed })
  }
  return result
}
