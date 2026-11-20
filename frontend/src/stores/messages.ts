import { apiNotifications } from '@/api/client'
import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { ChatConnection, type ConnectionState } from '@/api/realtime'
import type { ServerChat, ServerMessage, MessagePage } from '@/api/cabinet.types'
import type { Chat, Message } from '@/pages/messages/messages.types'

let connection: ChatConnection | null = null

function mapMessage(message: ServerMessage): Message {
  return {
    id: String(message.id),
    text: message.body,
    sender: message.sender_name,
    own: message.sender_id === useAuthStore().user?.id,
    sentAt: message.created_at,
    editedAt: message.edited_at,
    files: message.files,
  }
}

function mapChat(chat: ServerChat, previous?: Chat): Chat {
  const name = chat.peer?.name || chat.title
  return {
    id: String(chat.id),
    unreadCount: chat.unread_count ?? 0,
    lastReadMessageId: chat.last_read_message_id ?? 0,
    name,
    group: chat.kind === 'group',
    initials: name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join(''),
    description: chat.kind === 'group' ? 'Групповой чат' : 'Личный диалог',
    messages: previous?.messages ?? [],
    lastMessage: chat.last_message
      ? {
          ...mapMessage({ ...chat.last_message, chat_id: chat.id, files: [], edited_at: null }),
          text: chat.last_message.body || (chat.last_message.has_files ? 'Вложение' : ''),
        }
      : undefined,
  }
}

