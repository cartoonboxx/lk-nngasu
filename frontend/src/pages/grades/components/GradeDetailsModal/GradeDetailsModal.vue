<template>
  <Modal
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :width="640"
  >
    <article v-if="assignment" :class="style.details">
      <header :class="style.detailsHeader">
        <div>
          <p :class="style.muted">{{ assignment.subject }}</p>
          <h2 id="grade-detail-heading">{{ assignment.title }}</h2>
          <p :class="style.muted">{{ assignment.topic }}</p>
        </div>
        <IconButton variant="ghost" size="md" round @click="$emit('update:modelValue', false)"
          ><PhX :size="20"
        /></IconButton>
      </header>
      <div :class="style.detailMeta">
        <GradeStatus :status="assignment.status" />
        <span v-if="assignment.grade !== null"
          >Оценка: <strong>{{ assignment.grade }}</strong></span
        >
        <span v-if="assignment.submittedAt" :class="style.muted"
          >Сдано
          {{
            assignment.submittedAt
              ? gradeDateFormatter.format(new Date(`${assignment.submittedAt}T12:00:00`))
              : '—'
          }}</span
        >
      </div>
      <section :class="style.detailSection">
        <h3>Ваш ответ</h3>
        <p>{{ assignment.answer || 'Работа ещё не отправлена.' }}</p>
      </section>
      <section :class="style.detailSection">
        <h3>Комментарий преподавателя</h3>
        <p>
          {{
            assignment.teacherComment ||
            (assignment.status === 'pending'
              ? 'Работа ожидает проверки преподавателя.'
              : 'Комментария пока нет.')
          }}
        </p>
      </section>
    </article>
  </Modal>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhX } from '@phosphor-icons/vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import GradeStatus from '../GradeStatus/GradeStatus.vue'
import type { GradeAssignment } from '../../grades.types'
import { gradeDateFormatter } from '../../grades.format'
import style from './GradeDetailsModal.module.scss'

export default defineComponent({
  name: 'GradeDetailsModal',
  components: { Modal, IconButton, PhX, GradeStatus },
  props: {
    modelValue: { type: Boolean, default: false },
    assignment: { type: Object as PropType<GradeAssignment | null>, default: null },
  },
  emits: { 'update:modelValue': (value: boolean) => typeof value === 'boolean' },
  data() {
    return { style, gradeDateFormatter }
  },
})
</script>
