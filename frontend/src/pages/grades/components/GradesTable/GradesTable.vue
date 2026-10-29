<template>
  <section :class="style.results">
    <div :class="style.resultsHeader">
      <h2 id="grades-heading">Оценки за задания</h2>
      <span>Показано {{ assignments.length }} из {{ total }}</span>
    </div>
    <table v-if="assignments.length" :class="style.table">
      <caption :class="style.srOnly">
        Задания, оценки и комментарии преподавателей
      </caption>
      <thead>
        <tr>
          <th scope="col">Задание</th>
          <th scope="col">Дисциплина</th>
          <th scope="col">Сдано</th>
          <th scope="col">Оценка</th>
          <th scope="col">Статус</th>
          <th scope="col">Комментарий преподавателя</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="assignment in assignments" :key="assignment.id">
          <td :class="style.taskCell">
            <button type="button" :class="style.task" @click="$emit('select', assignment)">
              <span>{{ assignment.title }}</span>
              <small>{{ assignment.topic }}</small>
            </button>
          </td>
          <td data-label="Дисциплина">{{ assignment.subject }}</td>
          <td data-label="Сдано" :class="style.date">
            {{
              assignment.submittedAt
                ? gradeDateFormatter.format(new Date(`${assignment.submittedAt}T12:00:00`))
                : '—'
            }}
          </td>
          <td data-label="Оценка">
            <span
              v-if="assignment.grade !== null"
              :class="[style.grade, assignment.grade >= 5 ? style.excellent : style.good]"
              >{{ assignment.grade }}</span
            >
            <span v-else :class="style.muted">—</span>
          </td>
          <td data-label="Статус">
            <GradeStatus :status="assignment.status" />
          </td>
          <td data-label="Комментарий" :class="style.comment">
            {{ assignment.teacherComment || '—' }}
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else :class="style.empty" role="status">
      <PhMagnifyingGlass :size="32" />
      <h3>Задания не найдены</h3>
      <p>Попробуйте изменить запрос или выбрать другие фильтры.</p>
      <Button variant="secondary" @click="$emit('reset')">Сбросить фильтры</Button>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import GradeStatus from '../GradeStatus/GradeStatus.vue'
import type { GradeAssignment } from '../../grades.types'
import { gradeDateFormatter } from '../../grades.format'
import style from './GradesTable.module.scss'

export default defineComponent({
  name: 'GradesTable',
  components: { Button, PhMagnifyingGlass, GradeStatus },
  props: {
    assignments: { type: Array as PropType<GradeAssignment[]>, required: true },
    total: { type: Number, required: true },
  },
  emits: { select: (assignment: GradeAssignment) => Boolean(assignment), reset: () => true },
  data() {
    return { style, gradeDateFormatter }
  },
})
</script>
