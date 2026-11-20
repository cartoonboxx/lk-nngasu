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

      <MediaButton
        v-for="file in assignment.files"
        :key="file.id"
        @click="download(file.id, file.name)"
        :title="file.name"
        :filename="file.name"
      />
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
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import { apiNotifications } from '@/api/client'
import { defineComponent, type PropType } from 'vue'
import { PhX } from '@phosphor-icons/vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import { useFilesStore } from '@/stores/files'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import GradeStatus from '../GradeStatus/GradeStatus.vue'
import type { GradeAssignment } from '../../grades.types'
import { gradeDateFormatter } from '../../grades.format'
import style from './GradeDetailsModal.module.scss'

export default defineComponent({
  name: 'GradeDetailsModal',
  components: { MediaButton, Modal, IconButton, PhX, GradeStatus },
  props: {
    modelValue: { type: Boolean, default: false },
    assignment: { type: Object as PropType<GradeAssignment | null>, default: null },
  },
  watch: {
    modelValue() {},
  },
  methods: {
    async download(id: number, name: string) {
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        apiNotifications.onError?.(error)
      }
    },
  },
  emits: { 'update:modelValue': (value: boolean) => typeof value === 'boolean' },
  data() {
    return { style, gradeDateFormatter }
  },
})
</script>
