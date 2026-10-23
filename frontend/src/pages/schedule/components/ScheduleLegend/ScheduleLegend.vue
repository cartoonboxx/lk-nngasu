<template>
  <ul :class="style.legend">
    <li v-for="item in lessonTypes" :key="item.type">
      <span :class="[style.legendDot, style[item.type]]" />{{ item.label }}
    </li>
  </ul>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './ScheduleLegend.module.scss'
import { lessonTypes } from '../../schedule.data'

export default defineComponent({
  name: 'ScheduleLegend',
  data() {
    return { style, lessonTypes }
  },
})
</script>
