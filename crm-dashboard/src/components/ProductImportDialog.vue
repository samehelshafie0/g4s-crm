<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Upload, FileSpreadsheet, FileText, AlertTriangle, CheckCircle2, RotateCcw } from 'lucide-vue-next'
import AppDialog from '@/components/shared/AppDialog.vue'
import { productsService, manufacturersService, exchangeRatesService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import {
  parseFile,
  parseCSVText,
  downloadTemplate,
  getMappedValue,
  getMappedNumber,
  getMappedInt,
  type ParseResult,
} from '@/utils/fileParser'
import type { Currency, Manufacturer } from '@/types'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ 'update:open': [value: boolean]; imported: [] }>()

type Step = 'upload' | 'review' | 'done'
const step = ref<Step>('upload')
const busy = ref(false)
const error = ref('')

// ── Parsing ──────────────────────────────────────────────────
const fileName = ref('')
const pasteText = ref('')
const parsed = ref<ParseResult | null>(null)

interface Draft {
  include: boolean
  sku: string
  name: string
  description: string
  unitCost: number
  qty: number
  leadTimeDays: number
}

const drafts = ref<Draft[]>([])

function buildDrafts(result: ParseResult) {
  const { headers, mappedColumns } = result
  drafts.value = result.rows
    .map(row => ({
      include: true,
      sku: getMappedValue(row, 'sku', headers, mappedColumns).trim(),
      name: getMappedValue(row, 'name', headers, mappedColumns).trim(),
      description: getMappedValue(row, 'name', headers, mappedColumns).trim(),
      unitCost: getMappedNumber(row, 'unitCost', headers, mappedColumns),
      qty: getMappedInt(row, 'qty', headers, mappedColumns),
      leadTimeDays: getMappedInt(row, 'leadTime', headers, mappedColumns),
    }))
    // A row with no part number or no price is not something we can cost.
    .filter(d => d.sku || d.name)
    .map(d => ({ ...d, include: Boolean(d.sku && d.name && d.unitCost > 0) }))
  parsed.value = result
  step.value = 'review'
}

async function readFile(file?: File) {
  if (!file) return
  error.value = ''
  busy.value = true
  fileName.value = file.name
  const out = await parseFile(file)
  busy.value = false
  if ('error' in out) {
    error.value = `${out.error} If this is a scanned PDF, paste the rows below or use the Excel template.`
    return
  }
  buildDrafts(out.result)
}

function onFileChosen(event: Event) {
  const input = event.target as HTMLInputElement
  void readFile(input.files?.[0])
  input.value = ''
}

function usePastedText() {
  if (!pasteText.value.trim()) return
  const result = parseCSVText(pasteText.value)
  if (!result.rows.length) {
    error.value = 'No rows could be read from the pasted text. Include a header row such as: Part Number, Description, Price.'
    return
  }
  fileName.value = fileName.value || 'Pasted rows'
  buildDrafts(result)
}

// ── Import settings ──────────────────────────────────────────
const manufacturers = ref<Manufacturer[]>([])
const rates = ref<Record<string, number>>({})

const settings = ref({
  vendorName: '',
  manufacturerId: '',
  categoryId: '',
  source: 'vendor-catalog' as 'vendor-catalog' | 'supplier-quote',
  productType: 'import' as 'import' | 'local',
  originCurrency: 'USD' as Currency,
  fxRate: 3.75,
  freightPercent: 4,
  customsPercent: 5,
  clearancePercent: 1.5,
  targetMarginPercent: 30,
  updateExisting: true,
})

const categories = computed(
  () => manufacturers.value.find(m => m.id === settings.value.manufacturerId)?.categories ?? [],
)

watch(() => settings.value.manufacturerId, () => { settings.value.categoryId = '' })

watch(() => settings.value.originCurrency, currency => {
  if (currency === 'SAR') { settings.value.fxRate = 1; settings.value.productType = 'local'; return }
  settings.value.productType = 'import'
  const rate = rates.value[currency]
  if (rate) settings.value.fxRate = rate
})

async function loadReferenceData() {
  try {
    const [mfrs, fx] = await Promise.all([
      allPages(manufacturersService.list),
      exchangeRatesService.list(),
    ])
    manufacturers.value = mfrs
    rates.value = Object.fromEntries(fx.data.map(r => [r.fromCurrency, r.currentRate]))
    const rate = rates.value[settings.value.originCurrency]
    if (rate) settings.value.fxRate = rate
  } catch (e) {
    error.value = errorMessage(e)
  }
}

