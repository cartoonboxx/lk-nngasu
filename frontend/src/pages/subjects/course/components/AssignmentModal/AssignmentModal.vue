<template>
  <Modal
    :style="{ '--modal-content-padding': '0px' }"
    :model-value="modelValue"
    :width="800"
    @update:model-value="!saving && $emit('update:modelValue', $event)"
  >
    <div :class="style.layout">
      <header :class="style.header">
        <div :class="style.heading">
          <p :class="style.eyebrow">
            {{ courseTitle }} <span v-if="topicIndex >= 0">· Тема {{ topicIndex + 1 }}</span>
          </p>
          <div :class="style.titleRow">
            <h2>{{ assignment?.title || 'Задание' }}</h2>
            <StatusBadge
              v-if="assignment"
              :variant="
                reviewed ? 'success' : assignment.submission?.submitted_at ? 'info' : 'neutral'
              "
            >
              {{
                reviewed
                  ? 'Работа проверена'
                  : assignment.submission?.submitted_at
                    ? 'Ожидает проверки'
                    : assignment.submission
                      ? 'Черновик'
                      : 'Не отправлено'
              }}
            </StatusBadge>
          </div>
          <p v-if="assignment?.deadline_at" :class="style.deadline">
            <PhCalendarBlank :size="23" />
            <span
              >Срок сдачи:
              <strong>{{ new Date(assignment.deadline_at).toLocaleString('ru-RU') }}</strong></span
            >
          </p>
        </div>
        <IconButton
          :class="style.close"
          variant="ghost"
          size="md"
          round
          title="Закрыть"
          :disabled="saving"
          @click="$emit('update:modelValue', false)"
          ><PhX :size="24"
        /></IconButton>
      </header>
      <p v-if="loading">Загрузка задания…</p>

      <Button v-if="!loading && !assignment" variant="secondary" @click="load">Повторить</Button>
      <template v-if="assignment && !loading">
        <section :class="style.task">
          <h3>Задание</h3>
          <p>{{ assignment.description }}</p>
          <MediaButton
            v-for="file in assignment.files"
            :key="file.id"
            @click="download(file.id, file.name)"
            :title="file.name"
            :filename="file.name"
          />
        </section>
        <section v-if="assignment.submission && !isEditing" :class="style.form">
          <div :class="style.answerHeading">
            <h3>Ваш ответ</h3>
            <p v-if="assignment.submission.submitted_at" :class="style.muted">
              Отправлен {{ new Date(assignment.submission.submitted_at).toLocaleString('ru-RU') }}
            </p>
          </div>
          <p v-if="assignment.submission.answer_text" :class="style.comment">
            {{ assignment.submission.answer_text }}
          </p>
          <ul v-if="assignment.submission.files.length" :class="style.files">
            <li v-for="file in assignment.submission.files" :key="file.id">
              <MediaButton
                :title="file.name"
                :filename="file.name"
                @click="download(file.id, file.name)"
              />
            </li>
          </ul>
          <section v-if="reviewed" :class="style.review">
            <div :class="style.gradeBlock">
              <p :class="style.muted">Оценка работы</p>
              <p :class="style.grade">{{ assignment.submission.grade }}</p>
            </div>
            <div :class="style.feedback">
              <h3>Комментарий преподавателя</h3>
              <p :class="style.comment">
                {{
                  assignment.submission.review_comment || 'Преподаватель не оставил комментарий.'
                }}
              </p>
            </div>
          </section>
          <footer :class="style.actions">
            <template v-if="!reviewed && auth.user?.student"
              ><Button variant="danger" :disabled="saving" @click="withdrawOpen = true"
                >Удалить ответ</Button
              ><Button variant="secondary" @click="isEditing = true"
                >Редактировать ответ</Button
              ></template
            >
            <Button v-else variant="secondary" @click="$emit('update:modelValue', false)"
              >Закрыть</Button
            >
          </footer>
        </section>
        <form
          v-if="auth.user?.student && isEditing && !reviewed"
          :class="style.form"
          @submit.prevent="submit(true)"
        >
          <div :class="style.field">
            <div :class="style.labelRow">
              <label :for="`${id}-comment`">Комментарий к ответу</label
              ><span :class="style.muted">Необязательно</span>
            </div>
            <Input
              :id="`${id}-comment`"
              v-model="comment"
              multiline
              full-width
              :base-height="104"
              :maxlength="2000"
              :disabled="saving"
              placeholder="Добавьте пояснение к работе"
            />
            <p :class="style.counter">{{ comment.length }} / 2000</p>
          </div>
          <div :class="style.field">
            <label :for="`${id}-files`">Файлы работы</label>
            <div
              :class="[style.dropzone, dragging && style.dragging]"
              @dragover.prevent="!saving && (dragging = true)"
              @dragleave="dragging = false"
              @drop.prevent="
                (event: DragEvent) => {
                  dragging = false
                  if (!saving) appendFiles(Array.from(event.dataTransfer?.files ?? []))
                }
              "
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
                :disabled="saving"
                @change="addFiles"
              />
            </div>
            <p :class="style.muted">До 5 файлов любого типа без ограничения размера.</p>
            <ul v-if="attachments.length" :class="style.files">
              <li v-for="attachment in attachments" :key="attachment.id">
                <MediaButton
                  :title="attachment.name"
                  :filename="attachment.name"
                  :meta="attachment.size ? `${(attachment.size / 1024).toFixed(1)} КБ` : ''"
                  @click="downloadAttachment(attachment)"
                >
                  <template #actions
                    ><IconButton
                      variant="ghost"
                      size="sm"
                      round
                      :disabled="saving"
                      @click="attachments = attachments.filter((item) => item.id !== attachment.id)"
                      ><PhX :size="20" /></IconButton
                  ></template>
                </MediaButton>
              </li>
            </ul>
          </div>
          <footer v-if="!reviewed" :class="style.actions">
            <Button
              variant="secondary"
              :disabled="saving"
              @click="assignment.submission ? resetDraft() : $emit('update:modelValue', false)"
              >Отмена</Button
            >
            <Button
              v-if="!assignment.submission?.submitted_at"
              variant="secondary"
              :disabled="saving"
              @click="submit(false)"
              >Сохранить черновик</Button
            >
            <Button
              v-if="assignment.submission"
              variant="ghost"
              :disabled="saving"
              @click="withdrawOpen = true"
              >Удалить ответ</Button
            >
            <Button
              type="submit"
              :loading="saving"
              :disabled="saving || (!comment.trim() && !attachments.length)"
              >Отправить преподавателю</Button
            >
          </footer>
        </form>
      </template>
    </div>
    <Modal
      :model-value="withdrawOpen"
      :width="400"
      @update:model-value="!saving && (withdrawOpen = $event)"
    >
      <div :class="style.confirm">
        <h2>Удалить ответ?</h2>
        <p>
          Текст и прикреплённые файлы ответа будут удалены. После этого можно отправить новый ответ.
        </p>
        <div :class="style.confirmActions">
          <Button variant="ghost" :disabled="saving" @click="withdrawOpen = false">Отмена</Button>
          <Button variant="danger" :loading="saving" @click="withdraw">Удалить</Button>
        </div>
      </div>
    </Modal>
  </Modal>
