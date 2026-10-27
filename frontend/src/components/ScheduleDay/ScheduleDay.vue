<template>
  <section :class="[style.daySchedule, layout === 'grid' && style.grid]">
    <header :class="[style.dayHeader, today ? style.today : null]">
      <slot name="header" :title="title" :date="date" :today="today">
        <h2>
          <time v-if="date" :datetime="date">{{ title }}</time>
          <template v-else>{{ title }}</template>
        </h2>
      </slot>
      <div v-if="$slots['header-actions']" :class="style.headerActions">
        <slot name="header-actions" />
      </div>
    </header>
    <ul v-if="lessons.length" :class="style.dayLessons">
      <li
        v-for="lesson in lessons"
        :key="lesson.id"
        :class="[
          style.dayLesson,
          style[lesson.type],
          $slots['lesson-trailing'] && style.withTrailing,
        ]"
      >
        <div :class="style.lessonTime">
          <span>{{ lesson.start }}</span
          ><span :class="style.timeDash"> – </span><span>{{ lesson.end }}</span>
        </div>
        <div :class="style.lessonDetails">
          <h3>{{ lesson.subject }}</h3>
          <p :class="style.lessonMeta">
            <span v-if="lesson.typeLabel">{{ lesson.typeLabel }}</span>
            <span v-if="lesson.typeLabel && lesson.location" aria-hidden="true">·</span>
            <span v-if="lesson.location">{{ lesson.location }}</span>
          </p>
          <p v-if="lesson.teacher" :class="style.lessonTeacher">
            <PhUser :size="13" weight="fill" aria-hidden="true" />{{ lesson.teacher }}
          </p>
        </div>
        <div v-if="$slots['lesson-trailing']" :class="style.lessonTrailing">
          <slot name="lesson-trailing" :lesson="lesson" />
        </div>
      </li>
    </ul>
    <div v-else :class="style.emptyDay">
      <slot name="empty">На этот день занятий нет</slot>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhUser } from '@phosphor-icons/vue'
import type { ScheduleDayItem } from './schedule-day.types'
import style from './ScheduleDay.module.scss'

export default defineComponent({
  name: 'ScheduleDay',
  components: { PhUser },
  props: {
    layout: {
      type: String as PropType<'list' | 'grid'>,
      default: 'list',
    },
    title: {
      type: String,
      default: '',
    },
    date: {
      type: String,
      default: undefined,
    },
    today: {
      type: Boolean,
      default: false,
    },
    lessons: {
      type: Array as PropType<readonly ScheduleDayItem[]>,
      required: true,
    },
  },
  data() {
    return { style }
  },
})
</script>
