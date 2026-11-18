<template>
  <section :class="style.page">
    <SectionHeader header="Новости" paragraph="Публикации личного кабинета" />
    <p :class="style.notice">Управление новостями пока недоступно.</p>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import style from './Admin.module.scss'

export default defineComponent({
  name: 'AdminNewsPage',
  components: { SectionHeader },
  data() {
    return { style }
  },
})
</script>
