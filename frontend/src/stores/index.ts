import type { Plugin } from 'vue'
import { createPinia, disposePinia } from 'pinia'
import { useMessageStore, messagesStoragePrefix } from './messages'
import { useUtilityStore } from './utility'
import { useSubjectsStore } from './subjects'
import { useScheduleStore } from './schedule'

const stores: Plugin = {
  install(app) {
    const pinia = createPinia()
    app.use(pinia)

    app.config.globalProperties.$utilityStore = useUtilityStore(pinia)
    app.config.globalProperties.$subjectsStore = useSubjectsStore(pinia)
    app.config.globalProperties.$scheduleStore = useScheduleStore(pinia)

    const messageStore = useMessageStore(pinia)
    messageStore.initialize()
    app.config.globalProperties.$messageStore = messageStore

    const syncMessages = (event: StorageEvent) => {
      if (event.key === null || event.key.startsWith(messagesStoragePrefix)) {
        messageStore.refresh()
      }
    }
    window.addEventListener('storage', syncMessages)
    const cleanup = () => window.removeEventListener('storage', syncMessages)
    app.onUnmount(() => {
      cleanup()
      disposePinia(pinia)
    })
    import.meta.hot?.dispose(cleanup)
  },
}

export default stores
