export type GradeStatus = 'reviewed' | 'pending' | 'not-submitted' | 'missing'

export interface GradeAssignment {
  id: number
  title: string
  topic: string
  subject: string
  semester: string
  submittedAt: string | null
  grade: number | null
  status: GradeStatus
  teacherComment: string
  files: { id: number; name: string }[]
  answer: string
}

export const gradeStatuses: Record<GradeStatus, string> = {
  reviewed: 'Проверено',
  pending: 'На проверке',
  'not-submitted': 'Не отправлено',
  missing: 'Просрочено',
}
