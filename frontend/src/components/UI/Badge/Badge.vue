<template>
  <span :class="[style.root, inline && style.inline]">
    <slot />
    <span :class="style.badge">
      {{ value }}
    </span>
  </span>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import style from './Badge.module.scss'

export default defineComponent({
  name: 'Badge',
  props: {
    inline: {
      type: Boolean,
      default: false,
    },
    value: {
      type: Number,
      required: true,
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
