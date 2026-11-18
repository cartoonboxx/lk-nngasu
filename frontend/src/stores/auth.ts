import { defineStore } from 'pinia'
import { ApiError, apiRequest, type ApiRequestOptions } from '@/api/client'

interface AuthUser {
  id: number
  last_name: string
  first_name: string
  middle_names: string | null
  full_name: string
  email: string
  is_active: boolean
  is_admin: boolean
  can_manage_admins: boolean
  avatar: { id: number; url: string } | null
  student: {
    group_id: number
    group_name: string
    record_book_number: string | null
    is_headman: boolean
  } | null
  teacher: { department_id: number; department_name: string } | null
}

interface AuthToken {
  access_token: string
  token_type: 'Bearer'
  expires_at: string
  user: AuthUser
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    // Access-токен и профиль живут только в памяти; refresh-cookie управляет браузер.
    session: null as AuthToken | null,
    initialized: false,
    refreshPromise: null as Promise<void> | null,
  }),
  getters: {
    user: (state): AuthUser | null => state.session?.user ?? null,
    roleLabel(): string {
      return [
        this.user?.is_admin && (this.user.can_manage_admins ? 'Владелец' : 'Администратор'),
        this.user?.student && 'Студент',
        this.user?.teacher && 'Преподаватель',
      ]
        .filter(Boolean)
        .join(', ')
    },
  },
  actions: {
    async request<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
      await this.ensureSession()
      if (!this.session) throw new ApiError(401, 'session_expired')
      const headers = new Headers(options.headers)
      headers.set('Authorization', `Bearer ${this.session.access_token}`)
      try {
        return await apiRequest<T>(path, { ...options, headers })
      } catch (error) {
        if (error instanceof ApiError && error.code === 'token_expired') {
          await this.refresh()
          if (!this.session) throw new ApiError(401, 'session_expired')
          headers.set('Authorization', `Bearer ${this.session.access_token}`)
          return apiRequest<T>(path, { ...options, headers })
        }
        if (error instanceof ApiError && error.status === 401) this.session = null
        throw error
      }
    },

    async refresh(): Promise<void> {
      if (this.refreshPromise) return this.refreshPromise
      this.refreshPromise = (async () => {
        try {
          const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf')
          this.session = await apiRequest<AuthToken>('/auth/refresh', {
            method: 'POST',
            headers: { 'X-CSRF-Token': csrf.csrf_token },
          })
          this.initialized = true
        } catch (error) {
          if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
            this.session = null
            this.initialized = true
            if (error.status === 401) return
          }
          throw error
        } finally {
          this.refreshPromise = null
        }
      })()
      return this.refreshPromise
    },

    async ensureSession() {
      if (
        !this.initialized ||
        (this.session && Date.parse(this.session.expires_at) <= Date.now() + 30000)
      ) {
        await this.refresh()
      }
    },

    async signIn(email: string, password: string, rememberMe: boolean) {
      if (this.refreshPromise) await this.refreshPromise
      this.session = await apiRequest<AuthToken>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: email.trim(), password, remember_me: rememberMe }),
      })
      this.initialized = true
    },

    async signOut() {
      if (this.refreshPromise) await this.refreshPromise
      try {
        const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf')
        await apiRequest<void>('/auth/logout', {
          method: 'POST',
          headers: { 'X-CSRF-Token': csrf.csrf_token },
        })
      } catch (error) {
        if (!(error instanceof ApiError) || error.status !== 401) throw error
      }
      this.session = null
      this.initialized = true
    },
  },
})
