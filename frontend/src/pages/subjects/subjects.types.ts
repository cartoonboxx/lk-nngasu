export interface Subject {
  id: number
  label: string
  value: string
  semesterLabel: string
  semester: string
}

export interface SubjectDetails extends Subject {
  topics: CourseTopic[]
}
import type { CourseTopic } from './course/course.types'
