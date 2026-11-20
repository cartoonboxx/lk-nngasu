<template>
  <AdminDialog
    :model-value="modelValue"
    title="Материалы и задания курса"
    :width="960"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <p v-if="loading">Загрузка курса…</p>
    <p v-if="error"><Button variant="ghost" @click="load">Повторить</Button></p>
    <template v-if="course && !loading">
      <h3>{{ course.subject_name }} · {{ course.group_name }}</h3>
      <Button :disabled="saving" @click="edit('topic')">Добавить тему</Button>
      <p v-if="!course.topics.length">Темы пока не добавлены.</p>
      <section v-for="topic in course.topics" :key="topic.id" :class="style.form">
        <h3>{{ topic.title }}</h3>
        <p>{{ topic.description }}</p>
        <div :class="style.actions">
          <Button size="sm" variant="ghost" :disabled="saving" @click="edit('topic', topic)"
            >Изменить тему</Button
          >
          <Button
            size="sm"
            variant="ghost"
            :disabled="saving"
            @click="removal = { path: `/topics/${topic.id}`, label: 'тему и её материалы' }"
            >Удалить тему</Button
          >
          <Button size="sm" variant="secondary" :disabled="saving" @click="edit('material', topic)"
            >Добавить материал</Button
          >
          <Button
            v-if="!topic.assignment"
            size="sm"
            variant="secondary"
            :disabled="saving"
            @click="edit('assignment', topic)"
            >Добавить задание</Button
          >
        </div>
        <div v-for="material in topic.materials" :key="material.id" :class="style.actions">
          <Button
            v-if="material.file"
            size="sm"
            variant="ghost"
            @click="download(material.file.id, material.file.name)"
            >{{ material.file.name }}</Button
          >
          <a
            v-if="material.external_url"
            :href="material.external_url"
            target="_blank"
            rel="noopener noreferrer"
            >{{ material.external_url }}</a
          >
          <Button
            size="sm"
            variant="ghost"
            :disabled="saving"
            @click="edit('material', topic, material)"
            >Изменить</Button
          >
          <Button
            size="sm"
            variant="ghost"
            :disabled="saving"
            @click="removal = { path: `/materials/${material.id}`, label: 'материал' }"
            >Удалить</Button
          >
        </div>
        <template v-if="topic.assignment">
          <h4>{{ topic.assignment.title }}</h4>
          <p>{{ topic.assignment.description }}</p>
          <p>
            Срок:
            {{
              topic.assignment.deadline_at
                ? new Date(topic.assignment.deadline_at).toLocaleString('ru-RU')
                : 'без срока'
            }}
          </p>
          <div :class="style.actions">
            <Button size="sm" variant="ghost" :disabled="saving" @click="edit('assignment', topic)"
              >Изменить задание</Button
            >
            <Button
              size="sm"
              variant="ghost"
              :disabled="saving"
              @click="removal = { path: `/assignments/${topic.assignment.id}`, label: 'задание' }"
              >Удалить задание</Button
            >
            <Button
              size="sm"
              variant="secondary"
              :disabled="saving"
              @click="showWorks(topic.assignment)"
              >Работы студентов</Button
            >
          </div>
        </template>
      </section>
    </template>
    <AdminDialog
      :model-value="editorOpen"
      :title="
        kind === 'topic' ? 'Тема курса' : kind === 'material' ? 'Учебный материал' : 'Задание'
      "
      @update:model-value="!saving && (editorOpen = $event)"
    >
      <form :class="style.form" @submit.prevent="save">
        <TextInput
          v-if="kind !== 'material'"
          v-model="draft.title"
          label="Название"
          required
          full-width
          :maxlength="200"
          :disabled="saving"
        />
        <TextInput
          v-if="kind !== 'material'"
          v-model="draft.description"
          label="Описание"
          multiline
          full-width
          :maxlength="5000"
          :disabled="saving"
        />
        <TextInput
          v-if="kind === 'material'"
          v-model="draft.url"
          label="Внешняя ссылка"
          type="url"
          full-width
          :disabled="saving"
        />
        <label v-if="kind === 'assignment'" :class="style.field"
          >Срок сдачи<input v-model="draft.deadline" type="datetime-local" :disabled="saving"
        /></label>
        <template v-if="kind !== 'topic'">
          <ul>
            <li v-for="(file, index) in draft.files" :key="index">
              {{ file.name }}
              <button type="button" :disabled="saving" @click="draft.files.splice(index, 1)">
                Убрать
              </button>
            </li>
          </ul>
          <label :class="style.field"
            >Файлы<input
              type="file"
              :multiple="kind === 'assignment'"
              accept=".pdf,.docx,.pptx,.xlsx,.xls,.zip,.rar"
              :disabled="saving"
              @change="addFiles"
          /></label>
          <p :class="style.muted">
            PDF, DOCX, PPTX, XLSX, XLS, ZIP, RAR, до 15 МБ каждый.
            {{ kind === 'material' ? 'Один файл или внешняя ссылка.' : 'До пяти файлов.' }}
          </p>
        </template>
        <div :class="style.dialogFooter">
          <Button variant="ghost" :disabled="saving" @click="editorOpen = false">Отмена</Button
          ><Button type="submit" :loading="saving">Сохранить</Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog
      :model-value="!!removal"
      title="Удалить запись?"
      @update:model-value="!saving && !$event && (removal = null)"
    >
      <p>
        Удалить {{ removal?.label }}? Действие нельзя отменить. Сервер запретит удаление, если
        запись используется.
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" :disabled="saving" @click="removal = null">Отмена</Button
        ><Button variant="danger" :loading="saving" @click="remove">Удалить</Button>
      </div>
    </AdminDialog>
    <AdminDialog v-model="worksOpen" title="Работы студентов" :width="900">
      <p v-if="worksLoading">Загрузка работ…</p>
      <template v-else-if="worksAssignment">
        <Button
          variant="secondary"
          @click="download(`/assignments/${worksAssignment.id}/submissions/archive`, 'works.zip')"
          >Скачать все работы ZIP</Button
        >
        <AdminTable
          :columns="['Студент', 'Статус', 'Оценка', 'Работа']"
          :total="works.length"
          v-slot="{ start, end }"
        >
          <tr v-for="row in works.slice(start, end)" :key="row.student_id">
            <td>{{ row.student_name }}</td>
            <td>{{ submissionStatuses[row.status] ?? row.status }}</td>
            <td>{{ row.grade ?? '—' }}</td>
            <td>
              <Button
                v-if="row.submission_id"
                size="sm"
                variant="ghost"
                @click="showSubmission(row.submission_id)"
                >Посмотреть</Button
              >
            </td>
          </tr>
        </AdminTable>
      </template>
    </AdminDialog>
    <AdminDialog
      :model-value="!!submission"
      title="Ответ студента"
      @update:model-value="!$event && (submission = null)"
    >
      <template v-if="submission"
        ><p style="white-space: pre-wrap">{{ submission.answer_text }}</p>
        <p>Оценка: {{ submission.grade ?? '—' }}</p>
        <p>{{ submission.review_comment }}</p>
        <Button
          v-for="file in submission.files"
          :key="file.id"
          variant="ghost"
          @click="download(file.id, file.name)"
          >{{ file.name }}</Button
        >
        <Button
          variant="secondary"
          @click="download(`/submissions/${submission.id}/archive`, 'submission.zip')"
          >Скачать ZIP</Button
        >
      </template>
    </AdminDialog>
  </AdminDialog>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useFilesStore } from '@/stores/files'
