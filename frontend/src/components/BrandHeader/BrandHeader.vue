<template>
  <header :class="[style.brand, compact ? style.compact : null]">
    <span :class="style.brandMark"><img src="$images/logo.svg" alt="" /></span>
    <strong :class="style.universityName">ННГАСУ</strong>
    <div :class="style.portalName">
      <strong>Личный кабинет</strong>
      <span>студента и сотрудника</span>
    </div>
  </header>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import style from './BrandHeader.module.scss'

export default defineComponent({
  name: 'BrandHeader',
  props: {
    compact: {
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
