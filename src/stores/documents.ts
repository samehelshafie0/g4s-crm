import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Document, DocumentType, DocumentCategory } from '@/types'

export const useDocumentsStore = defineStore('documents', () => {
  const documents = ref<Document[]>([
    // Sample documents for demonstration
    {
      id: 'DOC-001',
      name: 'Standard Terms and Conditions 2024',
      fileName: 'terms-and-conditions-2024.pdf',
      fileType: 'application/pdf',
      fileSize: 245760,
      category: 'legal',
      type: 'terms',
      description: 'Standard T&C for all contracts',
      fileUrl: '/documents/terms-2024.pdf',
      uploadedBy: 'USR-001',
      uploadedAt: '2024-01-01T00:00:00Z',
      tags: ['legal', 'standard', 'terms'],
      version: 1
    },
    {
      id: 'DOC-002',
      name: 'Delivery Policy Saudi Arabia',
      fileName: 'delivery-policy-sa.pdf',
      fileType: 'application/pdf',
      fileSize: 180000,
      category: 'general',
      type: 'delivery',
      description: 'Standard delivery terms for KSA',
      fileUrl: '/documents/delivery-policy.pdf',
      uploadedBy: 'USR-001',
      uploadedAt: '2024-01-01T00:00:00Z',
      tags: ['delivery', 'policy'],
      version: 1
    }
  ])

  // Computed
  const activeDocuments = computed(() =>
    documents.value.filter(d => !d.linkedToQuotes || d.linkedToQuotes.length === 0)
  )

  const documentsByCategory = computed(() => {
    const grouped: Record<DocumentCategory, Document[]> = {
      contract: [],
      quote: [],
      general: [],
      compliance: [],
      legal: []
    }

    documents.value.forEach(doc => {
      grouped[doc.category].push(doc)
    })

    return grouped
  })

  const documentsByType = computed(() => {
    const grouped: Record<DocumentType, Document[]> = {
      terms: [],
      delivery: [],
      technical: [],
      warranty: [],
      sla: [],
      other: []
    }

    documents.value.forEach(doc => {
      grouped[doc.type].push(doc)
    })

    return grouped
  })

  // Actions
  function getDocumentById(id: string): Document | undefined {
    return documents.value.find(d => d.id === id)
  }

  function getDocumentsByType(type: DocumentType): Document[] {
    return documents.value.filter(d => d.type === type)
  }

  function getDocumentsByCategory(category: DocumentCategory): Document[] {
    return documents.value.filter(d => d.category === category)
  }

  function searchDocuments(query: string): Document[] {
    const lowerQuery = query.toLowerCase()
    return documents.value.filter(
      d =>
        d.name.toLowerCase().includes(lowerQuery) ||
        d.fileName.toLowerCase().includes(lowerQuery) ||
        d.description?.toLowerCase().includes(lowerQuery) ||
        d.tags?.some(tag => tag.toLowerCase().includes(lowerQuery))
    )
  }

  function addDocument(doc: Omit<Document, 'id' | 'uploadedAt'>): Document {
    const id = `DOC-${String(documents.value.length + 1).padStart(3, '0')}`

    const newDoc: Document = {
      ...doc,
      id,
      uploadedAt: new Date().toISOString(),
      version: 1
    }

    documents.value.push(newDoc)
    return newDoc
  }

  function updateDocument(id: string, updates: Partial<Document>): Document | null {
    const index = documents.value.findIndex(d => d.id === id)
    if (index !== -1) {
      documents.value[index] = {
        ...documents.value[index],
        ...updates
      }
      return documents.value[index]
    }
    return null
  }

  function deleteDocument(id: string): boolean {
    const index = documents.value.findIndex(d => d.id === id)
    if (index !== -1) {
      documents.value.splice(index, 1)
      return true
    }
    return false
  }

  // Link management
  function linkToQuote(documentId: string, quoteId: string): boolean {
    const doc = getDocumentById(documentId)
    if (doc) {
      if (!doc.linkedToQuotes) {
        doc.linkedToQuotes = []
      }
      if (!doc.linkedToQuotes.includes(quoteId)) {
        doc.linkedToQuotes.push(quoteId)
      }
      return true
    }
    return false
  }

  function unlinkFromQuote(documentId: string, quoteId: string): boolean {
    const doc = getDocumentById(documentId)
    if (doc && doc.linkedToQuotes) {
      const index = doc.linkedToQuotes.indexOf(quoteId)
      if (index !== -1) {
        doc.linkedToQuotes.splice(index, 1)
        return true
      }
    }
    return false
  }

  function linkToContract(documentId: string, contractId: string): boolean {
    const doc = getDocumentById(documentId)
    if (doc) {
      if (!doc.linkedToContracts) {
        doc.linkedToContracts = []
      }
      if (!doc.linkedToContracts.includes(contractId)) {
        doc.linkedToContracts.push(contractId)
      }
      return true
    }
    return false
  }

  function unlinkFromContract(documentId: string, contractId: string): boolean {
    const doc = getDocumentById(documentId)
    if (doc && doc.linkedToContracts) {
      const index = doc.linkedToContracts.indexOf(contractId)
      if (index !== -1) {
        doc.linkedToContracts.splice(index, 1)
        return true
      }
    }
    return false
  }

  // File upload helper (for demo - stores base64)
  function handleFileUpload(file: File): Promise<{ fileUrl: string; fileData: string }> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = (e) => {
        const fileData = e.target?.result as string
        resolve({
          fileUrl: `/uploads/${file.name}`,
          fileData
        })
      }
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
  }

  return {
    documents,
    activeDocuments,
    documentsByCategory,
    documentsByType,
    getDocumentById,
    getDocumentsByType,
    getDocumentsByCategory,
    searchDocuments,
    addDocument,
    updateDocument,
    deleteDocument,
    linkToQuote,
    unlinkFromQuote,
    linkToContract,
    unlinkFromContract,
    handleFileUpload
  }
})
