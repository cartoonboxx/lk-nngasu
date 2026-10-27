<template>
  <Modal
    :model-value="modelValue"
    :width="800"
    :style="{ '--modal-content-padding': '0px' }"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div :class="style.layout">
      <header :class="style.header">
        <div :class="style.heading">
          <p :class="style.eyebrow">
            {{ courseTitle }} <span>· Тема {{ topicIndex + 1 }}</span>
          </p>
          <div :class="style.titleRow">
            <h2>Практическая работа №{{ topicIndex + 1 }}</h2>
            <p
              :class="[
                style.status,
                reviewed ? style.reviewed : null,
                answer && !reviewed ? style.pending : null,
              ]"
            >
              {{ reviewed ? 'Работа проверена' : answer ? 'Ожидает проверки' : 'Не отправлено' }}
            </p>
          </div>
          <p :class="style.deadline">
            <PhCalendarBlank :size="23" />
            <span
              >Срок сдачи: <strong>{{ topic.deadline }}, 23:59</strong></span
            >
          </p>
        </div>
        <IconButton
          :class="style.close"
          variant="ghost"
          size="md"
          round
          @click="$emit('update:modelValue', false)"
        >
          <PhX :size="24" />
        </IconButton>
      </header>

      <section :class="style.task">
        <h3>Задание</h3>
        <p>{{ topic.task }}</p>
      </section>

      <form v-if="isEditing && !reviewed" :class="style.form" @submit.prevent="submit">
        <div :class="style.field">
          <div :class="style.labelRow">
            <label :for="`${id}-comment`"
              >Комментарий <span :class="style.srOnly">к ответу</span></label
            >
            <span :class="style.muted">Необязательно</span>
          </div>
          <Input
            :id="`${id}-comment`"
            v-model="comment"
            multiline
            full-width
            :base-height="104"
            :maxlength="2000"
            placeholder="Добавьте пояснение к работе"
          />
          <p :class="style.counter">{{ comment.length }} / 2000</p>
        </div>
        <div :class="style.field">
          <label :for="`${id}-files`">Файлы работы</label>
          <div
            :class="[style.dropzone, dragging ? style.dragging : null]"
            @dragover.prevent="dragging = true"
            @dragleave="dragging = false"
            @drop.prevent="onDrop"
          >
            <PhUploadSimple :size="34" />
            <div :class="style.uploadCopy">
              <p>Перетащите файлы сюда</p>
              <span>или выберите на устройстве</span>
            </div>
            <input
              :id="`${id}-files`"
              :class="style.fileInput"
              type="file"
              multiple
              @change="addFiles"
            />
          </div>
          <p :class="style.muted">До 10 файлов, не более 20 МБ каждый</p>
          <ul v-if="attachments.length" :class="style.files">
            <li v-for="attachment in attachments" :key="attachment.id">
              <MediaButton :title="attachment.name" :meta="formatSize(attachment.size)" readonly>
                <template #icon>
                  <component
                    :is="fileIcon(attachment.name)"
                    :class="attachment.name.toLowerCase().endsWith('.pdf') ? style.pdf : null"
                    :size="34"
                    weight="duotone"
                  />
                </template>
                <template #actions>
                  <IconButton variant="ghost" size="sm" @click="removeFile(attachment.id)">
                    <PhX :size="20" />
                  </IconButton>
                </template>
              </MediaButton>
            </li>
          </ul>
        </div>
        <p v-if="error" :class="style.error">{{ error }}</p>
        <p :class="style.sessionNote">
          Демонстрационный режим: ответ хранится до перезагрузки страницы.
        </p>
        <footer :class="style.actions">
          <Button
            variant="secondary"
            @click="answer ? cancelEditing() : $emit('update:modelValue', false)"
            >{{ answer ? 'Отменить' : 'Отмена' }}</Button
          >
          <Button
            type="submit"
            :disabled="(!attachments.length && !comment.trim()) || comment.length > 2000"
          >
            {{ answer ? 'Сохранить изменения' : 'Отправить преподавателю' }}
          </Button>
        </footer>
      </form>

      <section v-else-if="answer" :class="style.form">
        <div :class="style.answerHeading">
          <h3>Ваш ответ</h3>
          <p :class="style.muted">Отправлен {{ formatDate(answer.submittedAt) }}</p>
        </div>
        <p v-if="answer.comment" :class="style.comment">{{ answer.comment }}</p>
        <ul v-if="answer.attachments.length" :class="style.files">
          <li v-for="attachment in answer.attachments" :key="attachment.id">
            <MediaButton
              :title="attachment.name"
              :meta="formatSize(attachment.size)"
              :readonly="!attachment.file && !attachment.url"
              @click="download(attachment)"
            >
              <template #icon>
                <component
                  :is="fileIcon(attachment.name)"
                  :class="attachment.name.toLowerCase().endsWith('.pdf') ? style.pdf : null"
                  :size="34"
                  weight="duotone"
                />
              </template>
            </MediaButton>
          </li>
        </ul>
        <section v-if="answer.review" :class="style.review">
          <div :class="style.gradeBlock">
            <p :class="style.muted">Оценка работы</p>
            <p :class="style.grade">{{ answer.review.grade }}</p>
          </div>
          <div :class="style.feedback">
            <h3>Комментарий преподавателя</h3>
            <p :class="style.comment">
              {{ answer.review.comment || 'Преподаватель не оставил комментарий.' }}
            </p>
          </div>
        </section>
        <p :class="style.sessionNote">
          Демонстрационный режим: ответ хранится до перезагрузки страницы.
        </p>
        <footer v-if="!reviewed && !confirmDelete" :class="style.actions">
          <Button variant="danger" @click="confirmDelete = true">Удалить ответ</Button>
          <Button variant="secondary" @click="startEditing">Редактировать ответ</Button>
        </footer>
        <div v-if="confirmDelete && !reviewed" :class="style.confirm">
          <p>Удалить ответ вместе с файлами и комментарием?</p>
          <div :class="style.confirmActions">
            <Button variant="secondary" @click="confirmDelete = false">Отмена</Button>
            <Button variant="danger" @click="deleteAnswer">Подтвердить удаление</Button>
          </div>
        </div>
        <footer v-if="reviewed" :class="style.actions">
          <Button variant="secondary" @click="$emit('update:modelValue', false)">Закрыть</Button>
        </footer>
      </section>
    </div>
  </Modal>
