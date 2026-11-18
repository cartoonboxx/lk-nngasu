export type { Topic as CourseTopic } from '@/api/cabinet.types'

export interface AssignmentAttachment {
  id: string
  name: string
  size: number
  file?: File
}

export interface AssignmentAnswerDraft {
  attachments: AssignmentAttachment[]
  comment: string
}
