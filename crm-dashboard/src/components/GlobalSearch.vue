<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from 'lucide-vue-next'
import { customersService, opportunitiesService, quotesService, productsService } from '@/services'
import { useAuthStore } from '@/stores/auth'

interface Hit { key: string; group: string; label: string; detail: string; to: string }

const router = useRouter()
const auth = useAuthStore()
const query = ref('')
const hits = ref<Hit[]>([])
const open = ref(false)
const loading = ref(false)
const failed = ref(false)
const active = ref(0)
const input = ref<HTMLInputElement | null>(null)
const isMac = navigator.platform.toLowerCase().includes('mac')

// Each source is searched only when the user may read it, five results each.
const sources = computed(() => [
  auth.can('customers:read') && (async (q: string): Promise<Hit[]> =>
    (await customersService.list({ q, limit: 5 })).data.map(c => ({ key: `c${c.id}`, group: 'Customers', label: c.companyName, detail: [c.region, c.sector].filter(Boolean).join(' · '), to: `/customers?open=${c.id}` }))),
  auth.can('opportunities:read') && (async (q: string): Promise<Hit[]> =>
    (await opportunitiesService.list({ q, limit: 5 })).data.map(o => ({ key: `o${o.id}`, group: 'Opportunities', label: `${o.opportunityNumber} · ${o.title}`, detail: o.customerName, to: `/opportunities/${o.id}` }))),
  auth.can('quotes:read') && (async (q: string): Promise<Hit[]> =>
    (await quotesService.list({ q, limit: 5 })).data.map(qt => ({ key: `q${qt.id}`, group: 'Quotes', label: `${qt.quoteNumber} v${qt.version}`, detail: qt.customerName, to: `/quotes/${qt.id}/builder` }))),
  auth.can('products:read') && (async (q: string): Promise<Hit[]> =>
    (await productsService.list({ q, limit: 5 })).data.map(p => ({ key: `p${p.id}`, group: 'Products', label: p.sku, detail: p.name, to: `/products?open=${p.id}` }))),
].filter((s): s is (q: string) => Promise<Hit[]> => Boolean(s)))

let timer: number | undefined
let generation = 0
watch(query, (value) => {
  window.clearTimeout(timer)
  const q = value.trim()
  if (q.length < 2) { hits.value = []; loading.value = false; return }
  loading.value = true
  timer = window.setTimeout(async () => {
    const mine = ++generation
    const settled = await Promise.allSettled(sources.value.map(search => search(q)))
    if (mine !== generation) return
    hits.value = settled.flatMap(r => (r.status === 'fulfilled' ? r.value : []))
    failed.value = settled.some(r => r.status === 'rejected')
    active.value = 0
    loading.value = false
  }, 250)
})

const groups = computed(() => {
  const order: string[] = []
  const byGroup: Record<string, Hit[]> = {}
  for (const hit of hits.value) {
    if (!byGroup[hit.group]) { byGroup[hit.group] = []; order.push(hit.group) }
    byGroup[hit.group]!.push(hit)
  }
  return order.map(name => ({ name, items: byGroup[name]! }))
})

function go(hit?: Hit) {
  if (!hit) return
  router.push(hit.to)
  query.value = ''
  open.value = false
  input.value?.blur()
}

function onKey(event: KeyboardEvent) {
  if (event.key === 'ArrowDown') { event.preventDefault(); active.value = Math.min(active.value + 1, hits.value.length - 1) }
  else if (event.key === 'ArrowUp') { event.preventDefault(); active.value = Math.max(active.value - 1, 0) }
  else if (event.key === 'Enter') { event.preventDefault(); go(hits.value[active.value]) }
  else if (event.key === 'Escape') { open.value = false; input.value?.blur() }
}

function shortcut(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    input.value?.focus()
  }
}
onMounted(() => window.addEventListener('keydown', shortcut))
onBeforeUnmount(() => { window.removeEventListener('keydown', shortcut); window.clearTimeout(timer) })

