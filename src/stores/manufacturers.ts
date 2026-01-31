import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Manufacturer, ManufacturerCategory, ManufacturerProduct, ProductCategory } from '@/types'

export const useManufacturersStore = defineStore('manufacturers', () => {
  const manufacturers = ref<Manufacturer[]>([
    {
      id: 'MFG-001',
      name: 'Hikvision',
      code: 'HIK',
      description: 'Leading provider of security products and solutions',
      country: 'China',
      website: 'https://www.hikvision.com',
      contactEmail: 'info@hikvision.com',
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      categories: [
        {
          id: 'CAT-HIK-001',
          name: 'CCTV Systems',
          description: 'IP Cameras, NVRs, and Video Surveillance',
          products: [
            {
              id: 'PROD-HIK-001',
              name: '8MP IR Fixed Bullet Network Camera',
              modelNumber: 'DS-2CD2385G1-I',
              description: '8 MP high performance IR fixed bullet network camera',
              category: 'CCTV Systems',
              specifications: '8MP, 2.8mm/4mm/6mm lens, H.265+, IR 30m',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-HIK-002',
              name: '4MP IR Fixed Dome Network Camera',
              modelNumber: 'DS-2CD2143G0-I',
              description: '4 MP IR fixed dome network camera',
              category: 'CCTV Systems',
              specifications: '4MP, 2.8mm/4mm/6mm lens, H.265+, IR 30m',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-HIK-003',
              name: '32-Channel NVR',
              modelNumber: 'DS-7732NI-I4',
              description: '32-channel embedded NVR',
              category: 'CCTV Systems',
              specifications: '32CH, 4K output, H.265+, 4 SATA',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        },
        {
          id: 'CAT-HIK-002',
          name: 'Access Control Systems',
          description: 'Access controllers and readers',
          products: [
            {
              id: 'PROD-HIK-004',
              name: 'Network Access Controller',
              modelNumber: 'DS-K2604',
              description: '4-door network access controller',
              category: 'Access Control Systems',
              specifications: '4-door, TCP/IP, RS-485',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-HIK-005',
              name: 'Card Reader',
              modelNumber: 'DS-K1107E',
              description: 'EM card reader',
              category: 'Access Control Systems',
              specifications: 'EM 125KHz, Wiegand 26/34',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        }
      ]
    },
    {
      id: 'MFG-002',
      name: 'Dahua',
      code: 'DAH',
      description: 'Video-centric smart IoT solution and service provider',
      country: 'China',
      website: 'https://www.dahuasecurity.com',
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      categories: [
        {
          id: 'CAT-DAH-001',
          name: 'CCTV Systems',
          description: 'IP and Analog cameras',
          products: [
            {
              id: 'PROD-DAH-001',
              name: '8MP IR Bullet Camera',
              modelNumber: 'IPC-HFW2831S-S-S2',
              description: '8MP IR Fixed-focal Bullet Network Camera',
              category: 'CCTV Systems',
              specifications: '8MP, 2.8mm/3.6mm lens, H.265, IR 40m',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-DAH-002',
              name: '16-Channel NVR',
              modelNumber: 'NVR4216-16P-4KS2',
              description: '16-channel 1U 16PoE 4K H.265 Lite Network Video Recorder',
              category: 'CCTV Systems',
              specifications: '16CH, 4K, H.265, 16 PoE ports',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        },
        {
          id: 'CAT-DAH-002',
          name: 'Alarm Systems',
          description: 'Intrusion detection and alarm panels',
          products: [
            {
              id: 'PROD-DAH-003',
              name: 'Wireless Alarm Control Panel',
              modelNumber: 'ARC3000B-W8',
              description: 'Wireless/Wired Alarm Control Panel',
              category: 'Alarm Systems',
              specifications: '8 wireless zones, GSM/GPRS',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        }
      ]
    },
    {
      id: 'MFG-003',
      name: 'Dell',
      code: 'DELL',
      description: 'Computer technology company',
      country: 'USA',
      website: 'https://www.dell.com',
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      categories: [
        {
          id: 'CAT-DELL-001',
          name: 'Servers',
          description: 'Enterprise server solutions',
          products: [
            {
              id: 'PROD-DELL-001',
              name: 'PowerEdge R640',
              modelNumber: 'R640',
              description: '1U rack server',
              category: 'Servers',
              specifications: '1U, Dual Xeon, up to 3TB RAM',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-DELL-002',
              name: 'PowerEdge R740',
              modelNumber: 'R740',
              description: '2U rack server',
              category: 'Servers',
              specifications: '2U, Dual Xeon, up to 3TB RAM',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        },
        {
          id: 'CAT-DELL-002',
          name: 'Storage',
          description: 'Storage arrays and solutions',
          products: [
            {
              id: 'PROD-DELL-003',
              name: 'PowerVault ME4024',
              modelNumber: 'ME4024',
              description: 'SAN storage array',
              category: 'Storage',
              specifications: '2U, 24-bay, SAS/iSCSI',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        }
      ]
    },
    {
      id: 'MFG-004',
      name: 'Cisco',
      code: 'CISCO',
      description: 'Networking and cybersecurity solutions',
      country: 'USA',
      website: 'https://www.cisco.com',
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      categories: [
        {
          id: 'CAT-CISCO-001',
          name: 'Networking',
          description: 'Switches, routers, and wireless',
          products: [
            {
              id: 'PROD-CISCO-001',
              name: 'Catalyst 9300 Switch',
              modelNumber: 'C9300-48P',
              description: '48-port PoE+ switch',
              category: 'Networking',
              specifications: '48x1G PoE+, 4x10G uplinks',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-CISCO-002',
              name: 'ISR 4331 Router',
              modelNumber: 'ISR4331/K9',
              description: 'Integrated Services Router',
              category: 'Networking',
              specifications: '3xGE WAN, 2xGE LAN',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        },
        {
          id: 'CAT-CISCO-002',
          name: 'Security Equipment',
          description: 'Firewalls and security appliances',
          products: [
            {
              id: 'PROD-CISCO-003',
              name: 'Firepower 2100 Series',
              modelNumber: 'FPR2110',
              description: 'Next-generation firewall',
              category: 'Security Equipment',
              specifications: '10Gbps throughput, NGFW',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        }
      ]
    },
    {
      id: 'MFG-005',
      name: 'Bosch',
      code: 'BOSCH',
      description: 'Security and safety systems',
      country: 'Germany',
      website: 'https://www.boschsecurity.com',
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      categories: [
        {
          id: 'CAT-BOSCH-001',
          name: 'Fire Safety Equipment',
          description: 'Fire detection and alarm systems',
          products: [
            {
              id: 'PROD-BOSCH-001',
              name: 'Fire Alarm Control Panel',
              modelNumber: 'FPA-5000',
              description: 'Modular fire panel',
              category: 'Fire Safety Equipment',
              specifications: 'EN54 certified, modular',
              createdAt: '2024-01-01T00:00:00Z'
            },
            {
              id: 'PROD-BOSCH-002',
              name: 'Smoke Detector',
              modelNumber: 'FAP-440',
              description: 'Optical smoke detector',
              category: 'Fire Safety Equipment',
              specifications: 'Optical, EN54-7',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        },
        {
          id: 'CAT-BOSCH-002',
          name: 'CCTV Systems',
          description: 'IP cameras and recording',
          products: [
            {
              id: 'PROD-BOSCH-003',
              name: 'FLEXIDOME IP 5000',
              modelNumber: 'NDI-5503-A',
              description: '5MP indoor dome camera',
              category: 'CCTV Systems',
              specifications: '5MP, H.265, IVA',
              createdAt: '2024-01-01T00:00:00Z'
            }
          ]
        }
      ]
    }
  ])

  // Computed properties
  const activeManufacturers = computed(() => {
    return manufacturers.value.filter(m => m.status === 'active')
  })

  const manufacturersList = computed(() => {
    return activeManufacturers.value.map(m => ({
      id: m.id,
      name: m.name,
      code: m.code,
      country: m.country
    }))
  })

  // Get manufacturer by ID
  function getManufacturerById(id: string) {
    return manufacturers.value.find(m => m.id === id)
  }

  // Get manufacturer by name
  function getManufacturerByName(name: string) {
    return manufacturers.value.find(m =>
      m.name.toLowerCase() === name.toLowerCase()
    )
  }

  // Get all categories for a manufacturer
  function getManufacturerCategories(manufacturerId: string) {
    const manufacturer = getManufacturerById(manufacturerId)
    return manufacturer?.categories || []
  }

  // Get all products for a manufacturer
  function getManufacturerProducts(manufacturerId: string) {
    const manufacturer = getManufacturerById(manufacturerId)
    if (!manufacturer) return []

    const allProducts: ManufacturerProduct[] = []
    manufacturer.categories.forEach(category => {
      allProducts.push(...category.products)
    })
    return allProducts
  }

  // Get products by category for a manufacturer
  function getProductsByCategory(manufacturerId: string, categoryName: ProductCategory) {
    const manufacturer = getManufacturerById(manufacturerId)
    if (!manufacturer) return []

    const category = manufacturer.categories.find(c => c.name === categoryName)
    return category?.products || []
  }

  // Get product by ID across all manufacturers
  function getProductById(productId: string) {
    for (const manufacturer of manufacturers.value) {
      for (const category of manufacturer.categories) {
        const product = category.products.find(p => p.id === productId)
        if (product) {
          return {
            product,
            manufacturer,
            category
          }
        }
      }
    }
    return null
  }

  // Add new manufacturer
  function addManufacturer(manufacturerData: Omit<Manufacturer, 'id' | 'createdAt'>) {
    const newManufacturer: Manufacturer = {
      ...manufacturerData,
      id: `MFG-${String(manufacturers.value.length + 1).padStart(3, '0')}`,
      createdAt: new Date().toISOString(),
      categories: manufacturerData.categories || []
    }
    manufacturers.value.push(newManufacturer)
    return newManufacturer
  }

  // Update manufacturer
  function updateManufacturer(id: string, updates: Partial<Manufacturer>) {
    const index = manufacturers.value.findIndex(m => m.id === id)
    if (index !== -1) {
      manufacturers.value[index] = {
        ...manufacturers.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
    }
  }

  // Delete manufacturer
  function deleteManufacturer(id: string) {
    const index = manufacturers.value.findIndex(m => m.id === id)
    if (index !== -1) {
      manufacturers.value.splice(index, 1)
    }
  }

  // Add category to manufacturer
  function addCategory(manufacturerId: string, category: ManufacturerCategory) {
    const manufacturer = getManufacturerById(manufacturerId)
    if (manufacturer) {
      manufacturer.categories.push(category)
      manufacturer.updatedAt = new Date().toISOString()
    }
  }

  // Add product to category
  function addProductToCategory(manufacturerId: string, categoryName: ProductCategory, product: ManufacturerProduct) {
    const manufacturer = getManufacturerById(manufacturerId)
    if (manufacturer) {
      const category = manufacturer.categories.find(c => c.name === categoryName)
      if (category) {
        category.products.push(product)
        manufacturer.updatedAt = new Date().toISOString()
      }
    }
  }

  // Remove product from category
  function removeProductFromCategory(manufacturerId: string, categoryName: ProductCategory, productId: string) {
    const manufacturer = getManufacturerById(manufacturerId)
    if (manufacturer) {
      const category = manufacturer.categories.find(c => c.name === categoryName)
      if (category) {
        const index = category.products.findIndex(p => p.id === productId)
        if (index !== -1) {
          category.products.splice(index, 1)
          manufacturer.updatedAt = new Date().toISOString()
        }
      }
    }
  }

  return {
    manufacturers,
    activeManufacturers,
    manufacturersList,
    getManufacturerById,
    getManufacturerByName,
    getManufacturerCategories,
    getManufacturerProducts,
    getProductsByCategory,
    getProductById,
    addManufacturer,
    updateManufacturer,
    deleteManufacturer,
    addCategory,
    addProductToCategory,
    removeProductFromCategory
  }
})
