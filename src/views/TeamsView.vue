<template>
  <div class="teams-view">
    <div class="view-header">
      <div class="header-actions">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Search teams or members..."
          class="search-input"
        />
        <select v-model="departmentFilter" class="filter-select">
          <option value="">All Departments</option>
          <option value="Sales">Sales</option>
          <option value="Pre-Sales">Pre-Sales</option>
          <option value="Marketing">Marketing</option>
          <option value="Technical">Technical</option>
          <option value="Support">Support</option>
          <option value="Management">Management</option>
          <option value="Operations">Operations</option>
          <option value="Other">Other</option>
        </select>
        <button class="btn btn-primary" @click="showAddTeamModal">
          <span>➕</span> New Team
        </button>
      </div>
    </div>

    <!-- Teams Grid -->
    <div class="teams-grid">
      <div
        v-for="team in filteredTeams"
        :key="team.id"
        class="team-card"
        :class="{ inactive: team.status === 'inactive' }"
      >
        <div class="team-header">
          <div class="team-info">
            <h3 class="team-name">{{ team.name }}</h3>
            <BaseBadge :variant="team.status === 'active' ? 'success' : 'default'" size="sm">
              {{ team.status }}
            </BaseBadge>
            <BaseBadge variant="info" size="sm">{{ team.department }}</BaseBadge>
          </div>
          <div class="team-actions">
            <button class="action-btn" title="Edit Team" @click="editTeam(team)">✏️</button>
            <button class="action-btn" title="Add Member" @click="showAddMemberModal(team)">➕👤</button>
            <button class="action-btn" title="Toggle Status" @click="toggleTeamStatus(team.id)">
              {{ team.status === 'active' ? '🔒' : '🔓' }}
            </button>
          </div>
        </div>

        <p class="team-description">{{ team.description }}</p>

        <div class="team-stats">
          <div class="stat-item">
            <span class="stat-label">Members:</span>
            <span class="stat-value">{{ team.members.length }}</span>
          </div>
          <div class="stat-item">
            <span class="stat-label">Active:</span>
            <span class="stat-value">{{ getActiveMembers(team).length }}</span>
          </div>
          <div class="stat-item" v-if="team.leaderId">
            <span class="stat-label">Leader:</span>
            <span class="stat-value">{{ getTeamLeaderName(team) }}</span>
          </div>
        </div>

        <div class="team-members">
          <div class="members-header">
            <strong>Team Members</strong>
          </div>
          <div class="members-list">
            <div
              v-for="member in team.members"
              :key="member.id"
              class="member-item"
              :class="{ inactive: member.status === 'inactive', leader: member.id === team.leaderId }"
            >
              <div class="member-info">
                <div class="member-avatar">{{ member.name.charAt(0) }}</div>
                <div class="member-details">
                  <div class="member-name">
                    {{ member.name }}
                    <span v-if="member.id === team.leaderId" class="leader-badge">👑 Leader</span>
                  </div>
                  <div class="member-role">{{ member.role }}</div>
                  <div class="member-contact">{{ member.email }}</div>
                </div>
              </div>
              <div class="member-actions">
                <button class="action-btn-sm" title="Edit Member" @click="editMember(team, member)">✏️</button>
                <button class="action-btn-sm" title="Remove Member" @click="removeMember(team, member)">🗑️</button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="team.notes" class="team-notes">
          <strong>Notes:</strong> {{ team.notes }}
        </div>
      </div>
    </div>

    <!-- Add/Edit Team Modal -->
    <BaseModal
      v-model="showTeamModal"
      :title="editingTeam ? 'Edit Team' : 'Create New Team'"
      size="md"
      @confirm="saveTeam"
    >
      <div class="modal-form">
        <div class="form-group">
          <label class="form-label">Team Name *</label>
          <input v-model="teamForm.name" type="text" class="form-input" placeholder="e.g., Sales Team" />
        </div>

        <div class="form-group">
          <label class="form-label">Department *</label>
          <select v-model="teamForm.department" class="form-input">
            <option value="">Select department</option>
            <option value="Sales">Sales</option>
            <option value="Pre-Sales">Pre-Sales</option>
            <option value="Marketing">Marketing</option>
            <option value="Technical">Technical</option>
            <option value="Support">Support</option>
            <option value="Management">Management</option>
            <option value="Operations">Operations</option>
            <option value="Other">Other</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">Description</label>
          <textarea v-model="teamForm.description" class="form-input" rows="3" placeholder="Team description..."></textarea>
        </div>

        <div class="form-group">
          <label class="form-label">Status *</label>
          <select v-model="teamForm.status" class="form-input">
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">Notes</label>
          <textarea v-model="teamForm.notes" class="form-input" rows="2" placeholder="Additional notes..."></textarea>
        </div>
      </div>
    </BaseModal>

    <!-- Add/Edit Member Modal -->
    <BaseModal
      v-model="showMemberModal"
      :title="editingMember ? 'Edit Team Member' : 'Add Team Member'"
      size="md"
      @confirm="saveMember"
    >
      <div class="modal-form">
        <div class="form-group">
          <label class="form-label">Full Name *</label>
          <input v-model="memberForm.name" type="text" class="form-input" placeholder="e.g., Ahmed Al-Salem" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Email *</label>
            <input v-model="memberForm.email" type="email" class="form-input" placeholder="email@company.com" />
          </div>
          <div class="form-group">
            <label class="form-label">Phone</label>
            <input v-model="memberForm.phone" type="tel" class="form-input" placeholder="+966 XX XXX XXXX" />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Role/Title *</label>
            <input v-model="memberForm.role" type="text" class="form-input" placeholder="e.g., Sales Executive" />
          </div>
          <div class="form-group">
            <label class="form-label">Status *</label>
            <select v-model="memberForm.status" class="form-input">
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useTeamsStore } from '@/stores/teams'
import { useToast } from '@/composables/useToast'
import BaseModal from '@/components/UI/BaseModal.vue'
import BaseBadge from '@/components/UI/BaseBadge.vue'
import type { Team, TeamMember, Status } from '@/types'

