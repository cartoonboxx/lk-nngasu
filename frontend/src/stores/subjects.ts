import { apiNotifications } from '@/api/client'
import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { useFilesStore } from './files'
import type { Assignment, Course, CourseDetail, Submission } from '@/api/cabinet.types'
import type { Subject, SubjectDetails } from '@/pages/subjects/subjects.types'
import type { AssignmentAnswerDraft } from '@/pages/subjects/course/course.types'

export const useSubjectsStore = defineStore('subjects', {
  state: () => ({
    subjects: [] as Subject[],
    currentSubject: null as SubjectDetails | null,
    currentAssignment: null as Assignment | null,
    assignmentVersion: 0,
    loading: false,
    error: '',
    requestVersion: 0,
  }),
  getters: {
    topics: (state) => state.currentSubject?.topics ?? [],
  },
  actions: {
    async fetchSubjects() {
      const version = ++this.requestVersion
      this.loading = true
      this.error = ''
      this.subjects = []
      try {
        const courses = await useAuthStore().list<Course>('/courses')
        if (version !== this.requestVersion) return
        this.subjects = courses.map((course) => ({
          id: course.id,
          label: `${course.subject_name} · ${course.group_name}`,
          value: String(course.id),
          semester: String(course.semester_id),
          semesterLabel: course.semester_name,
        }))
      } catch (error) {
        if (version === this.requestVersion) {
          apiNotifications.onError?.(error)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить предметы.'
        }
      } finally {
        if (version === this.requestVersion) this.loading = false
      }
    },
    async fetchSubject(id: number | string) {
      const version = ++this.requestVersion
      this.loading = true
      this.error = ''
      this.currentSubject = null
      try {
        const course = await useAuthStore().request<CourseDetail>(
          `/courses/${encodeURIComponent(id)}`,
        )
        if (version !== this.requestVersion) return
        this.currentSubject = {
          id: course.id,
          label: course.subject_name,
          value: String(course.id),
          semester: String(course.semester_id),
          semesterLabel: course.semester_name,
          topics: course.topics,
        }
      } catch (error) {
        if (version === this.requestVersion) {
          apiNotifications.onError?.(error)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить курс.'
        }
      } finally {
        if (version === this.requestVersion) this.loading = false
      }
    },
    async getAssignment(id: number) {
      const version = ++this.assignmentVersion
      this.currentAssignment = null
      const assignment = await useAuthStore().request<Assignment>(`/assignments/${id}`)
      if (version !== this.assignmentVersion) throw new Error('Загрузка задания отменена.')
      this.currentAssignment = assignment
      return assignment
    },
    async submitAnswer(id: number, draft: AssignmentAnswerDraft, submit = true) {
      const version = this.requestVersion
      const fileIds: number[] = []
      for (const attachment of draft.attachments) {
        if (version !== this.requestVersion) throw new Error('Страница курса закрыта.')
        if (attachment.file) {
          const uploaded = await useFilesStore().upload(attachment.file)
          // Сохраняем id в черновике: повтор после ошибки не загружает файл заново.
          attachment.id = String(uploaded.id)
          delete attachment.file
        }
        fileIds.push(Number(attachment.id))
      }
      if (version !== this.requestVersion) throw new Error('Страница курса закрыта.')
      const submission = await useAuthStore().request<Submission>(`/assignments/${id}/submission`, {
        method: 'PUT',
        body: JSON.stringify({ answer_text: draft.comment || null, file_ids: fileIds, submit }),
      })
      if (version === this.requestVersion) {
        const assignment = this.currentSubject?.topics.find(
          (topic) => topic.assignment?.id === id,
        )?.assignment
        if (assignment)
          assignment.mine = {
            status: submission.status,
            submitted_at: submission.submitted_at,
            grade: submission.grade,
          }
      }
      return submission
    },
    clearSession() {
      this.requestVersion++
      this.assignmentVersion++
      this.currentAssignment = null
      this.subjects = []
      this.currentSubject = null
      this.loading = false
      this.error = ''
    },
  },
})
