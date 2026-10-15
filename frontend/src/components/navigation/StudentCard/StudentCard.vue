<template>
  <button :class="style.card" type="button" :aria-label="`Открыть профиль ${name}`">
    <img v-if="imageSrc" :class="style.avatar" :src="imageSrc" :alt="imageAlt" />
    <span v-else :class="style.avatar" aria-hidden="true">АИ</span>
    <span :class="style.copy">
      <strong>{{ name }}</strong>
      <span>{{ role }}</span>
    </span>
    <PhCaretRight :class="style.arrow" :size="18" aria-hidden="true" />
  </button>
</template>

<script lang="ts">
import { PhCaretRight } from '@phosphor-icons/vue'
import { defineComponent } from 'vue'

import style from './StudentCard.module.scss'

export default defineComponent({
  name: 'StudentCard',
  components: {
    PhCaretRight,
  },
  props: {
    name: {
      type: String,
      required: true,
    },
    role: {
      type: String,
      required: true,
    },
    imageSrc: {
      type: String,
      default: undefined,
    },
    imageAlt: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
