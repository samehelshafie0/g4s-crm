import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, tokenStorage } from '@/services'
import type { AuthUser } from '@/services/auth.service'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value && !!tokenStorage.getAccess())
  const userFullName = computed(() =>
    user.value ? `${user.value.firstName} ${user.value.lastName}` : '',
  )
  const userRole = computed(() => user.value?.role ?? '')
  const userInitials = computed(() => {
    if (!user.value) return ''
    return `${user.value.firstName?.[0] ?? ''}${user.value.lastName?.[0] ?? ''}`.toUpperCase()
  })

  async function login(email: string, password: string) {
    loading.value = true
    error.value = null
    try {
      const res = await authService.login({ email, password })
      if (res.success && res.data) {
        tokenStorage.setTokens(res.data.accessToken, res.data.refreshToken)
        user.value = res.data.user
      }
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: { message?: string } } } })?.response?.data
        ?.error?.message
      error.value = msg ?? 'Login failed. Please check your credentials.'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    const refreshToken = tokenStorage.getRefresh()
    const accessToken = tokenStorage.getAccess()
    tokenStorage.clear()
    user.value = null
    try {
      if (refreshToken) await authService.logout(refreshToken, accessToken)
    } catch { /* Local logout still completes if the server is unavailable. */ }
    // Recreate every store to prevent data from the previous account leaking into a new session.
    window.location.replace('/login')
  }

  async function fetchMe() {
    if (!tokenStorage.getAccess()) return
    try {
      const res = await authService.me()
      if (res.success) user.value = res.data
    } catch {
      tokenStorage.clear()
      user.value = null
    }
  }

  async function changePassword(currentPassword: string, newPassword: string) {
    const res = await authService.changePassword(currentPassword, newPassword)
    if (!res.success) throw new Error(res.error?.message ?? 'Failed to change password')
  }

  return {
    user,
    loading,
    error,
    isAuthenticated,
    userFullName,
    userRole,
    userInitials,
    login,
    logout,
    fetchMe,
    changePassword,
  }
})
