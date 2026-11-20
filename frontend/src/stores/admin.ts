import { defineStore } from 'pinia'
import { apiNotifications, ApiError, type ApiPage } from '@/api/client'
import { useAuthStore } from './auth'
import {
  adminRepository,
  catalogPaths,
  normalizeRequest,
  normalizeUser,
  type AdminUserResponse,
} from '@/pages/admin/admin.repository'
import type {
  AdminData,
  AdminUser,
  AdminLesson,
  AdminWorkload,
  CatalogKind,
  CatalogEntry,
  CourseOffering,
  RegistrationRequest,
  RequestQuery,
} from '@/pages/admin/admin.types'

function userPayload(user: AdminUser) {
  return {
    role: user.is_admin ? 'admin' : user.student ? 'student' : 'teacher',
    last_name: user.last_name.trim(),
    first_name: user.first_name.trim(),
    middle_names: user.middle_names.trim() || null,
    email: user.email.trim(),
    passport: user.passport.trim(),
    snils: user.snils.trim(),
    student: user.student
      ? {
          group_id: user.student.group_id,
          record_book_number: user.student.record_book_number.trim() || null,
          is_headman: user.student.is_headman,
        }
      : null,
    teacher: user.teacher ? { department_id: user.teacher.department_id } : null,
  }
}

export function lessonConflicts(
  data: AdminData,
  lesson: AdminLesson,
  peers = data.lessons,
): string[] {
  const workload = data.workloads.find((item) => item.id === lesson.workload_id)
  const course = data.courses.find((item) => item.id === workload?.course_id)
  if (!workload || !course) return ['Выберите существующую нагрузку.']
  const conflicts: string[] = []
  for (const other of peers) {
    if (
      other.id === lesson.id ||
      other.week_number !== lesson.week_number ||
      other.day_of_week !== lesson.day_of_week ||
      lesson.starts_at >= other.ends_at ||
      lesson.ends_at <= other.starts_at
    )
      continue
    const otherWorkload = data.workloads.find((item) => item.id === other.workload_id)
    const otherCourse = data.courses.find((item) => item.id === otherWorkload?.course_id)
    if (otherCourse?.semester_id !== course.semester_id) continue
    const resources = [
      otherCourse?.group_id === course.group_id ? 'группа' : '',
      otherWorkload?.teacher_id === workload.teacher_id ? 'преподаватель' : '',
      other.room_id === lesson.room_id ? 'аудитория' : '',
    ].filter(Boolean)
    if (resources.length)
      conflicts.push(
        `${other.starts_at}–${other.ends_at}: заняты ${resources.join(', ')} (занятие №${other.id}).`,
      )
  }
  return conflicts
}

