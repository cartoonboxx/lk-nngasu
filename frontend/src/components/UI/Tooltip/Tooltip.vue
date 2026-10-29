<template>
  <div
    ref="trigger"
    :class="style.root"
    @pointerenter="onPointerEnter"
    @pointerleave="
      () => {
        hovered = false
        dismissed = false
      }
    "
    @pointerdown="
      () => {
        hovered = false
        dismissed = false
      }
    "
  >
    <slot />
    <Teleport to="body">
      <Transition
        :enter-active-class="style.enterTransition"
        :leave-active-class="style.transition"
        :enter-from-class="style.hidden"
        :leave-to-class="style.hidden"
      >
        <div
          v-if="visible"
          :class="[style.tooltip, style[position]]"
          :style="{ '--tooltip-x': `${x}px`, '--tooltip-y': `${y}px` }"
        >
          <slot name="content">{{ text }}</slot>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './Tooltip.module.scss'

export default defineComponent({
  name: 'Tooltip',
  props: {
    text: {
      type: String,
      default: '',
    },
    position: {
      type: String as PropType<TooltipPosition>,
      default: 'top',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    isShow: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      style,
      hovered: false,
      dismissed: false,
      x: 0,
      y: 0,
      animationFrame: 0,
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
        this.trackPosition()
        document.addEventListener('keydown', this.onKeydown)
      } else {
        window.cancelAnimationFrame(this.animationFrame)
        document.removeEventListener('keydown', this.onKeydown)
      }
    },
    position() {
      if (this.visible) this.updatePosition()
    },
  },
  beforeUnmount() {
    window.cancelAnimationFrame(this.animationFrame)
    document.removeEventListener('keydown', this.onKeydown)
  },
  methods: {
    trackPosition() {
      this.updatePosition()
      this.animationFrame = window.requestAnimationFrame(this.trackPosition)
    },
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
    onPointerEnter(event: PointerEvent) {
      if (
        this.isShow &&
        event.pointerType === 'mouse' &&
        window.matchMedia('(hover: hover)').matches
      ) {
        this.hovered = true
      }
    },
    onKeydown(event: KeyboardEvent) {
      if (event.key === 'Escape') this.dismissed = true
    },
  },
})
</script>
