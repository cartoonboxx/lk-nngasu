<template>
  <section :class="style.mobileSchedule">
    <section v-for="day in days" :key="day.date" :class="style.daySchedule">
      <header :class="[style.dayHeader, day.isToday ? style.today : null]">
        <h2>
          <time :datetime="day.date">{{ day.fullLabel }}</time>
        </h2>
      </header>
      <ul v-if="day.lessons.length" :class="style.dayLessons">
        <li
          v-for="period in day.lessons"
          :key="period.start"
          :class="[style.dayLesson, style[period.lesson.type]]"
        >
          <div :class="style.lessonTime">
            <span>{{ period.start }}</span
            ><span :class="style.timeDash"> – </span><span>{{ period.end }}</span>
          </div>
          <div :class="style.lessonDetails">
            <h3>{{ period.lesson.subject }}</h3>
            <p :class="style.lessonMeta">
              <span>{{ period.typeLabel }}</span
              ><span>·</span><span>{{ period.lesson.location }}</span>
            </p>
            <p :class="style.lessonTeacher">
              <PhUser :size="13" weight="fill" />{{ period.lesson.teacher }}
            </p>
          </div>
        </li>
      </ul>
      <p v-else :class="style.emptyDay">На этот день занятий нет</p>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './MobileSchedule.module.scss'
import { PhUser } from '@phosphor-icons/vue'
import { useSchedule } from '../../useSchedule'

export default defineComponent({
  name: 'MobileSchedule',
  components: { PhUser },
  data() {
    const { days } = useSchedule()
    return { style, days }
  },
})
</script>
