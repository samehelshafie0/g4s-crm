<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
import { useRoute } from 'vue-router'
import AppSidebar from '@/components/AppSidebar.vue'
import AppHeader from '@/components/AppHeader.vue'

const mobile = useMediaQuery('(max-width: 768px)')
const sidebarCollapsed = ref(mobile.value)
const route = useRoute()

watch(() => route.path, () => { if (mobile.value) sidebarCollapsed.value = true })
watch(mobile, value => { sidebarCollapsed.value = value })
const pageTitle = computed(() => (route.meta.title as string) || 'Dashboard')

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}
</script>

<template>
  <div class="app-layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <button v-if="mobile && !sidebarCollapsed" class="nav-scrim" aria-label="Close navigation" @click="sidebarCollapsed = true" />
    <AppSidebar :collapsed="sidebarCollapsed" @toggle="toggleSidebar" />
    <div class="app-main">
      <AppHeader :title="pageTitle" :sidebar-collapsed="sidebarCollapsed" @toggle-sidebar="toggleSidebar" />
      <main class="app-content">
        <router-view v-slot="{ Component }">
          <transition name="page" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.app-layout {
  width: 100%;
  min-width: 0;
  display: flex;
  min-height: 100vh;
}

.app-main {
  flex: 1;
  margin-left: var(--sidebar-width);
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin-left var(--transition-normal);
}

.sidebar-collapsed .app-main {
  margin-left: var(--sidebar-collapsed-width);
}

.app-content {
  flex: 1;
  padding: var(--space-6) var(--space-8);
  overflow-y: auto;
}

.page-enter-active,
.page-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(4px);
}

.page-leave-to {
  opacity: 0;
}
.nav-scrim { position: fixed; inset: 0; background: rgb(0 0 0 / 40%); border: 0; z-index: 99; }
@media (max-width: 768px) {
  .app-main, .sidebar-collapsed .app-main { margin-left: 0; }
  .app-content { padding: var(--space-4); }
  :deep(.sidebar) { width: min(280px, 85vw); transition: transform var(--transition-normal); }
  :deep(.sidebar.collapsed) { width: min(280px, 85vw); transform: translateX(-100%); visibility: hidden; }
}
</style>
