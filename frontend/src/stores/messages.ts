import { defineStore, acceptHMRUpdate } from 'pinia'
import { createChats, isChat } from '@/pages/messages/messages.data'
import type { Chat, Message } from '@/pages/messages/messages.types'

export const messagesStoragePrefix = 'lk-nngasu-messages-v2:'
const baseKey = `${messagesStoragePrefix}base`
const draftsKey = 'lk-nngasu-message-drafts-v1'
const chatKey = (id: string) => `${messagesStoragePrefix}chat:${encodeURIComponent(id)}`
const messageKey = (chatId: string, id: string) =>
  `${messagesStoragePrefix}message:${encodeURIComponent(chatId)}:${encodeURIComponent(id)}`
const readKey = (id: string) => `${messagesStoragePrefix}read:${encodeURIComponent(id)}`

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

export const useMessageStore = defineStore('messages', {
  state: () => ({
    chats: [] as Chat[],
    drafts: {} as Record<string, string>,
    draftError: false,
    baseline: createChats(),
    pending: {} as Record<string, string>,
  }),
  getters: {
    unreadMessages: (state): number => state.chats.reduce((sum, chat) => sum + chat.unread, 0),
    storageError: (state): boolean => Object.keys(state.pending).length > 0,
  },
  actions: {
    initialize() {
      try {
        const saved = parseChats(browserStorage('localStorage')?.getItem(baseKey) ?? null)
        this.baseline =
          saved ??
          parseChats(browserStorage('localStorage')?.getItem('lk-nngasu-messages-v1') ?? null) ??
          this.baseline
        if (!saved) browserStorage('localStorage')?.setItem(baseKey, JSON.stringify(this.baseline))
      } catch {
        /* Sending still works in memory when storage is unavailable. */
      }
      try {
        const saved: unknown = JSON.parse(
          browserStorage('sessionStorage')?.getItem(draftsKey) ?? '{}',
        )
        if (saved && typeof saved === 'object' && !Array.isArray(saved)) {
          this.drafts = Object.fromEntries(
            Object.entries(saved).filter(
              (entry): entry is [string, string] => typeof entry[1] === 'string',
            ),
          )
        }
      } catch {
        /* A draft remains usable in memory. */
      }
      this.refresh()
    },
    write(key: string, value: string): boolean {
      const storage = browserStorage('localStorage')
      try {
        if (!storage) throw new Error('Browser storage is unavailable')
        storage.setItem(key, value)
        delete this.pending[key]
        return true
      } catch {
        this.pending[key] = value
        return false
      }
    },
    refresh() {
      const storage = browserStorage('localStorage')
      try {
        const persistedBase = parseChats(storage?.getItem(baseKey) ?? null)
        if (persistedBase) this.baseline = persistedBase
      } catch {
        /* Preserve the in-memory conversations. */
      }
      const merged = new Map(
        this.baseline.map((chat) => [chat.id, { ...chat, messages: [...chat.messages] }]),
      )
      const records = new Map<string, string>()
      try {
        if (storage)
          for (let i = 0; i < storage.length; i++) {
            const key = storage.key(i)
            if (key?.startsWith(messagesStoragePrefix)) {
              const value = storage.getItem(key)
              if (value !== null) records.set(key, value)
            }
          }
      } catch {
        /* Pending writes are overlaid below. */
      }
      for (const [key, value] of Object.entries(this.pending)) records.set(key, value)
      for (const [key, value] of records) {
        if (!key.startsWith(`${messagesStoragePrefix}chat:`)) continue
        try {
          const chat: unknown = JSON.parse(value)
          if (isChat(chat) && chatKey(chat.id) === key && !merged.has(chat.id))
            merged.set(chat.id, { ...chat, messages: [] })
        } catch {
          /* Ignore malformed individual records. */
        }
      }
      for (const [key, value] of records) {
        if (!key.startsWith(`${messagesStoragePrefix}message:`)) continue
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
      for (const current of this.chats) {
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
      this.chats = [...merged.values()]
    },
    setDraft(id: string, text: string) {
      const draftStorage = browserStorage('sessionStorage')
      this.drafts[id] = text
      try {
        if (!draftStorage) throw new Error('Draft storage is unavailable')
        draftStorage.setItem(draftsKey, JSON.stringify(this.drafts))
        this.draftError = false
      } catch {
        this.draftError = true
      }
    },
    ensureChat(contact: Omit<Chat, 'messages' | 'unread'>) {
      this.refresh()
      if (this.chats.some((chat) => chat.id === contact.id)) return
      const chat: Chat = { ...contact, messages: [], unread: 0 }
      this.write(chatKey(chat.id), JSON.stringify(chat))
      this.refresh()
    },
    send(id: string, text: string): Message | undefined {
      this.refresh()
      const chat = this.chats.find((item) => item.id === id)
      if (!chat || !text.trim() || text.trim().length > 4000) return
      const message: Message = {
        id: messageId(),
        sender: 'Вы',
        own: true,
        text: text.trim(),
        sentAt: new Date().toISOString(),
      }
      // Ensure migrated history is recoverable if storage was blocked on initialization.
      if (!this.storageHasBase()) this.write(baseKey, JSON.stringify(this.baseline))
      this.write(messageKey(id, message.id), JSON.stringify({ chatId: id, message }))
      this.refresh()
      this.setDraft(id, '')
      return message
    },
    storageHasBase(): boolean {
      const storage = browserStorage('localStorage')
      try {
        return !!storage?.getItem(baseKey)
      } catch {
        return false
      }
    },
    markRead(id: string) {
      const chat = this.chats.find((item) => item.id === id)
      if (!chat?.unread) return
      this.write(readKey(id), '1')
      this.refresh()
    },
    isSaved(chatId: string, message: Message): boolean {
      return (
        this.pending[messageKey(chatId, message.id)] === undefined &&
        this.pending[chatKey(chatId)] === undefined &&
        this.pending[baseKey] === undefined
      )
    },
    retry() {
      for (const [key, value] of Object.entries(this.pending)) this.write(key, value)
      this.refresh()
    },
  },
})

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useMessageStore, import.meta.hot))
}
