<template>
  <div :class="style.tableScroll" tabindex="0">
    <p v-if="!$scheduleStore.days.length" :class="style.empty">На эту неделю занятий нет</p>
    <table v-else :class="style.timetable">
      <thead>
        <tr>
          <th :class="style.timeHeading" />
          <th
            v-for="day in $scheduleStore.days"
            :key="day.date"
            :class="
              new Date(`${day.date}T00:00:00`).toDateString() === new Date().toDateString()
                ? style.today
                : null
            "
          >
            {{
              new Date(`${day.date}T00:00:00`)
                .toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' })
                .replace(/\.$/, '')
                .replace(/^./, (letter) => letter.toUpperCase())
            }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(period, index) in rows" :key="period.start">
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
import { periods } from '../../schedule.data'

export default defineComponent({
  name: 'ScheduleTable',
  components: { LessonCard },
  data() {
    return { style }
  },
  computed: {
    rows() {
      return periods.map((period) => ({
        ...period,
        lessons: this.$scheduleStore.days.map((day) =>
          day.lessons.find((lesson) => lesson.start === period.start),
        ),
      }))
    },
  },
})
</script>
