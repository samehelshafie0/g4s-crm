import * as XLSX from 'xlsx'
import { getDocument, GlobalWorkerOptions } from 'pdfjs-dist'

GlobalWorkerOptions.workerSrc = new URL(
  'pdfjs-dist/build/pdf.worker.min.mjs',
  import.meta.url,
).toString()

// ─── Types ──────────────────────────────────────────────────

export interface ParsedRow {
  [key: string]: string
}

export interface ColumnMapping {
  field: string
  aliases: string[]
  keywords: string[]
  /** Negative keywords — if header contains these, skip it for this field */
  antiKeywords?: string[]
  contentDetectable?: boolean
  contentPattern?: (values: string[]) => number
}

export interface ParseResult {
  headers: string[]
  rows: ParsedRow[]
  mappedColumns: Record<string, number>
  rawRowCount: number
  detectionMethod: Record<string, string>
}

export type FileType = 'csv' | 'excel' | 'pdf'

export function detectFileType(fileName: string): FileType | null {
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  if (['csv', 'tsv', 'txt'].includes(ext)) return 'csv'
  if (['xlsx', 'xls', 'xlsb', 'xlsm', 'ods'].includes(ext)) return 'excel'
  if (ext === 'pdf') return 'pdf'
  return null
}

// ─── Content pattern detectors ──────────────────────────────

function skuPattern(values: string[]): number {
  if (values.length === 0) return 0
  let hits = 0
  for (const v of values) {
    const t = v.trim()
    if (!t) continue
    const hasLetter = /[a-zA-Z]/.test(t)
    const hasDigit = /\d/.test(t)
    const hasSep = /[-_./]/.test(t)
    const shortEnough = t.length <= 40
    const noSpaces = !t.includes(' ') || t.split(' ').length <= 3
    if (hasLetter && hasDigit && shortEnough && noSpaces) hits++
    else if (hasSep && (hasLetter || hasDigit) && shortEnough) hits++
  }
  return hits / values.length
}

function namePattern(values: string[]): number {
  if (values.length === 0) return 0
  let hits = 0
  for (const v of values) {
    const t = v.trim()
    if (!t) continue
    const hasLetter = /[a-zA-Z]/.test(t)
    const wordy = t.length >= 8
    const multiWord = t.split(/\s+/).length >= 2
    if (hasLetter && (wordy || multiWord)) hits++
  }
  return hits / values.length
}

function pricePattern(values: string[]): number {
  if (values.length === 0) return 0
  let hits = 0
  for (const v of values) {
    const t = v.trim().replace(/[$€£¥,\s]|SAR|AED|USD|GBP|EUR/gi, '')
    if (!t) continue
    const num = parseFloat(t)
    if (!isNaN(num) && num > 0 && num < 10_000_000) {
      if (t.includes('.') || num >= 1) hits++
    }
  }
  return hits / values.length
}

function qtyPattern(values: string[]): number {
  if (values.length === 0) return 0
  let hits = 0
  for (const v of values) {
    const t = v.trim().replace(/[,\s]/g, '')
    if (!t) continue
    const num = parseInt(t, 10)
    if (!isNaN(num) && num > 0 && num === parseFloat(t) && num < 100_000) hits++
  }
  return hits / values.length
}

// ─── Column aliases ─────────────────────────────────────────

