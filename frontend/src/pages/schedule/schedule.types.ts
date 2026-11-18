export type LessonType = 'lecture' | 'practice' | 'lab' | 'consultation' | 'exam'

export interface ScheduleLesson {
  id: string | number
  start: string
  end: string
  subject: string
  type: LessonType
  location: string
  teacher: string
}

export interface ScheduleDay {
  date: string
  lessons: ScheduleLesson[]
}

export type SchedulePeriodPickerLocation = 'header' | 'controls'

export const lessonTypes: { type: LessonType; label: string }[] = [
  { type: 'lecture', label: 'Лекция' },
  { type: 'practice', label: 'Практика' },
  { type: 'lab', label: 'Лабораторная' },
  { type: 'consultation', label: 'Консультация' },
  { type: 'exam', label: 'Экзамен' },
]
