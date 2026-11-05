<template>
  <span :class="[style.root, style[variant], (inline || !$slots.default) && style.inline]">
    <slot />
    <span v-if="value > 0 || showZero" :class="style.badge">
      {{ value > max ? `${max}+` : value }}
    </span>
  </span>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './Badge.module.scss'

export default defineComponent({
  name: 'Badge',
  props: {
    inline: Boolean,
    value: {
      type: Number,
      required: true,
    },
    variant: {
      type: String as PropType<BadgeVariant>,
      default: 'accent',
    },
    max: {
      type: Number,
      default: 99,
    },
    showZero: Boolean,
  },
  data() {
    return { style }
  },
})
</script>
