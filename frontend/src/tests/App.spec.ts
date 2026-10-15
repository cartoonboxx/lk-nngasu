import { describe, it, expect } from 'vitest'

import { mount } from '@vue/test-utils'
import App from '../App.vue'
import router from '../router'

describe('App', () => {
  it('mounts renders properly', async () => {
    router.push('/ui-kit')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('Компоненты интерфейса')
  })

  it('renders nested page content inside the home layout', async () => {
    await router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.get('h1').text()).toBe('Главная')

    await router.push('/schedule')
    await wrapper.vm.$nextTick()

    expect(wrapper.get('h1').text()).toBe('Расписание')
  })
})
