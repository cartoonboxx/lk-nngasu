<template>
  <button
    v-press
    :class="[
      style.button,
      style[size],
      style[variant],
      state !== 'default' ? style[state] : null,
      round ? style.round : null,
    ]"
    :type="type"
    :disabled="disabled || state === 'disabled'"
  >
    <slot />
  </button>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent, type PropType } from 'vue'
import style from './IconButton.module.scss'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'IconButton',
  props: {
    variant: {
      type: String as PropType<IconButtonVariant>,
      default: 'secondary',
    },
    size: {
      type: String as PropType<IconButtonSize>,
      default: 'md',
    },
    state: {
      type: String as PropType<IconButtonState>,
      default: 'default',
    },
    type: {
      type: String as PropType<ButtonType>,
      default: 'button',
    },
    round: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
