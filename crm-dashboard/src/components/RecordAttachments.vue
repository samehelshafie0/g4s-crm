<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Paperclip, Download, X, Upload } from 'lucide-vue-next'
import { documentsService } from '@/services'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import type { Document as CrmDocument } from '@/types'

/**
 * Files that belong to one record — the supplier's quotation PDF on a supplier
 * quote, the signed order on a purchase order, the customer's paperwork on a
 * quote. The file is uploaded to the document library once and linked here, so
 * the same document can sit on the quote and the order it produced.
 */
const props = defineProps<{
  entityType: 'quote' | 'purchase-order' | 'supplier-quote' | 'project'
  entityId: string
  /** Shown in the uploaded document's name so the library stays readable */
  entityLabel?: string
  disabled?: boolean
}>()

const auth = useAuthStore()
const documents = ref<CrmDocument[]>([])
const loading = ref(false)
const busy = ref(false)
const error = ref('')

const canAttach = computed(() => auth.can('documents:create') && auth.can('documents:update') && !props.disabled)
const canRead = computed(() => auth.can('documents:read'))

async function load() {
  if (!props.entityId || !canRead.value) return
  loading.value = true
  error.value = ''
  try {
    const res = await documentsService.list({ entityId: props.entityId, limit: 100 })
    documents.value = res.data ?? []
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}

watch(() => props.entityId, load, { immediate: true })

async function attach(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || busy.value) return
  busy.value = true
  error.value = ''
  try {
    const form = new FormData()
    form.append('file', file)
    form.append('name', props.entityLabel ? `${props.entityLabel} — ${file.name}` : file.name)
    form.append('category', 'general')
    form.append('documentType', 'terms')
    const uploaded = await documentsService.upload(form)
    await documentsService.addLink(uploaded.data.id, props.entityType, props.entityId)
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}

async function download(doc: CrmDocument) {
  error.value = ''
  try {
    await documentsService.download(doc.id, doc.fileName ?? doc.name)
  } catch (e) {
    error.value = errorMessage(e)
  }
}

async function unlink(doc: CrmDocument) {
  const link = (doc.links ?? []).find(l => l.entityId === props.entityId)
  if (!link || busy.value) return
  busy.value = true
  error.value = ''
  try {
    await documentsService.deleteLink(doc.id, link.id)
    await load()
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="att">
    <header class="att-head">
      <h4 class="att-title"><Paperclip :size="14" /> Attached files</h4>
      <label v-if="canAttach" class="btn btn-sm att-upload">
        <input type="file" class="att-input" :disabled="busy" @change="attach" />
        <Upload :size="13" />
        {{ busy ? 'Uploading…' : 'Attach a file' }}
      </label>
    </header>

    <p v-if="error" class="att-error" role="alert">{{ error }}</p>

    <p v-if="!canRead" class="att-empty">You do not have access to documents.</p>
    <p v-else-if="loading" class="att-empty">Loading…</p>
    <ul v-else-if="documents.length" class="att-list">
      <li v-for="doc in documents" :key="doc.id" class="att-item">
        <button type="button" class="att-name" :title="`Download ${doc.fileName}`" @click="download(doc)">
          <Download :size="13" />
          <span class="att-label">{{ doc.name }}</span>
          <span class="att-meta">{{ doc.fileName }} · {{ doc.fileSize }}</span>
        </button>
        <button
          v-if="canAttach"
          type="button"
          class="att-remove"
          :disabled="busy"
          :aria-label="`Remove ${doc.name} from this record`"
          @click="unlink(doc)"
        >
          <X :size="13" />
        </button>
      </li>
    </ul>
    <p v-else class="att-empty">
      No files yet. Attach the supplier's quotation or price list so it stays with this record.
    </p>
  </section>
</template>

<style scoped>
.att { display: flex; flex-direction: column; gap: var(--space-2); }
.att-head { display: flex; align-items: center; justify-content: space-between; gap: var(--space-3); }
.att-title { display: inline-flex; align-items: center; gap: var(--space-2); font-size: var(--text-sm); font-weight: var(--font-semibold); color: var(--color-neutral-800); }
.att-upload { position: relative; overflow: hidden; cursor: pointer; }
.att-input { position: absolute; inset: 0; opacity: 0; cursor: pointer; }
.att-error { font-size: var(--text-xs); color: var(--color-danger-dark); background: var(--color-danger-light); padding: var(--space-2); border-radius: var(--radius-md); }
.att-empty { font-size: var(--text-xs); color: var(--color-neutral-500); }
.att-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 2px; }
.att-item { display: flex; align-items: center; gap: var(--space-1); }
.att-name { display: flex; align-items: center; gap: var(--space-2); flex: 1; min-width: 0; padding: var(--space-2); border: 0; background: none; font: inherit; color: inherit; text-align: left; cursor: pointer; border-radius: var(--radius-md); }
.att-name:hover, .att-name:focus-visible { background: var(--color-neutral-50); outline: none; }
.att-label { font-size: var(--text-sm); color: var(--color-neutral-800); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.att-meta { font-size: var(--text-xs); color: var(--color-neutral-500); white-space: nowrap; flex-shrink: 0; }
.att-remove { display: flex; align-items: center; justify-content: center; width: 26px; height: 26px; border: 0; background: none; color: var(--color-neutral-400); border-radius: var(--radius-md); cursor: pointer; flex-shrink: 0; }
.att-remove:hover:not(:disabled) { background: var(--color-danger-light); color: var(--color-danger-dark); }
</style>
