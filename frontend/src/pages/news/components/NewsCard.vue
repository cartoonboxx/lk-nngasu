<template>
  <RouterLink :class="[style.card, featured ? style.featured : '']" :to="url">
    <div
      v-if="image"
      :class="style.image"
      :style="{ backgroundImage: `url(${JSON.stringify(image)})` }"
    />

    <div :class="style.content">
      <time :datetime="date" :class="style.date">
        {{ formattedDate }}
      </time>

      <component :is="featured ? 'h2' : 'h3'" :class="style.title">
        {{ title }}
      </component>

      <div ref="descriptionBox" :class="style.descriptionBox">
        <p
          ref="descriptionText"
          :class="style.description"
          :style="{ WebkitLineClamp: descriptionLines }"
        >
          {{ description }}
        </p>
      </div>

      <div :class="style.readHint">
        Читать далее
        <PhArrowRight :size="14" />
      </div>
    </div>
  </RouterLink>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhArrowRight } from '@phosphor-icons/vue'
import style from './NewsCard.module.scss'
import { RouterLink } from 'vue-router'

export default defineComponent({
  name: 'NewsCard',

  components: {
    PhArrowRight,
    RouterLink,
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
      descriptionLines: 1,
      resizeObserver: null as ResizeObserver | null,
    }
  },

  computed: {
    canOpen(): boolean {
      return Boolean(this.url && this.url !== '#')
    },
    formattedDate(): string {
      return new Intl.DateTimeFormat('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).format(new Date(`${this.date}T00:00:00`))
    },
  },
  mounted() {
    const box = this.$refs.descriptionBox as HTMLElement

    this.resizeObserver = new ResizeObserver(this.updateDescriptionLines)
    this.resizeObserver.observe(box)
    this.updateDescriptionLines()
  },

  beforeUnmount() {
    this.resizeObserver?.disconnect()
  },

  methods: {
    updateDescriptionLines() {
      const box = this.$refs.descriptionBox as HTMLElement
      const text = this.$refs.descriptionText as HTMLElement

      const lineHeight = Number.parseFloat(window.getComputedStyle(text).lineHeight)

      if (!Number.isFinite(lineHeight) || lineHeight <= 0) return

      this.descriptionLines = Math.max(1, Math.floor(box.clientHeight / lineHeight))
    },
  },
})
</script>