export const useMessageStore = defineStore('messages', {
  state: () => ({
    chats: [] as Chat[],
    drafts: {} as Record<string, string>,
    loading: false,
    error: '',
    conversationError: '',
    pendingChatIds: {} as Record<string, boolean>,
    reloadChatIds: {} as Record<string, boolean>,
    activeChatId: '',
    pendingReads: {} as Record<string, boolean>,
    readTargets: {} as Record<string, number>,
    syncedMessageIds: {} as Record<string, number>,
    sending: false,
    sendError: '',
    hasOlder: {} as Record<string, boolean>,
    cursors: {} as Record<string, number | null>,
    connectionState: 'disconnected' as ConnectionState,
    revision: 0,
    changes: {} as Record<string, Record<string, { revision: number; message: Message }>>,
    version: 0,
  }),
  getters: {
    messagesLoading: (state) => !!state.pendingChatIds[state.activeChatId],
  },
  actions: {
    connect() {
      if (connection || !useAuthStore().user) return
      const version = this.version
      connection = new ChatConnection(
        (state) => {
          if (version === this.version) this.connectionState = state
        },
        (event) => {
          if (version !== this.version) return
          if (event.type === 'message.created') this.applyMessage(event.data.message)
          else if (event.type === 'chat.read') {
            this.applyRead(String(event.data.chat_id), event.data.last_read_message_id)
            void this.fetchChat(String(event.data.chat_id))
          } else void this.resync()
        },
        () => {
          if (version !== this.version) return
          void this.resync()
        },
      )
      void connection.start()
    },
    async resync() {
      const version = this.version
      await this.fetchChats()
      if (version !== this.version) return
      // Replay на сервере нет: после подключения перечитываем весь загруженный диапазон.
      await Promise.all(
        this.chats
          .filter((chat) => chat.messages.length || chat.id === this.activeChatId)
          .map((chat) => this.fetchMessages(chat.id)),
      )
    },
    applyChats(chats: ServerChat[], replace = false, since?: number) {
      const revision = since ?? this.revision
      const previous = new Map(this.chats.map((item) => [item.id, item]))
      const mapped = chats.map((chat) => {
        this.syncedMessageIds[String(chat.id)] = chat.last_message?.id ?? 0
        const old = previous.get(String(chat.id))
        const saved = mapChat(chat, old)
        if (
          old &&
          Object.values(this.changes[saved.id] ?? {}).some((change) => change.revision > revision)
        ) {
          saved.lastMessage = old.lastMessage
          saved.unreadCount += Object.values(this.changes[saved.id] ?? {}).filter(
            (change) =>
              change.revision > revision &&
              !change.message.own &&
              Number(change.message.id) > (chat.last_message?.id ?? 0) &&
              Number(change.message.id) > saved.lastReadMessageId,
          ).length
        }
        if (old && old.lastReadMessageId > saved.lastReadMessageId) {
          saved.lastReadMessageId = old.lastReadMessageId
          saved.unreadCount = old.unreadCount
        }
        return saved
      })
      this.chats = [
        ...mapped,
        ...this.chats.filter(
          (item) =>
            !mapped.some((chat) => chat.id === item.id) &&
            (!replace ||
              Object.values(this.changes[item.id] ?? {}).some(
                (change) => change.revision > revision,
              )),
        ),
      ]
    },
    async fetchChats() {
      if (this.loading) return
      const version = this.version
      const revision = this.revision
      this.loading = true
      this.error = ''
      try {
        const chats = await useAuthStore().list<ServerChat>('/chats')
        if (version === this.version) this.applyChats(chats, true, revision)
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить чаты.'
        }
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    async fetchChat(id: string) {
      const version = this.version
      const revision = this.revision
      try {
        const chat = await useAuthStore().request<ServerChat>(`/chats/${id}`, {
          notifyOnError: false,
        })
        if (version !== this.version) return
        this.applyChats([chat], false, revision)
        await this.fetchMessages(id)
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.conversationError =
            error instanceof Error ? error.message : 'Не удалось обновить чат.'
        }
      }
    },
    async openDirect(userId: number) {
      const version = this.version
      const chat = await useAuthStore().request<ServerChat>('/chats/direct', {
        method: 'POST',
        body: JSON.stringify({ user_id: userId }),
      })
      if (version !== this.version) throw new Error('Сессия завершена.')
      this.applyChats([chat])
      return String(chat.id)
    },
    async openGroup(groupId: number) {
      const version = this.version
      const chat = await useAuthStore().request<ServerChat>('/chats/group', {
        method: 'POST',
        body: JSON.stringify({ group_id: groupId }),
      })
      if (version !== this.version) throw new Error('Сессия завершена.')
      this.applyChats([chat])
      return String(chat.id)
    },
    async fetchMessages(id: string, older = false) {
      if (this.pendingChatIds[id]) {
        if (!older) this.reloadChatIds[id] = true
        return
      }
      const version = this.version
      const revision = this.revision
      const chat = this.chats.find((item) => item.id === id)
      if (!chat) return
      const oldest = older ? undefined : chat.messages[0]?.id
      const before = older ? this.cursors[id] : null
      if (older && !before) return
      this.pendingChatIds[id] = true
      this.conversationError = ''
      try {
        let page = await useAuthStore().request<MessagePage>(
          `/chats/${id}/messages?limit=100${before ? `&before=${before}` : ''}`,
        )
        const messages = [...page.items]
        while (
          !older &&
          oldest &&
          page.has_more &&
          page.next_before &&
          (!messages[0] || messages[0].id > Number(oldest))
        ) {
          page = await useAuthStore().request<MessagePage>(
            `/chats/${id}/messages?limit=100&before=${page.next_before}`,
          )
          if (version !== this.version) return
          messages.unshift(...page.items)
        }
        if (version !== this.version) return
        const current = this.chats.find((item) => item.id === id)
        if (!current) return
        const merged = new Map((older ? current.messages : []).map((item) => [item.id, item]))
        for (const message of messages) merged.set(String(message.id), mapMessage(message))
        for (const [messageId, change] of Object.entries(this.changes[id] ?? {})) {
          if (change.revision <= revision) continue
          merged.set(messageId, change.message)
        }
        current.messages = [...merged.values()].sort((a, b) => Number(a.id) - Number(b.id))
        this.hasOlder[id] = page.has_more
        this.cursors[id] = page.next_before
        current.lastMessage = current.messages.at(-1)
        const last = current.messages.at(-1)
        if (
          last &&
          id === this.activeChatId &&
          document.visibilityState === 'visible' &&
          document.hasFocus()
        )
          void this.markRead(id, Number(last.id))
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.conversationError =
            error instanceof Error ? error.message : 'Не удалось загрузить сообщения.'
        }
      } finally {
        if (version === this.version) {
          delete this.pendingChatIds[id]
          if (this.reloadChatIds[id]) {
            delete this.reloadChatIds[id]
            void this.fetchMessages(id)
          }
        }
      }
    },
    applyMessage(message: ServerMessage) {
      const id = String(message.chat_id)
      const saved = mapMessage(message)
      this.changes[id] ??= {}
      this.changes[id][saved.id] = { revision: ++this.revision, message: saved }
      const chat = this.chats.find((item) => item.id === id)
      if (!chat) {
        void this.fetchChat(id)
        return
      }
      const index = chat.messages.findIndex((item) => item.id === saved.id)
      if (index < 0) {
        chat.messages.push(saved)
        if (
          !saved.own &&
          Number(saved.id) > chat.lastReadMessageId &&
          Number(saved.id) > (this.syncedMessageIds[id] ?? 0)
        )
          chat.unreadCount++
      } else chat.messages.splice(index, 1, saved)
      chat.messages = [...chat.messages].sort((a, b) => Number(a.id) - Number(b.id))
      if (!chat.lastMessage || Number(saved.id) >= Number(chat.lastMessage.id))
        chat.lastMessage = saved
      if (saved.own) this.applyRead(id, Number(saved.id))
      if (id === this.activeChatId && document.visibilityState === 'visible' && document.hasFocus())
        void this.markRead(id, Number(saved.id))
    },
    applyRead(id: string, messageId: number) {
      const chat = this.chats.find((item) => item.id === id)
      if (!chat || messageId <= chat.lastReadMessageId) return
      chat.lastReadMessageId = messageId
      // Источник позиции — сервер; события после неё могут уже находиться в очереди клиента.
      const messages = new Map(chat.messages.map((message) => [message.id, message]))
      for (const change of Object.values(this.changes[id] ?? {}))
        messages.set(change.message.id, change.message)
      chat.unreadCount = [...messages.values()].filter(
        (message) => !message.own && Number(message.id) > messageId,
      ).length
    },
    async markRead(id: string, messageId: number) {
      const chat = this.chats.find((item) => item.id === id)
      if (!chat || messageId <= chat.lastReadMessageId) return
      this.readTargets[id] = Math.max(this.readTargets[id] ?? 0, messageId)
      if (this.pendingReads[id]) return
      const version = this.version
      this.pendingReads[id] = true
      try {
        while (
          version === this.version &&
          (this.readTargets[id] ?? 0) >
            (this.chats.find((item) => item.id === id)?.lastReadMessageId ?? Infinity)
        ) {
          const target = this.readTargets[id]!
          await useAuthStore().request(`/chats/${id}/read`, {
            method: 'POST',
            body: JSON.stringify({ message_id: target }),
            notifyOnError: false,
          })
          if (version !== this.version) return
          this.applyRead(id, target)
        }
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.conversationError =
            error instanceof Error ? error.message : 'Не удалось отметить сообщения прочитанными.'
        }
      } finally {
        if (version === this.version) delete this.pendingReads[id]
      }
    },
    async send(id: string, text: string, fileIds: number[] = []): Promise<boolean> {
      if (this.sending || !connection) return false
      const version = this.version
      this.sending = true
      this.sendError = ''
      try {
        const message = await connection.command<ServerMessage>('message.send', {
          chat_id: Number(id),
          body: text,
          file_ids: fileIds,
        })
        if (version !== this.version) return false
        this.applyMessage(message)
        if (this.drafts[id]?.trim() === text) delete this.drafts[id]
        return true
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.sendError =
            error instanceof Error ? error.message : 'Не удалось отправить сообщение.'
          void this.fetchMessages(id)
        }
        return false
      } finally {
        if (version === this.version) this.sending = false
      }
    },
    clearSession() {
      const version = this.version + 1
      this.version = version
      connection?.stop()
      connection = null
      this.$reset()
      // $reset не должен возвращать поколение к нулю: старые HTTP-ответы игнорируются.
      this.version = version
    },
  },
})
