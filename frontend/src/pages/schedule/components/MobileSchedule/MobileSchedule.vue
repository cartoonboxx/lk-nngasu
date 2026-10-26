<template>
  <section :class="style.mobileSchedule">
    <ScheduleDay
      v-for="day in scheduleDays"
      :key="day.date"
      :date="day.date"
      :title="day.fullLabel"
      :today="day.isToday"
      :lessons="day.lessons"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
import type { ScheduleDayItem } from '@/components/ScheduleDay/schedule-day.types'
import style from './MobileSchedule.module.scss'
import { useSchedule } from '../../useSchedule'

export default defineComponent({
  name: 'MobileSchedule',
  components: { ScheduleDay },
  data() {
    const { days } = useSchedule()
    return { style, days }
  },
  computed: {
    scheduleDays() {
      return this.days.map((day) => ({
        ...day,
        lessons: day.lessons.map((period): ScheduleDayItem => ({
          id: period.start,
          start: period.start,
          end: period.end,
          ...period.lesson,
          typeLabel: period.typeLabel,
        })),
      }))
    },
  },
})
</script>
