<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView } from 'vue-router'
import MainLayout from './components/Layout/MainLayout.vue'
import BaseToast from './components/UI/BaseToast.vue'
import { useToast } from './composables/useToast'
import { useTheme } from './composables/useTheme'

const toastRef = ref()
const { setToastInstance } = useToast()
const { initTheme, watchSystemTheme } = useTheme()

onMounted(() => {
  setToastInstance(toastRef.value)
  initTheme()
  watchSystemTheme()
})
</script>

<template>
  <MainLayout>
    <RouterView />
  </MainLayout>
  <BaseToast ref="toastRef" />
</template>

<style>
/* CSS Variables for theming */
:root {
  /* Light theme colors */
  --color-background: #f1f5f9;
  --color-surface: #ffffff;
  --color-surface-hover: #f8fafc;
  --color-border: #e2e8f0;
  --color-border-light: #f1f5f9;

  --color-text-primary: #1e293b;
  --color-text-secondary: #64748b;
  --color-text-tertiary: #94a3b8;

  --color-primary: #3b82f6;
  --color-primary-hover: #2563eb;
  --color-primary-light: #dbeafe;

  --color-success: #10b981;
  --color-success-light: #dcfce7;
  --color-warning: #f59e0b;
  --color-warning-light: #fef3c7;
  --color-danger: #ef4444;
  --color-danger-light: #fee2e2;
  --color-info: #3b82f6;
  --color-info-light: #e0f2fe;

  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.15);
  --shadow-lg: 0 10px 25px rgba(0, 0, 0, 0.2);
}

[data-theme='dark'] {
  /* Dark theme colors */
  --color-background: #0f172a;
  --color-surface: #1e293b;
  --color-surface-hover: #334155;
  --color-border: #334155;
  --color-border-light: #475569;

  --color-text-primary: #f1f5f9;
  --color-text-secondary: #cbd5e1;
  --color-text-tertiary: #94a3b8;

  --color-primary: #60a5fa;
  --color-primary-hover: #3b82f6;
  --color-primary-light: #1e3a8a;

  --color-success: #34d399;
  --color-success-light: #064e3b;
  --color-warning: #fbbf24;
  --color-warning-light: #78350f;
  --color-danger: #f87171;
  --color-danger-light: #7f1d1d;
  --color-info: #60a5fa;
  --color-info-light: #1e3a8a;

  --shadow-sm: 0 1px 3px rgba(0, 0, 0, 0.3);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.4);
  --shadow-lg: 0 10px 25px rgba(0, 0, 0, 0.5);
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: var(--color-background);
  color: var(--color-text-primary);
  transition: background-color 0.3s ease, color 0.3s ease;
}

#app {
  width: 100%;
  min-height: 100vh;
}
</style>
