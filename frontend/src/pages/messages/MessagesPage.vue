<template>
  <section
    ref="page"
    :class="style.page"
    :style="viewportStyle"
    aria-label="Сообщения"
    @focusin="updateViewport"
    @focusout="onFocusOut"
  >
    <div :class="[style.messenger, selectedChat && style.conversationOpen]">
      <aside :class="style.sidebar">
        <header :class="style.sidebarHeader">
          <h2 ref="sidebarTitle" tabindex="-1">
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
        <ul
          v-if="filteredChats.length"
          ref="chatList"
          :class="style.chatList"
          aria-label="Список диалогов"
        >
          <li v-for="chat in filteredChats" :key="chat.id">
            <StyleTab
              :class="style.chatRow"
              :active="selectedChat?.id === chat.id"
              :trigger-attrs="{
                'aria-current': selectedChat?.id === chat.id ? 'true' : undefined,
                'data-chat-id': chat.id,
              }"
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
            <h2 ref="conversationTitle" tabindex="-1">{{ selectedChat.name }}</h2>
            <p>{{ selectedChat.description }}</p>
          </div>
        </header>
        <div
          ref="history"
          :class="style.history"
          role="log"
          :aria-label="`История сообщений: ${selectedChat.name}`"
        >
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
                <PhCheck
                  v-if="message.own && messageStore.isSaved(selectedChat.id, message)"
                  :size="14"
                  aria-label="Сохранено"
                />
                <Button
                  v-else-if="message.own"
                  variant="ghost"
                  size="sm"
                  @click="messageStore.retry()"
                  >Повторить сохранение</Button
                >
              </div>
            </article>
          </template>
        </div>
        <form ref="composer" :class="style.composer" @submit.prevent="sendMessage">
          <div
            v-if="storageError || messageStore.state.draftError"
            :class="style.storageError"
            role="alert"
          >
            <span>{{
              storageError
                ? 'Не удалось сохранить изменения. Не перезагружайте страницу.'
                : 'Не удалось сохранить черновик. Он доступен до перезагрузки страницы.'
            }}</span>
            <Button v-if="storageError" variant="ghost" size="sm" @click="messageStore.retry()"
              >Повторить</Button
            >
          </div>
          <div :class="style.composerRow">
            <TextInput
              ref="messageField"
              v-model="draft"
              :class="style.composerInput"
              :maxlength="4000"
              base-height="var(--control-height-md)"
              :max-height="composerMaxHeight"
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
              :disabled="!draft.trim()"
              @pointerdown="onSendPointerDown"
              @pointerup="onSendPointerUp"
              @pointercancel="sendPointer = null"
              @pointerleave="sendPointer = null"
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
import { defineComponent, type CSSProperties } from 'vue'
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
import { contacts, type Chat } from './messages.data'
import { useMessagesStore } from '@/stores/messages'
import breakpoints from '@/styles/breakpoints.module.scss'
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
      messageStore: useMessagesStore(),
      viewportHeight: 0,
      viewportTop: 0,
      viewportWidth: 0,
      unfocusedHeight: 0,
      keyboardOpen: false,
      composerMaxHeight: 144,
      viewportObserver: null as ResizeObserver | null,
      lastChatId: '',
      sendPointer: null as { id: number; x: number; y: number } | null,
      search: '',
      unreadOnly: false,
      contactSearch: '',
    }
  },
  computed: {
    chats(): Chat[] {
      return this.messageStore.state.chats
    },
    storageError(): boolean {
      return this.messageStore.storageError.value
    },
    viewportStyle(): CSSProperties {
      return {
        '--messages-viewport-height': this.viewportHeight ? `${this.viewportHeight}px` : undefined,
        '--messages-viewport-top': `${this.viewportTop}px`,
        '--messages-navigation-space': this.keyboardOpen ? '0px' : undefined,
      } as CSSProperties
    },
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
        return this.selectedChat ? (this.messageStore.state.drafts[this.selectedChat.id] ?? '') : ''
      },
      set(value: string) {
        if (this.selectedChat) this.messageStore.setDraft(this.selectedChat.id, value)
      },
    },
    unreadChats(): number {
      return this.chats.filter((chat) => chat.unread > 0).length
    },
    filteredChats(): Chat[] {
      return this.chats
        .filter(
          (chat) =>
            (!this.unreadOnly || chat.unread > 0) &&
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
  mounted() {
    window.visualViewport?.addEventListener('resize', this.updateViewport)
    window.visualViewport?.addEventListener('scroll', this.updateViewport)
    window.addEventListener('resize', this.updateViewport)
    this.viewportObserver = new ResizeObserver(this.updateComposerLimit)
    this.viewportObserver.observe(this.$refs.page as HTMLElement)
    this.updateViewport()
  },
  beforeUnmount() {
    window.visualViewport?.removeEventListener('resize', this.updateViewport)
    window.visualViewport?.removeEventListener('scroll', this.updateViewport)
    window.removeEventListener('resize', this.updateViewport)
    this.viewportObserver?.disconnect()
    delete document.body.dataset.messageKeyboard
  },
  watch: {
    'selectedChat.id': {
      immediate: true,
      handler(id: string | undefined, previous: string | undefined) {
        if (id) {
          this.lastChatId = id
          this.messageStore.markRead(id)
        }
        void this.scrollToEnd()
        void this.$nextTick(() => {
          if (id)
            (this.$refs.conversationTitle as HTMLElement | undefined)?.focus({
              preventScroll: true,
            })
          else if (previous) this.restoreChatFocus()
          this.updateViewport()
        })
      },
    },
    storageError() {
      void this.$nextTick(this.updateComposerLimit)
    },
    'selectedChat.messages.length'() {
      void this.scrollToEnd()
    },
  },
  methods: {
    onSendPointerDown(event: PointerEvent) {
      if (event.pointerType !== 'mouse') {
        this.sendPointer = { id: event.pointerId, x: event.clientX, y: event.clientY }
      }
      const field = this.$refs.messageField as InstanceType<typeof TextInput> | undefined
      if (field?.$el.contains(document.activeElement)) event.preventDefault()
    },
    onSendPointerUp(event: PointerEvent) {
      const start = this.sendPointer
      this.sendPointer = null
      if (
        !start ||
        start.id !== event.pointerId ||
        Math.hypot(event.clientX - start.x, event.clientY - start.y) > 10
      )
        return
      const bounds = (event.currentTarget as HTMLElement).getBoundingClientRect()
      if (
        event.clientX < bounds.left ||
        event.clientX > bounds.right ||
        event.clientY < bounds.top ||
        event.clientY > bounds.bottom
      )
        return
      // Submit before a touch-generated blur can resize the keyboard or move the button.
      event.preventDefault()
      this.sendMessage()
    },
    onFocusOut() {
      void this.$nextTick(this.updateViewport)
    },
    updateViewport() {
      const mobile = window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      const viewport = window.visualViewport
      const height = viewport?.height ?? window.innerHeight
      const width = viewport?.width ?? window.innerWidth
      const focused =
        (this.$refs.messageField as InstanceType<typeof TextInput> | undefined)?.$el.contains(
          document.activeElement,
        ) ?? false
      if (!focused || Math.abs(width - this.viewportWidth) > 10) this.unfocusedHeight = height
      this.viewportWidth = width
      this.viewportHeight = mobile ? height : 0
      this.viewportTop = mobile ? (viewport?.offsetTop ?? 0) : 0
      this.keyboardOpen =
        mobile && focused && Math.max(this.unfocusedHeight, window.innerHeight) - height > 100
      if (this.keyboardOpen) document.body.dataset.messageKeyboard = 'true'
      else delete document.body.dataset.messageKeyboard
      void this.$nextTick(this.updateComposerLimit)
    },
    updateComposerLimit() {
      const page = this.$refs.page as HTMLElement | undefined
      const composer = this.$refs.composer as HTMLElement | undefined
      const header = this.$refs.conversationTitle as HTMLElement | undefined
      if (!page || !composer || !header) return
      const minimum =
        parseFloat(getComputedStyle(page).getPropertyValue('--control-height-md')) || 46
      const padding = getComputedStyle(composer)
      const warning = composer.querySelector('[role="alert"]')?.getBoundingClientRect().height ?? 0
      const available =
        page.clientHeight -
        (header.closest('header')?.getBoundingClientRect().height ?? 80) -
        parseFloat(padding.paddingTop) -
        parseFloat(padding.paddingBottom) -
        warning -
        32
      this.composerMaxHeight = Math.max(minimum, Math.min(144, available))
    },
    restoreChatFocus() {
      const list = this.$refs.chatList as HTMLElement | undefined
      const target = [...(list?.querySelectorAll<HTMLElement>('[data-chat-id]') ?? [])].find(
        (item) => item.dataset.chatId === this.lastChatId,
      )
      ;(target ?? (this.$refs.sidebarTitle as HTMLElement | undefined))?.focus({
        preventScroll: true,
      })
    },
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
    async scrollToEnd() {
      await this.$nextTick()
      const history = this.$refs.history as HTMLElement | undefined
      if (history) history.scrollTop = history.scrollHeight
    },
    sendMessage() {
      const chat = this.selectedChat
      const text = this.draft.trim()
      if (!chat || !text || text.length > 4000) return
      this.messageStore.send(chat.id, text)
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
        this.messageStore.ensureChat({ ...contact, group: false })
      }
      this.openChat(id)
    },
  },
})
</script>