import { apiNotifications } from '@/api/client'
import type { Assignment, CourseDetail, Material, Submission, Topic } from '@/api/cabinet.types'
import Button from '@/components/UI/buttons/Button/Button.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import AdminDialog from './AdminDialog.vue'
import AdminTable from './AdminTable.vue'
import style from '../Admin.module.scss'

type WorkRow = {
  student_id: number
  student_name: string
  submission_id: number | null
  status: string
  grade: number | null
}

export default defineComponent({
  components: { Button, TextInput, AdminDialog, AdminTable },
  props: { modelValue: Boolean, courseId: { type: Number, required: true } },
  emits: ['update:modelValue'],
  data() {
    return {
      style,
      course: null as CourseDetail | null,
      loading: false,
      error: '',
      saving: false,
      version: 0,
      editorOpen: false,
      kind: 'topic' as 'topic' | 'material' | 'assignment',
      draft: {
        id: 0,
        topicId: 0,
        title: '',
        description: '',
        url: '',
        deadline: '',
        files: [] as { id?: number; name: string; file?: File }[],
      },
      removal: null as { path: string; label: string } | null,
      worksOpen: false,
      worksLoading: false,
      works: [] as WorkRow[],
      worksAssignment: null as Assignment | null,
      submission: null as Submission | null,
      submissionStatuses: {
        not_submitted: 'Не отправлено',
        missing: 'Просрочено',
        submitted: 'На проверке',
        reviewed: 'Проверено',
      } as Record<string, string>,
    }
  },
  watch: {
    modelValue: {
      immediate: true,
      handler(open: boolean) {
        if (open) void this.load()
        else {
          this.version++
          this.editorOpen = false
          this.worksOpen = false
          this.submission = null
          this.removal = null
        }
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
      this.error = ''
      try {
        const course = await useAuthStore().request<CourseDetail>(`/courses/${this.courseId}`)
        if (version === this.version) this.course = course
      } catch (error) {
        if (version === this.version) {
          apiNotifications.onError?.(error)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить курс.'
        }
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    edit(kind: 'topic' | 'material' | 'assignment', topic?: Topic, material?: Material) {
      this.kind = kind
      const assignment = kind === 'assignment' ? topic?.assignment : null
      const deadline = assignment?.deadline_at ? new Date(assignment.deadline_at) : null
      if (deadline) deadline.setMinutes(deadline.getMinutes() - deadline.getTimezoneOffset())
      this.draft = {
        id:
          kind === 'topic'
            ? (topic?.id ?? 0)
            : kind === 'material'
              ? (material?.id ?? 0)
              : (assignment?.id ?? 0),
        topicId: topic?.id ?? 0,
        title: kind === 'topic' ? (topic?.title ?? '') : (assignment?.title ?? ''),
        description:
          kind === 'topic' ? (topic?.description ?? '') : (assignment?.description ?? ''),
        url: material?.external_url ?? '',
        deadline: deadline?.toISOString().slice(0, 16) ?? '',
        files:
          kind === 'material'
            ? material?.file
              ? [{ id: material.file.id, name: material.file.name }]
              : []
            : (assignment?.files ?? []).map((file) => ({ id: file.id, name: file.name })),
      }
      this.editorOpen = true
    },
    addFiles(event: Event) {
      const input = event.target as HTMLInputElement
      for (const file of Array.from(input.files ?? [])) {
        if (
          this.draft.files.length >= (this.kind === 'material' ? 1 : 5) ||
          file.size > 15 * 1024 * 1024 ||
          !/\.(pdf|docx|pptx|xlsx|xls|zip|rar)$/i.test(file.name)
        ) {
          this.$utilityStore.notifyError('Проверьте тип, размер и количество файлов.')
          continue
        }
        this.draft.files.push({ name: file.name, file })
      }
      input.value = ''
    },
    async save() {
      if (this.saving) return
      const version = this.version
      this.saving = true
      this.$adminStore.saving = true
      try {
        for (const file of this.draft.files) {
          if (file.file && !file.id) file.id = (await useFilesStore().upload(file.file)).id
          if (version !== this.version) return
        }
        let path: string
        let body: Record<string, unknown>
        if (this.kind === 'topic') {
          path = this.draft.id ? `/topics/${this.draft.id}` : `/courses/${this.courseId}/topics`
          body = {
            title: this.draft.title.trim(),
            description: this.draft.description.trim() || null,
          }
        } else if (this.kind === 'material') {
          path = this.draft.id
            ? `/materials/${this.draft.id}`
            : `/topics/${this.draft.topicId}/materials`
          body = {
            file_id: this.draft.files[0]?.id ?? null,
            external_url: this.draft.url.trim() || null,
          }
        } else {
          path = this.draft.id
            ? `/assignments/${this.draft.id}`
            : `/courses/${this.courseId}/assignments`
          body = {
            topic_id: this.draft.topicId,
            title: this.draft.title.trim(),
            description: this.draft.description.trim() || null,
            deadline_at: this.draft.deadline ? new Date(this.draft.deadline).toISOString() : null,
            file_ids: this.draft.files.map((file) => file.id!),
          }
        }
        await useAuthStore().request(path, {
          method: this.draft.id ? 'PUT' : 'POST',
          body: JSON.stringify(body),
        })
        if (version !== this.version) return
        this.editorOpen = false
        await this.load()
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
        this.$adminStore.saving = false
      }
    },
    async remove() {
      if (this.saving || !this.removal) return
      const version = this.version
      this.saving = true
      this.$adminStore.saving = true
      try {
        await useAuthStore().request(this.removal.path, { method: 'DELETE' })
        if (version !== this.version) return
        this.removal = null
        await this.load()
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
        this.$adminStore.saving = false
      }
    },
    async showWorks(assignment: Assignment) {
      const version = this.version
      this.worksAssignment = assignment
      this.works = []
      this.worksOpen = true
      this.worksLoading = true
      try {
        const works = await useAuthStore().list<WorkRow>(
          `/assignments/${assignment.id}/submissions`,
        )
        if (version === this.version && this.worksAssignment?.id === assignment.id)
          this.works = works
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        if (version === this.version && this.worksAssignment?.id === assignment.id)
          this.worksLoading = false
      }
    },
    async showSubmission(id: number) {
      const version = this.version
      try {
        const submission = await useAuthStore().request<Submission>(`/submissions/${id}`)
        if (version === this.version) this.submission = submission
      } catch (error) {
        apiNotifications.onError?.(error)
      }
    },
    async download(id: number | string, name: string) {
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        apiNotifications.onError?.(error)
      }
    },
  },
})
</script>
