export type CatalogKind =
  'groups' | 'specializations' | 'departments' | 'subjects' | 'semesters' | 'rooms'
export type LessonKind = 'lecture' | 'practice' | 'laboratory' | 'consultation' | 'exam'
export interface CatalogEntry {
  id: number
  name: string
  archived: boolean
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
  reviewed_by: string | null
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
  planned_hours: number
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
export interface AdminNews {
  id: number
  title: string
  body: string
  cover: string
  author: string
  status: 'draft' | 'published' | 'archived'
  is_pinned: boolean
  created_at: string
  updated_at: string
  published_at: string | null
}
export interface AdminData {
  version: 1
  requests: RegistrationRequest[]
  users: AdminUser[]
  catalogs: Record<CatalogKind, CatalogEntry[]>
  courses: CourseOffering[]
  workloads: AdminWorkload[]
  lessons: AdminLesson[]
  news: AdminNews[]
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
  draft: 'Черновик',
  published: 'Опубликована',
  archived: 'В архиве',
}
export const statusVariants: Record<
  RegistrationRequest['status'] | AdminNews['status'],
  StatusBadgeVariant
> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'error',
  draft: 'warning',
  published: 'success',
  archived: 'neutral',
}
export const adminDate = new Intl.DateTimeFormat('ru-RU', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export const adminDays = ['Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота']
