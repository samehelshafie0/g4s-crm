import { http, type ApiResponse } from './http'

export interface LoginPayload {
  email: string
  password: string
}

export interface AuthUser {
  id: string
  email: string
  firstName: string
  lastName: string
  role: string
  department: string
  isActive: boolean
  lastLoginAt?: string
}

export interface AuthTokens {
  user: AuthUser
  accessToken: string
  refreshToken: string
}

export const authService = {
  login: (payload: LoginPayload) =>
    http.post<ApiResponse<AuthTokens>>('/auth/login', payload).then((r) => r.data),

  logout: (refreshToken: string, accessToken: string | null) =>
    http.post<ApiResponse<null>>('/auth/logout', { refreshToken }, { headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {} }).then((r) => r.data),

  refresh: (refreshToken: string) =>
    http.post<ApiResponse<AuthTokens>>('/auth/refresh', { refreshToken }).then((r) => r.data),

  me: () => http.get<ApiResponse<AuthUser>>('/auth/me').then((r) => r.data),

  changePassword: (currentPassword: string, newPassword: string) =>
    http
      .patch<ApiResponse<null>>('/auth/change-password', { currentPassword, newPassword })
      .then((r) => r.data),
}
