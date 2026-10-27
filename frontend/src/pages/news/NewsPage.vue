<template>
  <section :class="style.newsPage">
    <SectionHeader header="Новости" paragraph="Новости университета, объявления и события кампуса">
      <template #actions>
        <div :class="style.search">
          <StyleInput v-model="searchQuery" placeholder="Поиск новостей" :full-width="true">
            <template #prefix>
              <PhMagnifyingGlass :size="20" aria-hidden="true" />
            </template>
          </StyleInput>
        </div>
      </template>
    </SectionHeader>

    <template v-if="featuredNews">
      <NewsCard
        :title="featuredNews.title"
        :description="featuredNews.description"
        :date="featuredNews.date"
        :url="featuredNews.url"
        :image="featuredNews.image"
        featured
      />

      <div v-if="visibleNews.length" :class="style.newsGrid">
        <NewsCard
          v-for="item in visibleNews"
          :key="item.id"
          :title="item.title"
          :description="item.description"
          :date="item.date"
          :url="item.url"
          :image="item.image"
        />
      </div>


    </template>

    <p v-else :class="style.empty" role="status">По вашему запросу ничего не найдено.</p>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'

import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import NewsCard from './components/NewsCard.vue'
import style from './NewsPage.module.scss'

import Nngasu from '$images/photoNews/Nngasu.png'
import ImageNews2 from '$images/photoNews/news2.png'
import ImageNews3 from '$images/photoNews/news3.png'
import ImageNews4 from '$images/photoNews/news4.png'

interface NewsItem {
  id: number
  title: string
  description: string
  date: string
  url: string
  image?: string
}

const PAGE_SIZE = 6

export default defineComponent({
  name: 'NewsPage',

  components: {
    SectionHeader,
    StyleInput,
    PhMagnifyingGlass,
    NewsCard,
  },

  data() {
    return {
      style,
      searchQuery: '',
      visibleCount: PAGE_SIZE,

      news: [
        {
          id: 1,
          title: 'Открыта регистрация на летнюю практику',
          description:
            'Студенты могут подать заявку на прохождение летней практики в компаниях-партнёрах университета.',
          date: '2025-04-12',
          url: '#',
          image: Nngasu,
        },
        {
          id: 2,
          title: 'Конкурс «Наука в каждом шаге»»',
          description:
            'Наука в каждом шаге» знакомит школьников и студентов с маршрутами научно-популярного туризма Нижегородской области. Проект работает при поддержке Минобрнауки России в рамках Десятилетия науки и технологий.',
          date: '2025-04-10',
          url: '#',
          image: ImageNews2,
        },
        {
          id: 3,
          title: 'Встреча с работодателями',
          description: 'Приглашаются все желающие. Регистрация открыта.',
          date: '2025-04-08',
          url: '#',
          image: ImageNews3,
        },
        {
          id: 4,
          title: 'Изменения в расписании на весенний семестр',
          description: 'Проверьте актуальную информацию в разделе «Расписание».',
          date: '2025-04-05',
          url: '#',
        },
        {
          id: 5,
          title: 'Результаты научной конференции',
          description:
            'Подводим итоги ежегодной студенческой конференции. Поздравляем победителей!',
          date: '2025-04-03',
          url: '#',
          image: ImageNews4,
        },
        {
          id: 6,
          title: 'День открытых дверей',
          description: 'Приглашаем абитуриентов и их родителей познакомиться с университетом.',
          date: '2025-04-01',
          url: '#',
        },
        {
          id: 7,
          title: 'Изменения в расписании мероприятий',
          description: 'Обновлён график мероприятий на апрель. Ознакомьтесь с изменениями.',
          date: '2025-03-28',
          url: '#',
          image: '',
        },
        {
          id: 8,
          title: 'Открыт набор в студенческие объединения',
          description: 'Выберите интересующее направление и присоединяйтесь к команде.',
          date: '2025-03-25',
          url: '#',
        },
      ] as NewsItem[],
    }
  },

  computed: {
    filteredNews(): NewsItem[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.news
        .filter((item) => `${item.title} ${item.description}`.toLowerCase().includes(query))
        .sort((a, b) => b.date.localeCompare(a.date))
    },

    featuredNews(): NewsItem | undefined {
      return this.filteredNews[0]
    },

    remainingNews(): NewsItem[] {
      return this.filteredNews.slice(1)
    },

    visibleNews(): NewsItem[] {
      return this.remainingNews.slice(0, this.visibleCount)
    },

    hasMoreNews(): boolean {
      return this.visibleCount < this.remainingNews.length
    },
  },

  watch: {
    searchQuery() {
      this.visibleCount = PAGE_SIZE
    },
  },

  methods: {
    formatDate(date: string): string {
      return new Intl.DateTimeFormat('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).format(new Date(`${date}T00:00:00`))
    },

    showMore() {
      this.visibleCount += PAGE_SIZE
    },
  },
})
</script>