watch(() => props.open, open => { if (open) { reset(); void loadReferenceData() } })

// ── Costing preview, mirroring the server's calculation ──────
function landedCost(unitCost: number): number {
  const base = settings.value.productType === 'local' ? unitCost : unitCost * settings.value.fxRate
  const s = settings.value
  return base * (1 + (s.freightPercent + s.customsPercent + s.clearancePercent) / 100)
}

function sellingPrice(unitCost: number): number {
  return landedCost(unitCost) / (1 - settings.value.targetMarginPercent / 100)
}

function money(value: number): string {
  return value.toLocaleString('en-GB', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const selected = computed(() => drafts.value.filter(d => d.include && d.sku && d.name))
const unusable = computed(() => drafts.value.filter(d => !d.sku || !d.name || d.unitCost <= 0).length)

// ── Import ───────────────────────────────────────────────────
interface Outcome { sku: string; status: string; reason?: string; landedCostSAR?: number; sellingPrice?: number }
const result = ref<{ created: number; updated: number; skipped: number; failed: number; rows: Outcome[] } | null>(null)

async function runImport() {
  if (!selected.value.length || !settings.value.vendorName.trim()) return
  busy.value = true
  error.value = ''
  try {
    const response = await productsService.importProducts({
      vendorName: settings.value.vendorName.trim(),
      manufacturerId: settings.value.manufacturerId || undefined,
      categoryId: settings.value.categoryId || undefined,
      sourceRef: fileName.value,
      source: settings.value.source,
      productType: settings.value.productType,
      originCurrency: settings.value.originCurrency,
      fxRate: settings.value.fxRate,
      freightPercent: settings.value.freightPercent,
      customsPercent: settings.value.customsPercent,
      clearancePercent: settings.value.clearancePercent,
      targetMarginPercent: settings.value.targetMarginPercent,
      updateExisting: settings.value.updateExisting,
      rows: selected.value.map(d => ({
        sku: d.sku,
        name: d.name.slice(0, 255),
        description: d.description,
        unitCost: d.unitCost,
        qty: d.qty,
        leadTimeDays: d.leadTimeDays,
      })),
    })
    result.value = response.data
    step.value = 'done'
    emit('imported')
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

function reset() {
  step.value = 'upload'
  error.value = ''
  fileName.value = ''
  pasteText.value = ''
  parsed.value = null
  drafts.value = []
  result.value = null
}

function close() {
  if (!busy.value) emit('update:open', false)
}
</script>

<template>
  <AppDialog
    :open="open"
    :busy="busy"
    title="Import products from a vendor file"
    @update:open="emit('update:open', $event)"
  >
    <p v-if="error" class="imp-error" role="alert"><AlertTriangle :size="15" /> {{ error }}</p>

    <!-- Step 1: choose a file -->
    <div v-if="step === 'upload'" class="imp-step">
      <p class="imp-lead">
        Upload the price list or quotation the vendor sent. Excel, CSV and text-based PDF are read
        directly; the part number, description and price columns are detected automatically.
      </p>

      <label class="imp-drop">
        <input type="file" accept=".xlsx,.xls,.xlsb,.xlsm,.ods,.csv,.tsv,.txt,.pdf" class="imp-file" @change="onFileChosen" />
        <Upload :size="22" />
        <span class="imp-drop-main">{{ busy ? 'Reading the file…' : 'Choose a vendor file' }}</span>
        <span class="imp-drop-sub">XLSX · XLS · CSV · PDF</span>
      </label>

      <div class="imp-alt">
        <span class="imp-alt-label">Or paste rows copied from the vendor's sheet</span>
        <textarea
          id="import-paste"
          v-model="pasteText"
          class="form-textarea"
          rows="4"
          placeholder="Part Number, Description, Price&#10;DS-2CD2143G2-I, 4 MP dome camera, 78.00"
        />
        <div class="imp-alt-actions">
          <button type="button" class="btn btn-sm" :disabled="!pasteText.trim()" @click="usePastedText">Read pasted rows</button>
          <button type="button" class="btn btn-ghost btn-sm" @click="downloadTemplate('product-catalog')">
            <FileSpreadsheet :size="14" /> Download the Excel template
          </button>
        </div>
      </div>
    </div>

    <!-- Step 2: settings and review -->
    <div v-else-if="step === 'review'" class="imp-step">
      <div class="imp-source">
        <FileText :size="15" />
        <span>{{ fileName }}</span>
        <span class="imp-source-meta">{{ drafts.length }} rows read</span>
        <button type="button" class="btn btn-ghost btn-sm" @click="reset"><RotateCcw :size="13" /> Choose another file</button>
      </div>

      <div class="imp-grid">
        <label class="imp-field">Vendor or supplier
          <input v-model.trim="settings.vendorName" class="input form-input" maxlength="255" required placeholder="Ingram Micro" />
        </label>
        <label class="imp-field">This file is a
          <select v-model="settings.source" class="select form-select">
            <option value="vendor-catalog">Price list</option>
            <option value="supplier-quote">Quotation</option>
          </select>
        </label>
        <label class="imp-field">Manufacturer
          <select v-model="settings.manufacturerId" class="select form-select">
            <option value="">Not set</option>
            <option v-for="m in manufacturers" :key="m.id" :value="m.id">{{ m.name }}</option>
          </select>
        </label>
        <label class="imp-field">Category
          <select v-model="settings.categoryId" class="select form-select" :disabled="!categories.length">
            <option value="">Not set</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
        </label>
        <label class="imp-field">Prices are in
          <select v-model="settings.originCurrency" class="select form-select">
            <option v-for="cur in ['SAR', 'USD', 'EUR', 'GBP', 'AED', 'CNY']" :key="cur" :value="cur">{{ cur }}</option>
          </select>
        </label>
        <label class="imp-field">Rate to SAR
          <input v-model.number="settings.fxRate" class="input form-input" type="number" min="0.000001" step="0.000001" :disabled="settings.originCurrency === 'SAR'" />
        </label>
        <label class="imp-field">Freight %
          <input v-model.number="settings.freightPercent" class="input form-input" type="number" min="0" max="100" step="0.1" />
        </label>
        <label class="imp-field">Customs %
          <input v-model.number="settings.customsPercent" class="input form-input" type="number" min="0" max="100" step="0.1" />
        </label>
        <label class="imp-field">Clearance %
          <input v-model.number="settings.clearancePercent" class="input form-input" type="number" min="0" max="100" step="0.1" />
        </label>
        <label class="imp-field">Target margin %
          <input v-model.number="settings.targetMarginPercent" class="input form-input" type="number" min="0" max="99" step="0.1" />
        </label>
        <label class="imp-check">
          <input v-model="settings.updateExisting" type="checkbox" />
          Refresh part numbers already in the catalog
        </label>
      </div>

      <p v-if="unusable" class="imp-note">
        {{ unusable }} of {{ drafts.length }} rows have no part number, description or price and start unticked.
        Fill them in below or leave them out.
      </p>

      <div class="table-container imp-table">
        <table class="table">
          <thead>
            <tr>
              <th class="imp-tick"><span class="sr-only">Include</span></th>
              <th>Part number</th>
              <th>Description</th>
              <th class="text-right">Cost ({{ settings.originCurrency }})</th>
              <th class="text-right">Landed (SAR)</th>
              <th class="text-right">Sell (SAR)</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in drafts" :key="index" :class="{ 'imp-row--off': !row.include }">
              <td><input v-model="row.include" type="checkbox" :aria-label="`Include ${row.sku || 'row ' + (index + 1)}`" /></td>
              <td><input v-model.trim="row.sku" class="input imp-cell imp-cell--sku" maxlength="100" /></td>
              <td><input v-model.trim="row.name" class="input imp-cell" maxlength="255" /></td>
              <td><input v-model.number="row.unitCost" class="input imp-cell imp-cell--num" type="number" min="0" step="0.01" /></td>
              <td class="text-right num">{{ money(landedCost(row.unitCost)) }}</td>
              <td class="text-right num">{{ money(sellingPrice(row.unitCost)) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Step 3: result -->
    <div v-else class="imp-step">
      <p class="imp-done"><CheckCircle2 :size="18" /> {{ result?.created }} created · {{ result?.updated }} updated · {{ result?.skipped }} skipped · {{ result?.failed }} failed</p>
      <div v-if="result?.rows.length" class="table-container imp-table">
        <table class="table">
          <thead><tr><th>Part number</th><th>Result</th><th class="text-right">Landed (SAR)</th><th class="text-right">Sell (SAR)</th></tr></thead>
          <tbody>
            <tr v-for="row in result.rows" :key="row.sku + row.status">
              <td class="text-mono">{{ row.sku }}</td>
              <td><span class="badge" :class="row.status === 'failed' ? 'badge-danger' : row.status === 'skipped' ? 'badge-neutral' : 'badge-success'">{{ row.status }}</span> <span class="imp-reason">{{ row.reason }}</span></td>
              <td class="text-right num">{{ row.landedCostSAR ? money(row.landedCostSAR) : '—' }}</td>
              <td class="text-right num">{{ row.sellingPrice ? money(row.sellingPrice) : '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template #footer>
      <button v-if="step !== 'done'" type="button" class="btn" :disabled="busy" @click="close">Cancel</button>
      <button v-if="step === 'review'" type="button" class="btn btn-primary" :disabled="busy || !selected.length || !settings.vendorName.trim()" @click="runImport">
        {{ busy ? 'Importing…' : `Import ${selected.length} product${selected.length === 1 ? '' : 's'}` }}
      </button>
      <button v-if="step === 'done'" type="button" class="btn" @click="reset">Import another file</button>
      <button v-if="step === 'done'" type="button" class="btn btn-primary" @click="close">Done</button>
    </template>
  </AppDialog>
</template>

<style scoped>
.imp-step { display: flex; flex-direction: column; gap: var(--space-4); min-width: min(78vw, 900px); }
.imp-lead { font-size: var(--text-sm); color: var(--color-neutral-600); max-width: 62ch; line-height: var(--leading-relaxed); }
.imp-error { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); color: var(--color-danger-dark); background: var(--color-danger-light); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); margin-bottom: var(--space-3); }

.imp-drop { position: relative; display: flex; flex-direction: column; align-items: center; gap: var(--space-1); padding: var(--space-8); border: 1px dashed var(--color-neutral-300); border-radius: var(--radius-lg); cursor: pointer; color: var(--color-neutral-500); transition: border-color var(--transition-fast), background-color var(--transition-fast); }
.imp-drop:hover, .imp-drop:focus-within { border-color: var(--color-primary); background: var(--color-primary-50); }
.imp-file { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
.imp-drop-main { font-size: var(--text-sm); font-weight: var(--font-medium); color: var(--color-neutral-800); }
.imp-drop-sub { font-size: var(--text-xs); }

.imp-alt { display: flex; flex-direction: column; gap: var(--space-2); }
.imp-alt-label { font-size: var(--text-xs); font-weight: var(--font-semibold); text-transform: uppercase; letter-spacing: 0.06em; color: var(--color-neutral-500); }
.imp-alt-actions { display: flex; gap: var(--space-2); flex-wrap: wrap; }

.imp-source { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); color: var(--color-neutral-700); flex-wrap: wrap; }
.imp-source-meta { color: var(--color-neutral-500); font-size: var(--text-xs); }

.imp-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: var(--space-3); }
.imp-field { display: flex; flex-direction: column; gap: 4px; font-size: var(--text-xs); font-weight: var(--font-medium); color: var(--color-neutral-600); }
.imp-check { grid-column: 1 / -1; display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); color: var(--color-neutral-700); }

.imp-note { font-size: var(--text-xs); color: var(--color-warning-dark); background: var(--color-warning-light); padding: var(--space-2) var(--space-3); border-radius: var(--radius-md); }
.imp-table { max-height: 360px; overflow: auto; }
.imp-tick { width: 32px; }
.imp-cell { width: 100%; min-width: 90px; font-size: var(--text-xs); padding: 4px 6px; }
.imp-cell--sku { font-family: var(--font-mono); min-width: 130px; }
.imp-cell--num { max-width: 110px; text-align: right; }
.imp-row--off { opacity: 0.45; }
.num { font-variant-numeric: tabular-nums; }
.imp-reason { font-size: var(--text-xs); color: var(--color-neutral-500); }
.imp-done { display: flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: var(--font-medium); color: var(--color-success-dark); }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; border: 0; }

@media (max-width: 640px) {
  .imp-step { min-width: 0; }
}
</style>
