import { defineStore } from 'pinia'
import {
  ApiError,
  apiRequest,
  apiNotifications,
  type ApiPage,
  type ApiRequestOptions,
} from '@/api/client'

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
    avatarUrl: '',
    avatarError: '',
    avatarVersion: 0,
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
      try {
        await this.ensureSession()
        if (!this.session) throw new ApiError(401, 'session_expired')
        const userId = this.user?.id
        const headers = new Headers(options.headers)
        headers.set('Authorization', `Bearer ${this.session.access_token}`)
        try {
          const result = await apiRequest<T>(path, { ...options, headers, notifyOnError: false })
          if (this.user?.id !== userId) throw new ApiError(401, 'session_expired')
          return result
        } catch (error) {
          if (
            error instanceof ApiError &&
            error.code === 'token_expired' &&
            this.user?.id === userId
          ) {
            await this.refresh()
            if (!this.session || this.user?.id !== userId)
              throw new ApiError(401, 'session_expired')
            headers.set('Authorization', `Bearer ${this.session.access_token}`)
            try {
              const result = await apiRequest<T>(path, {
                ...options,
                headers,
                notifyOnError: false,
              })
              if (this.user?.id !== userId) throw new ApiError(401, 'session_expired')
              return result
            } catch (retryError) {
              if (
                retryError instanceof ApiError &&
                retryError.status === 401 &&
                this.user?.id === userId
              )
                this.session = null
              throw retryError
            }
          }
          if (error instanceof ApiError && error.status === 401 && this.user?.id === userId)
            this.session = null
          throw error
        }
      } catch (error) {
        if (options.notifyOnError !== false && !options.signal?.aborted)
          apiNotifications.onError?.(error)
        throw error
      }
    },

    async loadAvatar() {
      const version = ++this.avatarVersion
      if (this.avatarUrl) URL.revokeObjectURL(this.avatarUrl)
      this.avatarUrl = ''
      this.avatarError = ''
      const id = this.user?.avatar?.id
      if (!id) return
      try {
        const blob = await this.request<Blob>(`/files/${id}`, { responseType: 'blob' })
        if (version === this.avatarVersion) this.avatarUrl = URL.createObjectURL(blob)
      } catch (error) {
        if (version === this.avatarVersion)
          this.avatarError = error instanceof Error ? error.message : 'Не удалось загрузить фото.'
      }
    },

    async getMe() {
      const userId = this.user?.id
      const user = await this.request<AuthUser>('/me')
      if (this.session && this.user?.id === userId) {
        this.session.user = { ...this.session.user, ...user }
      }
      return user
    },

    async updateMe(values: { avatar_file_id?: number | null }) {
      const userId = this.user?.id
      const user = await this.request<AuthUser>('/me', {
        method: 'PATCH',
        body: JSON.stringify(values),
      })
      if (this.session && this.user?.id === userId) {
        this.session.user = { ...this.session.user, ...user }
      }
      return user
    },

    async changePassword(oldPassword: string, newPassword: string) {
      await this.request<void>('/me/password', {
        method: 'POST',
        body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
      })
      this.session = null
      this.initialized = true
    },

    async list<T>(path: string): Promise<T[]> {
      const result: T[] = []
      let offset = 0
      while (true) {
        const page = await this.request<ApiPage<T>>(
          `${path}${path.includes('?') ? '&' : '?'}limit=200&offset=${offset}`,
          { paginated: true },
        )
        result.push(...page.data)
        offset += page.data.length
        if (offset >= page.meta.total || !page.data.length) {
          return result
        }
      }
    },

    async refresh(): Promise<void> {
      if (this.refreshPromise) return this.refreshPromise
      this.refreshPromise = (async () => {
        try {
          const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf', {
            notifyOnError: false,
          })
          this.session = await apiRequest<AuthToken>('/auth/refresh', {
            notifyOnError: false,
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
          apiNotifications.onError?.(error)
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
      const session = await apiRequest<AuthToken>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: email.trim(), password, remember_me: rememberMe }),
      })
      this.initialized = true
      this.session = session
    },

    async signOut() {
      if (this.refreshPromise) await this.refreshPromise
      try {
        const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf', {
          notifyOnError: false,
        })
        await apiRequest<void>('/auth/logout', {
          notifyOnError: false,
          method: 'POST',
          headers: { 'X-CSRF-Token': csrf.csrf_token },
        })
      } catch (error) {
        if (!(error instanceof ApiError) || error.status !== 401) {
          apiNotifications.onError?.(error)
          throw error
        }
      }
      this.session = null
      this.initialized = true
    },
  },
})
