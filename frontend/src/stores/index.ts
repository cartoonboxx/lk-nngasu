import { watch, type Plugin } from 'vue'
import { createPinia, disposePinia } from 'pinia'
import { useMessageStore } from './messages'
import { useUtilityStore } from './utility'
import { useSubjectsStore } from './subjects'
import { useScheduleStore } from './schedule'
import { useAdminStore } from './admin'
import { useAuthStore } from './auth'
import { useCabinetStore } from './cabinet'
import { apiNotifications } from '@/api/client'

const stores: Plugin = {
  install(app) {
    const pinia = createPinia()
    app.use(pinia)

    app.config.globalProperties.$utilityStore = useUtilityStore(pinia)
    app.config.globalProperties.$subjectsStore = useSubjectsStore(pinia)
    app.config.globalProperties.$scheduleStore = useScheduleStore(pinia)
    const adminStore = useAdminStore(pinia)
    app.config.globalProperties.$adminStore = adminStore
    const authStore = useAuthStore(pinia)
    const cabinetStore = useCabinetStore(pinia)
    const messageStore = useMessageStore(pinia)
    app.config.globalProperties.$messageStore = messageStore
    const stopSessionSync = watch(
      () => authStore.user?.id,
      (id) => {
        adminStore.clearSession()
        app.config.globalProperties.$subjectsStore.clearSession()
        app.config.globalProperties.$scheduleStore.clearSession()
        messageStore.clearSession()
        cabinetStore.clearSession()
        if (id) messageStore.connect()
      },
      { flush: 'sync' },
    )
    const stopAvatarSync = watch(
      () => [authStore.user?.id, authStore.user?.avatar?.id],
      async () => {
        await authStore.loadAvatar()
      },
      { flush: 'sync' },
    )
    const utilityStore = app.config.globalProperties.$utilityStore
    const reportedErrors = new WeakSet<object>()
    apiNotifications.onError = (error) => {
      if (error !== null && typeof error === 'object') {
        if (reportedErrors.has(error)) return
        reportedErrors.add(error)
      }
      utilityStore.notifyError(error)
    }
    const stopThemeSync = watch(
      () => utilityStore.themePreference,
      (value) => {
        document.documentElement.dataset.theme = value
        try {
          localStorage.setItem('lk-nngasu-theme', value)
        } catch {
          /* Настройка остаётся в памяти. */
        }
      },
      { immediate: true, flush: 'sync' },
    )
    const syncTheme = (event: StorageEvent) => {
      if (event.key === 'lk-nngasu-theme' || event.key === null) {
        utilityStore.themePreference =
          event.newValue === 'light' || event.newValue === 'dark' ? event.newValue : 'system'
      }
    }
    window.addEventListener('storage', syncTheme)
    const cleanup = () => {
      apiNotifications.onError = null
      stopSessionSync()
      stopAvatarSync()
      stopThemeSync()
      window.removeEventListener('storage', syncTheme)
      authStore.avatarVersion++
      messageStore.clearSession()
      if (authStore.avatarUrl) URL.revokeObjectURL(authStore.avatarUrl)
    }
    app.onUnmount(() => {
      cleanup()
      disposePinia(pinia)
    })
    import.meta.hot?.dispose(cleanup)
  },
}

export default stores