const teamsStore = useTeamsStore()
const { success, info } = useToast()

const searchQuery = ref('')
const departmentFilter = ref('')
const showTeamModal = ref(false)
const showMemberModal = ref(false)
const editingTeam = ref<Team | null>(null)
const editingMember = ref<TeamMember | null>(null)
const selectedTeam = ref<Team | null>(null)

const teamForm = ref({
  name: '',
  description: '',
  department: '' as 'Sales' | 'Pre-Sales' | 'Technical' | 'Support' | 'Marketing' | 'Management' | 'Operations' | 'Other' | '',
  status: 'active' as Status,
  notes: ''
})

const memberForm = ref({
  name: '',
  email: '',
  role: '',
  phone: '',
  status: 'active' as Status
})

// Computed
const filteredTeams = computed(() => {
  let teams = teamsStore.teams

  if (departmentFilter.value) {
    teams = teams.filter(t => t.department === departmentFilter.value)
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    teams = teams.filter(t =>
      t.name.toLowerCase().includes(query) ||
      t.description.toLowerCase().includes(query) ||
      t.members.some(m =>
        m.name.toLowerCase().includes(query) ||
        m.email.toLowerCase().includes(query)
      )
    )
  }

  return teams
})

// Helper functions
function getActiveMembers(team: Team): TeamMember[] {
  return team.members.filter(m => m.status === 'active')
}

function getTeamLeaderName(team: Team): string {
  const leader = team.members.find(m => m.id === team.leaderId)
  return leader ? leader.name : 'No leader assigned'
}

// Team Operations
function showAddTeamModal() {
  editingTeam.value = null
  teamForm.value = {
    name: '',
    description: '',
    department: '',
    status: 'active',
    notes: ''
  }
  showTeamModal.value = true
}

function editTeam(team: Team) {
  editingTeam.value = team
  teamForm.value = {
    name: team.name,
    description: team.description,
    department: team.department,
    status: team.status,
    notes: team.notes || ''
  }
  showTeamModal.value = true
}

function saveTeam() {
  if (!teamForm.value.name || !teamForm.value.department) {
    info('Please fill in all required fields')
    return
  }

  if (editingTeam.value) {
    teamsStore.updateTeam(editingTeam.value.id, {
      name: teamForm.value.name,
      description: teamForm.value.description,
      department: teamForm.value.department as any,
      status: teamForm.value.status,
      notes: teamForm.value.notes
    })
    success('Team updated successfully')
  } else {
    teamsStore.addTeam({
      name: teamForm.value.name,
      description: teamForm.value.description,
      department: teamForm.value.department as any,
      members: [],
      status: teamForm.value.status,
      notes: teamForm.value.notes
    })
    success('Team created successfully')
  }

  showTeamModal.value = false
  editingTeam.value = null
}

