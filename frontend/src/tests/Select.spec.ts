import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import Select from '@/components/UI/inputs/Select/Select.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'

const options = [
  { value: 'math', label: 'Математика' },
  { value: 'archive', label: 'Архив', disabled: true },
  { value: 'physics', label: 'Физика' },
]
const wrappers: ReturnType<typeof mount>[] = []
afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
})

function render(props = {}) {
  const wrapper = mount(Select, { props: { options, ...props }, attachTo: document.body })
  wrappers.push(wrapper)
  return wrapper
}

describe('Select', () => {
  it('shows a placeholder, selects a value and restores focus to the trigger', async () => {
    const wrapper = render({ name: 'subject' })
    const trigger = wrapper.get('button')
    expect(wrapper.findAllComponents(StyleTab)).toHaveLength(1)
    expect(wrapper.getComponent(StyleTab).get('button').element).toBe(trigger.element)
    expect(trigger.attributes('role')).toBe('button')
    expect(trigger.attributes('aria-selected')).toBeUndefined()
    expect(trigger.text()).toBe('Выберите значение')
    await trigger.trigger('click')
    expect(wrapper.findAll('[role="option"]')).toHaveLength(3)
    const tabs = wrapper.findAllComponents(StyleTab)
    expect(tabs).toHaveLength(options.length + 1)
    expect(tabs[3]!.get('button').attributes('role')).toBe('option')
    await wrapper.findAll('[role="option"]')[2]!.trigger('click')
    expect(wrapper.emitted('update:modelValue')).toEqual([['physics']])
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    expect(document.activeElement).toBe(trigger.element)
    await wrapper.setProps({ modelValue: 'physics' })
    expect(trigger.text()).toBe('Физика')
    expect(wrapper.get('input').element.value).toBe('physics')
  })

  it('skips disabled options with arrows and only commits on activation', async () => {
    const wrapper = render({ modelValue: 'math' })
    await wrapper.get('button').trigger('keydown', { key: 'ArrowDown' })
    const entries = wrapper.findAll('[role="option"]')
    await entries[0]!.trigger('keydown', { key: 'ArrowDown' })
    expect(document.activeElement).toBe(entries[2]!.element)
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
    await entries[2]!.trigger('keydown', { key: 'Escape' })
    expect(wrapper.get('button').attributes('aria-expanded')).toBe('false')
    expect(wrapper.get('button').text()).toBe('Математика')
  })

  it('closes on outside pointer and focus departure and handles empty options', async () => {
    const wrapper = render({ options: [] })
    await wrapper.get('button').trigger('click')
    expect(wrapper.get('[role="listbox"]').text()).toContain('Нет доступных вариантов')
    document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
    await wrapper.get('button').trigger('click')
    await wrapper.get('[role="listbox"]').trigger('focusout', { relatedTarget: document.body })
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('prevents opening while disabled', async () => {
    const wrapper = render({ disabled: true })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    await wrapper.get('button').trigger('keydown', { key: 'ArrowDown' })
    expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
  })

  it('preserves custom option content inside StyleTab and never selects a disabled option', async () => {
    const wrapper = mount(Select, {
      props: { options, modelValue: 'math' },
      slots: {
        option: '<template #option="{ option }"><strong>{{ option.label }}</strong></template>',
      },
    })
    wrappers.push(wrapper)
    await wrapper.get('button').trigger('click')
    const tabs = wrapper
      .findAllComponents(StyleTab)
      .filter((tab) => tab.get('button').attributes('role') === 'option')
    expect(tabs[0]!.props('active')).toBe(true)
    expect(tabs[0]!.get('[role="option"] strong').text()).toBe('Математика')
    expect(tabs[1]!.get('button').attributes('disabled')).toBeDefined()
    await tabs[1]!.get('button').trigger('click')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
