<script setup lang="ts">
import { computed, onMounted } from 'vue'
import {
  DollarSign,
  Target,
  FileText,
  ScrollText,
  TrendingUp,
  TrendingDown,
  ArrowUpRight,
  Clock,
  User,
  Building2,
  CheckCircle2,
  Send,
  FilePlus,
  Handshake,
  PhoneCall,
  CalendarCheck,
  AlertTriangle,
  RefreshCw,
  BarChart3,
} from 'lucide-vue-next'
import { useDashboardStore } from '@/stores/dashboard'

const dashStore = useDashboardStore()
onMounted(() => dashStore.fetchAll())

interface KpiCard {
  label: string
  value: string
  subtitle: string
  change: number
  icon: typeof DollarSign
  iconBg: string
  iconColor: string
}

interface PipelineStage {
  name: string
  count: number
  value: number
  color: string
}

interface Quote {
  number: string
  customer: string
  status: 'draft' | 'pending-approval' | 'approved' | 'sent' | 'accepted' | 'declined' | 'expired'
  total: number
  margin: number
  date: string
}

interface Activity {
  id: number
  icon: typeof DollarSign
  iconBg: string
  iconColor: string
  description: string
  timestamp: string
  user: string
}

interface TopCustomer {
  name: string
  value: number
  opportunities: number
}

const kpiCards = computed<KpiCard[]>(() => {
  const k = dashStore.kpis as Record<string, { value?: number; totalValue?: number; expiringIn30Days?: number }> | undefined
  return [
    {
      label: 'Total Revenue',
      value: k?.totalRevenue ? formatCurrency(k.totalRevenue.value ?? 0) : 'SAR —',
      subtitle: 'From accepted quotes',
      change: 0,
      icon: DollarSign,
      iconBg: 'var(--color-success-light)',
      iconColor: 'var(--color-success)',
    },
    {
      label: 'Active Opportunities',
      value: k?.activeOpportunities ? String(k.activeOpportunities.value ?? 0) : '—',
      subtitle: 'In pipeline',
      change: 0,
      icon: Target,
      iconBg: 'var(--color-primary-light)',
      iconColor: 'var(--color-primary)',
    },
    {
      label: 'Open Quotes',
      value: k?.openQuotes ? String(k.openQuotes.value ?? 0) : '—',
      subtitle: k?.openQuotes ? formatCurrency(k.openQuotes.totalValue ?? 0) + ' total' : 'Loading...',
      change: 0,
      icon: FileText,
      iconBg: 'var(--color-warning-light)',
      iconColor: 'var(--color-warning)',
    },
    {
      label: 'Active Contracts',
      value: k?.activeContracts ? String(k.activeContracts.value ?? 0) : '—',
      subtitle: k?.activeContracts
        ? `${k.activeContracts.expiringIn30Days ?? 0} expiring in 30d`
        : 'Loading...',
      change: 0,
      icon: ScrollText,
      iconBg: 'var(--color-primary-100)',
      iconColor: 'var(--color-primary-700)',
    },
  ]
})

const stageColors: Record<string, string> = {
  qualification: 'var(--color-primary)',
  proposal: 'var(--color-warning)',
  negotiation: 'var(--color-success)',
  'closed-won': '#10b981',
  'closed-lost': 'var(--color-danger)',
}

const pipelineStages = computed<PipelineStage[]>(() => {
  const raw = dashStore.pipeline as Array<{ stage: string; count: number; value: number }>
  if (!raw || raw.length === 0) return []
  return raw.map((s) => ({
    name: s.stage.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
    count: s.count,
    value: s.value,
    color: stageColors[s.stage] ?? 'var(--color-neutral-400)',
  }))
})

const maxPipelineValue = computed(() =>
  Math.max(...pipelineStages.value.map((s) => s.value), 1),
)

