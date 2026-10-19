<template>
  <Teleport to="body">
    <dialog
      ref="dialog"
      v-bind="$attrs"
      :class="style.overlay"
      @cancel.self.prevent="close"
      @close="onNativeClose"
      @pointerdown="backdropPressed = $event.target === $event.currentTarget"
      @click="onBackdropClick"
    >
      <div
        :class="style.modal"
        :style="{ width: typeof width === 'number' ? `${width}px` : width }"
      >
        <div :class="style.content">
          <slot :close="close" />
        </div>
      </div>
    </dialog>
  </Teleport>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './BaseModal.module.scss'

// Keep the page locked until the last modal closes.
const scrollLocks = new Set<HTMLDialogElement>()
let previousOverflow = ''

function unlockScroll(dialog: HTMLDialogElement) {
  if (scrollLocks.delete(dialog) && !scrollLocks.size) {
    document.body.style.overflow = previousOverflow
  }
}

export default defineComponent({
  name: 'Modal',
  inheritAttrs: false,
  props: {
    modelValue: { type: Boolean, default: false },
    width: { type: [Number, String], default: 540 },
    closeOnOverlay: { type: Boolean, default: true },
  },
  emits: {
    'update:modelValue': (value: boolean) => typeof value === 'boolean',
    close: () => true,
  },
  data() {
    return { style, backdropPressed: false }
  },
  watch: {
    modelValue() {
      this.syncDialog()
    },
  },
  mounted() {
    this.syncDialog()
  },
  beforeUnmount() {
    const dialog = this.$refs.dialog as HTMLDialogElement
    if (dialog.open) dialog.close()
    unlockScroll(dialog)
  },
  methods: {
    syncDialog() {
      const dialog = this.$refs.dialog as HTMLDialogElement | undefined
      if (!dialog) return
      if (this.modelValue && !dialog.open) {
        dialog.showModal()
        if (!scrollLocks.size) {
          previousOverflow = document.body.style.overflow
          document.body.style.overflow = 'hidden'
        }
        scrollLocks.add(dialog)
      } else if (!this.modelValue && dialog.open) {
        dialog.close()
        unlockScroll(dialog)
      }
    },
    close() {
      this.$emit('update:modelValue', false)
    },
    onBackdropClick(event: MouseEvent) {
      if (this.closeOnOverlay && this.backdropPressed && event.target === event.currentTarget) {
        this.close()
      }
      this.backdropPressed = false
    },
    onNativeClose() {
      const dialog = this.$refs.dialog as HTMLDialogElement | undefined
      // A queued close event can arrive after the dialog has reopened.
      if (!dialog || dialog.open) return
      unlockScroll(dialog)
      if (this.modelValue) this.close()
      this.$emit('close')
    },
  },
})
</script>
