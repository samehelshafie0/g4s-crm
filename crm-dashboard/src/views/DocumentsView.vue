<script setup lang="ts">
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

const docsStore = useDocumentsStore()
onMounted(() => docsStore.fetchDocuments())

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

const documents = ref<Document[]>([
  {
    id: uid(), name: 'Aramco Master Agreement 2025', category: 'contract', documentType: 'terms',
    tags: ['aramco', 'master-agreement', 'security'], version: '2.1', fileSize: '2.4 MB', fileType: 'pdf',
    linkedEntities: [{ type: 'Customer', id: 'c1', name: 'Saudi Aramco' }],
    uploadedBy: 'Ahmed Al-Dosari',
    createdAt: '2025-03-15T08:00:00Z', updatedAt: '2025-11-20T10:00:00Z',
  },
  {
    id: uid(), name: 'Q1 2026 Quotation Template', category: 'quote', documentType: 'terms',
    tags: ['template', 'quotation', '2026'], version: '1.0', fileSize: '340 KB', fileType: 'docx',
    linkedEntities: [],
    uploadedBy: 'Omar Al-Zahrani',
    createdAt: '2026-01-05T08:00:00Z', updatedAt: '2026-01-05T08:00:00Z',
  },
  {
    id: uid(), name: 'Hikvision Product Specifications', category: 'general', documentType: 'technical',
    tags: ['hikvision', 'specifications', 'cctv'], version: '3.2', fileSize: '8.7 MB', fileType: 'pdf',
    linkedEntities: [{ type: 'Manufacturer', id: 'mfr-hik', name: 'Hikvision' }],
    uploadedBy: 'Salman Al-Mutairi',
    createdAt: '2025-06-10T08:00:00Z', updatedAt: '2026-01-15T08:00:00Z',
  },
  {
    id: uid(), name: 'ISO 27001 Compliance Checklist', category: 'compliance', documentType: 'terms',
    tags: ['iso-27001', 'compliance', 'audit'], version: '1.3', fileSize: '520 KB', fileType: 'xlsx',
    linkedEntities: [],
    uploadedBy: 'Lina Al-Asmari',
    createdAt: '2025-09-01T08:00:00Z', updatedAt: '2026-02-01T08:00:00Z',
  },
  {
    id: uid(), name: 'Standard SLA Document', category: 'legal', documentType: 'sla',
    tags: ['sla', 'standard', 'response-time'], version: '4.0', fileSize: '180 KB', fileType: 'pdf',
    linkedEntities: [],
    uploadedBy: 'Ahmed Al-Dosari',
    createdAt: '2024-06-15T08:00:00Z', updatedAt: '2025-12-20T08:00:00Z',
  },
  {
    id: uid(), name: 'KFSH Delivery Note — Phase 2', category: 'contract', documentType: 'delivery',
    tags: ['kfsh', 'delivery', 'phase-2', 'cctv'], version: '1.0', fileSize: '95 KB', fileType: 'pdf',
    linkedEntities: [{ type: 'Customer', id: 'c2', name: 'King Faisal Specialist Hospital' }],
    uploadedBy: 'Khalid Al-Qahtani',
    createdAt: '2026-01-20T08:00:00Z', updatedAt: '2026-01-20T08:00:00Z',
  },
  {
    id: uid(), name: 'Bosch Fire Panel Warranty Certificate', category: 'general', documentType: 'warranty',
    tags: ['bosch', 'warranty', 'fire-panel'], version: '1.0', fileSize: '1.1 MB', fileType: 'pdf',
    linkedEntities: [{ type: 'Manufacturer', id: 'mfr-bosch', name: 'Bosch Security' }],
    uploadedBy: 'Salman Al-Mutairi',
    createdAt: '2025-12-01T08:00:00Z', updatedAt: '2025-12-01T08:00:00Z',
  },
  {
    id: uid(), name: 'MOI Project Site Photos', category: 'general', documentType: 'technical',
    tags: ['moi', 'site-survey', 'photos'], version: '1.0', fileSize: '24.3 MB', fileType: 'zip',
    linkedEntities: [{ type: 'Customer', id: 'c5', name: 'Ministry of Interior' }],
    uploadedBy: 'Tariq Al-Sulaiman',
    createdAt: '2026-02-10T08:00:00Z', updatedAt: '2026-02-10T08:00:00Z',
  },
])

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
  form.value = defaultForm()
  showUploadModal.value = true
}

function uploadDocument() {
  const now = new Date().toISOString()
  const tags = form.value.tags.split(',').map(t => t.trim()).filter(Boolean)
  const linked = form.value.linkedEntityName
    ? [{ type: 'Custom', id: uid(), name: form.value.linkedEntityName }]
    : []

  documents.value.push({
    id: uid(),
    name: form.value.name,
    category: form.value.category,
    documentType: form.value.documentType,
    tags,
    version: form.value.version,
    fileSize: form.value.fileSize || '0 KB',
    fileType: form.value.fileType,
    linkedEntities: linked,
    uploadedBy: 'Current User',
    createdAt: now,
    updatedAt: now,
  })
  showUploadModal.value = false
}

function deleteDocument(id: string) {
  documents.value = documents.value.filter(d => d.id !== id)
}
</script>

<template>
  <div class="documents-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Documents</h1>
        <p class="page-header-subtitle">{{ filteredDocuments.length }} document{{ filteredDocuments.length !== 1 ? 's' : '' }}</p>
      </div>
      <button class="btn btn-primary" @click="openUploadModal">
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
                <button class="btn btn-ghost btn-icon btn-sm" title="Download">
                  <Download :size="14" />
                </button>
                <button class="btn btn-ghost btn-icon btn-sm" title="Delete" @click="deleteDocument(doc.id)">
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
                <input v-model="form.version" type="text" class="form-input" placeholder="1.0" />
              </div>
            </div>

            <!-- File upload area -->
            <div class="form-group">
              <label class="form-label">File</label>
              <div class="file-upload-area">
                <Upload :size="32" class="upload-icon" />
                <p class="upload-text">Click to select or drag & drop</p>
                <p class="upload-hint">PDF, DOCX, XLSX, PNG, JPG, ZIP up to 50MB</p>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">File Type</label>
                <select v-model="form.fileType" class="form-select">
                  <option value="pdf">PDF</option>
                  <option value="docx">DOCX</option>
                  <option value="xlsx">XLSX</option>
                  <option value="png">PNG</option>
                  <option value="jpg">JPG</option>
                  <option value="zip">ZIP</option>
                </select>
              </div>
              <div class="form-group">
                <label class="form-label">File Size</label>
                <input v-model="form.fileSize" type="text" class="form-input" placeholder="e.g. 2.4 MB" />
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">Link to Entity</label>
              <input v-model="form.linkedEntityName" type="text" class="form-input" placeholder="Customer or manufacturer name" />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showUploadModal = false">Cancel</button>
            <button
              class="btn btn-primary"
              :disabled="!form.name"
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
