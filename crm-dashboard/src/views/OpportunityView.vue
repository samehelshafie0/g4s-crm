<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Pencil, FileText, MapPin, Truck } from 'lucide-vue-next'
import { opportunitiesService, quotesService } from '@/services'
import { errorMessage } from '@/services/payload'
import type { Activity, QuoteVersion } from '@/services/workflowDtos'
import type { Fulfilment, Opportunity, OpportunityStage } from '@/types'
import { useAuthStore } from '@/stores/auth'
import RecordAttachments from '@/components/RecordAttachments.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

type Tab = 'overview' | 'quote' | 'fulfilment' | 'procurement' | 'documents' | 'activity'
const tabs: { key: Tab; label: string }[] = [
  { key: 'overview', label: 'Overview' },
  { key: 'quote', label: 'Quote & versions' },
  { key: 'fulfilment', label: 'Fulfilment' },
  { key: 'procurement', label: 'Procurement' },
  { key: 'documents', label: 'Documents' },
  { key: 'activity', label: 'Activity' },
]
const tab = ref<Tab>('overview')

const id = computed(() => route.params.id as string)
const opp = ref<Opportunity | null>(null)
const versions = ref<QuoteVersion[]>([])
const fulfilment = ref<Fulfilment | null>(null)
const activity = ref<Activity[]>([])
const loading = ref(false)
const error = ref('')

const stageLabel: Record<OpportunityStage, string> = {
  qualification: 'Qualification',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  'closed-won': 'Closed Won',
  'closed-lost': 'Closed Lost',
}
const stageBadge: Record<OpportunityStage, string> = {
  qualification: 'badge-primary',
  proposal: 'badge-warning',
  negotiation: 'badge-negotiation',
  'closed-won': 'badge-success',
  'closed-lost': 'badge-danger',
}
const stateBadge: Record<string, string> = {
  covered: 'badge-success',
  'on-order': 'badge-info',
  partial: 'badge-warning',
  open: 'badge-danger',
  'not-applicable': 'badge-neutral',
}

