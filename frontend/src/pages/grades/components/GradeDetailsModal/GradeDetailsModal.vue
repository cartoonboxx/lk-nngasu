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
              ? gradeDateFormatter.format(new Date(assignment.submittedAt))
              : '—'
          }}</span
        >
      </div>
      <section :class="style.detailSection">
        <h3>Ваш ответ</h3>
        <p>
          {{
            assignment.answer ||
            (assignment.submittedAt ? 'Текст ответа отсутствует.' : 'Работа ещё не отправлена.')
          }}
        </p>
      </section>
      <p v-if="error" role="alert">{{ error }}</p>
      <Button
        v-for="file in assignment.files"
        :key="file.id"
        variant="secondary"
        @click="download(file.id, file.name)"
        >{{ file.name }}</Button
      >
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
import Button from '@/components/UI/buttons/Button/Button.vue'
import { useFilesStore } from '@/stores/files'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import GradeStatus from '../GradeStatus/GradeStatus.vue'
import type { GradeAssignment } from '../../grades.types'
import { gradeDateFormatter } from '../../grades.format'
import style from './GradeDetailsModal.module.scss'

export default defineComponent({
  name: 'GradeDetailsModal',
  components: { Modal, IconButton, Button, PhX, GradeStatus },
  props: {
    modelValue: { type: Boolean, default: false },
    assignment: { type: Object as PropType<GradeAssignment | null>, default: null },
  },
  watch: {
    modelValue() {
      this.error = ''
    },
  },
  methods: {
    async download(id: number, name: string) {
      this.error = ''
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось скачать файл.'
      }
    },
  },
  emits: { 'update:modelValue': (value: boolean) => typeof value === 'boolean' },
  data() {
    return { style, gradeDateFormatter, error: '' }
  },
})
</script>
