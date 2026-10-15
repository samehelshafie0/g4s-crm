// Matching of manufacturer and category names the way the product import API
// resolves them: manufacturer by name or code, category by name within its
// manufacturer, both ignoring case and surrounding spaces.

export interface MasterManufacturer { id: string; name: string; code: string; categories: { id: string; name: string }[] }
export interface NamedRow { manufacturerName: string; categoryName: string }
export interface MissingMasters {
  manufacturers: string[]
  categories: { manufacturer: string; category: string }[]
  ambiguous: string[]
}

export const sameMaster = (a: string, b: string) => a.trim().replace(/\s+/g, ' ').toLowerCase() === b.trim().replace(/\s+/g, ' ').toLowerCase()

export function matchManufacturers<T extends MasterManufacturer>(list: T[], name: string): T[] {
  if (!name.trim()) return []
  return list.filter(m => sameMaster(m.name, name) || sameMaster(m.code, name))
}

/** A short unique code for a new manufacturer, e.g. "Bosch Security" → "BOSCHSECUR". */
export function suggestCode(name: string, taken: Iterable<string>): string {
  const used = new Set([...taken].map(code => code.trim().toUpperCase()))
  const base = name.toUpperCase().replace(/[^A-Z0-9]+/g, '').slice(0, 10) || 'MFR'
  let code = base
  for (let n = 2; used.has(code); n++) code = `${base.slice(0, 10 - String(n).length - 1)}-${n}`
  return code
}

/**
 * Manufacturer and category names in the rows that the catalog does not have
 * yet. A blank manufacturer uses the file default, if one was chosen.
 */
export function findMissingMasters(rows: NamedRow[], manufacturers: MasterManufacturer[], defaultManufacturerId = ''): MissingMasters {
  const missing: MissingMasters = { manufacturers: [], categories: [], ambiguous: [] }
  const seen = new Set<string>()
  const add = (key: string, push: () => void) => { if (!seen.has(key)) { seen.add(key); push() } }
  for (const row of rows) {
    const mName = row.manufacturerName.trim()
    const cName = row.categoryName.trim()
    let manufacturer: MasterManufacturer | undefined
    let label = ''
    if (mName) {
      const matches = matchManufacturers(manufacturers, mName)
      if (matches.length > 1) { add(`a:${mName.toLowerCase()}`, () => missing.ambiguous.push(mName)); continue }
      manufacturer = matches[0]
      label = manufacturer?.name ?? mName
      if (!manufacturer) add(`m:${mName.toLowerCase()}`, () => missing.manufacturers.push(mName))
    } else if (defaultManufacturerId) {
      manufacturer = manufacturers.find(m => m.id === defaultManufacturerId)
      label = manufacturer?.name ?? ''
    }
    if (!cName || !label) continue
    if (!manufacturer || !manufacturer.categories.some(c => sameMaster(c.name, cName))) {
      add(`c:${label.toLowerCase()}|${cName.toLowerCase()}`, () => missing.categories.push({ manufacturer: label, category: cName }))
    }
  }
  return missing
}