const money = (v: number, currency = 'SAR') =>
  `${currency} ${v.toLocaleString('en-SA', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
const stamp = (ts: string) =>
  new Date(ts).toLocaleString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
const day = (d?: string) => (d ? new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) : '—')
const label = (s: string) => s.replace(/-/g, ' ').replace(/^./, (c) => c.toUpperCase())

const orders = computed(() => {
  const seen = new Map<string, { id: string; poNumber: string; status: string }>()
  for (const line of fulfilment.value?.lines ?? [])
    for (const o of line.orders) seen.set(o.purchaseOrderId, { id: o.purchaseOrderId, poNumber: o.poNumber, status: o.status })
  return [...seen.values()]
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await opportunitiesService.get(id.value)
    opp.value = res.data
    versions.value = []
    fulfilment.value = null
    activity.value = []
    if (opp.value.currentQuote && auth.can('quotes:read')) {
      versions.value = (await quotesService.versions(opp.value.currentQuote.id)).data
    }
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

async function loadTab(t: Tab) {
  try {
    if (t === 'fulfilment' || t === 'procurement') {
      if (!fulfilment.value) fulfilment.value = (await opportunitiesService.fulfilment(id.value)).data
    } else if (t === 'activity') {
      activity.value = (await opportunitiesService.activity(id.value)).data
    }
  } catch (e) {
    error.value = errorMessage(e)
  }
}

onMounted(load)
watch(id, load)
watch(tab, loadTab)

function openQuote() {
  if (!opp.value) return
  if (opp.value.currentQuote) router.push(`/quotes/${opp.value.currentQuote.id}/builder`)
  else router.push({ path: '/quotes', query: { newFromOpp: opp.value.id } })
}
</script>

<template>
  <div class="opp-page">
    <button class="btn btn-ghost btn-sm back" @click="router.push('/leads')"><ArrowLeft :size="14" /> Opportunities</button>

    <p v-if="error" class="opp-error" role="alert">{{ error }}</p>
    <p v-if="loading && !opp" class="text-muted">Loading…</p>

    <template v-if="opp">
      <div class="page-header">
        <div>
          <h1 class="page-header-title"><span class="opp-number">{{ opp.opportunityNumber }}</span> {{ opp.title }}</h1>
          <p class="page-header-subtitle">
            {{ opp.customerName }}
            <span class="badge badge-dot" :class="stageBadge[opp.stage]">{{ stageLabel[opp.stage] }}</span>
            <span v-if="opp.deliveryStatus" class="badge badge-info">Delivery: {{ label(opp.deliveryStatus) }}</span>
          </p>
        </div>
        <div class="head-actions">
          <button v-if="auth.can('opportunities:update')" class="btn btn-secondary" @click="router.push({ path: '/leads', query: { edit: opp.id } })">
            <Pencil :size="16" /> Edit
          </button>
          <button v-if="auth.can('quotes:read')" class="btn btn-primary" @click="openQuote">
            <FileText :size="16" /> {{ opp.currentQuote ? 'Open quote' : 'Create quote' }}
          </button>
        </div>
      </div>

      <div class="tabs">
        <button v-for="t in tabs" :key="t.key" class="tab" :class="{ active: tab === t.key }" @click="tab = t.key">{{ t.label }}</button>
      </div>

      <!-- Overview -->
      <div v-if="tab === 'overview'" class="card pad">
        <dl class="facts">
          <div><dt>Estimated value</dt><dd>{{ money(opp.estimatedValue) }}</dd></div>
          <div><dt>Estimated cost</dt><dd>{{ money(opp.estimatedCost) }}</dd></div>
          <div><dt>Win probability</dt><dd>{{ opp.winProbability }}%</dd></div>
          <div><dt>Expected close</dt><dd>{{ day(opp.expectedCloseDate) }}</dd></div>
          <div><dt>Sales executive</dt><dd>{{ opp.salesExecutiveName || '—' }}</dd></div>
          <div><dt>Pre-sales</dt><dd>{{ opp.preSalesName || '—' }}</dd></div>
          <div v-if="opp.deliveryStatus"><dt>Project manager</dt><dd>{{ opp.projectManagerName || '—' }}</dd></div>
          <div v-if="opp.deliveryStatus"><dt>Delivery window</dt><dd>{{ day(opp.startDate) }} → {{ day(opp.targetEndDate) }}</dd></div>
        </dl>
        <h3 class="sub">Sites</h3>
        <ul v-if="opp.sites?.length" class="site-list">
          <li v-for="s in opp.sites" :key="s.id"><MapPin :size="14" /> {{ s.name }}<span v-if="s.city" class="text-muted"> · {{ s.city }}</span></li>
        </ul>
        <p v-else class="text-muted text-sm">No sites chosen. Edit the opportunity to pick from the customer's sites.</p>
        <template v-if="opp.notes">
          <h3 class="sub">Notes</h3>
          <p class="notes">{{ opp.notes }}</p>
        </template>
      </div>

      <!-- Quote & versions -->
      <div v-else-if="tab === 'quote'" class="card pad">
        <template v-if="opp.currentQuote">
          <p>
            <strong class="text-mono">{{ opp.currentQuote.quoteNumber }} v{{ opp.currentQuote.version }}</strong>
            · {{ label(opp.currentQuote.status) }} ·
            {{ money(opp.currentQuote.total, opp.currentQuote.currency) }}
            <span v-if="opp.currentQuote.currency !== 'SAR'" class="text-muted">(1 {{ opp.currentQuote.currency }} = {{ opp.currentQuote.fxRate }} SAR)</span>
          </p>
          <div class="table-container">
            <table class="table table-compact">
              <thead><tr><th>Version</th><th>Status</th><th class="text-right">Total</th><th class="text-right">Margin</th><th>Reason</th><th>Created</th></tr></thead>
              <tbody>
                <tr v-for="v in versions" :key="v.id" :class="{ superseded: !v.isCurrent }">
                  <td>
                    <a href="#" @click.prevent="router.push(`/quotes/${v.id}/builder`)">v{{ v.version }}</a>
                    <span v-if="v.isCurrent" class="badge badge-success">current</span>
                  </td>
                  <td>{{ label(v.status) }}</td>
                  <td class="text-right text-mono">{{ money(v.total, v.currency) }}</td>
                  <td class="text-right">{{ v.marginPercent.toFixed(1) }}%</td>
                  <td>{{ v.revisionReason || '—' }}</td>
                  <td>{{ stamp(v.createdAt) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
        <div v-else class="table-empty">
          <div class="empty-text">No quote yet</div>
          <div class="empty-subtext">Each opportunity has one quote number for its whole life; revisions become new versions.</div>
        </div>
      </div>

      <!-- Fulfilment -->
      <div v-else-if="tab === 'fulfilment'" class="card pad">
        <p v-if="!fulfilment" class="text-muted">Loading…</p>
        <p v-else-if="fulfilment.basis === 'none'" class="text-muted">There is no quote to fulfil yet.</p>
        <template v-else>
          <p class="text-sm text-muted">
            Based on the {{ fulfilment.basis }} version, {{ fulfilment.quoteNumber }} v{{ fulfilment.version }}.
          </p>
          <div class="table-container">
            <table class="table table-compact">
              <thead><tr><th>Item</th><th>Source</th><th class="text-right">Qty</th><th class="text-right">Reserved</th><th class="text-right">Ordered</th><th class="text-right">Received</th><th>State</th></tr></thead>
              <tbody>
                <tr v-for="l in fulfilment.lines" :key="l.quoteLineItemId">
                  <td><div class="font-semibold">{{ l.description }}</div><div class="text-xs text-muted text-mono">{{ l.sku }}</div></td>
                  <td>{{ label(l.source) }}</td>
                  <td class="text-right">{{ l.quantity }}</td>
                  <td class="text-right">{{ l.reserved }}</td>
                  <td class="text-right">{{ l.ordered }}</td>
                  <td class="text-right">{{ l.received }}<span v-if="l.damaged" class="text-danger text-xs"> +{{ l.damaged }} damaged</span></td>
                  <td><span class="badge" :class="stateBadge[l.state]">{{ label(l.state) }}</span></td>
                </tr>
                <tr v-if="!fulfilment.lines.length"><td colspan="7" class="text-muted">No lines.</td></tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>

      <!-- Procurement -->
      <div v-else-if="tab === 'procurement'" class="card pad">
        <p v-if="!fulfilment" class="text-muted">Loading…</p>
        <template v-else>
          <ul v-if="orders.length" class="site-list">
            <li v-for="o in orders" :key="o.id">
              <Truck :size="14" />
              <a href="#" @click.prevent="router.push('/procurement')">{{ o.poNumber }}</a>
              <span class="badge badge-neutral">{{ label(o.status) }}</span>
            </li>
          </ul>
          <p v-else class="text-muted text-sm">No purchase orders are linked to this job's quote lines yet.</p>
        </template>
      </div>

      <!-- Documents -->
      <div v-else-if="tab === 'documents'" class="card pad">
        <RecordAttachments entity-type="opportunity" :entity-id="opp.id" :entity-label="opp.title" />
      </div>

      <!-- Activity -->
      <div v-else class="card pad">
        <ul v-if="activity.length" class="activity">
          <li v-for="a in activity" :key="a.id">
            <div>{{ a.description || a.action }}</div>
            <div class="text-xs text-muted">{{ a.user ? `${a.user.firstName} ${a.user.lastName}` : 'System' }} · {{ stamp(a.createdAt) }}</div>
          </li>
        </ul>
        <p v-else class="text-muted text-sm">No activity recorded yet.</p>
      </div>
    </template>
  </div>
</template>

<style scoped>
.opp-number { font-family: var(--font-mono); font-size: 0.7em; color: var(--color-neutral-500); margin-right: var(--space-2); vertical-align: middle; }
.opp-page { padding: var(--space-6); }
.back { margin-bottom: var(--space-3); }
.head-actions { display: flex; gap: var(--space-2); }
.pad { padding: var(--space-5); }
.tabs { display: flex; gap: var(--space-1); border-bottom: 1px solid var(--color-neutral-200); margin-bottom: var(--space-4); overflow-x: auto; }
.tab { background: none; border: 0; padding: var(--space-2) var(--space-4); cursor: pointer; color: var(--color-neutral-600); border-bottom: 2px solid transparent; white-space: nowrap; }
.tab.active { color: var(--color-primary); border-bottom-color: var(--color-primary); font-weight: 600; }
.facts { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: var(--space-4); margin: 0 0 var(--space-4); }
.facts dt { font-size: var(--text-xs); color: var(--color-neutral-500); text-transform: uppercase; }
.facts dd { margin: 2px 0 0; font-weight: 600; }
.sub { font-size: var(--text-sm); font-weight: 600; margin: var(--space-4) 0 var(--space-2); }
.site-list { list-style: none; padding: 0; margin: 0; display: grid; gap: var(--space-2); }
.site-list li { display: flex; align-items: center; gap: var(--space-2); }
.notes { white-space: pre-wrap; }
.superseded { opacity: 0.65; }
.activity { list-style: none; padding: 0; margin: 0; display: grid; gap: var(--space-3); }
.opp-error { color: var(--color-danger); font-size: var(--text-sm); }
</style>
