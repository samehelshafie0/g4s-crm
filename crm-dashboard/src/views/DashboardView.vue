<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  Target,
  FileText,
  ScrollText,
  ArrowUpRight,
  ArrowRight,
  Building2,
  AlertTriangle,
  RefreshCw,
  Activity as ActivityIcon,
  ClipboardCheck,
  CalendarClock,
  Inbox,
} from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useDashboardStore } from '@/stores/dashboard'
import { useAuthStore } from '@/stores/auth'
import { dashboardService } from '@/services'
import type { Activity as ApiActivity } from '@/services/workflowDtos'
import type { Quote as ApiQuote } from '@/types'

const router = useRouter()
const dashStore = useDashboardStore()
const auth = useAuthStore()

// ── Data loading ─────────────────────────────────────────────
interface RecentQuote {
  id: string
  number: string
  customer: string
  status: ApiQuote['status']
  total: number
  currency: string
  margin: number
  date: string
}

interface ExpiringQuote {
  id: string
  number: string
  customer: string
  total: number
  currency: string
  expiresIn: number
}

interface SalesRep {
  name: string
  quotesCreated: number
  quotesWon: number
  revenue: number
  winRate: number
}

const recentQuotes = ref<RecentQuote[]>([])
const expiringQuotes = ref<ExpiringQuote[]>([])
const salesReps = ref<SalesRep[]>([])
const loaded = ref(false)

async function load() {
  await dashStore.fetchAll()
  try {
    const [recent, expiring, reps] = await Promise.all([
      dashboardService.recentQuotes(),
      dashboardService.expiringQuotes(),
      dashboardService.salesPerformance(),
    ])
    recentQuotes.value = recent.data.map((q) => ({
      id: q.id,
      number: q.quoteNumber,
      customer: q.customerName,
      status: q.status,
      total: q.total,
      currency: q.currency,
      margin: q.marginPercent,
      date: q.createdAt,
    }))
    expiringQuotes.value = expiring.data.map((q) => ({
      id: q.id,
      number: q.quoteNumber,
      customer: q.customerName,
      total: q.total,
      currency: q.currency,
      expiresIn: Math.max(0, Math.ceil((new Date(q.validUntil).getTime() - Date.now()) / 86_400_000)),
    }))
    salesReps.value = reps.data
  } catch {
    dashStore.error = 'Some dashboard data could not be loaded'
  } finally {
    loaded.value = true
  }
}

onMounted(load)

const showSkeleton = computed(() => !loaded.value && dashStore.loading)

// ── Header ───────────────────────────────────────────────────
const greeting = computed(() => {
  const hour = new Date().getHours()
  const part = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening'
  const name = auth.user?.firstName
  return name ? `${part}, ${name}` : part
})

const today = new Date().toLocaleDateString('en-GB', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  year: 'numeric',
})

