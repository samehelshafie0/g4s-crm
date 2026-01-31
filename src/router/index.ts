import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
    },
    {
      path: '/customers',
      name: 'customers',
      component: () => import('../views/CustomersView.vue'),
    },
    {
      path: '/leads',
      name: 'leads',
      component: () => import('../views/LeadsView.vue'),
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('../views/ProductsView.vue'),
    },
    {
      path: '/manufacturers',
      name: 'manufacturers',
      component: () => import('../views/ManufacturersView.vue'),
    },
    {
      path: '/pricebooks',
      name: 'pricebooks',
      component: () => import('../views/PriceBooksView.vue'),
    },
    {
      path: '/exchange-rates',
      name: 'exchange-rates',
      component: () => import('../views/ExchangeRatesView.vue'),
    },
    {
      path: '/inventory',
      name: 'inventory',
      component: () => import('../views/InventoryView.vue'),
    },
    {
      path: '/procurement',
      name: 'procurement',
      component: () => import('../views/ProcurementView.vue'),
    },
    {
      path: '/manpower',
      name: 'manpower',
      component: () => import('../views/ManpowerView.vue'),
    },
    {
      path: '/teams',
      name: 'teams',
      component: () => import('../views/TeamsView.vue'),
    },
    {
      path: '/recurring',
      name: 'recurring',
      component: () => import('../views/RecurringServicesView.vue'),
    },
    {
      path: '/quoting',
      name: 'quoting',
      component: () => import('../views/QuotingView.vue'),
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('../views/ProjectsView.vue'),
    },
    {
      path: '/documents',
      name: 'documents',
      component: () => import('../views/DocumentsView.vue'),
    },
    {
      path: '/workflows',
      name: 'workflows',
      component: () => import('../views/WorkflowsView.vue'),
    },
    {
      path: '/security',
      name: 'security',
      component: () => import('../views/SecurityView.vue'),
    },
  ],
})

export default router
