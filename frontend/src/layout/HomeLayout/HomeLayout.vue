<template>
  <NavBar @collapse-change="navigationCollapsed = $event" />
  <main :class="[style.mainContent, navigationCollapsed ? style.collapsed : null]">
    <RouterView />
  </main>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import NavBar from '@/layout/NavBar/NavBar.vue'
import style from './HomeLayout.module.scss'

export default defineComponent({
  name: 'HomeLayout',
  components: { NavBar },
  data() {
    return {
      style,
      navigationCollapsed: false,
    }
  },
})
</script>
