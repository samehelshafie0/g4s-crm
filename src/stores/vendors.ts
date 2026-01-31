import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export interface Vendor {
  id: string
  vendorCode: string
  name: string
  nameArabic?: string
  category: 'manufacturer' | 'distributor' | 'local_supplier' | 'service_provider'
  status: 'active' | 'inactive' | 'suspended'

  // Contact Information
  primaryContact: string
  email: string
  phone: string
  website?: string

  // Address
  address: string
  city: string
  country: string
  postalCode?: string

  // Commercial Details
  taxId?: string
  commercialRegistration?: string
  vatNumber?: string

  // Payment Terms
  paymentTerms: string // e.g., "Net 30", "Net 60", "COD"
  paymentTermsDays: number
  currency: 'SAR' | 'USD' | 'EUR' | 'CNY'
  creditLimit?: number

  // Lead Time & Delivery
  defaultLeadTimeDays: number
  minimumOrderValue?: number
  shippingMethod: string // e.g., "Air Freight", "Sea Freight", "Local Delivery"

  // Performance Metrics
  rating: number // 1-5
  onTimeDeliveryRate?: number // percentage
  qualityRating?: number // 1-5
  totalPurchaseValue?: number
  lastPurchaseDate?: string

  // Notes
  notes?: string

  // Audit
  createdAt: string
  createdBy: string
  updatedAt?: string
  updatedBy?: string
}

