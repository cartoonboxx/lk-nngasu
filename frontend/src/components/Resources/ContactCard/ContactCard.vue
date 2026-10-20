<template>
  <article :class="style.card">
    <div :class="style.icon">
      <component :is="icon" :size="28" aria-hidden="true" />
    </div>

    <div :class="style.content">
      <h3>{{ label }}</h3>
      <p>{{ description }}</p>

      <a :class="style.link" :href="url">
        {{ linkText }}
        <PhArrowSquareOut :size="18" aria-hidden="true" />
      </a>
    </div>
  </article>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import type { Component } from 'vue'
import { PhArrowSquareOut } from '@phosphor-icons/vue'

import style from './ContactCard.module.scss'

export default defineComponent({
  name: 'ContactCard',

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
      type: Object as () => Component,
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
