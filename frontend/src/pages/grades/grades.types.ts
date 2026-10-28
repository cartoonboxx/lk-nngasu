export type GradeStatus = 'reviewed' | 'pending' | 'not-submitted'

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
  answer: string
}
