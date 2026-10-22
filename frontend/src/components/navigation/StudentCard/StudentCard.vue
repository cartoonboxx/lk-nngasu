<template>
  <button
    :class="[style.card, compact ? style.compact : null]"
    type="button"
    :aria-label="`Открыть профиль ${name}`"
    :title="compact ? `${name} · ${role}` : undefined"
  >
    <img v-if="imageSrc" :class="style.avatar" :src="imageSrc" :alt="imageAlt" />
    <span v-else :class="style.avatar" aria-hidden="true">АИ</span>
    <span v-if="!compact" :class="style.copy">
      <strong>{{ name }}</strong>
      <span>{{ role }}</span>
    </span>
    <PhCaretRight v-if="!compact" :class="style.arrow" :size="18" aria-hidden="true" />
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
    compact: {
      type: Boolean,
      default: false,
    },
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
