<script setup lang="ts">
import type { QuoteOffer } from '@/services/customerPricing'
defineProps<{ items: QuoteOffer[] }>()
const date = (value: string) => new Date(value).toLocaleDateString('en-GB')
const money = (value: number, currency: string) => `${currency} ${value.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
</script>
<template>
 <p class="history-note">Each saved supplier offer is retained, including earlier prices. These are purchase costs, not customer selling prices.</p>
 <div v-if="items.length" class="table-container" tabindex="0" aria-label="Supplier price history, scroll horizontally for more columns">
  <table class="table"><thead><tr><th>Recorded</th><th>Supplier / offer</th><th>Customer quotation</th><th class="numeric">Unit cost</th><th class="numeric">Quantity</th><th>Offer state</th></tr></thead><tbody>
   <tr v-for="offer in items" :key="offer.id">
    <td class="nowrap">{{ date(offer.recordedAt) }}</td>
    <td><strong>{{ offer.supplierName }}</strong><div>{{ offer.supplierQuoteNumber }} · revision {{ offer.revision }}</div><div v-if="offer.supplierRef">Ref: {{ offer.supplierRef }}</div></td>
    <td><template v-if="offer.quoteId"><RouterLink :to="`/quotes/${offer.quoteId}/builder`">{{ offer.quoteNumber }} v{{ offer.version }}</RouterLink><div>{{ offer.customerName }}</div><div v-if="offer.projectName">{{ offer.projectName }}</div></template><span v-else>General supplier offer</span></td>
    <td class="numeric nowrap">{{ money(offer.unitCost, offer.currency) }}</td><td class="numeric">{{ offer.quantity }}</td>
    <td>{{ offer.isCurrent ? 'Latest saved offer' : 'Earlier offer' }}<div v-if="offer.validUntil">Valid until {{ date(offer.validUntil) }}</div></td>
   </tr>
  </tbody></table>
 </div>
 <p v-else class="history-note">No supplier offers recorded for this item yet. Link a product when entering the supplier quotation to keep its price history.</p>
</template>
<style scoped>
.history-note { margin-block: var(--space-3); color: var(--color-neutral-700); }
.numeric { text-align: right; font-variant-numeric: tabular-nums; }
.nowrap { white-space: nowrap; }
td { vertical-align: top; } td div { margin-top: var(--space-1); color: var(--color-neutral-600); }
.table-container { overflow: auto; }
</style>
