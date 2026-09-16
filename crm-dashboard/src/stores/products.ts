import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { productsService } from '@/services'
import type { Product } from '@/types'

export const useProductsStore = defineStore('products', () => {
  const products = ref<Product[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const meta = ref({ page: 1, limit: 25, total: 0, totalPages: 1 })

  const activeProducts = computed(() => products.value.filter((p) => p.isActive))

  async function fetchProducts(params?: Record<string, unknown>) {
    loading.value = true
    error.value = null
    try {
      const res = await productsService.list(params)
      if (res.success) {
        products.value = res.data
        if (res.meta) meta.value = res.meta
      }
    } catch (e) {
      error.value = 'Failed to load products'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  function getByManufacturer(mfrId: string) {
    return products.value.filter((p) => p.manufacturerId === mfrId)
  }

  function searchProducts(query: string) {
    const q = query.toLowerCase()
    return products.value.filter(
      (p) => p.name.toLowerCase().includes(q) || p.sku.toLowerCase().includes(q),
    )
  }

  async function addProduct(data: Partial<Product>) {
    const res = await productsService.create(data)
    if (res.success) {
      products.value.unshift(res.data)
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to create product')
  }

  async function updateProduct(id: string, data: Partial<Product>) {
    const res = await productsService.update(id, data)
    if (res.success) {
      const idx = products.value.findIndex((p) => p.id === id)
      if (idx !== -1) products.value[idx] = res.data
      return res.data
    }
    throw new Error(res.error?.message ?? 'Failed to update product')
  }

  async function deleteProduct(id: string) {
    await productsService.delete(id)
    products.value = products.value.filter((p) => p.id !== id)
  }

  return {
    products,
    loading,
    error,
    meta,
    activeProducts,
    fetchProducts,
    getByManufacturer,
    searchProducts,
    addProduct,
    updateProduct,
    deleteProduct,
  }
})