// ── Formatting ───────────────────────────────────────────────
function compact(value: number): string {
  const abs = Math.abs(value)
  if (abs >= 1_000_000) return `${(value / 1_000_000).toFixed(abs >= 10_000_000 ? 0 : 1)}M`
  if (abs >= 10_000) return `${(value / 1_000).toFixed(0)}K`
  if (abs >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return value.toLocaleString('en-GB', { maximumFractionDigits: 0 })
}

function money(value: number, currency = 'SAR'): string {
  return `${currency} ${compact(value)}`
}

function exactMoney(value: number, currency = 'SAR'): string {
  return `${currency} ${value.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.round(diff / 60_000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.round(hours / 24)
  if (days < 7) return `${days}d ago`
  return formatDate(dateStr)
}

// ── Alerts (needs attention) ─────────────────────────────────
interface Alert {
  type: string
  severity: string
  message: string
  count: number
}

const alertMeta: Record<string, { icon: typeof Target; to: string; tone: 'action' | 'warning' }> = {
  pending_approvals: { icon: ClipboardCheck, to: '/quotes', tone: 'action' },
  contract_expiring: { icon: CalendarClock, to: '/contracts', tone: 'warning' },
}

const alerts = computed(() =>
  (dashStore.alerts as Alert[]).map((a) => ({
    ...a,
    icon: alertMeta[a.type]?.icon ?? AlertTriangle,
    to: alertMeta[a.type]?.to ?? '/dashboard',
    tone: alertMeta[a.type]?.tone ?? 'warning',
  })),
)

// ── KPIs ─────────────────────────────────────────────────────
interface KpiValue {
  value?: number
  totalValue?: number
  expiringIn30Days?: number
}

const kpis = computed(() => dashStore.kpis as Record<string, KpiValue | undefined>)

const heroValue = computed(() => kpis.value.totalRevenue?.value)
const statTiles = computed(() => [
  {
    key: 'opps',
    label: 'Active opportunities',
    value: kpis.value.activeOpportunities?.value,
    note: 'Open in the pipeline',
    icon: Target,
    to: '/leads',
  },
  {
    key: 'quotes',
    label: 'Open quotes',
    value: kpis.value.openQuotes?.value,
    note: kpis.value.openQuotes ? `${money(kpis.value.openQuotes.totalValue ?? 0)} outstanding` : '',
    icon: FileText,
    to: '/quotes',
  },
  {
    key: 'contracts',
    label: 'Active contracts',
    value: kpis.value.activeContracts?.value,
    note: kpis.value.activeContracts
      ? `${kpis.value.activeContracts.expiringIn30Days ?? 0} ending within 30 days`
      : '',
    icon: ScrollText,
    to: '/contracts',
  },
])

// ── Pipeline (ordinal ramp, fixed funnel order) ──────────────
interface PipelineStage {
  key: string
  name: string
  count: number
  value: number
  step: number
}

const stageOrder = ['qualification', 'proposal', 'negotiation', 'closed-won', 'closed-lost']
const stageNames: Record<string, string> = {
  qualification: 'Qualification',
  proposal: 'Proposal',
  negotiation: 'Negotiation',
  'closed-won': 'Closed won',
  'closed-lost': 'Closed lost',
}

const pipelineStages = computed<PipelineStage[]>(() => {
  const raw = dashStore.pipeline as Array<{ stage: string; count: number; value: number }>
  const byStage = new Map(raw.map((s) => [s.stage, s]))
  return stageOrder.map((key, index) => ({
    key,
    name: stageNames[key] ?? key,
    count: byStage.get(key)?.count ?? 0,
    value: byStage.get(key)?.value ?? 0,
    step: index,
  }))
})

const hasPipeline = computed(() => pipelineStages.value.some((s) => s.count > 0))
const maxPipelineValue = computed(() => Math.max(...pipelineStages.value.map((s) => s.value), 1))
const openPipeline = computed(() =>
  pipelineStages.value.filter((s) => !s.key.startsWith('closed')).reduce((sum, s) => sum + s.value, 0),
)
const openDeals = computed(() =>
  pipelineStages.value.filter((s) => !s.key.startsWith('closed')).reduce((sum, s) => sum + s.count, 0),
)

// ── Recent quotes ────────────────────────────────────────────
const statusConfig: Record<ApiQuote['status'], { label: string; class: string }> = {
  draft: { label: 'Draft', class: 'badge-neutral' },
  'pending-approval': { label: 'Pending', class: 'badge-warning' },
  approved: { label: 'Approved', class: 'badge-primary' },
  sent: { label: 'Sent', class: 'badge-primary' },
  accepted: { label: 'Accepted', class: 'badge-success' },
  declined: { label: 'Declined', class: 'badge-danger' },
  expired: { label: 'Expired', class: 'badge-neutral' },
}

function openQuote(id: string) {
  router.push(`/quotes/${id}/builder`)
}

// ── Expiring quotes ──────────────────────────────────────────
function urgency(days: number): 'critical' | 'warning' | 'notice' {
  if (days <= 2) return 'critical'
  if (days <= 5) return 'warning'
  return 'notice'
}

function expiresLabel(days: number): string {
  if (days <= 0) return 'Today'
  if (days === 1) return 'Tomorrow'
  return `${days} days`
}

// ── Activity ─────────────────────────────────────────────────
const entityIcon: Record<string, typeof Target> = {
  quote: FileText,
  contract: ScrollText,
  customer: Building2,
  opportunity: Target,
}

const activities = computed(() =>
  (dashStore.recentActivity as ApiActivity[]).slice(0, 8).map((a) => ({
    id: a.id,
    icon: entityIcon[a.entityType] ?? ActivityIcon,
    description: a.description,
    createdAt: a.createdAt,
    user: a.user ? `${a.user.firstName} ${a.user.lastName}` : 'System',
  })),
)

// ── Sales reps ───────────────────────────────────────────────
const maxRepRevenue = computed(() => Math.max(...salesReps.value.map((r) => r.revenue), 1))

function initials(name: string): string {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

// ── Top customers ────────────────────────────────────────────
interface TopCustomer {
  id: string
  name: string
  value: number
  quotes: number
}

const topCustomers = computed<TopCustomer[]>(() => {
  const raw = dashStore.topCustomers as Array<{
    customerId: string
    companyName: string
    totalRevenue: number
    quoteCount: number
  }>
  return (raw ?? []).slice(0, 6).map((c) => ({
    id: c.customerId,
    name: c.companyName,
    value: c.totalRevenue ?? 0,
    quotes: c.quoteCount ?? 0,
  }))
})

const maxCustomerValue = computed(() => Math.max(...topCustomers.value.map((c) => c.value), 1))
</script>

<template>
  <div class="dashboard">
    <!-- Header -->
    <header class="dash-header">
      <div>
        <h1 class="dash-title">{{ greeting }}</h1>
        <p class="dash-subtitle">{{ today }}</p>
      </div>
      <button class="btn btn-ghost btn-sm" :disabled="dashStore.loading" @click="load">
        <RefreshCw :size="14" :class="{ spinning: dashStore.loading }" />
        Refresh
      </button>
    </header>

    <!-- Error -->
    <div v-if="dashStore.error" class="dash-error" role="alert">
      <AlertTriangle :size="16" />
      <span>{{ dashStore.error }}</span>
      <button class="btn btn-ghost btn-sm" @click="load">Try again</button>
    </div>

    <!-- Needs attention -->
    <section v-if="alerts.length" class="attention" aria-label="Needs attention">
      <RouterLink
        v-for="alert in alerts"
        :key="alert.type"
        :to="alert.to"
        class="attention-item"
        :class="`attention-item--${alert.tone}`"
      >
        <component :is="alert.icon" :size="16" class="attention-icon" />
        <span class="attention-count">{{ alert.count }}</span>
        <span class="attention-text">{{ alert.message }}</span>
        <ArrowRight :size="14" class="attention-arrow" />
      </RouterLink>
    </section>

    <!-- KPI row -->
    <section class="kpi-grid" aria-label="Key figures">
      <template v-if="showSkeleton">
        <div v-for="n in 4" :key="n" class="card kpi-tile" :class="{ 'kpi-tile--hero': n === 1 }">
          <div class="skeleton skeleton-text" style="width: 40%" />
          <div class="skeleton skeleton-title" style="width: 60%; margin-top: 12px" />
          <div class="skeleton skeleton-text" style="width: 50%; margin-top: 10px" />
        </div>
      </template>
      <template v-else>
        <RouterLink to="/quotes" class="card kpi-tile kpi-tile--hero">
          <span class="kpi-label">Accepted quote value</span>
          <span class="kpi-hero">
            <span class="kpi-currency">SAR</span>
            {{ heroValue === undefined ? '—' : compact(heroValue) }}
          </span>
          <span class="kpi-note">All accepted quotations, SAR only</span>
        </RouterLink>

        <RouterLink v-for="tile in statTiles" :key="tile.key" :to="tile.to" class="card kpi-tile">
          <span class="kpi-label">
            <component :is="tile.icon" :size="14" class="kpi-label-icon" />
            {{ tile.label }}
          </span>
          <span class="kpi-value">{{ tile.value === undefined ? '—' : tile.value.toLocaleString('en-GB') }}</span>
          <span class="kpi-note">{{ tile.note || ' ' }}</span>
        </RouterLink>
      </template>
    </section>

    <!-- Content grid -->
    <div class="content-grid">
      <!-- Left column -->
      <div class="column">
        <!-- Pipeline -->
        <section class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Pipeline</h3>
              <p class="card-caption">Estimated value by stage</p>
            </div>
            <div class="pipeline-summary">
              <span class="pipeline-summary-value">{{ money(openPipeline) }}</span>
              <span class="pipeline-summary-label">open across {{ openDeals }} deals</span>
            </div>
          </div>
          <div class="card-body">
            <div v-if="showSkeleton" class="stack">
              <div v-for="n in 5" :key="n" class="skeleton skeleton-text" />
            </div>
            <div v-else-if="!hasPipeline" class="empty">
              <Inbox :size="20" />
              <p>No opportunities yet. Create one to start the pipeline.</p>
            </div>
            <ol v-else class="pipeline" aria-label="Pipeline by stage">
              <li
                v-for="stage in pipelineStages"
                :key="stage.key"
                class="pipeline-row"
                :class="{ 'pipeline-row--lost': stage.key === 'closed-lost' }"
              >
                <span class="pipeline-name">{{ stage.name }}</span>
                <div
                  class="bar-track"
                  :title="`${stage.name}: ${stage.count} deals, ${exactMoney(stage.value)}`"
                >
                  <div
                    class="bar-fill"
                    :class="[`bar-fill--step-${stage.step}`, { 'bar-fill--lost': stage.key === 'closed-lost' }]"
                    :style="{ width: `${(stage.value / maxPipelineValue) * 100}%` }"
                  />
                </div>
                <span class="pipeline-count">{{ stage.count }}</span>
                <span class="pipeline-value">{{ money(stage.value) }}</span>
              </li>
            </ol>
          </div>
        </section>

        <!-- Expiring quotes -->
        <section v-if="expiringQuotes.length" class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Expiring soon</h3>
              <p class="card-caption">Sent or approved quotes past their validity within 30 days</p>
            </div>
            <span class="card-header-meta">{{ expiringQuotes.length }}</span>
          </div>
          <div class="card-body card-body--list">
            <button
              v-for="eq in expiringQuotes"
              :key="eq.id"
              type="button"
              class="row-button"
              @click="openQuote(eq.id)"
            >
              <span class="urgency-stripe" :class="`urgency-stripe--${urgency(eq.expiresIn)}`" aria-hidden="true" />
              <span class="row-main">
                <span class="row-title">{{ eq.customer }}</span>
                <span class="row-sub text-mono">{{ eq.number }}</span>
              </span>
              <span class="row-side">
                <span class="row-amount">{{ money(eq.total, eq.currency) }}</span>
                <span class="urgency-pill" :class="`urgency-pill--${urgency(eq.expiresIn)}`">
                  {{ expiresLabel(eq.expiresIn) }}
                </span>
              </span>
            </button>
          </div>
        </section>

        <!-- Recent quotes -->
        <section class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Recent quotes</h3>
              <p class="card-caption">Latest ten quotations</p>
            </div>
            <RouterLink to="/quotes" class="btn btn-ghost btn-sm">
              View all
              <ArrowUpRight :size="14" />
            </RouterLink>
          </div>
          <div v-if="showSkeleton" class="card-body stack">
            <div v-for="n in 5" :key="n" class="skeleton skeleton-text" />
          </div>
          <div v-else-if="!recentQuotes.length" class="card-body empty">
            <Inbox :size="20" />
            <p>No quotations yet.</p>
          </div>
          <div v-else class="table-container table-container--embedded">
            <table class="table quotes-table">
              <thead>
                <tr>
                  <th>Quote</th>
                  <th>Customer</th>
                  <th>Status</th>
                  <th class="text-right">Total</th>
                  <th class="text-right">Margin</th>
                  <th class="text-right">Created</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="quote in recentQuotes" :key="quote.id" class="quote-row" @click="openQuote(quote.id)">
                  <td class="text-mono font-medium">{{ quote.number }}</td>
                  <td class="quote-customer">{{ quote.customer }}</td>
                  <td>
                    <span class="badge badge-dot" :class="statusConfig[quote.status]?.class ?? 'badge-neutral'">
                      {{ statusConfig[quote.status]?.label ?? quote.status }}
                    </span>
                  </td>
                  <td class="text-right num">{{ exactMoney(quote.total, quote.currency) }}</td>
                  <td class="text-right num">
                    <span class="margin" :class="{ 'margin--low': quote.margin < 30 }">
                      <i class="margin-dot" aria-hidden="true" />
                      {{ quote.margin.toFixed(1) }}%
                    </span>
                  </td>
                  <td class="text-right text-muted">{{ formatDate(quote.date) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <!-- Right column -->
      <div class="column">
        <!-- Sales performance -->
        <section class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Sales performance</h3>
              <p class="card-caption">This quarter · accepted SAR value</p>
            </div>
          </div>
          <div class="card-body">
            <div v-if="showSkeleton" class="stack">
              <div v-for="n in 3" :key="n" class="skeleton skeleton-text" />
            </div>
            <div v-else-if="!salesReps.length" class="empty">
              <Inbox :size="20" />
              <p>No quotes created this quarter.</p>
            </div>
            <ol v-else class="rep-list">
              <li v-for="rep in salesReps" :key="rep.name" class="rep-row">
                <span class="avatar">{{ initials(rep.name) }}</span>
                <div class="rep-body">
                  <div class="rep-head">
                    <span class="rep-name">{{ rep.name }}</span>
                    <span class="rep-revenue">{{ money(rep.revenue) }}</span>
                  </div>
                  <div class="bar-track bar-track--thin" :title="`${rep.name}: ${exactMoney(rep.revenue)}`">
                    <div class="bar-fill bar-fill--accent" :style="{ width: `${(rep.revenue / maxRepRevenue) * 100}%` }" />
                  </div>
                  <div class="rep-foot">
                    <span>{{ rep.quotesWon }} of {{ rep.quotesCreated }} won</span>
                    <span class="rep-winrate" :class="{ 'rep-winrate--low': rep.winRate < 50 }">
                      {{ Math.round(rep.winRate) }}% win rate
                    </span>
                  </div>
                </div>
              </li>
            </ol>
          </div>
        </section>

        <!-- Top customers -->
        <section class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Top customers</h3>
              <p class="card-caption">By accepted quote value, SAR</p>
            </div>
          </div>
          <div class="card-body">
            <div v-if="showSkeleton" class="stack">
              <div v-for="n in 4" :key="n" class="skeleton skeleton-text" />
            </div>
            <div v-else-if="!topCustomers.length" class="empty">
              <Inbox :size="20" />
              <p>No accepted quotes yet.</p>
            </div>
            <ol v-else class="customer-list">
              <li v-for="(customer, index) in topCustomers" :key="customer.id" class="customer-row">
                <span class="customer-rank">{{ index + 1 }}</span>
                <div class="customer-body">
                  <div class="customer-head">
                    <span class="customer-name">{{ customer.name }}</span>
                    <span class="customer-value">{{ money(customer.value) }}</span>
                  </div>
                  <div class="bar-track bar-track--thin" :title="`${customer.name}: ${exactMoney(customer.value)} across ${customer.quotes} quotes`">
                    <div class="bar-fill bar-fill--accent" :style="{ width: `${(customer.value / maxCustomerValue) * 100}%` }" />
                  </div>
                  <span class="customer-quotes">{{ customer.quotes }} accepted {{ customer.quotes === 1 ? 'quote' : 'quotes' }}</span>
                </div>
              </li>
            </ol>
          </div>
        </section>

        <!-- Activity -->
        <section class="card">
          <div class="card-header">
            <div>
              <h3 class="card-title">Recent activity</h3>
              <p class="card-caption">Latest changes across the CRM</p>
            </div>
          </div>
          <div class="card-body">
            <div v-if="showSkeleton" class="stack">
              <div v-for="n in 5" :key="n" class="skeleton skeleton-text" />
            </div>
            <div v-else-if="!activities.length" class="empty">
              <Inbox :size="20" />
              <p>Nothing has happened yet.</p>
            </div>
            <ol v-else class="activity">
              <li v-for="item in activities" :key="item.id" class="activity-item">
                <span class="activity-icon">
                  <component :is="item.icon" :size="14" />
                </span>
                <div class="activity-body">
                  <p class="activity-text">{{ item.description }}</p>
                  <p class="activity-meta">
                    <span>{{ item.user }}</span>
                    <span aria-hidden="true">·</span>
                    <time :datetime="item.createdAt" :title="new Date(item.createdAt).toLocaleString('en-GB')">
                      {{ relativeTime(item.createdAt) }}
                    </time>
                  </p>
                </div>
              </li>
            </ol>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Local tokens: the pipeline's ordinal ramp (validated light & dark) ── */
.dashboard {
  --ramp-0: #60a5fa;
  --ramp-1: #3b82f6;
  --ramp-2: #1d4ed8;
  --ramp-3: #1e3a8a;
  --ramp-track: var(--color-primary-50);
  --lost-fill: var(--color-neutral-400);
  --lost-track: var(--color-neutral-100);
  max-width: 1400px;
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

:global([data-theme='dark']) .dashboard {
  --ramp-0: #1d4ed8;
  --ramp-1: #3b82f6;
  --ramp-2: #60a5fa;
  --ramp-3: #bfdbfe;
  --ramp-track: rgba(96, 165, 250, 0.12);
  --lost-fill: var(--color-neutral-400);
  --lost-track: rgba(148, 163, 184, 0.12);
}

/* ── Header ── */
.dash-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.dash-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
  letter-spacing: -0.01em;
  text-wrap: balance;
}

.dash-subtitle {
  margin-top: var(--space-1);
  font-size: var(--text-sm);
  color: var(--color-neutral-500);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* ── Error ── */
.dash-error {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--color-danger-light);
  background: var(--color-danger-light);
  color: var(--color-danger-dark);
  border-radius: var(--radius-lg);
  font-size: var(--text-sm);
}

.dash-error span {
  flex: 1;
}

/* ── Needs attention ── */
.attention {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.attention-item {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3) var(--space-2) var(--space-3);
  border-radius: var(--radius-full);
  border: 1px solid transparent;
  font-size: var(--text-sm);
  color: var(--color-neutral-800);
  text-decoration: none;
  transition: border-color var(--transition-fast), background-color var(--transition-fast);
}

.attention-item--action {
  background: var(--color-primary-light);
  border-color: var(--color-primary-200);
}

.attention-item--action .attention-icon {
  color: var(--color-primary-dark);
}

.attention-item--warning {
  background: var(--color-warning-light);
  border-color: transparent;
}

.attention-item--warning .attention-icon {
  color: var(--color-warning-dark);
}

.attention-item:hover,
.attention-item:focus-visible {
  border-color: var(--color-neutral-400);
  outline: none;
}

.attention-count {
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
}

.attention-arrow {
  color: var(--color-neutral-500);
  margin-left: var(--space-1);
}

/* ── KPI row ── */
.kpi-grid {
  display: grid;
  grid-template-columns: 1.6fr repeat(3, 1fr);
  gap: var(--space-4);
}

.kpi-tile {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-5);
  text-decoration: none;
  color: inherit;
  min-width: 0;
}

.kpi-tile:hover,
.kpi-tile:focus-visible {
  border-color: var(--color-primary-200);
  outline: none;
}

.kpi-tile--hero {
  background: linear-gradient(135deg, var(--color-primary-50), var(--content-surface) 70%);
}

.kpi-label {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-neutral-500);
}

.kpi-label-icon {
  color: var(--color-neutral-400);
}

.kpi-hero {
  display: flex;
  align-items: baseline;
  gap: var(--space-2);
  margin-top: var(--space-2);
  font-size: 2.5rem;
  font-weight: var(--font-semibold);
  line-height: 1;
  letter-spacing: -0.02em;
  color: var(--color-neutral-900);
}

.kpi-currency {
  font-size: var(--text-md);
  font-weight: var(--font-medium);
  letter-spacing: 0.04em;
  color: var(--color-neutral-500);
}

.kpi-value {
  margin-top: var(--space-2);
  font-size: var(--text-3xl);
  font-weight: var(--font-semibold);
  line-height: 1;
  letter-spacing: -0.02em;
  color: var(--color-neutral-900);
}

.kpi-note {
  margin-top: var(--space-2);
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ── Content grid ── */
.content-grid {
  display: grid;
  grid-template-columns: 3fr 2fr;
  gap: var(--space-6);
  align-items: start;
}

.column {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  min-width: 0;
}

.card-caption {
  margin-top: 2px;
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.card-header-meta {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-600);
  padding: 2px var(--space-2);
  background: var(--color-neutral-100);
  border-radius: var(--radius-full);
}

.card-body--list {
  padding: 0 var(--space-2) var(--space-2);
}

.stack {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-6) 0;
  color: var(--color-neutral-400);
  font-size: var(--text-sm);
  text-align: center;
}

/* ── Bars (shared mark spec: thin, rounded data-end, square baseline) ── */
.bar-track {
  position: relative;
  height: 10px;
  border-radius: 0 4px 4px 0;
  background: var(--ramp-track);
  overflow: hidden;
  min-width: 0;
}

.bar-track--thin {
  height: 6px;
}

.bar-fill {
  height: 100%;
  border-radius: 0 4px 4px 0;
  min-width: 2px;
  transition: width var(--transition-slow);
}

.bar-fill--step-0 { background: var(--ramp-0); }
.bar-fill--step-1 { background: var(--ramp-1); }
.bar-fill--step-2 { background: var(--ramp-2); }
.bar-fill--step-3 { background: var(--ramp-3); }
.bar-fill--accent { background: var(--color-primary); }

.pipeline-row--lost .bar-track { background: var(--lost-track); }
.bar-fill--lost { background: var(--lost-fill); }

/* ── Pipeline ── */
.pipeline-summary {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.pipeline-summary-value {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
  letter-spacing: -0.01em;
}

.pipeline-summary-label {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.pipeline {
  list-style: none;
  display: grid;
  grid-template-columns: minmax(6.5rem, auto) 1fr 2.5rem minmax(5.5rem, auto);
  column-gap: var(--space-3);
  row-gap: var(--space-4);
  align-items: center;
  margin: 0;
  padding: 0;
}

.pipeline-row {
  display: contents;
}

.pipeline-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-800);
}

.pipeline-row--lost .pipeline-name,
.pipeline-row--lost .pipeline-count,
.pipeline-row--lost .pipeline-value {
  color: var(--color-neutral-500);
}

.pipeline-count,
.pipeline-value {
  font-size: var(--text-sm);
  font-variant-numeric: tabular-nums;
  text-align: right;
  color: var(--color-neutral-600);
}

.pipeline-value {
  font-weight: var(--font-medium);
  color: var(--color-neutral-900);
}

/* ── Row buttons (expiring quotes) ── */
.row-button {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: var(--space-3);
  border: 0;
  border-radius: var(--radius-lg);
  background: transparent;
  text-align: left;
  font: inherit;
  color: inherit;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.row-button:hover,
.row-button:focus-visible {
  background: var(--color-neutral-50);
  outline: none;
}

.urgency-stripe {
  width: 3px;
  align-self: stretch;
  border-radius: var(--radius-full);
  flex-shrink: 0;
}

.urgency-stripe--critical { background: var(--color-danger); }
.urgency-stripe--warning { background: var(--color-warning); }
.urgency-stripe--notice { background: var(--color-neutral-300); }

.row-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.row-title {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-900);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.row-sub {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.row-side {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-shrink: 0;
}

.row-amount {
  font-size: var(--text-sm);
  font-variant-numeric: tabular-nums;
  color: var(--color-neutral-700);
}

.urgency-pill {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  padding: 2px var(--space-2);
  border-radius: var(--radius-full);
  min-width: 4.5rem;
  text-align: center;
}

.urgency-pill--critical {
  background: var(--color-danger-light);
  color: var(--color-danger-dark);
}

.urgency-pill--warning {
  background: var(--color-warning-light);
  color: var(--color-warning-dark);
}

.urgency-pill--notice {
  background: var(--color-neutral-100);
  color: var(--color-neutral-600);
}

/* ── Recent quotes table ── */
.quote-row {
  cursor: pointer;
}

.quote-customer {
  max-width: 14rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.num {
  font-variant-numeric: tabular-nums;
}

.margin {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.margin-dot {
  width: 6px;
  height: 6px;
  border-radius: var(--radius-full);
  background: var(--color-success);
}

.margin--low .margin-dot {
  background: var(--color-warning);
}

/* ── Sales reps ── */
.rep-list,
.customer-list,
.activity {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
}

.rep-list {
  gap: var(--space-4);
}

.rep-row {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: var(--color-neutral-100);
  color: var(--color-neutral-700);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}

.rep-body,
.customer-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.rep-head,
.customer-head,
.rep-foot {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: var(--space-3);
}

.rep-name,
.customer-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-900);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rep-revenue,
.customer-value {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  font-variant-numeric: tabular-nums;
  color: var(--color-neutral-900);
  flex-shrink: 0;
}

.rep-foot,
.customer-quotes {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.rep-winrate {
  font-weight: var(--font-medium);
  color: var(--color-success-dark);
}

.rep-winrate--low {
  color: var(--color-warning-dark);
}

/* ── Top customers ── */
.customer-list {
  gap: var(--space-4);
}

.customer-row {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
}

.customer-rank {
  width: 24px;
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  font-variant-numeric: tabular-nums;
  color: var(--color-neutral-400);
  text-align: right;
  padding-top: 2px;
  flex-shrink: 0;
}

/* ── Activity ── */
.activity {
  gap: 0;
}

.activity-item {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-3) 0;
  border-top: 1px solid var(--color-neutral-100);
}

.activity-item:first-child {
  border-top: 0;
  padding-top: 0;
}

.activity-item:last-child {
  padding-bottom: 0;
}

.activity-icon {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full);
  background: var(--color-neutral-100);
  color: var(--color-neutral-600);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-body {
  min-width: 0;
  flex: 1;
}

.activity-text {
  font-size: var(--text-sm);
  color: var(--color-neutral-800);
  line-height: var(--leading-normal);
}

.activity-meta {
  display: flex;
  gap: var(--space-2);
  margin-top: 2px;
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

/* ── Responsive ── */
@media (max-width: 1100px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .kpi-tile--hero {
    grid-column: 1 / -1;
  }
}

@media (max-width: 640px) {
  .kpi-grid {
    grid-template-columns: 1fr;
  }

  .dash-header {
    flex-direction: column;
  }

  .pipeline {
    grid-template-columns: 1fr auto;
    row-gap: var(--space-2);
  }

  .pipeline-name {
    grid-column: 1;
  }

  .pipeline-value {
    grid-column: 2;
  }

  .pipeline .bar-track {
    grid-column: 1 / -1;
  }

  .pipeline-count {
    display: none;
  }

  .pipeline-row + .pipeline-row .pipeline-name {
    margin-top: var(--space-2);
  }

  .row-side {
    flex-direction: column;
    align-items: flex-end;
    gap: var(--space-1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .bar-fill,
  .attention-item,
  .row-button {
    transition: none;
  }

  .spinning {
    animation: none;
  }
}
</style>
