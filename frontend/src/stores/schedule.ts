import { apiNotifications, ApiError } from '@/api/client'
import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import type { CalendarLesson } from '@/api/cabinet.types'
import type { ScheduleDay, LessonType } from '@/pages/schedule/schedule.types'

export const useScheduleStore = defineStore('schedule', {
  state: () => {
    const weekStart = new Date()
    weekStart.setHours(0, 0, 0, 0)
    weekStart.setDate(weekStart.getDate() - ((weekStart.getDay() + 6) % 7))
    return {
      schedule: [] as ScheduleDay[],
      weekStart,
      loading: false,
      error: '',
      requestVersion: 0,
      groups: [] as { id: number; name: string }[],
      groupId: '',
      groupsError: '',
    }
  },
  getters: {
    days: (state) => state.schedule,
    periods: (state) =>
      [
        ...new Map(
          state.schedule
            .flatMap((day) => day.lessons)
            .map((lesson) => [
              `${lesson.start}-${lesson.end}`,
              { start: lesson.start, end: lesson.end },
            ]),
        ).values(),
      ].sort((a, b) => a.start.localeCompare(b.start)),
  },
  actions: {
    async fetchGroups() {
      const userId = useAuthStore().user?.id
      this.groupsError = ''
      try {
        const groups = await useAuthStore().list<{ id: number; name: string }>('/schedule/groups')
        if (useAuthStore().user?.id === userId) this.groups = groups
      } catch (error) {
        if (useAuthStore().user?.id === userId) {
          apiNotifications.onError?.(error)
          this.groupsError = error instanceof Error ? error.message : 'Не удалось загрузить группы.'
        }
      }
    },
    async fetchSchedule(from?: Date, to?: Date) {
      from = from ?? this.weekStart
      const version = ++this.requestVersion
      this.loading = true
      this.error = ''
      this.schedule = []
      const end = to ? new Date(to) : new Date(from)
      if (!to) end.setDate(end.getDate() + 6)
      const startDate = `${from.getFullYear()}-${String(from.getMonth() + 1).padStart(2, '0')}-${String(from.getDate()).padStart(2, '0')}`
      const endDate = `${end.getFullYear()}-${String(end.getMonth() + 1).padStart(2, '0')}-${String(end.getDate()).padStart(2, '0')}`
      try {
        const lessons = await useAuthStore().request<CalendarLesson[]>(
          `/schedule?date_from=${startDate}&date_to=${endDate}${this.groupId ? `&group_id=${this.groupId}` : ''}`,
        )
        if (version !== this.requestVersion) return
        const days = new Map<string, ScheduleDay>()
        for (const lesson of lessons) {
          if (!lesson.date) continue
          const day = days.get(lesson.date) ?? { date: lesson.date, lessons: [] }
          day.lessons.push({
            id: lesson.id,
            start: lesson.starts_at.slice(0, 5),
            end: lesson.ends_at.slice(0, 5),
            subject: lesson.subject,
            type: (lesson.lesson_type === 'laboratory' ? 'lab' : lesson.lesson_type) as LessonType,
            location: lesson.room,
            teacher: lesson.teacher_name,
          })
          days.set(lesson.date, day)
        }
        this.schedule = [...days.values()].sort((a, b) => a.date.localeCompare(b.date))
        this.schedule.forEach((day) => day.lessons.sort((a, b) => a.start.localeCompare(b.start)))
      } catch (error) {
        if (error instanceof ApiError && error.code === 'current_semester_not_set') return
        if (version === this.requestVersion) {
          apiNotifications.onError?.(error)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить расписание.'
        }
      } finally {
        if (version === this.requestVersion) this.loading = false
      }
    },
    shiftWeek(direction: number) {
      const date = new Date(this.weekStart)
      date.setDate(date.getDate() + direction * 7)
      this.weekStart = date
      void this.fetchSchedule()
    },
    selectWeek(date: Date) {
      const start = new Date(date)
      start.setHours(0, 0, 0, 0)
      start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
      this.weekStart = start
      void this.fetchSchedule()
    },
    clearSession() {
      this.requestVersion++
      this.schedule = []
      this.groups = []
      this.groupId = ''
      this.groupsError = ''
      this.loading = false
      this.error = ''
      const start = new Date()
      start.setHours(0, 0, 0, 0)
      start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
      this.weekStart = start
    },
  },
})
