import { h } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'

const wrappers: ReturnType<typeof mount>[] = []

afterEach(() => {
  wrappers.forEach((wrapper) => wrapper.unmount())
  wrappers.length = 0
})

describe('Tooltip', () => {
  it('wraps a component without intercepting clicks and shows content on hover', async () => {
    const onClick = vi.fn<() => void>()
    const wrapper = mount(Tooltip, {
      props: { text: 'Описание действия', position: 'right' },
      slots: { default: () => h(Button, { onClick }, () => 'Действие') },
      attachTo: document.body,
    })
    wrappers.push(wrapper)

    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    await wrapper.trigger('mouseenter')
    expect(document.querySelector('[role="tooltip"]')?.textContent).toBe('Описание действия')
    expect(wrapper.element.contains(document.querySelector('[role="tooltip"]'))).toBe(false)
    await wrapper.get('button').trigger('click')
    expect(onClick).toHaveBeenCalledOnce()
    await wrapper.trigger('mouseleave')
    await flushPromises()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
  })

  it('ignores focus, shows only on hover, and dismisses on Escape', async () => {
    const wrapper = mount(Tooltip, {
      props: { text: 'Подсказка' },
      slots: { default: '<button aria-describedby="existing">Кнопка</button>' },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    const button = wrapper.get('button')
    ;(button.element as HTMLButtonElement).focus()
    await flushPromises()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    expect(button.attributes('aria-describedby')).toBe('existing')

    await wrapper.trigger('mouseenter')
    expect(document.querySelector('[role="tooltip"]')).not.toBeNull()
    await wrapper.trigger('mouseleave')
    await flushPromises()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()

    await wrapper.trigger('mouseenter')

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    expect(button.attributes('aria-describedby')).toBe('existing')

    ;(button.element as HTMLButtonElement).blur()
    await flushPromises()
    ;(button.element as HTMLButtonElement).focus()
    await flushPromises()
    expect(document.querySelector('[role="tooltip"]')).toBeNull()

    await wrapper.trigger('mouseleave')
    await wrapper.trigger('mouseenter')
    expect(document.querySelector('[role="tooltip"]')).not.toBeNull()
  })

  it('supports markup in the content slot and can be disabled', async () => {
    const wrapper = mount(Tooltip, {
      props: { disabled: true },
      slots: { default: '<span>Текст</span>', content: '<strong>Подсказка</strong>' },
      attachTo: document.body,
    })
    wrappers.push(wrapper)
    await wrapper.trigger('mouseenter')
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
    await wrapper.setProps({ disabled: false })
    expect(document.querySelector('[role="tooltip"] strong')?.textContent).toBe('Подсказка')
    await wrapper.setProps({ disabled: true })
    expect(document.querySelector('[role="tooltip"]')).toBeNull()
  })
})
