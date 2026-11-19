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
          <p :class="style.eyebrow">{{ courseTitle }}</p>
          <h2>{{ assignment?.title || 'Задание' }}</h2>
          <p v-if="assignment?.deadline_at" :class="style.deadline">
            Срок сдачи: {{ new Date(assignment.deadline_at).toLocaleString('ru-RU') }}
          </p>
        </div>
        <Button variant="secondary" :disabled="saving" @click="$emit('update:modelValue', false)"
          >Закрыть</Button
        >
      </header>
      <p v-if="loading">Загрузка задания…</p>

      <Button v-if="!loading && !assignment" variant="secondary" @click="load">Повторить</Button>
      <template v-if="assignment && !loading">
        <section :class="style.task">
          <h3>Задание</h3>
          <p>{{ assignment.description }}</p>
          <Button
            v-for="file in assignment.files"
            :key="file.id"
            variant="secondary"
            @click="download(file.id, file.name)"
            >{{ file.name }}</Button
          >
        </section>
        <section v-if="assignment.submission?.status === 'reviewed'" :class="style.review">
          <h3>Оценка: {{ assignment.submission.grade }}</h3>
          <p>{{ assignment.submission.review_comment }}</p>
        </section>
        <p v-if="success" role="status">{{ success }}</p>
        <form v-if="auth.user?.student" :class="style.form" @submit.prevent="submit(true)">
          <p v-if="assignment.submission?.submitted_at">
            Отправлено: {{ new Date(assignment.submission.submitted_at).toLocaleString('ru-RU') }}
          </p>
          <TextInput
            v-model="comment"
            label="Ответ"
            multiline
            full-width
            :maxlength="2000"
            :disabled="saving || reviewed"
          />
          <ul :class="style.files">
            <li v-for="attachment in attachments" :key="attachment.id">
              <Button variant="secondary" type="button" @click="downloadAttachment(attachment)">{{
                attachment.name
              }}</Button>
              <Button
                v-if="!reviewed"
                variant="ghost"
                type="button"
                :disabled="saving"
                @click="attachments = attachments.filter((item) => item.id !== attachment.id)"
                >Убрать</Button
              >
            </li>
          </ul>
          <label v-if="!reviewed" :class="style.field"
            >Прикрепить файлы
            <input
              type="file"
              multiple
              accept=".pdf,.docx,.pptx,.xlsx,.xls,.zip,.rar"
              :disabled="saving"
              @change="addFiles"
            />
          </label>
          <p v-if="!reviewed" :class="style.muted">
            До 5 файлов: PDF, DOCX, PPTX, XLSX, XLS, ZIP, RAR. До 15 МБ каждый.
          </p>
          <footer v-if="!reviewed" :class="style.actions">
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
      <h2>Удалить ответ?</h2>
      <p>
        Текст и прикреплённые файлы ответа будут удалены. После этого можно отправить новый ответ.
      </p>
      <Button variant="ghost" :disabled="saving" @click="withdrawOpen = false">Отмена</Button>
      <Button variant="danger" :loading="saving" @click="withdraw">Удалить</Button>
    </Modal>
  </Modal>
</template>
<script lang="ts">
import { apiNotifications } from '@/api/client'
import { defineComponent } from 'vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import { useAuthStore } from '@/stores/auth'
import { useFilesStore } from '@/stores/files'
import type { Assignment } from '@/api/cabinet.types'
import type { AssignmentAttachment } from '../../course.types'
import style from './AssignmentModal.module.scss'
export default defineComponent({
  name: 'AssignmentModal',
  components: { Modal, Button, TextInput },
  props: {
    modelValue: Boolean,
    assignmentId: { type: Number, required: true },
    courseTitle: { type: String, default: '' },
  },
  emits: ['update:modelValue', 'submitted'],
  data() {
    return {
      style,
      assignment: null as Assignment | null,
      attachments: [] as AssignmentAttachment[],
      comment: '',
      loading: false,
      saving: false,
      success: '',
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

      this.success = ''
      this.assignment = null
      this.attachments = []
      this.comment = ''
      try {
        const assignment = await this.$subjectsStore.getAssignment(this.assignmentId)
        if (version !== this.version) return
        this.assignment = assignment
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

      for (const file of Array.from(input.files ?? [])) {
        if (
          this.attachments.length >= 5 ||
          file.size > 15 * 1024 * 1024 ||
          !/\.(pdf|docx|pptx|xlsx|xls|zip|rar)$/i.test(file.name)
        ) {
          this.$utilityStore.notifyError('Проверьте тип, размер и количество файлов.')
          continue
        }
        this.attachments.push({ id: crypto.randomUUID(), name: file.name, size: file.size, file })
      }
      input.value = ''
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

      this.success = ''
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
        this.success = send ? 'Ответ отправлен на проверку.' : 'Черновик сохранён на сервере.'
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
