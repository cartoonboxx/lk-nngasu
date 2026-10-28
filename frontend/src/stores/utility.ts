import { defineStore } from 'pinia'
import breakpoints from '@/styles/breakpoints.module.scss'

export const useUtilityStore = defineStore('utility', {
  actions: {
    isMobileVersion(): boolean {
      return (
        typeof window !== 'undefined' &&
        window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      )
    },
  },
})
