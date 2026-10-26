import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { h } from 'vue'
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
import type { ScheduleDayItem } from '@/components/ScheduleDay/schedule-day.types'

const lesson: ScheduleDayItem = {
  id: 'math',
  start: '08:30',
  end: '10:00',
  subject: 'Математический анализ',
  type: 'lecture',
  typeLabel: 'Лекция',
  location: 'Корпус А, ауд. 302',
  teacher: 'Иванова Елена Сергеевна',
}

describe('ScheduleDay', () => {
  it('renders independently and accepts caller-owned navigation and lesson statuses', () => {
    const wrapper = mount(ScheduleDay, {
      props: { title: 'Расписание на сегодня', date: '2026-10-26', lessons: [lesson] },
      slots: {
        'header-actions': () => h('button', 'Следующий день'),
        'lesson-trailing': ({ lesson: item }: { lesson: ScheduleDayItem }) =>
          h('span', `${item.id}: Идёт`),
      },
    })
    expect(wrapper.get('h2').text()).toBe('Расписание на сегодня')
    expect(wrapper.get('time').attributes('datetime')).toBe('2026-10-26')
    expect(wrapper.get('button').text()).toBe('Следующий день')
    const row = wrapper.get('li')
    expect(row.get('h3').text()).toBe(lesson.subject)
    for (const text of [
      lesson.start,
      lesson.end,
      lesson.typeLabel!,
      lesson.location,
      lesson.teacher,
      'math: Идёт',
    ]) {
      expect(row.text()).toContain(text)
    }
  })

  it('updates its content and supports an empty state without a date', async () => {
    const wrapper = mount(ScheduleDay, {
      props: { title: 'Занятия', lessons: [lesson] },
      slots: { empty: 'Сегодня можно отдохнуть' },
    })
    expect(wrapper.find('time').exists()).toBe(false)
    await wrapper.setProps({ lessons: [] })
    expect(wrapper.find('ul').exists()).toBe(false)
    expect(wrapper.text()).toContain('Сегодня можно отдохнуть')
  })
})
