<template>
  <a
    :href="link.href"
    :class="[
      link.isActive && 'router-link-active',
      link.isExactActive && 'router-link-exact-active',
    ]"
    @click="link.navigate"
  >
    <slot />
  </a>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { useLink, type RouteLocationRaw } from 'vue-router'

export default defineComponent({
  name: 'AppLink',
  props: {
    to: { type: [String, Object] as PropType<RouteLocationRaw>, required: true },
    replace: Boolean,
  },
  data() {
    return { link: useLink(this.$props) }
  },
})
</script>
