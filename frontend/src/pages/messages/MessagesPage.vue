<template>
  <section :class="style.page" aria-label="Сообщения">
    <div :class="[style.messenger, selectedChat && style.conversationOpen]">
      <aside :class="style.sidebar">
        <header :class="style.sidebarHeader">
          <h2>
            Чаты
            <span :class="style.chatCount">{{ chats.length }}</span>
          </h2>
          <IconButton
            variant="ghost"
            size="sm"
            round
            :class="style.headerAction"
            aria-label="Новый диалог"
            title="Новый диалог"
            @click="openCompose"
          >
            <PhNotePencil :size="21" />
          </IconButton>
        </header>
        <div :class="style.search">
          <Input
            v-model="search"
            type="search"
            aria-label="Поиск чатов"
            placeholder="Поиск по чатам"
            full-width
          >
            <template #prefix><PhMagnifyingGlass :size="18" aria-hidden="true" /></template>
          </Input>
        </div>
        <div :class="style.filters" role="group" aria-label="Фильтр чатов">
          <Button
            size="sm"
            round
            variant="ghost"
            :aria-pressed="!unreadOnly"
            @click="unreadOnly = false"
          >
            Все
          </Button>
          <Button
            size="sm"
            round
            variant="ghost"
            :aria-pressed="unreadOnly"
            @click="unreadOnly = true"
          >
            Непрочитанные
            <Badge v-if="unreadChats" :value="unreadChats" inline />
          </Button>
        </div>
        <ul v-if="filteredChats.length" :class="style.chatList" aria-label="Список диалогов">
          <li v-for="chat in filteredChats" :key="chat.id">
            <StyleTab
              :class="style.chatRow"
              :active="selectedChat?.id === chat.id"
              :trigger-attrs="{ 'aria-current': selectedChat?.id === chat.id ? 'true' : undefined }"
              @select="openChat(chat.id)"
            >
              <span :class="style.chatContent">
                <span :class="style.avatar" aria-hidden="true">
                  <PhUsersThree v-if="chat.group" :size="22" />
                  <template v-else>
                    {{ chat.initials }}
                  </template>
                </span>
                <span :class="style.chatCopy">
                  <span :class="style.chatLine">
                    <span :class="[style.chatName, chat.unread && style.unreadName]">
                      {{ chat.name }}
                    </span>
                    <time v-if="lastMessage(chat)" :class="style.chatTime">
                      {{ listTime(lastMessage(chat)!.sentAt) }}
                    </time>
                  </span>
                  <span :class="style.chatLine">
                    <span :class="style.preview">{{ preview(chat) }}</span>
                    <Badge
                      v-if="chat.unread"
                      :value="chat.unread"
                      :aria-label="`Непрочитанных сообщений: ${chat.unread}`"
                      inline
                    />
                  </span>
                </span>
              </span>
            </StyleTab>
          </li>
        </ul>
        <div v-else :class="style.listEmpty">
          <PhMagnifyingGlass :size="25" aria-hidden="true" />
          <p>{{ search ? 'Чаты не найдены' : 'Всё прочитано' }}</p>
          <span>
            {{ search ? 'Попробуйте другое имя или название' : 'Новые сообщения появятся здесь' }}
          </span>
        </div>
        <p :class="style.localNote">Переписки сохраняются в этом браузере</p>
      </aside>

      <section v-if="selectedChat" :class="style.conversation">
        <header :class="style.conversationHeader">
          <IconButton
            :class="[style.backButton, style.headerAction]"
            variant="ghost"
            size="sm"
            aria-label="Назад к чатам"
            round
            @click="backToChats"
          >
            <PhArrowLeft :size="21" />
          </IconButton>
          <span :class="style.avatar" aria-hidden="true">
            <PhUsersThree v-if="selectedChat.group" :size="22" />
            <template v-else>
              {{ selectedChat.initials }}
            </template>
          </span>
          <div :class="style.conversationTitle">
            <h2>{{ selectedChat.name }}</h2>
            <p>{{ selectedChat.description }}</p>
          </div>
        </header>
        <div ref="history" :class="style.history" role="log">
          <div v-if="!selectedChat.messages.length" :class="style.startConversation">
            <PhChatCircleText :size="32" aria-hidden="true" />
            <p>Начните разговор</p>
            <span>Ваше первое сообщение появится здесь</span>
          </div>
          <template v-for="(message, index) in selectedChat.messages" :key="message.id">
            <div v-if="isNewDay(index)" :class="style.dateDivider">
              <span>{{ dayLabel(message.sentAt) }}</span>
            </div>
            <article
              :class="[style.message, message.own && style.ownMessage]"
              :aria-label="`${message.sender}, ${messageTime(message.sentAt)}`"
            >
              <span v-if="selectedChat.group && !message.own" :class="style.sender">
                {{ message.sender }}
              </span>
              <p>{{ message.text }}</p>
              <div :class="style.messageMeta">
                <time :datetime="message.sentAt">
                  {{ messageTime(message.sentAt) }}
                </time>
                <PhCheck v-if="message.own" :size="14" aria-label="Сохранено" />
              </div>
            </article>
          </template>
        </div>
        <form :class="style.composer" @submit.prevent="sendMessage">
          <div :class="style.composerRow">
            <TextInput
              ref="messageField"
              v-model="draft"
              :class="style.composerInput"
              :maxlength="4000"
              base-height="var(--control-height-md)"
              :max-height="144"
              :error="
                storageError
                  ? 'Не удалось сохранить переписку в браузере. Сообщения доступны до перезагрузки страницы.'
                  : ''
              "
              aria-label="Сообщение"
              placeholder="Напишите сообщение…"
              multiline
              full-width
              @keydown="onComposerKeydown"
            />
            <IconButton
              type="submit"
              variant="primary"
              round
              v-if="draft.trim()"
              aria-label="Отправить сообщение"
              title="Отправить сообщение"
            >
              <PhPaperPlaneRight :size="21" weight="bold" />
            </IconButton>
          </div>
          <span :class="style.composerHint">Enter — отправить · Shift + Enter — новая строка</span>
        </form>
      </section>
      <div v-else :class="style.welcome">
        <span :class="style.welcomeIcon">
          <PhChatsCircle :size="36" weight="light" aria-hidden="true" />
        </span>
        <h2>Хороший разговор начинается здесь</h2>
        <p>Выберите чат слева или напишите<br />кому-нибудь первым.</p>
        <Button variant="secondary" @click="openCompose">
          <PhNotePencil :size="18" aria-hidden="true" />
          Новый диалог
        </Button>
      </div>
    </div>

    <Modal v-model="composeOpen" :width="440" aria-labelledby="new-chat-title">
      <header :class="style.modalHeader">
        <div>
          <h2 id="new-chat-title">Новый диалог</h2>
          <p>Кому хотите написать?</p>
        </div>
        <IconButton
          variant="ghost"
          size="sm"
          :class="style.headerAction"
          aria-label="Закрыть"
          round
          @click="composeOpen = false"
        >
          <PhX :size="20" />
        </IconButton>
      </header>
      <Input
        v-model="contactSearch"
        type="search"
        aria-label="Поиск собеседника"
        placeholder="Имя или фамилия"
        full-width
      >
        <template #prefix>
          <PhMagnifyingGlass :size="18" aria-hidden="true" />
        </template>
      </Input>
      <ul :class="style.contacts" aria-label="Собеседники">
        <li v-for="contact in filteredContacts" :key="contact.id">
          <StyleTab :class="style.contact" @select="startChat(contact.id)">
            <span :class="style.contactContent">
              <span :class="style.avatar" aria-hidden="true">{{ contact.initials }}</span>
              <span :class="style.contactCopy">
                <strong>{{ contact.name }}</strong>
                <span>{{ contact.description }}</span>
              </span>
            </span>
            <template #trailing>
              <PhArrowUpRight :size="17" aria-hidden="true" />
            </template>
          </StyleTab>
        </li>
      </ul>
      <p v-if="!filteredContacts.length" :class="style.contactsEmpty">
        Никого не нашли. Попробуйте другое имя.
      </p>
    </Modal>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import {
  PhNotePencil,
  PhMagnifyingGlass,
  PhUsersThree,
  PhArrowLeft,
  PhArrowUpRight,
  PhCheck,
  PhChatsCircle,
  PhChatCircleText,
  PhX,
  PhPaperPlaneRight,
} from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import Badge from '@/components/UI/Badge/Badge.vue'
import { contacts, loadChats, saveChats, type Chat } from './messages.data'
import style from './MessagesPage.module.scss'

