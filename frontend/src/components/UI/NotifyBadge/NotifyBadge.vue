<template>
  <Teleport :to="notificationTarget">
    <p
      v-if="$utilityStore.notification"
      :class="[style.badge, style[$utilityStore.notification.variant]]"
      :role="$utilityStore.notification.variant === 'error' ? 'alert' : 'status'"
    >
      {{ $utilityStore.notification.message }}
    </p>
  </Teleport>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import style from './NotifyBadge.module.scss'

export default defineComponent({
  name: 'NotifyBadge',
  data() {
    return {
      style,
      notificationTarget: 'body' as string | HTMLElement,
      dialogStack: [] as HTMLDialogElement[],
      observer: null as MutationObserver | null,
      timer: 0,
    }
  },
  mounted() {
    this.syncDialogs([])
    this.observer = new MutationObserver(this.syncDialogs)
    this.observer.observe(document.body, {
      subtree: true,
      childList: true,
      attributes: true,
      attributeFilter: ['open'],
    })
  },
  watch: {
    '$utilityStore.notification': {
      immediate: true,
      handler() {
        window.clearTimeout(this.timer)
        if (this.$utilityStore.notification) {
          this.timer = window.setTimeout(() => {
            this.$utilityStore.notification = null
          }, 3500)
        }
      },
    },
  },
  beforeUnmount() {
    this.observer?.disconnect()
    window.clearTimeout(this.timer)
  },
  methods: {
    syncDialogs(records: MutationRecord[]) {
      // Порядок открытия может отличаться от порядка диалогов в DOM.
      const stack = this.dialogStack.filter(
        (dialog) => dialog.isConnected && dialog.matches(':modal'),
      )
      for (const record of records) {
        if (record.type !== 'attributes' || !(record.target instanceof HTMLDialogElement)) continue
        const index = stack.indexOf(record.target)
        if (index !== -1) stack.splice(index, 1)
        if (record.target.matches(':modal')) stack.push(record.target)
      }
      for (const dialog of document.querySelectorAll<HTMLDialogElement>('dialog:modal')) {
        if (!stack.includes(dialog)) stack.push(dialog)
      }
      this.dialogStack = stack
      this.notificationTarget = stack.at(-1) ?? 'body'
    },
  },
})
</script>