export const useVendorsStore = defineStore('vendors', () => {
  const vendors = ref<Vendor[]>([
    {
      id: 'VEN-001',
      vendorCode: 'HIK-MENA',
      name: 'Hikvision MENA',
      nameArabic: 'هيكفيجن الشرق الأوسط',
      category: 'manufacturer',
      status: 'active',
      primaryContact: 'Ahmed Hassan',
      email: 'sales@hikvision-mena.com',
      phone: '+971-4-1234567',
      website: 'https://www.hikvision.com',
      address: 'Dubai Silicon Oasis',
      city: 'Dubai',
      country: 'UAE',
      postalCode: '00000',
      taxId: 'UAE-TAX-001',
      vatNumber: '100123456700003',
      paymentTerms: 'Net 45',
      paymentTermsDays: 45,
      currency: 'USD',
      creditLimit: 500000,
      defaultLeadTimeDays: 30,
      minimumOrderValue: 5000,
      shippingMethod: 'Air Freight',
      rating: 4.5,
      onTimeDeliveryRate: 92,
      qualityRating: 5,
      totalPurchaseValue: 1250000,
      lastPurchaseDate: '2024-11-15',
      notes: 'Preferred supplier for IP cameras and NVR systems. Good after-sales support.',
      createdAt: '2024-01-10T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'VEN-002',
      vendorCode: 'DAH-TECH',
      name: 'Dahua Technology Middle East',
      nameArabic: 'داهوا التقنية الشرق الأوسط',
      category: 'manufacturer',
      status: 'active',
      primaryContact: 'Mohammed Ali',
      email: 'info@dahuame.com',
      phone: '+971-4-7654321',
      website: 'https://www.dahuasecurity.com',
      address: 'Business Bay',
      city: 'Dubai',
      country: 'UAE',
      postalCode: '00000',
      vatNumber: '100987654300003',
      paymentTerms: 'Net 30',
      paymentTermsDays: 30,
      currency: 'USD',
      creditLimit: 400000,
      defaultLeadTimeDays: 25,
      minimumOrderValue: 4000,
      shippingMethod: 'Air Freight',
      rating: 4.2,
      onTimeDeliveryRate: 88,
      qualityRating: 4,
      totalPurchaseValue: 890000,
      lastPurchaseDate: '2024-11-20',
      notes: 'Competitive pricing on bullet cameras and PTZ systems.',
      createdAt: '2024-01-12T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'VEN-003',
      vendorCode: 'AXIS-KSA',
      name: 'Axis Communications Saudi Arabia',
      nameArabic: 'أكسيس للاتصالات السعودية',
      category: 'manufacturer',
      status: 'active',
      primaryContact: 'Lars Anderson',
      email: 'sales.sa@axis.com',
      phone: '+966-11-4567890',
      website: 'https://www.axis.com',
      address: 'Riyadh Business Park',
      city: 'Riyadh',
      country: 'Saudi Arabia',
      postalCode: '11564',
      taxId: 'SA-TAX-003',
      vatNumber: '300123456700003',
      paymentTerms: 'Net 60',
      paymentTermsDays: 60,
      currency: 'EUR',
      creditLimit: 600000,
      defaultLeadTimeDays: 35,
      minimumOrderValue: 8000,
      shippingMethod: 'Air Freight',
      rating: 4.8,
      onTimeDeliveryRate: 95,
      qualityRating: 5,
      totalPurchaseValue: 2100000,
      lastPurchaseDate: '2024-11-10',
      notes: 'Premium quality, higher price point. Excellent for enterprise projects.',
      createdAt: '2024-01-08T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'VEN-004',
      vendorCode: 'ELTEC-RY',
      name: 'Eltec Electronics - Riyadh',
      nameArabic: 'التيك للإلكترونيات - الرياض',
      category: 'local_supplier',
      status: 'active',
      primaryContact: 'Khalid Al-Mutairi',
      email: 'sales@eltec-ksa.com',
      phone: '+966-11-2345678',
      address: 'Al-Malaz District',
      city: 'Riyadh',
      country: 'Saudi Arabia',
      postalCode: '11432',
      taxId: 'SA-TAX-004',
      commercialRegistration: '1010123456',
      vatNumber: '300234567800003',
      paymentTerms: 'Net 15',
      paymentTermsDays: 15,
      currency: 'SAR',
      creditLimit: 200000,
      defaultLeadTimeDays: 7,
      minimumOrderValue: 1000,
      shippingMethod: 'Local Delivery',
      rating: 4.0,
      onTimeDeliveryRate: 85,
      qualityRating: 4,
      totalPurchaseValue: 450000,
      lastPurchaseDate: '2024-12-01',
      notes: 'Good for quick turnaround and local stock items. Cables and accessories.',
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'VEN-005',
      vendorCode: 'HNET-DIS',
      name: 'HNet Distribution',
      nameArabic: 'اتش نت للتوزيع',
      category: 'distributor',
      status: 'active',
      primaryContact: 'Fatima Al-Shahrani',
      email: 'orders@hnetdist.com',
      phone: '+966-12-3456789',
      website: 'https://www.hnetdist.com',
      address: 'King Abdullah Economic City',
      city: 'Jeddah',
      country: 'Saudi Arabia',
      postalCode: '23964',
      taxId: 'SA-TAX-005',
      commercialRegistration: '4030567890',
      vatNumber: '300345678900003',
      paymentTerms: 'Net 30',
      paymentTermsDays: 30,
      currency: 'SAR',
      creditLimit: 350000,
      defaultLeadTimeDays: 14,
      minimumOrderValue: 2500,
      shippingMethod: 'Air Freight',
      rating: 4.3,
      onTimeDeliveryRate: 90,
      qualityRating: 4,
      totalPurchaseValue: 780000,
      lastPurchaseDate: '2024-11-28',
      notes: 'Carries multiple brands. Good stock availability for standard items.',
      createdAt: '2024-01-18T00:00:00Z',
      createdBy: 'USR-001'
    },
    {
      id: 'VEN-006',
      vendorCode: 'PANASONIC-ME',
      name: 'Panasonic Middle East',
      nameArabic: 'باناسونيك الشرق الأوسط',
      category: 'manufacturer',
      status: 'active',
      primaryContact: 'Yuki Tanaka',
      email: 'b2b@panasonic-me.com',
      phone: '+971-4-8889999',
      website: 'https://www.panasonic.com',
      address: 'JAFZA',
      city: 'Dubai',
      country: 'UAE',
      postalCode: '00000',
      vatNumber: '100456789000003',
      paymentTerms: 'Net 45',
      paymentTermsDays: 45,
      currency: 'USD',
      creditLimit: 300000,
      defaultLeadTimeDays: 40,
      minimumOrderValue: 6000,
      shippingMethod: 'Sea Freight',
      rating: 4.6,
      onTimeDeliveryRate: 93,
      qualityRating: 5,
      totalPurchaseValue: 650000,
      lastPurchaseDate: '2024-10-20',
      notes: 'Excellent quality networking equipment and intercoms.',
      createdAt: '2024-01-20T00:00:00Z',
      createdBy: 'USR-001'
    }
  ])

  // Computed
  const activeVendors = computed(() => vendors.value.filter(v => v.status === 'active'))

  const vendorsByCategory = computed(() => {
    return {
      manufacturer: vendors.value.filter(v => v.category === 'manufacturer'),
      distributor: vendors.value.filter(v => v.category === 'distributor'),
      local_supplier: vendors.value.filter(v => v.category === 'local_supplier'),
      service_provider: vendors.value.filter(v => v.category === 'service_provider')
    }
  })

  const topRatedVendors = computed(() => {
    return [...vendors.value]
      .filter(v => v.status === 'active')
      .sort((a, b) => b.rating - a.rating)
      .slice(0, 5)
  })

  const vendorStats = computed(() => {
    const active = vendors.value.filter(v => v.status === 'active').length
    const total = vendors.value.length
    const totalPurchaseValue = vendors.value.reduce((sum, v) => sum + (v.totalPurchaseValue || 0), 0)
    const avgRating = vendors.value.reduce((sum, v) => sum + v.rating, 0) / vendors.value.length
    const avgOnTimeDelivery = vendors.value
      .filter(v => v.onTimeDeliveryRate)
      .reduce((sum, v) => sum + (v.onTimeDeliveryRate || 0), 0) /
      vendors.value.filter(v => v.onTimeDeliveryRate).length

    return {
      total,
      active,
      inactive: total - active,
      totalPurchaseValue,
      avgRating: Math.round(avgRating * 10) / 10,
      avgOnTimeDelivery: Math.round(avgOnTimeDelivery * 10) / 10
    }
  })

  // Actions
  function getVendorById(id: string): Vendor | undefined {
    return vendors.value.find(v => v.id === id)
  }

  function getVendorByCode(code: string): Vendor | undefined {
    return vendors.value.find(v => v.vendorCode === code)
  }

  function getActiveVendors() {
    return vendors.value.filter(v => v.status === 'active')
  }

  function getVendorsByCategory(category: Vendor['category']) {
    return vendors.value.filter(v => v.category === category && v.status === 'active')
  }

  function addVendor(vendor: Omit<Vendor, 'id' | 'createdAt'>) {
    const newId = `VEN-${String(vendors.value.length + 1).padStart(3, '0')}`
    const newVendor: Vendor = {
      ...vendor,
      id: newId,
      createdAt: new Date().toISOString()
    }
    vendors.value.push(newVendor)
    return newVendor
  }

  function updateVendor(id: string, updates: Partial<Vendor>) {
    const index = vendors.value.findIndex(v => v.id === id)
    if (index !== -1) {
      vendors.value[index] = {
        ...vendors.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return vendors.value[index]
    }
    return null
  }

  function updateVendorRating(id: string, rating: number, onTimeDeliveryRate?: number, qualityRating?: number) {
    const vendor = vendors.value.find(v => v.id === id)
    if (vendor) {
      vendor.rating = rating
      if (onTimeDeliveryRate !== undefined) vendor.onTimeDeliveryRate = onTimeDeliveryRate
      if (qualityRating !== undefined) vendor.qualityRating = qualityRating
      vendor.updatedAt = new Date().toISOString()
    }
  }

  function updateVendorPurchaseValue(id: string, additionalValue: number) {
    const vendor = vendors.value.find(v => v.id === id)
    if (vendor) {
      vendor.totalPurchaseValue = (vendor.totalPurchaseValue || 0) + additionalValue
      vendor.lastPurchaseDate = new Date().toISOString()
      vendor.updatedAt = new Date().toISOString()
    }
  }

  function deleteVendor(id: string) {
    const index = vendors.value.findIndex(v => v.id === id)
    if (index !== -1) {
      vendors.value.splice(index, 1)
      return true
    }
    return false
  }

  function suspendVendor(id: string, reason: string) {
    const vendor = vendors.value.find(v => v.id === id)
    if (vendor) {
      vendor.status = 'suspended'
      vendor.notes = (vendor.notes || '') + `\n[SUSPENDED: ${new Date().toLocaleDateString()}] ${reason}`
      vendor.updatedAt = new Date().toISOString()
    }
  }

  function activateVendor(id: string) {
    const vendor = vendors.value.find(v => v.id === id)
    if (vendor) {
      vendor.status = 'active'
      vendor.updatedAt = new Date().toISOString()
    }
  }

  return {
    vendors,
    activeVendors,
    vendorsByCategory,
    topRatedVendors,
    vendorStats,
    getVendorById,
    getVendorByCode,
    getActiveVendors,
    getVendorsByCategory,
    addVendor,
    updateVendor,
    updateVendorRating,
    updateVendorPurchaseValue,
    deleteVendor,
    suspendVendor,
    activateVendor
  }
})
