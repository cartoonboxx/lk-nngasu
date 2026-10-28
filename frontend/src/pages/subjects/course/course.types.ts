import type { Component } from 'vue'

export interface CourseTopic {
  id: string
  title: string
  description: string
  task: string
  deadline: string
  steps: string[]
}

export interface CourseMaterial {
  kind: 'lecture' | 'presentation' | 'practice'
  title: string
  meta: string
  icon: Component
}

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
