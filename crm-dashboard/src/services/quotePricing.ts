export function priceForMargin(cost: number, margin: number, discount = 0): number {
  if (![cost, margin, discount].every(Number.isFinite) || cost <= 0 || margin >= 100 || discount < 0 || discount >= 100) throw new Error('Use a positive cost, a margin below 100%, and a discount below 100%.')
  const price = cost / ((1 - margin / 100) * (1 - discount / 100))
  if (price > 1_000_000_000) throw new Error('The resulting selling price is too large.')
  return Math.round(price * 100) / 100
}
export function inPriceBook(bookId: string, entries: ReadonlySet<string>, kind: string, id: string): boolean {
  return !bookId || entries.has(`${kind}:${id}`)
}

// Reuse an actual net sale price, converting through the historical pinned SAR
// rate and adjusting for discounts already present on the new quotation.
export function priceFromSale(netPrice: number, historicalFX: number, quoteFX: number, lineDiscount = 0, quoteDiscount = 0): number {
 if (![netPrice, historicalFX, quoteFX, lineDiscount, quoteDiscount].every(Number.isFinite) || netPrice < 0 || historicalFX <= 0 || quoteFX <= 0 || lineDiscount < 0 || lineDiscount >= 100 || quoteDiscount < 0 || quoteDiscount >= 100) throw new Error('Set valid exchange rates and discounts below 100% before applying a sale price.')
 const price = netPrice * historicalFX / quoteFX / (1-lineDiscount/100) / (1-quoteDiscount/100)
 if (price > 1_000_000_000) throw new Error('The resulting selling price is too large.')
 return Math.round(price*100)/100
}