</template>

<script lang="ts">
import { defineComponent, useId, type PropType } from 'vue'
import {
  PhCalendarBlank,
  PhFileCode,
  PhFilePdf,
  PhFileText,
  PhUploadSimple,
  PhX,
} from '@phosphor-icons/vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import type { CourseTopic } from '../../course.types'
import type {
  AssignmentAnswer,
  AssignmentAnswerDraft,
  AssignmentAttachment,
} from './assignment.types'
import style from './AssignmentModal.module.scss'

export default defineComponent({
  name: 'AssignmentModal',
  components: {
    Modal,
    Button,
    MediaButton,
    IconButton,
    Input,
    PhCalendarBlank,
    PhFileCode,
    PhFilePdf,
    PhFileText,
    PhUploadSimple,
    PhX,
  },
  props: {
    courseTitle: { type: String, default: 'Задание курса' },
    modelValue: {
      type: Boolean,
      default: false,
    },
    topic: {
      type: Object as PropType<CourseTopic>,
      required: true,
    },
    topicIndex: {
      type: Number,
      required: true,
    },
    answer: {
      type: Object as PropType<AssignmentAnswer | null>,
      default: null,
    },
  },
  emits: {
    'update:modelValue': (value: boolean) => typeof value === 'boolean',
    submit: (answer: AssignmentAnswerDraft) => !!answer,
    delete: () => true,
  },
  data() {
    return {
      style,
      id: useId(),
      isEditing: !this.answer,
      comment: this.answer?.comment ?? '',
      attachments: [...(this.answer?.attachments ?? [])] as AssignmentAttachment[],
      error: '',
      confirmDelete: false,
      dragging: false,
    }
  },
  computed: {
    reviewed(): boolean {
      return !!this.answer?.review
    },
  },
  watch: {
    modelValue(open: boolean) {
      if (open) this.reset()
    },
    answer: {
      deep: true,
      handler() {
        this.reset()
      },
    },
    'topic.id'() {
      this.reset()
    },
  },
  methods: {
    reset() {
      this.comment = this.answer?.comment ?? ''
      this.attachments = [...(this.answer?.attachments ?? [])]
      this.isEditing = !this.answer
      this.confirmDelete = false
      this.error = ''
      this.dragging = false
    },
    startEditing() {
      this.reset()
      this.isEditing = true
    },
    cancelEditing() {
      this.reset()
    },
    addFiles(event: Event) {
      const input = event.target as HTMLInputElement
      this.appendFiles(Array.from(input.files ?? []))
      input.value = ''
    },
    onDrop(event: DragEvent) {
      this.dragging = false
      this.appendFiles(Array.from(event.dataTransfer?.files ?? []))
    },
    appendFiles(files: File[]) {
      const errors: string[] = []
      for (const file of files) {
        if (
          this.attachments.some(
            (item) =>
              item.file &&
              item.name === file.name &&
              item.size === file.size &&
              item.file.lastModified === file.lastModified,
          )
        )
          continue
        if (file.size > 20 * 1024 * 1024) {
          errors.push(`Файл «${file.name}» превышает 20 МБ.`)
          continue
        }
        if (this.attachments.length >= 10) {
          errors.push('Можно прикрепить не более 10 файлов.')
          break
        }
        this.attachments.push({ id: crypto.randomUUID(), name: file.name, size: file.size, file })
      }
      this.error = errors.join(' ')
    },
    removeFile(id: string) {
      this.attachments = this.attachments.filter((item) => item.id !== id)
      this.error = ''
    },
    submit() {
      if (this.reviewed) return
      if (this.comment.length > 2000) {
        this.error = 'Комментарий должен содержать не более 2000 символов.'
        return
      }
      if (!this.attachments.length && !this.comment.trim()) {
        this.error = 'Прикрепите файлы или напишите комментарий.'
        return
      }
      this.$emit('submit', { attachments: [...this.attachments], comment: this.comment.trim() })
    },
    deleteAnswer() {
      if (!this.reviewed) this.$emit('delete')
    },
    fileIcon(name: string): string {
      if (name.toLowerCase().endsWith('.pdf')) return 'PhFilePdf'
      if (/\.(pl|py|js|ts|html|css|cpp|c|java|json)$/i.test(name)) return 'PhFileCode'
      return 'PhFileText'
    },
    formatSize(size: number): string {
      if (size < 1024) return `${size} Б`
      if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} КБ`
      return `${(size / (1024 * 1024)).toFixed(1)} МБ`
    },
    formatDate(date: string): string {
      return new Date(date).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
    },
    download(attachment: AssignmentAttachment) {
      const url = attachment.file ? URL.createObjectURL(attachment.file) : attachment.url
      if (!url) return
      const link = document.createElement('a')
      link.href = url
      link.download = attachment.name
      link.click()
      if (attachment.file) setTimeout(() => URL.revokeObjectURL(url), 1000)
    },
  },
})
</script>
