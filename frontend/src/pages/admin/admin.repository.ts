import { createAdminData } from './admin.data'
import { catalogLabels, type AdminData, type CatalogKind } from './admin.types'

// A local snapshot is not an HTTP contract. Production needs entity commands and server pagination.
export interface AdminRepository {
  load(): Promise<AdminData>
  save(data: AdminData): Promise<void>
}
export const adminStorageKey = 'unihub-admin-demo-v1'
export const adminRepository: AdminRepository = {
  async load() {
    let saved: string | null
    try {
      saved = localStorage.getItem(adminStorageKey)
    } catch {
      throw new Error(
        'Хранилище браузера недоступно. Разрешите сохранение данных и повторите загрузку.',
      )
    }
    if (!saved) return createAdminData()
    try {
      const data = JSON.parse(saved) as AdminData
      if (
        !data ||
        data.version !== 1 ||
        !data.catalogs ||
        !['requests', 'users', 'courses', 'workloads', 'lessons', 'news'].every((key) =>
          Array.isArray(data[key as keyof AdminData]),
        )
      )
        throw new Error()
      for (const kind of Object.keys(catalogLabels) as CatalogKind[]) {
        if (
          !Array.isArray(data.catalogs[kind]) ||
          data.catalogs[kind].some(
            (item) =>
              !item ||
              !Number.isInteger(item.id) ||
              typeof item.name !== 'string' ||
              typeof item.archived !== 'boolean',
          )
        )
          throw new Error()
      }
      if (
        data.users.some(
          (item) =>
            !item ||
            !Number.isInteger(item.id) ||
            !['last_name', 'first_name', 'middle_names', 'email', 'avatar_url'].every(
              (key) => typeof item[key as keyof typeof item] === 'string',
            ) ||
            typeof item.is_active !== 'boolean' ||
            typeof item.is_admin !== 'boolean',
        )
      )
        throw new Error()
      if (
        data.users.some(
          (item) =>
            (item.student &&
              (!data.catalogs.groups.some((group) => group.id === item.student?.group_id) ||
                typeof item.student.record_book_number !== 'string')) ||
            (item.teacher &&
              !data.catalogs.departments.some((dept) => dept.id === item.teacher?.department_id)),
        )
      )
        throw new Error()
      if (
        data.courses.some(
          (item) =>
            !item ||
            !data.catalogs.groups.some((group) => group.id === item.group_id) ||
            !data.catalogs.semesters.some((semester) => semester.id === item.semester_id) ||
            !data.catalogs.subjects.some((subject) => subject.id === item.subject_id),
        )
      )
        throw new Error()
      if (
        data.workloads.some(
          (item) =>
            !item ||
            !data.courses.some((course) => course.id === item.course_id) ||
            !data.users.some((user) => user.id === item.teacher_id && user.teacher),
        )
      )
        throw new Error()
      if (
        data.lessons.some(
          (item) =>
            !item ||
            !data.workloads.some((workload) => workload.id === item.workload_id) ||
            !data.catalogs.rooms.some((room) => room.id === item.room_id) ||
            !/^([01]\d|2[0-3]):[0-5]\d$/.test(item.starts_at) ||
            !/^([01]\d|2[0-3]):[0-5]\d$/.test(item.ends_at),
        )
      )
        throw new Error()
      if (
        data.requests.some(
          (item) =>
            !item ||
            !['pending', 'approved', 'rejected'].includes(item.status) ||
            !Number.isFinite(Date.parse(item.created_at)) ||
            typeof item.email !== 'string' ||
            typeof item.last_name !== 'string' ||
            typeof item.first_name !== 'string',
        )
      )
        throw new Error()
      if (
        data.news.some(
          (item) =>
            !item ||
            !['draft', 'published', 'archived'].includes(item.status) ||
            typeof item.title !== 'string' ||
            typeof item.body !== 'string' ||
            !Number.isFinite(Date.parse(item.updated_at)),
        )
      )
        throw new Error()
      return {
        version: data.version,
        requests: data.requests,
        users: data.users,
        catalogs: data.catalogs,
        courses: data.courses,
        workloads: data.workloads,
        lessons: data.lessons,
        news: data.news,
      }
    } catch {
      throw new Error(
        'Локальные данные повреждены или имеют другую версию. Проверьте хранилище браузера и повторите загрузку.',
      )
    }
  },
  async save(data) {
    try {
      localStorage.setItem(adminStorageKey, JSON.stringify(data))
    } catch {
      throw new Error(
        'Не удалось сохранить данные. Проверьте доступность и свободное место в хранилище браузера; попробуйте уменьшить изображения.',
      )
    }
  },
}
