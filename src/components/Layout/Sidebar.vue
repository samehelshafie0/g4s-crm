<template>
  <aside class="sidebar" :class="{ collapsed: props.isCollapsed }">
    <div class="sidebar-header">
      <div class="logo">
        <span class="logo-icon">🏢</span>
        <span v-if="!props.isCollapsed" class="logo-text">CRM Pro</span>
      </div>
      <button class="collapse-btn" @click="onToggleClick">
        {{ props.isCollapsed ? '→' : '←' }}
      </button>
    </div>

    <nav class="sidebar-nav">
      <div v-for="item in menuItems" :key="item.id" class="nav-item">
        <router-link 
          :to="item.route" 
          class="nav-link"
          :class="{ active: isActive(item.route) }"
          :title="item.description"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span v-if="!props.isCollapsed" class="nav-label">{{ item.label }}</span>
          <span v-if="item.badge && !props.isCollapsed" class="nav-badge">{{ item.badge }}</span>
        </router-link>
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="user-info">
        <span class="user-icon">👤</span>
        <span v-if="!props.isCollapsed" class="user-name">Admin User</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { menuItems } from '@/types/menu'

const props = defineProps<{
  isCollapsed: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
}>()

function onToggleClick() {
  emit('toggle')
}

const route = useRoute()


const isActive = (path: string) => {
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<style scoped>
.sidebar {
  width: 280px;
  height: 100vh;
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 1000;
  box-shadow: 2px 0 10px rgba(0, 0, 0, 0.1);
}

.sidebar.collapsed {
  width: 70px;
}

.sidebar-header {
  padding: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 700;
  font-size: 1.25rem;
}

.logo-icon {
  font-size: 1.5rem;
}

.logo-text {
  color: #60a5fa;
}

.collapse-btn {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: #e2e8f0;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 1rem 0;
}

.sidebar-nav::-webkit-scrollbar {
  width: 6px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.nav-item {
  margin: 0.25rem 0.75rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  color: #cbd5e1;
  text-decoration: none;
  border-radius: 8px;
  transition: all 0.2s;
  font-size: 0.9rem;
  position: relative;
}

.nav-link:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  transform: translateX(2px);
}

.nav-link.active {
  background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.nav-icon {
  font-size: 1.25rem;
  min-width: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-badge {
  background: #ef4444;
  color: #fff;
  font-size: 0.7rem;
  padding: 0.2rem 0.5rem;
  border-radius: 12px;
  font-weight: 600;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.user-icon {
  font-size: 1.5rem;
}

.user-name {
  font-size: 0.9rem;
  font-weight: 500;
}

.sidebar.collapsed .nav-label,
.sidebar.collapsed .nav-badge,
.sidebar.collapsed .logo-text,
.sidebar.collapsed .user-name {
  display: none;
}

.sidebar.collapsed .nav-link {
  justify-content: center;
}
</style>
