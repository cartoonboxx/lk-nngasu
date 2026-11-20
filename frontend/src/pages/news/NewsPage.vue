<template>
  <section :class="style.newsPage">
    <SectionHeader header="Новости" paragraph="Новости и объявления университета" />
    <StyleInput v-model="searchQuery" placeholder="Поиск новостей" full-width />
    <PageLoader v-if="cabinet.newsLoading" />
    <p v-else-if="cabinet.newsError">
      <button @click="cabinet.fetchNews()">Повторить</button>
    </p>
    <p v-else-if="!news.length" :class="style.empty">
      {{ searchQuery ? 'По вашему запросу ничего не найдено.' : 'Новостей пока нет.' }}
    </p>
    <div v-else :class="style.newsGrid">
      <NewsCard
        v-for="item in news"
        :key="item.id"
        :title="item.excerpt"
        :description="item.excerpt"
        :date="item.published_at"
        :url="`/news/${item.id}`"
      />
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import NewsCard from './components/NewsCard.vue'
import { useCabinetStore } from '@/stores/cabinet'
import style from './NewsPage.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'NewsPage',
  components: { SectionHeader, StyleInput, NewsCard,PageLoader },
  data() {
    return { style, searchQuery: '' }
  },
  async created() {
    await this.cabinet.fetchNews()
  },
  computed: {
    cabinet() {
      return useCabinetStore()
    },
    news() {
      const query = this.searchQuery.trim().toLocaleLowerCase('ru')
      return this.cabinet.news.filter((item) =>
        `${item.excerpt} ${item.excerpt}`.toLocaleLowerCase('ru').includes(query),
      )
    },
  },
})
</script>
