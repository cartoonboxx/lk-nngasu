<template>
  <section :class="style.resoursePage">
    <SectionHeader
      header="Ресурсы и сервисы"
      paragraph="Учебные платформы и контакты университета"
    />
    <p>Ресурсы и контакты пока не опубликованы.</p>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import style from './ResourcesPage.module.scss'
export default defineComponent({
  name: 'ResourcesPage',
  components: { SectionHeader },
  data() {
    return { style }
  },
})
</script>
