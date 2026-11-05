import { defineStore } from 'pinia'
import { adminRepository } from '@/pages/admin/admin.repository'
import type {
  AdminData,
  AdminUser,
  AdminNews,
  AdminLesson,
  AdminWorkload,
  CatalogKind,
  CatalogEntry,
  CourseOffering,
} from '@/pages/admin/admin.types'

function validateUser(data: AdminData, user: AdminUser) {
  if (!user.last_name.trim() || !user.first_name.trim()) throw new Error('Укажите имя и фамилию.')
  if (user.last_name.length > 50 || user.first_name.length > 50 || user.middle_names.length > 50)
    throw new Error('Имя, фамилия и отчество — не более 50 символов каждое.')
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(user.email) || user.email.length > 255)
    throw new Error('Укажите корректный email.')
  if (
    data.users.some(
      (item) => item.id !== user.id && item.email.toLowerCase() === user.email.toLowerCase(),
    )
  )
    throw new Error('Пользователь с таким email уже существует.')
  if (user.student) {
    if (!data.catalogs.groups.some((item) => item.id === user.student?.group_id && !item.archived))
      throw new Error('Выберите действующую учебную группу.')
    if (!user.student.record_book_number.trim() || user.student.record_book_number.length > 50)
      throw new Error('Укажите номер зачётной книжки до 50 символов.')
    if (
      data.users.some(
        (item) =>
          item.id !== user.id &&
          item.student?.record_book_number === user.student?.record_book_number,
      )
    )
      throw new Error('Номер зачётной книжки уже используется.')
  }
  if (
    user.teacher &&
    !data.catalogs.departments.some(
      (item) => item.id === user.teacher?.department_id && !item.archived,
    )
  )
    throw new Error('Выберите действующую кафедру.')
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
  }),
  actions: {
    async initialize(force = false) {
      if (this.loading || (this.data && !force)) return
      this.loading = true
      this.error = ''
      try {
        this.data = await adminRepository.load()
      } catch (error) {
        this.error =
          error instanceof Error ? error.message : 'Не удалось загрузить примерные данные.'
      } finally {
        this.loading = false
      }
    },
    async commit(update: (data: AdminData) => void) {
      if (!this.data || this.saving) return false
      this.saving = true
      this.error = ''
      this.success = ''
      try {
        const next = JSON.parse(JSON.stringify(this.data)) as AdminData
        update(next)
        await adminRepository.save(next)
        this.data = next
        this.success = 'Изменения сохранены в демо-данных.'
        return true
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось сохранить изменения.'
        return false
      } finally {
        this.saving = false
      }
    },
    async reviewRequest(id: number, user: AdminUser | null) {
      return this.commit((data) => {
        const request = data.requests.find((item) => item.id === id)
        if (!request || request.status !== 'pending')
          throw new Error('Заявка уже обработана или не найдена.')
        if (user) {
          const created = {
            ...user,
            id: Math.max(0, ...data.users.map((item) => item.id)) + 1,
            email: request.email.toLowerCase(),
            is_admin: false,
            is_active: true,
          }
          if (Boolean(created.student) === Boolean(created.teacher))
            throw new Error('Выберите одну учебную роль для регистрации.')
          validateUser(data, created)
          data.users.push(created)
          request.user_id = created.id
          request.status = 'approved'
        } else {
          request.status = 'rejected'
        }
        request.reviewed_at = new Date().toISOString()
        request.reviewed_by = 'Анна Иванова'
      })
    },
    async saveUser(user: AdminUser) {
      return this.commit((data) => {
        const existing = data.users.find((item) => item.id === user.id)
        const next = {
          ...user,
          email: user.email.trim().toLowerCase(),
          is_admin: existing?.is_admin ?? false,
          is_active: existing?.is_active ?? true,
          id: user.id || Math.max(0, ...data.users.map((item) => item.id)) + 1,
        }
        if (
          existing?.teacher &&
          !next.teacher &&
          data.workloads.some((item) => item.teacher_id === existing.id)
        )
          throw new Error('Нельзя снять роль преподавателя: у пользователя есть учебная нагрузка.')
        validateUser(data, next)
        if (existing) data.users.splice(data.users.indexOf(existing), 1, next)
        else data.users.push(next)
      })
    },
    async toggleUser(id: number) {
      return this.commit((data) => {
        const user = data.users.find((item) => item.id === id)
        if (!user) throw new Error('Пользователь не найден.')
        if (
          user.is_active &&
          user.is_admin &&
          data.users.filter((item) => item.is_admin && item.is_active).length <= 1
        )
          throw new Error('Нельзя заблокировать последнего действующего администратора.')
        user.is_active = !user.is_active
      })
    },
    async saveCatalog(kind: CatalogKind, entry: CatalogEntry) {
      return this.commit((data) => {
        const entries = data.catalogs[kind]
        const existing = entries.find((item) => item.id === entry.id)
        const next = {
          ...entry,
          name: entry.name.trim(),
          id: entry.id || Math.max(0, ...entries.map((item) => item.id)) + 1,
        }
        if (!next.name || next.name.length > 200)
          throw new Error('Укажите название до 200 символов.')
        if (
          entries.some(
            (item) => item.id !== next.id && item.name.toLowerCase() === next.name.toLowerCase(),
          )
        )
          throw new Error('Запись с таким названием уже существует.')
        if (
          kind === 'specializations' &&
          (!next.code?.trim() ||
            entries.some((item) => item.id !== next.id && item.code === next.code))
        )
          throw new Error('Укажите уникальный код направления.')
        if (
          kind === 'groups' &&
          (!next.admission_year ||
            next.admission_year < 2000 ||
            !Number.isInteger(next.admission_year) ||
            !data.catalogs.specializations.some(
              (item) => item.id === next.specialization_id && !item.archived,
            ))
        )
          throw new Error('Укажите направление и год поступления не ранее 2000.')
        if (
          kind === 'semesters' &&
          (!Number.isInteger(next.academic_year_start) ||
            !next.academic_year_start ||
            next.academic_year_start < 2000 ||
            ![1, 2].includes(next.number ?? 0) ||
            entries.some(
              (item) =>
                item.id !== next.id &&
                item.academic_year_start === next.academic_year_start &&
                item.number === next.number,
            ))
        )
          throw new Error(
            'Проверьте учебный год и номер семестра: сочетание должно быть уникальным.',
          )
        if (existing) entries.splice(entries.indexOf(existing), 1, next)
        else entries.push(next)
      })
    },
    async archiveCatalog(kind: CatalogKind, id: number) {
      return this.commit((data) => {
        const entry = data.catalogs[kind].find((item) => item.id === id)
        if (!entry || entry.archived) throw new Error('Запись уже архивирована или не найдена.')
        const used =
          kind === 'groups'
            ? data.users.some((item) => item.student?.group_id === id) ||
              data.courses.some((item) => item.group_id === id)
            : kind === 'departments'
              ? data.users.some((item) => item.teacher?.department_id === id)
              : kind === 'specializations'
                ? data.catalogs.groups.some(
                    (item) => item.specialization_id === id && !item.archived,
                  )
                : kind === 'subjects'
                  ? data.courses.some((item) => item.subject_id === id)
                  : kind === 'semesters'
                    ? data.courses.some((item) => item.semester_id === id)
                    : data.lessons.some((item) => item.room_id === id)
        if (used)
          throw new Error(
            'Запись используется в учебных данных. Сначала перенесите связанные записи.',
          )
        entry.archived = true
      })
    },
    async saveCourse(course: CourseOffering) {
      return this.commit((data) => {
        for (const [kind, id] of [
          ['groups', course.group_id],
          ['semesters', course.semester_id],
          ['subjects', course.subject_id],
        ] as const) {
          if (!data.catalogs[kind].some((item) => item.id === id && !item.archived))
            throw new Error('Выберите действующие группу, семестр и дисциплину.')
        }
        if (
          data.courses.some(
            (item) =>
              item.id !== course.id &&
              item.group_id === course.group_id &&
              item.semester_id === course.semester_id &&
              item.subject_id === course.subject_id,
          )
        )
          throw new Error('Дисциплина уже назначена этой группе в выбранном семестре.')
        const existing = data.courses.find((item) => item.id === course.id)
        if (existing && data.workloads.some((item) => item.course_id === course.id))
          throw new Error('Сначала удалите связанную нагрузку перед изменением назначения.')
        const next = {
          ...course,
          id: course.id || Math.max(0, ...data.courses.map((item) => item.id)) + 1,
        }
        if (existing) data.courses.splice(data.courses.indexOf(existing), 1, next)
        else data.courses.push(next)
      })
    },
    async saveWorkload(workload: AdminWorkload) {
      return this.commit((data) => {
        if (
          !data.courses.some((item) => item.id === workload.course_id) ||
          !data.users.some(
            (item) => item.id === workload.teacher_id && item.teacher && item.is_active,
          )
        )
          throw new Error('Выберите курс и действующего преподавателя.')
        if (!Number.isFinite(workload.planned_hours) || workload.planned_hours < 0)
          throw new Error('Количество часов должно быть неотрицательным числом.')
        if (
          data.workloads.some(
            (item) =>
              item.id !== workload.id &&
              item.course_id === workload.course_id &&
              item.teacher_id === workload.teacher_id &&
              item.lesson_type === workload.lesson_type,
          )
        )
          throw new Error('Такая нагрузка уже существует.')
        const existing = data.workloads.find((item) => item.id === workload.id)
        if (
          existing &&
          data.lessons.some((item) => item.workload_id === existing.id) &&
          (existing.teacher_id !== workload.teacher_id ||
            existing.lesson_type !== workload.lesson_type)
        )
          throw new Error('Нагрузка используется в расписании. Сначала измените занятия.')
        const next = {
          ...workload,
          id: workload.id || Math.max(0, ...data.workloads.map((item) => item.id)) + 1,
        }
        if (existing) data.workloads.splice(data.workloads.indexOf(existing), 1, next)
        else data.workloads.push(next)
      })
    },
    async saveLesson(lesson: AdminLesson) {
      return this.commit((data) => {
        if (
          ![1, 2].includes(lesson.week_number) ||
          lesson.day_of_week < 1 ||
          lesson.day_of_week > 6 ||
          !/^\d{2}:\d{2}$/.test(lesson.starts_at) ||
          !/^\d{2}:\d{2}$/.test(lesson.ends_at) ||
          lesson.starts_at >= lesson.ends_at
        )
          throw new Error('Проверьте неделю, день и интервал занятия.')
        if (!data.catalogs.rooms.some((item) => item.id === lesson.room_id && !item.archived))
          throw new Error('Выберите действующую аудиторию.')
        const conflicts = lessonConflicts(data, lesson)
        if (conflicts.length) throw new Error(conflicts.join(' '))
        const existing = data.lessons.find((item) => item.id === lesson.id)
        const next = {
          ...lesson,
          id: lesson.id || Math.max(0, ...data.lessons.map((item) => item.id)) + 1,
        }
        if (existing) data.lessons.splice(data.lessons.indexOf(existing), 1, next)
        else data.lessons.push(next)
      })
    },
    async removeAcademic(kind: 'courses' | 'workloads' | 'lessons', id: number) {
      return this.commit((data) => {
        if (kind === 'courses' && data.workloads.some((item) => item.course_id === id))
          throw new Error('Сначала удалите связанную нагрузку.')
        if (kind === 'workloads' && data.lessons.some((item) => item.workload_id === id))
          throw new Error('Нагрузка используется в расписании.')
        const entries = data[kind]
        const index = entries.findIndex((item) => item.id === id)
        if (index < 0) throw new Error('Запись не найдена.')
        entries.splice(index, 1)
      })
    },
    async copyLessons(
      group: number,
      sourceSemester: number,
      targetSemester: number,
      week: number,
      preview = true,
    ) {
      const data = this.data
      if (!data) return null
      const source = data.lessons.filter((item) => {
        const workload = data.workloads.find((value) => value.id === item.workload_id)
        const course = data.courses.find((value) => value.id === workload?.course_id)
        return (
          course?.group_id === group &&
          course.semester_id === sourceSemester &&
          item.week_number === week
        )
      })
      const candidates: AdminLesson[] = []
      const conflicts: string[] = []
      for (const [index, item] of source.entries()) {
        const original = data.workloads.find((value) => value.id === item.workload_id)!
        const course = data.courses.find((value) => value.id === original.course_id)!
        const target = data.courses.find(
          (value) =>
            value.group_id === group &&
            value.semester_id === targetSemester &&
            value.subject_id === course.subject_id,
        )
        const workload = data.workloads.find(
          (value) =>
            value.course_id === target?.id &&
            value.teacher_id === original.teacher_id &&
            value.lesson_type === original.lesson_type,
        )
        if (!workload) {
          conflicts.push(`Нет соответствующей нагрузки для занятия №${item.id} в целевом семестре.`)
          continue
        }
        const next = {
          ...item,
          id: Math.max(0, ...data.lessons.map((value) => value.id)) + index + 1,
          workload_id: workload.id,
          week_number: sourceSemester === targetSemester ? 2 : week,
        }
        conflicts.push(...lessonConflicts(data, next, [...data.lessons, ...candidates]))
        candidates.push(next)
      }
      if (!preview && candidates.length && !conflicts.length) {
        const saved = await this.commit((next) => {
          next.lessons.push(...candidates)
        })
        if (!saved) return null
      }
      return { candidates, conflicts }
    },
    async saveNews(news: AdminNews) {
      return this.commit((data) => {
        if (!news.title.trim() || !news.body.trim())
          throw new Error('Заполните заголовок и текст новости.')
        if (news.title.length > 200)
          throw new Error('Заголовок должен быть не длиннее 200 символов.')
        const existing = data.news.find((item) => item.id === news.id)
        const now = new Date().toISOString()
        const next = {
          ...news,
          id: news.id || Math.max(0, ...data.news.map((item) => item.id)) + 1,
          author: existing?.author ?? 'Анна Иванова',
          created_at: existing?.created_at ?? now,
          updated_at: now,
          published_at: news.status === 'published' ? (existing?.published_at ?? now) : null,
        }
        if (existing) data.news.splice(data.news.indexOf(existing), 1, next)
        else data.news.push(next)
      })
    },
    async deleteNews(id: number) {
      return this.commit((data) => {
        const index = data.news.findIndex((item) => item.id === id)
        if (index < 0) throw new Error('Новость не найдена.')
        if (data.news[index]?.status === 'published')
          throw new Error('Перед удалением снимите новость с публикации.')
        data.news.splice(index, 1)
      })
    },
  },
})
