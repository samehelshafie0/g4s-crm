import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dashboardService } from '@/services'

export const useDashboardStore = defineStore('dashboard', () => {
  const kpis = ref<Record<string, unknown>>({})
  const pipeline = ref<unknown[]>([])
  const recentActivity = ref<unknown[]>([])
  const topCustomers = ref<unknown[]>([])
  const alerts = ref<unknown[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchAll() {
    loading.value = true
    error.value = null
    try {
      const [kpisRes, pipelineRes, activityRes, topRes, alertsRes] = await Promise.all([
        dashboardService.kpis(),
        dashboardService.pipeline(),
        dashboardService.recentActivity(),
        dashboardService.topCustomers(),
        dashboardService.alerts(),
      ])
      if (kpisRes.success) kpis.value = kpisRes.data as Record<string, unknown>
      if (pipelineRes.success) pipeline.value = (pipelineRes.data as { stages: unknown[] }).stages ?? []
      if (activityRes.success) recentActivity.value = activityRes.data
      if (topRes.success) topCustomers.value = topRes.data
      if (alertsRes.success) alerts.value = alertsRes.data
    } catch (e) {
      error.value = 'Failed to load dashboard data'
      console.error(e)
    } finally {
      loading.value = false
    }
  }

  return { kpis, pipeline, recentActivity, topCustomers, alerts, loading, error, fetchAll }
})
