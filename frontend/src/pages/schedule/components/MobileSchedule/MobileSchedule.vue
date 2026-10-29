<template>
  <section :class="style.mobileSchedule">
    <p v-if="!$scheduleStore.days.length">На эту неделю занятий нет</p>
    <ScheduleDay
      v-for="day in $scheduleStore.days"
      :key="day.date"
      :date="day.date"
      :title="
        new Date(`${day.date}T00:00:00`)
          .toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })
          .replace(/^./, (letter) => letter.toUpperCase())
      "
      :today="new Date(`${day.date}T00:00:00`).toDateString() === new Date().toDateString()"
      :lessons="day.lessons"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
import style from './MobileSchedule.module.scss'

export default defineComponent({
  name: 'MobileSchedule',
  components: { ScheduleDay },
  data() {
    return { style }
  },
})
</script>
