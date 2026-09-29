import test from 'node:test'
import assert from 'node:assert/strict'
import {KSA_CITIES, KSA_REGIONS, regionForCity} from '../src/utils/ksa.ts'
test('every city maps to a listed region and lookups ignore case and spacing', () => {
 for (const region of Object.values(KSA_CITIES)) assert.ok(KSA_REGIONS.includes(region), region)
 assert.equal(regionForCity('  jeddah '), 'Makkah')
 assert.equal(regionForCity('Dammam'), 'Eastern Province')
 assert.equal(regionForCity('Atlantis'), undefined)
})
