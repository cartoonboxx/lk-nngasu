import { defineStore } from 'pinia'
import { scheduleDays } from '@/pages/schedule/schedule.data'

export const useScheduleStore = defineStore('schedule', {
  state: () => {
    const weekStart = new Date()
    weekStart.setHours(0, 0, 0, 0)
    weekStart.setDate(weekStart.getDate() - ((weekStart.getDay() + 6) % 7))
    return { schedule: scheduleDays, weekStart }
  },
  getters: {
    days: (state) => {
      const end = new Date(state.weekStart)
      end.setDate(end.getDate() + 6)
      return state.schedule
        .filter((day) => {
          const date = new Date(`${day.date}T00:00:00`)
          return date >= state.weekStart && date < end
        })
        .sort((a, b) => a.date.localeCompare(b.date))
    },
  },
  actions: {
    shiftWeek(direction: number) {
      const date = new Date(this.weekStart)
      date.setDate(date.getDate() + direction * 7)
      this.weekStart = date
    },
    selectWeek(date: Date) {
      const start = new Date(date)
      start.setHours(0, 0, 0, 0)
      start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
      this.weekStart = start
    },
  },
})
