import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Contract, ContractStatus, ContractType, Currency } from '@/types'

export const useContractsStore = defineStore('contracts', () => {
  const contracts = ref<Contract[]>([
    {
      id: 'CNT-001',
      contractNumber: 'CONTRACT-2024-001',
      name: 'Annual Maintenance Contract - ABC Corp',
      type: 'maintenance',
      status: 'active',
      quoteId: 'QTE-003',
      opportunityId: 'OPP-004',
      customerId: 'CUST-001',
      contactId: 'CNT-001',
      priceBookId: 'PB-004',
      startDate: '2024-01-15',
      endDate: '2025-01-14',
      signedDate: '2024-01-12',
      totalValue: 120000,
      currency: 'SAR',
      paymentTerms: 'Quarterly in advance',
      autoRenewal: true,
      renewalNoticeDays: 30,
      documentIds: [],
      terms: 'Standard maintenance terms and conditions apply',
      deliverables: 'Quarterly maintenance visits, 24/7 support hotline, preventive maintenance',
      sla: 'Response time: 4 hours, Resolution time: 24 hours',
      createdAt: '2024-01-12T00:00:00Z',
      createdBy: 'USR-002',
      approvedBy: 'USR-004',
      approvedAt: '2024-01-12T10:00:00Z',
      signedBy: 'Ahmed Al-Rashid (ABC Corp CEO)'
    }
  ])

  // Computed
  const activeContracts = computed(() =>
    contracts.value.filter(c => c.status === 'active')
  )

  const expiringContracts = computed(() => {
    const today = new Date()
    const thirtyDaysFromNow = new Date()
    thirtyDaysFromNow.setDate(today.getDate() + 30)

    return contracts.value.filter(c => {
      if (c.status !== 'active') return false
      const endDate = new Date(c.endDate)
      return endDate >= today && endDate <= thirtyDaysFromNow
    })
  })

  const contractsByType = computed(() => {
    const grouped: Record<ContractType, Contract[]> = {
      sales: [],
      maintenance: [],
      service: [],
      project: [],
      subscription: []
    }

    contracts.value.forEach(contract => {
      grouped[contract.type].push(contract)
    })

    return grouped
  })

  const contractStats = computed(() => {
    const total = contracts.value.length
    const active = contracts.value.filter(c => c.status === 'active').length
    const expiring = expiringContracts.value.length
    const totalValue = activeContracts.value.reduce((sum, c) => sum + c.totalValue, 0)

    return {
      total,
      active,
      expiring,
      totalValue
    }
  })

  // Actions
  function getContractById(id: string): Contract | undefined {
    return contracts.value.find(c => c.id === id)
  }

  function getContractByNumber(contractNumber: string): Contract | undefined {
    return contracts.value.find(c => c.contractNumber === contractNumber)
  }

  function getContractsByCustomer(customerId: string): Contract[] {
    return contracts.value.filter(c => c.customerId === customerId)
  }

  function getContractByQuote(quoteId: string): Contract | undefined {
    return contracts.value.find(c => c.quoteId === quoteId)
  }

  function addContract(contract: Omit<Contract, 'id' | 'contractNumber' | 'createdAt'>) {
    const contractCount = contracts.value.length + 1
    const year = new Date().getFullYear()
    const id = `CNT-${String(contractCount).padStart(3, '0')}`
    const contractNumber = `CONTRACT-${year}-${String(contractCount).padStart(3, '0')}`

    const newContract: Contract = {
      ...contract,
      id,
      contractNumber,
      createdAt: new Date().toISOString()
    }

    contracts.value.push(newContract)
    return newContract
  }

  function updateContract(id: string, updates: Partial<Contract>) {
    const index = contracts.value.findIndex(c => c.id === id)
    if (index !== -1) {
      contracts.value[index] = {
        ...contracts.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return contracts.value[index]
    }
    return null
  }

  function deleteContract(id: string) {
    const index = contracts.value.findIndex(c => c.id === id)
    if (index !== -1) {
      contracts.value.splice(index, 1)
      return true
    }
    return false
  }

  function activateContract(id: string, signedBy: string) {
    const contract = getContractById(id)
    if (contract) {
      contract.status = 'active'
      contract.signedDate = new Date().toISOString().split('T')[0]
      contract.signedBy = signedBy
      contract.updatedAt = new Date().toISOString()
      return contract
    }
    return null
  }

  function terminateContract(id: string, reason?: string) {
    const contract = getContractById(id)
    if (contract) {
      contract.status = 'terminated'
      contract.updatedAt = new Date().toISOString()
      if (reason) {
        contract.notes = `Terminated: ${reason}. ${contract.notes || ''}`
      }
      return contract
    }
    return null
  }

  function renewContract(id: string, newEndDate: string) {
    const oldContract = getContractById(id)
    if (!oldContract) return null

    // Create new contract from old one
    const newContract = addContract({
      ...oldContract,
      startDate: oldContract.endDate,
      endDate: newEndDate,
      signedDate: undefined,
      status: 'draft',
      renewedFromId: oldContract.id,
      documentIds: [],
      approvedBy: undefined,
      approvedAt: undefined,
      signedBy: undefined,
      createdBy: 'USR-001'  // Current user
    })

    // Link old contract to new one
    oldContract.renewedToId = newContract.id
    oldContract.status = 'renewed'
    oldContract.updatedAt = new Date().toISOString()

    return newContract
  }

  function addDocumentToContract(contractId: string, documentId: string) {
    const contract = getContractById(contractId)
    if (contract) {
      if (!contract.documentIds.includes(documentId)) {
        contract.documentIds.push(documentId)
        contract.updatedAt = new Date().toISOString()
      }
      return contract
    }
    return null
  }

  function removeDocumentFromContract(contractId: string, documentId: string) {
    const contract = getContractById(contractId)
    if (contract) {
      const index = contract.documentIds.indexOf(documentId)
      if (index !== -1) {
        contract.documentIds.splice(index, 1)
        contract.updatedAt = new Date().toISOString()
      }
      return contract
    }
    return null
  }

  return {
    contracts,
    activeContracts,
    expiringContracts,
    contractsByType,
    contractStats,
    getContractById,
    getContractByNumber,
    getContractsByCustomer,
    getContractByQuote,
    addContract,
    updateContract,
    deleteContract,
    activateContract,
    terminateContract,
    renewContract,
    addDocumentToContract,
    removeDocumentFromContract
  }
})
