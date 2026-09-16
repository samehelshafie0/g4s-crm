<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  Plus,
  Pencil,
  Trash2,
  X,
  Users,
  UserPlus,
  UserMinus,
  Crown,
  Mail,
  Phone,
  ChevronDown,
  ChevronUp,
} from 'lucide-vue-next'
import type { Team, TeamMember, Department } from '@/types'
import { useTeamsStore } from '@/stores/teams'

function uid(): string {
  return Math.random().toString(36).slice(2, 11)
}

const teamsStore = useTeamsStore()
onMounted(() => teamsStore.fetchTeams())

const departmentLabels: Record<Department, string> = {
  sales: 'Sales',
  'pre-sales': 'Pre-Sales',
  technical: 'Technical',
  support: 'Support',
  marketing: 'Marketing',
  management: 'Management',
  operations: 'Operations',
}

const departmentBadge: Record<Department, string> = {
  sales: 'badge-primary',
  'pre-sales': 'badge-info',
  technical: 'badge-warning',
  support: 'badge-success',
  marketing: 'badge-danger',
  management: 'badge-gray',
  operations: 'badge-primary',
}

const teams = ref<Team[]>([
  {
    id: uid(), name: 'Enterprise Sales', department: 'sales',
    description: 'Handles key accounts and enterprise-level security solution sales across Saudi Arabia.',
    leaderId: 'm1', leaderName: 'Ahmed Al-Dosari', isActive: true,
    members: [
      { id: 'm1', name: 'Ahmed Al-Dosari', email: 'ahmed.dosari@g4s.sa', phone: '+966 50 111 2233', role: 'Sales Director', department: 'sales', isLeader: true },
      { id: 'm2', name: 'Fahad Al-Otaibi', email: 'fahad.otaibi@g4s.sa', phone: '+966 50 222 3344', role: 'Senior Account Executive', department: 'sales', isLeader: false },
      { id: 'm3', name: 'Noura Al-Shammari', email: 'noura.shammari@g4s.sa', phone: '+966 50 333 4455', role: 'Account Executive', department: 'sales', isLeader: false },
      { id: 'm4', name: 'Youssef Al-Harbi', email: 'youssef.harbi@g4s.sa', phone: '+966 50 444 5566', role: 'Sales Coordinator', department: 'sales', isLeader: false },
    ],
    createdAt: '2024-01-01T08:00:00Z', updatedAt: '2026-02-10T08:00:00Z',
  },
  {
    id: uid(), name: 'Solutions Engineering', department: 'pre-sales',
    description: 'Designs and architects integrated security solutions. Supports sales with technical proposals and BoMs.',
    leaderId: 'm5', leaderName: 'Omar Al-Zahrani', isActive: true,
    members: [
      { id: 'm5', name: 'Omar Al-Zahrani', email: 'omar.zahrani@g4s.sa', phone: '+966 50 555 6677', role: 'Head of Pre-Sales', department: 'pre-sales', isLeader: true },
      { id: 'm6', name: 'Salman Al-Mutairi', email: 'salman.mutairi@g4s.sa', phone: '+966 50 666 7788', role: 'Solutions Architect', department: 'pre-sales', isLeader: false },
      { id: 'm7', name: 'Reem Al-Ghamdi', email: 'reem.ghamdi@g4s.sa', phone: '+966 50 777 8899', role: 'Pre-Sales Engineer', department: 'pre-sales', isLeader: false },
    ],
    createdAt: '2024-01-15T08:00:00Z', updatedAt: '2026-01-20T08:00:00Z',
  },
  {
    id: uid(), name: 'Field Operations', department: 'operations',
    description: 'Manages on-site installations, maintenance, and guarding operations across all regions.',
    leaderId: 'm8', leaderName: 'Khalid Al-Qahtani', isActive: true,
    members: [
      { id: 'm8', name: 'Khalid Al-Qahtani', email: 'khalid.qahtani@g4s.sa', phone: '+966 50 888 9900', role: 'Operations Manager', department: 'operations', isLeader: true },
      { id: 'm9', name: 'Tariq Al-Sulaiman', email: 'tariq.sulaiman@g4s.sa', phone: '+966 50 999 0011', role: 'Field Supervisor', department: 'operations', isLeader: false },
      { id: 'm10', name: 'Hassan Al-Rashid', email: 'hassan.rashid@g4s.sa', phone: '+966 50 100 1122', role: 'Installation Technician', department: 'operations', isLeader: false },
      { id: 'm11', name: 'Majed Al-Faisal', email: 'majed.faisal@g4s.sa', phone: '+966 50 100 2233', role: 'Installation Technician', department: 'operations', isLeader: false },
      { id: 'm12', name: 'Ibrahim Al-Turki', email: 'ibrahim.turki@g4s.sa', phone: '+966 50 100 3344', role: 'Maintenance Engineer', department: 'operations', isLeader: false },
    ],
    createdAt: '2024-02-01T08:00:00Z', updatedAt: '2026-02-15T08:00:00Z',
  },
  {
    id: uid(), name: 'Customer Support', department: 'support',
    description: 'Provides 24/7 helpdesk and escalation support for all active contracts and service-level agreements.',
    leaderId: 'm13', leaderName: 'Lina Al-Asmari', isActive: true,
    members: [
      { id: 'm13', name: 'Lina Al-Asmari', email: 'lina.asmari@g4s.sa', phone: '+966 50 100 4455', role: 'Support Team Lead', department: 'support', isLeader: true },
      { id: 'm14', name: 'Maha Al-Dossary', email: 'maha.dossary@g4s.sa', phone: '+966 50 100 5566', role: 'Support Specialist', department: 'support', isLeader: false },
      { id: 'm15', name: 'Saad Al-Jaber', email: 'saad.jaber@g4s.sa', phone: '+966 50 100 6677', role: 'Support Specialist', department: 'support', isLeader: false },
    ],
    createdAt: '2024-03-01T08:00:00Z', updatedAt: '2026-02-20T08:00:00Z',
  },
])