const DEFAULT_COLUMN_ALIASES: ColumnMapping[] = [
  {
    field: 'sku',
    aliases: [
      'sku', 'part number', 'part no', 'part#', 'part #', 'partno', 'part',
      'partnumber', 'pn', 'p/n', 'mpn', 'mfr part', 'mfr part number',
      'manufacturer part number', 'manufacturer part no',
      'item code', 'itemcode', 'item no', 'item#', 'item #', 'item number',
      'product code', 'product no', 'product#', 'product number',
      'catalog number', 'catalog no', 'catalogue number', 'cat no', 'cat#',
      'code', 'model', 'model number', 'model no', 'model#',
      'article', 'article number', 'article no', 'art no', 'artno', 'art#',
      'ref', 'reference', 'ref no', 'reference number', 'ref#',
      'stock code', 'stock no', 'material number', 'material no', 'mat no',
      'material code', 'mat code', 'order code', 'order no',
      'vendor sku', 'vendor part', 'vendor part number',
      'supplier part', 'supplier part number', 'supplier code',
      'product id', 'item id',
      'partcode', 'part code',
      'product name',
      'رقم القطعة', 'رمز المنتج', 'كود', 'رقم الصنف',
    ],
    keywords: [
      'sku', 'part', 'mpn', 'p/n', 'partcode', 'catalog', 'catalogue',
      'article', 'model', 'stock code', 'material', 'mat no', 'item code',
      'item no', 'order code', 'ref no', 'product code',
    ],
    antiKeywords: ['description', 'desc', 'الوصف'],
    contentDetectable: true,
    contentPattern: skuPattern,
  },
  {
    field: 'name',
    aliases: [
      'name', 'description', 'desc', 'item description', 'item name',
      'product description', 'product desc', 'product title',
      'productname', 'productdescription',
      'title', 'label', 'material description', 'material desc',
      'material name', 'item', 'product', 'goods description',
      'full description', 'short description', 'long description',
      'specification', 'spec', 'details', 'item details',
      'الوصف', 'اسم المنتج', 'البيان', 'الصنف',
    ],
    keywords: [
      'desc', 'description', 'name', 'title', 'specification',
      'label', 'details', 'الوصف', 'البيان',
    ],
    antiKeywords: [],
    contentDetectable: true,
    contentPattern: namePattern,
  },
  {
    field: 'manufacturer',
    aliases: [
      'manufacturer', 'mfr', 'mfg', 'brand', 'make', 'vendor', 'oem',
      'manufacturer name', 'brand name', 'vendor name', 'supplier',
      'supplier name', 'company', 'producer', 'origin',
      'الشركة المصنعة', 'العلامة التجارية', 'المصنع',
    ],
    keywords: [
      'manufacturer', 'mfr', 'mfg', 'brand', 'make', 'oem',
      'المصنع',
    ],
  },
  {
    field: 'qty',
    aliases: [
      'qty', 'quantity', 'qnty', 'q-ty', 'amount', 'units', 'count',
      'pcs', 'pieces', 'ea', 'nos', 'no of units', 'number of units',
      'order qty', 'order quantity', 'required qty', 'required quantity',
      'requested qty', 'demand', 'demand qty', 'need', 'total qty',
      'الكمية', 'عدد',
    ],
    keywords: [
      'qty', 'quantity', 'units', 'pieces', 'pcs', 'count',
      'demand', 'الكمية',
    ],
    antiKeywords: ['price', 'cost', 'total'],
    contentDetectable: true,
    contentPattern: qtyPattern,
  },
  {
    field: 'unitCost',
    aliases: [
      'unit cost', 'unit price', 'unitcost', 'unitprice',
      'cost', 'price', 'rate', 'value',
      'net price', 'sell price', 'buy price', 'buying price',
      'selling price', 'list price', 'trade price',
      'each', 'per unit', 'price each', 'cost each',
      'up', 'u/p', 'u.p', 'u.p.',
      'fob price', 'cif price', 'offer price', 'quoted price',
      'bid price', 'tender price',
      'msrp', 'msrp unit purchase price',
      'distributor unit purchase price', 'distributor price',
      'discounted unit price', 'disc unit price', 'discounted price',
      'dealer cost', 'dealer price',
      'سعر الوحدة', 'السعر', 'التكلفة',
    ],
    keywords: [
      'price', 'cost', 'rate', 'السعر', 'التكلفة',
      'msrp', 'dealer', 'discounted',
    ],
    antiKeywords: ['total', 'extended', 'subtotal', 'الإجمالي'],
    contentDetectable: true,
    contentPattern: pricePattern,
  },
  {
    field: 'totalCost',
    aliases: [
      'total', 'total cost', 'total price', 'total amount',
      'line total', 'extended price', 'ext price', 'subtotal',
      'net amount', 'net total', 'gross total',
      'total discounted purchase price', 'total purchase price',
      'الإجمالي', 'المبلغ',
    ],
    keywords: ['total', 'extended', 'subtotal', 'الإجمالي'],
  },
  {
    field: 'discount',
    aliases: [
      'discount', 'disc', 'disc%', 'discount %', 'discount percent',
      'additional discount', 'volume discount',
      'discount category', 'خصم',
    ],
    keywords: ['discount', 'disc', 'خصم'],
  },
  {
    field: 'leadTimeDays',
    aliases: [
      'lead time', 'leadtime', 'lead time days', 'lead days',
      'delivery days', 'delivery time', 'delivery',
      'tat', 'turnaround', 'turnaround time',
      'days', 'working days', 'business days',
      'eta', 'estimated delivery', 'duration',
      'مدة التسليم', 'وقت التسليم',
    ],
    keywords: [
      'lead', 'delivery', 'tat', 'turnaround', 'eta', 'duration', 'التسليم',
    ],
  },
  {
    field: 'moq',
    aliases: [
      'moq', 'min qty', 'min quantity', 'minimum order',
      'minimum quantity', 'minimum order qty', 'min order',
      'min order qty', 'minimum', 'minimum order quantity',
      'الحد الأدنى للطلب',
    ],
    keywords: ['moq', 'minimum', 'min order', 'min qty'],
  },
  {
    field: 'currency',
    aliases: [
      'currency', 'curr', 'ccy', 'cur', 'currency code', 'العملة',
    ],
    keywords: ['currency', 'curr', 'ccy', 'العملة'],
  },
  {
    field: 'uom',
    aliases: [
      'uom', 'unit', 'unit of measure', 'measure', 'unit of measurement',
      'packaging', 'pack size', 'pack', 'وحدة القياس',
    ],
    keywords: ['uom', 'measure', 'packaging'],
    antiKeywords: ['price', 'cost'],
  },
  {
    field: 'associatedProducts',
    aliases: [
      'associated products', 'related products', 'compatible',
      'accessories', 'works with', 'compatible with',
    ],
    keywords: ['associated', 'related', 'compatible', 'accessories'],
  },
]

