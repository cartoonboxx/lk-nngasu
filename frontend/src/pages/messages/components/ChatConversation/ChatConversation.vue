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
    <MessageHistory :chat="chat" :is-saved="$messageStore.isSaved" @retry="$messageStore.retry()" />
    <MessageComposer
      ref="composer"
      v-model="draft"
      :max-height="composerMaxHeight"
      :storage-error="storageError"
      :draft-error="$messageStore.draftError"
      @send="sendMessage"
      @retry="$messageStore.retry()"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './ChatConversation.module.scss'
import { PhArrowLeft } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import MessageHistory from '../MessageHistory/MessageHistory.vue'
import MessageComposer from '../MessageComposer/MessageComposer.vue'
import type { Chat } from '../../messages.types'
export default defineComponent({
  name: 'ChatConversation',
  components: {
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
    storageError(): boolean {
      return this.$messageStore.storageError
    },
    draft: {
      get(): string {
        return this.$messageStore.drafts[this.chat.id] ?? ''
      },
      set(value: string) {
        this.$messageStore.setDraft(this.chat.id, value)
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
    storageError() {
      void this.$nextTick(this.updateComposerLimit)
    },
    '$messageStore.draftError'() {
      void this.$nextTick(this.updateComposerLimit)
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
      this.$messageStore.send(this.chat.id, text)
      void this.$nextTick(() => {
        ;(this.$refs.composer as InstanceType<typeof MessageComposer> | undefined)?.focus()
      })
    },
  },
})
</script>
