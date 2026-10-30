import { onScopeDispose, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import breakpoints from '@/styles/breakpoints.module.scss'

export type ThemePreference = 'system' | 'light' | 'dark'

export const useUtilityStore = defineStore('utility', () => {
  const initialTheme = document.documentElement.dataset.theme
  const themePreference = ref<ThemePreference>(
    initialTheme === 'light' || initialTheme === 'dark' ? initialTheme : 'system',
  )

  watch(
    themePreference,
    (value) => {
      document.documentElement.dataset.theme = value
      try {
        localStorage.setItem('lk-nngasu-theme', value)
      } catch {
        // Keep theme switching available when browser storage is blocked.
      }
    },
    { immediate: true, flush: 'sync' },
  )

  const syncPreference = (event: StorageEvent) => {
    if (
      event.storageArea !== localStorage ||
      (event.key !== null && event.key !== 'lk-nngasu-theme')
    )
      return
    themePreference.value =
      event.newValue === 'light' || event.newValue === 'dark' ? event.newValue : 'system'
  }
  window.addEventListener('storage', syncPreference)
  onScopeDispose(() => window.removeEventListener('storage', syncPreference))

  function isMobileVersion(): boolean {
    return (
      typeof window !== 'undefined' &&
      window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
    )
  }

  return { themePreference, isMobileVersion }
})
