<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  LayoutDashboard,
  Users,
  Target,
  Factory,
  Package,
  Warehouse,
  FileText,
  BookOpen,
  ArrowLeftRight,
  ScrollText,
  RefreshCw,
  UserCog,
  FolderOpen,
  ChevronLeft,
  ChevronRight,
  Shield,
  ShoppingCart,
  FolderKanban,
} from 'lucide-vue-next'

defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  toggle: []
}>()

const route = useRoute()
const router = useRouter()

const navGroups = computed(() => [
  {
    title: 'Overview',
    items: [
      { name: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
    ],
  },
  {
    title: 'Sales',
    items: [
      { name: 'Customers', path: '/customers', icon: Users },
      { name: 'Opportunities', path: '/leads', icon: Target },
      { name: 'Quotations', path: '/quotes', icon: FileText },
      { name: 'Contracts', path: '/contracts', icon: ScrollText },
      { name: 'Projects', path: '/projects', icon: FolderKanban },
    ],
  },
  {
    title: 'Catalog',
    items: [
      { name: 'Manufacturers', path: '/manufacturers', icon: Factory },
      { name: 'Products', path: '/products', icon: Package },
      { name: 'Price Books', path: '/price-books', icon: BookOpen },
      { name: 'Inventory', path: '/inventory', icon: Warehouse },
      { name: 'Procurement', path: '/procurement', icon: ShoppingCart },
    ],
  },
  {
    title: 'Operations',
    items: [
      { name: 'Recurring Services', path: '/recurring-services', icon: RefreshCw },
      { name: 'Exchange Rates', path: '/exchange-rates', icon: ArrowLeftRight },
      { name: 'Teams', path: '/teams', icon: UserCog },
      { name: 'Documents', path: '/documents', icon: FolderOpen },
    ],
  },
])

function isActive(path: string): boolean {
  return route.path === path || route.path.startsWith(path + '/')
}

function navigate(path: string) {
  router.push(path)
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed }">
    <div class="sidebar-brand" @click="navigate('/dashboard')">
      <div class="brand-icon">
        <Shield :size="24" />
      </div>
      <transition name="fade">
        <div v-if="!collapsed" class="brand-text">
          <span class="brand-name">G4S</span>
          <span class="brand-label">CRM</span>
        </div>
      </transition>
    </div>

    <nav class="sidebar-nav">
      <div v-for="group in navGroups" :key="group.title" class="nav-group">
        <transition name="fade">
          <div v-if="!collapsed" class="nav-group-title">{{ group.title }}</div>
        </transition>
        <button
          v-for="item in group.items"
          :key="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          :title="collapsed ? item.name : undefined"
          @click="navigate(item.path)"
        >
          <component :is="item.icon" :size="20" class="nav-icon" />
          <transition name="fade">
            <span v-if="!collapsed" class="nav-label">{{ item.name }}</span>
          </transition>
        </button>
      </div>
    </nav>

    <button class="sidebar-toggle" @click="emit('toggle')">
      <ChevronLeft v-if="!collapsed" :size="18" />
      <ChevronRight v-else :size="18" />
    </button>
  </aside>
</template>

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: var(--sidebar-width);
  background: var(--sidebar-bg);
  display: flex;
  flex-direction: column;
  z-index: 100;
  transition: width var(--transition-slow);
  overflow: hidden;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-5) var(--space-5);
  cursor: pointer;
  border-bottom: 1px solid var(--sidebar-border);
  flex-shrink: 0;
  min-height: 64px;
}

.brand-icon {
  width: 36px;
  height: 36px;
  background: var(--color-primary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.brand-text {
  display: flex;
  flex-direction: column;
  white-space: nowrap;
}

.brand-name {
  font-size: var(--text-xl);
  font-weight: 700;
  color: var(--sidebar-text-active);
  line-height: 1.2;
}

.brand-label {
  font-size: var(--text-xs);
  color: var(--sidebar-text);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-3) var(--space-3);
}

.nav-group {
  margin-bottom: var(--space-4);
}

.nav-group-title {
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--sidebar-text);
  padding: var(--space-2) var(--space-3);
  white-space: nowrap;
  opacity: 0.6;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  color: var(--sidebar-text);
  font-size: var(--text-sm);
  font-weight: 400;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
  text-align: left;
}

.nav-item:hover {
  background: var(--sidebar-hover-bg);
  color: var(--sidebar-text-active);
}

.nav-item.active {
  background: var(--sidebar-active-bg);
  color: var(--color-primary);
  font-weight: 500;
}

.nav-icon {
  flex-shrink: 0;
}

.nav-label {
  white-space: nowrap;
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: var(--space-4);
  border: none;
  background: transparent;
  color: var(--sidebar-text);
  cursor: pointer;
  border-top: 1px solid var(--sidebar-border);
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.sidebar-toggle:hover {
  color: var(--sidebar-text-active);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
