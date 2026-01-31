import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { PriceBook, PriceBookEntry, PriceBookType, Status } from '@/types'

export const usePriceBooksStore = defineStore('priceBooks', () => {
  const priceBooks = ref<PriceBook[]>([
    {
      id: 'PB-001',
      name: 'Standard Pricing 2024',
      description: 'Standard pricing for all products - Base pricing structure',
      type: 'standard',
      status: 'active',
      validFrom: '2024-01-01',
      customerIds: [],  // Applies to all customers
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-001',
          priceBookId: 'PB-001',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 1008,
          discountPercent: 0,
          discountAmount: 0
        },
        {
          id: 'PBE-002',
          priceBookId: 'PB-001',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          standardPrice: 1232,
          customPrice: 1232,
          discountPercent: 0,
          discountAmount: 0
        }
      ],
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'PB-002',
      name: 'VIP Customer Pricing',
      description: 'Special pricing for VIP customers with 15% discount',
      type: 'customer_specific',
      status: 'active',
      validFrom: '2024-01-01',
      validUntil: '2024-12-31',
      customerIds: ['CUST-001', 'CUST-003'],  // Specific customers only
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-003',
          priceBookId: 'PB-002',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 857,
          discountPercent: 15,
          discountAmount: 151
        },
        {
          id: 'PBE-004',
          priceBookId: 'PB-002',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          standardPrice: 1232,
          customPrice: 1047,
          discountPercent: 15,
          discountAmount: 185
        }
      ],
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-004'
    },
    {
      id: 'PB-003',
      name: 'Volume Discount - Bulk Orders',
      description: 'Volume-based pricing for orders over 50 units',
      type: 'volume',
      status: 'active',
      validFrom: '2024-01-01',
      customerIds: [],
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-005',
          priceBookId: 'PB-003',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 908,
          discountPercent: 10,
          discountAmount: 100,
          minQuantity: 50,
          maxQuantity: 99
        },
        {
          id: 'PBE-006',
          priceBookId: 'PB-003',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 857,
          discountPercent: 15,
          discountAmount: 151,
          minQuantity: 100
        }
      ],
      createdAt: '2024-02-01T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'PB-004',
      name: 'Annual Maintenance Contract Pricing',
      description: 'Special pricing for annual maintenance contracts',
      type: 'contract',
      status: 'active',
      validFrom: '2024-01-01',
      validUntil: '2024-12-31',
      customerIds: ['CUST-001'],
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-007',
          priceBookId: 'PB-004',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 807,
          discountPercent: 20,
          discountAmount: 201
        }
      ],
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-002'
    },
    {
      id: 'PB-005',
      name: 'ACME Corp - Security Services Contract',
      description: 'Comprehensive security services package for ACME Corporation including manpower and maintenance',
      type: 'customer_specific',
      status: 'active',
      validFrom: '2024-01-01',
      validUntil: '2025-12-31',
      customerIds: ['CUST-002'],
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-008',
          priceBookId: 'PB-005',
          productId: 'PRD-017',
          productSku: 'SRV-GUARD-BASIC',
          productName: 'Security Guard - Basic Level',
          standardPrice: 6429,
          customPrice: 6000,
          discountPercent: 6.67,
          discountAmount: 429
        },
        {
          id: 'PBE-009',
          priceBookId: 'PB-005',
          productId: 'PRD-018',
          productSku: 'SRV-GUARD-SUPERVISOR',
          productName: 'Security Supervisor',
          standardPrice: 9028,
          customPrice: 8500,
          discountPercent: 5.85,
          discountAmount: 528
        },
        {
          id: 'PBE-010',
          priceBookId: 'PB-005',
          productId: 'PRD-019',
          productSku: 'SRV-MAINT-CCTV',
          productName: 'CCTV System Maintenance',
          standardPrice: 1333,
          customPrice: 1200,
          discountPercent: 9.98,
          discountAmount: 133
        },
        {
          id: 'PBE-011',
          priceBookId: 'PB-005',
          productId: 'PRD-021',
          productSku: 'SRV-PATROL-SERVICE',
          productName: 'Mobile Patrol Service',
          standardPrice: 5385,
          customPrice: 5000,
          discountPercent: 7.15,
          discountAmount: 385
        },
        {
          id: 'PBE-012',
          priceBookId: 'PB-005',
          productId: 'PRD-001',
          productSku: 'CAM-HK-DS2CD2385G1',
          productName: 'Hikvision 8MP IP Dome Camera',
          standardPrice: 1008,
          customPrice: 900,
          discountPercent: 10.71,
          discountAmount: 108
        },
        {
          id: 'PBE-013',
          priceBookId: 'PB-005',
          productId: 'PRD-002',
          productSku: 'CAM-DH-IPC8542E',
          productName: 'Dahua 5MP Bullet Camera',
          standardPrice: 1232,
          customPrice: 1100,
          discountPercent: 10.71,
          discountAmount: 132
        }
      ],
      createdAt: '2024-02-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Long-term contract with recurring monthly services and equipment discount'
    },
    {
      id: 'PB-006',
      name: 'Tech Plaza Mall - Facility Management',
      description: 'Complete facility security and maintenance package for Tech Plaza Mall',
      type: 'customer_specific',
      status: 'active',
      validFrom: '2024-03-01',
      validUntil: '2025-02-28',
      customerIds: ['CUST-003'],
      contractIds: [],
      priceBookEntries: [
        {
          id: 'PBE-014',
          priceBookId: 'PB-006',
          productId: 'PRD-017',
          productSku: 'SRV-GUARD-BASIC',
          productName: 'Security Guard - Basic Level',
          standardPrice: 6429,
          customPrice: 5800,
          discountPercent: 9.78,
          discountAmount: 629
        },
        {
          id: 'PBE-015',
          priceBookId: 'PB-006',
          productId: 'PRD-018',
          productSku: 'SRV-GUARD-SUPERVISOR',
          productName: 'Security Supervisor',
          standardPrice: 9028,
          customPrice: 8200,
          discountPercent: 9.17,
          discountAmount: 828
        },
        {
          id: 'PBE-016',
          priceBookId: 'PB-006',
          productId: 'PRD-019',
          productSku: 'SRV-MAINT-CCTV',
          productName: 'CCTV System Maintenance',
          standardPrice: 1333,
          customPrice: 1150,
          discountPercent: 13.73,
          discountAmount: 183
        },
        {
          id: 'PBE-017',
          priceBookId: 'PB-006',
          productId: 'PRD-020',
          productSku: 'SRV-MAINT-ACCESS',
          productName: 'Access Control System Maintenance',
          standardPrice: 1000,
          customPrice: 900,
          discountPercent: 10.0,
          discountAmount: 100
        },
        {
          id: 'PBE-018',
          priceBookId: 'PB-006',
          productId: 'PRD-022',
          productSku: 'SRV-MONITORING-24-7',
          productName: '24/7 Monitoring Center Service',
          standardPrice: 4545,
          customPrice: 4200,
          discountPercent: 7.59,
          discountAmount: 345
        }
      ],
      createdAt: '2024-02-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Premium recurring services package for large shopping mall facility'
    }
  ])

  // Computed
  const activePriceBooks = computed(() =>
    priceBooks.value.filter(pb => pb.status === 'active')
  )

  const priceBooksByType = computed(() => {
    const grouped: Record<PriceBookType, PriceBook[]> = {
      standard: [],
      volume: [],
      contract: [],
      promotional: [],
      customer_specific: []
    }

    priceBooks.value.forEach(pb => {
      grouped[pb.type].push(pb)
    })

    return grouped
  })

  // Actions
  function getPriceBookById(id: string): PriceBook | undefined {
    return priceBooks.value.find(pb => pb.id === id)
  }

  function getPriceBooksForCustomer(customerId: string): PriceBook[] {
    const now = new Date()
    return priceBooks.value.filter(pb => {
      // Check if active
      if (pb.status !== 'active') return false

      // Check validity
      const validFrom = new Date(pb.validFrom)
      if (validFrom > now) return false

      if (pb.validUntil) {
        const validUntil = new Date(pb.validUntil)
        if (validUntil < now) return false
      }

      // Check if applies to this customer
      if (pb.customerIds.length === 0) return true  // Applies to all
      return pb.customerIds.includes(customerId)
    })
  }

  function getProductPrice(productId: string, priceBookId: string, quantity: number = 1): number {
    const priceBook = getPriceBookById(priceBookId)
    if (!priceBook) return 0

    // For volume pricing, find the appropriate tier
    if (priceBook.type === 'volume') {
      const entries = priceBook.priceBookEntries
        .filter(e => e.productId === productId)
        .filter(e => {
          if (e.minQuantity && quantity < e.minQuantity) return false
          if (e.maxQuantity && quantity > e.maxQuantity) return false
          return true
        })
        .sort((a, b) => (b.minQuantity || 0) - (a.minQuantity || 0))

      if (entries.length > 0) {
        return entries[0].customPrice
      }
    }

    // For other types, just find the product entry
    const entry = priceBook.priceBookEntries.find(e => e.productId === productId)
    return entry ? entry.customPrice : 0
  }

  function addPriceBook(priceBook: Omit<PriceBook, 'id' | 'createdAt'>) {
    const id = `PB-${String(priceBooks.value.length + 1).padStart(3, '0')}`

    const newPriceBook: PriceBook = {
      ...priceBook,
      id,
      createdAt: new Date().toISOString()
    }

    priceBooks.value.push(newPriceBook)
    return newPriceBook
  }

  function updatePriceBook(id: string, updates: Partial<PriceBook>) {
    const index = priceBooks.value.findIndex(pb => pb.id === id)
    if (index !== -1) {
      priceBooks.value[index] = {
        ...priceBooks.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return priceBooks.value[index]
    }
    return null
  }

  function deletePriceBook(id: string) {
    const index = priceBooks.value.findIndex(pb => pb.id === id)
    if (index !== -1) {
      priceBooks.value.splice(index, 1)
      return true
    }
    return false
  }

  function addPriceBookEntry(priceBookId: string, entry: Omit<PriceBookEntry, 'id' | 'priceBookId'>) {
    const priceBook = getPriceBookById(priceBookId)
    if (!priceBook) return null

    const id = `PBE-${String(priceBooks.value.reduce((sum, pb) => sum + pb.priceBookEntries.length, 0) + 1).padStart(3, '0')}`

    const newEntry: PriceBookEntry = {
      ...entry,
      id,
      priceBookId
    }

    priceBook.priceBookEntries.push(newEntry)
    priceBook.updatedAt = new Date().toISOString()

    return newEntry
  }

  function updatePriceBookEntry(priceBookId: string, entryId: string, updates: Partial<PriceBookEntry>) {
    const priceBook = getPriceBookById(priceBookId)
    if (!priceBook) return null

    const index = priceBook.priceBookEntries.findIndex(e => e.id === entryId)
    if (index !== -1) {
      priceBook.priceBookEntries[index] = {
        ...priceBook.priceBookEntries[index],
        ...updates
      }
      priceBook.updatedAt = new Date().toISOString()
      return priceBook.priceBookEntries[index]
    }
    return null
  }

  function deletePriceBookEntry(priceBookId: string, entryId: string) {
    const priceBook = getPriceBookById(priceBookId)
    if (!priceBook) return false

    const index = priceBook.priceBookEntries.findIndex(e => e.id === entryId)
    if (index !== -1) {
      priceBook.priceBookEntries.splice(index, 1)
      priceBook.updatedAt = new Date().toISOString()
      return true
    }
    return false
  }

  function calculateDiscountedPrice(standardPrice: number, customPrice: number) {
    const discountAmount = standardPrice - customPrice
    const discountPercent = (discountAmount / standardPrice) * 100
    return {
      discountAmount,
      discountPercent
    }
  }

  // Get price books for a specific customer (alias for getPriceBooksForCustomer)
  function getPriceBooksByCustomerId(customerId: string): PriceBook[] {
    return getPriceBooksForCustomer(customerId)
  }

  // Get a specific price book entry
  function getPriceBookEntry(priceBookId: string, productId: string): PriceBookEntry | undefined {
    const priceBook = getPriceBookById(priceBookId)
    if (!priceBook) return undefined
    return priceBook.priceBookEntries.find(e => e.productId === productId)
  }

  return {
    priceBooks,
    activePriceBooks,
    priceBooksByType,
    getPriceBookById,
    getPriceBooksForCustomer,
    getPriceBooksByCustomerId,
    getPriceBookEntry,
    getProductPrice,
    addPriceBook,
    updatePriceBook,
    deletePriceBook,
    addPriceBookEntry,
    updatePriceBookEntry,
    deletePriceBookEntry,
    calculateDiscountedPrice
  }
})
