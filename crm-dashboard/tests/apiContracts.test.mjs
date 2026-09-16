import test from 'node:test'
import assert from 'node:assert/strict'
import { allPages } from '../src/services/collections.ts'
import { editable } from '../src/services/payload.ts'
import { toPurchaseOrder, toGoodsReceipt } from '../src/services/procurementDtos.ts'

test('locally filtered tables load records beyond the first API page', async () => {
 const calls = []
 const records = await allPages(async params => {
   calls.push(params)
   return { success: true, data: [params.page], meta: { totalPages: 3 } }
 }, { search: 'camera' })
 assert.deepEqual(records, [1, 2, 3])
 assert.deepEqual(calls.map(p => p.page), [1, 2, 3])
 assert.ok(calls.every(p => p.limit === 100 && p.search === 'camera'))
})
test('a failed collection page rejects instead of silently returning partial records', async () => {
 await assert.rejects(allPages(async ({page}) => page === 1
   ? {success:true, data:[1], meta:{totalPages:2}}
   : {success:false, error:{message:'Database unavailable'}}), /Database unavailable/)
})
test('write payloads exclude display and protected fields while retaining false and zero', () => {
 assert.deepEqual(editable({id:'local', status:'approved', customerName:'Demo', customerId:'', isActive:false, price:0}, ['customerId','isActive','price','notes']), {customerId:null,isActive:false,price:0})
})
test('procurement responses preserve nested display values and safe empty arrays', () => {
 const po = toPurchaseOrder({id:'po', approvedBy:{firstName:'Jane',lastName:'Manager'}, items:[{product:{sku:'CAM-1',name:'Camera',manufacturer:{name:'Acme'}}}]})
 assert.equal(po.approvedBy,'Jane Manager'); assert.equal(po.items[0].productSku,'CAM-1'); assert.equal(po.items[0].manufacturerName,'Acme')
 const receipt = toGoodsReceipt({po:{poNumber:'PO-1'}, receivedBy:{firstName:'Ali',lastName:'Store'}})
 assert.equal(receipt.poNumber,'PO-1'); assert.equal(receipt.receivedBy,'Ali Store'); assert.deepEqual(receipt.items,[])
})
