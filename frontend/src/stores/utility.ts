import { ApiError } from '@/api/client'
import { defineStore } from 'pinia'
import breakpoints from '@/styles/breakpoints.module.scss'

export type NotificationVariant = 'error' | 'success' | 'neutral'

export interface Notification {
  message: string
  variant: NotificationVariant
}

export type ThemePreference = 'system' | 'light' | 'dark'

export const useUtilityStore = defineStore('utility', {
  state: () => {
    const initialTheme =
      typeof document === 'undefined' ? 'system' : document.documentElement.dataset.theme
    return {
      notification: null as Notification | null,
      themePreference: (initialTheme === 'light' || initialTheme === 'dark'
        ? initialTheme
        : 'system') as ThemePreference,
    }
  },
  actions: {
    notifyError(error: unknown) {
      if (error instanceof ApiError && error.code === 'current_semester_not_set') return
      const message =
        error instanceof Error ? error.message : typeof error === 'string' ? error : ''
      this.notify(message || 'Не удалось выполнить действие. Попробуйте позже.', 'error')
    },
    notify(message: string, variant: NotificationVariant = 'neutral') {
      this.notification = { message, variant }
    },
    isMobileVersion(): boolean {
      return (
        typeof window !== 'undefined' &&
        window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      )
    },
  },
})
