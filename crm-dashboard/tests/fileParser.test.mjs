import test from 'node:test'
import assert from 'node:assert/strict'
import * as XLSX from 'xlsx'
import { parseFile, parseCSVText, getMappedValue, getMappedNumber } from '../src/utils/fileParser.ts'

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
