import { PhHouse } from '@phosphor-icons/vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'

describe('StyleTab', () => {
  it('renders text and an optional image from props', () => {
    const wrapper = mount(StyleTab, {
      props: {
        text: 'Расписание',
        imageSrc: '/schedule.png',
        imageAlt: 'Календарь',
      },
    })

    expect(wrapper.get('[role="option"]').text()).toBe('Расписание')
    expect(wrapper.get('img').attributes()).toMatchObject({
      src: '/schedule.png',
      alt: 'Календарь',
    })
  })

  it('renders actions outside the tab button', () => {
    const wrapper = mount(StyleTab, {
      props: { text: 'Предметы' },
      slots: {
        actions: '<button data-action type="button">Меню</button>',
      },
    })

    expect(wrapper.find('[role="option"] [data-action]').exists()).toBe(false)
    expect(wrapper.get('[data-action]').text()).toBe('Меню')
  })

  it('renders an icon from a prop without an icon slot', () => {
    const wrapper = mount(StyleTab, {
      props: {
        text: 'Главная',
        icon: PhHouse,
      },
    })

    expect(wrapper.find('[role="option"] svg').exists()).toBe(true)
  })

  it('exposes active and disabled states to the native tab', () => {
    const wrapper = mount(StyleTab, {
      props: {
        text: 'Оценки',
        active: true,
        disabled: true,
      },
    })
    const tab = wrapper.get('[role="option"]')

    expect(tab.attributes('aria-selected')).toBe('true')
    expect((tab.element as HTMLButtonElement).disabled).toBe(true)
  })

  it('forwards listbox semantics to the native button and selects through it', async () => {
    const wrapper = mount(StyleTab, {
      props: {
        text: 'Предметы',
        triggerAttrs: {
          id: 'subject-option',
          role: 'option',
          tabindex: -1,
          'aria-controls': 'subjects',
        },
        active: true,
      },
    })
    const button = wrapper.get('button')
    expect(button.attributes()).toMatchObject({
      id: 'subject-option',
      role: 'option',
      tabindex: '-1',
      'aria-controls': 'subjects',
      'aria-selected': 'true',
    })
    expect(button.find('svg').exists()).toBe(false)
    await button.trigger('click')
    expect(wrapper.emitted('select')).toHaveLength(1)
  })
})
