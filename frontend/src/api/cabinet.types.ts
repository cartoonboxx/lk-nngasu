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
  file: FileInfo | null
  external_url: string | null
}

export interface Topic {
  id: number
  title: string
  description: string | null
  materials: Material[]
  assignment: Assignment | null
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
  excerpt: string
  author: { id: number; name: string }
  published_at: string
}

export interface NewsDetail extends News {
  body: string
}

export interface AdminNews extends Omit<NewsDetail, 'published_at'> {
  published_at: string | null
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
}

export interface ServerMessage {
  id: number
  chat_id: number
  sender_id: number
  sender_name: string
  body: string
  created_at: string
  edited_at: string | null
  files: { id: number; name: string }[]
}

export interface ServerChat {
  unread_count: number
  last_read_message_id: number
  id: number
  kind: 'direct' | 'group'
  title: string
  peer: { id: number; name: string } | null
  last_message:
    (Omit<ServerMessage, 'chat_id' | 'files' | 'edited_at'> & { has_files: boolean }) | null
}

export interface MessagePage {
  items: ServerMessage[]
  has_more: boolean
  next_before: number | null
}

export interface Contact {
  id: number
  name: string
  description: string | null
  avatar: { id: number; url: string } | null
}

export interface GradeRow {
  assignment_id: number
  assignment_title: string
  topic_title: string
  subject_name: string
  semester_id: number
  semester_name: string
  status: 'not_submitted' | 'missing' | 'submitted' | 'reviewed'
  submitted_at: string | null
  grade: number | null
  review_comment: string | null
}

export interface Dashboard {
  week: { date: string; week_number: number; position: string } | null
  lessons: CalendarLesson[]
  chats: ServerChat[]
  news: News[]
}
