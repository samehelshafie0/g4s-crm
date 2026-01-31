import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Status } from '@/types'

export type RecurringServiceType = 'product' | 'labor'
export type RecurringServiceSource = 'owned' | 'rented' // owned = from your products/services, rented = from vendor

export interface VendorQuote {
  id: string
  vendorId: string
  vendorName: string
  quoteNumber: string
  quoteDate: string
  validUntil: string
  monthlyRate: number
  minimumPeriodMonths: number
  setupCost?: number
  notes?: string
  attachmentUrl?: string // For storing quote document
  attachmentFileName?: string
  // Extracted data from uploaded file
  extractedItems?: Array<{
    sku: string
    itemName: string
    price: number
    description?: string
  }>
}

export interface VendorProduct {
  sku: string
  name: string
  description?: string
  monthlyRate: number
  setupCost?: number
}

export interface RecurringService {
  id: string
  name: string
  description: string
  type: RecurringServiceType
  source: RecurringServiceSource // 'owned' or 'rented'

  // Source reference (for owned products/services from your inventory)
  sourceId?: string // Product ID or Labor Position ID (optional for rented)
  sourceName: string

  // Vendor Product Information (only for source='rented')
  vendorProduct?: VendorProduct

  // Pricing (what you charge to customer)
  monthlyPrice: number
  minimumMonths: number
  setupFee?: number

  // Costing (for margin calculation)
  monthlyCost: number

  // Vendor Rental Information (only for source='rented')
  vendorQuote?: VendorQuote
  currentRentalPeriod?: {
    startDate: string
    endDate: string
    monthsRemaining: number
  }
  rentalHistory?: Array<{
    periodId: string
    startDate: string
    endDate: string
    vendorQuoteId: string
    monthlyRate: number
    totalPaid: number
  }>

  // Status
  status: Status

  // Metadata
  createdAt: string
  createdBy: string
  updatedAt?: string
  notes?: string
}

