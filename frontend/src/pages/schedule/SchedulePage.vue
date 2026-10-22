<template>
  <section :class="style.schedulePage">
    <SectionHeader header="Расписание" paragraph="Расписание группы ПРИ23/1" />
    <div :class="style.filterBlock">
      <div :class="style.dailyPart">
        <IconButton variant="ghost" @click="shiftWeek(-1)">
          <PhCaretLeft />
        </IconButton>
        <span :class="style.weekLabel">{{ weekLabel }}</span>
        <IconButton variant="ghost" @click="shiftWeek(1)">
          <PhCaretRight />
        </IconButton>
        <Button variant="primary" @click="resetWeek">Сегодня</Button>
        <IconButton variant="ghost">
          <PhCalendar />
        </IconButton>
      </div>
      <Input placeholder="Введите группу" :class="style.inputGroup">
        <template #prefix>
          <PhMagnifyingGlass />
        </template>
      </Input>
    </div>

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
    <ul :class="style.legend">
      <li v-for="item in lessonTypes" :key="item.type">
        <span :class="[style.legendDot, style[item.type]]" />{{ item.label }}
      </li>
    </ul>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SchedulePage.module.scss'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import LessonCard from './components/LessonCard/LessonCard.vue'
import { lessonTypes, periods } from './schedule.data'
import { PhCalendar, PhCaretLeft, PhCaretRight, PhMagnifyingGlass } from '@phosphor-icons/vue'

function monday(date: Date) {
  const start = new Date(date)
  start.setHours(0, 0, 0, 0)
  start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
  return start
}

export default defineComponent({
  name: 'SchedulePage',
  components: {
    LessonCard,
    PhCalendar,
    PhMagnifyingGlass,
    PhCaretRight,
    PhCaretLeft,
    Input,
    Button,
    IconButton,
    SectionHeader,
  },
  data() {
    return {
      style,
      periods,
      lessonTypes,
      weekStart: monday(new Date()),
    }
  },
  computed: {
    days() {
      const today = new Date().toDateString()
      return Array.from({ length: 6 }, (_, index) => {
        const date = new Date(this.weekStart)
        date.setDate(date.getDate() + index)
        return {
          date: date.toDateString(),
          label: date
            .toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' })
            .replace(/\.$/, '')
            .replace(/^./, (letter) => letter.toUpperCase()),
          isToday: date.toDateString() === today,
        }
      })
    },
    weekLabel() {
      const end = new Date(this.weekStart)
      end.setDate(end.getDate() + 5)
      const endLabel = end.toLocaleDateString('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      })
      if (
        this.weekStart.getMonth() === end.getMonth() &&
        this.weekStart.getFullYear() === end.getFullYear()
      ) {
        return `${this.weekStart.getDate()} – ${endLabel}`
      }
      return `${this.weekStart.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })} – ${endLabel}`
    },
  },
  methods: {
    shiftWeek(direction: number) {
      const date = new Date(this.weekStart)
      date.setDate(date.getDate() + direction * 7)
      this.weekStart = date
    },
    resetWeek() {
      this.weekStart = monday(new Date())
    },
  },
})
</script>
