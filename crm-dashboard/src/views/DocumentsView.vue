<script setup lang="ts">
import DocumentDetails from '@/components/DocumentDetails.vue'
import { useAuthStore } from '@/stores/auth'
import { documentsService, customersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import { onMounted, ref, computed } from 'vue'
import {
  Plus,
  Search,
  Trash2,
  X,
  Download,
  Upload,
  FileText,
  FileSpreadsheet,
  FileImage,
  File,
  FileArchive,
  Tag,
} from 'lucide-vue-next'
import type { Document, DocumentCategory, DocumentType } from '@/types'
import { useDocumentsStore } from '@/stores/documents'

function uid(): string {
  return Math.random().toString(36).slice(2, 11)
}

const auth = useAuthStore()
const detailId = ref('')
const detailsOpen = ref(false)
const loadError = ref('')
async function reloadDocuments() { documents.value = await allPages(documentsService.list) }
onMounted(async () => { try { await reloadDocuments(); if (auth.can('customers:read')) customerOptions.value = (await customersService.lookup()).data.map(c => ({id:c.id, name:c.companyName})) } catch(e) { loadError.value = errorMessage(e) } })
function openDetails(id:string) { detailId.value = id; detailsOpen.value = true }

const categoryLabels: Record<DocumentCategory, string> = {
  contract: 'Contract',
  quote: 'Quote',
  general: 'General',
  compliance: 'Compliance',
  legal: 'Legal',
}

const categoryBadge: Record<DocumentCategory, string> = {
  contract: 'badge-primary',
  quote: 'badge-info',
  general: 'badge-gray',
  compliance: 'badge-warning',
  legal: 'badge-danger',
}

const typeLabels: Record<DocumentType, string> = {
  terms: 'Terms & Conditions',
  delivery: 'Delivery Note',
  technical: 'Technical Spec',
  warranty: 'Warranty',
  sla: 'SLA',
}

const typeBadge: Record<DocumentType, string> = {
  terms: 'badge-primary',
  delivery: 'badge-success',
  technical: 'badge-info',
  warranty: 'badge-warning',
  sla: 'badge-gray',
}

const fileTypeIcons: Record<string, typeof FileText> = {
  pdf: FileText,
  docx: FileText,
  xlsx: FileSpreadsheet,
  png: FileImage,
  jpg: FileImage,
  zip: FileArchive,
}

function getFileIcon(fileType: string) {
  return fileTypeIcons[fileType] ?? File
}

const documents = ref<Document[]>([])
const searchQuery = ref('')
const filterCategory = ref<DocumentCategory | ''>('')
const filterType = ref<DocumentType | ''>('')

const filteredDocuments = computed(() => {
  let list = documents.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter(d =>
      d.name.toLowerCase().includes(q) ||
      d.tags.some(t => t.toLowerCase().includes(q)),
    )
  }
  if (filterCategory.value) list = list.filter(d => d.category === filterCategory.value)
  if (filterType.value) list = list.filter(d => d.documentType === filterType.value)
  return list
})

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

// Upload modal
const showUploadModal = ref(false)

interface DocForm {
  name: string
  category: DocumentCategory
  documentType: DocumentType
  tags: string
  version: string
  fileSize: string
  fileType: string
  linkedEntityName: string
}

const defaultForm = (): DocForm => ({
  name: '',
  category: 'general',
  documentType: 'technical',
  tags: '',
  version: '1.0',
  fileSize: '',
  fileType: 'pdf',
  linkedEntityName: '',
})

const form = ref<DocForm>(defaultForm())

function openUploadModal() {
  selectedFile.value = null
  form.value = defaultForm()
  showUploadModal.value = true
}

const selectedFile = ref<globalThis.File | null>(null)
const uploading = ref(false)
const customerOptions = ref<{id:string;name:string}[]>([])
function chooseFile(event: Event) { selectedFile.value = (event.target as HTMLInputElement).files?.[0] ?? null }
async function uploadDocument() {
 if (!selectedFile.value || uploading.value) return
 uploading.value = true
 try {
   const data = new FormData(); data.append('file', selectedFile.value); data.append('name', form.value.name); data.append('category', form.value.category); data.append('documentType', form.value.documentType); data.append('tags', form.value.tags)
   const result = await documentsService.upload(data)
   if (auth.can('documents:update') && form.value.linkedEntityName) await documentsService.addLink(result.data.id, 'customer', form.value.linkedEntityName)
   documents.value = await allPages(documentsService.list); showUploadModal.value = false; selectedFile.value = null
 } catch (e) { window.alert(errorMessage(e)) } finally { uploading.value = false }
}
async function deleteDocument(id: string) {
 try { await documentsService.delete(id); documents.value = documents.value.filter(d => d.id !== id) } catch (e) { window.alert(errorMessage(e)) }
}
async function downloadDocument(doc: Document) { try { await documentsService.download(doc.id, doc.fileName || doc.name) } catch (e) { window.alert(errorMessage(e)) } }
</script>

