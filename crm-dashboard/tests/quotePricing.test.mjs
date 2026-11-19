import test from 'node:test'
import assert from 'node:assert/strict'
import { priceForMargin, inPriceBook } from '../src/services/quotePricing.ts'
test('margin changes selling price while preserving cost, including line discounts', () => {
 const cost = 80
 assert.equal(priceForMargin(cost, 20), 100)
 assert.equal(priceForMargin(cost, 20, 10), 111.11)
 assert.equal(priceForMargin(cost, 50), 160)
 assert.equal(cost, 80)
 assert.equal(priceForMargin(cost, -25), 64)
 for (const args of [[80,100],[80,20,100],[0,20],[80,NaN]]) assert.throws(() => priceForMargin(...args))
})
test('selected price books restrict each source independently and standard pricing remains unrestricted', () => {
 const entries = new Set(['product:camera', 'service:install'])
 assert.equal(inPriceBook('book', entries, 'product','camera'), true)
 assert.equal(inPriceBook('book', entries, 'product','other'), false)
 assert.equal(inPriceBook('book', entries, 'service','camera'), false)
 assert.equal(inPriceBook('book', entries, 'service','install'), true)
 assert.equal(inPriceBook('', entries, 'product','other'), true)
 assert.equal(inPriceBook('empty', new Set(), 'product','camera'), false)
})

test('accepted sale prices preserve net pricing across currencies and both discounts', async () => {
 const { priceFromSale } = await import('../src/services/quotePricing.ts')
 assert.equal(priceFromSale(171,1,1,10,5),200)
 assert.equal(priceFromSale(100,3.75,1),375)
 assert.equal(priceFromSale(375,1,3.75),100)
 assert.equal(priceFromSale(0,1,1),0)
 for (const args of [[100,0,1],[100,1,0],[100,1,1,100],[100,1,1,0,100],[-1,1,1],[NaN,1,1]]) assert.throws(() => priceFromSale(...args))
})
