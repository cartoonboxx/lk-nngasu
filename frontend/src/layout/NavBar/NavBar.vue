<template>
  <DesktopNavBar @collapse-change="$emit('collapse-change', $event)" />
  <MobileNavBar @mobile-menu-change="$emit('mobile-menu-change', $event)" />
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import DesktopNavBar from './DesktopNavBar/DesktopNavBar.vue'
import MobileNavBar from './MobileNavBar/MobileNavBar.vue'

export default defineComponent({
  name: 'NavBar',
  components: { DesktopNavBar, MobileNavBar },
  emits: ['collapse-change', 'mobile-menu-change'],
})
</script>
