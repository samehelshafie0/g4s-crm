<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { usersService } from '@/services'
import { errorMessage } from '@/services/payload'
import { passwordError, roleLabel } from '@/services/userOptions'
const auth = useAuthStore()
const form = ref({firstName:auth.user?.firstName ?? '', lastName:auth.user?.lastName ?? '', email:auth.user?.email ?? '', phone:auth.user?.phone ?? ''})
const password = ref({current:'', next:'', confirm:''})
const saving = ref(false)
const changing = ref(false)
const error = ref('')
const passwordIssue = ref('')
const success = ref('')
async function save() {
 if (saving.value) return
 saving.value = true; error.value = ''; success.value = ''
 try { auth.user = (await usersService.updateProfile(form.value)).data; success.value = 'Profile updated.' }
 catch(e) { error.value = errorMessage(e) } finally { saving.value = false }
}
async function changePassword() {
 if (changing.value) return
 passwordIssue.value = passwordError(password.value.next, password.value.confirm)
 if (passwordIssue.value) return
 changing.value = true
 try { await auth.changePassword(password.value.current, password.value.next); password.value = {current:'',next:'',confirm:''}; await auth.logout() }
 catch(e) { passwordIssue.value = errorMessage(e) } finally { changing.value = false }
}
</script>
<template>
 <section class="profile-page">
  <h1 class="page-header-title">My Profile</h1><p class="page-header-subtitle">{{ roleLabel(auth.userRole) }} · {{ auth.user?.department }}</p>
  <form class="profile-section" @submit.prevent="save"><h2>Personal details</h2><p v-if="error" class="form-error" role="alert">{{ error }}</p><p v-if="success" role="status">{{ success }}</p><div class="profile-fields"><label>First name<input v-model.trim="form.firstName" class="input form-input" required maxlength="100" autocomplete="given-name" /></label><label>Last name<input v-model.trim="form.lastName" class="input form-input" required maxlength="100" autocomplete="family-name" /></label><label>Email<input v-model.trim="form.email" class="input form-input" type="email" required maxlength="255" autocomplete="email" /></label><label>Phone<input v-model.trim="form.phone" class="input form-input" type="tel" maxlength="100" autocomplete="tel" /></label></div><button class="btn btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save profile' }}</button></form>
  <form class="profile-section" @submit.prevent="changePassword"><h2>Change password</h2><p class="form-hint">Use 8–72 characters. You will sign in again after changing your password.</p><p v-if="passwordIssue" class="form-error" role="alert">{{ passwordIssue }}</p><div class="profile-fields"><label>Current password<input v-model="password.current" class="input form-input" type="password" autocomplete="current-password" required /></label><span /><label>New password<input v-model="password.next" class="input form-input" type="password" minlength="8" maxlength="72" autocomplete="new-password" required /></label><label>Confirm new password<input v-model="password.confirm" class="input form-input" type="password" autocomplete="new-password" required /></label></div><button class="btn btn-primary" :disabled="changing">{{ changing ? 'Changing…' : 'Change password' }}</button></form>
 </section>
</template>
<style scoped>
.profile-page { max-width: 860px; }
.profile-section { padding-block: var(--space-6); border-bottom: 1px solid var(--color-neutral-200); }
h2 { font-size: var(--text-lg); font-weight: 600; margin-bottom: var(--space-3); }
.profile-fields { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-4); margin-block: var(--space-5); }
label { display: grid; gap: var(--space-2); font-size: var(--text-sm); }
@media (max-width: 650px) { .profile-fields { grid-template-columns: 1fr; } .profile-fields > span { display: none; } }
</style>
