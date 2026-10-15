import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import * as XLSX from 'xlsx'
import { parseFile, parseExcelFile, parseCSVText, getMappedValue, getMappedNumber } from '../src/utils/fileParser.ts'

test('generated import template imports only Products and preserves identifiers and mixed row metadata', async () => {
 // Produced by GET /products/import-template; refresh it if the template layout changes.
 const bytes = await readFile(new URL('./fixtures/product-import-template.xlsx', import.meta.url))
 const empty = await parseExcelFile(new File([bytes], 'g4s-product-import.xlsx'))
 assert.equal(empty.rows.length, 0, 'instructions and drop-down lists must never become catalog rows')
 const workbook = XLSX.read(bytes, { type: 'buffer' })
 assert.deepEqual(workbook.SheetNames, ['Products', 'Instructions', 'Lists'])
 // Fill a disposable in-memory copy of the exact delivered workbook, including a leading-zero SKU.
 XLSX.utils.sheet_add_aoa(workbook.Sheets.Products, [
  ['000123', 'Dome camera', 100, 'Actual Supplier', 'Camera Maker', 'Cameras', 'USD', 3.75, 21],
  ['LOCAL-001', 'Recorder', 500, 'Local Supplier', 'Recorder Maker', 'Recorders', 'SAR', 1, 7],
 ], { origin: 'A2' })
 workbook.SheetNames.reverse() // Products must still win if the user opens/reorders Instructions first.
 const loaded = await parseFile(new File([XLSX.write(workbook, { type: 'array', bookType: 'xlsx' })], 'filled.xlsx'))
 assert.ok('result' in loaded)
 const { rows, headers, mappedColumns } = loaded.result
 assert.equal(rows.length, 2)
 for (const [field, expected] of Object.entries({sku:'000123',name:'Dome camera',vendorName:'Actual Supplier',manufacturer:'Camera Maker',category:'Cameras',currency:'USD'})) {
  assert.equal(getMappedValue(rows[0], field, headers, mappedColumns), expected, field)
 }
 assert.equal(getMappedNumber(rows[0], 'fxRate', headers, mappedColumns), 3.75)
 assert.equal(getMappedNumber(rows[0], 'leadTimeDays', headers, mappedColumns), 21)
 assert.equal(getMappedNumber(rows[0], 'unitCost', headers, mappedColumns), 100)
 assert.equal(getMappedValue(rows[1], 'currency', headers, mappedColumns), 'SAR')
})

for (const bookType of ['xlsx', 'xls']) {
 test(`vendor ${bookType} import preserves values across blank spacer columns`, async () => {
  const workbook = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(workbook, XLSX.utils.aoa_to_sheet([
   ['Part Number', '', 'Description', 'Qty', 'Unit Price'],
   ['CAM-001', '', 'Security camera', 2, 125.5],
  ]), 'Catalog')
  const bytes = XLSX.write(workbook, { type: 'array', bookType })
  const parsed = await parseFile(new File([bytes], `vendor.${bookType}`))
  assert.ok('result' in parsed, JSON.stringify(parsed))
  const { headers, rows, mappedColumns } = parsed.result
  assert.equal(rows.length, 1)
  assert.equal(getMappedValue(rows[0], 'sku', headers, mappedColumns), 'CAM-001')
  assert.equal(getMappedValue(rows[0], 'name', headers, mappedColumns), 'Security camera')
  assert.equal(getMappedNumber(rows[0], 'qty', headers, mappedColumns), 2)
  assert.equal(getMappedNumber(rows[0], 'unitCost', headers, mappedColumns), 125.5)
 })
}
test('CSV quoted descriptions retain commas and monetary values', () => {
 const parsed = parseCSVText('Part Number,Description,Qty,Unit Price\nCAM-001,"Camera, indoor",3,125.50')
 assert.equal(parsed.rows[0].Description, 'Camera, indoor')
 assert.equal(getMappedNumber(parsed.rows[0], 'unitCost', parsed.headers, parsed.mappedColumns), 125.5)
})
test('unsupported files produce an actionable error', async () => {
 const result = await parseFile(new File(['hello'], 'vendor.docx'))
 assert.deepEqual(result, { error: 'Unsupported file type: docx' })
})

test('a price list keeps its products and drops its category banners', async () => {
 // Shaped like the Ingram Micro Dell price list: a leading spacer column, a title
 // row above the header, and category banners between groups of products.
 const workbook = XLSX.utils.book_new()
 XLSX.utils.book_append_sheet(workbook, XLSX.utils.aoa_to_sheet([
  ['', 'Dell Pricelist For the Month', 'March 2026', ''],
  ['', 'Part No.', 'Description', 'Price', 'Availability'],
  ['', 'Windows Server Operating Systems - ROK', '', '', ''],
  ['', '634-CVFM', 'Dell Windows Server 2025, ROK, 16CORE', 3600, 'Ex-Stock'],
  ['', 'Dell PowerEdge Tower Servers :16G', '', '', ''],
  ['', 'EMEA_PET360SPL1', 'PowerEdge T360 Server', 7850, 'Limited Stock'],
 ]), 'Servers')
 const parsed = await parseFile(new File([XLSX.write(workbook, { type: 'array', bookType: 'xlsx' })], 'pricelist.xlsx'))
 assert.ok('result' in parsed, JSON.stringify(parsed))
 const { headers, rows, mappedColumns } = parsed.result
 assert.equal(rows.length, 2, 'category banners must not become products')
 assert.deepEqual(rows.map(r => getMappedValue(r, 'sku', headers, mappedColumns)), ['634-CVFM', 'EMEA_PET360SPL1'])
 assert.equal(getMappedNumber(rows[1], 'unitCost', headers, mappedColumns), 7850)
})

test('the payable price wins over the list price when a sheet has both', async () => {
 // Shaped like the ZGT card-printing quote, which prices each line three times.
 const workbook = XLSX.utils.book_new()
 XLSX.utils.book_append_sheet(workbook, XLSX.utils.aoa_to_sheet([
  ['Qty', 'Part Number', 'Description', 'MSRP Unit Purchase Price', 'Distributor Unit Purchase Price', 'Total Discounted Purchase Price'],
  [500, 'NSI-PROD', 'HID Proximity cards', 8, 5.6, 2800],
 ]), 'QUOTE')
 const parsed = await parseFile(new File([XLSX.write(workbook, { type: 'array', bookType: 'xlsx' })], 'quote.xlsx'))
 assert.ok('result' in parsed, JSON.stringify(parsed))
 const { headers, rows, mappedColumns } = parsed.result
 assert.equal(getMappedNumber(rows[0], 'unitCost', headers, mappedColumns), 5.6)
 assert.equal(getMappedNumber(rows[0], 'qty', headers, mappedColumns), 500)
})
