import { defineStore } from 'pinia'
import { ref } from 'vue'
import { documentsService } from '@/services'

export const useDocumentsStore = defineStore('documents', () => {
  const documents = ref<unknown[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  async function fetchDocuments(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await documentsService.list(params)
      if (res.success) {
        documents.value = res.data
        if (res.meta) meta.value = res.meta
      }
    } catch (e) { error.value = 'Failed to load documents'; console.error(e) }
    finally { loading.value = false }
  }

  async function upload(formData: FormData) {
    const res = await documentsService.upload(formData)
    if (res.success) { documents.value.unshift(res.data as never); return res.data }
    throw new Error('Upload failed')
  }

  async function deleteDocument(id: string) {
    await documentsService.delete(id)
    documents.value = (documents.value as Array<{ id: string }>).filter((d) => d.id !== id)
  }

  function downloadUrl(id: string) {
    return documentsService.downloadUrl(id)
  }

  return { documents, loading, error, meta, fetchDocuments, upload, deleteDocument, downloadUrl }
})
