import { computed, reactive, markRaw } from 'vue'
import { createChats, isChat, type Chat, type Message } from '@/pages/messages/messages.data'

const prefix = 'lk-nngasu-messages-v2:'
const baseKey = `${prefix}base`
const draftsKey = 'lk-nngasu-message-drafts-v1'
const chatKey = (id: string) => `${prefix}chat:${encodeURIComponent(id)}`
const messageKey = (chatId: string, id: string) =>
  `${prefix}message:${encodeURIComponent(chatId)}:${encodeURIComponent(id)}`
const readKey = (id: string) => `${prefix}read:${encodeURIComponent(id)}`

function browserStorage(kind: 'localStorage' | 'sessionStorage'): Storage | undefined {
  try {
    return window[kind]
  } catch {
    return undefined
  }
}

function parseChats(raw: string | null): Chat[] | undefined {
  try {
    const data: unknown = JSON.parse(raw ?? 'null')
    if (Array.isArray(data) && data.length && data.every(isChat)) return data
  } catch {
    /* Invalid browser data falls back to the demo conversations. */
  }
}

function messageId(): string {
  if (typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  // randomUUID may be unavailable when a phone opens the dev server over local HTTP.
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  bytes[6] = (bytes[6]! & 0x0f) | 0x40
  bytes[8] = (bytes[8]! & 0x3f) | 0x80
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

/** Each message has its own key: independent tabs never overwrite another message's history. */
export function createMessagesStore(storage?: Storage, draftStorage?: Storage) {
  const pending = reactive(new Map<string, string>())
  const state = reactive({
    chats: [] as Chat[],
    drafts: {} as Record<string, string>,
    draftError: false,
  })
  let baseline = createChats()
  try {
    const saved = parseChats(storage?.getItem(baseKey) ?? null)
    baseline = saved ?? parseChats(storage?.getItem('lk-nngasu-messages-v1') ?? null) ?? baseline
    if (!saved) storage?.setItem(baseKey, JSON.stringify(baseline))
  } catch {
    /* Sending still works in memory when storage is unavailable. */
  }

  function write(key: string, value: string): boolean {
    try {
      if (!storage) throw new Error('Browser storage is unavailable')
      storage.setItem(key, value)
      pending.delete(key)
      return true
    } catch {
      pending.set(key, value)
      return false
    }
  }

  function refresh() {
    try {
      const persistedBase = parseChats(storage?.getItem(baseKey) ?? null)
      if (persistedBase) baseline = persistedBase
    } catch {
      /* Preserve the in-memory conversations. */
    }
    const merged = new Map(
      baseline.map((chat) => [chat.id, { ...chat, messages: [...chat.messages] }]),
    )
    const records = new Map<string, string>()
    try {
      if (storage)
        for (let i = 0; i < storage.length; i++) {
          const key = storage.key(i)
          if (key?.startsWith(prefix)) {
            const value = storage.getItem(key)
            if (value !== null) records.set(key, value)
          }
        }
    } catch {
      /* Pending writes are overlaid below. */
    }
    for (const [key, value] of pending) records.set(key, value)
    for (const [key, value] of records) {
      if (!key.startsWith(`${prefix}chat:`)) continue
      try {
        const chat: unknown = JSON.parse(value)
        if (isChat(chat) && chatKey(chat.id) === key && !merged.has(chat.id))
          merged.set(chat.id, { ...chat, messages: [] })
      } catch {
        /* Ignore malformed individual records. */
      }
    }
    for (const [key, value] of records) {
      if (!key.startsWith(`${prefix}message:`)) continue
      try {
        const record = JSON.parse(value) as { chatId: string; message: Message }
        const chat = merged.get(record.chatId)
        if (
          chat &&
          isChat({ ...chat, messages: [record.message] }) &&
          key === messageKey(chat.id, record.message.id) &&
          !chat.messages.some((message) => message.id === record.message.id)
        )
          chat.messages.push(record.message)
      } catch {
        /* Ignore malformed individual records. */
      }
    }
    // Preserve visible conversations even if access to storage fails between operations.
    for (const current of state.chats) {
      const chat = merged.get(current.id)
      if (!chat) merged.set(current.id, { ...current, messages: [...current.messages] })
      else
        for (const message of current.messages) {
          if (!chat.messages.some((item) => item.id === message.id)) chat.messages.push(message)
        }
    }
    for (const chat of merged.values()) {
      chat.messages.sort((a, b) => Date.parse(a.sentAt) - Date.parse(b.sentAt))
      if (records.has(readKey(chat.id))) chat.unread = 0
    }
    state.chats = [...merged.values()]
  }

  try {
    const saved: unknown = JSON.parse(draftStorage?.getItem(draftsKey) ?? '{}')
    if (saved && typeof saved === 'object' && !Array.isArray(saved)) {
      state.drafts = Object.fromEntries(
        Object.entries(saved).filter(
          (entry): entry is [string, string] => typeof entry[1] === 'string',
        ),
      )
    }
  } catch {
    /* A draft remains usable in memory. */
  }
  refresh()

  function setDraft(id: string, text: string) {
    state.drafts[id] = text
    try {
      if (!draftStorage) throw new Error('Draft storage is unavailable')
      draftStorage.setItem(draftsKey, JSON.stringify(state.drafts))
      state.draftError = false
    } catch {
      state.draftError = true
    }
  }

  function ensureChat(contact: Omit<Chat, 'messages' | 'unread'>) {
    refresh()
    if (state.chats.some((chat) => chat.id === contact.id)) return
    const chat: Chat = { ...contact, messages: [], unread: 0 }
    write(chatKey(chat.id), JSON.stringify(chat))
    refresh()
  }

  function send(id: string, text: string) {
    refresh()
    const chat = state.chats.find((item) => item.id === id)
    if (!chat || !text.trim() || text.trim().length > 4000) return
    const message: Message = {
      id: messageId(),
      sender: 'Вы',
      own: true,
      text: text.trim(),
      sentAt: new Date().toISOString(),
    }
    // Ensure migrated history is recoverable if storage was blocked on initialization.
    if (!storageHasBase()) write(baseKey, JSON.stringify(baseline))
    write(messageKey(id, message.id), JSON.stringify({ chatId: id, message }))
    refresh()
    setDraft(id, '')
    return message
  }

  function storageHasBase(): boolean {
    try {
      return !!storage?.getItem(baseKey)
    } catch {
      return false
    }
  }
  function markRead(id: string) {
    const chat = state.chats.find((item) => item.id === id)
    if (!chat?.unread) return
    write(readKey(id), '1')
    refresh()
  }
  function isSaved(chatId: string, message: Message): boolean {
    return (
      !pending.has(messageKey(chatId, message.id)) &&
      !pending.has(chatKey(chatId)) &&
      !pending.has(baseKey)
    )
  }
  function retry() {
    for (const [key, value] of pending) write(key, value)
    refresh()
  }
  return markRaw({
    state,
    refresh,
    setDraft,
    ensureChat,
    send,
    markRead,
    isSaved,
    retry,
    unreadMessages: computed(() => state.chats.reduce((sum, chat) => sum + chat.unread, 0)),
    storageError: computed(() => pending.size > 0),
  })
}

let instance: ReturnType<typeof createMessagesStore> | undefined
export function useMessagesStore() {
  if (!instance) {
    instance = createMessagesStore(browserStorage('localStorage'), browserStorage('sessionStorage'))
    const sync = (event: StorageEvent) => {
      if (event.key === null || event.key.startsWith(prefix)) instance?.refresh()
    }
    window.addEventListener('storage', sync)
    import.meta.hot?.dispose(() => window.removeEventListener('storage', sync))
  }
  return instance
}
