<script setup lang="ts">
import { ref, watch } from 'vue'
import { customerPriceHistory, customerSalesHistory, type CustomerPrice, type SalePrice } from '@/services/customerPricing'
import { errorMessage } from '@/services/payload'
const props = defineProps<{ customerId: string }>()
const items = ref<CustomerPrice[]>([])
const sales = ref<SalePrice[]>([])
const scope = ref('sold')
const page = ref(1)
const pages = ref(1)
const loading = ref(false)
const error = ref('')
let request = 0
const date = (value: string) => new Date(value).toLocaleDateString('en-GB')
async function load() {
 const current = ++request
 loading.value = true; error.value = ''; items.value = []; sales.value = []
 try {
  if (scope.value === 'sold') { const result = await customerSalesHistory(props.customerId,page.value); if (current === request) { sales.value = result.data; pages.value = result.meta?.totalPages ?? 1 } }
  else { const result = await customerPriceHistory(props.customerId); if (current === request) items.value = result }
 } catch(e) { if (current === request) error.value = errorMessage(e) }
 finally { if (current === request) loading.value = false }
}
function changeScope() { page.value = 1; void load() }
function move(delta:number) { page.value += delta; void load() }
watch(() => props.customerId, () => { page.value = 1; void load() }, { immediate: true })
</script>
<template>
 <section>
 <label class="history-filter">Show prices<select v-model="scope" class="select form-select" @change="changeScope"><option value="sold">Accepted sales</option><option value="quoted">All quotations and revisions</option></select></label>
 <p>{{ scope === 'sold' ? 'Items sold on accepted quotations. Net unit prices include line and quotation discounts, before VAT.' : 'Quoted prices, including drafts and earlier revisions. Only accepted quotations represent sales.' }}</p>
 <p v-if="error" class="form-error" role="alert">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
 <p v-if="loading" role="status">Loading customer prices…</p>
 <template v-else-if="scope === 'sold' && sales.length">
  <div class="table-container" tabindex="0" aria-label="Customer sale prices"><table class="table"><thead><tr><th>Accepted</th><th>Quotation</th><th>Item</th><th>Quantity × multiplier</th><th>Net unit price</th></tr></thead><tbody><tr v-for="item in sales" :key="item.id"><td>{{ date(item.acceptedAt) }}<div v-if="item.acceptedAtEstimated">Estimated legacy date</div></td><td><RouterLink :to="`/quotes/${item.quoteId}/builder`">{{ item.quoteNumber }} v{{ item.version }}</RouterLink></td><td>{{ item.sku }} · {{ item.description }}</td><td>{{ item.quantity }} × {{ item.multiplier || 1 }}</td><td>{{ item.currency }} {{ item.netUnitPrice.toFixed(2) }}</td></tr></tbody></table></div>
  <div class="history-pages"><button class="btn btn-sm" :disabled="page <= 1" @click="move(-1)">Previous</button><span>Page {{ page }} of {{ pages }}</span><button class="btn btn-sm" :disabled="page >= pages" @click="move(1)">Next</button></div>
 </template>
 <div v-else-if="scope === 'quoted' && items.length" class="table-container" tabindex="0" aria-label="Customer quotation prices"><table class="table"><thead><tr><th>Quote</th><th>Item</th><th>Status</th><th>Quantity</th><th>Unit price</th><th>Line discount</th></tr></thead><tbody><tr v-for="item in items" :key="item.id"><td><RouterLink :to="`/quotes/${item.quoteId}/builder`">{{ item.quoteNumber }} v{{ item.version }}</RouterLink></td><td>{{ item.sku }} · {{ item.description }}</td><td>{{ item.status }}</td><td>{{ item.quantity }}</td><td>{{ item.currency }} {{ item.unitPrice.toFixed(2) }}</td><td>{{ item.discountPercent }}%</td></tr></tbody></table></div>
 <p v-else-if="!error">{{ scope === 'sold' ? 'No accepted sales recorded for this customer yet.' : 'No quotations recorded for this customer.' }}</p>
 </section>
</template>
<style scoped>
.history-filter { display: grid; gap: var(--space-2); max-width: 340px; margin-bottom: var(--space-3); }
p { margin-block: var(--space-3); }
.history-pages { display: flex; align-items: center; gap: var(--space-3); margin-top: var(--space-3); }
.table-container { overflow: auto; } td { font-variant-numeric: tabular-nums; }
</style>