function lastMessage(chat: Chat) {
  return chat.messages[chat.messages.length - 1]
}
function messageTime(date: string): string {
  return new Date(date).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
}
function dayLabel(date: string): string {
  const day = new Date(date).toDateString()
  const now = new Date()
  if (day === now.toDateString()) return 'Сегодня'
  now.setDate(now.getDate() - 1)
  if (day === now.toDateString()) return 'Вчера'
  return new Date(date).toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' })
}

export default defineComponent({
  name: 'MessagesPage',
  components: {
    PhPaperPlaneRight,
    Input,
    TextInput,
    StyleTab,
    Badge,
    IconButton,
    Button,
    Modal,
    PhNotePencil,
    PhMagnifyingGlass,
    PhUsersThree,
    PhArrowLeft,
    PhArrowUpRight,
    PhCheck,
    PhChatsCircle,
    PhChatCircleText,
    PhX,
  },
  data() {
    return {
      style,
      chats: loadChats(),
      search: '',
      unreadOnly: false,
      contactSearch: '',
      storageError: false,
      drafts: {} as Record<string, string>,
    }
  },
  computed: {
    selectedChat(): Chat | undefined {
      return this.chats.find((chat) => chat.id === this.$route.query.chat)
    },
    composeOpen: {
      get(): boolean {
        return this.$route.query.compose === '1'
      },
      set(open: boolean) {
        void this.$router.replace({
          query: { ...this.$route.query, compose: open ? '1' : undefined },
        })
      },
    },
    draft: {
      get(): string {
        return this.selectedChat ? (this.drafts[this.selectedChat.id] ?? '') : ''
      },
      set(value: string) {
        if (this.selectedChat) this.drafts[this.selectedChat.id] = value
      },
    },
    unreadChats(): number {
      return this.chats.filter((chat) => chat.unread > 0).length
    },
    filteredChats(): Chat[] {
      return this.chats
        .filter(
          (chat) =>
            (!this.unreadOnly || chat.unread > 0 || chat.id === this.selectedChat?.id) &&
            chat.name.toLocaleLowerCase('ru').includes(this.search.trim().toLocaleLowerCase('ru')),
        )
        .sort((a, b) => this.chatTimestamp(b) - this.chatTimestamp(a))
    },
    filteredContacts(): typeof contacts {
      return contacts.filter((contact) =>
        contact.name
          .toLocaleLowerCase('ru')
          .includes(this.contactSearch.trim().toLocaleLowerCase('ru')),
      )
    },
  },
  watch: {
    'selectedChat.id': {
      immediate: true,
      handler() {
        if (this.selectedChat?.unread) {
          this.selectedChat.unread = 0
          this.persist()
        }
        void this.scrollToEnd()
      },
    },
  },
  methods: {
    lastMessage,
    messageTime,
    dayLabel,
    chatTimestamp(chat: Chat): number {
      const message = lastMessage(chat)
      return message ? Date.parse(message.sentAt) : Number.MAX_SAFE_INTEGER
    },
    preview(chat: Chat): string {
      const message = lastMessage(chat)
      if (!message) return 'Начните разговор'
      return `${message.own ? 'Вы: ' : chat.group ? `${message.sender.split(' ')[0]}: ` : ''}${message.text}`
    },
    listTime(date: string): string {
      return new Date(date).toDateString() === new Date().toDateString()
        ? messageTime(date)
        : dayLabel(date)
    },
    isNewDay(index: number): boolean {
      const messages = this.selectedChat?.messages
      const current = messages?.[index]
      const previous = messages?.[index - 1]
      return (
        !!current &&
        (!previous ||
          new Date(current.sentAt).toDateString() !== new Date(previous.sentAt).toDateString())
      )
    },
    persist() {
      this.storageError = !saveChats(this.chats)
    },
    async scrollToEnd() {
      await this.$nextTick()
      const history = this.$refs.history as HTMLElement | undefined
      if (history) history.scrollTop = history.scrollHeight
    },
    sendMessage() {
      const chat = this.selectedChat
      const text = this.draft.trim()
      if (!chat || !text || text.length > 4000) return
      chat.messages.push({
        id: crypto.randomUUID(),
        text,
        sender: 'Вы',
        own: true,
        sentAt: new Date().toISOString(),
      })
      this.draft = ''
      this.persist()
      void this.scrollToEnd()
      void this.$nextTick(() => {
        ;(this.$refs.messageField as InstanceType<typeof TextInput> | undefined)?.focus()
      })
    },
    onComposerKeydown(event: KeyboardEvent) {
      if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
        event.preventDefault()
        this.sendMessage()
      }
    },
    openCompose() {
      this.contactSearch = ''
      this.composeOpen = true
    },
    backToChats() {
      void this.$router.push({ path: '/messages' })
    },
    openChat(id: string) {
      void this.$router.push({ path: '/messages', query: { chat: id } })
    },
    startChat(id: string) {
      const contact = contacts.find((item) => item.id === id)
      if (!contact) return
      if (!this.chats.some((chat) => chat.id === id)) {
        this.chats.unshift({ ...contact, group: false, unread: 0, messages: [] })
        this.persist()
      }
      this.openChat(id)
    },
  },
})
</script>