const activeDepartment = ref<Department | 'all'>('all')
const expandedTeams = ref<Set<string>>(new Set())

const filteredTeams = computed(() => {
  if (activeDepartment.value === 'all') return teams.value
  return teams.value.filter(t => t.department === activeDepartment.value)
})

function toggleExpand(teamId: string) {
  if (expandedTeams.value.has(teamId)) {
    expandedTeams.value.delete(teamId)
  } else {
    expandedTeams.value.add(teamId)
  }
}

function getInitials(name: string): string {
  return name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
}

// Modal state
const showModal = ref(false)
const editingId = ref<string | null>(null)

interface TeamForm {
  name: string
  department: Department
  description: string
  leaderId: string
  leaderName: string
  members: TeamMember[]
  isActive: boolean
}

const defaultForm = (): TeamForm => ({
  name: '',
  department: 'sales',
  description: '',
  leaderId: '',
  leaderName: '',
  members: [],
  isActive: true,
})

const form = ref<TeamForm>(defaultForm())

const newMember = ref({
  name: '',
  email: '',
  phone: '',
  role: '',
})

function openAddModal() {
  editingId.value = null
  form.value = defaultForm()
  showModal.value = true
}

function openEditModal(t: Team) {
  editingId.value = t.id
  form.value = {
    name: t.name,
    department: t.department,
    description: t.description,
    leaderId: t.leaderId,
    leaderName: t.leaderName,
    members: JSON.parse(JSON.stringify(t.members)),
    isActive: t.isActive,
  }
  showModal.value = true
}

function addMember() {
  if (!newMember.value.name) return
  const member: TeamMember = {
    id: uid(),
    name: newMember.value.name,
    email: newMember.value.email,
    phone: newMember.value.phone,
    role: newMember.value.role,
    department: form.value.department,
    isLeader: false,
  }
  form.value.members.push(member)
  newMember.value = { name: '', email: '', phone: '', role: '' }
}

function removeMember(id: string) {
  form.value.members = form.value.members.filter(m => m.id !== id)
  if (form.value.leaderId === id) {
    form.value.leaderId = ''
    form.value.leaderName = ''
  }
}

function setLeader(member: TeamMember) {
  form.value.members.forEach(m => (m.isLeader = false))
  member.isLeader = true
  form.value.leaderId = member.id
  form.value.leaderName = member.name
}

function saveTeam() {
  const now = new Date().toISOString()
  const base = {
    name: form.value.name,
    department: form.value.department,
    description: form.value.description,
    leaderId: form.value.leaderId,
    leaderName: form.value.leaderName,
    members: form.value.members,
    isActive: form.value.isActive,
  }

  if (editingId.value) {
    const idx = teams.value.findIndex(t => t.id === editingId.value)
    if (idx !== -1) {
      teams.value[idx] = { ...teams.value[idx], ...base, updatedAt: now } as Team
    }
  } else {
    teams.value.push({ id: uid(), ...base, createdAt: now, updatedAt: now })
  }
  showModal.value = false
}

function deleteTeam(id: string) {
  teams.value = teams.value.filter(t => t.id !== id)
}
</script>

