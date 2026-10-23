<template>
  <div :class="style.tableScroll" tabindex="0">
    <table :class="style.timetable">
      <thead>
        <tr>
          <th :class="style.timeHeading" />
          <th v-for="day in days" :key="day.date" :class="day.isToday ? style.today : null">
            {{ day.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(period, index) in periods" :key="period.start">
          <th :class="style.period">
            <span :class="style.periodNumber">{{ index + 1 }} пара</span>
            <span :class="style.periodTime">{{ period.start }} – {{ period.end }}</span>
          </th>
          <td v-for="(lesson, dayIndex) in period.lessons" :key="dayIndex">
            <LessonCard v-if="lesson" :lesson="lesson" />
            <span v-else :class="style.empty">—</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './ScheduleTable.module.scss'
import LessonCard from '../LessonCard/LessonCard.vue'
import { useSchedule } from '../../useSchedule'

export default defineComponent({
  name: 'ScheduleTable',
  components: { LessonCard },
  data() {
    const { days, periods } = useSchedule()
    return { style, days, periods }
  },
})
</script>