function formatCurrency(value: number): string {
  if (value >= 1_000_000) return `SAR ${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1_000) return `SAR ${(value / 1_000).toFixed(0)}K`
  return `SAR ${value.toLocaleString()}`
}

const recentQuotes: Quote[] = [
  { number: 'QT-2026-0147', customer: 'Saudi Aramco', status: 'sent', total: 485_000, margin: 32, date: '2026-02-21' },
  { number: 'QT-2026-0146', customer: 'SABIC', status: 'accepted', total: 1_230_000, margin: 28, date: '2026-02-20' },
  { number: 'QT-2026-0145', customer: 'Riyadh Municipality', status: 'pending-approval', total: 320_000, margin: 35, date: '2026-02-19' },
  { number: 'QT-2026-0144', customer: 'King Fahd Medical City', status: 'draft', total: 178_500, margin: 41, date: '2026-02-18' },
  { number: 'QT-2026-0143', customer: 'NEOM', status: 'declined', total: 2_100_000, margin: 22, date: '2026-02-17' },
]

const statusConfig: Record<Quote['status'], { label: string; class: string }> = {
  draft: { label: 'Draft', class: 'badge-neutral' },
  'pending-approval': { label: 'Pending', class: 'badge-warning' },
  approved: { label: 'Approved', class: 'badge-primary' },
  sent: { label: 'Sent', class: 'badge-primary' },
  accepted: { label: 'Accepted', class: 'badge-success' },
  declined: { label: 'Declined', class: 'badge-danger' },
  expired: { label: 'Expired', class: 'badge-neutral' },
}

function formatQuoteTotal(value: number): string {
  return `SAR ${value.toLocaleString()}`
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const activities: Activity[] = [
  {
    id: 1,
    icon: CheckCircle2,
    iconBg: 'var(--color-success-light)',
    iconColor: 'var(--color-success)',
    description: 'Closed deal with Saudi Aramco — CCTV upgrade for Eastern Province facilities',
    timestamp: '25 min ago',
    user: 'Khalid Al-Rashid',
  },
  {
    id: 2,
    icon: Send,
    iconBg: 'var(--color-primary-light)',
    iconColor: 'var(--color-primary)',
    description: 'Sent quotation QT-2026-0147 to Saudi Aramco for access control system',
    timestamp: '1 hour ago',
    user: 'Noura Al-Dosari',
  },
  {
    id: 3,
    icon: FilePlus,
    iconBg: 'var(--color-warning-light)',
    iconColor: 'var(--color-warning)',
    description: 'Created new opportunity: Riyadh Municipality — perimeter security',
    timestamp: '2 hours ago',
    user: 'Ahmed bin Saleh',
  },
  {
    id: 4,
    icon: Handshake,
    iconBg: 'var(--color-success-light)',
    iconColor: 'var(--color-success)',
    description: 'Contract CON-2026-089 signed with SABIC for annual maintenance',
    timestamp: '4 hours ago',
    user: 'Khalid Al-Rashid',
  },
  {
    id: 5,
    icon: PhoneCall,
    iconBg: 'var(--color-neutral-100)',
    iconColor: 'var(--color-neutral-600)',
    description: 'Follow-up call with NEOM project manager regarding proposal revisions',
    timestamp: '5 hours ago',
    user: 'Noura Al-Dosari',
  },
  {
    id: 6,
    icon: CalendarCheck,
    iconBg: 'var(--color-primary-light)',
    iconColor: 'var(--color-primary)',
    description: 'Scheduled site survey at King Fahd Medical City for intrusion detection',
    timestamp: 'Yesterday',
    user: 'Ahmed bin Saleh',
  },
]

const topCustomers = computed<TopCustomer[]>(() => {
  const raw = dashStore.topCustomers as Array<{
    companyName: string
    totalRevenue: number
    quoteCount: number
  }>
  return (raw ?? []).map((c) => ({
    name: c.companyName,
    value: c.totalRevenue ?? 0,
    opportunities: c.quoteCount ?? 0,
  }))
})

const maxCustomerValue = computed(() =>
  Math.max(...topCustomers.value.map((c) => c.value), 1),
)

// ── Expiring Quotes ──────────────────────────────────────────
interface ExpiringQuote {
  number: string
  customer: string
  total: number
  expiresIn: number
}

const expiringQuotes: ExpiringQuote[] = [
  { number: 'QT-2026-0147', customer: 'Saudi Aramco', total: 485_000, expiresIn: 2 },
  { number: 'QT-2026-0144', customer: 'King Faisal Specialist Hospital', total: 178_500, expiresIn: 5 },
  { number: 'QT-2026-0143', customer: 'NEOM', total: 2_100_000, expiresIn: 7 },
]

function urgencyClass(days: number): string {
  if (days <= 2) return 'urgency-critical'
  if (days <= 5) return 'urgency-warning'
  return 'urgency-notice'
}

// ── Sales Rep Performance ────────────────────────────────────
interface SalesRep {
  name: string
  quotesCreated: number
  quotesWon: number
  revenue: number
  winRate: number
}

const salesReps: SalesRep[] = [
  { name: 'Khalid Al-Rashid', quotesCreated: 12, quotesWon: 8, revenue: 1_850_000, winRate: 67 },
  { name: 'Noura Al-Dosari', quotesCreated: 9, quotesWon: 5, revenue: 1_220_000, winRate: 56 },
  { name: 'Ahmed bin Saleh', quotesCreated: 7, quotesWon: 4, revenue: 680_000, winRate: 57 },
  { name: 'Fatima Al-Harbi', quotesCreated: 6, quotesWon: 2, revenue: 340_000, winRate: 33 },
]

const maxRepRevenue = computed(() =>
  Math.max(...salesReps.map((r) => r.revenue), 1),
)
</script>

<template>
  <div class="dashboard">
    <!-- Page Header -->
    <div class="dashboard-header">
      <div>
        <h1 class="dashboard-title">Dashboard</h1>
        <p class="dashboard-subtitle">Welcome back. Here's what's happening today.</p>
      </div>
      <div class="header-date">
        <Clock :size="16" />
        <span>{{ new Date().toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }) }}</span>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="kpi-grid">
      <div v-for="card in kpiCards" :key="card.label" class="kpi-card card">
        <div class="kpi-top">
          <div>
            <p class="kpi-label">{{ card.label }}</p>
            <p class="kpi-value">{{ card.value }}</p>
            <p class="kpi-subtitle">{{ card.subtitle }}</p>
          </div>
          <div class="kpi-icon" :style="{ backgroundColor: card.iconBg, color: card.iconColor }">
            <component :is="card.icon" :size="22" />
          </div>
        </div>
        <div class="kpi-change" :class="card.change >= 0 ? 'positive' : 'negative'">
          <TrendingUp v-if="card.change >= 0" :size="14" />
          <TrendingDown v-else :size="14" />
          <span>{{ Math.abs(card.change) }}% vs last quarter</span>
        </div>
      </div>
    </div>

    <!-- Main Content Grid -->
    <div class="content-grid">
      <!-- Left Column -->
      <div class="column-left">
        <!-- Pipeline Overview -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Pipeline Overview</h3>
            <span class="card-header-meta">{{ pipelineStages.reduce((s, stage) => s + stage.count, 0) }} opportunities</span>
          </div>
          <div class="card-body">
            <div class="pipeline-stages">
              <div v-for="stage in pipelineStages" :key="stage.name" class="pipeline-row">
                <div class="pipeline-info">
                  <span class="pipeline-name">{{ stage.name }}</span>
                  <span class="pipeline-meta">{{ stage.count }} deals &middot; {{ formatCurrency(stage.value) }}</span>
                </div>
                <div class="pipeline-bar-track">
                  <div
                    class="pipeline-bar-fill"
                    :style="{
                      width: `${(stage.value / maxPipelineValue) * 100}%`,
                      backgroundColor: stage.color,
                    }"
                  />
                </div>
              </div>
            </div>
            <div class="pipeline-total">
              <span>Total Pipeline Value</span>
              <strong>{{ formatCurrency(pipelineStages.reduce((s, stage) => s + stage.value, 0)) }}</strong>
            </div>
          </div>
        </div>

        <!-- Expiring Quotes Alert -->
        <div v-if="expiringQuotes.length" class="card card--expiring">
          <div class="card-header">
            <h3 class="card-title expiring-title">
              <AlertTriangle :size="16" class="expiring-icon" />
              Expiring Soon
            </h3>
            <span class="card-header-meta">{{ expiringQuotes.length }} quotes</span>
          </div>
          <div class="card-body">
            <div class="expiring-list">
              <div v-for="eq in expiringQuotes" :key="eq.number" class="expiring-item">
                <div class="expiring-info">
                  <span class="expiring-number text-mono">{{ eq.number }}</span>
                  <span class="expiring-customer">{{ eq.customer }}</span>
                </div>
                <div class="expiring-right">
                  <span class="expiring-total text-mono">{{ formatCurrency(eq.total) }}</span>
                  <span :class="['expiring-badge', urgencyClass(eq.expiresIn)]">
                    {{ eq.expiresIn === 1 ? 'Tomorrow' : `${eq.expiresIn} days` }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Quotes -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Recent Quotes</h3>
            <button class="btn btn-ghost btn-sm">
              View All
              <ArrowUpRight :size="14" />
            </button>
          </div>
          <div class="card-body card-body--flush">
            <div class="table-container table-container--embedded">
              <table class="table">
                <thead>
                  <tr>
                    <th>Quote #</th>
                    <th>Customer</th>
                    <th>Status</th>
                    <th class="text-right">Total</th>
                    <th class="text-right">Margin</th>
                    <th>Date</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="quote in recentQuotes" :key="quote.number">
                    <td class="text-mono font-medium">{{ quote.number }}</td>
                    <td>
                      <div class="customer-cell">
                        <Building2 :size="14" class="customer-icon" />
                        {{ quote.customer }}
                      </div>
                    </td>
                    <td>
                      <span class="badge badge-dot" :class="statusConfig[quote.status].class">
                        {{ statusConfig[quote.status].label }}
                      </span>
                    </td>
                    <td class="text-right text-mono">{{ formatQuoteTotal(quote.total) }}</td>
                    <td class="text-right">
                      <span :class="quote.margin >= 30 ? 'margin-good' : 'margin-low'">
                        {{ quote.margin }}%
                      </span>
                    </td>
                    <td class="text-muted">{{ formatDate(quote.date) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column -->
      <div class="column-right">
        <!-- Activity Feed -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Recent Activity</h3>
            <button class="btn btn-ghost btn-sm">
              View All
              <ArrowUpRight :size="14" />
            </button>
          </div>
          <div class="card-body">
            <div class="activity-feed">
              <div v-for="activity in activities" :key="activity.id" class="activity-item">
                <div class="activity-icon" :style="{ backgroundColor: activity.iconBg, color: activity.iconColor }">
                  <component :is="activity.icon" :size="16" />
                </div>
                <div class="activity-content">
                  <p class="activity-description">{{ activity.description }}</p>
                  <div class="activity-meta">
                    <User :size="12" />
                    <span>{{ activity.user }}</span>
                    <span class="activity-dot">&middot;</span>
                    <span>{{ activity.timestamp }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Sales Rep Performance -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">
              <BarChart3 :size="16" />
              Sales Rep Performance
            </h3>
            <span class="card-header-meta">This quarter</span>
          </div>
          <div class="card-body">
            <div class="rep-list">
              <div v-for="rep in salesReps" :key="rep.name" class="rep-row">
                <div class="rep-info">
                  <div class="rep-avatar">{{ rep.name.charAt(0) }}</div>
                  <div class="rep-details">
                    <span class="rep-name">{{ rep.name }}</span>
                    <div class="rep-stats">
                      <span>{{ rep.quotesWon }}/{{ rep.quotesCreated }} won</span>
                      <span class="rep-dot">&middot;</span>
                      <span>{{ formatCurrency(rep.revenue) }}</span>
                    </div>
                  </div>
                </div>
                <div class="rep-right">
                  <div class="rep-winrate-bar">
                    <div class="rep-winrate-fill" :style="{ width: `${rep.winRate}%` }" :class="rep.winRate >= 50 ? 'fill-good' : 'fill-low'" />
                  </div>
                  <span class="rep-winrate-label" :class="rep.winRate >= 50 ? 'margin-good' : 'margin-low'">{{ rep.winRate }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Top Customers -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Top Customers</h3>
            <span class="card-header-meta">By opportunity value</span>
          </div>
          <div class="card-body">
            <div class="top-customers">
              <div v-for="(customer, index) in topCustomers" :key="customer.name" class="customer-row">
                <div class="customer-rank">{{ index + 1 }}</div>
                <div class="customer-details">
                  <div class="customer-name-row">
                    <span class="customer-name">{{ customer.name }}</span>
                    <span class="customer-value">{{ formatCurrency(customer.value) }}</span>
                  </div>
                  <div class="customer-bar-track">
                    <div
                      class="customer-bar-fill"
                      :style="{ width: `${(customer.value / maxCustomerValue) * 100}%` }"
                    />
                  </div>
                  <span class="customer-opps">{{ customer.opportunities }} opportunities</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 1400px;
}

/* Header */
.dashboard-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.dashboard-title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
}

.dashboard-subtitle {
  font-size: var(--text-sm);
  color: var(--color-neutral-500);
  margin-top: var(--space-1);
}

.header-date {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  color: var(--color-neutral-500);
  padding: var(--space-2) var(--space-3);
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-lg);
}

/* KPI Grid */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: var(--space-5);
  margin-bottom: var(--space-6);
}

.kpi-card {
  padding: var(--space-5);
}

.kpi-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-4);
}

.kpi-label {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-500);
  margin-bottom: var(--space-1);
}

.kpi-value {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
  line-height: var(--leading-tight);
}

.kpi-subtitle {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  margin-top: var(--space-1);
}

.kpi-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.kpi-change {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  font-weight: var(--font-medium);
  padding: 2px var(--space-2);
  border-radius: var(--radius-full);
}

.kpi-change.positive {
  color: var(--color-success-dark);
  background-color: var(--color-success-light);
}

.kpi-change.negative {
  color: var(--color-danger-dark);
  background-color: var(--color-danger-light);
}

/* Content Grid */
.content-grid {
  display: grid;
  grid-template-columns: 3fr 2fr;
  gap: var(--space-6);
}

.column-left,
.column-right {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* Pipeline */
.pipeline-stages {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.pipeline-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.pipeline-info {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.pipeline-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-800);
}

.pipeline-meta {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.pipeline-bar-track {
  width: 100%;
  height: 8px;
  background-color: var(--color-neutral-100);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.pipeline-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.pipeline-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--space-4);
  margin-top: var(--space-4);
  border-top: 1px solid var(--color-neutral-100);
  font-size: var(--text-sm);
  color: var(--color-neutral-600);
}

.pipeline-total strong {
  font-weight: var(--font-bold);
  color: var(--color-neutral-900);
}

/* Quotes Table */
.card-body--flush {
  padding-left: 0;
  padding-right: 0;
  padding-bottom: 0;
}

.table-container--embedded {
  border: none;
  border-radius: 0;
  box-shadow: none;
}

.customer-cell {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.customer-icon {
  color: var(--color-neutral-400);
  flex-shrink: 0;
}

.margin-good {
  color: var(--color-success);
  font-weight: var(--font-semibold);
}

.margin-low {
  color: var(--color-warning-dark);
  font-weight: var(--font-semibold);
}

/* Activity Feed */
.activity-feed {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.activity-item {
  display: flex;
  gap: var(--space-3);
  padding: var(--space-3) 0;
  position: relative;
}

.activity-item:not(:last-child) {
  border-bottom: 1px solid var(--color-neutral-100);
}

.activity-icon {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-description {
  font-size: var(--text-sm);
  color: var(--color-neutral-700);
  line-height: var(--leading-normal);
  margin-bottom: var(--space-1);
}

.activity-meta {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
}

.activity-dot {
  margin: 0 2px;
}

/* Top Customers */
.top-customers {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.customer-row {
  display: flex;
  gap: var(--space-3);
  align-items: flex-start;
}

.customer-rank {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background-color: var(--color-neutral-100);
  color: var(--color-neutral-600);
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 1px;
}

.customer-details {
  flex: 1;
  min-width: 0;
}

.customer-name-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: var(--space-2);
}

.customer-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-800);
}

.customer-value {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
}

.customer-bar-track {
  width: 100%;
  height: 6px;
  background-color: var(--color-neutral-100);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: var(--space-1);
}

.customer-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  background: linear-gradient(90deg, var(--color-primary), var(--color-primary-600));
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.customer-opps {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
}

/* Card header meta text */
.card-header-meta {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  font-weight: var(--font-medium);
}

/* Expiring Quotes */
.card--expiring { border-left: 3px solid var(--color-warning); }
.expiring-title { display: flex; align-items: center; gap: var(--space-2); }
.expiring-icon { color: var(--color-warning); }

.expiring-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.expiring-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) 0;
}

.expiring-item:not(:last-child) {
  border-bottom: 1px solid var(--color-neutral-100);
}

.expiring-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.expiring-number {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.expiring-customer {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.expiring-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.expiring-total {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-700);
}

.expiring-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 1px 8px;
  border-radius: var(--radius-full);
  white-space: nowrap;
}

.urgency-critical {
  background: var(--color-danger-light);
  color: var(--color-danger);
}

.urgency-warning {
  background: var(--color-warning-light);
  color: var(--color-warning-dark);
}

.urgency-notice {
  background: var(--color-neutral-100);
  color: var(--color-neutral-600);
}

/* Sales Rep Performance */
.rep-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.rep-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
}

.rep-info {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex: 1;
  min-width: 0;
}

.rep-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: var(--color-primary-light, #eff6ff);
  color: var(--color-primary);
  font-size: var(--text-sm);
  font-weight: var(--font-bold);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.rep-details {
  min-width: 0;
}

.rep-name {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-800);
  display: block;
}

.rep-stats {
  font-size: var(--text-xs);
  color: var(--color-neutral-400);
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.rep-dot { margin: 0 2px; }

.rep-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.rep-winrate-bar {
  width: 60px;
  height: 6px;
  background: var(--color-neutral-100);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.rep-winrate-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width 0.6s ease;
}

.rep-winrate-fill.fill-good { background: var(--color-success); }
.rep-winrate-fill.fill-low { background: var(--color-warning); }

.rep-winrate-label {
  font-size: var(--text-xs);
  font-weight: var(--font-semibold);
  width: 32px;
  text-align: right;
}

/* Responsive */
@media (max-width: 1400px) {
  .kpi-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 1200px) {
  .kpi-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .content-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .dashboard-header {
    flex-direction: column;
    gap: var(--space-3);
  }

  .kpi-grid {
    grid-template-columns: 1fr;
  }
}
</style>
