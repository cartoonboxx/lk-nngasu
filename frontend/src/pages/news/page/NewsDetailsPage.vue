<template>
  <section :class="style.newsDetailsPage">
    <RouterLink to="/news">Назад к новостям</RouterLink>
    <PageLoader v-if="$subjectsStore.loading" />
    <p v-else-if="cabinet.detailError">
      <button @click="cabinet.fetchNewsDetail(String($route.params.idNews))">Повторить</button>
    </p>
    <article v-else-if="currentNews">
      <header :class="style.header">
        <time :datetime="currentNews.published_at">{{
          new Date(currentNews.published_at).toLocaleDateString('ru-RU')
        }}</time>
        <h1>{{ currentNews.excerpt }}</h1>
      </header>
      <p :class="style.body" style="white-space: pre-wrap">{{ currentNews.body }}</p>
    </article>
    <p v-else>Новость не найдена.</p>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { useCabinetStore } from '@/stores/cabinet'
import style from './NewsDetailsPage.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'NewsDetailsPage',
  components: { RouterLink, PageLoader },
  data() {
    return { style }
  },
  watch: {
    '$route.params.idNews': {
      immediate: true,
      handler(id: string) {
        void this.cabinet.fetchNewsDetail(id)
      },
    },
  },
  computed: {
    cabinet() {
      return useCabinetStore()
    },
    currentNews() {
      return this.cabinet.currentNews
    },
  },
})
</script>
