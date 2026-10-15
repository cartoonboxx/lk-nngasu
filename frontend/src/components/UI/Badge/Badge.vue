<template>
  <div :class="style.root">
    <slot />
    <span :class="style.badge" role="status" :aria-label="`Количество: ${value}`">
      {{ value }}
    </span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import style from './Badge.module.scss'

export default defineComponent({
  name: 'Badge',
  props: {
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
