import { ref } from 'vue'
import { defineStore } from 'pinia'
import type { LaborPosition, Status } from '@/types'

export const useLaborPositionsStore = defineStore('laborPositions', () => {
  const positions = ref<LaborPosition[]>([
    {
      id: 'LP-001',
      name: 'Security Guard - Grade A',
      description: 'Senior security guard with 5+ years experience, licensed and trained',
      department: 'Security Operations',
      costPerHour: 45,
      sellingRatePerHour: 65,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Includes uniform, equipment, and supervision'
    },
    {
      id: 'LP-002',
      name: 'Security Guard - Grade B',
      description: 'Standard security guard with 2-5 years experience',
      department: 'Security Operations',
      costPerHour: 35,
      sellingRatePerHour: 50,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Standard training and certification required'
    },
    {
      id: 'LP-003',
      name: 'Security Guard - Grade C',
      description: 'Entry-level security guard with basic training',
      department: 'Security Operations',
      costPerHour: 28,
      sellingRatePerHour: 40,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Basic training and orientation included'
    },
    {
      id: 'LP-004',
      name: 'Security Supervisor',
      description: 'Shift supervisor for security operations',
      department: 'Security Operations',
      costPerHour: 60,
      sellingRatePerHour: 85,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Manages team of 10-15 guards, reporting and coordination'
    },
    {
      id: 'LP-005',
      name: 'Senior Engineer - Installation',
      description: 'Senior technical engineer for system installation and configuration',
      department: 'Technical Services',
      costPerHour: 300,
      sellingRatePerHour: 400,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Certified in multiple security systems, 10+ years experience'
    },
    {
      id: 'LP-006',
      name: 'Technical Engineer',
      description: 'Mid-level engineer for installation and maintenance',
      department: 'Technical Services',
      costPerHour: 200,
      sellingRatePerHour: 280,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: '5+ years experience with CCTV, access control systems'
    },
    {
      id: 'LP-007',
      name: 'Junior Technician',
      description: 'Entry-level technician for basic installation and support',
      department: 'Technical Services',
      costPerHour: 120,
      sellingRatePerHour: 180,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Assists senior engineers, basic troubleshooting'
    },
    {
      id: 'LP-008',
      name: 'Project Manager',
      description: 'Project manager for security system deployments',
      department: 'Project Management',
      costPerHour: 350,
      sellingRatePerHour: 480,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Full project lifecycle management, reporting, client coordination'
    },
    {
      id: 'LP-009',
      name: 'System Consultant',
      description: 'Security systems consultant and designer',
      department: 'Consulting',
      costPerHour: 400,
      sellingRatePerHour: 550,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'System design, risk assessment, compliance consulting'
    },
    {
      id: 'LP-010',
      name: 'Maintenance Technician',
      description: 'Regular maintenance and support technician',
      department: 'Maintenance',
      costPerHour: 150,
      sellingRatePerHour: 220,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Preventive maintenance, troubleshooting, minor repairs'
    },
    {
      id: 'LP-011',
      name: 'Control Room Operator',
      description: 'CCTV and alarm system monitoring operator',
      department: 'Monitoring Services',
      costPerHour: 38,
      sellingRatePerHour: 55,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: '24/7 monitoring, incident response, reporting'
    },
    {
      id: 'LP-012',
      name: 'Training Specialist',
      description: 'Security training and development specialist',
      department: 'Training',
      costPerHour: 250,
      sellingRatePerHour: 350,
      status: 'active',
      createdAt: '2024-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Staff training, certification programs, compliance training'
    }
  ])

  // Getters
  function getPositionById(id: string): LaborPosition | undefined {
    return positions.value.find(p => p.id === id)
  }

  function getActivePositions(): LaborPosition[] {
    return positions.value.filter(p => p.status === 'active')
  }

  function getPositionsByDepartment(department: string): LaborPosition[] {
    return positions.value.filter(p => p.department === department)
  }

  // CRUD Operations
  function addPosition(positionData: Omit<LaborPosition, 'id' | 'createdAt' | 'createdBy'>): LaborPosition {
    const newPosition: LaborPosition = {
      ...positionData,
      id: `LP-${String(positions.value.length + 1).padStart(3, '0')}`,
      createdAt: new Date().toISOString(),
      createdBy: 'USR-001' // Should come from auth
    }
    positions.value.push(newPosition)
    return newPosition
  }

  function updatePosition(id: string, updates: Partial<LaborPosition>): boolean {
    const index = positions.value.findIndex(p => p.id === id)
    if (index !== -1) {
      positions.value[index] = {
        ...positions.value[index],
        ...updates
      }
      return true
    }
    return false
  }

  function deletePosition(id: string): boolean {
    const index = positions.value.findIndex(p => p.id === id)
    if (index !== -1) {
      positions.value.splice(index, 1)
      return true
    }
    return false
  }

  function togglePositionStatus(id: string): boolean {
    const position = positions.value.find(p => p.id === id)
    if (position) {
      position.status = position.status === 'active' ? 'inactive' : 'active'
      return true
    }
    return false
  }

  // Calculate monthly cost/rate
  function calculateMonthlyCost(hourlyRate: number, hoursPerMonth: number = 160): number {
    return hourlyRate * hoursPerMonth
  }

  function calculateMarginPercent(costRate: number, sellingRate: number): number {
    if (sellingRate === 0) return 0
    return ((sellingRate - costRate) / sellingRate) * 100
  }

  return {
    positions,
    getPositionById,
    getActivePositions,
    getPositionsByDepartment,
    addPosition,
    updatePosition,
    deletePosition,
    togglePositionStatus,
    calculateMonthlyCost,
    calculateMarginPercent
  }
})
