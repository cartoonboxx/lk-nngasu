<template>
  <section :class="style.newsBlock">
    <header :class="style.header">
      <h2 :class="style.heading">Новости</h2>
      <RouterLink to="/news">Все новости</RouterLink>
    </header>
    <p v-if="cabinet.newsLoading">Загрузка новостей…</p>
    <p v-else-if="cabinet.newsError" role="alert">
      {{ cabinet.newsError }} <button @click="cabinet.fetchNews()">Повторить</button>
    </p>
    <p v-else-if="!cabinet.news.length">Новостей пока нет</p>
    <ul v-else :class="style.list">
      <li v-for="item in cabinet.news.slice(0, 3)" :key="item.id">
        <RouterLink :class="style.newsItem" :to="`/news/${item.id}`">
          <div :class="style.copy">
            <time :datetime="item.published_at">{{
              new Date(item.published_at).toLocaleDateString('ru-RU')
            }}</time>
            <h3 :class="style.title">{{ item.title }}</h3>
            <p :class="style.description">{{ item.text }}</p>
          </div>
        </RouterLink>
      </li>
    </ul>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { useCabinetStore } from '@/stores/cabinet'
import style from './NewsComponent.module.scss'
export default defineComponent({
  name: 'NewsComponent',
  components: { RouterLink },
  data() {
    return { style, cabinet: useCabinetStore() }
  },
  created() {
    void this.cabinet.fetchNews()
  },
})
</script>
