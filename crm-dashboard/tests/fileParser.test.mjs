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
