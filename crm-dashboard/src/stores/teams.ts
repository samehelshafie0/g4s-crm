import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { teamsService } from '@/services'

export const useTeamsStore = defineStore('teams', () => {
  const teams = ref<unknown[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeTeams = computed(() =>
    (teams.value as Array<{ isActive?: boolean }>).filter((t) => t.isActive !== false),
  )
  const totalMembers = computed(() =>
    (teams.value as Array<{ members?: unknown[] }>).reduce(
      (s, t) => s + (t.members?.length ?? 0),
      0,
    ),
  )

  async function fetchTeams(params?: Record<string, unknown>) {
    loading.value = true
    try {
      const res = await teamsService.list(params)
      if (res.success) teams.value = res.data
    } catch (e) { error.value = 'Failed to load teams'; console.error(e) }
    finally { loading.value = false }
  }

  async function addTeam(data: Record<string, unknown>) {
    const res = await teamsService.create(data)
    if (res.success) { teams.value.unshift(res.data as never); return res.data }
    throw new Error('Failed to create team')
  }

  async function updateTeam(id: string, data: Record<string, unknown>) {
    const res = await teamsService.update(id, data)
    if (res.success) {
      const idx = (teams.value as Array<{ id: string }>).findIndex((t) => t.id === id)
      if (idx !== -1) teams.value[idx] = res.data as never
    }
  }

  return { teams, loading, error, activeTeams, totalMembers, fetchTeams, addTeam, updateTeam }
})
