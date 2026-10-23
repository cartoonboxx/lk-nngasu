<template>
  <article :class="[style.card, featured ? style.featured : '']">
    <div
      v-if="image"
      :class="style.image"
      :style="{ backgroundImage: `url(${JSON.stringify(image)})` }"
      aria-hidden="true"
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

      <div :class="style.actions">
        <Button :variant="featured ? 'primary' : 'secondary'" @click="openNews">
        Читать подробнее
        <PhArrowRight :size="20" aria-hidden="true" />
      </Button>
      </div>
      
    </div>
  </article>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhArrowRight } from '@phosphor-icons/vue'
import style from './NewsCard.module.scss'
import Button from '@/components/UI/buttons/Button/Button.vue'

export default defineComponent({
  name: 'NewsCard',

  components: {
    PhArrowRight,
    Button,
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
  methods: {
    openNews() {
      if (!this.url || this.url === '#') return

      window.location.assign(this.url)
    },
  },
})
</script>