</template>
<script lang="ts">
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import { apiNotifications } from '@/api/client'
import { defineComponent, useId } from 'vue'
import { PhCalendarBlank, PhUploadSimple, PhX } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import StatusBadge from '@/components/UI/StatusBadge/StatusBadge.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import { useAuthStore } from '@/stores/auth'
import { useFilesStore } from '@/stores/files'
import type { Assignment } from '@/api/cabinet.types'
import type { AssignmentAttachment } from '../../course.types'
import style from './AssignmentModal.module.scss'
export default defineComponent({
  name: 'AssignmentModal',
  components: {
    MediaButton,
    Modal,
    Button,
    Input,
    IconButton,
    StatusBadge,
    PhCalendarBlank,
    PhUploadSimple,
    PhX,
  },
  props: {
    modelValue: Boolean,
    assignmentId: { type: Number, required: true },
    courseTitle: { type: String, default: '' },
    topicIndex: { type: Number, default: -1 },
  },
  emits: ['update:modelValue', 'submitted'],
  data() {
    return {
      style,
      id: useId(),
      dragging: false,
      isEditing: true,
      assignment: null as Assignment | null,
      attachments: [] as AssignmentAttachment[],
      comment: '',
      loading: false,
      saving: false,
      withdrawOpen: false,
      version: 0,
    }
  },
  computed: {
    auth() {
      return useAuthStore()
    },
    reviewed(): boolean {
      return this.assignment?.submission?.status === 'reviewed'
    },
  },
  watch: {
    modelValue: {
      immediate: true,
      async handler(open: boolean) {
        if (open) await this.load()
        else this.version++
      },
    },
  },
  beforeUnmount() {
    this.version++
  },
  methods: {
    async load() {
      const version = ++this.version
      this.loading = true

      this.withdrawOpen = false
      this.assignment = null
      this.attachments = []
      this.comment = ''
      try {
        const assignment = await this.$subjectsStore.getAssignment(this.assignmentId)
        if (version !== this.version) return
        this.assignment = assignment
        this.isEditing = !assignment.submission
        this.dragging = false
        this.comment = assignment.submission?.answer_text ?? ''
        this.attachments = (assignment.submission?.files ?? []).map((file) => ({
          id: String(file.id),
          name: file.name,
          size: 0,
        }))
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    addFiles(event: Event) {
      const input = event.target as HTMLInputElement

      this.appendFiles(Array.from(input.files ?? []))
      input.value = ''
    },
    appendFiles(files: File[]) {
      if (this.saving || this.reviewed) return
      for (const file of files) {
        if (this.attachments.length >= 5) {
          this.$utilityStore.notifyError('Можно прикрепить до 5 файлов.')
          continue
        }
        this.attachments.push({ id: crypto.randomUUID(), name: file.name, size: file.size, file })
      }
    },
    resetDraft() {
      this.comment = this.assignment?.submission?.answer_text ?? ''
      this.attachments = (this.assignment?.submission?.files ?? []).map((file) => ({
        id: String(file.id),
        name: file.name,
        size: 0,
      }))
      this.isEditing = false
      this.dragging = false
    },
    async downloadAttachment(attachment: AssignmentAttachment) {
      if (attachment.file) {
        const url = URL.createObjectURL(attachment.file)
        const link = document.createElement('a')
        link.href = url
        link.download = attachment.name
        link.click()
        window.setTimeout(() => URL.revokeObjectURL(url), 1000)
      } else await this.download(Number(attachment.id), attachment.name)
    },
    async download(id: number, name: string) {
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        apiNotifications.onError?.(error)
      }
    },
    async submit(send = true) {
      if (this.saving || this.reviewed || !this.assignment) return
      const version = this.version
      this.saving = true

      try {
        const submission = await this.$subjectsStore.submitAnswer(
          this.assignment.id,
          {
            attachments: this.attachments,
            comment: this.comment.trim(),
          },
          send,
        )
        if (version !== this.version) return
        this.assignment.submission = submission
        this.isEditing = false
        this.$utilityStore.notify(
          send ? 'Ответ отправлен на проверку.' : 'Черновик сохранён на сервере.',
          'success',
        )
        this.$emit('submitted')
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
      }
    },
    async withdraw() {
      if (this.saving || this.reviewed || !this.assignment) return
      const version = this.version
      this.saving = true
      try {
        await this.auth.request(`/assignments/${this.assignment.id}/submission`, {
          method: 'DELETE',
        })
        if (version !== this.version) return
        this.withdrawOpen = false
        this.$emit('submitted')
        await this.load()
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
      }
    },
  },
})
</script>
