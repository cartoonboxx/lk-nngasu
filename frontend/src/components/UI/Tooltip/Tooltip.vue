<template>
  <div
    ref="trigger"
    :class="style.root"
    :aria-describedby="visible ? id : undefined"
    @mouseenter="hovered = true"
    @mouseleave="onMouseleave"
  >
    <slot />
    <Teleport to="body">
      <Transition
        :enter-active-class="style.transition"
        :leave-active-class="style.transition"
        :enter-from-class="style.hidden"
        :leave-to-class="style.hidden"
      >
        <div
          v-if="visible"
          :id="id"
          :class="[style.tooltip, style[position]]"
          :style="{ '--tooltip-x': `${x}px`, '--tooltip-y': `${y}px` }"
          role="tooltip"
        >
          <slot name="content">{{ text }}</slot>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType, useId } from 'vue'
import style from './Tooltip.module.scss'

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right'

export default defineComponent({
  name: 'Tooltip',
  props: {
    text: { type: String, default: '' },
    position: { type: String as PropType<TooltipPosition>, default: 'top' },
    disabled: Boolean,
  },
  data() {
    return {
      style,
      id: `tooltip-${useId()}`,
      hovered: false,
      dismissed: false,
      x: 0,
      y: 0,
    }
  },
  computed: {
    visible(): boolean {
      return (
        !this.disabled &&
        !this.dismissed &&
        this.hovered &&
        Boolean(this.text || this.$slots.content)
      )
    },
  },
  watch: {
    visible(value: boolean) {
      if (value) {
        this.updatePosition()
        window.addEventListener('scroll', this.updatePosition, true)
        window.addEventListener('resize', this.updatePosition)
        document.addEventListener('keydown', this.onKeydown)
      } else {
        this.removeListeners()
      }
    },
    position() {
      if (this.visible) this.updatePosition()
    },
  },
  beforeUnmount() {
    this.removeListeners()
  },
  methods: {
    updatePosition() {
      const trigger = this.$refs.trigger as HTMLElement | undefined
      if (!trigger) return
      const rect = trigger.getBoundingClientRect()
      this.x =
        this.position === 'left'
          ? rect.left
          : this.position === 'right'
            ? rect.right
            : rect.left + rect.width / 2
      this.y =
        this.position === 'top'
          ? rect.top
          : this.position === 'bottom'
            ? rect.bottom
            : rect.top + rect.height / 2
    },
    onMouseleave() {
      this.hovered = false
      this.dismissed = false
    },
    onKeydown(event: KeyboardEvent) {
      if (event.key === 'Escape') this.dismissed = true
    },
    removeListeners() {
      window.removeEventListener('scroll', this.updatePosition, true)
      window.removeEventListener('resize', this.updatePosition)
      document.removeEventListener('keydown', this.onKeydown)
    },
  },
})
</script>
