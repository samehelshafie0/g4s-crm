<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import AppDialog from './shared/AppDialog.vue'
import SupplierOfferHistory from './SupplierOfferHistory.vue'
import { productSalesHistory, productQuoteOffers, type SalePrice, type QuoteOffer } from '@/services/customerPricing'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
const props = defineProps<{ open: boolean; productId: string; sku: string; productName: string; customerId: string; customerName: string; disabled: boolean }>()
const emit = defineEmits<{ 'update:open': [boolean]; usePrice: [SalePrice] }>()
const auth = useAuthStore()
type Scope = 'customer' | 'others' | 'supplier'
const scope = ref<Scope>('customer')
const sales = ref<SalePrice[]>([])
const offers = ref<QuoteOffer[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pages = ref(1)
const total = ref(0)
let request = 0
const sections = computed(() => [
 { key: 'customer' as Scope, label: 'This customer' },
 ...(auth.can('quotes:compare-prices') ? [{ key: 'others' as Scope, label: 'Other customers' }] : []),
 ...(auth.can('quotes:view-margins') ? [{ key: 'supplier' as Scope, label: 'Supplier offers' }] : []),
])
const date = (value: string) => new Date(value).toLocaleDateString('en-GB')
const money = (value: number, currency: string) => `${currency} ${value.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
async function load() {
 if (!props.open || !props.productId) return
 const current = ++request
 loading.value = true; error.value = ''; sales.value = []; offers.value = []
 try {
  if (scope.value === 'supplier') {
   const result = await productQuoteOffers(props.productId)
   if (current === request) offers.value = result
  } else {
   const result = await productSalesHistory(props.productId, props.customerId, scope.value, page.value)
   if (current === request) { sales.value = result.data; pages.value = result.meta?.totalPages ?? 1; total.value = result.meta?.total ?? result.data.length }
  }
 } catch(e) { if (current === request) error.value = errorMessage(e) }
 finally { if (current === request) loading.value = false }
}
function choose(value: Scope) { scope.value = value; page.value = 1; void load() }
function move(delta: number) { page.value += delta; void load() }
watch(() => [props.open, props.productId, props.customerId] as const, () => {
 request++
 if (props.open) { scope.value = 'customer'; page.value = 1; void load() }
}, { immediate: true })
</script>
<template>
 <AppDialog :open="open" title="Item price history" wide @update:open="emit('update:open', $event)">
  <p class="product-context"><strong>{{ sku }}</strong> · {{ productName }}</p>
  <div class="history-tabs" role="group" aria-label="Price history source">
   <button v-for="section in sections" :key="section.key" class="btn" :class="scope === section.key ? 'btn-primary' : 'btn-secondary'" :aria-pressed="scope === section.key" @click="choose(section.key)">{{ section.label }}</button>
  </div>
  <template v-if="scope !== 'supplier'">
   <p class="history-note">{{ scope === 'customer' ? `Accepted sales to ${customerName}.` : 'Accepted sales to other customers you are permitted to view.' }} Drafts, declined quotations and unselected optional items are excluded.</p>
   <p class="history-note">Net unit prices include line and quotation discounts, before VAT. Using a price converts it at the pinned rates and accounts for current discounts. Unit cost stays fixed.</p>
  </template>
  <p v-if="loading" role="status" class="state">Loading price history…</p>
  <p v-else-if="error" role="alert" class="form-error state">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
  <SupplierOfferHistory v-else-if="scope === 'supplier'" :items="offers" />
  <template v-else-if="sales.length">
   <div class="table-container" tabindex="0" aria-label="Accepted sale prices, scroll horizontally for more columns"><table class="table history-table"><thead><tr><th>Accepted</th><th>Customer / quotation</th><th class="numeric">Quantity × multiplier</th><th class="numeric">Unit price</th><th>Discounts</th><th class="numeric">Net unit price</th><th v-if="!disabled">Action</th></tr></thead><tbody>
    <tr v-for="sale in sales" :key="sale.id">
     <td class="nowrap">{{ date(sale.acceptedAt) }}<div v-if="sale.acceptedAtEstimated" class="secondary">Estimated legacy date</div></td>
     <td><strong>{{ sale.customerName }}</strong><div><RouterLink :to="`/quotes/${sale.quoteId}/builder`">{{ sale.quoteNumber }} v{{ sale.version }}</RouterLink></div></td>
     <td class="numeric">{{ sale.quantity }} × {{ sale.multiplier || 1 }}</td>
     <td class="numeric nowrap">{{ money(sale.unitPrice, sale.currency) }}</td>
     <td class="nowrap">Line {{ sale.discountPercent }}%<div class="secondary">Quote {{ sale.quoteDiscountPercent }}%</div></td>
     <td class="numeric nowrap"><strong>{{ money(sale.netUnitPrice, sale.currency) }}</strong><div v-if="sale.currency !== 'SAR'" class="secondary">{{ money(sale.netUnitPrice * sale.fxRate, 'SAR') }} at sale rate</div></td>
     <td v-if="!disabled"><button class="btn btn-sm btn-primary" :aria-label="`Use net price from ${sale.quoteNumber}`" @click="emit('usePrice', sale)">Use price</button></td>
    </tr>
   </tbody></table></div>
   <div class="pagination"><span>{{ total }} accepted sale {{ total === 1 ? 'line' : 'lines' }} · Page {{ page }} of {{ pages }}</span><div><button class="btn btn-sm" :disabled="page <= 1" @click="move(-1)">Previous</button><button class="btn btn-sm" :disabled="page >= pages" @click="move(1)">Next</button></div></div>
  </template>
  <p v-else class="state">{{ scope === 'customer' ? 'No accepted sales of this item to this customer yet.' : 'No accepted sales of this item to other accessible customers yet.' }}</p>
  <template #footer><button class="btn btn-secondary" @click="emit('update:open', false)">Close history</button></template>
 </AppDialog>
</template>
<style scoped>
.product-context { margin-bottom: var(--space-4); overflow-wrap: anywhere; }
.history-tabs { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.history-tabs .btn { min-height: 44px; }
.history-tabs .btn-primary, .history-table .btn-primary { background: var(--color-primary-light); color: var(--color-neutral-900); border-color: var(--color-primary); }
.history-note { color: var(--color-neutral-700); margin-block: var(--space-3); max-width: 80ch; }
.state { padding-block: var(--space-6); }
.table-container { overflow: auto; margin-top: var(--space-4); }
.numeric { text-align: right; font-variant-numeric: tabular-nums; }
.nowrap { white-space: nowrap; }
.history-table td { vertical-align: top; }
.secondary { color: var(--color-neutral-600); margin-top: var(--space-1); font-size: var(--text-xs); }
.pagination { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: var(--space-3); margin-top: var(--space-4); }
.pagination div { display: flex; gap: var(--space-2); }
</style>
