import { computed, inject, readonly, ref, type InjectionKey } from 'vue'
import { lessonTypes, periods } from './schedule.data'

export function createSchedule() {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  today.setDate(today.getDate() - ((today.getDay() + 6) % 7))
  const weekStart = ref(today)

  const days = computed(() => {
    const today = new Date().toDateString()
    return Array.from({ length: 6 }, (_, index) => {
      const date = new Date(weekStart.value)
      date.setDate(date.getDate() + index)
      return {
        date: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`,
        fullLabel: date
          .toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })
          .replace(/^./, (letter) => letter.toUpperCase()),
        lessons: periods.flatMap((period) => {
          const lesson = period.lessons[index]
          return lesson
            ? [
                {
                  ...period,
                  lesson,
                  typeLabel: lessonTypes.find((item) => item.type === lesson.type)?.label,
                },
              ]
            : []
        }),
        label: date
          .toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' })
          .replace(/\.$/, '')
          .replace(/^./, (letter) => letter.toUpperCase()),
        isToday: date.toDateString() === today,
      }
    })
  })

  const weekLabel = computed(() => {
    const start = weekStart.value
    const end = new Date(start)
    end.setDate(end.getDate() + 5)
    const endLabel = end.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })
    if (start.getMonth() === end.getMonth() && start.getFullYear() === end.getFullYear()) {
      return `${start.getDate()} – ${endLabel}`
    }
    return `${start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' })} – ${endLabel}`
  })

  function shiftWeek(direction: number) {
    const date = new Date(weekStart.value)
    date.setDate(date.getDate() + direction * 7)
    weekStart.value = date
  }

  function selectWeek(date: Date) {
    const start = new Date(date)
    start.setHours(0, 0, 0, 0)
    start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
    weekStart.value = start
  }

  return {
    days,
    periods: readonly(periods),
    weekStart: readonly(weekStart),
    weekLabel,
    shiftWeek,
    selectWeek,
  }
}

export const scheduleKey: InjectionKey<ReturnType<typeof createSchedule>> = Symbol('schedule')

export function useSchedule() {
  const schedule = inject(scheduleKey)
  if (!schedule) throw new Error('useSchedule requires a SchedulePage provider')
  return schedule
}
