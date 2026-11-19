<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { customersService } from '@/services'
import { errorMessage } from '@/services/payload'
import { useAuthStore } from '@/stores/auth'
import type { CustomerContact, CustomerSite } from '@/types'

const props = defineProps<{ customerId: string; startAdding?: boolean }>()
const emit = defineEmits<{ changed: [contacts: CustomerContact[]] }>()
const auth = useAuthStore()
const canEdit = computed(() => auth.can('customers:update'))
const contacts = ref<CustomerContact[]>([])
const sites = ref<CustomerSite[]>([])
const loading = ref(false)
const busy = ref(false)
const error = ref('')
const editing = ref<string | null>(null)
const showForm = ref(false)
const blank = () => ({ name: '', email: '', phone: '', position: '', siteId: null as string | null, isPrimary: false })
const form = ref(blank())
async function load() {
  loading.value = true; error.value = ''
  try {
    const [people, locations] = await Promise.all([customersService.contacts(props.customerId), customersService.sites(props.customerId)])
    contacts.value = people.data; sites.value = locations.data
    emit('changed', contacts.value)
  } catch (e) { error.value = errorMessage(e) }
  finally { loading.value = false }
}
watch(() => props.customerId, () => { showForm.value = false; void load(); if (props.startAdding && canEdit.value) edit() }, { immediate: true })
function edit(contact?: CustomerContact) {
  editing.value = contact?.id ?? null
  form.value = contact ? { ...blank(), ...contact, siteId: contact.siteId ?? null } : blank()
  showForm.value = true
}
async function save() {
  if (busy.value) return
  busy.value = true; error.value = ''
  try {
    if (editing.value) await customersService.updateContact(props.customerId, editing.value, form.value)
    else await customersService.addContact(props.customerId, form.value)
    showForm.value = false; await load()
  } catch (e) { error.value = errorMessage(e) }
  finally { busy.value = false }
}
async function remove(contact: CustomerContact) {
  if (busy.value || !window.confirm(`Delete contact “${contact.name}”?`)) return
  busy.value = true; error.value = ''
  try { await customersService.deleteContact(props.customerId, contact.id); await load() }
  catch (e) { error.value = errorMessage(e) }
  finally { busy.value = false }
}
</script>

<template>
  <section>
    <p v-if="error" class="form-error" role="alert">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
    <button v-if="canEdit && !showForm" class="btn btn-primary btn-sm" @click="edit()">Add contact</button>
    <form v-if="showForm" class="contact-form" @submit.prevent="save">
      <label>Name<input v-model.trim="form.name" class="input form-input" required maxlength="255" autocomplete="name" /></label>
      <label>Position<input v-model.trim="form.position" class="input form-input" maxlength="255" /></label>
      <label>Email<input v-model.trim="form.email" type="email" class="input form-input" maxlength="255" autocomplete="email" /></label>
      <label>Phone<input v-model.trim="form.phone" type="tel" class="input form-input" maxlength="100" autocomplete="tel" /></label>
      <label>Site (optional)<select v-model="form.siteId" class="select form-select"><option :value="null">Customer-wide</option><option v-for="site in sites" :key="site.id" :value="site.id">{{ site.name }}</option></select></label>
      <label><span><input v-model="form.isPrimary" type="checkbox" /> Primary customer contact</span></label>
      <div class="contact-actions"><button type="button" class="btn btn-sm" :disabled="busy" @click="showForm = false">Cancel</button><button class="btn btn-primary btn-sm" :disabled="busy">{{ busy ? 'Saving…' : 'Save contact' }}</button></div>
    </form>
    <p v-if="loading" role="status">Loading contacts…</p>
    <div v-else-if="contacts.length" class="table-container">
      <table class="table"><thead><tr><th>Name</th><th>Email / phone</th><th>Site</th><th v-if="canEdit">Actions</th></tr></thead><tbody>
        <tr v-for="contact in contacts" :key="contact.id">
          <td><strong>{{ contact.name }}</strong><span v-if="contact.isPrimary" class="badge">Primary</span><div>{{ contact.position }}</div></td>
          <td><div>{{ contact.email || '—' }}</div><div>{{ contact.phone || '—' }}</div></td>
          <td>{{ sites.find(site => site.id === contact.siteId)?.name || 'Customer-wide' }}</td>
          <td v-if="canEdit"><button class="btn btn-ghost btn-sm" :disabled="busy" @click="edit(contact)">Edit</button><button class="btn btn-ghost btn-sm" :disabled="busy" @click="remove(contact)">Delete</button></td>
        </tr>
      </tbody></table>
    </div>
    <p v-else-if="!error">No contacts yet. Add a contact to this customer, with an optional site.</p>
  </section>
</template>

<style scoped>
.contact-form { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: var(--space-4); margin-block: var(--space-4); }
label { display: grid; gap: var(--space-2); }
.contact-actions { grid-column: 1 / -1; display: flex; justify-content: flex-end; gap: var(--space-2); }
.table-container { margin-top: var(--space-4); }
@media (max-width: 600px) { .contact-form { grid-template-columns: 1fr; } }
</style>
