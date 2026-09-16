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

import { teamsService, usersService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import type { UserLookup } from '@/services/workflowDtos'
const directory = ref<UserLookup[]>([])
const selectedMemberId = ref('')

const teamsStore = useTeamsStore()
onMounted(async()=>{try{teams.value=await allPages(teamsService.list);directory.value=(await usersService.lookup()).data}catch(e){window.alert(errorMessage(e))}})

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

const teams = ref<Team[]>([])
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
 const user=directory.value.find(u=>u.id===selectedMemberId.value)
 if(!user||form.value.members.some(m=>m.id===user.id))return
 form.value.members.push({id:user.id,name:`${user.firstName} ${user.lastName}`,email:user.email,phone:'',role:user.role,department:form.value.department,isLeader:false});selectedMemberId.value=''
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

const saving = ref(false)
async function saveTeam() {
 if(saving.value)return
 saving.value=true
 try {const old=editingId.value?(await teamsService.get(editingId.value)).data:null;const data={...form.value};const result=editingId.value?await teamsService.update(editingId.value,data):await teamsService.create(data);editingId.value=result.data.id
 for(const member of old?.members??[])if(!form.value.members.some(m=>m.id===member.id))await teamsService.removeMember(result.data.id,member.id)
 for(const member of form.value.members)if(!old?.members.some(m=>m.id===member.id))await teamsService.addMember(result.data.id,member.id)
 teams.value=await allPages(teamsService.list);showModal.value=false
 }catch(e){window.alert(errorMessage(e))}finally{saving.value=false}
}
async function deleteTeam(id:string){try{await teamsService.delete(id);teams.value=teams.value.filter(t=>t.id!==id)}catch(e){window.alert(errorMessage(e))}}
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
              <h5 class="add-member-title">Add an existing user</h5>
              <label>User<select v-model="selectedMemberId" class="select"><option value="">Select a user</option><option v-for="user in directory" :key="user.id" :value="user.id">{{ user.firstName }} {{ user.lastName }} — {{ user.email }}</option></select></label>
              <button class="btn btn-secondary btn-sm" :disabled="!selectedMemberId" @click="addMember">
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
