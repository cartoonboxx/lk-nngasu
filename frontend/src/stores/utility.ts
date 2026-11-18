import { ApiError } from '@/api/client'
import { defineStore } from 'pinia'
import breakpoints from '@/styles/breakpoints.module.scss'

export interface Notification {
  id: number
  code: string
  message: string
  status?: number
  count: number
  expiresAt: number
}

export type ThemePreference = 'system' | 'light' | 'dark'

export const useUtilityStore = defineStore('utility', {
  state: () => {
    const initialTheme =
      typeof document === 'undefined' ? 'system' : document.documentElement.dataset.theme
    return {
      notifications: [] as Notification[],
      nextNotificationId: 0,
      themePreference: (initialTheme === 'light' || initialTheme === 'dark'
        ? initialTheme
        : 'system') as ThemePreference,
    }
  },
  actions: {
    notifyError(error: unknown) {
      const isApiError = error instanceof ApiError
      const message =
        error instanceof Error ? error.message : typeof error === 'string' ? error : ''
      return this.notify({
        code: isApiError ? error.code : 'client_error',
        status: isApiError && error.status > 0 ? error.status : undefined,
        message: message || 'Не удалось выполнить действие. Попробуйте позже.',
      })
    },
    notify(notification: Pick<Notification, 'code' | 'message' | 'status'>, duration = 10000) {
      const previous = this.notifications.find(
        (item) =>
          item.code === notification.code &&
          item.status === notification.status &&
          item.message === notification.message,
      )
      if (previous) {
        previous.count++
        previous.expiresAt = Date.now() + duration
        return previous.id
      }
      const id = ++this.nextNotificationId
      this.notifications.unshift({
        ...notification,
        id,
        count: 1,
        expiresAt: Date.now() + duration,
      })
      this.notifications = this.notifications.slice(0, 4)
      return id
    },
    dismissNotification(id: number) {
      this.notifications = this.notifications.filter((item) => item.id !== id)
    },
    isMobileVersion(): boolean {
      return (
        typeof window !== 'undefined' &&
        window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      )
    },
  },
})
