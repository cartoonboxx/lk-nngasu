import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import type { Assignment, Course, News, GradeReview } from '@/api/cabinet.types'
import type { GradeAssignment } from '@/pages/grades/grades.types'

export const useCabinetStore = defineStore('cabinet', {
  state: () => ({
    news: [] as News[],
    grades: [] as GradeAssignment[],
    newsLoading: false,
    gradesLoading: false,
    newsError: '',
    gradesError: '',
    version: 0,
  }),
  actions: {
    async fetchNews() {
      if (this.newsLoading) return
      const version = this.version
      this.newsLoading = true
      this.newsError = ''
      this.news = []
      try {
        const news = await useAuthStore().request<News[]>('/news?limit=50')
        if (version === this.version) this.news = news
      } catch (error) {
        if (version === this.version)
          this.newsError = error instanceof Error ? error.message : 'Не удалось загрузить новости.'
      } finally {
        if (version === this.version) this.newsLoading = false
      }
    },
    async fetchGrades() {
      const version = this.version
      this.gradesLoading = true
      this.gradesError = ''
      this.grades = []
      try {
        const [courses, reviews] = await Promise.all([
          useAuthStore().request<(Course & { assignments: Assignment[] })[]>('/grades'),
          useAuthStore().request<GradeReview[]>('/me/grades'),
        ])
        const comments = new Map(
          reviews.map((review) => [review.assignment_id, review.review_comment]),
        )
        if (version !== this.version) return
        this.grades = courses.flatMap((course) =>
          course.assignments.map((assignment) => ({
            id: assignment.id,
            title: assignment.title,
            topic: '',
            subject: course.subject_name,
            semester: course.semester_name,
            submittedAt: assignment.mine?.submitted_at ?? null,
            grade: assignment.mine?.grade ?? null,
            status:
              assignment.mine?.status === 'reviewed'
                ? 'reviewed'
                : assignment.mine?.status === 'submitted'
                  ? 'pending'
                  : 'not-submitted',
            teacherComment: comments.get(assignment.id) ?? '',
            files: [],
            answer: '',
          })),
        )
      } catch (error) {
        if (version === this.version)
          this.gradesError = error instanceof Error ? error.message : 'Не удалось загрузить оценки.'
      } finally {
        if (version === this.version) this.gradesLoading = false
      }
    },
    clearSession() {
      this.version++
      this.news = []
      this.grades = []
      this.newsLoading = false
      this.gradesLoading = false
      this.newsError = ''
      this.gradesError = ''
    },
  },
})
