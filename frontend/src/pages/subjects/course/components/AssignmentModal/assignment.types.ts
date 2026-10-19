export interface AssignmentAttachment {
  id: string
  name: string
  size: number
  file?: File
  url?: string
}

export interface AssignmentAnswer {
  attachments: AssignmentAttachment[]
  comment: string
  submittedAt: string
  review?: {
    grade: string | number
    comment: string
  }
}

export type AssignmentAnswerDraft = Pick<AssignmentAnswer, 'attachments' | 'comment'>
