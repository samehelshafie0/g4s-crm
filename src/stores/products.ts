import { ref } from 'vue'
import { defineStore } from 'pinia'
import { useExchangeRatesStore } from './exchangeRates'
import type { Product, ProductCategory, Currency } from '@/types'

export const useProductsStore = defineStore('products', () => {
  // Use centralized exchange rates store
  const exchangeRatesStore = useExchangeRatesStore()
  const products = ref<Product[]>([
    {
      id: 'PRD-001',
      sku: 'CAM-HK-DS2CD2385G1',
      name: 'Hikvision 8MP IP Dome Camera',
      description: '8MP (4K) outdoor IP dome camera with IR, WDR, H.265+ compression',
      category: 'CCTV Systems',
      vendorId: 'VEN-001',
      supplier: 'Hikvision MENA',
      supplierPartNumber: 'DS-2CD2385G1-I',
      manufacturer: 'Hikvision',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 180,
      fxRate: 3.75,
      costInSAR: 675,
      freightPercent: 5,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 756,
      targetMarginPercent: 25,
      sellingPrice: 1008,
      unitOfMeasure: 'Unit',
      leadTimeDays: 30,
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Popular model for outdoor applications, supports ONVIF'
    },
    {
      id: 'PRD-002',
      sku: 'CAM-DH-IPC8542E',
      name: 'Dahua 5MP Bullet Camera',
      description: '5MP bullet camera with motorized lens and smart detection',
      category: 'CCTV Systems',
      supplier: 'Dahua Technology',
      supplierPartNumber: 'IPC-HFW8542E-ZE',
      manufacturer: 'Dahua',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 220,
      fxRate: 3.75,
      costInSAR: 825,
      freightPercent: 5,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 924,
      targetMarginPercent: 25,
      sellingPrice: 1232,
      unitOfMeasure: 'Unit',
      leadTimeDays: 25,
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Advanced AI features, perimeter protection'
    },
    {
      id: 'PRD-003',
      sku: 'NVR-HK-DS96256NI',
      name: 'Hikvision 256-Channel NVR',
      description: 'Network Video Recorder with 256 channels, 12MP recording resolution',
      category: 'CCTV Systems',
      supplier: 'Hikvision MENA',
      supplierPartNumber: 'DS-96256NI-I24',
      manufacturer: 'Hikvision',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 3500,
      fxRate: 3.75,
      costInSAR: 13125,
      freightPercent: 8,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 15094,
      targetMarginPercent: 22,
      sellingPrice: 19351,
      unitOfMeasure: 'Unit',
      leadTimeDays: 45,
      createdAt: '2024-01-15T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Enterprise-grade NVR for large installations'
    },
    {
      id: 'PRD-004',
      sku: 'ACS-ZK-C3400',
      name: 'ZKTeco C3-400 Access Controller',
      description: '4-door access controller with TCP/IP, supports 100K users',
      category: 'Access Control Systems',
      supplier: 'ZKTeco Middle East',
      supplierPartNumber: 'C3-400',
      manufacturer: 'ZKTeco',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 450,
      fxRate: 3.75,
      costInSAR: 1687.5,
      freightPercent: 6,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 1907,
      targetMarginPercent: 23,
      minimumMarginPercent: 17,
      sellingPrice: 2477,
      unitOfMeasure: 'Unit',
      leadTimeDays: 35,
      createdAt: '2024-01-16T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Scalable solution for multi-door applications'
    },
    {
      id: 'PRD-005',
      sku: 'ACS-HID-5365',
      name: 'HID Proximity Card Reader',
      description: 'ProxPoint Plus wall-mount card reader, Wiegand output',
      category: 'Access Control Systems',
      supplier: 'HID Global',
      supplierPartNumber: '5365EGK00',
      manufacturer: 'HID Global',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 85,
      fxRate: 3.75,
      costInSAR: 318.75,
      freightPercent: 5,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 357,
      targetMarginPercent: 28,
      minimumMarginPercent: 20,
      sellingPrice: 496,
      unitOfMeasure: 'Unit',
      leadTimeDays: 20,
      createdAt: '2024-01-16T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Industry standard reader, very reliable'
    },
    {
      id: 'PRD-006',
      sku: 'ALM-BOSCH-B9512G',
      name: 'Bosch B9512G Control Panel',
      description: 'Professional intrusion alarm control panel, 512 zones',
      category: 'Alarm Systems',
      supplier: 'Bosch Security',
      supplierPartNumber: 'B9512G',
      manufacturer: 'Bosch',
      isImport: true,
      originCurrency: 'EUR',
      unitCost: 890,
      fxRate: 4.10,
      costInSAR: 3649,
      freightPercent: 7,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 4160,
      targetMarginPercent: 24,
      minimumMarginPercent: 16,
      sellingPrice: 5474,
      unitOfMeasure: 'Unit',
      leadTimeDays: 40,
      createdAt: '2024-01-17T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Ideal for large commercial buildings'
    },
    {
      id: 'PRD-007',
      sku: 'FIRE-NOTIFIER-NFS320',
      name: 'Notifier NFS2-3030 Fire Alarm Panel',
      description: 'Intelligent fire alarm control panel, 636 addressable points',
      category: 'Fire Safety Equipment',
      supplier: 'Honeywell Fire Safety',
      supplierPartNumber: 'NFS2-3030',
      manufacturer: 'Notifier by Honeywell',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 2800,
      fxRate: 3.75,
      costInSAR: 10500,
      freightPercent: 8,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 12075,
      targetMarginPercent: 20,
      minimumMarginPercent: 14,
      sellingPrice: 15094,
      unitOfMeasure: 'Unit',
      leadTimeDays: 50,
      createdAt: '2024-01-17T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'UL/FM approved, meets Saudi Civil Defense requirements'
    },
    {
      id: 'PRD-008',
      sku: 'UNI-GUARD-SET01',
      name: 'Security Guard Uniform Set',
      description: 'Complete uniform set: shirt, pants, cap, belt, badge holder',
      category: 'Uniforms & Apparel',
      supplier: 'Saudi Uniform Factory',
      manufacturer: 'Local Saudi Manufacturer',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 180,
      fxRate: 1,
      costInSAR: 180,
      landedCostSAR: 180,
      targetMarginPercent: 30,
      minimumMarginPercent: 22,
      sellingPrice: 257,
      unitOfMeasure: 'Set',
      leadTimeDays: 7,
      createdAt: '2024-01-18T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Local production, fast delivery, multiple sizes available'
    },
    {
      id: 'PRD-009',
      sku: 'COM-MOT-DP4400',
      name: 'Motorola DP4400 Two-Way Radio',
      description: 'Digital portable radio, VHF/UHF, 1000 channels',
      category: 'Communication Equipment',
      supplier: 'Motorola Solutions ME',
      supplierPartNumber: 'DP4400e',
      manufacturer: 'Motorola Solutions',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 320,
      fxRate: 3.75,
      costInSAR: 1200,
      freightPercent: 5,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 1344,
      targetMarginPercent: 26,
      minimumMarginPercent: 19,
      sellingPrice: 1816,
      unitOfMeasure: 'Unit',
      leadTimeDays: 30,
      createdAt: '2024-01-18T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Requires CITC license for operation in Saudi Arabia'
    },
    {
      id: 'PRD-010',
      sku: 'VEH-PATROL-01',
      name: 'Mobile Patrol Vehicle - Toyota Hilux',
      description: 'Toyota Hilux Double Cab 4x4 with security branding and equipment',
      category: 'Vehicles',
      supplier: 'Abdul Latif Jameel',
      manufacturer: 'Toyota',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 95000,
      fxRate: 1,
      costInSAR: 95000,
      landedCostSAR: 95000,
      targetMarginPercent: 15,
      minimumMarginPercent: 10,
      sellingPrice: 111765,
      unitOfMeasure: 'Unit',
      leadTimeDays: 60,
      createdAt: '2024-01-19T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Includes branding, light bar, radio equipment'
    },
    {
      id: 'PRD-011',
      sku: 'SW-MS-VMS50',
      name: 'Milestone XProtect Corporate - 50 Camera License',
      description: 'Video management software license for 50 cameras, perpetual',
      category: 'Software & Licenses',
      supplier: 'Milestone Systems ME',
      supplierPartNumber: 'XPCOBASE50',
      manufacturer: 'Milestone Systems',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 4500,
      fxRate: 3.75,
      costInSAR: 16875,
      freightPercent: 0,
      customsPercent: 0,
      clearancePercent: 0,
      landedCostSAR: 16875,
      targetMarginPercent: 22,
      sellingPrice: 21635,
      unitOfMeasure: 'License',
      leadTimeDays: 5,
      createdAt: '2024-01-20T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Software license, delivered electronically'
    },
    {
      id: 'PRD-012',
      sku: 'SRV-INSTALL-CAM',
      name: 'Camera Installation Service',
      description: 'Professional installation service per camera including mounting, cabling, configuration',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 200,
      fxRate: 1,
      costInSAR: 200,
      landedCostSAR: 200,
      targetMarginPercent: 35,
      minimumMarginPercent: 25,
      sellingPrice: 308,
      unitOfMeasure: 'Unit',
      leadTimeDays: 1,
      createdAt: '2024-01-20T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Labor and materials for standard camera installation'
    },
    {
      id: 'PRD-013',
      sku: 'SVR-DELL-R750',
      name: 'Dell PowerEdge R750 Server',
      description: 'Rack server: 2x Xeon Gold, 128GB RAM, 4x 2TB SSD RAID10',
      category: 'Servers',
      supplier: 'Dell EMC Middle East',
      supplierPartNumber: 'PER750-001',
      manufacturer: 'Dell',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 8500,
      fxRate: 3.75,
      costInSAR: 31875,
      freightPercent: 6,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 36019,
      targetMarginPercent: 18,
      minimumMarginPercent: 12,
      sellingPrice: 43926,
      unitOfMeasure: 'Unit',
      leadTimeDays: 45,
      createdAt: '2024-01-21T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Enterprise server for VMS and access control applications'
    },
    {
      id: 'PRD-014',
      sku: 'NET-CISCO-SG350',
      name: 'Cisco SG350-28P PoE Switch',
      description: '28-port Gigabit PoE+ managed switch, 195W power budget',
      category: 'Networking',
      supplier: 'Cisco Middle East',
      supplierPartNumber: 'SG350-28P-K9',
      manufacturer: 'Cisco',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 650,
      fxRate: 3.75,
      costInSAR: 2437.5,
      freightPercent: 5,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 2730,
      targetMarginPercent: 24,
      minimumMarginPercent: 17,
      sellingPrice: 3592,
      unitOfMeasure: 'Unit',
      leadTimeDays: 25,
      createdAt: '2024-01-21T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Ideal for IP camera and access control networks'
    },
    {
      id: 'PRD-015',
      sku: 'STG-QNAP-TS1685',
      name: 'QNAP TS-1685 NAS Storage',
      description: '16-bay NAS with Xeon processor, 10GbE, 96TB capacity',
      category: 'Storage',
      supplier: 'QNAP Middle East',
      supplierPartNumber: 'TS-1685-D1531-64G',
      manufacturer: 'QNAP',
      isImport: true,
      originCurrency: 'USD',
      unitCost: 3200,
      fxRate: 3.75,
      costInSAR: 12000,
      freightPercent: 7,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 13680,
      targetMarginPercent: 21,
      sellingPrice: 17316,
      unitOfMeasure: 'Unit',
      leadTimeDays: 35,
      createdAt: '2024-01-22T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'High-capacity storage for video surveillance archives'
    },
    {
      id: 'PRD-016',
      sku: 'SEC-EQ-METAL-DET',
      name: 'Walk-Through Metal Detector',
      description: 'Multi-zone walk-through metal detector with alarm',
      category: 'Security Equipment',
      supplier: 'CEIA Middle East',
      supplierPartNumber: 'CEIA-HI-PE',
      manufacturer: 'CEIA',
      isImport: true,
      originCurrency: 'EUR',
      unitCost: 4200,
      fxRate: 4.10,
      costInSAR: 17220,
      freightPercent: 10,
      customsPercent: 5,
      clearancePercent: 2,
      landedCostSAR: 20147,
      targetMarginPercent: 20,
      minimumMarginPercent: 14,
      sellingPrice: 25184,
      unitOfMeasure: 'Unit',
      leadTimeDays: 60,
      createdAt: '2024-01-22T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'High-end metal detector for critical facilities'
    },
    {
      id: 'PRD-017',
      sku: 'SRV-GUARD-BASIC',
      name: 'Security Guard - Basic Level',
      description: 'Basic level security guard service (per guard per month)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 4500,
      fxRate: 1,
      costInSAR: 4500,
      landedCostSAR: 4500,
      targetMarginPercent: 30,
      minimumMarginPercent: 20,
      sellingPrice: 6429,
      unitOfMeasure: 'Month',
      leadTimeDays: 7,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - Basic level guard with standard training'
    },
    {
      id: 'PRD-018',
      sku: 'SRV-GUARD-SUPERVISOR',
      name: 'Security Supervisor',
      description: 'Security supervisor service (per supervisor per month)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 6500,
      fxRate: 1,
      costInSAR: 6500,
      landedCostSAR: 6500,
      targetMarginPercent: 28,
      sellingPrice: 9028,
      unitOfMeasure: 'Month',
      leadTimeDays: 7,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - Supervisor with management experience'
    },
    {
      id: 'PRD-019',
      sku: 'SRV-MAINT-CCTV',
      name: 'CCTV System Maintenance',
      description: 'Monthly CCTV system maintenance service (per 10 cameras)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 800,
      fxRate: 1,
      costInSAR: 800,
      landedCostSAR: 800,
      targetMarginPercent: 40,
      minimumMarginPercent: 30,
      sellingPrice: 1333,
      unitOfMeasure: 'Month',
      leadTimeDays: 3,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - Includes cleaning, testing, minor repairs'
    },
    {
      id: 'PRD-020',
      sku: 'SRV-MAINT-ACCESS',
      name: 'Access Control System Maintenance',
      description: 'Monthly access control system maintenance (per 5 doors)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 600,
      fxRate: 1,
      costInSAR: 600,
      landedCostSAR: 600,
      targetMarginPercent: 40,
      minimumMarginPercent: 30,
      sellingPrice: 1000,
      unitOfMeasure: 'Month',
      leadTimeDays: 3,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - Software updates, card management, testing'
    },
    {
      id: 'PRD-021',
      sku: 'SRV-PATROL-SERVICE',
      name: 'Mobile Patrol Service',
      description: 'Mobile patrol service (per location per month, 2 visits daily)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 3500,
      fxRate: 1,
      costInSAR: 3500,
      landedCostSAR: 3500,
      targetMarginPercent: 35,
      minimumMarginPercent: 25,
      sellingPrice: 5385,
      unitOfMeasure: 'Month',
      leadTimeDays: 5,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - 2 patrol visits daily with reporting'
    },
    {
      id: 'PRD-022',
      sku: 'SRV-MONITORING-24-7',
      name: '24/7 Monitoring Center Service',
      description: 'Remote 24/7 monitoring service (per site per month)',
      category: 'Professional Services',
      supplier: 'Internal',
      manufacturer: 'Internal',
      isImport: false,
      originCurrency: 'SAR',
      unitCost: 2500,
      fxRate: 1,
      costInSAR: 2500,
      landedCostSAR: 2500,
      targetMarginPercent: 45,
      minimumMarginPercent: 35,
      sellingPrice: 4545,
      unitOfMeasure: 'Month',
      leadTimeDays: 3,
      createdAt: '2024-01-23T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Monthly recurring service - 24/7 alarm monitoring with instant response'
    }
  ])

  // Exchange Rate Management (would typically come from API)
  const exchangeRates = ref<Record<Currency, number>>({
    SAR: 1,
    USD: 3.75,
    EUR: 4.10,
    GBP: 4.80,
    AED: 1.02,
    CNY: 0.52
  })

  // Getters
  function getProductById(id: string): Product | undefined {
    return products.value.find(p => p.id === id)
  }

  function getProductBySku(sku: string): Product | undefined {
    return products.value.find(p => p.sku === sku)
  }

  function getActiveProducts(): Product[] {
    return products.value // Return all products since Product type doesn't have status field
  }

  function getProductsByCategory(category: ProductCategory): Product[] {
    return products.value.filter(p => p.category === category)
  }

  function getImportProducts(): Product[] {
    return products.value.filter(p => p.isImport)
  }

  function getLocalProducts(): Product[] {
    return products.value.filter(p => !p.isImport)
  }

  // Calculation Utilities
  function calculateCostInSAR(originCurrency: Currency, unitCost: number): number {
    // Use centralized exchange rates from the exchange rates store
    const fxRate = exchangeRatesStore.getCurrentRate(originCurrency)
    const costInSAR = unitCost * fxRate
    // Always ceil (round up) the cost
    return Math.ceil(costInSAR)
  }

  function calculateLandedCost(
    costInSAR: number,
    isImport: boolean,
    freightPercent: number = 0,
    customsPercent: number = 0,
    clearancePercent: number = 0
  ): number {
    if (!isImport) return costInSAR

    const totalImportPercent = (freightPercent + customsPercent + clearancePercent) / 100
    const landedCost = costInSAR * (1 + totalImportPercent)
    // Always ceil (round up) the landed cost
    return Math.ceil(landedCost)
  }

  function calculateSellingPrice(landedCost: number, targetMarginPercent: number): number {
    // Selling Price = Landed Cost / (1 - Target Margin%)
    const price = landedCost / (1 - targetMarginPercent / 100)
    // Always ceil (round up) the price
    return Math.ceil(price)
  }

  function calculateMarginPercent(cost: number, sellingPrice: number): number {
    if (sellingPrice === 0) return 0
    return ((sellingPrice - cost) / sellingPrice) * 100
  }

  function calculateMarginAmount(cost: number, sellingPrice: number): number {
    return sellingPrice - cost
  }

  // CRUD Operations
  function addProduct(productData: Omit<Product, 'id' | 'createdAt' | 'createdBy'>): Product {
    // Calculate derived values
    const costInSAR = calculateCostInSAR(productData.originCurrency, productData.unitCost)
    const landedCostSAR = calculateLandedCost(
      costInSAR,
      productData.isImport,
      productData.freightPercent,
      productData.customsPercent,
      productData.clearancePercent
    )
    const sellingPrice = calculateSellingPrice(landedCostSAR, productData.targetMarginPercent)

    const newProduct: Product = {
      ...productData,
      id: `PRD-${String(products.value.length + 1).padStart(3, '0')}`,
      costInSAR,
      landedCostSAR,
      sellingPrice,
      fxRate: exchangeRatesStore.getCurrentRate(productData.originCurrency),
      createdAt: new Date().toISOString(),
      createdBy: 'USR-001' // Should come from auth
    }

    products.value.push(newProduct)
    return newProduct
  }

  function updateProduct(id: string, updates: Partial<Product>): boolean {
    const index = products.value.findIndex(p => p.id === id)
    if (index !== -1) {
      const currentProduct = products.value[index]
      const updatedProduct = { ...currentProduct, ...updates } as Product

      // Recalculate if costing-related fields changed
      if (
        updates.unitCost !== undefined ||
        updates.originCurrency !== undefined ||
        updates.freightPercent !== undefined ||
        updates.customsPercent !== undefined ||
        updates.clearancePercent !== undefined ||
        updates.targetMarginPercent !== undefined
      ) {
        updatedProduct.fxRate = exchangeRates.value[updatedProduct.originCurrency]
        updatedProduct.costInSAR = calculateCostInSAR(
          updatedProduct.originCurrency,
          updatedProduct.unitCost
        )
        updatedProduct.landedCostSAR = calculateLandedCost(
          updatedProduct.costInSAR,
          updatedProduct.isImport,
          updatedProduct.freightPercent || 0,
          updatedProduct.customsPercent || 0,
          updatedProduct.clearancePercent || 0
        )
        updatedProduct.sellingPrice = calculateSellingPrice(
          updatedProduct.landedCostSAR,
          updatedProduct.targetMarginPercent
        )
      }

      products.value[index] = updatedProduct
      return true
    }
    return false
  }

  function deleteProduct(id: string): boolean {
    const index = products.value.findIndex(p => p.id === id)
    if (index !== -1) {
      products.value.splice(index, 1)
      return true
    }
    return false
  }

  function toggleProductStatus(id: string): boolean {
    const product = products.value.find(p => p.id === id)
    if (product) {
      product.status = product.status === 'active' ? 'inactive' : 'active'
      return true
    }
    return false
  }

  // Analytics
  function getProductStats() {
    const activeProducts = getActiveProducts()
    const importProducts = getImportProducts()

    const avgMargin = activeProducts.length > 0
      ? activeProducts.reduce((sum, p) =>
          sum + calculateMarginPercent(p.landedCostSAR, p.sellingPrice), 0
        ) / activeProducts.length
      : 0

    const totalInventoryValue = activeProducts.reduce((sum, p) => sum + p.landedCostSAR, 0)

    return {
      totalProducts: products.value.length,
      activeProducts: activeProducts.length,
      importProducts: importProducts.length,
      localProducts: products.value.length - importProducts.length,
      avgMargin,
      totalInventoryValue
    }
  }

  return {
    products,
    exchangeRates,
    getProductById,
    getProductBySku,
    getActiveProducts,
    getProductsByCategory,
    getImportProducts,
    getLocalProducts,
    calculateCostInSAR,
    calculateLandedCost,
    calculateSellingPrice,
    calculateMarginPercent,
    calculateMarginAmount,
    addProduct,
    updateProduct,
    deleteProduct,
    toggleProductStatus,
    getProductStats
  }
})
