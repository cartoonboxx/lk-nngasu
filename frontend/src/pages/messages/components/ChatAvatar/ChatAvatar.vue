<template>
  <span :class="style.avatar" aria-hidden="true">
    <PhUsersThree v-if="group" :size="22" />
    <template v-else>{{ initials }}</template>
  </span>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhUsersThree } from '@phosphor-icons/vue'
import style from './ChatAvatar.module.scss'
export default defineComponent({
  name: 'ChatAvatar',
  components: { PhUsersThree },
  props: { initials: { type: String, required: true }, group: Boolean },
  data() {
    return { style }
  },
})
</script>
