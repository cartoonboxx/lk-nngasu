<template>
  <article :class="style.card">
    <div :class="style.content">
      <h3>{{ label }}</h3>
      <p>{{ description }}</p>

      <a :class="style.link" :href="url" target="_blank" rel="noopener noreferrer">
        {{ linkText }}
        <PhArrowSquareOut :size="18" aria-hidden="true" />
      </a>
    </div>

    <div :class="style.icon">
      <img
        v-if="typeof icon === 'string'"
        :src="icon"
        :class="style.image"
        :alt="`${label} — логотип`"
      />

      <component :is="icon" v-else :size="56" weight="fill" aria-hidden="true" />
    </div>
  </article>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import type { Component } from 'vue'
import { PhArrowSquareOut } from '@phosphor-icons/vue'

import style from './ServiceCard.module.scss'

export default defineComponent({
  name: 'ServiceCard',

  components: {
    PhArrowSquareOut,
  },

  props: {
    label: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      required: true,
    },
    url: {
      type: String,
      required: true,
    },
    linkText: {
      type: String,
      required: true,
    },
    icon: {
      type: [String, Object] as unknown as () => string | Component,
      required: true,
    },
  },

  data() {
    return {
      style,
    }
  },
})
</script>
