<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppDialog from '@/components/shared/AppDialog.vue'
import { documentsService, http } from '@/services'
import { allPages } from '@/services/collections'
import type { ApiResponse } from '@/services/http'
import type { DocumentVersion } from '@/services/workflowDtos'
import type { Document as CrmDocument } from '@/types'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
type Link = {id:string;entityId:string;entityType:string;entityName:string}
type DetailedDocument = CrmDocument & {links:Link[]}
const props = defineProps<{id:string;open:boolean}>()
const emit = defineEmits<{'update:open':[boolean];changed:[]}>()
const auth = useAuthStore()
const doc = ref<DetailedDocument>()
const versions = ref<DocumentVersion[]>([])
const tab = ref('versions')
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const success = ref('')
const file = ref<File | null>(null)
const fileInput = ref<HTMLInputElement>()
const metadata = ref({name:'',category:'general' as CrmDocument['category'],documentType:'technical' as CrmDocument['documentType'],tags:''})
const linkTypes = [ ['customer','customers','Customer'], ['opportunity','opportunities','Opportunity'], ['quote','quotes','Quote'], ['contract','contracts','Contract'], ['project','projects','Project'], ['product','products','Product'], ['purchase-order','procurement/purchase-orders','Purchase order'] ] as const
const allowedLinks = computed(() => linkTypes.filter(([,path]) => auth.can(`${path.split('/')[0]}:read`)))
const linkType = ref<string>(allowedLinks.value[0]?.[0] ?? '')
const query = ref('')
const results = ref<{id:string;name:string}[]>([])
const entityId = ref('')
async function load() {
 loading.value = true; error.value = ''
 try { const [record, history] = await Promise.all([documentsService.get(props.id),documentsService.versions(props.id)]); doc.value = record.data as DetailedDocument; versions.value = history.data; metadata.value = {name:record.data.name,category:record.data.category,documentType:record.data.documentType,tags:record.data.tags.join(', ')} }
 catch(e) { error.value = errorMessage(e) } finally { loading.value = false }
}
onMounted(load)
async function run(operation:()=>Promise<unknown>, message:string) {
 if (busy.value) return
 busy.value = true; error.value = ''; success.value = ''
 try { await operation(); await load(); emit('changed'); if (!error.value) success.value = message }
 catch(e) { error.value = errorMessage(e) } finally { busy.value = false }
}
function chooseFile(event:Event) { file.value = (event.target as HTMLInputElement).files?.[0] ?? null }
async function uploadVersion() {
 if (!file.value || busy.value) return
 if (!file.value.size || file.value.size > 50 * 1024 * 1024) { error.value = 'Choose a nonempty file of at most 50 MB.'; return }
 const data = new FormData(); data.append('file',file.value)
 await run(async () => { await documentsService.uploadVersion(props.id,data); file.value = null; if (fileInput.value) fileInput.value.value = '' },'New version uploaded. Previous versions remain available.')
}
async function download(version:DocumentVersion) { try { await documentsService.download(props.id,version.fileName,version.id) } catch(e) { error.value = errorMessage(e) } }
async function searchLinks() {
 const path = allowedLinks.value.find(([type]) => type === linkType.value)?.[1]
 if (!path || busy.value) return
 busy.value = true; error.value = ''; results.value = []; entityId.value = ''
 try { const rows = await allPages(async params => (await http.get<ApiResponse<Record<string,unknown>[]>>('/'+path,{params})).data); const term = query.value.trim().toLowerCase(); results.value = rows.map(item => ({id:String(item.id),name:String(item.companyName ?? item.title ?? item.name ?? item.quoteNumber ?? item.poNumber ?? item.id)})).filter(item => item.name.toLowerCase().includes(term)).slice(0,100) }
 catch(e) { error.value = errorMessage(e) } finally { busy.value = false }
}
async function link() { if (!entityId.value) return; await run(()=>documentsService.addLink(props.id,linkType.value,entityId.value),'Record linked.'); entityId.value = '' }
</script>
<template>
 <AppDialog :open="open" :title="doc?.name ?? 'Document details'" :busy="busy" @update:open="emit('update:open',$event)">
  <p v-if="loading" role="status">Loading document…</p><p v-if="error" class="form-error" role="alert">{{ error }}</p><p v-if="success" role="status">{{ success }}</p>
  <template v-if="doc">
   <nav class="doc-tabs" aria-label="Document sections"><button v-for="item in ['versions','details','links']" :key="item" class="btn btn-sm" :aria-pressed="tab === item" @click="tab = item">{{ item === 'versions' ? 'Version history' : item === 'details' ? 'Details' : 'Linked records' }}</button></nav>
   <section v-if="tab === 'versions'"><p class="form-hint">Current version: {{ doc.version }}. Download any saved version.</p><div class="table-container"><table class="table"><thead><tr><th>Version / file</th><th>Uploaded</th><th>Size</th><th>Action</th></tr></thead><tbody><tr v-for="version in versions" :key="version.id"><td>v{{ version.version }} · {{ version.fileName }}</td><td>{{ new Date(version.createdAt).toLocaleString() }}</td><td>{{ (version.fileSize/1024).toFixed(1) }} KB</td><td><button class="btn btn-sm" :aria-label="`Download version ${version.version}`" @click="download(version)">Download</button></td></tr></tbody></table></div><form v-if="auth.can('documents:update')" class="doc-form" @submit.prevent="uploadVersion"><label>Upload new version<input ref="fileInput" class="file-input form-input" type="file" required :disabled="busy" @change="chooseFile" /></label><p class="form-hint">Maximum 50 MB. The existing file will remain in the version history.</p><button class="btn btn-primary" :disabled="busy || !file">{{ busy ? 'Uploading…' : 'Upload new version' }}</button></form></section>
   <form v-else-if="tab === 'details'" class="doc-form" @submit.prevent="run(()=>documentsService.update(id,{...metadata,tags:metadata.tags.split(',').map(t=>t.trim()).filter(Boolean)}),'Document details saved.')"><fieldset :disabled="busy || !auth.can('documents:update')"><label>Name<input v-model.trim="metadata.name" class="input form-input" required maxlength="255" /></label><label>Category<select v-model="metadata.category" class="select form-select"><option v-for="value in ['contract','quote','general','compliance','legal']" :key="value">{{ value }}</option></select></label><label>Document type<select v-model="metadata.documentType" class="select form-select"><option v-for="value in ['terms','delivery','technical','warranty','sla']" :key="value">{{ value }}</option></select></label><label>Tags (comma separated)<input v-model="metadata.tags" class="input form-input" /></label></fieldset><button v-if="auth.can('documents:update')" class="btn btn-primary" :disabled="busy">Save details</button></form>
   <section v-else><ul class="document-links"><li v-for="item in doc.links" :key="item.id"><span>{{ item.entityName || item.entityId }} · {{ item.entityType }}</span><button v-if="auth.can('documents:update')" class="btn btn-sm" :disabled="busy" :aria-label="`Unlink ${item.entityName}`" @click="run(()=>documentsService.deleteLink(id,item.id),'Record unlinked.')">Unlink</button></li></ul><p v-if="!doc.links?.length">No linked records.</p><form v-if="auth.can('documents:update') && allowedLinks.length" class="doc-form" @submit.prevent="link"><label>Record type<select v-model="linkType" class="select form-select" :disabled="busy" @change="results = []; entityId = ''"><option v-for="[type,,label] in allowedLinks" :key="type" :value="type">{{ label }}</option></select></label><label>Search records<input v-model="query" class="input form-input" type="search" :disabled="busy" /></label><button type="button" class="btn btn-sm" :disabled="busy" @click="searchLinks">Find records</button><label>Record<select v-model="entityId" class="select form-select" required :disabled="busy"><option value="">Select a record</option><option v-for="item in results" :key="item.id" :value="item.id">{{ item.name }}</option></select></label><p class="form-hint">Search returns up to 100 records. Refine the search if needed.</p><button class="btn btn-primary" :disabled="busy || !entityId">Link record</button></form></section>
  </template>
 </AppDialog>
</template>
<style scoped>
.doc-tabs { display: flex; gap: var(--space-2); margin-bottom: var(--space-5); flex-wrap: wrap; }
.doc-tabs [aria-pressed="true"] { background: var(--color-neutral-200); }
.doc-form { display: grid; gap: var(--space-4); margin-top: var(--space-6); }
fieldset { border: 0; padding: 0; display: grid; gap: var(--space-4); }
label { display: grid; gap: var(--space-2); font-size: var(--text-sm); }
.document-links { list-style: none; padding: 0; }
.document-links li { display: flex; justify-content: space-between; align-items: center; gap: var(--space-3); padding-block: var(--space-3); border-bottom: 1px solid var(--color-neutral-200); }
.document-links span { overflow-wrap: anywhere; }
</style>
