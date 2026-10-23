<template>
  <article :class="[style.card, featured ? style.featured : '']">
    <img
      v-if="image"
      :src="image"
      alt=""
      :loading="featured ? 'eager' : 'lazy'"
      :class="style.image"
    />

    <div :class="style.content">
      <time :datetime="date" :class="style.date">
        {{ formattedDate }}
      </time>

      <component :is="featured ? 'h2' : 'h3'" :class="style.title">
        <a :href="url" :class="style.titleLink">
          {{ title }}
        </a>
      </component>

      <p :class="style.description">
        {{ description }}
      </p>

      <a v-if="featured" :href="url" :class="style.readMore">
        Читать подробнее
        <PhArrowRight :size="20" aria-hidden="true" />
      </a>
    </div>
  </article>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhArrowRight } from '@phosphor-icons/vue'
import style from './NewsCard.module.scss'

export default defineComponent({
  name: 'NewsCard',

  components: {
    PhArrowRight,
  },

  props: {
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      required: true,
    },
    date: {
      type: String,
      required: true,
    },
    url: {
      type: String,
      required: true,
    },
    image: {
      type: String,
      default: '',
    },
    featured: {
      type: Boolean,
      default: false,
    },
  },

  data() {
    return {
      style,
    }
  },

  computed: {
    formattedDate(): string {
      return new Intl.DateTimeFormat('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).format(new Date(`${this.date}T00:00:00`))
    },
  },
})
</script>