function toggleTeamStatus(teamId: string) {
  if (teamsStore.toggleTeamStatus(teamId)) {
    success('Team status updated')
  }
}

// Member Operations
function showAddMemberModal(team: Team) {
  selectedTeam.value = team
  editingMember.value = null
  memberForm.value = {
    name: '',
    email: '',
    role: '',
    phone: '',
    status: 'active'
  }
  showMemberModal.value = true
}

function editMember(team: Team, member: TeamMember) {
  selectedTeam.value = team
  editingMember.value = member
  memberForm.value = {
    name: member.name,
    email: member.email,
    role: member.role,
    phone: member.phone || '',
    status: member.status
  }
  showMemberModal.value = true
}

function saveMember() {
  if (!memberForm.value.name || !memberForm.value.email || !memberForm.value.role) {
    info('Please fill in all required fields')
    return
  }

  if (!selectedTeam.value) return

  if (editingMember.value) {
    teamsStore.updateTeamMember(selectedTeam.value.id, editingMember.value.id, {
      name: memberForm.value.name,
      email: memberForm.value.email,
      role: memberForm.value.role,
      phone: memberForm.value.phone,
      status: memberForm.value.status
    })
    success('Member updated successfully')
  } else {
    teamsStore.addTeamMember(selectedTeam.value.id, {
      name: memberForm.value.name,
      email: memberForm.value.email,
      role: memberForm.value.role,
      phone: memberForm.value.phone,
      status: memberForm.value.status
    })
    success('Member added successfully')
  }

  showMemberModal.value = false
  editingMember.value = null
  selectedTeam.value = null
}

function removeMember(team: Team, member: TeamMember) {
  if (confirm(`Remove ${member.name} from ${team.name}?`)) {
    if (teamsStore.removeTeamMember(team.id, member.id)) {
      success('Member removed successfully')
    }
  }
}
</script>

<style scoped>
.teams-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 1rem;
  width: 100%;
}

.search-input {
  flex: 1;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-select {
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-primary);
  border-radius: 8px;
  font-size: 0.95rem;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-primary {
  background: var(--color-primary);
  color: #fff;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
  transform: translateY(-2px);
}

/* Teams Grid */
.teams-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(450px, 1fr));
  gap: 1.5rem;
}

.team-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 1.5rem;
  transition: all 0.3s;
}

.team-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.team-card.inactive {
  opacity: 0.6;
}

.team-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.team-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.team-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.team-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  background: none;
  border: none;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.action-btn:hover {
  background: var(--color-surface-hover);
  transform: scale(1.1);
}

.team-description {
  color: var(--color-text-secondary);
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.team-stats {
  display: flex;
  gap: 1.5rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-primary);
}

/* Team Members */
.team-members {
  margin-top: 1rem;
}

.members-header {
  margin-bottom: 0.75rem;
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.member-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  transition: all 0.2s;
}

.member-item:hover {
  border-color: var(--color-primary);
  background: var(--color-surface);
}

.member-item.inactive {
  opacity: 0.5;
}

.member-item.leader {
  border-color: #f59e0b;
  background: rgba(245, 158, 11, 0.05);
}

.member-info {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.member-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1rem;
}

.member-details {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.member-name {
  font-weight: 600;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.leader-badge {
  font-size: 0.75rem;
  padding: 0.125rem 0.5rem;
  background: #f59e0b;
  color: white;
  border-radius: 12px;
}

.member-role {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.member-contact {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.member-actions {
  display: flex;
  gap: 0.25rem;
}

.action-btn-sm {
  background: none;
  border: none;
  font-size: 0.9rem;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.action-btn-sm:hover {
  background: var(--color-surface-hover);
  transform: scale(1.1);
}

.team-notes {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

/* Modal Form Styles */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-weight: 500;
  color: var(--color-text-primary);
  font-size: 0.9rem;
}

.form-input {
  padding: 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-surface);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}
</style>
