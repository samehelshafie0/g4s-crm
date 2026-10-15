<script setup lang="ts">
// Set VITE_ENVIRONMENT_LABEL at build time (ENVIRONMENT_LABEL in .env for Docker)
// so testers can never mistake a test server for the live system.
const label = (import.meta.env.VITE_ENVIRONMENT_LABEL ?? '').trim()
</script>

<template>
  <div v-if="label" class="env-banner" role="status">{{ label }}</div>
</template>

<style scoped>
.env-banner {
  position: fixed;
  bottom: 12px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  padding: 4px 14px;
  border-radius: 999px;
  background: #b45309;
  color: #fff;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  box-shadow: 0 2px 8px rgb(0 0 0 / 25%);
  pointer-events: none;
}
</style>
