<template>
  <div class="documents-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search documents..."
          class="search-input"
        />
        <button class="btn btn-primary" @click="showUploadModal = true">
          <span>📤</span>
          Upload Document
        </button>
      </div>
    </div>

    <div class="filters-bar">
      <div class="filter-group">
        <label>Type:</label>
        <select v-model="typeFilter" class="filter-select">
          <option value="all">All Types</option>
          <option value="terms">Terms & Conditions</option>
          <option value="delivery">Delivery Terms</option>
          <option value="technical">Technical Specs</option>
          <option value="warranty">Warranty</option>
          <option value="sla">SLA</option>
          <option value="other">Other</option>
        </select>
      </div>

      <div class="filter-group">
        <label>Category:</label>
        <select v-model="categoryFilter" class="filter-select">
          <option value="all">All Categories</option>
          <option value="contract">Contract</option>
          <option value="quote">Quote</option>
          <option value="general">General</option>
          <option value="compliance">Compliance</option>
          <option value="legal">Legal</option>
        </select>
      </div>

      <div class="stats-summary">
        <div class="stat-item">
          <span class="stat-label">Total</span>
          <span class="stat-value">{{ filteredDocuments.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Terms Docs</span>
          <span class="stat-value text-primary">{{ termsCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Size</span>
          <span class="stat-value">{{ formatSize(totalSize) }}</span>
        </div>
      </div>
    </div>

    <BaseTable
      :columns="tableColumns"
      :data="filteredDocuments"
      :loading="false"
    >
      <template #cell-name="{ row }">
        <div class="name-cell">
          <span class="file-icon">{{ getFileIcon(row.fileType) }}</span>
          <div>
            <div class="doc-name">{{ row.name }}</div>
            <div class="doc-version">{{ row.fileName }}</div>
          </div>
        </div>
      </template>

      <template #cell-type="{ row }">
        <BaseBadge :variant="getTypeVariant(row.type)">
          {{ formatType(row.type) }}
        </BaseBadge>
      </template>

      <template #cell-category="{ row }">
        <div class="category-cell">
          {{ formatCategory(row.category) }}
        </div>
      </template>

      <template #cell-size="{ row }">
        <div class="size-cell">
          {{ formatSize(row.fileSize) }}
        </div>
      </template>

      <template #cell-modified="{ row }">
        <div class="date-cell">
          <div>{{ formatDate(row.uploadedAt) }}</div>
          <div class="text-secondary">{{ row.uploadedBy }}</div>
        </div>
      </template>

      <template #cell-actions="{ row }">
        <div class="action-buttons">
          <button class="action-btn" @click="viewDocument(row)" title="View">
            👁️
          </button>
          <button class="action-btn" @click="editDocument(row)" title="Edit">
            ✏️
          </button>
          <button class="action-btn" @click="deleteDocument(row)" title="Delete">
            🗑️
          </button>
        </div>
      </template>
    </BaseTable>

    <!-- Upload Modal -->
    <BaseModal
      v-model="showUploadModal"
      title="Upload Document"
      size="md"
      @confirm="handleUpload"
    >
      <div class="upload-form">
        <div class="form-group">
          <label class="form-label">Document Name *</label>
          <input
            v-model="uploadForm.name"
            type="text"
            class="form-input"
            placeholder="e.g., Standard Terms and Conditions 2024"
            required
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Type *</label>
            <select v-model="uploadForm.type" class="form-input" required>
              <option value="">Select type</option>
              <option value="terms">Terms & Conditions</option>
              <option value="delivery">Delivery Terms</option>
              <option value="technical">Technical Specifications</option>
              <option value="warranty">Warranty</option>
              <option value="sla">Service Level Agreement</option>
              <option value="other">Other</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">Category *</label>
            <select v-model="uploadForm.category" class="form-input" required>
              <option value="">Select category</option>
              <option value="contract">Contract</option>
              <option value="quote">Quote</option>
              <option value="general">General</option>
              <option value="compliance">Compliance</option>
              <option value="legal">Legal</option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Description</label>
          <textarea
            v-model="uploadForm.description"
            class="form-input"
            rows="2"
            placeholder="Brief description of the document"
          ></textarea>
        </div>

        <div class="form-group">
          <label class="form-label">Select File *</label>
          <input
            ref="fileInput"
            type="file"
            class="form-input file-input"
            accept=".pdf,.doc,.docx"
            @change="onFileSelect"
            required
          />
          <small class="form-hint">Accepted formats: PDF, DOC, DOCX (Max 10MB)</small>
        </div>

        <div v-if="selectedFile" class="file-preview">
          <span class="file-icon-large">{{ getFileIcon(selectedFile.type) }}</span>
          <div class="file-info">
            <div class="file-name">{{ selectedFile.name }}</div>
            <div class="file-size">{{ formatSize(selectedFile.size) }}</div>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useDocumentsStore } from '@/stores/documents'
import BaseTable from '@/components/UI/BaseTable.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import BaseModal from '@/components/UI/BaseModal.vue'
import { useToast } from '@/composables/useToast'
import type { DocumentType, DocumentCategory } from '@/types'

const documentsStore = useDocumentsStore()
const { success, info } = useToast()

const searchQuery = ref('')
const typeFilter = ref('all')
const categoryFilter = ref('all')
const showUploadModal = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)

const uploadForm = ref({
  name: '',
  type: '' as DocumentType | '',
  category: '' as DocumentCategory | '',
  description: ''
})

const tableColumns = [
  { key: 'name', label: 'Document Name', sortable: true },
  { key: 'type', label: 'Type', sortable: true },
  { key: 'category', label: 'Category', sortable: true },
  { key: 'size', label: 'Size', sortable: true },
  { key: 'modified', label: 'Uploaded', sortable: true },
  { key: 'actions', label: 'Actions', sortable: false }
]

const filteredDocuments = computed(() => {
  let filtered = searchQuery.value
    ? documentsStore.searchDocuments(searchQuery.value)
    : documentsStore.documents

  if (typeFilter.value !== 'all') {
    filtered = filtered.filter(doc => doc.type === typeFilter.value)
  }

  if (categoryFilter.value !== 'all') {
    filtered = filtered.filter(doc => doc.category === categoryFilter.value)
  }

  return filtered
})

const termsCount = computed(() =>
  documentsStore.documents.filter(doc => doc.type === 'terms').length
)

const totalSize = computed(() =>
  documentsStore.documents.reduce((sum, doc) => sum + doc.fileSize, 0)
)

function onFileSelect(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0]
  }
}

