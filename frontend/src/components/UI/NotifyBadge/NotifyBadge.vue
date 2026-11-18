<template>
  <Teleport :to="notificationTarget">
    <div
      :class="style.notifications"
      @keydown.esc.stop.prevent="focusedId !== null && $utilityStore.dismissNotification(focusedId)"
    >
      <TransitionGroup
        :enter-active-class="style.enterActive"
        :leave-active-class="style.leaveActive"
        :enter-from-class="style.enterFrom"
        :leave-to-class="style.leaveTo"
        :move-class="style.move"
      >
        <article
          v-for="notification in $utilityStore.notifications"
          :key="notification.id"
          :class="style.badge"
          role="alert"
          @mouseenter="hoveredId = notification.id"
          @mouseleave="hoveredId = null"
          @focusin="focusedId = notification.id"
          @focusout="focusedId = null"
        >
          <PhWarningCircle :class="style.icon" :size="24" weight="fill" />
          <div :class="style.content">
            <div :class="style.heading">
              <strong
                >Ошибка<span v-if="notification.status">
                  HTTP {{ notification.status }}</span
                ></strong
              ><span v-if="notification.count > 1" :class="style.count"
                >×{{ notification.count }}</span
              >
            </div>
            <code :class="style.code">{{ notification.code }}</code>
            <p :class="style.message">{{ notification.message }}</p>
          </div>
          <IconButton
            variant="ghost"
            size="sm"
            title="Закрыть уведомление"
            @click="$utilityStore.dismissNotification(notification.id)"
            ><PhX :size="18"
          /></IconButton>
        </article>
      </TransitionGroup>
    </div>
  </Teleport>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { PhWarningCircle, PhX } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import style from './NotifyBadge.module.scss'

export default defineComponent({
  name: 'NotifyBadge',
  components: { IconButton, PhWarningCircle, PhX },
  data() {
    return {
      style,
      notificationTarget: 'body' as string | HTMLElement,
      dialogStack: [] as HTMLDialogElement[],
      observer: null as MutationObserver | null,
      timer: 0,
      hoveredId: null as number | null,
      focusedId: null as number | null,
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
    this.timer = window.setInterval(() => {
      for (const notification of this.$utilityStore.notifications) {
        if (
          notification.id !== this.hoveredId &&
          notification.id !== this.focusedId &&
          notification.expiresAt <= Date.now()
        )
          this.$utilityStore.dismissNotification(notification.id)
      }
    }, 500)
  },
  beforeUnmount() {
    this.observer?.disconnect()
    window.clearInterval(this.timer)
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
