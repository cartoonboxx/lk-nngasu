<template>
  <button
    v-press
    :class="[style.tab, style[size], style[variant], state !== 'default' ? style[state] : null]"
    type="button"
    :disabled="disabled || state === 'disabled'"
  >
    <slot />
  </button>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent, type PropType } from 'vue'
import style from './IconTab.module.scss'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'IconTab',
  props: {
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
