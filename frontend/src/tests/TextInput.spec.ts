import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import inputStyle from '@/components/UI/inputs/Input/Input.module.scss'

describe('TextInput errors', () => {
  it.each([false, true])(
    'associates a dynamic error with the field (multiline: %s)',
    async (multiline) => {
      const wrapper = mount(TextInput, { props: { multiline, modelValue: 'value' } })
      const field = () => wrapper.get(multiline ? 'textarea' : 'input')

      expect(wrapper.find('[role="alert"]').exists()).toBe(false)
      expect(field().attributes('aria-invalid')).toBeUndefined()

      await wrapper.setProps({ error: 'Проверьте значение' })
      const message = wrapper.get('[role="alert"]')
      expect(message.text()).toBe('Проверьте значение')
      expect(field().attributes('aria-invalid')).toBe('true')
      expect(field().attributes('aria-describedby')).toBe(message.attributes('id'))
      expect(message.attributes('id')).toBe(`${field().attributes('id')}-error`)
      expect(field().classes()).toContain(inputStyle.error)

      await field().setValue('new value')
      expect(wrapper.emitted('update:modelValue')?.slice(-1)[0]).toEqual(['new value'])

      await wrapper.setProps({ error: '' })
      expect(wrapper.find('[role="alert"]').exists()).toBe(false)
      expect(field().attributes('aria-invalid')).toBeUndefined()
      expect(field().attributes('aria-describedby')).toBeUndefined()
      expect(field().classes()).not.toContain(inputStyle.error)
    },
  )

  it('preserves an explicit id and label and keeps disabled fields disabled', () => {
    const wrapper = mount(TextInput, {
      props: { id: 'email', label: 'Почта', error: 'Некорректная почта', variant: 'disabled' },
    })
    expect(wrapper.get('label').attributes('for')).toBe('email')
    expect(wrapper.get('input').attributes('aria-describedby')).toBe('email-error')
    expect(wrapper.get('input').attributes('disabled')).toBeDefined()
  })

  it('generates distinct error ids for separate fields', () => {
    const first = mount(TextInput, { props: { error: 'Ошибка' } })
    const second = mount(TextInput, { props: { error: 'Ошибка' } })
    expect(first.get('[role="alert"]').attributes('id')).not.toBe(
      second.get('[role="alert"]').attributes('id'),
    )
  })
})
