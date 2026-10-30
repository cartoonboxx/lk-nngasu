<template>
  <section :class="style.newsDetailsPage">
    <div :class="style.toolbar">
      <div :class="style.back">
        <IconButton
          size="sm"
          variant="ghost"
          type="button"
          title="Назад к новостям"
          @click="$router.push('/news')"
        >
          <PhArrowLeft  />
        </IconButton>

        <span>Новости</span>
      </div>

      <IconButton
        v-if="currentNews"
        :class="style.shareButton"
        size="sm"
        variant="ghost"
        type="button"
        title="Поделиться новостью"
        @click="shareNews"
      >
        <PhShareNetwork />
      </IconButton>
    </div>

    <template v-if="currentNews">
      <header :class="style.header">
        <time :datetime="currentNews.date" :class="style.date">
          {{ formattedDate }}
        </time>

        <h1>{{ currentNews.title }}</h1>

        <p :class="style.description">
          {{ currentNews.description }}
        </p>
      </header>

      <div
        v-if="currentNews.image"
        :class="style.image"
        :style="{
          backgroundImage: `url(${JSON.stringify(currentNews.image)})`,
        }"
      />

      <div :class="style.body">
        <p
          v-for="(paragraph, index) in currentNews.paragraphs"
          :key="index"
        >
          {{ paragraph }}
        </p>
      </div>
    </template>

    <p v-else role="status">
      Новость не найдена.
    </p>

    <p v-if="shareMessage" :class="style.shareMessage" role="status">
      {{ shareMessage }}
    </p>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import {
  PhArrowLeft,
  PhShareNetwork,
} from '@phosphor-icons/vue'

import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'

import Nngasu from '$images/photoNews/Nngasu.png'
import style from './NewsDetailsPage.module.scss'

interface NewsDetails {
  id: number
  title: string
  description: string
  date: string
  image?: string
  paragraphs: string[]
}

export default defineComponent({
  name: 'NewsDetailsPage',

  components: {
    IconButton,
    PhArrowLeft,
    PhShareNetwork,
  },

  data() {
    return {
      style,
      shareMessage: '',

      news: [
        {
          id: 1,
          title: 'Открыта регистрация на летнюю практику',
          description:
            'Студенты могут подать заявку на прохождение летней практики в компаниях-партнёрах университета.',
          date: '2025-04-12',
          image: Nngasu,
          paragraphs: [
            'Университет объявляет старт регистрации на летнюю практику. Программа позволяет студентам применить полученные знания и получить профессиональный опыт.',
            'Подробности участия и сроки подачи заявок будут опубликованы дополнительно.',
          ],
        },
      ] as NewsDetails[],
    }
  },

  computed: {
    currentNews(): NewsDetails | undefined {
      const idNews = this.$route.params.idNews

      if (typeof idNews !== 'string' || !/^\d+$/.test(idNews)) {
        return undefined
      }

      return this.news.find((item) => item.id === Number(idNews))
    },

    formattedDate(): string {
      if (!this.currentNews) return ''

      return new Intl.DateTimeFormat('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).format(new Date(`${this.currentNews.date}T00:00:00`))
    },
  },

  watch: {
    '$route.params.idNews'() {
      this.shareMessage = ''
    },
  },

  methods: {
    async shareNews() {
      if (!this.currentNews) return

      this.shareMessage = ''

      try {
        if (navigator.share) {
          await navigator.share({
            title: this.currentNews.title,
            url: window.location.href,
          })
        } else if (navigator.clipboard) {
          await navigator.clipboard.writeText(window.location.href)
          this.shareMessage = 'Ссылка скопирована.'
        } else {
          this.shareMessage =
            'Скопируйте адрес новости из строки браузера.'
        }
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') {
          return
        }

        this.shareMessage =
          'Не удалось поделиться. Скопируйте адрес из строки браузера.'
      }
    },
  },
})
</script>