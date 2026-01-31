<template>
  <span class="base-badge" :class="[variantClass, sizeClass]">
    <slot></slot>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info'
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'md'
})

const variantClass = computed(() => `badge-${props.variant}`)
const sizeClass = computed(() => `badge-${props.size}`)
</script>

<style scoped>
.base-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 500;
  border-radius: 12px;
  white-space: nowrap;
}

/* Sizes */
.badge-sm {
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
}

.badge-md {
  padding: 0.25rem 0.75rem;
  font-size: 0.8rem;
}

.badge-lg {
  padding: 0.375rem 1rem;
  font-size: 0.875rem;
}

/* Variants */
.badge-default {
  background: #f1f5f9;
  color: #475569;
}

.badge-primary {
  background: #dbeafe;
  color: #1e40af;
}

.badge-success {
  background: #dcfce7;
  color: #166534;
}

.badge-warning {
  background: #fef3c7;
  color: #92400e;
}

.badge-danger {
  background: #fee2e2;
  color: #991b1b;
}

.badge-info {
  background: #e0f2fe;
  color: #075985;
}
</style>
