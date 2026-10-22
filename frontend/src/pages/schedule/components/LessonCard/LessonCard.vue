<template>
  <article :class="[style.card, style[lesson.type]]">
    <h3 :class="style.subject">{{ lesson.subject }}</h3>
    <span :class="style.type">{{ typeLabel }}</span>
    <p :class="style.location">{{ lesson.location }}</p>
    <p :class="style.teacher"><PhUser :size="12" weight="fill" />{{ lesson.teacher }}</p>
  </article>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhUser } from '@phosphor-icons/vue'
import { lessonTypes, type ScheduleLesson } from '../../schedule.data'
import style from './LessonCard.module.scss'

export default defineComponent({
  name: 'LessonCard',
  components: { PhUser },
  props: {
    lesson: { type: Object as PropType<ScheduleLesson>, required: true },
  },
  data() {
    return { style }
  },
  computed: {
    typeLabel() {
      return lessonTypes.find((item) => item.type === this.lesson.type)?.label
    },
  },
})
</script>
