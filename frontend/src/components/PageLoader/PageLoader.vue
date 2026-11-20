<template>
  <Teleport to="body">
    <div :class="style.overlay">
      <span :class="style.loader" />
    </div>
  </Teleport>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './PageLoader.module.scss'

export default defineComponent({
  name: 'PageLoader',

  data() {
    return {
      style,
    }
  },
})
</script>