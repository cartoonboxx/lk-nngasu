export type CatalogKind =
  'groups' | 'specializations' | 'departments' | 'subjects' | 'semesters' | 'rooms'
export type LessonKind = 'lecture' | 'practice' | 'laboratory' | 'consultation' | 'exam'
export interface CatalogEntry {
  id: number
  name: string
  is_current?: boolean
  code?: string
  specialization_id?: number
  admission_year?: number
  academic_year_start?: number
  number?: number
}
export interface AdminUser {
  id: number
  last_name: string
  first_name: string
  middle_names: string
  email: string
  passport: string
  snils: string
  avatar_url: string
  is_active: boolean
  is_admin: boolean
  student: { group_id: number; record_book_number: string; is_headman: boolean } | null
  teacher: { department_id: number } | null
}
export interface RegistrationRequest {
  id: number
  email: string
  passport: string
  snils: string
  last_name: string
  first_name: string
  middle_names: string
  comment: string
  status: 'pending' | 'approved' | 'rejected'
  created_at: string
  reviewed_at: string | null
  reviewed_by: { id: number; name: string } | null
  user_id: number | null
}
export interface CourseOffering {
  id: number
  group_id: number
  semester_id: number
  subject_id: number
}
export interface AdminWorkload {
  id: number
  course_id: number
  teacher_id: number
  lesson_type: LessonKind
  planned_hours: number | null
}
export interface AdminLesson {
  id: number
  workload_id: number
  week_number: number
  day_of_week: number
  starts_at: string
  ends_at: string
  room_id: number
}
export interface AdminData {
  requests: RegistrationRequest[]
  users: AdminUser[]
  catalogs: Record<CatalogKind, CatalogEntry[]>
  courses: CourseOffering[]
  workloads: AdminWorkload[]
  lessons: AdminLesson[]
}

export interface RequestQuery {
  q: string
  status: string
  sort: string
  limit: number
  offset: number
}

export const catalogLabels: Record<CatalogKind, string> = {
  groups: 'Учебные группы',
  specializations: 'Направления',
  departments: 'Кафедры',
  subjects: 'Дисциплины',
  semesters: 'Семестры',
  rooms: 'Аудитории',
}
export const lessonLabels: Record<LessonKind, string> = {
  lecture: 'Лекция',
  practice: 'Практика',
  laboratory: 'Лабораторная',
  consultation: 'Консультация',
  exam: 'Экзамен',
}
export const statusLabels: Record<string, string> = {
  pending: 'Ожидает решения',
  approved: 'Одобрена',
  rejected: 'Отклонена',
}
export const statusVariants: Record<RegistrationRequest['status'], StatusBadgeVariant> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
}
export const adminDate = new Intl.DateTimeFormat('ru-RU', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export const adminDays = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота']
