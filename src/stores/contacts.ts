import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Contact } from '@/types'

export const useContactsStore = defineStore('contacts', () => {
  const contacts = ref<Contact[]>([
    {
      id: 'CNT-001',
      companyId: 'CUST-001',
      firstName: 'Ahmed',
      lastName: 'Al-Salem',
      fullName: 'Ahmed Al-Salem',
      title: 'IT Director',
      department: 'Information Technology',
      email: 'ahmed@abc-corp.com',
      phone: '+966 11 234 5678',
      mobile: '+966 50 123 4567',
      isPrimary: true,
      status: 'active',
      createdAt: '2023-06-15T10:00:00Z',
      createdBy: 'admin',
      notes: 'Key decision maker for IT projects'
    },
    {
      id: 'CNT-002',
      companyId: 'CUST-001',
      firstName: 'Sarah',
      lastName: 'Al-Mansour',
      fullName: 'Sarah Al-Mansour',
      title: 'Procurement Manager',
      department: 'Procurement',
      email: 'sarah@abc-corp.com',
      phone: '+966 11 234 5679',
      mobile: '+966 50 123 4568',
      isPrimary: false,
      status: 'active',
      createdAt: '2023-06-20T11:30:00Z',
      createdBy: 'admin',
      notes: 'Handles purchasing and vendor management'
    },
    {
      id: 'CNT-003',
      companyId: 'CUST-002',
      firstName: 'Fatima',
      lastName: 'Al-Otaibi',
      fullName: 'Fatima Al-Otaibi',
      title: 'Operations Director',
      department: 'Operations',
      email: 'fatima@xyz-ind.com',
      phone: '+966 12 345 6789',
      mobile: '+966 50 234 5678',
      isPrimary: true,
      status: 'active',
      createdAt: '2023-08-22T09:00:00Z',
      createdBy: 'admin',
      notes: 'Primary contact for all operations-related projects'
    },
    {
      id: 'CNT-004',
      companyId: 'CUST-003',
      firstName: 'Mohammed',
      lastName: 'Al-Qahtani',
      fullName: 'Mohammed Al-Qahtani',
      title: 'IT Manager',
      department: 'Information Technology',
      email: 'mohammed@global-sol.com',
      phone: '+966 13 456 7890',
      mobile: '+966 50 345 6789',
      isPrimary: true,
      status: 'active',
      createdAt: '2023-09-10T14:00:00Z',
      createdBy: 'admin',
      notes: 'Technical decision maker'
    },
    {
      id: 'CNT-005',
      companyId: 'CUST-003',
      firstName: 'Nora',
      lastName: 'Al-Mutairi',
      fullName: 'Nora Al-Mutairi',
      title: 'Finance Manager',
      department: 'Finance',
      email: 'nora@global-sol.com',
      phone: '+966 13 456 7891',
      mobile: '+966 50 345 6790',
      isPrimary: false,
      status: 'active',
      createdAt: '2023-09-15T10:30:00Z',
      createdBy: 'admin',
      notes: 'Approves all financial transactions'
    }
  ])

  // Computed
  const activeContacts = computed(() => {
    return contacts.value.filter(c => c.status === 'active')
  })

  const primaryContacts = computed(() => {
    return contacts.value.filter(c => c.isPrimary)
  })

  // Actions
  function addContact(contact: Contact) {
    contacts.value.push(contact)
  }

  function updateContact(id: string, updates: Partial<Contact>) {
    const index = contacts.value.findIndex(c => c.id === id)
    if (index !== -1) {
      contacts.value[index] = { ...contacts.value[index], ...updates }
    }
  }

  function deleteContact(id: string) {
    const index = contacts.value.findIndex(c => c.id === id)
    if (index !== -1) {
      contacts.value.splice(index, 1)
    }
  }

  function getContactById(id: string) {
    return contacts.value.find(c => c.id === id)
  }

  function getContactsByCompany(companyId: string) {
    return contacts.value.filter(c => c.companyId === companyId)
  }

  function getPrimaryContactByCompany(companyId: string) {
    return contacts.value.find(c => c.companyId === companyId && c.isPrimary)
  }

  function setPrimaryContact(companyId: string, contactId: string) {
    // Remove primary flag from all contacts of this company
    contacts.value.forEach(contact => {
      if (contact.companyId === companyId) {
        contact.isPrimary = false
      }
    })

    // Set new primary contact
    const contact = contacts.value.find(c => c.id === contactId)
    if (contact) {
      contact.isPrimary = true
    }
  }

  return {
    contacts,
    activeContacts,
    primaryContacts,
    addContact,
    updateContact,
    deleteContact,
    getContactById,
    getContactsByCompany,
    getPrimaryContactByCompany,
    setPrimaryContact
  }
})
