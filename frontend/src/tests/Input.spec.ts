import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import Input from '../components/inputs/Input/Input.vue'
import style from '../components/inputs/Input/Input.module.scss'

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
})
