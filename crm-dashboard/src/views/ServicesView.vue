<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppDialog from '@/components/shared/AppDialog.vue'
import { catalogServicesService } from '@/services'
import type { ServiceItem } from '@/services/workflowDtos'
import { allPages } from '@/services/collections'
import { errorMessage } from '@/services/payload'
import { departments } from '@/services/userOptions'
import { useAuthStore } from '@/stores/auth'
const auth = useAuthStore()
const services = ref<ServiceItem[]>([])
const loading = ref(true)
const busy = ref(false)
const error = ref('')
const formError = ref('')
const success = ref('')
const search = ref('')
const status = ref('')
const showEditor = ref(false)
const showDelete = ref(false)
const target = ref<ServiceItem | null>(null)
const blank = (): Omit<ServiceItem,'id'> => ({sku:'',name:'',description:'',department:'technical',rateType:'hour',unitCost:0,unitPrice:0,isActive:true})
const form = ref(blank())
const filtered = computed(() => services.value.filter(s => `${s.sku} ${s.name}`.toLowerCase().includes(search.value.trim().toLowerCase()) && (!status.value || s.isActive === (status.value === 'active'))))
const money = (value: number) => new Intl.NumberFormat('en-SA',{style:'currency',currency:'SAR'}).format(value)
async function load() { loading.value = true; error.value = ''; try { services.value = await allPages(catalogServicesService.list) } catch(e) { error.value = errorMessage(e) } finally { loading.value = false } }
onMounted(load)
function edit(item?: ServiceItem) { target.value = item ?? null; form.value = item ? {sku:item.sku,name:item.name,description:item.description,department:item.department,rateType:item.rateType,unitCost:item.unitCost,unitPrice:item.unitPrice,isActive:item.isActive} : blank(); formError.value = ''; showEditor.value = true }
async function save() {
 if (busy.value) return
 busy.value = true; formError.value = ''
 try { if (target.value) await catalogServicesService.update(target.value.id,form.value); else await catalogServicesService.create(form.value); showEditor.value = false; success.value = 'Service saved. Active services are available in quote builders and price books.'; await load() }
 catch(e) { formError.value = errorMessage(e) } finally { busy.value = false }
}
function remove(item: ServiceItem) { target.value = item; formError.value = ''; showDelete.value = true }
async function confirmDelete() { if (busy.value || !target.value) return; busy.value = true; try { await catalogServicesService.delete(target.value.id); showDelete.value = false; success.value = 'Service deleted.'; await load() } catch(e) { formError.value = errorMessage(e) } finally { busy.value = false } }
</script>
<template>
 <section>
  <div class="page-header"><div><h1 class="page-header-title">Service Catalog</h1><p class="page-header-subtitle">One-time labor and service rates for quotes and price books. Prices are in SAR.</p></div><button v-if="auth.can('products:create')" class="btn btn-primary" @click="edit()">Add catalog service</button></div>
  <p v-if="success" role="status">{{ success }}</p><p v-if="error" class="form-error" role="alert">{{ error }} <button class="btn btn-sm" @click="load">Retry</button></p>
  <div class="service-filters"><label>Search services<input v-model="search" class="input form-input" type="search" placeholder="SKU or service name" /></label><label>Status<select v-model="status" class="select form-select"><option value="">All services</option><option value="active">Active</option><option value="inactive">Inactive</option></select></label></div>
  <p v-if="loading" role="status">Loading catalog…</p>
  <div v-else-if="filtered.length" class="table-container"><table class="table"><thead><tr><th>SKU / service</th><th>Department</th><th>Rate unit</th><th>Unit cost</th><th>Selling price</th><th>Status</th><th>Actions</th></tr></thead><tbody><tr v-for="item in filtered" :key="item.id"><td><strong>{{ item.name }}</strong><div>{{ item.sku }}</div></td><td>{{ item.department }}</td><td>{{ item.rateType }}</td><td>{{ money(item.unitCost) }}</td><td>{{ money(item.unitPrice) }}</td><td>{{ item.isActive ? 'Active' : 'Inactive' }}</td><td><div class="table-actions"><button v-if="auth.can('products:update')" class="btn btn-sm" @click="edit(item)">Edit</button><button v-if="auth.can('products:delete')" class="btn btn-sm" @click="remove(item)">Delete</button></div></td></tr></tbody></table></div>
  <p v-else-if="!error">No catalog services match these filters. Add a service to make it available for quotation.</p>
  <AppDialog v-model:open="showEditor" :title="target ? 'Edit catalog service' : 'Add catalog service'" :busy="busy"><form id="catalog-service-form" class="service-form" @submit.prevent="save"><p v-if="formError" class="form-error wide" role="alert">{{ formError }}</p><label>SKU<input v-model.trim="form.sku" class="input form-input" required maxlength="100" /></label><label>Name<input v-model.trim="form.name" class="input form-input" required maxlength="200" /></label><label>Department<select v-model="form.department" class="select form-select"><option v-for="dept in departments" :key="dept">{{ dept }}</option></select></label><label>Rate unit<input v-model.trim="form.rateType" class="input form-input" required maxlength="50" list="service-rate-units" /><datalist id="service-rate-units"><option value="hour" /><option value="day" /><option value="visit" /><option value="job" /></datalist></label><label>Unit cost (SAR)<input v-model.number="form.unitCost" class="input form-input" type="number" min="0" max="1000000000" step="0.01" required /></label><label>Selling price (SAR)<input v-model.number="form.unitPrice" class="input form-input" type="number" min="0" max="1000000000" step="0.01" required /></label><label class="wide">Description<textarea v-model="form.description" class="form-textarea" maxlength="5000" rows="3" /></label><label class="wide"><span><input v-model="form.isActive" type="checkbox" /> Available for new quotes and price books</span></label></form><template #footer><button class="btn" :disabled="busy" @click="showEditor = false">Cancel</button><button class="btn btn-primary" form="catalog-service-form" type="submit" :disabled="busy">{{ busy ? 'Saving…' : 'Save service' }}</button></template></AppDialog>
  <AppDialog v-model:open="showDelete" title="Delete catalog service" :busy="busy"><p>Delete {{ target?.name }}? Existing quote line snapshots will be retained. To stop using a service without deleting it, edit it and turn off availability.</p><p v-if="formError" class="form-error" role="alert">{{ formError }}</p><template #footer><button class="btn" :disabled="busy" @click="showDelete = false">Cancel</button><button class="btn btn-primary" :disabled="busy" @click="confirmDelete">Delete service</button></template></AppDialog>
 </section>
</template>
<style scoped>
.service-form, .service-filters { display: grid; grid-template-columns: 1fr 1fr; gap: var(--space-4); margin-block: var(--space-5); }
.service-filters { grid-template-columns: 2fr 1fr; }
label { display: grid; gap: var(--space-2); font-size: var(--text-sm); }
.wide { grid-column: 1 / -1; }
@media (max-width: 700px) { .service-form, .service-filters { grid-template-columns: 1fr; } .page-header { flex-wrap: wrap; gap: var(--space-3); } }
</style>
