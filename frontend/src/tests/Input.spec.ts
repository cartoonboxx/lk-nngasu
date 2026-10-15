import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import Input from '@/components/UI/inputs/Input/Input.vue'
import style from '@/components/UI/inputs/Input/Input.module.scss'

describe('Input values', () => {
  it('detects autofill without an input event and returns to default when cleared', async () => {
    const wrapper = mount(Input, { props: { type: 'password' } })
    const field = wrapper.get('input')
    field.element.value = 'saved-password'
    await field.trigger('animationstart')
    expect(field.classes()).toContain(style.filled)
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['saved-password'])
    await field.setValue('')
    expect(field.classes()).toContain(style.default)
    expect(field.classes()).not.toContain(style.filled)
  })

  it('keeps the error state when autofilled', async () => {
    const wrapper = mount(Input, { props: { variant: 'error' } })
    const field = wrapper.get('input')
    field.element.value = 'saved-login'
    await field.trigger('animationstart')
    expect(field.classes()).toContain(style.error)
    expect(field.classes()).not.toContain(style.filled)
  })

  it('updates filled state when the model changes', async () => {
    const wrapper = mount(Input)
    await wrapper.setProps({ modelValue: 'login' })
    expect(wrapper.get('input').classes()).toContain(style.filled)
    await wrapper.setProps({ modelValue: '' })
    expect(wrapper.get('input').classes()).toContain(style.default)
  })

  it('renders prefix slot before the input', () => {
    const wrapper = mount(Input, {
      slots: {
        prefix: '<span data-test="prefix">Иконка</span>',
      },
    })
    const prefix = wrapper.get('[data-test="prefix"]')
    const field = wrapper.get('input')

    expect(prefix.element.compareDocumentPosition(field.element)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    )
    expect(field.classes()).toContain(style.withPrefix)
  })
})
