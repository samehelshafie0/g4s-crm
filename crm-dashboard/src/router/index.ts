import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { tokenStorage } from '@/services/http'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // ─── Public ──────────────────────────────────────────────
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { title: 'Sign In', public: true },
    },

    // ─── App root redirect ────────────────────────────────────
    {
      path: '/',
      redirect: '/dashboard',
    },

    // ─── Protected ───────────────────────────────────────────
    {
      path: '/dashboard',
      name: 'dashboard',
      component: () => import('@/views/DashboardView.vue'),
      meta: { title: 'Dashboard' },
    },
    {
      path: '/customers',
      name: 'customers',
      component: () => import('@/views/CustomersView.vue'),
      meta: { title: 'Customers' },
    },
    {
      path: '/leads',
      name: 'leads',
      component: () => import('@/views/LeadsView.vue'),
      meta: { title: 'Opportunities & Leads' },
    },
    {
      path: '/manufacturers',
      name: 'manufacturers',
      component: () => import('@/views/ManufacturersView.vue'),
      meta: { title: 'Manufacturers' },
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('@/views/ProductsView.vue'),
      meta: { title: 'Products' },
    },
    {
      path: '/inventory',
      name: 'inventory',
      component: () => import('@/views/InventoryView.vue'),
      meta: { title: 'Inventory' },
    },
    {
      path: '/quotes',
      name: 'quotes',
      component: () => import('@/views/QuotesView.vue'),
      meta: { title: 'Quotations' },
    },
    {
      path: '/quotes/:id/builder',
      name: 'quote-builder',
      component: () => import('@/views/QuoteBuilderView.vue'),
      meta: { title: 'Quote Builder' },
    },
    {
      path: '/price-books',
      name: 'price-books',
      component: () => import('@/views/PriceBooksView.vue'),
      meta: { title: 'Price Books' },
    },
    {
      path: '/exchange-rates',
      name: 'exchange-rates',
      component: () => import('@/views/ExchangeRatesView.vue'),
      meta: { title: 'Exchange Rates' },
    },
    {
      path: '/contracts',
      name: 'contracts',
      component: () => import('@/views/ContractsView.vue'),
      meta: { title: 'Contracts' },
    },
    {
      path: '/recurring-services',
      name: 'recurring-services',
      component: () => import('@/views/RecurringServicesView.vue'),
      meta: { title: 'Recurring Services' },
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('@/views/ProjectsView.vue'),
      meta: { title: 'Projects' },
    },
    {
      path: '/procurement',
      name: 'procurement',
      component: () => import('@/views/ProcurementView.vue'),
      meta: { title: 'Procurement' },
    },
    {
      path: '/teams',
      name: 'teams',
      component: () => import('@/views/TeamsView.vue'),
      meta: { title: 'Teams' },
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('@/views/DocumentsView.vue'),
      meta: { title: 'Documents' },
    },

    { path: '/users', name: 'users', component: () => import('@/views/UsersView.vue'), meta: { title: 'User Management', permission: 'users:update' } },
    { path: '/profile', name: 'profile', component: () => import('@/views/ProfileView.vue'), meta: { title: 'My Profile' } },
    { path: '/services', name: 'services', component: () => import('@/views/ServicesView.vue'), meta: { title: 'Service Catalog', permission: 'products:read' } },

    // ─── Catch-all ────────────────────────────────────────────
    {
      path: '/:pathMatch(.*)*',
      redirect: '/dashboard',
    },
  ],
})

// ─── Auth Guard ──────────────────────────────────────────────
router.beforeEach(async (to) => {
  document.title = `${to.meta.title || 'CRM'} | G4S CRM`

  const isPublic = to.meta.public === true
  const hasToken = !!tokenStorage.getAccess()

  if (!isPublic && !hasToken) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  const auth = useAuthStore()
  if (!isPublic && hasToken && !auth.user) await auth.fetchMe()
  if (!isPublic && !auth.user) return { name: 'login', query: { redirect: to.fullPath } }
  const resource = to.path.split('/')[1]
  const permission = to.meta.permission as string | undefined ?? (resource === 'profile' ? undefined : `${resource === 'leads' ? 'opportunities' : resource}:read`)
  if (!isPublic && permission && !auth.can(permission)) return { name: 'dashboard' }

  if (to.name === 'login' && hasToken) {
    return { name: 'dashboard' }
  }

  return true
})

export default router
