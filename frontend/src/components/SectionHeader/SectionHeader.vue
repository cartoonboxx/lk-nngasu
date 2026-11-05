<template>
  <section :class="[style.sectionHeader, nm ? style.noMargin : null]">
    <div :class="style.information">
      <h1>{{ header }}</h1>
      <p>{{ paragraph }}</p>
    </div>
    <div :class="style.actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SectionHeader.module.scss'

export default defineComponent({
  name: 'SectionHeader',
  props: {
    header: {
      type: String,
      default: '',
    },
    paragraph: {
      type: String,
      default: '',
    },
    nm: {
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
