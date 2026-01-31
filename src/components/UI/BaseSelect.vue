<template>
  <div class="select-group" :class="{ error: error }">
    <label v-if="label" class="select-label" :for="selectId">
      {{ label }}
      <span v-if="required" class="required">*</span>
    </label>
    <div class="select-wrapper">
      <select
        :id="selectId"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        @change="handleChange"
        class="select-field"
      >
        <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
        <option
          v-for="option in options"
          :key="getOptionValue(option)"
          :value="getOptionValue(option)"
        >
          {{ getOptionLabel(option) }}
        </option>
      </select>
      <span class="select-icon">▼</span>
    </div>
    <p v-if="error" class="select-error">{{ error }}</p>
    <p v-else-if="hint" class="select-hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue: string | number
  options: any[]
  label?: string
  placeholder?: string
  error?: string
  hint?: string
  disabled?: boolean
  required?: boolean
  valueKey?: string
  labelKey?: string
}

const props = withDefaults(defineProps<Props>(), {
  valueKey: 'value',
  labelKey: 'label'
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const selectId = computed(() => `select-${Math.random().toString(36).substr(2, 9)}`)

const getOptionValue = (option: any) => {
  return typeof option === 'object' ? option[props.valueKey] : option
}

const getOptionLabel = (option: any) => {
  return typeof option === 'object' ? option[props.labelKey] : option
}

const handleChange = (event: Event) => {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', target.value)
}
</script>

<style scoped>
.select-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.select-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.required {
  color: #ef4444;
}

.select-wrapper {
  position: relative;
}

.select-field {
  width: 100%;
  padding: 0.625rem 2.5rem 0.625rem 0.875rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.95rem;
  color: #1f2937;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s;
  appearance: none;
}

.select-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.select-field:disabled {
  background: #f9fafb;
  color: #9ca3af;
  cursor: not-allowed;
}

.select-icon {
  position: absolute;
  right: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: #6b7280;
  font-size: 0.7rem;
}

.select-error {
  font-size: 0.8rem;
  color: #ef4444;
  margin: 0;
}

.select-hint {
  font-size: 0.8rem;
  color: #6b7280;
  margin: 0;
}

.select-group.error .select-field {
  border-color: #ef4444;
}

.select-group.error .select-field:focus {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}
</style>
