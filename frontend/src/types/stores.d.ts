import type { useMessageStore } from '@/stores/messages'
import type { useUtilityStore } from '@/stores/utility'
import type { useSubjectsStore } from '@/stores/subjects'

declare module 'vue' {
  interface ComponentCustomProperties {
    $messageStore: ReturnType<typeof useMessageStore>
    $utilityStore: ReturnType<typeof useUtilityStore>
    $subjectsStore: ReturnType<typeof useSubjectsStore>
  }
}
