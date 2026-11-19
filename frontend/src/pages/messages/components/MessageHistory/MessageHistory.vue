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
        <form v-if="editingId === message.id" @submit.prevent="saveEdit(message)">
          <TextInput
            v-model="editText"
            label="Текст сообщения"
            multiline
            full-width
            :maxlength="4000"
          />
          <Button
            type="submit"
            size="sm"
            :disabled="!!$messageStore.mutationId || $messageStore.connectionState !== 'connected'"
            >Сохранить</Button
          >
          <Button size="sm" variant="ghost" @click="editingId = ''">Отмена</Button>
        </form>
        <p v-else>{{ message.text }}</p>
        <Button
          v-for="file in message.files"
          :key="file.id"
          size="sm"
          variant="ghost"
          @click="download(file.id, file.name)"
        >
          {{ file.name }}
        </Button>
        <div :class="style.messageMeta">
          <span v-if="message.editedAt">изменено</span>
          <time :datetime="message.sentAt">
            {{ messageTimeFormatter.format(new Date(message.sentAt)) }}
          </time>
          <PhCheck v-if="message.own" :size="14" title="Сохранено на сервере" />
          <IconButton
            v-if="message.own"
            size="sm"
            variant="ghost"
            title="Изменить"
            :disabled="!!$messageStore.mutationId || $messageStore.connectionState !== 'connected'"
            @click="
              () => {
                editingId = message.id
                editText = message.text
              }
            "
          >
            <PhPencilSimple :size="14" />
          </IconButton>
          <IconButton
            v-if="message.own"
            size="sm"
            variant="ghost"
            title="Удалить"
            :disabled="!!$messageStore.mutationId || $messageStore.connectionState !== 'connected'"
            @click="removeTarget = message"
          >
            <PhTrash :size="14" />
          </IconButton>
        </div>
      </article>
    </template>
    <Modal
      :model-value="!!removeTarget"
      :width="400"
      @update:model-value="!$event && (removeTarget = null)"
    >
      <h2>Удалить сообщение?</h2>
      <p>Сообщение исчезнет у всех участников чата.</p>
      <Button variant="ghost" :disabled="!!$messageStore.mutationId" @click="removeTarget = null"
        >Отмена</Button
      >
      <Button variant="danger" :loading="!!$messageStore.mutationId" @click="removeMessage"
        >Удалить</Button
      >
    </Modal>
  </div>
</template>

<script lang="ts">
import { apiNotifications } from '@/api/client'
import { defineComponent, type PropType } from 'vue'
import style from './MessageHistory.module.scss'
import { PhChatCircleText, PhCheck, PhPencilSimple, PhTrash } from '@phosphor-icons/vue'
import { useFilesStore } from '@/stores/files'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import type { Chat, Message } from '../../messages.types'
import { messageTimeFormatter, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'MessageHistory',
  components: {
    PhChatCircleText,
    PhCheck,
    PhPencilSimple,
    PhTrash,
    Modal,
    Button,
    IconButton,
    TextInput,
  },
  props: {
    chat: { type: Object as PropType<Chat>, required: true },
  },
  data() {
    return {
      style,
      editingId: '',
      editText: '',
      removeTarget: null as Message | null,

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
    await this.scrollToEnd()
  },
  beforeUnmount() {
    this.resizeObserver?.disconnect()
  },
  watch: {
    async 'chat.id'() {
      this.editingId = ''
      this.removeTarget = null
      await this.scrollToEnd()
    },
    async 'chat.messages'(messages: Message[], previous: Message[]) {
      const history = this.$refs.history as HTMLElement | undefined
      const before = history?.scrollHeight ?? 0
      if (previous[0] && messages[0] && Number(messages[0].id) < Number(previous[0].id)) {
        await this.$nextTick()
        if (history) history.scrollTop += history.scrollHeight - before
      } else if (this.followingLatest || messages.at(-1)?.own) await this.scrollToEnd()
    },
  },
  methods: {
    async saveEdit(message: Message) {
      if (!this.editText.trim() && !message.files.length) return
      const id = this.chat.id
      const saved = await this.$messageStore.mutateMessage(id, message.id, this.editText.trim())
      if (saved && this.chat.id === id) {
        this.editingId = ''
        this.editText = ''
      }
    },
    async removeMessage() {
      if (!this.removeTarget) return
      const id = this.chat.id
      const saved = await this.$messageStore.mutateMessage(id, this.removeTarget.id)
      if (saved && this.chat.id === id) {
        this.removeTarget = null
        this.editingId = ''
      }
    },
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
    },
    async onResize() {
      const history = this.$refs.history as HTMLElement
      this.historyHeight = history.clientHeight
      this.historyWidth = history.clientWidth
      // Keep the bottom anchored when the keyboard or the growing composer changes the viewport.
      if (this.followingLatest) await this.scrollToEnd()
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
