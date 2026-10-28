import type { useMessageStore } from '@/stores/messages'
import type { useUtilityStore } from '@/stores/utility'

declare module 'vue' {
  interface ComponentCustomProperties {
    $messageStore: ReturnType<typeof useMessageStore>
    $utilityStore: ReturnType<typeof useUtilityStore>
  }
}
