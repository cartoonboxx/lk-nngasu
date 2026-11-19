import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { useMessageStore } from './messages'
import type { News, NewsDetail, GradeRow, Dashboard } from '@/api/cabinet.types'
import type { GradeAssignment } from '@/pages/grades/grades.types'

export const useCabinetStore = defineStore('cabinet', {
  state: () => ({
    news: [] as News[],
    currentNews: null as NewsDetail | null,
    dashboard: null as Dashboard | null,
    grades: [] as GradeAssignment[],
    newsLoading: false,
    detailLoading: false,
    dashboardLoading: false,
    gradesLoading: false,
    newsError: '',
    detailError: '',
    dashboardError: '',
    gradesError: '',
    version: 0,
    detailVersion: 0,
  }),
  actions: {
    async fetchNews() {
      if (this.newsLoading) return
      const version = this.version
      this.newsLoading = true
      this.newsError = ''
      try {
        const news = await useAuthStore().list<News>('/news')
        if (version === this.version) this.news = news
      } catch (error) {
        if (version === this.version)
          this.newsError = error instanceof Error ? error.message : 'Не удалось загрузить новости.'
      } finally {
        if (version === this.version) this.newsLoading = false
      }
    },
    async fetchNewsDetail(id: string) {
      const version = ++this.detailVersion
      this.detailLoading = true
      this.detailError = ''
      this.currentNews = null
      try {
        const news = await useAuthStore().request<NewsDetail>(`/news/${encodeURIComponent(id)}`)
        if (version === this.detailVersion) this.currentNews = news
      } catch (error) {
        if (version === this.detailVersion)
          this.detailError =
            error instanceof Error ? error.message : 'Не удалось загрузить новость.'
      } finally {
        if (version === this.detailVersion) this.detailLoading = false
      }
    },
    async fetchDashboard() {
      if (this.dashboardLoading) return
      const version = this.version
      this.dashboardLoading = true
      this.dashboardError = ''
      try {
        const data = await useAuthStore().request<Dashboard>('/dashboard')
        if (version !== this.version) return
        this.dashboard = data
        this.news = data.news
        useMessageStore().applyChats(data.chats)
      } catch (error) {
        if (version === this.version)
          this.dashboardError =
            error instanceof Error ? error.message : 'Не удалось загрузить главную.'
      } finally {
        if (version === this.version) this.dashboardLoading = false
      }
    },
    async fetchGrades() {
      if (this.gradesLoading) return
      const version = this.version
      this.gradesLoading = true
      this.gradesError = ''
      this.grades = []
      try {
        const auth = useAuthStore()
        const semesters = await auth.list<{ id: number }>('/semesters')
        const pages = await Promise.all(
          semesters.map((semester) => auth.list<GradeRow>(`/grades?semester_id=${semester.id}`)),
        )
        if (version !== this.version) return
        this.grades = pages.flat().map((row) => ({
          id: row.assignment_id,
          title: row.assignment_title,
          topic: row.topic_title,
          subject: row.subject_name,
          semester: row.semester_name,
          submittedAt: row.submitted_at,
          grade: row.grade,
          status:
            row.status === 'reviewed'
              ? 'reviewed'
              : row.status === 'submitted'
                ? 'pending'
                : row.status === 'missing'
                  ? 'missing'
                  : 'not-submitted',
          teacherComment: row.review_comment ?? '',
          files: [],
          answer: '',
        }))
      } catch (error) {
        if (version === this.version)
          this.gradesError = error instanceof Error ? error.message : 'Не удалось загрузить оценки.'
      } finally {
        if (version === this.version) this.gradesLoading = false
      }
    },
    clearSession() {
      this.version++
      this.detailVersion++
      this.news = []
      this.currentNews = null
      this.dashboard = null
      this.grades = []
      this.newsLoading = false
      this.detailLoading = false
      this.dashboardLoading = false
      this.gradesLoading = false
      this.newsError = ''
      this.detailError = ''
      this.dashboardError = ''
      this.gradesError = ''
    },
  },
})
