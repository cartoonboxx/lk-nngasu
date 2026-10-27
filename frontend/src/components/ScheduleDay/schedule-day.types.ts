export interface ScheduleDayItem {
  id: string | number
  start: string
  end: string
  subject: string
  type: 'lecture' | 'practice' | 'lab'
  typeLabel?: string
  location: string
  teacher: string
}
