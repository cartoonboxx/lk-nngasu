<template>
  <div ref="history" :class="style.history" role="log" @scroll.passive="onScroll">
    <p v-if="error" role="alert">{{ error }}</p>
    <div v-if="!chat.messages.length" :class="style.startConversation">
      <PhChatCircleText :size="32" />
      <p>Начните разговор</p>
      <span>Ваше первое сообщение появится здесь</span>
    </div>
    <template v-for="(message, index) in chat.messages" :key="message.id">
      <div v-if="isNewDay(index)" :class="style.dateDivider">
        <span>{{ dayLabel(message.sentAt) }}</span>
      </div>
      <article :class="[style.message, message.own && style.ownMessage]">
        <span v-if="chat.group && !message.own" :class="style.sender">
          {{ message.sender }}
        </span>
        <p>{{ message.text }}</p>
        <button v-for="file in message.files" :key="file.id" @click="download(file.id, file.name)">
          {{ file.name }}
        </button>
        <div :class="style.messageMeta">
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
import { defineComponent, type PropType } from 'vue'
import style from './MessageHistory.module.scss'
import { PhChatCircleText, PhCheck } from '@phosphor-icons/vue'
import { useFilesStore } from '@/stores/files'
import type { Chat, Message } from '../../messages.types'
import { messageTimeFormatter, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'MessageHistory',
  components: { PhChatCircleText, PhCheck },
  props: {
    chat: { type: Object as PropType<Chat>, required: true },
  },
  data() {
    return {
      style,
      error: '',
      messageTimeFormatter,
      followingLatest: true,
      historyHeight: 0,
      historyWidth: 0,
      resizeObserver: null as ResizeObserver | null,
    }
  },
  mounted() {
    const history = this.$refs.history as HTMLElement
    this.historyHeight = history.clientHeight
    this.historyWidth = history.clientWidth
    this.resizeObserver = new ResizeObserver(this.onResize)
    this.resizeObserver.observe(history)
    void this.scrollToEnd()
  },
  beforeUnmount() {
    this.resizeObserver?.disconnect()
  },
  watch: {
    'chat.id'() {
      void this.scrollToEnd()
    },
    'chat.messages'(messages: Message[], previous: Message[]) {
      const history = this.$refs.history as HTMLElement | undefined
      const before = history?.scrollHeight ?? 0
      if (previous[0] && messages[0] && Number(messages[0].id) < Number(previous[0].id)) {
        void this.$nextTick(() => {
          if (history) history.scrollTop += history.scrollHeight - before
        })
      } else if (this.followingLatest || messages.at(-1)?.own) void this.scrollToEnd()
    },
  },
  methods: {
    async download(id: number, name: string) {
      this.error = ''
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось скачать файл.'
      }
    },
    dayLabel,
    onScroll() {
      const history = this.$refs.history as HTMLElement
      // WebKit can emit scroll before ResizeObserver when the composer changes height.
      if (history.clientHeight !== this.historyHeight || history.clientWidth !== this.historyWidth)
        return
      this.followingLatest = history.scrollHeight - history.clientHeight - history.scrollTop <= 48
    },
    onResize() {
      const history = this.$refs.history as HTMLElement
      this.historyHeight = history.clientHeight
      this.historyWidth = history.clientWidth
      // Keep the bottom anchored when the keyboard or the growing composer changes the viewport.
      if (this.followingLatest) void this.scrollToEnd()
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
