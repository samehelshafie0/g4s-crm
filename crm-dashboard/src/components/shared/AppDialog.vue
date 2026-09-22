<script setup lang="ts">
import { ref, watch, nextTick, useId } from 'vue'
import { X } from 'lucide-vue-next'
const props = defineProps<{ open: boolean; title: string; busy?: boolean }>()
const emit = defineEmits<{ 'update:open': [value: boolean] }>()
const dialog = ref<HTMLDialogElement>()
const titleId = useId()
watch(() => props.open, async value => {
  await nextTick()
  if (value && !dialog.value?.open) dialog.value?.showModal()
  if (!value && dialog.value?.open) dialog.value?.close()
}, { immediate: true })
function cancel(event: Event) {
  event.preventDefault()
  if (!props.busy) emit('update:open', false)
}
</script>
<template>
  <Teleport to="body">
    <dialog ref="dialog" class="modal crm-dialog" :aria-labelledby="titleId" @cancel="cancel" @close="emit('update:open', false)">
      <div class="modal-box">
        <header class="modal-header">
          <h2 :id="titleId" class="modal-title">{{ title }}</h2>
          <form method="dialog"><button class="btn btn-ghost btn-sm" aria-label="Close dialog" :disabled="busy"><X :size="18" /></button></form>
        </header>
        <div class="modal-body"><slot /></div>
        <footer v-if="$slots.footer" class="modal-footer"><slot name="footer" /></footer>
      </div>
    </dialog>
  </Teleport>
</template>
<style scoped>
.crm-dialog { margin: auto; padding: 0; border: 1px solid var(--color-neutral-200); color: var(--color-neutral-900); width: min(720px, calc(100vw - 32px)); max-width: 720px; animation: none; }
.crm-dialog:not([open]) { display: none; }
.crm-dialog[open] { display: block; }
.crm-dialog::backdrop { background: rgb(0 0 0 / 45%); }
.modal-box { width: 100%; }
.modal-body { max-height: calc(100dvh - 220px); overflow: auto; }
</style>
