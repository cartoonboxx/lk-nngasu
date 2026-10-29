export type LessonType = 'lecture' | 'practice' | 'lab'

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
