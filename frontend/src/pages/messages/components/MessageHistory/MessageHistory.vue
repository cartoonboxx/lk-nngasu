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
      <article :class="[style.message, message.own && style.ownMessage]">
        <span v-if="chat.group && !message.own" :class="style.sender">
          {{ message.sender }}
        </span>
        <p>{{ message.text }}</p>
        <div :class="style.messageMeta">
          <time :datetime="message.sentAt">
            {{ messageTimeFormatter.format(new Date(message.sentAt)) }}
          </time>
          <PhCheck v-if="message.own && isSaved(chat.id, message)" :size="14" title="Сохранено" />
          <Button v-else-if="message.own" variant="ghost" size="sm" @click="$emit('retry')"
            >Повторить сохранение</Button
          >
        </div>
      </article>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './MessageHistory.module.scss'
import { PhChatCircleText, PhCheck } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import type { Chat, MessageSavedChecker } from '../../messages.types'
import { messageTimeFormatter, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'MessageHistory',
  components: { PhChatCircleText, PhCheck, Button },
  props: {
    chat: { type: Object as PropType<Chat>, required: true },
    isSaved: {
      type: Function as PropType<MessageSavedChecker>,
      required: true,
    },
  },
  emits: ['retry'],
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
    'chat.messages.length'() {
      const lastMessage = this.chat.messages[this.chat.messages.length - 1]
      if (this.followingLatest || lastMessage?.own) void this.scrollToEnd()
    },
  },
  methods: {
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
