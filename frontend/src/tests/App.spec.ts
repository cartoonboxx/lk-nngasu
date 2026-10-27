import { afterEach, beforeEach, describe, it, expect, vi } from 'vitest'

import { mount } from '@vue/test-utils'
import App from '../App.vue'
import router from '../router'
import stores from '@/stores'

describe('App', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'matchMedia',
      vi.fn(() => ({
        matches: false,
        addEventListener: vi.fn<MediaQueryList['addEventListener']>(),
        removeEventListener: vi.fn<MediaQueryList['removeEventListener']>(),
      })),
    )
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('mounts renders properly', async () => {
    router.push('/ui-kit')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [stores, router],
      },
    })

    expect(wrapper.text()).toContain('Компоненты интерфейса')
    wrapper.unmount()
  })

  it('renders nested page content inside the home layout', async () => {
    await router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(App, {
      global: {
        plugins: [stores, router],
      },
    })

    expect(wrapper.get('h1').text()).toBe('Главная')

    await router.push('/schedule')
    await wrapper.vm.$nextTick()

    expect(wrapper.get('h1').text()).toBe('Расписание')
    wrapper.unmount()
  })
})
