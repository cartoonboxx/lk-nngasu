import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import type { ServerChat, ServerMessage } from '@/api/cabinet.types'
import type { Chat, Message } from '@/pages/messages/messages.types'

function mapMessage(message: ServerMessage): Message {
  return {
    id: String(message.id),
    text: message.deleted
      ? 'Сообщение удалено'
      : message.body || (message.files.length ? 'Вложение' : ''),
    sender: message.sender_name,
    own: message.sender_id === useAuthStore().user?.id,
    sentAt: message.created_at,
    files: message.deleted ? [] : message.files,
  }
}

function mapChat(chat: ServerChat, previous?: Chat): Chat {
  const name = chat.peer?.name || chat.title
  return {
    id: String(chat.id),
    name,
    group: chat.kind === 'group',
    initials: name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0])
      .join(''),
    description: chat.kind === 'group' ? 'Групповой чат' : 'Личный диалог',
    unread: chat.unread,
    messages: previous?.messages ?? [],
    lastMessage: chat.last_message
      ? {
          ...mapMessage({ ...chat.last_message, chat_id: chat.id, files: [] }),
          text: chat.last_message.deleted
            ? 'Сообщение удалено'
            : chat.last_message.body || (chat.last_message.has_files ? 'Вложение' : ''),
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
    activeChatId: '',
    sending: false,
    sendError: '',
    hasOlder: {} as Record<string, boolean>,
    version: 0,
  }),
  getters: {
    unreadMessages: (state) => state.chats.reduce((sum, chat) => sum + chat.unread, 0),
    messagesLoading: (state) => !!state.pendingChatIds[state.activeChatId],
  },
  actions: {
    async fetchChats() {
      if (this.loading) return
      const version = this.version
      this.loading = true
      this.error = ''
      try {
        const chats = await useAuthStore().request<ServerChat[]>('/chats')
        if (version !== this.version) return
        this.chats = chats.map((chat) =>
          mapChat(
            chat,
            this.chats.find((item) => item.id === String(chat.id)),
          ),
        )
      } catch (error) {
        if (version === this.version)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить чаты.'
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    async openDirect(userId: number) {
      const version = this.version
      const chat = await useAuthStore().request<ServerChat>('/chats/direct', {
        method: 'POST',
        body: JSON.stringify({ user_id: userId }),
      })
      if (version !== this.version) throw new Error('Сессия завершена.')
      const previous = this.chats.find((item) => item.id === String(chat.id))
      if (!previous) this.chats.unshift(mapChat(chat))
      return String(chat.id)
    },
    async openGroup(groupId: number) {
      const version = this.version
      const chat = await useAuthStore().request<ServerChat>('/chats/group', {
        method: 'POST',
        body: JSON.stringify({ group_id: groupId }),
      })
      if (version !== this.version) throw new Error('Сессия завершена.')
      if (!this.chats.some((item) => item.id === String(chat.id))) this.chats.unshift(mapChat(chat))
      return String(chat.id)
    },
    async fetchMessages(id: string, older = false) {
      if (this.pendingChatIds[id]) return
      const version = this.version
      const initial = this.chats.find((item) => item.id === id)
      if (!initial) return
      this.pendingChatIds[id] = true
      this.conversationError = ''
      const first = older ? initial.messages[0]?.id : undefined
      const latest = initial.messages.at(-1)?.id
      try {
        const messages = await useAuthStore().request<ServerMessage[]>(
          `/chats/${id}/messages?limit=50${first ? `&before=${first}` : ''}`,
        )
        const pageSize = messages.length
        // После паузы может прийти больше одной страницы: догружаем пропущенный промежуток.
        if (!older && latest && messages[0] && messages[0].id > Number(latest)) {
          let after = Number(latest)
          while (version === this.version) {
            const page = await useAuthStore().request<ServerMessage[]>(
              `/chats/${id}/messages?limit=50&after=${after}`,
            )
            messages.push(...page)
            if (page.length < 50) break
            after = page[page.length - 1]!.id
          }
        }
        if (version !== this.version) return
        const chat = this.chats.find((item) => item.id === id)
        if (!chat) return
        const merged = new Map(chat.messages.map((item) => [item.id, item]))
        for (const message of messages) merged.set(String(message.id), mapMessage(message))
        chat.messages = [...merged.values()].sort((a, b) => Number(a.id) - Number(b.id))
        if (older || !(id in this.hasOlder)) this.hasOlder[id] = pageSize === 50
        const last = chat.messages.at(-1)
        if (last) chat.lastMessage = last
        if (!older && last && this.activeChatId === id && document.visibilityState === 'visible') {
          await useAuthStore().request<void>(`/chats/${id}/read`, {
            method: 'POST',
            body: JSON.stringify({ message_id: Number(last.id) }),
          })
          const current = this.chats.find((item) => item.id === id)
          if (version === this.version && current) current.unread = 0
        }
      } catch (error) {
        if (version === this.version && this.activeChatId === id)
          this.conversationError =
            error instanceof Error ? error.message : 'Не удалось загрузить сообщения.'
      } finally {
        if (version === this.version) delete this.pendingChatIds[id]
      }
    },
    async send(id: string, text: string) {
      if (this.sending) return
      const version = this.version
      this.sending = true
      this.sendError = ''
      try {
        const message = await useAuthStore().request<ServerMessage>(`/chats/${id}/messages`, {
          method: 'POST',
          body: JSON.stringify({ body: text, file_ids: [] }),
        })
        if (version !== this.version) return
        const chat = this.chats.find((item) => item.id === id)
        if (chat) {
          const saved = mapMessage(message)
          if (!chat.messages.some((item) => item.id === saved.id)) chat.messages.push(saved)
          chat.lastMessage = saved
        }
        if (this.drafts[id]?.trim() === text) delete this.drafts[id]
      } catch (error) {
        if (version === this.version && this.activeChatId === id)
          this.sendError =
            error instanceof Error ? error.message : 'Не удалось отправить сообщение.'
      } finally {
        if (version === this.version) this.sending = false
      }
    },
    clearSession() {
      this.version++
      this.chats = []
      this.drafts = {}
      this.loading = false
      this.pendingChatIds = {}
      this.activeChatId = ''
      this.sending = false
      this.error = ''
      this.conversationError = ''
      this.sendError = ''
      this.hasOlder = {}
    },
  },
})
