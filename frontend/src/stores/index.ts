import { watch, type Plugin } from 'vue'
import { createPinia, disposePinia } from 'pinia'
import { useMessageStore } from './messages'
import { useUtilityStore } from './utility'
import { useSubjectsStore } from './subjects'
import { useScheduleStore } from './schedule'
import { useAdminStore } from './admin'
import { useAuthStore } from './auth'
import { useCabinetStore } from './cabinet'

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
      () => {
        adminStore.clearSession()
        app.config.globalProperties.$subjectsStore.clearSession()
        app.config.globalProperties.$scheduleStore.clearSession()
        messageStore.clearSession()
        cabinetStore.clearSession()
      },
      { flush: 'sync' },
    )
    const stopAvatarSync = watch(
      () => [authStore.user?.id, authStore.user?.avatar?.id],
      () => {
        void authStore.loadAvatar()
      },
      { flush: 'sync' },
    )
    const utilityStore = app.config.globalProperties.$utilityStore
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
      stopSessionSync()
      stopAvatarSync()
      stopThemeSync()
      window.removeEventListener('storage', syncTheme)
      authStore.avatarVersion++
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
