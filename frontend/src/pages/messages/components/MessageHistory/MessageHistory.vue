<template>
  <div ref="history" :class="style.history" role="log" @scroll.passive="onScroll">
    <div v-if="!chat.messages.length" :class="style.startConversation">
      <PhChatCircleText :size="32" />
      <p>Начните разговор</p>
      <span>Ваше первое сообщение появится здесь</span>
    </div>
    <template v-for="(message, index) in chat.messages" :key="message.id">
      <div v-if="isNewDay(index)" :class="style.dateDivider">
        <span>{{ dayLabel(message.sentAt) }}</span>
      </div>
      <article
        :class="[style.message, message.own && style.ownMessage]"
        :data-message-id="message.own ? undefined : message.id"
      >
        <span v-if="chat.group && !message.own" :class="style.sender">
          {{ message.sender }}
        </span>
        <p>{{ message.text }}</p>
        <MediaButton
          v-for="file in message.files"
          :key="file.id"
          @click="download(file.id, file.name)"
          :title="file.name"
          :filename="file.name"
        />
        <div :class="style.messageMeta">
          <span v-if="message.editedAt">изменено</span>
          <time :datetime="message.sentAt">
            {{ messageTimeFormatter.format(new Date(message.sentAt)) }}
          </time>
          <PhCheck v-if="message.own" :size="14" title="Сохранено на сервере" />
        </div>
      </article>
    </template>
  </div>
</template>

<script lang="ts">
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import { apiNotifications } from '@/api/client'
import { defineComponent, type PropType } from 'vue'
import style from './MessageHistory.module.scss'
import { PhChatCircleText, PhCheck } from '@phosphor-icons/vue'
import { useFilesStore } from '@/stores/files'
import type { Chat, Message } from '../../messages.types'
import { messageTimeFormatter, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'MessageHistory',
  components: { MediaButton, PhChatCircleText, PhCheck },
  props: {
    chat: { type: Object as PropType<Chat>, required: true },
  },
  data() {
    return {
      style,
      messageTimeFormatter,
      followingLatest: true,
      historyHeight: 0,
      historyWidth: 0,
      resizeObserver: null as ResizeObserver | null,
    }
  },
  async mounted() {
    const history = this.$refs.history as HTMLElement
    this.historyHeight = history.clientHeight
    this.historyWidth = history.clientWidth
    this.resizeObserver = new ResizeObserver(this.onResize)
    this.resizeObserver.observe(history)
    document.addEventListener('visibilitychange', this.markVisibleMessages)
    document.addEventListener('focusin', this.markVisibleMessages)
    document.addEventListener('close', this.markVisibleMessages, true)
    window.addEventListener('focus', this.markVisibleMessages)
    await this.scrollToEnd()
    this.markVisibleMessages()
  },
  beforeUnmount() {
    this.resizeObserver?.disconnect()
    document.removeEventListener('visibilitychange', this.markVisibleMessages)
    document.removeEventListener('focusin', this.markVisibleMessages)
    document.removeEventListener('close', this.markVisibleMessages, true)
    window.removeEventListener('focus', this.markVisibleMessages)
  },
  watch: {
    async 'chat.id'() {
      await this.scrollToEnd()
      this.markVisibleMessages()
    },
    async 'chat.messages'(messages: Message[], previous: Message[]) {
      const history = this.$refs.history as HTMLElement | undefined
      const before = history?.scrollHeight ?? 0
      if (previous[0] && messages[0] && Number(messages[0].id) < Number(previous[0].id)) {
        await this.$nextTick()
        if (history) history.scrollTop += history.scrollHeight - before
      } else if (this.followingLatest || messages.at(-1)?.own) await this.scrollToEnd()
      else await this.$nextTick()
      this.markVisibleMessages()
    },
  },
  methods: {
    async download(id: number, name: string) {
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        apiNotifications.onError?.(error)
      }
    },
    dayLabel,
    onScroll() {
      const history = this.$refs.history as HTMLElement
      // WebKit can emit scroll before ResizeObserver when the composer changes height.
      if (history.clientHeight !== this.historyHeight || history.clientWidth !== this.historyWidth)
        return
      this.followingLatest = history.scrollHeight - history.clientHeight - history.scrollTop <= 48
      this.markVisibleMessages()
    },
    async onResize() {
      const history = this.$refs.history as HTMLElement
      this.historyHeight = history.clientHeight
      this.historyWidth = history.clientWidth
      // Keep the bottom anchored when the keyboard or the growing composer changes the viewport.
      if (this.followingLatest) await this.scrollToEnd()
      this.markVisibleMessages()
    },
    markVisibleMessages() {
      if (
        this.$messageStore.activeChatId !== this.chat.id ||
        document.visibilityState !== 'visible' ||
        !document.hasFocus()
      )
        return
      const history = this.$refs.history as HTMLElement | undefined
      if (!history) return
      const bounds = history.getBoundingClientRect()
      const viewport = window.visualViewport
      const top = Math.max(bounds.top, viewport?.offsetTop ?? 0)
      const bottom = Math.min(
        bounds.bottom,
        (viewport?.offsetTop ?? 0) + (viewport?.height ?? innerHeight),
      )
      const left = Math.max(bounds.left, viewport?.offsetLeft ?? 0)
      const right = Math.min(
        bounds.right,
        (viewport?.offsetLeft ?? 0) + (viewport?.width ?? innerWidth),
      )
      const elements = history.querySelectorAll<HTMLElement>('[data-message-id]')
      for (let index = elements.length - 1; index >= 0; index--) {
        const element = elements[index]!
        const rect = element.getBoundingClientRect()
        const visibleTop = Math.max(rect.top, top)
        const visibleBottom = Math.min(rect.bottom, bottom)
        const visibleLeft = Math.max(rect.left, left)
        const visibleRight = Math.min(rect.right, right)
        if (
          visibleBottom - visibleTop < Math.min(32, rect.height / 2) ||
          visibleRight <= visibleLeft
        )
          continue
        const visibleElement = document.elementFromPoint(
          (visibleLeft + visibleRight) / 2,
          (visibleTop + visibleBottom) / 2,
        )
        if (!visibleElement || !element.contains(visibleElement)) continue
        const messageId = Number(element.dataset.messageId)
        if (messageId > this.chat.lastReadMessageId)
          void this.$messageStore.markRead(this.chat.id, messageId)
        break
      }
    },
    isNewDay(index: number): boolean {
      const messages = this.chat.messages
      const current = messages?.[index]
      const previous = messages?.[index - 1]
      return (
        !!current &&
        (!previous ||
          new Date(current.sentAt).toDateString() !== new Date(previous.sentAt).toDateString())
      )
    },
    async scrollToEnd() {
      this.followingLatest = true
      await this.$nextTick()
      const history = this.$refs.history as HTMLElement | undefined
      if (history) history.scrollTop = history.scrollHeight
    },
  },
})
</script>
