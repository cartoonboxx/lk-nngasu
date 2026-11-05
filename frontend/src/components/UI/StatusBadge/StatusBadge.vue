<template>
  <span :class="[style.root, style[variant]]"><slot /></span>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './StatusBadge.module.scss'

export default defineComponent({
  name: 'StatusBadge',
  props: {
    variant: { type: String as PropType<StatusBadgeVariant>, default: 'neutral' },
  },
  data() {
    return { style }
  },
})
</script>
