<script setup lang="ts">
import { ref, watch } from 'vue'
import { http, type ApiResponse } from '@/services/http'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import RecordAttachments from './RecordAttachments.vue'
export interface SupplierLine { id: string; productId?: string; productSku: string; productName: string; unitCost: number; quantity: number; leadTimeDays: number }
interface SupplierQuote { id: string; sqNumber: string; supplierName: string; supplierRef: string; currency: string; items: SupplierLine[] }
const props = defineProps<{ quoteId: string; disabled: boolean; currency: string }>()
const emit = defineEmits<{ usePrice: [item: SupplierLine, currency: string] }>()
const auth = useAuthStore()
const linked = ref<SupplierQuote[]>([])
const available = ref<SupplierQuote[]>([])
const selection = ref('')
const error = ref('')
const busy = ref(false)
async function load() {
 error.value = ''
 try {
  const result = await http.get<ApiResponse<SupplierQuote[]>>(`/quotes/${props.quoteId}/supplier-quotes`)
  linked.value = result.data.data
  if (!props.disabled && auth.can('quotes:documents')) available.value = (await http.get<ApiResponse<SupplierQuote[]>>(`/quotes/${props.quoteId}/supplier-quotes`, { params: { available: true } })).data.data
 } catch(e) { error.value = errorMessage(e) }
}
watch(() => props.quoteId, load, { immediate: true })
async function link() {
 if (busy.value || !selection.value) return
 busy.value = true; error.value = ''
 try { await http.post(`/quotes/${props.quoteId}/supplier-quotes`, { supplierQuoteId: selection.value }); selection.value = ''; await load() }
 catch(e) { error.value = errorMessage(e) }
 finally { busy.value = false }
}
</script>
<template>
 <section class="supplier-quotes">
  <h3>Supplier quotations for this project</h3>
  <p>Link supplier pricing and its original files to this customer quotation. Each added line keeps its own cost snapshot.</p>
  <p v-if="error" class="form-error" role="alert">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
  <form v-if="!disabled && auth.can('quotes:documents')" class="supplier-link" @submit.prevent="link"><label>Supplier quotation<select v-model="selection" class="select form-select"><option value="">Choose an unlinked quotation</option><option v-for="quote in available" :key="quote.id" :value="quote.id">{{ quote.sqNumber }} · {{ quote.supplierName }} · {{ quote.supplierRef }}</option></select></label><button class="btn btn-sm" :disabled="busy || !selection">{{ busy ? 'Linking…' : 'Link quotation' }}</button></form>
  <p v-if="!linked.length && !error">No supplier quotations linked yet. Create one in Procurement, then link it here.</p>
  <details v-for="quote in linked" :key="quote.id" open>
   <summary>{{ quote.sqNumber }} · {{ quote.supplierName }} · {{ quote.currency }}</summary>
   <div class="table-container"><table class="table"><thead><tr><th>Product</th><th>Supplier unit cost</th><th>Quantity</th><th>Action</th></tr></thead><tbody><tr v-for="item in quote.items" :key="item.id"><td>{{ item.productSku }} · {{ item.productName }}</td><td>{{ quote.currency }} {{ item.unitCost.toFixed(2) }}</td><td>{{ item.quantity }}</td><td><button v-if="!disabled && item.productId && auth.can('quotes:pricing')" class="btn btn-sm" @click="emit('usePrice', item, quote.currency)">Add at this cost</button><span v-else-if="!item.productId">Link product in Procurement first</span></td></tr></tbody></table></div>
   <RecordAttachments quote-context entity-type="supplier-quote" :entity-id="quote.id" :entity-label="quote.sqNumber" :disabled="disabled || !auth.can('quotes:documents')" />
  </details>
 </section>
</template>
<style scoped>
.supplier-quotes { margin-block: var(--space-5); }
.supplier-link { display: flex; align-items: end; flex-wrap: wrap; gap: var(--space-3); margin-block: var(--space-4); }
label { display: grid; gap: var(--space-2); flex: 1; }
details { padding-block: var(--space-3); border-bottom: 1px solid var(--color-neutral-200); }
summary { cursor: pointer; font-weight: 600; }
</style>
