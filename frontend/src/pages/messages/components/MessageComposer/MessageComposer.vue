<template>
  <form :class="style.composer" @submit.prevent="sendMessage">
    <p v-if="error" :class="style.storageError" role="alert">
      {{ error }} Черновик сохранён в форме; можно отправить повторно.
    </p>
    <div :class="style.composerRow">
      <TextInput
        ref="messageField"
        v-model="draft"
        :class="style.composerInput"
        :maxlength="4000"
        base-height="var(--control-height-md)"
        :max-height="maxHeight"
        placeholder="Напишите сообщение…"
        multiline
        full-width
        @keydown="onComposerKeydown"
      />
      <IconButton
        type="submit"
        variant="primary"
        round
        :disabled="sending || !draft.trim()"
        @pointerdown="onSendPointerDown"
        @pointerup="onSendPointerUp"
        @pointercancel="sendPointer = null"
        @pointerleave="sendPointer = null"
        title="Отправить сообщение"
      >
        <PhPaperPlaneRight :size="21" weight="bold" />
      </IconButton>
    </div>
    <span :class="style.composerHint">Enter — отправить · Shift + Enter — новая строка</span>
  </form>
</template>

<script lang="ts">
import type { SendPointer } from '../../messages.types'
import { defineComponent } from 'vue'
import { PhPaperPlaneRight } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import style from './MessageComposer.module.scss'
export default defineComponent({
  name: 'MessageComposer',
  components: { PhPaperPlaneRight, IconButton, TextInput },
  props: {
    modelValue: { type: String, required: true },
    maxHeight: { type: Number, default: 144 },
    error: { type: String, default: '' },
    sending: Boolean,
  },
  emits: {
    'update:modelValue': (value: string) => typeof value === 'string',
    send: () => true,
  },
  data() {
    return { style, sendPointer: null as SendPointer | null }
  },
  computed: {
    draft: {
      get(): string {
        return this.modelValue
      },
      set(value: string) {
        this.$emit('update:modelValue', value)
      },
    },
  },
  methods: {
    focus() {
      ;(this.$refs.messageField as InstanceType<typeof TextInput> | undefined)?.focus()
    },
    sendMessage() {
      if (!this.sending && this.draft.trim()) this.$emit('send')
    },
    onSendPointerDown(event: PointerEvent) {
      if (event.pointerType !== 'mouse') {
        this.sendPointer = { id: event.pointerId, x: event.clientX, y: event.clientY }
      }
      const field = this.$refs.messageField as InstanceType<typeof TextInput> | undefined
      if (field?.$el.contains(document.activeElement)) event.preventDefault()
    },
    onSendPointerUp(event: PointerEvent) {
      const start = this.sendPointer
      this.sendPointer = null
      if (
        !start ||
        start.id !== event.pointerId ||
        Math.hypot(event.clientX - start.x, event.clientY - start.y) > 10
      )
        return
      const bounds = (event.currentTarget as HTMLElement).getBoundingClientRect()
      if (
        event.clientX < bounds.left ||
        event.clientX > bounds.right ||
        event.clientY < bounds.top ||
        event.clientY > bounds.bottom
      )
        return
      // Submit before a touch-generated blur can resize the keyboard or move the button.
      event.preventDefault()
      this.sendMessage()
    },
    onComposerKeydown(event: KeyboardEvent) {
      if (
        event.key === 'Enter' &&
        !event.shiftKey &&
        !event.isComposing &&
        !this.$utilityStore.isMobileVersion()
      ) {
        event.preventDefault()
        this.sendMessage()
      }
    },
  },
})
</script>
