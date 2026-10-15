import test from 'node:test'
import assert from 'node:assert/strict'
import { findMissingMasters, matchManufacturers, suggestCode } from '../src/services/catalogMasters.ts'

const catalog = [
  { id: 'hik', name: 'Hikvision', code: 'HIK', categories: [{ id: 'c1', name: 'IP Cameras' }] },
  { id: 'axis', name: 'Axis Communications', code: 'AXIS', categories: [] },
]

test('manufacturers match by name or code, ignoring case and spacing', () => {
  assert.equal(matchManufacturers(catalog, '  hikvision ')[0]?.id, 'hik')
  assert.equal(matchManufacturers(catalog, 'axis')[0]?.id, 'axis')
  assert.equal(matchManufacturers(catalog, 'Axis  Communications')[0]?.id, 'axis')
  assert.equal(matchManufacturers(catalog, 'Dahua').length, 0)
})

test('suggested codes are short, upper case and unique', () => {
  assert.equal(suggestCode('Bosch Security Systems', ['HIK']), 'BOSCHSECUR')
  assert.equal(suggestCode('Hik', ['HIK']), 'HIK-2')
  assert.equal(suggestCode('Hik', ['HIK', 'hik-2']), 'HIK-3')
  assert.equal(suggestCode('***', []), 'MFR')
})

test('new manufacturers and categories are listed once; known ones are not', () => {
  const rows = [
    { manufacturerName: 'Hikvision', categoryName: 'ip cameras' },
    { manufacturerName: 'HIK', categoryName: 'Thermal' },
    { manufacturerName: 'Dahua', categoryName: 'Recorders' },
    { manufacturerName: 'dahua', categoryName: 'recorders' },
    { manufacturerName: '', categoryName: 'Speakers' },
    { manufacturerName: '', categoryName: '' },
  ]
  const missing = findMissingMasters(rows, catalog, 'axis')
  assert.deepEqual(missing.manufacturers, ['Dahua'])
  assert.deepEqual(missing.categories, [
    { manufacturer: 'Hikvision', category: 'Thermal' },
    { manufacturer: 'Dahua', category: 'Recorders' },
    { manufacturer: 'Axis Communications', category: 'Speakers' },
  ])
  assert.deepEqual(missing.ambiguous, [])
})

test('a name shared by two manufacturers is reported as ambiguous, not created', () => {
  const twins = [...catalog, { id: 'h2', name: 'Hikvision', code: 'HIK2', categories: [] }]
  const missing = findMissingMasters([{ manufacturerName: 'Hikvision', categoryName: 'X' }], twins)
  assert.deepEqual(missing, { manufacturers: [], categories: [], ambiguous: ['Hikvision'] })
})
