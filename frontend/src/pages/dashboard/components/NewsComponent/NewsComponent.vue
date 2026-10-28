<template>
  <section :class="style.newsBlock">
    <header :class="style.header">
      <h2 id="dashboard-news-title" :class="style.heading"><PhNewspaper :size="22" />Новости</h2>
      <AppLink v-press :class="style.allLink" to="/news">
        Все новости <PhArrowRight :size="14" />
      </AppLink>
    </header>
    <ul :class="style.list">
      <li v-for="item in news" :key="item.id">
        <AppLink v-press :class="style.newsItem" :to="`/news/${item.id}`">
          <img :class="style.image" :src="item.image" alt="" loading="lazy" />
          <div :class="style.copy">
            <div :class="style.meta">
              <time :class="style.date" :datetime="item.date">{{ formatDate(item.date) }}</time>
            </div>
            <h3 :class="style.title">{{ item.title }}</h3>
            <p :class="style.description">{{ item.description }}</p>
            <span :class="style.readMore"> Читать новость <PhArrowRight :size="16" /> </span>
          </div>
        </AppLink>
      </li>
    </ul>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import AppLink from '@/components/navigation/AppLink/AppLink.vue'
import { PhNewspaper, PhArrowRight } from '@phosphor-icons/vue'
import { pressFeedback } from '@/directives/pressFeedback'
import campus from '$images/photoNews/Nngasu.png'
import technology from '$images/photoNews/news4.png'
import meeting from '$images/photoNews/news3.png'
import style from './NewsComponent.module.scss'

export default defineComponent({
  name: 'NewsComponent',
  components: { AppLink, PhNewspaper, PhArrowRight },
  directives: { press: pressFeedback },
  data() {
    return {
      style,
      news: [
        {
          id: 'internship',
          title: 'Открыта регистрация на летнюю практику 2025',
          description: 'Студенты 3–5 курсов могут подать заявку до 25 апреля.',
          date: '2025-04-12',
          image: campus,
        },
        {
          id: 'it-course',
          title: 'Новый курс: «Современные технологии в IT»',
          description: 'Запуск курса 20 апреля. Подробности в разделе «Курсы».',
          date: '2025-04-10',
          image: technology,
        },
        {
          id: 'employers',
          title: 'Встреча с работодателями',
          description: '22 апреля в 16:00 в аудитории 301. Приглашаются все желающие.',
          date: '2025-04-08',
          image: meeting,
        },
      ],
    }
  },
  methods: {
    formatDate(date: string) {
      return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
        .format(new Date(`${date}T00:00:00`))
        .replace(/ г\.$/, '')
    },
  },
})
</script>
