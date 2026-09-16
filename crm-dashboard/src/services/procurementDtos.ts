import type { GoodsReceipt, GoodsReceiptItem, PurchaseOrder, PurchaseOrderItem, SupplierQuote, SupplierQuoteLineItem } from '@/types'

interface ProductSummary { sku: string; name: string; manufacturer?: { name: string } }
type POItemDto = Omit<PurchaseOrderItem, 'productSku' | 'productName' | 'manufacturerName'> & { product?: ProductSummary }
export type PurchaseOrderDto = Omit<PurchaseOrder, 'items' | 'expectedDelivery' | 'approvedBy'> & {
  items?: POItemDto[]; expectedDelivery?: string; approvedBy?: { firstName: string; lastName: string }
}
export type SupplierQuoteDto = Omit<SupplierQuote, 'items' | 'validFrom' | 'validUntil'> & {
  items?: SupplierQuoteLineItem[]; validFrom?: string; validUntil?: string
}
type GRItemDto = Omit<GoodsReceiptItem, 'productSku' | 'productName'> & { product?: ProductSummary }
export type GoodsReceiptDto = Omit<GoodsReceipt, 'items' | 'poNumber' | 'receivedBy'> & {
  items?: GRItemDto[]; po?: { poNumber: string }; receivedBy?: { firstName: string; lastName: string }
}

export function toPurchaseOrder(dto: PurchaseOrderDto): PurchaseOrder {
  return { ...dto, expectedDelivery: dto.expectedDelivery ?? '',
    approvedBy: dto.approvedBy ? `${dto.approvedBy.firstName} ${dto.approvedBy.lastName}` : undefined,
    items: (dto.items ?? []).map(item => ({ ...item, productSku: item.product?.sku ?? '',
      productName: item.product?.name ?? '', manufacturerName: item.product?.manufacturer?.name ?? '' })) }
}
export function toSupplierQuote(dto: SupplierQuoteDto): SupplierQuote {
  return { ...dto, items: dto.items ?? [], validFrom: dto.validFrom ?? '', validUntil: dto.validUntil ?? '' }
}
export function toGoodsReceipt(dto: GoodsReceiptDto): GoodsReceipt {
  return { ...dto, poNumber: dto.po?.poNumber ?? '',
    receivedBy: dto.receivedBy ? `${dto.receivedBy.firstName} ${dto.receivedBy.lastName}` : '',
    items: (dto.items ?? []).map(item => ({ ...item, productSku: item.product?.sku ?? '', productName: item.product?.name ?? '' })) }
}
