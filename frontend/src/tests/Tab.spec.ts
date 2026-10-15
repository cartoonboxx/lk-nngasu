import { PhHouse } from '@phosphor-icons/vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import Tab from '@/components/UI/tabs/Tab/Tab.vue'

describe('Tab', () => {
  it('renders text and an optional image from props', () => {
    const wrapper = mount(Tab, {
      props: {
        text: 'Расписание',
        imageSrc: '/schedule.png',
        imageAlt: 'Календарь',
      },
    })

    expect(wrapper.get('[role="tab"]').text()).toBe('Расписание')
    expect(wrapper.get('img').attributes()).toMatchObject({
      src: '/schedule.png',
      alt: 'Календарь',
    })
  })

  it('renders actions outside the tab button', () => {
    const wrapper = mount(Tab, {
      props: { text: 'Предметы' },
      slots: {
        actions: '<button data-action type="button">Меню</button>',
      },
    })

    expect(wrapper.find('[role="tab"] [data-action]').exists()).toBe(false)
    expect(wrapper.get('[data-action]').text()).toBe('Меню')
  })

  it('renders an icon from a prop without an icon slot', () => {
    const wrapper = mount(Tab, {
      props: {
        text: 'Главная',
        icon: PhHouse,
      },
    })

    expect(wrapper.find('[role="tab"] svg').exists()).toBe(true)
  })

  it('exposes active and disabled states to the native tab', () => {
    const wrapper = mount(Tab, {
      props: {
        text: 'Оценки',
        active: true,
        disabled: true,
      },
    })
    const tab = wrapper.get('[role="tab"]')

    expect(tab.attributes('aria-selected')).toBe('true')
    expect((tab.element as HTMLButtonElement).disabled).toBe(true)
  })
})