// ─── Smart column detection ────────────────────────────────

function normalizeHeader(h: string): string {
  return h
    .toLowerCase()
    .replace(/[_\-./\\#()\[\]{}:]/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/\s+/g, ' ')
    .trim()
}

function tokenize(s: string): string[] {
  return s.toLowerCase().replace(/[^a-z0-9\u0600-\u06FF ]/g, ' ').split(/\s+/).filter(t => t.length > 0)
}

function tokenOverlap(a: string[], b: string[]): number {
  if (a.length === 0 || b.length === 0) return 0
  const setB = new Set(b)
  let matches = 0
  for (const t of a) { if (setB.has(t)) matches++ }
  return matches / Math.max(a.length, b.length)
}

interface ColumnScore {
  colIdx: number
  score: number
  method: string
}

function detectColumns(
  headers: string[],
  aliases: ColumnMapping[],
  rows?: ParsedRow[],
): { mapping: Record<string, number>; methods: Record<string, string> } {
  const normalized = headers.map(normalizeHeader)
  const candidates: Record<string, ColumnScore[]> = {}

  for (const mapping of aliases) {
    const scores: ColumnScore[] = []

    for (let i = 0; i < normalized.length; i++) {
      const h = normalized[i] ?? ''
      if (!h) continue

      // Check anti-keywords first — skip if header strongly belongs to another field
      if (mapping.antiKeywords?.some(ak => h.includes(ak.toLowerCase()))) continue

      let bestScore = 0
      let method = ''

      // Tier 1: Exact match (score 1.0)
      if (mapping.aliases.some(a => normalizeHeader(a) === h)) {
        bestScore = 1.0
        method = 'exact'
      }

      // Tier 2: Header contains alias or alias contains header (score 0.8)
      if (bestScore < 0.8) {
        for (const alias of mapping.aliases) {
          const na = normalizeHeader(alias)
          if (na.length > 2 && h.length > 2 && (h.includes(na) || na.includes(h))) {
            bestScore = 0.8
            method = 'contains'
            break
          }
        }
      }

      // Tier 3: Keyword in header words (score 0.6)
      if (bestScore < 0.6 && mapping.keywords.length > 0) {
        const hWords = h.split(' ')
        for (const kw of mapping.keywords) {
          const kwLower = kw.toLowerCase()
          if (hWords.some(w => w === kwLower) || h.includes(kwLower)) {
            bestScore = 0.6
            method = 'keyword'
            break
          }
        }
      }

      // Tier 4: Token overlap (score up to 0.5)
      if (bestScore < 0.4) {
        const hTokens = tokenize(h)
        for (const alias of mapping.aliases) {
          const aTokens = tokenize(alias)
          const overlap = tokenOverlap(hTokens, aTokens)
          if (overlap > 0.4) {
            const s = Math.min(0.5, overlap * 0.6)
            if (s > bestScore) { bestScore = s; method = 'fuzzy' }
          }
        }
      }

      if (bestScore > 0) scores.push({ colIdx: i, score: bestScore, method })
    }

    candidates[mapping.field] = scores
  }

  // Greedy assignment: highest confidence first, no double assignment
  const allEntries: { field: string; colIdx: number; score: number; method: string }[] = []
  for (const [field, scores] of Object.entries(candidates)) {
    for (const s of scores) {
      allEntries.push({ field, colIdx: s.colIdx, score: s.score, method: s.method })
    }
  }
  allEntries.sort((a, b) => b.score - a.score)

  const assignedCols = new Set<number>()
  const assignedFields = new Set<string>()
  const resultMapping: Record<string, number> = {}
  const methods: Record<string, string> = {}

  for (const entry of allEntries) {
    if (assignedFields.has(entry.field) || assignedCols.has(entry.colIdx)) continue
    resultMapping[entry.field] = entry.colIdx
    methods[entry.field] = entry.method
    assignedFields.add(entry.field)
    assignedCols.add(entry.colIdx)
  }

  // Tier 5: Content-based detection for unmapped critical fields
  if (rows && rows.length > 0) {
    const sampleSize = Math.min(rows.length, 20)
    const sampleRows = rows.slice(0, sampleSize)

    for (const def of aliases) {
      if (assignedFields.has(def.field) || !def.contentDetectable || !def.contentPattern) continue

      let bestCol = -1
      let bestConfidence = 0.5

      for (let i = 0; i < headers.length; i++) {
        if (assignedCols.has(i)) continue
        const hdr = headers[i]
        if (!hdr) continue
        const values = sampleRows.map(r => r[hdr] || '').filter(v => v.trim())
        if (values.length < 2) continue
        const confidence = def.contentPattern(values)
        if (confidence > bestConfidence) { bestConfidence = confidence; bestCol = i }
      }

      if (bestCol >= 0) {
        resultMapping[def.field] = bestCol
        methods[def.field] = 'content'
        assignedFields.add(def.field)
        assignedCols.add(bestCol)
      }
    }
  }

  return { mapping: resultMapping, methods }
}

// ─── Row filtering — remove subtotals, section headers, junk ─

const SUBTOTAL_PATTERNS = /^(subtotal|sub total|total|grand total|net total|gross|amount due|المجموع|الإجمالي|charges total)/i
const SECTION_NOISE = /^(page \d|prices exclude|errors and omissions|terms and conditions|all orders are|thank you|currency:|dealer|lapsed coverage|amnesty)/i

function isJunkRow(row: ParsedRow, headers: string[], mappedColumns: Record<string, number>): boolean {
  const values = headers.map(h => (row[h] || '').trim())
  const nonEmpty = values.filter(v => v.length > 0)

  if (nonEmpty.length === 0) return true

  if (nonEmpty.length === 1) {
    const v = nonEmpty[0] || ''
    if (SUBTOTAL_PATTERNS.test(v) || SECTION_NOISE.test(v)) return true
    if (v.length > 3 && !/\d/.test(v) && !v.includes('$') && !v.includes('£')) return true
  }

  for (const v of nonEmpty) {
    if (SUBTOTAL_PATTERNS.test(v)) return true
    if (SECTION_NOISE.test(v)) return true
  }

  const skuIdx = mappedColumns['sku']
  const nameIdx = mappedColumns['name']
  const skuHeader = skuIdx !== undefined ? headers[skuIdx] : undefined
  const nameHeader = nameIdx !== undefined ? headers[nameIdx] : undefined
  const skuVal = skuHeader ? (row[skuHeader] || '').trim() : ''
  const nameVal = nameHeader ? (row[nameHeader] || '').trim() : ''

  if (skuHeader && nameHeader && !skuVal && !nameVal) {
    const priceIdx = mappedColumns['unitCost']
    const qtyIdx = mappedColumns['qty']
    const priceHeader = priceIdx !== undefined ? headers[priceIdx] : undefined
    const qtyHeader = qtyIdx !== undefined ? headers[qtyIdx] : undefined
    const hasPrice = priceHeader ? parseFloat((row[priceHeader] || '').replace(/[^0-9.-]/g, '')) : 0
    const hasQty = qtyHeader ? parseInt((row[qtyHeader] || '').replace(/[^0-9]/g, ''), 10) : 0
    if (!hasPrice && !hasQty) return true
  }

  return false
}

// ─── Data cleaning helpers ──────────────────────────────────

/** Strip currency symbols and trailing currency codes from a value */
function cleanCurrencyValue(v: string): string {
  return v
    .replace(/[$€£¥]/g, '')
    .replace(/\s*(SAR|AED|USD|GBP|EUR|QAR|KWD|BHD|OMR)\s*/gi, '')
    .replace(/,/g, '')
    .replace(/\s+/g, '')
    .replace(/^\((.+)\)$/, '-$1')  // Handle (5,000.00) as negative
    .trim()
}

// ─── CSV / TSV parsing ──────────────────────────────────────

function splitCSVLine(line: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (ch === '"') { inQuotes = !inQuotes; continue }
    if (!inQuotes && (ch === ',' || ch === '\t' || ch === ';')) {
      result.push(current.trim())
      current = ''
      continue
    }
    current += ch
  }
  result.push(current.trim())
  return result
}

export function parseCSVText(
  text: string,
  columnAliases: ColumnMapping[] = DEFAULT_COLUMN_ALIASES,
): ParseResult {
  const lines = text.split(/\r?\n/).filter(l => l.trim())
  if (lines.length < 1) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  const firstLine = lines[0] ?? ''
  const headerCells = splitCSVLine(firstLine)
  const headers = headerCells.map(h => h.replace(/^["']|["']$/g, '').trim())

  const rows: ParsedRow[] = []
  for (let i = 1; i < lines.length; i++) {
    const line = lines[i]
    if (!line) continue
    const cells = splitCSVLine(line)
    if (cells.length < 2 || cells.every(c => !c.trim())) continue
    const row: ParsedRow = {}
    headers.forEach((h, idx) => {
      const cell = idx < cells.length ? cells[idx] : undefined
      if (cell !== undefined) row[h] = cell.replace(/^["']|["']$/g, '').trim()
    })
    rows.push(row)
  }

  const { mapping, methods } = detectColumns(headers, columnAliases, rows)

  const filtered = rows.filter(r => !isJunkRow(r, headers, mapping))

  return { headers, rows: filtered, mappedColumns: mapping, rawRowCount: filtered.length, detectionMethod: methods }
}

// ─── Excel parsing ──────────────────────────────────────────

function findHeaderRow(data: (string | number)[][], aliases: ColumnMapping[]): number {
  let bestIdx = 0
  let bestScore = 0

  for (let i = 0; i < Math.min(data.length, 20); i++) {
    const row = data[i]
    if (!row) continue
    const cells = row.map(c => String(c).trim()).filter(c => c.length > 0)
    if (cells.length < 2) continue

    let score = 0
    const normalizedCells = cells.map(normalizeHeader)

    for (const def of aliases) {
      for (const cell of normalizedCells) {
        // Exact alias match
        if (def.aliases.some(a => normalizeHeader(a) === cell)) { score += 3; break }
        // Contains alias
        if (def.aliases.some(a => { const na = normalizeHeader(a); return na.length > 2 && (cell.includes(na) || na.includes(cell)) })) { score += 2; break }
        // Keyword hit
        if (def.keywords.some(k => cell.includes(k.toLowerCase()))) { score += 1; break }
      }
    }

    // Bonus for having more non-empty cells (wider tables are more likely to be headers)
    score += Math.min(cells.length * 0.2, 2)

    if (score > bestScore) { bestScore = score; bestIdx = i }
  }

  return bestIdx
}

export async function parseExcelFile(
  file: File,
  columnAliases: ColumnMapping[] = DEFAULT_COLUMN_ALIASES,
): Promise<ParseResult> {
  const buffer = await file.arrayBuffer()
  const workbook = XLSX.read(buffer, { type: 'array' })
  const sheetName = workbook.SheetNames[0]
  if (!sheetName) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  const sheet = workbook.Sheets[sheetName]
  if (!sheet) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }
  const jsonData: (string | number)[][] = XLSX.utils.sheet_to_json(sheet, { header: 1, defval: '' })
  if (jsonData.length < 1) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  const headerIdx = findHeaderRow(jsonData, columnAliases)
  const hRow = jsonData[headerIdx]
  if (!hRow) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }
  const headerRow = hRow.map(h => String(h).trim())
  const headers = headerRow.filter(h => h.length > 0)

  const rows: ParsedRow[] = []
  for (let i = headerIdx + 1; i < jsonData.length; i++) {
    const cells = jsonData[i]
    if (!cells || cells.every(c => !String(c).trim())) continue
    const row: ParsedRow = {}
    headers.forEach((h, idx) => {
      row[h] = idx < cells.length ? String(cells[idx]).trim() : ''
    })
    rows.push(row)
  }

  const { mapping, methods } = detectColumns(headers, columnAliases, rows)
  const filtered = rows.filter(r => !isJunkRow(r, headers, mapping))

  return { headers, rows: filtered, mappedColumns: mapping, rawRowCount: filtered.length, detectionMethod: methods }
}

// ─── PDF parsing ────────────────────────────────────────────

interface TextItem {
  str: string
  x: number
  y: number
  width: number
  page: number
}

function findPDFHeaderRow(
  rowGroups: TextItem[][],
  aliases: ColumnMapping[],
): number {
  let bestIdx = 0
  let bestScore = 0

  for (let i = 0; i < Math.min(rowGroups.length, 15); i++) {
    const group = rowGroups[i]
    if (!group || group.length < 2) continue
    const texts = group.map(g => normalizeHeader(g.str))
    let score = 0

    for (const def of aliases) {
      for (const text of texts) {
        if (def.aliases.some(a => normalizeHeader(a) === text)) { score += 3; break }
        if (def.aliases.some(a => { const na = normalizeHeader(a); return na.length > 2 && (text.includes(na) || na.includes(text)) })) { score += 2; break }
        if (def.keywords.some(k => text.includes(k.toLowerCase()))) { score += 1; break }
      }
    }

    score += Math.min(group.length * 0.2, 2)
    if (score > bestScore) { bestScore = score; bestIdx = i }
  }

  return bestIdx
}

export async function parsePDFFile(
  file: File,
  columnAliases: ColumnMapping[] = DEFAULT_COLUMN_ALIASES,
): Promise<ParseResult> {
  const buffer = await file.arrayBuffer()
  const pdf = await getDocument({ data: new Uint8Array(buffer) }).promise
  const allItems: TextItem[] = []

  for (let p = 1; p <= pdf.numPages; p++) {
    const page = await pdf.getPage(p)
    const textContent = await page.getTextContent()
    for (const item of textContent.items) {
      if (!('str' in item) || !item.str.trim()) continue
      const tx = (item as any).transform || [1, 0, 0, 1, 0, 0]
      allItems.push({
        str: item.str.trim(),
        x: tx[4],
        y: p * 100000 - tx[5], // Flatten pages into one coordinate space (top-to-bottom)
        width: (item as any).width || item.str.length * 5,
        page: p,
      })
    }
  }

  if (allItems.length === 0) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  // Group by y-coordinate
  const yThreshold = 3
  allItems.sort((a, b) => a.y - b.y || a.x - b.x)

  const rowGroups: TextItem[][] = []
  const firstItem = allItems[0]
  if (!firstItem) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }
  let currentGroup: TextItem[] = [firstItem]
  let currentY = firstItem.y

  for (let i = 1; i < allItems.length; i++) {
    const item = allItems[i]
    if (!item) continue
    if (Math.abs(item.y - currentY) <= yThreshold) {
      currentGroup.push(item)
    } else {
      rowGroups.push(currentGroup.sort((a, b) => a.x - b.x))
      currentGroup = [item]
      currentY = item.y
    }
  }
  rowGroups.push(currentGroup.sort((a, b) => a.x - b.x))

  if (rowGroups.length < 2) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  // Find the best header row
  const headerRowIdx = findPDFHeaderRow(rowGroups, columnAliases)
  const headerGroup = rowGroups[headerRowIdx]
  if (!headerGroup) return { headers: [], rows: [], mappedColumns: {}, rawRowCount: 0, detectionMethod: {} }

  // Detect page header pattern (repeated text blocks across pages) to filter them
  const pageHeaderTexts = new Set<string>()
  if (pdf.numPages > 1) {
    const page1Rows = rowGroups.filter(g => g[0]?.page === 1).slice(0, 5)
    const page2Rows = rowGroups.filter(g => g[0]?.page === 2).slice(0, 5)
    for (const r1 of page1Rows) {
      const t1 = r1.map(i => i.str).join(' ')
      for (const r2 of page2Rows) {
        const t2 = r2.map(i => i.str).join(' ')
        if (t1 === t2 && t1.length > 5) pageHeaderTexts.add(t1)
      }
    }
  }

  const colBoundaries = headerGroup.map(item => ({
    x: item.x,
    width: item.width,
    header: item.str,
  }))

  const headers = colBoundaries.map(cb => cb.header)

  function findColumnIndex(x: number): number {
    let bestIdx = 0
    let bestDist = Infinity
    for (let i = 0; i < colBoundaries.length; i++) {
      const cb = colBoundaries[i]
      if (!cb) continue
      const dist = Math.abs(x - cb.x)
      if (dist < bestDist) { bestDist = dist; bestIdx = i }
    }
    return bestIdx
  }

  const rows: ParsedRow[] = []
  for (let i = headerRowIdx + 1; i < rowGroups.length; i++) {
    const group = rowGroups[i]
    if (!group) continue

    const rowText = group.map(g => g.str).join(' ')
    if (pageHeaderTexts.has(rowText)) continue
    if (SECTION_NOISE.test(rowText)) continue

    const row: ParsedRow = {}
    headers.forEach(h => row[h] = '')

    for (const item of group) {
      const colIdx = findColumnIndex(item.x)
      const header = headers[colIdx]
      if (header) {
        row[header] = row[header] ? `${row[header]} ${item.str}` : item.str
      }
    }

    if (Object.values(row).some(v => v.trim())) rows.push(row)
  }

  const { mapping, methods } = detectColumns(headers, columnAliases, rows)
  const filtered = rows.filter(r => !isJunkRow(r, headers, mapping))

  return { headers, rows: filtered, mappedColumns: mapping, rawRowCount: filtered.length, detectionMethod: methods }
}

// ─── Unified parse function ─────────────────────────────────

export async function parseFile(
  file: File,
  columnAliases?: ColumnMapping[],
): Promise<{ result: ParseResult; fileType: FileType } | { error: string }> {
  const fileType = detectFileType(file.name)
  if (!fileType) return { error: `Unsupported file type: ${file.name.split('.').pop()}` }

  try {
    let result: ParseResult

    switch (fileType) {
      case 'csv': {
        const text = await file.text()
        result = parseCSVText(text, columnAliases)
        break
      }
      case 'excel':
        result = await parseExcelFile(file, columnAliases)
        break
      case 'pdf':
        result = await parsePDFFile(file, columnAliases)
        break
    }

    if (result.rows.length === 0) {
      return { error: 'No data rows could be parsed from the file.' }
    }

    return { result, fileType }
  } catch (err) {
    return { error: `Failed to parse file: ${err instanceof Error ? err.message : 'Unknown error'}` }
  }
}

// ─── Value extraction helpers ───────────────────────────────

export function getMappedValue(
  row: ParsedRow,
  field: string,
  headers: string[],
  mappedColumns: Record<string, number>,
  fallback = '',
): string {
  const colIdx = mappedColumns[field]
  if (colIdx === undefined) return fallback
  const header = headers[colIdx]
  if (!header) return fallback
  return row[header] ?? fallback
}

export function getMappedNumber(
  row: ParsedRow,
  field: string,
  headers: string[],
  mappedColumns: Record<string, number>,
  fallback = 0,
): number {
  const val = getMappedValue(row, field, headers, mappedColumns)
  const cleaned = cleanCurrencyValue(val)
  const num = parseFloat(cleaned)
  return isNaN(num) ? fallback : num
}

export function getMappedInt(
  row: ParsedRow,
  field: string,
  headers: string[],
  mappedColumns: Record<string, number>,
  fallback = 0,
): number {
  const val = getMappedValue(row, field, headers, mappedColumns)
  const cleaned = val.replace(/[,\s]/g, '')
  const num = parseInt(cleaned.replace(/[^0-9-]/g, ''), 10)
  return isNaN(num) ? fallback : num
}

// ─── Downloadable Excel templates ───────────────────────────

export type TemplateType = 'supplier-quote' | 'purchase-order' | 'product-catalog'

const TEMPLATES: Record<TemplateType, { headers: string[]; sampleRow: (string | number)[] }> = {
  'supplier-quote': {
    headers: ['Part Number', 'Description', 'Manufacturer', 'Qty', 'Unit Price', 'Lead Time (Days)', 'MOQ', 'Currency'],
    sampleRow: ['CAM-4MP-001', '4MP IP Dome Camera', 'Hikvision', 10, 245.00, 30, 1, 'USD'],
  },
  'purchase-order': {
    headers: ['Part Number', 'Description', 'Manufacturer', 'Qty', 'Unit Price', 'Lead Time (Days)'],
    sampleRow: ['CAM-4MP-001', '4MP IP Dome Camera', 'Hikvision', 10, 245.00, 30],
  },
  'product-catalog': {
    headers: ['Part Number', 'Description', 'Unit Price', 'MOQ', 'Lead Time (Days)', 'Currency'],
    sampleRow: ['CAM-4MP-001', '4MP IP Dome Camera', 245.00, 1, 21, 'USD'],
  },
}

export function downloadTemplate(type: TemplateType, fileName?: string) {
  const tmpl = TEMPLATES[type]
  const wb = XLSX.utils.book_new()
  const data = [tmpl.headers, tmpl.sampleRow]
  const ws = XLSX.utils.aoa_to_sheet(data)

  // Auto-size columns
  ws['!cols'] = tmpl.headers.map((h, i) => {
    const sampleLen = String(tmpl.sampleRow[i] ?? '').length
    return { wch: Math.max(h.length, sampleLen) + 4 }
  })

  XLSX.utils.book_append_sheet(wb, ws, 'Template')
  const name = fileName || `${type}-template.xlsx`
  XLSX.writeFile(wb, name)
}