export const useRecurringServicesStore = defineStore('recurringServices', () => {
  const recurringServices = ref<RecurringService[]>([
    {
      id: 'REC-001',
      name: 'Security Guard - Monthly Service',
      description: 'Monthly service for basic security guard from our team (160 hours/month)',
      type: 'labor',
      source: 'owned',
      sourceId: 'LP-003',
      sourceName: 'Security Guard - Grade C',
      monthlyPrice: 6500,
      minimumMonths: 3,
      setupFee: 500,
      monthlyCost: 4480,
      status: 'active',
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Standard 8-hour shifts, 20 working days per month - Our own staff'
    },
    {
      id: 'REC-002',
      name: 'CCTV Camera - Rental from Hikvision Vendor',
      description: 'Monthly rental for Hikvision 8MP IP Dome Camera rented from vendor',
      type: 'product',
      source: 'rented',
      sourceId: 'PRD-001',
      sourceName: 'Hikvision 8MP IP Dome Camera',
      monthlyPrice: 250,
      minimumMonths: 12,
      setupFee: 1000,
      monthlyCost: 150,
      vendorQuote: {
        id: 'VQ-001',
        vendorId: 'VEN-001',
        vendorName: 'Hikvision MENA',
        quoteNumber: 'HIK-Q-2024-001',
        quoteDate: '2024-01-10',
        validUntil: '2024-12-31',
        monthlyRate: 150,
        minimumPeriodMonths: 12,
        setupCost: 500,
        notes: 'Bulk rental agreement for 50+ cameras',
        attachmentUrl: '/quotes/HIK-Q-2024-001.pdf'
      },
      currentRentalPeriod: {
        startDate: '2024-01-15',
        endDate: '2025-01-15',
        monthsRemaining: 10
      },
      status: 'active',
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Rented from Hikvision - Back-to-back rental needed'
    },
    {
      id: 'REC-003',
      name: 'Security Supervisor - Monthly Service',
      description: 'Monthly security supervisor service from our team (160 hours/month)',
      type: 'labor',
      source: 'owned',
      sourceId: 'LP-004',
      sourceName: 'Security Supervisor',
      monthlyPrice: 9500,
      minimumMonths: 6,
      setupFee: 1000,
      monthlyCost: 6800,
      status: 'active',
      createdAt: '2024-01-16T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Includes reporting, team management, and incident coordination - Our own staff'
    },
    {
      id: 'REC-004',
      name: 'Access Control System - Rental from ZKTeco',
      description: 'Complete access control system rented from ZKTeco vendor (per 5 doors)',
      type: 'product',
      source: 'rented',
      sourceId: 'PRD-006',
      sourceName: 'ZKTeco inBio460 Access Control Panel',
      monthlyPrice: 800,
      minimumMonths: 24,
      setupFee: 3000,
      monthlyCost: 500,
      vendorQuote: {
        id: 'VQ-002',
        vendorId: 'VEN-002',
        vendorName: 'ZKTeco Middle East',
        quoteNumber: 'ZK-Q-2024-015',
        quoteDate: '2024-01-12',
        validUntil: '2025-01-12',
        monthlyRate: 500,
        minimumPeriodMonths: 24,
        setupCost: 2000,
        notes: 'Long-term rental agreement with maintenance included'
      },
      currentRentalPeriod: {
        startDate: '2024-01-17',
        endDate: '2026-01-17',
        monthsRemaining: 22
      },
      status: 'active',
      createdAt: '2024-01-17T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Rented from ZKTeco - Includes installation, configuration, and monthly maintenance'
    },
    {
      id: 'REC-005',
      name: 'Network Video Recorder - Our Inventory',
      description: 'NVR rental from our own inventory for up to 32 cameras with storage',
      type: 'product',
      source: 'owned',
      sourceId: 'PRD-003',
      sourceName: 'Hikvision 32CH 4K NVR',
      monthlyPrice: 450,
      minimumMonths: 12,
      setupFee: 800,
      monthlyCost: 280,
      status: 'active',
      createdAt: '2024-01-18T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'From our inventory - Includes 4TB storage, setup, and configuration'
    }
  ])

  // Computed
  const activeRecurringServices = computed(() =>
    recurringServices.value.filter(rs => rs.status === 'active')
  )

  const recurringServicesByType = computed(() => {
    const grouped: Record<RecurringServiceType, RecurringService[]> = {
      product: [],
      labor: []
    }

    recurringServices.value.forEach(rs => {
      grouped[rs.type].push(rs)
    })

    return grouped
  })

  // Getters
  function getRecurringServiceById(id: string): RecurringService | undefined {
    return recurringServices.value.find(rs => rs.id === id)
  }

  function getActiveRecurringServices(): RecurringService[] {
    return activeRecurringServices.value
  }

  function getRecurringServicesByType(type: RecurringServiceType): RecurringService[] {
    return recurringServices.value.filter(rs => rs.type === type && rs.status === 'active')
  }

  // CRUD Operations
  function addRecurringService(serviceData: Omit<RecurringService, 'id' | 'createdAt' | 'createdBy'>): RecurringService {
    const newService: RecurringService = {
      ...serviceData,
      id: `REC-${String(recurringServices.value.length + 1).padStart(3, '0')}`,
      createdAt: new Date().toISOString(),
      createdBy: 'USR-001' // Should come from auth
    }
    recurringServices.value.push(newService)
    return newService
  }

  function updateRecurringService(id: string, updates: Partial<RecurringService>): boolean {
    const index = recurringServices.value.findIndex(rs => rs.id === id)
    if (index !== -1) {
      recurringServices.value[index] = {
        ...recurringServices.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return true
    }
    return false
  }

  function deleteRecurringService(id: string): boolean {
    const index = recurringServices.value.findIndex(rs => rs.id === id)
    if (index !== -1) {
      recurringServices.value.splice(index, 1)
      return true
    }
    return false
  }

  function toggleRecurringServiceStatus(id: string): boolean {
    const service = recurringServices.value.find(rs => rs.id === id)
    if (service) {
      service.status = service.status === 'active' ? 'inactive' : 'active'
      service.updatedAt = new Date().toISOString()
      return true
    }
    return false
  }

  // Calculate total cost for a period
  function calculateTotalCost(serviceId: string, months: number, includeSetup: boolean = true): number {
    const service = getRecurringServiceById(serviceId)
    if (!service) return 0

    let total = service.monthlyPrice * months
    if (includeSetup && service.setupFee) {
      total += service.setupFee
    }
    return total
  }

  // Calculate margin
  function calculateMarginPercent(serviceId: string): number {
    const service = getRecurringServiceById(serviceId)
    if (!service || service.monthlyPrice === 0) return 0

    return ((service.monthlyPrice - service.monthlyCost) / service.monthlyPrice) * 100
  }

  // Get services by source
  function getServicesBySource(source: RecurringServiceSource): RecurringService[] {
    return recurringServices.value.filter(rs => rs.source === source && rs.status === 'active')
  }

  // Get rented services with expiring rentals (within next 3 months)
  function getExpiringRentals(withinMonths: number = 3): RecurringService[] {
    const now = new Date()
    const thresholdDate = new Date(now.setMonth(now.getMonth() + withinMonths))

    return recurringServices.value.filter(rs => {
      if (rs.source !== 'rented' || !rs.currentRentalPeriod) return false
      const endDate = new Date(rs.currentRentalPeriod.endDate)
      return endDate <= thresholdDate
    })
  }

  // Update vendor quote for a rented service
  function updateVendorQuote(serviceId: string, quote: VendorQuote): boolean {
    const service = recurringServices.value.find(rs => rs.id === serviceId)
    if (!service || service.source !== 'rented') return false

    service.vendorQuote = quote
    service.monthlyCost = quote.monthlyRate
    service.updatedAt = new Date().toISOString()
    return true
  }

  // Start new rental period (for back-to-back renewals)
  function startNewRentalPeriod(
    serviceId: string,
    startDate: string,
    months: number,
    vendorQuoteId?: string
  ): boolean {
    const service = recurringServices.value.find(rs => rs.id === serviceId)
    if (!service || service.source !== 'rented') return false

    // Archive current period to history
    if (service.currentRentalPeriod) {
      if (!service.rentalHistory) service.rentalHistory = []
      service.rentalHistory.push({
        periodId: `RP-${Date.now()}`,
        startDate: service.currentRentalPeriod.startDate,
        endDate: service.currentRentalPeriod.endDate,
        vendorQuoteId: service.vendorQuote?.id || '',
        monthlyRate: service.monthlyCost,
        totalPaid: service.monthlyCost * (
          (new Date(service.currentRentalPeriod.endDate).getTime() -
           new Date(service.currentRentalPeriod.startDate).getTime()) /
          (1000 * 60 * 60 * 24 * 30.44)
        )
      })
    }

    // Set new rental period
    const start = new Date(startDate)
    const end = new Date(start)
    end.setMonth(end.getMonth() + months)

    service.currentRentalPeriod = {
      startDate: start.toISOString().split('T')[0],
      endDate: end.toISOString().split('T')[0],
      monthsRemaining: months
    }

    service.updatedAt = new Date().toISOString()
    return true
  }

  // Calculate months remaining in current rental
  function updateRemainingMonths(serviceId: string): boolean {
    const service = recurringServices.value.find(rs => rs.id === serviceId)
    if (!service || service.source !== 'rented' || !service.currentRentalPeriod) return false

    const now = new Date()
    const endDate = new Date(service.currentRentalPeriod.endDate)
    const monthsRemaining = Math.max(0, Math.ceil((endDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24 * 30.44)))

    service.currentRentalPeriod.monthsRemaining = monthsRemaining
    return true
  }

  return {
    recurringServices,
    activeRecurringServices,
    recurringServicesByType,
    getRecurringServiceById,
    getActiveRecurringServices,
    getRecurringServicesByType,
    addRecurringService,
    updateRecurringService,
    deleteRecurringService,
    toggleRecurringServiceStatus,
    calculateTotalCost,
    calculateMarginPercent,
    getServicesBySource,
    getExpiringRentals,
    updateVendorQuote,
    startNewRentalPeriod,
    updateRemainingMonths
  }
})
