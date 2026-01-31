import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Quote, QuoteLineItem, QuoteStatus, ApprovalHistoryEntry } from '@/types'

export const useQuotesStore = defineStore('quotes', () => {
  const quotes = ref<Quote[]>([
    {
      id: 'QTE-001',
      quoteNumber: 'Q-2024-001',
      version: 1,
      opportunityId: 'OPP-001',
      companyId: 'CUST-001',
      contactId: 'CNT-001',
      quoteDate: '2024-01-20',
      validUntil: '2024-02-20',
      expiryDate: '2024-02-20',
      currency: 'SAR',
      exchangeRate: 1.0,
      status: 'approved',
      subtotal: 306652,
      discountPercent: 5,
      discountAmount: 15333,
      subtotalAfterDiscount: 291319,
      vatPercent: 15,
      vatAmount: 43698,
      total: 335017,
      totalCost: 227658,
      marginAmount: 63661,
      marginPercent: 21.9,
      paymentTerms: '30% advance, 40% on delivery, 30% after installation',
      deliveryTerms: '45 days from PO',
      requiresApproval: false,
      createdBy: 'USR-001',
      createdAt: '2024-01-20T10:00:00Z',
      approvedBy: 'USR-001',
      approvedAt: '2024-01-20T10:00:00Z',
      customerNotes: 'Complete CCTV system with professional installation and 12-month maintenance',
      approvalHistory: [
        {
          date: '2024-01-20T10:00:00Z',
          status: 'draft',
          userId: 'USR-001',
          userName: 'Ahmed Al-Rashid',
          comments: 'Initial quote created for CCTV installation project'
        },
        {
          date: '2024-01-20T14:00:00Z',
          status: 'approved',
          userId: 'USR-001',
          userName: 'Ahmed Al-Rashid',
          comments: 'Auto-approved - margin above threshold'
        }
      ]
    },
    {
      id: 'QTE-002',
      quoteNumber: 'Q-2024-002',
      version: 1,
      opportunityId: 'OPP-002',
      companyId: 'CUST-002',
      contactId: 'CNT-003',
      quoteDate: '2024-01-22',
      validUntil: '2024-02-22',
      expiryDate: '2024-02-22',
      currency: 'SAR',
      exchangeRate: 1.0,
      status: 'sent',
      subtotal: 187841,
      discountPercent: 3,
      discountAmount: 5635,
      subtotalAfterDiscount: 182206,
      vatPercent: 15,
      vatAmount: 27331,
      total: 209537,
      totalCost: 136244,
      marginAmount: 45962,
      marginPercent: 25.2,
      paymentTerms: '50% advance, 50% on completion',
      deliveryTerms: '30 days from PO',
      requiresApproval: false,
      createdBy: 'USR-002',
      createdAt: '2024-01-22T11:00:00Z',
      approvedBy: 'USR-002',
      approvedAt: '2024-01-22T11:00:00Z',
      sentAt: '2024-01-23T15:00:00Z',
      customerNotes: 'Complete access control solution with 1-year maintenance',
      approvalHistory: [
        {
          date: '2024-01-22T11:00:00Z',
          status: 'draft',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Quote created for access control system'
        },
        {
          date: '2024-01-22T16:00:00Z',
          status: 'approved',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Auto-approved - margin above threshold'
        },
        {
          date: '2024-01-23T15:00:00Z',
          status: 'sent',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Quote sent to customer via email'
        }
      ]
    },
    {
      id: 'QTE-003',
      quoteNumber: 'Q-2024-003',
      version: 2,
      opportunityId: 'OPP-003',
      companyId: 'CUST-003',
      contactId: 'CNT-004',
      quoteDate: '2024-01-18',
      validUntil: '2024-03-18',
      expiryDate: '2024-03-18',
      currency: 'SAR',
      exchangeRate: 1.0,
      status: 'pending_approval',
      subtotal: 701018,
      discountPercent: 7,
      discountAmount: 49071,
      subtotalAfterDiscount: 651947,
      vatPercent: 15,
      vatAmount: 97792,
      total: 749739,
      totalCost: 523460,
      marginAmount: 128487,
      marginPercent: 19.7,
      paymentTerms: '20% advance, 30% on delivery, 30% on installation, 20% after acceptance',
      deliveryTerms: '60 days from PO',
      requiresApproval: true,
      approvalReason: 'Margin below 20% and value exceeds 500,000 SAR',
      createdBy: 'USR-001',
      createdAt: '2024-01-18T09:00:00Z',
      updatedAt: '2024-01-19T16:00:00Z',
      internalNotes: 'Client requested 7% discount due to large order volume',
      customerNotes: 'Comprehensive integrated security solution for entire facility',
      approvalHistory: [
        {
          date: '2024-01-18T09:00:00Z',
          status: 'draft',
          userId: 'USR-001',
          userName: 'Ahmed Al-Rashid',
          comments: 'Initial quote created for comprehensive security project'
        },
        {
          date: '2024-01-19T16:00:00Z',
          status: 'pending_approval',
          userId: 'USR-001',
          userName: 'Ahmed Al-Rashid',
          comments: 'Revised with 7% discount. Awaiting manager approval for large deal'
        }
      ]
    },
    {
      id: 'QTE-004',
      quoteNumber: 'Q-2024-004',
      version: 1,
      opportunityId: 'OPP-004',
      companyId: 'CUST-001',
      contactId: 'CNT-001',
      quoteDate: '2024-01-05',
      validUntil: '2024-01-20',
      expiryDate: '2024-01-20',
      currency: 'SAR',
      exchangeRate: 1.0,
      status: 'accepted',
      subtotal: 224304,
      discountPercent: 0,
      discountAmount: 0,
      subtotalAfterDiscount: 224304,
      vatPercent: 15,
      vatAmount: 33646,
      total: 257950,
      totalCost: 142800,
      marginAmount: 81504,
      marginPercent: 36.3,
      paymentTerms: 'Monthly billing',
      deliveryTerms: 'Continuous service',
      requiresApproval: false,
      createdBy: 'USR-002',
      createdAt: '2024-01-05T10:00:00Z',
      approvedBy: 'USR-002',
      approvedAt: '2024-01-05T10:00:00Z',
      sentAt: '2024-01-06T09:00:00Z',
      acceptedAt: '2024-01-12T14:00:00Z',
      customerNotes: 'Annual comprehensive security services package',
      approvalHistory: [
        {
          date: '2024-01-05T10:00:00Z',
          status: 'draft',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Created annual recurring services contract quote'
        },
        {
          date: '2024-01-05T15:00:00Z',
          status: 'approved',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Auto-approved - margin above threshold'
        },
        {
          date: '2024-01-06T09:00:00Z',
          status: 'sent',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Sent to customer'
        },
        {
          date: '2024-01-12T14:00:00Z',
          status: 'accepted',
          userId: 'USR-002',
          userName: 'Fatima Al-Zahrani',
          comments: 'Customer accepted the quote'
        }
      ]
    }
  ])

  const lineItems = ref<QuoteLineItem[]>([
    // Quote QTE-001 line items - CCTV Installation Project
    {
      id: 'QLI-001',
      quoteId: 'QTE-001',
      lineNumber: 1,
      type: 'product',
      category: 'materials',
      productId: 'PRD-001',
      productSku: 'CAM-HK-DS2CD2385G1',
      productName: 'Hikvision 8MP IP Dome Camera',
      description: '8MP (4K) outdoor IP dome camera with IR, WDR, H.265+ compression',
      quantity: 50,
      unitCost: 756,
      unitPrice: 1008,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 50400,
      totalCost: 37800,
      marginAmount: 12600,
      marginPercent: 25.0
    },
    {
      id: 'QLI-002',
      quoteId: 'QTE-001',
      lineNumber: 2,
      type: 'product',
      category: 'materials',
      productId: 'PRD-002',
      productSku: 'CAM-DH-IPC8542E',
      productName: 'Dahua 5MP Bullet Camera',
      description: '5MP bullet camera with motorized lens and smart detection',
      quantity: 30,
      unitCost: 924,
      unitPrice: 1232,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 36960,
      totalCost: 27720,
      marginAmount: 9240,
      marginPercent: 25.0
    },
    {
      id: 'QLI-003',
      quoteId: 'QTE-001',
      lineNumber: 3,
      type: 'product',
      category: 'materials',
      productId: 'PRD-003',
      productSku: 'NVR-HK-DS96256NI',
      productName: 'Hikvision 256-Channel NVR',
      description: 'Network Video Recorder with 256 channels, 12MP recording resolution',
      quantity: 2,
      unitCost: 15094,
      unitPrice: 19351,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 38702,
      totalCost: 30188,
      marginAmount: 8514,
      marginPercent: 22.0
    },
    {
      id: 'QLI-004',
      quoteId: 'QTE-001',
      lineNumber: 4,
      type: 'product',
      category: 'materials',
      productId: 'PRD-014',
      productSku: 'NET-CISCO-SG350',
      productName: 'Cisco SG350-28P PoE Switch',
      description: '28-port Gigabit PoE+ managed switch, 195W power budget',
      quantity: 4,
      unitCost: 2730,
      unitPrice: 3592,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 14368,
      totalCost: 10920,
      marginAmount: 3448,
      marginPercent: 24.0
    },
    {
      id: 'QLI-005',
      quoteId: 'QTE-001',
      lineNumber: 5,
      type: 'product',
      category: 'materials',
      productId: 'PRD-015',
      productSku: 'STG-QNAP-TS1685',
      productName: 'QNAP TS-1685 NAS Storage',
      description: '16-bay NAS with Xeon processor, 10GbE, 96TB capacity',
      quantity: 1,
      unitCost: 13680,
      unitPrice: 17316,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 17316,
      totalCost: 13680,
      marginAmount: 3636,
      marginPercent: 21.0
    },
    {
      id: 'QLI-006',
      quoteId: 'QTE-001',
      lineNumber: 6,
      type: 'product',
      category: 'materials',
      productId: 'PRD-011',
      productSku: 'SW-MS-VMS50',
      productName: 'Milestone XProtect Corporate - 50 Camera License',
      description: 'Video management software license for 50 cameras, perpetual',
      quantity: 2,
      unitCost: 16875,
      unitPrice: 21635,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 43270,
      totalCost: 33750,
      marginAmount: 9520,
      marginPercent: 22.0
    },
    {
      id: 'QLI-007',
      quoteId: 'QTE-001',
      lineNumber: 7,
      type: 'labor',
      category: 'manpower',
      productId: 'PRD-012',
      productSku: 'SRV-INSTALL-CAM',
      productName: 'Camera Installation Service',
      description: 'Professional installation service per camera including mounting, cabling, configuration',
      laborPosition: 'Installation Technician',
      quantity: 80,
      unitCost: 200,
      unitPrice: 308,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 24640,
      totalCost: 16000,
      marginAmount: 8640,
      marginPercent: 35.0
    },
    {
      id: 'QLI-008',
      quoteId: 'QTE-001',
      lineNumber: 8,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-019',
      productSku: 'SRV-MAINT-CCTV',
      productName: 'CCTV System Maintenance',
      description: 'Monthly CCTV system maintenance service (per 10 cameras)',
      months: 12,
      quantity: 12,
      unitCost: 800,
      unitPrice: 1333,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 15996,
      totalCost: 9600,
      marginAmount: 6396,
      marginPercent: 40.0
    },
    {
      id: 'QLI-009',
      quoteId: 'QTE-001',
      lineNumber: 9,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Cabling & Accessories',
      description: 'Network cables, conduits, brackets, and mounting hardware',
      quantity: 1,
      unitCost: 15000,
      unitPrice: 21000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 21000,
      totalCost: 15000,
      marginAmount: 6000,
      marginPercent: 28.6
    },
    {
      id: 'QLI-010',
      quoteId: 'QTE-001',
      lineNumber: 10,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Project Management & Commissioning',
      description: 'Project management, testing, and system commissioning',
      quantity: 1,
      unitCost: 18000,
      unitPrice: 25000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 25000,
      totalCost: 18000,
      marginAmount: 7000,
      marginPercent: 28.0
    },
    // Quote QTE-002 line items - Access Control System
    {
      id: 'QLI-011',
      quoteId: 'QTE-002',
      lineNumber: 1,
      type: 'product',
      category: 'materials',
      productId: 'PRD-004',
      productSku: 'ACS-ZK-C3400',
      productName: 'ZKTeco C3-400 Access Controller',
      description: '4-door access controller with TCP/IP, supports 100K users',
      quantity: 15,
      unitCost: 1907,
      unitPrice: 2477,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 37155,
      totalCost: 28605,
      marginAmount: 8550,
      marginPercent: 23.0
    },
    {
      id: 'QLI-012',
      quoteId: 'QTE-002',
      lineNumber: 2,
      type: 'product',
      category: 'materials',
      productId: 'PRD-005',
      productSku: 'ACS-HID-5365',
      productName: 'HID Proximity Card Reader',
      description: 'ProxPoint Plus wall-mount card reader, Wiegand output',
      quantity: 60,
      unitCost: 357,
      unitPrice: 496,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 29760,
      totalCost: 21420,
      marginAmount: 8340,
      marginPercent: 28.0
    },
    {
      id: 'QLI-013',
      quoteId: 'QTE-002',
      lineNumber: 3,
      type: 'product',
      category: 'materials',
      productId: 'PRD-013',
      productSku: 'SVR-DELL-R750',
      productName: 'Dell PowerEdge R750 Server',
      description: 'Rack server: 2x Xeon Gold, 128GB RAM, 4x 2TB SSD RAID10',
      quantity: 1,
      unitCost: 36019,
      unitPrice: 43926,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 43926,
      totalCost: 36019,
      marginAmount: 7907,
      marginPercent: 18.0
    },
    {
      id: 'QLI-014',
      quoteId: 'QTE-002',
      lineNumber: 4,
      type: 'labor',
      category: 'manpower',
      productName: 'Access Control Installation',
      description: 'Installation, cabling, and configuration of access control system',
      laborPosition: 'System Integrator',
      quantity: 120,
      unitCost: 250,
      unitPrice: 375,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 45000,
      totalCost: 30000,
      marginAmount: 15000,
      marginPercent: 33.3
    },
    {
      id: 'QLI-015',
      quoteId: 'QTE-002',
      lineNumber: 5,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-020',
      productSku: 'SRV-MAINT-ACCESS',
      productName: 'Access Control System Maintenance',
      description: 'Monthly access control system maintenance (per 5 doors)',
      months: 12,
      quantity: 12,
      unitCost: 600,
      unitPrice: 1000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 12000,
      totalCost: 7200,
      marginAmount: 4800,
      marginPercent: 40.0
    },
    {
      id: 'QLI-016',
      quoteId: 'QTE-002',
      lineNumber: 6,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Access Cards & Accessories',
      description: 'Proximity cards, door strikes, power supplies, cables',
      quantity: 1,
      unitCost: 8000,
      unitPrice: 12000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 12000,
      totalCost: 8000,
      marginAmount: 4000,
      marginPercent: 33.3
    },
    {
      id: 'QLI-017',
      quoteId: 'QTE-002',
      lineNumber: 7,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Training & Documentation',
      description: 'Staff training and system documentation',
      quantity: 1,
      unitCost: 5000,
      unitPrice: 8000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 8000,
      totalCost: 5000,
      marginAmount: 3000,
      marginPercent: 37.5
    },
    // Quote QTE-003 line items - Comprehensive Security Solution
    {
      id: 'QLI-018',
      quoteId: 'QTE-003',
      lineNumber: 1,
      type: 'product',
      category: 'materials',
      productId: 'PRD-001',
      productSku: 'CAM-HK-DS2CD2385G1',
      productName: 'Hikvision 8MP IP Dome Camera',
      description: '8MP (4K) outdoor IP dome camera with IR, WDR, H.265+ compression',
      quantity: 100,
      unitCost: 756,
      unitPrice: 1008,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 100800,
      totalCost: 75600,
      marginAmount: 25200,
      marginPercent: 25.0
    },
    {
      id: 'QLI-019',
      quoteId: 'QTE-003',
      lineNumber: 2,
      type: 'product',
      category: 'materials',
      productId: 'PRD-003',
      productSku: 'NVR-HK-DS96256NI',
      productName: 'Hikvision 256-Channel NVR',
      description: 'Network Video Recorder with 256 channels, 12MP recording resolution',
      quantity: 3,
      unitCost: 15094,
      unitPrice: 19351,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 58053,
      totalCost: 45282,
      marginAmount: 12771,
      marginPercent: 22.0
    },
    {
      id: 'QLI-020',
      quoteId: 'QTE-003',
      lineNumber: 3,
      type: 'product',
      category: 'materials',
      productId: 'PRD-004',
      productSku: 'ACS-ZK-C3400',
      productName: 'ZKTeco C3-400 Access Controller',
      description: '4-door access controller with TCP/IP, supports 100K users',
      quantity: 25,
      unitCost: 1907,
      unitPrice: 2477,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 61925,
      totalCost: 47675,
      marginAmount: 14250,
      marginPercent: 23.0
    },
    {
      id: 'QLI-021',
      quoteId: 'QTE-003',
      lineNumber: 4,
      type: 'product',
      category: 'materials',
      productId: 'PRD-006',
      productSku: 'ALM-BOSCH-B9512G',
      productName: 'Bosch B9512G Control Panel',
      description: 'Professional intrusion alarm control panel, 512 zones',
      quantity: 5,
      unitCost: 4160,
      unitPrice: 5474,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 27370,
      totalCost: 20800,
      marginAmount: 6570,
      marginPercent: 24.0
    },
    {
      id: 'QLI-022',
      quoteId: 'QTE-003',
      lineNumber: 5,
      type: 'product',
      category: 'materials',
      productId: 'PRD-007',
      productSku: 'FIRE-NOTIFIER-NFS320',
      productName: 'Notifier NFS2-3030 Fire Alarm Panel',
      description: 'Intelligent fire alarm control panel, 636 addressable points',
      quantity: 3,
      unitCost: 12075,
      unitPrice: 15094,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 45282,
      totalCost: 36225,
      marginAmount: 9057,
      marginPercent: 20.0
    },
    {
      id: 'QLI-023',
      quoteId: 'QTE-003',
      lineNumber: 6,
      type: 'product',
      category: 'materials',
      productId: 'PRD-013',
      productSku: 'SVR-DELL-R750',
      productName: 'Dell PowerEdge R750 Server',
      description: 'Rack server: 2x Xeon Gold, 128GB RAM, 4x 2TB SSD RAID10',
      quantity: 2,
      unitCost: 36019,
      unitPrice: 43926,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 87852,
      totalCost: 72038,
      marginAmount: 15814,
      marginPercent: 18.0
    },
    {
      id: 'QLI-024',
      quoteId: 'QTE-003',
      lineNumber: 7,
      type: 'product',
      category: 'materials',
      productId: 'PRD-014',
      productSku: 'NET-CISCO-SG350',
      productName: 'Cisco SG350-28P PoE Switch',
      description: '28-port Gigabit PoE+ managed switch, 195W power budget',
      quantity: 8,
      unitCost: 2730,
      unitPrice: 3592,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 28736,
      totalCost: 21840,
      marginAmount: 6896,
      marginPercent: 24.0
    },
    {
      id: 'QLI-025',
      quoteId: 'QTE-003',
      lineNumber: 8,
      type: 'labor',
      category: 'manpower',
      productName: 'System Integration & Installation',
      description: 'Complete system installation, integration, and commissioning',
      laborPosition: 'Senior System Integrator',
      quantity: 480,
      unitCost: 300,
      unitPrice: 450,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 216000,
      totalCost: 144000,
      marginAmount: 72000,
      marginPercent: 33.3
    },
    {
      id: 'QLI-026',
      quoteId: 'QTE-003',
      lineNumber: 9,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Materials & Accessories Package',
      description: 'Comprehensive cabling, hardware, and accessory package',
      quantity: 1,
      unitCost: 35000,
      unitPrice: 50000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 50000,
      totalCost: 35000,
      marginAmount: 15000,
      marginPercent: 30.0
    },
    {
      id: 'QLI-027',
      quoteId: 'QTE-003',
      lineNumber: 10,
      type: 'product',
      category: 'miscellaneous',
      productName: 'Project Management & Support',
      description: 'Full project management, testing, and warranty support',
      quantity: 1,
      unitCost: 25000,
      unitPrice: 35000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 35000,
      totalCost: 25000,
      marginAmount: 10000,
      marginPercent: 28.6
    },
    // Quote QTE-004 line items - Recurring Services Contract
    {
      id: 'QLI-028',
      quoteId: 'QTE-004',
      lineNumber: 1,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-017',
      productSku: 'SRV-GUARD-BASIC',
      productName: 'Security Guard - Basic Level',
      description: 'Basic level security guard service (per guard per month)',
      months: 12,
      quantity: 12,
      unitCost: 4500,
      unitPrice: 6429,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 77148,
      totalCost: 54000,
      marginAmount: 23148,
      marginPercent: 30.0
    },
    {
      id: 'QLI-029',
      quoteId: 'QTE-004',
      lineNumber: 2,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-019',
      productSku: 'SRV-MAINT-CCTV',
      productName: 'CCTV System Maintenance',
      description: 'Monthly CCTV system maintenance service (per 10 cameras)',
      months: 12,
      quantity: 12,
      unitCost: 800,
      unitPrice: 1333,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 15996,
      totalCost: 9600,
      marginAmount: 6396,
      marginPercent: 40.0
    },
    {
      id: 'QLI-030',
      quoteId: 'QTE-004',
      lineNumber: 3,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-020',
      productSku: 'SRV-MAINT-ACCESS',
      productName: 'Access Control System Maintenance',
      description: 'Monthly access control system maintenance (per 5 doors)',
      months: 12,
      quantity: 12,
      unitCost: 600,
      unitPrice: 1000,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 12000,
      totalCost: 7200,
      marginAmount: 4800,
      marginPercent: 40.0
    },
    {
      id: 'QLI-031',
      quoteId: 'QTE-004',
      lineNumber: 4,
      type: 'service',
      category: 'manpower',
      productId: 'PRD-021',
      productSku: 'SRV-PATROL-SERVICE',
      productName: 'Mobile Patrol Service',
      description: 'Mobile patrol service (per location per month, 2 visits daily)',
      months: 12,
      quantity: 12,
      unitCost: 3500,
      unitPrice: 5385,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 64620,
      totalCost: 42000,
      marginAmount: 22620,
      marginPercent: 35.0
    },
    {
      id: 'QLI-032',
      quoteId: 'QTE-004',
      lineNumber: 5,
      type: 'service',
      category: 'miscellaneous',
      productId: 'PRD-022',
      productSku: 'SRV-MONITORING-24-7',
      productName: '24/7 Monitoring Center Service',
      description: 'Remote 24/7 monitoring service (per site per month)',
      months: 12,
      quantity: 12,
      unitCost: 2500,
      unitPrice: 4545,
      discountPercent: 0,
      discountAmount: 0,
      lineTotal: 54540,
      totalCost: 30000,
      marginAmount: 24540,
      marginPercent: 45.0
    }
  ])

  // Computed
  const pendingApprovalQuotes = computed(() => {
    return quotes.value.filter(q => q.status === 'pending_approval')
  })

  const draftQuotes = computed(() => {
    return quotes.value.filter(q => q.status === 'draft')
  })

  const sentQuotes = computed(() => {
    return quotes.value.filter(q => q.status === 'sent')
  })

  const acceptedQuotes = computed(() => {
    return quotes.value.filter(q => q.status === 'accepted')
  })

  const totalQuoteValue = computed(() => {
    return quotes.value.reduce((sum, q) => sum + q.total, 0)
  })

  const averageMargin = computed(() => {
    if (quotes.value.length === 0) return 0
    const totalMargin = quotes.value.reduce((sum, q) => sum + q.marginPercent, 0)
    return totalMargin / quotes.value.length
  })

  // Actions
  function addQuote(quote: Quote) {
    quotes.value.push(quote)
  }

  function updateQuote(id: string, updates: Partial<Quote>) {
    const index = quotes.value.findIndex(q => q.id === id)
    if (index !== -1) {
      quotes.value[index] = {
        ...quotes.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
    }
  }

  function deleteQuote(id: string) {
    const index = quotes.value.findIndex(q => q.id === id)
    if (index !== -1) {
      quotes.value.splice(index, 1)
      // Also delete associated line items
      lineItems.value = lineItems.value.filter(li => li.quoteId !== id)
    }
  }

  function getQuoteById(id: string) {
    return quotes.value.find(q => q.id === id)
  }

  function getLineItemsByQuoteId(quoteId: string) {
    return lineItems.value.filter(li => li.quoteId === quoteId)
  }

  function addLineItem(lineItem: QuoteLineItem) {
    lineItems.value.push(lineItem)
  }

  function updateLineItem(id: string, updates: Partial<QuoteLineItem>) {
    const index = lineItems.value.findIndex(li => li.id === id)
    if (index !== -1) {
      lineItems.value[index] = { ...lineItems.value[index], ...updates }
    }
  }

  function deleteLineItem(id: string) {
    const index = lineItems.value.findIndex(li => li.id === id)
    if (index !== -1) {
      lineItems.value.splice(index, 1)
    }
  }

  function updateQuoteStatus(id: string, status: QuoteStatus) {
    updateQuote(id, { status })
  }

  function sendQuote(id: string) {
    updateQuote(id, {
      status: 'sent',
      sentAt: new Date().toISOString()
    })
  }

  function acceptQuote(id: string) {
    updateQuote(id, {
      status: 'accepted',
      acceptedAt: new Date().toISOString()
    })
  }

  function declineQuote(id: string, reason: string) {
    updateQuote(id, {
      status: 'declined',
      declinedAt: new Date().toISOString(),
      declineReason: reason
    })
  }

  function approveQuote(id: string, approverId: string) {
    updateQuote(id, {
      status: 'approved',
      approvedBy: approverId,
      approvedAt: new Date().toISOString()
    })
  }

  function rejectQuote(id: string, reason: string) {
    updateQuote(id, {
      status: 'draft',
      rejectionReason: reason
    })
  }

  function cloneQuote(id: string): Quote | null {
    const original = getQuoteById(id)
    if (!original) return null

    const newId = `QTE-${Date.now()}`
    const newQuote: Quote = {
      ...original,
      id: newId,
      quoteNumber: `Q-${Date.now()}`,
      version: 1,
      status: 'draft',
      quoteDate: new Date().toISOString().split('T')[0],
      createdAt: new Date().toISOString(),
      approvedBy: undefined,
      approvedAt: undefined,
      sentAt: undefined,
      acceptedAt: undefined,
      declinedAt: undefined
    }

    addQuote(newQuote)

    // Clone line items
    const originalLineItems = getLineItemsByQuoteId(id)
    originalLineItems.forEach(li => {
      const newLineItem: QuoteLineItem = {
        ...li,
        id: `QLI-${Date.now()}-${li.lineNumber}`,
        quoteId: newId
      }
      addLineItem(newLineItem)
    })

    return newQuote
  }

  function getQuotesByCompany(companyId: string) {
    return quotes.value.filter(q => q.companyId === companyId)
  }

  function getQuotesByOpportunity(opportunityId: string) {
    return quotes.value.filter(q => q.opportunityId === opportunityId)
  }

  function recalculateQuoteTotals(quoteId: string) {
    const quote = getQuoteById(quoteId)
    if (!quote) return

    const items = getLineItemsByQuoteId(quoteId)

    const subtotal = items.reduce((sum, item) => sum + item.lineTotal, 0)
    const discountAmount = (subtotal * quote.discountPercent) / 100
    const subtotalAfterDiscount = subtotal - discountAmount
    const vatAmount = (subtotalAfterDiscount * quote.vatPercent) / 100
    const total = subtotalAfterDiscount + vatAmount

    const totalCost = items.reduce((sum, item) => sum + item.totalCost, 0)
    const marginAmount = subtotalAfterDiscount - totalCost
    const marginPercent = subtotalAfterDiscount > 0 ? (marginAmount / subtotalAfterDiscount) * 100 : 0

    updateQuote(quoteId, {
      subtotal,
      discountAmount,
      subtotalAfterDiscount,
      vatAmount,
      total,
      totalCost,
      marginAmount,
      marginPercent
    })
  }

  return {
    quotes,
    lineItems,
    pendingApprovalQuotes,
    draftQuotes,
    sentQuotes,
    acceptedQuotes,
    totalQuoteValue,
    averageMargin,
    addQuote,
    updateQuote,
    deleteQuote,
    getQuoteById,
    getLineItemsByQuoteId,
    addLineItem,
    updateLineItem,
    deleteLineItem,
    updateQuoteStatus,
    sendQuote,
    acceptQuote,
    declineQuote,
    approveQuote,
    rejectQuote,
    cloneQuote,
    getQuotesByCompany,
    getQuotesByOpportunity,
    recalculateQuoteTotals
  }
})
