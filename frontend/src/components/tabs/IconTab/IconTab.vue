<template>
  <button
    :class="[style.tab, style[size], style[variant], state !== 'default' ? style[state] : null]"
    type="button"
    role="tab"
    :aria-label="ariaLabel"
    :aria-selected="variant === 'selected'"
    :disabled="disabled || state === 'disabled'"
  >
    <slot />
  </button>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './IconTab.module.scss'

export default defineComponent({
  name: 'IconTab',
  props: {
    ariaLabel: {
      type: String,
      required: true,
    },
    variant: {
      type: String as PropType<IconTabVariant>,
      default: 'default',
    },
    size: {
      type: String as PropType<IconTabSize>,
      default: 'md',
    },
    state: {
      type: String as PropType<IconTabState>,
      default: 'default',
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