function closeSoon() { window.setTimeout(() => { open.value = false }, 150) }
</script>

<template>
  <div class="global-search" role="search">
    <Search :size="16" class="search-icon" />
    <input
      ref="input"
      v-model="query"
      type="search"
      class="search-input"
      placeholder="Search customers, opportunities, quotes, products..."
      aria-label="Search the CRM"
      autocomplete="off"
      role="combobox"
      :aria-expanded="open && query.trim().length >= 2"
      aria-controls="global-search-results"
      @focus="open = true"
      @input="open = true"
      @blur="closeSoon"
      @keydown="onKey"
    />
    <kbd class="search-kbd">{{ isMac ? '⌘K' : 'Ctrl K' }}</kbd>

    <div v-if="open && query.trim().length >= 2" id="global-search-results" class="search-results" role="listbox">
      <p v-if="loading" class="search-note">Searching…</p>
      <template v-else>
        <div v-for="group in groups" :key="group.name" class="search-group">
          <p class="search-group-title">{{ group.name }}</p>
          <button
            v-for="hit in group.items"
            :key="hit.key"
            type="button"
            role="option"
            class="search-hit"
            :class="{ 'search-hit--active': hits[active]?.key === hit.key }"
            :aria-selected="hits[active]?.key === hit.key"
            @mousedown.prevent="go(hit)"
          >
            <span class="search-hit-label">{{ hit.label }}</span>
            <span v-if="hit.detail" class="search-hit-detail">{{ hit.detail }}</span>
          </button>
        </div>
        <p v-if="!hits.length" class="search-note">No matches for “{{ query.trim() }}”.</p>
        <p v-if="failed" class="search-note search-note--warn">Some results could not be loaded.</p>
      </template>
    </div>
  </div>
</template>

<style scoped>
.global-search { position: relative; display: flex; align-items: center; width: 100%; }
.search-icon { position: absolute; left: var(--space-3); color: var(--color-neutral-400); pointer-events: none; }
.search-input {
  width: 100%; height: 38px; padding: 0 56px 0 36px; border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md); background: var(--color-neutral-50); font-size: var(--text-sm);
  color: var(--color-neutral-800); outline: none; transition: all var(--transition-fast);
}
.search-input:focus { background: var(--content-surface); border-color: var(--color-primary); box-shadow: 0 0 0 3px var(--color-primary-200); }
.search-input::placeholder { color: var(--color-neutral-400); }
.search-input::-webkit-search-cancel-button { display: none; }
.search-kbd {
  position: absolute; right: 8px; padding: 2px 6px; font-size: 0.625rem; font-family: inherit; background: var(--content-surface);
  border: 1px solid var(--color-neutral-200); border-radius: 4px; color: var(--color-neutral-400); pointer-events: none;
}
.search-results {
  position: absolute; top: calc(100% + 6px); left: 0; right: 0; z-index: 60; max-height: 420px; overflow-y: auto;
  background: var(--content-surface); border: 1px solid var(--color-neutral-200); border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg); padding: var(--space-2);
}
.search-group + .search-group { margin-top: var(--space-2); }
.search-group-title { font-size: var(--text-xs); font-weight: 600; text-transform: uppercase; letter-spacing: 0.04em; color: var(--color-neutral-500); padding: var(--space-1) var(--space-2); }
.search-hit { display: flex; flex-direction: column; align-items: flex-start; width: 100%; text-align: left; padding: var(--space-2); border-radius: var(--radius-md); background: none; border: 0; cursor: pointer; color: inherit; }
.search-hit:hover, .search-hit--active { background: var(--color-neutral-100); }
.search-hit-label { font-size: var(--text-sm); font-weight: 500; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.search-hit-detail { font-size: var(--text-xs); color: var(--color-neutral-500); max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.search-note { font-size: var(--text-sm); color: var(--color-neutral-500); padding: var(--space-2); }
.search-note--warn { color: var(--color-warning-dark, #b45309); }
@media (max-width: 768px) { .search-kbd { display: none; } }
</style>
