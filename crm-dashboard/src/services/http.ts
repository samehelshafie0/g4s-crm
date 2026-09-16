import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios'

// ─── Axios Instance ─────────────────────────────────────────
export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
})

// ─── Token helpers ──────────────────────────────────────────
const TOKEN_KEY = 'g4s_access_token'
const REFRESH_KEY = 'g4s_refresh_token'

export const tokenStorage = {
  getAccess: () => localStorage.getItem(TOKEN_KEY),
  getRefresh: () => localStorage.getItem(REFRESH_KEY),
  setTokens: (access: string, refresh: string) => {
    localStorage.setItem(TOKEN_KEY, access)
    localStorage.setItem(REFRESH_KEY, refresh)
  },
  clear: () => {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

// ─── Request Interceptor — attach Bearer token ───────────────
http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStorage.getAccess()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// ─── Response Interceptor — auto-refresh on 401 ─────────────
// One bounded refresh request shared by all expired requests. Every request retries once.
let refreshPromise: Promise<string> | null = null
export const refreshHttp = axios.create({ baseURL: '/api/v1', timeout: 15_000 })

function expireSession() {
  tokenStorage.clear()
  if (window.location.pathname !== '/login') window.location.replace('/login')
}

http.interceptors.response.use(
  response => response,
  async (error: AxiosError) => {
    const request = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined
    if (error.response?.status !== 401 || !request || request.url?.startsWith('/auth/login') || request.url?.startsWith('/auth/refresh') || request.url?.startsWith('/auth/logout')) throw error
    if (request._retry) { expireSession(); throw error }
    request._retry = true
    const refresh = tokenStorage.getRefresh()
    if (!refresh) { expireSession(); throw error }
    try {
      if (!refreshPromise) {
        refreshPromise = refreshHttp.post<ApiResponse<{ accessToken: string; refreshToken: string }>>('/auth/refresh', { refreshToken: refresh })
          .then(({ data }) => {
            if (!data.success || !data.data?.accessToken || !data.data?.refreshToken) throw new Error('Invalid refresh response')
            // A logout or different login while refresh was in flight must win.
            if (tokenStorage.getRefresh() !== refresh) throw new Error('Session changed')
            tokenStorage.setTokens(data.data.accessToken, data.data.refreshToken)
            return data.data.accessToken
          }).finally(() => { refreshPromise = null })
      }
      const token = await refreshPromise
      request.headers.Authorization = `Bearer ${token}`
      return http(request)
    } catch (refreshError) {
      if (tokenStorage.getRefresh() === refresh) expireSession()
      throw refreshError
    }
  },
)

// ─── API response shape ─────────────────────────────────────
export interface ApiResponse<T> {
  success: boolean
  data: T
  meta?: {
    page: number
    limit: number
    total: number
    totalPages: number
  }
  error?: {
    code: string
    message: string
    details?: Array<{ field: string; message: string; code: string }>
  }
}

export interface PaginationParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  q?: string
  [key: string]: unknown
}
