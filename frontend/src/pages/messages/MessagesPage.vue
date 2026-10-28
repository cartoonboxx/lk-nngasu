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
      <ChatSidebar
        ref="sidebar"
        :class="style.sidebar"
        :chats="chats"
        :selected-id="selectedChat?.id"
        @select="openChat"
        @compose="openCompose"
      />
      <ChatConversation
        v-if="selectedChat"
        ref="conversation"
        :class="style.conversation"
        :chat="selectedChat"
        @back="backToChats"
      />
      <MessagesWelcome v-else :class="style.welcome" @compose="openCompose" />
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
import { contacts } from './messages.data'
import type { Chat } from './messages.types'
import breakpoints from '@/styles/breakpoints.module.scss'
import style from './MessagesPage.module.scss'
export default defineComponent({
  name: 'MessagesPage',
  components: { ChatSidebar, ChatConversation, MessagesWelcome, NewChatModal },
  data() {
    return {
      style,
      viewportHeight: 0,
      viewportTop: 0,
      viewportWidth: 0,
      unfocusedHeight: 0,
      keyboardOpen: false,
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
          this.$messageStore.markRead(id)
        }
        void this.$nextTick(() => {
          if (!id && previous)
            (this.$refs.sidebar as InstanceType<typeof ChatSidebar> | undefined)?.restoreFocus(
              this.lastChatId,
            )
          this.updateViewport()
        })
      },
    },
  },
  methods: {
    onFocusOut() {
      void this.$nextTick(this.updateViewport)
    },
    updateViewport() {
      const mobile = window.matchMedia(`(max-width: ${breakpoints.phoneMax})`).matches
      const viewport = window.visualViewport
      const height = viewport?.height ?? window.innerHeight
      const width = viewport?.width ?? window.innerWidth
      const focused =
        (
          this.$refs.conversation as InstanceType<typeof ChatConversation> | undefined
        )?.isComposerFocused() ?? false
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
      ;(
        this.$refs.conversation as InstanceType<typeof ChatConversation> | undefined
      )?.updateComposerLimit()
    },
    openCompose() {
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
        this.$messageStore.ensureChat({ ...contact, group: false })
      }
      this.openChat(id)
    },
  },
})
</script>