export const useAdminStore = defineStore('adminStore', {
  state: () => ({
    data: null as AdminData | null,
    loading: false,
    saving: false,
    error: '',
    success: '',
    section: '',
    loadVersion: 0,
    sessionVersion: 0,
    requestsVersion: 0,
    requestsLoading: false,
    requestsTotal: 0,
    pendingRequestsTotal: 0,
    requestsQuery: {
      q: '',
      status: 'pending',
      sort: '-created_at',
      limit: 8,
      offset: 0,
    } as RequestQuery,
  }),
  actions: {
    clearSession() {
      this.loadVersion++
      this.requestsVersion++
      this.sessionVersion++
      this.data = null
      this.loading = false
      this.saving = false
      this.requestsLoading = false
      this.requestsTotal = 0
      this.pendingRequestsTotal = 0
      this.error = ''
      this.success = ''
      this.requestsQuery = { q: '', status: 'pending', sort: '-created_at', limit: 8, offset: 0 }
    },
    async initialize(section = 'admin-requests') {
      const version = ++this.loadVersion
      this.requestsVersion++
      this.requestsLoading = false
      this.section = section
      this.loading = true
      this.error = ''
      this.success = ''
      this.data = null
      try {
        const [data, pending] = await Promise.all([
          adminRepository.load(section),
          useAuthStore().request<ApiPage<RegistrationRequest>>(
            '/registration-requests?status=pending&limit=1&offset=0',
            { paginated: true },
          ),
        ])
        if (version === this.loadVersion) {
          this.data = data
          this.pendingRequestsTotal = pending.meta.total
        }
      } catch (error) {
        if (version === this.loadVersion) {
          apiNotifications.onError?.(error)
          this.error =
            error instanceof Error ? error.message : 'Не удалось загрузить данные с сервера.'
        }
      } finally {
        if (version === this.loadVersion) this.loading = false
      }
    },
    async loadRequests(query?: RequestQuery, fresh = false): Promise<boolean> {
      const filters = query ?? this.requestsQuery
      const version = ++this.requestsVersion
      this.requestsQuery = { ...filters }
      this.requestsLoading = true
      this.error = ''
      if (this.data) this.data.requests = []
      try {
        const params = new URLSearchParams({
          ...filters,
          limit: String(filters.limit),
          offset: String(filters.offset),
        })
        const auth = useAuthStore()
        const [page, pending] = await Promise.all([
          auth.request<ApiPage<RegistrationRequest>>(`/registration-requests?${params}`, {
            paginated: true,
            fresh,
          }),
          auth.request<ApiPage<RegistrationRequest>>(
            '/registration-requests?status=pending&limit=1&offset=0',
            { paginated: true, fresh },
          ),
        ])
        if (version !== this.requestsVersion || !this.data) return false
        this.data.requests = page.data.map(normalizeRequest)
        this.requestsTotal = page.meta.total
        this.pendingRequestsTotal = pending.meta.total
        return true
      } catch (error) {
        if (version === this.requestsVersion) {
          apiNotifications.onError?.(error)
          this.requestsTotal = 0
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить заявки.'
        }
        return false
      } finally {
        if (version === this.requestsVersion) this.requestsLoading = false
      }
    },
    async execute(operation: () => Promise<void>, message: string): Promise<boolean> {
      if (this.saving) return false
      const version = this.sessionVersion
      this.saving = true
      this.error = ''
      this.success = ''
      try {
        await operation()
        if (version !== this.sessionVersion) return false
        this.success = message
        return true
      } catch (error) {
        if (version === this.sessionVersion) {
          apiNotifications.onError?.(error)
          this.error =
            error instanceof Error ? error.message : 'Не удалось сохранить изменения на сервере.'
        }
        return false
      } finally {
        if (version === this.sessionVersion) this.saving = false
      }
    },
    async reviewRequest(id: number, user: AdminUser | null) {
      return this.execute(
        async () => {
          if (user && Boolean(user.student) === Boolean(user.teacher)) {
            throw new Error('Выберите одну учебную роль для регистрации.')
          }
          if (user?.student && !user.student.group_id)
            throw new Error('Выберите учебную группу из списка.')
          if (user?.teacher && !user.teacher.department_id)
            throw new Error('Выберите кафедру из списка.')
          const body = user
            ? {
                decision: 'approve',
                user: {
                  role: user.student ? 'student' : 'teacher',
                  last_name: user.last_name.trim(),
                  first_name: user.first_name.trim(),
                  middle_names: user.middle_names.trim() || null,
                  passport: user.passport.trim(),
                  snils: user.snils.trim(),
                  student: user.student
                    ? {
                        ...user.student,
                        record_book_number: user.student.record_book_number.trim() || null,
                      }
                    : null,
                  teacher: user.teacher,
                },
              }
            : { decision: 'reject' }
          try {
            await useAuthStore().request<RegistrationRequest>(
              `/registration-requests/${id}/decision`,
              { method: 'POST', body: JSON.stringify(body) },
            )
          } catch (error) {
            if (error instanceof ApiError && error.code === 'request_already_reviewed') {
              await this.loadRequests()
            }
            throw error
          }
        },
        user
          ? 'Заявка одобрена. Пользователь создан на сервере и может войти.'
          : 'Заявка отклонена.',
      )
    },
    async saveUser(user: AdminUser) {
      const data = this.data
      return this.execute(
        async () => {
          if (!user.is_admin && Boolean(user.student) === Boolean(user.teacher))
            throw new Error('Выберите одну учебную роль.')
          if (user.student && !user.student.group_id)
            throw new Error('Выберите учебную группу из списка.')
          if (user.teacher && !user.teacher.department_id)
            throw new Error('Выберите кафедру из списка.')
          const saved = await useAuthStore().request<AdminUserResponse>(
            user.id ? `/users/${user.id}` : '/users',
            {
              method: user.id ? 'PATCH' : 'POST',
              body: JSON.stringify(userPayload(user)),
            },
          )
          if (!data) return
          const index = data.users.findIndex((item) => item.id === saved.id)
          if (index < 0) data.users.push(normalizeUser(saved))
          else data.users.splice(index, 1, normalizeUser(saved))
        },
        user.id
          ? 'Пользователь обновлён.'
          : 'Пользователь создан. Приглашение поставлено в очередь отправки.',
      )
    },
    async toggleUser(id: number) {
      const data = this.data
      return this.execute(async () => {
        const user = data?.users.find((item) => item.id === id)
        if (!user) throw new Error('Пользователь не найден.')
        const saved = await useAuthStore().request<AdminUserResponse>(`/users/${id}`, {
          method: 'PATCH',
          body: JSON.stringify({ is_active: !user.is_active }),
        })
        data!.users.splice(data!.users.indexOf(user), 1, normalizeUser(saved))
      }, 'Доступ пользователя обновлён.')
    },
    async saveCatalog(kind: CatalogKind, entry: CatalogEntry) {
      const data = this.data
      return this.execute(async () => {
        const body: Record<string, unknown> =
          kind === 'rooms'
            ? { number: entry.name.trim() }
            : kind === 'semesters'
              ? {}
              : { name: entry.name.trim() }
        if (kind === 'specializations') body.code = entry.code?.trim()
        if (kind === 'groups')
          Object.assign(body, {
            specialization_id: entry.specialization_id,
            admission_year: entry.admission_year,
          })
        if (kind === 'semesters')
          Object.assign(body, {
            academic_year_start: entry.academic_year_start,
            number: entry.number,
          })
        const saved = await useAuthStore().request<CatalogEntry>(
          `${catalogPaths[kind]}${entry.id ? `/${entry.id}` : ''}`,
          {
            method: entry.id ? 'PUT' : 'POST',
            body: JSON.stringify(body),
          },
        )
        if (!data) return
        const entries = data.catalogs[kind]
        const index = entries.findIndex((item) => item.id === saved.id)
        const normalized = kind === 'rooms' ? { ...saved, name: entry.name.trim() } : saved
        if (index < 0) entries.push(normalized)
        else entries.splice(index, 1, normalized)
      }, 'Запись справочника сохранена.')
    },
    async removeCatalog(kind: CatalogKind, id: number) {
      const data = this.data
      return this.execute(async () => {
        await useAuthStore().request<void>(`${catalogPaths[kind]}/${id}`, {
          method: 'DELETE',
        })
        const entries = data?.catalogs[kind]
        const index = entries?.findIndex((item) => item.id === id) ?? -1
        if (entries && index >= 0) entries.splice(index, 1)
      }, 'Запись удалена.')
    },
    async setCurrentSemester(id: number) {
      const data = this.data
      return this.execute(async () => {
        await useAuthStore().request(`/semesters/${id}/current`, { method: 'POST' })
        for (const semester of data?.catalogs.semesters ?? [])
          semester.is_current = semester.id === id
      }, 'Текущий семестр изменён.')
    },
    async saveCourse(course: CourseOffering) {
      const data = this.data
      return this.execute(async () => {
        const saved = await useAuthStore().request<CourseOffering>(
          course.id ? `/courses/${course.id}` : '/courses',
          {
            method: course.id ? 'PATCH' : 'POST',
            body: JSON.stringify(
              course.id
                ? { subject_id: course.subject_id }
                : {
                    group_id: course.group_id,
                    semester_id: course.semester_id,
                    subject_id: course.subject_id,
                  },
            ),
          },
        )
        if (!data) return
        const index = data.courses.findIndex((item) => item.id === saved.id)
        if (index < 0) data.courses.push(saved)
        else data.courses.splice(index, 1, saved)
        course.id = saved.id
      }, 'Назначение дисциплины сохранено.')
    },
    async saveWorkload(workload: AdminWorkload) {
      const data = this.data
      return this.execute(async () => {
        const saved = await useAuthStore().request<AdminWorkload>(
          workload.id ? `/workloads/${workload.id}` : `/courses/${workload.course_id}/workloads`,
          {
            method: workload.id ? 'PUT' : 'POST',
            body: JSON.stringify({
              teacher_id: workload.teacher_id,
              lesson_type: workload.lesson_type,
              planned_hours: workload.planned_hours,
            }),
          },
        )
        if (!data) return
        const index = data.workloads.findIndex((item) => item.id === saved.id)
        if (index < 0) data.workloads.push(saved)
        else data.workloads.splice(index, 1, saved)
      }, 'Нагрузка сохранена.')
    },
    async saveLesson(lesson: AdminLesson) {
      const data = this.data
      return this.execute(async () => {
        const body = {
          week_number: lesson.week_number,
          day_of_week: lesson.day_of_week,
          starts_at: lesson.starts_at,
          room_id: lesson.room_id,
          ...(!lesson.id ? { workload_id: lesson.workload_id } : {}),
        }
        const saved = await useAuthStore().request<AdminLesson>(
          lesson.id ? `/schedule/lessons/${lesson.id}` : '/schedule/lessons',
          { method: lesson.id ? 'PUT' : 'POST', body: JSON.stringify(body) },
        )
        if (!data) return
        const index = data.lessons.findIndex((item) => item.id === saved.id)
        if (index < 0) data.lessons.push(saved)
        else data.lessons.splice(index, 1, saved)
      }, 'Занятие сохранено.')
    },
    async removeAcademic(kind: 'courses' | 'workloads' | 'lessons', id: number) {
      const data = this.data
      return this.execute(async () => {
        await useAuthStore().request<void>(
          `${kind === 'lessons' ? '/schedule/lessons' : `/${kind}`}/${id}`,
          { method: 'DELETE' },
        )
        const entries = data?.[kind]
        const index = entries?.findIndex((item) => item.id === id) ?? -1
        if (entries && index >= 0) entries.splice(index, 1)
      }, 'Запись удалена.')
    },
  },
})