<template>
  <div class="teams-page">
    <div class="page-header">
      <div>
        <h1 class="page-header-title">Teams</h1>
        <p class="page-header-subtitle">{{ filteredTeams.length }} team{{ filteredTeams.length !== 1 ? 's' : '' }}</p>
      </div>
      <button class="btn btn-primary" @click="openAddModal">
        <Plus :size="18" />
        Add Team
      </button>
    </div>

    <!-- Department Tabs -->
    <div class="card mb-6">
      <div class="toolbar">
        <div class="tab-row">
          <button
            :class="['tab-btn', activeDepartment === 'all' && 'tab-btn--active']"
            @click="activeDepartment = 'all'"
          >All</button>
          <button
            v-for="(label, key) in departmentLabels"
            :key="key"
            :class="['tab-btn', activeDepartment === key && 'tab-btn--active']"
            @click="activeDepartment = key as Department"
          >{{ label }}</button>
        </div>
      </div>
    </div>

    <!-- Teams Grid -->
    <div class="teams-grid">
      <div v-for="team in filteredTeams" :key="team.id" class="card team-card">
        <div class="card-body">
          <div class="team-card-header">
            <div>
              <h3 class="team-name">{{ team.name }}</h3>
              <span :class="['badge', departmentBadge[team.department]]">{{ departmentLabels[team.department] }}</span>
            </div>
            <span :class="['badge badge-dot', team.isActive ? 'badge-success' : 'badge-danger']">
              {{ team.isActive ? 'Active' : 'Inactive' }}
            </span>
          </div>

          <p class="team-desc">{{ team.description }}</p>

          <!-- Leader highlight -->
          <div class="leader-card">
            <div class="avatar" :class="departmentBadge[team.department]">{{ getInitials(team.leaderName) }}</div>
            <div>
              <div class="leader-name">
                {{ team.leaderName }}
                <Crown :size="14" class="leader-crown" />
              </div>
              <div class="leader-role">Team Leader</div>
            </div>
          </div>

          <!-- Member count and expand toggle -->
          <button class="members-toggle" @click="toggleExpand(team.id)">
            <Users :size="16" />
            <span>{{ team.members.length }} member{{ team.members.length !== 1 ? 's' : '' }}</span>
            <component :is="expandedTeams.has(team.id) ? ChevronUp : ChevronDown" :size="16" />
          </button>

          <!-- Expandable member list -->
          <div v-if="expandedTeams.has(team.id)" class="members-list">
            <div v-for="member in team.members" :key="member.id" class="member-row">
              <div class="avatar avatar-sm" :class="departmentBadge[team.department]">{{ getInitials(member.name) }}</div>
              <div class="member-info">
                <div class="member-name">
                  {{ member.name }}
                  <Crown v-if="member.isLeader" :size="12" class="leader-crown" />
                </div>
                <div class="member-role">{{ member.role }}</div>
                <div class="member-contact">
                  <span><Mail :size="12" /> {{ member.email }}</span>
                  <span><Phone :size="12" /> {{ member.phone }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="team-card-footer">
            <button class="btn btn-secondary btn-sm" @click="openEditModal(team)">
              <Pencil :size="14" />
              Edit
            </button>
            <button class="btn btn-ghost btn-sm btn-danger" @click="deleteTeam(team.id)">
              <Trash2 :size="14" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!filteredTeams.length" class="empty-state">
      <Users :size="48" class="empty-state-icon" />
      <h3 class="empty-state-title">No teams found</h3>
      <p class="empty-state-text">Try selecting a different department or add a new team.</p>
    </div>

    <!-- Add / Edit Team Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
        <div class="modal modal-lg">
          <div class="modal-header">
            <h2 class="modal-title">{{ editingId ? 'Edit Team' : 'Add Team' }}</h2>
            <button class="modal-close" @click="showModal = false"><X :size="20" /></button>
          </div>
          <div class="modal-body">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">Team Name <span class="required">*</span></label>
                <input v-model="form.name" type="text" class="form-input" placeholder="Team name" />
              </div>
              <div class="form-group">
                <label class="form-label">Department</label>
                <select v-model="form.department" class="form-select">
                  <option v-for="(label, key) in departmentLabels" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Description</label>
              <textarea v-model="form.description" class="form-textarea" rows="2" placeholder="Team description…" />
            </div>

            <div class="form-group">
              <label class="form-label">Team Leader</label>
              <select v-model="form.leaderId" class="form-select" @change="() => { const m = form.members.find(x => x.id === form.leaderId); if (m) setLeader(m); }">
                <option value="" disabled>Select leader from members</option>
                <option v-for="m in form.members" :key="m.id" :value="m.id">{{ m.name }} — {{ m.role }}</option>
              </select>
            </div>

            <label class="form-checkbox">
              <input v-model="form.isActive" type="checkbox" />
              Active
            </label>

            <div class="divider" />

            <!-- Members management -->
            <h4 class="section-title mb-4">Members ({{ form.members.length }})</h4>

            <div v-if="form.members.length" class="modal-members-list">
              <div v-for="member in form.members" :key="member.id" class="modal-member-row">
                <div class="avatar avatar-sm" :class="departmentBadge[form.department]">{{ getInitials(member.name) }}</div>
                <div class="modal-member-info">
                  <span class="font-medium">{{ member.name }}</span>
                  <span class="text-muted text-sm">{{ member.role }}</span>
                </div>
                <Crown v-if="member.isLeader" :size="14" class="leader-crown" />
                <button class="btn btn-ghost btn-icon btn-sm" title="Remove" @click="removeMember(member.id)">
                  <UserMinus :size="14" />
                </button>
              </div>
            </div>

            <div class="add-member-section">
              <h5 class="add-member-title"><UserPlus :size="14" /> Add Member</h5>
              <div class="form-row">
                <div class="form-group">
                  <input v-model="newMember.name" type="text" class="form-input" placeholder="Name" />
                </div>
                <div class="form-group">
                  <input v-model="newMember.role" type="text" class="form-input" placeholder="Role" />
                </div>
              </div>
              <div class="form-row">
                <div class="form-group">
                  <input v-model="newMember.email" type="email" class="form-input" placeholder="Email" />
                </div>
                <div class="form-group">
                  <input v-model="newMember.phone" type="tel" class="form-input" placeholder="Phone" />
                </div>
              </div>
              <button class="btn btn-secondary btn-sm" :disabled="!newMember.name" @click="addMember">
                <UserPlus :size="14" />
                Add Member
              </button>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary" @click="showModal = false">Cancel</button>
            <button
              class="btn btn-primary"
              :disabled="!form.name"
              @click="saveTeam"
            >
              {{ editingId ? 'Save Changes' : 'Add Team' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.teams-page {
  padding: var(--space-6);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  flex-wrap: wrap;
}

.tab-row {
  display: flex;
  gap: var(--space-1);
  flex-wrap: wrap;
}

.tab-btn {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  background: var(--content-surface);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-600);
  cursor: pointer;
  transition: all 0.15s;
}

.tab-btn:hover { background: var(--color-neutral-50); }

.tab-btn--active {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}

.teams-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: var(--space-4);
}

.team-card .card-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.team-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}