<template>
  <div class="documents-page">
    <p v-if="loadError" class="form-error" role="alert">{{ loadError }}</p>
    <DocumentDetails v-if="detailsOpen" :id="detailId" :key="detailId" v-model:open="detailsOpen" @changed="reloadDocuments().catch(e => loadError = errorMessage(e))" />
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Documents</h1>
        <p class="page-header-subtitle">{{ filteredDocuments.length }} document{{ filteredDocuments.length !== 1 ? 's' : '' }}</p>
      </div>
      <button v-if="auth.can('documents:create')" class="btn btn-primary" @click="openUploadModal">
        <Upload :size="18" />
        Upload Document
      </button>
    </div>

    <!-- Filter Toolbar -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="search-input toolbar-search">
          <Search :size="18" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Search by name or tag…"
          />
        </div>
        <select v-model="filterCategory" class="form-select toolbar-select">
          <option value="">All Categories</option>
          <option v-for="(label, key) in categoryLabels" :key="key" :value="key">{{ label }}</option>
        </select>
        <select v-model="filterType" class="form-select toolbar-select">
          <option value="">All Types</option>
          <option v-for="(label, key) in typeLabels" :key="key" :value="key">{{ label }}</option>
        </select>
      </div>
    </div>

    <!-- Documents Table -->
    <div v-if="filteredDocuments.length" class="table-container">
      <table class="table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Category</th>
            <th>Type</th>
            <th>Tags</th>
            <th>Version</th>
            <th>Size</th>
            <th>Linked To</th>
            <th>Uploaded By</th>
            <th>Date</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="doc in filteredDocuments" :key="doc.id">
            <td>
              <div class="doc-name-cell">
                <component :is="getFileIcon(doc.fileType)" :size="18" class="doc-file-icon" />
                <div>
                  <div class="font-medium text-dark">{{ doc.name }}</div>
                  <div class="text-xs text-muted">.{{ doc.fileType }}</div>
                </div>
              </div>
            </td>
            <td><span :class="['badge', categoryBadge[doc.category]]">{{ categoryLabels[doc.category] }}</span></td>
            <td><span :class="['badge', typeBadge[doc.documentType]]">{{ typeLabels[doc.documentType] }}</span></td>
            <td>
              <div class="tag-list">
                <span v-for="tag in doc.tags" :key="tag" class="tag-badge">{{ tag }}</span>
              </div>
            </td>
            <td class="text-mono">v{{ doc.version }}</td>
            <td class="whitespace-nowrap">{{ doc.fileSize }}</td>
            <td>
              <span v-if="doc.linkedEntities.length">{{ doc.linkedEntities.map(e => e.name).join(', ') }}</span>
              <span v-else class="text-muted">—</span>
            </td>
            <td>{{ doc.uploadedBy }}</td>
            <td class="whitespace-nowrap">{{ formatDate(doc.createdAt) }}</td>
            <td>
              <div class="table-actions">
                <button class="btn btn-ghost btn-icon btn-sm" title="Download" @click="downloadDocument(doc)">
                  <Download :size="14" />
                </button>
                <button class="btn btn-sm" @click="openDetails(doc.id)">Details & versions</button>
                <button v-if="auth.can('documents:delete')" class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deleteDocument(doc.id)">
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="empty-state">
      <FileText :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No documents found</h3>
      <p class="empty-state-text">Try adjusting your filters or upload a new document.</p>
    </div>

    <!-- Upload Document Modal -->
    <Teleport to="body">
      <div v-if="showUploadModal" class="modal-backdrop" @click.self="showUploadModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">Upload Document</h2>
            <button class="modal-close" @click="showUploadModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">Document Name <span class="required">*</span></label>
              <input v-model="form.name" type="text" class="form-input" placeholder="Document name" />
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Category</label>
                <select v-model="form.category" class="form-select">
                  <option v-for="(label, key) in categoryLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">Document Type</label>
                <select v-model="form.documentType" class="form-select">
                  <option v-for="(label, key) in typeLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Tags (comma-separated)</label>
                <input v-model="form.tags" type="text" class="form-input" placeholder="e.g. security, cctv, proposal" />
              </div>
              <div class="form-group">
                <label class="form-label">Version</label>
                <input v-model="form.version" readonly type="text" class="form-input" placeholder="1.0" />
              </div>
            </div>

            <label>File (up to 50 MB)<input type="file" class="file-input" @change="chooseFile" /></label>
            <label>Link to customer<select v-model="form.linkedEntityName" :disabled="!auth.can('documents:update')" class="select"><option value="">No customer link</option><option v-for="customer in customerOptions" :key="customer.id" :value="customer.id">{{ customer.name }}</option></select></label>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showUploadModal = false">Cancel</button>
            <button
              class="btn btn-primary"
              :disabled="!form.name || !selectedFile || uploading"
              @click="uploadDocument"
            >
              <Upload :size="14" />
              Upload
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.documents-page {
  padding: var(--space-6);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  flex-wrap: wrap;
}

.toolbar-search {
  flex: 1;
  min-width: 220px;
}

.toolbar-select {
  width: 180px;
  flex-shrink: 0;
}

.doc-name-cell {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.doc-file-icon {
  color: var(--color-neutral-400);
  flex-shrink: 0;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tag-badge {
  display: inline-block;
  padding: 1px 8px;
  background: var(--color-neutral-100);
  color: var(--color-neutral-600);
  border-radius: var(--radius-full, 9999px);
  font-size: 11px;
  font-weight: var(--font-medium);
  line-height: 1.6;
}

.text-muted { color: var(--color-neutral-400); }
.text-xs { font-size: var(--text-xs); }

/* File upload area */
.file-upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-8, 2rem) var(--space-6);
  border: 2px dashed var(--color-neutral-300);
  border-radius: var(--radius-lg);
  background: var(--color-neutral-50);
  cursor: pointer;
  transition: border-color 0.15s;
}

.file-upload-area:hover {
  border-color: var(--color-primary);
}

.upload-icon {
  color: var(--color-neutral-400);
}

.upload-text {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-700);
  margin: 0;
}

.upload-hint {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  margin: 0;
}

@media (max-width: 768px) {
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-select { width: 100%; }
}
</style>
