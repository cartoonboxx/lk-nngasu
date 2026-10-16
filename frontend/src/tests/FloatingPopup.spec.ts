import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import FloatingPopup from '@/components/UI/popup/FloatingPopup/FloatingPopup.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'

const items = [
  { value: 'schedule', label: 'Расписание' },
  { value: 'archive', label: 'Архив', disabled: true },
  { value: 'grades', label: 'Оценки' },
]

describe('FloatingPopup', () => {
  it('renders a vertical tab list and switches content without closing', async () => {
    const wrapper = mount(FloatingPopup, {
      props: { items },
      slots: { schedule: 'Занятия', grades: 'Результаты' },
    })
    await wrapper.get('button').trigger('click')
    expect(wrapper.findAllComponents(NavTab)).toHaveLength(items.length)
    expect(wrapper.findAllComponents(StyleTab)).toHaveLength(0)
    expect(wrapper.get('[role="tablist"]').attributes('aria-orientation')).toBe('vertical')
    const tabs = wrapper.findAll('[role="tab"]')
    await tabs[2]!.trigger('click')
    expect(tabs[2]!.attributes('aria-selected')).toBe('true')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
    expect(wrapper.findAll('[role="tabpanel"]')[1]!.isVisible()).toBe(true)
    expect(wrapper.findAll('[role="tabpanel"]')[0]!.isVisible()).toBe(false)
    expect(wrapper.emitted('update:modelValue')).toEqual([['grades']])
    wrapper.unmount()
  })

  it('supports controlled selection and keyboard switching while skipping disabled tabs', async () => {
    const wrapper = mount(FloatingPopup, {
      props: { items, modelValue: 'schedule' },
      attachTo: document.body,
    })
    try {
      await wrapper.get('button').trigger('click')
      const tabs = wrapper.findAll('[role="tab"]')
      await tabs[0]!.trigger('keydown', { key: 'ArrowDown' })
      expect(document.activeElement).toBe(tabs[2]!.element)
      expect(wrapper.emitted('update:modelValue')).toEqual([['grades']])
      expect(tabs[0]!.attributes('aria-selected')).toBe('true')
      await wrapper.setProps({ modelValue: 'grades' })
      expect(tabs[2]!.attributes('aria-selected')).toBe('true')
      await wrapper.setProps({ items: items.slice(0, 2) })
      expect(tabs[0]!.attributes('aria-selected')).toBe('true')
    } finally {
      wrapper.unmount()
    }
  })
})