async function handleUpload() {
  if (!uploadForm.value.name || !uploadForm.value.type || !uploadForm.value.category || !selectedFile.value) {
    info('Please fill in all required fields')
    return
  }

  try {
    // Handle file upload
    const { fileUrl, fileData } = await documentsStore.handleFileUpload(selectedFile.value)

    // Create document
    documentsStore.addDocument({
      name: uploadForm.value.name,
      fileName: selectedFile.value.name,
      fileType: selectedFile.value.type,
      fileSize: selectedFile.value.size,
      category: uploadForm.value.category as DocumentCategory,
      type: uploadForm.value.type as DocumentType,
      description: uploadForm.value.description || undefined,
      fileUrl,
      fileData,
      uploadedBy: 'USR-001' // Current user
    })

    success('Document uploaded successfully')
    resetUploadForm()
    showUploadModal.value = false
  } catch (error) {
    info('Error uploading document')
  }
}

function resetUploadForm() {
  uploadForm.value = {
    name: '',
    type: '' as DocumentType | '',
    category: '' as DocumentCategory | '',
    description: ''
  }
  selectedFile.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function getFileIcon(fileType: string): string {
  if (fileType.includes('pdf')) return '📄'
  if (fileType.includes('doc')) return '📝'
  if (fileType.includes('sheet') || fileType.includes('excel')) return '📊'
  return '📎'
}

function formatType(type: DocumentType): string {
  const types: Record<DocumentType, string> = {
    terms: 'Terms & Conditions',
    delivery: 'Delivery Terms',
    technical: 'Technical Specs',
    warranty: 'Warranty',
    sla: 'SLA',
    other: 'Other'
  }
  return types[type]
}

function formatCategory(category: DocumentCategory): string {
  const categories: Record<DocumentCategory, string> = {
    contract: 'Contract',
    quote: 'Quote',
    general: 'General',
    compliance: 'Compliance',
    legal: 'Legal'
  }
  return categories[category]
}

function getTypeVariant(type: DocumentType): 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const variants: Record<DocumentType, 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
    terms: 'warning',
    delivery: 'info',
    technical: 'primary',
    warranty: 'success',
    sla: 'info',
    other: 'default'
  }
  return variants[type]
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  })
}

function viewDocument(doc: any) {
  info(`Viewing: ${doc.name}`)
  // In real app, would open document viewer
}

function editDocument(doc: any) {
  info(`Editing: ${doc.name}`)
  // In real app, would open edit modal
}

function deleteDocument(doc: any) {
  if (confirm(`Delete "${doc.name}"?`)) {
    documentsStore.deleteDocument(doc.id)
    success('Document deleted')
  }
}
</script>

<style scoped>
.documents-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: 100%;
  padding: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1.25rem;
  width: 100%;
}

.search-input {
  flex: 1;
  padding: 1rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 10px;
  font-size: 1.05rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn {
  padding: 1rem 2rem;
  border: none;
  border-radius: 10px;
  font-size: 1.05rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.625rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
  transform: translateY(-2px);
}

.filters-bar {
  display: flex;
  gap: 2.5rem;
  align-items: center;
  flex-wrap: wrap;
  padding: 1.75rem 2rem;
  background: var(--color-surface);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.filter-group label {
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.filter-select {
  padding: 0.75rem 1.5rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
  min-width: 200px;
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
}

.stats-summary {
  display: flex;
  gap: 2.5rem;
  margin-left: auto;
  padding-left: 2.5rem;
  border-left: 1px solid var(--color-border);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-primary);
}

.text-primary {
  color: var(--color-primary) !important;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 0.75rem;
  min-width: 350px;
}

.file-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.doc-name {
  font-weight: 600;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}

.doc-version {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.category-cell {
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
  color: var(--color-text-primary);
}

.size-cell {
  padding: 1rem 0.75rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-primary);
}

.date-cell {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding: 1rem 0.75rem;
  min-width: 200px;
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: 0.95rem;
}

.action-buttons {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  padding: 0.5rem;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.3rem;
  cursor: pointer;
  padding: 0.35rem 0.65rem;
  transition: transform 0.2s;
  opacity: 0.7;
}

.action-btn:hover {
  transform: scale(1.2);
  opacity: 1;
}

/* Upload Form */
.upload-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.form-input {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text-primary);
  border-radius: 6px;
  font-size: 0.95rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.form-hint {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-style: italic;
}

.file-input {
  padding: 0.5rem;
  cursor: pointer;
}

.file-preview {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--color-surface);
  border-radius: 8px;
  border: 1px solid var(--color-border);
}

.file-icon-large {
  font-size: 2.5rem;
}

.file-info {
  flex: 1;
}

.file-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.file-size {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}
</style>
