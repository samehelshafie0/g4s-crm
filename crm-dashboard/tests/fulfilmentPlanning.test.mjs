import test from 'node:test'
import assert from 'node:assert/strict'
import { planQuoteHolds } from '../src/services/fulfilmentPlanning.ts'
import { canonicalVendorName, vendorKey, usableVendorName } from '../src/services/vendorNames.ts'
const line = (id, quantity, extra = {}) => ({ id, productId: 'p', sku: 'CAM', description: id, quantity, multiplier: 1, ...extra })
const stock = [{ id: 'r', productId: 'p', productSku: 'CAM', warehouseLocation: 'riyadh-main', availableQty: 4 }, { id: 'j', productId: 'p', productSku: 'CAM', warehouseLocation: 'jeddah-branch', availableQty: 3 }]
test('holds carry exact quote line IDs and never reuse stock across repeated products', () => {
  const plan = planQuoteHolds([line('first', 5), line('second', 5)], stock, [], [])
  assert.deepEqual(plan.map(p => [p.quoteLineItemId, p.warehouseLocation, p.holdQty, p.availableQty]), [['first','riyadh-main',4,4],['first','jeddah-branch',1,1],['second','jeddah-branch',2,2],['second','riyadh-main',3,0]])
})
test('quantity multiplier, fulfilled holds and draft orders reduce demand; cancelled orders and released holds do not', () => {
  const holds = [{quoteLineItemId:'first', qty:2, status:'fulfilled'}, {quoteLineItemId:'first', qty:100, status:'released'}]
  const orders = [{status:'draft',items:[{quoteLineItemId:'first',quantity:3,receivedQty:1}]}, {status:'cancelled',items:[{quoteLineItemId:'first',quantity:100,receivedQty:0}]}]
  assert.equal(planQuoteHolds([line('first',2.5,{multiplier:2})],stock,holds,orders)[0].holdQty,1)
})
test('unselected optional and service lines are excluded; fully committed demand is absent', () => {
  const lines = [line('optional',5,{isOptional:true,isSelected:false}),line('service',5,{fulfilmentSource:'service'}),line('done',2)]
  assert.deepEqual(planQuoteHolds(lines,stock,[{quoteLineItemId:'done',qty:2,status:'active'}],[]),[])
})
test('verified vendor aliases share a supplier option and incomplete vendor names cannot be selected', () => {
  assert.equal(canonicalVendorName('  alfanar  Electric (Riyadh) '),'alfanar Electric')
  assert.equal(vendorKey('alfanar Electric (Riyadh)'),vendorKey('alfanar Electric'))
  assert.equal(usableVendorName(' in '),false)
  assert.equal(canonicalVendorName('Ingram Micro'),'Ingram Micro')
})

test('warehouse plan reserves only available units and respects server demand beyond local order pages', () => {
  const plan = planQuoteHolds([line('first', 20, {remainingQty: 8}), line('done', 2, {remainingQty: 0})], stock, [], [])
  assert.equal(plan.filter(p => p.selected).reduce((sum, p) => sum + p.holdQty, 0), 7)
  assert.equal(plan.find(p => p.availableQty === 0).selected, false)
  assert.equal(plan.reduce((sum, p) => sum + p.holdQty, 0), 8)
})
