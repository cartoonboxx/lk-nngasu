<template>
  <section ref="page" :class="style.page" :style="viewportStyle">
    <p v-if="$messageStore.loading && !chats.length">Загрузка чатов…</p>
    <p v-if="$messageStore.error">
      <button @click="refreshMessages">Повторить</button>
    </p>
    <p v-if="$route.query.chat && !selectedChat && !$messageStore.loading">Чат не найден.</p>
    <div :class="[style.messenger, selectedChat && style.conversationOpen]">
      <ChatSidebar
        ref="sidebar"
        :class="style.sidebar"
        :chats="chats"
        :selected-id="selectedChat?.id"
        @compose="composeOpen = true"
      />
      <ChatConversation
        v-if="selectedChat"
        ref="conversation"
        :class="style.conversation"
        :chat="selectedChat"
        @back="$router.push({ path: '/messages' })"
      />
      <MessagesWelcome v-else :class="style.welcome" @compose="composeOpen = true" />
    </div>
    <NewChatModal v-model="composeOpen" @select="startChat" />
  </section>
</template>

<script lang="ts">
import { defineComponent, type CSSProperties } from 'vue'
import ChatSidebar from './components/ChatSidebar/ChatSidebar.vue'
import ChatConversation from './components/ChatConversation/ChatConversation.vue'
import MessagesWelcome from './components/MessagesWelcome/MessagesWelcome.vue'
import NewChatModal from './components/NewChatModal/NewChatModal.vue'
import type { Chat } from './messages.types'
import style from './MessagesPage.module.scss'
export default defineComponent({
  name: 'MessagesPage',
  components: { ChatSidebar, ChatConversation, MessagesWelcome, NewChatModal },
  data() {
    return {
      style,
      refreshing: false,
      viewportHeight: 0,
      viewportTop: 0,
      viewportWidth: 0,
      unfocusedHeight: 0,
      keyboardOpen: false,
      viewportFrame: 0,
      viewportObserver: null as ResizeObserver | null,
      lastChatId: '',
    }
  },
  computed: {
    chats(): Chat[] {
      return this.$messageStore.chats
    },
    viewportStyle(): CSSProperties {
      return {
        '--messages-viewport-height': this.viewportHeight ? `${this.viewportHeight}px` : undefined,
        '--messages-viewport-top': `${this.viewportTop}px`,
        '--messages-navigation-space': this.keyboardOpen || this.selectedChat ? '0px' : undefined,
      } as CSSProperties
    },
    selectedChat(): Chat | undefined {
      return this.chats.find((chat) => chat.id === this.$route.query.chat)
    },
    composeOpen: {
      get(): boolean {
        return this.$route.query.compose === '1'
      },
      async set(open: boolean) {
        await this.$router.replace({
          query: { ...this.$route.query, compose: open ? '1' : undefined },
        })
      },
    },
  },
  async mounted() {
    this.$messageStore.connect()
    document.documentElement.dataset.messagesPage = 'true'
    // Search in NewChatModal is teleported outside the page element.
    document.addEventListener('focusin', this.scheduleViewportUpdate)
    document.addEventListener('focusout', this.scheduleViewportUpdate)
    window.visualViewport?.addEventListener('resize', this.scheduleViewportUpdate)
    window.visualViewport?.addEventListener('scroll', this.scheduleViewportUpdate)
    window.addEventListener('resize', this.scheduleViewportUpdate)
    this.viewportObserver = new ResizeObserver(this.updateComposerLimit)
    this.viewportObserver.observe(this.$refs.page as HTMLElement)
    this.updateViewport()
    await this.refreshMessages()
  },
  beforeUnmount() {
    this.$messageStore.activeChatId = ''
    cancelAnimationFrame(this.viewportFrame)
    delete document.documentElement.dataset.messagesPage
    document.removeEventListener('focusin', this.scheduleViewportUpdate)
    document.removeEventListener('focusout', this.scheduleViewportUpdate)
    window.visualViewport?.removeEventListener('resize', this.scheduleViewportUpdate)
    window.visualViewport?.removeEventListener('scroll', this.scheduleViewportUpdate)
    window.removeEventListener('resize', this.scheduleViewportUpdate)
    this.viewportObserver?.disconnect()
    delete document.body.dataset.messageKeyboard
  },
  watch: {
    'selectedChat.id': {
      immediate: true,
      async handler(id: string | undefined, previous: string | undefined) {
        this.$messageStore.activeChatId = id ?? ''
        this.$messageStore.conversationError = ''
        this.$messageStore.sendError = ''
        if (id) {
          this.lastChatId = id
          await this.$messageStore.fetchMessages(id)
        }
        await this.$nextTick()
        if (!id && previous)
          (this.$refs.sidebar as InstanceType<typeof ChatSidebar> | undefined)?.restoreFocus(
            this.lastChatId,
          )
        this.updateViewport()
      },
    },
  },
  methods: {
    scheduleViewportUpdate() {
      if (!this.$refs.page || this.viewportFrame) return
      this.viewportFrame = requestAnimationFrame(() => {
        this.viewportFrame = 0
        this.updateViewport()
      })
    },
    updateViewport() {
      const mobile = this.$utilityStore.isMobileVersion()
      const viewport = window.visualViewport
      const height = viewport?.height ?? window.innerHeight
      const width = viewport?.width ?? window.innerWidth
      const active = document.activeElement
      const focused =
        active instanceof HTMLElement && active.matches('input, textarea, [contenteditable="true"]')
      if (!this.viewportWidth || Math.abs(width - this.viewportWidth) > 10) {
        this.unfocusedHeight = height
      } else if (!focused && !this.keyboardOpen) {
        this.unfocusedHeight = Math.max(this.unfocusedHeight, height)
      }
      this.viewportWidth = width
      this.viewportHeight = mobile ? height : 0
      this.viewportTop = mobile ? (viewport?.offsetTop ?? 0) : 0
      this.keyboardOpen =
        mobile &&
        (focused || this.keyboardOpen) &&
        Math.max(this.unfocusedHeight, window.innerHeight) - height > 100
      if (this.keyboardOpen) document.body.dataset.messageKeyboard = 'true'
      else delete document.body.dataset.messageKeyboard
    },
    updateComposerLimit() {
      ;(
        this.$refs.conversation as InstanceType<typeof ChatConversation> | undefined
      )?.updateComposerLimit()
    },
    async refreshMessages() {
      if (this.refreshing) return
      this.refreshing = true
      try {
        await this.$messageStore.fetchChats()
        if (this.selectedChat) await this.$messageStore.fetchMessages(this.selectedChat.id)
      } finally {
        this.refreshing = false
      }
    },
    async startChat(id: string) {
      await this.$router.push({ path: '/messages', query: { chat: id } })
    },
  },
})
</script>
