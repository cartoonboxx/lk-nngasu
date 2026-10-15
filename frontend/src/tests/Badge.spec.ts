import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import Badge from '@/components/UI/Badge/Badge.vue'

describe('Badge', () => {
  it('renders a numeric value next to slotted content', () => {
    const wrapper = mount(Badge, {
      props: { value: 7 },
      slots: { default: '<button type="button">Уведомления</button>' },
    })

    expect(wrapper.get('button').text()).toBe('Уведомления')
    expect(wrapper.get('[role="status"]').text()).toBe('7')
    expect(wrapper.get('[role="status"]').attributes('aria-label')).toBe('Количество: 7')
  })
})
