import type { useMessageStore } from '@/stores/messages'

declare module 'vue' {
  interface ComponentCustomProperties {
    $messageStore: ReturnType<typeof useMessageStore>
  }
}