.team-name {
  font-size: var(--text-base);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-900);
  margin: 0 0 var(--space-1) 0;
}

.team-desc {
  font-size: var(--text-sm);
  color: var(--color-neutral-500);
  line-height: var(--leading-relaxed);
  margin: 0;
}

/* Leader card */
.leader-card {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--color-neutral-50);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-neutral-200);
}

.avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--text-xs);
  font-weight: var(--font-bold, 700);
  flex-shrink: 0;
}

.avatar-sm {
  width: 32px;
  height: 32px;
  font-size: 10px;
}

.avatar.badge-primary { background: var(--color-primary-light, #e0e7ff); color: var(--color-primary); }
.avatar.badge-info { background: #dbeafe; color: #2563eb; }
.avatar.badge-warning { background: #fef3c7; color: #d97706; }
.avatar.badge-success { background: #d1fae5; color: #059669; }
.avatar.badge-danger { background: #fee2e2; color: #dc2626; }
.avatar.badge-gray { background: var(--color-neutral-100); color: var(--color-neutral-600); }

.leader-name {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.leader-crown {
  color: #d97706;
}

.leader-role {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

/* Members toggle */
.members-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  width: 100%;
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  background: var(--content-surface);
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  color: var(--color-neutral-600);
  cursor: pointer;
  transition: background 0.15s;
}

.members-toggle:hover {
  background: var(--color-neutral-50);
}

/* Members list */
.members-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.member-row {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-neutral-100);
  border-radius: var(--radius-md);
}

.member-info {
  flex: 1;
  min-width: 0;
}

.member-name {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-800);
}

.member-role {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
  margin-bottom: 2px;
}

.member-contact {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.member-contact span {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--color-neutral-500);
}

.team-card-footer {
  display: flex;
  gap: var(--space-2);
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-neutral-200);
}

/* Modal members */
.section-title {
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-neutral-500);
}

.modal-members-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-bottom: var(--space-4);
}

.modal-member-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
}

.modal-member-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.add-member-section {
  padding: var(--space-4);
  background: var(--color-neutral-50);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--color-neutral-300);
}

.add-member-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-sm);
  font-weight: var(--font-semibold);
  color: var(--color-neutral-700);
  margin: 0 0 var(--space-3) 0;
}

.text-muted { color: var(--color-neutral-500); }
.text-sm { font-size: var(--text-sm); }

@media (max-width: 768px) {
  .teams-grid { grid-template-columns: 1fr; }
  .tab-row { flex-wrap: wrap; }
}
</style>
