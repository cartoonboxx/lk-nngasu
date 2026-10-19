import { mount, enableAutoUnmount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'

enableAutoUnmount(afterEach)

describe('MediaButton', () => {
  it('activates the media action with a native button that does not submit forms by default', async () => {
    const wrapper = mount(MediaButton, {
      props: { title: 'Лекция 1', meta: 'Конспект · MD', ariaLabel: 'Скачать лекцию' },
    })
    const button = wrapper.get('button')
    expect(button.attributes('type')).toBe('button')
    expect(button.attributes('aria-label')).toBe('Скачать лекцию')
    expect(button.text()).toContain('Лекция 1')
    expect(button.text()).toContain('Конспект · MD')
    await button.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it('keeps readonly content non-interactive while allowing an independent action', async () => {
    const wrapper = mount(MediaButton, {
      props: { readonly: true, title: 'main.pl', meta: '4 КБ' },
      slots: { actions: '<button type="button" aria-label="Удалить файл">Удалить</button>' },
    })
    expect(wrapper.findAll('button')).toHaveLength(1)
    expect(wrapper.find('button button').exists()).toBe(false)
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
    await wrapper.get('span').trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
  })

  it('does not activate the main action when a separate action is clicked', async () => {
    const wrapper = mount(MediaButton, {
      props: { title: 'Материал' },
      slots: { actions: '<button type="button">Удалить</button>' },
    })
    const buttons = wrapper.findAll('button')
    expect(buttons).toHaveLength(2)
    expect(wrapper.find('button button').exists()).toBe(false)
    await buttons[1]!.trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
    await buttons[0]!.trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it('blocks disabled media actions and supports restoring interactivity', async () => {
    const wrapper = mount(MediaButton, {
      props: { title: 'Файл', disabled: true },
      slots: { default: '<span>Отчёт</span><small>PDF</small>' },
    })
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
    await wrapper.setProps({ disabled: false })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
    expect(wrapper.text()).toContain('Отчёт')
  })
})
