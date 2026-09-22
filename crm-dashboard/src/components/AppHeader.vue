<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Menu,
  Search,
  Bell,
  Settings,
  User,
  ChevronDown,
  Sun,
  Moon,
  Monitor,
  LogOut,
} from 'lucide-vue-next'
import { useTheme } from '@/composables/useTheme'
import { useAuthStore } from '@/stores/auth'

defineProps<{
  title: string
  sidebarCollapsed: boolean
}>()

defineEmits<{
  toggleSidebar: []
}>()

const { mode, cycleMode } = useTheme()
const authStore = useAuthStore()
const router = useRouter()

const showProfileMenu = ref(false)
const searchQuery = ref('')

async function handleLogout() {
  showProfileMenu.value = false
  await authStore.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <header class="app-header">
    <div class="header-left">
      <button class="btn btn-ghost btn-icon" aria-label="Toggle navigation" @click="$emit('toggleSidebar')">
        <Menu :size="20" />
      </button>
      <div class="header-title-section">
        <h1 class="header-title">{{ title }}</h1>
      </div>
    </div>

    <div class="header-center">
      <div class="header-search">
        <Search :size="16" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search customers, quotes, products..."
          class="search-input"
        />
        <kbd class="search-kbd">⌘K</kbd>
      </div>
    </div>

    <div class="header-right">
      <button class="btn btn-ghost btn-icon theme-toggle" :title="'Theme: ' + mode" @click="cycleMode">
        <Sun v-if="mode === 'light'" :size="20" />
        <Moon v-else-if="mode === 'dark'" :size="20" />
        <Monitor v-else :size="20" />
      </button>

      <button class="btn btn-ghost btn-icon notification-btn">
        <Bell :size="20" />
        <span class="notification-dot" />
      </button>

      <button class="btn btn-ghost btn-icon" aria-label="My profile settings" @click="router.push('/profile')">
        <Settings :size="20" />
      </button>

      <div class="profile-menu" role="button" tabindex="0" aria-label="Account menu" :aria-expanded="showProfileMenu" @keydown.enter.self="showProfileMenu = !showProfileMenu" @keydown.space.prevent.self="showProfileMenu = !showProfileMenu" @keydown.escape="showProfileMenu = false" @click="showProfileMenu = !showProfileMenu">
        <div class="profile-avatar">
          <span v-if="authStore.userInitials">{{ authStore.userInitials }}</span>
          <User v-else :size="18" />
        </div>
        <div class="profile-info">
          <span class="profile-name">{{ authStore.userFullName || 'User' }}</span>
          <span class="profile-role">{{ authStore.userRole || '—' }}</span>
        </div>
        <ChevronDown :size="14" class="text-muted" />

        <div v-if="showProfileMenu" class="profile-dropdown" @click.stop>
          <button class="dropdown-item" @click="showProfileMenu = false; router.push('/profile')">My Profile</button>

          <div class="dropdown-divider" />
          <button class="dropdown-item text-danger" @click="handleLogout">
            <LogOut :size="14" style="margin-right: 6px; vertical-align: middle" />
            Sign Out
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  height: var(--header-height);
  background: var(--content-surface);
  border-bottom: 1px solid var(--color-neutral-200);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-6);
  gap: var(--space-6);
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 50;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-shrink: 0;
}

.header-title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-neutral-900);
}

.header-center {
  flex: 1;
  max-width: 480px;
}

.header-search {
  position: relative;
  display: flex;
  align-items: center;
}

.header-search .search-icon {
  position: absolute;
  left: 12px;
  color: var(--color-neutral-400);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 38px;
  padding: 0 44px 0 36px;
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  background: var(--color-neutral-50);
  color: var(--color-neutral-800);
  outline: none;
  transition: all var(--transition-fast);
}

.search-input:focus {
  background: var(--content-surface);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-200);
}

.search-input::placeholder {
  color: var(--color-neutral-400);
}

.search-kbd {
  position: absolute;
  right: 8px;
  padding: 2px 6px;
  font-size: 0.625rem;
  font-family: inherit;
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: 4px;
  color: var(--color-neutral-400);
  pointer-events: none;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
}

.theme-toggle {
  position: relative;
}

.notification-btn {
  position: relative;
}

.notification-dot {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  background: var(--color-danger);
  border-radius: 50%;
  border: 2px solid var(--content-surface);
}

.profile-menu {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background var(--transition-fast);
  position: relative;
  margin-left: var(--space-2);
}

.profile-menu:hover {
  background: var(--color-neutral-100);
}

.profile-avatar {
  width: 32px;
  height: 32px;
  background: var(--color-primary-100);
  color: var(--color-primary);
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-info {
  display: flex;
  flex-direction: column;
}

.profile-name {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-neutral-800);
  line-height: 1.2;
}

.profile-role {
  font-size: var(--text-xs);
  color: var(--color-neutral-500);
}

.profile-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 180px;
  background: var(--content-surface);
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  padding: var(--space-1);
  z-index: 200;
  animation: slideUp var(--transition-fast);
}

.dropdown-item {
  display: block;
  width: 100%;
  padding: 8px 12px;
  font-size: var(--text-sm);
  text-align: left;
  background: none;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: var(--color-neutral-800);
  transition: background var(--transition-fast);
}

.dropdown-item:hover {
  background: var(--color-neutral-100);
}

.dropdown-divider {
  height: 1px;
  background: var(--color-neutral-200);
  margin: var(--space-1) 0;
}
@media (max-width: 768px) {
  .app-header { padding-inline: var(--space-3); gap: var(--space-2); }
  .header-left { flex: 1; min-width: 0; gap: var(--space-2); }
  .header-title-section { min-width: 0; }
  .header-title { font-size: var(--text-base); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .header-center, .profile-info, .notification-btn, [aria-label="My profile settings"] { display: none; }
  .profile-menu { padding-inline: var(--space-1); }
  .header-right { gap: 0; }
}
</style>
