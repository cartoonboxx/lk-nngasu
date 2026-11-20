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
      <div
        :class="[style.headerInfo, chat.group && style.groupHeader]"
        :title="chat.group ? 'Участники группы' : undefined"
        @click="chat.group && (participantsOpen = true)"
      >
        <ChatAvatar :initials="chat.initials" :group="chat.group" />
        <span :class="style.conversationTitle">
          <span ref="conversationTitle" :class="style.chatName" tabindex="-1">{{ chat.name }}</span>
          <span :class="style.chatDescription">
            {{ chat.description }} ·
            {{
              $messageStore.connectionState === 'connected'
                ? 'На связи'
                : $messageStore.connectionState === 'connecting'
                  ? 'Подключение…'
                  : 'Восстанавливаем соединение…'
            }}
          </span>
        </span>
        <PhCaretDown v-if="chat.group" :size="16" :class="style.participantsIndicator" />
      </div>
    </header>
    <p v-if="$messageStore.messagesLoading">Загрузка сообщений…</p>
    <p v-if="$messageStore.conversationError">
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
    <ul v-if="attachments.length" :class="style.attachments">
      <li v-for="(attachment, index) in attachments" :key="index">
        {{ attachment.file.name }}
        <button
          :disabled="uploading || $messageStore.sending"
          @click="attachments.splice(index, 1)"
        >
          Убрать
        </button>
      </li>
    </ul>
    <ChatParticipantsModal
      v-model="participantsOpen"
      :chat-id="chat.id"
      :chat-name="chat.name"
      @select="$router.push({ path: '/messages', query: { chat: $event } })"
    />
    <MessageComposer
      ref="composer"
      v-model="draft"
      :max-height="composerMaxHeight"
      :sending="$messageStore.sending || uploading"
      :connected="$messageStore.connectionState === 'connected'"
      :has-attachments="attachments.length > 0"
      @attach="addFiles"
      @send="sendMessage"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './ChatConversation.module.scss'
import { PhArrowLeft, PhCaretDown } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import ChatParticipantsModal from '../ChatParticipantsModal/ChatParticipantsModal.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import MessageHistory from '../MessageHistory/MessageHistory.vue'
import MessageComposer from '../MessageComposer/MessageComposer.vue'
import type { Chat } from '../../messages.types'
import { useFilesStore } from '@/stores/files'
import { useAuthStore } from '@/stores/auth'
import { apiNotifications } from '@/api/client'
export default defineComponent({
  name: 'ChatConversation',
  components: {
    Button,
    PhArrowLeft,
    PhCaretDown,
    ChatParticipantsModal,
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
      uploading: false,
      participantsOpen: false,
      fileDrafts: {} as Record<string, { file: File; id?: number }[]>,
    }
  },
  computed: {
    attachments() {
      return this.fileDrafts[this.chat.id] ?? []
    },
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
    async 'chat.id'() {
      this.participantsOpen = false
      await this.$nextTick()
      ;(this.$refs.conversationTitle as HTMLElement | undefined)?.focus({ preventScroll: true })
      this.updateComposerLimit()
    },
  },
  methods: {
    addFiles(files: File[]) {
      if (this.uploading || this.$messageStore.sending) return
      this.fileDrafts[this.chat.id] ??= []
      for (const file of files) {
        if (
          this.attachments.length >= 5 ||
          file.size > 15 * 1024 * 1024 ||
          !/\.(pdf|docx|pptx|xlsx|xls|zip|rar)$/i.test(file.name)
        ) {
          this.$utilityStore.notifyError(
            'До 5 файлов: PDF, DOCX, PPTX, XLSX, XLS, ZIP, RAR, не более 15 МБ каждый.',
          )
          continue
        }
        this.fileDrafts[this.chat.id]!.push({ file })
      }
    },
    updateComposerLimit() {
      const conversation = this.$el as HTMLElement
      const composer = (this.$refs.composer as InstanceType<typeof MessageComposer> | undefined)
        ?.$el as HTMLElement | undefined
      const header = this.$refs.conversationTitle as HTMLElement | undefined
      if (!composer || !header) return
      const minimum =
        parseFloat(getComputedStyle(conversation).getPropertyValue('--control-height-md')) || 46
      const padding = getComputedStyle(composer)
      const available =
        conversation.clientHeight -
        (header.closest('header')?.getBoundingClientRect().height ?? 80) -
        parseFloat(padding.paddingTop) -
        parseFloat(padding.paddingBottom) -
        32
      this.composerMaxHeight = Math.max(minimum, Math.min(144, available))
    },
    async sendMessage() {
      const text = this.draft.trim()
      if (
        (!text && !this.attachments.length) ||
        [...text].length > 4000 ||
        this.uploading ||
        this.$messageStore.sending ||
        this.$messageStore.connectionState !== 'connected'
      )
        return
      const id = this.chat.id
      const userId = useAuthStore().user?.id
      const attachments = this.attachments
      this.uploading = true
      try {
        for (const attachment of attachments) {
          if (!attachment.id) attachment.id = (await useFilesStore().upload(attachment.file)).id
          if (useAuthStore().user?.id !== userId) return
        }
        const sent = await this.$messageStore.send(
          id,
          text,
          attachments.map((attachment) => attachment.id!),
        )
        if (sent) delete this.fileDrafts[id]
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.uploading = false
      }
      await this.$nextTick()
      ;(this.$refs.composer as InstanceType<typeof MessageComposer> | undefined)?.focus()
    },
  },
})
</script>
