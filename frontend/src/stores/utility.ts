import { defineStore } from 'pinia'
import breakpoints from '@/styles/breakpoints.module.scss'

export type ThemePreference = 'system' | 'light' | 'dark'

export const useUtilityStore = defineStore('utility', {
  state: () => {
    const initialTheme = document.documentElement.dataset.theme
    return {
      themePreference: (initialTheme === 'light' || initialTheme === 'dark'
        ? initialTheme
        : 'system') as ThemePreference,
    }
  },
  actions: {
    isMobileVersion(): boolean {
      return (
        typeof window !== 'undefined' &&
        window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      )
    },
  },
})
