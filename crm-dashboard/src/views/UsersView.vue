<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppDialog from '@/components/shared/AppDialog.vue'
import { usersService, teamsService } from '@/services'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import type { AuthUser } from '@/services/auth.service'
import type { Team } from '@/types'
import { roles, departments, roleLabel, passwordError } from '@/services/userOptions'
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
const users = ref<AuthUser[]>([])
const teams = ref<Team[]>([])
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const success = ref('')
const search = ref('')
const role = ref('')
const status = ref('active')
const editing = ref<AuthUser | null>(null)
const showEditor = ref(false)
const showReset = ref(false)
const showDeactivate = ref(false)
const target = ref<AuthUser | null>(null)
const formError = ref('')
const reset = ref({password:'', confirmation:''})
const blank = () => ({firstName:'', lastName:'', email:'', phone:'', role:'sales_executive', department:'sales', teamId:'', password:'', confirmation:'', isActive:true})
const form = ref(blank())
const filtered = computed(() => users.value.filter(user => `${user.firstName} ${user.lastName} ${user.email}`.toLowerCase().includes(search.value.trim().toLowerCase()) && (!role.value || user.role === role.value) && (!status.value || user.isActive === (status.value === 'active'))))
const lastAdmin = computed(() => editing.value?.role === 'admin' && editing.value.isActive && users.value.filter(u => u.role === 'admin' && u.isActive).length === 1)
async function load() {
 loading.value = true; error.value = ''
 try { [users.value, teams.value] = await Promise.all([allPages(usersService.list), allPages(teamsService.list)]) }
 catch(e) { error.value = errorMessage(e) } finally { loading.value = false }
}
onMounted(load)
function edit(user?: AuthUser) {
 editing.value = user ?? null
 form.value = user ? {...blank(), ...user, phone:user.phone ?? '', teamId:user.teamId ?? ''} : blank()
 formError.value = ''; showEditor.value = true
}
async function save() {
 if (busy.value) return
 formError.value = editing.value ? '' : passwordError(form.value.password, form.value.confirmation)
 if (formError.value) return
 busy.value = true
 try {
  if (editing.value) await usersService.update(editing.value.id, form.value)
  else await usersService.create(form.value)
  showEditor.value = false; form.value = blank(); success.value = editing.value ? 'User updated.' : 'User created. Share the password with them securely.'
  await load()
  if (editing.value?.id === auth.user?.id) await auth.fetchMe()
 } catch(e) { formError.value = errorMessage(e) } finally { busy.value = false }
}
function openReset(user: AuthUser) { target.value = user; reset.value = {password:'', confirmation:''}; formError.value = ''; showReset.value = true }
async function resetPassword() {
 if (busy.value || !target.value) return
 formError.value = passwordError(reset.value.password, reset.value.confirmation)
 if (formError.value) return
 busy.value = true
 try { await usersService.resetPassword(target.value.id, reset.value.password); reset.value = {password:'',confirmation:''}; showReset.value = false; success.value = 'Password reset. Share the new password securely.' }
 catch(e) { formError.value = errorMessage(e) } finally { busy.value = false }
}
function deactivate(user: AuthUser) { target.value = user; formError.value = ''; showDeactivate.value = true }
async function confirmDeactivate() {
 if (busy.value || !target.value) return
 busy.value = true
 try { await usersService.delete(target.value.id); showDeactivate.value = false; success.value = 'User deactivated. Their records are retained.'; await load() }
 catch(e) { formError.value = errorMessage(e) } finally { busy.value = false }
}
</script>
<template>
 <section>
  <div class="page-header"><div><h1 class="page-header-title">User Management</h1><p class="page-header-subtitle">Manage sign-in access, roles and team assignments.</p></div><button class="btn btn-primary" @click="edit()">Add user</button></div>
  <p v-if="success" class="form-hint" role="status">{{ success }}</p>
  <p v-if="error" class="form-error" role="alert">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
  <div class="user-filters">
   <label>Search users<input v-model="search" class="input form-input" type="search" placeholder="Name or email" /></label>
   <label>Role<select v-model="role" class="select form-select"><option value="">All roles</option><option v-for="[value,label] in roles" :key="value" :value="value">{{ label }}</option></select></label>
   <label>Status<select v-model="status" class="select form-select"><option value="">All users</option><option value="active">Active</option><option value="inactive">Inactive</option></select></label>
  </div>
  <p v-if="loading" role="status">Loading users…</p>
  <div v-else-if="filtered.length" class="table-container"><table class="table"><thead><tr><th>Name / email</th><th>Role</th><th>Department</th><th>Team</th><th>Status</th><th>Actions</th></tr></thead><tbody>
   <tr v-for="user in filtered" :key="user.id"><td><strong>{{ user.firstName }} {{ user.lastName }}</strong><div>{{ user.email }}</div></td><td>{{ roleLabel(user.role) }}</td><td>{{ user.department }}</td><td>{{ teams.find(t => t.id === user.teamId)?.name ?? 'Unassigned' }}</td><td>{{ user.isActive ? 'Active' : 'Inactive' }}</td><td><div class="user-actions"><button class="btn btn-sm" @click="edit(user)">Edit</button><button v-if="user.id !== auth.user?.id" class="btn btn-sm" @click="openReset(user)">Reset password</button><button v-if="user.isActive && user.id !== auth.user?.id" class="btn btn-sm" @click="deactivate(user)">Deactivate</button></div></td></tr>
  </tbody></table></div>
  <p v-else-if="!error">No users match these filters.</p>
  <AppDialog v-model:open="showEditor" :title="editing ? 'Edit user' : 'Add user'" :busy="busy">
   <form id="user-form" class="user-form" @submit.prevent="save">
    <p v-if="formError" class="form-error wide" role="alert">{{ formError }}</p>
    <label>First name<input v-model.trim="form.firstName" class="input form-input" autocomplete="given-name" required maxlength="100" /></label>
    <label>Last name<input v-model.trim="form.lastName" class="input form-input" autocomplete="family-name" required maxlength="100" /></label>
    <label class="wide">Email<input v-model.trim="form.email" class="input form-input" type="email" autocomplete="off" required maxlength="255" /></label>
    <label>Phone<input v-model.trim="form.phone" class="input form-input" type="tel" maxlength="100" /></label>
    <label>Department<select v-model="form.department" class="select form-select"><option v-for="dept in departments" :key="dept">{{ dept }}</option></select></label>
    <label>Role<select v-model="form.role" class="select form-select" :disabled="lastAdmin"><option v-for="[value,label] in roles" :key="value" :value="value">{{ label }}</option></select></label>
    <label>Team<select v-model="form.teamId" class="select form-select"><option value="">Unassigned</option><option v-for="team in teams" :key="team.id" :value="team.id">{{ team.name }}</option></select></label>
    <template v-if="!editing"><label>Password<input v-model="form.password" class="input form-input" type="password" autocomplete="new-password" required minlength="8" maxlength="72" /></label><label>Confirm password<input v-model="form.confirmation" class="input form-input" type="password" autocomplete="new-password" required /></label></template>
    <label v-else class="wide"><input v-model="form.isActive" type="checkbox" :disabled="lastAdmin || editing.id === auth.user?.id" /> Active account</label>
    <p v-if="lastAdmin" class="form-hint wide">At least one active administrator must remain.</p>
    <p v-if="!editing" class="form-hint wide">Use 8–72 characters. The password is not emailed automatically.</p>
   </form>
   <template #footer><button class="btn" :disabled="busy" @click="showEditor = false">Cancel</button><button class="btn btn-primary" form="user-form" type="submit" :disabled="busy">{{ busy ? 'Saving…' : 'Save user' }}</button></template>
  </AppDialog>
  <AppDialog v-model:open="showReset" :title="`Reset password for ${target?.firstName ?? ''}`" :busy="busy">
   <form id="reset-user-password" class="user-form" @submit.prevent="resetPassword"><p class="wide">{{ target?.email }}. Existing refresh sessions will be revoked.</p><p v-if="formError" class="form-error wide" role="alert">{{ formError }}</p><label>New password<input v-model="reset.password" class="input form-input" type="password" autocomplete="new-password" minlength="8" maxlength="72" required /></label><label>Confirm password<input v-model="reset.confirmation" class="input form-input" type="password" autocomplete="new-password" required /></label></form>
   <template #footer><button class="btn" :disabled="busy" @click="showReset = false">Cancel</button><button class="btn btn-primary" form="reset-user-password" type="submit" :disabled="busy">{{ busy ? 'Resetting…' : 'Reset password' }}</button></template>
  </AppDialog>
  <AppDialog v-model:open="showDeactivate" title="Deactivate user" :busy="busy"><p>{{ target?.email }} will lose access immediately. Their records will be retained.</p><p v-if="formError" class="form-error" role="alert">{{ formError }}</p><template #footer><button class="btn" :disabled="busy" @click="showDeactivate = false">Cancel</button><button class="btn btn-primary" :disabled="busy" @click="confirmDeactivate">Deactivate user</button></template></AppDialog>
 </section>
</template>
<style scoped>
.user-filters, .user-form { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-4); margin-block: var(--space-5); }
.user-filters { grid-template-columns: 2fr 1fr 1fr; }
label { display: grid; gap: var(--space-2); font-size: var(--text-sm); }
.wide { grid-column: 1 / -1; }
.user-actions { display: flex; gap: var(--space-2); flex-wrap: wrap; }
@media (max-width: 700px) { .user-filters, .user-form { grid-template-columns: 1fr; } .page-header { align-items: flex-start; flex-wrap: wrap; gap: var(--space-3); } }
</style>
