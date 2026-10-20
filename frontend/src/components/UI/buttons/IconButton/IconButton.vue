<template>
  <button
    :class="[style.button, style[size], style[variant], state !== 'default' ? style[state] : null]"
    :type="type"
    :disabled="disabled || state === 'disabled'"
  >
    <slot />
  </button>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './IconButton.module.scss'

export default defineComponent({
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
