export type LessonType = 'lecture' | 'practice' | 'lab'

export interface ScheduleLesson {
  subject: string
  type: LessonType
  location: string
  teacher: string
}

export type SchedulePeriodPickerLocation = 'header' | 'controls'
