import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Company, CompanySite } from '@/types'

// Keep old types for backward compatibility, map to new types
export type Customer = Company

export interface CustomerHistory {
  customerId: string
  totalQuotes: number
  wonQuotes: number
  lostQuotes: number
  totalRevenue: number
  avgMargin: number
  lastQuoteDate: string
}

export const useCustomersStore = defineStore('customers', () => {
  const customers = ref<Company[]>([
    {
      id: 'CUST-001',
      name: 'ABC Corporation',
      sector: 'Technology',
      regions: ['Riyadh', 'Eastern Province'],
      cities: ['Riyadh', 'Dammam', 'Khobar'],
      contactPerson: 'Ahmed Al-Salem',
      email: 'info@abc-corp.com',
      phone: '+966 11 234 5678',
      website: 'www.abc-corp.com',
      address: 'King Fahd Road, Riyadh 12271',
      crNumber: 'CR-1234567890',
      vatNumber: 'VAT-300123456700003',
      status: 'get',
      createdAt: '2023-06-15T10:00:00Z',
      createdBy: 'admin',
      notes: 'Key account - enterprise solutions'
    },
    {
      id: 'CUST-002',
      name: 'XYZ Industries',
      sector: 'Retail & Manufacturing',
      regions: ['Makkah'],
      cities: ['Jeddah', 'Makkah'],
      contactPerson: 'Fatima Al-Otaibi',
      email: 'info@xyz-ind.com',
      phone: '+966 12 345 6789',
      website: 'www.xyz-industries.com',
      address: 'Industrial Area, Jeddah 23541',
      crNumber: 'CR-9876543210',
      vatNumber: 'VAT-300987654300003',
      status: 'grow',
      createdAt: '2023-08-22T09:30:00Z',
      createdBy: 'admin',
      notes: 'Manufacturing equipment supplier'
    },
    {
      id: 'CUST-003',
      name: 'Global Solutions',
      sector: 'Healthcare',
      regions: ['Eastern Province'],
      cities: ['Dammam', 'Dhahran', 'Khobar'],
      contactPerson: 'Mohammed Al-Qahtani',
      email: 'info@global-sol.com',
      phone: '+966 13 456 7890',
      website: 'www.globalsolutions.com',
      address: 'Medical District, Dammam 32254',
      crNumber: 'CR-5555444433',
      vatNumber: 'VAT-300555444433003',
      status: 'get',
      createdAt: '2023-09-10T11:00:00Z',
      createdBy: 'admin',
      notes: 'Healthcare IT infrastructure'
    },
    {
      id: 'CUST-004',
      name: 'National Bank',
      sector: 'Banking & Finance',
      regions: ['Riyadh', 'Makkah', 'Eastern Province'],
      cities: ['Riyadh', 'Jeddah', 'Dammam'],
      contactPerson: 'Khalid Al-Mutairi',
      email: 'it@nationalbank.com',
      phone: '+966 11 888 9999',
      website: 'www.nationalbank.com',
      address: 'Financial District, Riyadh 11564',
      crNumber: 'CR-1111222233',
      vatNumber: 'VAT-301111222233003',
      status: 'grow',
      createdAt: '2023-07-01T08:00:00Z',
      createdBy: 'admin',
      notes: 'Financial sector customer'
    },
    {
      id: 'CUST-005',
      name: 'Royal Hotel Group',
      sector: 'Hospitality',
      regions: ['Makkah', 'Madinah'],
      cities: ['Makkah', 'Madinah'],
      contactPerson: 'Nora Al-Harbi',
      email: 'operations@royalhotel.com',
      phone: '+966 12 777 8888',
      website: 'www.royalhotelgroup.com',
      address: 'Aziziyah District, Makkah 24231',
      crNumber: 'CR-6666777788',
      vatNumber: 'VAT-306666777788003',
      status: 'get',
      createdAt: '2023-10-15T13:00:00Z',
      createdBy: 'admin',
      notes: 'Hospitality sector - seasonal demands'
    }
  ])

  const sites = ref<CompanySite[]>([
    {
      id: 'SITE-001',
      companyId: 'CUST-001',
      name: 'ABC Corp - Headquarters',
      region: 'Riyadh',
      address: 'King Fahd Road, Riyadh 12271',
      city: 'Riyadh',
      isPrimary: true,
      contactPerson: 'Ahmed Al-Salem',
      phone: '+966 11 234 5678',
      email: 'riyadh@abc-corp.com',
      createdAt: '2023-06-15T10:00:00Z'
    },
    {
      id: 'SITE-002',
      companyId: 'CUST-001',
      name: 'ABC Corp - Jeddah Branch',
      region: 'Jeddah',
      address: 'Corniche Road, Jeddah 23433',
      city: 'Jeddah',
      isPrimary: false,
      contactPerson: 'Sara Al-Otaibi',
      phone: '+966 12 555 6666',
      email: 'jeddah@abc-corp.com',
      createdAt: '2023-07-20T10:00:00Z',
      notes: 'Branch office opened in July 2023'
    },
    {
      id: 'SITE-003',
      companyId: 'CUST-002',
      name: 'XYZ Industries - Main Factory',
      region: 'Jeddah',
      address: 'Industrial Area, Jeddah 23541',
      city: 'Jeddah',
      isPrimary: true,
      contactPerson: 'Fatima Al-Otaibi',
      phone: '+966 12 345 6789',
      email: 'factory@xyz-ind.com',
      createdAt: '2023-08-22T09:30:00Z'
    },
    {
      id: 'SITE-004',
      companyId: 'CUST-002',
      name: 'XYZ Industries - Warehouse',
      region: 'Dammam',
      address: 'Logistics Zone, Dammam 32234',
      city: 'Dammam',
      isPrimary: false,
      contactPerson: 'Khalid Al-Zahrani',
      phone: '+966 13 777 8888',
      email: 'warehouse@xyz-ind.com',
      createdAt: '2023-09-01T08:00:00Z',
      notes: 'Primary distribution center'
    },
    {
      id: 'SITE-005',
      companyId: 'CUST-003',
      name: 'Global Solutions - Head Office',
      region: 'Dammam',
      address: 'Medical District, Dammam 32254',
      city: 'Dammam',
      isPrimary: true,
      contactPerson: 'Mohammed Al-Qahtani',
      phone: '+966 13 456 7890',
      email: 'hq@global-sol.com',
      createdAt: '2023-09-10T11:00:00Z'
    },
    {
      id: 'SITE-006',
      companyId: 'CUST-004',
      name: 'National Bank - Main Branch',
      region: 'Riyadh',
      address: 'Financial District, Riyadh 11564',
      city: 'Riyadh',
      isPrimary: true,
      contactPerson: 'Khalid Al-Mutairi',
      phone: '+966 11 888 9999',
      email: 'mainbranch@nationalbank.com',
      createdAt: '2023-07-01T08:00:00Z'
    }
  ])

  const customerHistory = ref<Record<string, CustomerHistory>>({
    'CUST-001': {
      customerId: 'CUST-001',
      totalQuotes: 45,
      wonQuotes: 31,
      lostQuotes: 14,
      totalRevenue: 4500000,
      avgMargin: 18.5,
      lastQuoteDate: '2024-01-20'
    },
    'CUST-002': {
      customerId: 'CUST-002',
      totalQuotes: 32,
      wonQuotes: 18,
      lostQuotes: 14,
      totalRevenue: 2800000,
      avgMargin: 15.2,
      lastQuoteDate: '2024-01-18'
    },
    'CUST-003': {
      customerId: 'CUST-003',
      totalQuotes: 28,
      wonQuotes: 20,
      lostQuotes: 8,
      totalRevenue: 3200000,
      avgMargin: 21.3,
      lastQuoteDate: '2024-01-22'
    },
    'CUST-004': {
      customerId: 'CUST-004',
      totalQuotes: 52,
      wonQuotes: 41,
      lostQuotes: 11,
      totalRevenue: 6200000,
      avgMargin: 22.8,
      lastQuoteDate: '2024-01-24'
    },
    'CUST-005': {
      customerId: 'CUST-005',
      totalQuotes: 18,
      wonQuotes: 12,
      lostQuotes: 6,
      totalRevenue: 1800000,
      avgMargin: 19.5,
      lastQuoteDate: '2024-01-15'
    }
  })

  const activeCustomers = computed(() => {
    return customers.value.filter(c => c.status === 'active')
  })

  const sectors = computed(() => {
    return Array.from(new Set(customers.value.map(c => c.sector)))
  })

  const regions = computed(() => {
    return Array.from(new Set(customers.value.map(c => c.region)))
  })

  function addCustomer(customer: Company) {
    customers.value.push(customer)
  }

  function updateCustomer(id: string, updates: Partial<Company>) {
    const index = customers.value.findIndex(c => c.id === id)
    if (index !== -1) {
      customers.value[index] = { ...customers.value[index], ...updates }
    }
  }

  function deleteCustomer(id: string) {
    const index = customers.value.findIndex(c => c.id === id)
    if (index !== -1) {
      customers.value.splice(index, 1)
    }
  }

  function getCustomerById(id: string) {
    return customers.value.find(c => c.id === id)
  }

  function getSitesByCustomerId(customerId: string) {
    return sites.value.filter(s => s.companyId === customerId)
  }

  function addSite(site: CompanySite) {
    sites.value.push(site)
  }

  function updateSite(id: string, updates: Partial<CompanySite>) {
    const index = sites.value.findIndex(s => s.id === id)
    if (index !== -1) {
      sites.value[index] = { ...sites.value[index], ...updates }
    }
  }

  function deleteSite(id: string) {
    const index = sites.value.findIndex(s => s.id === id)
    if (index !== -1) {
      sites.value.splice(index, 1)
    }
  }

  function getCustomerHistory(customerId: string) {
    return customerHistory.value[customerId]
  }

  return {
    customers,
    sites,
    customerHistory,
    activeCustomers,
    sectors,
    regions,
    addCustomer,
    updateCustomer,
    deleteCustomer,
    getCustomerById,
    getSitesByCustomerId,
    addSite,
    updateSite,
    deleteSite,
    getCustomerHistory
  }
})
