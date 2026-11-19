import { useAuthStore } from '@/stores/auth'
import type { ApiPage } from '@/api/client'
import type {
  AdminData,
  AdminUser,
  AdminLesson,
  AdminWorkload,
  CatalogKind,
  CatalogEntry,
  CourseOffering,
  RegistrationRequest,
} from './admin.types'

export const catalogPaths: Record<CatalogKind, string> = {
  groups: '/groups',
  specializations: '/specializations',
  departments: '/departments',
  subjects: '/catalog/subjects',
  semesters: '/semesters',
  rooms: '/rooms',
}

export interface AdminUserResponse extends Omit<
  AdminUser,
  'middle_names' | 'passport' | 'snils' | 'avatar_url' | 'student'
> {
  middle_names: string | null
  passport?: string | null
  snils?: string | null
  avatar: { id: number; url: string } | null
  student: { group_id: number; record_book_number: string | null; is_headman: boolean } | null
}

export function normalizeUser(user: AdminUserResponse): AdminUser {
  return {
    ...user,
    middle_names: user.middle_names ?? '',
    passport: user.passport ?? '',
    snils: user.snils ?? '',
    avatar_url: user.avatar?.url ?? '',
    student: user.student
      ? {
          group_id: user.student.group_id,
          record_book_number: user.student.record_book_number ?? '',
          is_headman: user.student.is_headman,
        }
      : null,
    teacher: user.teacher ? { department_id: user.teacher.department_id } : null,
  }
}

export function normalizeRequest(item: RegistrationRequest): RegistrationRequest {
  return {
    ...item,
    first_name: item.first_name ?? '',
    last_name: item.last_name ?? '',
    middle_names: item.middle_names ?? '',
    comment: item.comment ?? '',
  }
}

export const adminRepository = {
  async list<T>(path: string): Promise<T[]> {
    const items: T[] = []
    let offset = 0
    while (true) {
      const separator = path.includes('?') ? '&' : '?'
      const page = await useAuthStore().request<ApiPage<T>>(
        `${path}${separator}limit=200&offset=${offset}`,
        { paginated: true },
      )
      items.push(...page.data)
      offset += page.data.length
      if (offset >= page.meta.total || !page.data.length) return items
    }
  },
  async load(section: string): Promise<AdminData> {
    const data: AdminData = {
      requests: [],
      users: [],
      courses: [],
      workloads: [],
      lessons: [],
      catalogs: {
        groups: [],
        specializations: [],
        departments: [],
        subjects: [],
        semesters: [],
        rooms: [],
      },
    }
    if (section === 'admin-news' || section === 'admin-administrators') return data
    const kinds: CatalogKind[] =
      section === 'admin-requests' || section === 'admin-users'
        ? ['groups', 'departments']
        : (Object.keys(catalogPaths) as CatalogKind[])
    await Promise.all(
      kinds.map(async (kind) => {
        if (kind === 'rooms') {
          const rooms = await this.list<{ id: number; number: string }>(catalogPaths[kind])
          data.catalogs.rooms = rooms.map((room) => ({ id: room.id, name: room.number }))
        } else data.catalogs[kind] = await this.list<CatalogEntry>(catalogPaths[kind])
      }),
    )
    if (['admin-users', 'admin-courses', 'admin-schedule'].includes(section)) {
      data.users = (await this.list<AdminUserResponse>('/users')).map(normalizeUser)
    }
    if (section === 'admin-courses' || section === 'admin-schedule') {
      data.courses = await this.list<CourseOffering>('/courses')
      const workloads = await Promise.all(
        data.courses.map((course) =>
          useAuthStore().request<AdminWorkload[]>(`/courses/${course.id}/workloads`),
        ),
      )
      data.workloads = workloads.flat()
      const lessons = await Promise.all(
        data.catalogs.semesters.map((semester) =>
          useAuthStore().request<AdminLesson[]>(`/schedule/lessons?semester_id=${semester.id}`),
        ),
      )
      data.lessons = lessons.flat()
    }
    return data
  },
}
