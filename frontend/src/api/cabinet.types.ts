export interface FileInfo {
  id: number
  name: string
  size_bytes: number
  mime_type: string
}

export interface Submission {
  id: number
  answer_text: string | null
  submitted_at: string | null
  status: string
  review_comment: string | null
  grade: number | null
  files: { id: number; name: string }[]
}

export interface Assignment {
  id: number
  course_id: number
  topic_id: number
  title: string
  description: string | null
  deadline_at: string | null
  files: { id: number; name: string }[]
  mine: { status: string; submitted_at: string | null; grade: number | null } | null
  submission?: Submission | null
}

export interface Material {
  id: number
  title: string
  file: FileInfo | null
  external_url: string | null
}

export interface Topic {
  id: number
  title: string
  description: string | null
  materials: Material[]
  assignments: Assignment[]
}

export interface Course {
  id: number
  subject_name: string
  group_name: string
  semester_id: number
  semester_name: string
  semester_number: number
}

export interface CourseDetail extends Course {
  topics: Topic[]
  workloads: { teacher_id: number; teacher_name: string }[]
}

export interface CalendarLesson {
  id: number
  date: string | null
  starts_at: string
  ends_at: string
  subject: string
  lesson_type: string
  room: string
  teacher_name: string
}

export interface News {
  id: number
  title: string
  text: string
  published_at: string
}

export interface ServerMessage {
  id: number
  chat_id: number
  sender_id: number
  sender_name: string
  body: string
  created_at: string
  deleted: boolean
  files: { id: number; name: string }[]
}

export interface ServerChat {
  id: number
  kind: 'direct' | 'group'
  title: string
  peer: { id: number; name: string } | null
  unread: number
  last_message: (Omit<ServerMessage, 'chat_id' | 'files'> & { has_files: boolean }) | null
}

export interface GradeReview {
  assignment_id: number
  grade: number
  review_comment: string | null
}
