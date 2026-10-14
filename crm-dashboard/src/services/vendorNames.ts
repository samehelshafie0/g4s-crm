export function canonicalVendorName(name: string): string {
  const cleaned = name.trim().replace(/\s+/g, ' ')
  if (['alfanar electric', 'alfanar electric (riyadh)'].includes(cleaned.toLowerCase())) return 'alfanar Electric'
  return cleaned
}
export function vendorKey(name: string): string { return canonicalVendorName(name).toLowerCase() }
export function usableVendorName(name: string): boolean { return !!vendorKey(name) && vendorKey(name) !== 'in' }
