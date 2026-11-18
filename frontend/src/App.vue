<template>
  <router-view />
  <NotifyBadge />
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import NotifyBadge from '@/components/UI/NotifyBadge/NotifyBadge.vue'

export default defineComponent({
  name: 'App',
  components: { NotifyBadge },
})
</script>
