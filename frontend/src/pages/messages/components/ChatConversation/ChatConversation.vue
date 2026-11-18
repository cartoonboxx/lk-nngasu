<template>
  <section :class="style.conversation">
    <header :class="style.conversationHeader">
      <IconButton
        :class="[style.backButton, style.headerAction]"
        variant="ghost"
        size="sm"
        title="Назад к чатам"
        round
        @click="$emit('back')"
      >
        <PhArrowLeft :size="21" />
      </IconButton>
      <ChatAvatar :initials="chat.initials" :group="chat.group" />
      <div :class="style.conversationTitle">
        <h2 ref="conversationTitle" tabindex="-1">{{ chat.name }}</h2>
        <p>{{ chat.description }}</p>
      </div>
    </header>
    <p v-if="$messageStore.messagesLoading">Загрузка сообщений…</p>
    <p v-if="$messageStore.conversationError" role="alert">
      {{ $messageStore.conversationError }}
      <button @click="$messageStore.fetchMessages(chat.id)">Повторить</button>
    </p>
    <Button
      v-if="$messageStore.hasOlder[chat.id]"
      variant="ghost"
      :disabled="$messageStore.messagesLoading"
      @click="$messageStore.fetchMessages(chat.id, true)"
      >Предыдущие сообщения</Button
    >
    <MessageHistory :chat="chat" />
    <MessageComposer
      ref="composer"
      v-model="draft"
      :max-height="composerMaxHeight"
      :error="$messageStore.sendError"
      :sending="$messageStore.sending"
      @send="sendMessage"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './ChatConversation.module.scss'
import { PhArrowLeft } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import MessageHistory from '../MessageHistory/MessageHistory.vue'
import MessageComposer from '../MessageComposer/MessageComposer.vue'
import type { Chat } from '../../messages.types'
export default defineComponent({
  name: 'ChatConversation',
  components: {
    Button,
    PhArrowLeft,
    IconButton,
    ChatAvatar,
    MessageHistory,
    MessageComposer,
  },
  props: {
    chat: {
      type: Object as PropType<Chat>,
      required: true,
    },
  },
  emits: ['back'],
  data() {
    return {
      style,
      composerMaxHeight: 144,
    }
  },
  computed: {
    draft: {
      get(): string {
        return this.$messageStore.drafts[this.chat.id] ?? ''
      },
      set(value: string) {
        this.$messageStore.drafts[this.chat.id] = value
      },
    },
  },
  mounted() {
    ;(this.$refs.conversationTitle as HTMLElement | undefined)?.focus({ preventScroll: true })
    this.updateComposerLimit()
  },
  watch: {
    'chat.id'() {
      void this.$nextTick(() => {
        ;(this.$refs.conversationTitle as HTMLElement | undefined)?.focus({ preventScroll: true })
        this.updateComposerLimit()
      })
    },
  },
  methods: {
    updateComposerLimit() {
      const conversation = this.$el as HTMLElement
      const composer = (this.$refs.composer as InstanceType<typeof MessageComposer> | undefined)
        ?.$el as HTMLElement | undefined
      const header = this.$refs.conversationTitle as HTMLElement | undefined
      if (!composer || !header) return
      const minimum =
        parseFloat(getComputedStyle(conversation).getPropertyValue('--control-height-md')) || 46
      const padding = getComputedStyle(composer)
      const warning = composer.querySelector('[role="alert"]')?.getBoundingClientRect().height ?? 0
      const available =
        conversation.clientHeight -
        (header.closest('header')?.getBoundingClientRect().height ?? 80) -
        parseFloat(padding.paddingTop) -
        parseFloat(padding.paddingBottom) -
        warning -
        32
      this.composerMaxHeight = Math.max(minimum, Math.min(144, available))
    },
    sendMessage() {
      const text = this.draft.trim()
      if (!text || text.length > 4000) return
      void this.$messageStore.send(this.chat.id, text)
      void this.$nextTick(() => {
        ;(this.$refs.composer as InstanceType<typeof MessageComposer> | undefined)?.focus()
      })
    },
  },
})
</script>
