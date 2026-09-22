<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowUp, ArrowDown, Trash2, Plus, Upload } from 'lucide-vue-next'
import { documentsService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import type { Document as CrmDocument } from '@/types'
import type { QuoteAppendix } from '@/services/workflowDtos'
import { useAuthStore } from '@/stores/auth'
const props = defineProps<{ modelValue: QuoteAppendix[]; customerId: string; disabled: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [QuoteAppendix[]]; busy: [boolean] }>()
const auth = useAuthStore()
const documents = ref<CrmDocument[]>([])
const search = ref('')
const scope = ref('customer')
const loading = ref(false)
const busy = ref(false)
const error = ref('')
const selectedFile = ref<File>()
const uploadName = ref('')
const uploadInput = ref<HTMLInputElement>()
const extensions = /\.(pdf|png|jpe?g|docx?|xlsx?)$/i
const matches = computed(() => documents.value.filter(doc => {
  const customerLinks = (doc as CrmDocument & {links?: {entityType:string;entityId:string}[]}).links?.filter(link => link.entityType === 'customer') ?? []
  return extensions.test(doc.fileName ?? '') && doc.name.toLowerCase().includes(search.value.toLowerCase()) && (scope.value === 'all' || !customerLinks.length || customerLinks.some(link => link.entityId === props.customerId))
}))
async function load() {
  if (!auth.can('documents:read')) return
  loading.value = true
  try { documents.value = await allPages(documentsService.list) } catch(e) { error.value = errorMessage(e) } finally { loading.value = false }
}
onMounted(load)
function update(items: QuoteAppendix[]) { if (!props.disabled && !busy.value) emit('update:modelValue', items) }
function rename(index:number, event:Event) { update(props.modelValue.map((item,i) => i === index ? {...item,label:(event.target as HTMLInputElement).value} : item)) }
function move(index:number, direction:number) { const items = [...props.modelValue]; const [item] = items.splice(index,1); if (item) items.splice(index+direction,0,item); update(items) }
async function pin(doc:CrmDocument) {
  if (props.modelValue.length >= 20) throw new Error('A quote can include up to 20 documents.')
  const versions = (await documentsService.versions(doc.id)).data
  const version = versions.find(item => item.version === doc.version)
  if (!version) throw new Error('This document version is unavailable. Reload the library and try again.')
  if (props.modelValue.some(item => item.documentVersionId === version.id)) throw new Error('This version is already included.')
  if (props.modelValue.reduce((sum,item) => sum+item.fileSize,0)+version.fileSize > 100*1024*1024) throw new Error('Combined appendix files must not exceed 100 MB.')
  emit('update:modelValue',[...props.modelValue,{ documentId:doc.id,documentVersionId:version.id,label:doc.name,documentName:doc.name,version:version.version,fileName:version.fileName,fileType:version.fileType,fileSize:version.fileSize }])
}
async function run(action:()=>Promise<void>) {
  if (props.disabled || busy.value) return
  busy.value = true; emit('busy',true); error.value = ''
  try { await action() } catch(e) { error.value = errorMessage(e) } finally { busy.value = false; emit('busy',false) }
}
async function upload() {
  await run(async () => {
    const file = selectedFile.value
    if (!file || !extensions.test(file.name)) throw new Error('Choose a PDF, PNG, JPG, Word or Excel file.')
    if (!file.size || file.size>50*1024*1024) throw new Error('Choose a nonempty file of at most 50 MB.')
    if (props.modelValue.reduce((sum,item) => sum+item.fileSize,0)+file.size > 100*1024*1024) throw new Error('Combined appendix files must not exceed 100 MB.')
    const data = new FormData(); data.append('file',file); data.append('name',uploadName.value.trim() || file.name); data.append('category','quote'); data.append('documentType','technical')
    const doc = (await documentsService.upload(data)).data
    documents.value.unshift(doc)
    selectedFile.value = undefined; uploadName.value = ''; if (uploadInput.value) uploadInput.value.value = ''
    await pin(doc)
  })
}
</script>
<template>
  <section class="appendices-panel" aria-labelledby="appendices-heading">
    <h2 id="appendices-heading">Documents included in the PDF</h2>
    <p class="form-hint">These files become labeled appendix pages after the quotation. The export includes an index and continuous page numbers. Save the draft to keep your selection.</p>
    <p v-if="error" class="form-error" role="alert">{{ error }}</p>
    <p v-if="!modelValue.length">No supporting documents selected.</p>
    <ol v-else class="appendix-list">
      <li v-for="(item,index) in modelValue" :key="item.documentVersionId" class="appendix-row">
        <div class="appendix-info"><label :for="`appendix-label-${index}`">Appendix {{ index+1 }} label</label><input :id="`appendix-label-${index}`" class="input form-input" :value="item.label" :disabled="disabled || busy" maxlength="255" required @input="rename(index,$event)" /><p class="form-hint">{{ item.fileName }} · Version {{ item.version }} · {{ (item.fileSize/1024).toFixed(0) }} KB</p></div>
        <div class="appendix-actions"><button class="btn btn-sm" :disabled="disabled || busy || index === 0" :aria-label="`Move appendix ${index+1} up`" @click="move(index,-1)"><ArrowUp :size="16" /></button><button class="btn btn-sm" :disabled="disabled || busy || index === modelValue.length-1" :aria-label="`Move appendix ${index+1} down`" @click="move(index,1)"><ArrowDown :size="16" /></button><button class="btn btn-sm" :disabled="disabled || busy" :aria-label="`Remove appendix ${index+1}`" @click="update(modelValue.filter((_,i) => i !== index))"><Trash2 :size="16" /></button></div>
      </li>
    </ol>
    <p v-if="modelValue.length" class="form-hint">Versions are fixed when selected. New library uploads do not replace these files. Remove and reselect a document in a draft to use its latest version.</p>
    <template v-if="!disabled && auth.can('documents:read')">
      <h3>Select from Documents</h3>
      <div class="appendix-filters"><label>Search documents<input v-model="search" type="search" class="input form-input" /></label><label>Show<select v-model="scope" class="select form-select"><option value="customer">This customer's and general documents</option><option value="all">All accessible documents</option></select></label></div>
      <p v-if="loading" role="status">Loading documents…</p>
      <p v-else-if="!matches.length">No supported documents match. Upload one below or change the search.</p>
      <ul v-else class="library-list"><li v-for="doc in matches" :key="doc.id"><div><strong>{{ doc.name }}</strong><p class="form-hint">{{ doc.fileName }} · Version {{ doc.version }}</p></div><button class="btn btn-sm" :disabled="busy || modelValue.length >= 20 || modelValue.some(item => item.documentId === doc.id && item.version === doc.version)" :aria-label="`Include ${doc.name}`" @click="run(()=>pin(doc))"><Plus :size="14" /> Include</button></li></ul>
      <button v-if="error" class="btn btn-sm" :disabled="loading || busy" @click="load">Reload library</button>
      <form v-if="auth.can('documents:create')" class="appendix-upload" @submit.prevent="upload">
        <h3>Upload a new document</h3><p class="form-hint">PDF, PNG/JPG, Word (DOC/DOCX) or Excel (XLS/XLSX). Up to 50 MB per file, 20 files and 100 MB per quote. Word/Excel pages use the file's print layout.</p>
        <label>Document name<input v-model="uploadName" class="input form-input" maxlength="255" :disabled="busy" /></label><label>File<input ref="uploadInput" class="file-input form-input" type="file" accept=".pdf,.png,.jpg,.jpeg,.doc,.docx,.xls,.xlsx" required :disabled="busy" @change="selectedFile = ($event.target as HTMLInputElement).files?.[0]" /></label><button class="btn btn-primary" :disabled="busy || !selectedFile || modelValue.length >= 20"><Upload :size="16" /> {{ busy ? 'Adding document…' : 'Upload and include' }}</button>
      </form>
    </template>
  </section>
</template>
<style scoped>
.appendices-panel { max-width: 1000px; }
h2 { font-size: var(--text-lg); margin-bottom: var(--space-2); }
h3 { font-size: var(--text-base); margin-top: var(--space-6); margin-bottom: var(--space-3); }
label { display: grid; gap: var(--space-2); font-size: var(--text-sm); }
.appendix-list, .library-list { list-style: none; padding: 0; margin-block: var(--space-4); }
.appendix-row, .library-list li { display: flex; align-items: center; gap: var(--space-4); padding-block: var(--space-3); border-bottom: 1px solid var(--color-neutral-200); }
.appendix-info, .library-list li > div { flex: 1; min-width: 0; overflow-wrap: anywhere; }
.appendix-actions { display: flex; gap: var(--space-2); }
.appendix-filters { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-4); }
.library-list { max-height: 360px; overflow-y: auto; }
.appendix-upload { display: grid; gap: var(--space-3); max-width: 640px; }
.appendix-upload .btn { justify-self: start; }
@media(max-width: 700px) { .appendix-filters { grid-template-columns: 1fr; } .appendix-row { align-items: stretch; flex-direction: column; } .appendix-actions { justify-content: flex-end; } }
</style>
