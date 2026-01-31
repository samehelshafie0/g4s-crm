import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Team, TeamMember, Status } from '@/types'

export const useTeamsStore = defineStore('teams', () => {
  const teams = ref<Team[]>([
    {
      id: 'TEAM-001',
      name: 'Sales Team',
      description: 'Main sales team responsible for closing deals and customer relationships',
      department: 'Sales',
      leaderId: 'MEM-001',
      members: [
        {
          id: 'MEM-001',
          name: 'Ahmed Al-Salem',
          email: 'ahmed.salem@company.com',
          role: 'Sales Manager',
          phone: '+966 50 123 4567',
          status: 'active',
          joinedAt: '2023-01-15T00:00:00Z'
        },
        {
          id: 'MEM-002',
          name: 'Fatima Al-Otaibi',
          email: 'fatima.otaibi@company.com',
          role: 'Senior Sales Executive',
          phone: '+966 50 234 5678',
          status: 'active',
          joinedAt: '2023-02-01T00:00:00Z'
        },
        {
          id: 'MEM-003',
          name: 'Mohammed Al-Qahtani',
          email: 'mohammed.qahtani@company.com',
          role: 'Sales Executive',
          phone: '+966 50 345 6789',
          status: 'active',
          joinedAt: '2023-03-10T00:00:00Z'
        }
      ],
      status: 'active',
      createdAt: '2023-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Primary sales team handling all major accounts'
    },
    {
      id: 'TEAM-002',
      name: 'Pre-Sales Team',
      description: 'Technical pre-sales team for solution design and proposals',
      department: 'Pre-Sales',
      leaderId: 'MEM-004',
      members: [
        {
          id: 'MEM-004',
          name: 'Khalid Al-Mutairi',
          email: 'khalid.mutairi@company.com',
          role: 'Pre-Sales Manager',
          phone: '+966 50 456 7890',
          status: 'active',
          joinedAt: '2023-01-15T00:00:00Z'
        },
        {
          id: 'MEM-005',
          name: 'Nora Al-Harbi',
          email: 'nora.harbi@company.com',
          role: 'Pre-Sales Consultant',
          phone: '+966 50 567 8901',
          status: 'active',
          joinedAt: '2023-02-15T00:00:00Z'
        },
        {
          id: 'MEM-006',
          name: 'Omar Al-Dosari',
          email: 'omar.dosari@company.com',
          role: 'Technical Consultant',
          phone: '+966 50 678 9012',
          status: 'active',
          joinedAt: '2023-03-01T00:00:00Z'
        }
      ],
      status: 'active',
      createdAt: '2023-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Pre-sales and technical consulting team'
    },
    {
      id: 'TEAM-003',
      name: 'Marketing Team',
      description: 'Marketing and lead generation team',
      department: 'Marketing',
      leaderId: 'MEM-007',
      members: [
        {
          id: 'MEM-007',
          name: 'Sara Al-Fahad',
          email: 'sara.fahad@company.com',
          role: 'Marketing Manager',
          phone: '+966 50 789 0123',
          status: 'active',
          joinedAt: '2023-01-20T00:00:00Z'
        },
        {
          id: 'MEM-008',
          name: 'Abdullah Al-Saud',
          email: 'abdullah.saud@company.com',
          role: 'Marketing Specialist',
          phone: '+966 50 890 1234',
          status: 'active',
          joinedAt: '2023-02-20T00:00:00Z'
        }
      ],
      status: 'active',
      createdAt: '2023-01-01T00:00:00Z',
      createdBy: 'USR-001',
      notes: 'Lead generation and marketing campaigns'
    }
  ])

  // Getters
  const activeTeams = computed(() => teams.value.filter(t => t.status === 'active'))

  function getTeamById(id: string): Team | undefined {
    return teams.value.find(t => t.id === id)
  }

  function getTeamsByDepartment(department: string): Team[] {
    return teams.value.filter(t => t.department === department)
  }

  function getTeamMemberById(teamId: string, memberId: string): TeamMember | undefined {
    const team = getTeamById(teamId)
    return team?.members.find(m => m.id === memberId)
  }

  function getAllTeamMembers(): TeamMember[] {
    const allMembers: TeamMember[] = []
    teams.value.forEach(team => {
      allMembers.push(...team.members)
    })
    return allMembers
  }

  function getActiveSalesPeople(): TeamMember[] {
    const salesTeams = teams.value.filter(t =>
      (t.department === 'Sales' || t.department === 'Pre-Sales') &&
      t.status === 'active'
    )
    const salesPeople: TeamMember[] = []
    salesTeams.forEach(team => {
      salesPeople.push(...team.members.filter(m => m.status === 'active'))
    })
    return salesPeople
  }

  // CRUD Operations
  function addTeam(teamData: Omit<Team, 'id' | 'createdAt' | 'createdBy'>): Team {
    const newTeam: Team = {
      ...teamData,
      id: `TEAM-${String(teams.value.length + 1).padStart(3, '0')}`,
      createdAt: new Date().toISOString(),
      createdBy: 'USR-001' // Should come from auth
    }
    teams.value.push(newTeam)
    return newTeam
  }

  function updateTeam(id: string, updates: Partial<Team>): boolean {
    const index = teams.value.findIndex(t => t.id === id)
    if (index !== -1) {
      teams.value[index] = {
        ...teams.value[index],
        ...updates,
        updatedAt: new Date().toISOString()
      }
      return true
    }
    return false
  }

  function deleteTeam(id: string): boolean {
    const index = teams.value.findIndex(t => t.id === id)
    if (index !== -1) {
      teams.value.splice(index, 1)
      return true
    }
    return false
  }

  function toggleTeamStatus(id: string): boolean {
    const team = teams.value.find(t => t.id === id)
    if (team) {
      team.status = team.status === 'active' ? 'inactive' : 'active'
      team.updatedAt = new Date().toISOString()
      return true
    }
    return false
  }

  // Team Member Operations
  function addTeamMember(teamId: string, memberData: Omit<TeamMember, 'id' | 'joinedAt'>): TeamMember | null {
    const team = getTeamById(teamId)
    if (!team) return null

    const newMember: TeamMember = {
      ...memberData,
      id: `MEM-${String(getAllTeamMembers().length + 1).padStart(3, '0')}`,
      joinedAt: new Date().toISOString()
    }

    team.members.push(newMember)
    team.updatedAt = new Date().toISOString()
    return newMember
  }

  function updateTeamMember(teamId: string, memberId: string, updates: Partial<TeamMember>): boolean {
    const team = getTeamById(teamId)
    if (!team) return false

    const memberIndex = team.members.findIndex(m => m.id === memberId)
    if (memberIndex !== -1) {
      team.members[memberIndex] = {
        ...team.members[memberIndex],
        ...updates
      }
      team.updatedAt = new Date().toISOString()
      return true
    }
    return false
  }

  function removeTeamMember(teamId: string, memberId: string): boolean {
    const team = getTeamById(teamId)
    if (!team) return false

    const memberIndex = team.members.findIndex(m => m.id === memberId)
    if (memberIndex !== -1) {
      team.members.splice(memberIndex, 1)
      team.updatedAt = new Date().toISOString()
      return true
    }
    return false
  }

  return {
    teams,
    activeTeams,
    getTeamById,
    getTeamsByDepartment,
    getTeamMemberById,
    getAllTeamMembers,
    getActiveSalesPeople,
    addTeam,
    updateTeam,
    deleteTeam,
    toggleTeamStatus,
    addTeamMember,
    updateTeamMember,
    removeTeamMember
  }
})
