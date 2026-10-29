import type { ScheduleLesson } from '@/pages/schedule/schedule.types'

export interface ScheduleDayItem extends ScheduleLesson {
  typeLabel?: string
}
