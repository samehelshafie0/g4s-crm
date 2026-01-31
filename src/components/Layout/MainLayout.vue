<template>
  <div class="main-layout">
    <Sidebar :isCollapsed="isCollapsed" @toggle="toggleSidebar" />
    <div class="content-wrapper" :style="{ marginLeft:  isCollapsed ? '70px' : '280px' }">
      <header class="top-header">
        <div class="header-left">
          <h1 class="page-title">{{ pageTitle }}</h1>
        </div>
        <div class="header-right">
          <button class="header-btn" @click="toggleTheme" :title="`Switch to ${currentTheme === 'light' ? 'dark' : 'light'} mode`">
            <span>{{ themeIcon }}</span>
          </button>
          <button class="header-btn">
            <span>🔔</span>
            <span class="notification-badge">3</span>
          </button>
          <button class="header-btn">
            <span>⚙️</span>
          </button>
        </div>
      </header>
      <main class="main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './Sidebar.vue'
import { menuItems } from '@/types/menu'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const { currentTheme, toggleTheme } = useTheme()
const isCollapsed = ref(false)

const toggleSidebar = () => {
  isCollapsed.value = !isCollapsed.value
}

const pageTitle = computed(() => {
  const item = menuItems.find(item => item.route === route.path)
  return item?.label || 'Dashboard'
})

const themeIcon = computed(() => {
  return currentTheme.value === 'light' ? '🌙' : '☀️'
})
</script>

<style scoped>
.main-layout {
  display: flex;
  min-height: 100vh;
  background: var(--color-background);
  transition: background-colors 0.3s ease;
}

.content-wrapper {
  flex: 1;
  margin-left: 280px;
  display: flex;
  flex-direction: column;
  transition: margin-left 0.3s ease;
   /* min-width: 0; */
   width: calc(100% - 280px);

}

.top-header {
  background: var(--color-surface);
  padding: 1.25rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: var(--shadow-sm);
  position: sticky;
  top: 0;
  z-index: 100;
  transition: background-color 0.3s ease;
}

.header-left {
  flex: 1;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
  transition: color 0.3s ease;
}

.header-right {
  display: flex;
  gap: 0.75rem;
}

.header-btn {
  position: relative;
  background: var(--color-background);
  border: none;
  width: 42px;
  height: 42px;
  border-radius: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  transition: all 0.2s;
}

.header-btn:hover {
  background: var(--color-surface-hover);
  transform: translateY(-2px);
}

.notification-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: #ef4444;
  color: #fff;
  font-size: 0.7rem;
  padding: 0.2rem 0.4rem;
  border-radius: 10px;
  font-weight: 600;
  min-width: 18px;
}

.main-content {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
  min-width: 0;
}
</style>
