<template>
  <div
    ref="history"
    :class="style.history"
    role="log"
    :aria-label="`История сообщений: ${chat.name}`"
  >
    <div v-if="!chat.messages.length" :class="style.startConversation">
      <PhChatCircleText :size="32" aria-hidden="true" />
      <p>Начните разговор</p>
      <span>Ваше первое сообщение появится здесь</span>
    </div>
    <template v-for="(message, index) in chat.messages" :key="message.id">
      <div v-if="isNewDay(index)" :class="style.dateDivider">
        <span>{{ dayLabel(message.sentAt) }}</span>
      </div>
      <article
        :class="[style.message, message.own && style.ownMessage]"
        :aria-label="`${message.sender}, ${messageTime(message.sentAt)}`"
      >
        <span v-if="chat.group && !message.own" :class="style.sender">
          {{ message.sender }}
        </span>
        <p>{{ message.text }}</p>
        <div :class="style.messageMeta">
          <time :datetime="message.sentAt">
            {{ messageTime(message.sentAt) }}
          </time>
          <PhCheck
            v-if="message.own && isSaved(chat.id, message)"
            :size="14"
            aria-label="Сохранено"
          />
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
import { messageTime, dayLabel } from '../../messages.format'
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
    return { style }
  },
  mounted() {
    void this.scrollToEnd()
  },
  watch: {
    'chat.id'() {
      void this.scrollToEnd()
    },
    'chat.messages.length'() {
      void this.scrollToEnd()
    },
  },
  methods: {
    messageTime,
    dayLabel,
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
      await this.$nextTick()
      const history = this.$refs.history as HTMLElement | undefined
      if (history) history.scrollTop = history.scrollHeight
    },
  },
})
</script>
