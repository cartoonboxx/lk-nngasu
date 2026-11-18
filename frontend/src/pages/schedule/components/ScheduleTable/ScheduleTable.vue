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
        <tr v-for="(period, index) in rows" :key="`${period.start}-${period.end}`">
          <th :class="style.period">
            <span :class="style.periodNumber">{{ index + 1 }} пара</span>
            <span :class="style.periodTime">{{ period.start }} – {{ period.end }}</span>
          </th>
          <td v-for="(lessons, dayIndex) in period.lessons" :key="dayIndex">
            <LessonCard v-for="lesson in lessons" :key="lesson.id" :lesson="lesson" />
            <span v-if="!lessons.length" :class="style.empty"></span>
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

export default defineComponent({
  name: 'ScheduleTable',
  components: { LessonCard },
  data() {
    return { style }
  },
  computed: {
    rows() {
      return this.$scheduleStore.periods.map((period) => ({
        ...period,
        lessons: this.$scheduleStore.days.map((day) =>
          day.lessons.filter(
            (lesson) => lesson.start === period.start && lesson.end === period.end,
          ),
        ),
      }))
    },
  },
})
</script>
