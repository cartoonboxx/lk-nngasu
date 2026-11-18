<template>
  <section :class="style.newsDetailsPage">
    <RouterLink to="/news">Назад к новостям</RouterLink>
    <p v-if="cabinet.newsLoading">Загрузка новости…</p>
    <p v-else-if="cabinet.newsError" role="alert">
      {{ cabinet.newsError }} <button @click="cabinet.fetchNews()">Повторить</button>
    </p>
    <article v-else-if="currentNews">
      <header :class="style.header">
        <time :datetime="currentNews.published_at">{{
          new Date(currentNews.published_at).toLocaleDateString('ru-RU')
        }}</time>
        <h1>{{ currentNews.title }}</h1>
      </header>
      <p :class="style.body" style="white-space: pre-wrap">{{ currentNews.text }}</p>
    </article>
    <p v-else>Новость не найдена.</p>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { useCabinetStore } from '@/stores/cabinet'
import style from './NewsDetailsPage.module.scss'
export default defineComponent({
  name: 'NewsDetailsPage',
  components: { RouterLink },
  data() {
    return { style, cabinet: useCabinetStore() }
  },
  created() {
    void this.cabinet.fetchNews()
  },
  computed: {
    currentNews() {
      return this.cabinet.news.find((item) => String(item.id) === this.$route.params.idNews)
    },
  },
})
</script>
