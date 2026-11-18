import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ApiError, apiRequest } from '@/api/client'

interface AuthUser {
  id: number
  last_name: string
  first_name: string
  middle_names: string | null
  full_name: string
  email: string
  is_active: boolean
  is_admin: boolean
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

export const useAuthStore = defineStore('auth', () => {
  // Access-токен и профиль живут только в памяти; refresh-cookie управляет браузер.
  const session = ref<AuthToken | null>(null)
  const initialized = ref(false)
  let refreshPromise: Promise<void> | null = null
  const user = computed(() => session.value?.user ?? null)
  const roleLabel = computed(() =>
    [
      user.value?.is_admin && 'Администратор',
      user.value?.student && 'Студент',
      user.value?.teacher && 'Преподаватель',
    ]
      .filter(Boolean)
      .join(', '),
  )

  async function refresh() {
    if (refreshPromise) return refreshPromise
    refreshPromise = (async () => {
      try {
        const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf')
        session.value = await apiRequest<AuthToken>('/auth/refresh', {
          method: 'POST',
          headers: { 'X-CSRF-Token': csrf.csrf_token },
        })
        initialized.value = true
      } catch (error) {
        if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
          session.value = null
          initialized.value = true
          if (error.status === 401) return
        }
        throw error
      } finally {
        refreshPromise = null
      }
    })()
    return refreshPromise
  }

  async function ensureSession() {
    if (
      !initialized.value ||
      (session.value && Date.parse(session.value.expires_at) <= Date.now() + 30000)
    ) {
      await refresh()
    }
  }

  async function signIn(email: string, password: string, rememberMe: boolean) {
    if (refreshPromise) await refreshPromise
    session.value = await apiRequest<AuthToken>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: email.trim(), password, remember_me: rememberMe }),
    })
    initialized.value = true
  }

  async function signOut() {
    if (refreshPromise) await refreshPromise
    try {
      const csrf = await apiRequest<{ csrf_token: string }>('/auth/csrf')
      await apiRequest<void>('/auth/logout', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrf.csrf_token },
      })
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error
    }
    session.value = null
    initialized.value = true
  }

  return { session, user, roleLabel, ensureSession, refresh, signIn, signOut }
})
